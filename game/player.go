package game

import (
	"fmt"
)

// PlayerState holds all data for a single player.
type PlayerState struct {
	Name           string
	PlayerBoard    Board  // Player's own board with their ships
	TrackingBoard  Board  // Player's view of opponent's board (shots fired)
	Ships          []Ship // List of ships for this player
	ShipsRemaining int    // Count of unsunk ships
	ReadyForBattle bool   // True when all ships are placed
}

// NewPlayer creates and initializes a new PlayerState.
func NewPlayer(name string) *PlayerState {
	ships := make([]Ship, len(DefaultShips))
	copy(ships, DefaultShips) // Deep copy default ships

	for i := range ships {
		ships[i].Hits = make([]bool, ships[i].Size) // Initialize hits slice
	}

	return &PlayerState{
		Name:           name,
		PlayerBoard:    NewBoard(),
		TrackingBoard:  NewBoard(),
		Ships:          ships,
		ShipsRemaining: len(ships),
	}
}

// GetShipByType returns a pointer to a ship of the given type, or nil if not found.
func (p *PlayerState) GetShipByType(shipType ShipType) *Ship {
	for i := range p.Ships {
		if p.Ships[i].Type == shipType {
			return &p.Ships[i]
		}
	}
	return nil
}

// RecordHit registers a hit on one of the player's ships.
// It returns the resulting CellState (Hit/Miss), the type of ship if sunk, and any error.
func (p *PlayerState) RecordHit(coord Coordinate) (CellState, ShipType, error) {
	// Apply shot to PlayerBoard
	cellResult, err := p.PlayerBoard.ApplyShot(coord)
	if err != nil {
		return Water, -1, err
	}

	if cellResult == Hit {
		// Find the actual ship that was hit and mark it
		for i := range p.Ships {
			ship := &p.Ships[i]
			for j, shipCoord := range ship.Coordinates {
				if shipCoord == coord {
					if !ship.Hits[j] { // Only record if not already hit
						ship.Hits[j] = true
					}

					// Check if the ship is sunk
					isSunk := true
					for _, hit := range ship.Hits {
						if !hit {
							isSunk = false
							break
						}
					}
					if isSunk {
						p.PlayerBoard.UpdateCellsAsSunk(ship.Coordinates) // Mark cells as SunkShip
						p.ShipsRemaining--
						return SunkShip, ship.Type, nil
					}
					return Hit, ship.Type, nil // Return ship type on hit
				}
			}
		}
	}
	return cellResult, -1, nil // Return result, no ship sunk for misses
}

// RemoveShip removes a ship from a player's board (for re-placement).
func (p *PlayerState) RemoveShip(shipType ShipType) error {
	ship := p.GetShipByType(shipType)
	if ship == nil {
		return fmt.Errorf("ship type %v not found", shipType)
	}
	if len(ship.Coordinates) == 0 {
		return fmt.Errorf("ship %v not placed", shipType)
	}

	// Clear cells on the board
	for _, coord := range ship.Coordinates {
		if IsValidCoordinate(coord) { // Defensive check
			p.PlayerBoard[coord.Row][coord.Col] = Water
		}
	}
	ship.Coordinates = nil              // Clear coordinates
	ship.Hits = make([]bool, ship.Size) // Reset hits
	p.ReadyForBattle = false            // Player is no longer ready if a ship is removed
	return nil
}
