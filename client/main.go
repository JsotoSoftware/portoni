package main

import (
	"flag"
	"fmt"

	"github.com/JsotoSoftware/portoni/config"
)

var localPort int

func main() {
	config.Load()

	flag.IntVar(&localPort, "port", 3000, "Local port to forward traffic to")
	flag.Parse()

	fmt.Println("Tunnel client started, forwarding traffic to port", localPort)

	handleControl()
}
