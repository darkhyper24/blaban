package orders

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderListener struct {
	collection *mongo.Collection
	publisher  *mqtt.Publisher
}

func NewOrderListener(collection *mongo.Collection, publisher *mqtt.Publisher) *OrderListener {
	return &OrderListener{
		collection: collection,
		publisher:  publisher,
	}
}

func (l *OrderListener) StartListening(ctx context.Context) error {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"operationType", bson.D{
				{"$in", bson.A{"insert", "update"}},
			}},
		}}},
	}

	options := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	stream, err := l.collection.Watch(ctx, pipeline, options)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var changeEvent struct {
			OperationType string `bson:"operationType"`
			FullDocument  struct {
				ID        string    `bson:"id"`
				Status    string    `bson:"status"`
				UpdatedAt time.Time `bson:"updated_at"`
			} `bson:"fullDocument"`
		}

		if err := stream.Decode(&changeEvent); err != nil {
			log.Printf("Error decoding change event: %v", err)
			continue
		}

		// Publish notification
		message := ""
		switch changeEvent.OperationType {
		case "insert":
			message = "Order created and is pending"
		case "update":
			message = "Order status updated to " + changeEvent.FullDocument.Status
		}

		err = l.publisher.PublishOrderStatus(mqtt.OrderStatusUpdate{
			OrderID: changeEvent.FullDocument.ID,
			Status:  changeEvent.FullDocument.Status,
			Message: message,
		})

		if err != nil {
			log.Printf("Error publishing notification: %v", err)
		}
	}

	return nil
}
