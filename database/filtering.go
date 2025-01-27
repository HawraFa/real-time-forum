package database
import (
	"database/sql"
	"fmt"
)
// FilterPostsByCategory retrieves posts associated with specific category IDs
func FilterPostsByCategory(db *sql.DB, categoryIDs []int) ([]Post, error) {
	var posts []Post
	// Iterate over the slice of category IDs to retrieve posts for each category
	for _, categoryID := range categoryIDs {
		postIDs, err := QueryPostsByCategory(db, categoryID)
		if err != nil {
			return nil, err
		}
		for _, postID := range postIDs {
			post, err := QueryPostDetails(db, postID)
			if err != nil {
				continue
			}
			posts = append(posts, post)
		}
	}
	// Optionally, you could remove duplicate posts if needed
	uniquePosts := removeDuplicatePosts(posts)
	return uniquePosts, nil
}
// removeDuplicatePosts removes duplicate posts from a slice
func removeDuplicatePosts(posts []Post) []Post {
	seen := make(map[int]bool) // Assuming Post has a unique ID of type int
	var uniquePosts []Post
	for _, post := range posts {
		if !seen[post.ID] { // Replace post.ID with the actual field representing the unique identifier
			seen[post.ID] = true
			uniquePosts = append(uniquePosts, post)
		}
	}
	return uniquePosts
}
// FilterCreatedPosts retrieves all posts created by a specific user
func FilterCreatedPosts(db *sql.DB, userID int) ([]Post, error) {
	posts, err := QueryPosts(db, &userID)
	if err != nil {
		return nil, err
	}
	return posts, nil
}
// FilterLikedPosts retrieves all posts liked by a specific user
func FilterLikedPosts(db *sql.DB, userID int) ([]Post, error) {
	querySQL := `SELECT post_id FROM Reactions WHERE user_id = ? AND type = 'Like';`
	rows, err := db.Query(querySQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			return nil, err
		}
		postIDs = append(postIDs, postID)
	}
	var likedPosts []Post
	for _, postID := range postIDs {
		post, err := QueryPostDetails(db, postID)
		if err != nil {
			continue
		}
		likedPosts = append(likedPosts, post)
	}
	return likedPosts, nil
}
// FilterPostsByCriteria filters posts based on a given criterion
func FilterPostsByCriteria(db *sql.DB, userID int, filterType string, filterValues []int) ([]Post, error) {
	switch filterType {
	case "category":
		return FilterPostsByCategory(db, filterValues)
	case "created":
		return FilterCreatedPosts(db, userID)
	case "liked":
		return FilterLikedPosts(db, userID)
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterType)
	}
}