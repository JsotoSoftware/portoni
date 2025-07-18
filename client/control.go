package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/JsotoSoftware/portoni/config"
)

var tunnelID string

func handleControl() {
	controlPort := config.Get("CONTROL_PORT", "1994")
	serverHost := config.Get("SERVER_HOST", "localhost")
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverHost, controlPort))
	if err != nil {
		log.Fatal("Failed to connect to tunnel server control port:", controlPort, err)
	}
	defer conn.Close()

	fmt.Println("Connected to server control port")

	fmt.Fprintf(conn, "REGISTER %d\n", localPort)

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to receive tunnel ID:", err)
		return
	}

	tunnelID = strings.TrimSpace(line)
	fmt.Printf("Assigned public URL: https://%s.portoni.josuesyc.com\n", tunnelID)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Control connection closed:", err)
			return
		}

		line = strings.TrimSpace(line)
		fmt.Println("Received control message:", line)

		if line == "REQ" {
			go openTunnel()
		}
	}
}
