package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/mike-cellini/chirpy/internal/database"
)

type userHandler struct {
    db *database.DB
}

func (uh *userHandler) create(w http.ResponseWriter, r *http.Request) {
    type request struct {
        Email string `json:"email"`
    }

    decoder := json.NewDecoder(r.Body)
    req := request{}
    err := decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }

    u, err := uh.db.CreateUser(req.Email)
    respondWithJSON(w, 201, u)
}
