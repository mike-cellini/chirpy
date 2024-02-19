package database

import (
    "log"
)

func (db *DB) CreateUser(email string) (User, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return User {}, err
    }

    id := len(dbStructure.Users) + 1
    
    u := User { 
        Id: id, 
        Email: email,
    }
    dbStructure.Users[id] = u

    err = db.writeDB(dbStructure)
    if err != nil {
        log.Printf("ERROR: Could not write DB: %v", err.Error())
        return User {}, err
    }

    return u, nil
}
