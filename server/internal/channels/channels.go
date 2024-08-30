package channels

import (
	"slices"
	"sync"
	"time"

	"github.com/ogzhanolguncu/go-chat/protocol"
	"github.com/sirupsen/logrus"
)

const (
	chAlreadyExists     = "Channel name already exists."
	chDoesNotExist      = "Channel does not exist."
	incorrectChPassword = "Incorrect channel password."
	chAtCapacity        = "Channel is full. Try again later."
	notInTheCh          = "User not in the channel."
	unknownChAction     = "Unknown channel action."
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

type ChannelDetails struct {
	ChName       string
	ChPass       string
	ChSize       int
	Owner        string
	Users        []string
	LastActivity int64
	BannedUsers  []string
	Visibility   string
}

type Manager struct {
	chMap map[string]*ChannelDetails
	lock  sync.RWMutex
}

func NewChannelManager() *Manager {
	logger.Info("Initializing new ChannelManager")
	return &Manager{
		chMap: make(map[string]*ChannelDetails),
	}
}

func (m *Manager) Handle(payload protocol.Payload) protocol.ChannelPayload {
	logger.WithFields(logrus.Fields{
		"action":  payload.ChannelPayload.ChannelAction,
		"channel": payload.ChannelPayload.ChannelName,
		"user":    payload.ChannelPayload.Requester,
	}).Info("Handling channel action")

	switch payload.ChannelPayload.ChannelAction {
	case protocol.CreateChannel:
		return m.createChannel(*payload.ChannelPayload)
	case protocol.JoinChannel:
		return m.joinChannel(*payload.ChannelPayload)
	case protocol.LeaveChannel:
		return m.leaveChannel(*payload.ChannelPayload)
	case protocol.GetChannels:
		return m.getChannels(*payload.ChannelPayload)
	case protocol.GetUsers:
		return m.getUsers(*payload.ChannelPayload)
	case protocol.MessageChannel:
		return m.messageChannel(*payload.ChannelPayload)
	default:
		logger.WithField("action", payload.ChannelPayload.ChannelAction).Warn("Unknown channel action")
		return protocol.ChannelPayload{
			OptionalChannelArgs: &protocol.OptionalChannelArgs{
				Status: protocol.StatusFail,
				Reason: unknownChAction,
			},
		}
	}
}

func (m *Manager) createChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"owner":   chPayload.Requester,
	}).Info("Attempting to create channel")

	if _, exists := m.chMap[chPayload.ChannelName]; exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel already exists")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chAlreadyExists,
		}
		return chPayload
	}

	m.chMap[chPayload.ChannelName] = &ChannelDetails{
		ChName:       chPayload.ChannelName,
		ChPass:       chPayload.ChannelPassword,
		Owner:        chPayload.Requester,
		ChSize:       chPayload.ChannelSize,
		Users:        []string{chPayload.Requester},
		LastActivity: time.Now().Unix(),
		Visibility:   string(chPayload.OptionalChannelArgs.Visibility),
	}

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"owner":   chPayload.Requester,
	}).Info("channel created successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return chPayload
}

func (m *Manager) joinChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("Attempting to join channel")

	// Missing channel
	channel, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}

	// Wrong password
	if channel.ChPass != chPayload.ChannelPassword {
		logger.WithFields(logrus.Fields{
			"channel": chPayload.ChannelName,
			"user":    chPayload.Requester,
		}).Warn("Incorrect channel password")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: incorrectChPassword,
		}
		return chPayload
	}

	// Channel is full
	if len(channel.Users) >= channel.ChSize {
		logger.WithFields(logrus.Fields{
			"channel": chPayload.ChannelName,
			"user":    chPayload.Requester,
		}).Warn("Channel at capacity")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chAtCapacity,
		}
		return chPayload
	}

	channel.Users = append(channel.Users, chPayload.Requester)
	channel.LastActivity = time.Now().Unix()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("User joined channel successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return chPayload
}

func (m *Manager) leaveChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("Attempting to leave channel")

	channel, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}

	userIndex := slices.Index(channel.Users, chPayload.Requester)
	if userIndex == -1 {
		logger.WithFields(logrus.Fields{
			"channel": chPayload.ChannelName,
			"user":    chPayload.Requester,
		}).Warn("User not in the channel")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: notInTheCh,
		}
		return chPayload
	}

	channel.Users = slices.Delete(channel.Users, userIndex, userIndex+1)
	channel.LastActivity = time.Now().Unix()

	if len(channel.Users) == 0 {
		delete(m.chMap, chPayload.ChannelName)
		logger.WithField("channel", chPayload.ChannelName).Info("Channel deleted as it's empty")
	}

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("User left channel successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return chPayload
}

func (m *Manager) getChannels(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	logger.Info("Getting list of channels")

	channels := make([]string, 0, len(m.chMap))
	for channelName := range m.chMap {
		channels = append(channels, channelName)
	}

	logger.WithField("channelCount", len(channels)).Info("Channel list retrieved")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status:   protocol.StatusSuccess,
		Channels: channels,
	}
	return chPayload
}

func (m *Manager) getUsers(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	logger.Info("Getting list of users")
	selectedCh, exists := m.chMap[chPayload.ChannelName]
	//Missing channel check
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}
	users := make([]string, 0, len(selectedCh.Users))
	logger.WithField("userCount", len(users)).Info("User list retrieved")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
		Users:  users,
	}
	return chPayload
}

func (m *Manager) messageChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	//Check if requesting user is in the channel
	found := slices.Contains(m.chMap[chPayload.ChannelName].Users, chPayload.Requester)
	if !found {
		logger.WithField("channel", chPayload.ChannelName).Warn("User not in the channel")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: notInTheCh,
		}
		return chPayload
	}

	logger.Info("Getting list of users")
	selectedCh, exists := m.chMap[chPayload.ChannelName]
	//Missing channel check
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}
	users := make([]string, 0, len(selectedCh.Users))
	logger.WithField("userCount", len(users)).Info("User list retrieved")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
		Users:  users,
	}
	return chPayload
}
