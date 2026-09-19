# Growww / NBSE WebSocket Streaming Protocol Specification

## 1. Overview & Architecture

The Growww / NBSE WebSocket Gateway (`services/market-data-service` and `services/depth-broadcaster`) provides real-time streaming market data and private trade execution events over secure WebSockets (`wss://`).

### Connection Endpoints
- **Public Market Data (Mainnet):** `wss://stream.nbse.in/ws/market`
- **Public Market Data (Testnet / Demo):** `wss://stream.sandbox.nbse.in/ws/market`
- **Private User Stream (Authenticated):** `wss://stream.nbse.in/ws/user`

### Framing & Serialization
- Default wire format: JSON (UTF-8).
- High-frequency institutional format: Binary Protocol Buffers (`Content-Type: application/x-protobuf`).
- Heartbeat interval: Server sends `{"type": "ping"}` every 15 seconds; client must respond with `{"type": "pong"}` within 5 seconds.

---

## 2. Public Market Data Channels

### 2.1 BTC/USDT Ticker Channel (`ticker@BTC-USDT`)
- **Subscription Request:**
  ```json
  {
    "action": "subscribe",
    "channel": "ticker",
    "params": {
      "symbols": ["BTC-USDT"]
    }
  }
  ```
- **Real-Time Stream Payload:**
  ```json
  {
    "stream": "ticker@BTC-USDT",
    "data": {
      "symbol": "BTC-USDT",
      "price": "64520.50",
      "best_bid": "64520.00",
      "best_ask": "64520.50",
      "high_24h": "65100.00",
      "low_24h": "63850.00",
      "volume_24h": "1428.52000000",
      "fair_index_price": "64519.80",
      "price_change_pct_24h": "+1.05",
      "timestamp_ms": 1789726800125
    }
  }
  ```

### 2.2 Level-2 Order Book Depth Channel (`depth20@BTC-USDT`)
Emits conflated 50ms snapshots of the top 20 bid and ask price levels with an incremental sequence number and CRC32 integrity checksum.
- **Subscription Request:**
  ```json
  {
    "action": "subscribe",
    "channel": "depth",
    "params": {
      "symbols": ["BTC-USDT"],
      "levels": 20,
      "interval_ms": 50
    }
  }
  ```
- **Stream Payload:**
  ```json
  {
    "stream": "depth20@BTC-USDT",
    "data": {
      "symbol": "BTC-USDT",
      "sequence": 9940192,
      "checksum": 38401928,
      "bids": [
        ["64520.00", "2.45000000"],
        ["64519.50", "5.12000000"],
        ["64518.00", "1.89000000"]
      ],
      "asks": [
        ["64520.50", "1.15000000"],
        ["64521.00", "3.80000000"],
        ["64522.50", "4.25000000"]
      ],
      "timestamp_ms": 1789726800200
    }
  }
  ```

### 2.3 Live Candlestick Stream (`kline_1m@BTC-USDT`)
- **Subscription Request:**
  ```json
  {
    "action": "subscribe",
    "channel": "kline",
    "params": {
      "symbol": "BTC-USDT",
      "interval": "1m"
    }
  }
  ```
- **Stream Payload:**
  ```json
  {
    "stream": "kline_1m@BTC-USDT",
    "data": {
      "symbol": "BTC-USDT",
      "open_time_ms": 1789726800000,
      "close_time_ms": 1789726859999,
      "open": "64500.00",
      "high": "64535.00",
      "low": "64495.00",
      "close": "64520.50",
      "volume": "12.85000000",
      "trades_count": 482,
      "is_closed": false
    }
  }
  ```

---

## 3. Private User Execution Stream (`/ws/user`)

### 3.1 Handshake & In-Band Authentication
Clients initiate connection with a temporary authorization token, and refresh tokens in-band without dropping socket connections:
```json
{
  "action": "authenticate",
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
**Server Acknowledgement:**
```json
{
  "event": "authenticated",
  "account_id": "acc_ind_992019482",
  "environment": "MAINNET",
  "expires_in_seconds": 3600
}
```

### 3.2 Real-Time Order Execution Report
Broadcast the instant an order is accepted, partially filled, filled, or cancelled:
```json
{
  "event": "execution_report",
  "order_id": "ord_99401928401",
  "client_order_id": "my_client_ref_001",
  "symbol": "BTC-USDT",
  "side": "BUY",
  "order_type": "LIMIT",
  "exec_type": "TRADE",
  "order_status": "FILLED",
  "last_price": "64520.00",
  "last_quantity": "0.15000000",
  "cumulative_quantity": "0.15000000",
  "leaves_quantity": "0.00000000",
  "fee_amount": "0.967800",
  "fee_asset": "USDT",
  "fee_rate": "0.0000",
  "trade_id": "trd_88301928",
  "dvp_settlement_tx": "0x892a0b3c892019b83049281a0b3c892019b83049281a0b3c892019b83049281a",
  "timestamp_ms": 1789726800312
}
```

### 3.3 Balance & Margin Update Stream
```json
{
  "event": "balance_update",
  "asset": "USDT",
  "delta": "-9678.967800",
  "new_free": "2821.532200",
  "new_locked": "0.000000",
  "reason": "TRADE_BUY_SETTLEMENT",
  "timestamp_ms": 1789726800315
}
```

---

## 4. Backpressure & Conflation Controls

To protect constrained mobile clients and slow connections from latency drift or out-of-memory crashes:
1. **Per-Client Ring Buffer:** Each WebSocket connection has a bounded 256-packet ring buffer.
2. **Conflated Fallback Mode:** If client consumption falls behind by $> 256$ frames, the gateway transitions the client into `CONFLATED_SNAPSHOT` mode, emitting 200ms consolidated depth frames until backpressure subsides.
3. **Slow Consumer Eviction:** If the buffer exceeds 512 frames, the gateway terminates the connection with WebSocket close code `1008 (Policy Violation: Slow Consumer Buffer Overflow)`.
