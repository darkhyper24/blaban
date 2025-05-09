// menu-service/cmd/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/darkhyper24/blaban/menu-service/internal/db"
	"github.com/darkhyper24/blaban/menu-service/internal/routes"
)

var redisClient *redis.Client
var cacheTTL = 15 * time.Minute

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	menuDB, err := db.NewMenuDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer menuDB.Close()

	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis")
	}

	routes.SetupRoutes(app, menuDB)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	log.Println("Menu service started on port 8083")
	log.Fatal(app.Listen(":8083"))
}
