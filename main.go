package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const adminMetricsPage = `
<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>
`

func main() {
	const port = "8080"
	ac := apiConfig{}
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	
	api_router := chi.NewRouter()
	api_router.Get("/healthz", healthzCallback)
	api_router.Get("/reset", ac.resetCallback)
	api_router.Post("/validate_chirp", validateChirpHandler)
	r.Mount("/api", api_router)

	admin_router := chi.NewRouter()
	admin_router.Get("/metrics", ac.metricsCallback)
	r.Mount("/admin", admin_router)

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

// TODO: I'll write it out for sake of repetition when learning, but it needs a refactor
// probably generic handleEror and sendJson functions will do nicely
func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applicaton/json")
	type chirpParams struct {
		Body string `json:"body"`
	}

	type errorBody struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpParams{}
	err := decoder.Decode(&chirp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		error := errorBody {
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(error)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(dat)
		return
	}

	if len(chirp.Body) > 140 {
                 error := errorBody {
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(error)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	type resp struct {
		Valid bool `json:"valid"`
	}

	dat, err := json.Marshal(resp{
		Valid: true,
	})

	w.Write(dat)
	w.WriteHeader(http.StatusOK)

}

func healthzCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func (cfg *apiConfig) metricsCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(adminMetricsPage, cfg.fileserverHits)))
	return
}

func (cfg *apiConfig) resetCallback(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
