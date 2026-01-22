package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/natnael-alemayehu/chirpy/internal/auth"
	"github.com/natnael-alemayehu/chirpy/internal/database"
)

type ChirpApp struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	var param parameters
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		respondWithError(w, http.StatusInternalServerError, "decoding param", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Expect's a Bearer token", err)
		return
	}

	uid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not valid", err)
		return
	}

	chripBody := validateChirpBody(w, param.Body)

	chrp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      chripBody,
		UserID:    uid,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Create chirp error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, ChirpApp{
		ID:        chrp.ID.String(),
		CreatedAt: chrp.CreatedAt,
		UpdatedAt: chrp.UpdatedAt,
		Body:      chrp.Body,
		UserID:    chrp.UserID.String(),
	})

}

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, r *http.Request) {
	chrps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "list chirps error", err)
	}

	authorID := uuid.Nil
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		authorID, err = uuid.Parse(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}
	}
	chirpApps := []ChirpApp{}
	for _, v := range chrps {
		if authorID != uuid.Nil && v.UserID != authorID {
			continue
		}
		chirpApps = append(chirpApps, ChirpApp{
			ID:        v.ID.String(),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Body:      v.Body,
			UserID:    v.UserID.String(),
		})
	}
	respondWithJSON(w, http.StatusOK, chirpApps)
}

func (cfg *apiConfig) handlerGetChirpsByID(w http.ResponseWriter, r *http.Request) {
	paramChirpID := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(paramChirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing chirp UUID", err)
		return
	}

	chrp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "chirp not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "error fetching chirp by id", err)
		return
	}

	respondWithJSON(w, http.StatusOK, ChirpApp{
		ID:        chrp.ID.String(),
		CreatedAt: chrp.CreatedAt,
		UpdatedAt: chrp.UpdatedAt,
		Body:      chrp.Body,
		UserID:    chrp.UserID.String(),
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HELPERS
// ============================================
func validateChirpBody(w http.ResponseWriter, body string) (clean string) {
	badwords := []string{"kerfuffle", "sharbert", "fornax"}

	final_output := getCleanedBody(body, badwords)

	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	return final_output
}

func getCleanedBody(body string, badwords []string) string {
	lst := strings.Fields(body)

	output := []string{}
	for _, v := range lst {
		word := strings.ToLower(v)
		if slices.Contains(badwords, word) {
			output = append(output, "****")
			continue
		}
		output = append(output, v)
	}

	return strings.Join(output, " ")
}
