package main

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/natnael-alemayehu/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirpid format", err)
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token not formatted properly", err)
		return
	}

	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token not formatted properly", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "chirp not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "chirp error", err)
	}

	if chirp.UserID != userid {
		respondWithError(w, http.StatusForbidden, "You can not delete this chirp", err)
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), chirp.ID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirp delition not successful", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
