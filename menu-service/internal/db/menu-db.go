package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MenuDB manages database connections and operations
type MenuDB struct {
	Pool *pgxpool.Pool
}

// NewMenuDB creates a new database connection pool
func NewMenuDB() (*MenuDB, error) {
	// Read connection string from environment variable or use default
	dbConnString := os.Getenv("MENU_DB_DSN")
	if dbConnString == "" {
		// For local development, connect to localhost
		dbConnString = "postgres://postgres:password@postgres:5432/menu?sslmode=disable"
	}

	dbPool, err := pgxpool.New(context.Background(), dbConnString)
	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &MenuDB{Pool: dbPool}, nil
}

// Close closes the database connection
func (m *MenuDB) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
