package main

import (
    "fmt"
    "log"
    "net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Print("Incrementing file server hits")
        cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    log.Print("Retrieving file server metrics")
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(200)
    var body = fmt.Sprintln(
        "<html>\n", 
        "<body>", 
        "<h1>Welcome, Chirpy Admin</h1>",
        "<p>Chirpy has been visited %d times!</p>",
        "</body>",
        "</html>")
    w.Write([]byte(fmt.Sprintf(body, cfg.fileserverHits)))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
    log.Print("Reseting file server metrics")
    w.WriteHeader(200)
    cfg.fileserverHits = 0
}
