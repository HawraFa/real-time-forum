package database

import (
	"database/sql"
	"log"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash from password string
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ValidateUser checks if username/password combo is valid
func ValidateUser(db *sql.DB, username, password string) (int, bool) {
	var (
		id int
		storedPassword string
	)
	
	// Query for the stored password hash and user ID
	err := db.QueryRow("SELECT id, password FROM Users WHERE username = ?", username).Scan(&id, &storedPassword)
	if err != nil {
		log.Printf("Failed to find user with username %s: %v", username, err)
		return 0, false
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		log.Println("Invalid password")
		return 0, false
	}

	return id, true
}

// RegisterUser creates a new user in the database
func RegisterUser(db *sql.DB, username, email, password string, age int, gender string) error {
	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Insert the new user
	_, err = db.Exec(`
		INSERT INTO Users (username, email, password, age, gender)
		VALUES (?, ?, ?, ?, ?)`,
		username, email, hashedPassword, age, gender)
	
	return err
} 