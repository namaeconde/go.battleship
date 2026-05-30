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
	if !strings.Contains(err.Error(), "required flag(s) \"server\" not set") {
		t.Fatalf("expected required server flag error, got %v", err)
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
	if !strings.Contains(err.Error(), "required flag(s) \"server\" not set") {
		t.Fatalf("expected required server flag error, got %v", err)
	}
}

type mockGameUI struct {
	messages  []string
	drawCalls int
}

func (m *mockGameUI) SetMessage(msg string) {
	m.messages = append(m.messages, msg)
}

func (m *mockGameUI) Draw(gs *game.GameState, currentShip *game.Ship, currentOrientation game.Orientation) {
	m.drawCalls++
}

func TestInitializeGameSessionTransitionsToPlacement(t *testing.T) {
	ui := &mockGameUI{}
	gs := game.NewGame("Host", "Joiner", ui)

	initializeGameSession(gs)

	if gs.Phase != game.PhasePlacement {
		t.Fatalf("expected phase %v, got %v", game.PhasePlacement, gs.Phase)
	}
	if len(ui.messages) == 0 {
		t.Fatal("expected startup message to be shown")
	}
	if got := ui.messages[len(ui.messages)-1]; got != "Connection established! Place your ships." {
		t.Fatalf("expected placement message, got %q", got)
	}
	if ui.drawCalls == 0 {
		t.Fatal("expected UI redraw during startup")
	}
}
