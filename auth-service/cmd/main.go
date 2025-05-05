package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/darkhyper24/blaban/auth-service/internal/db"
	"github.com/darkhyper24/blaban/auth-service/internal/tokens"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {
	// Load .env file before anything else runs
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	} else {
		log.Println("Loaded .env file")
	}
}

var (
	googleOauthConfig *oauth2.Config
	tokenService      *tokens.TokenService
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database connection
	dsn := os.Getenv("AUTH_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:admin@localhost:5432/authdb?sslmode=disable"
	}

	database, err := db.Connect(dsn)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer database.Close()

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8082/api/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Initialize token service
	tokenService = tokens.NewTokenService(
		database,
		os.Getenv("JWT_SECRET"),
		15*time.Minute, // Access token expiry
		7*24*time.Hour, // Refresh token expiry
	)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173", //local frontend server
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))
	app.Use(logger.New())

	// Auth routes
	app.Post("/api/auth/tokens", handleUserRegistration)
	app.Get("/api/auth/google/login", handleGoogleLogin)
	app.Get("/api/auth/google/callback", handleGoogleCallback)
	app.Post("/api/auth/refresh", handleRefreshToken)
	app.Get("/api/auth/verify", handleVerifyToken)
	app.Post("/api/auth/logout", handleLogout)

	// Health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Fatal(app.Listen(":8082"))
}

func handleGoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("state")
	return c.Redirect(url)
}

func handleGoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if state != "state" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	token, err := googleOauthConfig.Exchange(c.Context(), code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	client := googleOauthConfig.Client(c.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode user info",
		})
	}

	// Generate tokens
	accessToken, refreshToken, err := tokenService.GenerateTokens(googleUser.ID, "user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Store refresh token
	if err := tokenService.StoreRefreshToken(googleUser.ID, refreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}
	responseData := fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
		"user": fiber.Map{
			"id":    googleUser.ID,
			"email": googleUser.Email,
		},
	}

	// Log the response data
	jsonBytes, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
	} else {
		log.Printf("Auth response data:\n%s", string(jsonBytes))
	}

	return c.Redirect("http://localhost:5173", fiber.StatusSeeOther)
}

func handleRefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate refresh token
	userID, err := tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := tokenService.GenerateTokens(userID, "user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Revoke old refresh token and store new one
	if err := tokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		log.Printf("Failed to revoke old refresh token: %v", err)
	}

	if err := tokenService.StoreRefreshToken(userID, newRefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "bearer",
	})
}

func handleVerifyToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	// Remove "Bearer " prefix if present
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &tokens.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	claims, ok := token.Claims.(*tokens.CustomClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	return c.JSON(fiber.Map{
		"valid":   true,
		"user_id": claims.UserID,
		"roles":   claims.Role,
	})
}

func handleUserRegistration(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate tokens
	accessToken, refreshToken, err := tokenService.GenerateTokens(req.UserID, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Store refresh token
	if err := tokenService.StoreRefreshToken(req.UserID, refreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
		"expires_in":    900, // 15 minutes in seconds
	})
}

func handleLogout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	if err := tokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}
