package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Function to fetch all posts from the database
func FetchAllPosts(db *sql.DB) ([]Post, error) {
	return QueryPosts(db, nil)
}

// Function to fetch (retrieve) comments by post ID from the database
func FetchCommentsByPostID(db *sql.DB, postID int) ([]Comment, error) {
	return QueryComments(db, &postID)
}