package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// CellState represents the state of a single cell on the game board.
type CellState int

const (
	Water CellState = iota
	StateShip
	Hit
	Miss
	SunkShip
)

// ShipType represents the type of a ship.
type ShipType int

const (
	Carrier ShipType = iota // Size 5
	Battleship              // Size 4
	Cruiser                 // Size 3
	Submarine               // Size 3
	Destroyer               // Size 2
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
	case StateShip:
		return "Ship"
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
	switch o {
	case Horizontal:
		return "Horizontal"
	case Vertical:
		return "Vertical"
	default:
		return "Unknown"
	}
}

// String returns the string representation of a GamePhase.
func (gp GamePhase) String() string {
	switch gp {
	case PhaseConnection:
		return "Connection"
	case PhasePlacement:
		return "Placement"
	case PhaseBattle:
		return "Battle"
	case PhaseGameOver:
		return "GameOver"
	default:
		return "Unknown"
	}
}

// String returns the string representation of a Coordinate (e.g., "A1").
func (c Coordinate) String() string {
	return fmt.Sprintf("%c%d", 'A'+c.Col, c.Row+1)
}

// ParseCoordinate parses a string like "A1" into a Coordinate.
// Returns an error if the input is invalid or out of bounds (A-J, 1-10).
func ParseCoordinate(s string) (Coordinate, error) {
	if len(s) < 2 || len(s) > 3 {
		return Coordinate{}, errors.New("invalid coordinate format")
	}

	colChar := strings.ToUpper(string(s[0]))[0]
	if colChar < 'A' || colChar > 'J' {
		return Coordinate{}, errors.New("column out of bounds (A-J)")
	}

	rowVal, err := strconv.Atoi(s[1:])
	if err != nil || rowVal < 1 || rowVal > 10 {
		return Coordinate{}, errors.New("row out of bounds (1-10)")
	}

	return Coordinate{
		Row: rowVal - 1,
		Col: int(colChar - 'A'),
	}, nil
}

// ParseShipType parses a string into a ShipType.
// Returns -1 if the string does not match any known ShipType.
func ParseShipType(s string) ShipType {
	switch strings.ToLower(s) {
	case "carrier":
		return Carrier
	case "battleship":
		return Battleship
	case "cruiser":
		return Cruiser
	case "submarine":
		return Submarine
	case "destroyer":
		return Destroyer
	default:
		return -1
	}
}

// UIEvent is the interface all game→UI events implement.
type UIEvent interface{ uiEvent() }

// MessageEvent carries a status/info message to display in the UI.
type MessageEvent struct{ Text string }

// PhaseChangedEvent notifies the UI of a game phase transition.
type PhaseChangedEvent struct{ Phase GamePhase }

// ShotResultEvent notifies the UI of a shot result to animate.
// Board is "fleet" for incoming opponent shots, "tracking" for outgoing shots.
type ShotResultEvent struct {
	Coord    Coordinate
	Result   string // "hit", "miss", or "sunk"
	ShipType string // ship name when Result == "sunk", empty otherwise
	Board    string // "fleet" | "tracking"
}

// GameOverEvent notifies the UI that the game has ended.
type GameOverEvent struct{ Winner string }

// ReplayEvent notifies the UI that both players agreed to replay.
type ReplayEvent struct{}

// QuitEvent tells the UI to shut down cleanly.
type QuitEvent struct{}

func (MessageEvent) uiEvent()      {}
func (PhaseChangedEvent) uiEvent() {}
func (ShotResultEvent) uiEvent()   {}
func (GameOverEvent) uiEvent()     {}
func (ReplayEvent) uiEvent()       {}
func (QuitEvent) uiEvent()         {}
