package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/internal/auth"
	"github.com/ogzhanolguncu/go-chat/server/internal/block_user"
	"github.com/ogzhanolguncu/go-chat/server/internal/chat_history"
	"github.com/ogzhanolguncu/go-chat/server/internal/connection"
	"github.com/stretchr/testify/assert"
)

// TestConn implements net.Conn for testing
type TestConn struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
}

func (tc *TestConn) Read(b []byte) (n int, err error)   { return tc.ReadBuffer.Read(b) }
func (tc *TestConn) Write(b []byte) (n int, err error)  { return tc.WriteBuffer.Write(b) }
func (tc *TestConn) Close() error                       { return nil }
func (tc *TestConn) LocalAddr() net.Addr                { return nil }
func (tc *TestConn) RemoteAddr() net.Addr               { return nil }
func (tc *TestConn) SetDeadline(t time.Time) error      { return nil }
func (tc *TestConn) SetReadDeadline(t time.Time) error  { return nil }
func (tc *TestConn) SetWriteDeadline(t time.Time) error { return nil }

func TestConnectionHandler_Authenticate(t *testing.T) {
	// Setup in-memory SQLite database for testing
	dbPath := ":memory:"

	tests := []struct {
		name           string
		input          string
		setupAuth      func(*auth.AuthManager) error
		expectedResult bool
		expectedWrite  string
	}{
		{
			name:  "Successful Authentication",
			input: "USR|1234567890|testuser|Test1234.|\r\n",
			setupAuth: func(am *auth.AuthManager) error {
				return am.AddUser("testuser", "Test1234.")
			},
			expectedResult: true,
			expectedWrite:  "testuser||success\r\n",
		},
		{
			name:           "Failed Authentication",
			input:          "USR|1234567890|testuser|wrongpassword|\r\n",
			expectedResult: false,
			expectedWrite:  "Authentication failed||fail\r\n",
		},
		{
			name:           "Authentication Error - Weak Password",
			input:          "USR|1234567890|weakuser|weak|\r\n",
			expectedResult: false,
			expectedWrite:  "Password does not meet strength requirements||fail\r\n",
		},
		{
			name:           "Authentication Error - Short Username",
			input:          "USR|1234567890|o|Test1234.|\r\n",
			expectedResult: false,
			expectedWrite:  "Username must be at least 2 characters long||fail\r\n",
		},
		{
			name:           "Invalid Data Format",
			input:          "INVALID|DATA|FORMAT|\r\n",
			expectedResult: false,
			expectedWrite:  "Invalid data format||fail\r\n",
		},
	}
	authManager, err := auth.NewAuthManager(dbPath)
	assert.NoError(t, err)
	defer authManager.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			testConn := &TestConn{
				ReadBuffer:  bytes.NewBufferString(tt.input),
				WriteBuffer: &bytes.Buffer{},
			}

			// Create actual components

			historyManager, err := chat_history.NewChatHistory(false, dbPath)
			assert.NoError(t, err)
			defer historyManager.Close()

			blockUserManager, err := block_user.NewBlockUserManager(dbPath)
			assert.NoError(t, err)
			defer blockUserManager.Close()

			server := &TCPServer{
				connectionManager: connection.NewConnectionManager(),
				historyManager:    historyManager,
				authManager:       authManager,
				blockUserManager:  blockUserManager,
				encodeFn:          protocol.InitEncodeProtocol(false),
				decodeFn:          protocol.InitDecodeProtocol(false),
			}

			handler := NewConnectionHandler(testConn, server)

			// Execute
			result := handler.authenticate()

			// Assert
			assert.Equal(t, tt.expectedResult, result)
			assert.Contains(t, testConn.WriteBuffer.String(), tt.expectedWrite)
		})
	}
}
