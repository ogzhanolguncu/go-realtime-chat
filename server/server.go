package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	protocol "github.com/ogzhanolguncu/go-chat/protocol"
)

// ConnectionInfo holds connection-related information.
type ConnectionInfo struct {
	Connection net.Conn
	OwnerName  string
}

type TCPServer struct {
	connectionMap sync.Map
}

func newServer() *TCPServer {
	return &TCPServer{}
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
	value, ok := s.connectionMap.LoadAndDelete(c)
	if !ok {
		return nil, false
	}
	info, ok := value.(*ConnectionInfo)
	return info, ok
}

func (s *TCPServer) handleNewConnection(c net.Conn) {
	name := s.handleUsernameSet(c)

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

	s.sendSystemNotice(name, c, "joined")
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
	defer func() {
		info, _ := s.getConnectionInfoAndDelete(c)
		if info != nil {
			s.sendSystemNotice(info.OwnerName, nil, "left")
		}
		c.Close()
	}()

	connReader := bufio.NewReader(c)

	for {
		data, err := connReader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", c.RemoteAddr().String(), err)
			break
		}

		rawMessage := strings.TrimSpace(data)
		log.Printf("Message from %s: %s\n", c.RemoteAddr().String(), rawMessage)

		msgPayload, err := protocol.DecodeMessage(rawMessage)
		if err != nil {
			// Write back to client that their message is malformed
			c.Write([]byte(err.Error()))
		}

		switch msgPayload.MessageType {
		case protocol.MessageTypeMSG:
			s.broadcastMessage(msgPayload, c)
		case protocol.MessageTypeWSP:
			s.sendWhisper(msgPayload, c)
		case protocol.MessageTypeACT_USRS:
			activeUsers := s.getActiveUsers()
			msg := []byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeACT_USRS, ActiveUsers: activeUsers, Status: "res"}))
			c.Write(msg)
		default:
			log.Printf("Unknown message type received from %s\n", c.RemoteAddr().String())
		}
	}
	log.Printf("Connection closed for %s\n", c.RemoteAddr().String())
}

// sendWhisper looks up the recipient's connection in the connectionList. If found, it sends a whisper message to the recipient.
// If not found, it sends a system failure message back to the sender.
func (s *TCPServer) sendWhisper(msgPayload protocol.Payload, sender net.Conn) {
	// Look up the recipient's connection by their name in the connectionList
	recipientConn, found := s.findConnectionByOwnerName(msgPayload.Recipient)

	// If the recipient is not found or their connection is lost, send a system failure message to the sender
	if !found || recipientConn == nil {
		// Encode and send a system message indicating the recipient was not found or the connection was lost
		_, err := sender.Write([]byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeSYS, Content: "Recipient not found or connection lost", Status: "fail"})))
		if err != nil {
			log.Println("Error sending recipient not found message:", err)
		}
		return
	}

	// If the recipient's connection is found, send the whisper message to the recipient
	_, err := recipientConn.Write([]byte(protocol.EncodeMessage(msgPayload)))
	if err != nil {
		log.Println("Error sending whisper:", err)
	}
}

// broadcastMessage sends a message to all connections except the sender
func (s *TCPServer) broadcastMessage(msgPayload protocol.Payload, sender net.Conn) {
	msg := []byte(protocol.EncodeMessage(msgPayload))
	s.broadcastToAll(msg, "Error broadcasting message", sender)
}

// sendSystemNotice sends a system notice to all connections except the sender.
// The notice informs about an action performed by the sender (e.g., joining or leaving the chat).
func (s *TCPServer) sendSystemNotice(senderName string, sender net.Conn, action string) {
	msg := []byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeSYS, Content: fmt.Sprintf("%s has %s the chat.", senderName, action), Status: "success"}))
	s.broadcastToAll(msg, "Error sending system notice", sender)
}

// broadcastMessage sends a message to all connections except the sender
func (s *TCPServer) broadcastToAll(b []byte, errLog string, excludeConn net.Conn) {
	s.connectionMap.Range(func(key, value interface{}) bool {
		conn := key.(net.Conn)
		if conn != excludeConn {
			_, err := conn.Write(b)
			if err != nil {
				log.Printf("%s %s\n", errLog, err)
			}
		}
		return true
	})
}

func (s *TCPServer) handleUsernameSet(conn net.Conn) string {
	requiredMsg := protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeUSR, Status: "required"})
	conn.Write([]byte(requiredMsg))
	connReader := bufio.NewReader(conn)

	var name string

	for {
		data, err := connReader.ReadString('\n')

		if err != nil {
			break
		}
		payload, err := protocol.DecodeMessage(data)
		_, nameIsAlreadyInUse := s.findConnectionByOwnerName(payload.Username)
		if err != nil || len(payload.Username) < 2 {
			conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: fmt.Sprintf("Username '%s' cannot be empty or less than two characters.", payload.Username), Status: "fail"})))
		} else if nameIsAlreadyInUse {
			conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: fmt.Sprintf("Username '%s' is already in use.", payload.Username), Status: "fail"})))
		} else {
			name = payload.Username
			conn.Write([]byte(protocol.EncodeMessage(protocol.Payload{MessageType: protocol.MessageTypeUSR, Username: payload.Username, Status: "success"})))
			break
		}
	}
	return name
}
