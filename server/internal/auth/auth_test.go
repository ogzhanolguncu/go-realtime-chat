package auth

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const testDBPath = "test_auth.db"

func setupTestDB(t *testing.T) *AuthManager {
	am, err := NewAuthManager(testDBPath)
	assert.NoError(t, err, "Error creating AuthManager")
	return am
}

func cleanupTestDB(t *testing.T, am *AuthManager) {
	err := am.Close()
	assert.NoError(t, err, "Error closing database")
	err = os.Remove(testDBPath)
	assert.NoError(t, err, "Error removing test database")
}

func TestAddUser(t *testing.T) {
	am := setupTestDB(t)
	defer cleanupTestDB(t, am)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"Valid user", "testuser", "P@ssw0rd", false},
		{"Duplicate username", "testuser", "AnotherP@ss1", true},
		{"Weak password", "newuser", "weak", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := am.AddUser(tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	am := setupTestDB(t)
	defer cleanupTestDB(t, am)

	err := am.AddUser("testuser", "P@ssw0rd")
	assert.NoError(t, err, "Error adding test user")

	tests := []struct {
		name     string
		username string
		password string
		want     bool
		wantErr  bool
	}{
		{"Valid credentials", "testuser", "P@ssw0rd", true, false},
		{"Invalid password", "testuser", "wrongpassword", false, true},
		{"Non-existent user", "nonexistent", "anypassword", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := am.AuthenticateUser(tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"Strong password", "P@ssw0rd", true},
		{"Weak password", "password", false},
		{"No uppercase", "p@ssw0rd", false},
		{"No lowercase", "P@SSW0RD", false},
		{"No digit", "P@ssword", false},
		{"No special char", "Passw0rd", false},
		{"Too short", "P@ss1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPasswordStrength(tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}
