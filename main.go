package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"real-time-forum/database"

	_ "github.com/mattn/go-sqlite3" // Import for SQLite driver
)

// type RegisterRequest struct {
// 	Username string `json:"username"`
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// 	Age      int    `json:"age"`
// 	Gender   string `json:"gender"`
// }

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
	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}
	fmt.Println("SQLite driver is installed, and the database connection is successful!")
	// Call the functions to set up the database
	database.CreateTables(db)
	// Query data for verification (optional)
	database.QueryUsers(db)
	database.QueryPosts(db, nil)
	postID := 1 // Assuming you already have a postID, e.g., 1
	database.QueryComments(db, &postID)
	database.QueryReactions(db)
	database.QueryCategories(db)

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve index.html for root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Register handler
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

		err := database.RegisterUser(db, req.Username, req.FirstName, req.LastName, req.Email, req.Password, req.Age, req.Gender)
		if err != nil {
			log.Printf("Registration error: %v", err)
			http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
	}))

	// Login handler
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

		// Send back user info
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": userID,
			"username": req.Username,
		})
	}))

	// Start the HTTP server
	log.Println("Server started at http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
