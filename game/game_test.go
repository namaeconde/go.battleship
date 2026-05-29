// game/game_test.go
package game

import (
	"testing"
	"go.battleship/network"
)

// Mock UI for testing GameState without actual tcell screen.
type MockUI struct {
	Messages []string
	DrawCalls int
}

func (m *MockUI) SetMessage(msg string) {
	m.Messages = append(m.Messages, msg)
}
func (m *MockUI) Draw(gs *GameState, currentShip *Ship, currentOrientation Orientation) {
	m.DrawCalls++
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
