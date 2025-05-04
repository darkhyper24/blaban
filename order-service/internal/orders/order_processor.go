package orders

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/darkhyper24/blaban/order-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
)

// OrderProcessor handles automatic order processing and status updates
type OrderProcessor struct {
	orderService *OrderService
	processing   map[string]*processingOrder
	mu           sync.Mutex
	stopChan     chan struct{}
}

// processingOrder tracks an order in the processing pipeline
type processingOrder struct {
	order       *models.Order
	processedAt time.Time
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor(orderService *OrderService) *OrderProcessor {
	return &OrderProcessor{
		orderService: orderService,
		processing:   make(map[string]*processingOrder),
		stopChan:     make(chan struct{}),
	}
}

// Start begins the order processing loop
func (p *OrderProcessor) Start() {
	// Poll for pending orders every 10 seconds
	pendingTicker := time.NewTicker(10 * time.Second)

	// Process orders that have been in "processing" state for 30+ seconds
	processingTicker := time.NewTicker(5 * time.Second)

	log.Println("Order processor started")

	go func() {
		for {
			select {
			case <-pendingTicker.C:
				p.findAndProcessPendingOrders()
			case <-processingTicker.C:
				p.completeProcessedOrders()
			case <-p.stopChan:
				pendingTicker.Stop()
				processingTicker.Stop()
				return
			}
		}
	}()
}

// Stop halts the order processor
func (p *OrderProcessor) Stop() {
	close(p.stopChan)
	log.Println("Order processor stopped")
}

// findAndProcessPendingOrders finds orders in "pending" status and moves them to "processing"
func (p *OrderProcessor) findAndProcessPendingOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find all pending orders
	filter := bson.M{"status": "pending"}
	cursor, err := p.orderService.collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Error finding pending orders: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var pendingOrders []models.Order
	if err := cursor.All(ctx, &pendingOrders); err != nil {
		log.Printf("Error decoding pending orders: %v", err)
		return
	}

	// Process each pending order
	for _, order := range pendingOrders {
		// Update order status to "processing"
		err := p.orderService.UpdateOrderStatus(ctx, order.ID, "processing")
		if err != nil {
			log.Printf("Error updating order %s to processing: %v", order.ID, err)
			continue
		}

		// Add to processing queue with current timestamp
		p.mu.Lock()
		p.processing[order.ID] = &processingOrder{
			order:       &order,
			processedAt: time.Now(),
		}
		p.mu.Unlock()

		log.Printf("Order %s moved to processing state", order.ID)
	}
}

// completeProcessedOrders completes orders that have been processing for at least 30 seconds
func (p *OrderProcessor) completeProcessedOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()

	for orderID, procOrder := range p.processing {
		// If order has been processing for at least 30 seconds
		if now.Sub(procOrder.processedAt) >= 30*time.Second {
			// Update order status to "completed"
			err := p.orderService.UpdateOrderStatus(ctx, orderID, "completed")
			if err != nil {
				log.Printf("Error completing order %s: %v", orderID, err)
				continue
			}

			log.Printf("Order %s automatically completed after 30 seconds", orderID)

			// Remove from processing queue
			delete(p.processing, orderID)
		}
	}
}
