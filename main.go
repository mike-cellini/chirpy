package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/mike-cellini/chirpy/internal/database"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
    fileserverHits int
    jwtSecret string
    polkaKey string
}

var apiCfg apiConfig

func main() {
    godotenv.Load()
    const filepathRoot = "."
    const port = "8080"
    apiCfg = apiConfig {
        fileserverHits: 0,
        jwtSecret: os.Getenv("JWT_SECRET"),
        polkaKey: os.Getenv("POLKA_KEY"),
    }
    
    path, err := os.Executable()
    if err != nil {
        log.Fatal("FATAL: Could not get execution path")
    }

    db, err := database.NewDB(filepath.Dir(path))
    if err != nil {
        log.Fatal("FATAL: Could not retrieve or create DB")
    }

    chirpHandler := chirpHandler { db: db }
    userHandler := userHandler { db: db }
    polkaHandler := polkaHandler { db: db }

    r := chi.NewRouter()
    rApi := chi.NewRouter()
    rAdmin := chi.NewRouter()

    rApi.Get("/healthz", readinessHandler)
    rApi.HandleFunc("/reset", apiCfg.resetMetricsHandler)
    rApi.Post("/chirps", chirpHandler.create)
    rApi.Get("/chirps", chirpHandler.retrieve)
    rApi.Get("/chirps/{chirpid}", chirpHandler.retrieveById)
    rApi.Delete("/chirps/{chirpid}", chirpHandler.delete)
    rApi.Post("/users", userHandler.create)
    rApi.Post("/login", userHandler.authenticate)
    rApi.Put("/users", userHandler.update)
    rApi.Post("/refresh", userHandler.refresh)
    rApi.Post("/revoke", userHandler.revoke)
    rApi.Post("/polka/webhooks", polkaHandler.handleEvent)
    r.Mount("/api", rApi)
    
    handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
    rAdmin.Get("/metrics", apiCfg.metricsHandler)
    r.Mount("/admin", rAdmin)

    r.Handle("/app", apiCfg.middlewareMetricsInc(handler))
    r.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))

    corsMux := middlewareCors(r)
    server := &http.Server {
        Addr: ":" + port,
        Handler: corsMux,
    }

    log.Printf("Serving files from %s port: %s\n", filepathRoot, port)
    server.ListenAndServe()
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

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}
