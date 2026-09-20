import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    order_burst: {
      executor: 'ramping-arrival-rate',
      startRate: 5000,
      timeUnit: '1s',
      preAllocatedVUs: 2000,
      maxVUs: 10000,
      stages: [
        { duration: '2m', target: 10000 },  // Warm-up to 10k req/s
        { duration: '5m', target: 50000 },  // Ramp to 50k req/s
        { duration: '15m', target: 50000 }, // Sustain 50k req/s peak load
        { duration: '3m', target: 0 },      // Cool-down
      ],
    },
  },
  thresholds: {
    'http_req_duration{type:order_place}': ['p(95)<30', 'p(99)<50'], // API Gateway p99 < 50ms
    'http_req_failed': ['rate<0.001'],                               // Error rate < 0.1%
    'matching_engine_p99_latency': ['value<1000'],                   // In-engine p99 < 1ms
  },
};

const SYMBOLS = ['INE002A01018', 'INE467B01029', 'INE040A01034', 'INE009A01021', 'INE090A01021'];

export default function () {
  const symbol = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
  const isBuy = Math.random() > 0.5;
  const side = isBuy ? 'BUY' : 'SELL';
  const price = (2800.0 + (Math.random() * 40 - 20)).toFixed(2);
  const qty = (Math.random() * 10 + 1).toFixed(2);

  const payload = JSON.stringify({
    isin: symbol,
    order_type: 'LIMIT',
    side: side,
    quantity: qty,
    price: price,
    idempotency_key: `load-${__VU}-${__ITER}-${Date.now()}`
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.TEST_AUTH_TOKEN || 'mock-institutional-token'}`,
      'x-user-tier': 'tier2_institutional',
      'x-request-timestamp': Math.floor(Date.now() / 1000).toString(),
    },
    tags: { type: 'order_place' },
  };

  const res = http.post(`${__ENV.API_BASE_URL || 'http://localhost:8443'}/api/v1/orders`, payload, params);
  check(res, {
    'order accepted': (r) => r.status === 201 || r.status === 200,
  });
}
