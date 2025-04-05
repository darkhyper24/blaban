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

	app.Post("/api/users/signup", handleSignup)
	app.Post("/api/users/login", handleLogin)
	app.Get("/api/users/profile", handleGetProfile)
	app.Put("/api/users/profile", handleUpdateProfile)

	log.Fatal(app.Listen(":3001"))
}

func handleSignup(c *fiber.Ctx) error {
	// TODO
	return c.SendString("User signup")
}

func handleLogin(c *fiber.Ctx) error {
	// TODO
	return c.SendString("User login")
}

func handleGetProfile(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Get user profile")
}

func handleUpdateProfile(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Update user profile")
}
