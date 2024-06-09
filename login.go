package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type responseModel struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, errLogin := cfg.db.Login(params.Email, params.Password)
	if errLogin != nil {
		respondWithError(w, http.StatusUnauthorized, "Something went wrong!")
		return
	}

	now := time.Now()
	expTime := time.Hour * 24
	if params.ExpiresInSeconds != 0 {
		expTime = time.Second * time.Duration(params.ExpiresInSeconds)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now.UTC()),
		ExpiresAt: jwt.NewNumericDate(now.Add(expTime)),
		Subject:   strconv.Itoa(user.Id),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString([]byte(cfg.jwtSecret))

	if signErr != nil {
		log.Println(signErr)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	responseData := responseModel{
		Id:    user.Id,
		Email: user.Email,
		Token: tokenString,
	}

	respondWithJSON(w, http.StatusOK, responseData)
}
