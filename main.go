package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//go:embed static
var static embed.FS

const (
	// Filename is the name of the file where the image data will be stored
	LiveDataFilename = "imgcast.data"
)

type sseClient struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

var (
	staticServer http.Handler
	clients      = make(map[*sseClient]bool)
	broadcast    = make(chan struct{})
	mu           sync.Mutex
	imagePath    string
	apiKey       string
	basePath     string
)

func init() {
	fSys, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}
	staticServer = http.FileServer(http.FS(fSys))
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		slog.Warn("API_KEY is not set, using default key")
		apiKey = "secret"
	}
	basePath = getBasePath()
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

// joinPath joins base path with a relative path
func joinPath(relativePath string) string {
	if basePath == "/" {
		return relativePath
	}
	return strings.TrimSuffix(basePath, "/") + relativePath
}

func main() {
	// Setup routes with base path support
	if basePath == "/" {
		http.Handle("/", staticServer)
	} else {
		http.Handle(basePath, http.StripPrefix(strings.TrimSuffix(basePath, "/"), staticServer))
	}

	http.HandleFunc(joinPath("/upload"), handleUpload)
	http.HandleFunc(joinPath("/events"), handleSSE)
	http.HandleFunc(joinPath("/live"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, filepath.Join(os.TempDir(), LiveDataFilename))
	})

	go broadcaster()

	port := getPort()
	slog.Info("Starting server", "port", port, "basePath", basePath)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	clientKey := r.Header.Get("X-API-Key")
	if clientKey != apiKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mu.Lock()
	// Create a temporary directory to store the uploaded file
	tmpDir := os.TempDir()
	dstPath := filepath.Join(tmpDir, LiveDataFilename)

	// remove the old file if it exists
	os.Remove(dstPath)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Unable to create live data", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Unable to write live data", http.StatusInternalServerError)
		return
	}

	mu.Unlock()

	slog.Info("Live data updated", "path", dstPath)
	broadcast <- struct{}{}

	w.WriteHeader(http.StatusOK)
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
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

	client := &sseClient{
		w:       w,
		flusher: flusher,
	}

	mu.Lock()
	clients[client] = true
	mu.Unlock()

	// Send initial message
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	// Keep connection alive until client disconnects
	<-r.Context().Done()

	mu.Lock()
	delete(clients, client)
	mu.Unlock()
}

func broadcaster() {
	for {
		<-broadcast
		slog.Info("Broadcasting update to clients", "count", len(clients))
		mu.Lock()
		for client := range clients {
			_, err := fmt.Fprintf(client.w, "data: updated\n\n")
			if err != nil {
				delete(clients, client)
				continue
			}
			client.flusher.Flush()
		}
		mu.Unlock()
	}
}
