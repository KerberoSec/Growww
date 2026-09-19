# 273 - Demo / Paper Trading Virtual Matching & Execution Engine (Rust / Go)

## Purpose
Retail and institutional market participants entering digital asset and cryptocurrency markets encounter severe volatility, complex order types, and rapid price discovery regimes. Directly committing real capital into high-beta pairs such as BTC/USDT without operational familiarity risks catastrophic user drawdowns caused by poor execution timing, misunderstanding of order mechanics (such as slippage, market impact, and partial fills), or lack of algorithmic strategy validation.

Conventional paper trading systems deployed by retail brokers suffer from severe design flaws: they execute orders naively against top-of-book mid-prices without modeling Level-2 order book depth depletion, queue priority, latency jitter, or realistic slippage. This imparts a deceptive impression of perfect liquidity and instantaneous zero-cost execution, leading traders to adopt unviable strategies that collapse in production environments.

The **Demo / Paper Trading Virtual Matching & Execution Engine** (`services/demo-matching-engine`) delivers an institutional-grade, zero-risk simulated trading environment. It empowers retail and institutional users to practice BTC/USDT spot and derivative trading against real-time, live-moving market prices using virtual funds. 

The engine models realistic liquidity dynamics, synthetic order book depth, queue priority, and configurable slippage models. Furthermore, it operates under strict architectural, process, and data isolation from production trading infrastructure: virtual funds and orders are managed through dedicated demo pipelines, and simulated on-chain receipts settle on an isolated Hyperledger Besu Testnet with guaranteed zero state bleed into the Mainnet clearing infrastructure.

---

## What You Are Building
A microsecond-capable, deterministic virtual execution engine engineered in Rust (or Go 1.22+) situated at `services/demo-matching-engine`. Core architectural components include:

- **Live Market Data Ingestion & Synthetic L2 Depth Cache:** Consumes streaming tick-by-tick trades, best bid/offer (BBO), and Level-2 order book depth from the Live Feeder (Prompt 272) and Market Data Service (Prompt 207). Dynamically updates an in-memory synthetic limit order book for BTC/USDT and supported pairs.
- **In-Memory Virtual Matching Engine:** Implements high-speed price-time priority matching algorithms capable of handling Market, Limit, Stop-Loss, Stop-Limit, Take-Profit, Immediate-Or-Cancel (IOC), and Fill-Or-Kill (FOK) virtual orders.
- **Realistic Queue Position & Fill Rate Simulator:** Instead of granting immediate fills to resting limit orders when price touches the limit price, the engine places virtual limit orders at the back of the real-market queue at that price level. Fills occur only as actual traded volume is consumed in the live external feed at or through that price level.
- **Configurable Dynamic Slippage & Market Impact Engine:** Evaluates incoming virtual market and large limit orders against synthetic order book depth. Calculates dynamic slippage using parameterized linear and square-root market impact curves ($\Delta P \propto \sigma \sqrt{V_{\text{order}} / V_{\text{book}}}$), accurately penalizing aggressive sizing in thin markets.
- **Synthetic Latency Injector:** Introduces configurable execution latency (e.g., 20ms to 80ms) and network jitter to emulate realistic gateway round-trips, preventing artificial latency-arbitrage exploits in paper trading.
- **Demo Balance Reservation & Settlement Coordinator:** Interacts via low-latency gRPC with the Demo Wallet Service (Prompt 274) to verify and lock virtual USDT/BTC balances before order acceptance, debiting and crediting virtual ledgers upon execution.
- **Simulated On-Chain DvP Relayer:** Emits simulated settlement receipts to an isolated Hyperledger Besu Testnet (QBFT consensus, Chain ID 1338) via `DemoSettlementDvP.sol`. Mints educational ERC-20/ERC-3643 tokens representing virtual assets for compliance and smart contract auditing workflows without touching Mainnet.
- **Kafka Event Streamer:** Emits virtual order status transitions, execution reports, and synthetic trade ticks to dedicated demo topics (`demo.orders.events.v1`, `demo.trades.executed.v1`).

---

## Scope Boundaries

