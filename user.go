package main

import (
	"encoding/json"

	"net/http"

	"github.com/mike-cellini/chirpy/internal/database"

	"golang.org/x/crypto/bcrypt"
)

type userHandler struct {
    db *database.DB
}

type request struct {
    Email string `json:"email"`
    Password string `json:"password"`
}

type response struct {
    Email string `json:"email"`
    Id int `json:"id"`
}

func (uh *userHandler) create(w http.ResponseWriter, r *http.Request) {

    decoder := json.NewDecoder(r.Body)
    req := request{}
    err := decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }

    pHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        respondWithError(w, 400, "Unable to hash password")
        return 
    }

    u, err := uh.db.CreateUser(req.Email, string(pHash))
    res := response {
        Id: u.Id,
        Email: u.Email,
    }
    respondWithJSON(w, 201, res)
}

func (uh *userHandler) authenticate(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    req := request{}
    err := decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }

    u, err := uh.db.RetrieveUserByEmail(req.Email)
    if err != nil {
        respondWithError(w, 401, "Email and/or password are invalid.")
        return
    }
    
    err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
    if err != nil {
        respondWithError(w, 401, "Email and/or password are invalid.")
    }

    res := response {
        Id: u.Id,
        Email: u.Email,
    }
    respondWithJSON(w, 200, res)
}
