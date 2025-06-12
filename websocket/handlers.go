package websocket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"real-time-forum/session"
	"database/sql"
	"log"
    "real-time-forum/database"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use the global session store
		if session.Store == nil {
			http.Error(w, "Session store not initialized", http.StatusInternalServerError)
			return
		}

		// Get session
		sess, err := session.Store.Get(r, "forum-session")
		if err != nil {
			http.Error(w, "Session error", http.StatusUnauthorized)
			return
		}

		// Validate authentication
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := sess.Values["user_id"].(int)
		log.Printf("User ID from session: %d", userID)
		if !ok || userID == 0 {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		// Get user details
		user, err := database.GetUserByID(db, userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{
			Conn:     conn,
			UserID:   userID,
			Username: user.Username,
			Avatar:   user.Avatar,
			Send:     make(chan []byte, 256),
		}

		clientsMu.Lock()
		Clients[userID] = client
		clientsMu.Unlock()

		go client.WritePump()
		go client.ReadPump()
	}
}