package main

import (
    "github.com/go-chi/chi/v5"

    "encoding/json"
    "fmt"
    "log"
    "net/http"
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

func main() {
    const filepathRoot = "."
    const port = "8080"
    var apiCfg = apiConfig {}

    r := chi.NewRouter()
    rApi := chi.NewRouter()
    rAdmin := chi.NewRouter()

    rApi.Get("/healthz", readinessHandler)
    rApi.HandleFunc("/reset", apiCfg.resetMetricsHandler)
    rApi.HandleFunc("/validate_chirp", chirpHandler)
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

func chirpHandler(w http.ResponseWriter, r *http.Request) {
    const maxChirpLen = 160

    type chirp struct {
        Body string `json:"body"`
    }

    type serverError struct {
        Error string `json:"error"`
    }

    type chirpValidation struct {
        Valid bool `json:"valid"`
    }

    decoder := json.NewDecoder(r.Body)
    c := chirp{}
    err := decoder.Decode(&c)

    var errResponse serverError

    if err != nil {
        errResponse = serverError { Error: "Something went wrong" }
    } else if len(c.Body) > maxChirpLen {
        errResponse = serverError { Error: "Chirp is too long" }
    } else {
        respBody := chirpValidation { Valid: true }
        dat, err := json.Marshal(respBody)

        if err != nil {
            w.WriteHeader(500)
            log.Printf("Error marhsalling JSON: %s", err)
            return
        } else {
            w.WriteHeader(200)
            w.Header().Set("Content-Type", "application/json")
            w.Write(dat)
            return
        }
    }
    
    dat, err := json.Marshal(errResponse)

    if err != nil {
        w.WriteHeader(500)
        log.Printf("Error marshalling JSON: %s", err)
        return
    }
    
    w.WriteHeader(400)
    w.Header().Set("Content-Type", "application/json")
    w.Write(dat)
}
