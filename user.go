package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseModel struct {
		Id    int    `json:"id"`
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

	data, err := cfg.db.CreateUser(params.Email, params.Password)
	responseData := responseModel{
		Id:    data.Id,
		Email: data.Email,
	}

	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	respondWithJSON(w, http.StatusCreated, responseData)
}

func (cfg *apiConfig) handlerPutUsers(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type responseModel struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

	authToken, authtokenError := getTokenFromHeader(r)
	if authtokenError != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "params")
		return
	}

	token, err := jwt.ParseWithClaims(authToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}

		return []byte(cfg.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	subject, getSubjectError := token.Claims.GetSubject()
	if getSubjectError != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	id, convErr := strconv.Atoi(subject)
	if convErr != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	user, updateError := cfg.db.UpdateUser(id, params.Email, params.Password)
	if updateError != nil {
		respondWithError(w, http.StatusBadRequest, "")
		return
	}

	responseData := responseModel{
		Id:    user.Id,
		Email: user.Email,
	}

	respondWithJSON(w, http.StatusOK, responseData)
}
