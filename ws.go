package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader config — allows all origins for testing
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// NOTE: In production, add proper origin checks!
		return true
	},
}
