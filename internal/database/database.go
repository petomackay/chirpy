package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Couldn't loadDB: " + err.Error())
		return Chirp{}, err
	}

	id := len(dbStruct.Chirps) + 1
	chirp := Chirp{
		Id:   id,
		Body: body,
	}

	dbStruct.Chirps[id] = chirp

	if err := db.writeDB(dbStruct); err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, nil
	}
	chirp, ok := dbStruct.Chirps[id]
	if ok {
		return chirp, nil
	}
	return Chirp{}, errors.New("Chirp not found.")
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStruct, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := make([]Chirp, 0, len(dbStruct.Chirps))
	for _, chirp := range dbStruct.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, nil
}

func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if _, err := os.ReadFile(db.path); errors.Is(err, os.ErrNotExist) {
		log.Printf("The DB file %s does not exist. Attempting to create it.\n", db.path)
		if _, err := os.Create(db.path); err != nil {
			log.Printf("Couldn't create file %s\n", db.path)
			return err
		}
		if err := db.writeDB(DBStructure{Chirps: make(map[int]Chirp)}); err != nil {
			log.Println("Error when initializing the DB: " + err.Error())
			return err
		}
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	contents, err := os.ReadFile(db.path)
	if err != nil {
		log.Println("Couldn't read DB file: " + err.Error())
		return DBStructure{}, err
	}

	dbStruct := DBStructure{}
	if err := json.Unmarshal(contents, &dbStruct); err != nil {
		log.Println("Couldn't unmarshall DB file: " + err.Error())
		return DBStructure{}, err
	}
	return dbStruct, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	if err := os.WriteFile(db.path, dat, 0666); err != nil {
		return err
	}
	return nil
}
