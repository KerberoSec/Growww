# Growww / NBSE gRPC Service Definitions & Protocol Buffers Specification

## 1. Overview & Protocol Architecture

All internal inter-service communication across the Growww / NBSE platform is mediated by high-performance gRPC services utilizing Protocol Buffers v3 (`proto3`).

### Key Communication Guidelines
- **Transport:** HTTP/2 over mTLS with SPIFFE/SPIRE workload attestation.
- **Wire Compatibility:** Strict backward compatibility rules governed by Prompt 114 (field tags are immutable, newly added fields are optional, zero tag reuse).
- **Deadlines:** Every gRPC RPC carries an explicit context deadline (default 250ms for matching/risk RPCs; 1000ms for database/settlement RPCs).

---

## 2. Core Service Definitions

### 2.1 Order Service & Matching Engine Protobuf Schema
```protobuf
syntax = "proto3";

package nbse.order.v1;

option go_package = "github.com/growww/proto/order/v1;orderv1";

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BUY = 1;
  ORDER_SIDE_SELL = 2;
}

enum OrderType {
  ORDER_TYPE_UNSPECIFIED = 0;
  ORDER_TYPE_LIMIT = 1;
  ORDER_TYPE_MARKET = 2;
  ORDER_TYPE_STOP_LIMIT = 3;
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_ACCEPTED = 2;
  ORDER_STATUS_PARTIALLY_FILLED = 3;
  ORDER_STATUS_FILLED = 4;
  ORDER_STATUS_CANCELLED = 5;
  ORDER_STATUS_REJECTED = 6;
}

message PlaceOrderRequest {
  string idempotency_key = 1;
  string account_id = 2;
  string symbol = 3;
  OrderSide side = 4;
  OrderType type = 5;
  int64 price_scaled = 6;     // Scaled by 10^6
  int64 quantity_scaled = 7;  // Scaled by 10^8
  optional int64 stop_price_scaled = 8;
  string client_order_id = 9;
  bool is_demo = 10;
}

message PlaceOrderResponse {
  string order_id = 1;
  OrderStatus status = 2;
  int64 executed_quantity_scaled = 3;
  int64 fee_amount_scaled = 4;
  int64 fee_rate_bps = 5; // Exactly 0 bps (0.00% fee at launch) (0.00% (Zero Fee))
  int64 created_at_unix_ms = 6;
}

message CancelOrderRequest {
  string order_id = 1;
  string account_id = 2;
  string symbol = 3;
}

message CancelOrderResponse {
  string order_id = 1;
  OrderStatus status = 2;
  int64 unexecuted_quantity_scaled = 3;
}

service BtcOrderService {
  rpc PlaceOrder (PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc CancelOrder (CancelOrderRequest) returns (CancelOrderResponse);
}
```

---

### 2.2 Demo Paper Trading Matching Service
```protobuf
syntax = "proto3";

package nbse.demo.v1;

option go_package = "github.com/growww/proto/demo/v1;demov1";

import "nbse/order/v1/order.proto";

message DemoOrderRequest {
  string idempotency_key = 1;
  string account_id = 2;
  string symbol = 3;
  nbse.order.v1.OrderSide side = 4;
  nbse.order.v1.OrderType type = 5;
  int64 price_scaled = 6;
  int64 quantity_scaled = 7;
}

message DemoOrderResponse {
  string order_id = 1;
  nbse.order.v1.OrderStatus status = 2;
  int64 fill_price_scaled = 3;
  int64 simulated_fee_scaled = 4;
  int64 new_virtual_usdt_balance = 5;
  int64 new_virtual_btc_balance = 6;
}

message FaucetClaimRequest {
  string account_id = 1;
}

message FaucetClaimResponse {
  bool success = 1;
  int64 credited_vusdt = 2;
  int64 credited_vbtc = 3;
  int64 next_claim_allowed_unix_sec = 4;
}

service DemoMatchingService {
  rpc ExecuteDemoOrder (DemoOrderRequest) returns (DemoOrderResponse);
  rpc ClaimDemoFaucet (FaucetClaimRequest) returns (FaucetClaimResponse);
}
```

---

### 2.3 Double-Entry Wallet Ledger Service
```protobuf
syntax = "proto3";

package nbse.wallet.v1;

option go_package = "github.com/growww/proto/wallet/v1;walletv1";

message ReserveBalanceRequest {
  string idempotency_key = 1;
  string account_id = 2;
  string asset = 3;
  int64 amount_scaled = 4;
  string reason = 5;
}

message ReserveBalanceResponse {
  bool success = 1;
  string reservation_id = 2;
  int64 remaining_free_balance = 3;
}

message ReleaseBalanceRequest {
  string reservation_id = 1;
  string account_id = 2;
  int64 amount_scaled = 3;
}

message ReleaseBalanceResponse {
  bool success = 1;
  int64 new_free_balance = 2;
}

service WalletService {
  rpc ReserveBalance (ReserveBalanceRequest) returns (ReserveBalanceResponse);
  rpc ReleaseBalance (ReleaseBalanceRequest) returns (ReleaseBalanceResponse);
}
```
