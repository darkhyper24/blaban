package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/darkhyper24/blaban/menu-service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MenuDBOperations defines the interface for database operations that need to be mockable for testing
type MenuDBOperations interface {
	GetMenuItemByID(itemID string) (models.MenuItem, string, error)
	GetCategoryID(categoryName string) (string, error)
	GetAllCategories() ([]models.Category, error)
	GetAllMenuItems() ([]models.MenuItem, []string, error)
	SearchMenuItems(query string) ([]models.MenuItem, []string, []string, error)
	CreateMenuItem(item models.MenuItem, categoryID string) (string, error)
	UpdateMenuItem(item models.MenuItem) error
	DeleteMenuItem(itemID string) (int64, error)
	UpdateItemDiscount(itemID string, discountValue float64, active bool) (int64, error)
	FilterMenuItems(categoryID, minPrice, maxPrice, hasDiscount, isAvailable string) ([]models.MenuItem, []string, []string, error)

	GetPool() *pgxpool.Pool
}

// AuthVerifier defines the interface for authentication verification operations
type AuthVerifier interface {
	VerifyManagerRole(authHeader string) error
}

// DefaultAuthVerifier is the default implementation of AuthVerifier using HTTP requests to auth service
type DefaultAuthVerifier struct{}

// VerifyManagerRole checks if the request is from a manager
func (a *DefaultAuthVerifier) VerifyManagerRole(authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("missing authorization header")
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8082/api/auth/verify", nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request")
	}
	req.Header.Add("Authorization", "Bearer "+tokenString)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify token: %v", err)
	}
	defer resp.Body.Close()

	var authResp struct {
		Valid bool   `json:"valid"`
		Roles string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response")
	}

	if !authResp.Valid || authResp.Roles != "manager" {
		return fmt.Errorf("only managers can perform this action")
	}

	return nil
}

// For backward compatibility with existing code
func VerifyManagerRole(authHeader string) error {
	verifier := &DefaultAuthVerifier{}
	return verifier.VerifyManagerRole(authHeader)
}

