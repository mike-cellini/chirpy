package main

import (
    "encoding/json"
    "net/http"
    "strings"
    
    "github.com/mike-cellini/chirpy/internal/database"
)

type polkaHandler struct {
    db *database.DB
}

func (ph *polkaHandler) handleEvent(w http.ResponseWriter, r *http.Request) {
    type requestBody struct {
        UserId int `json:"user_id"`
    }

    type request struct {
        Event string `json:"event"`
        Data requestBody `json:"data"` 
    }

    auth := r.Header.Get("Authorization")
    key := strings.TrimPrefix(auth, "ApiKey ")
    if key != apiCfg.polkaKey {
        w.WriteHeader(401)
        return
    }

    decoder := json.NewDecoder(r.Body)
    req := request{}
    err := decoder.Decode(&req)

    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    } else if req.Event != "user.upgraded" {
        return
    }
    _, ok, err := ph.db.UpgradeUser(req.Data.UserId)
    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    } else if !ok {
        w.WriteHeader(404)
        return
    }
    w.WriteHeader(200)
}
