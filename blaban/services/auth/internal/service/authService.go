package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"auth/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	db          *pgxpool.Pool
	supabaseURL string
	supabaseKey string
	redirectURL string
}

// NewAuthService creates a new authentication service
func NewAuthService(dbPool *pgxpool.Pool) *AuthService {
	return &AuthService{
		db:          dbPool,
		supabaseURL: os.Getenv("SUPABASE_URL"),
		supabaseKey: os.Getenv("SUPABASE_PASSWORD"),
		redirectURL: os.Getenv("GOOGLE_REDIRECT_URL"),
	}
}

// SignUp registers a new user with email and password
func (s *AuthService) SignUp(ctx context.Context, email, password, name string) (*models.Users, *models.TokenResponse, error) {
	// Register with Supabase
	tokens, err := s.supabaseRequest("/auth/v1/signup", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("signup failed: %w", err)
	}

	// Extract user ID from JWT token
	userID := s.getUserIDFromToken(tokens.AccessToken)

	// Create user record in our database
	user, err := s.saveUser(ctx, userID, email, name)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login authenticates a user
// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.Users, *models.TokenResponse, error) {
	// Login with Supabase
	tokens, err := s.supabaseRequest("/auth/v1/token?grant_type=password", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("supabase auth failed: %w", err)
	}

	// Get user from our database
	userID := s.getUserIDFromToken(tokens.AccessToken)
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		// User exists in Supabase but not in our database
		// Try to create them automatically
		user, err = s.saveUser(ctx, userID, email, email) // Use email as temporary name
		if err != nil {
			return nil, nil, fmt.Errorf("user authenticated but not found in database: %w", err)
		}
	}

	return user, tokens, nil
}

// GetGoogleAuthURL returns the Google OAuth URL
func (s *AuthService) GetGoogleAuthURL() string {
	baseURL := strings.TrimPrefix(s.supabaseURL, "https://")
	return fmt.Sprintf("https://%s/auth/v1/authorize?provider=google&redirect_to=%s",
		baseURL, s.redirectURL)
}

// HandleGoogleCallback processes the OAuth callback
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.Users, *models.TokenResponse, error) {
	// Exchange code for tokens
	tokens, err := s.supabaseRequest("/auth/v1/token?grant_type=authorization_code",
		map[string]string{"code": code})
	if err != nil {
		return nil, nil, err
	}

	// Get user info
	googleUser, err := s.getGoogleUserInfo(tokens.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	// Check if user exists, create if not
	user, err := s.getUserByEmail(ctx, googleUser.Email)
	if err != nil {
		// Create new user
		user, err = s.saveUser(ctx, googleUser.ID, googleUser.Email, googleUser.Name)
		if err != nil {
			return nil, nil, err
		}
	}

	return user, tokens, nil
}

// RefreshToken gets a new access token
func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	return s.supabaseRequest("/auth/v1/token?grant_type=refresh_token",
		map[string]string{"refresh_token": refreshToken})
}

// Helper methods

// supabaseRequest makes a request to the Supabase API
// supabaseRequest makes a request to the Supabase API
func (s *AuthService) supabaseRequest(path string, data map[string]string) (*models.TokenResponse, error) {
	// Prepare URL
	baseURL := strings.TrimPrefix(s.supabaseURL, "https://")
	url := fmt.Sprintf("https://%s%s", baseURL, path)

	// For debugging
	fmt.Printf("Making request to: %s\n", url)

	// Create request body
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.supabaseKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for errors and read response body for error details
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s - %s",
			resp.StatusCode, resp.Status, string(bodyBytes))
	}

	// Parse response
	var tokenResp models.TokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getUserIDFromToken extracts the user ID from a JWT token
func (s *AuthService) getUserIDFromToken(token string) string {
	// Quick and simple token parsing
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return uuid.New().String()
	}

	// Decode the payload
	payload, err := base64Decode(parts[1])
	if err != nil {
		return uuid.New().String()
	}

	// Extract the user ID
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.New().String()
	}

	if sub, ok := claims["sub"].(string); ok {
		return sub
	}

	return uuid.New().String()
}

// getGoogleUserInfo gets user info from Google
func (s *AuthService) getGoogleUserInfo(token string) (*models.GoogleUser, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info")
	}

	var user models.GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Database methods

// saveUser saves a user to the database
func (s *AuthService) saveUser(ctx context.Context, id, email, name string) (*models.Users, error) {
	user := models.Users{}
	err := s.db.QueryRow(ctx,
		"INSERT INTO users (id, email, name) VALUES ($1, $2, $3) RETURNING id, email, name",
		id, email, name,
	).Scan(&user.ID, &user.Email, &user.Name)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// getUserByID gets a user by ID
func (s *AuthService) getUserByID(ctx context.Context, id string) (*models.Users, error) {
	user := models.Users{}
	err := s.db.QueryRow(ctx,
		"SELECT id, email, name FROM users WHERE id = $1", id,
	).Scan(&user.ID, &user.Email, &user.Name)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// getUserByEmail gets a user by email
func (s *AuthService) getUserByEmail(ctx context.Context, email string) (*models.Users, error) {
	user := models.Users{}
	err := s.db.QueryRow(ctx,
		"SELECT id, email, name FROM users WHERE email = $1", email,
	).Scan(&user.ID, &user.Email, &user.Name)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// base64Decode decodes a base64-encoded string
func base64Decode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.RawURLEncoding.DecodeString(s)
}
