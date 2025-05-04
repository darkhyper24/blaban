package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Notification represents the structure of a notification message
type Notification struct {
	Type      string                 `json:"type"`
	OrderID   string                 `json:"order_id"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

func main() {
	// Connect to the WebSocket server
	fmt.Println("\n🔌 Connecting to notification service WebSocket...")
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8085/ws", nil)
	if err != nil {
		log.Fatalf("❌ Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to notification service!")
	fmt.Println("\n📢 Waiting for notifications...")
	fmt.Println("\n----------------------------------------")

	// Handle interrupt signal for clean shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Channel to receive messages
	messages := make(chan Notification)

	// Goroutine to read messages from WebSocket
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("❌ WebSocket read error: %v", err)
				close(messages)
				return
			}

			var notification Notification
			if err := json.Unmarshal(message, &notification); err != nil {
				log.Printf("❌ Error parsing notification: %v", err)
				continue
			}

			messages <- notification
		}
	}()

	// Main loop
	for {
		select {
		case notification, ok := <-messages:
			if !ok {
				return
			}
			displayNotification(notification)
		case <-interrupt:
			// Cleanly close the connection
			fmt.Println("\n👋 Closing connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("❌ Error during closing WebSocket: %v", err)
			}
			return
		}
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
