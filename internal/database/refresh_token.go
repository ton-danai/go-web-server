package database

import (
	"log"
	"time"
)

func (db *DB) StoreRefreshToken(id int, token string) error {
	limit := time.Hour * 24 * 60
	exp := time.Now().Add(limit).Unix()

	item := RefreshToken{
		UserId: id,
		Token:  token,
		ExpAt:  exp,
	}

	dbStructure := db.mapDBStructure()
	dbStructure.RefreshTokens[id] = item
	err := db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) VerifyRefreshToken(token string) (int, bool) {
	record, found := findToken(&db.refreshToken, &token)
	if !found {
		return 0, false
	}

	now := time.Now().Unix()
	log.Println(now)
	log.Println(record.ExpAt)
	if now > record.ExpAt {
		return 0, false
	}

	return record.UserId, true
}

func findToken(data *map[int]RefreshToken, token *string) (RefreshToken, bool) {
	for _, value := range *data {
		if value.Token == *token {
			return value, true
		}
	}

	return RefreshToken{}, false
}

func (db *DB) RevokeToken(token string) bool {
	record, found := findToken(&db.refreshToken, &token)
	if !found {
		return false
	}

	dbStructure := db.mapDBStructure()
	delete(dbStructure.RefreshTokens, record.UserId)

	return true
}
