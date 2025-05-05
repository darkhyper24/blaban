package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	"github.com/darkhyper24/blaban/order-service/internal/orders"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderService *orders.OrderService

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

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

	// Initialize order service
	db := client.Database("orderdb")
	collection := db.Collection("orders")
	orderService = orders.NewOrderService(collection)

	// Order routes
	app.Get("/api/orders", handleGetOrders)
	app.Get("/api/orders/:id", handleGetOrder)
	app.Post("/api/orders", handleCreateOrder)

	// Test route to check if the service is running
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Println("Order service started on port 8084")
	log.Fatal(app.Listen(":8084"))
}

func handleGetOrders(c *fiber.Ctx) error {
	userID, err := verifyToken(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	orders, err := orderService.GetOrders(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch orders: " + err.Error(),
		})
	}

	// Return an empty array instead of null if no orders found
	if orders == nil {
		orders = make([]models.Order, 0)
	}

	return c.JSON(fiber.Map{
		"orders": orders,
		"count":  len(orders),
	})
}

func handleGetOrder(c *fiber.Ctx) error {
	userID, err := verifyToken(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	orderID := c.Params("id")
	order, err := orderService.GetOrder(c.Context(), orderID, userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch order: " + err.Error(),
		})
	}

	return c.JSON(order)
}

func handleCreateOrder(c *fiber.Ctx) error {
	userID, err := verifyToken(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var order models.Order
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate and fetch menu items
	for i, item := range order.Items {
		menuItem, err := getMenuItem(item.ItemID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid menu item ID: %s", item.ItemID),
			})
		}
		// Update item details from menu
		order.Items[i].Name = menuItem.Name
		order.Items[i].Price = menuItem.Price
	}

	order.UserID = userID // Set the authenticated user's ID

	if err := orderService.CreateOrder(c.Context(), &order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// Helper function to validate menu items
func getMenuItem(itemID string) (*models.OrderItem, error) {
	menuURL := fmt.Sprintf("http://localhost:8083/api/menu/%s", itemID)
	resp, err := http.Get(menuURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("menu item not found (status: %d)", resp.StatusCode)
	}

	var response struct {
		Item struct {
			ID             string  `json:"id"`
			Name           string  `json:"name"`
			Price          float64 `json:"price"`
			EffectivePrice float64 `json:"effective_price"`
		} `json:"item"`
	}

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	// Use effective price if item has a discount
	price := response.Item.Price
	if response.Item.EffectivePrice > 0 && response.Item.EffectivePrice < price {
		price = response.Item.EffectivePrice
	}

	return &models.OrderItem{
		ItemID: response.Item.ID,
		Name:   response.Item.Name,
		Price:  price,
	}, nil
}

func verifyToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Missing authorization header")
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://auth-service:8082/api/auth/verify", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	if !authResp.Valid {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return authResp.UserID, nil
}
