// menu-service/cmd/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"

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
	app.Get("/api/menu/filter", handleFilterItems)
	app.Get("/api/menu/:id", handleGetMenuItem)
	app.Post("/api/menu", handleCreateMenuItem)
	app.Patch("/api/menu/:id", handleUpdateMenuItem)
	app.Delete("/api/menu/:id", handleDeleteMenuItem)
	app.Post("/api/menu/:id/discount", handleAddDiscount)

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

// getMenuItemByID retrieves a menu item by ID from the database
// Returns the item, category name, and any error that occurred
func getMenuItemByID(itemID string) (models.MenuItem, string, error) {
	var item models.MenuItem
	var categoryName string

	err := menuDB.Pool.QueryRow(context.Background(), `
		SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
			i.has_discount, i.discount_value, i.category_id, 
			c.name as category_name
		FROM items i
		JOIN category c ON i.category_id = c.category_id
		WHERE i.item_id = $1
	`, itemID).Scan(
		&item.ID, &item.Name, &item.Price, &item.IsAvailable,
		&item.Quantity, &item.HasDiscount, &item.DiscountValue,
		&item.CategoryID, &categoryName,
	)

	return item, categoryName, err
}

// retrieves a specific menu item by ID
func handleGetMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	item, categoryName, err := getMenuItemByID(id)

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

