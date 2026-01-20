package main

import (
	"embed"
	"log"
	"log/slog"
	"net/http"

	"github.com/ncarlier/imgcast/internal/auth"
	"github.com/ncarlier/imgcast/internal/config"
	"github.com/ncarlier/imgcast/internal/handlers"
	"github.com/ncarlier/imgcast/internal/room"
	"github.com/ncarlier/imgcast/internal/storage"
)

//go:embed static
var staticFS embed.FS

func main() {
	// Load configuration
	cfg := config.Load()
	slog.Info("Configuration loaded", "port", cfg.Port, "basePath", cfg.BasePath, "roomsBaseDir", cfg.RoomsBaseDir)

	// Initialize storage
	store := storage.New(cfg.RoomsBaseDir)
	if err := store.EnsureBaseDir(); err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}

	// Initialize admin authentication
	adminAuth := auth.NewAuthenticator(cfg.AdminHtpasswd)
	if !adminAuth.Exists() {
		slog.Warn("Admin htpasswd file not found - rooms cannot be created until admin credentials are set up",
			"path", cfg.AdminHtpasswd)
		slog.Info("To create admin credentials, run: mkdir -p var && htpasswd -nbB admin secret > var/.htpasswd")
	}

	// Initialize room manager
	roomManager := room.NewManager(store, adminAuth)

	// Initialize HTTP server
	server, err := handlers.NewServer(cfg, roomManager, store, staticFS)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Register routes
	server.RegisterRoutes()

	// Start server
	slog.Info("Starting imgcast server", "port", cfg.Port, "basePath", cfg.BasePath)
	log.Fatal(http.ListenAndServe(cfg.Port, nil))
}
