package database

import (
    "log"
    "sort"
)

func (db *DB) CreateChirp(authorId int, body string) (Chirp, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return Chirp {}, err
    }

    id := len(dbStructure.Chirps) + 1
    
    c := Chirp { 
        AuthorId: authorId,
        Id: id, 
        Body: body,
    }
    dbStructure.Chirps[id] = c

    err = db.writeDB(dbStructure)
    if err != nil {
        log.Printf("ERROR: Could not write DB: %v", err.Error())
        return Chirp {}, err
    }

    return c, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return []Chirp {}, err
    }

    vals := make([]Chirp, 0, len(dbStructure.Chirps))
    for _, v := range dbStructure.Chirps {
        vals = append(vals, v)
    }

    sort.Slice(vals, func (a, b int) bool { return vals[a].Id < vals[b].Id })
    return vals, nil
}

func (db *DB) GetChirpById(id int) (chirp Chirp, ok bool, err error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database: %s", err.Error())
        return Chirp {}, false, err
    }

    chirp, ok = dbStructure.Chirps[id]

    return chirp, ok, nil
}

func (db *DB) DeleteChirp(id int) (err error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database: %s", err.Error())
        return err
    }

    delete(dbStructure.Chirps, id)
    return nil
}

