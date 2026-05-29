package main

import (
	"fmt"
	"os"
	"time"
	"net"

	"go.battleship/cmd"
	"go.battleship/game" // Import the game package
	"go.battleship/ui" // Import ui package
	"go.battleship/network" // Import network package
	"github.com/gdamore/tcell/v2"
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
		conn, err = network.StartHost(gs.CancelCtx, appConfig.Port) // Use game's context
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
				}
				// Keep cursor within bounds
				if gameUI.Cursor.Row < 0 { gameUI.Cursor.Row = 0 }
				if gameUI.Cursor.Row >= 10 { gameUI.Cursor.Row = 10 - 1 }
				if gameUI.Cursor.Col < 0 { gameUI.Cursor.Col = 0 }
				if gameUI.Cursor.Col >= 10 { gameUI.Cursor.Col = 10 - 1 }
				
				gs.UI.Draw(gs, nil, game.Horizontal)
			case *tcell.EventResize:
				s.Sync()
				gs.UI.Draw(gs, nil, game.Horizontal)
			}
		}
	}()

	<-quit
	fmt.Println("Exiting Battleship...")
}
