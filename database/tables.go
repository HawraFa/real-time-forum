package database

import (
	"database/sql"                  // Provides database-related functions (Open, Query, etc.)
	"fmt"                           // For formatted I/O (e.g., printing messages)
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
	"log"                           // For logging errors and important information
)

//var db *sql.DB
// createTables creates the necessary tables in the database if they don't already exist
/* This is a function declaration. It takes a parameter db which is a pointer to the SQLite database object (*sql.DB). This allows the function to interact with the database.*/
func CreateTables(db *sql.DB) {
	// The backticks ` allow multi-line strings in Go.
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS Users ( 
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        avatar TEXT,
        gender TEXT NOT NULL,  
        age INTEGER NOT NULL,
         is_online BOOLEAN DEFAULT FALSE   
    );`
	// Create the Categories table
	createCategoriesTable := `
    CREATE TABLE IF NOT EXISTS Categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );`
	// Create the Posts table
	createPostsTable := `
    CREATE TABLE IF NOT EXISTS Posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        category_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
         like_count INTEGER DEFAULT 0,
        dislike_count INTEGER DEFAULT 0,
        comment_count INTEGER DEFAULT 0,
        FOREIGN KEY (user_id) REFERENCES Users(id),
        FOREIGN KEY (category_id) REFERENCES Categories(id)
    );`
	// Create the Comments table
	createCommentsTable := `
    CREATE TABLE IF NOT EXISTS Comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        like_count INTEGER DEFAULT 0,
        dislike_count INTEGER DEFAULT 0,
        FOREIGN KEY (post_id) REFERENCES Posts(id),
        FOREIGN KEY (user_id) REFERENCES Users(id)
    );`
	// Create the Reactions table
	createReactionsTable := `
    CREATE TABLE IF NOT EXISTS Reactions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        post_id INTEGER,
        comment_id INTEGER,
        type TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES Users(id),
        FOREIGN KEY (post_id) REFERENCES Posts(id),
        FOREIGN KEY (comment_id) REFERENCES Comments(id),
         CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL),
         UNIQUE (user_id, post_id, comment_id)  
     );`
	// Create the post_categories table
	createPostsCategoriesTable := `
      CREATE TABLE IF NOT EXISTS post_categories (
         post_id INTEGER,                     
         category_id INTEGER,                 
         FOREIGN KEY (post_id) REFERENCES Posts(id),     
         FOREIGN KEY (category_id) REFERENCES Categories(id), 
         PRIMARY KEY (post_id, category_id)
     ); `

	// Private Messages table
	createPrivateMessagesTable := `
    CREATE TABLE IF NOT EXISTS PrivateMessages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sender_id INTEGER NOT NULL,
        receiver_id INTEGER NOT NULL,
        message_text TEXT NOT NULL,
        sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (sender_id) REFERENCES Users(id),
        FOREIGN KEY (receiver_id) REFERENCES Users(id)
    );`

	// Execute all the CREATE TABLE commands
	tables := []string{createUsersTable, createCategoriesTable, createPostsTable, createCommentsTable, createReactionsTable, createPostsCategoriesTable, createPrivateMessagesTable}
	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
	// Call insertInitialCategories to populate the Categories table with predefined categories
	InsertInitialCategories(db)
	// Print a message to indicate the table creation was successful
	fmt.Println("Tables created successfully!")
}
