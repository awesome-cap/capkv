package net

import (
    "log"
    "net"
)

type tcp struct {
    addr string
}

func NewTcp(addr string) *tcp{
    return &tcp{addr: addr}
}

func (t *tcp) Serve(handle func(args []string) ([]string, error)) error{
    listener, err := net.Listen("tcp", t.addr)
    if err != nil{
        return err
    }
    log.Println("Tcp server listening on ", t.addr)
    for {
        conn, err := listener.Accept()
        if err != nil{
            return err
        }
        go func() {
            err = NewConn(conn).Accept(func(args []string, c *Conn) {
                results, err := handle(args)
                if err != nil {
                    _ = c.Write([]string{"0", err.Error()})
                    return
                }
                _ = c.Write(append([]string{"1"}, results...))
            })
            if err != nil{
                log.Println(err)
            }
        }()
    }
}
