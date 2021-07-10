package net

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
)

type Conn struct {
	reader *bufio.Reader
	writer *bufio.Writer
	bytes []byte
}

func NewConn(conn net.Conn) *Conn{
	return &Conn{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		bytes: make([]byte, 0),
	}
}

func (c *Conn) writeUint16(i uint16) error {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, i)
	_, err := c.writer.Write(data)
	return err
}

func (c *Conn) writeUint32(i uint32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, i)
	_, err := c.writer.Write(data)
	return err
}

func (c *Conn) nextUint16() (uint16, error){
	data := make([]byte, 2)
	_, err := io.ReadFull(c.reader, data)
	if err != nil{
		return 0, err
	}
	return binary.BigEndian.Uint16(data), nil
}

func (c *Conn) nextUint32() (uint32, error){
	data := make([]byte, 4)
	_, err := io.ReadFull(c.reader, data)
	if err != nil{
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

func (c *Conn) nextBytes(size int) ([]byte, error){
	data := make([]byte, size)
	_, err := io.ReadFull(c.reader, data)
	if err != nil{
		return nil, err
	}
	return data, nil
}

func (c *Conn) Read() ([]string, error){
	count, err := c.nextUint16()
	if err != nil{
		return nil, err
	}
	args := make([]string, count)
	for i := 0; i < int(count); i ++{
		size, err := c.nextUint32()
		if err != nil{
			return nil, err
		}
		data, err := c.nextBytes(int(size))
		if err != nil{
			return nil, err
		}
		args[i] = string(data)
	}
	return args, nil
}

func (c *Conn) Write(args []string) error{
	err := c.writeUint16(uint16(len(args)))
	if err != nil{
		return err
	}
	for _, data := range args{
		err = c.writeUint32(uint32(len(data)))
		if err != nil{
			return err
		}
		_, err := c.writer.Write([]byte(data))
		if err != nil{
			return err
		}
	}
	return c.writer.Flush()
}

func (c *Conn) Accept(apply func(args []string, c *Conn)) error{
	for {
		args, err := c.Read()
		if err != nil{
			return err
		}
		apply(args, c)
	}
}


