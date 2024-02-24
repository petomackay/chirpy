package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type webhookParams struct {
	Event string `json:"event"`
	Data  struct {
		UserId int `json:"user_id"`
	} `json:"data"`
}

func (ac *apiConfig) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	if err := authenticatePolka(ac.polkaApiKey, r); err != nil {
		log.Printf("Couldn't authenticate polka webhook request: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := webhookParams{}
	if err := decoder.Decode(&params); err != nil {
		log.Println("Couldn't decode polka webhook body" + err.Error())
		handleError("Oops", http.StatusBadRequest, w)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusOK)
		return
	}

	user, err := ac.db.FindUserById(params.Data.UserId)
	if err != nil {
		log.Printf("Couldn't find user with ID %d in polka webhook: %v\n", params.Data.UserId, err)
		handleError("Not found", http.StatusNotFound, w)
		return
	}

	user.ChirpyRed = true
	if err := ac.db.UpdateUser(user); err != nil {
		log.Printf("Couldn't upgrade user id:%d to chirpy red in polka webhook: %v\n", params.Data.UserId, err)
		handleError("Couldn't upgrade user, soz", http.StatusInternalServerError, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func authenticatePolka(expectedApiKey string, r *http.Request) error {
	apiKey, found := strings.CutPrefix(r.Header.Get("Authorization"), "ApiKey")
	if !found {
		log.Println("Couldn't find an ApiKey in the polka webhook request.")
		return errors.New("Couldn't find ApiKey")
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey != expectedApiKey {
		log.Printf("Polka ApiKey doesn't match: %s\n", apiKey)
		return errors.New("Wrong ApiKey")
	}

	return nil
}
