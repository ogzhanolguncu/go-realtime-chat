package auth

import (
	"database/sql"
	"fmt"
	"unicode"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 14
const minimumPasswordLength = 8

type AuthManager struct {
	db *sql.DB
}

func NewAuthManager(dbPath string) (*AuthManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create users table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    )
    `)
	if err != nil {
		return nil, err
	}

	return &AuthManager{db: db}, nil
}

func (am *AuthManager) Close() error {
	return am.db.Close()
}

func (am *AuthManager) AddUser(username, password string) error {
	var exists string
	query := "SELECT username FROM users WHERE username = ?"
	err := am.db.QueryRow(query, username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing user: %v", err)
	}

	if exists != "" {
		return fmt.Errorf("username already exists")
	}

	if len(username) < 2 {
		return fmt.Errorf("username '%s' cannot be empty or less than two characters", username)
	}

	if !checkPasswordStrength(password, minimumPasswordLength) {
		return fmt.Errorf("password is not strong enough. It has to contain at least %d characters, one upper, one lower, one symbol and one digit", minimumPasswordLength)
	}

	hashedPass, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("could not hash user: %s's password - %v", username, err)
	}

	_, err = sq.Insert("users").Columns("username", "password").Values(username, hashedPass).RunWith(am.db).Exec()
	if err != nil {
		return fmt.Errorf("could not add user to users table - %v", err)
	}
	return nil
}

func checkPasswordStrength(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}

	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (am *AuthManager) AuthenticateUser(username, password string) (bool, error) {
	var storedPassword string

	err := sq.Select("password").
		From("users").
		Where(sq.Eq{"username": username}).
		RunWith(am.db).
		QueryRow().
		Scan(&storedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("authentication failed")
		}
		return false, fmt.Errorf("error querying user %s: %v", username, err)
	}

	// Compare the stored password with the provided password
	if checkPasswordHash(password, storedPassword) {
		return true, nil
	}

	return false, nil
}
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
