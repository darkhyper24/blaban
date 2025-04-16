package controllers

import (
	"github.com/gofiber/fiber/v2"

	"auth/internal/service"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp handles user registration
func (c *AuthController) SignUp(ctx *fiber.Ctx) error {
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, password, and name are required",
		})
	}

	// Sign up user
	user, err := c.authService.SignUp(ctx.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":    user,
		"message": "User created successfully",
	})
}

// Login handles user authentication
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Login user
	user, tokens, err := c.authService.Login(ctx.Context(), req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    tokens.TokenType,
		"expiresIn":    tokens.ExpiresIn,
	})
}

// GetGoogleAuthURL provides the URL for Google OAuth login
func (c *AuthController) GetGoogleAuthURL(ctx *fiber.Ctx) error {
	url := c.authService.GetGoogleAuthURL()

	return ctx.JSON(fiber.Map{
		"url": url,
	})
}

// HandleGoogleCallback processes the OAuth callback from Google
func (c *AuthController) HandleGoogleCallback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Authorization code is required",
		})
	}

	user, tokens, err := c.authService.HandleGoogleCallback(ctx.Context(), code)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    tokens.TokenType,
		"expiresIn":    tokens.ExpiresIn,
	})
}

// RefreshToken refreshes an access token using a refresh token
func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	// Parse request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate input
	if req.RefreshToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	// Refresh token
	tokens, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    tokens.TokenType,
		"expiresIn":    tokens.ExpiresIn,
	})
}