### In Scope
- Real-time ingestion of live market ticker, trade ticks, and L2 depth diffs for BTC/USDT, ETH/USDT, and tokenized digital assets from Live Feeder (Prompt 272) and Market Data Service (Prompt 207).
- In-memory order book simulation maintaining synthetic resting liquidity and order queue depth.
- Full virtual order lifecycle management: placement, validation, conditional triggering, partial fills, full fills, cancellations, and replacements.
- Support for order types: Market, Limit, Stop-Loss Market, Stop-Loss Limit, Take-Profit Market, Take-Profit Limit, GTC, IOC, and FOK.
- Configurable slippage simulation based on live order book volume curves and historical tick volatility.
- Queue priority tracking for virtual limit orders matching against live market trade prints.
- Virtual fund holds, releases, and double-entry settlements coordinating with Demo Wallet Service (Prompt 274).
- Asynchronous dispatch of simulated trade execution receipts to Hyperledger Besu Testnet.
- Complete logical, physical, and process isolation between Demo accounts and production Mainnet accounts.
- Protobuf contracts, PostgreSQL schemas for demo order history, Redis key conventions, and Kafka event producers.

### Out of Scope / Handled Elsewhere
- Live external exchange WebSocket connections and raw feed normalization (handled in Live Feeder, Prompt 272).
- Production central limit order book matching and real-money equity/crypto execution (handled in Order Matching Engine, Prompt 205).
- Real fiat deposits, banking payment gateways, and UPI/IMPS collections (handled in Payment Gateway Service, Prompt 212).
- Real-money double-entry wallet accounts and bank integrations (handled in Wallet & Account Service, Prompt 203).
- Real-money margin lending, collateral haircuts, and liquidation engine (handled in Risk Engine Prompt 206 and ADL Resolver Prompt 268).
- Production Hyperledger Besu Mainnet settlement contracts (handled in Trade Settlement Service Prompt 208 and SettlementDvP Prompt 306).
- User identity onboarding, PAN/Aadhaar verification, and CKYC checks (handled in User Service Prompt 201 and KYC Service Prompt 202).

---

## Technology to Use
- **Primary Language & Runtime:** Rust 1.78+ (Tokio, Tonic, Prost, Crossbeam) or Go 1.22+ (Goroutines, gRPC-Go). Rust is the preferred runtime for the virtual matching engine core due to its deterministic execution latency, zero garbage collection pauses, memory safety, and high-density cache-friendly data structures.
- **In-Memory Synthetic Order Book:** `std::collections::BTreeMap` (Rust) or concurrent skip-lists / red-black trees (Go) mapping discrete price ticks to FIFO order volume queues.
- **High-Precision Financial Arithmetic:** Fixed-point 128-bit integers or `rust_decimal` / `shopspring/decimal` using 8 decimal places for BTC ($10^{-8}$ satoshis) and 6 decimal places for USDT ($10^{-6}$ micro-USDT), eliminating floating-point rounding errors.
- **In-Memory Cache & Ephemeral State:** Redis 7.2+ Cluster (dedicated demo keyspaces or dedicated demo cluster) for active virtual order indexes, ticker caches, user active order sets, and distributed locks via atomic Lua scripts.
- **Relational Persistence & Historical Audit:** PostgreSQL 16+ utilizing `sqlx` (Rust) or `pgx/v5` / `sqlc` (Go) for persistent demo order records, fill audits, and historical performance metrics.
- **Message Broker & Event Streaming:** Apache Kafka 3.7+ (`rdkafka` or `segmentio/kafka-go`) with strictly isolated demo topic namespaces (`demo.orders.*`, `demo.trades.*`).
- **Simulated Blockchain Client:** `ethers-rs` (Rust) or `go-ethereum/ethclient` (Go) connecting to isolated Hyperledger Besu Testnet nodes (Chain ID 1338, QBFT consensus).
- **Telemetry & Observability:** OpenTelemetry (OTel) instrumentation with Prometheus metrics (`demo_matching_latency_microseconds`, `demo_slippage_bps`, `demo_orders_active_total`) and Jaeger tracing.

---

## Backend / Infra Touchpoints
- **Live Feeder (Prompt 272):** Streams ultra-low-latency, normalized market data ticks, best bid/ask updates, and order book snapshots for BTC/USDT directly to the demo matching engine over gRPC or zero-copy IPC channels.
- **Market Data Service (Prompt 207):** Provides 24-hour volume statistics, consolidated mark prices, index prices, and historical OHLCV candlestick data for chart rendering in paper trading mode.
- **Demo Wallet Service (Prompt 274):** The authority for virtual user balances. Handles pre-trade virtual balance holds, trade execution settlements, fee deductions in virtual USDT, and account resets (e.g., replenishing 100,000 virtual USDT faucet).
- **Flutter & Web Client (Prompts 509, 527, 603):** Renders the simulated trading interface. Subscribes to bi-directional gRPC/WebSocket streams for virtual order placement, depth charts, real-time fill notifications, and position PnL dashboards. Displays prominent visual indicators ("DEMO TRADING / PAPER MODE") across all UI views.
- **Audit & Analytics Service (Prompt 218):** Ingests simulated execution metrics to generate trader execution reports, comparing paper execution quality and slippage against historical benchmark benchmarks.

