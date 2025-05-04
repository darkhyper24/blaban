package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/mqtt"
	"notification-service/internal/ws"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Get MQTT client options
	opts := mqtt.GetMQTTOptions(hub)

	// Create a new MQTT client
	mqttClient := mqttlib.NewClient(opts)

	// Connect to the MQTT broker with retry mechanism
	log.Println("Connecting to MQTT broker...")
	connected := false
	maxRetries := 5
	retryCount := 0

	for !connected && retryCount < maxRetries {
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("Failed to connect to MQTT broker (attempt %d/%d): %v",
				retryCount+1, maxRetries, token.Error())
			retryCount++

			if retryCount < maxRetries {
				// Exponential backoff for retries
				backoffTime := time.Duration(1<<uint(retryCount)) * time.Second
				log.Printf("Retrying in %v...", backoffTime)
				time.Sleep(backoffTime)
			}
		} else {
			connected = true
			log.Println("Successfully connected to MQTT broker")
		}
	}

	if !connected {
		log.Fatalf("Failed to connect to MQTT broker after %d attempts", maxRetries)
	}

	// Setup HTTP server for WebSockets
	go ws.ServeHTTP(":8085", hub)

	log.Println("Notification service is running. WebSocket server on :8085")
	log.Println("MQTT client is subscribed to order status updates")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a termination signal
	<-sigChan

	// Clean shutdown
	log.Println("Shutting down...")
	if mqttClient.IsConnected() {
		mqttClient.Disconnect(250) // Wait 250ms to complete any in-progress requests
	}
	log.Println("MQTT client disconnected")

	// Allow time for other graceful shutdowns
	time.Sleep(500 * time.Millisecond)
}
