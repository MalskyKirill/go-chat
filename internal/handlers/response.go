package handlers

import (
	"encoding/json"
	"net/http"
)

func ReadJSON(r *http.Request, destination any) error {
	return json.NewDecoder(r.Body).Decode(destination)
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
