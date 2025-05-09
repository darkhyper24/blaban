package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/darkhyper24/blaban/menu-service/internal/db"
	"github.com/darkhyper24/blaban/menu-service/internal/models"
)

// MenuService handles all menu related operations
type MenuService struct {
	DB       db.MenuDBOperations
	Verifier db.AuthVerifier
}

//  creates a new menu service with DB dependency injected
func NewMenuService(dbOps db.MenuDBOperations, verifier ...db.AuthVerifier) *MenuService {
	var authVerifier db.AuthVerifier
	if len(verifier) > 0 && verifier[0] != nil {
		authVerifier = verifier[0]
	} else {
		authVerifier = &db.DefaultAuthVerifier{}
	}

	return &MenuService{
		DB:       dbOps,
		Verifier: authVerifier,
	}
}

// HandleGetCategories retrieves all menu categories
func (s *MenuService) HandleGetCategories(c *fiber.Ctx) error {
	categories, err := s.DB.GetAllCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch categories: " + err.Error(),
		})
	}

	categoriesMap := make([]fiber.Map, len(categories))
	for i, category := range categories {
		categoriesMap[i] = fiber.Map{
			"id":      category.ID,
			"name":    category.Name,
			"picture": category.Picture,
		}
	}

	response := fiber.Map{
		"categories": categoriesMap,
	}

	return c.JSON(response)
}

// HandleGetMenu gets all menu items and the category they belong to
func (s *MenuService) HandleGetMenu(c *fiber.Ctx) error {
	items, categoryNames, err := s.DB.GetAllMenuItems()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu items: " + err.Error(),
		})
	}

	menuByCategory := make(map[string]fiber.Map)

	for i, item := range items {
		categoryName := categoryNames[i]

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

// HandleGetMenuItem retrieves a specific menu item by ID
func (s *MenuService) HandleGetMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	item, categoryName, err := s.DB.GetMenuItemByID(id)

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

// HandleSearchItems searches menu items by name
func (s *MenuService) HandleSearchItems(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	items, categoryNames, categoryPics, err := s.DB.SearchMenuItems(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search menu items: " + err.Error(),
		})
	}

	searchResults := []fiber.Map{}
	for i, item := range items {
		searchResults = append(searchResults, fiber.Map{
			"id":              item.ID,
			"name":            item.Name,
			"price":           item.Price,
			"effective_price": item.GetEffectivePrice(),
			"has_discount":    item.HasDiscount,
			"discount_value":  item.DiscountValue,
			"category": fiber.Map{
				"id":      item.CategoryID,
				"name":    categoryNames[i],
				"picture": categoryPics[i],
			},
		})
	}

	return c.JSON(fiber.Map{
		"results": searchResults,
		"count":   len(searchResults),
	})
}

// HandleCreateMenuItem creates a new menu item
func (s *MenuService) HandleCreateMenuItem(c *fiber.Ctx) error {
	err := s.Verifier.VerifyManagerRole(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var requestItem struct {
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		CategoryName string  `json:"category_name"`
		Quantity     int     `json:"quantity"`
		IsAvailable  bool    `json:"is_available"`
	}

	if err := c.BodyParser(&requestItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if requestItem.Name == "" || requestItem.Price <= 0 || requestItem.CategoryName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, price and category name are required",
		})
	}

	categoryID, err := s.DB.GetCategoryID(requestItem.CategoryName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	itemID := uuid.New().String()
	if requestItem.Quantity <= 0 {
		requestItem.Quantity = 0
	}
	if !requestItem.IsAvailable {
		requestItem.IsAvailable = true
	}

	item := models.MenuItem{
		ID:            itemID,
		Name:          requestItem.Name,
		Price:         requestItem.Price,
		IsAvailable:   requestItem.IsAvailable,
		Quantity:      requestItem.Quantity,
		HasDiscount:   false,
		DiscountValue: 0,
		CategoryID:    categoryID,
	}

	_, err = s.DB.CreateMenuItem(item, categoryID)
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

// HandleUpdateMenuItem updates an existing menu item
func (s *MenuService) HandleUpdateMenuItem(c *fiber.Ctx) error {
	err := s.Verifier.VerifyManagerRole(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	itemID := c.Params("id")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	currentItem, categoryName, err := s.DB.GetMenuItemByID(itemID)
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
		var categoryErr error
		updatedCategoryID, categoryErr = s.DB.GetCategoryID(update.CategoryName)
		if categoryErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": categoryErr.Error(),
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

	currentItem.CategoryID = updatedCategoryID

	err = s.DB.UpdateMenuItem(currentItem)
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

// HandleDeleteMenuItem deletes a menu item
func (s *MenuService) HandleDeleteMenuItem(c *fiber.Ctx) error {
	err := s.Verifier.VerifyManagerRole(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	itemID := c.Params("id")
	if itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Item ID is required",
		})
	}

	_, _, err = s.DB.GetMenuItemByID(itemID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	rowsAffected, err := s.DB.DeleteMenuItem(itemID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu item: " + err.Error(),
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete menu item",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Menu item deleted successfully",
		"item_id": itemID,
	})
}

// HandleAddDiscount applies or removes a discount from a menu item
func (s *MenuService) HandleAddDiscount(c *fiber.Ctx) error {
	err := s.Verifier.VerifyManagerRole(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
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

	currentItem, _, err := s.DB.GetMenuItemByID(itemID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Menu item not found",
		})
	}

	rowsAffected, err := s.DB.UpdateItemDiscount(itemID, discountReq.DiscountValue, discountReq.Active)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update discount: " + err.Error(),
		})
	}

	if rowsAffected == 0 {
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

// HandleFilterItems handles filtering menu items based on category, price, discount, and availability
func (s *MenuService) HandleFilterItems(c *fiber.Ctx) error {
	categoryID := c.Query("category_id")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	hasDiscount := c.Query("has_discount")
	isAvailable := c.Query("is_available")

	items, categoryNames, categoryPics, err := s.DB.FilterMenuItems(
		categoryID, minPrice, maxPrice, hasDiscount, isAvailable)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to filter menu items: " + err.Error(),
		})
	}

	results := []fiber.Map{}
	for i, item := range items {
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
				"name":    categoryNames[i],
				"picture": categoryPics[i],
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