// GetAllCategories retrieves all menu categories
func (db *MenuDB) GetAllCategories() ([]models.Category, error) {
	rows, err := db.Pool.Query(context.Background(), `
        SELECT category_id, name, category_pic 
        FROM category
        ORDER BY name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []models.Category{}

	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Picture)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetAllMenuItems retrieves all menu items with their categories
func (db *MenuDB) GetAllMenuItems() ([]models.MenuItem, []string, error) {
	rows, err := db.Pool.Query(context.Background(), `
        SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
               i.has_discount, i.discount_value, i.category_id, 
               c.name as category_name
        FROM items i
        JOIN category c ON i.category_id = c.category_id
        ORDER BY c.name, i.name
    `)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := []models.MenuItem{}
	categoryNames := []string{}

	for rows.Next() {
		var item models.MenuItem
		var categoryName string

		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.IsAvailable,
			&item.Quantity, &item.HasDiscount, &item.DiscountValue,
			&item.CategoryID, &categoryName,
		)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, item)
		categoryNames = append(categoryNames, categoryName)
	}

	return items, categoryNames, nil
}

// SearchMenuItems searches for menu items matching the query
func (db *MenuDB) SearchMenuItems(query string) ([]models.MenuItem, []string, []string, error) {
	rows, err := db.Pool.Query(context.Background(), `
        SELECT i.item_id, i.name, i.price, i.is_available, i.quantity, 
               i.has_discount, i.discount_value, i.category_id, 
               c.name as category_name, c.category_pic
        FROM items i
        JOIN category c ON i.category_id = c.category_id
        WHERE i.name LIKE $1 AND i.is_available = true
        ORDER BY i.name
    `, "%"+query+"%")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	items := []models.MenuItem{}
	categoryNames := []string{}
	categoryPics := []string{}

	for rows.Next() {
		var item models.MenuItem
		var categoryName, categoryPic string

		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.IsAvailable,
			&item.Quantity, &item.HasDiscount, &item.DiscountValue,
			&item.CategoryID, &categoryName, &categoryPic,
		)
		if err != nil {
			return nil, nil, nil, err
		}

		items = append(items, item)
		categoryNames = append(categoryNames, categoryName)
		categoryPics = append(categoryPics, categoryPic)
	}

	return items, categoryNames, categoryPics, nil
}

// CreateMenuItem creates a new menu item
func (db *MenuDB) CreateMenuItem(item models.MenuItem, categoryID string) (string, error) {
	_, err := db.Pool.Exec(context.Background(), `
		INSERT INTO items (
			item_id, name, price, is_available, quantity, 
			has_discount, discount_value, category_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, item.ID, item.Name, item.Price, item.IsAvailable, item.Quantity,
		item.HasDiscount, item.DiscountValue, categoryID)

	if err != nil {
		return "", err
	}

	return item.ID, nil
}

// UpdateMenuItem updates an existing menu item
func (db *MenuDB) UpdateMenuItem(item models.MenuItem) error {
	_, err := db.Pool.Exec(context.Background(), `
		UPDATE items 
		SET name = $1, price = $2, is_available = $3, quantity = $4, category_id = $5,
		    has_discount = $6, discount_value = $7
		WHERE item_id = $8
	`,
		item.Name, item.Price, item.IsAvailable,
		item.Quantity, item.CategoryID, item.HasDiscount, item.DiscountValue, item.ID)

	return err
}

// DeleteMenuItem deletes a menu item by ID
func (db *MenuDB) DeleteMenuItem(itemID string) (int64, error) {
	cmdTag, err := db.Pool.Exec(context.Background(),
		"DELETE FROM items WHERE item_id = $1",
		itemID)

	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

// UpdateItemDiscount updates the discount on a menu item
func (db *MenuDB) UpdateItemDiscount(itemID string, discountValue float64, active bool) (int64, error) {
	var query string
	var args []interface{}

	if !active {
		query = "UPDATE items SET has_discount = false, discount_value = 0 WHERE item_id = $1"
		args = []interface{}{itemID}
	} else {
		query = "UPDATE items SET has_discount = true, discount_value = $2 WHERE item_id = $1"
		args = []interface{}{itemID, discountValue}
	}

	cmdTag, err := db.Pool.Exec(context.Background(), query, args...)
	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

// FilterMenuItems filters menu items based on various criteria
func (db *MenuDB) FilterMenuItems(categoryID, minPrice, maxPrice, hasDiscount, isAvailable string) ([]models.MenuItem, []string, []string, error) {
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

	rows, err := db.Pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	items := []models.MenuItem{}
	categoryNames := []string{}
	categoryPics := []string{}

	for rows.Next() {
		var item models.MenuItem
		var categoryName, categoryPic string

		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.IsAvailable,
			&item.Quantity, &item.HasDiscount, &item.DiscountValue,
			&item.CategoryID, &categoryName, &categoryPic,
		)
		if err != nil {
			return nil, nil, nil, err
		}

		items = append(items, item)
		categoryNames = append(categoryNames, categoryName)
		categoryPics = append(categoryPics, categoryPic)
	}

	return items, categoryNames, categoryPics, nil
}

// GetCategoryID gets the category ID from the category name
func (db *MenuDB) GetCategoryID(categoryName string) (string, error) {
	var categoryID string
	err := db.Pool.QueryRow(context.Background(),
		"SELECT category_id FROM category WHERE name = $1",
		categoryName).Scan(&categoryID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", fmt.Errorf("category not found: %s", categoryName)
		}
		return "", fmt.Errorf("failed to fetch category: %v", err)
	}

	return categoryID, nil
}

// GetMenuItemByID retrieves a menu item by ID from the database
func (db *MenuDB) GetMenuItemByID(itemID string) (models.MenuItem, string, error) {
	var item models.MenuItem
	var categoryName string

	err := db.Pool.QueryRow(context.Background(), `
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

// Helper functions for NULL handling
func NullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func NullIfZero(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func NullIfNegative(i int) interface{} {
	if i < 0 {
		return nil
	}
	return i
}

func NullIfDefault(b bool, defaultVal bool) interface{} {
	if b == defaultVal {
		return nil
	}
	return b
}
