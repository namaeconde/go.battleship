package cmd

import (
	"io"
	"strings"
	"testing"

	"go.battleship/game"
)

func executeRoot(t *testing.T, args ...string) error {
	t.Helper()
	serverURL = ""
	hostCmd.Flags().Lookup("server").Changed = false
	joinCmd.Flags().Lookup("server").Changed = false
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs(args)
	_, err := rootCmd.ExecuteC()
	return err
}

func TestHostCommandRequiresServerFlag(t *testing.T) {
	err := executeRoot(t, "host")
	if err == nil {
		t.Fatal("expected missing server flag error")
	}
	if !strings.Contains(err.Error(), "server URL is required") {
		t.Fatalf("expected server URL required error, got %v", err)
	}
}

func TestJoinCommandRequiresGameID(t *testing.T) {
	err := executeRoot(t, "join", "--server", "http://localhost:8080")
	if err == nil {
		t.Fatal("expected missing game id error")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 0") {
		t.Fatalf("expected missing game id error, got %v", err)
	}
}

func TestJoinCommandRequiresServerFlag(t *testing.T) {
	err := executeRoot(t, "join", "ABC123")
	if err == nil {
		t.Fatal("expected missing server flag error")
	}
	if !strings.Contains(err.Error(), "server URL is required") {
		t.Fatalf("expected server URL required error, got %v", err)
	}
}

type mockGameUI struct {
	events []game.UIEvent
}

func (m *mockGameUI) Send(event game.UIEvent) {
	m.events = append(m.events, event)
}

func TestTransitionToPlacementSendsPhaseEvent(t *testing.T) {
	ui := &mockGameUI{}
	gs := game.NewGame("Host", "Joiner", ui)

	gs.TransitionPhase(game.PhasePlacement)

	if gs.Phase != game.PhasePlacement {
		t.Fatalf("expected phase %v, got %v", game.PhasePlacement, gs.Phase)
	}
	if len(ui.events) == 0 {
		t.Fatal("expected at least one UIEvent")
	}
	found := false
	for _, ev := range ui.events {
		if pce, ok := ev.(game.PhaseChangedEvent); ok && pce.Phase == game.PhasePlacement {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected PhaseChangedEvent{PhasePlacement}, got %v", ui.events)
	}
}
