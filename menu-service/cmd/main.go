// menu-service/cmd/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	// "github.com/google/uuid"

	"github.com/darkhyper24/blaban/menu-service/internal/db"
	"github.com/darkhyper24/blaban/menu-service/internal/models"
)

var menuDB *db.MenuDB
var redisClient *redis.Client
var cacheTTL = 15 * time.Minute

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// postgres connection
	var err error
	menuDB, err = db.NewMenuDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer menuDB.Close()

	// Redis connection
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		// Continue without Redis
	} else {
		log.Println("Connected to Redis")
	}

	// Routes
	app.Get("/api/categories", handleGetCategories)
	app.Get("/api/menu", handleGetMenu)
	app.Get("/api/menu/search", handleSearchItems)
	app.Get("/api/menu/:id", handleGetMenuItem)
	// app.Post("/api/menu", handleCreateMenuItem)           // manager only
	// app.Put("/api/menu/:id", handleUpdateMenuItem)        // manager only
	// app.Post("/api/menu/:id/discount", handleAddDiscount) // manager only
	// app.Get("/api/menu/filter", handleFilterItems)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	log.Println("Menu service started on port 8083")
	log.Fatal(app.Listen(":8083"))
}

// retrieves all menu categories
func handleGetCategories(c *fiber.Ctx) error {

	rows, err := menuDB.Pool.Query(context.Background(), `
        SELECT category_id, name, category_pic 
        FROM category
        ORDER BY name
    `)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch categories: " + err.Error(),
		})
	}
	defer rows.Close()

	categories := []fiber.Map{}

	for rows.Next() {
		var id, name, picture string

		err := rows.Scan(&id, &name, &picture)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan category: " + err.Error(),
			})
		}
		categories = append(categories, fiber.Map{
			"id":      id,
			"name":    name,
			"picture": picture,
		})
	}

	response := fiber.Map{
		"categories": categories,
	}

	return c.JSON(response)
}

// gets all menu items and the category they belong to
func handleGetMenu(c *fiber.Ctx) error {

	rows, err := menuDB.Pool.Query(context.Background(), `
        SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
               i.has_discount, i.discount_value, i.category_id, 
               c.name as category_name
        FROM items i
        JOIN category c ON i.category_id = c.category_id
        ORDER BY c.name, i.name
    `)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu items: " + err.Error(),
		})
	}
	defer rows.Close()

	menuByCategory := make(map[string]fiber.Map)

	for rows.Next() {
		var item models.MenuItem
		var categoryName string

		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.IsAvailable,
			&item.Quantity, &item.HasDiscount, &item.DiscountValue,
			&item.CategoryID, &categoryName,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan menu item: " + err.Error(),
			})
		}
		if _, exists := menuByCategory[item.CategoryID]; !exists {
			menuByCategory[item.CategoryID] = fiber.Map{
				"id":    item.CategoryID,
				"name":  categoryName,
				"items": []fiber.Map{},
			}
		}
		categoryItems := menuByCategory[item.CategoryID]["items"].([]fiber.Map)
		categoryItems = append(categoryItems, fiber.Map{
			"id":              item.ID,
			"name":            item.Name,
			"price":           item.Price,
			"effective_price": item.GetEffectivePrice(),
			"is_available":    item.IsAvailable,
			"quantity":        item.Quantity,
			"has_discount":    item.HasDiscount,
			"discount_value":  item.DiscountValue,
		})

		menuByCategory[item.CategoryID]["items"] = categoryItems
	}

	categories := make([]fiber.Map, 0, len(menuByCategory))
	for _, category := range menuByCategory {
		categories = append(categories, category)
	}

	response := fiber.Map{
		"menu": categories,
	}

	return c.JSON(response)
}

// retrieves a specific menu item by ID
func handleGetMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	var item models.MenuItem
	var categoryName string
	err := menuDB.Pool.QueryRow(context.Background(), `
        SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
               i.has_discount, i.discount_value, i.category_id, 
               c.name as category_name
        FROM items i
        JOIN category c ON i.category_id = c.category_id
        WHERE i.item_id = $1
    `, id).Scan(
		&item.ID, &item.Name, &item.Price, &item.IsAvailable,
		&item.Quantity, &item.HasDiscount, &item.DiscountValue,
		&item.CategoryID, &categoryName,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	response := fiber.Map{
		"item": fiber.Map{
			"id":              item.ID,
			"name":            item.Name,
			"price":           item.Price,
			"effective_price": item.GetEffectivePrice(),
			"is_available":    item.IsAvailable,
			"quantity":        item.Quantity,
			"has_discount":    item.HasDiscount,
			"discount_value":  item.DiscountValue,
			"category": fiber.Map{
				"id":   item.CategoryID,
				"name": categoryName,
			},
		},
	}

	return c.JSON(response)
}

// searches menu items by name
func handleSearchItems(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}
	rows, err := menuDB.Pool.Query(context.Background(), `
        SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
               i.has_discount, i.discount_value, i.category_id, 
               c.name as category_name, c.category_pic
        FROM items i
        JOIN category c ON i.category_id = c.category_id
        WHERE i.name LIKE $1 AND i.is_available = true
        ORDER BY i.name
    `, "%"+query+"%")

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search menu items: " + err.Error(),
		})
	}
	defer rows.Close()
	searchResults := []fiber.Map{}

	for rows.Next() {
		var item models.MenuItem
		var categoryName, categoryPic string

		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.IsAvailable,
			&item.Quantity, &item.HasDiscount, &item.DiscountValue,
			&item.CategoryID, &categoryName, &categoryPic,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan menu item: " + err.Error(),
			})
		}

		searchResults = append(searchResults, fiber.Map{
			"id":              item.ID,
			"name":            item.Name,
			"price":           item.Price,
			"effective_price": item.GetEffectivePrice(),
			"has_discount":    item.HasDiscount,
			"discount_value":  item.DiscountValue,
			"category": fiber.Map{
				"id":      item.CategoryID,
				"name":    categoryName,
				"picture": categoryPic,
			},
		})
	}

	return c.JSON(fiber.Map{
		"results": searchResults,
		"count":   len(searchResults),
	})
}

// Helper functions for NULL handling
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullIfZero(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nullIfNegative(i int) interface{} {
	if i < 0 {
		return nil
	}
	return i
}

func nullIfDefault(b bool, defaultVal bool) interface{} {
	if b == defaultVal {
		return nil
	}
	return b
}
