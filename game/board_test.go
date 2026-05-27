package game

import (
	"reflect"
	"testing"
)

// TestNewBoard ensures a new board is correctly initialized with Water cells.
func TestNewBoard(t *testing.T) {
	board := NewBoard()
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if board[r][c] != Water {
				t.Errorf("NewBoard cell (%d,%d) is %v, expected Water", r, c, board[r][c])
			}
		}
	}
}

// TestPlaceShip_Valid ensures a ship can be placed in a valid position.
func TestPlaceShip_Valid(t *testing.T) {
	board := NewBoard()
	ship := Ship{
		Type:        Destroyer,
		Size:        2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{Row: 0, Col: 0}, {Row: 0, Col: 1}},
		Hits:        make([]bool, 2),
	}

	err := board.PlaceShip(ship)
	if err != nil {
		t.Fatalf("PlaceShip failed for valid placement: %v", err)
	}

	if board[0][0] != ActiveShip || board[0][1] != ActiveShip {
		t.Errorf("Ship not placed correctly. Cells are %v, %v, expected Ship", board[0][0], board[0][1])
	}
}

// TestPlaceShip_Overlap ensures a ship cannot overlap with another.
func TestPlaceShip_Overlap(t *testing.T) {
	board := NewBoard()
	ship1 := Ship{
		Type:        Destroyer,
		Size:        2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{Row: 0, Col: 0}, {Row: 0, Col: 1}},
		Hits:        make([]bool, 2),
	}
	// Place ship1
	_ = board.PlaceShip(ship1)

	ship2 := Ship{
		Type:        Destroyer,
		Size:        2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{Row: 0, Col: 0}, {Row: 0, Col: 1}}, // Overlapping
		Hits:        make([]bool, 2),
	}

	err := board.PlaceShip(ship2)
	if err == nil {
		t.Fatalf("PlaceShip did not return error for overlapping ship")
	}
	if err.Error() != "ship overlap detected" {
		t.Errorf("Expected 'ship overlap detected' error, got '%v'", err.Error())
	}
}

// TestPlaceShip_OutOfBounds ensures a ship cannot be placed out of bounds.
func TestPlaceShip_OutOfBounds(t *testing.T) {
	board := NewBoard()
	ship := Ship{
		Type:        Destroyer,
		Size:        2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{Row: 0, Col: 9}, {Row: 0, Col: 10}}, // Out of bounds
		Hits:        make([]bool, 2),
	}

	err := board.PlaceShip(ship)
	if err == nil {
		t.Fatalf("PlaceShip did not return error for out of bounds ship")
	}
	if err.Error() != "ship out of bounds" {
		t.Errorf("Expected 'ship out of bounds' error, got '%v'", err.Error())
	}
}

// TestApplyShot ensures a shot correctly updates the board and returns result.
func TestApplyShot(t *testing.T) {
	board := NewBoard()
	ship := Ship{
		Type:        Destroyer,
		Size:        2,
		Orientation: Horizontal,
		Coordinates: []Coordinate{{Row: 0, Col: 0}, {Row: 0, Col: 1}},
		Hits:        make([]bool, 2),
	}
	// Place ship for testing hits
	_ = board.PlaceShip(ship)

	// Test a hit
	result, err := board.ApplyShot(Coordinate{Row: 0, Col: 0})
	if err != nil {
		t.Fatalf("ApplyShot failed for valid hit: %v", err)
	}
	if result != Hit {
		t.Errorf("Expected Hit, got %v", result)
	}
	if board[0][0] != Hit {
		t.Errorf("Board cell not updated to Hit, got %v", board[0][0])
	}

	// Test a miss
	result, err = board.ApplyShot(Coordinate{Row: 9, Col: 9})
	if err != nil {
		t.Fatalf("ApplyShot failed for valid miss: %v", err)
	}
	if result != Miss {
		t.Errorf("Expected Miss, got %v", result)
	}
	if board[9][9] != Miss {
		t.Errorf("Board cell not updated to Miss, got %v", board[9][9])
	}

	// Test hitting an already hit cell
	result, err = board.ApplyShot(Coordinate{Row: 0, Col: 0})
	if err == nil {
		t.Fatalf("ApplyShot did not return error for already targeted cell")
	}
	if err.Error() != "coordinate already targeted" {
		t.Errorf("Expected 'coordinate already targeted' error, got '%v'", err.Error())
	}
}

// TestNewPlayer ensures a new player is correctly initialized.
func TestNewPlayer(t *testing.T) {
	player := NewPlayer("Player1")
	if player.Name != "Player1" {
		t.Errorf("Player name expected 'Player1', got '%s'", player.Name)
	}
	if len(player.Ships) != len(DefaultShips) {
		t.Errorf("Expected %d default ships, got %d", len(DefaultShips), len(player.Ships))
	}
	if player.ShipsRemaining != len(DefaultShips) {
		t.Errorf("Expected ShipsRemaining %d, got %d", len(DefaultShips), player.ShipsRemaining)
	}
	// Deep equality check for boards (initially empty water)
	expectedBoard := NewBoard()
	if !reflect.DeepEqual(player.PlayerBoard, expectedBoard) {
		t.Errorf("PlayerBoard not initialized correctly")
	}
	if !reflect.DeepEqual(player.TrackingBoard, expectedBoard) {
		t.Errorf("TrackingBoard not initialized correctly")
	}
}

// TestMarkShipSunk ensures that sunk ship state is correctly updated.
func TestMarkShipSunk(t *testing.T) {
	player := NewPlayer("Player1")
	// Find a ship to modify, e.g., Destroyer
	var destroyer *Ship
	for i := range player.Ships {
		if player.Ships[i].Type == Destroyer {
			destroyer = &player.Ships[i]
			break
		}
	}
	if destroyer == nil {
		t.Fatalf("Destroyer not found in player ships")
	}

	// Simulate placing the ship
	destroyer.Coordinates = []Coordinate{{Row: 0, Col: 0}, {Row: 0, Col: 1}}
	_ = player.PlayerBoard.PlaceShip(*destroyer) // Place on board

	// Simulate hits to sink it
	player.RecordHit(Coordinate{Row: 0, Col: 0})
	player.RecordHit(Coordinate{Row: 0, Col: 1})

	if player.ShipsRemaining != len(DefaultShips)-1 {
		t.Errorf("Expected ShipsRemaining %d, got %d after sinking one", len(DefaultShips)-1, player.ShipsRemaining)
	}

	// Verify cells on PlayerBoard are updated to SunkShip
	for _, coord := range destroyer.Coordinates {
		if player.PlayerBoard[coord.Row][coord.Col] != SunkShip {
			t.Errorf("Ship cell (%d,%d) for sunk Destroyer expected SunkShip, got %v", coord.Row, coord.Col, player.PlayerBoard[coord.Row][coord.Col])
		}
	}
}

// TestRemoveShip ensures a ship is correctly removed from the board.
func TestRemoveShip(t *testing.T) {
	player := NewPlayer("Player1")
	ship := player.GetShipByType(Destroyer)
	if ship == nil {
		t.Fatalf("Destroyer not found")
	}
	ship.Coordinates = []Coordinate{{0, 0}, {0, 1}}
	_ = player.PlayerBoard.PlaceShip(*ship)

	player.RemoveShip(Destroyer)

	if player.PlayerBoard[0][0] != Water || player.PlayerBoard[0][1] != Water {
		t.Errorf("Expected ship cells to be Water, got %v, %v", player.PlayerBoard[0][0], player.PlayerBoard[0][1])
	}
	if len(ship.Coordinates) != 0 {
		t.Errorf("Ship coordinates not cleared after removal")
	}
}
