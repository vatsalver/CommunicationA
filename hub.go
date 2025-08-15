package main

import (
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var userSockets = make(map[string]*websocket.Conn) // userID -> WebSocket
var groups = make(map[string][]string)             // groupID -> []userIDs

var broadcast = make(chan Message, 8)

type Message struct {
	Type      string `json:"type"`      // broadcast, dm, group, create_group, join_group, register
	From      string `json:"from"`      // sender
	To        string `json:"to"`        // for dm or p2p calls
	Group     string `json:"group"`     // group name
	Content   string `json:"content"`   // text
	SDP       string `json:"sdp"`       // offer/answer payload
	Candidate string `json:"candidate"` // ICE candidate JSON
}

func handleMessages() {
	for {
		msg := <-broadcast

		switch msg.Type {

		// Existing Messaging Functions
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

		// New: WebRTC Signaling
		case "start_video":
			// Notify target or group
			if msg.To != "" {
				if target, ok := userSockets[msg.To]; ok {
					target.WriteJSON(msg)
				}
			} else if msg.Group != "" {
				if members, ok := groups[msg.Group]; ok {
					for _, memberID := range members {
						if memberID != msg.From {
							if conn, connected := userSockets[memberID]; connected {
								conn.WriteJSON(msg)
							}
						}
					}
				}
			}

		case "video_offer", "video_answer", "ice_candidate":
			if msg.To != "" {
				if target, ok := userSockets[msg.To]; ok {
					target.WriteJSON(msg)
				}
			} else if msg.Group != "" {
				if members, ok := groups[msg.Group]; ok {
					for _, memberID := range members {
						if memberID != msg.From {
							if conn, connected := userSockets[memberID]; connected {
								conn.WriteJSON(msg)
							}
						}
					}
				}
			}
		}
	}
}
