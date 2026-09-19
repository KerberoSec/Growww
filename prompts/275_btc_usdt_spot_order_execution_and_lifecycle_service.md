# 275 - Real-Money BTC/USDT Spot Order Execution & Lifecycle Service (Go)

## Purpose
The **Real-Money BTC/USDT Spot Order Execution & Lifecycle Service** (`services/btc-order-service`) serves as the production-grade Order Management System (OMS) and transaction lifecycle state orchestrator for real-money Bitcoin/Tether (BTC/USDT) spot trading on the National Blockchain Stock Exchange (NBSE) platform.

In an institutional-grade digital asset exchange, retail and institutional participants require sub-millisecond order intake, deterministic pre-trade financial reservations, bulletproof anti-double-spend guarantees, and absolute fidelity to statutory fee invariants. This service sits at the nexus between investor trade intent, high-speed order book matching, pre-trade balance validation, real-time risk checks, and post-trade Delivery-versus-Payment (DvP) settlement on the permissioned Hyperledger Besu mainnet.

The service manages the full state lifecycle of every BTC/USDT order from inception to final settlement or cancellation. It guarantees:
- Atomic pre-trade balance reservation holds across fiat/stablecoin and digital asset balances before an order ever reaches the matching engine.
- Zero-loss in-flight reservation locking preventing double-spending during bursts of concurrent order submissions.
- Strict mathematical enforcement of the statutory 0.00% (No fee at all) (0.00% fee / 0 bps at launch) platform fee invariant computed in USDT.
- Deterministic state machine transitions with millisecond-precision audit logging for regulatory surveillance.
- Seamless post-match handoff to the on-chain atomic `SettlementDvP.sol` smart contract on Hyperledger Besu.

---

## What You Are Building
A high-performance, ultra-low latency Go microservice (`services/btc-order-service`) built with Go 1.22+ that handles all order routing and lifecycle operations for the spot BTC/USDT market. Concrete deliverables include:

- **Multi-Order Intake & Validation Pipeline:** Ingestion endpoints supporting Market, Limit, and Stop-Limit orders with strict tick-size (0.01 USDT), lot-size (0.00001 BTC / 1,000 satoshis), and minimum notional (5.00 USDT) enforcement.
- **Pre-Trade Balance Reservation Coordinator:** A distributed coordinator that interfaces synchronously with the Real Wallet Service (Prompt 203) to place pessimistic reservation holds:
  - *Buy Orders:* Holds `(Quantity * Price) + Fee` in USDT for Limit orders, or `Estimated Notional + Slippage Buffer + Fee` for Market orders.
  - *Sell Orders:* Holds `Quantity` in BTC (8 decimal places).
- **In-Flight Reservation Lock Manager:** High-speed distributed locking engine utilizing Redis Cluster and atomic Lua scripts to prevent double-spending or negative balance excursions during in-flight order routing.
- **Strict 0.00% (Zero Fee) Platform Fee Calculation Engine:** Enforces the invariant 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) platform fee on turnover charged in USDT, computed via 128-bit fixed-point decimal arithmetic with Banker's Rounding (half-to-even) and verified against Fee Engine (Prompt 210).
- **Deterministic Order State Machine:** Non-reversible, acyclic state engine managing state transitions: `SUBMITTED`, `PENDING_RESERVATION`, `RESERVED`, `ROUTED_TO_ENGINE`, `PARTIALLY_FILLED`, `FILLED`, `PENDING_CANCELLATION`, `CANCELLED`, `REJECTED`, and `EXPIRED`.
- **Stop-Limit Trigger Engine:** Real-time observer consuming live trade price feeds from Market Data Service (Prompt 207) that monitors stop conditions (Stop Price $\ge$ Mark Price or Stop Price $\le$ Mark Price) and automatically promotes triggered Stop-Limit orders to active Limit orders.
- **Kafka Command Publisher & Execution Consumer:**
  - Publishes validated, funded orders to Kafka topic `btc.matching.commands.v1` partitioned strictly by symbol (`BTC-USDT`).
  - Consumes trade match execution events from `btc.matching.events.v1` to update filled quantities, release surplus reservations, and trigger settlement handoffs.
- **Atomic DvP Settlement Handoff Dispatcher:** Packages matched bilateral trade executions and emits verified settlement intents to Trade Settlement Service (Prompt 208) for on-chain execution via `SettlementDvP.sol` on Hyperledger Besu Mainnet.

