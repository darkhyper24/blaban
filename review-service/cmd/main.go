package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/api/reviews", func(c *fiber.Ctx) error {
		return c.SendString("Reviews service")
	})

	log.Fatal(app.Listen(":3003"))
}
