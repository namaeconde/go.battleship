package network

import (
	"encoding/json"
	"fmt"
	"net"
)

type Command string

const (
	CmdConnectRequest Command = "CONNECT_REQUEST"
	CmdConnectAck     Command = "CONNECT_ACK"
	CmdPlacementDone  Command = "PLACEMENT_DONE"
	CmdShot           Command = "SHOT"
	CmdShotResult     Command = "SHOT_RESULT"
	CmdQuit           Command = "QUIT"
)

type Message struct {
	Command Command           `json:"command"`
	Args    map[string]string `json:"args,omitempty"`
}

// SerializeMessage serializes a message to JSON.
func SerializeMessage(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}

// DeserializeMessage deserializes a message from JSON.
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}
	return &msg, nil
}

// Send sends a message over the connection.
func Send(conn net.Conn, msg Message) error {
	data, err := SerializeMessage(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(data, '\n'))
	return err
}

// Receive receives a message from the connection.
func Receive(conn net.Conn) (*Message, error) {
	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