---

## Scope Boundaries

### In Scope
- Order intake, syntax validation, tick/lot size enforcement, and price collar validation for the `BTC/USDT` trading pair.
- Supported order types: Market, Limit, and Stop-Limit.
- Supported Time-In-Force (TIF) instructions: Good-Til-Cancelled (GTC), Immediate-Or-Cancel (IOC), and Fill-Or-Kill (FOK).
- Synchronous pre-trade balance checks and pessimistic fund/coin reservation holds via Real Wallet Service (Prompt 203).
- Ephemeral in-flight reservation locking in Redis Cluster to eliminate concurrent balance racing.
- Exact 0.00% (Zero Fee) platform fee computation in USDT on all order executions.
- Stop-Limit order resting, trigger condition monitoring, and activation dispatch.
- Order cancellation intake, cancellation routing to matching engine, and automated reservation hold releases.
- Deterministic order state persistence and state transition history auditing in PostgreSQL 16.
- Publishing to `btc.matching.commands.v1` and consuming execution reports from `btc.matching.events.v1`.
- Post-trade execution packaging and settlement intent dispatching for Hyperledger Besu DvP settlement.

### Out of Scope / Handled Elsewhere
- Continuous in-memory Central Limit Order Book (CLOB) price-time priority matching algorithm (handled in Prompt 205).
- Physical on-chain Bitcoin UTXO management, node RPC, Lightning Network channels, and MPC custody vaults (handled in Prompt 234 and Prompt 237).
- Fiat/USDT banking gateway deposit rails and payment settlement (handled in Prompt 212 and Prompt 214).
- KYC verification, AML screening, and user account onboarding (handled in Prompt 201 and Prompt 202).
- Global cross-asset margin lending, collateral haircuts, and liquidation engine (handled in Prompt 206 and Prompt 241).
- Direct EVM smart contract execution and blockchain gas relayer management (handled in Prompt 208, Prompt 245, and Prompt 306).
- Market data aggregation, candlestick charting, and public WebSocket broadcasting (handled in Prompt 207).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Selected for sub-millisecond execution overhead, lightweight goroutine concurrency, strict type safety, zero-allocation memory optimizations, and predictable garbage collection pauses.
- **Relational Persistence & Audit Ledger:** PostgreSQL 16+ utilizing `jackc/pgx/v5` with connection pooling, prepared statements, row-level pessimistic locking (`SELECT ... FOR UPDATE`), and transactional outbox pattern.
- **Compile-Time Type-Safe SQL:** `sqlc` for generating zero-allocation, compile-time verified Go database access code directly from strict SQL schema definitions.
- **In-Memory Cache & Distributed Lock Store:** Redis 7.2+ Cluster utilizing Redis Hashes, sorted sets for Stop-Limit triggers, and atomic Lua scripts for sub-millisecond lock acquisition and single-use idempotency enforcement.
- **Message Streaming & Broker:** Apache Kafka 3.7+ (`segmentio/kafka-go`) with idempotent producers (`enable.idempotence=true`, `acks=all`), strict sequential symbol partition keys, and manual offset commits.
- **RPC Framework & Serialization:** gRPC (`google.golang.org/grpc`) over HTTP/2 with Protocol Buffers v3, structured status codes, and mTLS mutual authentication.
- **High-Precision Fixed-Point Arithmetic:** `github.com/shopspring/decimal` for exact 128-bit decimal calculations without IEEE 754 floating-point rounding errors (8 decimal places for BTC, 6 decimal places for USDT).
- **Observability & Telemetry:** OpenTelemetry Go SDK (`go.opentelemetry.io/otel`) with microsecond-precision trace propagation, Prometheus metrics instrumentation, and structured JSON logging via `uber-go/zap`.

---

