// ui/draw.go
package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"go.battleship/game"
)

const (
	BoardOffsetX = 4 // X-offset for the first board (labels take space)
	BoardOffsetY = 2 // Y-offset for both boards
	BoardWidth   = 10
	BoardHeight  = 10
	CellWidth    = 2                              // Width of each cell (e.g., "~ ")
	MessageRow   = BoardOffsetY + BoardHeight + 4 // Row for messages
)

// drawBoards draws both the local player's board and the tracking board.
func drawBoards(s tcell.Screen, localPlayer *game.PlayerState, remotePlayer *game.PlayerState, cursor game.Coordinate, targetCoord *game.Coordinate, phase game.GamePhase, currentShip *game.Ship, currentOrientation game.Orientation) {
	// Draw Local Player's Board (Left)
	if localPlayer != nil && localPlayer.Board != nil {
		drawBoard(s, localPlayer.Board, BoardOffsetX, BoardOffsetY, cursor, nil, true, phase == game.PhasePlacement, phase, currentShip, currentOrientation)
	}

	// Draw Remote Player's Tracking Board (Right) — shows local player's shots against opponent
	if remotePlayer != nil && remotePlayer.TrackingBoard != nil {
		remoteBoardOffsetX := BoardOffsetX + (BoardWidth*CellWidth + 10)
		drawBoard(s, remotePlayer.TrackingBoard, remoteBoardOffsetX, BoardOffsetY, cursor, targetCoord, false, phase == game.PhaseBattle, phase, nil, game.Horizontal)
	}
}

// drawBoard draws a single game board with coordinate labels.
func drawBoard(s tcell.Screen, board *game.Board, offsetX, offsetY int, cursor game.Coordinate, targetCoord *game.Coordinate, isPlayerBoard bool, showCursor bool, phase game.GamePhase, currentShip *game.Ship, currentOrientation game.Orientation) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	labelStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	hitStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
	missStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	sunkStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkRed).Background(tcell.ColorBlack)
	shipStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	cursorStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	potentialPlacementStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue)
	potentialShotStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorOrange)

	// Draw column headers (A-J)
	for c := 0; c < BoardWidth; c++ {
		s.SetContent(offsetX+(c*CellWidth)+1, offsetY-1, rune('A'+c), nil, labelStyle)
	}
	// Draw row headers (1-10)
	for r := 0; r < BoardHeight; r++ {
		rowStr := fmt.Sprintf("%d", r+1)
		for i, char := range rowStr {
			s.SetContent(offsetX-len(rowStr)+i, offsetY+r, char, nil, labelStyle)
		}
	}

	if board == nil {
		return
	}

	for r := 0; r < BoardHeight; r++ {
		for c := 0; c < BoardWidth; c++ {
			char := ' '
			cellStyle := style

			switch board.Grid[r][c] {
			case game.Water:
				char = '~'
			case game.StateShip:
				if isPlayerBoard {
					char = '#'
					cellStyle = shipStyle
				} else {
					char = '~'
				}
			case game.Hit:
				char = 'X'
				cellStyle = hitStyle
			case game.Miss:
				char = 'O'
				cellStyle = missStyle
			case game.SunkShip:
				char = 'S'
				cellStyle = sunkStyle
			}

			// Highlight potential ship placement
			if isPlayerBoard && phase == game.PhasePlacement && showCursor && currentShip != nil {
				isPotentialPlacementCell := false
				for i := 0; i < currentShip.Size; i++ {
					var potentialCoord game.Coordinate
					if currentOrientation == game.Horizontal {
						potentialCoord = game.Coordinate{Row: cursor.Row, Col: cursor.Col + i}
					} else {
						potentialCoord = game.Coordinate{Row: cursor.Row + i, Col: cursor.Col}
					}
					if potentialCoord.Row == r && potentialCoord.Col == c {
						isPotentialPlacementCell = true
						break
					}
				}
				if isPotentialPlacementCell {
					cellStyle = potentialPlacementStyle
					char = '#'
				}
			} else if !isPlayerBoard && phase == game.PhaseBattle && showCursor {
				// Highlight confirmed target in orange (takes priority over cursor yellow)
				if targetCoord != nil && targetCoord.Row == r && targetCoord.Col == c {
					cellStyle = potentialShotStyle
				}
			}

			// Draw cell content
			s.SetContent(offsetX+(c*CellWidth), offsetY+r, char, nil, cellStyle)

			// Draw cursor (overrides cell style)
			if showCursor && cursor.Row == r && cursor.Col == c {
				s.SetContent(offsetX+(c*CellWidth), offsetY+r, char, nil, cursorStyle)
			}
		}
	}
}

// drawMessage draws a message at the bottom of the screen.
func drawMessage(s tcell.Screen, msg string) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	// Clear previous message line
	w, _ := s.Size()
	for i := 0; i < w; i++ {
		s.SetContent(i, MessageRow, ' ', nil, style)
	}
	// Draw new message
	for i, r := range msg {
		s.SetContent(BoardOffsetX+i, MessageRow, r, nil, style)
	}
}

// drawGameOverScreen draws a full-screen game over message with replay/quit options.
func drawGameOverScreen(s tcell.Screen, winner string, localPlayerName string) {
	s.Clear()
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	winnerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)
	winStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack).Bold(true)
	loseStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack).Bold(true)

	width, height := s.Size()
	gameOverText := "GAME OVER"
	winnerText := fmt.Sprintf("%s WINS!", winner)
	var outcomeText string
	var outcomeStyle tcell.Style
	if winner == localPlayerName {
		outcomeText = "YOU WIN!"
		outcomeStyle = winStyle
	} else {
		outcomeText = "YOU LOSE!"
		outcomeStyle = loseStyle
	}
	replayText := "Press R to play again  |  Press Q to quit"

	for i, r := range gameOverText {
		s.SetContent((width-len(gameOverText))/2+i, height/2-3, r, nil, style)
	}
	for i, r := range winnerText {
		s.SetContent((width-len(winnerText))/2+i, height/2-1, r, nil, winnerStyle)
	}
	for i, r := range outcomeText {
		s.SetContent((width-len(outcomeText))/2+i, height/2+1, r, nil, outcomeStyle)
	}
	for i, r := range replayText {
		s.SetContent((width-len(replayText))/2+i, height/2+3, r, nil, style)
	}
	s.Show()
}
