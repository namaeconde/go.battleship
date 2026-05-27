package game

import (
	"errors"
)

const BoardSize = 10

// Board represents a 10x10 game board.
type Board [BoardSize][BoardSize]CellState

// NewBoard creates and returns a new 10x10 board initialized with Water cells.
func NewBoard() Board {
	board := Board{}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			board[r][c] = Water
		}
	}
	return board
}

// IsValidCoordinate checks if a coordinate is within board boundaries.
func IsValidCoordinate(coord Coordinate) bool {
	return coord.Row >= 0 && coord.Row < BoardSize &&
		coord.Col >= 0 && coord.Col < BoardSize
}

// PlaceShip attempts to place a ship on the board.
// It returns an error if the placement is invalid (out of bounds, overlaps, etc.).
func (b *Board) PlaceShip(ship Ship) error {
	// Validate placement
	for _, coord := range ship.Coordinates {
		if !IsValidCoordinate(coord) {
			return errors.New("ship out of bounds")
		}
		if b[coord.Row][coord.Col] != Water {
			return errors.New("ship overlap detected")
		}
	}

	// Place the ship
	for _, coord := range ship.Coordinates {
		b[coord.Row][coord.Col] = ActiveShip
	}
	return nil
}

// ApplyShot processes a shot at a given coordinate and returns the result and the type of ship hit (if any).
func (b *Board) ApplyShot(coord Coordinate) (CellState, error) {
	if !IsValidCoordinate(coord) {
		return Water, errors.New("shot out of bounds")
	}

	switch b[coord.Row][coord.Col] {
	case Hit, Miss, SunkShip:
		return Water, errors.New("coordinate already targeted")
	case Water:
		b[coord.Row][coord.Col] = Miss
		return Miss, nil
	case ActiveShip:
		b[coord.Row][coord.Col] = Hit
		return Hit, nil
	}
	return Water, nil
}

// UpdateCellsAsSunk marks all cells of a specific ship type as SunkShip.
func (b *Board) UpdateCellsAsSunk(shipCoords []Coordinate) {
	for _, coord := range shipCoords {
		if IsValidCoordinate(coord) {
			b[coord.Row][coord.Col] = SunkShip
		}
	}
}
