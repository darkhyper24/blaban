package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
	// rate limiter b2a w keda
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // we're rate limiting by IP address
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	}))

	setupRoutes(app)

	log.Fatal(app.Listen("0.0.0.0:8080"))
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
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return func(c *fiber.Ctx) error {
		// get service URL from environment variable
		serviceURL := os.Getenv(strings.ToUpper(service) + "_URL")
		if serviceURL == "" {
			// fallback to local development
			serviceURL = fmt.Sprintf("http://%s:%d", service, port)
		}

		// build the target URL
		path := strings.TrimPrefix(c.Path(), "/api"+"/"+strings.Split(service, "-")[0])
		targetURL := serviceURL + path

		// create a new request
		req, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create request",
				"details": err.Error(),
			})
		}

		// copy headers
		for key, values := range c.GetReqHeaders() {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// execute the request
		resp, err := client.Do(req)
		if err != nil {
			// implement circuit breaker pattern
			log.Printf("Service %s is unreachable: %v", service, err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service temporarily unavailable",
			})
		}
		defer resp.Body.Close()

		// read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read response",
			})
		}

		// set response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Set(key, value)
			}
		}

		// return response
		return c.Status(resp.StatusCode).Send(body)
	}
}
