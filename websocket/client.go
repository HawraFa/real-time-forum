package websocket

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	Clients   = make(map[int]*Client) // Map of user IDs to clients
	clientsMu sync.Mutex              // Mutex to protect Clients map
)

type Client struct {
	ID       int
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   int
	Username string
	Avatar   string
	TabID    string 
}

func (c *Client) ReadPump() {
	defer func() {
		// Remove client from map and set offline status
		clientsMu.Lock()
		delete(Clients, c.UserID)
		clientsMu.Unlock()
		
		// Update user status to offline
		statusManager := GetStatusManager()
		statusManager.SetOffline(int64(c.UserID))
		
		// Broadcast status change
		BroadcastUserStatus(int64(c.UserID), false, c.Username, c.Avatar)
		
		c.Conn.Close()
	}()

	// Add client to map
	clientsMu.Lock()
	Clients[c.UserID] = c
	clientsMu.Unlock()

	// Update user status to online
	statusManager := GetStatusManager()
	statusManager.SetOnline(int64(c.UserID), c.Username, c.Avatar)
	
	// Broadcast status change
	BroadcastUserStatus(int64(c.UserID), true, c.Username, c.Avatar)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		handleIncomingMessage(message)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Channel was closed
				return
			}
			
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}