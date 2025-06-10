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
	"real-time-forum/session"
	websocket "real-time-forum/websocket"

	_ "github.com/mattn/go-sqlite3"
	"strconv"
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
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

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
			http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
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

		userID, valid := database.ValidateUser(db, req.Username, req.Password)
		if !valid {
			http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
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
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query("SELECT id, username, first_name, last_name, email, avatar, gender, age FROM Users")
		if err != nil {
			http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []database.User
		for rows.Next() {
			var u database.User
			err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Avatar, &u.Gender, &u.Age)
			if err != nil {
				http.Error(w, `{"error": "Failed to parse users"}`, http.StatusInternalServerError)
				return
			}

			// Get online status
			var isOnline bool
			_ = db.QueryRow("SELECT is_online FROM user_status WHERE user_id = ?", u.ID).Scan(&isOnline)
			u.IsOnline = isOnline
			users = append(users, u)
		}

		json.NewEncoder(w).Encode(users)
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
		updatedUser, err := database.GetUserByID(db, id)  // NOT fmt.Sprintf

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

		rows, _ := db.Query("SELECT id, user_id, title FROM Posts")
		defer rows.Close()
		for rows.Next() {
			var id int
			var uid int
			var title string
			rows.Scan(&id, &uid, &title)
			fmt.Printf("\U0001F4E6 Post #%d by user_id=%d — Title: %s\n", id, uid, title)
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
	}))

	http.HandleFunc("/api/posts", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		posts, err := database.QueryPosts(db, nil)
	if err != nil {
		log.Printf("❌ QueryPosts failed: %v", err)
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

	http.HandleFunc("/api/comments", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			postIDStr := r.URL.Query().Get("post_id")
			log.Println("✅ Reached /api/comments handler")
			log.Println("Query param post_id =", postIDStr)

			if postIDStr == "" {
				http.Error(w, `{"error": "Missing post_id"}`, http.StatusBadRequest)
				return
			}

			var postID int
			fmt.Sscanf(postIDStr, "%d", &postID)

			comments, err := database.QueryComments(db, &postID)
			if err != nil {
				log.Printf("❌ Failed to fetch comments: %v", err)
				http.Error(w, `{"error": "Failed to fetch comments"}`, http.StatusInternalServerError)
				return
			}

			log.Printf("✅ Returning %d comments", len(comments))
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
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		otherUserID, _ := strconv.Atoi(userID)
		offset, _ := strconv.Atoi(offsetStr)

		messages, err := database.GetUserMessages(db, int64(currentUserID), int64(otherUserID), offset)
		if err != nil {
			http.Error(w, `{"error": "Failed to load messages"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}))

	// 🧾 NEW: Mark Messages as Read
	http.HandleFunc("/api/chat/mark-read", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		senderIDStr := r.URL.Query().Get("sender_id")
		if senderIDStr == "" {
			http.Error(w, `{"error": "Missing sender_id"}`, http.StatusBadRequest)
			return
		}

		senderID, err := strconv.Atoi(senderIDStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid sender_id"}`, http.StatusBadRequest)
			return
		}

		currentUserID, err := session.GetUserIDFromSession(r)
		if err != nil {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		err = database.MarkMessagesAsRead(db, int64(senderID), int64(currentUserID))
		if err != nil {
			http.Error(w, `{"error": "Failed to mark messages as read"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))

	log.Println("Server started at http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
