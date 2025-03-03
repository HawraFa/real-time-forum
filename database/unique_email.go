package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Function to check if email exists
func EmailExists(email string) bool {
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Printf("Failed to connect to the database: %v", err)
		return false
	}
	defer db.Close()
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check if email exists: %v", err)
		return false
	}
	return exists
}