func handleCreateMenuItem(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create auth request",
		})
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify token: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid bool   `json:"valid"`
		Roles string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode auth response",
		})
	}

	if !authResp.Valid || authResp.Roles != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only managers can create menu items",
		})
	}

	var item struct {
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		CategoryName string  `json:"category_name"`
		Quantity     int     `json:"quantity"`
		IsAvailable  bool    `json:"is_available"`
	}

	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if item.Name == "" || item.Price <= 0 || item.CategoryName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, price and category name are required",
		})
	}

	var categoryID string
	err = menuDB.Pool.QueryRow(context.Background(),
		"SELECT category_id FROM category WHERE name = $1",
		item.CategoryName).Scan(&categoryID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Category not found: " + item.CategoryName,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch category: " + err.Error(),
		})
	}

	itemID := uuid.New().String()

	if item.Quantity <= 0 {
		item.Quantity = 0
	}
	if !item.IsAvailable {
		item.IsAvailable = true
	}

	_, err = menuDB.Pool.Exec(context.Background(), `
		INSERT INTO items (
			item_id, name, price, is_available, quantity, 
			has_discount, discount_value, category_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, itemID, item.Name, item.Price, item.IsAvailable, item.Quantity, false, 0, categoryID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create menu item: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Menu item created successfully",
		"item": fiber.Map{
			"id":           itemID,
			"name":         item.Name,
			"price":        item.Price,
			"is_available": item.IsAvailable,
			"quantity":     item.Quantity,
			"has_discount": false,
			"category_id":  categoryID,
		},
	})
}

func handleUpdateMenuItem(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create auth request",
		})
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify token: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid bool   `json:"valid"`
		Roles string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode auth response",
		})
	}

	if !authResp.Valid || authResp.Roles != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only managers can update menu items",
		})
	}

	itemID := c.Params("id")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	currentItem, categoryName, err := getMenuItemByID(itemID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	var update struct {
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		CategoryName string  `json:"category_name"`
		Quantity     *int    `json:"quantity"`
		IsAvailable  *bool   `json:"is_available"`
	}

	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updatedCategoryID := currentItem.CategoryID
	if update.CategoryName != "" {
		err = menuDB.Pool.QueryRow(context.Background(),
			"SELECT category_id FROM category WHERE name = $1",
			update.CategoryName).Scan(&updatedCategoryID)

		if err != nil {
			if err.Error() == "no rows in result set" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Category not found: " + update.CategoryName,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch category: " + err.Error(),
			})
		}

		categoryName = update.CategoryName
	}

	if update.Name != "" {
		currentItem.Name = update.Name
	}

	if update.Price > 0 {
		currentItem.Price = update.Price
	}

	if update.Quantity != nil {
		currentItem.Quantity = *update.Quantity
	}

	if update.IsAvailable != nil {
		currentItem.IsAvailable = *update.IsAvailable
	}

	_, err = menuDB.Pool.Exec(context.Background(), `
		UPDATE items 
		SET name = $1, price = $2, is_available = $3, quantity = $4, category_id = $5 
		WHERE item_id = $6
	`,
		currentItem.Name, currentItem.Price, currentItem.IsAvailable,
		currentItem.Quantity, updatedCategoryID, itemID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu item: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Menu item updated successfully",
		"item": fiber.Map{
			"id":              itemID,
			"name":            currentItem.Name,
			"price":           currentItem.Price,
			"effective_price": currentItem.GetEffectivePrice(),
			"is_available":    currentItem.IsAvailable,
			"quantity":        currentItem.Quantity,
			"has_discount":    currentItem.HasDiscount,
			"discount_value":  currentItem.DiscountValue,
			"category": fiber.Map{
				"id":   updatedCategoryID,
				"name": categoryName,
			},
		},
	})
}

func handleDeleteMenuItem(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create auth request",
		})
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify token: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid bool   `json:"valid"`
		Roles string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode auth response",
		})
	}

	if !authResp.Valid || authResp.Roles != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only managers can delete menu items",
		})
	}

	itemID := c.Params("id")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	var exists bool
	err = menuDB.Pool.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM items WHERE item_id = $1)",
		itemID).Scan(&exists)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query item: " + err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	cmdTag, err := menuDB.Pool.Exec(context.Background(),
		"DELETE FROM items WHERE item_id = $1",
		itemID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu item: " + err.Error(),
		})
	}

	if cmdTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu item",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Menu item deleted successfully",
		"item_id": itemID,
	})
}

func handleAddDiscount(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create auth request",
		})
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify token: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid bool   `json:"valid"`
		Roles string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode auth response",
		})
	}

	if !authResp.Valid || authResp.Roles != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only managers can apply discounts",
		})
	}

	itemID := c.Params("id")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	var discountReq struct {
		DiscountValue float64 `json:"discount_value"`
		Active        bool    `json:"active"`
	}

	if err := c.BodyParser(&discountReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if discountReq.Active && (discountReq.DiscountValue <= 0 || discountReq.DiscountValue >= 100) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Discount value must be between 0 and 100",
		})
	}

	currentItem, _, err := getMenuItemByID(itemID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	var query string
	var args []interface{}

	if !discountReq.Active {
		query = "UPDATE items SET has_discount = false, discount_value = 0 WHERE item_id = $1"
		args = []interface{}{itemID}
	} else {
		query = "UPDATE items SET has_discount = true, discount_value = $2 WHERE item_id = $1"
		args = []interface{}{itemID, discountReq.DiscountValue}
	}

	cmdTag, err := menuDB.Pool.Exec(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update discount: " + err.Error(),
		})
	}

	if cmdTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to apply discount",
		})
	}

	var effectivePrice float64
	if discountReq.Active {
		effectivePrice = currentItem.Price * (1 - discountReq.DiscountValue/100)
	} else {
		effectivePrice = currentItem.Price
	}

	var responseDiscountValue float64
	if discountReq.Active {
		responseDiscountValue = discountReq.DiscountValue
	} else {
		responseDiscountValue = 0
	}

	return c.JSON(fiber.Map{
		"message": "Discount updated successfully",
		"item_id": itemID,
		"discount": fiber.Map{
			"active":          discountReq.Active,
			"discount_value":  responseDiscountValue,
			"original_price":  currentItem.Price,
			"effective_price": effectivePrice,
		},
	})
}

// handles filtering menu items based on category, price, discount, and availability
func handleFilterItems(c *fiber.Ctx) error {
	categoryID := c.Query("category_id")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	hasDiscount := c.Query("has_discount")
	isAvailable := c.Query("is_available")

	query := `
		SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
			   i.has_discount, i.discount_value, i.category_id, 
			   c.name as category_name, c.category_pic
		FROM items i
		JOIN category c ON i.category_id = c.category_id
		WHERE 1=1
	`
	var args []interface{}
	argPosition := 1

	if categoryID != "" {
		query += fmt.Sprintf(" AND i.category_id = $%d", argPosition)
		args = append(args, categoryID)
		argPosition++
	}

	if minPrice != "" {
		minPriceFloat, err := strconv.ParseFloat(minPrice, 64)
		if err == nil && minPriceFloat > 0 {
			query += fmt.Sprintf(" AND i.price >= $%d", argPosition)
			args = append(args, minPriceFloat)
			argPosition++
		}
	}

	if maxPrice != "" {
		maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64)
		if err == nil && maxPriceFloat > 0 {
			query += fmt.Sprintf(" AND i.price <= $%d", argPosition)
			args = append(args, maxPriceFloat)
			argPosition++
		}
	}

	if hasDiscount == "true" {
		query += " AND i.has_discount = true"
	} else if hasDiscount == "false" {
		query += " AND i.has_discount = false"
	}

	if isAvailable == "true" {
		query += " AND i.is_available = true"
	} else if isAvailable == "false" {
		query += " AND i.is_available = false"
	}

	query += " ORDER BY c.name, i.name"

	rows, err := menuDB.Pool.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to filter menu items: " + err.Error(),
		})
	}
	defer rows.Close()

	results := []fiber.Map{}
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

		results = append(results, fiber.Map{
			"id":              item.ID,
			"name":            item.Name,
			"price":           item.Price,
			"effective_price": item.GetEffectivePrice(),
			"is_available":    item.IsAvailable,
			"quantity":        item.Quantity,
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
		"results": results,
		"count":   len(results),
		"filters": fiber.Map{
			"category_id":  categoryID,
			"min_price":    minPrice,
			"max_price":    maxPrice,
			"has_discount": hasDiscount,
			"is_available": isAvailable,
		},
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
