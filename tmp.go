package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type chirpSaver struct {
	id int
}
func getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type chirpsResp struct {
		Id int `json:"id"`
		Body string `json:"body"`
	}

	chirps := []chirpsResp{}
	chirps = append(chirps,
		chirpsResp{
			Id: 0,
			Body: "This is a dummy chirp.",
		},
		chirpsResp{
			Id: 1,
			Body: "This is another dummy chirp.",
		},
	)
	dat, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		type errorBody struct {
			Error string `json:"error"`
		}
		dat, err := json.Marshal(errorBody{
			Error: err.Error(),
		})
		if err != nil {
		        log.Printf("Fudge there was an error: %s", err)
		}
		w.Write(dat)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}


