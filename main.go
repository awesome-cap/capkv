package main

import (
	"github.com/awesome-cap/dkv/config"
	"github.com/awesome-cap/dkv/engine"
	"github.com/awesome-cap/dkv/net"
	"log"
)

func main() {
	e, err := engine.New(config.Config{
		Storage: config.Storage{
			Dir: "D:\\kv\\data",
			Log: config.Log{
				Enable: true,
			},
			DB: config.DB{
				Enable:        true,
				FlushMethod:   1,
				FlushInterval: 60,
				FilingSize:    1048576,
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	e.Registry(engine.Get)
	e.Registry(engine.Set)
	e.Registry(engine.Del)

	tcpServer := net.NewTcp(":8888")
	err = tcpServer.Serve(func(args []string) ([]string, error) {
		return e.Exec(args)
	})
	if err != nil {
		log.Fatalln(err)
	}
}