---

## Blockchain Interaction (Simulated On-Chain Receipts on Hyperledger Besu Testnet, Isolated from Mainnet)
- **Dedicated Testnet Deployment (`DemoSettlementDvP.sol`):** The demo execution engine submits simulated trade settlement receipts exclusively to an isolated Hyperledger Besu Testnet running under QBFT consensus (Chain ID 1338, 1-second block time, zero real economic value).
- **Educational On-Chain DvP:** For every executed virtual match, the engine asynchronously submits an educational DvP receipt to `DemoSettlementDvP.sol`. The contract mints and transfers simulated ERC-20 / ERC-3643 demo tokens (e.g., `dUSDT`, `dBTC`) between virtual user escrow accounts.
- **Complete Mainnet Firefall:** 
  - The demo engine has zero network connectivity, credentials, or RPC endpoints pointing to Hyperledger Besu Mainnet (Chain ID 1337).
  - Testnet relayer accounts use throwaway private keys funded by a testnet faucet.
  - Hardcoded runtime assertions prevent execution if connected chain ID matches production Mainnet.
- **Zero PII Invariant:** Testnet transactions record only pseudonymized virtual account hashes (`0x...`), synthetic asset identifiers, quantities, execution rates, and the internal `demo_trade_id`. No personal data or user identifiers are ever written to the blockchain.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Workspace:** Initialize the Rust crate (or Go module) at `services/demo-matching-engine` with standardized directory layout:
   - `src/book/`: In-memory synthetic limit order book and queue structures.
   - `src/matcher/`: Matching engine state machine and order execution algorithms.
   - `src/slippage/`: Configurable slippage, market impact, and latency models.
   - `src/feeder/`: Ingestion clients for Live Feeder (Prompt 272) and Market Data (Prompt 207).
   - `src/wallet/`: gRPC client for Demo Wallet Service (Prompt 274).
   - `src/blockchain/`: Hyperledger Besu Testnet DvP relayer client.
   - `src/storage/`: PostgreSQL persistence models and repository queries.
   - `src/api/`: gRPC and REST server handlers (`DemoOrderService`).
2. **Define Protobuf Specifications:** Author `proto/growww/demo/v1/demo_order_service.proto` defining RPCs: `PlaceDemoOrder`, `CancelDemoOrder`, `ReplaceDemoOrder`, `GetDemoOrder`, `ListActiveDemoOrders`, `StreamDemoOrders`, and `StreamDemoL2Depth`.
3. **Compile Protobuf & gRPC Stubs:** Generate type-safe Rust/Go stubs using `buf` or `tonic-build` / `protoc-gen-go`.
4. **Design PostgreSQL Database Migrations:** Write database migration scripts creating tables: `demo_accounts`, `demo_orders`, `demo_trades`, `demo_positions`, and `demo_slippage_configs`. Create indexes optimized for active order retrieval and user fill lookups.
5. **Implement Fixed-Point Arithmetic & Precision Utilities:** Implement arithmetic structures for BTC quantity (8 decimal places) and USDT price/value (6 decimal places) with strict overflow protection and Banker's Rounding (`ROUND_HALF_EVEN`).
6. **Build In-Memory Synthetic Order Book:** Implement `SyntheticOrderBook` using sorted trees (`BTreeMap<Price, LevelQueue>`). Populate and update bids and asks dynamically from incoming L2 market depth updates from Live Feeder (Prompt 272).
7. **Implement Market Data Consumer:** Construct a high-throughput gRPC and Kafka consumer subscribing to market updates from Live Feeder (Prompt 272). Maintain the latest BBO, microsecond trade ticks, and 24-hour VWAP in lock-free memory structures.
8. **Build Virtual Order Lifecycle State Machine:** Implement atomic order status transitions: `PENDING` -> `OPEN` -> `PARTIALLY_FILLED` -> `FILLED`, or termination states `CANCELLED`, `REJECTED`, `EXPIRED`.
9. **Implement Virtual Market Order Execution & Slippage Engine:**
   - Consume synthetic order book depth.
   - Calculate volume-weighted average price (VWAP) through the book.
   - Apply dynamic slippage formula:
     $$P_{\text{fill}} = P_{\text{mid}} \times \left(1 \pm \frac{\text{Spread}}{2} \pm \gamma \cdot \left(\frac{Q_{\text{order}}}{Q_{\text{depth}}}\right)^{\alpha}\right)$$
     where $\gamma$ is the slippage coefficient (e.g., 0.05) and $\alpha$ is the market impact exponent (typically 0.5 to 1.0).
