package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"auth/internal/models"
)

type AuthService struct {
	dbPool         *pgxpool.Pool
	supabaseURL    string
	supabaseKey    string
	googleRedirect string
	jwtSecret      string
}

func NewAuthService(dbPool *pgxpool.Pool) *AuthService {
	return &AuthService{
		dbPool:         dbPool,
		supabaseURL:    os.Getenv("SUPABASE_URL"),
		supabaseKey:    os.Getenv("SUPABASE_PASSWORD"),
		googleRedirect: os.Getenv("GOOGLE_REDIRECT_URL"),
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}
}

// SignUp registers a new user with email/password
func (s *AuthService) SignUp(ctx context.Context, email, password, name string) (*models.User, *models.TokenResponse, error) {
	// Create user in Supabase Auth
	tokenResp, err := s.signUpWithSupabase(email, password)
	if err != nil {
		return nil, nil, err
	}

	// Extract user ID from token
	userID := getUserIDFromToken(tokenResp.AccessToken)

	// Create user record in our database
	user, err := s.createUserRecord(ctx, userID, email, name, "email")
	if err != nil {
		return nil, nil, err
	}

	return user, tokenResp, nil
}

// Login authenticates a user with email/password
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *models.TokenResponse, error) {
	// Login with Supabase Auth
	tokenResp, err := s.loginWithSupabase(email, password)
	if err != nil {
		return nil, nil, err
	}

	// Get user from database
	userID := getUserIDFromToken(tokenResp.AccessToken)
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokenResp, nil
}

// GetGoogleAuthURL generates the Google OAuth URL
func (s *AuthService) GetGoogleAuthURL() string {
	return fmt.Sprintf(
		"https://%s/auth/v1/authorize?provider=google&redirect_to=%s",
		s.supabaseURL,
		s.googleRedirect,
	)
}

// HandleGoogleCallback processes the Google OAuth callback
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, *models.TokenResponse, error) {
	// Exchange code for tokens with Supabase
	tokenResp, err := s.exchangeCodeForToken(code)
	if err != nil {
		return nil, nil, err
	}

	// Get user info from Google
	googleUser, err := s.getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	// Check if user exists, create if not
	user, err := s.getUserByEmail(ctx, googleUser.Email)
	if err != nil {
		// Create new user
		user, err = s.createUserRecord(ctx, googleUser.ID, googleUser.Email, googleUser.Name, "google")
		if err != nil {
			return nil, nil, err
		}
	}

	return user, tokenResp, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("https://%s/auth/v1/token?grant_type=refresh_token", s.supabaseURL)

	// Prepare request body
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonData, err := json.Marshal(reqBody)
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

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh token: %s", resp.Status)
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Helper methods for Supabase API interactions
func (s *AuthService) signUpWithSupabase(email, password string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("https://%s/auth/v1/signup", s.supabaseURL)

	// Prepare request body
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonData, err := json.Marshal(reqBody)
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

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to sign up: %s - %s", resp.Status, string(bodyBytes))
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *AuthService) loginWithSupabase(email, password string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("https://%s/auth/v1/token?grant_type=password", s.supabaseURL)

	// Prepare request body
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonData, err := json.Marshal(reqBody)
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

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to login: %s - %s", resp.Status, string(bodyBytes))
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *AuthService) exchangeCodeForToken(code string) (*models.TokenResponse, error) {
	// This is a simplified version. Actual implementation would depend on Supabase's OAuth2 flow
	// For a real implementation, you would need to exchange the authorization code for tokens

	// This is a placeholder
	url := fmt.Sprintf("https://%s/auth/v1/token?grant_type=authorization_code", s.supabaseURL)

	// Prepare request body
	reqBody := map[string]string{
		"code": code,
	}
	jsonData, err := json.Marshal(reqBody)
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

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to exchange code: %s", resp.Status)
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *AuthService) getGoogleUserInfo(accessToken string) (*models.GoogleUser, error) {
	// Real implementation would call Google's API to get user info
	// This is a placeholder
	url := "https://www.googleapis.com/oauth2/v1/userinfo"

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var googleUser models.GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&googleUser)
	if err != nil {
		return nil, err
	}

	return &googleUser, nil
}

// Database operations
func (s *AuthService) createUserRecord(ctx context.Context, id, email, name, provider string) (*models.User, error) {
	// Insert user into our database
	query := `
        INSERT INTO users (id, email, name, provider, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $5)
        RETURNING id, email, name, provider, created_at, updated_at
    `

	now := time.Now()
	user := models.User{}

	err := s.dbPool.QueryRow(ctx, query, id, email, name, provider, now).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	return &user, nil
}

func (s *AuthService) getUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
        SELECT id, email, name, provider, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	user := models.User{}

	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, name, provider, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	user := models.User{}

	err := s.dbPool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Helper function to get user ID from JWT token
func getUserIDFromToken(token string) string {
	// This is a placeholder. In a real implementation, you would decode the JWT and extract the user ID
	return "user-id-from-token"
}
