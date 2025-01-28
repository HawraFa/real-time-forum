package middleware

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
)

// AuthMiddleware wraps an http.HandlerFunc and adds user authentication
func AuthMiddleware(next http.HandlerFunc, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for certain paths
        if r.Method == "OPTIONS" || 
           r.URL.Path == "/api/login" || 
           r.URL.Path == "/api/register" ||
           (r.Method == "GET" && (strings.HasPrefix(r.URL.Path, "/api/posts") || 
                                strings.HasPrefix(r.URL.Path, "/api/categories"))) {
            next(w, r)
            return
        }

        // Get user ID from request header or cookie
        userIDStr := r.Header.Get("X-User-ID")
        if userIDStr == "" {
            cookie, err := r.Cookie("userID")
            if err == nil {
                userIDStr = cookie.Value
            }
        }

        if userIDStr == "" {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{
                "message": "Unauthorized",
            })
            return
        }

        userID, err := strconv.Atoi(userIDStr)
        if err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{
                "message": "Invalid user ID",
            })
            return
        }

        // Add user ID to request context
        ctx := context.WithValue(r.Context(), "currentUser", userID)
        next(w, r.WithContext(ctx))
    }
} 