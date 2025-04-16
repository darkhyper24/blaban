package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"

	"auth/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	dbPool         *pgxpool.Pool
	supabaseURL    string
	supabaseKey    string
	googleRedirect string
	jwtSecret      string
}

// Helper function to get user ID from JWT token
func getUserIDFromToken(tokenString string) string {
	// Split the token to get the payload part
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		// Not a valid JWT token format
		return uuid.New().String() // Fallback to random UUID if token format is invalid
	}

	// Decode the payload part (second part of the JWT)
	payload, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return uuid.New().String()
	}

	// Parse the payload as JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.New().String()
	}

	// Supabase tokens store the user ID in the "sub" claim
	if sub, ok := claims["sub"].(string); ok {
		return sub // This is the Supabase auth user ID
	}

	// Fallback if we can't extract the ID
	return uuid.New().String()
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

func (s *AuthService) getSupabaseAPIURL(path string) string {
	// Remove https:// if present in the base URL
	baseURL := strings.TrimPrefix(s.supabaseURL, "https://")

	// Ensure path starts with slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("https://%s%s", baseURL, path)
}

// SignUp registers a new user with email/password
func (s *AuthService) SignUp(ctx context.Context, email, password, name string) (*models.Users, error) {
	// Create user in Supabase Auth
	tokenResp, err := s.signUpWithSupabase(email, password)
	if err != nil {
		return nil, err
	}

	// Extract user ID from token
	userID := getUserIDFromToken(tokenResp.AccessToken)

	// Create user record in our database
	user, err := s.createUserRecord(ctx, userID, email, name)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user with email/password
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.Users, *models.TokenResponse, error) {
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
	baseURL := strings.TrimPrefix(s.supabaseURL, "https://")
	return fmt.Sprintf(
		"https://%s/auth/v1/authorize?provider=google&redirect_to=%s",
		baseURL,
		s.googleRedirect,
	)
}

// HandleGoogleCallback processes the Google OAuth callback
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.Users, *models.TokenResponse, error) {
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
		user, err = s.createUserRecord(ctx, googleUser.ID, googleUser.Email, googleUser.Name)
		if err != nil {
			return nil, nil, err
		}
	}

	return user, tokenResp, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	url := s.getSupabaseAPIURL("/auth/v1/token?grant_type=refresh_token")

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
	url := s.getSupabaseAPIURL("/auth/v1/signup")

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
	url := s.getSupabaseAPIURL("/auth/v1/token?grant_type=password")

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
	url := s.getSupabaseAPIURL("/auth/v1/token?grant_type=authorization_code")

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
func (s *AuthService) createUserRecord(ctx context.Context, id string, email string, name string) (*models.Users, error) {
	// Insert user into our database
	query := `
        INSERT INTO users (id, email, name)
        VALUES ($1, $2, $3)
        RETURNING id, email, name
    `

	user := models.Users{}

	err := s.dbPool.QueryRow(ctx, query, id, email, name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	return &user, nil
}

func (s *AuthService) getUserByID(ctx context.Context, id string) (*models.Users, error) {
	query := `
        SELECT id, email, name 
        FROM users
        WHERE id = $1
    `

	user := models.Users{}

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

func (s *AuthService) getUserByEmail(ctx context.Context, email string) (*models.Users, error) {
	query := `
        SELECT id, email, name
        FROM users
        WHERE email = $1
    `

	user := models.Users{}

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
