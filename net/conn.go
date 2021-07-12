package net

import (
	"bufio"
	"github.com/awesome-cap/dkv/ptl"
	"net"
)

type Conn struct {
	reader *bufio.Reader
	writer *bufio.Writer
	bytes  []byte
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		bytes:  make([]byte, 0),
	}
}

func (c *Conn) Read() ([]string, error) {
	return ptl.UnMarshal(c.reader)
}

func (c *Conn) Write(args []string) error {
	bytes, err := ptl.Marshal(args)
	if err != nil {
		return err
	}
	_, err = c.writer.Write(bytes)
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Conn) Accept(apply func(args []string, c *Conn)) error {
	for {
		args, err := c.Read()
		if err != nil {
			return err
		}
		apply(args, c)
	}
}
