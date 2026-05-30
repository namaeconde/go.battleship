package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/cobra"
	"go.battleship/game"
	"go.battleship/network"
	"go.battleship/ui"
)

var (
	serverURL string
)

var rootCmd = &cobra.Command{
	Use:   "battleship",
	Short: "Terminal Battleship",
}

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Create a new game and wait for opponent",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, gameID, err := network.CreateGame(context.Background(), serverURL)
		if err != nil {
			return fmt.Errorf("creating game: %w", err)
		}
		fmt.Printf("Game created! Share this code with your opponent: %s\n", gameID)
		fmt.Println("Waiting for opponent to join...")
		RunGame(conn, true)
		return nil
	},
}

var joinCmd = &cobra.Command{
	Use:   "join GAME_ID",
	Short: "Join an existing game",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameID := args[0]
		conn, err := network.JoinGame(context.Background(), serverURL, gameID)
		if err != nil {
			return fmt.Errorf("joining game: %w", err)
		}
		fmt.Printf("Joined game %s!\n", gameID)
		RunGame(conn, false)
		return nil
	},
}

func init() {
	hostCmd.Flags().StringVar(&serverURL, "server", "", "gRPC relay server URL (required)")
	hostCmd.MarkFlagRequired("server")
	joinCmd.Flags().StringVar(&serverURL, "server", "", "gRPC relay server URL (required)")
	joinCmd.MarkFlagRequired("server")
	rootCmd.AddCommand(hostCmd, joinCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initializeGameSession(gs *game.GameState) {
	gs.TransitionPhase(game.PhasePlacement)
	gs.UI.SetMessage("Connection established! Place your ships.")
	gs.UI.Draw(gs, nil, game.Horizontal)
}

func RunGame(conn game.NetworkConn, host bool) {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer s.Fini()

	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	s.Clear()

	gameUI := ui.NewUI(s)

	localPlayerName, remotePlayerName := "Joiner", "Host"
	if host {
		localPlayerName, remotePlayerName = "Host", "Joiner"
	}

	gs := game.NewGame(localPlayerName, remotePlayerName, gameUI)
	defer gs.Close()
	gs.Connection = conn

	initializeGameSession(gs)
	go gs.StartGameLoop()

	currentShip := gs.LocalPlayer.GetShipByType(game.Carrier)
	currentOrientation := game.Horizontal
	lastPhase := gs.Phase

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventInterrupt:
				return
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					close(quit)
					return
				}
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
						if gameUI.TargetCoord == nil {
							gameUI.SetMessage("Aim first: move cursor with arrows, press Space to confirm target, then Enter to fire.")
						} else {
							err := gs.FireShot(*gameUI.TargetCoord)
							if err != nil {
								gameUI.SetMessage(fmt.Sprintf("Cannot fire: %v", err))
							} else {
								gameUI.TargetCoord = nil
							}
						}
					}
				case tcell.KeyRune:
					switch {
					case ev.Rune() == ' ' && gs.Phase == game.PhasePlacement:
						if currentShip != nil {
							err := gs.RemoveShip(currentShip.Type)
							if err != nil {
								gameUI.SetMessage(fmt.Sprintf("Error: %v", err))
							} else {
								gameUI.SetMessage("Ship removed!")
							}
						}
					case ev.Rune() == ' ' && gs.Phase == game.PhaseBattle:
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

	select {
	case <-quit:
	case <-gs.CancelCtx.Done():
	}
	// Wake the input goroutine if blocked on PollEvent so it can exit cleanly.
	s.PostEvent(tcell.NewEventInterrupt(struct{}{}))
	fmt.Println("Exiting Battleship...")
}
