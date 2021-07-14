package net

import (
	"io"
	"log"
	"net"
)

type Network interface {
	Serve(addr ...string) error
}

type tcp struct {
	addr string
}

func NewTcp(addr string) *tcp {
	return &tcp{addr: addr}
}

func (t *tcp) Serve(handle func(args []string) ([]string, error)) error {
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}
	log.Println("Tcp server listening on ", t.addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			err = NewConn(conn).Accept(func(args []string, c *Conn) {
				results, err := handle(args)
				if err != nil {
					_ = c.Write([]string{"fail", err.Error()})
					return
				}
				_ = c.Write(append([]string{"ok"}, results...))
			})
			if err != nil && err != io.EOF {
				log.Println(err)
			}
		}()
	}
}
