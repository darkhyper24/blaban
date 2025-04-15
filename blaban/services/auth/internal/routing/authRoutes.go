package routing

import (
	"github.com/gofiber/fiber/v2"

	"auth/internal/controllers"
)

func SetupRoutes(app *fiber.App, authController *controllers.AuthController) {
	// Root route handler (add this)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Auth API",
			"status":  "running",
			"version": "1.0.0",
		})
	})

	// Auth routes
	auth := app.Group("/auth")

	// Registration and login routes
	auth.Post("/signup", authController.SignUp)
	auth.Post("/login", authController.Login)

	// Token refresh route
	auth.Post("/refresh", authController.RefreshToken)

	// Google OAuth routes
	auth.Get("/google", authController.GetGoogleAuthURL)
	auth.Get("/google/callback", authController.HandleGoogleCallback)
}
