package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Access  string = "chirpy-access"
	Refresh        = "chirpy-refresh"
)

type loginResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	logCall(r)

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		EIS      int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	id, err := cfg.DB.AuthorizeUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := cfg.createJwt(id, Access)
	if err != nil {
		log.Print(cfg.jwt)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	refresh_token, err := cfg.createJwt(id, Refresh)
	if err != nil {
		log.Print(cfg.jwt)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = cfg.DB.SaveRefreshToken(refresh_token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
	respondWithJSON(w, http.StatusOK, loginResponse{
		Email:        params.Email,
		ID:           id,
		Token:        token,
		RefreshToken: refresh_token})
}

func (cfg *apiConfig) createJwt(id int, issuer string) (string, error) {
	idasstring := strconv.Itoa(id)
	expires := time.Hour
	if issuer == "chirpy-refresh" {
		expires = 60 * 24 * time.Hour
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
		Subject:   idasstring,
	})

	signedToken, err := token.SignedString([]byte(cfg.jwt))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}
