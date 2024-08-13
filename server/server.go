package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"sync"

	"github.com/elliotchance/pie/v2"
	protocol "github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/auth"
	"github.com/ogzhanolguncu/go-chat/server/block_user"
	"github.com/ogzhanolguncu/go-chat/server/chat_history"
	"github.com/ogzhanolguncu/go-chat/server/utils"
)

// ConnectionInfo holds connection-related information.
const dbName = "/chat.db"

type ConnectionInfo struct {
	Connection net.Conn
	OwnerName  string
}

type TCPServer struct {
	connectionMap    sync.Map
	history          *chat_history.ChatHistory
	authManager      *auth.AuthManager
	blockUserManager *block_user.BlockUserManager
	encodeFn         func(payload protocol.Payload) string
	decodeFn         func(message string) (protocol.Payload, error)
}

func newServer() *TCPServer {
	encoding := flag.Bool("encoding", false, "enable encoding")
	flag.Parse()

	var encodingType string
	if *encoding {
		encodingType = "BASE64"
	} else {
		encodingType = "PLAIN-TEXT"
	}

	log.Printf("------ ENCODING SET TO %s ------", encodingType)
	dbPath := fmt.Sprintf(utils.RootDir() + dbName)

	authManager, err := auth.NewAuthManager(dbPath)
	if err != nil {
		log.Printf("Failed to initialize auth manager: %v", err)
	}
	chatManager, err := chat_history.NewChatHistory(*encoding, dbPath)
	if err != nil {
		log.Printf("Failed to initialize auth manager: %v", err)
	}
	blockUserManager, err := block_user.NewBlockUserManager(dbPath)
	if err != nil {
		log.Printf("Failed to initialize auth manager: %v", err)
	}

	return &TCPServer{
		decodeFn:         protocol.InitDecodeProtocol(*encoding),
		encodeFn:         protocol.InitEncodeProtocol(*encoding),
		history:          chatManager,
		blockUserManager: blockUserManager,
		authManager:      authManager,
	}
}

func (s *TCPServer) addConnection(c net.Conn, info *ConnectionInfo) {
	s.connectionMap.Store(c, info)
}

func (s *TCPServer) getConnectedUsersCount() int {
	count := 0
	s.connectionMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *TCPServer) getActiveUsers() []string {
	var users []string
	s.connectionMap.Range(func(key, value interface{}) bool {
		info := value.(*ConnectionInfo)
		users = append(users, info.OwnerName)
		return true
	})
	return users
}

func (s *TCPServer) getConnectionInfoAndDelete(c net.Conn) (*ConnectionInfo, bool) {
	info, ok := s.getConnectionInfo(c)
	if !ok {
		return nil, false
	}
	s.connectionMap.Delete(c)
	return info, ok
}

func (s *TCPServer) getConnectionInfo(c net.Conn) (*ConnectionInfo, bool) {
	value, ok := s.connectionMap.Load(c)
	if !ok {
		return nil, false
	}
	info, ok := value.(*ConnectionInfo)
	return info, ok
}

func (s *TCPServer) handleNewConnection(c net.Conn) {
	name := s.handleAuth(c)
	// If the username is an empty string after exhausting retries,
	// close the connection to prevent clients with no username from connecting.
	if len(name) == 0 {
		c.Close()
		return
	}

	log.Printf("Recently joined user's name: %s\n", name)
	s.addConnection(c, &ConnectionInfo{Connection: c, OwnerName: name})
	connectedUsers := s.getConnectedUsersCount()
	log.Printf("Connection from %s\n", c.RemoteAddr().String())
	log.Printf("Connected users: %d\n", connectedUsers)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		//Todo: Prevent blocked users to see each other join
		s.sendSysNotice(name, c, "joined")

	}()

	go func() {
		defer wg.Done()
		//TODO: Prevent blocked users to see each other join
		s.sendActiveUsers(c)
	}()

	// Wait for both messages to be sent
	wg.Wait()

	s.handleConnection(c)
}

