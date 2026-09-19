# 204 - Order Management & Lifecycle Service (Go)

## Purpose
The Order Management & Lifecycle Service acts as the transaction orchestration hub for investor trade intents. It ingests buy and sell orders from mobile and web clients, validates market parameters (ISIN validity, tick size, price bands, minimum/fractional lot constraints), coordinates pre-trade financial reservations (cash holds for buys via Wallet Service, share holds for sells via Portfolio Service), enforces pre-trade risk policies via Risk Engine, and dispatches validated orders to the high-frequency matching engine.

The service manages the full state lifecycle of every order throughout its lifetime (from submission to final settlement or cancellation), ensuring deterministic state progression, idempotent client submissions, and audit-grade telemetry for SEBI market surveillance.

## What You Are Building
A high-throughput, low-latency Go microservice (`services/order-service`). Concrete deliverables include:
- REST and gRPC endpoints for placing orders, cancelling active orders, and querying order status/history.
- Deterministic order state machine engine managing transitions across `SUBMITTED`, `VALIDATING`, `RESERVED`, `ROUTED_TO_ENGINE`, `PARTIALLY_FILLED`, `FILLED`, `CANCELLED`, `REJECTED`, and `EXPIRED`.
- Distributed pre-trade coordinator calling Wallet Service (Prompt 203) for INR cash holds and Portfolio Service (Prompt 209) for fractional share holds.
- Kafka producer publishing validated orders to `order.matching.commands.v1` (partitioned by ISIN).
- Kafka consumer ingesting execution reports and match events from the matching engine (`engine.matches.v1`) to update persistent order state and trigger downstream notifications.

## Scope Boundaries
- **In Scope:**
 - Order intake and validation (Limit, Market, Immediate-or-Cancel IOC, Fill-or-Kill FOK).
 - Fractional share quantity validation (up to 6 decimal places, e.g., 0.125000 units of RELIANCE).
 - Pre-trade reservation coordination and rollback on rejection.
 - Order state persistence and historical order book queries for users.
 - Cancellation intent dispatching and state reconciliation.
- **Out of Scope / Handled Elsewhere:**
 - In-memory order book matching algorithm (Prompt 205).
 - Pre-trade risk policy evaluation rules and price bands (Prompt 206).
 - Double-entry cash balance ledger storage (Prompt 203).
 - DvP on-chain smart contract settlement orchestration (Prompt 208).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Selected for its exceptional concurrency handling, sub-millisecond execution overhead, predictable garbage collection, and robust standard library.
