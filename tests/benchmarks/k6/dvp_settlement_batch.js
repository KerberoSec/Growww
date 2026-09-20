import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    settlement_batch: {
      executor: 'constant-arrival-rate',
      rate: 100, // 100 batch submissions per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    'http_req_duration{type:dvp_batch}': ['p(95)<200', 'p(99)<500'],
    'http_req_failed': ['rate<0.005'],
  },
};

export default function () {
  const batchTrades = [];
  for (let i = 0; i < 50; i++) {
    batchTrades.push({
      trade_id: `trade-${__VU}-${__ITER}-${i}-${Date.now()}`,
      buyer_account: `0x71C${Math.floor(Math.random() * 10000000).toString(16).padStart(37, '0')}`,
      seller_account: `0x98A${Math.floor(Math.random() * 10000000).toString(16).padStart(37, '0')}`,
      security_token: '0x1234567890123456789012345678901234567890',
      quantity: '1000000000000000000',
      gross_amount_inr_paisa: 285000,
    });
  }

  const payload = JSON.stringify({
    batch_id: `batch-${__VU}-${__ITER}-${Date.now()}`,
    trades_count: batchTrades.length,
    trades: batchTrades,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.RELAYER_TOKEN || 'mock-settlement-relayer-token'}`,
    },
    tags: { type: 'dvp_batch' },
  };

  const res = http.post(`${__ENV.SETTLEMENT_SERVICE_URL || 'http://localhost:9002'}/api/v1/settlement/batch`, payload, params);
  check(res, {
    'batch submitted': (r) => r.status === 202 || r.status === 200,
  });
}
