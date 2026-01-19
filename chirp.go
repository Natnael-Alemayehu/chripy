package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	badwords := []string{"kerfuffle", "sharbert", "fornax"}

	final_output := getCleanedBody(params.Body, badwords)

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid: final_output,
	})
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