10. **Implement Virtual Limit Order & Queue Priority Simulator:**
    - When a limit order is placed inside the spread, register it in the virtual book.
    - When a limit order is placed at or behind the current market price, record the accumulated market volume ahead of it in the queue ($V_{\text{ahead}}$).
    - Track incoming real-market trades from the live feed. Decrement $V_{\text{ahead}}$ by the volume of trades occurring at that price level.
    - Trigger virtual fills only when $V_{\text{ahead}} \le 0$, realistically simulating true queue delay.
11. **Implement Stop-Loss, Take-Profit, and Trigger Monitor:** Build an in-memory trigger evaluation loop that checks pending conditional orders against live mark and last-trade prices on every tick. Immediately activate triggered orders into executable Market or Limit orders.
12. **Integrate with Demo Wallet Service (Prompt 274):** Implement gRPC two-phase virtual balance holds:
    - *Pre-Order Check:* Request virtual balance reservation (e.g., virtual USDT for buy orders, virtual BTC for sell orders).
    - *Fill Execution:* Commit debit/credit instructions to update virtual cash and asset holdings upon trade match.
    - *Order Cancellation:* Release unused balance holds immediately.
13. **Build Kafka Event Publisher:** Publish virtual execution reports and trade ticks to Kafka topics `demo.orders.events.v1` and `demo.trades.executed.v1`.
14. **Implement Hyperledger Besu Testnet Relayer:** Construct an asynchronous background worker that batches virtual trade matches, constructs DvP transactions, and submits them to `DemoSettlementDvP.sol` on the Besu Testnet.
15. **Implement Comprehensive Test Suites:** Author unit and integration tests using `testcontainers`:
    - Verify zero fill on limit orders when market trades do not exceed queue volume ahead.
    - Verify dynamic slippage matches theoretical curves across order sizes.
    - Verify 100% data and account isolation between Demo and Production systems.
    - Measure matching latency under synthetic load of 50,000 orders/sec.

---

## Interfaces / Contracts

### Protobuf Definition (`demo_order_service.proto`)

