package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type refreshResponse struct {
	RefreshToken string
}

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	logCall(r)

	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	log.Printf("Tokenstring: %v", tokenString)

	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwt), nil },
	)

	if err != nil {
		log.Print("Failed while parsing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Print("Failed while getting issuer")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if issuer != Refresh {
		respondWithError(w, http.StatusBadRequest, "Issuer not Refresh.")
		return
	}

	isvalid, err := cfg.DB.IsRevoked(tokenString)
	if err != nil || !isvalid {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return

	}

	respondWithJSON(w, http.StatusOK, refreshResponse{RefreshToken: tokenString})
}
