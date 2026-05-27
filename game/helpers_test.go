package game

import "testing"

func TestCoordinateString(t *testing.T) {
	c := Coordinate{Row: 0, Col: 0}
	if c.String() != "A1" {
		t.Errorf("Expected A1, got %s", c.String())
	}
	c = Coordinate{Row: 9, Col: 9}
	if c.String() != "J10" {
		t.Errorf("Expected J10, got %s", c.String())
	}
}

func TestParseCoordinate(t *testing.T) {
	tests := []struct {
		input    string
		expected Coordinate
		wantErr  bool
	}{
		{"A1", Coordinate{0, 0}, false},
		{"J10", Coordinate{9, 9}, false},
		{"B5", Coordinate{4, 1}, false},
		{"A0", Coordinate{}, true},
		{"K1", Coordinate{}, true},
		{"A11", Coordinate{}, true},
		{"", Coordinate{}, true},
		{"Z99", Coordinate{}, true},
	}
	for _, tt := range tests {
		got, err := ParseCoordinate(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseCoordinate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ParseCoordinate(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseShipType(t *testing.T) {
	if ParseShipType("Carrier") != Carrier {
		t.Error("Failed to parse Carrier")
	}
	if ParseShipType("Battleship") != Battleship {
		t.Error("Failed to parse Battleship")
	}
	if ParseShipType("Cruiser") != Cruiser {
		t.Error("Failed to parse Cruiser")
	}
	if ParseShipType("Submarine") != Submarine {
		t.Error("Failed to parse Submarine")
	}
	if ParseShipType("Destroyer") != Destroyer {
		t.Error("Failed to parse Destroyer")
	}
	if ParseShipType("Invalid") != -1 {
		t.Error("Should return -1 for invalid ship type")
	}
}
