package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Port          string
	BasePath      string
	RoomsBaseDir  string
	AdminHtpasswd string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:          getPort(),
		BasePath:      getBasePath(),
		RoomsBaseDir:  getRoomsBaseDir(),
		AdminHtpasswd: getAdminHtpasswd(),
	}
}

// getPort returns the port from environment variable or default port
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Validate port is a valid number
	if _, err := strconv.Atoi(port); err != nil {
		slog.Warn("Invalid PORT value, using default", "port", port)
		port = "8080"
	}
	return ":" + port
}

// getBasePath returns the base path from environment variable
func getBasePath() string {
	path := os.Getenv("BASE_PATH")
	if path == "" {
		return "/"
	}
	// Ensure base path starts with / and ends with /
	if path[0] != '/' {
		path = "/" + path
	}
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	return path
}

// getRoomsBaseDir returns the rooms base directory from environment variable
func getRoomsBaseDir() string {
	dir := os.Getenv("ROOMS_BASE_DIR")
	if dir == "" {
		dir = "var"
	}
	return dir
}

// getAdminHtpasswd returns the path to the admin htpasswd file
func getAdminHtpasswd() string {
	dir := getRoomsBaseDir()
	return dir + "/.htpasswd"
}

// JoinPath joins base path with a relative path
func (c *Config) JoinPath(relativePath string) string {
	if c.BasePath == "/" {
		return relativePath
	}
	return strings.TrimSuffix(c.BasePath, "/") + relativePath
}
