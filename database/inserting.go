package database
import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	"fmt"          
	"log"         
	//"time"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)
// InsertUser adds a new user to the database
func InsertUser(db *sql.DB, username, password, email, avatar, gender string, age int, firstName, lastName string) error {
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}
	query := `
		INSERT INTO Users (username, password, email, avatar, gender, age, first_name, last_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = db.Exec(query, username, string(hashedPassword), email, avatar, gender, age, firstName, lastName)
	if err != nil {
		return fmt.Errorf("error inserting user: %v", err)
	}
	
	return nil
}
// InsertInitialCategories adds predefined categories to the Categories table
func InsertInitialCategories(db *sql.DB) error {
	categories := []string{
		"Technology",
		"Gaming",
		"Sports",
		"Movies",
		"Music",
		"Books",
		"Food",
		"Travel",
		"Art",
		"Science",
		"Health",
		"Fashion",
		"Education",
		"Politics",
		"Other",
	}
	for _, category := range categories {
		query := `INSERT OR IGNORE INTO Categories (name) VALUES (?)`
		_, err := db.Exec(query, category)
		if err != nil {
			return fmt.Errorf("error inserting category %s: %v", category, err)
		}
	}
	return nil
}
// InsertPost inserts a new post into the Posts table
func InsertPost(db *sql.DB, userID *int, categoryID *int, title, content, image string) error {
	// If categoryID is nil, retrieve the ID for the "Other" category
	if categoryID == nil {
		var otherCategoryID int
		err := db.QueryRow("SELECT id FROM Categories WHERE name = ?", "Other").Scan(&otherCategoryID)
		if err != nil {
			return fmt.Errorf("failed to retrieve 'Other' category ID: %v", err)
		}
		categoryID = &otherCategoryID
	}
	query := `
		INSERT INTO Posts (user_id, category_id, title, content, image, like_count, dislike_count, comment_count) 
		VALUES (?, ?, ?, ?, ?, 0, 0, 0)
	`
	_, err := db.Exec(query, *userID, *categoryID, title, content, image)
	if err != nil {
		return fmt.Errorf("failed to insert post: %v", err)
	}
	return nil
}
func InsertComment(db *sql.DB, postID, userID int, content string) error {
	// Insert the comment
	insertCommentSQL := `INSERT INTO Comments (post_id, user_id, content) VALUES (?, ?, ?);`
	_, err := db.Exec(insertCommentSQL, postID, userID, content)
	if err != nil {
			log.Printf("Failed to insert comment: %v", err)
			return err
	}
	// Increment the comment count for the post
	err = IncrementCommentCount(db, postID) // Ensure IncrementCommentCount uses db
	if err != nil {
			log.Printf("Failed to increment comment count for post: %v", err)
			return err
	}
	fmt.Println("Comment inserted successfully, and comment count updated!")
	return nil
}
func InsertReaction(db *sql.DB, userID, postID, commentID int, reactionType string) error {
	// Insert or update the reaction
	insertReactionSQL := `INSERT INTO reactions (user_id, post_id, comment_id, type)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(user_id, post_id, comment_id)
	DO UPDATE SET type = excluded.type;`
	_, err := db.Exec(insertReactionSQL, userID, postID, commentID, reactionType)
	if err != nil {
			log.Printf("Failed to insert reaction: %v", err)
			return err
	}
	// Update like and dislike counts based on the reactionType (for posts)
	if postID != 0 {
			switch reactionType {
			case "Like":
					err = IncrementLikeCount(db, postID) // Pass db
					if err != nil {
							log.Printf("Failed to increment like count: %v", err)
							return err
					}
			case "Dislike":
					err = IncrementDislikeCount(db, postID) // Pass db
					if err != nil {
							log.Printf("Failed to increment dislike count: %v", err)
							return err
					}
			}
	} else if commentID != 0 {
			// Handle comment-specific reactions
			switch reactionType {
			case "Like":
					err = IncrementLikeCountForComment(db, commentID) // Pass db
			case "Dislike":
					err = IncrementDislikeCountForComment(db, commentID) // Pass db
			}
			if err != nil {
					log.Printf("Failed to increment comment reaction count: %v", err)
					return err
			}
	}
	fmt.Println("Reaction inserted successfully!")
	return nil
}
// InsertPostCategory inserts a relation between a post and a category
func InsertPostCategory(db *sql.DB, postID, categoryID int) error {
	insertSQL := `INSERT INTO post_categories (post_id, category_id) VALUES (?, ?);`
	_, err := db.Exec(insertSQL, postID, categoryID)
	if err != nil {
		log.Printf("Failed to insert post-category relation: %v", err)
		return err
	}
	fmt.Println("Post-Category relation inserted successfully!")
	return nil
}
// Private message-related insertions
func InsertPrivateMessage(db *sql.DB, senderId, receiverId int, content string) (int64, error) {
	query := `
		INSERT INTO private_messages (sender_id, receiver_id, content)
		VALUES (?, ?, ?)`
	
	result, err := db.Exec(query, senderId, receiverId, content)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
// Mark private message as read
func MarkMessageAsRead(db *sql.DB, messageId int) error {
	query := `
		UPDATE private_messages 
		SET is_read = true 
		WHERE id = ?`
	
	_, err := db.Exec(query, messageId)
	return err
}