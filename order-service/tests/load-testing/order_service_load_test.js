import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

export const errorRate = new Rate('http_error_rate');
export const durationTrend = new Trend('request_duration');

const baseUrl = 'http://localhost:8084/api';
const authUrl = 'http://localhost:8082/api';

// User ID for the test user we created
const testUserId = '66687428-eabd-4544-8b33-c701cba2914b'; // Replace with your actual test user ID

export const options = {
  scenarios: {
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 5 },
        { duration: '20s', target: 10 },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '5s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<1000'],
    'http_error_rate': ['rate<0.05'],
    'http_req_failed': ['rate<0.05'],
    'request_duration': ['avg<800'],
  },
};

const createOrderData = {
  items: [
    { 
      item_id: "d704184b-7858-4450-a163-b17335b2c4f7", // Changed to an existing menu item ID
      quantity: 6 
    }
  ],
  delivery_address: {
    street: "123 Main St",
    city: "Cairo",
    zip_code: "12345"
  },
  payment_method: "cash_on_delivery",
  notes: "Please deliver to front door",
  test: true
};

// This array is scoped per VU, not shared across all
let orderIds = [];

export default function () {
  // Get token directly from auth service instead of login
  const tokenRes = http.post(`${authUrl}/auth/tokens`, JSON.stringify({
    user_id: testUserId,
    role: "user"
  }), {
    headers: { 'Content-Type': 'application/json' }
  });

  check(tokenRes, {
    'token generation succeeded': (res) => res.status === 200,
    'received access token': (res) => !!res.json('access_token'),
  });

  const token = tokenRes.json('access_token');
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };

  // 1. Create order
  const postStart = Date.now();
  const createRes = http.post(`${baseUrl}/orders`, JSON.stringify(createOrderData), { headers });
  durationTrend.add(Date.now() - postStart);

  check(createRes, {
    'order creation status is 201': (r) => r.status === 201 || r.status === 200,
    'order ID received': (r) => !!r.json('id') || !!r.json('orderId')
  });

  // Try to get the order ID, supporting both potential response formats
  let createdOrderId = createRes.json('id');
  if (!createdOrderId) {
    createdOrderId = createRes.json('orderId');
  }
  
  if (createdOrderId) {
    orderIds.push(createdOrderId);
  }

  // 2. Get orders (both specific and all orders)
  if (Math.random() > 0.5) {
    // 50% chance: Get all orders
    const getStart = Date.now();
    const getRes = http.get(`${baseUrl}/orders`, { headers });
    durationTrend.add(Date.now() - getStart);

    const ok = check(getRes, {
      'GET all orders 200': (r) => r.status === 200,
    });

    errorRate.add(!ok);
  } else {
    // 50% chance: Get specific order (if we have one)
    const randomId = orderIds.length > 0
      ? orderIds[Math.floor(Math.random() * orderIds.length)]
      : null;

    if (randomId) {
      const getStart = Date.now();
      const getRes = http.get(`${baseUrl}/orders/${randomId}`, { headers });
      durationTrend.add(Date.now() - getStart);

      const ok = check(getRes, {
        'GET specific order 200': (r) => r.status === 200,
      });

      errorRate.add(!ok);
    }
  }

  sleep(Math.random() * 1 + 0.5); // Sleep between 0.5–1.5s
}
