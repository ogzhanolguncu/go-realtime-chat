package server

import (
	"fmt"
	"log"
	"net"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/internal/auth"
	"github.com/ogzhanolguncu/go-chat/server/internal/block_user"
	"github.com/ogzhanolguncu/go-chat/server/internal/chat_history"
	"github.com/ogzhanolguncu/go-chat/server/internal/connection"
)

type TCPServer struct {
	listener          net.Listener
	connectionManager *connection.Manager
	historyManager    *chat_history.ChatHistory
	authManager       *auth.AuthManager
	blockUserManager  *block_user.BlockUserManager
	messageRouter     *MessageRouter
	encodeFn          func(payload protocol.Payload) string
	decodeFn          func(message string) (protocol.Payload, error)
}

// Server Initialization
// -----------------------------

func NewServer(port int, dbPath string, encoding bool) (*TCPServer, error) {
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

	server := &TCPServer{
		listener:          listener,
		connectionManager: cm,
		historyManager:    hm,
		authManager:       am,
		blockUserManager:  bum,
		encodeFn:          protocol.InitEncodeProtocol(encoding),
		decodeFn:          protocol.InitDecodeProtocol(encoding),
	}

	server.messageRouter = NewMessageRouter(server)

	return server, nil
}

func (s *TCPServer) Start() {
	log.Printf("Chat server started on %s\n", s.listener.Addr().String())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
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
	return nil
}

// Connection Handling
// -----------------------------

func (s *TCPServer) handleNewConnection(conn net.Conn) {
	handler := NewConnectionHandler(conn, s)
	handler.Handle()
}

// Client Event Handlers
// -----------------------------

func (s *TCPServer) OnClientJoin(info *connection.ConnectionInfo) {
	s.connectionManager.AddConnection(info.Connection, info)
	s.broadcastSystemNotice(fmt.Sprintf("%s has joined the chat.", info.OwnerName), info.Connection)
	s.broadcastActiveUsers()
}

// This function broadcasts to all clients when someone leaves or joins. Blocked and blocker users are also taken
// into consideration. The order of the function calls is crucial - do not rearrange them. Otherwise, "left" messages will be received by blocker and blocked clients.
func (s *TCPServer) OnClientLeave(info *connection.ConnectionInfo) {
	s.broadcastSystemNotice(fmt.Sprintf("%s has left the chat.", info.OwnerName), info.Connection)
	s.connectionManager.DeleteConnection(info.Connection)
	s.broadcastActiveUsers()
}

func (s *TCPServer) OnMessageReceived(info *connection.ConnectionInfo, message string) {
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
		s.messageRouter.sendSysResponse(excludeConn, "Failed to get blocker/blocked users", "fail")
	}

	encodedMsg := []byte(s.encodeFn(payload))
	s.messageRouter.broadcastToAll(encodedMsg, "Error sending system notice", excludedConns...)
}

func (s *TCPServer) broadcastActiveUsers() {
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

	conn.Write(msg)
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
