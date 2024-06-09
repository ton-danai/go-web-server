package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ton-danai/go-web-server/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type responseModel struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, errLogin := cfg.login(params.Email, params.Password)
	if errLogin != nil {
		respondWithError(w, http.StatusUnauthorized, "Something went wrong!")
		return
	}

	tokenString, tokenStringErr := cfg.generateAccessToken(user.Id)
	if tokenStringErr != nil {
		log.Println(tokenStringErr)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	refreshToken, refreshTokenErr := cfg.generateRefreshToken()
	if refreshTokenErr != nil {
		log.Println(refreshTokenErr)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!")
		return
	}

	cfg.db.StoreRefreshToken(user.Id, refreshToken)

	responseData := responseModel{
		Id:           user.Id,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, responseData)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type responseModel struct {
		Token string `json:"token"`
	}

	token, tokenError := getTokenFromHeader(r)
	if tokenError != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	id, isPass := cfg.db.VerifyRefreshToken(token)
	if !isPass {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	tokenString, accessTokenError := cfg.generateAccessToken(id)
	if accessTokenError != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}
	respondWithJSON(w, http.StatusOK, responseModel{Token: tokenString})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, headerError := getTokenFromHeader(r)

	if headerError != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	ok := cfg.db.RevokeToken(token)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) login(email, password string) (database.User, error) {
	user, found := cfg.db.FindUserByEmail(email)
	if !found {
		return database.User{}, errors.New("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return database.User{}, errors.New("password incorrect")
	}

	return user, nil
}

func (cfg *apiConfig) generateAccessToken(id int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now.UTC()),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 1)),
		Subject:   strconv.Itoa(id),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString([]byte(cfg.jwtSecret))

	return tokenString, signErr
}

func (cfg *apiConfig) generateRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)
	n, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	if n != len(randomBytes) {
		fmt.Println("Unexpected number of bytes read:", n)
		return "", fmt.Errorf("unexpected number of bytes read: %d", n)
	}

	refreshToken := hex.EncodeToString([]byte(randomBytes))
	return refreshToken, nil
}

func getTokenFromHeader(r *http.Request) (string, error) {
	type responseModel struct {
		Token string `json:"token"`
	}

	authString := r.Header.Get("Authorization")

	if authString == "" {
		return "", errors.New("not found")
	}

	arr := strings.Split(authString, " ")
	if arr[0] != "Bearer" {
		return "", errors.New("not found")
	}

	return arr[1], nil
}
