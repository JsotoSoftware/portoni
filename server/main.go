package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var (
	controlConn    net.Conn
	controlLock    sync.Mutex
	tunnelConnChan = make(chan net.Conn)
)

func handleControlConnection(conn net.Conn) {
	controlLock.Lock()
	controlConn = conn
	controlLock.Unlock()
	fmt.Println("Tunnel client connected to control port")
}

func listenControlPort() {
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal("Failed to listen to 9090:", err)
	}
	defer listener.Close()

	fmt.Println("Control port listening on :9090")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept control connection:", err)
			continue
		}
		go handleControlConnection(conn)
	}
}

func acceptTunnelConnections() {
	listener, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatal("Failed to listen to 9091:", err)
	}
	defer listener.Close()

	fmt.Println("Tunnel port listening on :9091")

	for {
		tunnelConn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept tunnel connection:", err)
			continue
		}
		tunnelConnChan <- tunnelConn
	}
}

func listenPublicPort() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen to 8080:", err)
	}
	defer listener.Close()

	fmt.Println("Public server listening on :8080")

	for {
		publicConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting public connection:", err)
			continue
		}

		go func(publicConn net.Conn) {
			fmt.Println("Public request received, Sending REQ to client...")

			controlLock.Lock()
			if controlConn == nil {
				log.Println("No control connection available")
				publicConn.Close()
				controlLock.Unlock()
				return
			}

			_, err := controlConn.Write([]byte("REQ\n"))
			controlLock.Unlock()

			if err != nil {
				log.Println("Error sending REQ to client:", err)
				publicConn.Close()
				return
			}

			tunnelConn := <-tunnelConnChan

			fmt.Println("Tunnel connection established, forwarding traffic...")

			go io.Copy(tunnelConn, publicConn)
			io.Copy(publicConn, tunnelConn)
		}(publicConn)
	}
}

func main() {
	go listenControlPort()
	go acceptTunnelConnections()
	go listenPublicPort()

	// Keep the main thread alive
	select {}
}
