package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/darkhyper24/blaban/menu-service/internal/models"
	"github.com/darkhyper24/blaban/menu-service/services"
	"github.com/darkhyper24/blaban/menu-service/tests/unit/db/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleGetCategories(t *testing.T) {
	app := fiber.New()
	mockDB := new(mocks.MockDB)
	menuService := services.NewMenuService(mockDB)
	categories := []models.Category{
		{
			ID:      "cat1",
			Name:    "Appetizers",
			Picture: "appetizers.jpg",
		},
		{
			ID:      "cat2",
			Name:    "Main Courses",
			Picture: "main-courses.jpg",
		},
		{
			ID:      "cat3",
			Name:    "Desserts",
			Picture: "desserts.jpg",
		},
	}

	t.Run("Get all categories successfully", func(t *testing.T) {
		mockDB.On("GetAllCategories").Return(categories, nil).Once()
		app.Get("/api/categories", menuService.HandleGetCategories)
		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		categoriesResp, ok := result["categories"].([]interface{})
		assert.True(t, ok, "Response should have a 'categories' array field")
		assert.Len(t, categoriesResp, 3, "Response should contain 3 categories")

		firstCategory := categoriesResp[0].(map[string]interface{})
		assert.Equal(t, "cat1", firstCategory["id"], "Category ID should match")
		assert.Equal(t, "Appetizers", firstCategory["name"], "Category name should match")
		assert.Equal(t, "appetizers.jpg", firstCategory["picture"], "Category picture should match")
	})

	// Test case 2: Database error
	t.Run("Database error", func(t *testing.T) {
		mockDB.On("GetAllCategories").Return([]models.Category{}, errors.New("database connection failed")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to fetch categories", "Error message should indicate fetch failure")
	})

	mockDB.AssertExpectations(t)
}