- **Database & Storage:** PostgreSQL 16+ using `pgx/v5` with connection pooling for transactional order records; Redis 7.2 for active order caching and idempotency locks.
- **SQL Generation:** `sqlc` for compile-time verified, high-performance database access.
- **Messaging & Streaming:** `segmentio/kafka-go` with manual commit strategies and producer delivery guarantees (`acks=all`).
- **RPC & Serialization:** `google.golang.org/grpc` for high-throughput inter-service communication.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `orders`, `order_state_transitions`, `order_executions`.
- **Redis 7.2:** Caches active order IDs by user and handles fast idempotency deduplication keys.
- **Apache Kafka:** Publishes to `order.matching.commands.v1` (keyed by ISIN); consumes from `engine.matches.v1`.
- **Wallet Service (Prompt 203):** Invoked via gRPC for cash holds on buy orders.
- **Risk Engine (Prompt 206):** Invoked via gRPC for pre-trade limit checks.
- **Portfolio Service (Prompt 209):** Invoked via gRPC for fractional share holds on sell orders.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Token Contract Association:** Each tradable equity symbol is mapped to its canonical on-chain `DigitalSecurityToken.sol` contract deployed on Hyperledger Besu representing 1:1 custody-backed physical shares.
- **Settlement Binding:** When an order is placed, the order record associates the user's registered public ledger address (`0x...`). When trades are matched, execution receipts link the order ID to the on-chain atomic `SettlementDvP.sol` transaction hash once executed on Hyperledger Besu.
- **Consensus Alignment:** State progressions reflect finalized on-chain settlement blocks under QBFT consensus (2-second finality).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/order-service` with strict linter rules, Makefile, and standard DDD folder layout.
2. **Define Protobuf Contracts:** Create `proto/growww/order/v1/order_service.proto` for order submission, cancellation, and status lookup.
3. **Generate Protobuf Stubs:** Compile `.proto` files into Go gRPC server and client stubs.
4. **Design PostgreSQL Schema:** Write SQL migrations for `orders`, `order_state_transitions`, and `order_executions` with comprehensive indexes.
5. **Configure `sqlc`:** Set up SQL queries for order insertion, state updating, and history retrieval with type-safe Go code generation.
6. **Implement Order Validation Module:** Build strict validation for order type, ISIN validity, tick size (0.05 INR increments), fractional share lot sizes, and positive non-zero quantities.
7. **Build Pre-Trade Reservation Coordinator:** Implement async workflow: for buy orders, call `WalletService.ReserveFunds`; for sell orders, call `PortfolioService.ReserveHoldings`; on failure, transition order to `REJECTED` and abort.
8. **Integrate Pre-Trade Risk Checks:** Call `RiskService.EvaluateOrderRisk` to verify price bands and user limits before dispatching to the matching engine.
9. **Implement Order State Machine:** Write deterministic transition logic enforcing valid state transitions and persisting every state change to `order_state_transitions`.
10. **Implement Kafka Command Publisher:** Publish validated orders to `order.matching.commands.v1` with Kafka message key set to `isin` to guarantee partition ordering.
11. **Implement Execution Report Consumer:** Build Kafka consumer for `engine.matches.v1` that updates order filled quantities, transitions orders to `PARTIALLY_FILLED` or `FILLED`, and triggers settlement workflows.
12. **Implement Cancellation Pipeline:** Build `CancelOrder` endpoint that routes cancellation requests to the matching engine and releases remaining cash/share holds upon engine confirmation.
13. **Expose REST & gRPC Handlers:** Implement REST endpoints (via grpc-gateway or Chi router) and gRPC services with structured error codes.
14. **Configure Telemetry & Metrics:** Expose Prometheus metrics tracking order submission rate, state transition latencies, and rejection rate counters.
15. **Write Comprehensive Test Suite:** Implement unit tests and integration tests using `testcontainers-go` covering edge cases (partial fills, rapid cancellations, timeout rollbacks).

## Interfaces / Contracts

### Protobuf Definition (`order_service.proto`)
```protobuf
syntax = "proto3";

package growww.order.v1;

option go_package = "growww/order/v1;orderv1";

service OrderService {
  rpc PlaceOrder (PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc CancelOrder (CancelOrderRequest) returns (CancelOrderResponse);
  rpc GetOrder (GetOrderRequest) returns (GetOrderResponse);
  rpc ListOrders (ListOrdersRequest) returns (ListOrdersResponse);
}

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BUY = 1;
  ORDER_SIDE_SELL = 2;
}

enum OrderType {
  ORDER_TYPE_UNSPECIFIED = 0;
  ORDER_TYPE_LIMIT = 1;
  ORDER_TYPE_MARKET = 2;
  ORDER_TYPE_IOC = 3; // Immediate-or-Cancel
  ORDER_TYPE_FOK = 4; // Fill-or-Kill
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_SUBMITTED = 1;
  ORDER_STATUS_ACCEPTED = 2;
  ORDER_STATUS_PARTIALLY_FILLED = 3;
  ORDER_STATUS_FILLED = 4;
  ORDER_STATUS_CANCELLED = 5;
  ORDER_STATUS_REJECTED = 6;
  ORDER_STATUS_EXPIRED = 7;
}

message PlaceOrderRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string isin = 3; // e.g., "INE002A01018"
  OrderSide side = 4;
  OrderType type = 5;
  string price = 6; // Limit price in INR (e.g., "2450.50"), optional for MARKET
  string quantity = 7; // Fractional quantity (e.g., "0.500000")
}

message PlaceOrderResponse {
  string order_id = 1;
  OrderStatus status = 2;
  string isin = 3;
  string price = 4;
  string quantity = 5;
  int64 submitted_at_unix = 6;
}

