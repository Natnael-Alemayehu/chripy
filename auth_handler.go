package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/natnael-alemayehu/chirpy/internal/auth"
	"github.com/natnael-alemayehu/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(usr.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation error", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh Token Error", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    usr.ID,
		ExpiresAt: time.Now().Add(1440 * time.Hour),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          usr.ID,
			CreatedAt:   usr.CreatedAt,
			UpdatedAt:   usr.UpdatedAt,
			Email:       usr.Email,
			IsChirpyRed: usr.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})

}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	reftoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Ref Token not found in Header", err)
	}

	refToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), reftoken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Ref Token not found", err)
	}

	if time.Now().After(refToken.ExpiresAt) {
		respondWithError(w, http.StatusBadRequest, "Refresh token expired", err)
		return
	}

	oneHour := 3600 * time.Second
	expirationTime := oneHour

	token, err := auth.MakeJWT(refToken.UserID, cfg.secret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation error", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	reftoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Ref Token not found", err)
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), reftoken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't revoke Token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
