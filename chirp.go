package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    
    "github.com/mike-cellini/chirpy/internal/database"
)

type chirpHandler struct {
    db *DB
}

func cleanChirp(in string) string {
    const sep string = " "
    const sub string = "****"
    var badWords = [3]string { "kerfuffle", "sharbert", "fornax" }
    words := strings.Split(in, sep)
    for i, w := range words {
        for _, b := range badWords {
            if strings.ToLower(w) == b {
                words[i] = sub
            }
        }
    }
    return strings.Join(words, sep)
}

func (ch *chirpHandler) create(w http.ResponseWriter, r *http.Request) {
    const maxChirpLen = 160

    type newChirp struct {
        Body string `json:"body"`
    }

    type serverError struct {
        Error string `json:"error"`
    }

    type chirpValidation struct {
        CleanedBody string `json:"cleaned_body"`
    }

    decoder := json.NewDecoder(r.Body)
    c := newChirp{}
    err := decoder.Decode(&c)

    var errResponse serverError

    if err != nil {
        errResponse = serverError { Error: "Something went wrong" }
    } else if len(c.Body) > maxChirpLen {
        errResponse = serverError { Error: "Chirp is too long" }
    } else {
        respBody := chirpValidation { CleanedBody: cleanChirp(c.Body) }
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
