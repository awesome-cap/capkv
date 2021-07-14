package main

import (
	"bufio"
	"bytes"
	netx "github.com/awesome-cap/capkv/net"
	"log"
	"net"
	"os"
	"strings"
)

var (
	in = bufio.NewReader(os.Stdin)
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":8888")
	if err != nil {
		log.Fatalln(err)
	}
	nativeConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalln(err)
	}
	conn := netx.NewConn(nativeConn)
	go func() {
		for {
			inputs, err := in.ReadBytes('\n')
			if err != nil {
				log.Println(err)
				continue
			}
			line := strings.TrimSpace(string(inputs))
			buffer := bytes.Buffer{}
			args := make([]string, 0)
			state := 0
			for i, c := range line {
				if i == 0 && c == '"' {
					log.Println("Invalid command. ")
					continue
				}
				if c == '"' && line[i-1] != '\\' {
					state ^= 1
					continue
				} else if (c == ' ' || c == '\t') && state == 0 {
					if buffer.Len() > 0 {
						args = append(args, buffer.String())
						buffer.Reset()
					}
				} else {
					buffer.WriteRune(c)
				}
			}
			if buffer.Len() > 0 {
				args = append(args, buffer.String())
			}
			err = conn.Write(args)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()
	err = conn.Accept(func(args []string, c *netx.Conn) {
		log.Println(args)
	})
	if err != nil {
		log.Fatalln(err)
	}
}
