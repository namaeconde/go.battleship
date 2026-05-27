package game

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	b := NewBoard()
	for r := 0; r < 10; r++ {
		for c := 0; c < 10; c++ {
			if b.Grid[r][c] != Water {
				t.Errorf("Expected Water at %d,%d, got %v", r, c, b.Grid[r][c])
			}
		}
	}
}

func TestPlaceShip(t *testing.T) {
	b := NewBoard()
	ship := Ship{
		Type:      Destroyer,
		Size:      2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{0, 0}, {0, 1}},
	}

	err := b.PlaceShip(ship)
	if err != nil {
		t.Errorf("Unexpected error placing ship: %v", err)
	}

	if b.Grid[0][0] != StateShip || b.Grid[0][1] != StateShip {
		t.Errorf("Ship not placed correctly on grid")
	}

	// Overlap
	err = b.PlaceShip(ship)
	if err == nil {
		t.Error("Expected error when placing overlapping ship")
	}

	// Out of bounds
	badShip := Ship{
		Type: Destroyer,
		Size: 2,
		Coordinates: []Coordinate{{9, 9}, {9, 10}},
	}
	err = b.PlaceShip(badShip)
	if err == nil {
		t.Error("Expected error when placing ship out of bounds")
	}
}

func TestApplyShot(t *testing.T) {
	b := NewBoard()
	ship := Ship{
		Type:      Destroyer,
		Size:      2,
		Coordinates: []Coordinate{{0, 0}, {0, 1}},
	}
	_ = b.PlaceShip(ship)

	// Hit
	res, err := b.ApplyShot(Coordinate{0, 0})
	if err != nil || res != Hit {
		t.Errorf("Expected Hit, got %v, err: %v", res, err)
	}

	// Miss
	res, err = b.ApplyShot(Coordinate{1, 1})
	if err != nil || res != Miss {
		t.Errorf("Expected Miss, got %v, err: %v", res, err)
	}

	// Already shot (Hit)
	_, err = b.ApplyShot(Coordinate{0, 0})
	if err == nil {
		t.Error("Expected error when shooting same coordinate twice (Hit)")
	}

	// Already shot (Miss)
	_, err = b.ApplyShot(Coordinate{1, 1})
	if err == nil {
		t.Error("Expected error when shooting same coordinate twice (Miss)")
	}
}

func TestUpdateCellsAsSunk(t *testing.T) {
	b := NewBoard()
	coords := []Coordinate{{0, 0}, {0, 1}}
	for _, c := range coords {
		b.Grid[c.Row][c.Col] = StateShip
	}

	b.UpdateCellsAsSunk(coords)
	for _, c := range coords {
		if b.Grid[c.Row][c.Col] != SunkShip {
			t.Errorf("Expected SunkShip at %v, got %v", c, b.Grid[c.Row][c.Col])
		}
	}
}

func TestNewPlayer(t *testing.T) {
	p := NewPlayer("Alice")
	if p.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", p.Name)
	}
	if len(p.Ships) != 5 {
		t.Errorf("Expected 5 ships, got %d", len(p.Ships))
	}
}

func TestGetShipByType(t *testing.T) {
	p := NewPlayer("Alice")
	ship := p.GetShipByType(Carrier)
	if ship == nil || ship.Type != Carrier {
		t.Errorf("Expected Carrier ship")
	}

	ship = p.GetShipByType(ShipType(-1))
	if ship != nil {
		t.Errorf("Expected nil for invalid ship type")
	}
}

func TestRecordHit(t *testing.T) {
	p := NewPlayer("Alice")
	// Place a ship manually for testing
	ship := p.GetShipByType(Destroyer)
	ship.Coordinates = []Coordinate{{0, 0}, {0, 1}}
	ship.Hits = make([]bool, ship.Size)
	for _, c := range ship.Coordinates {
		p.Board.Grid[c.Row][c.Col] = StateShip
	}

	// First hit
	state, shipType, err := p.RecordHit(Coordinate{0, 0})
	if err != nil || state != Hit || shipType != Destroyer {
		t.Errorf("Expected Hit on Destroyer, got state=%v, shipType=%v, err=%v", state, shipType, err)
	}

	// Second hit (sinking)
	state, shipType, err = p.RecordHit(Coordinate{0, 1})
	if err != nil || state != SunkShip || shipType != Destroyer {
		t.Errorf("Expected SunkShip on Destroyer, got state=%v, shipType=%v, err=%v", state, shipType, err)
	}
	
	if p.Board.Grid[0][0] != SunkShip || p.Board.Grid[0][1] != SunkShip {
		t.Errorf("Board not updated to SunkShip")
	}
}

func TestRemoveShip(t *testing.T) {
	p := NewPlayer("Alice")
	ship := p.GetShipByType(Destroyer)
	ship.Coordinates = []Coordinate{{0, 0}, {0, 1}}
	for _, c := range ship.Coordinates {
		p.Board.Grid[c.Row][c.Col] = StateShip
	}

	err := p.RemoveShip(Destroyer)
	if err != nil {
		t.Errorf("Unexpected error removing ship: %v", err)
	}

	if p.Board.Grid[0][0] != Water || p.Board.Grid[0][1] != Water {
		t.Error("Board not cleared after removing ship")
	}
	if len(ship.Coordinates) != 0 {
		t.Error("Ship coordinates not cleared")
	}
}
