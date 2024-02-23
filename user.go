package main

import (
	"encoding/json"
    "errors"
    "fmt"
	"net/http"
    "log"
    "time"
    "strings"
    "strconv"

	"github.com/mike-cellini/chirpy/internal/database"

	"github.com/golang-jwt/jwt/v5"
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
    id, err := validateToken(token)
    if err != nil {
        respondWithErrorCode(w, 401)
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

    if req.ExpiresInSeconds == 0 {
        req.ExpiresInSeconds = 86400
    }

    token := createJwt(req.ExpiresInSeconds, u.Id)

    res := response {
        Id: u.Id,
        Email: u.Email,
        Token: token,
    }
    respondWithJSON(w, 200, res)
}

func createJwt(expiresInSeconds int, userId int) string {
    issuedAt := time.Now().UTC()
    expiresAt := issuedAt.Add(time.Duration(expiresInSeconds * int(time.Second)))

    var claims = jwt.RegisteredClaims {
        Issuer: "chirpy",
        IssuedAt: jwt.NewNumericDate(issuedAt),
        ExpiresAt: jwt.NewNumericDate(expiresAt),
        Subject: fmt.Sprint(userId),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    val, err := token.SignedString([]byte(apiCfg.jwtSecret))
    if err != nil {
        log.Printf("Unable to create token: %s", err.Error())
    }

    return val
}

func validateToken(token string) (int, error) {
    claims := jwt.RegisteredClaims {}
    t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
        return []byte(apiCfg.jwtSecret), nil
    })

    if err != nil {
        log.Printf("Token parse error: %s", err.Error())
        return 0, err
    }

    s, err := t.Claims.GetSubject()
    if err != nil {
        return 0, errors.New("Unable to parse claims")
    }

    log.Print("Converting", s)
    id, _ := strconv.Atoi(s)
    return id, nil
}
