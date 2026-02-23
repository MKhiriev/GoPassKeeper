package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteJSON serializes the given data to JSON and writes it to the HTTP response.
//
// It sets the "Content-Type" header to "application/json" and writes
// the provided HTTP status code before sending the response body.
//
// If marshaling fails, it responds with 500 Internal Server Error
// and returns a wrapped error.
//
// Parameters:
//
//	w          - the HTTP response writer to write the response to
//	data       - any value to be serialized as JSON (struct, map, slice, nil, etc.)
//	statusCode - HTTP status code to set in the response (e.g. http.StatusOK)
//
// Returns:
//
//	int   - number of bytes written to the response body
//	error - non-nil if JSON marshaling fails
//
// Example usage:
//
//	WriteJSON(w, map[string]string{"status": "ok"}, http.StatusOK)
//	WriteJSON(w, map[string]string{"error": "not found"}, http.StatusNotFound)
func WriteJSON(w http.ResponseWriter, data any, statusCode int) (int, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "error writing data to JSON", http.StatusInternalServerError)
		return 0, fmt.Errorf("error writing data to JSON: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return w.Write(jsonData)
}
