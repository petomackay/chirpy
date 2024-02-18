package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userBody struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Expires  int    `json:"expires_in_seconds"`
}

type userResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}
type userLoginResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
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

	tokenExpirationTime := userBody.Expires
	defaultTokenExpiration := 24 * 60 * 60
	if tokenExpirationTime <= 0 || tokenExpirationTime > defaultTokenExpiration {
		tokenExpirationTime = defaultTokenExpiration
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

	currentTime := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(tokenExpirationTime) * time.Second)),
		Subject:   strconv.Itoa(user.Id),
	})
	ss, err := token.SignedString([]byte(ac.jwtSecret))
	if err != nil {
		handleError(err.Error(), http.StatusInternalServerError, w)
		return
	}

	sendJson(userLoginResponse{Id: user.Id, Email: user.Email, Token: ss}, http.StatusOK, w)
}

func (ac *apiConfig) putUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userBody := userBody{}
	if err := decoder.Decode(&userBody); err != nil {
		handleError("Couldn't decode use json: "+err.Error(), http.StatusBadRequest, w)
		return
	}
	tokenString, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer")
	if !found {
		handleError("Unknown auth type", http.StatusBadRequest, w)
		return
	}
	tokenString = strings.TrimSpace(tokenString)

	log.Println("The tokenString was: " + tokenString)

	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(ac.jwtSecret), nil
	})
	if err != nil {
		handleError("Unauthorized"+err.Error(), http.StatusUnauthorized, w)
		return
	}
	fmt.Printf("token.Claims: %v\n", token.Claims)
	idFromToken, _ := strconv.Atoi(claims.Subject)
	log.Println("The id was: " + claims.Subject)

	user, err := ac.db.FindUserById(idFromToken)
	if err != nil {
		handleError("Couldn't find user", http.StatusUnauthorized, w)
		return
	}
	user.Email = userBody.Email
	user.Password = userBody.Password

	if err := ac.db.UpdateUser(user); err != nil {
		handleError("Error when updating user: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	sendJson(userResponse{Id: user.Id, Email: user.Email}, http.StatusOK, w)
}
