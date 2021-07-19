package tests

import (
	"github.com/awesome-cap/kv/client"
	"github.com/awesome-cap/kv/config"
	"github.com/awesome-cap/kv/engine"
	"github.com/awesome-cap/kv/net"
	"log"
	"strconv"
	"testing"
)

const addr = ":9999"

var connect *client.Connect

func init() {
	e, err := engine.New(config.Default())
	if err != nil {
		log.Panicln(err)
	}

	go func() {
		tcpServer := net.NewTcp(addr)
		err = tcpServer.Serve(func(args []string) ([]string, error) {
			return e.Exec(args)
		})
		if err != nil {
			panic(err)
		}
	}()

	c := client.New(addr)
	connect, err = c.Connect()
	if err != nil {
		panic(err)
	}
}

func BenchmarkSet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		_, err := connect.Cmd("set", is, is)
		if err != nil {
			b.Fatal(err, i)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is := strconv.Itoa(i)
		_, err := connect.Cmd("get", is)
		if err != nil {
			b.Fatal(err, i)
		}
	}
}
