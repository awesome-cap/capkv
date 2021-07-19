package client

import (
	"errors"
	netx "github.com/awesome-cap/kv/net"
	"log"
	"net"
)

type Client struct {
	addr string
}

type Connect struct {
	conn *netx.Conn
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Connect() (*Connect, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.addr)
	if err != nil {
		log.Panicln(err)
	}
	nativeConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Panicln(err)
	}
	return &Connect{
		conn: netx.NewConn(nativeConn),
	}, nil
}

func (c *Connect) Cmd(args ...string) ([]string, error) {
	err := c.conn.Write(args)
	if err != nil {
		return nil, err
	}
	resp, err := c.conn.Read()
	if err != nil {
		return nil, err
	}
	if len(resp) < 2 {
		return nil, errors.New("Server response error. ")
	}
	if resp[0] == "fail" {
		return nil, errors.New(resp[1])
	}
	return resp[1:], nil
}
