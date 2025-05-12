package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/mqtt"
	"notification-service/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
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
	go ws.ServeHTTP(":8087", hub)

	http.Handle("/metrics", promhttp.Handler())

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		method := c.Method()
		endpoint := c.Route().Path

		httpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		return err
	})

	go func() {
		log.Println("Starting Prometheus metrics server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start Prometheus metrics server: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Fiber app on :3000")
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("Failed to start Fiber app: %v", err)
		}
	}()

	log.Println("Notification service is running. WebSocket server on :8087")

	log.Println("Notification service is running. WebSocket server on :8087")
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
