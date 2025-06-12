package websocket

import (
	"sync"
	"time"
)

type UserStatus struct {
	UserID    int64
	IsOnline  bool
	LastSeen  time.Time
	Username  string
	Avatar    string
}

type UserStatusManager struct {
	mu     sync.RWMutex
	users  map[int64]*UserStatus
}

var (
	statusManager *UserStatusManager
	once          sync.Once
)

func GetStatusManager() *UserStatusManager {
	once.Do(func() {
		statusManager = &UserStatusManager{
			users: make(map[int64]*UserStatus),
		}
	})
	return statusManager
}

func (m *UserStatusManager) SetOnline(userID int64, username, avatar string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if user, exists := m.users[userID]; exists {
		user.IsOnline = true
		user.LastSeen = time.Now()
	} else {
		m.users[userID] = &UserStatus{
			UserID:    userID,
			IsOnline:  true,
			LastSeen:  time.Now(),
			Username:  username,
			Avatar:    avatar,
		}
	}
}

func (m *UserStatusManager) SetOffline(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if user, exists := m.users[userID]; exists {
		user.IsOnline = false
		user.LastSeen = time.Now()
	}
}

func (m *UserStatusManager) GetStatus(userID int64) (bool, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if user, exists := m.users[userID]; exists {
		return user.IsOnline, user.LastSeen
	}
	return false, time.Time{}
}

func (m *UserStatusManager) GetAllOnlineUsers() []*UserStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var onlineUsers []*UserStatus
	for _, user := range m.users {
		if user.IsOnline {
			onlineUsers = append(onlineUsers, user)
		}
	}
	return onlineUsers
}

func (m *UserStatusManager) GetUser(userID int64) (*UserStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[userID]
	return user, exists
}