## Backend / Infra Touchpoints
- **Order Matching Engine (Prompt 205):** Ingests validated orders from `btc.matching.commands.v1` and emits trade match reports on `btc.matching.events.v1`.
- **Real Wallet Service (Prompt 203):** Invoked via synchronous gRPC (`HoldFunds`, `ReleaseHold`, `CommitSettlement`) to execute pessimistic double-entry balance reservations in user wallets.
- **Pre-Trade Risk & Margin Engine (Prompt 206):** Invoked via gRPC (`EvaluateOrderRisk`) before order routing to verify user trading status, maximum notional size, and price band limits relative to mark price.
- **Fee & Realized PnL Engine (Prompt 210):** Reconciles the invariant 0.00% (No fee at all) platform fee in USDT and validates downstream fee distribution (0.00% fee at launch (governed by FeeController.sol)).
- **Real-Time Market Data Service (Prompt 207):** Streams top-of-book BBO (Best Bid/Offer) and last-trade prices to evaluate Stop-Limit triggers and market order price collars.
- **Trade Settlement & DvP Orchestration Service (Prompt 208):** Receives completed execution packages from `btc-order-service` to coordinate atomic Delivery-versus-Payment settlement on Hyperledger Besu.
- **Immutable Audit Log Service (Prompt 218):** Ingests cryptographic audit trails for every order placement, rejection, transition, and cancellation.

---

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Token Asset Representation:** BTC traded in this spot market is mapped to canonical tokenized Bitcoin (`TokenizedBTC.sol` / `WBTC` equivalent) on Hyperledger Besu, backed 1:1 by physical Bitcoin held in institutional multi-party computation (MPC) cold vaults (Prompt 234 / Prompt 237). USDT is mapped to verified fiat-collateralized stablecoin contracts (`NBSE-USDT`).
- **Atomic DvP Smart Contract Trigger (`SettlementDvP.sol`):**
  - When matching engine trade match events arrive on `btc.matching.events.v1`, the service packages the execution parameters into a DvP settlement payload: `buyer_address`, `seller_address`, `trade_id`, `btc_amount` (in satoshis, $10^{-8}$ precision), `usdt_gross_amount` ($10^{-6}$ precision), and `usdt_fee_amount` (0.00% (Zero Fee) flat fee).
  - This payload is handed off to the Trade Settlement Service (Prompt 208), which signs and invokes `SettlementDvP.sol.settleTrade()` on the Hyperledger Besu Mainnet.
  - Under QBFT consensus (2-second block finality), the smart contract atomically transfers BTC tokens from seller to buyer while simultaneously transferring USDT from buyer to seller and routing the 0.00% fee (No fee at all) to exchange Treasury and SGF vaults.
- **Zero On-Chain PII Invariant:** The Besu ledger records only cryptographic addresses (`0x...`), smart contract token addresses, satoshi/micro-USDT quantities, and the cryptographic hash of `order_id` and `trade_id`. No usernames, email addresses, KYC IDs, PAN numbers, or IP addresses are written to the blockchain.
- **Settlement Binding & Nonce Protection:** Each matched trade is assigned a deterministic `trade_id` hash verified by `SettlementDvP.sol` replay protection. If the on-chain transaction reverts, the settlement coordinator initiates an automated reconciliation saga (Prompt 215) and notifies `btc-order-service` to maintain order state integrity.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Workspace:** Initialize Go module `services/btc-order-service` adhering to clean Domain-Driven Design (DDD):
   - `cmd/server/`: Application initialization, signal handling, and dependency wiring.
   - `internal/domain/order/`: Core entities (`Order`, `Reservation`, `Execution`), value objects (`Price`, `Quantity`, `Fee`), and state transition rules.
   - `internal/application/`: Use-case orchestrators (`PlaceOrderUseCase`, `CancelOrderUseCase`, `ProcessMatchUseCase`, `TriggerStopLimitUseCase`).
   - `internal/infrastructure/postgres/`: SQL migrations, `sqlc` generated queries, and transactional repositories.
   - `internal/infrastructure/redis/`: Distributed lock manager and Stop-Limit sorted sets.
   - `internal/infrastructure/kafka/`: Producer for matching commands and consumer for match events.
   - `internal/infrastructure/grpc/`: gRPC server endpoints and client adapters for external services.
