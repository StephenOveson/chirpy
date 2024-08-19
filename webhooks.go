package main

import (
	"encoding/json"
	"net/http"

	"github.com/stephenoveson/chirpy/auth"
)

func (api *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to access this route")
		return
	}

	if apiKey != api.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unable to access this route")
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = api.db.UpgradeUser(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
