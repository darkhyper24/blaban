package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	collection *mongo.Collection
}

func NewOrderService(collection *mongo.Collection, publisher *MQTTPublisher) *OrderService {
	return &OrderService{
		collection: collection,
	}
}

func (s *OrderService) GetOrders(ctx context.Context, userID string) ([]models.Order, error) {
	filter := bson.M{}
	// Only filter by user ID if it's provided
	if userID != "" {
		filter["user_id"] = userID
	}

	cursor, err := s.collection.Find(ctx, filter)
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
	filter := bson.M{"id": id}

	// Only filter by user ID if it's provided
	if userID != "" {
		filter["user_id"] = userID
	}

	err := s.collection.FindOne(ctx, filter).Decode(&order)
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

	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	_, err := s.collection.InsertOne(ctx, order)
	return err
}

// UpdateOrderStatus updates the status of an order and notifies via MQTT
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	filter := bson.M{"id": orderID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Update the order in the database
	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Publish the status update
	if s.publisher != nil {
		message := fmt.Sprintf("Order #%s status updated to %s", orderID, status)
		if err := s.publisher.PublishOrderStatus(orderID, status, message); err != nil {
			// Just log the error but don't fail the update
			fmt.Printf("Failed to publish order status update: %v\n", err)
		}
	}

	return nil
}
