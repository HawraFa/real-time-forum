package database

import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	"fmt"          // For formatted I/O (e.g., printing messages)
	"log"          // For logging errors and important information

	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)

func InsertUser(db *sql.DB, username, first_name, last_name, password, email, avatar, gender string, age int) error {
	//ensure the email is not duplicated in the DB.
	isUnique := EmailExists(email)

	if !isUnique {
		log.Println("Failed to insert user: email already exists")
		return fmt.Errorf("email already exists")
	}
	// SQL query to insert a new user with placeholders for username, password, and email
	/* VALUES (?, ?, ?): These placeholders (?) are used to avoid SQL injection. The actual values will be provided later. */
	insertUserSQL := `INSERT INTO Users (username, first_name, last_name, password, email, avatar, gender, age) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	// Execute the SQL query and provide values for the placeholders (username, password, email).
	// db.Exec: This executes the INSERT SQL command.
	_, err := db.Exec(insertUserSQL, username, first_name, last_name, password, email, avatar, gender, age)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return err
	}
	// Print a message to indicate the user was successfully inserted
	fmt.Println("User inserted successfully!")
	return nil // Return nil if successful
}

// insertInitialCategories adds predefined categories to the Categories table
func InsertInitialCategories(db *sql.DB) error {
	categories := []string{"General", "Technology", "Arts", "History", "Music", "Cooking", "Fashion", "Travel", "Politics", "Other"}
	for _, category := range categories {
		insertCategorySQL := `INSERT OR IGNORE INTO Categories (name) VALUES (?);`
		_, err := db.Exec(insertCategorySQL, category)
		if err != nil {
			log.Printf("Failed to insert category %s: %v", category, err)
			return err // Return the error if any category fails to be inserted
		}
	}
	fmt.Println("Initial categories inserted successfully!")
	return nil // Return nil if all categories are inserted successfully
}

// insertPost inserts a new post into the Posts table
func InsertPost(db *sql.DB, userID *int, categoryID *int, title, content string) error {
	// If categoryID is nil, retrieve the ID for the "Other" category
	if categoryID == nil {
		var otherCategoryID int
		err := db.QueryRow("SELECT id FROM Categories WHERE name = ?", "Other").Scan(&otherCategoryID)
		if err != nil {
			log.Printf("Failed to retrieve 'Other' category ID: %v", err)
			return err // Return the error if category lookup fails
		}
		categoryID = &otherCategoryID // Use the ID for the "Other" category
	}
	insertPostSQL := `INSERT INTO Posts (user_id, category_id, title, content, like_count, dislike_count, comment_count) VALUES (?, ?, ?, ?, 0, 0, 0);`
	_, err := db.Exec(insertPostSQL, *userID, *categoryID, title, content)
	if err != nil {
		log.Printf("Failed to insert post: %v", err)
		return err // Return the error if the post insertion fails
	}
	fmt.Println("Post inserted successfully!")
	return nil // Return nil if successful
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
	var prevType string
	err := db.QueryRow(
		`SELECT type FROM reactions WHERE user_id = ? AND post_id = ? AND comment_id = ?`,
		userID, postID, commentID,
	).Scan(&prevType)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking previous reaction: %v", err)
		return err
	}

	// If the same reaction already exists, remove it (toggle off)
	if prevType == reactionType {
		_, err := db.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id = ? AND comment_id = ?`, userID, postID, commentID)
		if err != nil {
			log.Printf("Failed to delete toggled reaction: %v", err)
			return err
		}
		if postID != 0 {
			if reactionType == "Like" {
				_ = DecrementLikeCount(db, postID)
			} else if reactionType == "Dislike" {
				_ = DecrementDislikeCount(db, postID)
			}
		}
		return nil // Stop here since it's toggle off
	}

	// If switching reactions (Like <-> Dislike)
	if prevType == "Like" && reactionType == "Dislike" {
		_ = DecrementLikeCount(db, postID)
		_ = IncrementDislikeCount(db, postID)
	} else if prevType == "Dislike" && reactionType == "Like" {
		_ = DecrementDislikeCount(db, postID)
		_ = IncrementLikeCount(db, postID)
	} else if prevType == "" {
		// First time reacting
		if postID != 0 {
			if reactionType == "Like" {
				_ = IncrementLikeCount(db, postID)
			} else if reactionType == "Dislike" {
				_ = IncrementDislikeCount(db, postID)
			}
		}
	}

	// Now insert or update the reaction
	insertSQL := `
		INSERT INTO reactions (user_id, post_id, comment_id, type)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, post_id, comment_id)
		DO UPDATE SET type = excluded.type;
	`
	_, err = db.Exec(insertSQL, userID, postID, commentID, reactionType)
	if err != nil {
		log.Printf("Failed to insert/update reaction: %v", err)
		return err
	}

	fmt.Println("✅ Reaction inserted/updated successfully!")
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

