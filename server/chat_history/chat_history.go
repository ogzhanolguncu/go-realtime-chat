package chat_history

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

type ChatHistory struct {
	db       *sql.DB
	encoding bool
}

func NewChatHistory(encoding bool, dbPath string) (*ChatHistory, error) {
	db, err := sql.Open("sqlite3", dbPath)
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

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender TEXT NOT NULL,
			recipient TEXT NOT NULL,
			message_type TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS blocked_users (
			blocker TEXT NOT NULL,
			blocked TEXT NOT NULL,
			PRIMARY KEY (blocker, blocked)
		);
		CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	`)
	return err
}

func (ch *ChatHistory) AddMessage(message string) error {
	decodedMsg, err := protocol.InitDecodeProtocol(ch.encoding)(message)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	_, err = ch.db.Exec(
		"INSERT INTO messages (sender, recipient, message_type, content) VALUES (?, ?, ?, ?)",
		decodedMsg.Sender, decodedMsg.Recipient, decodedMsg.MessageType, decodedMsg.Content,
	)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

// TODO: When new message is received server can't return it to clients without restart fix here.
func (ch *ChatHistory) GetHistory(user string, messageTypes ...string) ([]string, error) {
	const messageLimit = 200
	query := `
	SELECT sender, recipient, message_type, content, timestamp
	FROM messages
	WHERE (sender = ? OR recipient = ?)
	AND sender NOT IN (SELECT blocked FROM blocked_users WHERE blocker = ?)
	AND recipient NOT IN (SELECT blocked FROM blocked_users WHERE blocker = ?)
	`
	params := []interface{}{user, user, user, user}

	if len(messageTypes) > 0 {
		query += fmt.Sprintf("AND message_type IN (%s) ", strings.Repeat("?,", len(messageTypes)-1)+"?")
		for _, msgType := range messageTypes {
			params = append(params, msgType)
		}
	}

	query += `
	ORDER BY timestamp DESC
	LIMIT ?
	`
	params = append(params, messageLimit)

	log.Printf("Executing query: %s with params: %v", query, params)
	rows, err := ch.db.Query(query, params...)
	if err != nil {

		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var encodedMessages []string
	for rows.Next() {
		var msg protocol.Payload
		var timestamp string
		if err := rows.Scan(&msg.Sender, &msg.Recipient, &msg.MessageType, &msg.Content, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		encodedMessage := protocol.InitEncodeProtocol(ch.encoding)(msg)
		encodedMessage = strings.TrimSpace(encodedMessage)
		encodedMessages = append(encodedMessages, encodedMessage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return encodedMessages, nil
}

func (ch *ChatHistory) Close() error {
	if err := ch.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