```protobuf
syntax = "proto3";

package growww.demo.v1;

option go_package = "growww/demo/v1;demov1";

// DemoOrderService provides virtual / paper trading matching and execution.
service DemoOrderService {
  // Place a new virtual order against live market prices.
  rpc PlaceDemoOrder (PlaceDemoOrderRequest) returns (PlaceDemoOrderResponse);

  // Cancel an active virtual order.
  rpc CancelDemoOrder (CancelDemoOrderRequest) returns (CancelDemoOrderResponse);

  // Replace or modify an active virtual order (price or quantity).
  rpc ReplaceDemoOrder (ReplaceDemoOrderRequest) returns (ReplaceDemoOrderResponse);

  // Retrieve details of a specific virtual order.
  rpc GetDemoOrder (GetDemoOrderRequest) returns (GetDemoOrderResponse);

  // List all active or historical virtual orders for a demo account.
  rpc ListActiveDemoOrders (ListActiveDemoOrdersRequest) returns (ListActiveDemoOrdersResponse);

  // Stream real-time order lifecycle events and fill notifications.
  rpc StreamDemoOrders (StreamDemoOrdersRequest) returns (stream DemoOrderEvent);

  // Stream synthetic Level-2 order book depth for the demo environment.
  rpc StreamDemoL2Depth (StreamDemoL2DepthRequest) returns (stream DemoL2DepthSnapshot);

  // Reset a demo account balance and purge active virtual positions.
  rpc ResetDemoAccount (ResetDemoAccountRequest) returns (ResetDemoAccountResponse);
}

enum DemoOrderSide {
  DEMO_ORDER_SIDE_UNSPECIFIED = 0;
  DEMO_ORDER_SIDE_BUY = 1;
  DEMO_ORDER_SIDE_SELL = 2;
}

enum DemoOrderType {
  DEMO_ORDER_TYPE_UNSPECIFIED = 0;
  DEMO_ORDER_TYPE_LIMIT = 1;
  DEMO_ORDER_TYPE_MARKET = 2;
  DEMO_ORDER_TYPE_STOP_LOSS = 3;
  DEMO_ORDER_TYPE_STOP_LIMIT = 4;
  DEMO_ORDER_TYPE_TAKE_PROFIT = 5;
  DEMO_ORDER_TYPE_TAKE_PROFIT_LIMIT = 6;
}

enum DemoTimeInForce {
  DEMO_TIME_IN_FORCE_UNSPECIFIED = 0;
  DEMO_TIME_IN_FORCE_GTC = 1; // Good 'Til Cancelled
  DEMO_TIME_IN_FORCE_IOC = 2; // Immediate Or Cancel
  DEMO_TIME_IN_FORCE_FOK = 3; // Fill Or Kill
}

enum DemoOrderStatus {
  DEMO_ORDER_STATUS_UNSPECIFIED = 0;
  DEMO_ORDER_STATUS_PENDING = 1;
  DEMO_ORDER_STATUS_OPEN = 2;
  DEMO_ORDER_STATUS_PARTIALLY_FILLED = 3;
  DEMO_ORDER_STATUS_FILLED = 4;
  DEMO_ORDER_STATUS_CANCELLED = 5;
  DEMO_ORDER_STATUS_REJECTED = 6;
  DEMO_ORDER_STATUS_EXPIRED = 7;
}

message PlaceDemoOrderRequest {
  string demo_account_id = 1;
  string symbol = 2;                  // e.g., "BTC-USDT", "ETH-USDT"
  DemoOrderSide side = 3;
  DemoOrderType order_type = 4;
  DemoTimeInForce time_in_force = 5;
  string quantity = 6;                // Fixed-point string (e.g., "0.50000000")
  string price = 7;                   // Optional for market orders (e.g., "65250.500000")
  string trigger_price = 8;           // Required for stop/take-profit orders
  string idempotency_key = 9;
  string slippage_model_id = 10;      // Optional override for simulation profile
}

message PlaceDemoOrderResponse {
  string demo_order_id = 1;
  string demo_account_id = 2;
  string symbol = 3;
  DemoOrderStatus status = 4;
  string original_quantity = 5;
  string executed_quantity = 6;
  string cumulative_quote_value = 7;
  string average_fill_price = 8;
  string estimated_slippage_bps = 9;
  int64 created_at_unix_ms = 10;
}

message CancelDemoOrderRequest {
  string demo_order_id = 1;
  string demo_account_id = 2;
  string symbol = 3;
}

message CancelDemoOrderResponse {
  string demo_order_id = 1;
  DemoOrderStatus status = 2;
  string unexecuted_quantity = 3;
  int64 cancelled_at_unix_ms = 4;
}

message ReplaceDemoOrderRequest {
  string demo_order_id = 1;
  string demo_account_id = 2;
  string symbol = 3;
  string new_price = 4;
  string new_quantity = 5;
  string idempotency_key = 6;
}

message ReplaceDemoOrderResponse {
  string new_demo_order_id = 1;
  string previous_demo_order_id = 2;
  DemoOrderStatus status = 3;
  string updated_price = 4;
  string updated_quantity = 5;
  int64 updated_at_unix_ms = 6;
}

message GetDemoOrderRequest {
  string demo_order_id = 1;
  string demo_account_id = 2;
}

message GetDemoOrderResponse {
  DemoOrderDetail order = 1;
}

message ListActiveDemoOrdersRequest {
  string demo_account_id = 1;
  string symbol = 2;                  // Optional filter
  int32 page_size = 3;
  string page_token = 4;
}

message ListActiveDemoOrdersResponse {
  repeated DemoOrderDetail orders = 1;
  string next_page_token = 2;
}

message DemoOrderDetail {
  string demo_order_id = 1;
  string demo_account_id = 2;
  string symbol = 3;
  DemoOrderSide side = 4;
  DemoOrderType order_type = 5;
  DemoTimeInForce time_in_force = 6;
  DemoOrderStatus status = 7;
  string original_quantity = 8;
  string executed_quantity = 9;
  string remaining_quantity = 10;
  string limit_price = 11;
  string trigger_price = 12;
  string average_fill_price = 13;
  string virtual_fee_debited = 14;
  int64 created_at_unix_ms = 15;
  int64 updated_at_unix_ms = 16;
}

message StreamDemoOrdersRequest {
  string demo_account_id = 1;
}

message DemoOrderEvent {
  string event_id = 1;
  string demo_order_id = 2;
  string demo_account_id = 3;
  string symbol = 4;
  DemoOrderStatus status = 5;
  string fill_quantity = 6;
  string fill_price = 7;
  string virtual_fee = 8;
  string slippage_bps = 9;
  string on_chain_tx_hash = 10;       // Besu Testnet receipt hash
  int64 event_time_unix_ms = 11;
}

message StreamDemoL2DepthRequest {
  string symbol = 1;
  int32 depth_levels = 2;             // e.g., 20 or 50
}

message DepthLevel {
  string price = 1;
  string quantity = 2;
  int32 order_count = 3;
}

message DemoL2DepthSnapshot {
  string symbol = 1;
  repeated DepthLevel bids = 2;
  repeated DepthLevel asks = 3;
  int64 sequence_number = 4;
  int64 timestamp_unix_ms = 5;
}

message ResetDemoAccountRequest {
  string demo_account_id = 1;
  string initial_virtual_usdt = 2;   // e.g., "100000.000000"
}

message ResetDemoAccountResponse {
  string demo_account_id = 1;
  bool success = 2;
  string reset_balance_usdt = 3;
  int32 cancelled_orders_count = 4;
  int64 reset_at_unix_ms = 5;
}
```

