package room

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/ncarlier/imgcast/internal/auth"
	"github.com/ncarlier/imgcast/internal/broadcaster"
	"github.com/ncarlier/imgcast/internal/storage"
	"github.com/ncarlier/imgcast/pkg/validator"
)

// Room represents a multi-room instance
type Room struct {
	Name        string
	broadcaster *broadcaster.Broadcaster
	auth        *auth.Authenticator
}

// Manager manages multiple rooms
type Manager struct {
	rooms     map[string]*Room
	storage   *storage.Storage
	adminAuth *auth.Authenticator
	mu        sync.RWMutex
}

// NewManager creates a new room manager
func NewManager(storage *storage.Storage, adminAuth *auth.Authenticator) *Manager {
	return &Manager{
		rooms:     make(map[string]*Room),
		storage:   storage,
		adminAuth: adminAuth,
	}
}

// GetRoom retrieves a room by name, creating it if it doesn't exist (for viewing)
func (m *Manager) GetRoom(roomName string) (*Room, error) {
	// Validate room name
	if !validator.IsValidRoomName(roomName) {
		return nil, fmt.Errorf("invalid room name: must be alphanumeric with dash/underscore only")
	}

	m.mu.RLock()
	room, exists := m.rooms[roomName]
	m.mu.RUnlock()

	if exists {
		return room, nil
	}

	// Check if room exists on disk
	if !m.storage.RoomExists(roomName) {
		return nil, fmt.Errorf("room does not exist")
	}

	// Create room instance
	return m.loadRoom(roomName)
}

// CreateRoom creates a new room with authentication
func (m *Manager) CreateRoom(roomName, username, password string) (*Room, error) {
	// Validate room name
	if !validator.IsValidRoomName(roomName) {
		return nil, fmt.Errorf("invalid room name: must be alphanumeric with dash/underscore only")
	}

	// Check if room already exists
	if m.storage.RoomExists(roomName) {
		return nil, fmt.Errorf("room already exists")
	}

	// Authenticate against admin htpasswd
	authenticated, err := m.adminAuth.Authenticate(username, password)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}
	if !authenticated {
		return nil, fmt.Errorf("admin authentication failed")
	}

	// Create room directory
	if err := m.storage.CreateRoom(roomName); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Create room htpasswd with the authenticated admin user
	roomAuth := auth.NewAuthenticator(m.storage.GetRoomHtpasswdPath(roomName))
	if err := roomAuth.CreateWithUser(username, password); err != nil {
		return nil, fmt.Errorf("failed to create room authentication: %w", err)
	}

	slog.Info("Room created", "room", roomName, "creator", username)

	// Load the room
	return m.loadRoom(roomName)
}

// AuthenticateForRoom authenticates a user against a room's htpasswd
func (m *Manager) AuthenticateForRoom(roomName, username, password string) (bool, error) {
	room, err := m.GetRoom(roomName)
	if err != nil {
		return false, err
	}

	return room.auth.Authenticate(username, password)
}

// GetOrCreateRoom gets a room if it exists, or creates it if the user is an admin
func (m *Manager) GetOrCreateRoom(roomName, username, password string) (*Room, bool, error) {
	// Validate room name
	if !validator.IsValidRoomName(roomName) {
		return nil, false, fmt.Errorf("invalid room name: must be alphanumeric with dash/underscore only")
	}

	// Check if room exists
	if m.storage.RoomExists(roomName) {
		// Room exists, authenticate against room htpasswd
		room, err := m.GetRoom(roomName)
		if err != nil {
			return nil, false, err
		}

		authenticated, err := room.auth.Authenticate(username, password)
		if err != nil {
			return nil, false, fmt.Errorf("authentication error: %w", err)
		}
		if !authenticated {
			return nil, false, fmt.Errorf("room authentication failed")
		}

		return room, false, nil
	}

	// Room doesn't exist, try to create it (requires admin auth)
	room, err := m.CreateRoom(roomName, username, password)
	if err != nil {
		return nil, false, err
	}

	return room, true, nil
}

// loadRoom loads an existing room from disk
func (m *Manager) loadRoom(roomName string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check if room was loaded by another goroutine
	if room, exists := m.rooms[roomName]; exists {
		return room, nil
	}

	// Create room instance
	room := &Room{
		Name:        roomName,
		broadcaster: broadcaster.New(),
		auth:        auth.NewAuthenticator(m.storage.GetRoomHtpasswdPath(roomName)),
	}

	m.rooms[roomName] = room
	slog.Info("Room loaded", "room", roomName)

	return room, nil
}

// GetBroadcaster returns the broadcaster for a room
func (r *Room) GetBroadcaster() *broadcaster.Broadcaster {
	return r.broadcaster
}
