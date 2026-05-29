package network

import "net"

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
