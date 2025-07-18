package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func openTunnel() {
	serverConn, err := net.Dial("tcp", "localhost:9091")
	if err != nil {
		log.Println("Failed to connect to tunnel port:", err)
		return
	}
	defer serverConn.Close()

	fmt.Println("Tunnel connection established")

	localConn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Println("Failed to connect to local service:", err)
		serverConn.Close()
		return
	}
	defer localConn.Close()

	fmt.Println("Local service connection established")

	go io.Copy(localConn, serverConn)
	io.Copy(serverConn, localConn)
}
