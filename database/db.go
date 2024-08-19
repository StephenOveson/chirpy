package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	Id           int       `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := db.ensureDB()
	if err != nil {
		return &DB{}, nil
	}

	return db, nil
}

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to create chirp")
		return Chirp{}, err
	}

	id := len(data.Chirps) + 1
	chirp := Chirp{
		Body:     body,
		Id:       id,
		AuthorId: authorId,
	}

	data.Chirps[id] = chirp

	db.writeDB(data)

	return chirp, nil
}

func (db *DB) GetChirps(authorId int, sortBy string) ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}

	chirps := make([]Chirp, 0, len(data.Chirps))
	for _, chirp := range data.Chirps {
		if authorId == chirp.AuthorId {
			chirps = append(chirps, chirp)
		} else if authorId == 0 {
			chirps = append(chirps, chirp)
		}
	}

	if sortBy == "" || sortBy == "asc" {
		sort.Slice(chirps, func(i int, j int) bool {
			return chirps[i].Id < chirps[j].Id
		})
	} else {
		sort.Slice(chirps, func(i int, j int) bool {
			return chirps[i].Id > chirps[j].Id
		})
	}

	return chirps, nil
}

func (db *DB) GetChirpById(id int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := data.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("unable to find entry")
	}

	return chirp, nil
}

func (db *DB) DeleteChirpById(chirpId, userId int) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}

	chirp, ok := data.Chirps[chirpId]
	if !ok {
		return errors.New("could not find chirp")
	}

	if chirp.AuthorId != userId {
		return errors.New("unable to delete chirp")
	}

	delete(data.Chirps, chirpId)

	return db.writeDB(data)
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, u := range data.Users {
		if u.Email == email {
			return u, nil
		}
	}

	return User{}, errors.New("unable to find entry")
}

func (db *DB) CreateUser(email, password string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to create chirp")
		return User{}, err
	}

	hasDupEmail := db.checkDuplicateEmails(email)
	if hasDupEmail {
		return User{}, errors.New("email already in use")
	}

	id := len(data.Users) + 1
	user := User{
		Id:       id,
		Email:    email,
		Password: password,
	}

	data.Users[id] = user

	db.writeDB(data)

	return user, nil
}

func (db *DB) UpdateUser(id int, u User) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to load db")
		return User{}, err
	}
	user, ok := data.Users[id]
	if !ok {
		return User{}, errors.New("unable to find user")
	}

	if user.RefreshToken == "" {
		user.RefreshToken = u.RefreshToken
		user.ExpiresAt = u.ExpiresAt
	}
	user.Email = u.Email
	user.Password = u.Password

	data.Users[id] = user

	return user, db.writeDB(data)
}

func (db *DB) ConfirmUserToken(token string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to load db")
		return User{}, err
	}

	user, err := db.findUserByToken(data, token)
	if err != nil {
		return User{}, err
	}

	if time.Now().Before(user.ExpiresAt) {
		return user, nil
	}

	return User{}, errors.New("token does not exist on user")
}

func (db *DB) RevokeUserToken(token string) error {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to load db")
		return err
	}

	user, err := db.findUserByToken(data, token)
	if err != nil {
		return err
	}

	user.RefreshToken = ""
	user.ExpiresAt = time.Time{}
	data.Users[user.Id] = user

	return db.writeDB(data)
}

func (db *DB) UpgradeUser(userId int) error {
	data, err := db.loadDB()
	if err != nil {
		log.Fatal("Unable to load db")
		return err
	}

	user, ok := data.Users[userId]
	if !ok {
		return errors.New("unable to get user.")
	}

	user.IsChirpyRed = true

	data.Users[userId] = user

	return db.writeDB(data)

}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) findUserByToken(data DBStructure, token string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	for _, u := range data.Users {
		if u.RefreshToken == token {
			return u, nil
		}
	}

	return User{}, errors.New("Unable to find user token")
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatal("Unable to read from database.")
		return DBStructure{}, err
	}
	chirps := DBStructure{}
	err = json.Unmarshal(data, &chirps)
	if err != nil {
		return DBStructure{}, err
	}

	return chirps, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, data, 0600)
	if err != nil {
		log.Fatal("Unable to write to database.")
		return err
	}
	return nil
}

func (db *DB) checkDuplicateEmails(email string) bool {
	data, err := db.loadDB()
	if err != nil {
		return true
	}

	db.mux.RLock()
	defer db.mux.RUnlock()
	for _, user := range data.Users {
		if user.Email == email {
			return true
		}
	}

	return false
}
