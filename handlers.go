package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (ac *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpParams{}
	err := decoder.Decode(&chirp)
	if err != nil {
		handleError("Couldn't decode json", http.StatusInternalServerError, w)
		return
	}

	if len(chirp.Body) > 140 {
		handleError("Chirp is too long", 400, w)
		return
	}

	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	sanitized := re.ReplaceAllString(chirp.Body, "****")

	responseData, err := ac.db.CreateChirp(sanitized)
	if err != nil {
		handleError("Couldn't create a new chirp"+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(responseData, http.StatusCreated, w)
}

func (ac *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := ac.db.GetChirps()
	if err != nil {
		handleError(fmt.Sprintf("Couldn't retrieve chirps: %v", err), http.StatusInternalServerError, w)
		return
	}
	sendJson(chirps, http.StatusOK, w)
}

func (ac *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		handleError("Invalid ID format: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	chirp, err := ac.db.GetChirp(id)
	if err != nil {
		handleError(fmt.Sprintf("Couldn't retrieve chirp: %s", err), http.StatusNotFound, w)
		return
	}
	sendJson(chirp, http.StatusOK, w)

}

func (ac *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	type userBody struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	user := userBody{}
	if err := decoder.Decode(&user); err != nil {
		handleError("Couldn't decode user json: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	responseData, err := ac.db.CreateUser(user.Email)
	if err != nil {
		handleError("Couldn't create a new user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(responseData, http.StatusCreated, w)

}
