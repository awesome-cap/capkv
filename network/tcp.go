package network

import (
    "fmt"
    "log"
    "net"
)

type tcp struct {
    addr string
    exec *Executor
}

func NewTcp(addr string, exec *Executor) *tcp{
    return &tcp{addr: addr}
}

func (t *tcp) Serve() error{
    listener, err := net.Listen("tcp", t.addr)
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
                v, err := t.exec.Exec(args)
                if err != nil {
                    _ = c.Write([]string{"1", err.Error()})
                    return
                }
                _ = c.Write([]string{"0", fmt.Sprintf("%v", v)})
            })
            if err != nil{
                log.Println(err)
            }
        }()
    }
}
