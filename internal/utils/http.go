package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
