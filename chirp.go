package main

import (
    "encoding/json"
    "net/http"
    "strings"
    
    "github.com/mike-cellini/chirpy/internal/database"
)

type chirpHandler struct {
    db *database.DB
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

    type request struct {
        Body string `json:"body"`
    }

    type chirpValidation struct {
        CleanedBody string `json:"cleaned_body"`
    }

    decoder := json.NewDecoder(r.Body)
    req := request{}
    err := decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
    } else if len(req.Body) > maxChirpLen {
        respondWithError(w, 400, "Chirp is too long")
    }

    c, err := ch.db.CreateChirp(req.Body)
    respondWithJSON(w, 201, c)
}

func (ch *chirpHandler) retrieve(w http.ResponseWriter, r *http.Request) {
    data, err := ch.db.GetChirps()
     if err != nil {
         respondWithError(w, 400, "Something went wrong")
     }
     respondWithJSON(w, 200, data)
}
