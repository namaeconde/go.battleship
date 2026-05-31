package network

import (
	"net"
)

// TCPConn wraps a net.Conn and implements the NetworkConn interface.
type TCPConn struct {
	conn net.Conn
}

// NewTCPConn creates a new TCPConn wrapper from a net.Conn.
func NewTCPConn(conn net.Conn) *TCPConn {
	return &TCPConn{conn: conn}
}

// Send sends a message over the connection.
func (tc *TCPConn) Send(msg Message) error {
	return Send(tc.conn, msg)
}

// Receive receives a message from the connection.
func (tc *TCPConn) Receive() (*Message, error) {
	return Receive(tc.conn)
}

// Close closes the connection.
func (tc *TCPConn) Close() error {
	return tc.conn.Close()
}
