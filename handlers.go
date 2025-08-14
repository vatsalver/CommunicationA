package main

import (
	"log"
	"net/http"
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	clients[ws] = true
	log.Println("New client connected:", ws.RemoteAddr())

	defer func() {
		log.Println("Client disconnected:", ws.RemoteAddr())
		delete(clients, ws)
		for uid, conn := range userSockets {
			if conn == ws {
				delete(userSockets, uid)
				break
			}
		}
		ws.Close()
	}()

	for {
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		if msg.Type == "register" {
			userSockets[msg.From] = ws
			log.Println("Registered user:", msg.From)
			continue
		}

		broadcast <- msg
	}
}
