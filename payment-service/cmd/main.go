package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	client, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI("mongodb://mongo:27017"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

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

	// Payment routes
	app.Post("/api/payments", handleCreatePayment)
	app.Get("/api/payments/:id", handleGetPayment)
	app.Post("/api/payments/webhook", handlePaymentWebhook)
	app.Get("/api/payments/order/:orderId", handleGetPaymentByOrder)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	log.Fatal(app.Listen(":8085"))
}

func handleCreatePayment(c *fiber.Ctx) error {
	// TODO: Implement payment processing
	return c.SendString("Process payment")
}

func handleGetPayment(c *fiber.Ctx) error {
	// TODO: Implement fetching payment details
	paymentId := c.Params("id")
	return c.SendString("Get payment details for ID: " + paymentId)
}

func handlePaymentWebhook(c *fiber.Ctx) error {
	// TODO: Implement handling payment provider webhooks
	return c.SendString("Payment webhook processed")
}

func handleGetPaymentByOrder(c *fiber.Ctx) error {
	// TODO: Implement fetching payment by order ID
	orderId := c.Params("orderId")
	return c.SendString("Get payment for order ID: " + orderId)
}
