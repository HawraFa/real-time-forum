package main
import (
	"database/sql"
	"log"
	"net/http"
	"real-time-forum/database"
	"real-time-forum/serve"
	_ "github.com/mattn/go-sqlite3"
)
func main() {
	// Database setup
	db, err := sql.Open("sqlite3", "./real-time-forum.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	// Create tables if they don't exist
	err = database.CreateTables(db)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	// Insert initial categories
	err = database.InsertInitialCategories(db)
	if err != nil {
		log.Printf("Warning: Failed to insert initial categories: %v", err)
	}
	// Serve static files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/pictures/", fs)
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// Serve index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
		http.NotFound(w, r)
	})
	// Setup API routes
	serve.SetupRoutes(db)
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}