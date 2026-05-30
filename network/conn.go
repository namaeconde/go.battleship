package network

import (
	"context"
	"fmt"
	"net"

	battleshipgrpc "go.battleship/proto/battleshipgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func StartHost(ctx context.Context, port string, addrChan chan<- string) (*TCPConn, error) {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	if addrChan != nil {
		addrChan <- ln.Addr().String()
	}

	// Wait for connection
	type res struct {
		conn net.Conn
		err  error
	}
	ch := make(chan res)
	go func() {
		conn, err := ln.Accept()
		ch <- res{conn, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		if r.err != nil {
			return nil, r.err
		}
		return NewTCPConn(r.conn), nil
	}
}

func ConnectToHost(ctx context.Context, addr string) (*TCPConn, error) {
	d := net.Dialer{}
	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewTCPConn(c), nil
}

// CreateGame dials the gRPC relay server, creates a new game session, and returns
// the open GameStream connection plus the game_id to share with Player B.
func CreateGame(ctx context.Context, serverAddr string) (*GRPCConn, string, error) {
	grpcConn, err := dialServer(serverAddr)
	if err != nil {
		return nil, "", fmt.Errorf("CreateGame dial: %w", err)
	}

	client := battleshipgrpc.NewBattleshipRelayClient(grpcConn)

	resp, err := client.CreateGame(ctx, &battleshipgrpc.CreateGameRequest{})
	if err != nil {
		grpcConn.Close()
		return nil, "", fmt.Errorf("CreateGame RPC: %w", err)
	}
	gameID := resp.GameId

	stream, err := client.GameStream(ctx)
	if err != nil {
		grpcConn.Close()
		return nil, "", fmt.Errorf("CreateGame GameStream: %w", err)
	}

	return NewGRPCConn(stream, gameID), gameID, nil
}

// JoinGame dials the gRPC relay server, joins an existing game session by game_id,
// and returns the open GameStream connection.
func JoinGame(ctx context.Context, serverAddr string, gameID string) (*GRPCConn, error) {
	grpcConn, err := dialServer(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("JoinGame dial: %w", err)
	}

	client := battleshipgrpc.NewBattleshipRelayClient(grpcConn)

	joinResp, err := client.JoinGame(ctx, &battleshipgrpc.JoinGameRequest{GameId: gameID})
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("JoinGame RPC: %w", err)
	}
	if !joinResp.Success {
		grpcConn.Close()
		return nil, fmt.Errorf("JoinGame rejected: %s", joinResp.ErrorMessage)
	}

	stream, err := client.GameStream(ctx)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("JoinGame GameStream: %w", err)
	}

	return NewGRPCConn(stream, gameID), nil
}

// dialServer opens a gRPC client connection to the relay server.
// Uses TLS for https:// addresses, plaintext for everything else.
func dialServer(addr string) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if len(addr) >= 5 && addr[:5] == "https" {
		creds = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	return grpc.NewClient(addr, creds)
}
