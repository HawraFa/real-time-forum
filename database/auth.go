package database

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash from password string
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ValidateUser checks if username/email and password combo is valid
func ValidateUser(db *sql.DB, usernameOrEmail, password string) (int, bool) {
	var (
		id             int
		storedPassword string
	)

	// Query for the stored password hash and user ID using either username or email
	err := db.QueryRow("SELECT id, password FROM Users WHERE username = ? OR email = ?", usernameOrEmail, usernameOrEmail).Scan(&id, &storedPassword)
	if err != nil {
		log.Printf("Failed to find user with username/email %s: %v", usernameOrEmail, err)
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
func RegisterUser(db *sql.DB, username, first_name, last_name, email, password, avatar string, age int, gender string) error {
	// Validate password strength first
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Log the values being inserted
	log.Printf("Registering: username=%s, email=%s, avatar=%s, age=%d, gender=%s", username, email, avatar, age, gender)

	// Insert the new user
	result, err := db.Exec(`
    INSERT INTO Users (username, first_name, last_name, email, password, avatar, age, gender)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		username, first_name, last_name, email, hashedPassword, avatar, age, gender)

	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return err
	}

	// Log successful insertion
	id, _ := result.LastInsertId()
	log.Printf("Successfully inserted user with ID: %d", id)

	return nil
}

// func GetUserByID(db *sql.DB, userID string) (*User, error) {
//   var user User
//   err := db.QueryRow("SELECT id, username, first_name, last_name, email, avatar, gender, age FROM Users WHERE id = ?", userID).
//       Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Avatar, &user.Gender, &user.Age)

//   if err != nil {
//       return nil, err
//   }
//   return &user, nil
// }

func GetUserByID(db *sql.DB, userID int) (*User, error) {
	var user User
	var avatar sql.NullString

	err := db.QueryRow(`SELECT id, username, first_name, last_name, email, avatar, gender, age FROM Users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &avatar, &user.Gender, &user.Age)

	if err != nil {
		log.Printf("❌ Failed to get user by ID: %v", err)
		return nil, err
	}

	if avatar.Valid {
		user.Avatar = avatar.String
	} else {
		user.Avatar = "/static/images/profile.png"
	}

	return &user, nil
}

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
