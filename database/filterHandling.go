package database

import (
	"database/sql"
	"encoding/json"
	//"forum/functions"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// HandleFilterPosts handles filtering posts based on query parameters
func HandleFilterPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Full Request URL:", r.URL.String())
		log.Println("Raw Query Parameters:", r.URL.Query())
		// Parse filter type and value from query parameters
		filterType := r.URL.Query().Get("type")
		filterValueStr := r.URL.Query().Get("value")
		log.Println("Filter Type:", filterType)
		log.Println("Filter Value String:", filterValueStr)
		// Ensure a filter value is provided
		if filterValueStr == "" {
			log.Printf("Missing filter value")
			//functions.ErrorHandler(w, r, http.StatusBadRequest)
			return
		}
		var filterValues []int
		var err error
		filterType = strings.TrimSpace(filterType)
		// Handle multiple values for the "category" type, single value for others
		if filterType == "category" {
			// Split the filter value string into a slice of strings
			filterValueStrings := strings.Split(filterValueStr, ",")
			for _, valueStr := range filterValueStrings {
				value, err := GetCategoryIDByName(db, strings.TrimSpace(valueStr))
				if err != nil {
					log.Printf("Invalid filter value")
					//functions.ErrorHandler(w, r, http.StatusBadRequest)

					return
				}
				filterValues = append(filterValues, value)
			}
		} else {
			// For other filter types, just convert the single value
			filterValue, err := strconv.Atoi(filterValueStr)
			if err != nil {
				log.Printf("Invalid filter value")
				//functions.ErrorHandler(w, r, http.StatusBadRequest)

				return
			}
			filterValues = []int{filterValue} // Wrap in a slice
		}
		// Call the FilterPostsByCriteria function
		userID := 0 // Replace this with the actual user ID you want to use
		posts, err := FilterPostsByCriteria(db, userID, filterType, filterValues)
		if err != nil {
			log.Printf("Failed to filter posts")
			//functions.ErrorHandler(w, r, http.StatusInternalServerError)

			return
		}
		// Return filtered posts as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(posts); err != nil {
			log.Printf("Failed to encode posts")
			//functions.ErrorHandler(w, r, http.StatusInternalServerError)

		}
	}
}