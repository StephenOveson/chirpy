package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/stephenoveson/chirpy/auth"
	"github.com/stephenoveson/chirpy/database"
)

type userBody struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type userSuccess struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (api *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	user := userBody{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	password, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Issue hashing password")
		return
	}

	savedEmail, err := api.db.CreateUser(user.Email, password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(w, http.StatusCreated, savedEmail)
}

func (api *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Email        string `json:"email"`
		Id           int    `json:"id"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}
	decoder := json.NewDecoder(r.Body)
	user := userBody{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	u, err := api.db.GetUserByEmail(user.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = auth.CheckPasswordHash(user.Password, u.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid password")
		return
	}

	refreshToken, err := auth.GetRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	additionalTime := time.Duration(((60*60)*24)*60) * time.Second
	expiresAt := time.Now().Add(additionalTime)
	u.ExpiresAt = expiresAt
	u.RefreshToken = refreshToken

	dbUser, err := api.db.UpdateUser(u.Id, u)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}

	defaultExpiration := 60 * 60 * 1
	if user.ExpiresInSeconds == 0 {
		user.ExpiresInSeconds = defaultExpiration
	} else if user.ExpiresInSeconds > defaultExpiration {
		user.ExpiresInSeconds = defaultExpiration
	}

	token, err := auth.MakeJWT(u.Id, api.secret, time.Duration(user.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJson(w, http.StatusOK, response{
		Email:        dbUser.Email,
		Id:           dbUser.Id,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed,
	})
}

func (api *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	userId, err := auth.ValidateJWT(token, api.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Validate JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	user := database.User{}
	err = decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to properly handle password")
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	user.Password = hash
	user.Id = userIdInt

	u, err := api.db.UpdateUser(userIdInt, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJson(w, http.StatusOK, userSuccess{
		Id:          u.Id,
		Email:       u.Email,
		IsChirpyRed: u.IsChirpyRed,
	})
}

func (api *apiConfig) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := api.db.ConfirmUserToken(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := auth.MakeJWT(user.Id, api.secret, time.Duration(60*60)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJson(w, http.StatusOK, response{
		Token: token,
	})
}

func (api *apiConfig) handleTokenRevoke(w http.ResponseWriter, r *http.Request) {
	type response struct{}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to authenticate token")
		return
	}

	err = api.db.RevokeUserToken(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token to revoke")
		return
	}

	respondWithJson(w, http.StatusNoContent, response{})
}
