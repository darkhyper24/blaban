import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

export const errorRate = new Rate('http_error_rate');
export const durationTrend = new Trend('request_duration');

const baseUrl = 'http://localhost:8084/api';
const authUrl = 'http://localhost:8082/api';

const testUserId = '66687428-eabd-4544-8b33-c701cba2914b';

let orderIds = [];

export const options = {
  scenarios: {
    stress_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 100,
      stages: [
        { duration: '15s', target: 10 },
        { duration: '30s', target: 30 },
        { duration: '30s', target: 50 },
        { duration: '15s', target: 0 },
      ],
    },
  },

  thresholds: {
    'http_req_duration': ['p(95)<1000'],
    'http_error_rate': ['rate<0.10'], 
    'http_req_failed': ['rate<0.10'], 
    'request_duration': ['avg<1000'],      
  },
};

const createOrderData = {
  items: [
    { 
      item_id: "d704184b-7858-4450-a163-b17335b2c4f7",
      quantity: 2 
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

export default function () {
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
  
  const operation = Math.random() < 0.7 ? 'get' : 'create';
  
  if (operation === 'create') {
    const postStart = Date.now();
    const createRes = http.post(`${baseUrl}/orders`, JSON.stringify(createOrderData), { headers });
    const duration = Date.now() - postStart;
    durationTrend.add(duration);

    const createOk = check(createRes, {
      'order created successfully': (r) => r.status === 201 || r.status === 200,
      'order ID received': (r) => !!r.json('id') || !!r.json('orderId')
    });

    errorRate.add(!createOk);

    let createdOrderId = createRes.json('id');
    if (!createdOrderId) {
      createdOrderId = createRes.json('orderId');
    }
    
    if (createdOrderId) {
      orderIds.push(createdOrderId);
    }
  } else {
    const getSpecific = orderIds.length > 0 && Math.random() < 0.5;
    
    const getStart = Date.now();
    let res;
    
    if (getSpecific) {
      const randomId = orderIds[Math.floor(Math.random() * orderIds.length)];
      res = http.get(`${baseUrl}/orders/${randomId}`, { headers });
    } else {
      res = http.get(`${baseUrl}/orders`, { headers });
    }
    
    const duration = Date.now() - getStart;
    durationTrend.add(duration);

    const getOk = check(res, {
      'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    });

    errorRate.add(!getOk);
  }

  sleep(Math.random() * 0.5 + 0.3);
} 