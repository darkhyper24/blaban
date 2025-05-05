import React, { useEffect, useState } from 'react';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const OrderNotifications = () => {
  const [connected, setConnected] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [socket, setSocket] = useState(null);

  // Connect to the WebSocket server when the component mounts
  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket('ws://localhost:8085/ws');
    
    ws.onopen = () => {
      console.log('Connected to notification service');
      setConnected(true);
      toast.success('Connected to notification service');
    };
    
    ws.onmessage = (event) => {
      try {
        const notification = JSON.parse(event.data);
        console.log('Received notification:', notification);
        
        // Add the notification to our state
        setNotifications(prev => [notification, ...prev].slice(0, 10));
        
        // Show a toast notification
        if (notification.type === 'order_status_update') {
          switch (notification.status) {
            case 'completed':
              toast.success(`Your order #${notification.order_id} is ready!`);
              break;
            case 'processing':
              toast.info(`Your order #${notification.order_id} is being prepared.`);
              break;
            case 'cancelled':
              toast.error(`Your order #${notification.order_id} has been cancelled.`);
              break;
            default:
              toast.info(`Order #${notification.order_id}: ${notification.message}`);
          }
        }
      } catch (error) {
        console.error('Error parsing notification:', error);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setConnected(false);
      toast.error('Connection to notification service failed');
    };
    
    ws.onclose = () => {
      console.log('Disconnected from notification service');
      setConnected(false);
      toast.warn('Disconnected from notification service');
      
      // Try to reconnect after 5 seconds
      setTimeout(() => {
        toast.info('Attempting to reconnect...');
      }, 5000);
    };
    
    setSocket(ws);
    
    // Clean up the WebSocket connection when the component unmounts
    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, []);
  
  return (
    <div className="order-notifications">
      <h2>Order Notifications</h2>
      <div className="connection-status">
        Status: {connected ? 
          <span className="connected">Connected</span> : 
          <span className="disconnected">Disconnected</span>
        }
      </div>
      
      <div className="notifications-list">
        <h3>Recent Updates</h3>
        {notifications.length === 0 ? (
          <p>No notifications yet</p>
        ) : (
          <ul>
            {notifications.map((notification, index) => (
              <li key={index} className={`notification ${notification.status}`}>
                <div className="notification-time">
                  {new Date(notification.timestamp * 1000).toLocaleTimeString()}
                </div>
                <div className="notification-content">
                  <strong>Order #{notification.order_id}</strong>: {notification.message}
                </div>
                <div className="notification-status">
                  Status: <span className={notification.status}>{notification.status}</span>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
      
      <ToastContainer 
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
      />
      
      <style jsx>{`
        .order-notifications {
          background: #f8f9fa;
          border-radius: 8px;
          padding: 20px;
          margin: 20px 0;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .connection-status {
          margin-bottom: 15px;
        }
        
        .connected {
          color: green;
          font-weight: bold;
        }
        
        .disconnected {
          color: red;
          font-weight: bold;
        }
        
        .notifications-list {
          max-height: 300px;
          overflow-y: auto;
          border: 1px solid #ddd;
          border-radius: 4px;
          padding: 10px;
        }
        
        .notification {
          padding: 10px;
          margin-bottom: 10px;
          border-radius: 4px;
          background: white;
          border-left: 4px solid #ccc;
        }
        
        .notification.completed {
          border-left-color: green;
        }
        
        .notification.processing {
          border-left-color: blue;
        }
        
        .notification.cancelled {
          border-left-color: red;
        }
        
        .notification.pending {
          border-left-color: orange;
        }
        
        .notification-time {
          font-size: 0.8em;
          color: #666;
        }
        
        .notification-content {
          margin: 5px 0;
        }
        
        .notification-status .completed {
          color: green;
        }
        
        .notification-status .processing {
          color: blue;
        }
        
        .notification-status .cancelled {
          color: red;
        }
        
        .notification-status .pending {
          color: orange;
        }
      `}</style>
    </div>
  );
};

export default OrderNotifications; 