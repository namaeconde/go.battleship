// game/game.go
package game

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

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
	Connection   net.Conn    // Network connection
	TurnOwner    string      // Name of the player whose turn it is
	UI           UIInterface // Reference to the UI

	// Channels for network communication
	incomingMsgs        chan *network.Message
	outgoingMsgs        chan network.Message
	LastShot            *Coordinate
	Winner              string             // Name of the winning player
	WaitingReplay       bool               // True if this player requested a replay
	OpponentWantsReplay bool               // True if the opponent requested a replay
	CancelCtx           context.Context    // Context for cancellation of goroutines
	CancelFunc          context.CancelFunc // Function to cancel context
	wg                  sync.WaitGroup     // WaitGroup for goroutines
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
	if newPhase == PhaseBattle {
		// Host always fires first
		gs.TurnOwner = "Host"
	}
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
		gs.CancelFunc() // Signal goroutines to stop
	}
	gs.wg.Wait() // Wait for them to finish
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

// PlaceShip places a ship on the local board.
func (gs *GameState) PlaceShip(t ShipType, coords []Coordinate, orientation Orientation) error {
	ship := gs.LocalPlayer.GetShipByType(t)
	if ship == nil {
		return fmt.Errorf("ship %v not found", t)
	}
	ship.Coordinates = coords
	ship.Orientation = orientation
	err := gs.LocalPlayer.Board.PlaceShip(*ship)
	if err != nil {
		ship.Coordinates = nil // Reset
		return err
	}
	return nil
}

// RemoveShip removes a ship from the local board.
func (gs *GameState) RemoveShip(t ShipType) error {
	return gs.LocalPlayer.RemoveShip(t)
}

// FireShot fires a shot at the given coordinate on the opponent's tracking board.
// Returns an error if it is not the local player's turn, the phase is wrong,
// or the coordinate was already targeted.
func (gs *GameState) FireShot(coord Coordinate) error {
	if gs.Phase != PhaseBattle {
		return fmt.Errorf("not in battle phase")
	}
	if gs.TurnOwner != gs.LocalPlayer.Name {
		return fmt.Errorf("not your turn")
	}
	if gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] != Water {
		return fmt.Errorf("coordinate already targeted")
	}
	gs.LastShot = &coord
	gs.TurnOwner = "" // Clear while waiting for response
	gs.SendMessage(network.CmdShot, "coord", coord.String())
	gs.UI.SetMessage(fmt.Sprintf("Firing at %s...", coord.String()))
	gs.UI.Draw(gs, nil, Horizontal)
	return nil
}

// SetReady marks the local player as ready and sends a message to the opponent.
func (gs *GameState) SetReady() {
	gs.LocalPlayer.IsReady = true
	gs.SendMessage(network.CmdPlacementDone)

	if gs.RemotePlayer.IsReady {
		gs.TransitionPhase(PhaseBattle)
	} else {
		gs.UI.SetMessage("Waiting for opponent...")
		gs.UI.Draw(gs, nil, Horizontal)
	}
}

// RequestReplay sends a replay request to the opponent.
// If the opponent has already requested one, the game resets immediately.
func (gs *GameState) RequestReplay() {
	if gs.Phase != PhaseGameOver {
		return
	}
	gs.WaitingReplay = true
	gs.SendMessage(network.CmdReplayRequest)
	if gs.OpponentWantsReplay {
		gs.doReset()
	} else {
		gs.UI.SetMessage("Replay requested. Waiting for opponent to agree...")
		gs.UI.Draw(gs, nil, Horizontal)
	}
}

// doReset resets game state for a new game while keeping the network connection open.
func (gs *GameState) doReset() {
	localName := gs.LocalPlayer.Name
	remoteName := gs.RemotePlayer.Name
	gs.LocalPlayer = NewPlayer(localName)
	gs.RemotePlayer = NewPlayer(remoteName)
	gs.Phase = PhasePlacement
	gs.TurnOwner = ""
	gs.LastShot = nil
	gs.Winner = ""
	gs.WaitingReplay = false
	gs.OpponentWantsReplay = false
	gs.UI.SetMessage("New game! Place your ships.")
	gs.UI.Draw(gs, nil, Horizontal)
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
	case network.CmdPlacementDone:
		gs.RemotePlayer.IsReady = true
		if gs.LocalPlayer.IsReady {
			gs.TransitionPhase(PhaseBattle)
		} else {
			gs.UI.SetMessage("Opponent is ready.")
			gs.UI.Draw(gs, nil, Horizontal)
		}
	case network.CmdShot:
		if gs.Phase != PhaseBattle {
			gs.UI.SetMessage("Shot received outside battle phase.")
			gs.UI.Draw(gs, nil, Horizontal)
			return
		}
		coordStr := msg.Args["coord"]
		coord, err := ParseCoordinate(coordStr)
		if err != nil {
			gs.UI.SetMessage(fmt.Sprintf("Invalid coordinate: %v", coordStr))
			gs.UI.Draw(gs, nil, Horizontal)
			return
		}

		cellState, shipType, err := gs.LocalPlayer.RecordHit(coord)
		if err != nil {
			gs.SendMessage(network.CmdShotResult, "result", "miss")
			return
		}

		result := ""
		switch cellState {
		case Hit:
			result = "hit"
		case Miss:
			result = "miss"
		case SunkShip:
			result = "sunk"
		}

		args := []string{"result", result}
		if result == "sunk" {
			args = append(args, "shipType", shipType.String())
		}
		gs.SendMessage(network.CmdShotResult, args...)
		// Detect if all local ships are sunk (we lost)
		if result == "sunk" && gs.LocalPlayer.ShipsRemaining == 0 {
			gs.Winner = gs.RemotePlayer.Name
			gs.TransitionPhase(PhaseGameOver)
			return
		}
		// Opponent just fired — now it is the local player's turn
		gs.TurnOwner = gs.LocalPlayer.Name
		gs.UI.SetMessage(fmt.Sprintf("Your turn! Opponent fired at %s (%s). Fire back.", coordStr, result))
		gs.UI.Draw(gs, nil, Horizontal)
	case network.CmdShotResult:
		if gs.Phase != PhaseBattle {
			return
		}

		result := msg.Args["result"]
		coord := *gs.LastShot

		// Update TrackingBoard and account for sunk ships
		switch result {
		case "hit":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = Hit
			gs.UI.SetMessage(fmt.Sprintf("Hit at %s! It's now opponent's turn.", coord.String()))
		case "miss":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = Miss
			gs.UI.SetMessage(fmt.Sprintf("Miss at %s. Opponent's turn.", coord.String()))
		case "sunk":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = SunkShip
			gs.RemotePlayer.ShipsRemaining--
			if gs.RemotePlayer.ShipsRemaining <= 0 {
				gs.Winner = gs.LocalPlayer.Name
				gs.TransitionPhase(PhaseGameOver)
				gs.UI.Draw(gs, nil, Horizontal)
				return
			}
			gs.UI.SetMessage(fmt.Sprintf("Sunk %s at %s! Opponent's turn.", msg.Args["shipType"], coord.String()))
		}
		// Local player just received their shot result — now it is opponent's turn
		gs.TurnOwner = gs.RemotePlayer.Name
		gs.UI.Draw(gs, nil, Horizontal)
	case network.CmdReplayRequest:
		if gs.Phase != PhaseGameOver {
			return
		}
		gs.OpponentWantsReplay = true
		if gs.WaitingReplay {
			gs.doReset()
		} else {
			gs.UI.SetMessage("Opponent wants to replay! Press R to agree or Q to quit.")
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
