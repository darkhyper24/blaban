# Order Service Notification System Integration

The Order Service now integrates with the Notification Service to provide real-time order status updates to clients.

## Overview

When an order status changes (e.g., from "pending" to "processing" or "completed"), the Order Service publishes a message to the MQTT broker. The Notification Service subscribes to these messages and forwards them to connected clients via WebSockets.

## Automatic Order Processing

The system now includes an automatic order processor that uses timers to simulate the order preparation process:

1. When an order is created, it is initially set to "pending" status
2. Every 10 seconds, the system scans for pending orders and moves them to "processing" status
3. Once an order has been in "processing" status for 30 seconds, it is automatically moved to "completed" status
4. Each status change triggers an MQTT notification that is delivered to clients via WebSockets

This automated flow simulates a real-world restaurant order process where orders are prepared and completed after a certain time.

## Requirements

- MQTT Broker (e.g., Mosquitto) running on port 1883
- Notification Service running on port 8085

## Usage

### Updating Order Status

The Order Service automatically publishes status updates in the following scenarios:

1. When an order is created (status: "pending")
2. When the order status is updated automatically by the timer-based processor
3. When the order status is updated manually via the API

### API Endpoint for Status Updates

```
PUT /api/orders/:id/status
```

Request body:
```json
{
  "status": "completed"
}
```

Possible status values:
- "pending" - Order has been received but not processed
- "processing" - Order is being prepared
- "completed" - Order is ready for pickup/delivery
- "cancelled" - Order has been cancelled

### Example Usage

```bash
# Update an order status to "completed"
curl -X PUT http://localhost:8084/api/orders/your-order-id/status \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

## MQTT Message Format

The Order Service publishes messages to the topic `orders/{order_id}/status` with the following format:

```json
{
  "order_id": "order-123",
  "status": "completed",
  "message": "Order #order-123 status updated to completed",
  "timestamp": 1617293965
}
```

## Notification Flow

1. User submits an order via the Order Service
2. Order Service creates the order with "pending" status
3. Order Service publishes a status update to MQTT
4. Notification Service receives the MQTT message
5. Notification Service forwards the update to all connected WebSocket clients
6. After about 10 seconds, the automatic processor changes the status to "processing"
7. After 30 seconds in "processing" status, the order is marked as "completed"
8. Each status change is published to MQTT and delivered to clients in real-time

## Testing

You can test the automatic notification flow by using the provided test script:

```bash
# Run the test script to create an order and watch it change status
go run cmd/test/create_test_order.go
```

This script:
1. Creates a new test order
2. Polls the order status every 5 seconds
3. Displays status changes as they occur
4. The order should move from "pending" to "processing" to "completed" automatically

## Error Handling

The Order Service is designed to gracefully handle MQTT connection failures:

- If the MQTT broker is unavailable at startup, the service will log a warning but continue to function without notifications
- If publishing a notification fails, the error is logged but the order operation still succeeds 