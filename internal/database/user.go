package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) findUserByEmail(email string) (User, bool) {
	for _, value := range db.users {
		if value.Email == email {
			return value, true
		}
	}

	return User{}, false
}

func (db *DB) CreateUser(email, password string) (User, error) {
	nextId := db.currentUserId + 1
	hashPassword, hasError := bcrypt.GenerateFromPassword([]byte(password), 14)

	if hasError != nil {
		return User{}, hasError
	}

	data := User{
		Id:       nextId,
		Email:    email,
		Password: string(hashPassword),
	}

	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}

	dbStructure.Chirps = db.chirps
	dbStructure.Users = db.users

	dbStructure.Users[nextId] = data

	err := db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	db.users = dbStructure.Users
	db.currentUserId = nextId
	return data, nil
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	data, found := db.users[id]

	if !found {
		return User{}, errors.New("not found")
	}

	hashPassword, hasError := bcrypt.GenerateFromPassword([]byte(password), 14)
	if hasError != nil {
		return User{}, hasError
	}

	data.Email = email
	data.Password = string(hashPassword)

	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}

	dbStructure.Chirps = db.chirps
	dbStructure.Users = db.users

	dbStructure.Users[id] = data

	err := db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	db.users = dbStructure.Users

	return data, nil
}
