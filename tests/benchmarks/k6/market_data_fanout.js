import ws from 'k6/ws';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    ws_fanout: {
      executor: 'ramping-vus',
      startVUs: 1000,
      stages: [
        { duration: '3m', target: 20000 },
        { duration: '5m', target: 50000 },
        { duration: '10m', target: 100000 },
        { duration: '2m', target: 0 },
      ],
    },
  },
  thresholds: {
    'ws_connecting{status:success}': ['rate>0.99'],
    'ws_msgs_received': ['count>1000000'],
  },
};

export default function () {
  const url = `${__ENV.WS_BASE_URL || 'ws://localhost:8443'}/ws/v1/market-data?symbol=RELIANCE`;
  const params = { tags: { my_tag: 'market_data_ws' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({ action: 'subscribe', channel: 'orderbook_l2', symbol: 'RELIANCE' }));
    });

    socket.on('message', (data) => {
      check(data, {
        'received message': (d) => d.length > 0,
      });
    });

    socket.on('error', (e) => {
      console.error('WebSocket error: ', e.error());
    });

    socket.setTimeout(() => {
      socket.close();
    }, 60000);
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
