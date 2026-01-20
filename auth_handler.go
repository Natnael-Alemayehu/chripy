package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/natnael-alemayehu/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	var param parameter
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		respondWithError(w, http.StatusBadRequest, "bad request", err)
	}

	usr, err := cfg.db.GetUserByEmail(r.Context(), param.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
			return
		}
		respondWithError(w, http.StatusUnauthorized, "Fetch by email error", err)
		return
	}

	match, err := auth.CheckPasswordHash(param.Password, usr.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "compare hash err", err)
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if param.ExpiresInSeconds == 0 {
		param.ExpiresInSeconds = 3600
	}

	expirationTime := time.Duration(param.ExpiresInSeconds) * time.Second

	token, err := auth.MakeJWT(usr.ID, cfg.secret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        usr.ID,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
		Token:     token,
	})

}
