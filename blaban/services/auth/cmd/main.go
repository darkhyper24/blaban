package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"auth/internal/controllers"
	"auth/internal/db"
	"auth/internal/routing"
	"auth/internal/service"
)

func main() {
	// Connect to the database
	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create service and controller
	authService := service.NewAuthService(pool)
	authController := controllers.NewAuthController(authService)

	// Create Fiber app
	app := fiber.New()

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	routing.SetupRoutes(app, authController)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting auth service on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
