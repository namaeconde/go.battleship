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
func drawBoards(s tcell.Screen, localPlayer *game.PlayerState, remotePlayer *game.PlayerState, cursor game.Coordinate, phase game.GamePhase, currentShip *game.Ship, currentOrientation game.Orientation) {
	// Draw Local Player's Board (Left)
	if localPlayer != nil && localPlayer.Board != nil {
		drawBoard(s, localPlayer.Board, BoardOffsetX, BoardOffsetY, cursor, true, phase == game.PhasePlacement, phase, currentShip, currentOrientation)
	}

	// Draw Remote Player's Board (Right) - This acts as the tracking board for the local player
	if remotePlayer != nil && remotePlayer.Board != nil {
		remoteBoardOffsetX := BoardOffsetX + (BoardWidth*CellWidth + 10)
		drawBoard(s, remotePlayer.Board, remoteBoardOffsetX, BoardOffsetY, cursor, false, phase == game.PhaseBattle, phase, nil, game.Horizontal)
	}
}

// drawBoard draws a single game board with coordinate labels.
func drawBoard(s tcell.Screen, board *game.Board, offsetX, offsetY int, cursor game.Coordinate, isPlayerBoard bool, showCursor bool, phase game.GamePhase, currentShip *game.Ship, currentOrientation game.Orientation) {
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
			} else if !isPlayerBoard && phase == game.PhaseBattle && showCursor && cursor.Row == r && cursor.Col == c {
				// Highlight potential shot target
				cellStyle = potentialShotStyle
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

// drawGameOverScreen draws a full-screen game over message.
func drawGameOverScreen(s tcell.Screen, winner string) {
	s.Clear()
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	winnerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)

	width, height := s.Size()
	gameOverText := "GAME OVER"
	winnerText := fmt.Sprintf("%s WINS!", winner)
	exitText := "Press ESC or Ctrl+C to exit."

	// Draw Game Over
	for i, r := range gameOverText {
		s.SetContent((width-len(gameOverText))/2+i, height/2-2, r, nil, style)
	}

	// Draw Winner
	for i, r := range winnerText {
		s.SetContent((width-len(winnerText))/2+i, height/2, r, nil, winnerStyle)
	}

	// Draw Exit instructions
	for i, r := range exitText {
		s.SetContent((width-len(exitText))/2+i, height/2+2, r, nil, style)
	}
	s.Show()
}