func (s *TCPServer) findConnectionByOwnerName(ownerName string) (net.Conn, bool) {
	var foundConn net.Conn
	var found bool

	s.connectionMap.Range(func(key, value interface{}) bool {
		conn := key.(net.Conn)
		info := value.(*ConnectionInfo)

		if info.OwnerName == ownerName {
			foundConn = conn
			found = true
			return false
		}
		return true
	})
	return foundConn, found
}

func (s *TCPServer) handleConnection(c net.Conn) {
	defer func(conn net.Conn) {
		info, _ := s.getConnectionInfoAndDelete(conn)
		if info != nil {
			s.sendSysNotice(info.OwnerName, conn, "left")
			s.sendActiveUsers(conn)
		}
		conn.Close()
	}(c)

	connReader := bufio.NewReader(c)

	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}

		rawMessage := strings.TrimSpace(data)
		s.history.AddMessage(rawMessage)
		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), rawMessage)

		msgPayload, err := s.decodeFn(rawMessage)
		if err != nil {
			// Write back to client that their message is malformed
			c.Write([]byte(err.Error()))
		}

		switch msgPayload.MessageType {
		case protocol.MessageTypeMSG:
			s.broadcastMessage(msgPayload, c)
		case protocol.MessageTypeWSP:
			s.sendWhisper(msgPayload, c)
		case protocol.MessageTypeBLCK_USR:
			//TODO: check if msgPayload.recipient (blockedUser) actually exist in users table
			if msgPayload.Content == "block" {
				err := s.blockUserManager.BlockUser(msgPayload.Sender, msgPayload.Recipient)
				if err != nil {
					log.Printf("Failed to block %v", err)
					s.sendSysResponse(c, fmt.Sprintf("Could not block %s due to an error", msgPayload.Recipient), "fail")
				}
				log.Printf("%s blocked %s", msgPayload.Sender, msgPayload.Recipient)
			}
			if msgPayload.Content == "unblock" {
				err := s.blockUserManager.UnblockUser(msgPayload.Sender, msgPayload.Recipient)
				if err != nil {
					log.Printf("Failed to unblock %v", err)
					s.sendSysResponse(c, fmt.Sprintf("Could not unblock %s due to an error", msgPayload.Recipient), "fail")
				}
				log.Printf("%s unblocked %s", msgPayload.Sender, msgPayload.Recipient)
			} else {
				log.Printf("Unknown block message received from %s\n", c.RemoteAddr().String())
			}

		case protocol.MessageTypeHSTRY:
			history, err := s.history.GetHistory(msgPayload.Sender, "MSG", "WSP")
			if err != nil {
				errMsg := s.encodeFn(protocol.Payload{
					MessageType: protocol.MessageTypeSYS,
					Content:     "Chat history not available",
					Status:      "fail",
				})
				c.Write([]byte(errMsg))
			}

			historyMsg := s.encodeFn(protocol.Payload{
				MessageType:        protocol.MessageTypeHSTRY,
				Sender:             msgPayload.Sender,
				EncodedChatHistory: history,
				Status:             "res",
			})
			log.Printf("Requested chat history length: %d", len(history))
			_, err = c.Write([]byte(historyMsg))
			if err != nil {
				log.Printf("failed to write history message: %v", err)
			}
		case protocol.MessageTypeACT_USRS:
			s.sendActiveUsers(c)
		default:
			log.Printf("Unknown message type received from %s\n", c.RemoteAddr().String())
		}
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

