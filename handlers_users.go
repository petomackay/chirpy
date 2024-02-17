package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type userBody struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func (ac *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user := userBody{}
	if err := decoder.Decode(&user); err != nil {
		handleError("Couldn't decode user json: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	responseData, err := ac.db.CreateUser(user.Email, user.Password)
	if err != nil {
		handleError("Couldn't create a new user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(userResponse{Id: responseData.Id, Email: responseData.Email}, http.StatusCreated, w)
}

func (ac *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userBody := userBody{}
	if err := decoder.Decode(&userBody); err != nil {
		handleError("Couldn't decode user json: "+err.Error(), http.StatusBadRequest, w)
		return
	}
	user, err := ac.db.FindUserByEmail(userBody.Email)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, w)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userBody.Password)); err != nil {
		handleError(err.Error(), http.StatusUnauthorized, w)
		return
	}
	sendJson(userResponse{Id: user.Id, Email: user.Email}, http.StatusOK, w)
}
