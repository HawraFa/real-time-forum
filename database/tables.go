package database
import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	"fmt"          // For formatted I/O (e.g., printing messages)
	//"log"          // For logging errors and important information
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)
//var db *sql.DB
// createTables creates the necessary tables in the database if they don't already exist
/* This is a function declaration. It takes a parameter db which is a pointer to the SQLite database object (*sql.DB). This allows the function to interact with the database.*/
func CreateTables(db *sql.DB) error {
	// Create Users table with all required columns
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS Users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            first_name TEXT,
            last_name TEXT,
            avatar TEXT DEFAULT 'pictures/profile.png',
            gender TEXT,
            age INTEGER,
            is_online BOOLEAN DEFAULT FALSE,
            last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("error creating Users table: %v", err)
	}
	// Create the Categories table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS Categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );`)
	if err != nil {
		return fmt.Errorf("error creating Categories table: %v", err)
	}
	// Create the Posts table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS Posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        category_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        image TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        like_count INTEGER DEFAULT 0,
        dislike_count INTEGER DEFAULT 0,
        comment_count INTEGER DEFAULT 0,
        FOREIGN KEY (user_id) REFERENCES Users(id),
        FOREIGN KEY (category_id) REFERENCES Categories(id)
    )
`)
	if err != nil {
		return fmt.Errorf("error creating Posts table: %v", err)
	}
	// Create the Comments table
	_, err = db.Exec(`
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
    );`)
	if err != nil {
		return fmt.Errorf("error creating Comments table: %v", err)
	}
	// Create the Reactions table
	_, err = db.Exec(`
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
     );`)
	if err != nil {
		return fmt.Errorf("error creating Reactions table: %v", err)
	}
	// Create the post_categories table
	_, err = db.Exec(`
      CREATE TABLE IF NOT EXISTS post_categories (
         post_id INTEGER,                     
         category_id INTEGER,                 
         FOREIGN KEY (post_id) REFERENCES Posts(id),     
         FOREIGN KEY (category_id) REFERENCES Categories(id), 
         PRIMARY KEY (post_id, category_id)
     ); `)
	if err != nil {
		return fmt.Errorf("error creating post_categories table: %v", err)
	}
	fmt.Println("Tables created successfully!")
	return nil
}