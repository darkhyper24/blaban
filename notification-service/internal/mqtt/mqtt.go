package mqtt

import (
	"log"
	"time"

	"notification-service/internal/ws"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// GetMQTTOptions returns configured MQTT client options
func GetMQTTOptions(hub *ws.Hub) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("go_mqtt_client").
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetCleanSession(false).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(1 * time.Minute)

	// Set up a callback for when the client connects to the broker
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("Connected to MQTT broker")

		// Subscribe to topics when connection is established
		// In mqtt.go, replace the handleOrderStatusMessage call with:
		if token := c.Subscribe("orders/+/status", 1, func(client mqtt.Client, msg mqtt.Message) {
			ws.HandleOrderStatusMessage(hub, msg)
		}); token.Wait() && token.Error() != nil {
			log.Printf("Error subscribing to order status topic: %v", token.Error())
		}
	}

	// Set up a callback for when the client disconnects from the broker
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	}

	return opts
}
