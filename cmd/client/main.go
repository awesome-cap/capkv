package main

import (
	"bufio"
	"bytes"
	"github.com/awesome-cap/capkv/client"
	"log"
	"os"
	"strings"
)

var (
	in = bufio.NewReader(os.Stdin)
)

func main() {
	connect, err := client.New(":8888").Connect()
	if err != nil {
		log.Panicln(err)
	}
	for {
		inputs, err := in.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		line := strings.TrimSpace(string(inputs))
		if line == "exit" {
			break
		}
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
		resp, err := connect.Cmd(args...)
		log.Println("error: ", err, "response: ", resp)
	}
}
