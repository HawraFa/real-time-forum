package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"real-time-forum/database"

	_ "github.com/mattn/go-sqlite3"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Connect to SQLite DB
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}
	fmt.Println("Database connected.")

	// Create tables and insert default data
	database.CreateTables(db)
	database.QueryUsers(db)
	database.QueryPosts(db, nil)
	database.QueryComments(db, nil)
	database.QueryReactions(db)
	database.QueryCategories(db)

	// Serve static files (JS, CSS, Images)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Register endpoint
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
			http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
	}))

	// Login endpoint
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

		userID, valid := database.ValidateUser(db, req.Username, req.Password)
		if !valid {
			http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByID(db, fmt.Sprintf("%d", userID))
		if err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(user)
	}))

	// 🔄 NEW: Edit Profile endpoint
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

		id := r.FormValue("id")
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		email := r.FormValue("email")
		age := r.FormValue("age")
		gender := r.FormValue("gender")

		// Optional: handle uploaded profile picture
		var avatarPath *string = nil

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
		

		// 🔄 NEW: Update user in DB
		err = database.UpdateUserProfile(db, id, firstName, lastName, email, age, gender, avatarPath)
		if err != nil {
			http.Error(w, `{"error": "Failed to update user"}`, http.StatusInternalServerError)
			return
		}

		updatedUser, err := database.GetUserByID(db, id)
		if err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(updatedUser)
	}))

	http.HandleFunc("/api/posts/create", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
	
		var req struct {
			Title      string `json:"title"`
			Content    string `json:"content"`
			CategoryID int    `json:"category_id"`
			AuthorID   int    `json:"author_id"`
		}
	
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}
	
		err := database.InsertPost(db, &req.AuthorID, &req.CategoryID, req.Title, req.Content)
		if err != nil {
			http.Error(w, `{"error": "Failed to create post"}`, http.StatusInternalServerError)
			return
		}
	
		json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
	}))

	http.HandleFunc("/api/posts", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	
		posts, err := database.QueryPosts(db, nil)
		if err != nil {
			log.Printf("QueryPosts failed: %v", err) // ✅ show in console
			http.Error(w, `{"error": "Failed to fetch posts"}`, http.StatusInternalServerError)
			return
		}
	
		if posts == nil {
			log.Println("QueryPosts returned nil — setting to empty list")
			posts = []database.Post{}
		}
	
		log.Printf("Returning %d posts", len(posts)) // ✅ show how many posts found
		json.NewEncoder(w).Encode(posts)
	}))
	
	
	

	// Start server
	log.Println("Server started at http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
