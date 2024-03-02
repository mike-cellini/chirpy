package database

import (
    "log"
    "time"
)

func(db *DB) GetRevokedToken(token string) (revokedToken RevokedToken, ok bool, err error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return RevokedToken {}, false, err
    }

    revokedToken, ok = dbStructure.RevokedTokens[token]

    return revokedToken, ok, nil
}

func(db *DB) CreateRevokedToken(token string) error {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return err
    }

    rt := RevokedToken {
        Token: token, 
        RevokeDate: time.Now().UTC(),
    }

    dbStructure.RevokedTokens[token] = rt
    err = db.writeDB(dbStructure)
    if err != nil {
        log.Printf("ERROR: Unable to write data to database")
        return err
    }
    return nil
}
