// game/game.go
package game

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
	"strings"

	"go.battleship/network"
)

// UIInterface defines the interface for the game UI.
// Defined here to avoid circular dependency with the ui package.
type UIInterface interface {
	SetMessage(msg string)
	Draw(gs *GameState, currentShip *Ship, currentOrientation Orientation)
}

// GameState holds the entire state of the game.
type GameState struct {
	LocalPlayer  *PlayerState
	RemotePlayer *PlayerState
	Phase        GamePhase
	Connection   net.Conn // Network connection
	TurnOwner    string   // Name of the player whose turn it is
	UI           UIInterface   // Reference to the UI

	// Channels for network communication
	incomingMsgs chan *network.Message
	outgoingMsgs chan network.Message
	CancelCtx    context.Context    // Context for cancellation of goroutines
	CancelFunc   context.CancelFunc // Function to cancel context
	wg           sync.WaitGroup     // WaitGroup for goroutines
}

// NewGame initializes a new GameState with two players.
func NewGame(localPlayerName, remotePlayerName string, gameUI UIInterface) *GameState {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameState{
		LocalPlayer:  NewPlayer(localPlayerName),
		RemotePlayer: NewPlayer(remotePlayerName),
		Phase:        PhaseConnection, // Start in connection phase
		UI:           gameUI,
		incomingMsgs: make(chan *network.Message, 100), // Buffered channel
		outgoingMsgs: make(chan network.Message, 100),
		CancelCtx:    ctx,
		CancelFunc:   cancel,
	}
}

// TransitionPhase changes the current game phase.
func (gs *GameState) TransitionPhase(newPhase GamePhase) {
	gs.Phase = newPhase
	gs.UI.SetMessage(fmt.Sprintf("Game phase transitioned to: %v", newPhase))
	gs.UI.Draw(gs, nil, Horizontal) // Redraw with updated phase
}

// StartGameLoop is the main game loop, integrating UI and network events.
func (gs *GameState) StartGameLoop() {
	gs.UI.SetMessage(fmt.Sprintf("Game loop started. Current phase: %v", gs.Phase))
	gs.UI.Draw(gs, nil, Horizontal)

	// Start network listener/sender goroutines
	gs.wg.Add(2)
	go gs.networkReader()
	go gs.networkWriter()

	// Main loop for game logic and UI updates
	for {
		select {
		case <-gs.CancelCtx.Done():
			fmt.Println("Game loop cancelled.")
			return
		case msg := <-gs.incomingMsgs:
			gs.handleIncomingMessage(msg)
		default:
			// Prevents busy-waiting on the select statement
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Close gracefully shuts down the game state.
func (gs *GameState) Close() {
	if gs.CancelFunc != nil {
		gs.CancelFunc()      // Signal goroutines to stop
	}
	gs.wg.Wait()        // Wait for them to finish
	if gs.Connection != nil {
		gs.Connection.Close() // Close network connection
	}
	// Note: We don't close channels here because they might still be read/written by goroutines
	// that haven't finished yet, even after wg.Wait().
	// Actually wg.Wait() ensures they are finished, but closing channels is generally safe after wg.Wait().
	// But let's be careful.
	fmt.Println("GameState closed.")
}

// networkReader reads messages from the network and sends them to incomingMsgs channel.
func (gs *GameState) networkReader() {
	defer gs.wg.Done()
	if gs.Connection == nil {
		fmt.Println("networkReader: No connection to read from.")
		return
	}
	for {
		select {
		case <-gs.CancelCtx.Done():
			fmt.Println("networkReader: Shutting down.")
			return
		default:
			msg, err := network.Receive(gs.Connection)
			if err != nil {
				if err.Error() == "EOF" || strings.Contains(err.Error(), "use of closed network connection") {
					gs.UI.SetMessage("Opponent disconnected.")
				} else {
					gs.UI.SetMessage(fmt.Sprintf("Network read error: %v", err))
				}
				gs.UI.Draw(gs, nil, Horizontal)
				gs.CancelFunc() // Signal game loop to stop on network error
				return
			}
			gs.incomingMsgs <- msg
		}
	}
}

// networkWriter sends messages from outgoingMsgs channel over the network.
func (gs *GameState) networkWriter() {
	defer gs.wg.Done()
	if gs.Connection == nil {
		fmt.Println("networkWriter: No connection to write to.")
		return
	}
	for {
		select {
		case <-gs.CancelCtx.Done():
			fmt.Println("networkWriter: Shutting down.")
			return
		case msg := <-gs.outgoingMsgs:
			err := network.Send(gs.Connection, msg)
			if err != nil {
				gs.UI.SetMessage(fmt.Sprintf("Network write error: %v", err))
				gs.UI.Draw(gs, nil, Horizontal)
				gs.CancelFunc() // Signal game loop to stop on network error
				return
			}
		}
	}
}

// handleIncomingMessage processes a received network message.
func (gs *GameState) handleIncomingMessage(msg *network.Message) {
	gs.UI.SetMessage(fmt.Sprintf("Received: %v %v", msg.Command, msg.Args))
	gs.UI.Draw(gs, nil, Horizontal) // Redraw to show message immediately

	switch msg.Command {
	case network.CmdConnectRequest:
		if gs.LocalPlayer.Name == "Host" {
			gs.SendMessage(network.CmdConnectAck)
			gs.TransitionPhase(PhasePlacement)
			gs.UI.SetMessage("Opponent connected. Ready for ship placement. Press Enter when done.")
			gs.UI.Draw(gs, nil, Horizontal)
		}
	case network.CmdConnectAck:
		if gs.LocalPlayer.Name == "Joiner" {
			gs.TransitionPhase(PhasePlacement)
			gs.UI.SetMessage("Connected to host. Ready for ship placement. Press Enter when done.")
			gs.UI.Draw(gs, nil, Horizontal)
		}
	case network.CmdQuit:
		gs.UI.SetMessage("Opponent quit the game.")
		gs.UI.Draw(gs, nil, Horizontal)
		gs.CancelFunc()
	default:
		gs.UI.SetMessage(fmt.Sprintf("Unhandled command: %s", msg.Command))
		gs.UI.Draw(gs, nil, Horizontal)
	}
}

// SendMessage adds a message to the outgoing queue.
func (gs *GameState) SendMessage(cmd network.Command, args ...string) {
	argMap := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			argMap[args[i]] = args[i+1]
		}
	}
	msg := network.Message{Command: cmd, Args: argMap}
	select {
	case gs.outgoingMsgs <- msg:
		// Message sent to queue
	case <-gs.CancelCtx.Done():
		fmt.Println("SendMessage: Game is shutting down, cannot send message.")
	default:
		fmt.Println("SendMessage: Outgoing message queue full, dropping message.")
	}
}

