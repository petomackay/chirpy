package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/petomackay/chirpy/internal/database"
)

type chirpParams struct {
	Body string `json:"body"`
}

func (ac *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, found := extractTokenString(r)
	if !found {
		handleError("Unauthorized", http.StatusUnauthorized, w)
		return
	}

	user, err := ac.authenticateUserWithToken(tokenString)
	if err != nil {
		handleError("Unauthorized", http.StatusUnauthorized, w)
	}

	userId := user.Id

	decoder := json.NewDecoder(r.Body)
	chirp := chirpParams{}
	err = decoder.Decode(&chirp)
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

	responseData, err := ac.db.CreateChirp(sanitized, userId)
	if err != nil {
		handleError("Couldn't create a new chirp"+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(responseData, http.StatusCreated, w)
}

func (ac *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if authorId == "" {
		chirps, err = ac.db.GetChirps()
		if err != nil {
			handleError(fmt.Sprintf("Couldn't retrieve chirps: %v", err), http.StatusInternalServerError, w)
			return
		}
	} else {
		authorIdInt, err := strconv.Atoi(authorId)

		if err != nil {
			handleError("BAD REQUEST", http.StatusBadRequest, w)
			return
		}
		chirps, err = ac.db.GetChirpsByAuthor(authorIdInt)
		if err != nil {
			handleError(fmt.Sprintf("Couldn't retrieve chirps: %v", err), http.StatusInternalServerError, w)
			return
		}
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

func (ac *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, found := extractTokenString(r)
	if !found {
		handleError("Unathorized", http.StatusUnauthorized, w)
		return
	}
	user, err := ac.authenticateUserWithToken(tokenString)
	if err != nil {
		handleError("Unathorized", http.StatusUnauthorized, w)
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		handleError("Invalid ID format: "+err.Error(), http.StatusBadRequest, w)
	}
	chirp, err := ac.db.GetChirp(id)
	if err != nil {
		handleError("Chirp not found", http.StatusNotFound, w)
		return
	}
	if chirp.UserId != user.Id {
		handleError("Forbidden", http.StatusForbidden, w)
		return
	}
	if err := ac.db.DeleteChirp(id); err != nil {
		handleError("Couldn't delete chirp", http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
