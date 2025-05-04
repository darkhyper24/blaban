package orders

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTPublisher handles publishing order status messages to MQTT
type MQTTPublisher struct {
	client mqtt.Client
}

// NewMQTTPublisher creates a new MQTT publisher
func NewMQTTPublisher(brokerURL string) (*MQTTPublisher, error) {
	// Set up MQTT client options
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("order_service_publisher").
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(1 * time.Minute)

	// Set up connection lost handler
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	// Create and connect client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	log.Println("Connected to MQTT broker for order status updates")
	return &MQTTPublisher{client: client}, nil
}

// PublishOrderStatus publishes an order status update to MQTT
func (p *MQTTPublisher) PublishOrderStatus(orderID, status, message string) error {
	// Create the status update message
	update := map[string]interface{}{
		"order_id":  orderID,
		"status":    status,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	// Marshal the message to JSON
	payload, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling order status update: %v", err)
		return err
	}

	// Publish to the order status topic
	topic := "orders/" + orderID + "/status"
	token := p.client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("Error publishing order status update: %v", token.Error())
		return token.Error()
	}

	log.Printf("Published status update for order %s: %s", orderID, status)
	return nil
}

// Close closes the MQTT connection
func (p *MQTTPublisher) Close() {
	if p.client.IsConnected() {
		p.client.Disconnect(250)
	}
}
