package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed static
var static embed.FS

const (
	// Filename is the name of the file where the image data will be stored
	LiveDataFilename = "imgcast.data"
)

var (
	staticServer http.Handler
	clients      = make(map[*websocket.Conn]bool)
	broadcast    = make(chan struct{})
	upgrader     = websocket.Upgrader{}
	mu           sync.Mutex
	imagePath    string
	apiKey       string
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
}

func main() {
	http.Handle("/", staticServer)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(os.TempDir(), LiveDataFilename))
	})

	go broadcaster()

	slog.Info("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	if imagePath != "" {
		conn.WriteMessage(websocket.TextMessage, []byte(imagePath))
	}
	mu.Unlock()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
	}
}

func broadcaster() {
	for {
		<-broadcast
		slog.Info("Broadcasting update to clients")
		mu.Lock()
		for conn := range clients {
			err := conn.WriteMessage(websocket.TextMessage, []byte("updated"))
			if err != nil {
				conn.Close()
				delete(clients, conn)
			}
		}
		mu.Unlock()
	}
}
