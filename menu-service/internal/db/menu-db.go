package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MenuDB manages database connections and operations
type MenuDB struct {
	Pool *pgxpool.Pool
}

// NewMenuDB creates a new database connection pool
func NewMenuDB() (*MenuDB, error) {
	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		dockerConn := "postgres://postgres:password@postgres:5432/menu"
		localConn := "postgres://postgres:password@localhost:5432/menu"
		// Try Docker connection first, then local if that fails
		dbPool, err := pgxpool.New(context.Background(), dockerConn)
		if err == nil {
			if err := dbPool.Ping(context.Background()); err == nil {
				fmt.Println("Connected to Docker PostgreSQL database")
				return &MenuDB{Pool: dbPool}, nil
			}
		}
		dbConnString = localConn
		fmt.Println("Docker connection failed, trying local PostgreSQL connection")
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

// GetPool returns the underlying connection pool
func (m *MenuDB) GetPool() *pgxpool.Pool {
	return m.Pool
}

// Close closes the database connection
func (m *MenuDB) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
