package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// LiveDataFilename is the name of the file where the image data is stored
	LiveDataFilename = "imgcast.data"
	// HtpasswdFilename is the name of the htpasswd file
	HtpasswdFilename = ".htpasswd"
)

// Storage handles file operations for rooms
type Storage struct {
	baseDir string
}

// New creates a new storage instance
func New(baseDir string) *Storage {
	return &Storage{
		baseDir: baseDir,
	}
}

// GetRoomDir returns the directory path for a room
func (s *Storage) GetRoomDir(roomName string) string {
	return filepath.Join(s.baseDir, "rooms", roomName)
}

// GetRoomImagePath returns the path to the room's image file
func (s *Storage) GetRoomImagePath(roomName string) string {
	return filepath.Join(s.GetRoomDir(roomName), LiveDataFilename)
}

// GetRoomHtpasswdPath returns the path to the room's htpasswd file
func (s *Storage) GetRoomHtpasswdPath(roomName string) string {
	return filepath.Join(s.GetRoomDir(roomName), HtpasswdFilename)
}

// GetAdminHtpasswdPath returns the path to the admin htpasswd file
func (s *Storage) GetAdminHtpasswdPath() string {
	return filepath.Join(s.baseDir, HtpasswdFilename)
}

// RoomExists checks if a room directory exists
func (s *Storage) RoomExists(roomName string) bool {
	if roomName == "" {
		return false
	}
	_, err := os.Stat(s.GetRoomDir(roomName))
	return err == nil
}

// CreateRoom creates a new room directory
func (s *Storage) CreateRoom(roomName string) error {
	roomDir := s.GetRoomDir(roomName)
	if err := os.MkdirAll(roomDir, 0755); err != nil {
		return fmt.Errorf("failed to create room directory: %w", err)
	}
	return nil
}

// SaveImage saves an image to the room's storage
func (s *Storage) SaveImage(roomName string, reader io.Reader) error {
	imagePath := s.GetRoomImagePath(roomName)

	// Remove old file if it exists
	os.Remove(imagePath)

	// Create the image file
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Copy the image data
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// EnsureBaseDir ensures the base directory structure exists
func (s *Storage) EnsureBaseDir() error {
	// Create base directory
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create rooms directory
	roomsDir := filepath.Join(s.baseDir, "rooms")
	if err := os.MkdirAll(roomsDir, 0755); err != nil {
		return fmt.Errorf("failed to create rooms directory: %w", err)
	}

	return nil
}
