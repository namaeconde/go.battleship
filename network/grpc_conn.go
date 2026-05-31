// network/grpc_conn.go
package network

import (
	battleshipgrpc "go.battleship/proto/battleshipgrpc"
	"google.golang.org/grpc"
)

// GRPCConn wraps a BattleshipRelay_GameStreamClient and implements game.NetworkConn.
type GRPCConn struct {
	stream battleshipgrpc.BattleshipRelay_GameStreamClient
	gameID string
	conn   *grpc.ClientConn // underlying transport; closed by Close()
}

// NewGRPCConn creates a GRPCConn from an open GameStream client and its parent connection.
// Pass nil for conn when managing the connection lifecycle externally (e.g. in tests).
func NewGRPCConn(stream battleshipgrpc.BattleshipRelay_GameStreamClient, gameID string, conn *grpc.ClientConn) *GRPCConn {
	return &GRPCConn{stream: stream, gameID: gameID, conn: conn}
}

// Send encodes a network.Message as a GameMessage proto and sends it on the stream.
func (c *GRPCConn) Send(msg Message) error {
	protoArgs := make(map[string]string, len(msg.Args))
	for k, v := range msg.Args {
		protoArgs[k] = v
	}
	return c.stream.Send(&battleshipgrpc.GameMessage{
		GameId:  c.gameID,
		Command: string(msg.Command),
		Args:    protoArgs,
	})
}

// Receive reads the next GameMessage from the stream and decodes it into a network.Message.
func (c *GRPCConn) Receive() (*Message, error) {
	protoMsg, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	args := make(map[string]string, len(protoMsg.Args))
	for k, v := range protoMsg.Args {
		args[k] = v
	}
	return &Message{
		Command: Command(protoMsg.Command),
		Args:    args,
	}, nil
}

// Close half-closes the stream and shuts down the underlying connection.
func (c *GRPCConn) Close() error {
	err := c.stream.CloseSend()
	if c.conn != nil {
		if cerr := c.conn.Close(); err == nil {
			err = cerr
		}
	}
	return err
}
