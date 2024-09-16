package server

import (
	"fmt"
	"log"
	"net"
	"slices"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/ogzhanolguncu/go-chat/server/internal/connection"
)

type MessageRouter struct {
	server *TCPServer
}

func NewMessageRouter(server *TCPServer) *MessageRouter {
	return &MessageRouter{
		server: server,
	}
}

// Main Message Routing
// -----------------------------

func (mr *MessageRouter) RouteMessage(info *connection.ConnectionInfo, message string) {
	payload, err := mr.server.decodeFn(message)
	if err != nil {
		mr.sendSysResponse(info.Connection, err.Error(), "fail")
		return
	}

	switch payload.MessageType {
	case protocol.MessageTypeCH:
		mr.handleChannelMessage(payload, info)
	case protocol.MessageTypeMSG:
		mr.handleGroupMessage(payload, info)
	case protocol.MessageTypeWSP:
		mr.handleWhisper(payload, info)
	case protocol.MessageTypeBLCK_USR:
		mr.handleBlockUser(payload, info)
	case protocol.MessageTypeHSTRY:
		mr.handleChatHistory(payload, info)
	case protocol.MessageTypeACT_USRS:
		mr.handleActiveUsers(info)
	default:
		log.Printf("Unknown message type received from %s\n", info.Connection.RemoteAddr().String())
	}
}

// Message Handlers
// -----------------------------
func (mr *MessageRouter) handleChannelMessage(payload protocol.Payload, info *connection.ConnectionInfo) {
	roomPayload, noticePayload := mr.server.channelManager.Handle(payload)
	payload.Timestamp = time.Now().Unix()
	payload.ChannelPayload = &roomPayload

	// Broadcast when kick, join, ban and close room happens
	go func() {
		noticePayloadCopy := payload
		noticePayloadCopy.ChannelPayload = &noticePayload

		if noticePayloadCopy.ChannelPayload.ChannelAction == protocol.NoticeChannel {
			roomMsg := []byte(mr.server.encodeFn(noticePayloadCopy))
			mr.broadcastToUsers(roomMsg, noticePayloadCopy.ChannelPayload.OptionalChannelArgs.Users, info.Connection)
		}
	}()

	// Broadcast newly created channel to users if visibility is public or closed channel
	go func() {
		isSuccess := payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusSuccess
		if payload.ChannelPayload.ChannelAction == protocol.CreateChannel && isSuccess &&
			payload.ChannelPayload.OptionalChannelArgs.Visibility == protocol.VisibilityPublic {
			mr.server.broadcastSystemNotice(fmt.Sprintf("Channel '%s' has been created by '%s'", payload.ChannelPayload.ChannelName, info.OwnerName), info.Connection)
		}
	}()

	if payload.ChannelPayload.ChannelAction == protocol.MessageChannel {
		roomMsg := []byte(mr.server.encodeFn(payload))
		mr.broadcastToUsers(roomMsg, payload.ChannelPayload.OptionalChannelArgs.Users, info.Connection)
		return
	}

	if payload.ChannelPayload.ChannelAction == protocol.KickUser || payload.ChannelPayload.ChannelAction == protocol.BanUser {
		// Fail cases in kickUser should be recieved by requester
		user := roomPayload.OptionalChannelArgs.TargetUser
		if payload.ChannelPayload.OptionalChannelArgs.Status == protocol.StatusFail {
			user = payload.ChannelPayload.Requester
		}
		userConn, found := mr.server.connectionManager.FindConnectionByOwnerName(user)
		//Skip if connection not found
		if !found {
			return
		}
		writeToAConn(mr, payload, userConn)
		return
	}

	writeToAConn(mr, payload, info.Connection)
}

func (mr *MessageRouter) handleGroupMessage(payload protocol.Payload, info *connection.ConnectionInfo) {
	excludedConns, err := mr.getExcludedConnections(info.Connection)
	if err != nil {
		mr.sendSysResponse(info.Connection, fmt.Sprintf("Error preparing message broadcast: %v", err), "fail")
		return
	}

	msg := []byte(mr.server.encodeFn(payload))
	mr.broadcastToAll(msg, "Error broadcasting message", excludedConns...)
}

func (mr *MessageRouter) handleWhisper(payload protocol.Payload, info *connection.ConnectionInfo) {
	recipientConn, found := mr.server.connectionManager.FindConnectionByOwnerName(payload.Recipient)
	if !found || recipientConn == nil {
		mr.sendSysResponse(info.Connection, "Recipient not found or connection lost", "fail")
		return
	}

	excludedConns, err := mr.getExcludedConnections(info.Connection)
	if err != nil {
		mr.sendSysResponse(info.Connection, fmt.Sprintf("Error preparing whisper message: %v", err), "fail")
		return
	}

	if !containsConnection(excludedConns, recipientConn) {
		_, err := recipientConn.Write([]byte(mr.server.encodeFn(payload)))
		if err != nil {
			log.Println("Error sending whisper:", err)
		}
	}
}

