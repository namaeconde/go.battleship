// game/network_conn_test.go
package game

import (
	"io"
	"testing"
	"go.battleship/network"
)

// fakeConn implements NetworkConn for tests.
type fakeConn struct {
	sent   []network.Message
	inbox  []network.Message
	closed bool
}

func (f *fakeConn) Send(msg network.Message) error {
	f.sent = append(f.sent, msg)
	return nil
}

func (f *fakeConn) Receive() (*network.Message, error) {
	if len(f.inbox) == 0 {
		return nil, io.EOF
	}
	msg := f.inbox[0]
	f.inbox = f.inbox[1:]
	return &msg, nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func TestNetworkConnInterface(t *testing.T) {
	// Verify fakeConn satisfies NetworkConn at compile time.
	var _ NetworkConn = (*fakeConn)(nil)
}
