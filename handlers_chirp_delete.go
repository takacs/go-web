
package main

import (
	"net/http"
	"strconv"
	"strings"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	logCall(r)
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwt), nil },
	)

	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	strid, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	author_id, err := strconv.Atoi(strid)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	chirpid, err := strconv.Atoi(chi.URLParam(r, "chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID.")
		return
	}

	chirp, err := cfg.DB.GetChirpById(chirpid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No chirp found.")
		return
	}

	if chirp.AuthorID != author_id {
		respondWithError(w, http.StatusForbidden, "Can't delete tweet with different author")
		return
	}
	err = cfg.DB.DeleteChirp(chirpid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
