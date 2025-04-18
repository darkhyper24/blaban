package models

// Category represents a menu category
type Category struct {
	ID      string `json:"category_id"`
	Name    string `json:"name"`
	Picture string `json:"category_pic"`
}