func (mr *MessageRouter) handleBlockUser(payload protocol.Payload, info *connection.ConnectionInfo) {
	if payload.Content == "block" {
		err := mr.server.blockUserManager.BlockUser(payload.Sender, payload.Recipient)
		if err != nil {
			log.Printf("Failed to block %v", err)
			mr.sendSysResponse(info.Connection, fmt.Sprintf("Could not block %s due to an error", payload.Recipient), "fail")
		}
		log.Printf("%s blocked %s", payload.Sender, payload.Recipient)
	} else if payload.Content == "unblock" {
		err := mr.server.blockUserManager.UnblockUser(payload.Sender, payload.Recipient)
		if err != nil {
			log.Printf("Failed to unblock %v", err)
			mr.sendSysResponse(info.Connection, fmt.Sprintf("Could not unblock %s due to an error", payload.Recipient), "fail")
		}
		log.Printf("%s unblocked %s", payload.Sender, payload.Recipient)
	} else {
		log.Printf("Unknown block message received from %s\n", info.Connection.RemoteAddr().String())
	}
}

func (mr *MessageRouter) handleChatHistory(payload protocol.Payload, info *connection.ConnectionInfo) {
	history, err := mr.server.historyManager.GetHistory(payload.Sender, "MSG", "WSP")
	if err != nil {
		mr.sendSysResponse(info.Connection, "Chat history not available", "fail")
		return
	}

	historyMsg := mr.server.encodeFn(protocol.Payload{
		MessageType:        protocol.MessageTypeHSTRY,
		Sender:             payload.Sender,
		EncodedChatHistory: history,
		Status:             "res",
	})
	log.Printf("Requested chat history length: %d", len(history))
	_, err = info.Connection.Write([]byte(historyMsg))
	if err != nil {
		log.Printf("failed to write history message: %v", err)
	}
}

func (mr *MessageRouter) handleActiveUsers(info *connection.ConnectionInfo) {
	mr.server.sendActiveUsers(info.Connection)
}

// User Filtering
// -----------------------------

// Finds connections to exclude when routing messages. This is used for filtering recipients based on block status and sender.
func (mr *MessageRouter) getExcludedConnections(sender net.Conn) ([]net.Conn, error) {
	senderInfo, ok := mr.server.connectionManager.GetConnectionInfo(sender)
	if !ok {
		return nil, fmt.Errorf("failed to get sender info")
	}

	blockedUsers, err := mr.server.blockUserManager.GetBlockedUsers(senderInfo.OwnerName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch blocked users: %w", err)
	}

	blockerUsers, err := mr.server.blockUserManager.GetBlockerUsers(senderInfo.OwnerName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch blocker users: %w", err)
	}

	namesToExclude := append(blockedUsers, blockerUsers...)
	namesToExclude = append(namesToExclude, senderInfo.OwnerName) // Exclude sender

	var excludedConns []net.Conn
	mr.server.connectionManager.RangeConnections(func(conn net.Conn, info *connection.ConnectionInfo) bool {
		if contains(namesToExclude, info.OwnerName) {
			excludedConns = append(excludedConns, conn)
		}
		return true
	})

	return excludedConns, nil
}

// Broadcasting Methods
// -----------------------------

// broadcastToAll sends a message to all connections except those in the exclude list
func (mr *MessageRouter) broadcastToAll(b []byte, errLog string, excludeConn ...net.Conn) {
	mr.server.connectionManager.RangeConnections(func(conn net.Conn, _ *connection.ConnectionInfo) bool {
		if !containsConnection(excludeConn, conn) {
			_, err := conn.Write(b)
			if err != nil {
				log.Printf("%s %s\n", errLog, err)
			}
		}
		return true
	})
}

func (mr *MessageRouter) broadcastToUsers(b []byte, users []string, excludeConn ...net.Conn) {
	mr.server.connectionManager.RangeConnections(func(conn net.Conn, details *connection.ConnectionInfo) bool {
		// Exclude initator from broadcast and make sure user is in connection list
		if !containsConnection(excludeConn, conn) && slices.Contains(users, details.OwnerName) {
			_, err := conn.Write(b)
			if err != nil {
				log.Printf("%s %s\n", "Couldn't send broadcast message", err)
			}
		}
		return true
	})
}

// sendSysResponse sends a system response message to a specific connection
func (mr *MessageRouter) sendSysResponse(conn net.Conn, message, status string) {
	conn.Write([]byte(mr.server.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeSYS,
		Content:     message,
		Status:      status,
	})))
}

// Helper Functions
// -----------------------------

// containsConnection checks if a given connection is present in a slice of connections
func containsConnection(slice []net.Conn, conn net.Conn) bool {
	for _, v := range slice {
		if v == conn {
			return true
		}
	}
	return false
}

// Write to a connection mostly used for channel messages
func writeToAConn(mr *MessageRouter, payload protocol.Payload, userConn net.Conn) {
	roomMsg := []byte(mr.server.encodeFn(payload))
	_, err := userConn.Write([]byte(roomMsg))
	if err != nil {
		log.Printf("failed to write history message: %v", err)
	}
}
