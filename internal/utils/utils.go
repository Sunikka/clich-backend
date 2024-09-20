package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	content, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshall json response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(content)
}

func GetUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return userID, err
	}
	return userID, nil
}
