package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(pwd string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Couldn't hash user password: %s", err)
		return "", err
	}
	return string(hashed), nil
}