---

### PostgreSQL Database Schema DDL

```sql
-- PostgreSQL 16+ DDL Schema for Demo / Paper Trading Engine
-- Strictly isolated in dedicated database or schema: growww_demo

CREATE SCHEMA IF NOT EXISTS growww_demo;
SET search_path TO growww_demo, public;

CREATE TYPE demo_order_side_enum AS ENUM ('BUY', 'SELL');
CREATE TYPE demo_order_type_enum AS ENUM (
    'LIMIT', 
    'MARKET', 
    'STOP_LOSS', 
    'STOP_LIMIT', 
    'TAKE_PROFIT', 
    'TAKE_PROFIT_LIMIT'
);
CREATE TYPE demo_time_in_force_enum AS ENUM ('GTC', 'IOC', 'FOK');
CREATE TYPE demo_order_status_enum AS ENUM (
    'PENDING', 
    'OPEN', 
    'PARTIALLY_FILLED', 
    'FILLED', 
    'CANCELLED', 
    'REJECTED', 
    'EXPIRED'
);

-- Table 1: Demo Trading Accounts (Zero real capital, isolated from Mainnet)
CREATE TABLE demo_accounts (
    demo_account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,                -- References user in User Service, but tagged as demo
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    initial_balance_usdt NUMERIC(28, 6) NOT NULL DEFAULT 100000.000000,
    current_balance_usdt NUMERIC(28, 6) NOT NULL DEFAULT 100000.000000,
    locked_balance_usdt NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    reset_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 2: Simulated Orders
CREATE TABLE demo_orders (
    demo_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    demo_account_id UUID NOT NULL REFERENCES demo_accounts(demo_account_id) ON DELETE CASCADE,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    symbol VARCHAR(32) NOT NULL,                 -- e.g., 'BTC-USDT'
    side demo_order_side_enum NOT NULL,
    order_type demo_order_type_enum NOT NULL,
    time_in_force demo_time_in_force_enum NOT NULL DEFAULT 'GTC',
    status demo_order_status_enum NOT NULL DEFAULT 'PENDING',
    original_quantity NUMERIC(28, 8) NOT NULL CHECK (original_quantity > 0),
    executed_quantity NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000 CHECK (executed_quantity >= 0),
    remaining_quantity NUMERIC(28, 8) GENERATED ALWAYS AS (original_quantity - executed_quantity) STORED,
    limit_price NUMERIC(28, 6),
    trigger_price NUMERIC(28, 6),
    average_fill_price NUMERIC(28, 6) DEFAULT 0.000000,
    virtual_fee_debited NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    queue_volume_ahead NUMERIC(28, 8) DEFAULT 0.00000000,
    rejection_reason TEXT,
    client_submitted_at_unix_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 3: Simulated Trades / Fills
CREATE TABLE demo_trades (
    demo_trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    demo_order_id UUID NOT NULL REFERENCES demo_orders(demo_order_id) ON DELETE RESTRICT,
    demo_account_id UUID NOT NULL REFERENCES demo_accounts(demo_account_id),
    symbol VARCHAR(32) NOT NULL,
    side demo_order_side_enum NOT NULL,
    fill_price NUMERIC(28, 6) NOT NULL CHECK (fill_price > 0),
    fill_quantity NUMERIC(28, 8) NOT NULL CHECK (fill_quantity > 0),
    quote_amount NUMERIC(28, 6) NOT NULL CHECK (quote_amount > 0),
    virtual_fee NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    slippage_bps NUMERIC(10, 4) NOT NULL DEFAULT 0.0000,
    reference_mid_price NUMERIC(28, 6) NOT NULL,
    testnet_tx_hash VARCHAR(66),                 -- 0x-prefixed 32-byte hash on Besu Testnet
    executed_at_unix_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 4: Current Virtual Open Positions
CREATE TABLE demo_positions (
    position_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    demo_account_id UUID NOT NULL REFERENCES demo_accounts(demo_account_id) ON DELETE CASCADE,
    symbol VARCHAR(32) NOT NULL,
    net_quantity NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000,
    entry_vwap NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    realized_pnl_usdt NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    unrealized_pnl_usdt NUMERIC(28, 6) NOT NULL DEFAULT 0.000000,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_demo_account_symbol UNIQUE (demo_account_id, symbol)
);

-- Table 5: Configurable Slippage & Simulation Models
CREATE TABLE demo_slippage_configs (
    model_id VARCHAR(64) PRIMARY KEY,
    model_name VARCHAR(128) NOT NULL,
    linear_coefficient NUMERIC(10, 6) NOT NULL DEFAULT 0.010000,
    sqrt_impact_coefficient NUMERIC(10, 6) NOT NULL DEFAULT 0.050000,
    min_latency_ms INT NOT NULL DEFAULT 20,
    max_latency_ms INT NOT NULL DEFAULT 80,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Default Simulation Profiles
INSERT INTO demo_slippage_configs (model_id, model_name, linear_coefficient, sqrt_impact_coefficient, min_latency_ms, max_latency_ms)
VALUES
    ('DEFAULT_RETAIL', 'Standard Retail Volatility & Latency', 0.005000, 0.030000, 25, 65),
    ('INSTITUTIONAL_FAST', 'Direct Market Access (Low Latency)', 0.001000, 0.015000, 5, 15),
    ('HIGH_SLIPPAGE_STRESS', 'Stress Simulation (Extreme Volatility)', 0.020000, 0.100000, 50, 150);

-- Indexes for rapid order book lookups, triggers, and user queries
CREATE INDEX idx_demo_orders_active ON demo_orders (symbol, status) WHERE status IN ('OPEN', 'PARTIALLY_FILLED');
CREATE INDEX idx_demo_orders_triggers ON demo_orders (symbol, trigger_price) WHERE status = 'PENDING' AND trigger_price IS NOT NULL;
CREATE INDEX idx_demo_orders_account ON demo_orders (demo_account_id, created_at DESC);
CREATE INDEX idx_demo_trades_order ON demo_trades (demo_order_id);
CREATE INDEX idx_demo_trades_account ON demo_trades (demo_account_id, created_at DESC);
```

