package ws

import (
	"encoding/json"
	"log"
	"sync"
)

// WebSocketConn defines the interface for websocket connections
type WebSocketConn interface {
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type NotificationMessage struct {
	Type      string                 `json:"type"`
	OrderID   string                 `json:"order_id"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

type Hub struct {
	clients map[WebSocketConn]bool
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[WebSocketConn]bool),
	}
}

// Run starts the hub processing
func (h *Hub) Run() {
	log.Println("WebSocket hub is running")
}

func (h *Hub) AddClient(conn WebSocketConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
	log.Printf("WebSocket client added. Total clients: %d", len(h.clients))
}

func (h *Hub) RemoveClient(conn WebSocketConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
	log.Printf("WebSocket client removed. Total clients: %d", len(h.clients))
}

// Broadcast sends the given data to all connected clients.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		// Set TextMessage type (1)
		err := conn.WriteMessage(1, msg)
		if err != nil {
			log.Printf("Error sending message to a client: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
	log.Printf("Broadcasted message to %d clients", len(h.clients))
}

// BroadcastJSON marshals and broadcasts the given notification message to all connected clients
func (h *Hub) BroadcastJSON(notification NotificationMessage) {
	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling message for broadcast: %v", err)
		return
	}

	h.Broadcast(data)
}
