package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/natnael-alemayehu/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized request: API key err", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized request: API key err", err)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to read body", err)
	}

	var param parameter
	if err := json.Unmarshal(data, &param); err != nil {
		respondWithError(w, http.StatusBadRequest, "json not formatted correctly", err)
		return
	}

	if param.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, struct{}{})
	}

	userID, err := uuid.Parse(param.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadGateway, "user_id not formatted correctly", err)
	}

	_, err = cfg.db.UpdateUserChirpyRed(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error saving data to db", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