---

### Kafka Event Schemas

#### Topic: `demo.trades.executed.v1`
Key: `symbol` (e.g., `"BTC-USDT"`)
```json
{
  "event_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "demo_trade_id": "c1f2b6e4-4d1e-4c7b-b3f8-8a8c8e1f5d2b",
  "demo_order_id": "7a3e2b1c-9d8f-4e5a-8b2c-1f4e6d8a9b0c",
  "demo_account_id": "4d2c1b0a-8e7f-4a3b-9c1d-5e2f3a4b5c6d",
  "symbol": "BTC-USDT",
  "side": "BUY",
  "fill_price": "65240.250000",
  "fill_quantity": "0.50000000",
  "quote_amount": "32620.125000",
  "virtual_fee": "3.262013",
  "slippage_bps": "2.4500",
  "reference_mid_price": "65224.300000",
  "testnet_tx_hash": "0x8f3c4e2b1a9d0f7e6c5b4a3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f",
  "executed_at_unix_ms": 1726738081000
}
```

---

## Security & Compliance Notes

- **Absolute Zero State Bleed & Complete Isolation:**
  - *Logical Isolation:* Demo accounts, demo orders, and demo positions are partitioned into dedicated database tables and dedicated Redis instances. No demo order or transaction can ever cross-reference real-money production order books.
  - *Cryptographic Separation:* User authentication tokens for demo mode contain explicit JWT scope claims (`mode: "DEMO"`, `sub_type: "DEMO_TRADER"`). The API Gateway (Prompt 219) and microservice layers strictly reject demo JWT tokens when presented to production trading endpoints, and vice versa.
  - *Network and Blockchain Isolation:* Testnet relayer clients are physically restricted via Kubernetes NetworkPolicies from communicating with the Hyperledger Besu Mainnet RPC nodes. Chain ID checks are enforced at runtime (`assert!(chain_id == 1338)`).
