package block_user

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type BlockUserManager struct {
	db *sqlx.DB
}

type BlockedUserEntry struct {
	Blocker   string    `db:"blocker"`
	Blocked   string    `db:"blocked"`
	Timestamp time.Time `db:"timestamp"`
}

func NewBlockUserManager(dbPath string) (*BlockUserManager, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return &BlockUserManager{db: db}, nil
}

func createSchema(db *sqlx.DB) error {
	statement := `CREATE TABLE IF NOT EXISTS blocked_users (
        blocker TEXT NOT NULL,
        blocked TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (blocker, blocked)
    )`
	_, err := db.Exec(statement)
	return err
}

func (bu *BlockUserManager) BlockUser(blocker, blocked string) error {
	_, err := bu.db.Exec("INSERT OR IGNORE INTO blocked_users (blocker, blocked) VALUES (?, ?)", blocker, blocked)
	return err
}

func (bu *BlockUserManager) UnblockUser(blocker, blocked string) error {
	_, err := bu.db.Exec("DELETE FROM blocked_users WHERE blocker=? AND blocked=?", blocker, blocked)
	return err
}

func (bu *BlockUserManager) IsBlocked(blocker, blocked string) (bool, error) {
	var exists bool
	err := bu.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM blocked_users WHERE blocker=? AND blocked=?)", blocker, blocked)
	return exists, err
}

func (bu *BlockUserManager) GetBlockedUsers(blocker string) ([]string, error) {
	var blockedUsers []string
	err := bu.db.Select(&blockedUsers, "SELECT blocked FROM blocked_users WHERE blocker = ?", blocker)
	return blockedUsers, err
}

func (bu *BlockUserManager) GetBlockerUsers(blocked string) ([]string, error) {
	var blockerUsers []string
	err := bu.db.Select(&blockerUsers, "SELECT blocker FROM blocked_users WHERE blocked = ?", blocked)
	return blockerUsers, err
}

func (bu *BlockUserManager) GetBlockedUsersWithDetails(blocker string) ([]BlockedUserEntry, error) {
	var entries []BlockedUserEntry
	err := bu.db.Select(&entries, "SELECT * FROM blocked_users WHERE blocker = ?", blocker)
	return entries, err
}

func (bu *BlockUserManager) Close() error {
	return bu.db.Close()
}
