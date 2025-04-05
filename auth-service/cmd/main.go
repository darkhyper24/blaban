package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:3002/api/auth/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
)

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// Auth routes
	app.Get("/api/auth/login", handleGoogleLogin)
	app.Get("/api/auth/callback", handleGoogleCallback)
	app.Get("/api/auth/verify", verifyToken)

	log.Fatal(app.Listen(":3002"))
}

func handleGoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("state")
	return c.Redirect(url)
}

func handleGoogleCallback(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Authentication successful")
}

func verifyToken(c *fiber.Ctx) error {
	// TODO
	return c.SendString("Token verified")
}
