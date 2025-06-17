package database
import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)
// CountLikesForPost counts the total number of likes for a specific post
func CountLikesForPost(db *sql.DB, postID int) (int, error) {
	var totalLikes int
	err := db.QueryRow("SELECT COUNT(*) FROM reactions WHERE post_id = ? AND type = 'like'", postID).Scan(&totalLikes)
	return totalLikes, err
}
// CountDislikesForPost counts the total number of dislikes for a specific post
func CountDislikesForPost(db *sql.DB, postID int) (int, error) {
	var totalDislikes int
	err := db.QueryRow("SELECT COUNT(*) FROM reactions WHERE post_id = ? AND type = 'dislike'", postID).Scan(&totalDislikes)
	return totalDislikes, err
}

// CountCommentsForPost counts the total number of comments for a specific post
func CountCommentsForPost(db *sql.DB, postID int) (int, error) {
	var totalComments int
	err := db.QueryRow("SELECT COUNT(*) FROM Comments WHERE post_id = ?", postID).Scan(&totalComments)
	return totalComments, err
}
// IncrementLikeCount increments the like count for a post within a transaction.
func IncrementLikeCount(db *sql.DB, postID int) error {
	_, err := db.Exec(`UPDATE Posts SET like_count = like_count + 1 WHERE id = ?`, postID)
	return err
}
// Increment the dislike count for a post
func IncrementDislikeCount(db *sql.DB, postID int) error {
	_, err := db.Exec(`UPDATE Posts SET dislike_count = dislike_count + 1 WHERE id = ?`, postID)
	return err
}
// Increment the comment count for a post
func IncrementCommentCount(db *sql.DB, postID int) error {
	_, err := db.Exec(`UPDATE Posts SET comment_count = comment_count + 1 WHERE id = ?`, postID)
	return err
}
// Decrement the like count for a post
func DecrementLikeCount(db *sql.DB, postID int) error {
	_, err := db.Exec(`UPDATE Posts SET like_count = like_count - 1 WHERE id = ?`, postID)
	return err
}
// Decrement the dislike count for a post
func DecrementDislikeCount(db *sql.DB, postID int) error {
	_, err := db.Exec(`UPDATE Posts SET dislike_count = dislike_count - 1 WHERE id = ?`, postID)
	return err
}