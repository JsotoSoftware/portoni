package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/JsotoSoftware/portoni/config"
)

func openTunnel() {
	tunnelPort := config.Get("TUNNEL_PORT", "9091")
	serverHost := config.Get("SERVER_HOST", "localhost")

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverHost, tunnelPort))
	if err != nil {
		log.Println("Failed to connect to tunnel port:", err)
		return
	}
	defer serverConn.Close()

	fmt.Println("Tunnel connection established")

	localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverHost, localPort))
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
