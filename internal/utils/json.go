package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type errJSON struct {
	Error string `json:"error"`
}

func RespondErrJSON(w http.ResponseWriter, errCode int, errMsg string) {
	if errCode > 499 {
		log.Println("Error message: ", errMsg)
	}
	RespondJSON(w, errCode, errJSON{
		Error: errMsg,
	})
}

func RespondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshalling failed on payload: %v", payload)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}