2. **Define Protobuf Service Specifications:** Author `proto/growww/btc_order/v1/btc_order_service.proto` defining RPC methods: `PlaceBtcOrder`, `CancelBtcOrder`, `GetBtcOrder`, `ListBtcOrders`, and `StreamBtcOrderUpdates`.
3. **Compile Protobuf & gRPC Stubs:** Generate Go structs and gRPC interfaces using `buf` or `protoc-gen-go` and `protoc-gen-go-grpc` with strict validation rules.
4. **Design PostgreSQL Schema & Migrations:** Write database migration files establishing tables: `btc_orders`, `btc_order_reservations`, `btc_order_state_transitions`, and `btc_order_executions` with composite indexes and strict check constraints.
5. **Configure `sqlc` Persistence Layer:** Define SQL queries for atomic order insertion, status updating, execution appending, and history queries, generating type-safe Go repositories.
6. **Implement In-Flight Reservation Locking in Redis:** Implement atomic Lua scripts (`acquire_order_lock.lua`, `release_order_lock.lua`) enforcing short-lived (500ms) distributed mutexes on `user_id + asset` to eliminate concurrent double-spend race conditions.
7. **Implement Idempotency Verification:** Author Redis and PostgreSQL idempotency validation checking incoming `idempotency_key`. Duplicate requests within 24 hours return the existing order record without re-executing balance holds or matching engine routing.
8. **Build Strict Order Parameter Validator:** Implement validation rules:
   - Symbol must be strictly `BTC-USDT`.
   - Side must be `BUY` or `SELL`.
   - Tick size validation: Price must be an integer multiple of 0.01 USDT ($10^{-2}$).
   - Lot size validation: BTC quantity must be an integer multiple of 0.00001 BTC (1,000 satoshis, $10^{-5}$ lot step).
   - Minimum notional: `Price * Quantity >= 5.00 USDT`.
   - Price collar check: Limit price must be within +/- 5% of current market mark price.
9. **Implement Invariant 0.00% fee (No fee at all) Calculator:** Implement high-precision fee engine using `decimal.Decimal`:
   - Calculate fee: $\\text{Fee}_{\\text{USDT}} = \\text{Turnover}_{\\text{USDT}} \\times 0.0000 = 0$.
   - Enforce Banker's Rounding (half-to-even) to 4 decimal places (minimum 0.0001 USDT).
   - Ensure buy order balance reservation accounts for both nominal order value and full estimated fee.
10. **Build Pre-Trade Balance Reservation Coordinator:** Coordinate pre-trade locks with Real Wallet Service (Prompt 203) and Pre-Trade Risk Engine (Prompt 206):
    - For Buy orders: Call `WalletService.HoldFunds` for `(Price * Quantity) + Fee` in USDT.
    - For Sell orders: Call `WalletService.HoldFunds` for `Quantity` in BTC.
    - If reservation fails, atomically transition order to `REJECTED` and abort pipeline.
11. **Implement Deterministic Order State Machine:** Code the finite state machine enforcing strict transition graphs (e.g., `SUBMITTED` -> `RESERVED` -> `ROUTED_TO_ENGINE` -> `PARTIALLY_FILLED` -> `FILLED`). Persist every transition to `btc_order_state_transitions` with microsecond timestamps.
12. **Implement Stop-Limit Trigger Observer:** Maintain a Redis Sorted Set of resting Stop-Limit orders indexed by stop price. Subscribe to live price ticks from Market Data Service (Prompt 207); when mark price crosses trigger threshold, automatically promote order to active Limit order and dispatch to matching engine.
13. **Implement Kafka Command Producer:** Build Kafka producer publishing validated orders to `btc.matching.commands.v1` with message key set to `BTC-USDT` to guarantee partition serialization, using `acks=all` and transactional delivery.
14. **Implement Match Event Consumer & DvP Settlement Handoff:** Build consumer for `btc.matching.events.v1`:
    - On match report: Update `filled_quantity`, calculate realized fee in USDT, append execution record to `btc_order_executions`.
    - If fully filled: Transition state to `FILLED` and dispatch DvP settlement package to Trade Settlement Service (Prompt 208) for Besu on-chain execution.
    - If partially filled: Transition state to `PARTIALLY_FILLED`.
    - Release unneeded balance reservation holds if fill price is better than limit price.
15. **Implement Cancellation Pipeline:** Build `CancelBtcOrder` endpoint:
    - Route cancellation request to matching engine via `btc.matching.commands.v1`.
    - Upon engine confirmation (`ORDER_CANCELLED`), update order state to `CANCELLED` and immediately release all unexecuted reservation holds in Wallet Service within 50ms.
