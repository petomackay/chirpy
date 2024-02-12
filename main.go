package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/petomackay/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func main() {
	const port = "8080"
	ac := apiConfig{}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzCallback)
	apiRouter.Get("/reset", ac.resetCallback)
	apiRouter.Get("/chirps", getChirpsHandler)
	apiRouter.Get("/chirps/{id}", getChirpByIDHandler)
	apiRouter.Post("/chirps", postChirpHandler)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", ac.metricsCallback)
	r.Mount("/admin", adminRouter)

	fsHandler := ac.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	corsMux := middlewareCors(r)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func postChirpHandler(w http.ResponseWriter, r *http.Request) {
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

	db, err := database.NewDB("database.json")
	if err != nil {
		handleError("Couldn't create a new database: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	responseData, err := db.CreateChirp(sanitized)
	if err != nil {
		handleError("Couldn't create a new chirp"+err.Error(), http.StatusInternalServerError, w)
		return
	}
	sendJson(responseData, http.StatusCreated, w)
}

func getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(fmt.Sprintf("Couldn't get a DB connection: %v", err), http.StatusInternalServerError, w)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		handleError(fmt.Sprintf("Couldn't retrieve chirps: %v", err), http.StatusInternalServerError, w)
		return
	}
	sendJson(chirps, http.StatusOK, w)
}

func getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		handleError("Invalid ID format: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(fmt.Sprintf("Couldn't get a DB connection: %s", err), http.StatusInternalServerError, w)
		return
	}
	chirp, err := db.GetChirp(id)
	if err != nil {
		handleError(fmt.Sprintf("Couldn't retrieve chirp: %s", err), http.StatusNotFound, w)
		return
	}
	sendJson(chirp, http.StatusOK, w)

}

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
