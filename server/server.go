package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	protocol "github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/auth"
	"github.com/ogzhanolguncu/go-chat/server/chat_history"
	"github.com/ogzhanolguncu/go-chat/server/utils"
)

// ConnectionInfo holds connection-related information.
const maxMessageLimit = 200
const groupKey = "SuperSecretGroupKey"

type ConnectionInfo struct {
	Connection   net.Conn
	OwnerName    string
	blockedUsers []net.Conn
}

type TCPServer struct {
	connectionMap sync.Map
	history       chat_history.ChatHistory
	authManager   *auth.AuthManager
	groupKey      string
	encodeFn      func(payload protocol.Payload) string
	decodeFn      func(message string) (protocol.Payload, error)
}

func newServer() *TCPServer {
	// Generate a 32-byte key
	key, err := generateSecureKey(32)
	if err != nil {
		log.Printf("Failed to create secure key moving forward with hardcoded key: %v", err)
		key = groupKey
	}

	encoding := flag.Bool("encoding", false, "enable encoding")
	flag.Parse()

	var encodingType string
	if *encoding {
		encodingType = "BASE64"
	} else {
		encodingType = "PLAIN-TEXT"
	}

	log.Printf("------ ENCODING SET TO %s ------", encodingType)

	authManager, err := auth.NewAuthManager(fmt.Sprintf(utils.RootDir() + "/chat.db"))
	if err != nil {
		log.Printf("Failed to initialize auth manager: %v", err)
	}

	return &TCPServer{
		groupKey:    key,
		decodeFn:    protocol.InitDecodeProtocol(*encoding),
		encodeFn:    protocol.InitEncodeProtocol(*encoding),
		history:     *chat_history.NewChatHistory(*encoding),
		authManager: authManager,
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
		s.sendSystemNotice(name, c, "joined")

	}()

	go func() {
		defer wg.Done()
		s.sendActiveUsers()
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
	var stopInterval chan bool
	defer func() {
		//Cleanup for chat history save disk
		if stopInterval != nil {
			stopInterval <- true
		}
		info, _ := s.getConnectionInfoAndDelete(c)
		if info != nil {
			s.sendSystemNotice(info.OwnerName, nil, "left")
			s.sendActiveUsers()
		}
		c.Close()
	}()

	//Save chat history to disk every five second.
	stopInterval = setInterval(func() {
		s.history.SaveToDisk(maxMessageLimit)
	}, 5*time.Second)

	connReader := bufio.NewReader(c)

	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}

		rawMessage := strings.TrimSpace(data)
		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), rawMessage)
		s.history.AddMessage(rawMessage)

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
			blocker, _ := s.connectionMap.Load(c)

			blockerUser := blocker.(*ConnectionInfo)
			blockedUser, ok := s.findConnectionByOwnerName(msgPayload.Recipient)
			if !ok {
				c.Write([]byte(s.encodeFn(protocol.Payload{
					MessageType: protocol.MessageTypeSYS,
					Content:     "Recipient not found or connection lost",
					Status:      "fail",
				})))
				continue
			}

			if msgPayload.Content == "block" {
				blockerUser.blockedUsers = append(blockerUser.blockedUsers, blockedUser)
				s.connectionMap.Store(c, blockerUser)
				log.Printf("%s blocked %s", msgPayload.Sender, msgPayload.Recipient)
			}
			if msgPayload.Content == "unblock" {
				blockerUser.blockedUsers = slices.DeleteFunc(blockerUser.blockedUsers, func(u net.Conn) bool {
					return u == blockedUser
				})
				s.connectionMap.Store(c, blockerUser)
				log.Printf("%s unblocked %s", msgPayload.Sender, msgPayload.Recipient)
			} else {
				log.Printf("Unknown block message received from %s\n", c.RemoteAddr().String())
			}

		case protocol.MessageTypeHSTRY:
			msg := []byte(s.encodeFn(protocol.Payload{
				MessageType:        protocol.MessageTypeHSTRY,
				Sender:             msgPayload.Sender,
				EncodedChatHistory: s.history.GetHistory(msgPayload.Sender, "MSG", "WSP"),
				Status:             "res"}))
			c.Write(msg)
		case protocol.MessageTypeACT_USRS:
			s.sendActiveUsers()
		case protocol.MessageTypeENC:
			s.sendEncryptionKey(msgPayload, c)
		default:
			log.Printf("Unknown message type received from %s\n", c.RemoteAddr().String())
		}
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

