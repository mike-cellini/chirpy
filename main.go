package main

import (
    "github.com/go-chi/chi/v5"
    "fmt"
    "net/http"
    "log"
)

type apiConfig struct {
    fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Print("Incrementing file server hits")
        cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    log.Print("Retrieving file server metrics")
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
    log.Print("Reseting file server metrics")
    w.WriteHeader(200)
    cfg.fileserverHits = 0
}

func main() {
    const filepathRoot = "."
    const port = "8080"
    var apiCfg = apiConfig {}

    r := chi.NewRouter()
    
    //mux := http.NewServeMux()
    handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
    r.Handle("/app", apiCfg.middlewareMetricsInc(handler))
    r.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))
    r.Get("/healthz", readinessHandler)
    r.Get("/metrics", apiCfg.metricsHandler)
    r.HandleFunc("/reset", apiCfg.resetMetricsHandler)

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

