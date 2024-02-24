package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenExpirationTime int = 60 * 60
const refreshtokenExpirationTime int = 60 * 60 * 60

func issueAccessToken(userId string, jwtSecret []byte) (string, error) {
	currentTime := time.Now()

	log.Printf("Issuing a new token for user with ID: %s\n", userId)

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(accessTokenExpirationTime) * time.Second)),
		Subject:   userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func issueRefreshToken(userId string, jwtSecret []byte) (string, error) {
	currentTime := time.Now()

	log.Printf("Issuing a new refresh token for user with ID: %s\n", userId)

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(refreshtokenExpirationTime) * time.Second)),
		Subject:   userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func extractClaims(tokenString string, jwtSecret []byte) (jwt.Claims, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		log.Println("Couldn't parse JWT: " + err.Error())
		log.Println("the token was:'" + tokenString + "'")
		return nil, err
	}
	log.Printf("the id from claims was: %s\n", claims.Subject)
	return claims, nil
}

func isAccessToken(tokenString string, jwtSecret []byte) bool {
	claims, err := extractClaims(tokenString, jwtSecret)
	if err != nil {
		return false
	}
	issuer, err := claims.GetIssuer()
	if err != nil {
		return false
	}
	return issuer == "chirpy-access"
}

func getIdFromToken(tokenString string, jwtSecret []byte) (int, error) {
	claims, err := extractClaims(tokenString, jwtSecret)
	if err != nil {
		return 0, err
	}
	subjectClaim, err := claims.GetSubject()
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(subjectClaim)
	if err != nil {
		return 0, err
	}
	log.Printf("id: %d\n", id)
	return id, nil
}

func (ac *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	tokenString, found := extractTokenString(r)

	if found {
		log.Println("FOUND")
	}

	log.Println("The token string:" + tokenString)

	if !found || isAccessToken(tokenString, []byte(ac.jwtSecret)) || ac.db.IsTokenRevoked(tokenString) {
		handleError("Auth error", http.StatusUnauthorized, w)
		return
	}
	id, _ := getIdFromToken(tokenString, []byte(ac.jwtSecret))
	accessToken, err := issueAccessToken(strconv.Itoa(id), []byte(ac.jwtSecret))
	if err != nil {
		handleError("Oops", http.StatusInternalServerError, w)
		return
	}
	sendJson(userLoginResponse{Token: accessToken}, http.StatusOK, w)
}

func (ac *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, found := extractTokenString(r)

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(ac.jwtSecret), nil
	})
	if err != nil || !found {
		handleError("Unauthorized", http.StatusUnauthorized, w)
	}

	err = ac.db.RevokeToken(tokenString)
	if err != nil {
		handleError("Couldn't revoke token", http.StatusInternalServerError, w)
	}
	w.WriteHeader(http.StatusOK)
	return
}

func extractTokenString(r *http.Request) (string, bool) {
	tokenString, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer")
	return strings.TrimSpace(tokenString), found
}
