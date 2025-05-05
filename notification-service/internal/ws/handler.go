package ws

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

// WebsocketHandler returns a Fiber handler function for WebSocket connections
func WebsocketHandler(hub *Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Extract user information from query params if needed
		userID := c.Query("user_id") // Optional: Extract user ID if available

		if userID != "" {
			log.Printf("WebSocket client connected for user: %s", userID)
		}

		// Create an adapter for Fiber websocket to work with our hub
		// that expects gorilla websocket connections
		adapter := &WebSocketAdapter{conn: c}
		hub.AddClient(adapter)

		// Keep connection open and handle client messages
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				hub.RemoveClient(adapter)
				break
			}

			// For now, just log received messages - could process client commands here
			log.Printf("Received message from client: %s (type: %d)", message, messageType)
		}
	})
}

// ClientCountHandler returns a handler that returns the current number of connected clients
func ClientCountHandler(hub *Hub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub.mu.Lock()
		count := len(hub.clients)
		hub.mu.Unlock()
		return c.JSON(fiber.Map{"client_count": count})
	}
}

// ServeHTTP starts the HTTP server for WebSocket connections
func ServeHTTP(addr string, hub *Hub) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// CORS middleware - Fix the insecure configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://localhost:8080", // Specific origins instead of wildcard
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin",
		AllowCredentials: true,
	}))

	// WebSocket middleware to filter websocket requests
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Routes
	app.Get("/ws", WebsocketHandler(hub))
	app.Get("/stats", ClientCountHandler(hub))
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Printf("Starting WebSocket server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start WebSocket server: %v", err)
	}
}

// HandleOrderStatusMessage processes order status messages from MQTT
func HandleOrderStatusMessage(hub *Hub, msg mqtt.Message) {
	log.Printf("Received message on topic: %s", msg.Topic())

	// Try to parse the message as JSON
	var orderUpdate map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &orderUpdate); err != nil {
		// If not valid JSON, just log the error and don't broadcast
		log.Printf("Failed to parse message as JSON: %v", err)
		return
	}

	// Create a standardized notification
	notification := NotificationMessage{
		Type:      "order_status_update",
		OrderID:   getStringValue(orderUpdate, "order_id"),
		Status:    getStringValue(orderUpdate, "status"),
		Message:   getStringValue(orderUpdate, "message"),
		Data:      orderUpdate,
		Timestamp: time.Now().Unix(),
	}

	// Broadcast the notification
	hub.BroadcastJSON(notification)
}

// Helper function to safely extract string values from a map
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
