package main

import (
	"encoding/json"
    "log"
	"net/http"
    "strings"

	"github.com/mike-cellini/chirpy/internal/database"

	"golang.org/x/crypto/bcrypt"
)

type userHandler struct {
    db *database.DB
}

type request struct {
    Email string `json:"email"`
    Password string `json:"password"`
    ExpiresInSeconds int `json:"expires_in_seconds"`
}

func (uh *userHandler) create(w http.ResponseWriter, r *http.Request) {
    type response struct {
        Email string `json:"email"`
        Id int `json:"id"`
    }

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

func (uh *userHandler) update(w http.ResponseWriter, r *http.Request) {
    type response struct {
        Email string `json:"email"`
        Id int `json:"id"`
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
    }

    pHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        respondWithError(w, 400, "Unable to hash password")
        return 
    }

    u, err := uh.db.UpdateUser(id, req.Email, string(pHash))
    res := response {
        Id: u.Id,
        Email: u.Email,
    }
    respondWithJSON(w, 200, res)
}

func (uh *userHandler) authenticate(w http.ResponseWriter, r *http.Request) {
    type response struct {
        Email string `json:"email"`
        Id int `json:"id"`
        Token string `json:"token"`
        RefreshToken string `json:"refresh_token"`
    }

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

    token := createToken(u.Id, getAccessIssuerAndExpiration)
    refreshToken := createToken(u.Id, getRefreshIssuerAndExpiration)

    res := response {
        Id: u.Id,
        Email: u.Email,
        Token: token,
        RefreshToken: refreshToken,
    }
    respondWithJSON(w, 200, res)
}

func (uh *userHandler) refresh(w http.ResponseWriter, r *http.Request) {
    type response struct {
        Token string `json:"token"`
    }

    auth := r.Header.Get("Authorization")
    token := strings.TrimPrefix(auth, "Bearer ")
    id, err := validateRefreshToken(token, uh.db)
    if err != nil {
        log.Print("Could not validate refresh token")
        w.WriteHeader(401)
        return
    }

    accessToken := createToken(id, getAccessIssuerAndExpiration)

    res := response {
        Token: accessToken,
    }
    respondWithJSON(w, 200, res)
}

func (uh *userHandler) revoke(w http.ResponseWriter, r *http.Request) {
    auth := r.Header.Get("Authorization")
    token := strings.TrimPrefix(auth, "Bearer ")

    err := revokeToken(token, uh.db)
    if err != nil {
        w.WriteHeader(401)
        return
    }
    w.WriteHeader(200)
}
