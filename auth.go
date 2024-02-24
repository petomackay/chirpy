package main

import (
	"errors"
	"github.com/petomackay/chirpy/internal/database"
	"log"
)

func (ac *apiConfig) authenticateUserWithToken(tokenString string) (database.User, error) {
	if !isAccessToken(tokenString, []byte(ac.jwtSecret)) {
		return database.User{}, errors.New("Not a valid access token.")
	}

	userId, err := getIdFromToken(tokenString, []byte(ac.jwtSecret))
	if err != nil {
		log.Println("Couldn't get ID from token during token auth: " + err.Error())
		return database.User{}, err
	}

	user, err := ac.db.FindUserById(userId)
	if err != nil {
		log.Printf("Couldn't find used with id %d during token auth: %v\n", userId, err)
		return database.User{}, err
	}

	return user, nil
}