func TestHandleGetMenu(t *testing.T) {
	app := fiber.New()

	mockDB := new(mocks.MockDB)

	menuService := services.NewMenuService(mockDB)

	items := []models.MenuItem{
		{
			ID:            "item1",
			Name:          "Burger",
			Price:         12.99,
			IsAvailable:   true,
			Quantity:      50,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		},
		{
			ID:            "item2",
			Name:          "Pizza",
			Price:         15.99,
			IsAvailable:   true,
			Quantity:      30,
			HasDiscount:   true,
			DiscountValue: 10,
			CategoryID:    "cat1",
		},
		{
			ID:            "item3",
			Name:          "Ice Cream",
			Price:         5.99,
			IsAvailable:   true,
			Quantity:      100,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat3",
		},
	}

	categoryNames := []string{"Fast Food", "Fast Food", "Desserts"}

	// Test case 1: Successful retrieval of menu
	t.Run("Get menu successfully", func(t *testing.T) {
		mockDB.On("GetAllMenuItems").Return(items, categoryNames, nil).Once()

		app.Get("/api/menu", menuService.HandleGetMenu)

		req := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		menu, ok := result["menu"].([]interface{})
		assert.True(t, ok, "Response should have a 'menu' array field")

		assert.Len(t, menu, 2, "Menu should contain 2 categories")

		var fastFoodCategory map[string]interface{}
		var dessertsCategory map[string]interface{}

		for _, cat := range menu {
			category := cat.(map[string]interface{})
			if category["name"] == "Fast Food" {
				fastFoodCategory = category
			} else if category["name"] == "Desserts" {
				dessertsCategory = category
			}
		}

		assert.NotNil(t, fastFoodCategory, "Fast Food category should exist")
		assert.Equal(t, "cat1", fastFoodCategory["id"], "Category ID should match")

		fastFoodItems, ok := fastFoodCategory["items"].([]interface{})
		assert.True(t, ok, "Category should have items array")
		assert.Len(t, fastFoodItems, 2, "Fast Food should have 2 items")

		firstItem := fastFoodItems[0].(map[string]interface{})
		assert.Equal(t, "item1", firstItem["id"], "Item ID should match")
		assert.Equal(t, "Burger", firstItem["name"], "Item name should match")
		assert.Equal(t, 12.99, firstItem["price"], "Item price should match")

		assert.NotNil(t, dessertsCategory, "Desserts category should exist")
		assert.Equal(t, "cat3", dessertsCategory["id"], "Category ID should match")

		dessertItems, ok := dessertsCategory["items"].([]interface{})
		assert.True(t, ok, "Category should have items array")
		assert.Len(t, dessertItems, 1, "Desserts should have 1 item")

		dessertItem := dessertItems[0].(map[string]interface{})
		assert.Equal(t, "item3", dessertItem["id"], "Item ID should match")
		assert.Equal(t, "Ice Cream", dessertItem["name"], "Item name should match")
	})

	// Test case 2: Database error
	t.Run("Database error", func(t *testing.T) {
		mockDB.On("GetAllMenuItems").Return([]models.MenuItem{}, []string{}, errors.New("database connection failed")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to fetch menu items", "Error message should indicate fetch failure")
	})

	// Verify that all expected mock calls were made
	mockDB.AssertExpectations(t)
}

func TestHandleGetMenuItem(t *testing.T) {
	app := fiber.New()

	mockDB := new(mocks.MockDB)

	menuService := services.NewMenuService(mockDB)

	validItem := models.MenuItem{
		ID:            "123",
		Name:          "Test Item",
		Price:         10.99,
		CategoryID:    "category1",
		IsAvailable:   true,
		Quantity:      50,
		HasDiscount:   true,
		DiscountValue: 10.0,
	}

	mockDB.On("GetMenuItemByID", "123").Return(validItem, "Test Category", nil)
	mockDB.On("GetMenuItemByID", "999").Return(models.MenuItem{}, "", errors.New("item not found"))

	app.Get("/api/menu/:id", menuService.HandleGetMenuItem)

	// Test case 1: Item exists
	t.Run("Valid item", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/menu/123", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		item, ok := result["item"].(map[string]interface{})
		assert.True(t, ok, "Response should have an 'item' field")
		assert.Equal(t, "123", item["id"], "Item ID should match")
		assert.Equal(t, "Test Item", item["name"], "Item name should match")
		assert.Equal(t, 10.99, item["price"], "Item price should match")

		category, ok := item["category"].(map[string]interface{})
		assert.True(t, ok, "Item should have a category field")
		assert.Equal(t, "category1", category["id"], "Category ID should match")
		assert.Equal(t, "Test Category", category["name"], "Category name should match")
	})

	// Test case 2: Item does not exist
	t.Run("Non-existent item", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/menu/999", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Menu item not found", errorMsg, "Error message should match")
	})

	mockDB.AssertExpectations(t)
}

