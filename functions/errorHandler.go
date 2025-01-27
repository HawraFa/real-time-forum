package functions

import (
	"html/template"
	"net/http"
)

// ErrorHandler handles error messages
func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	var fileName string
	// Choose the correct HTML file based on the error status
	switch status {
	case http.StatusNotFound:
		fileName = "templates/404.html"
	case http.StatusInternalServerError:
		fileName = "templates/500.html"
	case http.StatusBadRequest:
		fileName = "templates/400.html"
	case http.StatusMethodNotAllowed: //added this
		fileName="templates/405.html"	
	default:
		fileName = "templates/404.html" // Default to 404 if the status is unhandled
	}
	// Parse and serve the error page
	t, err := template.ParseFiles(fileName)
	if err != nil {
		http.ServeFile(w, r, fileName) // Serve file directly if parsing fails
		return
	}
	t.Execute(w, nil)
}