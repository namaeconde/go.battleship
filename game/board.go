package game

import (
	"errors"
)

// Board represents the 10x10 game board.
type Board struct {
	Grid [10][10]CellState
}

// NewBoard returns an initialized 10x10 grid with all cells set to Water.
func NewBoard() *Board {
	return &Board{}
}

// PlaceShip validates and places a ship on the board.
func (b *Board) PlaceShip(ship Ship) error {
	// Check if all coordinates are within bounds and not occupied
	for _, coord := range ship.Coordinates {
		if coord.Row < 0 || coord.Row >= 10 || coord.Col < 0 || coord.Col >= 10 {
			return errors.New("ship coordinate out of bounds")
		}
		if b.Grid[coord.Row][coord.Col] != Water {
			return errors.New("ship coordinate already occupied")
		}
	}

	// Place the ship
	for _, coord := range ship.Coordinates {
		b.Grid[coord.Row][coord.Col] = StateShip
	}

	return nil
}

// ApplyShot processes a shot result at the given coordinate.
// Returns Hit, Miss, or an error if the shot is invalid (out of bounds or already shot).
func (b *Board) ApplyShot(coord Coordinate) (CellState, error) {
	if coord.Row < 0 || coord.Row >= 10 || coord.Col < 0 || coord.Col >= 10 {
		return Water, errors.New("coordinate out of bounds")
	}

	state := b.Grid[coord.Row][coord.Col]
	switch state {
	case Water:
		b.Grid[coord.Row][coord.Col] = Miss
		return Miss, nil
	case StateShip:
		b.Grid[coord.Row][coord.Col] = Hit
		return Hit, nil
	case Hit, Miss, SunkShip:
		return state, errors.New("coordinate already shot")
	default:
		return Water, errors.New("unknown cell state")
	}
}

// UpdateCellsAsSunk marks the given coordinates as SunkShip.
func (b *Board) UpdateCellsAsSunk(coords []Coordinate) {
	for _, coord := range coords {
		if coord.Row >= 0 && coord.Row < 10 && coord.Col >= 0 && coord.Col < 10 {
			b.Grid[coord.Row][coord.Col] = SunkShip
		}
	}
}
