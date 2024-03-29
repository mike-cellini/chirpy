package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "sort"
    
    "github.com/mike-cellini/chirpy/internal/database"
    
    "github.com/go-chi/chi/v5"
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

    auth := r.Header.Get("Authorization")
    token := strings.TrimPrefix(auth, "Bearer ")
    id, err := validateAccessToken(token)
    if err != nil {
        w.WriteHeader(401)
        return
    }

    decoder := json.NewDecoder(r.Body)
    req := request{}
    err = decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    } else if len(req.Body) > maxChirpLen {
        respondWithError(w, 400, "Chirp is too long")
        return
    }

    c, err := ch.db.CreateChirp(id, req.Body)
    respondWithJSON(w, 201, c)
}

func (ch *chirpHandler) retrieve(w http.ResponseWriter, r *http.Request) {
    var err error
    authorIdStr := r.URL.Query().Get("author_id")
    sortOrder := r.URL.Query().Get("sort")

    authorId := 0
    if authorIdStr != "" {
        authorId, err = strconv.Atoi(authorIdStr)
        if err != nil {
            respondWithError(w, 400, fmt.Sprintf("%s is not a valid author id", authorIdStr))
        }
    }

    data, err := ch.db.GetChirps(authorId)
    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }

    if sortOrder == "desc" {
        sort.Slice(data, func (a, b int) bool { return data[a].Id > data[b].Id })
    } else {
        sort.Slice(data, func (a, b int) bool { return data[a].Id < data[b].Id })
    }
    respondWithJSON(w, 200, data)
}

func (ch *chirpHandler) retrieveById(w http.ResponseWriter, r *http.Request) {
    chirpID := chi.URLParam(r, "chirpid")
    id, err := strconv.Atoi(chirpID)
    if err != nil {
        respondWithError(w, 400, "Invalid Chirp ID")
        return
    }
    data, ok, err := ch.db.GetChirpById(id)
     if err != nil {
         respondWithError(w, 400, "Something went wrong")
         return
     } else if !ok {
         w.WriteHeader(404)
         return
     }
     respondWithJSON(w, 200, data)
}

func (ch *chirpHandler) delete(w http.ResponseWriter, r *http.Request) {
    auth := r.Header.Get("Authorization")
    token := strings.TrimPrefix(auth, "Bearer ")
    userId, err := validateAccessToken(token)
    if err != nil {
        w.WriteHeader(401)
        return
    }

    chirpID := chi.URLParam(r, "chirpid")
    id, err := strconv.Atoi(chirpID)
    if err != nil {
        respondWithError(w, 400, "Invalid Chirp ID")
        return
    }

    data, ok, err := ch.db.GetChirpById(id)
    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    } else if !ok {
        w.WriteHeader(404)
        return
    } else if data.AuthorId != userId {
        w.WriteHeader(403)
    }

    err = ch.db.DeleteChirp(id)
}
