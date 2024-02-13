package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func handleError(errorMsg string, statusCode int, w http.ResponseWriter) {
	type errorBody struct {
		Error string `json:"error"`
	}
	errorData := errorBody{
		Error: errorMsg,
	}
	sendJson(errorData, statusCode, w)
}

func sendJson(data interface{}, statusCode int, w http.ResponseWriter) {
	dat, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		handleError(fmt.Sprintf("Error marshalling JSON: %s", err), http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(dat)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}
