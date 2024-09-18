package server

import (
	"fmt"
	"net"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	chat_ratelimit "github.com/ogzhanolguncu/go-chat/ratelimit"
	"github.com/ogzhanolguncu/go-chat/server/internal/auth"
	"github.com/ogzhanolguncu/go-chat/server/internal/block_user"
	"github.com/ogzhanolguncu/go-chat/server/internal/channels"
	"github.com/ogzhanolguncu/go-chat/server/internal/chat_history"
	"github.com/ogzhanolguncu/go-chat/server/internal/connection"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

type TCPServer struct {
	listener net.Listener

	connectionManager *connection.Manager
	historyManager    *chat_history.ChatHistory
	authManager       *auth.AuthManager
	blockUserManager  *block_user.BlockUserManager
	channelManager    *channels.Manager

	messageRouter *MessageRouter
	encodeFn      func(payload protocol.Payload) string
	decodeFn      func(message string) (protocol.Payload, error)

	ratelimiter *chat_ratelimit.Ratelimit
}

// Server Initialization
// -----------------------------

func NewServer(port int, dbPath string, encoding bool) (*TCPServer, error) {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	cm := connection.NewConnectionManager()
	am, err := auth.NewAuthManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}
	hm, err := chat_history.NewChatHistory(encoding, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat history manager: %w", err)
	}
	bum, err := block_user.NewBlockUserManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize block user manager: %w", err)
	}
	chanm := channels.NewChannelManager(cm, protocol.InitEncodeProtocol(encoding))

	server := &TCPServer{
		listener: listener,

		connectionManager: cm,
		historyManager:    hm,
		authManager:       am,
		blockUserManager:  bum,
		channelManager:    chanm,

		encodeFn: protocol.InitEncodeProtocol(encoding),
		decodeFn: protocol.InitDecodeProtocol(encoding),

		ratelimiter: chat_ratelimit.NewRatelimit(
			chat_ratelimit.TokenBucket{
				RefillInterval: time.Second * 3,
				RefillRate:     1,
				BucketLimit:    10,
			}),
	}

	server.messageRouter = NewMessageRouter(server)

	return server, nil
}

func (s *TCPServer) Start() {
	logger.Info("Server started. Listening for connections...")
	for {
		conn, err := s.listener.Accept()
		s.ratelimiter.Add(conn)
		if err != nil {
			logger.WithError(err).Error("Error accepting connection")
			continue
		}
		go s.handleNewConnection(conn)
	}
}

func (s *TCPServer) Close() error {
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	if err := s.historyManager.Close(); err != nil {
		return fmt.Errorf("failed to close history manager: %w", err)
	}
	if err := s.authManager.Close(); err != nil {
		return fmt.Errorf("failed to close auth manager: %w", err)
	}
	if err := s.blockUserManager.Close(); err != nil {
		return fmt.Errorf("failed to close block user manager: %w", err)
	}
	logger.Info("Server closed successfully")
	return nil
}

// Connection Handling
// -----------------------------

func (s *TCPServer) handleNewConnection(conn net.Conn) {
	// This function mainly handles auth then forwards users to message router
	handler := NewConnectionHandler(conn, s)
	handler.Handle()
}

// Client Event Handlers
// -----------------------------

func (s *TCPServer) OnClientJoin(info *connection.ConnectionInfo) {
	s.connectionManager.AddConnection(info.Connection, info)
	logger.WithField("user", info.OwnerName).Info("Client joined the chat")
	s.broadcastSystemNotice(fmt.Sprintf("%s has joined the chat.", info.OwnerName), info.Connection)
	s.broadcastActiveUsers()
}

func (s *TCPServer) OnClientLeave(info *connection.ConnectionInfo) {
	logger.WithField("user", info.OwnerName).Info("Client left the chat")
	s.broadcastSystemNotice(fmt.Sprintf("%s has left the chat.", info.OwnerName), info.Connection)
	s.connectionManager.DeleteConnection(info.Connection)
	s.ratelimiter.Remove(info.Connection)
	s.broadcastActiveUsers()
}

func (s *TCPServer) OnMessageReceived(info *connection.ConnectionInfo, message string) {
	logger.WithFields(logrus.Fields{
		"user":    info.OwnerName,
		"message": message,
	}).Info("Message received")

	allowed := s.ratelimiter.Check(info.Connection)
	if !allowed {
		s.messageRouter.sendSysResponse(info.Connection, "Please wait a moment before sending your next message", "fail")
		return
	}

	s.historyManager.AddMessage(message)
	s.messageRouter.RouteMessage(info, message)
}

// Broadcasting Methods
// -----------------------------

func (s *TCPServer) broadcastSystemNotice(message string, excludeConn net.Conn) {
	payload := protocol.Payload{
		MessageType: protocol.MessageTypeSYS,
		Content:     message,
		Status:      "success",
	}

	excludedConns, err := s.messageRouter.getExcludedConnections(excludeConn)
	if err != nil {
		logger.WithError(err).Error("Failed to get blocker/blocked users")
		s.messageRouter.sendSysResponse(excludeConn, "Failed to get blocker/blocked users", "fail")
	}

	encodedMsg := []byte(s.encodeFn(payload))
	s.messageRouter.broadcastToAll(encodedMsg, "Error sending system notice", excludedConns...)
}

func (s *TCPServer) broadcastActiveUsers() {
	logger.Info("Broadcasting active users")
	s.connectionManager.RangeConnections(func(conn net.Conn, _ *connection.ConnectionInfo) bool {
		s.sendActiveUsers(conn)
		return true
	})
}

func (s *TCPServer) sendActiveUsers(conn net.Conn) {
	activeUsers := s.connectionManager.GetActiveUsers()
	connectionInfo, ok := s.connectionManager.GetConnectionInfo(conn)
	if ok {
		blockedUsers, _ := s.blockUserManager.GetBlockedUsers(connectionInfo.OwnerName)
		blockerUsers, _ := s.blockUserManager.GetBlockerUsers(connectionInfo.OwnerName)

		activeUsers = filterActiveUsers(activeUsers, append(blockedUsers, blockerUsers...))
	}

	msg := []byte(s.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeACT_USRS,
		ActiveUsers: activeUsers,
		Status:      "res",
	}))

	if _, err := conn.Write(msg); err != nil {
		logger.WithError(err).Error("Failed to send active users")
	}
}

// Helper Functions
// -----------------------------

func filterActiveUsers(activeUsers, excludeUsers []string) []string {
	filtered := make([]string, 0)
	for _, user := range activeUsers {
		if !contains(excludeUsers, user) {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
