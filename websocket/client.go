package websocket

import (
	"real-time-forum/database"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     int
	Conn   *websocket.Conn
	Send   chan []byte
	UserID int
}

func (c *Client) ReadPump() {
	defer func() {

		delete(Clients, c.UserID)
		// Update user status to offline when they disconnect
		database.UpdateUserStatus(DB, int64(c.UserID), false)
		BroadcastUserStatus(DB, c.UserID, false)
		c.Conn.Close()
	}()

	// Update user status to online when they connect
	database.UpdateUserStatus(DB, int64(c.UserID), true)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		handleIncomingMessage(message)
	}
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
