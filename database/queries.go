package database
import (
	"database/sql" // Provides database-related functions (Open, Query, etc.)
	"fmt"          // For formatting strings
	"log"          // For logging errors and important information
	//"time"         // For time-related functions
	"strings" // For string manipulation
	_ "github.com/mattn/go-sqlite3" // Blank import for SQLite3 driver, needed to interact with SQLite databases
)
// queryUsers retrieves all users from the Users table and prints their id, username, and email
func QueryUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, username, email, avatar, gender, age FROM Users;")
	if err != nil {
		log.Printf("Failed to query Users: %v", err)
		return nil, err
	}
	//rows.Close(): This closes the rows object, freeing up any resources it is using. It's important to close the rows object after you're done using it to prevent resource leaks.
	defer rows.Close()
	var users []User
	//Looping Through the Query Results
	//rows.Next(): This advances to the next row of results returned by the query. It returns true as long as there are more rows to process.
	for rows.Next() {
		var user User
		/*rows.Scan(): This reads the data from the current row into the variables id, username, email..., It "scans" the row and assigns the column values to these variables. */
		//Each column from the SQL query corresponds to a field in the User struct.
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Avatar, &user.Gender, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
// QueryCategories retrieves all categories from the Categories table
func QueryCategories(db *sql.DB) ([]Category, error) {
	query := "SELECT id, name FROM Categories ORDER BY name ASC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying categories: %v", err)
	}
	defer rows.Close()
	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		categories = append(categories, category)
	}
	// Initialize empty array if no categories found
	if categories == nil {
		categories = []Category{}
	}
	return categories, nil
}
// QueryPosts retrieves all posts with user information
func QueryPosts(db *sql.DB, userID *int) ([]Post, error) {
	var query string
	var rows *sql.Rows
	var err error
	if userID != nil {
		query = `
			SELECT p.id, p.user_id, p.category_id, p.title, p.content, p.image,
				   p.timestamp, p.like_count, p.dislike_count, p.comment_count,
				   u.username, u.avatar
			FROM Posts p
			JOIN Users u ON p.user_id = u.id
			WHERE p.user_id = ?
			ORDER BY p.timestamp DESC`
		rows, err = db.Query(query, *userID)
	} else {
		query = `
			SELECT p.id, p.user_id, p.category_id, p.title, p.content, p.image,
				   p.timestamp, p.like_count, p.dislike_count, p.comment_count,
				   u.username, u.avatar
			FROM Posts p
			JOIN Users u ON p.user_id = u.id
			ORDER BY p.timestamp DESC`
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %v", err)
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.CategoryID,
			&post.Title,
			&post.Content,
			&post.Image,
			&post.CreatedAt,
			&post.LikesCount,
			&post.DislikesCount,
			&post.CommentsCount,
			&post.Username,
			&post.Avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %v", err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %v", err)
	}
	return posts, nil
}
// QueryComments retrieves comments from the Comments table, optionally filtered by post ID
func QueryComments(db *sql.DB, postID *int) ([]Comment, error) {
	var rows *sql.Rows
	var err error
	// If postID is provided, filter by post ID; otherwise, retrieve all comments
	if postID != nil {
		rows, err = db.Query("SELECT id, post_id, user_id, content, timestamp, like_count, dislike_count  FROM Comments WHERE post_id = ?", *postID)
	} else {
		rows, err = db.Query("SELECT id, post_id, user_id, content, like_count, dislike_count timestamp FROM Comments;")
	}
	if err != nil {
		log.Printf("Failed to query Comments: %v", err)
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.LikesCount, &comment.DislikesCount); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	// Check if there was an error during row iteration
	if err = rows.Err(); err != nil {
		return nil, err // Return if there was an iteration error
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
// GetPrivateMessages retrieves messages between two users with pagination
func GetPrivateMessages(db *sql.DB, user1ID, user2ID, limit, offset int) ([]PrivateMessage, error) {
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.content, m.created_at, m.is_read,
			   u.username, u.avatar
		FROM private_messages m
		JOIN Users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := db.Query(query, user1ID, user2ID, user2ID, user1ID, limit, offset)
	if err != nil {
		log.Printf("Failed to query private messages: %v", err)
		return nil, err
	}
	defer rows.Close()
	var messages []PrivateMessage
	for rows.Next() {
		var msg PrivateMessage
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Content, &msg.CreatedAt, &msg.IsRead,
			&msg.Sender.Username, &msg.Sender.Avatar,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
// GetUnreadMessageCount gets the count of unread messages for a user
func GetUnreadMessageCount(db *sql.DB, userID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM private_messages 
		WHERE receiver_id = ? AND is_read = false`
	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		log.Printf("Failed to get unread message count: %v", err)
		return 0, err
	}
	return count, nil
}
// GetChatUsers retrieves all users with whom the current user has exchanged messages
// Ordered by the most recent message
func GetChatUsers(db *sql.DB, currentUserID int) ([]User, error) {
	query := `
		SELECT DISTINCT u.id, u.username, u.avatar, u.is_online,
			(SELECT created_at 
			 FROM private_messages 
			 WHERE (sender_id = ? AND receiver_id = u.id) 
				OR (sender_id = u.id AND receiver_id = ?)
			 ORDER BY created_at DESC 
			 LIMIT 1) as last_message_time
		FROM Users u
		JOIN private_messages m ON (m.sender_id = u.id OR m.receiver_id = u.id)
		WHERE (m.sender_id = ? OR m.receiver_id = ?) 
			AND u.id != ?
		ORDER BY last_message_time DESC`
	rows, err := db.Query(query, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID)
	if err != nil {
		log.Printf("Failed to query chat users: %v", err)
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		var lastMessageTime sql.NullTime
		err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &user.IsOnline, &lastMessageTime)
		if err != nil {
			return nil, err
		}
		if lastMessageTime.Valid {
			user.LastSeen = lastMessageTime.Time
		}
		users = append(users, user)
	}
	return users, nil
}
// GetOnlineUsers retrieves all currently online users
func GetOnlineUsers(db *sql.DB) ([]User, error) {
	query := `
		SELECT id, username, avatar, is_online, 
			   COALESCE(last_seen, CURRENT_TIMESTAMP) as last_seen
		FROM Users 
		WHERE is_online = true 
		ORDER BY username`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query online users: %v", err)
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Avatar,
			&user.IsOnline, &user.LastSeen,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
// MarkMessagesAsRead marks all messages from a specific sender to a receiver as read
func MarkMessagesAsRead(db *sql.DB, senderID, receiverID int) error {
	query := `
		UPDATE private_messages 
		SET is_read = true 
		WHERE sender_id = ? AND receiver_id = ? AND is_read = false`
	result, err := db.Exec(query, senderID, receiverID)
	if err != nil {
		log.Printf("Failed to mark messages as read: %v", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("Marked %d messages as read", rowsAffected)
	return nil
}
// GetPostCountByCategory returns the number of posts in a specific category
func GetPostCountByCategory(db *sql.DB, categoryID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM Posts p
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?`
	var count int
	err := db.QueryRow(query, categoryID).Scan(&count)
	if err != nil {
		log.Printf("Error getting post count for category %d: %v", categoryID, err)
		return 0, err
	}
	return count, nil
}
// GetPostsByCategory retrieves all posts for a specific category
func GetPostsByCategory(db *sql.DB, categoryID int) ([]Post, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.timestamp,
			   p.like_count as likes, p.dislike_count as dislikes, p.comment_count as comments,
			   u.username, u.avatar
		FROM Posts p
		JOIN post_categories pc ON p.id = pc.post_id
		JOIN Users u ON p.user_id = u.id
		WHERE pc.category_id = ?
		ORDER BY p.timestamp DESC`
	rows, err := db.Query(query, categoryID)
	if err != nil {
		log.Printf("Error getting posts for category %d: %v", categoryID, err)
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt,
			&post.LikesCount, &post.DislikesCount, &post.CommentsCount,
			&post.Username, &post.Avatar,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
// GetCategoryName retrieves the name of a category by its ID
func GetCategoryName(db *sql.DB, categoryID int) (string, error) {
	query := `SELECT name FROM Categories WHERE id = ?`
	var name string
	err := db.QueryRow(query, categoryID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("category not found")
		}
		log.Printf("Error getting category name for ID %d: %v", categoryID, err)
		return "", err
	}
	return name, nil
}
func GetUserProfile(db *sql.DB, userID int) (*UserProfile, error) {
	query := `
		SELECT u.id, u.username, u.email, u.first_name, u.last_name, 
			   u.age, u.gender, u.avatar, u.is_online, u.last_seen, u.created_at,
			   (SELECT COUNT(*) FROM Posts WHERE user_id = u.id) as post_count
		FROM Users u
		WHERE u.id = ?`
	var profile UserProfile
	err := db.QueryRow(query, userID).Scan(
		&profile.ID,
		&profile.Username,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.Age,
		&profile.Gender,
		&profile.Avatar,
		&profile.IsOnline,
		&profile.LastSeen,
		&profile.JoinDate,
		&profile.PostCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &profile, nil
}
func UpdateUserOnlineStatus(db *sql.DB, userID int, isOnline bool) error {
	// First check if user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking user existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}
	query := `
		UPDATE Users 
		SET is_online = ?,
				last_seen = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	_, err = db.Exec(query, isOnline, userID)
	if err != nil {
		return fmt.Errorf("failed to update online status: %v", err)
	}
	return nil
}
// UpdateUserProfile updates a user's profile information
func UpdateUserProfile(db *sql.DB, userID int, updates map[string]interface{}) error {
	// Start building the query
	var setFields []string
	var values []interface{}
	// Map JSON fields to database columns
	fieldMapping := map[string]string{
		"username":  "username",
		"email":    "email",
		"firstName": "first_name",
		"lastName":  "last_name",
		"age":      "age",
		"avatar":   "avatar",
	}
	// Build the SET clause based on provided updates
	for jsonField, value := range updates {
		if dbField, ok := fieldMapping[jsonField]; ok {
			setFields = append(setFields, fmt.Sprintf("%s = ?", dbField))
			values = append(values, value)
			log.Printf("Updating field %s to %v", dbField, value) // Debug log
		}
	}
	if len(setFields) == 0 {
		return fmt.Errorf("no valid fields to update")
	}
	// Build the complete query
	query := fmt.Sprintf("UPDATE Users SET %s WHERE id = ?", strings.Join(setFields, ", "))
	values = append(values, userID)
	log.Printf("Update query: %s", query) // Debug log
	log.Printf("Update values: %v", values) // Debug log
	// Execute the update
	result, err := db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
// GetNextPictureID gets the next available picture ID
func GetNextPictureID(db *sql.DB) (int, error) {
	query := `
		SELECT COALESCE(MAX(CAST(SUBSTR(avatar, INSTR(avatar, 'picture_') + 8, 
			INSTR(avatar, '.') - INSTR(avatar, 'picture_') - 8) AS INTEGER)), 0) + 1
		FROM Users 
		WHERE avatar LIKE 'pictures/picture_%'
	`
	
	var nextID int
	err := db.QueryRow(query).Scan(&nextID)
	if err != nil {
		return 0, fmt.Errorf("error getting next picture ID: %v", err)
	}
	
	return nextID, nil
}
// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, userID int) (*User, error) {
	query := `SELECT id, username, password FROM Users WHERE id = ?`
	
	var user User
	err := db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	
	// Log for debugging
	log.Printf("Retrieved user %d with hashed password: %s", userID, user.Password)
	return &user, nil
}
// UpdateUserPassword updates a user's password
func UpdateUserPassword(db *sql.DB, userID int, newPassword string) error {
	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}
	// Log for debugging
	log.Printf("Updating password for user %d. Hashed password: %s", userID, hashedPassword)
	// Update the password in the database
	query := `UPDATE Users SET password = ? WHERE id = ?`
	result, err := db.Exec(query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("error updating password: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	log.Printf("Successfully updated password for user %d", userID)
	return nil
}
// GetUserByLogin retrieves a user by email/username and password
func GetUserByLogin(db *sql.DB, identifier, password string) (*User, error) {
	// First, get the user by email or username to retrieve their stored hashed password
	query := `SELECT id, username, password FROM Users WHERE email = ? OR username = ?`
	
	var user User
	err := db.QueryRow(query, identifier, identifier).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, err
	}
	// Now check if the provided password matches the stored hash
	if !CheckPassword(user.Password, password) {
		log.Printf("Password mismatch for user %s", user.Username)
		return nil, fmt.Errorf("invalid credentials")
	}
	log.Printf("Successful login for user %s", user.Username)
	return &user, nil
}
