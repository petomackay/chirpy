package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type userBody struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	ChirpyRed bool   `json:"is_chirpy_red"`
}
type userLoginResponse struct {
	userResponse
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (ac *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user := userBody{}
	if err := decoder.Decode(&user); err != nil {
		handleError("Couldn't decode user json: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	hashedPwd, err := hashPassword(user.Password)
	if err != nil {
		handleError("Couldn't create a new user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	responseData, err := ac.db.CreateUser(user.Email, hashedPwd)
	if err != nil {
		handleError("Couldn't create a new user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(userResponse{Id: responseData.Id, Email: responseData.Email, ChirpyRed: responseData.ChirpyRed}, http.StatusCreated, w)
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

	userId := strconv.Itoa(user.Id)
	accessToken, err := issueAccessToken(userId, []byte(ac.jwtSecret))
	if err != nil {
		handleError(err.Error(), http.StatusInternalServerError, w)
		return
	}
	refreshToken, err := issueRefreshToken(userId, []byte(ac.jwtSecret))
	if err != nil {
		handleError(err.Error(), http.StatusInternalServerError, w)
		return
	}

	sendJson(userLoginResponse{userResponse: userResponse{Id: user.Id, Email: user.Email, ChirpyRed: user.ChirpyRed}, Token: accessToken, RefreshToken: refreshToken}, http.StatusOK, w)
}

func (ac *apiConfig) putUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userBody := userBody{}
	if err := decoder.Decode(&userBody); err != nil {
		handleError("Couldn't decode use json: "+err.Error(), http.StatusBadRequest, w)
		return
	}
	tokenString, found := extractTokenString(r)
	if !found {
		handleError("Unknown auth type", http.StatusBadRequest, w)
		return
	}

	user, err := ac.authenticateUserWithToken(tokenString)
	if err != nil {
		handleError("Couldn't find user", http.StatusUnauthorized, w)
		return
	}

	hashedPwd, err := hashPassword(userBody.Password)
	if err != nil {
		handleError("Couldn't create a new user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	user.Email = userBody.Email
	user.Password = hashedPwd

	if err := ac.db.UpdateUser(user); err != nil {
		handleError("Error when updating user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	sendJson(userResponse{Id: user.Id, Email: user.Email}, http.StatusOK, w)
}
