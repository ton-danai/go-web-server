package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseData, err := cfg.db.CreateUser(params.Email)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	respondWithJSON(w, http.StatusCreated, responseData)
}
