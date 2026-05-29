package network

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNetworkConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	port := ":0"
	
	// Try to start host
	connChan := make(chan net.Conn)
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
