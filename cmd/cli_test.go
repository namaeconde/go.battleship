package cmd

import (
	"io"
	"strings"
	"testing"
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
