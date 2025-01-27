package middleware

import (
	"log"
	"net/http"
)

// EnableCORS is middleware that adds CORS headers to the response
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS middleware: %s request to %s", r.Method, r.URL.Path) // Debug log

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5500")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Println("Handling OPTIONS request") // Debug log
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
} 