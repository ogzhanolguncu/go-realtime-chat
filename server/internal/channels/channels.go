package channels

import (
	"slices"
	"sync"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/sirupsen/logrus"
)

const (
	roomAlreadyExists     = "Room name already exists."
	roomDoesNotExist      = "Room does not exist."
	incorrectRoomPassword = "Incorrect room password."
	roomAtCapacity        = "Room is full. Try again later."
	notInTheRoom          = "User not in the room."
	unknownRoomAction     = "Unknown room action."
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

type RoomDetails struct {
	RoomName     string
	RoomPassword string
	RoomSize     int
	Owner        string
	Users        []string
	LastActivity int64
	BannedUsers  []string
	Visibility   string
}

type Manager struct {
	roomMap map[string]*RoomDetails
	lock    sync.RWMutex
}

func NewChannelManager() *Manager {
	logger.Info("Initializing new ChannelManager")
	return &Manager{
		roomMap: make(map[string]*RoomDetails),
	}
}

func (m *Manager) Handle(payload protocol.Payload) protocol.ChannelPayload {
	logger.WithFields(logrus.Fields{
		"action": payload.ChannelPayload.ChannelAction,
		"room":   payload.ChannelPayload.ChannelName,
		"user":   payload.ChannelPayload.Requester,
	}).Info("Handling room action")

	switch payload.ChannelPayload.ChannelAction {
	case protocol.CreateChannel:
		return m.createRoom(*payload.ChannelPayload)
	case protocol.JoinChannel:
		return m.joinRoom(*payload.ChannelPayload)
	case protocol.LeaveChannel:
		return m.leaveRoom(*payload.ChannelPayload)
	case protocol.GetChannels:
		return m.getRooms(*payload.ChannelPayload)
	default:
		logger.WithField("action", payload.ChannelPayload.ChannelAction).Warn("Unknown room action")
		return protocol.ChannelPayload{
			OptionalChannelArgs: &protocol.OptionalChannelArgs{
				Status: protocol.StatusFail,
				Reason: unknownRoomAction,
			},
		}
	}
}

func (m *Manager) createRoom(roomPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room":  roomPayload.ChannelName,
		"owner": roomPayload.Requester,
	}).Info("Attempting to create room")

	if _, exists := m.roomMap[roomPayload.ChannelName]; exists {
		logger.WithField("room", roomPayload.ChannelName).Warn("Room already exists")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: roomAlreadyExists,
		}
		return roomPayload
	}

	m.roomMap[roomPayload.ChannelName] = &RoomDetails{
		RoomName:     roomPayload.ChannelName,
		RoomPassword: roomPayload.ChannelPassword,
		Owner:        roomPayload.Requester,
		RoomSize:     roomPayload.ChannelSize,
		Users:        []string{roomPayload.Requester},
		LastActivity: time.Now().Unix(),
		Visibility:   string(roomPayload.OptionalChannelArgs.Visibility),
	}

	logger.WithFields(logrus.Fields{
		"room":  roomPayload.ChannelName,
		"owner": roomPayload.Requester,
	}).Info("Room created successfully")

	roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) joinRoom(roomPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.ChannelName,
		"user": roomPayload.Requester,
	}).Info("Attempting to join room")

	room, exists := m.roomMap[roomPayload.ChannelName]
	if !exists {
		logger.WithField("room", roomPayload.ChannelName).Warn("Room does not exist")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: roomDoesNotExist,
		}
		return roomPayload
	}

	if room.RoomPassword != roomPayload.ChannelPassword {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.ChannelName,
			"user": roomPayload.Requester,
		}).Warn("Incorrect room password")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: incorrectRoomPassword,
		}
		return roomPayload
	}

	if len(room.Users) >= room.RoomSize {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.ChannelName,
			"user": roomPayload.Requester,
		}).Warn("Room at capacity")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: roomAtCapacity,
		}
		return roomPayload
	}

	room.Users = append(room.Users, roomPayload.Requester)
	room.LastActivity = time.Now().Unix()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.ChannelName,
		"user": roomPayload.Requester,
	}).Info("User joined room successfully")

	roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) leaveRoom(roomPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.ChannelName,
		"user": roomPayload.Requester,
	}).Info("Attempting to leave room")

	room, exists := m.roomMap[roomPayload.ChannelName]
	if !exists {
		logger.WithField("room", roomPayload.ChannelName).Warn("Room does not exist")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: roomDoesNotExist,
		}
		return roomPayload
	}

	userIndex := slices.Index(room.Users, roomPayload.Requester)
	if userIndex == -1 {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.ChannelName,
			"user": roomPayload.Requester,
		}).Warn("User not in the room")
		roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: notInTheRoom,
		}
		return roomPayload
	}

	room.Users = slices.Delete(room.Users, userIndex, userIndex+1)
	room.LastActivity = time.Now().Unix()

	if len(room.Users) == 0 {
		delete(m.roomMap, roomPayload.ChannelName)
		logger.WithField("room", roomPayload.ChannelName).Info("Room deleted as it's empty")
	}

	logger.WithFields(logrus.Fields{
		"room": roomPayload.ChannelName,
		"user": roomPayload.Requester,
	}).Info("User left room successfully")

	roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) getRooms(roomPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	logger.Info("Getting list of rooms")

	rooms := make([]string, 0, len(m.roomMap))
	for roomName := range m.roomMap {
		rooms = append(rooms, roomName)
	}

	logger.WithField("roomCount", len(rooms)).Info("Room list retrieved")

	roomPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status:   protocol.StatusSuccess,
		Channels: rooms,
	}
	return roomPayload
}