func (s *TCPServer) sendActiveUsers(conn net.Conn) {
	activeUsers := s.getActiveUsers()
	connectionInfo, ok := s.getConnectionInfo(conn)
	// If we can find the connectionInfo start excludingUser from activeList
	if ok {
		log.Printf("Logged in user %+v", connectionInfo)
		blockedUsers, err := s.blockUserManager.GetBlockedUsers(connectionInfo.OwnerName)
		if err != nil {
			s.sendSysResponse(conn, "Could not fetch blocked users. Blocked users will be able to message you", "fail")
			return
		}

		blockerUsers, err := s.blockUserManager.GetBlockerUsers(connectionInfo.OwnerName)
		if err != nil {
			s.sendSysResponse(conn, "Could not fetch blocked users. Blocked users will be able to message you:", "fail")
			return
		}

		activeUsers = pie.Filter(activeUsers, func(user string) bool {
			return !pie.Contains(blockedUsers, user) && !pie.Contains(blockerUsers, user)
		})
	}

	log.Printf("Sending active user list %s", activeUsers)
	msg := []byte(s.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeACT_USRS,
		ActiveUsers: activeUsers,
		Status:      "res",
	}))

	s.broadcastToAll(msg, "Error broadcasting active users", nil)
}

// sendWhisper looks up the recipient's connection in the connectionList. If found, it sends a whisper message to the recipient.
// If not found, it sends a system failure message back to the sender.
func (s *TCPServer) sendWhisper(msgPayload protocol.Payload, sender net.Conn) {
	// Look up the recipient's connection by their name in the connectionList
	recipientConn, found := s.findConnectionByOwnerName(msgPayload.Recipient)

	// If the recipient is not found or their connection is lost, send a system failure message to the sender
	if !found || recipientConn == nil {
		// Encode and send a system message indicating the recipient was not found or the connection was lost
		_, err := sender.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeSYS, Content: "Recipient not found or connection lost", Status: "fail"})))
		if err != nil {
			log.Println("Error sending recipient not found message:", err)
		}
		return
	}

	// If the recipient's connection is found, send the whisper message to the recipient
	_, err := recipientConn.Write([]byte(s.encodeFn(msgPayload)))
	if err != nil {
		log.Println("Error sending whisper:", err)
	}
}

// broadcastMessage sends a message to all connections except the sender
func (s *TCPServer) broadcastMessage(msgPayload protocol.Payload, sender net.Conn) {
	excludedUsers, err := getExcludedConnections(s, sender)
	if err != nil {
		s.sendSysResponse(sender, err.Error(), "fail")
	}
	msg := []byte(s.encodeFn(msgPayload))
	s.broadcastToAll(msg, "Error broadcasting message", excludedUsers...)
}

// This function gives us users who are excluded when broadcasting, whisper or sending active users.
// Mainly used for blocking logic.
func getExcludedConnections(s *TCPServer, sender net.Conn) ([]net.Conn, error) {
	var excludedConns []net.Conn
	senderInfo, ok := s.getConnectionInfo(sender)
	if !ok {
		return nil, fmt.Errorf("failed to get sender info")
	}

	blockerUsers, err := s.blockUserManager.GetBlockerUsers(senderInfo.OwnerName)
	if err != nil {
		return []net.Conn{sender}, fmt.Errorf("could not fetch blocked users. Blocked users will be able to message you: %w", err)
	}

	blockedUsers, err := s.blockUserManager.GetBlockedUsers(senderInfo.OwnerName)
	if err != nil {
		return []net.Conn{sender}, fmt.Errorf("could not fetch blocked users. Blocked users will be able to message you: %w", err)
	}

	blockedSet := make(map[string]struct{}, len(blockedUsers)+len(blockerUsers))
	for _, user := range append(blockedUsers, blockerUsers...) {
		blockedSet[user] = struct{}{}
	}

	s.connectionMap.Range(func(key, value any) bool {
		conn := key.(net.Conn)
		connInfo := value.(*ConnectionInfo)
		if _, blocked := blockedSet[connInfo.OwnerName]; blocked {
			excludedConns = append(excludedConns, conn)
		}
		return true
	})

	excludedConns = append(excludedConns, sender)
	return excludedConns, nil
}