16. **Build Comprehensive Test Suite:** Author unit, property-based, and integration tests using `testcontainers-go` (PostgreSQL 16, Redis 7.2, Kafka 3.7) verifying concurrent balance reservations, Stop-Limit activations, partial fill progressions, and zero-slippage fee enforcement under heavy load.

---

## Interfaces / Contracts

### Protobuf Definition (`btc_order_service.proto`)

```protobuf
syntax = "proto3";

package growww.btc_order.v1;

option go_package = "growww/btc_order/v1;btcorderv1";

// BtcOrderService manages the lifecycle of spot BTC/USDT orders.
service BtcOrderService {
  // Place a new spot BTC/USDT order (Market, Limit, Stop-Limit).
  rpc PlaceBtcOrder (PlaceBtcOrderRequest) returns (PlaceBtcOrderResponse);

  // Cancel an active resting order.
  rpc CancelBtcOrder (CancelBtcOrderRequest) returns (CancelBtcOrderResponse);

  // Retrieve current state and execution history of an order.
  rpc GetBtcOrder (GetBtcOrderRequest) returns (GetBtcOrderResponse);

  // List orders for a specific user with pagination and status filters.
  rpc ListBtcOrders (ListBtcOrdersRequest) returns (ListBtcOrdersResponse);

  // Stream live order updates and execution fills for a user.
  rpc StreamBtcOrderUpdates (StreamBtcOrderUpdatesRequest) returns (stream BtcOrderUpdate);
}

enum BtcOrderSide {
  BTC_ORDER_SIDE_UNSPECIFIED = 0;
  BTC_ORDER_SIDE_BUY = 1;
  BTC_ORDER_SIDE_SELL = 2;
}

enum BtcOrderType {
  BTC_ORDER_TYPE_UNSPECIFIED = 0;
  BTC_ORDER_TYPE_LIMIT = 1;
  BTC_ORDER_TYPE_MARKET = 2;
  BTC_ORDER_TYPE_STOP_LIMIT = 3;
}

enum BtcTimeInForce {
  BTC_TIME_IN_FORCE_UNSPECIFIED = 0;
  BTC_TIME_IN_FORCE_GTC = 1; // Good-Til-Cancelled
  BTC_TIME_IN_FORCE_IOC = 2; // Immediate-Or-Cancel
  BTC_TIME_IN_FORCE_FOK = 3; // Fill-Or-Kill
}

enum BtcOrderStatus {
  BTC_ORDER_STATUS_UNSPECIFIED = 0;
  BTC_ORDER_STATUS_SUBMITTED = 1;
  BTC_ORDER_STATUS_PENDING_RESERVATION = 2;
  BTC_ORDER_STATUS_RESERVED = 3;
  BTC_ORDER_STATUS_ROUTED_TO_ENGINE = 4;
  BTC_ORDER_STATUS_PARTIALLY_FILLED = 5;
  BTC_ORDER_STATUS_FILLED = 6;
  BTC_ORDER_STATUS_PENDING_CANCELLATION = 7;
  BTC_ORDER_STATUS_CANCELLED = 8;
  BTC_ORDER_STATUS_REJECTED = 9;
  BTC_ORDER_STATUS_EXPIRED = 10;
}

enum BtcStopCondition {
  BTC_STOP_CONDITION_UNSPECIFIED = 0;
  BTC_STOP_CONDITION_GE = 1; // Trigger when Mark Price >= Stop Price
  BTC_STOP_CONDITION_LE = 2; // Trigger when Mark Price <= Stop Price
}

message PlaceBtcOrderRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string client_order_id = 3;
  BtcOrderSide side = 4;
  BtcOrderType order_type = 5;
  BtcTimeInForce time_in_force = 6;
  string quantity_btc = 7;     // Fixed-point string (e.g. "0.05000000")
  string price_usdt = 8;       // Fixed-point string (e.g. "64500.50"), optional for MARKET
  string stop_price_usdt = 9;  // Required for STOP_LIMIT orders
  BtcStopCondition stop_condition = 10;
}

message PlaceBtcOrderResponse {
  string order_id = 1;
  string client_order_id = 2;
  BtcOrderStatus status = 3;
  string symbol = 4;           // Always "BTC-USDT"
  BtcOrderSide side = 5;
  BtcOrderType order_type = 6;
  string quantity_btc = 7;
  string price_usdt = 8;
  string reserved_asset = 9;   // "USDT" for BUY, "BTC" for SELL
  string reserved_amount = 10;
  string estimated_fee_usdt = 11;
  int64 submitted_at_unix_ms = 12;
}

message CancelBtcOrderRequest {
  string order_id = 1;
  string user_id = 2;
  string reason = 3;
}

message CancelBtcOrderResponse {
  string order_id = 1;
  bool cancellation_accepted = 2;
  BtcOrderStatus status = 3;
  string released_asset = 4;
  string released_amount = 5;
  int64 cancelled_at_unix_ms = 6;
}

message GetBtcOrderRequest {
  string order_id = 1;
  string user_id = 2;
}

message BtcExecutionDetail {
  string execution_id = 1;
  string trade_id = 2;
  string match_price_usdt = 3;
  string match_quantity_btc = 4;
  string match_quote_usdt = 5;
  string fee_usdt = 6;
  string besu_settlement_tx_hash = 7;
  int64 executed_at_unix_ms = 8;
}

message GetBtcOrderResponse {
  string order_id = 1;
  string client_order_id = 2;
  string user_id = 3;
  string symbol = 4;
  BtcOrderSide side = 5;
  BtcOrderType order_type = 6;
  BtcTimeInForce time_in_force = 7;
  BtcOrderStatus status = 8;
  string price_usdt = 9;
  string quantity_btc = 10;
  string filled_quantity_btc = 11;
  string remaining_quantity_btc = 12;
  string average_fill_price_usdt = 13;
  string cumulative_fee_usdt = 14;
  string rejection_reason = 15;
  repeated BtcExecutionDetail executions = 16;
  int64 created_at_unix_ms = 17;
  int64 updated_at_unix_ms = 18;
}

message ListBtcOrdersRequest {
  string user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  repeated BtcOrderStatus status_filter = 4;
  BtcOrderSide side_filter = 5;
  int64 from_timestamp_unix_ms = 6;
  int64 to_timestamp_unix_ms = 7;
}

message ListBtcOrdersResponse {
  repeated GetBtcOrderResponse orders = 1;
  string next_page_token = 2;
}

message StreamBtcOrderUpdatesRequest {
  string user_id = 1;
}

message BtcOrderUpdate {
  string order_id = 1;
  string client_order_id = 2;
  BtcOrderStatus status = 3;
  string filled_quantity_btc = 4;
  string remaining_quantity_btc = 5;
  string latest_match_price_usdt = 6;
  string cumulative_fee_usdt = 7;
  int64 updated_at_unix_ms = 8;
}
```

