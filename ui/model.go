// ui/model.go
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.battleship/game"
	"go.battleship/network"
)

// animState tracks a single in-progress hit/miss animation.
type animState struct {
	active bool
	frame  int    // 0–3; frame 3 is the "settled" state
	coord  game.Coordinate
	board  string // "fleet" (incoming) | "tracking" (outgoing)
	result string // "hit", "miss", or "sunk"
}

// Model is the bubbletea model for the Battleship TUI.
type Model struct {
	gs      *game.GameState
	program *tea.Program // set via SetProgram before p.Run()

	// Input / placement state
	cursor      game.Coordinate
	targetCoord *game.Coordinate
	currentShip *game.Ship
	orientation game.Orientation

	// UI state
	message   string
	lastPhase game.GamePhase

	// Animation state
	anim animState
}

// NewModel creates a Model and the underlying GameState.
// SetProgram must be called before p.Run().
func NewModel(conn game.NetworkConn, host bool) *Model {
	localPlayerName, remotePlayerName := "Joiner", "Host"
	if host {
		localPlayerName, remotePlayerName = "Host", "Joiner"
	}

	m := &Model{
		orientation: game.Horizontal,
	}
	m.gs = game.NewGame(localPlayerName, remotePlayerName, m)
	m.gs.Connection = conn
	m.currentShip = m.gs.LocalPlayer.GetShipByType(game.Carrier)
	return m
}

// SetProgram wires the bubbletea Program into the model so Send() works.
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// Send implements game.UIInterface — forwards UIEvents into the bubbletea event loop.
func (m *Model) Send(event game.UIEvent) {
	if m.program != nil {
		m.program.Send(event)
	}
}

// Close shuts down the game network connection.
func (m *Model) Close() {
	if m.gs != nil {
		m.gs.Close()
	}
}

// Init starts the game loop goroutines and transitions to the placement phase.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		m.gs.TransitionPhase(game.PhasePlacement)
		m.gs.StartGameLoop()
		return nil
	}
}

