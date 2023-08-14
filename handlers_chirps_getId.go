package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) handlerChirpsGetId(w http.ResponseWriter, r *http.Request) {
	logCall(r)

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

	dat, err := json.Marshal(chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad JSON")
		return
	}
	w.Write(dat)
}
