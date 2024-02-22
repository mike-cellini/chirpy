package database

import (
    "errors"
    "log"
)

func (db *DB) CreateUser(email string, passHash string) (User, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return User {}, err
    }

    for _, u := range dbStructure.Users {
       if u.Email == email {
           return User {}, errors.New("User with that email already exists.")
       }
    }

    id := len(dbStructure.Users) + 1
    
    u := User { 
        Id: id,
        Email: email,
        PasswordHash: string(passHash),
    }
    dbStructure.Users[id] = u

    err = db.writeDB(dbStructure)
    if err != nil {
        log.Printf("ERROR: Could not write DB: %v", err.Error())
        return User {}, err
    }

    return u, nil
}

func (db *DB) RetrieveUserByEmail(email string) (User, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return User {}, err
    }

    for _, u := range dbStructure.Users {
       if u.Email == email {
           return u, nil
       }
    }

    err = errors.New("User does not exist")
    return User{}, err
}
