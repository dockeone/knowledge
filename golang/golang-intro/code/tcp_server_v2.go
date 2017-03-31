package main
import ("net"; "time";)

func main() {
    service := ":1200"
    tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)
    listener, _ := net.ListenTCP("tcp", tcpAddr)
    for {
        conn, _ := listener.Accept()
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    defer conn.Close()
    daytime := time.Now().String()
    conn.Write([]byte(daytime)) // don't care about return value
    // we're finished with this client
}
