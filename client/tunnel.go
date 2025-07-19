package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/JsotoSoftware/portoni/config"
)

func openTunnel() {
	tunnelPort := config.Get("TUNNEL_PORT", "2001")
	serverHost := config.Get("SERVER_HOST", "localhost")

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverHost, tunnelPort))
	if err != nil {
		log.Println("❌ Failed to connect to tunnel port:", err)
		return
	}
	log.Println("✅ Tunnel connection established")

	localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		log.Println("❌ Failed to connect to local service on localhost:", localPort, err)
		serverConn.Close()
		return
	}
	log.Println("✅ Local service connection established")

	// Start bidirectional streaming
	go func() {
		defer localConn.Close()
		defer serverConn.Close()
		_, err := io.Copy(localConn, serverConn)
		if err != nil {
			log.Println("Error copying server → local:", err)
		}
	}()

	go func() {
		defer localConn.Close()
		defer serverConn.Close()
		_, err := io.Copy(serverConn, localConn)
		if err != nil {
			log.Println("Error copying local → server:", err)
		}
	}()
}
