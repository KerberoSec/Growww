# Growww / NBSE REST API Specification (OpenAPI 3.1)

## 1. Overview & Architectural Standards

The Growww / NBSE REST API provides institutional and retail developers with programmatic access to account management, trading, market data, and ledger services. The API conforms to OpenAPI 3.1 standards, strictly enforcing:

- **Base URLs:**
  - **Production (Mainnet):** `https://api.nbse.in/v1`
  - **Sandbox (Testnet / Demo):** `https://sandbox.nbse.in/v1`
- **Wire Format:** Standard JSON UTF-8 with strict ASCII encoding.
- **Authentication:** Bearer token (JWT) passed in the `Authorization` header (`Authorization: Bearer <jwt_token>`).
- **Idempotency:** All state-mutating requests (`POST`, `PUT`, `DELETE`) require a unique UUIDv4 header: `Idempotency-Key: <uuid>`.
- **Fee Transparency:** Every trade execution confirms the platform flat **0.00% fee (No fee at all)** (1 basis point / 0 bps at launch).
- **Rate Limits:**
  - Public endpoints: 100 requests per second per IP.
  - Private trading endpoints: 50 requests per second per user account.
  - Rate limit headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.

---

## 2. Authentication & Identity Endpoints

### 2.1 Authenticate / Obtain JWT Token
- **Method & Path:** `POST /auth/token`
- **Description:** Authenticates user via OAuth2/OIDC credentials or API Key/Secret.
- **Request Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "grant_type": "api_key",
    "api_key": "nbse_live_89f0291a0b3c",
    "api_secret": "sec_892019b83049281a0b3c8920"
  }
  ```
- **Response (`200 OK`):**
  ```json
  {
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "environment": "MAINNET",
    "account_id": "acc_ind_992019482"
  }
  ```

---

## 3. Account, Balances & Demo Faucet

### 3.1 Get Account Balances
- **Method & Path:** `GET /accounts/balances`
- **Description:** Returns real-time unencumbered, locked, and total collateral balances.
- **Headers:** `Authorization: Bearer <jwt>`
- **Response (`200 OK`):**
  ```json
  {
    "account_id": "acc_ind_992019482",
    "environment": "MAINNET",
    "balances": [
      {
        "asset": "USDT",
        "free": "12500.500000",
        "locked": "2500.000000",
        "total": "15000.500000"
      },
      {
        "asset": "BTC",
        "free": "0.45000000",
        "locked": "0.10000000",
        "total": "0.55000000"
      },
      {
        "asset": "INR",
        "free": "100000.00",
        "locked": "0.00",
        "total": "100000.00"
      }
    ]
  }
  ```

### 3.2 Claim Demo Testnet Faucet (Paper Trading)
- **Method & Path:** `POST /demo/faucet/claim`
- **Description:** Dispenses 10,000 virtual USDT and 1.0 virtual BTC to a testnet paper trading account.
- **Headers:** `Authorization: Bearer <jwt>`, `Idempotency-Key: <uuid>`
- **Response (`200 OK`):**
  ```json
  {
    "status": "SUCCESS",
    "environment": "TESTNET",
    "credited": {
      "vUSDT": "10000.000000",
      "vBTC": "1.00000000"
    },
    "new_balances": {
      "vUSDT": "10000.000000",
      "vBTC": "1.00000000"
    },
    "rate_limit_reset_utc": "2026-09-20T10:00:00Z"
  }
  ```

### 3.3 Reset Demo Portfolio
- **Method & Path:** `POST /demo/portfolio/reset`
- **Description:** Closes all virtual open positions and resets virtual balances back to original faucet state.
- **Headers:** `Authorization: Bearer <jwt>`, `Idempotency-Key: <uuid>`
- **Response (`200 OK`):**
  ```json
  {
    "status": "RESET_COMPLETE",
    "environment": "TESTNET",
    "closed_orders_count": 3,
    "current_vUSDT": "10000.000000",
    "current_vBTC": "1.00000000"
  }
  ```

---

## 4. Spot & BTC/USDT Trading Endpoints

### 4.1 Place Order (Limit / Market / Stop-Limit)
- **Method & Path:** `POST /orders`
- **Description:** Places a new buy or sell order for spot or BTC/USDT trading.
- **Headers:** `Authorization: Bearer <jwt>`, `Idempotency-Key: <uuid>`, `X-NBSE-Environment: MAINNET | TESTNET`
- **Request Body:**
  ```json
  {
    "symbol": "BTC/USDT",
    "side": "BUY",
    "type": "LIMIT",
    "time_in_force": "GTC",
    "price": "64500.50",
    "quantity": "0.15000000",
    "stop_price": null,
    "client_order_id": "my_client_ref_001"
  }
  ```
- **Response (`201 Created`):**
  ```json
  {
    "order_id": "ord_99401928401",
    "client_order_id": "my_client_ref_001",
    "symbol": "BTC/USDT",
    "side": "BUY",
    "type": "LIMIT",
    "price": "64500.500000",
    "quantity": "0.15000000",
    "executed_quantity": "0.00000000",
    "status": "ACCEPTED",
    "fee_rate": "0.0000",
    "estimated_fee": "0.967507",
    "created_at_utc": "2026-09-19T10:15:30.124Z"
  }
  ```

### 4.2 Cancel Order
- **Method & Path:** `DELETE /orders/{order_id}`
- **Description:** Cancels an active resting limit or stop-limit order.
- **Headers:** `Authorization: Bearer <jwt>`, `Idempotency-Key: <uuid>`
- **Response (`200 OK`):**
  ```json
  {
    "order_id": "ord_99401928401",
    "symbol": "BTC/USDT",
    "status": "CANCELLED",
    "unfilled_quantity": "0.15000000",
    "released_collateral": "9675.075000"
  }
  ```

### 4.3 Get Open Orders
- **Method & Path:** `GET /orders/open?symbol=BTC/USDT`
- **Headers:** `Authorization: Bearer <jwt>`
- **Response (`200 OK`):**
  ```json
  {
    "orders": [
      {
        "order_id": "ord_99401928401",
        "symbol": "BTC/USDT",
        "side": "BUY",
        "type": "LIMIT",
        "price": "64500.500000",
        "quantity": "0.15000000",
        "executed_quantity": "0.00000000",
        "created_at_utc": "2026-09-19T10:15:30.124Z"
      }
    ]
  }
  ```

---

## 5. Market Data Endpoints

### 5.1 Get Order Book Depth (Level-2)
- **Method & Path:** `GET /market/depth?symbol=BTC/USDT&limit=20`
- **Description:** Returns the top N bid and ask price levels with aggregated volumes and checksum.
- **Response (`200 OK`):**
  ```json
  {
    "symbol": "BTC/USDT",
    "sequence": 88392019,
    "checksum": 294019284,
    "bids": [
      ["64500.00", "1.25000000"],
      ["64499.50", "3.42000000"],
      ["64498.00", "0.85000000"]
    ],
    "asks": [
      ["64501.00", "0.95000000"],
      ["64502.50", "2.10000000"],
      ["64505.00", "4.50000000"]
    ],
    "timestamp_utc": "2026-09-19T10:16:00.050Z"
  }
  ```

### 5.2 Get Candlesticks (OHLCV)
- **Method & Path:** `GET /market/candles?symbol=BTC/USDT&interval=1m&limit=100`
- **Description:** Returns historical and current OHLCV candlestick series.
- **Intervals Supported:** `1s`, `1m`, `5m`, `15m`, `1h`, `4h`, `1d`.
- **Response (`200 OK`):**
  ```json
  {
    "symbol": "BTC/USDT",
    "interval": "1m",
    "candles": [
      [1789726500000, "64480.00", "64510.00", "64475.50", "64502.00", "15.42000000"],
      [1789726560000, "64502.00", "64525.00", "64495.00", "64518.50", "18.89000000"]
    ]
  }
  ```

---

## 6. Crypto Deposit & Withdrawal Endpoints

### 6.1 Get Crypto Deposit Address
- **Method & Path:** `GET /wallet/deposit-address?asset=BTC&network=BITCOIN`
- **Headers:** `Authorization: Bearer <jwt>`
- **Response (`200 OK`):**
  ```json
  {
    "asset": "BTC",
    "network": "BITCOIN",
    "address": "bc1q98f0291a0b3c892019b83049281a0b3c8920",
    "address_type": "NATIVE_SEGWIT",
    "minimum_deposit": "0.00050000",
    "confirmations_required": 3
  }
  ```

### 6.2 Submit Withdrawal Request
- **Method & Path:** `POST /wallet/withdraw`
- **Headers:** `Authorization: Bearer <jwt>`, `Idempotency-Key: <uuid>`
- **Request Body:**
  ```json
  {
    "asset": "USDT",
    "network": "TRON",
    "destination_address": "TQ89f0291a0b3c892019b83049281a0b3c",
    "amount": "5000.000000",
    "two_factor_code": "892019"
  }
  ```
- **Response (`202 Accepted`):**
  ```json
  {
    "withdrawal_id": "wth_88301928401",
    "asset": "USDT",
    "network": "TRON",
    "amount": "5000.000000",
    "network_fee": "1.000000",
    "status": "AWAITING_MPC_SIGNATURE",
    "created_at_utc": "2026-09-19T10:18:00.124Z"
  }
  ```

---

## 7. Error Handling Standard

All API errors return a standardized RFC 7807 problem details object:

```json
{
  "type": "https://api.nbse.in/errors/INSUFFICIENT_FEE_COLLATERAL",
  "title": "Insufficient Collateral For Platform Fee",
  "status": 400,
  "detail": "Free collateral of 12.50 USDT is insufficient to cover gross notional 15000.00 USDT plus 0.00% (Zero Fee) platform fee (0.00 USDT).",
  "instance": "/orders",
  "error_code": "ERR_COLLATERAL_004",
  "timestamp": "2026-09-19T10:18:45.102Z"
}
```
