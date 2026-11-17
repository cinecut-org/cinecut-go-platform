package platform

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the JSON error every service returns.
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// WriteError writes an ErrorBody with the given status.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: code, Message: message})
}
