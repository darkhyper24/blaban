import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

// 📊 Custom metrics
export const errorRate = new Rate('http_error_rate');
export const durationTrend = new Trend('request_duration');

// 🌍 Base URL
const baseUrl = 'http://localhost:8083/api';

const menuItemIds = [
  'd704184b-7858-4450-a163-b17335b2c4f7',
  'a0a25e8c-c750-4e29-8741-de4a201beae0',
  '37907130-b513-42a5-9577-e9c59ee2e0a3',
  '39b8c4d6-20a6-4b3a-a2e9-1994d489ca82',
  'afded924-87f8-49a7-b759-4bff7147dba1'
];

export const options = {
  scenarios: {
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '15s', target: 10 },
        { duration: '30s', target: 30 },
        { duration: '30s', target: 50 },
        { duration: '15s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },

  thresholds: {
    'http_req_duration': ['p(95)<1000'],            // ✅ 95% of requests < 1s
    'http_error_rate': ['rate<0.05'],               // ✅ <5% error rate
    'http_req_failed': ['rate<0.05'],               // ✅ <5% failure rate
    'request_duration': ['avg<800'],                // ✅ Trend avg duration
  },
};

export default function () {
  const endpoints = [
    `${baseUrl}/categories`,
    `${baseUrl}/menu`,
    `${baseUrl}/menu/search?q=ام`,
    `${baseUrl}/menu/filter?category_id=f54c7e1d-34b4-4b89-8703-90090edf5fb1`,
    `${baseUrl}/menu/${menuItemIds[Math.floor(Math.random() * menuItemIds.length)]}`
  ];

  for (const url of endpoints) {
    const start = Date.now();
    const res = http.get(url);
    const duration = Date.now() - start;

    durationTrend.add(duration);

    const ok = check(res, {
      'status is 2xx/3xx': (r) => r.status >= 200 && r.status < 400,
    });

    errorRate.add(!ok);
  }

  sleep(Math.random() * 1 + 0.5);  // 0.5 to 1.5 sec sleep
}
