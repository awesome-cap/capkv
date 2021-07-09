package main

import (
    "bufio"
    netx "github.com/awesome-cap/dkv/net"
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
    if err != nil{
        log.Fatalln(err)
    }
    nativeConn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil{
        log.Fatalln(err)
    }
    conn := netx.NewConn(nativeConn)
    go func() {
        for {
            lines, err := in.ReadBytes('\n')
            if err != nil {
                log.Println(err)
                continue
            }
            args := strings.Split(strings.TrimSpace(string(lines)), " ")
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
    if err != nil{
        log.Fatalln(err)
    }
}