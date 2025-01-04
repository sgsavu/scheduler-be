package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JSONError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]string{
		"error": message,
	}

	JSON(w, statusCode, errorResponse)
}

func JSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		fmt.Println("Failed to encode response")
	}
}
