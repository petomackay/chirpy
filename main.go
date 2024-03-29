package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/petomackay/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	polkaApiKey    string
	db             *database.DB
}

func main() {
	godotenv.Load()

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		os.Remove("database.json")
	}

	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	ac := apiConfig{
		fileserverHits: 0,
		jwtSecret:      os.Getenv("JWT_SECRET"),
		polkaApiKey:    os.Getenv("POLKA_API_KEY"),
		db:             db,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzCallback)
	apiRouter.Get("/reset", ac.resetCallback)
	apiRouter.Get("/chirps", ac.getChirpsHandler)
	apiRouter.Get("/chirps/{id}", ac.getChirpByIDHandler)
	apiRouter.Post("/chirps", ac.postChirpHandler)
	apiRouter.Post("/users", ac.postUsersHandler)
	apiRouter.Post("/login", ac.userLoginHandler)
	apiRouter.Put("/users", ac.putUsersHandler)
	apiRouter.Post("/refresh", ac.handleRefresh)
	apiRouter.Post("/revoke", ac.handleRevoke)
	apiRouter.Delete("/chirps/{id}", ac.deleteChirpHandler)

	polkaRouter := chi.NewRouter()
	polkaRouter.Post("/webhooks", ac.handleWebhooks)
	apiRouter.Mount("/polka", polkaRouter)
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
