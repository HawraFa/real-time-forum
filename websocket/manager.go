package websocket

import (
	"database/sql"
	"encoding/json"
	// "fmt"
	"log"
	"time"
	"real-time-forum/database"
)

var DB *sql.DB

type WebSocketMessage struct {
	Type       string    `json:"type"`
	SenderID   int       `json:"senderId,omitempty"`
	ReceiverID int       `json:"receiverId,omitempty"`
	Content    string    `json:"content,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	Username   string    `json:"username,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
	UserIds    []int     `json:"userIds,omitempty"`
	UserID     int       `json:"userId,omitempty"`
}

func BroadcastTo(receiverID int, message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if conns, ok := Clients[receiverID]; ok {
		for client := range conns {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(conns, client)
			}
		}
	}
}

func savePrivateMessage(msg WebSocketMessage) error {
	query := `INSERT INTO PrivateMessages (sender_id, receiver_id, message_text, sent_at) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.Timestamp)
	return err
}

func handleIncomingMessage(raw []byte) {
	var msg WebSocketMessage
	_ = json.Unmarshal(raw, &msg)

	switch msg.Type {
	case "message":
		// Check if receiver is online
		clientsMu.Lock()
		receiverOnline := len(Clients[msg.ReceiverID]) > 0
		clientsMu.Unlock()

		if !receiverOnline {
			// Send error message back to sender
			errorMsg := WebSocketMessage{
				Type:    "error",
				Content: "Cannot send message: User is offline",
			}
			errorData, _ := json.Marshal(errorMsg)
			BroadcastTo(msg.SenderID, errorData)
			return
		}

		err := database.SaveMessage(DB, int64(msg.SenderID), int64(msg.ReceiverID), msg.Content)
		if err != nil {
			log.Println("❌ SaveMessage failed:", err)
		} else {
			log.Println("✅ SaveMessage succeeded for", msg.SenderID, "->", msg.ReceiverID)
		}

		if conns, ok := Clients[msg.SenderID]; ok {
			for c := range conns {
				msg.Username = c.Username
				msg.Avatar = c.Avatar
				break
			}
		}
		if msg.Timestamp.IsZero() {
			msg.Timestamp = time.Now()
		}
		response, _ := json.Marshal(msg)
		BroadcastTo(msg.ReceiverID, response)

	case "typing":
		// Check if receiver is online before sending typing status
		clientsMu.Lock()
		receiverOnline := len(Clients[msg.ReceiverID]) > 0
		clientsMu.Unlock()

		if receiverOnline {
			response, _ := json.Marshal(msg)
			BroadcastTo(msg.ReceiverID, response)
		}

	case "status":
		var username, avatar string
		if conns, ok := Clients[msg.SenderID]; ok {
			for c := range conns {
				username = c.Username
				avatar = c.Avatar
				break
			}
		}
		statusManager := GetStatusManager()
		if msg.Content == "online" {
			statusManager.SetOnline(int64(msg.SenderID), username, avatar)
		} else {
			statusManager.SetOffline(int64(msg.SenderID))
		}
		BroadcastUserStatus(int64(msg.SenderID), msg.Content == "online", username, avatar)
	}
}

func BroadcastUserStatus(userID int64, online bool, username, avatar string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	status := "offline"
	if online {
		status = "online"
	}

	msg := map[string]interface{}{
		"type":     "status",
		"senderId": userID,
		"content":  status,
		"username": username,
		"avatar":   avatar,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("❌ Error marshaling status message:", err)
		return
	}

	for uid, conns := range Clients {
		for client := range conns {
			select {
			case client.Send <- data:
				log.Printf("📤 Sent status '%s' of user %d to user %d", status, userID, uid)
			default:
				log.Printf("⚠️ Skipped sending to client %d due to full channel", client.UserID)
			}
		}
	}
}

func BroadcastPresenceToNewUser(newUserID int, newClient *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for uid, conns := range Clients {
		if uid == newUserID {
			continue
		}
		for client := range conns {
			msg := map[string]interface{}{
				"type":     "status",
				"senderId": uid,
				"content":  "online",
				"username": client.Username,
				"avatar":   client.Avatar,
			}
			data, _ := json.Marshal(msg)
			select {
			case newClient.Send <- data:
				log.Printf("👋 Telling user %d that %d is online", newUserID, uid)
			default:
				log.Printf("⚠️ Could not tell user %d that %d is online", newUserID, uid)
			}
			break
		}
	}
}

func GetOnlineUsers() []int {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	onlineUsers := make([]int, 0)
	for uid := range Clients {
		onlineUsers = append(onlineUsers, uid)
	}
	return onlineUsers
}
