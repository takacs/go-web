package main

import (
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	logCall(r)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "No auth header")
	}
	headers := strings.Split(authHeader, " ")
	if len(headers) < 2 || headers[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "malformed authorization header")
	}

	tokenString := headers[1]

	err := cfg.DB.RevokeToken(tokenString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
