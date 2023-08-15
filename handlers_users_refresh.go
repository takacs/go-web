package main

import (
	"log"
	"net/http"
	"strings"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type refreshResponse struct {
	RefreshToken string `json:"token"`
}

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	logCall(r)

	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwt), nil },
	)

	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if issuer != Refresh {
		respondWithError(w, http.StatusBadRequest, "Issuer not Refresh.")
		return
	}

	isvalid, err := cfg.DB.IsRevoked(tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return

	} else if !isvalid {
		respondWithError(w, http.StatusUnauthorized, "Provided token not valid.")
		return
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}

	strid, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
	refreshedAccess, err := cfg.createJwt(strid, Access)
	if err != nil {
		log.Print(cfg.jwt)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, refreshResponse{RefreshToken: refreshedAccess})
}