func (s *TCPServer) sendActiveUsers() {
	activeUsers := s.getActiveUsers()
	log.Printf("Sending active user list %s", activeUsers)
	msg := []byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeACT_USRS, ActiveUsers: activeUsers, Status: "res"}))
	s.broadcastToAll(msg, "Error broadcasting active users", nil)
}

func (s *TCPServer) sendEncryptionKey(msgPayload protocol.Payload, c net.Conn) {
	usersPublicKey, err := stringToPublicKey(msgPayload.EncryptedKey)
	if err != nil {
		log.Printf("Could not decode users public key, closing the connection: %v", err)
		c.Close()
		return
	}

	groupChatKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, usersPublicKey, []byte(s.groupKey), nil)
	if err != nil {
		log.Printf("Could not encrypt group chat key using users public key, closing the connection: %v", err)
		c.Close()
		return
	}

	base64EncryptedKey := base64.StdEncoding.EncodeToString(groupChatKey)
	msg := []byte(s.encodeFn(protocol.Payload{
		MessageType:  protocol.MessageTypeENC,
		EncryptedKey: base64EncryptedKey,
	}))

	_, err = c.Write(msg)
	if err != nil {
		log.Printf("Failed to send encrypted group chat key: %v", err)
		c.Close()
		return
	}
	log.Printf("Successfully sent encrypted group chat key to client")
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
	excludedUsers := getExcludedUsers(s, sender)
	msg := []byte(s.encodeFn(msgPayload))
	s.broadcastToAll(msg, "Error broadcasting message", excludedUsers...)
}

// This function gives us users who are excluded when broadcasting, whisper or sending active users.
// Mainly used for blocking logic.
func getExcludedUsers(s *TCPServer, sender net.Conn) []net.Conn {
	var excludedUsers []net.Conn
	senderInfo, _ := s.getConnectionInfo(sender)

	s.connectionMap.Range(func(key, value any) bool {
		conn, ok := key.(net.Conn)
		if !ok {
			fmt.Printf("Unexpected key type in connectionMap: %T\n", key)
			return true
		}

		info, ok := value.(*ConnectionInfo)
		if !ok {
			fmt.Printf("Unexpected value type in connectionMap: %T\n", value)
			return true
		}

		if slices.Contains(senderInfo.blockedUsers, conn) {
			excludedUsers = append(excludedUsers, conn)
		}

		if slices.Contains(info.blockedUsers, sender) {
			excludedUsers = append(excludedUsers, conn)
		}

		return true
	})

	excludedUsers = append(excludedUsers, sender)
	return excludedUsers
}

// sendSystemNotice sends a system notice to all connections except the sender.
// The notice informs about an action performed by the sender (e.g., joining or leaving the chat).
func (s *TCPServer) sendSystemNotice(senderName string, sender net.Conn, action string) {
	msg := []byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeSYS, Content: fmt.Sprintf("%s has %s the chat.", senderName, action), Status: "success"}))
	s.broadcastToAll(msg, "Error sending system notice", sender)
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
			conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: "Invalid data format", Status: "fail"})))
			continue
		}

		log.Printf("Login/register attempt for '%s' from %s", payload.Username, conn.RemoteAddr().String())
		err = s.authManager.AddUser(payload.Username, payload.Password)
		if err == nil {
			name = payload.Username
			conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: payload.Username, Status: "success"})))
			break
		}

		if err.Error() == "username already exists" {
			ok, authErr := s.authManager.AuthenticateUser(payload.Username, payload.Password)
			if authErr != nil {
				conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: authErr.Error(), Status: "fail"})))
			} else if ok {
				name = payload.Username
				conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: payload.Username, Status: "success"})))
				break
			} else {
				conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: "Invalid credentials", Status: "fail"})))
			}
		} else {
			errMsg := strings.ToUpper(err.Error()[:1]) + err.Error()[1:]
			conn.Write([]byte(s.encodeFn(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: errMsg, Status: "fail"})))
		}
	}
	return name
}

func setInterval(callback func(), interval time.Duration) chan bool {
	// Create a channel to signal the interval to stop
	stop := make(chan bool)

	go func() {
		for {
			select {
			case <-time.After(interval):
				callback()
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func generateSecureKey(keyLength int) (string, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	strKey := base64.StdEncoding.EncodeToString(key)
	log.Printf("Generating group chat key %s", strKey)

	return strKey, nil
}

func stringToPublicKey(keyStr string) (*rsa.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPubKey, nil
}
