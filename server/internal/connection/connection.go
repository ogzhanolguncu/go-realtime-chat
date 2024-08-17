package connection

import (
	"net"
	"sync"
)

type ConnectionInfo struct {
	Connection net.Conn
	OwnerName  string
}

type Manager struct {
	connectionMap sync.Map
}

func NewConnectionManager() *Manager {
	return &Manager{}
}

func (cm *Manager) AddConnection(c net.Conn, info *ConnectionInfo) {
	cm.connectionMap.Store(c, info)
}

func (cm *Manager) DeleteConnection(c net.Conn) {
	cm.connectionMap.Delete(c)
}

func (cm *Manager) GetConnectedUsersCount() int {
	count := 0
	cm.connectionMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (cm *Manager) GetActiveUsers() []string {
	var users []string
	cm.connectionMap.Range(func(key, value interface{}) bool {
		info := value.(*ConnectionInfo)
		users = append(users, info.OwnerName)
		return true
	})
	return users
}

func (cm *Manager) GetConnectionInfo(c net.Conn) (*ConnectionInfo, bool) {
	value, ok := cm.connectionMap.Load(c)
	if !ok {
		return nil, false
	}
	info, ok := value.(*ConnectionInfo)
	return info, ok
}

func (cm *Manager) FindConnectionByOwnerName(ownerName string) (net.Conn, bool) {
	var foundConn net.Conn
	var found bool

	cm.connectionMap.Range(func(key, value interface{}) bool {
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

// RangeConnections iterates over all connections and applies the given function
func (m *Manager) RangeConnections(f func(conn net.Conn, info *ConnectionInfo) bool) {
	m.connectionMap.Range(func(key, value interface{}) bool {
		conn := key.(net.Conn)
		info := value.(*ConnectionInfo)
		return f(conn, info)
	})
}
