package main

import (
    "fmt"
    "net"
    "log"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Got connection!")
	p := make([]byte, 2048)
	n, _ := conn.Read(p)
    fmt.Printf("Received : %s\n", p[:n]);
}

func main() {
    fmt.Println("start")
    //laddr, _ := net.ResolveTCPAddr("tcp", "129.241.187.159:20011")
    ln, err := net.Listen("tcp", ":20011")
    if err != nil {
        log.Fatal(err)
    }
    defer ln.Close()
    for {
        conn, err := ln.Accept() // this blocks until connection or error
        if err != nil {
            fmt.Println("Error occured")
            continue
        }
        go handleConnection(conn) // a goroutine handles conn so that the loop can accept other connections
    }
}
