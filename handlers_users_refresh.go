package main

import (
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"	
	
)

type refreshResponse struct {
	RefreshToken string
}

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwt), nil },
	)
	
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

        //It's a refresh token
	if issuer == Access {
		respondWithError(w, http.StatusBadRequest, "Issuer not Refresh.")
		return
	}

	isvalid, err := cfg.DB.IsRevoked(tokenString)
	if err != nil || !isvalid {
		respondWithError(w, http.StatusBadRequest, "Token is revoked")
		return
		
	}

	respondWithJSON(w, http.StatusOK, refreshResponse{RefreshToken: tokenString})
}
