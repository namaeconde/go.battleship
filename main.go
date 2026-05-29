package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"go.battleship/cmd"
	"go.battleship/game"    // Import the game package
	"go.battleship/network" // Import network package
	"go.battleship/ui"      // Import ui package
)

func main() {
	appConfig, err := cmd.ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer s.Fini() // Ensure screen is finalized on exit

	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	s.Clear()

	gameUI := ui.NewUI(s)

	var localPlayerName, remotePlayerName string
	if appConfig.IsHost {
		localPlayerName = "Host"
		remotePlayerName = "Joiner"
	} else {
		localPlayerName = "Joiner"
		remotePlayerName = "Host"
	}
	gs := game.NewGame(localPlayerName, remotePlayerName, gameUI)
	defer gs.Close()

	// --- Network Connection Setup ---
	var conn net.Conn
	if appConfig.IsHost {
		gameUI.SetMessage(fmt.Sprintf("Hosting on port %s. Waiting for opponent...", appConfig.Port))
		gameUI.Draw(gs, nil, game.Horizontal)
		conn, err = network.StartHost(gs.CancelCtx, appConfig.Port, nil) // Use game's context
	} else {
		gameUI.SetMessage(fmt.Sprintf("Connecting to %s...", appConfig.RemoteAddress))
		gameUI.Draw(gs, nil, game.Horizontal)
		conn, err = network.ConnectToHost(gs.CancelCtx, appConfig.RemoteAddress) // Use game's context
	}

	if err != nil {
		gameUI.SetMessage(fmt.Sprintf("Connection error: %v", err))
		gameUI.Draw(gs, nil, game.Horizontal)
		time.Sleep(3 * time.Second) // Give user time to read error
		os.Exit(1)
	}
	gs.Connection = conn // Assign the established connection to GameState

	gameUI.SetMessage("Connection established! Starting game...")
	gameUI.Draw(gs, nil, game.Horizontal)
	time.Sleep(2 * time.Second) // Show connection message briefly
	// --- End Network Connection Setup ---

	// Send initial handshake messages
	if appConfig.IsHost {
		gs.SendMessage(network.CmdConnectAck) // Host sends ACK after connection
	} else {
		gs.SendMessage(network.CmdConnectRequest) // Joiner sends Request
	}

	// Start the game loop (which will now draw the UI)
	go gs.StartGameLoop()

	// Placement state
	currentShip := gs.LocalPlayer.GetShipByType(game.Carrier)
	currentOrientation := game.Horizontal
	lastPhase := gs.Phase // track phase changes to detect game reset

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					close(quit)
					return
				}
				// Cursor movement logic (for testing UI interaction)
				switch ev.Key() {
				case tcell.KeyUp:
					gameUI.Cursor.Row--
				case tcell.KeyDown:
					gameUI.Cursor.Row++
				case tcell.KeyLeft:
					gameUI.Cursor.Col--
				case tcell.KeyRight:
					gameUI.Cursor.Col++
				case tcell.KeyEnter:
					switch gs.Phase {
					case game.PhasePlacement:
						if currentShip != nil {
							coords := []game.Coordinate{}
							for i := 0; i < currentShip.Size; i++ {
								if currentOrientation == game.Horizontal {
									coords = append(coords, game.Coordinate{Row: gameUI.Cursor.Row, Col: gameUI.Cursor.Col + i})
								} else {
									coords = append(coords, game.Coordinate{Row: gameUI.Cursor.Row + i, Col: gameUI.Cursor.Col})
								}
							}
							err := gs.PlaceShip(currentShip.Type, coords, currentOrientation)
							if err != nil {
								gameUI.SetMessage(fmt.Sprintf("Error: %v", err))
							} else {
								gameUI.SetMessage("Ship placed!")
								// Move to next ship
								found := false
								for i, ship := range gs.LocalPlayer.Ships {
									if ship.Type == currentShip.Type {
										if i+1 < len(gs.LocalPlayer.Ships) {
											currentShip = &gs.LocalPlayer.Ships[i+1]
										} else {
											currentShip = nil
										}
										found = true
										break
									}
								}
								if !found {
									currentShip = nil
								}
							}
						} else {
							gs.SetReady()
						}
					case game.PhaseBattle:
						// Fire at the confirmed target (set with Space)
						if gameUI.TargetCoord == nil {
							gameUI.SetMessage("Aim first: move cursor with arrows, press Space to confirm target, then Enter to fire.")
						} else {
							err := gs.FireShot(*gameUI.TargetCoord)
							if err != nil {
								gameUI.SetMessage(fmt.Sprintf("Cannot fire: %v", err))
							} else {
								gameUI.TargetCoord = nil // Clear confirmed target after firing
							}
						}
					}
				case tcell.KeyRune:
					switch {
					case ev.Rune() == ' ' && gs.Phase == game.PhasePlacement:
						// Remove current ship from board during placement
						if currentShip != nil {
							err := gs.RemoveShip(currentShip.Type)
							if err != nil {
								gameUI.SetMessage(fmt.Sprintf("Error: %v", err))
							} else {
								gameUI.SetMessage("Ship removed!")
							}
						}
					case ev.Rune() == ' ' && gs.Phase == game.PhaseBattle:
						// Confirm the targeting coordinate at the current cursor position
						target := game.Coordinate{Row: gameUI.Cursor.Row, Col: gameUI.Cursor.Col}
						gameUI.TargetCoord = &target
						gameUI.SetMessage(fmt.Sprintf("Targeting %s — press Enter to fire.", target.String()))
					case (ev.Rune() == 'r' || ev.Rune() == 'R') && gs.Phase == game.PhaseGameOver:
						gs.RequestReplay()
					case (ev.Rune() == 'q' || ev.Rune() == 'Q') && gs.Phase == game.PhaseGameOver:
						gs.SendMessage(network.CmdQuit)
						close(quit)
						return
					}
				}
				// Keep cursor within bounds
				if gameUI.Cursor.Row < 0 {
					gameUI.Cursor.Row = 0
				}
				if gameUI.Cursor.Row >= 10 {
					gameUI.Cursor.Row = 10 - 1
				}
				if gameUI.Cursor.Col < 0 {
					gameUI.Cursor.Col = 0
				}
				if gameUI.Cursor.Col >= 10 {
					gameUI.Cursor.Col = 10 - 1
				}

				// Reset placement state when a replay starts after game over
				if lastPhase == game.PhaseGameOver && gs.Phase == game.PhasePlacement {
					currentShip = gs.LocalPlayer.GetShipByType(game.Carrier)
					currentOrientation = game.Horizontal
					gameUI.TargetCoord = nil
				}
				lastPhase = gs.Phase
				gameUI.Draw(gs, currentShip, currentOrientation)
			case *tcell.EventResize:
				s.Sync()
				gameUI.Draw(gs, currentShip, currentOrientation)
			}
		}
	}()

	<-quit
	fmt.Println("Exiting Battleship...")
}
