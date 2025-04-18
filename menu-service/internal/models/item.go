package models

type MenuItem struct {
	ID            string  `json:"item_id"`
	CategoryID    string  `json:"category_id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	IsAvailable   bool    `json:"is_available"`
	Quantity      int     `json:"quantity"`
	HasDiscount   bool    `json:"has_discount"`
	DiscountValue float64 `json:"discount_value"`
}

// GetEffectivePrice returns the price after applying any discount
func (m *MenuItem) GetEffectivePrice() float64 {
	if m.HasDiscount {
		return m.Price * (1 - m.DiscountValue)
	}
	return m.Price
}
