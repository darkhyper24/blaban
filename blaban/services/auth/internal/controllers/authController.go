package controllers

import (
	"github.com/gofiber/fiber/v2"

	"auth/internal/service"
)

type AuthController struct {
	auth *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{auth: authService}
}

// SignUp handles user registration
func (c *AuthController) SignUp(ctx *fiber.Ctx) error {
	// Parse and validate request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Email, password, and name are required"})
	}

	// Sign up user
	user, tokens, err := c.auth.SignUp(ctx.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Return success response
	return ctx.Status(201).JSON(fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

// Login handles user authentication
// Login handles user authentication
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	// Parse and validate request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	if req.Email == "" || req.Password == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}

	// Login user
	user, tokens, err := c.auth.Login(ctx.Context(), req.Email, req.Password)
	if err != nil {
		// Return the actual error message for debugging
		return ctx.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// Return success response
	return ctx.Status(200).JSON(fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    tokens.TokenType,
		"expiresIn":    tokens.ExpiresIn,
	})
}

// GetGoogleAuthURL provides the URL for Google OAuth
func (c *AuthController) GetGoogleAuthURL(ctx *fiber.Ctx) error {
	url := c.auth.GetGoogleAuthURL()
	return ctx.JSON(fiber.Map{"url": url})
}

// HandleGoogleCallback processes the OAuth callback
func (c *AuthController) HandleGoogleCallback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Authorization code required"})
	}

	user, tokens, err := c.auth.HandleGoogleCallback(ctx.Context(), code)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

// RefreshToken refreshes an access token
func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := ctx.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Refresh token is required"})
	}

	tokens, err := c.auth.RefreshToken(req.RefreshToken)
	if err != nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	return ctx.JSON(fiber.Map{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    tokens.TokenType,
		"expiresIn":    tokens.ExpiresIn,
	})
}
