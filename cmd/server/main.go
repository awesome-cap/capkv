package main

import (
	"github.com/awesome-cap/capkv/config"
	"github.com/awesome-cap/capkv/engine"
	"github.com/awesome-cap/capkv/net"
	"log"
	"os"
)

func main() {
	conf := config.Default()
	if len(os.Args) > 1 {
		var err error
		conf, err = config.Parse(os.Args[1])
		if err != nil {
			log.Panicln(err)
		}
	}

	e, err := engine.New(conf)
	if err != nil {
		log.Panicln(err)
	}

	tcpServer := net.NewTcp(":8888")
	err = tcpServer.Serve(func(args []string) ([]string, error) {
		return e.Exec(args)
	})
	if err != nil {
		log.Panicln(err)
	}
}
