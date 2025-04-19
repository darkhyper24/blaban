package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Initialize order service
	collection := client.Database("orderdb").Collection("orders")
	orderService = orders.NewOrderService(collection)

	// Order routes
	app.Get("/api/orders", handleGetOrders)
	app.Get("/api/orders/:id", handleGetOrder)
	app.Post("/api/orders", handleCreateOrder)

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
			"error": "Failed to fetch orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
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
			"error": "Failed to fetch order",
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
	resp, err := http.Get(fmt.Sprintf("http://localhost:8083/api/menu/items/%s", itemID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("menu item not found")
	}

	var menuItem models.OrderItem
	if err := json.NewDecoder(resp.Body).Decode(&menuItem); err != nil {
		return nil, err
	}

	return &menuItem, nil
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
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
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
