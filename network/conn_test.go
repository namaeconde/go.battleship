package network

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNetworkConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	port := ":9000"
	
	// Try to start host
	connChan := make(chan net.Conn)
	errChan := make(chan error)

	go func() {
		conn, err := StartHost(ctx, port)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	// Try to connect to host
	conn, err := ConnectToHost(ctx, "localhost"+port)
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
