// menu-service/cmd/main.go
package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
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

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	app.Get("/api/menu", handleGetMenu)
	app.Get("/api/menu/:id", handleGetMenuItem)
	app.Post("/api/menu", handleCreateMenuItem)           // manager only
	app.Put("/api/menu/:id", handleUpdateMenuItem)        // manager only
	app.Post("/api/menu/:id/discount", handleAddDiscount) // manager only
	app.Get("/api/menu/search", handleSearchItems)
	app.Get("/api/menu/filter", handleFilterItems)

	log.Fatal(app.Listen(":8083"))
}

func handleGetMenu(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Get all menu items")
}

func handleGetMenuItem(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Get menu item by ID")
}

func handleCreateMenuItem(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Create menu item")
}

func handleUpdateMenuItem(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Update menu item")
}

func handleAddDiscount(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Add discount")
}

func handleSearchItems(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Search menu items")
}

func handleFilterItems(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Filter menu items")
}
