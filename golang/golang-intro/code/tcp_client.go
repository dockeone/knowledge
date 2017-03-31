package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "os"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s host:port ", os.Args[0])
        os.Exit(1)
    }
    service := os.Args[1]
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)
    conn, _ := net.DialTCP("tcp", nil, tcpAddr)
    _, _ = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
    result, _ := ioutil.ReadAll(conn)
    fmt.Println(string(result))
    os.Exit(0)
}
