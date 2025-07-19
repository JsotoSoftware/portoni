package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/JsotoSoftware/portoni/config"
)

type TunnelClient struct {
	conn      net.Conn
	localPort int
}

var (
	controlConn    net.Conn
	controlLock    sync.Mutex
	tunnelConnChan = make(chan net.Conn)
	tunnelRegistry = make(map[string]TunnelClient)
	tunnelMutex    = sync.Mutex{}
)

func handleControlConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Failed to read registration message:", err)
		conn.Close()
		return
	}

	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "REGISTER") {
		log.Println("Invalid registration message:", line)
		conn.Close()
		return
	}

	var localPort int
	fmt.Sscanf(line, "REGISTER %d", &localPort)

	tunnelID := generateTunnelID()

	tunnelMutex.Lock()
	tunnelRegistry[tunnelID] = TunnelClient{
		conn:      conn,
		localPort: localPort,
	}
	tunnelMutex.Unlock()

	controlLock.Lock()
	controlConn = conn
	controlLock.Unlock()

	fmt.Fprintf(conn, "%s\n", tunnelID)
	log.Printf("Registered new tunnel: %s → localhost:%d\n", tunnelID, localPort)

	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Tunnel %s disconnected.\n", tunnelID)
			tunnelMutex.Lock()
			delete(tunnelRegistry, tunnelID)
			tunnelMutex.Unlock()

			controlLock.Lock()
			if controlConn == conn {
				controlConn = nil
			}
			controlLock.Unlock()

			conn.Close()
			return
		}
	}
}

func listenControlPort() {
	controlPort := config.Get("CONTROL_PORT", "9090")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", controlPort))
	if err != nil {
		log.Fatal("Failed to listen to control port:", err)
	}
	defer listener.Close()

	fmt.Printf("Control port listening on %s\n", controlPort)

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
	tunnelPort := config.Get("TUNNEL_PORT", "9091")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", tunnelPort))
	if err != nil {
		log.Fatal("Failed to listen to tunnel port:", err)
	}
	defer listener.Close()

	fmt.Printf("Tunnel port listening on %s\n", tunnelPort)

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
	publicPort := config.Get("PUBLIC_PORT", "8080")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", publicPort))
	if err != nil {
		log.Fatal("Failed to listen to public port:", err)
	}
	defer listener.Close()

	fmt.Printf("Public server listening on %s\n", publicPort)

	for {
		publicConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting public connection:", err)
			continue
		}

		go func(publicConn net.Conn) {
			defer publicConn.Close()
			fmt.Println("Public request received, Sending REQ to client...")

			controlLock.Lock()
			if controlConn == nil {
				log.Println("No control connection available")
				controlLock.Unlock()
				return
			}

			_, err := controlConn.Write([]byte("REQ\n"))
			if err != nil {
				log.Println("Error sending REQ to client:", err)
				controlConn = nil
				controlLock.Unlock()
				return
			}
			controlLock.Unlock()

			tunnelConn := <-tunnelConnChan
			defer tunnelConn.Close()

			fmt.Println("✅ Tunnel connection established, forwarding traffic...")

			go io.Copy(tunnelConn, publicConn)
			io.Copy(publicConn, tunnelConn)
		}(publicConn)
	}
}

func generateTunnelID() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	for {
		id := make([]rune, 6)
		for i := range id {
			id[i] = letters[rand.Intn(len(letters))]
		}
		tunnelID := string(id)

		tunnelMutex.Lock()
		_, exists := tunnelRegistry[tunnelID]
		tunnelMutex.Unlock()

		if !exists {
			return tunnelID
		}
	}
}

func main() {
	config.Load()

	go listenControlPort()
	go acceptTunnelConnections()
	go listenPublicPort()

	select {}
}
