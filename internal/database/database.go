package database

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path      string
	mux       *sync.RWMutex
	data      map[int]Chirp
	currentId int
}

type DBChirps struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func New(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  new(sync.RWMutex),
		data: map[int]Chirp{},
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

	maxKey := 0
	for key := range data.Chirps {
		// v = append(v, value)
		if key > maxKey {
			maxKey = key
		}
	}

	db.data = data.Chirps
	db.currentId = maxKey

	log.Printf("MaxKey %d", maxKey)
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	nextId := db.currentId + 1
	data := Chirp{
		Id:   nextId,
		Body: body,
	}

	dbStructure := DBChirps{
		Chirps: map[int]Chirp{},
	}

	for _, item := range db.data {
		dbStructure.Chirps[item.Id] = item
	}

	dbStructure.Chirps[nextId] = data

	err := db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	db.data = dbStructure.Chirps
	db.currentId = nextId
	return data, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	v := make([]Chirp, 0, len(db.data))

	sort.Slice(v, func(i, j int) bool {
		return v[i].Id < v[j].Id
	})

	return v, nil
}

func (db *DB) GetChirpById(id int) (Chirp, bool) {
	data, found := db.data[id]
	return data, found
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()
	_, err := os.Stat(db.path)
	if err != nil {
		_, createError := os.Create(db.path)
		return createError
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBChirps, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	result := DBChirps{
		Chirps: map[int]Chirp{},
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
func (db *DB) writeDB(dbStructure DBChirps) error {
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
