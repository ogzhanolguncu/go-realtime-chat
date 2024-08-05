package auth

import (
	"database/sql"
	"fmt"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 8
const minimumPasswordLength = 8

type AuthManager struct {
	db              *sql.DB
	getUserStmt     *sql.Stmt
	addUserStmt     *sql.Stmt
	getPasswordStmt *sql.Stmt
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

	// Create index on username
	_, err = db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_username ON users(username)
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}

	am := &AuthManager{db: db}
	if err := am.prepareStatements(); err != nil {
		return nil, err
	}

	return am, nil
}

func (am *AuthManager) prepareStatements() error {
	var err error
	am.getUserStmt, err = am.db.Prepare("SELECT username FROM users WHERE username = ?")
	if err != nil {
		return err
	}
	am.addUserStmt, err = am.db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
	if err != nil {
		am.getUserStmt.Close()
		return err
	}
	am.getPasswordStmt, err = am.db.Prepare("SELECT password FROM users WHERE username = ?")
	if err != nil {
		am.getUserStmt.Close()
		am.addUserStmt.Close()
		return err
	}
	return nil
}

func (am *AuthManager) Close() error {
	am.getUserStmt.Close()
	am.addUserStmt.Close()
	am.getPasswordStmt.Close()
	return am.db.Close()
}

func (am *AuthManager) AddUser(username, password string) error {
	var exists string
	err := am.getUserStmt.QueryRow(username).Scan(&exists)
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

	_, err = am.addUserStmt.Exec(username, hashedPass)
	if err != nil {
		return fmt.Errorf("could not add user to users table - %v", err)
	}

	return nil
}

func checkPasswordStrength(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
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
		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}
	return false
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func (am *AuthManager) AuthenticateUser(username, password string) (bool, error) {
	var storedPassword string
	err := am.getPasswordStmt.QueryRow(username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("authentication failed")
		}
		return false, fmt.Errorf("error querying user %s: %v", username, err)
	}
	return checkPasswordHash(password, storedPassword), nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
