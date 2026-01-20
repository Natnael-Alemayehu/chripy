package main

import "net/http"

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if err := a.db.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Users Reset Failed", err)
		return
	}
	respondWithJSON(w, http.StatusOK, struct {
		Status string
	}{
		Status: "Reset Successful",
	})
}
