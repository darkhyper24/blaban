package main

import (
	"context"
	"log"

	"auth/internal/db"
)

func main() {
	// Connect to the database
	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	var version string
	err = pool.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	log.Printf("Connected to PostgreSQL: %s", version)

	// Continue with your application...
}
