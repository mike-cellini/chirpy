package database

import (
    "encoding/json"
    "log"
    "fmt"
    "sync"
    "os"
    "errors"
    "sort"
)

type DB struct {
    path string
    mux *sync.RWMutex
}

type Chirp struct {
    Body string `json:"body"`
    Id int `json:"id"`
}

type DBStructure struct {
    Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string)(*DB, error) {
    db := &DB { 
        path: fmt.Sprintf("%s/database.json", path), 
        mux: &sync.RWMutex{},
    }
    db.mux.Lock()
    _, err := os.ReadFile(db.path)
    db.mux.Unlock()

    if errors.Is(err, os.ErrNotExist) {
        dbMap := DBStructure { Chirps: make(map[int]Chirp)}

        data, err := json.Marshal(dbMap); 
        if err != nil {
            log.Printf("ERROR: Could not marshal the JSON: %v\n", err.Error())
            return db, err
        }

        err = os.WriteFile(db.path, data, 0666)
        if err != nil {
            log.Printf("ERROR: Could not create the database: %v\n", err.Error())
            return db, err
        }
    } else if err != nil {
        return db, err
    }
    
    return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
    dbStructure, err := db.loadDB()
    if err != nil {
        log.Printf("ERROR: Unable to load data from database")
        return Chirp {}, err
    }

    var id int = 1
    
    if len(dbStructure.Chirps) > 0 {
        keys := make([]int, 0, len(dbStructure.Chirps))
        for k := range dbStructure.Chirps {
            keys = append(keys, k)
        }

        sort.Slice(keys, func (a, b int) bool { return a > b })

        id = keys[0] + 1
    }

    c := Chirp { Id: id, Body: body }
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

func (db *DB) loadDB() (DBStructure, error) {
    var dbMap DBStructure

    db.mux.RLock()
    defer db.mux.RUnlock()

    data, err := os.ReadFile(db.path)
    if err != nil {
        log.Printf("ERROR: Could not load data from %v: %v\n", db.path, err.Error())
        return dbMap, err
    }
    err = json.Unmarshal(data, &dbMap)
    if err != nil {
        log.Printf("ERROR: Could not parse JSON: %v\n", err.Error())
        return dbMap, err
    }
    return dbMap, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
    db.mux.Lock()
    defer db.mux.Unlock()

    data, err := json.Marshal(dbStructure)
    if err != nil {
        log.Printf("ERROR: Could not marshal JSON: %v\n", err.Error())
        return err
    }
    err = os.WriteFile(db.path, data, 0666)
    if err != nil {
        log.Printf("ERROR: Could not write data to: %v: %v\n", db.path, err.Error())
        return err
    }
    return nil
}
