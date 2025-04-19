package models

type OrderItem struct {
	ItemID   string  `json:"item_id" bson:"item_id"`
	Name     string  `json:"name" bson:"name"`
	Quantity int     `json:"quantity" bson:"quantity"`
	Price    float64 `json:"price" bson:"price"`
}
