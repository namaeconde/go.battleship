package network

import "net"
import "context"

type Command string
const (
    CmdConnectRequest Command = "CONNECT_REQUEST"
    CmdConnectAck     Command = "CONNECT_ACK"
    CmdQuit           Command = "QUIT"
)

type Message struct {
    Command Command
    Args    []string
}

func ReadMessage(conn net.Conn) (string, error) { return "", nil }
func ParseMessage(s string) (*Message, error) { return &Message{}, nil }
func WriteMessage(conn net.Conn, msg string) error { return nil }
func CreateMessage(cmd Command, args ...string) string { return "" }
func StartHost(ctx context.Context, port string) (net.Conn, error) { return nil, nil }
func ConnectToHost(ctx context.Context, addr string) (net.Conn, error) { return nil, nil }