---

### PostgreSQL Database Schema DDL

```sql
-- Schema DDL: Spot BTC/USDT Order Management System
CREATE TYPE btc_order_side_enum AS ENUM ('BUY', 'SELL');
CREATE TYPE btc_order_type_enum AS ENUM ('LIMIT', 'MARKET', 'STOP_LIMIT');
CREATE TYPE btc_time_in_force_enum AS ENUM ('GTC', 'IOC', 'FOK');
CREATE TYPE btc_order_status_enum AS ENUM (
    'SUBMITTED',
    'PENDING_RESERVATION',
    'RESERVED',
    'ROUTED_TO_ENGINE',
    'PARTIALLY_FILLED',
    'FILLED',
    'PENDING_CANCELLATION',
    'CANCELLED',
    'REJECTED',
    'EXPIRED'
);
CREATE TYPE btc_stop_condition_enum AS ENUM ('GE', 'LE');

-- Master Orders Table
CREATE TABLE btc_orders (
    order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_order_id VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    wallet_address VARCHAR(42) NOT NULL, -- 0x... Hyperledger Besu address
    symbol VARCHAR(16) NOT NULL DEFAULT 'BTC-USDT',
    side btc_order_side_enum NOT NULL,
    order_type btc_order_type_enum NOT NULL,
    time_in_force btc_time_in_force_enum NOT NULL DEFAULT 'GTC',
    status btc_order_status_enum NOT NULL DEFAULT 'SUBMITTED',
    
    -- Numerical constraints: BTC precision 8 decimals, USDT precision 4 decimals
    price NUMERIC(18, 4), -- NULL for MARKET orders
    stop_price NUMERIC(18, 4), -- NULL unless STOP_LIMIT
    stop_condition btc_stop_condition_enum,
    quantity NUMERIC(18, 8) NOT NULL CHECK (quantity >= 0.00001000),
    filled_quantity NUMERIC(18, 8) NOT NULL DEFAULT 0.00000000 CHECK (filled_quantity >= 0.00000000),
    remaining_quantity NUMERIC(18, 8) NOT NULL CHECK (remaining_quantity >= 0.00000000),
    executed_quote_amount NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    
    -- Universal 0.00% Zero-Fee Invariant (No fee at all) tracking
    fee_rate_bps NUMERIC(6, 2) NOT NULL DEFAULT 0.00 CHECK (fee_rate_bps >= 0.00), -- Invariant 0.00% (Zero Fee at launch)
    cumulative_fee_usdt NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    
    -- Financial hold reference from Real Wallet Service (Prompt 203)
    hold_id UUID,
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Active Balance Reservation Holds Tracking
CREATE TABLE btc_order_reservations (
    reservation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES btc_orders(order_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    asset_symbol VARCHAR(8) NOT NULL CHECK (asset_symbol IN ('BTC', 'USDT')),
    reserved_amount NUMERIC(24, 8) NOT NULL CHECK (reserved_amount > 0),
    released_amount NUMERIC(24, 8) NOT NULL DEFAULT 0 CHECK (released_amount >= 0),
    committed_amount NUMERIC(24, 8) NOT NULL DEFAULT 0 CHECK (committed_amount >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'COMMITTED', 'RELEASED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Immutable State Transition Audit Log
CREATE TABLE btc_order_state_transitions (
    transition_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES btc_orders(order_id) ON DELETE CASCADE,
    from_status btc_order_status_enum NOT NULL,
    to_status btc_order_status_enum NOT NULL,
    reason TEXT,
    transitioned_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trade Execution Fills & Settlement Link Table
CREATE TABLE btc_order_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES btc_orders(order_id) ON DELETE CASCADE,
    trade_id UUID NOT NULL,
    match_id VARCHAR(64) NOT NULL,
    matched_price NUMERIC(18, 4) NOT NULL CHECK (matched_price > 0),
    matched_quantity NUMERIC(18, 8) NOT NULL CHECK (matched_quantity > 0),
    matched_quote_amount NUMERIC(18, 4) NOT NULL CHECK (matched_quote_amount > 0),
    fee_usdt NUMERIC(18, 4) NOT NULL CHECK (fee_usdt >= 0.0000), -- Min fee floor
    counterparty_order_id UUID NOT NULL,
    besu_settlement_tx_hash VARCHAR(66), -- 0x... hash once DvP committed
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance & Query Optimization Indexes
CREATE INDEX idx_btc_orders_user_status ON btc_orders(user_id, status);
CREATE INDEX idx_btc_orders_symbol_status ON btc_orders(symbol, status);
CREATE INDEX idx_btc_orders_client_order_id ON btc_orders(user_id, client_order_id);
CREATE INDEX idx_btc_orders_created_at ON btc_orders(created_at DESC);
CREATE INDEX idx_btc_order_executions_order_id ON btc_order_executions(order_id);
CREATE INDEX idx_btc_order_transitions_order_id ON btc_order_state_transitions(order_id);
```

