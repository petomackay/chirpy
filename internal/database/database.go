package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Chirp struct {
	Id     int    `json:"id"`
	Body   string `json:"body"`
	UserId int    `json:"author_id"`
}

type DBStructure struct {
	Chirps  map[int]Chirp    `json:"chirps"`
	Users   map[int]User     `json:"users"`
	Revoked map[string]int64 `json:"revoked"`
}

var ErrAlreadyExists = errors.New("already exists")
var ErrNotExist = errors.New("does not exist")

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

func (db *DB) CreateUser(email string, password string) (User, error) {
	if _, err := db.FindUserByEmail(email); !errors.Is(err, ErrNotExist) {
		return User{}, ErrAlreadyExists
	}

	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Couldn't load DB: " + err.Error())
		return User{}, err
	}

	id := len(dbStruct.Users) + 1
	pwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Couldn't hash user password: %s", err)
		return User{}, nil
	}

	user := User{
		Id:       id,
		Email:    email,
		Password: string(pwd),
	}
	dbStruct.Users[id] = user
	if err := db.writeDB(dbStruct); err != nil {
		return User{}, nil
	}
	return user, nil
}

func (db *DB) UpdateUser(user User) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Couldn't load DB: " + err.Error())
		return err
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Couldn't hash user password: %s", err)
		return err
	}
	user.Password = string(pwd)
	dbStruct.Users[user.Id] = user
	err = db.writeDB(dbStruct)
	return err

}
func (db *DB) FindUserByEmail(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Coudln't load DB: " + err.Error())
		return User{}, nil
	}

	for _, user := range dbStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, ErrNotExist
}

func (db *DB) FindUserById(id int) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Coudln't load DB: " + err.Error())
		return User{}, nil
	}

	if user, ok := dbStruct.Users[id]; ok {
		log.Println("found user")
		return user, nil
	}
	log.Printf("didn't find user with id, %d\n", id)
	return User{}, ErrNotExist
}

func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Couldn't loadDB: " + err.Error())
		return Chirp{}, err
	}

	id := len(dbStruct.Chirps) + 1
	chirp := Chirp{
		Id:     id,
		Body:   body,
		UserId: userId,
	}

	dbStruct.Chirps[id] = chirp

	if err := db.writeDB(dbStruct); err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
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
	if _, err := os.ReadFile(db.path); errors.Is(err, os.ErrNotExist) {
		log.Printf("The DB file %s does not exist. Attempting to create it.\n", db.path)
		if _, err := os.Create(db.path); err != nil {
			log.Printf("Couldn't create file %s\n", db.path)
			return err
		}
		emptyStruct := DBStructure{
			Chirps:  make(map[int]Chirp),
			Users:   make(map[int]User),
			Revoked: make(map[string]int64),
		}
		if err := db.writeDB(emptyStruct); err != nil {
			log.Println("Error when initializing the DB: " + err.Error())
			return err
		}
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

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
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	if err := os.WriteFile(db.path, dat, 0666); err != nil {
		return err
	}
	return nil
}

func (db *DB) RevokeToken(tokenString string) error {
	t := time.Now().UnixMilli()
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}
	dbStruct.Revoked[tokenString] = t
	return db.writeDB(dbStruct)
}

func (db *DB) IsTokenRevoked(tokenString string) bool {
	dbStruct, err := db.loadDB()
	if err != nil {
		log.Println("Couldn't load DB: " + err.Error())
		return true
	}
	_, ok := dbStruct.Revoked[tokenString]
	return ok
}
