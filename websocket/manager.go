package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"real-time-forum/database"
	"time"
)

var Clients = make(map[int]*Client) // map userID to client


func BroadcastTo(receiverID int, message []byte) {
	if client, ok := Clients[receiverID]; ok {
		client.Send <- message
	}
}

// WebSocketMessage represents the structure of incoming/outgoing messages
type WebSocketMessage struct {
	Type       string    `json:"type"`
	SenderID   int       `json:"senderId"`
	ReceiverID int       `json:"receiverId"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

func savePrivateMessage(msg WebSocketMessage) error {
	query := `
		INSERT INTO PrivateMessages (sender_id, receiver_id, message_text, sent_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := DB.Exec(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.Timestamp)
	return err
}

// DB is a global reference to your database connection (you should initialize this in main)
var DB *sql.DB

func handleIncomingMessage(raw []byte) {
	var msg WebSocketMessage
	err := json.Unmarshal(raw, &msg)
	if err != nil {
		fmt.Println("Error parsing message:", err)
		return
	}

	switch msg.Type {
	case "message":
		// 1. Save to database
		err = savePrivateMessage(msg)
		if err != nil {
			fmt.Println("Error saving message:", err)
		}

		// 2. Forward to receiver if online
		response, _ := json.Marshal(msg)
		BroadcastTo(msg.ReceiverID, response)

	case "typing":
		response, _ := json.Marshal(msg)
		BroadcastTo(msg.ReceiverID, response)

	case "status":
		// Update DB user status
		err := database.UpdateUserStatus(DB, int64(msg.SenderID), msg.Content == "online")
		if err != nil {
			fmt.Println("Failed to update user status:", err)
		}

		// Broadcast to all users (optional)
		response, _ := json.Marshal(msg)
		for _, client := range Clients {
			client.Send <- response
		}
	default:
		fmt.Println("Unknown message type:", msg.Type)
	}
}

func BroadcastUserStatus(db *sql.DB, userID int, isOnline bool) {
	msg := map[string]interface{}{
		"type": "status",
		"content": map[string]interface{}{
			"userId":   userID,
			"isOnline": isOnline,
		},
		"timestamp": time.Now().Unix(),
	}

	msgBytes, _ := json.Marshal(msg)

	for _, client := range Clients {
		if client.UserID != userID {
			select {
			case client.Send <- msgBytes:
			default:
				close(client.Send)
				delete(Clients, client.UserID)
			}
		}
	}
}
