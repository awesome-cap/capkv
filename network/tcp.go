package network

import (
    "log"
    "net"
)

type tcp struct {

}

func (t *tcp) Serve(addr string) error{
    listener, err := net.Listen("tcp", addr)
    if err != nil{
        return err
    }
    for {
        conn, err := listener.Accept()
        if err != nil{
            return err
        }
        go func() {
            err = newConn(conn).accept(func(args []string, c *Conn) {

            })
            if err != nil{
                log.Println(err)
            }
        }()
    }
}
