package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	ac := apiConfig{}
	mux.Handle("/app/", http.StripPrefix("/app/", ac.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", healthzCallback)
	mux.HandleFunc("/metrics", ac.metricsCallback)
	mux.HandleFunc("/reset", ac.resetCallback)

	corsMux := middlewareCors(mux)
	server := &http.Server {
		Addr: ":" + port,
		Handler: corsMux,
	}
	
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func healthzCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func (cfg *apiConfig) metricsCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
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
