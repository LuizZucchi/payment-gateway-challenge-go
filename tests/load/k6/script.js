import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '30s', target: 10 }, 
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const url = 'http://localhost:8090/api/payments';
  
  const payload = JSON.stringify({
    card_number: "1234567890123451",
    expiry_month: 12,
    expiry_year: 2030,
    currency: "USD",
    amount: 1500,
    cvv: "123"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'payment is authorized': (r) => r.json('payment_status') === 'Authorized',
    'has payment id': (r) => r.json('id') !== '',
  });

  sleep(1);
}