package models

import (
	"time"
)

type Order struct {
	ID        string      `json:"id" bson:"id"`
	UserID    string      `json:"user_id" bson:"user_id"`
	Items     []OrderItem `json:"items" bson:"items"`
	Total     float64     `json:"total" bson:"total"`
	Status    string      `json:"status" bson:"status"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" bson:"updated_at"`
}
