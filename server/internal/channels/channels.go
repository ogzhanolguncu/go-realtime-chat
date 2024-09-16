package channels

import (
	"fmt"
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
	notChannelOwner     = "Not a channel owner"
	emptyTargetUser     = "Target user cannot be empty"
	ownerCannotBeKicked = "Owner cannot be kicked"
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
	Users        map[string]bool
	LastActivity int64
	BannedUsers  map[string]bool
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

func (m *Manager) Handle(payload protocol.Payload) (protocol.ChannelPayload, protocol.ChannelPayload) {
	logger.WithFields(logrus.Fields{
		"action":  payload.ChannelPayload.ChannelAction,
		"channel": payload.ChannelPayload.ChannelName,
		"user":    payload.ChannelPayload.Requester,
	}).Info("Handling channel action")

	switch payload.ChannelPayload.ChannelAction {
	case protocol.CreateChannel:
		return m.createChannel(*payload.ChannelPayload), protocol.ChannelPayload{}
	case protocol.JoinChannel:
		return m.joinChannel(*payload.ChannelPayload)
	case protocol.LeaveChannel:
		return m.leaveChannel(*payload.ChannelPayload), protocol.ChannelPayload{}
	case protocol.GetChannels:
		return m.getChannels(*payload.ChannelPayload), protocol.ChannelPayload{}
	case protocol.GetUsers:
		return m.getUsers(*payload.ChannelPayload), protocol.ChannelPayload{}
	case protocol.MessageChannel:
		return m.messageChannel(*payload.ChannelPayload), protocol.ChannelPayload{}
	case protocol.KickUser:
		return m.kickUser(*payload.ChannelPayload), protocol.ChannelPayload{}
	default:
		logger.WithField("action", payload.ChannelPayload.ChannelAction).Warn("Unknown channel action")
		return protocol.ChannelPayload{
			OptionalChannelArgs: &protocol.OptionalChannelArgs{
				Status: protocol.StatusFail,
				Reason: unknownChAction,
			},
		}, protocol.ChannelPayload{}
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
		Users:        map[string]bool{chPayload.Requester: true},
		BannedUsers:  map[string]bool{},
		LastActivity: time.Now().Unix(),
		Visibility:   string(chPayload.OptionalChannelArgs.Visibility),
	}

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"owner":   chPayload.Requester,
	}).Info("channel created successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status:     protocol.StatusSuccess,
		Visibility: chPayload.OptionalChannelArgs.Visibility,
	}
	return chPayload
}

func (m *Manager) joinChannel(chPayload protocol.ChannelPayload) (joinPayload, noticePayload protocol.ChannelPayload) {
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
		return chPayload, protocol.ChannelPayload{}
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
		return chPayload, protocol.ChannelPayload{}
	}

	// Channel is full
	if len(channel.Users) >= channel.ChCapacity {
		logger.WithFields(logrus.Fields{
			"channel": chPayload.ChannelName,
			"user":    chPayload.Requester,
		}).Warn("Channel at capacity")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chAtCapacity,
		}
		return chPayload, protocol.ChannelPayload{}
	}

	channel.Users[chPayload.Requester] = true
	channel.LastActivity = time.Now().Unix()
	m.chMap[chPayload.ChannelName] = channel

	logger.WithFields(logrus.Fields{
		"channel": chPayload.ChannelName,
		"user":    chPayload.Requester,
	}).Info("User joined channel successfully")

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
	}

	chNoticePayload := m.prepareNoticePayload(chPayload, channel, fmt.Sprintf("'%s' has joined the channel", chPayload.Requester))
	return chPayload, chNoticePayload
}

func (*Manager) prepareNoticePayload(chPayload protocol.ChannelPayload, channel *ChannelDetails, message string) protocol.ChannelPayload {
	chNoticeResp := chPayload
	chNoticeResp.ChannelAction = protocol.NoticeChannel
	users := make([]string, 0, len(channel.Users))
	for user := range channel.Users {
		users = append(users, user)
	}
	chNoticeResp.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
		Notice: message,
		Users:  users,
	}
	return chNoticeResp
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

	if _, found := channel.Users[chPayload.Requester]; !found {
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
	for channelName, channelDetails := range m.chMap {
		if channelDetails.Visibility == string(protocol.VisibilityPublic) {
			channels = append(channels, channelName)
		}
	}

	logger.WithField("channelCount", len(channels)).Info("Channel list retrieved")
	// There is no public channel
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
	for user := range selectedCh.Users {
		users = append(users, user)
	}

	logger.WithField("userCount", len(selectedCh.Users)).Info("User list retrieved")
	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status: protocol.StatusSuccess,
		Users:  users,
	}
	return chPayload
}

func (m *Manager) messageChannel(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.RLock()
	defer m.lock.RUnlock()

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

	//Check if requesting user is in the channel
	if _, found := m.chMap[chPayload.ChannelName].Users[chPayload.Requester]; !found {
		logger.WithField("channel", chPayload.ChannelName).Warn("User not in the channel")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: notInTheCh,
		}
		return chPayload
	}

	users := make([]string, 0, len(selectedCh.Users))
	for user := range selectedCh.Users {
		users = append(users, user)
	}

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status:  protocol.StatusSuccess,
		Users:   users,
		Message: chPayload.OptionalChannelArgs.Message,
	}
	return chPayload
}

func (m *Manager) kickUser(chPayload protocol.ChannelPayload) protocol.ChannelPayload {
	m.lock.Lock()
	defer m.lock.Unlock()

	//Channel has to exist
	selectedCh, exists := m.chMap[chPayload.ChannelName]
	if !exists {
		logger.WithField("channel", chPayload.ChannelName).Warn("Channel does not exist")
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: chDoesNotExist,
		}
		return chPayload
	}

	//Requester has to be the owner
	if selectedCh.Owner != chPayload.Requester {
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: notChannelOwner,
		}
		return chPayload
	}

	//Payload has to have target user
	if chPayload.OptionalChannelArgs != nil &&
		chPayload.OptionalChannelArgs.TargetUser == protocol.EmptyChannelField {
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: emptyTargetUser,
		}
		return chPayload
	}

	//Payload has to have target user
	if chPayload.OptionalChannelArgs.TargetUser == selectedCh.Owner {
		chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
			Status: protocol.StatusFail,
			Reason: ownerCannotBeKicked,
		}
		return chPayload
	}
	//Delete user from user list
	delete(selectedCh.Users, chPayload.OptionalChannelArgs.TargetUser)

	chPayload.OptionalChannelArgs = &protocol.OptionalChannelArgs{
		Status:     protocol.StatusSuccess,
		TargetUser: chPayload.OptionalChannelArgs.TargetUser,
	}

	return chPayload
}
