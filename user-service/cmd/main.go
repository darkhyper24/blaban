package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/darkhyper24/blaban/user-service/internal/users"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
)

func main() {

	dsn := os.Getenv("USER_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect DB:", err)
	}
	defer db.Close()

	userService := users.NewUserService(db)
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	app.Post("/api/users/signup", func(c *fiber.Ctx) error {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Bio      string `json:"bio"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate required fields
		if req.Name == "" || req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Name, email and password are required",
			})
		}

		// Set default role if not provided
		if req.Role == "" {
			req.Role = "user"
		}

		// Validate role is either "user" or "manager"
		if req.Role != "user" && req.Role != "manager" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid role. Must be either 'user' or 'manager'",
			})
		}

		user, err := userService.RegisterUser(req.Name, req.Email, req.Password, req.Role, req.Bio)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "registered",
			"user": fiber.Map{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"bio":   user.Bio,
			},
		})
	})

	app.Post("/api/users/login", func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		user, err := userService.LoginUser(req.Email, req.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Call auth service to generate tokens
		authResp, err := http.Post(
			"http://localhost:8082/api/auth/tokens",
			"application/json",
			strings.NewReader(fmt.Sprintf(`{"user_id":"%s","role":"%s"}`, user.ID, user.Role)),
		)
		if err != nil {
			log.Printf("Failed to get auth tokens: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service unavailable",
			})
		}
		defer authResp.Body.Close()

		var tokenResp map[string]interface{}
		if err := json.NewDecoder(authResp.Body).Decode(&tokenResp); err != nil {
			log.Printf("Failed to decode auth response: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid authentication response",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user": fiber.Map{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"bio":   user.Bio,
			},
			"tokens": tokenResp,
		})
	})

	app.Post("/api/users/logout", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Call auth service to revoke token
		authResp, err := http.Post(
			"http://localhost:8082/api/auth/logout",
			"application/json",
			strings.NewReader(fmt.Sprintf(`{"refresh_token":"%s"}`, req.RefreshToken)),
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service unavailable",
			})
		}
		defer authResp.Body.Close()

		if authResp.StatusCode != http.StatusOK {
			return c.Status(authResp.StatusCode).JSON(fiber.Map{
				"error": "Failed to logout",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Successfully logged out",
		})
	})

	app.Get("/api/users/profile", handleGetProfile)
	app.Put("/api/users/profile", handleUpdateProfile)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Fatal(app.Listen(":8081"))
}

func handleGetProfile(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Get user profile")
}

func handleUpdateProfile(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Update user profile")
}
