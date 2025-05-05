package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	collection *mongo.Collection
	mqttClient mqtt.Client
}

func NewOrderService(collection *mongo.Collection, mqttClient mqtt.Client) *OrderService {
	return &OrderService{
		collection: collection,
		mqttClient: mqttClient,
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

	// Calculate total
	var total float64
	for _, item := range order.Items {
		// Don't overwrite the original item ID
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	_, err := s.collection.InsertOne(ctx, order)
	return err
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	filter := bson.M{"id": orderID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// ScheduleOrderCompletion handles the automatic order status updates
func (s *OrderService) ScheduleOrderCompletion(orderID string) {
	// Wait for 30 seconds then mark the order as completed
	go func() {
		time.Sleep(30 * time.Second)

		ctx := context.Background()
		if err := s.UpdateOrderStatus(ctx, orderID, "completed"); err != nil {
			log.Printf("Failed to update order status: %v", err)
			return
		}

		// Fetch updated order to include in notification
		order, err := s.GetOrder(ctx, orderID, "")
		if err != nil {
			log.Printf("Error fetching order after status update: %v", err)
		}

		// Format data for notification
		data := map[string]interface{}{
			"total":      order.Total,
			"item_count": len(order.Items),
		}

		// Publish notification
		message := fmt.Sprintf("Your order #%s is ready!", orderID[:6])
		if err := s.publishOrderStatusUpdate(orderID, "completed", message, data); err != nil {
			log.Printf("Failed to publish order status update: %v", err)
		} else {
			log.Printf("Order %s marked as completed and notification sent", orderID)
		}
	}()

	// Optionally, send initial notification that order is being processed
	message := "Your order is being prepared"
	if err := s.publishOrderStatusUpdate(orderID, "pending", message, nil); err != nil {
		log.Printf("Failed to publish initial order status: %v", err)
	}

	log.Printf("Order %s scheduled for completion in 30 seconds", orderID)
}

func (s *OrderService) publishOrderStatusUpdate(orderID, status, message string, data map[string]interface{}) error {
	if s.mqttClient == nil || !s.mqttClient.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	if data == nil {
		data = map[string]interface{}{}
	}

	// Add standard fields to data
	data["order_id"] = orderID
	data["status"] = status

	payload, err := json.Marshal(map[string]interface{}{
		"order_id": orderID,
		"status":   status,
		"message":  message,
		"data":     data,
	})
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("orders/%s/status", orderID)
	token := s.mqttClient.Publish(topic, 1, false, payload)
	return token.Error()
}
