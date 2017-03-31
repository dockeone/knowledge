package main

import (
    "net"
    "time"
)

func main() {
    service := ":1200"
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)
    listener, _ := net.ListenTCP("tcp", tcpAddr)
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        daytime := time.Now().String()
        conn.Write([]byte(daytime)) // don't care about return value
        conn.Close()                // we're finished with this client
    }
}
