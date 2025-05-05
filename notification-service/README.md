# Notification Service

The Notification Service provides real-time order status updates to clients using WebSockets and MQTT.

## Overview

This service:
- Creates a persistent WebSocket connection with clients
- Subscribes to order status updates via MQTT
- Broadcasts order status notifications to connected clients

## Architecture

- **WebSocket Server**: Handles client connections and delivers real-time notifications
- **MQTT Client**: Subscribes to order status topics and processes updates
- **Hub**: Manages WebSocket connections and broadcasts notifications

## Setup

1. Ensure you have an MQTT broker running (e.g., Mosquitto) on port 1883
2. Start the notification service:
   ```
   cd notification-service
   go run cmd/main.go
   ```

## Client Integration

### WebSocket Connection

Connect to the WebSocket endpoint:
```javascript
// Front-end JavaScript example
const ws = new WebSocket('ws://localhost:8085/ws');

ws.onopen = () => {
  console.log('Connected to notification service');
};

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('Received notification:', notification);
  
  // Handle order status update
  if (notification.type === 'order_status_update') {
    // Update UI based on order status
    if (notification.status === 'completed') {
      // Show order completion notification
      displayNotification(`Your order #${notification.order_id} is ready!`);
    } else if (notification.status === 'processing') {
      // Show processing notification
      displayNotification(`Your order #${notification.order_id} is being prepared.`);
    }
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from notification service');
  // Implement reconnection logic here
};

function displayNotification(message) {
  // Implement your UI notification logic here
  // For example:
  alert(message);
}
```

## Testing

You can test the notification service using the provided test script:

```bash
# Run the test script with an order ID
go run cmd/test/send_notification.go ORDER_ID [STATUS]

# Example: 
go run cmd/test/send_notification.go order-123 completed
```

## Notification Message Format

```json
{
  "type": "order_status_update",
  "order_id": "order-123",
  "status": "completed",
  "message": "Your order has been completed",
  "data": {
    "order_id": "order-123",
    "status": "completed",
    "timestamp": 1617293965
  },
  "timestamp": 1617293965
}
```

## Integration with Order Service

When an order status changes in the order service, it should publish a message to the MQTT broker on the topic `orders/{order_id}/status` with the following payload:

```json
{
  "order_id": "order-123",
  "status": "completed",
  "message": "Order has been prepared and is ready for pickup"
}
```

The notification service will then:
1. Receive the MQTT message
2. Format it as a notification
3. Broadcast it to all connected WebSocket clients 