message CancelOrderRequest {
  string order_id = 1;
  string user_id = 2;
  string reason = 3;
}

message CancelOrderResponse {
  string order_id = 1;
  bool cancellation_requested = 2;
  OrderStatus current_status = 3;
}

message GetOrderRequest {
  string order_id = 1;
  string user_id = 2;
}

message GetOrderResponse {
  string order_id = 1;
  string user_id = 2;
  string isin = 3;
  OrderSide side = 4;
  OrderType type = 5;
  OrderStatus status = 6;
  string price = 7;
  string quantity = 8;
  string filled_quantity = 9;
  string average_fill_price = 10;
  int64 created_at_unix = 11;
  int64 updated_at_unix = 12;
}

message ListOrdersRequest {
  string user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  repeated OrderStatus status_filter = 4;
}

message ListOrdersResponse {
  repeated GetOrderResponse orders = 1;
  string next_page_token = 2;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE order_side_enum AS ENUM ('BUY', 'SELL');
CREATE TYPE order_type_enum AS ENUM ('LIMIT', 'MARKET', 'IOC', 'FOK');
CREATE TYPE order_status_enum AS ENUM ('SUBMITTED', 'ACCEPTED', 'PARTIALLY_FILLED', 'FILLED', 'CANCELLED', 'REJECTED', 'EXPIRED');

CREATE TABLE orders (
    order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    isin VARCHAR(12) NOT NULL,
    side order_side_enum NOT NULL,
    order_type order_type_enum NOT NULL,
    status order_status_enum NOT NULL DEFAULT 'SUBMITTED',
    price NUMERIC(18, 4), -- NULL for MARKET orders
    quantity NUMERIC(18, 6) NOT NULL CHECK (quantity > 0),
    filled_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    remaining_quantity NUMERIC(18, 6) NOT NULL,
    hold_id UUID, -- Reference to Wallet hold (Buy) or Portfolio hold (Sell)
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_state_transitions (
    transition_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    from_status order_status_enum NOT NULL,
    to_status order_status_enum NOT NULL,
    reason TEXT,
    transitioned_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    trade_id UUID NOT NULL,
    matched_price NUMERIC(18, 4) NOT NULL,
    matched_quantity NUMERIC(18, 6) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_orders_isin_status ON orders(isin, status);
```

## Security & Compliance Notes
- **Strict Idempotency:** Duplicate submissions with identical `idempotency_key` return existing order state without double-reserving funds or duplicating orders in the matching engine.
- **Pre-Trade Hold Safety:** Orders are never routed to the matching engine without verified, active funds or holdings holds.
- **Audit Logging:** Every state transition is recorded in `order_state_transitions` with millisecond timestamps to satisfy SEBI trade surveillance audit requirements.

## Acceptance Criteria
- [ ] Order service accepts valid buy and sell orders with fractional quantities up to 6 decimal places.
- [ ] Buy orders atomically reserve cash in Wallet Service; sell orders atomically reserve fractional shares in Portfolio Service.
- [ ] Orders failing risk checks or funds reservation are immediately transitioned to `REJECTED` with clear error reasons.
- [ ] Kafka events on `order.matching.commands.v1` are keyed by `isin` ensuring strict sequential delivery per symbol.
- [ ] Matches received from matching engine correctly update filled quantities and transition states to `PARTIALLY_FILLED` or `FILLED`.
- [ ] Cancellation requests release unused holds back to the user's available balance within 100ms.
- [ ] End-to-end order intake latency (intake $\rightarrow$ pre-trade checks $\rightarrow$ Kafka dispatch) is $< 5\text{ms}$ under load.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 104 (Kafka Standards), 111 (Domain Model), 112 (Idempotency), 203 (Wallet Service), 206 (Risk Engine).
- **Parallel Tasks:** 205 (Matching Engine), 207 (Market Data Service).
- **Downstream Blockers:** 208 (Trade Settlement Service), 509 (Flutter Order Flow), 603 (Web Trading Dashboard).
