package database

func (db *DB) CreateUser(email string) (User, error) {
	nextId := db.currentUserId + 1
	data := User{
		Id:    nextId,
		Email: email,
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
