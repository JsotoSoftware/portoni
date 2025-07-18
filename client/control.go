package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleControl() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		log.Fatal("Failed to connect to tunnel server control port:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server control port")

	reader := bufio.NewReader(conn)
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
