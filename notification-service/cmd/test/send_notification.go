package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"notification-service/internal/mqtt"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Check if order ID was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run send_notification.go <order_id> [status]")
		os.Exit(1)
	}

	// Get order ID and optional status from command line args
	orderID := os.Args[1]
	status := "completed"
	if len(os.Args) > 2 {
		status = os.Args[2]
	}

	// Set up MQTT client options
	opts := mqttlib.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("go_mqtt_test_client").
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetAutoReconnect(true)

	// Create MQTT client
	client := mqttlib.NewClient(opts)

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Create order status update
	update := mqtt.OrderStatusUpdate{
		OrderID: orderID,
		Status:  status,
		Message: fmt.Sprintf("Order %s has been %s", orderID, status),
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}

	// Publish the update
	err := mqtt.PublishOrderStatus(client, update)
	if err != nil {
		log.Fatalf("Failed to publish order status: %v", err)
	}

	fmt.Printf("Successfully sent notification for order %s with status '%s'\n", orderID, status)
}
