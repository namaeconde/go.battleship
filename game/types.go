package game

// CellState represents the state of a single cell on the game board.
type CellState int

const (
	Water CellState = iota
	ActiveShip
	Hit
	Miss
	SunkShip
)

// ShipType represents the type of a ship.
type ShipType int

const (
	Carrier    ShipType = iota // Size 5
	Battleship                 // Size 4
	Cruiser                    // Size 3
	Submarine                  // Size 3
	Destroyer                  // Size 2
)

// Ship represents a single ship with its properties.
type Ship struct {
	Type        ShipType
	Size        int
	Coordinates []Coordinate // Slice of (row, col) pairs the ship occupies
	Hits        []bool       // Tracks which parts of the ship have been hit. len(Hits) == Size
	Orientation Orientation
}

// Orientation for ship placement.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Coordinate represents a row and column pair on the board.
type Coordinate struct {
	Row int
	Col int
}

// GamePhase represents the current phase of the game.
type GamePhase int

const (
	PhaseConnection GamePhase = iota
	PhasePlacement
	PhaseBattle
	PhaseGameOver
)

// DefaultShips defines the standard set of ships for Battleship.
var DefaultShips = []Ship{
	{Type: Carrier, Size: 5},
	{Type: Battleship, Size: 4},
	{Type: Cruiser, Size: 3},
	{Type: Submarine, Size: 3},
	{Type: Destroyer, Size: 2},
}

// String returns the string representation of a CellState.
func (cs CellState) String() string {
	switch cs {
	case Water:
		return "Water"
	case ActiveShip:
		return "ActiveShip"
	case Hit:
		return "Hit"
	case Miss:
		return "Miss"
	case SunkShip:
		return "SunkShip"
	default:
		return "Unknown"
	}
}

// String returns the string representation of a ShipType.
func (st ShipType) String() string {
	switch st {
	case Carrier:
		return "Carrier"
	case Battleship:
		return "Battleship"
	case Cruiser:
		return "Cruiser"
	case Submarine:
		return "Submarine"
	case Destroyer:
		return "Destroyer"
	default:
		return "Unknown"
	}
}

// String returns the string representation of an Orientation.
func (o Orientation) String() string {
	if o == Horizontal {
		return "Horizontal"
	}
	return "Vertical"
}
