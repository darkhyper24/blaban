package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type NotificationClient struct {
	conn     *websocket.Conn
	messages chan Notification
}

// Notification represents the structure of a notification message
type Notification struct {
	Type      string                 `json:"type"`
	OrderID   string                 `json:"order_id"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

func NewNotificationClient(wsURL string) (*NotificationClient, error) {
	fmt.Println("\n🔌 Connecting to notification service WebSocket...")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %v", err)
	}

	return &NotificationClient{
		conn:     conn,
		messages: make(chan Notification),
	}, nil
}

func (c *NotificationClient) Start() {
	fmt.Println("✅ Connected to notification service!")
	fmt.Println("\n📢 Waiting for notifications...")
	fmt.Println("\n----------------------------------------")

	// Handle interrupt signal for clean shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Goroutine to read messages from WebSocket
	go c.readMessages()

	// Main loop
	for {
		select {
		case notification, ok := <-c.messages:
			if !ok {
				return
			}
			c.displayNotification(notification)
		case <-interrupt:
			// Cleanly close the connection
			fmt.Println("\n👋 Closing connection...")
			err := c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				log.Printf("❌ Error during closing WebSocket: %v", err)
			}
			return
		}
	}
}

func (c *NotificationClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *NotificationClient) readMessages() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("❌ WebSocket read error: %v", err)
			close(c.messages)
			return
		}

		var notification Notification
		if err := json.Unmarshal(message, &notification); err != nil {
			log.Printf("❌ Error parsing notification: %v", err)
			continue
		}

		c.messages <- notification
	}
}

// displayNotification prints a notification in a nice format
func displayNotification(n Notification) {
	// Get timestamp as formatted time
	t := time.Unix(n.Timestamp, 0)
	timeStr := t.Format("15:04:05")

	// Choose emoji based on status
	statusEmoji := "🔄"
	switch n.Status {
	case "pending":
		statusEmoji = "⏳"
	case "processing":
		statusEmoji = "👨‍🍳"
	case "completed":
		statusEmoji = "✅"
	case "cancelled":
		statusEmoji = "❌"
	}

	fmt.Println("\n📩 NEW NOTIFICATION RECEIVED")
	fmt.Printf("⏰ Time: %s\n", timeStr)
	fmt.Printf("🔑 Order ID: %s\n", n.OrderID)
	fmt.Printf("📊 Status: %s %s\n", statusEmoji, n.Status)
	fmt.Printf("💬 Message: %s\n", n.Message)
	fmt.Println("----------------------------------------")
}
