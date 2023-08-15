package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps      map[int]Chirp         `json:"chirps"`
	Users       map[int]User          `json:"user"`
	Revocations map[string]Revocation `json:"refresh_tokens"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Revocation struct {
	Token     string    `json:"token"`
	RevokedAt time.Time `json:"revoked_at"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.userExists(email) {
		return User{}, errors.New("User with that email already exists.")
	}

	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashed_password),
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DBStructure) userExists(email string) bool {
	for _, user := range db.Users {
		if user.Email == email {
			return true
		}
	}
	return false
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetChirpById(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	val, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("No Chirp")
	}
	return val, nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps:      map[int]Chirp{},
		Users:       map[int]User{},
		Revocations: map[string]Revocation{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AuthorizeUser(email, password string) (int, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return 0, err
	}
	for _, user := range dbStructure.Users {
		if user.Email == email {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				return user.ID, nil
			}
			return 0, errors.New("Password invalid.")
		}
	}
	return 0, errors.New("User not found.")
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, errors.New("Failed to load DB.")
	}

	user, exists := dbStructure.Users[id]
	if exists {
		user.Email = email
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		user.Password = string(hashed_password)
		dbStructure.Users[id] = user
		db.writeDB(dbStructure)
		return user, nil
	}
	return User{}, errors.New("User not found")
}

func (db *DB) SaveRefreshToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return errors.New("Failed to load DB.")
	}
	dbStructure.Revocations[token] = Revocation{Token: token, RevokedAt: time.Time{}}

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) IsRevoked(token string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, errors.New("Failed to load DB.")
	}

	revocation, exists := dbStructure.Revocations[token]	
	if !exists {
		return false, errors.New("This token is not in the DB.")
	}

	if revocation.RevokedAt.IsZero() {
		return true, nil
	} 

	return false, nil
}

func (db *DB) RevokeToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return errors.New("Failed to load DB.")
	}
	
	revocation, exists := dbStructure.Revocations[token]
	if exists {
		revocation.RevokedAt = time.Now()
		dbStructure.Revocations[token] = revocation
		db.writeDB(dbStructure)
	}

	dbStructure, err = db.loadDB()
	if err != nil {
		return errors.New("Failed to load DB.")
	}
	return nil
}
