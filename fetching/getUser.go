package fetching
import (
	"database/sql"
	"fmt"
	"real-time-forum/database"
)
// GetUser retrieves a user by their ID from the database.
func GetUser(db *sql.DB, userID string) (database.User, error) {
	var user database.User
	query := "SELECT id, username, email, avatar, gender, age FROM users WHERE id = ?"
	err := db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Avatar, &user.Gender, &user.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return database.User{}, fmt.Errorf("no user found with ID: %s", userID)
		}
		return database.User{}, err
	}
	return user, nil
}
// GetUserID retrieves the userID based on the username from the database.
func GetUserID(db *sql.DB, username string) (int, error) {
	var userID int
	query := "SELECT id FROM users WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no user found with username: %s", username)
		}
		return 0, err
	}
	return userID, nil
}