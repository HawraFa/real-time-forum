package database
import (
	"database/sql"
)
// GetCategoryIDByName retrieves the category ID based on the category name
func GetCategoryIDByName(db *sql.DB, categoryName string) (int, error) {
	var id int
	query := "SELECT id FROM Categories WHERE name = ?"
	err := db.QueryRow(query, categoryName).Scan(&id)
	return id, err
}
