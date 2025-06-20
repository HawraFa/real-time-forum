package websocket

import (
	//"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	Clients   = make(map[int]map[*Client]bool) // userID → active connections
	clientsMu = sync.Mutex{}
)

type Client struct {
	Conn     *websocket.Conn
	UserID   int
	Username string
	Avatar   string
	Send     chan []byte
}

func (c *Client) ReadPump() {
	clientsMu.Lock()
	if Clients[c.UserID] == nil {
		Clients[c.UserID] = make(map[*Client]bool)
	}
	Clients[c.UserID][c] = true
	clientsMu.Unlock()

	// Immediately set online status and broadcast
	statusManager := GetStatusManager()
	statusManager.SetOnline(int64(c.UserID), c.Username, c.Avatar)
	BroadcastUserStatus(int64(c.UserID), true, c.Username, c.Avatar)
	BroadcastPresenceToNewUser(c.UserID, c)

	defer func() {
		clientsMu.Lock()
		if conns, ok := Clients[c.UserID]; ok {
			delete(conns, c)
			if len(conns) == 0 {
				delete(Clients, c.UserID)
				clientsMu.Unlock()

				statusManager := GetStatusManager()
				statusManager.SetOffline(int64(c.UserID))
				BroadcastUserStatus(int64(c.UserID), false, c.Username, c.Avatar)
			} else {
				clientsMu.Unlock()
			}
		} else {
			clientsMu.Unlock()
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Handle incoming messages
		handleIncomingMessage(message)
	}
}

func (c *Client) WritePump() {
	defer func() {
		clientsMu.Lock()
		if clients, ok := Clients[c.UserID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(Clients, c.UserID)
				// Immediately broadcast offline status
				statusManager := GetStatusManager()
				statusManager.SetOffline(int64(c.UserID))
				BroadcastUserStatus(int64(c.UserID), false, c.Username, c.Avatar)
			}
		}
		clientsMu.Unlock()

		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}
