package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/stephenoveson/chirpy/auth"
	"github.com/stephenoveson/chirpy/database"
)

func (api *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortBy := r.URL.Query().Get("sort")
	id, err := strconv.Atoi(authorId)
	var chirps []database.Chirp
	if authorId == "" {
		chirps, err = api.db.GetChirps(0, sortBy)
	} else {
		chirps, err = api.db.GetChirps(id, sortBy)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to read chirps from database.")
		return
	}

	respondWithJson(w, http.StatusOK, chirps)
}

func (api *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")
	id, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to convert parameter to integer.")
		return
	}
	chirp, err := api.db.GetChirpById(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to read chirps from database.")
		return
	}

	respondWithJson(w, http.StatusOK, chirp)
}

func (api *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to create a chirp")
		return
	}
	userId, err := auth.ValidateJWT(token, api.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to create a chirp")
		return
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpBody{}
	err = decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleanString, err := validateChirp(chirp.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "A problem has occurred on the server")
		return
	}

	savedChirp, err := api.db.CreateChirp(cleanString, userIdInt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(w, http.StatusCreated, savedChirp)
}

func (api *apiConfig) handleDeleteChrips(w http.ResponseWriter, r *http.Request) {
	type response struct{}
	id := r.PathValue("chirpID")
	chirpId, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to convert parameter to integer.")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to create a chirp")
		return
	}

	userId, err := auth.ValidateJWT(token, api.secret)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Not authorized to create a chirp")
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "A problem has occurred on the server")
		return
	}

	err = api.db.DeleteChirpById(chirpId, userIdInt)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Chirp unable to be found or you are not the author")
		return
	}

	respondWithJson(w, http.StatusNoContent, response{})
}

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", errors.New("chirp is too long")
	}

	cleanString := cleanBody(body)
	return cleanString, nil
}

func cleanBody(s string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	stringSlice := strings.Split(s, " ")
	for i, word := range stringSlice {
		lowercase := strings.ToLower(word)
		if _, ok := badWords[lowercase]; ok {
			stringSlice[i] = "****"
		}
	}

	cleanedString := strings.Join(stringSlice, " ")

	return cleanedString
}
