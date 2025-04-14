package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Global connection pool for reuse
var dbPool *pgxpool.Pool

// Connect establishes a connection to the Supabase PostgreSQL database
func Connect() (*pgxpool.Pool, error) {
	// Return existing connection if available
	if dbPool != nil {
		return dbPool, nil
	}

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		// Try looking in different locations
		err = godotenv.Load("../.env")
		if err != nil {
			err = godotenv.Load("../../.env")
			if err != nil {
				log.Println("Warning: Could not find .env file, will use environment variables")
			}
		}
	}

	// Try using the direct PostgreSQL URL first
	connectionString := os.Getenv("POSTGRES_URL")

	// If direct URL not available, build from Supabase credentials
	if connectionString == "" {
		supabaseURL := os.Getenv("SUPABASE_URL")
		supabasePassword := os.Getenv("SUPABASE_PASSWORD")

		if supabaseURL == "" || supabasePassword == "" {
			return nil, fmt.Errorf("missing required database credentials")
		}

		// Extract project ID from URL
		var projectID string
		fmt.Sscanf(supabaseURL, "https://%[^.].supabase.co", &projectID)
		if projectID == "" {
			return nil, fmt.Errorf("invalid Supabase URL format")
		}

		connectionString = fmt.Sprintf(
			"postgres://postgres:%s@db.%s.supabase.co:5432/postgres",
			supabasePassword,
			projectID,
		)
	}

	// Configure and create the connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Set reasonable defaults for connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour

	// Create and test the connection
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to Supabase PostgreSQL database")
	dbPool = pool
	return pool, nil
}

// Close closes the database connection pool
func Close() {
	if dbPool != nil {
		dbPool.Close()
		dbPool = nil
	}
}