- **Statutory Regulatory Disclaimers & Fair Presentation (SEBI / IFSCA):**
  - In compliance with SEBI and IFSCA investor protection guidelines, all responses, WebSocket feeds, and UI components must clearly present non-removable disclaimers:
    > "SIMULATED TRADING ENVIRONMENT: All assets, orders, and balances displayed are virtual and carry zero real-world economic value. Virtual profits cannot be redeemed, withdrawn, or transferred."
  - Prohibits gamified predatory mechanics (such as offering bonuses that trigger real-money deposits based on simulated paper performance) to comply with investor protection standards.
- **Idempotency & Double-Execution Prevention:**
  - All virtual order actions (`PlaceDemoOrder`, `CancelDemoOrder`, `ReplaceDemoOrder`) require client-generated UUIDv4 idempotency keys stored in Redis with a 24-hour TTL.
  - Network retries or duplicate client requests return the existing execution result without re-executing fills or creating redundant orders.
- **Abuse Prevention & Rate Limiting:**
  - Demo endpoints are protected by token-bucket rate limiters in the API Gateway (Prompt 220): maximum 20 orders/second per demo account, preventing denial-of-service degradation of the shared simulation infrastructure.
- **Deterministic Server-Authoritative Clock:**
  - All queue positions, order expirations, and fill evaluations rely on microsecond-precision NTP-synchronized server clocks, preventing client-side clock tampering or latency manipulation.

---

## Acceptance Criteria

- [ ] High-performance virtual execution engine achieves an internal matching and fill simulation latency $\le 2\text{ms}$ ($p99$) under normal load.
- [ ] In-memory synthetic order book dynamically updates against live market feeds from Live Feeder (Prompt 272) at rates exceeding 20,000 tick updates per second without memory leaks.
- [ ] Limit order queue simulation strictly enforces volume priority: virtual limit orders are not filled until actual market volume traded at that price level exceeds the volume ahead in the queue.
- [ ] Market order execution applies realistic dynamic slippage: test suite verifies that slippage scales monotonically with order size relative to synthetic book depth.
- [ ] Stop-Loss, Stop-Limit, and Take-Profit orders trigger deterministically within $\le 5\text{ms}$ of mark price crossing the configured trigger threshold.
- [ ] Virtual balance reservations and double-entry settlements with Demo Wallet Service (Prompt 274) execute with zero orphan holds or ledger discrepancies.
- [ ] Dual-environment isolation test: passing a demo JWT token to production `OrderService` (Prompt 204) results in HTTP 403 Forbidden; passing a production JWT token to `DemoOrderService` is strictly rejected.
- [ ] Simulated trade receipts are dispatched to `DemoSettlementDvP.sol` on Hyperledger Besu Testnet (Chain ID 1338); any connection attempt to Mainnet (Chain ID 1337) panics and fails fast.
- [ ] Resetting a demo account cleanly cancels all active orders, clears resting queue allocations, and restores virtual USDT balance to exactly 100,000.000000.
- [ ] All PostgreSQL schema DDL scripts compile and execute idempotently against PostgreSQL 16+.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: gRPC and Protobuf service standards.
  - `104_event_schema_and_kafka_topic_standards.md`: Event definitions and topic naming conventions.
  - `111_domain_model_core_entities.md`: Core trading pair, asset, and order entity models.
  - `112_idempotency_and_exactly_once_processing.md`: Idempotency key specifications.
  - `207_market_data_service.md`: Market ticker snapshots and candlestick chart data.
  - `272_live_market_data_feeder.md`: Ultra-low-latency real-time external crypto market data feed.
- **Parallel Tasks:**
  - `274_demo_wallet_and_virtual_faucet_service.md`: Virtual ledger, faucet replenishments, and balance holds.
  - `306_settlement_dvp_smart_contract.md`: Reference architecture for Delivery-versus-Payment smart contracts adapted for Testnet (`DemoSettlementDvP.sol`).
- **Downstream Blockers:**
  - `509_flutter_order_placement_and_trading_interface.md`: Integration of the Flutter mobile paper trading interface.
  - `527_flutter_environment_switcher_and_sandbox_mode.md`: Sandbox environment switcher toggle between Live and Demo modes.
  - `603_web_trading_dashboard_and_order_entry.md`: Web paper trading dashboard, depth chart, and simulated trade stream.
