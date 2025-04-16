package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// Payment routes
	app.Post("/api/payments", handleCreatePayment)
	app.Get("/api/payments/:id", handleGetPayment)
	app.Post("/api/payments/webhook", handlePaymentWebhook)
	app.Get("/api/payments/order/:orderId", handleGetPaymentByOrder)

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
