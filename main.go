package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"real-time-forum/database"
	"real-time-forum/session"
	websocket "real-time-forum/websocket"
	"time"

	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// generateUniqueFilename creates a unique filename for uploaded images
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	bytes := make([]byte, 16)
	rand.Read(bytes)
	uniqueID := hex.EncodeToString(bytes)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s%s", timestamp, uniqueID, ext)
}

func main() {
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	database.DB = db

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}
	fmt.Println("Database connected.")

	database.CreateTables(db)
	session.InitSessionStore("super-secret-key") // BEFORE starting routes
	database.QueryUsers(db)
	database.QueryPosts(db, nil)
	database.QueryComments(db, nil)
	database.QueryReactions(db)
	database.QueryCategories(db)

	websocket.DB = db // ⬅️ set DB connection for WebSocket
	http.HandleFunc("/ws", websocket.ServeWS(db))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/api/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req database.User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		err := database.RegisterUser(db, req.Username, req.FirstName, req.LastName, req.Email, req.Password, "", req.Age, req.Gender)
		if err != nil {
			log.Printf("Registration error: %v", err)
			// Check if it's a password validation error
			if err.Error() == "password must be at least 8 characters long" ||
				err.Error() == "password must contain at least one uppercase letter" ||
				err.Error() == "password must contain at least one lowercase letter" ||
				err.Error() == "password must contain at least one number" ||
				err.Error() == "password must contain at least one special character" {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "UNIQUE constraint failed: Users.username") {
				http.Error(w, `{"error": "Username is already taken"}`, http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "UNIQUE constraint failed: Users.email") {
				http.Error(w, `{"error": "Email is already registered"}`, http.StatusBadRequest)
			} else {
				http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
	}))

	http.HandleFunc("/api/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		userID, valid := database.ValidateUser(db, req.UsernameOrEmail, req.Password)
		if !valid {
			http.Error(w, `{"error": "Invalid username/email or password"}`, http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByID(db, userID)
		if err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusInternalServerError)
			return
		}

		sessionData, _ := session.Store.Get(r, "forum-session")
		sessionData.Values["authenticated"] = true
		sessionData.Values["user_id"] = user.ID
		sessionData.Save(r, w)

		json.NewEncoder(w).Encode(user)
	}))

	// Get all users endpoint
	http.HandleFunc("/api/users", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// Set content type first
		w.Header().Set("Content-Type", "application/json")

		// Only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Add pagination parameters (optional)
		limit := 100 // Default limit
		if l := r.URL.Query().Get("limit"); l != "" {
			if l, err := strconv.Atoi(l); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if o, err := strconv.Atoi(o); err == nil && o >= 0 {
				offset = o
			}
		}

		// Query database with pagination
		rows, err := db.Query(`
        SELECT id, username, first_name, last_name, avatar
        FROM Users 
        ORDER BY username ASC
        LIMIT ? OFFSET ?
    `, limit, offset)

		if err != nil {
			log.Printf("Database query error: %v", err)
			http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []database.User
		for rows.Next() {
			var u database.User
			if err := rows.Scan(
				&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Avatar,
			); err != nil {
				log.Printf("Row scan error: %v", err)
				continue // Skip problematic rows instead of failing entire request
			}

			// Get online status
			err := db.QueryRow(`
            SELECT is_online 
            FROM user_status 
            WHERE user_id = ?
        `, u.ID).Scan(&u.IsOnline)

			if err != nil && err != sql.ErrNoRows {
				log.Printf("Status query error: %v", err)
			}

			users = append(users, u)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Rows error: %v", err)
		}

		// Return empty array instead of null when no users
		if users == nil {
			users = []database.User{}
		}

		// Cache control headers
		// w.Header().Set("Cache-Control", "max-age=60") // Cache for 1 minute
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		json.NewEncoder(w).Encode(users)
	}))

	http.HandleFunc("/api/users/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract user ID from URL
		idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
			return
		}

		// Get user from database
		var user database.User
		err = db.QueryRow(`
			SELECT id, username, first_name, last_name, email, avatar, gender, age 
			FROM Users 
			WHERE id = ?
		`, userID).Scan(
			&user.ID, &user.Username, &user.FirstName, &user.LastName,
			&user.Email, &user.Avatar, &user.Gender, &user.Age,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			} else {
				log.Printf("Database error: %v", err)
				http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			}
			return
		}

		// Get online status from in-memory manager
		statusManager := websocket.GetStatusManager()
		if status, exists := statusManager.GetUser(int64(userID)); exists {
			user.IsOnline = status.IsOnline
		} else {
			user.IsOnline = false
		}

		// Cache control
		w.Header().Set("Cache-Control", "max-age=120") // Cache for 2 minutes

		json.NewEncoder(w).Encode(user)
	}))

	http.HandleFunc("/api/profile/update", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, `{"error": "Failed to parse form data"}`, http.StatusBadRequest)
			return
		}
		// Parse and convert string fields to proper types
		idStr := r.FormValue("id")
		ageStr := r.FormValue("age")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
			return
		}
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid age"}`, http.StatusBadRequest)
			return
		}
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		email := r.FormValue("email")
		gender := r.FormValue("gender")
		var avatarPath *string = nil
		fmt.Println("🧪 Form values 1:", idStr, firstName, lastName, email, ageStr, gender)
		// Handle optional profile picture upload
		file, handler, err := r.FormFile("profilePicture")
		if err == nil && handler != nil {
			defer file.Close()
			os.MkdirAll("static/images", os.ModePerm)
			filename := "static/images/" + handler.Filename
			f, err := os.Create(filename)
			if err == nil {
				defer f.Close()
				io.Copy(f, file)
				path := "/static/images/" + handler.Filename
				avatarPath = &path
			}
		}
		log.Println("🧾 Parsed FormData: id =", id, ", age =", age, ", email =", email)
		// Update user
		err = database.UpdateUserProfile(db, id, firstName, lastName, email, age, gender, avatarPath)
		if err != nil {
			log.Printf("❌ UpdateUserProfile failed: %v", err)
			http.Error(w, fmt.Sprintf(`{"error": "Failed to update user: %v"}`, err), http.StatusInternalServerError)
			return
		}

		// Retrieve updated user and return it
		// updatedUser, err := database.GetUserByID(db, fmt.Sprintf("%d", id))
		updatedUser, err := database.GetUserByID(db, id) // NOT fmt.Sprintf

		if err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updatedUser)
	}))

	http.HandleFunc("/api/categories", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		categories, err := database.QueryCategories(db)
		if err != nil {
			http.Error(w, `{"error":"Failed to load categories"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(categories)
	}))

	// API handler for creating a post
	http.HandleFunc("/api/posts/create", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Title       string `json:"title"`
			Content     string `json:"content"`
			AuthorID    int    `json:"author_id"`
			CategoryIDs []int  `json:"category_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		if len(req.CategoryIDs) == 0 {
			http.Error(w, `{"error": "At least one category must be selected"}`, http.StatusBadRequest)
			return
		}

		// Insert the post and get the ID
		postID, err := database.InsertPostAndReturnID(database.DB, &req.AuthorID, &req.CategoryIDs[0], req.Title, req.Content)
		if err != nil {
			log.Printf("Failed to insert post: %v", err)
			http.Error(w, `{"error": "Failed to create post"}`, http.StatusInternalServerError)
			return
		}

		// Link each category to the post
		for _, catID := range req.CategoryIDs {
			err := database.InsertPostCategory(database.DB, int(postID), catID)
			if err != nil {
				log.Printf("Failed to link post to category %d: %v", catID, err)
			}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"post_id": postID,
		})
	}))

	http.HandleFunc("/api/posts", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check if user_id is provided in the query
		var userID *int
		if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
			if id, err := strconv.Atoi(userIDStr); err == nil {
				userID = &id
			} else {
				http.Error(w, `{"error": "Invalid user_id"}`, http.StatusBadRequest)
				return
			}
		}

		// Pass userID to QueryPosts
		posts, err := database.QueryPosts(db, userID)
		if err != nil {
			log.Printf("QueryPosts failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Failed to fetch posts: %v", err),
			})
			return
		}

		if posts == nil {
			log.Println("QueryPosts returned nil — setting to empty list")
			posts = []database.Post{}
		}

		log.Printf("Returning %d posts", len(posts))
		json.NewEncoder(w).Encode(posts)
	}))

	http.HandleFunc("/api/posts/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid post ID"}`, http.StatusBadRequest)
			return
		}

		post, err := database.QueryPostDetails(database.DB, postID)
		if err != nil {
			log.Printf("Failed to fetch post: %v", err)
			http.Error(w, `{"error": "Post not found"}`, http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(post)
	}))

	http.HandleFunc("/api/react", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID    int    `json:"user_id"`
			PostID    int    `json:"post_id"`
			CommentID int    `json:"comment_id"`
			Type      string `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		err := database.InsertReaction(db, req.UserID, req.PostID, req.CommentID, req.Type)
		if err != nil {
			http.Error(w, `{"error": "Failed to save reaction"}`, http.StatusInternalServerError)
			return
		}

		likes, err := database.CountLikesForPost(db, req.PostID)
		if err != nil {
			http.Error(w, `{"error": "Failed to count likes"}`, http.StatusInternalServerError)
			return
		}

		dislikes, err := database.CountDislikesForPost(db, req.PostID)
		if err != nil {
			http.Error(w, `{"error": "Failed to count dislikes"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]int{
			"likes":    likes,
			"dislikes": dislikes,
		})
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// 🔁 NEW: Chat History API
	http.HandleFunc("/api/chat/history", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID := r.URL.Query().Get("user_id")
		offsetStr := r.URL.Query().Get("offset")

		currentUserID, err := session.GetUserIDFromSession(r)
		if err != nil {
			log.Println("❌ Unauthorized chat history request")
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		log.Printf("Current User ID: %d", currentUserID)

		otherUserID, err := strconv.Atoi(userID)
		if err != nil {
			log.Println("❌ Invalid user_id parameter:", userID)
			http.Error(w, `{"error": "Invalid user_id"}`, http.StatusBadRequest)
			return
		}

		offset, _ := strconv.Atoi(offsetStr)
		if offset < 0 {
			offset = 0
		}

		messages, err := database.GetUserMessages(db, int64(currentUserID), int64(otherUserID), offset)
		if err != nil {
			log.Println("❌ GetUserMessages failed:", err)
			http.Error(w, `{"error": "Failed to load messages"}`, http.StatusInternalServerError)
			return
		}

		if messages == nil {
			messages = []database.PrivateMessage{} // Ensure we never return null
		}

		json.NewEncoder(w).Encode(messages)
	}))

	http.HandleFunc("/api/chat/last-interactions", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		currentUserID, err := session.GetUserIDFromSession(r)
		if err != nil {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		interactions, err := database.GetUserChats(db, int64(currentUserID))
		if err != nil {
			http.Error(w, `{"error": "Failed to load interactions"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(interactions)
	}))

	http.HandleFunc("/api/posts/filter", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		categoriesParam := r.URL.Query().Get("categories")
		if categoriesParam == "" {
			http.Error(w, `{"error": "Missing categories param"}`, http.StatusBadRequest)
			return
		}
		parts := strings.Split(categoriesParam, ",")
		var categoryIDs []int
		for _, p := range parts {
			if id, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				categoryIDs = append(categoryIDs, id)
			}
		}
		posts, err := database.FilterPostsByMultipleCategories(db, categoryIDs)
		if err != nil {
			http.Error(w, `{"error": "Failed to filter posts"}`, http.StatusInternalServerError)
			return
		}

		if posts == nil {
			posts = []database.Post{}
		}

		json.NewEncoder(w).Encode(posts)
	}))

	http.HandleFunc("/api/comments", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			postIDStr := r.URL.Query().Get("post_id")
			log.Println("Reached /api/comments handler")
			log.Println("Query param post_id =", postIDStr)

			if postIDStr == "" {
				http.Error(w, `{"error": "Missing post_id"}`, http.StatusBadRequest)
				return
			}

			var postID int
			fmt.Sscanf(postIDStr, "%d", &postID)

			comments, err := database.QueryComments(db, &postID)
			if err != nil {
				log.Printf("Failed to fetch comments: %v", err)
				http.Error(w, `{"error": "Failed to fetch comments"}`, http.StatusInternalServerError)
				return
			}

			log.Printf("Returning %d comments", len(comments))
			json.NewEncoder(w).Encode(comments)
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				UserID  int    `json:"user_id"`
				PostID  int    `json:"post_id"`
				Content string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
				return
			}

			err := database.InsertComment(db, req.PostID, req.UserID, req.Content)
			if err != nil {
				http.Error(w, `{"error": "Failed to insert comment"}`, http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "Comment added"})
			return
		}

		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}))

	// Image upload endpoint
	http.HandleFunc("/api/upload-image", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, `{"error": "Failed to parse form"}`, http.StatusBadRequest)
			return
		}

		// Get the uploaded file
		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"error": "No image file provided"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Check file type
		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			http.Error(w, `{"error": "File must be an image"}`, http.StatusBadRequest)
			return
		}

		// Generate unique filename
		filename := generateUniqueFilename(header.Filename)
		filepath := fmt.Sprintf("static/images/%s", filename)

		// Create the file
		dst, err := os.Create(filepath)
		if err != nil {
			log.Printf("Failed to create file: %v", err)
			http.Error(w, `{"error": "Failed to save image"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Failed to copy file: %v", err)
			http.Error(w, `{"error": "Failed to save image"}`, http.StatusInternalServerError)
			return
		}

		// Return the image URL
		imageURL := fmt.Sprintf("/static/images/%s", filename)
		json.NewEncoder(w).Encode(map[string]string{
			"success":   "true",
			"image_url": imageURL,
		})
	}))

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
