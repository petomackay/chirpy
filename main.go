package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const port = "8080"
	ac := apiConfig{}
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzCallback)
	apiRouter.Get("/reset", ac.resetCallback)
	apiRouter.Post("/validate_chirp", validateChirpHandler)
	apiRouter.Get("/chirps", getChirpsHandler)
	apiRouter.Post("/chirps", validateChirpHandler)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", ac.metricsCallback)
	r.Mount("/admin", adminRouter)

	fsHandler := ac.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	corsMux := middlewareCors(r)
	server := &http.Server {
		Addr: ":" + port,
		Handler: corsMux,
	}
	
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpParams{}
	err := decoder.Decode(&chirp)
	if err != nil {
		handleError("Couldn't decode json", http.StatusInternalServerError, w)
		return
	}

	if len(chirp.Body) > 140 {
		handleError("Chirp is too long", 400, w)
		return
	}

	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	sanitized := re.ReplaceAllString(chirp.Body, "****")

	type resp struct {
		Id int `json:"id"`
		Body string `json:"body"`
	}

	id := 1
	responseData := resp {
		Id: id,
		Body: sanitized,

	}
	sendJson(responseData, http.StatusOK, w)
}

func handleError(errorMsg string, statusCode int, w http.ResponseWriter) {
	type errorBody struct {
		Error string `json:"error"`
	}
	errorData := errorBody {
		Error: errorMsg,
	}
	sendJson(errorData, statusCode, w)
}

func sendJson(data interface{}, statusCode int, w http.ResponseWriter) {
	dat, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(dat)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

