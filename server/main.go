package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handleClient(publicConn net.Conn, clientConn net.Conn) {
	defer publicConn.Close()
	defer clientConn.Close()

	go io.Copy(publicConn, clientConn)
	io.Copy(clientConn, publicConn)
}

func main() {
	publicListener, err := net.Listen("tcp", ":8080") // public port
	if err != nil {
		log.Fatal("Failed to start public listener: ", err)
	}

	defer publicListener.Close()

	fmt.Println("Public listener started on port :8080")

	for {
		publicConn, err := publicListener.Accept()
		if err != nil {
			log.Println("Failed to accept public connection: ", err)
			continue
		}

		go func(publicConn net.Conn) {
			fmt.Println("Waiting for tunnel client on :5173")
			clientConn, err := net.Dial("tcp", "localhost:5173")
			if err != nil {
				log.Println("Tunnel from client not available: ", err)
				publicConn.Close()
				return
			}

			fmt.Println("Tunnel client connected, forwarding traffic...")
			handleClient(publicConn, clientConn)
		}(publicConn)
	}
}
