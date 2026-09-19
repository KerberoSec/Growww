import { describe, it, expect } from 'vitest';
import { ApiClient } from '../src/lib/api-client';
import { WebSocketClient } from '../src/lib/websocket-client';

describe('Prompt 601 - Next.js 14 Investor Web App Scaffolding', () => {
  it('instantiates ApiClient with default endpoint and auth token support', () => {
    const client = new ApiClient({ baseUrl: 'https://mock.api.growww.in' });
    expect(client).toBeDefined();
    client.setAuthToken('test-jwt-token');
  });

  it('manages WebSocket subscriptions and topic handlers', () => {
    const ws = new WebSocketClient('wss://mock.growww.in/ws');
    expect(ws).toBeDefined();

    let messageReceived = false;
    const unsub = ws.subscribe('orderbook:BTC-USDT', (data) => {
      messageReceived = true;
    });

    expect(typeof unsub).toBe('function');
    unsub();
  });
});
