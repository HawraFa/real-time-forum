package serve

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"real-time-forum/database"
	"real-time-forum/middleware"
	"strconv"
	"strings"
	"time"
)

type UserRegistration struct {
	Username  string `json:"username"` // Changed from nickname to username
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// LoginCredentials struct for login requests
type LoginCredentials struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// SetupRoutes configures all the routes for the application
func SetupRoutes(db *sql.DB) {
	// Handle registration
	http.HandleFunc("/api/register", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		var user UserRegistration
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Error decoding registration data: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid request body",
			})
			return
		}

		log.Printf("Attempting to register user: %s", user.Username)

		// Add user to database
		err := database.InsertUser(
			db,
			user.Username,
			user.Password,
			user.Email,
			"pictures/profile.png",
			user.Gender,
			user.Age,
			user.FirstName,
			user.LastName,
		)

		if err != nil {
			log.Printf("Registration failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Registration successful",
		})
	}))

	// Handle login
	http.HandleFunc("/api/login", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request to /api/login", r.Method)

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		var credentials LoginCredentials
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			log.Printf("Error decoding request body: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid request body",
			})
			return
		}

		log.Printf("Login attempt for identifier: %s", credentials.Identifier)

		user, err := database.GetUserByLogin(db, credentials.Identifier, credentials.Password)
		if err != nil {
			log.Printf("Login failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid credentials",
			})
			return
		}

		// Set user as online
		err = database.UpdateUserOnlineStatus(db, user.ID, true)
		if err != nil {
			log.Printf("Failed to update online status: %v", err)
			// Continue anyway, as login was successful
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message":  "Login successful",
			"userID":   user.ID,
			"username": user.Username,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Internal server error",
			})
			return
		}

		log.Printf("Login successful for user: %s", user.Username)
	}))

	// Handle logout
	http.HandleFunc("/api/logout", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		// Get user ID from request
		var request struct {
			UserID int `json:"userId"`
		}

		// Debug log the request body
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Printf("Logout request body: %s", string(body))

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("Error decoding logout request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid request body",
			})
			return
		}

		log.Printf("Attempting to logout user ID: %d", request.UserID)

		// Update user's online status
		err := database.UpdateUserOnlineStatus(db, request.UserID, false)
		if err != nil {
			log.Printf("Error updating user online status: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Logout failed: %v", err),
			})
			return
		}

		log.Printf("Successfully logged out user ID: %d", request.UserID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Logged out successfully",
		})
	}))

	// Get all categories
	http.HandleFunc("/api/categories", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		categories, err := database.QueryCategories(db)
		if err != nil {
			log.Printf("Error getting categories: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Failed to get categories",
			})
			return
		}

		// Debug logging
		log.Printf("Sending categories: %+v", categories)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(categories); err != nil {
			log.Printf("Error encoding categories: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Failed to encode categories",
			})
			return
		}
	}))

	// Get post count for a category
	http.HandleFunc("/api/categories/", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		// Extract category ID and endpoint type from URL
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		categoryID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if len(pathParts) == 5 && pathParts[4] == "post-count" {
			// Handle post count request
			count, err := database.GetPostCountByCategory(db, categoryID)
			if err != nil {
				log.Printf("Error getting post count: %v", err)
				http.Error(w, "Failed to get post count", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": count})
			return
		}

		if len(pathParts) == 5 && pathParts[4] == "posts" {
			// Handle posts request
			posts, err := database.GetPostsByCategory(db, categoryID)
			if err != nil {
				log.Printf("Error getting posts: %v", err)
				http.Error(w, "Failed to get posts", http.StatusInternalServerError)
				return
			}

			categoryName, err := database.GetCategoryName(db, categoryID)
			if err != nil {
				log.Printf("Error getting category name: %v", err)
				http.Error(w, "Failed to get category name", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"posts":        posts,
				"categoryName": categoryName,
			})
			return
		}

		http.Error(w, "Invalid endpoint", http.StatusNotFound)
	}))

	// Get user profile
	http.HandleFunc("/api/profile/", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request to %s", r.Method, r.URL.Path) // Add debug logging

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		// Extract user ID from URL
		pathParts := strings.Split(r.URL.Path, "/")
		log.Printf("Path parts: %v", pathParts) // Debug log

		if len(pathParts) < 4 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid URL",
			})
			return
		}

		userID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid user ID",
			})
			return
		}

		log.Printf("Fetching profile for user ID: %d", userID) // Debug log

		profile, err := database.GetUserProfile(db, userID)
		if err != nil {
			log.Printf("Error getting profile: %v", err) // Debug log
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}))

	// Update user profile
	http.HandleFunc("/api/profile/update", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Failed to parse form data",
			})
			return
		}

		// Get profile updates from form data
		var updates map[string]interface{}
		if err := json.Unmarshal([]byte(r.FormValue("data")), &updates); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid update data",
			})
			return
		}

		// Handle file upload if present
		file, handler, err := r.FormFile("profilePicture")
		if err == nil {
			defer file.Close()

			// Check file extension
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".png" {
				log.Printf("Invalid file type: %s", ext)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Only .jpg and .png files are allowed",
				})
				return
			}

			// Create pictures directory if it doesn't exist
			if err := os.MkdirAll("pictures", 0755); err != nil {
				log.Printf("Error creating pictures directory: %v", err)
				return
			}

			// Get next picture ID
			nextID, err := database.GetNextPictureID(db)
			if err != nil {
				log.Printf("Error getting next picture ID: %v", err)
				nextID = int(time.Now().UnixNano() % 1000000000)
			}

			// Generate filename with incremental ID
			filename := fmt.Sprintf("picture_%d%s", nextID, ext)
			filepath := fmt.Sprintf("pictures/%s", filename)

			// Save the file
			dst, err := os.Create(filepath)
			if err != nil {
				log.Printf("Error saving file: %v", err)
				return
			}
			defer dst.Close()

			// Copy the file
			if _, err := io.Copy(dst, file); err != nil {
				log.Printf("Error copying file: %v", err)
				return
			}

			// Add avatar path to updates
			updates["avatar"] = filepath
			log.Printf("Updated avatar path to: %s", filepath) // Debug log
		}

		userID := int(updates["userID"].(float64))
		delete(updates, "userID")

		err = database.UpdateUserProfile(db, userID, updates)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Profile updated successfully",
		})
	}))

	// Handle password change
	http.HandleFunc("/api/change-password", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Method not allowed",
			})
			return
		}

		var passwordChange struct {
			UserID          int    `json:"userID"`
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}

		if err := json.NewDecoder(r.Body).Decode(&passwordChange); err != nil {
			log.Printf("Error decoding password change request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid request body",
			})
			return
		}

		log.Printf("Attempting to change password for user %d", passwordChange.UserID)

		// Verify current password
		user, err := database.GetUserByID(db, passwordChange.UserID)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "User not found",
			})
			return
		}

		// Check if current password matches
		if !database.CheckPassword(user.Password, passwordChange.CurrentPassword) {
			log.Printf("Current password mismatch for user %d", passwordChange.UserID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Current password is incorrect",
			})
			return
		}

		// Update password
		err = database.UpdateUserPassword(db, passwordChange.UserID, passwordChange.NewPassword)
		if err != nil {
			log.Printf("Error updating password: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Failed to update password",
			})
			return
		}

		log.Printf("Successfully changed password for user %d", passwordChange.UserID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Password updated successfully",
		})
	}))

	// Handle getting all posts or posts by category
	http.HandleFunc("/api/posts", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			posts, err := database.QueryPosts(db, nil)
			if err != nil {
				log.Printf("Error getting posts: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Failed to get posts",
				})
				return
			}

			// Initialize empty array if posts is nil
			if posts == nil {
				posts = []database.Post{}
			}

			json.NewEncoder(w).Encode(posts)
			return
		}

		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")

			// Parse JSON request body
			var post struct {
				UserId     int    `json:"userId"`
				CategoryId int    `json:"categoryId"`
				Title      string `json:"title"`
				Content    string `json:"content"`
			}

			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				log.Printf("Error decoding request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Invalid request body",
				})
				return
			}

			log.Printf("Received post data: %+v", post)

			err := database.InsertPost(db, &post.UserId, &post.CategoryId, post.Title, post.Content, "")
			if err != nil {
				log.Printf("Error creating post: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Failed to create post",
				})
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Post created successfully",
			})
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	// Handle getting posts by category
	http.HandleFunc("/api/posts/category/", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract category ID from URL
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		categoryID, err := strconv.Atoi(parts[4])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		posts, err := database.GetPostsByCategory(db, categoryID)
		if err != nil {
			log.Printf("Error getting posts for category %d: %v", categoryID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}))
}
