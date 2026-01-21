package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/natnael-alemayehu/chirpy/internal/auth"
	"github.com/natnael-alemayehu/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
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
		ID:        usr.ID,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	})
}
