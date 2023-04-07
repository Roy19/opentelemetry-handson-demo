package utils

import (
	"encoding/json"
	"net/http"
)

func Respond(w http.ResponseWriter, statusCode int, data any) {
	w.WriteHeader(statusCode)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
