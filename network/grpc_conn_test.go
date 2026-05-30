// network/grpc_conn_test.go
package network

import (
	"context"
	"net"
	"testing"

	battleshipgrpc "go.battleship/proto/battleshipgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// fakeRelayServer echoes every message back (single-player relay for test).
type fakeRelayServer struct {
	battleshipgrpc.UnimplementedBattleshipRelayServer
}

func (s *fakeRelayServer) GameStream(stream battleshipgrpc.BattleshipRelay_GameStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}

func setupGRPCTest(t *testing.T) (*GRPCConn, func()) {
	t.Helper()
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	battleshipgrpc.RegisterBattleshipRelayServer(srv, &fakeRelayServer{})
	go srv.Serve(lis)

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	client := battleshipgrpc.NewBattleshipRelayClient(conn)
	stream, err := client.GameStream(context.Background())
	if err != nil {
		t.Fatalf("GameStream: %v", err)
	}

	grpcConn := NewGRPCConn(stream, "TESTID", conn)

	cleanup := func() {
		grpcConn.Close()
		srv.Stop()
	}
	return grpcConn, cleanup
}

func TestGRPCConnSendReceive(t *testing.T) {
	grpcConn, cleanup := setupGRPCTest(t)
	defer cleanup()

	want := Message{Command: CmdShot, Args: map[string]string{"coord": "B3"}}
	if err := grpcConn.Send(want); err != nil {
		t.Fatalf("Send: %v", err)
	}

	got, err := grpcConn.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if got.Command != want.Command || got.Args["coord"] != want.Args["coord"] {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestNilArgsInSend(t *testing.T) {
	grpcConn, cleanup := setupGRPCTest(t)
	defer cleanup()

	// Send message with nil Args
	nilArgsMsg := Message{Command: CmdPlacementDone, Args: nil}
	if err := grpcConn.Send(nilArgsMsg); err != nil {
		t.Fatalf("Send with nil Args: %v", err)
	}

	received, err := grpcConn.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if received.Command != CmdPlacementDone {
		t.Fatalf("got %v, want %v", received.Command, CmdPlacementDone)
	}
}
