package network

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNetworkConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	port := "0"
	
	// Try to start host
	connChan := make(chan *TCPConn)
	errChan := make(chan error)
	addrChan := make(chan string)

	go func() {
		conn, err := StartHost(ctx, port, addrChan)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	// Wait for the host to be ready and get the address
	var hostAddr string
	select {
	case hostAddr = <-addrChan:
		// OK
	case err := <-errChan:
		t.Fatalf("Failed to start host: %v", err)
	case <-ctx.Done():
		t.Fatal("Timeout waiting for host address")
	}

	// Try to connect to host
	conn, err := ConnectToHost(ctx, hostAddr)
	if err != nil {
		t.Fatalf("Failed to connect to host: %v", err)
	}
	defer conn.Close()

	// Wait for host connection
	select {
	case hostConn := <-connChan:
		defer hostConn.Close()
	case err := <-errChan:
		t.Fatalf("Failed to start host: %v", err)
	case <-ctx.Done():
		t.Fatal("Timeout waiting for connection")
	}
}

func TestCreateAndJoinGame_Integration(t *testing.T) {
	// This test requires a running gRPC server. Skip if SERVER_URL not set.
	serverURL := os.Getenv("BATTLESHIP_SERVER_URL")
	if serverURL == "" {
		t.Skip("BATTLESHIP_SERVER_URL not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	hostConn, gameID, err := CreateGame(ctx, serverURL)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	defer hostConn.Close()

	if len(gameID) != 6 {
		t.Fatalf("expected 6-char game_id, got %q", gameID)
	}

	joinerConn, err := JoinGame(ctx, serverURL, gameID)
	if err != nil {
		t.Fatalf("JoinGame: %v", err)
	}
	defer joinerConn.Close()
}
