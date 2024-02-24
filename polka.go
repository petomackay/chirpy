package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type webhookParams struct {
	Event string `json:"event"`
	Data  struct {
		UserId int `json:"user_id"`
	} `json:"data"`
}

func (ac *apiConfig) handleWebhooks(w http.ResponseWriter, r *http.Request) {
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
