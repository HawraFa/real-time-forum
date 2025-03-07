package handlers

import (
	"encoding/json"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// ... existing code ...

	response := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"age":      user.Age,
		"gender":   user.Gender,
		// Add any other fields you want to send back
	}
	json.NewEncoder(w).Encode(response)
} 