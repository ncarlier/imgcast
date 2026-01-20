package handlers

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ncarlier/imgcast/internal/config"
	"github.com/ncarlier/imgcast/internal/room"
	"github.com/ncarlier/imgcast/internal/storage"
)

// Server holds the HTTP handlers and dependencies
type Server struct {
	config       *config.Config
	roomManager  *room.Manager
	storage      *storage.Storage
	staticServer http.Handler
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, roomManager *room.Manager, storage *storage.Storage, staticFS embed.FS) (*Server, error) {
	// Setup static file server
	fSys, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("failed to setup static file server: %w", err)
	}

	return &Server{
		config:       cfg,
		roomManager:  roomManager,
		storage:      storage,
		staticServer: http.FileServer(http.FS(fSys)),
	}, nil
}

// extractRoomName extracts the room name from a request path
// Path format: /{roomname}/upload or /{roomname}/live or /{roomname}/events or /{roomname}
func (s *Server) extractRoomName(path string) string {
	// Remove base path if present
	if s.config.BasePath != "/" {
		path = strings.TrimPrefix(path, strings.TrimSuffix(s.config.BasePath, "/"))
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Split by slash and get first part
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

// HandleUpload handles image upload to a room
func (s *Server) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract room name from path
	roomName := s.extractRoomName(r.URL.Path)
	if roomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	// Get basic auth credentials
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Room Upload"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get or create room with authentication
	room, created, err := s.roomManager.GetOrCreateRoom(roomName, username, password)
	if err != nil {
		slog.Error("Failed to get/create room", "room", roomName, "error", err)
		w.Header().Set("WWW-Authenticate", `Basic realm="Room Upload"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if created {
		slog.Info("New room created via upload", "room", roomName, "creator", username)
	}

	// Parse multipart form
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save the image
	if err := s.storage.SaveImage(roomName, file); err != nil {
		slog.Error("Failed to save image", "room", roomName, "error", err)
		http.Error(w, "Unable to save image", http.StatusInternalServerError)
		return
	}

	slog.Info("Image uploaded", "room", roomName, "user", username)

	// Notify all connected clients
	room.GetBroadcaster().Notify()

	w.WriteHeader(http.StatusOK)
}

// HandleLive serves the current image for a room
func (s *Server) HandleLive(w http.ResponseWriter, r *http.Request) {
	// Extract room name from path
	roomName := s.extractRoomName(r.URL.Path)
	if roomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	// Check if room exists
	if !s.storage.RoomExists(roomName) {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Set cache headers
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the image file
	imagePath := s.storage.GetRoomImagePath(roomName)
	http.ServeFile(w, r, imagePath)
}

// HandleSSE handles Server-Sent Events for a room
func (s *Server) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Extract room name from path
	roomName := s.extractRoomName(r.URL.Path)
	if roomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	// Get the room (creates if exists on disk)
	room, err := s.roomManager.GetRoom(roomName)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Add client to broadcaster
	client := room.GetBroadcaster().AddClient(w, flusher)

	// Keep connection alive until client disconnects
	<-r.Context().Done()

	// Remove client from broadcaster
	room.GetBroadcaster().RemoveClient(client)
}

// HandleStatic serves static files for a room
func (s *Server) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// Extract room name from path
	roomName := s.extractRoomName(r.URL.Path)

	// If request is for favicon.ico, serve it directly
	if roomName == "favicon.ico" {
		s.staticServer.ServeHTTP(w, r)
		return
	}

	// Check if room exists
	if !s.storage.RoomExists(roomName) {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Trim room name from URL path to serve static files correctly
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+roomName)
	s.staticServer.ServeHTTP(w, r)
}

// RegisterRoutes registers all HTTP routes
func (s *Server) RegisterRoutes() {
	// Main handler that routes to appropriate endpoints
	mainHandler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if this is a room-specific endpoint
		if strings.HasSuffix(path, "/upload") {
			s.HandleUpload(w, r)
		} else if strings.HasSuffix(path, "/live") {
			s.HandleLive(w, r)
		} else if strings.HasSuffix(path, "/events") {
			s.HandleSSE(w, r)
		} else {
			s.HandleStatic(w, r)
		}
	}

	// Register main handler
	if s.config.BasePath == "/" {
		http.HandleFunc("/", mainHandler)
	} else {
		http.Handle(s.config.BasePath, http.StripPrefix(strings.TrimSuffix(s.config.BasePath, "/"), http.HandlerFunc(mainHandler)))
	}
}
