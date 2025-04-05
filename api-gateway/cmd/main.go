package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	setupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}

func setupRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	api := app.Group("/api")

	api.All("/users/*", createServiceProxy("user-service", 3001))
	api.All("/auth/*", createServiceProxy("auth-service", 3002))
	api.All("/menu/*", createServiceProxy("menu-service", 3003))
	api.All("/orders/*", createServiceProxy("order-service", 3004))
	api.All("/payments/*", createServiceProxy("payment-service", 3005))
	api.All("/reviews/*", createServiceProxy("review-service", 3006))
}

func createServiceProxy(service string, port int) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// THIS IS A PLACEHOLDER IMPLEMENTATION JUST TO GUIDE Y'ALL, WHEN YOU IMPLEMENT
		// PLEASE MAKE IT FORWARD TO A REAL IMPLEMENTATION S'IL VOUS PLAIT
		return c.JSON(fiber.Map{
			"proxy_to": service,
			"port":     port,
			"path":     c.Path(),
			"method":   c.Method(),
		})
	}
}
