package broadcaster

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// Client represents an SSE client connection
type Client struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// Broadcaster manages SSE connections and broadcasts messages to clients
type Broadcaster struct {
	clients   map[*Client]bool
	broadcast chan struct{}
	mu        sync.Mutex
}

// New creates a new broadcaster instance
func New() *Broadcaster {
	b := &Broadcaster{
		clients:   make(map[*Client]bool),
		broadcast: make(chan struct{}, 10),
	}
	go b.run()
	return b
}

// run is the broadcaster's main loop that listens for broadcast events
func (b *Broadcaster) run() {
	for {
		<-b.broadcast
		b.sendToClients()
	}
}

// sendToClients sends the "updated" message to all connected clients
func (b *Broadcaster) sendToClients() {
	b.mu.Lock()
	defer b.mu.Unlock()

	slog.Info("Broadcasting update to clients", "count", len(b.clients))
	for client := range b.clients {
		_, err := fmt.Fprintf(client.w, "data: updated\n\n")
		if err != nil {
			delete(b.clients, client)
			continue
		}
		client.flusher.Flush()
	}
}

// AddClient registers a new SSE client
func (b *Broadcaster) AddClient(w http.ResponseWriter, flusher http.Flusher) *Client {
	client := &Client{
		w:       w,
		flusher: flusher,
	}

	b.mu.Lock()
	b.clients[client] = true
	b.mu.Unlock()

	// Send initial connected message
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	return client
}

// RemoveClient unregisters an SSE client
func (b *Broadcaster) RemoveClient(client *Client) {
	b.mu.Lock()
	delete(b.clients, client)
	b.mu.Unlock()
}

// Notify sends a broadcast event to all connected clients
func (b *Broadcaster) Notify() {
	select {
	case b.broadcast <- struct{}{}:
	default:
		// Channel full, skip this notification
		slog.Warn("Broadcast channel full, skipping notification")
	}
}

// ClientCount returns the number of connected clients
func (b *Broadcaster) ClientCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients)
}
