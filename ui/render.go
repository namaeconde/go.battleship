// ui/render.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"go.battleship/game"
)

// renderBoard renders a 10×10 board grid as a string.
// boardType: "fleet" (player's own ships) or "tracking" (shots at opponent).
func renderBoard(
	board *game.Board,
	boardType string,
	cursor game.Coordinate,
	showCursor bool,
	targetCoord *game.Coordinate,
	anim animState,
	currentShip *game.Ship,
	orientation game.Orientation,
	phase game.GamePhase,
) string {
	if board == nil {
		return ""
	}

	labelSt := lipgloss.NewStyle().Foreground(colorLabel)
	var sb strings.Builder

	// Column headers A–J
	sb.WriteString("    ")
	for c := 0; c < 10; c++ {
		sb.WriteString(labelSt.Render(string(rune('A'+c))) + " ")
	}
	sb.WriteString("\n")

	for r := 0; r < 10; r++ {
		sb.WriteString(labelSt.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < 10; c++ {
			sym, st := cellSymbolAndStyle(board, boardType, r, c, cursor, showCursor, targetCoord, anim, currentShip, orientation, phase)
			sb.WriteString(st.Render(sym) + " ")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func cellSymbolAndStyle(
	board *game.Board,
	boardType string,
	r, c int,
	cursor game.Coordinate,
	showCursor bool,
	targetCoord *game.Coordinate,
	anim animState,
	currentShip *game.Ship,
	orientation game.Orientation,
	phase game.GamePhase,
) (string, lipgloss.Style) {
	base := lipgloss.NewStyle()

	// Animation overlay takes highest priority.
	if anim.active && anim.board == boardType && anim.coord.Row == r && anim.coord.Col == c {
		return animSymbolAndStyle(anim)
	}

	cell := board.Grid[r][c]

	switch {
	case cell == game.Hit || cell == game.SunkShip:
		return "✕", base.Foreground(colorHit).Bold(true)

	case cell == game.Miss:
		return "○", base.Foreground(colorMiss)

	case boardType == "fleet" && cell == game.StateShip:
		return "#", base.Foreground(colorShip)

	default: // Water (or tracking board hiding unshot ship cells)
		// Ghost placement preview during placement phase.
		if boardType == "fleet" && isGhostCell(r, c, cursor, currentShip, orientation, phase) {
			return "#", base.Foreground(colorAccent).Background(lipgloss.Color("#0a2a5a"))
		}
		// Confirmed target highlight.
		if boardType == "tracking" && targetCoord != nil && targetCoord.Row == r && targetCoord.Col == c {
			return "▣", base.Foreground(lipgloss.Color("#ff8800"))
		}
		// Cursor.
		if showCursor && cursor.Row == r && cursor.Col == c {
			return "·", base.Background(lipgloss.Color("#1a3a6a")).Foreground(colorAccent)
		}
		return "~", base.Foreground(colorWater)
	}
}

// animSymbolAndStyle returns the symbol and style for the current animation frame.
func animSymbolAndStyle(anim animState) (string, lipgloss.Style) {
	base := lipgloss.NewStyle()
	f := anim.frame
	if f > 3 {
		f = 3
	}

	hitSymbols := [4]string{"*", "✸", "✕", "✕"}
	hitColors := [4]lipgloss.Color{colorAnimHit0, colorAnimHit1, colorAnimHit2, colorHit}
	missSymbols := [4]string{"·", "·", "○", "○"}
	missColors := [4]lipgloss.Color{colorAnimMiss0, colorAnimMiss1, colorAnimMiss2, colorMiss}

	if anim.result == "hit" || anim.result == "sunk" {
		return hitSymbols[f], base.Foreground(hitColors[f]).Bold(true)
	}
	return missSymbols[f], base.Foreground(missColors[f])
}

// isGhostCell returns true if (r, c) is covered by the ship ghost preview.
func isGhostCell(r, c int, cursor game.Coordinate, ship *game.Ship, orientation game.Orientation, phase game.GamePhase) bool {
	if ship == nil || phase != game.PhasePlacement {
		return false
	}
	for i := 0; i < ship.Size; i++ {
		var coord game.Coordinate
		if orientation == game.Horizontal {
			coord = game.Coordinate{Row: cursor.Row, Col: cursor.Col + i}
		} else {
			coord = game.Coordinate{Row: cursor.Row + i, Col: cursor.Col}
		}
		if coord.Row == r && coord.Col == c {
			return true
		}
	}
	return false
}

// renderFleetTracker returns a styled list of the player's ships showing which are sunk.
func renderFleetTracker(ships []game.Ship) string {
	var sb strings.Builder
	sb.WriteString(boardTitleStyle.Render("FLEET") + "\n")
	for _, ship := range ships {
		name := ship.Type.String()
		switch {
		case isShipSunk(ship):
			sb.WriteString(shipSunkStyle.Render("▣ "+name) + "\n")
		case ship.Coordinates != nil:
			sb.WriteString(shipIntactStyle.Render("▣ "+name) + "\n")
		default:
			sb.WriteString(shipUnplacedStyle.Render("▣ "+name) + "\n")
		}
	}
	return sb.String()
}

// isShipSunk returns true when every section of a placed ship has been hit.
func isShipSunk(ship game.Ship) bool {
	if len(ship.Hits) == 0 {
		return false
	}
	for _, h := range ship.Hits {
		if !h {
			return false
		}
	}
	return true
}

// actionBarText returns context-sensitive help text for the bottom action bar.
func actionBarText(gs *game.GameState, targetCoord *game.Coordinate, currentShip *game.Ship) string {
	switch gs.Phase {
	case game.PhasePlacement:
		if currentShip != nil {
			return fmt.Sprintf("Placing: %s (size %d)  ·  Arrows to move  ·  Enter to place  ·  Space to remove", currentShip.Type, currentShip.Size)
		}
		return "All ships placed!  ·  Press Enter to ready up."
	case game.PhaseBattle:
		if gs.TurnOwner != gs.LocalPlayer.Name {
			return "⏳  Waiting for opponent to fire..."
		}
		if targetCoord != nil {
			return fmt.Sprintf("🎯  Targeting %s  ·  Enter to fire", targetCoord.String())
		}
		return "🎯  YOUR TURN  ·  Arrows to aim  ·  Space to confirm  ·  Enter to fire"
	case game.PhaseGameOver:
		return "R to play again  ·  Q to quit"
	default:
		return "Connecting..."
	}
}

// renderGameOver returns a full-screen game-over view.
func renderGameOver(winnerName, localPlayer string) string {
	var outcomeText string
	if winnerName == localPlayer {
		outcomeText = lipgloss.NewStyle().Foreground(lipgloss.Color("#3fb950")).Bold(true).Render("🏆  YOU WIN!")
	} else {
		outcomeText = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Bold(true).Render("💀  YOU LOSE!")
	}
	winnerLine := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(winnerName + " WINS!")
	return lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("⚓   BATTLESHIP"),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(colorStatus).Render("GAME OVER"),
		"",
		winnerLine,
		outcomeText,
		"",
		lipgloss.NewStyle().Foreground(colorStatus).Render("R to play again  ·  Q to quit"),
	)
}
