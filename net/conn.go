package net

import (
	"net"
)

type Conn struct {
	conn net.Conn
	bytes []byte
}

func NewConn(conn net.Conn) *Conn{
	return &Conn{
		conn: conn,
		bytes: make([]byte, 0),
	}
}

func (c *Conn) read() error{
	buffered := make([]byte, 1024)
	n, err := c.conn.Read(buffered)
	if err != nil{
		return err
	}
	c.bytes = append(c.bytes, buffered[0:n]...)
	return nil
}

func (c *Conn) parse() ([][]string, error) {
	data := make([][]string, 0)
	for {
		msg, index, err := decode(c.bytes)
		if err != nil{
			break
		}
		data = append(data, msg)
		c.bytes = c.bytes[index:]
	}
	return data, nil
}

func (c *Conn) Accept(apply func(args []string, c *Conn)) error{
	for {
		err := c.read()
		if err != nil {
			return err
		}
		data, err := c.parse()
		if err != nil {
			return err
		}
		for _, args := range data{
			apply(args, c)
		}
	}
}

func (c *Conn) Write(args []string) error{
	_, err := c.conn.Write(encode(args))
	return err
}
