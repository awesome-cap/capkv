package main

import (
	"github.com/awesome-cap/dkv/engine"
	"github.com/awesome-cap/dkv/net"
	"log"
)

func main() {
	e := engine.New()
	e.RegistryHandler("get", engine.GetHandler)
	e.RegistryHandler("set", engine.SetHandler)
	e.RegistryHandler("del", engine.DelHandler)

	tcpServer := net.NewTcp(":8888")
	err := tcpServer.Serve(func(args []string) ([]string, error) {
		return e.Exec(args)
	})
	if err != nil {
		log.Fatalln(err)
	}
}
