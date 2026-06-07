// game/game.go
package game

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.battleship/network"
)

// UIInterface defines the interface for the game UI.
// Defined here to avoid circular dependency with the ui package.
type UIInterface interface {
	Send(event UIEvent)
}

// NetworkConn is the minimal interface both the legacy TCP connection and the
// new gRPC stream wrapper must satisfy.
type NetworkConn interface {
	Send(msg network.Message) error
	Receive() (*network.Message, error)
	Close() error
}

// GameState holds the entire state of the game.
type GameState struct {
	LocalPlayer  *PlayerState
	RemotePlayer *PlayerState
	Phase        GamePhase
	Connection   NetworkConn // Network connection
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
	mu                  sync.Mutex         // Protects ready/phase transition logic
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
		gs.TurnOwner = "Host"
	}
	gs.UI.Send(PhaseChangedEvent{Phase: newPhase})
}

// StartGameLoop is the main game loop, integrating UI and network events.
func (gs *GameState) StartGameLoop() {
	gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Game loop started. Current phase: %v", gs.Phase)})

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
			msg, err := gs.Connection.Receive()
			if err != nil {
				if err.Error() == "EOF" || strings.Contains(err.Error(), "use of closed network connection") {
					gs.UI.Send(MessageEvent{Text: "Opponent disconnected."})
				} else {
					gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Network read error: %v", err)})
				}
				gs.UI.Send(QuitEvent{})
				gs.CancelFunc()
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
			err := gs.Connection.Send(msg)
			if err != nil {
				gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Network write error: %v", err)})
				gs.UI.Send(QuitEvent{})
				gs.CancelFunc()
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
	gs.TurnOwner = ""
	gs.SendMessage(network.CmdShot, "coord", coord.String())
	gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Firing at %s...", coord.String())})
	return nil
}

// SetReady marks the local player as ready and sends a message to the opponent.
func (gs *GameState) SetReady() {
	gs.mu.Lock()
	gs.LocalPlayer.IsReady = true
	bothReady := gs.RemotePlayer.IsReady
	gs.mu.Unlock()

	gs.SendMessage(network.CmdPlacementDone)

	if bothReady {
		gs.TransitionPhase(PhaseBattle)
		gs.UI.Send(MessageEvent{Text: "Both players ready! Starting battle phase."})
	} else {
		gs.UI.Send(MessageEvent{Text: "Waiting for opponent..."})
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
		gs.UI.Send(MessageEvent{Text: "Replay requested. Waiting for opponent to agree..."})
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
	gs.UI.Send(PhaseChangedEvent{Phase: PhasePlacement})
	gs.UI.Send(MessageEvent{Text: "New game! Place your ships."})
}

// handleIncomingMessage processes a received network message.
func (gs *GameState) handleIncomingMessage(msg *network.Message) {
	switch msg.Command {
	case network.CmdPlacementDone:
		gs.mu.Lock()
		gs.RemotePlayer.IsReady = true
		bothReady := gs.LocalPlayer.IsReady
		gs.mu.Unlock()

		if bothReady {
			gs.TransitionPhase(PhaseBattle)
			gs.UI.Send(MessageEvent{Text: "Both players ready! Starting battle phase."})
		} else {
			gs.UI.Send(MessageEvent{Text: "Opponent is ready."})
		}

	case network.CmdShot:
		if gs.Phase != PhaseBattle {
			gs.UI.Send(MessageEvent{Text: "Shot received outside battle phase."})
			return
		} else {
			gs.UI.Send(MessageEvent{Text: "Fire in the hole!"})
		}

		coordStr := msg.Args["coord"]
		coord, err := ParseCoordinate(coordStr)
		if err != nil {
			gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Invalid coordinate: %v", coordStr)})
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

		gs.UI.Send(ShotResultEvent{Coord: coord, Result: result, ShipType: shipType.String(), Board: "fleet"})

		if result == "sunk" && gs.LocalPlayer.ShipsRemaining == 0 {
			gs.Winner = gs.RemotePlayer.Name
			gs.TransitionPhase(PhaseGameOver)
			gs.UI.Send(GameOverEvent{Winner: gs.Winner})
			return
		}
		gs.TurnOwner = gs.LocalPlayer.Name
		gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Your turn! Opponent fired at %s (%s). Fire back.", coordStr, result)})

	case network.CmdShotResult:
		if gs.Phase != PhaseBattle {
			return
		}
		result := msg.Args["result"]
		coord := *gs.LastShot

		switch result {
		case "hit":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = Hit
		case "miss":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = Miss
		case "sunk":
			gs.RemotePlayer.TrackingBoard.Grid[coord.Row][coord.Col] = SunkShip
			gs.RemotePlayer.ShipsRemaining--
		}

		gs.UI.Send(ShotResultEvent{Coord: coord, Result: result, ShipType: msg.Args["shipType"], Board: "tracking"})

		if result == "sunk" && gs.RemotePlayer.ShipsRemaining <= 0 {
			gs.Winner = gs.LocalPlayer.Name
			gs.TransitionPhase(PhaseGameOver)
			gs.UI.Send(GameOverEvent{Winner: gs.Winner})
			return
		}
		gs.TurnOwner = gs.RemotePlayer.Name

	case network.CmdReplayRequest:
		if gs.Phase != PhaseGameOver {
			return
		}
		gs.OpponentWantsReplay = true
		if gs.WaitingReplay {
			gs.doReset()
		} else {
			gs.UI.Send(MessageEvent{Text: "Opponent wants to replay! Press R to agree or Q to quit."})
		}

	case network.CmdQuit:
		gs.UI.Send(MessageEvent{Text: "Opponent quit the game."})
		gs.UI.Send(QuitEvent{})

	default:
		gs.UI.Send(MessageEvent{Text: fmt.Sprintf("Unhandled command: %s", msg.Command)})
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
