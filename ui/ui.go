// ui/ui.go
package ui

import (
	"github.com/gdamore/tcell/v2"
	"go.battleship/game"
)

// UIInterface defines the interface for our UI.
type UIInterface interface {
	SetMessage(msg string)
	Draw(gs *game.GameState, currentShip *game.Ship, currentOrientation game.Orientation)
}

// UI implements the UIInterface for tcell.
type UI struct {
	screen      tcell.Screen
	Cursor      game.Coordinate
	TargetCoord *game.Coordinate // Confirmed shot target during battle phase
	Message     string
}

// NewUI creates a new UI instance.
func NewUI(s tcell.Screen) *UI {
	return &UI{
		screen:  s,
		Cursor:  game.Coordinate{Row: 0, Col: 0},
		Message: "Welcome to Battleship!",
	}
}

// Draw updates the entire screen based on the current game state.
func (u *UI) Draw(gs *game.GameState, currentShip *game.Ship, currentOrientation game.Orientation) {
	if gs.Phase == game.PhaseGameOver {
		drawGameOverScreen(u.screen, gs.Winner, gs.LocalPlayer.Name)
		return
	}

	u.screen.Clear()
	drawBoards(u.screen, gs.LocalPlayer, gs.RemotePlayer, u.Cursor, u.TargetCoord, gs.Phase, currentShip, currentOrientation)
	drawMessage(u.screen, u.Message)
	u.screen.Show()
}

func (u *UI) SetMessage(msg string) {
	u.Message = msg
}
