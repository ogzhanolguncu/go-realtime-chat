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

func (m *Manager) Handle(payload protocol.Payload) protocol.RoomPayload {
	logger.WithFields(logrus.Fields{
		"action": payload.RoomPayload.RoomAction,
		"room":   payload.RoomPayload.RoomName,
		"user":   payload.RoomPayload.Requester,
	}).Info("Handling room action")

	switch payload.RoomPayload.RoomAction {
	case protocol.CreateRoom:
		return m.createRoom(*payload.RoomPayload)
	case protocol.JoinRoom:
		return m.joinRoom(*payload.RoomPayload)
	case protocol.LeaveRoom:
		return m.leaveRoom(*payload.RoomPayload)
	case protocol.GetRooms:
		return m.getRooms(*payload.RoomPayload)
	default:
		logger.WithField("action", payload.RoomPayload.RoomAction).Warn("Unknown room action")
		return protocol.RoomPayload{
			OptionalRoomArgs: &protocol.OptionalRoomArgs{
				Status: protocol.StatusFail,
				Reason: unknownRoomAction,
			},
		}
	}
}

func (m *Manager) createRoom(roomPayload protocol.RoomPayload) protocol.RoomPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room":  roomPayload.RoomName,
		"owner": roomPayload.Requester,
	}).Info("Attempting to create room")

	if _, exists := m.roomMap[roomPayload.RoomName]; exists {
		logger.WithField("room", roomPayload.RoomName).Warn("Room already exists")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: roomAlreadyExists,
		}
		return roomPayload
	}

	m.roomMap[roomPayload.RoomName] = &RoomDetails{
		RoomName:     roomPayload.RoomName,
		RoomPassword: roomPayload.RoomPassword,
		Owner:        roomPayload.Requester,
		RoomSize:     roomPayload.RoomSize,
		Users:        []string{roomPayload.Requester},
		LastActivity: time.Now().Unix(),
		Visibility:   string(roomPayload.OptionalRoomArgs.Visibility),
	}

	logger.WithFields(logrus.Fields{
		"room":  roomPayload.RoomName,
		"owner": roomPayload.Requester,
	}).Info("Room created successfully")

	roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) joinRoom(roomPayload protocol.RoomPayload) protocol.RoomPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.RoomName,
		"user": roomPayload.Requester,
	}).Info("Attempting to join room")

	room, exists := m.roomMap[roomPayload.RoomName]
	if !exists {
		logger.WithField("room", roomPayload.RoomName).Warn("Room does not exist")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: roomDoesNotExist,
		}
		return roomPayload
	}

	if room.RoomPassword != roomPayload.RoomPassword {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.RoomName,
			"user": roomPayload.Requester,
		}).Warn("Incorrect room password")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: incorrectRoomPassword,
		}
		return roomPayload
	}

	if len(room.Users) >= room.RoomSize {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.RoomName,
			"user": roomPayload.Requester,
		}).Warn("Room at capacity")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: roomAtCapacity,
		}
		return roomPayload
	}

	room.Users = append(room.Users, roomPayload.Requester)
	room.LastActivity = time.Now().Unix()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.RoomName,
		"user": roomPayload.Requester,
	}).Info("User joined room successfully")

	roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) leaveRoom(roomPayload protocol.RoomPayload) protocol.RoomPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"room": roomPayload.RoomName,
		"user": roomPayload.Requester,
	}).Info("Attempting to leave room")

	room, exists := m.roomMap[roomPayload.RoomName]
	if !exists {
		logger.WithField("room", roomPayload.RoomName).Warn("Room does not exist")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: roomDoesNotExist,
		}
		return roomPayload
	}

	userIndex := slices.Index(room.Users, roomPayload.Requester)
	if userIndex == -1 {
		logger.WithFields(logrus.Fields{
			"room": roomPayload.RoomName,
			"user": roomPayload.Requester,
		}).Warn("User not in the room")
		roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
			Status: protocol.StatusFail,
			Reason: notInTheRoom,
		}
		return roomPayload
	}

	room.Users = slices.Delete(room.Users, userIndex, userIndex+1)
	room.LastActivity = time.Now().Unix()

	if len(room.Users) == 0 {
		delete(m.roomMap, roomPayload.RoomName)
		logger.WithField("room", roomPayload.RoomName).Info("Room deleted as it's empty")
	}

	logger.WithFields(logrus.Fields{
		"room": roomPayload.RoomName,
		"user": roomPayload.Requester,
	}).Info("User left room successfully")

	roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
		Status: protocol.StatusSuccess,
	}
	return roomPayload
}

func (m *Manager) getRooms(roomPayload protocol.RoomPayload) protocol.RoomPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	logger.Info("Getting list of rooms")

	rooms := make([]string, 0, len(m.roomMap))
	for roomName := range m.roomMap {
		rooms = append(rooms, roomName)
	}

	logger.WithField("roomCount", len(rooms)).Info("Room list retrieved")

	roomPayload.OptionalRoomArgs = &protocol.OptionalRoomArgs{
		Status: protocol.StatusSuccess,
		Rooms:  rooms,
	}
	return roomPayload
}
