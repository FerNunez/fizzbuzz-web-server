package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// TODO: maybe fix this. If data is something that can't be decoded => Error
	_ = json.NewEncoder(w).Encode(APIResponse{
		Data:  data,
		Error: nil,
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}