// sendSysNotice sends a system notice to all connections except the sender.
// The notice informs about an action performed by the sender (e.g., joining or leaving the chat).
func (s *TCPServer) sendSysNotice(senderName string, sender net.Conn, action string) {
	blockedUsers, err := s.blockUserManager.GetBlockedUsers(senderName)
	if err != nil {
		s.sendSysResponse(sender, "Could not fetch blocked users. Blocked users will be able to message you", "fail")
		return
	}

	blockerUsers, err := s.blockUserManager.GetBlockerUsers(senderName)
	if err != nil {
		s.sendSysResponse(sender, "Could not fetch blocked users. Blocked users will be able to message you:", "fail")
		return
	}
	namesToExclude := append(blockedUsers, blockerUsers...)
	namesToExclude = pie.Unique(namesToExclude)

	var finalExcludedList []net.Conn
	for _, v := range namesToExclude {
		foundConn, ok := s.findConnectionByOwnerName(v)
		if !ok {
			continue
		}
		finalExcludedList = append(finalExcludedList, foundConn)
	}
	finalExcludedList = append(finalExcludedList, sender)

	msg := []byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeSYS, Content: fmt.Sprintf("%s has %s the chat.", senderName, action), Status: "success"}))
	s.broadcastToAll(msg, "Error sending system notice", finalExcludedList...)
}

// broadcastMessage sends a message to all connections except the sender
func (s *TCPServer) broadcastToAll(b []byte, errLog string, excludeConn ...net.Conn) {
	s.connectionMap.Range(func(key, value interface{}) bool {
		conn := key.(net.Conn)
		if !slices.Contains(excludeConn, conn) {
			_, err := conn.Write(b)
			if err != nil {
				log.Printf("%s %s\n", errLog, err)
			}
		}
		return true
	})
}

func (s *TCPServer) handleAuth(conn net.Conn) string {
	requiredMsg := s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Status: "required"})
	conn.Write([]byte(requiredMsg))
	connReader := bufio.NewReader(conn)
	var name string

	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading string: %v", err)
			break
		}

		payload, err := s.decodeFn(data)
		if err != nil {
			log.Printf("Failed to decode data: %s. Error: %v", data, err)
			s.sendAuthResponse(conn, "Invalid data format", "fail")
			continue
		}

		log.Printf("Login/register attempt for '%s' from %s", payload.Username, conn.RemoteAddr().String())

		// First, try to add the user (register)
		err = s.authManager.AddUser(payload.Username, payload.Password)
		if err == nil {
			// Registration successful
			name = payload.Username
			s.sendAuthResponse(conn, payload.Username, "success")
			break
		}

		// If registration failed, handle the specific error
		switch {
		case errors.Is(err, auth.ErrUserExists):
			// User exists, so this is a login attempt
			ok, authErr := s.authManager.AuthenticateUser(payload.Username, payload.Password)
			if authErr != nil {
				log.Printf("Authentication error: %v", authErr)
				s.sendAuthResponse(conn, "Invalid username or password", "fail")
			} else if ok {
				name = payload.Username
				s.sendAuthResponse(conn, payload.Username, "success")
				return name
			} else {
				s.sendAuthResponse(conn, "Invalid username or password", "fail")
			}
		case errors.Is(err, auth.ErrWeakPassword):
			s.sendAuthResponse(conn, "Password does not meet strength requirements", "fail")
		case errors.Is(err, auth.ErrInvalidUsername):
			s.sendAuthResponse(conn, "Username must be at least 2 characters long", "fail")
		default:
			log.Printf("Registration error: %s", err.Error())
			s.sendAuthResponse(conn, "Registration failed", "fail")
		}
	}
	return name
}

func (s *TCPServer) sendAuthResponse(conn net.Conn, message, status string) {
	conn.Write([]byte(s.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeUSR,
		Username:    message,
		Status:      status,
	})))
}

// Status is either fail or success
func (s *TCPServer) sendSysResponse(conn net.Conn, message, status string) {
	conn.Write([]byte(s.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeSYS,
		Content:     message,
		Status:      status,
	})))
}
