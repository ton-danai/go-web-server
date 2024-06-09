package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) Login(email, password string) (User, error) {
	user, found := db.findUserByEmail(email)
	if !found {
		return User{}, errors.New("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return User{}, errors.New("password incorrect")
	}

	return user, nil
}
