package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type loginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
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
	var eis int
	if eis = params.EIS; params.EIS == 0 {
		eis = 24 * 60 * 60
	}
	token, err := cfg.createJwt(id, eis)
	if err != nil {
		log.Print(cfg.jwt)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, loginResponse{Email: params.Email, ID: id, Token: token})
}

func (cfg *apiConfig) createJwt(id, eis int) (string, error) {
	idasstring := strconv.Itoa(id)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(eis))),
		Subject:   idasstring,
	})

	signedToken, err := token.SignedString([]byte(cfg.jwt))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}
