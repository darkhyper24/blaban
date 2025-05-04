package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Simple test script to create an order and watch its status changes
func main() {
	// Connect to MongoDB
	client, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI("mongodb://localhost:27017"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Ping the database to verify connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Create a test order directly in the database
	db := client.Database("orderdb")
	collection := db.Collection("orders")

	orderID := uuid.NewString()
	now := time.Now()

	order := models.Order{
		ID:        orderID,
		UserID:    "test-user-123",
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
		Items: []models.OrderItem{
			{
				ItemID:   "test-item-123",
				Name:     "Test Item",
				Price:    9.99,
				Quantity: 2,
			},
		},
		Total: 19.98, // 9.99 * 2
	}

	_, err = collection.InsertOne(context.Background(), order)
	if err != nil {
		log.Fatalf("Failed to insert test order: %v", err)
	}

	fmt.Printf("Created test order with ID: %s\n", orderID)
	fmt.Println("Watching order status changes...")
	fmt.Println("Order will automatically move to 'processing' and then 'completed' status")

	// Poll for status changes
	currentStatus := "pending"
	for i := 0; i < 12; i++ { // Check for 60 seconds (12 * 5s)
		time.Sleep(5 * time.Second)

		// Get the current order status from the database
		var updatedOrder models.Order
		err := collection.FindOne(
			context.Background(),
			bson.M{"id": orderID},
		).Decode(&updatedOrder)

		if err != nil {
			log.Printf("Error checking order status: %v", err)
			continue
		}

		status := updatedOrder.Status

		if status != currentStatus {
			fmt.Printf("Order status changed: %s -> %s\n", currentStatus, status)
			currentStatus = status

			if status == "completed" {
				fmt.Println("Order is now completed!")
				break
			}
		} else {
			fmt.Printf("Order status still: %s\n", status)
		}
	}
}

// The following functions are kept for reference but not used in the direct DB approach

// Create a test order via the API
func createTestOrderAPI() (string, error) {
	// Create a sample order payload
	orderData := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"item_id":  "item-123", // Replace with an actual item ID in your system
				"quantity": 2,
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return "", err
	}

	// Create the request
	req, err := http.NewRequest("POST", "http://localhost:8084/api/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Add auth header if needed (mock token for testing)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token") // Replace with a valid token for your system

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response to get the order ID
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response["id"].(string), nil
}

// Get the current status of an order via API
func getOrderStatusAPI(orderID string) (string, error) {
	// Make a request to get the order
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8084/api/orders/%s", orderID), nil)
	if err != nil {
		return "", err
	}

	// Add auth header if needed
	req.Header.Set("Authorization", "Bearer test-token") // Replace with a valid token for your system

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response to get the status
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response["status"].(string), nil
}
