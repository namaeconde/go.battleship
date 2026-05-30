// network/grpc_conn.go
package network

import (
	battleshipgrpc "go.battleship/proto/battleshipgrpc"
)

// GRPCConn wraps a BattleshipRelay_GameStreamClient and implements game.NetworkConn.
type GRPCConn struct {
	stream battleshipgrpc.BattleshipRelay_GameStreamClient
	gameID string
}

// NewGRPCConn creates a GRPCConn from an open GameStream client.
func NewGRPCConn(stream battleshipgrpc.BattleshipRelay_GameStreamClient, gameID string) *GRPCConn {
	return &GRPCConn{stream: stream, gameID: gameID}
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

// Close half-closes the stream (signals no more sends).
func (c *GRPCConn) Close() error {
	return c.stream.CloseSend()
}
