package ui

import "go.battleship/game"

type UIInterface interface {
	SetMessage(msg string)
	Draw(gs *game.GameState, currentShip *game.Ship, currentOrientation game.Orientation)
}

type UI struct {
    Cursor game.Coordinate
}
func NewUI(s interface{}) *UI { return &UI{} }
func (u *UI) SetMessage(msg string) {}
func (u *UI) Draw(gs *game.GameState, currentShip *game.Ship, currentOrientation game.Orientation) {}
