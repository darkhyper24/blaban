package orders

import (
	"context"
	"time"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	collection *mongo.Collection
}

func NewOrderService(collection *mongo.Collection) *OrderService {
	return &OrderService{collection: collection}
}

func (s *OrderService) GetOrders(ctx context.Context, userID string) ([]models.Order, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string, userID string) (*models.Order, error) {
	var order models.Order
	err := s.collection.FindOne(ctx, bson.M{
		"id":      id,
		"user_id": userID,
	}).Decode(&order)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	now := time.Now()
	order.ID = uuid.NewString()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Status = "pending"

	// Calculate total
	var total float64
	for _, item := range order.Items {
		item.ItemID = uuid.NewString()
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	_, err := s.collection.InsertOne(ctx, order)
	return err
}
