package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
    type serverError struct {
        Error string `json:"error"`
    }

    errResponse := serverError { Error: msg }

    respondWithJSON(w, code, errResponse)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    dat, err := json.Marshal(payload)

    if err != nil {
        w.WriteHeader(500)
        log.Printf("Error marhsalling JSON: %s", err)
        return
    }
    w.WriteHeader(code)
    w.Header().Set("Content-Type", "application/json")
    w.Write(dat)
}
