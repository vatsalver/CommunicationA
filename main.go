package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Setup routes
	http.HandleFunc("/ws", handleConnections)

	// Start a goroutine to handle messages
	go handleMessages()

	fmt.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
