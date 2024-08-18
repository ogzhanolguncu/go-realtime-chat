package chat_history

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

type ChatHistory struct {
	db       *sqlx.DB
	encoding bool
}

type MessageEntry struct {
	Sender       string               `db:"sender"`
	Recipient    string               `db:"recipient"`
	MessageType  protocol.MessageType `db:"message_type"`
	Content      string               `db:"content"`
	BlockedUsers string               `db:"blocked_users"`
	Timestamp    time.Time            `db:"timestamp"`
}

func NewChatHistory(encoding bool, dbPath string) (*ChatHistory, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := createSchema(db); err != nil {
		db.Close() // Close the database if schema creation fails
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return &ChatHistory{
		encoding: encoding,
		db:       db,
	}, nil
}

func createSchema(db *sqlx.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender TEXT NOT NULL,
			recipient TEXT,
			message_type TEXT NOT NULL,
			content TEXT NOT NULL,
			blocked_users TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp)`,
	}
	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *ChatHistory) AddMessage(message string) error {
	decodedMsg, err := protocol.InitDecodeProtocol(ch.encoding)(message)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	entry := MessageEntry{
		Sender:      decodedMsg.Sender,
		Recipient:   decodedMsg.Recipient,
		MessageType: decodedMsg.MessageType,
		Content:     decodedMsg.Content,
		Timestamp:   time.Now(),
	}

	// Dynamically fetches blocked users for that sender and puts them into blocked_users table without extra call.
	query := `
   INSERT INTO messages (sender, recipient, message_type, content, timestamp, blocked_users)
SELECT :sender, :recipient, :message_type, :content, :timestamp,
    COALESCE(
        (SELECT GROUP_CONCAT(blocked, ',')
         FROM blocked_users
         WHERE blocker = :sender),
        ''
    ) || ',' ||
    COALESCE(
        (SELECT GROUP_CONCAT(blocker, ',')
         FROM blocked_users
         WHERE blocked = :sender),
        ''
    )
	`

	_, err = ch.db.NamedExec(query, entry)

	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (ch *ChatHistory) GetHistory(user string, messageTypes ...string) ([]string, error) {
	const messageLimit = 200
	// Default message types if none provided
	if len(messageTypes) == 0 {
		messageTypes = []string{"WSP", "MSG"}
	}

	query := `
    SELECT sender, recipient, message_type, content, timestamp
    FROM messages
    WHERE message_type IN (:message_type)
    AND (sender = :user OR recipient = '' OR recipient = :user)
    AND (
		(blocked_users = '' OR blocked_users = ',')
        OR NOT(
            blocked_users LIKE :user || ',%'
            OR blocked_users LIKE '%,' || :user || ',%'
            OR blocked_users LIKE '%,' || :user
            OR blocked_users LIKE '%' || :user || '%'
        )
    )
    ORDER BY timestamp ASC
    LIMIT :limit
    `

	params := map[string]interface{}{
		"user":         user,
		"message_type": messageTypes,
		"limit":        messageLimit,
	}

	query, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, fmt.Errorf("error in named query: %w", err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error expanding IN clause: %w", err)
	}

	query = ch.db.Rebind(query)

	log.Printf("Executing query: %s with args: %v", query, args)

	rows, err := ch.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var entries []MessageEntry
	for rows.Next() {
		var entry MessageEntry
		err := rows.StructScan(&entry)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message entry: %w", err)
		}
		entries = append(entries, entry)
	}

	encodedMessages := make([]string, len(entries))
	for i, entry := range entries {
		msg := protocol.Payload{
			Sender:      entry.Sender,
			Recipient:   entry.Recipient,
			MessageType: entry.MessageType,
			Content:     entry.Content,
		}
		encodedMessage := protocol.InitEncodeProtocol(ch.encoding)(msg)
		encodedMessages[i] = strings.TrimSpace(encodedMessage)
	}
	return encodedMessages, nil
}
func (ch *ChatHistory) Close() error {
	if err := ch.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
