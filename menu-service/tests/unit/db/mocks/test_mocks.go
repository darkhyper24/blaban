package mocks

import (
	"github.com/darkhyper24/blaban/menu-service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

// MockDB implements the MenuDBOperations interface for testing
type MockDB struct {
	mock.Mock
}

// MockAuthVerifier implements the AuthVerifier interface for testing
type MockAuthVerifier struct {
	mock.Mock
}

// VerifyManagerRole mocks the authentication verification
func (m *MockAuthVerifier) VerifyManagerRole(authHeader string) error {
	args := m.Called(authHeader)
	return args.Error(0)
}

// GetMenuItemByID mocks the database operation
func (m *MockDB) GetMenuItemByID(id string) (models.MenuItem, string, error) {
	args := m.Called(id)
	return args.Get(0).(models.MenuItem), args.String(1), args.Error(2)
}

// GetCategoryID mocks getting a category ID from name
func (m *MockDB) GetCategoryID(categoryName string) (string, error) {
	args := m.Called(categoryName)
	return args.String(0), args.Error(1)
}

// GetAllCategories mocks retrieving all categories
func (m *MockDB) GetAllCategories() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

// GetAllMenuItems mocks retrieving all menu items
func (m *MockDB) GetAllMenuItems() ([]models.MenuItem, []string, error) {
	args := m.Called()
	return args.Get(0).([]models.MenuItem), args.Get(1).([]string), args.Error(2)
}

// SearchMenuItems mocks searching for menu items
func (m *MockDB) SearchMenuItems(query string) ([]models.MenuItem, []string, []string, error) {
	args := m.Called(query)
	return args.Get(0).([]models.MenuItem), args.Get(1).([]string), args.Get(2).([]string), args.Error(3)
}

// CreateMenuItem mocks creating a menu item
func (m *MockDB) CreateMenuItem(item models.MenuItem, categoryID string) (string, error) {
	args := m.Called(item, categoryID)
	return args.String(0), args.Error(1)
}

// UpdateMenuItem mocks updating a menu item
func (m *MockDB) UpdateMenuItem(item models.MenuItem) error {
	args := m.Called(item)
	return args.Error(0)
}

// DeleteMenuItem mocks deleting a menu item
func (m *MockDB) DeleteMenuItem(itemID string) (int64, error) {
	args := m.Called(itemID)
	return args.Get(0).(int64), args.Error(1)
}

// UpdateItemDiscount mocks updating an item's discount
func (m *MockDB) UpdateItemDiscount(itemID string, discountValue float64, active bool) (int64, error) {
	args := m.Called(itemID, discountValue, active)
	return args.Get(0).(int64), args.Error(1)
}

// FilterMenuItems mocks filtering menu items
func (m *MockDB) FilterMenuItems(categoryID, minPrice, maxPrice, hasDiscount, isAvailable string) ([]models.MenuItem, []string, []string, error) {
	args := m.Called(categoryID, minPrice, maxPrice, hasDiscount, isAvailable)
	return args.Get(0).([]models.MenuItem), args.Get(1).([]string), args.Get(2).([]string), args.Error(3)
}

// GetPool mocks access to the connection pool
func (m *MockDB) GetPool() *pgxpool.Pool {
	return nil // We won't use this in these tests
}
