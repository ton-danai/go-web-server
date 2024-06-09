package database

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path           string
	mux            *sync.RWMutex
	chirps         map[int]Chirp
	users          map[int]User
	currentChripId int
	currentUserId  int
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func New(path string) (*DB, error) {
	db := &DB{
		path:   path,
		mux:    new(sync.RWMutex),
		chirps: map[int]Chirp{},
		users:  map[int]User{},
	}

	err := db.ensureDB()
	if err != nil {
		log.Printf("db.ensureDB : %+v", err)
		return nil, err
	}
	// log.Println("this")
	data, loadError := db.loadDB()
	if loadError != nil {
		log.Printf("Load Data Error : %+v", loadError)
		return nil, loadError
	}

	maxKeyChirps := getMaxId(data.Chirps)
	maxKeyUsers := getMaxId(data.Users)

	db.chirps = data.Chirps
	db.users = data.Users
	db.currentChripId = maxKeyChirps
	db.currentUserId = maxKeyUsers

	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	nextId := db.currentChripId + 1
	data := Chirp{
		Id:   nextId,
		Body: body,
	}

	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}

	dbStructure.Chirps = db.chirps
	dbStructure.Users = db.users

	dbStructure.Chirps[nextId] = data

	err := db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	db.chirps = dbStructure.Chirps
	db.currentChripId = nextId
	return data, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	v := make([]Chirp, 0, len(db.chirps))

	sort.Slice(v, func(i, j int) bool {
		return v[i].Id < v[j].Id
	})

	return v, nil
}

func (db *DB) GetChirpById(id int) (Chirp, bool) {
	data, found := db.chirps[id]
	return data, found
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()
	_, err := os.Stat(db.path)
	if err != nil {
		_, createError := os.Create(db.path)
		if createError != nil {
			return createError
		}

		dbError := db.createDB()

		return dbError
	}

	return nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}
	rawJsonString, err := json.Marshal(dbStructure)

	if err != nil {
		return err
	}

	writeErr := os.WriteFile(db.path, rawJsonString, 0666)
	if writeErr != nil {
		return writeErr
	}

	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	result := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}

	data, err := os.ReadFile(db.path)

	if err != nil {
		return result, err
	}

	if len(data) > 0 {
		jsonError := json.Unmarshal(data, &result)

		if jsonError != nil {
			return result, jsonError
		}
	}

	return result, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	rawJsonString, err := json.Marshal(dbStructure)

	if err != nil {
		return err
	}

	writeErr := os.WriteFile(db.path, rawJsonString, 0666)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
