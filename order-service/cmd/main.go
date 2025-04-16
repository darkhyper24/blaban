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

	// Order routes
	app.Get("/api/orders", handleGetOrders)
	app.Get("/api/orders/:id", handleGetOrder)
	app.Post("/api/orders", handleCreateOrder)

	log.Fatal(app.Listen(":8084"))
}

func handleGetOrders(c *fiber.Ctx) error {
	// TODO: Implement fetching user's order history
	return c.SendString("Get order history")
}

func handleGetOrder(c *fiber.Ctx) error {
	// TODO: Implement fetching a specific order
	orderId := c.Params("id")
	return c.SendString("Get order details for ID: " + orderId)
}

func handleCreateOrder(c *fiber.Ctx) error {
	// TODO: Implement creating a new order from cart
	return c.SendString("Create new order")
}
