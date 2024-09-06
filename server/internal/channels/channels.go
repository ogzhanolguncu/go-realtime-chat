package channels

import (
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
	noActiveChannels    = "No active channels"
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
	ChCapacity   int
	Owner        string
	Users        map[string]struct{}
	LastActivity int64
	BannedUsers  map[string]struct{}
	Visibility   string
}

type Manager struct {
	chMap   map[string]*ChannelDetails
	chLocks map[string]*sync.RWMutex
	lock    sync.RWMutex
}

type ManagerConfig struct {
	CleanupInterval     time.Duration
	InactivityThreshold time.Duration
}

func NewChannelManager(config ManagerConfig) *Manager {
	logger.Info("Initializing new ChannelManager")

	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute // default cleanup interval
	}
	if config.InactivityThreshold == 0 {
		config.InactivityThreshold = 24 * time.Hour // default inactivity threshold
	}

	manager := &Manager{
		chMap:   make(map[string]*ChannelDetails),
		chLocks: make(map[string]*sync.RWMutex),
	}

	go manager.startCleanupRoutine(config.CleanupInterval, config.InactivityThreshold)

	return manager
}

func (m *Manager) startCleanupRoutine(cleanupInterval, inactivityThreshold time.Duration) {
	logger.WithFields(logrus.Fields{
		"cleanupInterval":     cleanupInterval,
		"inactivityThreshold": inactivityThreshold,
	}).Info("Starting cleanup routine")

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupInactiveChannels(inactivityThreshold)
	}
}

func (m *Manager) cleanupInactiveChannels(inactivityThreshold time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	now := time.Now().Unix()
	for channelName, channel := range m.chMap {
		if now-channel.LastActivity > int64(inactivityThreshold.Seconds()) {
			delete(m.chMap, channelName)
			delete(m.chLocks, channelName)
			logger.WithField("channel", channelName).Info("Removed inactive channel")
		}
	}
}

func (m *Manager) getChannelLock(channelName string) *sync.RWMutex {
	m.lock.Lock()
	defer m.lock.Unlock()

	if lock, exists := m.chLocks[channelName]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	m.chLocks[channelName] = lock
	return lock
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
		ChCapacity:   chPayload.ChannelSize,
		Users:        map[string]struct{}{chPayload.Requester: {}},
		LastActivity: time.Now().Unix(),
		BannedUsers:  make(map[string]struct{}),
		Visibility:   string(chPayload.OptionalChannelArgs.Visibility),
	}

	m.chLocks[chPayload.ChannelName] = &sync.RWMutex{}

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
	chLock := m.getChannelLock(chPayload.ChannelName)
	chLock.Lock()
	defer chLock.Unlock()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("Attempting to join channel")

	channel, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}

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

	if len(channel.Users) >= channel.ChCapacity {
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

	channel.Users[chPayload.Requester] = struct{}{}
	channel.LastActivity = time.Now().Unix()

	logger.WithFields(logrus.Fields{
		"channel":        chPayload.ChannelName,
		"user":           chPayload.Requester,
		"channelDetails": channel,
	}).Info("User joined channel successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}
	return chPayload
}

func (m *Manager) leaveChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	chLock := m.getChannelLock(chPayload.ChannelName)
	chLock.Lock()
	defer chLock.Unlock()

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

	if _, exists := channel.Users[chPayload.Requester]; !exists {
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

	delete(channel.Users, chPayload.Requester)
	channel.LastActivity = time.Now().Unix()

	if len(channel.Users) == 0 {
		m.lock.Lock()
		delete(m.chMap, chPayload.ChannelName)
		delete(m.chLocks, chPayload.ChannelName)
		m.lock.Unlock()
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
	for channelName, channelDetails := range m.chMap {
		if channelDetails.Visibility == string(protocol.VisibilityPublic) {
			channels = append(channels, channelName)
		}
	}

	logger.WithField("channelCount", len(channels)).Info("Channel list retrieved")
	if len(channels) == 0 {
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: noActiveChannels,
		}
	} else {
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status:   protocol.StatusSuccess,
			Channels: channels,
		}
	}

	return chPayload
}

func (m *Manager) getUsers(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	chLock := m.getChannelLock(chPayload.ChannelName)
	chLock.RLock()
	defer chLock.RUnlock()

	logger.Info("Getting list of users")
	selectedCh, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}

	users := make([]string, 0, len(selectedCh.Users))
	for user := range selectedCh.Users {
		users = append(users, user)
	}

	logger.WithField("userCount", len(users)).Info("User list retrieved")
	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
		Users:  users,
	}
	return chPayload
}

func (m *Manager) messageChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	chLock := m.getChannelLock(chPayload.ChannelName)
	chLock.RLock()
	defer chLock.RUnlock()

	selectedCh, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		return protocol.ChannelPayload{
			OptionalChannelArgs: &protocol.OptionalChannelArgs{
				Status: protocol.StatusFail,
				Reason: chDoesNotExist,
			},
		}
	}

	if _, found := selectedCh.Users[chPayload.Requester]; !found {
		logger.WithFields(logrus.Fields{
			"channel": chPayload.ChannelName,
			"user":    chPayload.Requester,
		}).Warn("User not in the channel")
		return protocol.ChannelPayload{
			OptionalChannelArgs: &protocol.OptionalChannelArgs{
				Status: protocol.StatusFail,
				Reason: notInTheCh,
			},
		}
	}

	go func() {
		chLock := m.getChannelLock(chPayload.ChannelName)
		chLock.Lock()
		defer chLock.Unlock()
		selectedCh.LastActivity = time.Now().Unix()
	}()

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Debug("Message sent in channel")

	return protocol.ChannelPayload{
		OptionalChannelArgs: &protocol.OptionalChannelArgs{
			Status:  protocol.StatusSuccess,
			Message: chPayload.OptionalChannelArgs.Message,
		},
	}
}
