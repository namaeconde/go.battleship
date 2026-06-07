// game/game_test.go
package game

import (
	"strings"
	"testing"

	"go.battleship/network"
)

// Mock UI for testing GameState without actual tcell screen.
type MockUI struct {
	Events []UIEvent
}

func (m *MockUI) Send(e UIEvent) {
	m.Events = append(m.Events, e)
}

// TestNewGame initializes a new game state.
func TestNewGame(t *testing.T) {
	mockUI := &MockUI{}
	gs := NewGame("Player1", "Player2", mockUI)

	if gs.LocalPlayer == nil || gs.RemotePlayer == nil {
		t.Fatal("Local or Remote player not initialized")
	}
	if gs.LocalPlayer.Name != "Player1" {
		t.Errorf("Expected local player name 'Player1', got '%s'", gs.LocalPlayer.Name)
	}
	if gs.RemotePlayer.Name != "Player2" {
		t.Errorf("Expected remote player name 'Player2', got '%s'", gs.RemotePlayer.Name)
	}
	if gs.Phase != PhaseConnection {
		t.Errorf("Expected initial phase PhaseConnection, got %v", gs.Phase)
	}
	if gs.CancelCtx == nil || gs.CancelFunc == nil {
		t.Error("Context not initialized")
	}
	if gs.incomingMsgs == nil || gs.outgoingMsgs == nil {
		t.Error("Message channels not initialized")
	}
}

// TestTransitionPhase ensures game phase transitions correctly.
func TestTransitionPhase(t *testing.T) {
	mockUI := &MockUI{}
	gs := NewGame("P1", "P2", mockUI)

	gs.TransitionPhase(PhasePlacement)
	if gs.Phase != PhasePlacement {
		t.Errorf("Expected phase PhasePlacement, got %v", gs.Phase)
	}

	gs.TransitionPhase(PhaseBattle)
	if gs.Phase != PhaseBattle {
		t.Errorf("Expected phase PhaseBattle, got %v", gs.Phase)
	}

	gs.TransitionPhase(PhaseGameOver)
	if gs.Phase != PhaseGameOver {
		t.Errorf("Expected phase PhaseGameOver, got %v", gs.Phase)
	}
}

// TestHandleCmdShot tests processing a SHOT command during Battle phase.
func TestHandleCmdShot(t *testing.T) {
	mockUI := &MockUI{}
	gs := NewGame("Host", "Joiner", mockUI)
	gs.Phase = PhaseBattle

	// Send SHOT command
	msg := &network.Message{
		Command: network.CmdShot,
		Args: map[string]string{
			"coord": "A1",
		},
	}

	gs.handleIncomingMessage(msg)

	// Verify response is in outgoing queue
	select {
	case response := <-gs.outgoingMsgs:
		if response.Command != network.CmdShotResult {
			t.Errorf("Expected CmdShotResult, got %s", response.Command)
		}
	default:
		t.Error("Expected response in outgoingMsgs, but channel empty")
	}
}

func TestLegacyConnectMessagesAreIgnored(t *testing.T) {
	tests := []struct {
		name      string
		localName string
		command   network.Command
	}{
		{name: "host ignores connect request", localName: "Host", command: network.Command("CONNECT_REQUEST")},
		{name: "joiner ignores connect ack", localName: "Joiner", command: network.Command("CONNECT_ACK")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUI := &MockUI{}
			gs := NewGame(tt.localName, "Opponent", mockUI)

			gs.handleIncomingMessage(&network.Message{Command: tt.command})

			if gs.Phase != PhaseConnection {
				t.Fatalf("expected phase to remain %v, got %v", PhaseConnection, gs.Phase)
			}

			select {
			case msg := <-gs.outgoingMsgs:
				t.Fatalf("expected no response message, got %v", msg.Command)
			default:
			}

			if len(mockUI.Events) == 0 {
				t.Fatal("expected UI event for ignored legacy command")
			}
			msgEvent, ok := mockUI.Events[len(mockUI.Events)-1].(MessageEvent)
			if !ok {
				t.Fatalf("expected MessageEvent as last event, got %T", mockUI.Events[len(mockUI.Events)-1])
			}
			if !strings.Contains(msgEvent.Text, "Unhandled command") {
				t.Fatalf("expected unhandled command message, got %q", msgEvent.Text)
			}
		})
	}
}
