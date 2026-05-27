package game

import (
	"errors"
	"fmt"
)

// PlayerState represents the state of a player in the game.
type PlayerState struct {
	Name  string
	Ships []Ship
	Board *Board
}

// NewPlayer initializes a player with the default set of ships and a new board.
func NewPlayer(name string) *PlayerState {
	ships := make([]Ship, len(DefaultShips))
	copy(ships, DefaultShips)
	
	// Ensure Hits slice is initialized for each ship
	for i := range ships {
		ships[i].Hits = make([]bool, ships[i].Size)
	}

	return &PlayerState{
		Name:  name,
		Ships: ships,
		Board: NewBoard(),
	}
}

// GetShipByType returns a pointer to the ship of the given type, or nil if not found.
func (p *PlayerState) GetShipByType(t ShipType) *Ship {
	for i := range p.Ships {
		if p.Ships[i].Type == t {
			return &p.Ships[i]
		}
	}
	return nil
}

// RecordHit updates the state when a coordinate is hit.
// It checks if a ship was hit, updates its Hits status, and returns if it was sunk.
func (p *PlayerState) RecordHit(coord Coordinate) (CellState, ShipType, error) {
	// Apply shot to the board
	state, err := p.Board.ApplyShot(coord)
	if err != nil {
		return Water, -1, err
	}

	if state == Miss {
		return Miss, -1, nil
	}

	// Find which ship was hit
	for i := range p.Ships {
		ship := &p.Ships[i]
		for j, shipCoord := range ship.Coordinates {
			if shipCoord == coord {
				ship.Hits[j] = true
				
				// Check if sunk
				isSunk := true
				for _, h := range ship.Hits {
					if !h {
						isSunk = false
						break
					}
				}

				if isSunk {
					p.Board.UpdateCellsAsSunk(ship.Coordinates)
					return SunkShip, ship.Type, nil
				}
				return Hit, ship.Type, nil
			}
		}
	}

	// This should not be reachable if Board.ApplyShot returned Hit
	return Hit, -1, fmt.Errorf("ship not found at coordinate %v", coord)
}

// RemoveShip clears a ship's coordinates from the board and the ship's own state.
func (p *PlayerState) RemoveShip(t ShipType) error {
	ship := p.GetShipByType(t)
	if ship == nil {
		return errors.New("ship type not found")
	}

	// Clear coordinates from board
	for _, coord := range ship.Coordinates {
		if coord.Row >= 0 && coord.Row < 10 && coord.Col >= 0 && coord.Col < 10 {
			p.Board.Grid[coord.Row][coord.Col] = Water
		}
	}

	// Clear coordinates from ship
	ship.Coordinates = nil
	ship.Hits = make([]bool, ship.Size)

	return nil
}