---

## Security & Compliance Notes
- **Strict Idempotency Keys:** Every order submission requires a client-supplied UUID v4 `idempotency_key`. The service validates keys against Redis (24-hour TTL) and PostgreSQL unique constraints. Re-submitting an existing key returns the current order record and rejects duplicate balance holds or duplicate engine command dispatches.
- **Pessimistic Locking on Balance Reservations:** Balance verification and hold placement execute under strict pessimistic row locks in PostgreSQL (`SELECT ... FOR UPDATE`) or atomic balance reservation calls to Real Wallet Service (Prompt 203). This completely prevents overdrafts, double-allocation, or race conditions during bursts of concurrent orders.
- **Strict 0.00% fee (No fee at all) Invariant:** The 0.00% (Zero Fee) (0 bps (0.00% fee at launch) / 0 bps at launch) platform fee on trade turnover is non-negotiable and immutable. Fee amounts are computed via fixed-point arithmetic (`decimal.Decimal`) using Banker's Rounding (`ROUND_HALF_EVEN`) to 4 decimal places in USDT. Fee evasion or truncation anomalies are prevented at both the application and database constraint levels (`CHECK (fee_rate_bps >= 0.00)`).
- **Price Band & Fat-Finger Protection:** Limit order prices are strictly validated against current mark price collars (+/- 5%). Orders submitted outside dynamic collar limits are rejected immediately at intake with `ERR_PRICE_COLLAR_EXCEEDED`.
- **Anti-Wash Trading Safeguards:** Orders include beneficial owner and client account tags. If an incoming order would match against an opposite resting order owned by the same user or entity, the matching engine applies self-trade prevention (Cancel-Oldest or Cancel-Newest) and logs the event for surveillance.
- **Zero On-Chain PII Invariant:** All data dispatched to `SettlementDvP.sol` on Hyperledger Besu consists exclusively of public Ethereum addresses (`0x...`), numeric quantities, and cryptographic hashes (`trade_id`, `order_id`). Investor identity data, KYC records, and IP addresses remain strictly stored in off-chain, DPDP-compliant domestic databases.

