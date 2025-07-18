package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request:", r.Method, r.URL.Path)
	fmt.Fprintf(w, "This is a local server for tunneling test!")
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Local server listening on port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
