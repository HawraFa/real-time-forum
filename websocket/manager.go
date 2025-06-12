package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

var DB *sql.DB // Global database connection

// WebSocketMessage represents the structure of incoming/outgoing messages
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
	
	if client, ok := Clients[receiverID]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(Clients, receiverID)
		}
	}
}

func savePrivateMessage(msg WebSocketMessage) error {
	query := `
		INSERT INTO PrivateMessages (sender_id, receiver_id, message_text, sent_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := DB.Exec(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.Timestamp)
	return err
}

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
			return
		}

		// 2. Add sender info to message
		if sender, ok := Clients[msg.SenderID]; ok {
			msg.Username = sender.Username
			msg.Avatar = sender.Avatar
		}

		// 3. Forward to receiver if online
		response, _ := json.Marshal(msg)
		BroadcastTo(msg.ReceiverID, response)

	case "typing":
		response, _ := json.Marshal(msg)
		BroadcastTo(msg.ReceiverID, response)

    case "get_online_users":
        // Send list of online users to requesting client
        onlineUsers := GetOnlineUsers()
        response := WebSocketMessage{
            Type:    "online_users",
            UserIds: onlineUsers,
        }
        msgBytes, _ := json.Marshal(response)
        BroadcastTo(msg.SenderID, msgBytes)	

	case "status":
		// Handle explicit status updates (e.g., "away", "busy")
		statusManager := GetStatusManager()
		if msg.Content == "online" {
			// Get client details if available
			var username, avatar string
			if client, ok := Clients[msg.SenderID]; ok {
				username = client.Username
				avatar = client.Avatar
			}
			statusManager.SetOnline(int64(msg.SenderID), username, avatar)
		} else {
			statusManager.SetOffline(int64(msg.SenderID))
		}
		
		// Broadcast the status change
		if client, ok := Clients[msg.SenderID]; ok {
			BroadcastUserStatus(int64(msg.SenderID), msg.Content == "online", client.Username, client.Avatar)
		}

	default:
		fmt.Println("Unknown message type:", msg.Type)
	}
}

func BroadcastUserStatus(userID int64, isOnline bool, username, avatar string) {
	statusMsg := WebSocketMessage{
		Type:      "status",
		SenderID:  int(userID),
		Content:   map[bool]string{true: "online", false: "offline"}[isOnline],
		Timestamp: time.Now(),
		Username:  username,
		Avatar:    avatar,
	}

	msgBytes, _ := json.Marshal(statusMsg)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	
	for _, client := range Clients {
		if client.UserID != int(userID) {
			select {
			case client.Send <- msgBytes:
			default:
				close(client.Send)
				delete(Clients, client.UserID)
			}
		}
	}
}

func GetOnlineUsers() []int {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	
	onlineUsers := make([]int, 0, len(Clients))
	for userID := range Clients {
		onlineUsers = append(onlineUsers, userID)
	}
	return onlineUsers
}