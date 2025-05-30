package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/leonardoklaser/Chirpy/models"
)

type errorResponse struct {
	Error string `json:"error"`
}

func RespondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := errorResponse{Error: msg}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

func RespondWithJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding json response: %v", err)
		}
	}
}

func RespondWithListJson(w http.ResponseWriter, statusCode int, data []models.Chirp) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding json response: %v", err)
		}
	}
}