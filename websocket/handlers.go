package websocket

import (
    "net/http"
    "github.com/gorilla/websocket"
    "real-time-forum/session"
    "real-time-forum/database"
    "database/sql"
    "log"
    

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
		session, err := session.Store.Get(r, "forum-session")
		if err != nil {
			http.Error(w, "Session error", http.StatusUnauthorized)
			return
		}

		// Validate authentication
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int)
        log.Printf("User ID from session: %d", userID)
		if !ok || userID == 0 {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

        conn, _ := upgrader.Upgrade(w, r, nil)
        client := &Client{
            Conn:   conn,
            UserID: userID,
            Send:   make(chan []byte),
        }
        Clients[userID] = client

        // Mark user as online in DB
        database.UpdateUserStatus(db, int64(userID), true)
        BroadcastUserStatus(db, userID, true)

        go client.ReadPump()
        go client.WritePump()
    }
}