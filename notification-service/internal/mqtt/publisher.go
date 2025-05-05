package mqtt

import (
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// OrderStatusUpdate represents an order status update message
type OrderStatusUpdate struct {
	OrderID string                 `json:"order_id"`
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// PublishOrderStatus publishes an order status update to MQTT
func PublishOrderStatus(client mqtt.Client, update OrderStatusUpdate) error {
	topic := "orders/" + update.OrderID + "/status"

	payload, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling order status update: %v", err)
		return err
	}

	token := client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("Error publishing order status update: %v", token.Error())
		return token.Error()
	}

	log.Printf("Published order status update for order %s: %s", update.OrderID, update.Status)
	return nil
}
