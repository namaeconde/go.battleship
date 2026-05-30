package network

import (
	"context"
	"net"
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
