package api

import (
	"encoding/json"
	"net/http"
)

// Response Envelope for standardized REST API JSON responses.
type ResponseEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONResponse sends a JSON response with status code and data payload.
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ResponseEnvelope{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	})
}

// JSONError sends a JSON response formatted error message.
func JSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ResponseEnvelope{
		Success: false,
		Error:   message,
	})
}
