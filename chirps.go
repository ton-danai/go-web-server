package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerPostChirps(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const maxLen = 140
	if len(params.Body) > maxLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	responseData, err := cfg.db.CreateChirp(params.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	cfg.currentId++

	respondWithJSON(w, http.StatusCreated, responseData)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	responseData, err := cfg.db.GetChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	respondWithJSON(w, http.StatusOK, responseData)
}
