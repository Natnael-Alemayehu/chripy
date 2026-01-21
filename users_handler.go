package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/natnael-alemayehu/chirpy/internal/auth"
	"github.com/natnael-alemayehu/chirpy/internal/database"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var param parameter

	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding", err)
		return
	}

	hash, err := auth.HashPassword(param.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating hash", err)
		return
	}

	dbparam := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          param.Email,
		HashedPassword: hash,
	}

	usr, err := cfg.db.CreateUser(r.Context(), dbparam)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Create User error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:          usr.ID,
		CreatedAt:   usr.CreatedAt,
		UpdatedAt:   usr.UpdatedAt,
		Email:       usr.Email,
		IsChirpyRed: usr.IsChirpyRed,
	})
}

func (cfg *apiConfig) hanlderUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}

	var param parameter
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		respondWithError(w, http.StatusBadGateway, "password or email not correct", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token error", err)
		return
	}

	uid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	hash, err := auth.HashPassword(param.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password hashing failed", err)
		return
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             uid,
		Email:          param.Email,
		HashedPassword: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed Writing to database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          updatedUser.ID,
			CreatedAt:   updatedUser.CreatedAt,
			UpdatedAt:   updatedUser.UpdatedAt,
			Email:       updatedUser.Email,
			IsChirpyRed: updatedUser.IsChirpyRed,
		},
	})
}

func (cfg *apiConfig) handlerUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
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