func TestHandleSearchItems(t *testing.T) {
	app := fiber.New()
	mockDB := new(mocks.MockDB)
	menuService := services.NewMenuService(mockDB)

	searchItems := []models.MenuItem{
		{
			ID:            "burger1",
			Name:          "Cheese Burger",
			Price:         8.99,
			IsAvailable:   true,
			Quantity:      30,
			HasDiscount:   true,
			DiscountValue: 5.0,
			CategoryID:    "cat1",
		},
		{
			ID:            "burger2",
			Name:          "Veggie Burger",
			Price:         7.99,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		},
	}

	categoryNames := []string{"Fast Food", "Fast Food"}
	categoryPics := []string{"fast-food.jpg", "fast-food.jpg"}

	app.Get("/api/menu/search", menuService.HandleSearchItems)

	// Test case 1: Successful search
	t.Run("Search items successfully", func(t *testing.T) {
		mockDB.On("SearchMenuItems", "burger").Return(searchItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/search?q=burger", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 2, "Results should contain 2 items")
		assert.Equal(t, float64(2), result["count"], "Count should be 2")

		firstResult := results[0].(map[string]interface{})
		assert.Equal(t, "burger1", firstResult["id"], "Item ID should match")
		assert.Equal(t, "Cheese Burger", firstResult["name"], "Item name should match")
		assert.Equal(t, 8.99, firstResult["price"], "Item price should match")
		assert.Equal(t, true, firstResult["has_discount"], "Has discount should match")
		assert.Equal(t, 5.0, firstResult["discount_value"], "Discount value should match")

		assert.InDelta(t, 8.54, firstResult["effective_price"], 0.01, "Effective price should be calculated correctly")

		category, ok := firstResult["category"].(map[string]interface{})
		assert.True(t, ok, "Item should have a category field")
		assert.Equal(t, "cat1", category["id"], "Category ID should match")
		assert.Equal(t, "Fast Food", category["name"], "Category name should match")
		assert.Equal(t, "fast-food.jpg", category["picture"], "Category picture should match")
	})

	// Test case 2: Empty search query
	t.Run("Empty search query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/menu/search?q=", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Search query is required", errorMsg, "Error message should match")
	})

	// Test case 3: Database error
	t.Run("Database error", func(t *testing.T) {
		mockDB.On("SearchMenuItems", "error").Return(
			[]models.MenuItem{},
			[]string{},
			[]string{},
			errors.New("database connection failed")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/search?q=error", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to search menu items", "Error message should indicate search failure")
	})

	// Test case 4: No results found (empty results but not an error)
	t.Run("No search results", func(t *testing.T) {
		mockDB.On("SearchMenuItems", "nonexistent").Return(
			[]models.MenuItem{},
			[]string{},
			[]string{},
			nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/search?q=nonexistent", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 0, "Results should be empty")
		assert.Equal(t, float64(0), result["count"], "Count should be 0")
	})

	mockDB.AssertExpectations(t)
}

func TestHandleCreateMenuItem(t *testing.T) {
	app := fiber.New()

	mockDB := new(mocks.MockDB)
	mockAuth := new(mocks.MockAuthVerifier)

	menuService := services.NewMenuService(mockDB, mockAuth)

	categoryID := "cat1"
	itemID := "new-item-123"

	app.Post("/api/menu", menuService.HandleCreateMenuItem)

	// Test case 1: Successful creation of menu item
	t.Run("Create menu item successfully", func(t *testing.T) {
		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetCategoryID", "Burgers").Return(categoryID, nil).Once()
		mockDB.On("CreateMenuItem", mock.AnythingOfType("models.MenuItem"), categoryID).Return(itemID, nil).Once()

		requestBody := `{
			"name": "Double Cheeseburger",
			"price": 12.99,
			"category_name": "Burgers",
			"quantity": 50,
			"is_available": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.Equal(t, "Menu item created successfully", result["message"])

		item, ok := result["item"].(map[string]interface{})
		assert.True(t, ok, "Response should have an 'item' field")
		assert.NotEmpty(t, item["id"], "Item ID should not be empty")
		assert.Equal(t, "Double Cheeseburger", item["name"], "Item name should match")
		assert.Equal(t, 12.99, item["price"], "Item price should match")
		assert.Equal(t, float64(50), item["quantity"], "Item quantity should match")
		assert.Equal(t, true, item["is_available"], "Item availability should match")
		assert.Equal(t, categoryID, item["category_id"], "Category ID should match")
	})

	// Test case 2: Authentication failure
	t.Run("Authentication failure", func(t *testing.T) {
		mockAuth.On("VerifyManagerRole", "invalid-token").Return(errors.New("only managers can perform this action")).Once()
		requestBody := `{
			"name": "Double Cheeseburger",
			"price": 12.99,
			"category_name": "Burgers",
			"quantity": 50,
			"is_available": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "invalid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "only managers can perform this action", errorMsg, "Error message should match")
	})

	// Test case 3: Invalid request body
	t.Run("Invalid request body", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app := fiber.New()
		app.Post("/api/menu", menuService.HandleCreateMenuItem)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

		invalidBody := `{"`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Invalid request body", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 4: Missing required fields
	t.Run("Missing required fields", func(t *testing.T) {
		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

		requestBody := `{
			"name": "",
			"price": 0,
			"category_name": ""
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Name, price and category name are required", errorMsg, "Error message should match")
	})

	// Test case 5: Invalid category
	t.Run("Invalid category", func(t *testing.T) {
		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetCategoryID", "NonexistentCategory").Return("", errors.New("category not found")).Once()

		requestBody := `{
			"name": "Double Cheeseburger",
			"price": 12.99,
			"category_name": "NonexistentCategory",
			"quantity": 50,
			"is_available": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "category not found", errorMsg, "Error message should match")
	})

	// Test case 6: Database error on creation
	t.Run("Database error on creation", func(t *testing.T) {
		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetCategoryID", "Burgers").Return(categoryID, nil).Once()
		mockDB.On("CreateMenuItem", mock.AnythingOfType("models.MenuItem"), categoryID).Return("", errors.New("database error")).Once()

		requestBody := `{
			"name": "Double Cheeseburger",
			"price": 12.99,
			"category_name": "Burgers",
			"quantity": 50,
			"is_available": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to create menu item", "Error message should indicate creation failure")
	})

	mockDB.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

func TestHandleUpdateMenuItem(t *testing.T) {
	existingItem := models.MenuItem{
		ID:            "item-123",
		Name:          "Cheeseburger",
		Price:         9.99,
		IsAvailable:   true,
		Quantity:      25,
		HasDiscount:   false,
		DiscountValue: 0,
		CategoryID:    "cat1",
	}

	t.Run("Update menu item successfully", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Patch("/api/menu/:id", menuService.HandleUpdateMenuItem)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("GetCategoryID", "Sandwiches").Return("cat2", nil).Once()

		mockDB.On("UpdateMenuItem", mock.MatchedBy(func(item models.MenuItem) bool {
			return item.ID == "item-123" &&
				item.Name == "Deluxe Cheeseburger" &&
				item.Price == 11.99 &&
				item.CategoryID == "cat2"
		})).Return(nil).Once()

		requestBody := `{
			"name": "Deluxe Cheeseburger",
			"price": 11.99,
			"category_name": "Sandwiches"
		}`

		req := httptest.NewRequest(http.MethodPatch, "/api/menu/item-123", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.Equal(t, "Menu item updated successfully", result["message"])

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Authentication failure", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Patch("/api/menu/:id", menuService.HandleUpdateMenuItem)

		mockAuth.On("VerifyManagerRole", "invalid-token").Return(errors.New("only managers can perform this action")).Once()

		requestBody := `{
			"name": "Deluxe Cheeseburger",
			"price": 11.99
		}`

		req := httptest.NewRequest(http.MethodPatch, "/api/menu/item-123", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "invalid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "only managers can perform this action", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}

func TestHandleDeleteMenuItem(t *testing.T) {
	t.Run("Delete menu item successfully", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         9.99,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("DeleteMenuItem", "item-123").Return(int64(1), nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/item-123", nil)
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.Equal(t, "Menu item deleted successfully", result["message"])
		assert.Equal(t, "item-123", result["item_id"])

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 2: Authentication failure
	t.Run("Authentication failure", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)

		mockAuth.On("VerifyManagerRole", "invalid-token").Return(errors.New("only managers can perform this action")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/item-123", nil)
		req.Header.Set("Authorization", "invalid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "only managers can perform this action", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 3: Item not found
	t.Run("Item not found", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "nonexistent-item").Return(models.MenuItem{}, "", errors.New("item not found")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/nonexistent-item", nil)
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Menu item not found", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 4: Database error
	t.Run("Database error", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         9.99,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("DeleteMenuItem", "item-123").Return(int64(0), errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/item-123", nil)
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to delete menu item", "Error message should indicate deletion failure")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 5: Zero rows affected
	t.Run("Zero rows affected", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         9.99,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("DeleteMenuItem", "item-123").Return(int64(0), nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/item-123", nil)
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Failed to delete menu item", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 6: Missing item ID
	t.Run("Missing item ID", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)

		app.Delete("/api/menu/", menuService.HandleDeleteMenuItem)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/menu/", nil)
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Item ID is required", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}

func TestHandleAddDiscount(t *testing.T) {
	// Test case 1: Successfully apply discount
	t.Run("Apply discount successfully", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         10.00,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("UpdateItemDiscount", "item-123", 20.0, true).Return(int64(1), nil).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.Equal(t, "Discount updated successfully", result["message"])
		assert.Equal(t, "item-123", result["item_id"])

		discount, ok := result["discount"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'discount' field")
		assert.Equal(t, true, discount["active"], "Discount should be active")
		assert.Equal(t, 20.0, discount["discount_value"], "Discount value should match")
		assert.Equal(t, 10.0, discount["original_price"], "Original price should match")
		assert.InDelta(t, 8.0, discount["effective_price"], 0.001, "Effective price should be calculated correctly")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 2: Remove discount
	t.Run("Remove discount successfully", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         10.00,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   true,
			DiscountValue: 20.0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("UpdateItemDiscount", "item-123", 0.0, false).Return(int64(1), nil).Once()

		requestBody := `{
			"discount_value": 0.0,
			"active": false
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.Equal(t, "Discount updated successfully", result["message"])

		discount, ok := result["discount"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'discount' field")
		assert.Equal(t, false, discount["active"], "Discount should be inactive")
		assert.Equal(t, 0.0, discount["discount_value"], "Discount value should be 0")
		assert.Equal(t, 10.0, discount["original_price"], "Original price should match")
		assert.Equal(t, 10.0, discount["effective_price"], "Effective price should match original price")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 3: Authentication failure
	t.Run("Authentication failure", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		mockAuth.On("VerifyManagerRole", "invalid-token").Return(errors.New("only managers can perform this action")).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "invalid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "only managers can perform this action", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 4: Missing item ID
	t.Run("Missing item ID", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)

		app.Post("/api/menu//discount", menuService.HandleAddDiscount)
		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu//discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Item ID is required", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 5: Invalid request body
	t.Run("Invalid request body", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

		invalidBody := `{"`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Invalid request body", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 6: Invalid discount value
	t.Run("Invalid discount value", func(t *testing.T) {
		// Negative discount value test
		t.Run("Negative discount value", func(t *testing.T) {
			app := fiber.New()
			mockDB := new(mocks.MockDB)
			mockAuth := new(mocks.MockAuthVerifier)
			menuService := services.NewMenuService(mockDB, mockAuth)
			app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

			mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
			requestBody := `{
				"discount_value": -10.0,
				"active": true
			}`

			req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "valid-token")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)

			errorMsg, ok := result["error"].(string)
			assert.True(t, ok, "Response should have an 'error' field")
			assert.Equal(t, "Discount value must be between 0 and 100", errorMsg, "Error message should match")

			mockDB.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})

		// Over 100% discount value test
		t.Run("Over 100% discount value", func(t *testing.T) {
			app := fiber.New()
			mockDB := new(mocks.MockDB)
			mockAuth := new(mocks.MockAuthVerifier)
			menuService := services.NewMenuService(mockDB, mockAuth)
			app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

			mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()

			requestBody := `{
				"discount_value": 150.0,
				"active": true
			}`

			req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "valid-token")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)

			errorMsg, ok := result["error"].(string)
			assert.True(t, ok, "Response should have an 'error' field")
			assert.Equal(t, "Discount value must be between 0 and 100", errorMsg, "Error message should match")

			mockDB.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})
	})

	// Test case 7: Item not found
	t.Run("Item not found", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "nonexistent-item").Return(models.MenuItem{}, "", errors.New("item not found")).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/nonexistent-item/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Menu item not found", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 8: Database error
	t.Run("Database error", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         10.00,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("UpdateItemDiscount", "item-123", 20.0, true).Return(int64(0), errors.New("database error")).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to update discount", "Error message should indicate update failure")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	// Test case 9: Zero rows affected
	t.Run("Zero rows affected", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		mockAuth := new(mocks.MockAuthVerifier)
		menuService := services.NewMenuService(mockDB, mockAuth)
		app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)

		existingItem := models.MenuItem{
			ID:            "item-123",
			Name:          "Cheeseburger",
			Price:         10.00,
			IsAvailable:   true,
			Quantity:      25,
			HasDiscount:   false,
			DiscountValue: 0,
			CategoryID:    "cat1",
		}

		mockAuth.On("VerifyManagerRole", "valid-token").Return(nil).Once()
		mockDB.On("GetMenuItemByID", "item-123").Return(existingItem, "Burgers", nil).Once()
		mockDB.On("UpdateItemDiscount", "item-123", 20.0, true).Return(int64(0), nil).Once()

		requestBody := `{
			"discount_value": 20.0,
			"active": true
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/menu/item-123/discount", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "valid-token")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Equal(t, "Failed to apply discount", errorMsg, "Error message should match")

		mockDB.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})
}

func TestHandleFilterItems(t *testing.T) {
	// Test case 1: Filter by all parameters
	t.Run("Filter with all parameters", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		filteredItems := []models.MenuItem{
			{
				ID:            "item1",
				Name:          "Premium Burger",
				Price:         15.99,
				IsAvailable:   true,
				Quantity:      30,
				HasDiscount:   true,
				DiscountValue: 10,
				CategoryID:    "cat1",
			},
		}
		categoryNames := []string{"Fast Food"}
		categoryPics := []string{"fastfood.jpg"}

		mockDB.On("FilterMenuItems", "cat1", "10.0", "20.0", "true", "true").
			Return(filteredItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter?category_id=cat1&min_price=10.0&max_price=20.0&has_discount=true&is_available=true", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 1, "Results should contain 1 item")
		assert.Equal(t, float64(1), result["count"], "Count should be 1")

		filters, ok := result["filters"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'filters' field")
		assert.Equal(t, "cat1", filters["category_id"], "Category ID filter should match")
		assert.Equal(t, "10.0", filters["min_price"], "Min price filter should match")
		assert.Equal(t, "20.0", filters["max_price"], "Max price filter should match")
		assert.Equal(t, "true", filters["has_discount"], "Has discount filter should match")
		assert.Equal(t, "true", filters["is_available"], "Is available filter should match")

		item := results[0].(map[string]interface{})
		assert.Equal(t, "item1", item["id"], "Item ID should match")
		assert.Equal(t, "Premium Burger", item["name"], "Item name should match")
		assert.Equal(t, 15.99, item["price"], "Item price should match")
		assert.Equal(t, true, item["has_discount"], "Has discount should match")
		assert.Equal(t, float64(10), item["discount_value"], "Discount value should match")

		category, ok := item["category"].(map[string]interface{})
		assert.True(t, ok, "Item should have a category field")
		assert.Equal(t, "cat1", category["id"], "Category ID should match")
		assert.Equal(t, "Fast Food", category["name"], "Category name should match")
		assert.Equal(t, "fastfood.jpg", category["picture"], "Category picture should match")

		mockDB.AssertExpectations(t)
	})

	// Test case 2: Filter with no parameters (get all)
	t.Run("Filter with no parameters", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		allItems := []models.MenuItem{
			{
				ID:            "item1",
				Name:          "Burger",
				Price:         12.99,
				IsAvailable:   true,
				Quantity:      50,
				HasDiscount:   false,
				DiscountValue: 0,
				CategoryID:    "cat1",
			},
			{
				ID:            "item2",
				Name:          "Pizza",
				Price:         15.99,
				IsAvailable:   true,
				Quantity:      30,
				HasDiscount:   true,
				DiscountValue: 10,
				CategoryID:    "cat1",
			},
			{
				ID:            "item3",
				Name:          "Ice Cream",
				Price:         5.99,
				IsAvailable:   true,
				Quantity:      100,
				HasDiscount:   false,
				DiscountValue: 0,
				CategoryID:    "cat3",
			},
		}
		categoryNames := []string{"Fast Food", "Fast Food", "Desserts"}
		categoryPics := []string{"fastfood.jpg", "fastfood.jpg", "desserts.jpg"}

		mockDB.On("FilterMenuItems", "", "", "", "", "").
			Return(allItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 3, "Results should contain 3 items")
		assert.Equal(t, float64(3), result["count"], "Count should be 3")

		mockDB.AssertExpectations(t)
	})

	// Test case 3: Filter by category
	t.Run("Filter by category only", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		categoryItems := []models.MenuItem{
			{
				ID:            "item1",
				Name:          "Burger",
				Price:         12.99,
				IsAvailable:   true,
				Quantity:      50,
				HasDiscount:   false,
				DiscountValue: 0,
				CategoryID:    "cat1",
			},
			{
				ID:            "item2",
				Name:          "Pizza",
				Price:         15.99,
				IsAvailable:   true,
				Quantity:      30,
				HasDiscount:   true,
				DiscountValue: 10,
				CategoryID:    "cat1",
			},
		}
		categoryNames := []string{"Fast Food", "Fast Food"}
		categoryPics := []string{"fastfood.jpg", "fastfood.jpg"}

		mockDB.On("FilterMenuItems", "cat1", "", "", "", "").
			Return(categoryItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter?category_id=cat1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 2, "Results should contain 2 items")
		assert.Equal(t, float64(2), result["count"], "Count should be 2")

		filters, ok := result["filters"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'filters' field")
		assert.Equal(t, "cat1", filters["category_id"], "Category ID filter should match")

		mockDB.AssertExpectations(t)
	})

	// Test case 4: Filter by price range
	t.Run("Filter by price range", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		priceRangeItems := []models.MenuItem{
			{
				ID:            "item2",
				Name:          "Pizza",
				Price:         15.99,
				IsAvailable:   true,
				Quantity:      30,
				HasDiscount:   true,
				DiscountValue: 10,
				CategoryID:    "cat1",
			},
		}
		categoryNames := []string{"Fast Food"}
		categoryPics := []string{"fastfood.jpg"}

		mockDB.On("FilterMenuItems", "", "15.0", "20.0", "", "").
			Return(priceRangeItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter?min_price=15.0&max_price=20.0", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 1, "Results should contain 1 item")
		assert.Equal(t, float64(1), result["count"], "Count should be 1")

		filters, ok := result["filters"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'filters' field")
		assert.Equal(t, "15.0", filters["min_price"], "Min price filter should match")
		assert.Equal(t, "20.0", filters["max_price"], "Max price filter should match")

		mockDB.AssertExpectations(t)
	})

	// Test case 5: Filter by availability and discount
	t.Run("Filter by availability and discount", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		discountedItems := []models.MenuItem{
			{
				ID:            "item2",
				Name:          "Pizza",
				Price:         15.99,
				IsAvailable:   true,
				Quantity:      30,
				HasDiscount:   true,
				DiscountValue: 10,
				CategoryID:    "cat1",
			},
		}
		categoryNames := []string{"Fast Food"}
		categoryPics := []string{"fastfood.jpg"}

		mockDB.On("FilterMenuItems", "", "", "", "true", "true").
			Return(discountedItems, categoryNames, categoryPics, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter?has_discount=true&is_available=true", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 1, "Results should contain 1 item")
		assert.Equal(t, float64(1), result["count"], "Count should be 1")

		filters, ok := result["filters"].(map[string]interface{})
		assert.True(t, ok, "Response should have a 'filters' field")
		assert.Equal(t, "true", filters["has_discount"], "Has discount filter should match")
		assert.Equal(t, "true", filters["is_available"], "Is available filter should match")

		item := results[0].(map[string]interface{})
		assert.Equal(t, true, item["has_discount"], "Item should have a discount")
		assert.Equal(t, true, item["is_available"], "Item should be available")

		mockDB.AssertExpectations(t)
	})

	// Test case 6: Empty result set
	t.Run("Filter with no matching items", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		mockDB.On("FilterMenuItems", "nonexistent", "", "", "", "").
			Return([]models.MenuItem{}, []string{}, []string{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter?category_id=nonexistent", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		results, ok := result["results"].([]interface{})
		assert.True(t, ok, "Response should have a 'results' array field")
		assert.Len(t, results, 0, "Results should be empty")
		assert.Equal(t, float64(0), result["count"], "Count should be 0")

		mockDB.AssertExpectations(t)
	})

	// Test case 7: Database error
	t.Run("Database error", func(t *testing.T) {
		app := fiber.New()
		mockDB := new(mocks.MockDB)
		menuService := services.NewMenuService(mockDB)
		app.Get("/api/menu/filter", menuService.HandleFilterItems)

		mockDB.On("FilterMenuItems", "", "", "", "", "").
			Return([]models.MenuItem{}, []string{}, []string{}, errors.New("database connection failed")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/menu/filter", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		errorMsg, ok := result["error"].(string)
		assert.True(t, ok, "Response should have an 'error' field")
		assert.Contains(t, errorMsg, "Failed to filter menu items", "Error message should indicate filtering failure")

		mockDB.AssertExpectations(t)
	})
}
