package serve

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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

		log.Printf("Attempting to register user: %+v", user) // Debug log

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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		categories, err := database.QueryCategories(db)
		if err != nil {
			log.Printf("Error fetching categories: %v", err)
			http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
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

	// Handle all post-related endpoints
	http.HandleFunc("/api/posts", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			// Get all posts
			posts, err := database.QueryPosts(db, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": fmt.Sprintf("Failed to get posts: %v", err),
				})
				return
			}
			json.NewEncoder(w).Encode(posts)

		case http.MethodPost:
			// Create new post
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Failed to parse form",
				})
				return
			}

			userID := r.Context().Value("currentUser").(int)
			title := r.FormValue("title")
			content := r.FormValue("content")
			categoryIDStr := r.FormValue("categoryId")

			if title == "" || content == "" || categoryIDStr == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Missing required fields",
				})
				return
			}

			categoryID, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Invalid category ID",
				})
				return
			}

			// Handle image upload
			var imagePath string
			file, handler, err := r.FormFile("image")
			if err == nil && file != nil {
				defer file.Close()

				// Create pictures directory if it doesn't exist
				if err := os.MkdirAll("pictures", 0755); err != nil {
					log.Printf("Error creating pictures directory: %v", err)
				}

				// Generate unique filename
				filename := fmt.Sprintf("picture_%d_%s", time.Now().UnixNano(), handler.Filename)
				imagePath = "pictures/" + filename

				// Create the file
				dst, err := os.Create(imagePath)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Failed to save image",
					})
					return
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Failed to save image",
					})
					return
				}
			}

			err = database.InsertPost(db, &userID, &categoryID, title, content, imagePath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": fmt.Sprintf("Failed to create post: %v", err),
				})
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Post created successfully",
			})
		}
	}))

	// Handle specific post operations (comments, reactions)
	http.HandleFunc("/api/posts/", middleware.EnableCORS(middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract path parts
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")

		// Handle root /api/posts/ endpoint
		if pathParts[0] == "" {
			switch r.Method {
			case http.MethodGet:
				// Get all posts
				posts, err := database.QueryPosts(db, nil)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"message": fmt.Sprintf("Failed to get posts: %v", err),
					})
					return
				}
				json.NewEncoder(w).Encode(posts)

			case http.MethodPost:
				// Create new post
				err := r.ParseMultipartForm(10 << 20)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Failed to parse form",
					})
					return
				}

				userID := r.Context().Value("currentUser").(int)
				title := r.FormValue("title")
				content := r.FormValue("content")
				categoryIDStr := r.FormValue("categoryId")

				if title == "" || content == "" || categoryIDStr == "" {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Missing required fields",
					})
					return
				}

				categoryID, err := strconv.Atoi(categoryIDStr)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Invalid category ID",
					})
					return
				}

				// Handle image upload
				var imagePath string
				file, handler, err := r.FormFile("image")
				if err == nil && file != nil {
					defer file.Close()

					// Create pictures directory if it doesn't exist
					if err := os.MkdirAll("pictures", 0755); err != nil {
						log.Printf("Error creating pictures directory: %v", err)
					}

					// Generate unique filename
					filename := fmt.Sprintf("picture_%d_%s", time.Now().UnixNano(), handler.Filename)
					imagePath = "pictures/" + filename

					// Create the file
					dst, err := os.Create(imagePath)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(map[string]string{
							"message": "Failed to save image",
						})
						return
					}
					defer dst.Close()

					if _, err := io.Copy(dst, file); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(map[string]string{
							"message": "Failed to save image",
						})
						return
					}
				}

				err = database.InsertPost(db, &userID, &categoryID, title, content, imagePath)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"message": fmt.Sprintf("Failed to create post: %v", err),
					})
					return
				}

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Post created successfully",
				})
				return
			}
			return
		}

		// Handle category filter
		if pathParts[0] == "category" && len(pathParts) > 1 {
			categoryID, err := strconv.Atoi(pathParts[1])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Invalid category ID",
				})
				return
			}

			posts, err := database.FilterPostsByCategory(db, []int{categoryID})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": fmt.Sprintf("Failed to get posts: %v", err),
				})
				return
			}

			json.NewEncoder(w).Encode(posts)
			return
		}

		// Handle specific post endpoints
		postID, err := strconv.Atoi(pathParts[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid post ID"})
			return
		}

		// Handle comments
		if len(pathParts) > 1 && pathParts[1] == "comments" {
			handleComments(w, r, db, postID)
			return
		}

		// Handle reactions
		if len(pathParts) > 1 && pathParts[1] == "reactions" {
			handleReactions(w, r, db, postID)
			return
		}

		// Get single post
		post, err := database.QueryPostDetails(db, postID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Failed to get post: %v", err),
			})
			return
		}

		json.NewEncoder(w).Encode(post)
	}, db)))

	// Handle user profile
	http.HandleFunc("/api/users/", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL
		userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid user ID",
			})
			return
		}

		// Get user profile
		profile, err := database.GetUserProfile(db, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Failed to get user profile: %v", err),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}))
}

func handleComments(w http.ResponseWriter, r *http.Request, db *sql.DB, postID int) {
	switch r.Method {
	case http.MethodGet:
		comments, err := database.GetPostComments(db, postID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Failed to get comments: %v", err),
			})
			return
		}
		json.NewEncoder(w).Encode(comments)

	case http.MethodPost:
		var comment struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid comment data",
			})
			return
		}

		userID := r.Context().Value("currentUser").(int)
		err := database.InsertComment(db, postID, userID, comment.Content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Failed to add comment: %v", err),
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Comment added successfully"})
	}
}

func handleReactions(w http.ResponseWriter, r *http.Request, db *sql.DB, postID int) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var reaction struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reaction); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid reaction data",
		})
		return
	}

	userID := r.Context().Value("currentUser").(int)
	err := database.HandleReaction(db, postID, userID, reaction.Type, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Failed to handle reaction: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reaction updated"})
}

func handleImageUpload(file io.Reader, handler *multipart.FileHeader) string {
	// Implement image upload logic here
	// This is a placeholder and should be replaced with the actual implementation
	return ""
}
