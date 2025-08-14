package main

import (
	"github.com/gorilla/websocket"
)

// ----- Global State -----

var clients = make(map[*websocket.Conn]bool)     // WebSocket connections
var userSockets = make(map[string]*websocket.Conn) // userID -> WebSocket
var groups = make(map[string][]string)           // groupID -> list of userIDs

var broadcast = make(chan Message, 8) // Outgoing messages

// ----- Message Structure -----
type Message struct {
	Type    string `json:"type"`    // broadcast, dm, group, create_group, join_group, register
	From    string `json:"from"`    // Sender userID
	To      string `json:"to"`      // Recipient userID (for DM)
	Group   string `json:"group"`   // Group name (for group chat)
	Content string `json:"content"` // Text message
}

// ----- Message Router -----
func handleMessages() {
	for {
		msg := <-broadcast

		switch msg.Type {

		case "broadcast":
			for client := range clients {
				client.WriteJSON(msg)
			}

		case "dm":
			if targetConn, ok := userSockets[msg.To]; ok {
				targetConn.WriteJSON(msg)
			}

		case "create_group":
			if _, exists := groups[msg.Group]; !exists {
				groups[msg.Group] = []string{msg.From}
			}
			if creator, ok := userSockets[msg.From]; ok {
				creator.WriteJSON(Message{Type: "system", Content: "Group " + msg.Group + " created"})
			}

		case "join_group":
			if _, exists := groups[msg.Group]; exists {
				// Avoid duplicates
				for _, m := range groups[msg.Group] {
					if m == msg.From {
						return
					}
				}
				groups[msg.Group] = append(groups[msg.Group], msg.From)
				if user, ok := userSockets[msg.From]; ok {
					user.WriteJSON(Message{Type: "system", Content: "Joined group " + msg.Group})
				}
			}

		case "group":
			if members, ok := groups[msg.Group]; ok {
				for _, memberID := range members {
					if conn, connected := userSockets[memberID]; connected {
						conn.WriteJSON(msg)
					}
				}
			}
		}
	}
}
