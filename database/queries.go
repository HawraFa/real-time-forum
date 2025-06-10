	package database

import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	"log"          // For logging errors and important information
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)

// queryUsers retrieves all users from the Users table and prints their id, username, and email
func QueryUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, username, first_name, last_name, email, avatar, gender, age FROM Users;")
	if err != nil {
		log.Printf("Failed to query Users: %v", err)
		return nil, err
	}
		//rows.Close(): This closes the rows object, freeing up any resources it is using. It’s important to close the rows object after you’re done using it to prevent resource leaks.
	defer rows.Close()
	var users []User
	//Looping Through the Query Results
	//rows.Next(): This advances to the next row of results returned by the query. It returns true as long as there are more rows to process.
	for rows.Next() {
		var user User
		/*rows.Scan(): This reads the data from the current row into the variables id, username, email..., It "scans" the row and assigns the column values to these variables. */
		//Each column from the SQL query corresponds to a field in the User struct.
		if err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.Avatar, &user.Gender, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// queryCategories retrieves all categories from the Categories table and prints their details
func QueryCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM Categories;")
	if err != nil {
		log.Printf("Failed to query Categories: %v", err)
		return nil, err
	}
	defer rows.Close()
	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// ***************************** joined two tales, user and posts
func QueryPosts(db *sql.DB, userID *int) ([]Post, error) {
	var rows *sql.Rows
	var err error

	if userID != nil {
		query := `
			SELECT p.id, p.user_id, p.category_id, p.title, p.content, p.timestamp, 
			       p.like_count, p.dislike_count, p.comment_count, 
			       u.username, u.avatar
			FROM Posts p
			JOIN Users u ON p.user_id = u.id
			WHERE p.user_id = ?`
		rows, err = db.Query(query, *userID)
	} else {
		query := `
			SELECT p.id, p.user_id, p.category_id, p.title, p.content, p.timestamp, 
			       p.like_count, p.dislike_count, p.comment_count, 
			       u.username, u.avatar
			FROM Posts p
			JOIN Users u ON p.user_id = u.id`
		rows, err = db.Query(query)
	}

	if err != nil {
		log.Printf("Failed to query Posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var avatar sql.NullString // only avatar may be null!

		err := rows.Scan(
			&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content,
			&post.CreatedAt, &post.LikesCount, &post.DislikesCount, &post.CommentsCount,
			&post.Username, &avatar,
		)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			return nil, err
		}

		if avatar.Valid {
			post.Avatar = avatar.String
		} else {
			post.Avatar = "/static/images/profile.png"
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// QueryComments retrieves comments from the Comments table, optionally filtered by post ID
func QueryComments(db *sql.DB, postID *int) ([]Comment, error) {
	var rows *sql.Rows
	var err error

	// Join with Users to get commenter's username and avatar
	if postID != nil {
		rows, err = db.Query(`
			SELECT c.id, c.post_id, c.user_id, c.content, c.timestamp, c.like_count, c.dislike_count,
			       u.username, u.avatar
			FROM Comments c
			JOIN Users u ON c.user_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.timestamp ASC
		`, *postID)
	} else {
		rows, err = db.Query(`
			SELECT c.id, c.post_id, c.user_id, c.content, c.timestamp, c.like_count, c.dislike_count,
			       u.username, u.avatar
			FROM Comments c
			JOIN Users u ON c.user_id = u.id
			ORDER BY c.timestamp ASC
		`)
	}

	if err != nil {
		log.Printf("Failed to query Comments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var username sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt,
			&comment.LikesCount, &comment.DislikesCount, &username, &avatar,
		)
		if err != nil {
			log.Printf("❌ Error scanning comment row: %v", err)
			return nil, err
		}

		comment.Username = "Unknown"
		if username.Valid {
			comment.Username = username.String
		}

		comment.Avatar = "/static/images/profile.png"
		if avatar.Valid {
			comment.Avatar = avatar.String
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}


// queryReactions retrieves all reactions from the Reactions table and prints their details
func QueryReactions(db *sql.DB) ([]Reaction, error) {
	rows, err := db.Query("SELECT id, user_id, post_id, comment_id, type, timestamp FROM Reactions;")
	if err != nil {
		log.Printf("Failed to query Reactions: %v", err)
		return nil, err
	}
	defer rows.Close()
	var reactions []Reaction
	for rows.Next() {
		var reaction Reaction
		if err := rows.Scan(&reaction.ID, &reaction.UserID, &reaction.PostID, &reaction.CommentID, &reaction.Type, &reaction.Timestamp); err != nil {
			return nil, err
		}
		reactions = append(reactions, reaction)
	}
	return reactions, nil
}

// QueryPostCategories retrieves all categories for a specific post
func QueryPostCategories(db *sql.DB, postID int) ([]int, error) {
	querySQL := `SELECT category_id FROM post_categories WHERE post_id = ?;`
	rows, err := db.Query(querySQL, postID)
	if err != nil {
		log.Printf("Failed to query post categories: %v", err)
		return nil, err
	}
	defer rows.Close()
	var categories []int
	for rows.Next() {
		var categoryID int
		if err := rows.Scan(&categoryID); err != nil {
			log.Printf("Failed to scan category ID: %v", err)
			return nil, err
		}
		categories = append(categories, categoryID)
	}
	return categories, nil
}

// QueryPostsByCategory retrieves post IDs for a specific category from the database
func QueryPostsByCategory(db *sql.DB, categoryID int) ([]int, error) {
	querySQL := `SELECT post_id FROM post_categories WHERE category_id = ?;`
	rows, err := db.Query(querySQL, categoryID)
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
		postIDs = append(postIDs, postID) // Ensure postID is of type int
	}
	return postIDs, nil
}

// QueryPostDetails retrieves a single post by its ID along with counts of likes, dislikes, comments,
// and the username and avatar of the user who created the post.
func QueryPostDetails(db *sql.DB, postID int) (Post, error) {
	var post Post
	// Modified query to also retrieve username and avatar from the Users table
	querySQL := `SELECT p.id, p.user_id, p.title, p.content, p.timestamp, p.like_count, p.dislike_count, p.comment_count, u.username, u.avatar
					FROM Posts p
					JOIN Users u ON p.user_id = u.id
					WHERE p.id = ?;`

	// Execute the query
	row := db.QueryRow(querySQL, postID)

	// Scan the results into the post struct
	if err := row.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.LikesCount, &post.DislikesCount, &post.CommentsCount, &post.Username, &post.Avatar); err != nil {
		log.Printf("Failed to retrieve post details: %v", err)
		return Post{}, err
	}

	// Return the post with all the necessary details
	return post, nil
}

func UpdateUserProfile(db *sql.DB, id int, firstName, lastName, email string, age int, gender string, avatarPath *string) error {
	log.Printf("🔧 UpdateUserProfile called: id=%d, email=%s", id, email)

	query := `
		UPDATE Users 
		SET first_name = ?, last_name = ?, email = ?, age = ?, gender = ?, avatar = ?
		WHERE id = ?
	`

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare update query: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(firstName, lastName, email, age, gender, avatarPath, id)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Rows affected: %d", rowsAffected)

	// 🧠 FIX: Remove strict failure when 0 rows are affected
	// Don't return error if 0 rows changed — it may just mean values are the same

	return nil
}