// Update handles incoming messages (key events, UIEvents, ticks).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// --- UIEvents from game goroutines ---
	case game.MessageEvent:
		m.message = msg.Text
		return m, nil

	case game.PhaseChangedEvent:
		m.lastPhase = msg.Phase
		if msg.Phase == game.PhasePlacement {
			// Replay reset: re-fetch ship pointer from refreshed slice.
			if len(m.gs.LocalPlayer.Ships) > 0 {
				m.currentShip = &m.gs.LocalPlayer.Ships[0]
			}
		} else if msg.Phase == game.PhaseBattle {
			m.currentShip = nil
		}
		return m, nil

	case game.ShotResultEvent:
		m.anim = animState{
			active: true,
			frame:  0,
			coord:  msg.Coord,
			board:  msg.Board,
			result: msg.Result,
		}
		return m, tickCmd()

	case game.GameOverEvent:
		m.gs.Winner = msg.Winner
		return m, nil

	case game.ReplayEvent:
		return m, nil

	case game.QuitEvent:
		return m, tea.Quit

	// --- Animation tick ---
	case tickMsg:
		if m.anim.active {
			m.anim.frame++
			if m.anim.frame >= 3 {
				m.anim.active = false
			} else {
				return m, tickCmd()
			}
		}
		return m, nil

	// --- Keyboard ---
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyUp:
		if m.cursor.Row > 0 {
			m.cursor.Row--
		}
	case tea.KeyDown:
		if m.cursor.Row < 9 {
			m.cursor.Row++
		}
	case tea.KeyLeft:
		if m.cursor.Col > 0 {
			m.cursor.Col--
		}
	case tea.KeyRight:
		if m.cursor.Col < 9 {
			m.cursor.Col++
		}

	case tea.KeyEnter:
		switch m.gs.Phase {
		case game.PhasePlacement:
			if m.currentShip != nil {
				coords := buildCoords(m.cursor, m.currentShip.Size, m.orientation)
				if err := m.gs.PlaceShip(m.currentShip.Type, coords, m.orientation); err != nil {
					m.message = fmt.Sprintf("Cannot place ship: %v", err)
				} else {
					m.message = "Ship placed!"
					m.currentShip = nextShip(m.gs.LocalPlayer.Ships, m.currentShip.Type)
				}
			} else {
				m.gs.SetReady()
			}

		case game.PhaseBattle:
			if m.targetCoord == nil {
				m.message = "Aim first: move cursor, press Space to confirm, then Enter to fire."
			} else {
				if err := m.gs.FireShot(*m.targetCoord); err != nil {
					m.message = fmt.Sprintf("Cannot fire: %v", err)
				} else {
					m.targetCoord = nil
				}
			}
		}

	case tea.KeyRunes:
		switch msg.Runes[0] {
		case ' ':
			switch m.gs.Phase {
			case game.PhasePlacement:
				if m.currentShip != nil {
					if err := m.gs.RemoveShip(m.currentShip.Type); err != nil {
						m.message = fmt.Sprintf("Cannot remove: %v", err)
					} else {
						m.message = "Ship removed."
					}
				}
			case game.PhaseBattle:
				target := game.Coordinate{Row: m.cursor.Row, Col: m.cursor.Col}
				m.targetCoord = &target
				m.message = fmt.Sprintf("Targeting %s — press Enter to fire.", target.String())
			}

		case 'r', 'R':
			if m.gs.Phase == game.PhaseGameOver {
				m.gs.RequestReplay()
			}

		case 'q', 'Q':
			if m.gs.Phase == game.PhaseGameOver {
				m.gs.SendMessage(network.CmdQuit)
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the full TUI screen.
func (m Model) View() string {
	if m.gs == nil {
		return "Connecting..."
	}

	if m.gs.Phase == game.PhaseGameOver {
		return renderGameOver(m.gs.Winner, m.gs.LocalPlayer.Name)
	}

	// Title bar
	title := titleStyle.Render("⚓   BATTLESHIP")

	// Phase badge
	var phaseText string
	switch m.gs.Phase {
	case game.PhasePlacement:
		phaseText = "📍  PLACEMENT PHASE"
	case game.PhaseBattle:
		phaseText = "⚔   BATTLE PHASE"
	default:
		phaseText = "⏳  CONNECTING..."
	}
	phase := phaseBadgeStyle.Render(phaseText)

	// Board panels
	showFleetCursor := m.gs.Phase == game.PhasePlacement
	showTrackCursor := m.gs.Phase == game.PhaseBattle

	fleetBoard := boardPanelStyle.Render(
		boardTitleStyle.Render("YOUR FLEET") + "\n" +
			renderBoard(m.gs.LocalPlayer.Board, "fleet", m.cursor, showFleetCursor,
				nil, m.anim, m.currentShip, m.orientation, m.gs.Phase),
	)

	trackBoard := boardPanelStyle.Render(
		boardTitleStyle.Render("ENEMY WATERS") + "\n" +
			renderBoard(m.gs.RemotePlayer.TrackingBoard, "tracking", m.cursor, showTrackCursor,
				m.targetCoord, m.anim, nil, game.Horizontal, m.gs.Phase),
	)

	fleetTracker := boardPanelStyle.Render(
		renderFleetTracker(m.gs.LocalPlayer.Ships),
	)

	boards := lipgloss.JoinHorizontal(lipgloss.Top, fleetBoard, "  ", trackBoard, "  ", fleetTracker)

	// Message + action bar
	msg := messageStyle.Render(m.message)
	action := actionBarStyle.Render(actionBarText(m.gs, m.targetCoord, m.currentShip))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		phase,
		"",
		boards,
		"",
		msg,
		action,
	)
}

// buildCoords computes ship cell coordinates from a cursor, size, and orientation.
func buildCoords(cursor game.Coordinate, size int, orientation game.Orientation) []game.Coordinate {
	coords := make([]game.Coordinate, size)
	for i := 0; i < size; i++ {
		if orientation == game.Horizontal {
			coords[i] = game.Coordinate{Row: cursor.Row, Col: cursor.Col + i}
		} else {
			coords[i] = game.Coordinate{Row: cursor.Row + i, Col: cursor.Col}
		}
	}
	return coords
}

// nextShip returns a pointer to the ship after shipType in the slice, or nil if it was the last.
func nextShip(ships []game.Ship, shipType game.ShipType) *game.Ship {
	for i, ship := range ships {
		if ship.Type == shipType {
			if i+1 < len(ships) {
				return &ships[i+1]
			}
			return nil
		}
	}
	return nil
}
