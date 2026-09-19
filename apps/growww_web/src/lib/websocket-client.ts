type MessageHandler = (data: any) => void;

export class WebSocketClient {
  private url: string;
  private ws: WebSocket | null = null;
  private listeners: Map<string, Set<MessageHandler>> = new Map();
  private reconnectIntervalMs: number = 2000;
  private isConnected: boolean = false;
  private heartbeatTimer: any = null;

  constructor(url: string = 'wss://api.growww.in/ws/v1/market') {
    this.url = url;
  }

  connect(): void {
    if (typeof window === 'undefined') return;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.isConnected = true;
        this.startHeartbeat();
        this.emit('connection', { status: 'CONNECTED' });
      };

      this.ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          const channel = parsed.channel || parsed.topic || 'default';
          this.emit(channel, parsed.data || parsed);
        } catch {
          this.emit('raw', event.data);
        }
      };

      this.ws.onclose = () => {
        this.isConnected = false;
        this.stopHeartbeat();
        this.emit('connection', { status: 'DISCONNECTED' });
        setTimeout(() => this.connect(), this.reconnectIntervalMs);
      };

      this.ws.onerror = (err) => {
        this.emit('error', err);
      };
    } catch {
      // Fallback for mocked/offline testing
      this.isConnected = true;
      this.emit('connection', { status: 'CONNECTED_MOCK' });
    }
  }

  subscribe(channel: string, handler: MessageHandler): () => void {
    if (!this.listeners.has(channel)) {
      this.listeners.set(channel, new Set());
    }
    this.listeners.get(channel)!.add(handler);

    if (this.isConnected && this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: 'subscribe', channel }));
    }

    return () => {
      this.listeners.get(channel)?.delete(handler);
    };
  }

  private emit(channel: string, data: any): void {
    this.listeners.get(channel)?.forEach((fn) => fn(data));
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ action: 'ping', timestamp: Date.now() }));
      }
    }, 5000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) clearInterval(this.heartbeatTimer);
  }

  disconnect(): void {
    this.stopHeartbeat();
    this.ws?.close();
    this.isConnected = false;
  }
}

export const wsClient = new WebSocketClient();
