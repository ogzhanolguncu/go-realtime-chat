package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost            = 8 // Increased from 8 for better security
	minimumPasswordLength = 8 // Increased from 8 for better security
)

var (
	ErrInvalidUsername      = errors.New("username must be at least 2 characters long")
	ErrWeakPassword         = errors.New("password does not meet strength requirements")
	ErrAuthenticationFailed = errors.New("invalid username or password")
)

type AuthManager struct {
	db              *sql.DB
	getUserStmt     *sql.Stmt
	addUserStmt     *sql.Stmt
	getPasswordStmt *sql.Stmt
}

func NewAuthManager(dbPath string) (*AuthManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	am := &AuthManager{db: db}
	if err := am.prepareStatements(); err != nil {
		db.Close()
		return nil, err
	}

	return am, nil
}

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_username ON users(username);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func (am *AuthManager) prepareStatements() error {
	var err error

	am.getUserStmt, err = am.db.Prepare("SELECT username FROM users WHERE username = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare getUserStmt: %w", err)
	}

	am.addUserStmt, err = am.db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
	if err != nil {
		am.getUserStmt.Close()
		return fmt.Errorf("failed to prepare addUserStmt: %w", err)
	}

	am.getPasswordStmt, err = am.db.Prepare("SELECT password FROM users WHERE username = ?")
	if err != nil {
		am.getUserStmt.Close()
		am.addUserStmt.Close()
		return fmt.Errorf("failed to prepare getPasswordStmt: %w", err)
	}

	return nil
}

func (am *AuthManager) Close() error {
	stmts := []*sql.Stmt{am.getUserStmt, am.addUserStmt, am.getPasswordStmt}
	for _, stmt := range stmts {
		if stmt != nil {
			stmt.Close()
		}
	}
	return am.db.Close()
}

func (am *AuthManager) AddUser(username, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	hashedPass, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("could not hash password: %w", err)
	}

	_, err = am.addUserStmt.Exec(username, hashedPass)
	if err != nil {
		return fmt.Errorf("could not add user to database: %w", err)
	}

	return nil
}

func validateUsername(username string) error {
	if len(username) < 2 {
		return ErrInvalidUsername
	}
	return nil
}

func validatePassword(password string) error {
	if !checkPasswordStrength(password) {
		return ErrWeakPassword
	}
	return nil
}

func checkPasswordStrength(password string) bool {
	if len(password) < minimumPasswordLength {
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
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func (am *AuthManager) AuthenticateUser(username, password string) (bool, error) {
	var storedPassword string
	err := am.getPasswordStmt.QueryRow(username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User does not exist in the database '%s' creating new record", username)
			err = am.AddUser(username, password)
			if err != nil {
				return false, err
			}
			return true, nil

		}
		return false, fmt.Errorf("error querying user: %w", err)
	}

	// This means we found the user in the database, but password is wrong
	if !checkPasswordHash(password, storedPassword) {
		log.Printf("User does exist in the database '%s' but password is wrong", username)
		return false, ErrAuthenticationFailed
	}

	log.Printf("User '%s' has successfully logged in", username)
	return true, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
