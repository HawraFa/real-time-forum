package database

import (
	"database/sql"
	 "fmt"
	"log"
	// "strings"
	"net/http"
	"encoding/json"
)

var DB *sql.DB

func InsertPostAndReturnID(db *sql.DB, userID *int, categoryID *int, title, content string) (int64, error) {
	if categoryID == nil {
		var otherCategoryID int
		err := db.QueryRow("SELECT id FROM Categories WHERE name = ?", "Other").Scan(&otherCategoryID)
		if err != nil {
			log.Printf("Failed to retrieve 'Other' category ID: %v", err)
			return 0, err
		}
		categoryID = &otherCategoryID
	}

	insertPostSQL := `
		INSERT INTO Posts (user_id, category_id, title, content, like_count, dislike_count, comment_count)
		VALUES (?, ?, ?, ?, 0, 0, 0);`

	res, err := db.Exec(insertPostSQL, *userID, *categoryID, title, content)
	if err != nil {
		log.Printf("Failed to insert post: %v", err)
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return 0, err
	}

	fmt.Println("Post inserted successfully with ID:", postID)
	return postID, nil
}

func FilterPostsByMultipleCategories(db *sql.DB, categoryIDs []int) ([]Post, error) {
	seen := map[int]bool{}
	var result []Post

	// Fetch all categories once to match IDs later
	allCategories, err := QueryCategories(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}

	for _, cid := range categoryIDs {
		postIDs, err := QueryPostsByCategory(db, cid)
		if err != nil {
			return nil, err
		}
		for _, pid := range postIDs {
			if seen[pid] {
				continue
			}
			seen[pid] = true

			post, err := QueryPostDetails(db, pid)
			if err == nil {
				catIDs, _ := QueryPostCategories(db, pid)
				post.Categories = ConvertCategoryIDsToNames(allCategories, catIDs)
				result = append(result, post)
			}
		}
	}

	return result, nil
}

func ConvertCategoryIDsToNames(allCategories []Category, ids []int) []Category {
	var matched []Category
	for _, id := range ids {
		for _, cat := range allCategories {
			if cat.ID == id {
				matched = append(matched, cat)
				break
			}
		}
	}
	return matched
}


type CreatePostRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	AuthorID    int    `json:"author_id"`
	CategoryIDs []int  `json:"category_ids"`
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Invalid method"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CreatePostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if len(req.CategoryIDs) == 0 {
		http.Error(w, `{"error": "At least one category must be selected"}`, http.StatusBadRequest)
		return
	}

	// Insert post and return post ID
	postID, err := InsertPostAndReturnID(DB, &req.AuthorID, &req.CategoryIDs[0], req.Title, req.Content)
	if err != nil {
		fmt.Println("❌ Failed to insert post:", err)
		http.Error(w, `{"error": "Failed to create post"}`, http.StatusInternalServerError)
		return
	}

	// Link categories in post_categories
	for _, catID := range req.CategoryIDs {
		err := InsertPostCategory(DB, int(postID), catID)
		if err != nil {
			fmt.Printf("⚠️ Failed to link category %d: %v\n", catID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"post_id": postID,
	})
}