---

## Acceptance Criteria
- [ ] Service successfully accepts Market, Limit, and Stop-Limit orders for `BTC-USDT` with quantities adhering to the 0.00001000 BTC minimum lot size and 0.01 USDT price tick.
- [ ] Pre-trade balance reservation atomically locks USDT for Buy orders (`nominal + 0.00% fee (No fee at all)`) and BTC for Sell orders via Real Wallet Service (Prompt 203) prior to matching engine dispatch.
- [ ] In-flight Redis distributed locks prevent double-spend excursions under 100 concurrent requests per user.
- [ ] Duplicate order submissions with identical `idempotency_key` return HTTP/gRPC success with original order state without duplicating financial holds.
- [ ] Orders failing price band limits, minimum notional (5.00 USDT), or balance reservations are immediately transitioned to `REJECTED` and logged.
- [ ] Kafka events on `btc.matching.commands.v1` are keyed strictly by `BTC-USDT`, preserving total order sequencing per symbol partition.
- [ ] Ingested trade match events from `btc.matching.events.v1` correctly increment `filled_quantity`, deduct the strict 0.00% (Zero Fee) USDT fee, and transition state to `PARTIALLY_FILLED` or `FILLED`.
- [ ] Stop-Limit orders remain resting in Redis until mark price crosses the trigger threshold, immediately promoting them to active Limit orders.
- [ ] Completed executions generate verifiable settlement intents handed off to Trade Settlement Service (Prompt 208) for atomic `SettlementDvP.sol` execution on Hyperledger Besu.
- [ ] Order cancellations and IOC/FOK expirations release remaining unexecuted balance holds back to the user's available wallet balance within 50ms.
- [ ] End-to-end order intake latency (intake -> balance hold -> Kafka publish) is $< 5\text{ms}$ at 5,000 requests/sec.

---

## Suggested Order / Dependencies
- **Prerequisites:**
  - `103_api_design_standards.md` (REST & gRPC Wire Protocol Conventions)
  - `104_event_schema_and_kafka_topic_standards.md` (CloudEvents & Kafka Topic Partitioning)
  - `111_domain_model_core_entities.md` (Canonical Domain Entities & Money Types)
  - `112_idempotency_and_exactly_once_processing.md` (Distributed Idempotency Standards)
  - `203_wallet_account_service.md` (Real Wallet & Double-Entry Account Ledger Service)
  - `206_risk_and_margin_checks_service.md` (Pre-Trade Risk & Margin Limits)
  - `210_fee_and_realized_pnl_engine.md` (Fixed Transaction Fee Engine)
  - `234_bitcoin_lightning_and_taproot_ingress_service.md` (Bitcoin Cold Vault Custody)
- **Parallel Tasks:**
  - `205_order_matching_engine.md` (High-Performance Rust Matching Engine)
  - `207_market_data_service.md` (WebSocket Streaming & Price Feeds)
  - `258_national_best_bid_offer_nbbo_consolidated_tape_engine.md` (Consolidated Mark Price Feed)
- **Downstream Blockers:**
  - `208_trade_settlement_service.md` (Trade Settlement & DvP Orchestration)
  - `306_settlement_dvp_smart_contract.md` (Atomic DvP Smart Contract on Hyperledger Besu)
  - `509_flutter_order_flow.md` (Mobile Crypto Trading Interface)
  - `603_web_trading_dashboard.md` (Web Trading Terminal Interface)
