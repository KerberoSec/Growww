# 265 - Copy Trading & Proportional Replication Service (Rust / Go)

## Purpose
The **Copy Trading & Proportional Replication Service** is an ultra-low-latency algorithmic mirror execution and asset allocation engine designed for tokenized financial markets. It empowers retail and institutional investors (followers) to mirror the real-time trading actions of verified lead traders (masters) with mathematical fidelity, proportional position sizing, and millisecond-level execution dispatching.

In fast-moving equity and digital security token markets, manual trade replication introduces fatal execution lag, adverse selection, and severe price slippage. This service listens to master execution fills, calculates scaled order quantities based on each follower's dedicated equity ratio, and fans out child orders to the matching infrastructure within strict sub-15ms latency budgets. 

The service enforces a robust investor protection architecture:
- Proportional order allocation supporting fractional equities up to 6 decimal places.
- Strict slippage barriers capping follower execution deviation to a maximum of +/- 0.5% relative to the master fill price.
- Dynamic follower emergency detachment circuit breakers based on maximum drawdown limits, account health, and strategy halts (ADR-0042, RUNBOOK-30).
- A mathematically rigorous 10% to 15% High-Water Mark (HWM) profit-sharing settlement engine that calculates performance fees exclusively on net new profits, orchestrating weekly settlements through off-chain ledger transfers and on-chain permissioned smart contracts.

## What You Are Building
A standalone, high-concurrency, dual-language microservice (`services/copy-trading-service`) engineered in Rust and Go. Deliverables include:

- **Master Execution Listener (Go / Kafka):** A high-throughput consumer subscribing to `engine.matches.v1` and `master.executions.v1` that ingests master fill events, validates strategy authorization, and pushes events into an in-memory lock-free dispatch ring.
- **Proportional Sizing Calculator (Rust Core):** A vectorized mathematical calculation module computing child order sizes across thousands of subscribed followers using real-time equity ratios ($Q_{\text{follower}} = Q_{\text{master}} \times \frac{\text{Equity}_{\text{follower}}}{\text{Equity}_{\text{master}}}$), lot-size quantization, and minimum ticket thresholds.
- **Sub-15ms Child Order Fanout Dispatcher (Rust Async):** A low-latency dispatcher leveraging asynchronous worker pools and connection multiplexing to validate pre-trade buying power and route parallel child orders to the Order Service (Prompt 204) and Order Matching Engine (Prompt 205) within 15 milliseconds of the master fill.
- **Slippage & Market Protection Guard:** A real-time pricing filter that checks current top-of-book market depth before child order submission; if market impact or spread divergence exceeds +/- 0.5% (50 bps) of the master execution price, child orders are safely transitioned to limit orders or aborted to shield followers.
- **Follower Emergency Detachment Circuit Breaker (ADR-0042, RUNBOOK-30):** An automated safety monitor evaluating real-time follower equity against peak allocation values. If portfolio drawdown breaches user-configured thresholds (e.g., 5%, 10%, 15%), the engine detaches the follower instantly, cancels pending child orders, and optionally unwinds open positions.
- **Weekly High-Water Mark (HWM) Profit-Share Settlement Engine (Go):** An automated weekly batch reconciliation daemon calculating net realized and unrealized gains across follower accounts. It enforces the classic High-Water Mark principle, computes the master's 10% to 15% performance fee cut, deducts applicable GST, and triggers settlement via the Fee Engine (Prompt 210) and Wallet Service (Prompt 203).
- **On-Chain Settlement Bridge:** An automated bridge emitting cryptographically signed performance fee distribution receipts to the `PerformanceFeeDistributor.sol` smart contract on Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
  - Master trader registration, strategy provisioning, and public performance indexing.
  - Follower subscription management, capital allocation limits, and replication mode configuration (Equity-Proportional, Fixed-Ratio, or Cash-Proportional).
  - Ingestion of master order fills and sub-15ms fanout calculation for up to 10,000 followers per master strategy.
  - Real-time pre-trade slippage validation (+/- 0.5% maximum allowable price dispersion).
  - Automated follower detachment circuit breakers triggered by maximum drawdown, master strategy suspension, or follower cash deficiency.
  - Weekly High-Water Mark profit-share calculation (10% to 15% configurable rate) with carry-forward loss recovery tracking.
  - Emission of child order intents to Order Management & Lifecycle Service (Prompt 204).
  - Dispatch of on-chain performance fee settlement attestations to Hyperledger Besu.
- **Out of Scope / Handled Elsewhere:**
  - Primary order matching algorithms and central limit order book state (handled in Prompt 205).
  - Direct custodial cash debits, bank payments, and double-entry ledger accounts (handled in Prompt 203 and Prompt 212).
  - Exchange-wide margin calculation, VaR, and initial margin haircut evaluation (handled in Prompt 206 and Prompt 229).
  - Social network feeds, community chat, user commentary, and unverified tipping systems.
  - Physical depository demat allocation with CDSL/NSDL (handled in Prompt 213).

## Technology to Use
- **Core Languages:**
  - **Rust 1.78+:** Core replication calculation engine, vectorized proportional sizing math, lock-free ring buffers, and the high-speed child order fanout dispatcher.
  - **Go 1.22+:** gRPC control plane APIs, Kafka event subscription coordinators, weekly HWM profit-share batch orchestrator, and external service connectors.
- **In-Memory Cache & State Store:** **Redis 7.2+ Cluster** utilizing Redis Hashes for active follower balances, Redis Sets for strategy-to-follower indexing, and Redis Streams for ultra-fast inter-thread dispatch pipelines.
- **Event Streaming & Messaging:** **Apache Kafka** (`segmentio/kafka-go` and `rdkafka-rust`) consuming from `engine.matches.v1` and publishing child orders to `copytrading.child_orders.v1` and emergency alerts to `copytrading.circuit_breaker.v1`.
- **Database & Persistence:** **PostgreSQL 16+** with `pgx/v5` and `sqlc` for Go / `sqlx` for Rust, storing strategy configurations, follower subscription states, execution audit logs, and weekly HWM settlement ledgers.
- **Inter-Service Communication:** **gRPC (HTTP/2 with Protobuf)** for synchronous balance checks and child order submission to Order Service (Prompt 204) and Risk Service (Prompt 206).
- **Concurrency & Concurrency Primitives:** Rust `tokio` multi-threaded runtime, crossbeam channels, and atomic bitmasks for tracking follower detachment flags without mutex contention.

## Backend / Infra Touchpoints
- **Order Matching Engine (Prompt 205):** Consumes real-time master fill events from `engine.matches.v1` with millisecond timestamps and microsecond match sequence numbers.
- **Order Service (Prompt 204):** Submits batched child replication orders via high-performance gRPC streaming or direct Kafka injection into `order.matching.commands.v1`.
- **Risk & Margin Checks Service (Prompt 206):** Ingests pre-trade risk validations to confirm that follower accounts have sufficient margin, leverage room, and unencumbered cash before dispatching child orders.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes trade execution costs, exchange transaction charges, and weekly 10% to 15% HWM performance fee splits with GST accounting.
- **Wallet & Account Service (Prompt 203):** Coordinates fund reservations and atomic cash balance transfers for weekly profit-share payouts from followers to master traders.
- **Portfolio & Holdings Service (Prompt 209):** Queries real-time follower equity, cash balances, and tokenized share quantities to compute live NAV and drawdown percentages.
- **Notification Service (Prompt 211):** Emits critical notifications (follower detachment, circuit breaker activation, slippage aborts, and weekly HWM settlement statements).

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Smart Contract Target:** `PerformanceFeeDistributor.sol` deployed on the permissioned Hyperledger Besu network under QBFT consensus (2-second finality).
- **On-Chain Profit-Share Distribution:** At the conclusion of each weekly HWM settlement cycle, the service publishes an aggregated, cryptographically signed settlement batch transaction. The smart contract validates the master's public key signature and transfers tokenized fee shares from the follower's on-chain escrow balance to the master trader's vault address.
- **Cryptographic State Attestation:** A Merkle root summarizing follower weekly PnL snapshots, fee deductions, and High-Water Mark thresholds is committed to the blockchain, creating an immutable audit trail for regulatory scrutiny.
- **Zero On-Chain PII Guarantee:** Blockchain interactions use strictly pseudonymous identifiers: `master_strategy_id` (UUIDv4 hash), follower `account_hash` (keccak256 of user ID and salt), and token smart contract addresses (`0x...`). User names, PAN numbers, contact details, and bank coordinates are strictly quarantined in off-chain encrypted vaults.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Directory & Build Configuration:** Create `services/copy-trading-service` with a dual-module workspace: a high-performance Rust crate (`crates/replication-core`) for sizing and fanout, and a Go application (`cmd/copy-trading-daemon`) for gRPC APIs, Kafka consumers, and settlement scheduling.
2. **Define Protobuf Contracts:** Create `proto/growww/copytrading/v1/copy_trading_service.proto` defining endpoints for strategy registration, follower subscriptions, allocation updates, detachment overrides, and fee settlement status.
3. **Generate Language Stubs:** Compile Protobuf contracts into Go stubs (`protoc-gen-go`, `protoc-gen-go-grpc`) and Rust stubs (`prost`, `tonic-build`).
4. **Author PostgreSQL DDL Migrations:** Implement database migration scripts for `master_strategies`, `follower_subscriptions`, `follower_allocations`, `replicated_orders`, `hwm_snapshots`, `profit_share_settlements`, and `circuit_breaker_events`.
5. **Configure Redis Cache Schemas:** Implement caching schemas in Redis 7.2 Cluster:
   - Hash `copytrading:master:{strategy_id}:meta` for strategy limits and status.
   - Set `copytrading:strategy:{strategy_id}:followers` for fast follower discovery.
   - Hash `copytrading:follower:{subscription_id}:balance` for real-time allocated equity and cash balance.
   - Bitfield / Set `copytrading:circuit_breaker:detached` for sub-millisecond detachment state lookups.
6. **Implement Master Execution Consumer:** Build Kafka consumer in Go subscribing to `engine.matches.v1`. When a fill matches a registered master account ID, extract ISIN, side, price, filled quantity, and execution timestamp, forwarding the event into the Rust replication core via shared memory or low-overhead Unix domain socket.
7. **Build Proportional Position Sizing Core (Rust):** Implement the vectorized calculation algorithm:
   - Query master total allocated equity ($E_m$) and follower allocated equity ($E_f$).
   - Compute base child quantity: $Q_f = Q_m \times \frac{E_f}{E_m}$.
   - Round $Q_f$ according to the symbol's tick size and lot constraints (up to 6 decimal places for fractional tokens).
   - If $Q_f \times \text{Price} < \text{MinTicketSize}$, drop child order to prevent dust positions.
8. **Implement Slippage Protection Filter:** Prior to child order dispatch, query Redis top-of-book cache for the symbol. Calculate price deviation:
   $$\Delta P = \frac{|P_{\text{current}} - P_{\text{master\_fill}}|}{P_{\text{master\_fill}}}$$
   If $\Delta P > 0.005$ (0.5% or 50 bps), reject the child replication order or convert it into a passive limit order at $P_{\text{master\_fill}} \times (1 \pm 0.005)$ based on user strategy preference.
9. **Implement Sub-15ms Child Order Fanout Dispatcher:** In Rust, utilize `tokio::spawn` with a bounded worker pool and pre-warmed gRPC connection channels to dispatch child orders in parallel batches to Order Service (Prompt 204). Complete end-to-end fanout (from master fill ingestion to child order submission) within 15 milliseconds for up to 5,000 concurrent followers.
10. **Implement Follower Emergency Detachment Circuit Breaker (ADR-0042, RUNBOOK-30):**
    - Continuously evaluate follower equity against current High-Water Mark.
    - Calculate current drawdown: $DD = \frac{\text{HWM} - \text{Equity}_{\text{current}}}{\text{HWM}}$.
    - If $DD \ge \text{MaxDrawdownLimit}$, atomically flag the subscription as `CIRCUIT_BREAKER_DETACHED` in Redis and PostgreSQL.
    - Cancel all open child orders for that follower and emit an event to `copytrading.circuit_breaker.v1`.
    - If configured by follower preferences, dispatch market exit orders to flatten open replicated positions.
11. **Implement High-Water Mark (HWM) Accounting Engine:** Implement daily and weekly equity tracking that maintains the historical peak portfolio valuation for each follower-strategy pair:
    - At weekly cut-off (Friday 23:59:59 IST), record Current NAV ($V_{\text{current}}$).
    - If $V_{\text{current}} > \text{HWM}_{\text{prev}}$, compute Net New Profit: $P_{\text{net}} = V_{\text{current}} - \text{HWM}_{\text{prev}}$.
    - Compute performance fee: $\text{Fee} = P_{\text{net}} \times r_{\text{perf}}$ (where $r_{\text{perf}} \in [0.10, 0.15]$).
    - Update $\text{HWM}_{\text{new}} = V_{\text{current}} - \text{Fee}$.
    - If $V_{\text{current}} \le \text{HWM}_{\text{prev}}$, fee is zero; carry forward deficit to subsequent cycles.
12. **Build Weekly Settlement Orchestrator:** Develop Go batch worker that iterates over active subscriptions, reserves performance fees from follower cash accounts via Wallet Service (Prompt 203), deducts applicable statutory GST (18%) via Fee Engine (Prompt 210), and credits the net performance fee to the master trader's wallet.
13. **Implement On-Chain Settlement Bridge:** Author Go integration client interfacing with `PerformanceFeeDistributor.sol` via JSON-RPC/EVM client. Publish batch Merkle root and execute token distribution transactions on Hyperledger Besu with QBFT finality.
14. **Implement Management gRPC Services:** Expose gRPC APIs for creating master strategies, subscribing/unsubscribing followers, adjusting allocation capital, modifying drawdown limits, and retrieving real-time replication status.
15. **Integrate Observability, Metrics, and Tracing:** Configure Prometheus metrics (`copytrading_fanout_duration_seconds`, `copytrading_child_orders_total`, `copytrading_slippage_aborts_total`, `copytrading_circuit_breaker_trips_total`) and export OpenTelemetry distributed traces for every master execution pipeline.

## Interfaces / Contracts

### Protobuf Definition (`copy_trading_service.proto`)
```protobuf
syntax = "proto3";

package growww.copytrading.v1;

option go_package = "growww/copytrading/v1;copytradingv1";

service CopyTradingService {
  // Strategy Management
  rpc CreateMasterStrategy (CreateMasterStrategyRequest) returns (CreateMasterStrategyResponse);
  rpc UpdateMasterStrategy (UpdateMasterStrategyRequest) returns (UpdateMasterStrategyResponse);
  rpc GetMasterStrategy (GetMasterStrategyRequest) returns (GetMasterStrategyResponse);
  rpc ListMasterStrategies (ListMasterStrategiesRequest) returns (ListMasterStrategiesResponse);

  // Follower Subscription & Allocation
  rpc SubscribeStrategy (SubscribeStrategyRequest) returns (SubscribeStrategyResponse);
  rpc UnsubscribeStrategy (UnsubscribeStrategyRequest) returns (UnsubscribeStrategyResponse);
  rpc UpdateAllocation (UpdateAllocationRequest) returns (UpdateAllocationResponse);
  rpc GetSubscriptionStatus (GetSubscriptionStatusRequest) returns (GetSubscriptionStatusResponse);

  // Circuit Breaker & Detachment Management
  rpc TriggerManualDetachment (TriggerManualDetachmentRequest) returns (TriggerManualDetachmentResponse);
  rpc ResumeSubscription (ResumeSubscriptionRequest) returns (ResumeSubscriptionResponse);

  // High-Water Mark & Settlement Queries
  rpc GetHwmSnapshot (GetHwmSnapshotRequest) returns (GetHwmSnapshotResponse);
  rpc ListSettlementHistory (ListSettlementHistoryRequest) returns (ListSettlementHistoryResponse);
}

enum ReplicationMode {
  REPLICATION_MODE_UNSPECIFIED = 0;
  REPLICATION_MODE_EQUITY_PROPORTIONAL = 1;
  REPLICATION_MODE_FIXED_RATIO = 2;
  REPLICATION_MODE_CASH_PROPORTIONAL = 3;
}

enum SubscriptionStatus {
  SUBSCRIPTION_STATUS_UNSPECIFIED = 0;
  SUBSCRIPTION_STATUS_ACTIVE = 1;
  SUBSCRIPTION_STATUS_PAUSED = 2;
  SUBSCRIPTION_STATUS_CIRCUIT_BREAKER_DETACHED = 3;
  SUBSCRIPTION_STATUS_CLOSED = 4;
}

enum DetachmentAction {
  DETACHMENT_ACTION_UNSPECIFIED = 0;
  DETACHMENT_ACTION_KEEP_POSITIONS = 1;
  DETACHMENT_ACTION_MARKET_UNWIND = 2;
}

message CreateMasterStrategyRequest {
  string master_user_id = 1;
  string strategy_name = 2;
  string description = 3;
  string profit_share_percentage = 4; // Range: "10.00" to "15.00"
  string max_followers = 5;
  string min_capital_inr = 6;
}

message CreateMasterStrategyResponse {
  string strategy_id = 1;
  string status = 2;
  int64 created_at_unix = 3;
}

message UpdateMasterStrategyRequest {
  string strategy_id = 1;
  string master_user_id = 2;
  bool is_active = 3;
  string description = 4;
}

message UpdateMasterStrategyResponse {
  string strategy_id = 1;
  bool success = 2;
  int64 updated_at_unix = 3;
}

message GetMasterStrategyRequest {
  string strategy_id = 1;
}

message GetMasterStrategyResponse {
  string strategy_id = 1;
  string master_user_id = 2;
  string strategy_name = 3;
  string description = 4;
  string profit_share_percentage = 5;
  int32 active_followers = 6;
  string total_allocated_equity = 7;
  string historical_roi_percentage = 8;
  string max_historical_drawdown = 9;
  bool is_active = 10;
}

message ListMasterStrategiesRequest {
  int32 page_size = 1;
  string page_token = 2;
  bool active_only = 3;
}

message ListMasterStrategiesResponse {
  repeated GetMasterStrategyResponse strategies = 1;
  string next_page_token = 2;
}

message SubscribeStrategyRequest {
  string follower_user_id = 1;
  string strategy_id = 2;
  ReplicationMode replication_mode = 3;
  string allocated_capital_inr = 4;
  string max_drawdown_percentage = 5; // e.g., "10.00" for 10%
  DetachmentAction detachment_action = 6;
  string max_slippage_tolerance = 7;  // Default: "0.0050" (0.5%)
}

message SubscribeStrategyResponse {
  string subscription_id = 1;
  SubscriptionStatus status = 2;
  string initial_hwm = 3;
  int64 subscribed_at_unix = 4;
}

message UnsubscribeStrategyRequest {
  string subscription_id = 1;
  string follower_user_id = 2;
  DetachmentAction action = 3;
}

message UnsubscribeStrategyResponse {
  string subscription_id = 1;
  SubscriptionStatus status = 2;
  int64 terminated_at_unix = 3;
}

message UpdateAllocationRequest {
  string subscription_id = 1;
  string follower_user_id = 2;
  string new_allocated_capital_inr = 3;
}

message UpdateAllocationResponse {
  string subscription_id = 1;
  string updated_allocated_capital_inr = 2;
  int64 updated_at_unix = 3;
}

message GetSubscriptionStatusRequest {
  string subscription_id = 1;
}

message GetSubscriptionStatusResponse {
  string subscription_id = 1;
  string follower_user_id = 2;
  string strategy_id = 3;
  SubscriptionStatus status = 4;
  string allocated_capital = 5;
  string current_nav = 6;
  string current_hwm = 7;
  string current_drawdown_percentage = 8;
  int64 last_replication_unix = 9;
}

message TriggerManualDetachmentRequest {
  string subscription_id = 1;
  string reason = 2;
  DetachmentAction action = 3;
}

message TriggerManualDetachmentResponse {
  string subscription_id = 1;
  SubscriptionStatus status = 2;
  int64 detached_at_unix = 3;
}

message ResumeSubscriptionRequest {
  string subscription_id = 1;
  string follower_user_id = 2;
  string new_allocated_capital_inr = 3;
}

message ResumeSubscriptionResponse {
  string subscription_id = 1;
  SubscriptionStatus status = 2;
  string reset_hwm = 3;
  int64 resumed_at_unix = 4;
}

message GetHwmSnapshotRequest {
  string subscription_id = 1;
}

message GetHwmSnapshotResponse {
  string subscription_id = 1;
  string current_hwm = 2;
  string peak_nav = 3;
  string cumulative_profit_shared = 4;
  int64 last_settled_at_unix = 5;
}

message ListSettlementHistoryRequest {
  string subscription_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message SettlementRecord {
  string settlement_id = 1;
  string cycle_start_unix = 2;
  string cycle_end_unix = 3;
  string pre_settlement_hwm = 4;
  string post_settlement_hwm = 5;
  string net_new_profit = 6;
  string profit_share_amount = 7;
  string gst_amount = 8;
  string on_chain_tx_hash = 9;
}

message ListSettlementHistoryResponse {
  repeated SettlementRecord records = 1;
  string next_page_token = 2;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Custom Types & Enums
CREATE TYPE copytrading_replication_mode AS ENUM (
    'EQUITY_PROPORTIONAL',
    'FIXED_RATIO',
    'CASH_PROPORTIONAL'
);

CREATE TYPE copytrading_subscription_status AS ENUM (
    'ACTIVE',
    'PAUSED',
    'CIRCUIT_BREAKER_DETACHED',
    'CLOSED'
);

CREATE TYPE copytrading_detachment_action AS ENUM (
    'KEEP_POSITIONS',
    'MARKET_UNWIND'
);

CREATE TYPE child_order_status AS ENUM (
    'PENDING_DISPATCH',
    'DISPATCHED',
    'FILLED',
    'PARTIALLY_FILLED',
    'SLIPPAGE_ABORTED',
    'RISK_REJECTED',
    'FAILED'
);

-- Master Strategy Registry
CREATE TABLE master_strategies (
    strategy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    master_user_id UUID NOT NULL,
    strategy_name VARCHAR(128) NOT NULL,
    description TEXT,
    profit_share_percentage NUMERIC(5, 2) NOT NULL CHECK (profit_share_percentage >= 10.00 AND profit_share_percentage <= 15.00),
    max_followers INT NOT NULL DEFAULT 5000,
    min_capital_inr NUMERIC(18, 4) NOT NULL DEFAULT 10000.0000,
    current_allocated_equity NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Follower Subscriptions
CREATE TABLE follower_subscriptions (
    subscription_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_user_id UUID NOT NULL,
    strategy_id UUID NOT NULL REFERENCES master_strategies(strategy_id) ON DELETE RESTRICT,
    replication_mode copytrading_replication_mode NOT NULL DEFAULT 'EQUITY_PROPORTIONAL',
    allocated_capital_inr NUMERIC(18, 4) NOT NULL CHECK (allocated_capital_inr > 0),
    max_drawdown_percentage NUMERIC(5, 2) NOT NULL DEFAULT 10.00 CHECK (max_drawdown_percentage > 0 AND max_drawdown_percentage <= 50.00),
    max_slippage_tolerance NUMERIC(5, 4) NOT NULL DEFAULT 0.0050 CHECK (max_slippage_tolerance > 0 AND max_slippage_tolerance <= 0.0200),
    detachment_action copytrading_detachment_action NOT NULL DEFAULT 'KEEP_POSITIONS',
    status copytrading_subscription_status NOT NULL DEFAULT 'ACTIVE',
    subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(follower_user_id, strategy_id)
);

-- High-Water Mark (HWM) State Snapshots
CREATE TABLE hwm_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES follower_subscriptions(subscription_id) ON DELETE CASCADE,
    high_water_mark NUMERIC(18, 4) NOT NULL,
    peak_nav NUMERIC(18, 4) NOT NULL,
    current_nav NUMERIC(18, 4) NOT NULL,
    cumulative_profit_shared NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    last_valuation_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Replicated Child Order Executions
CREATE TABLE replicated_orders (
    replicated_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES follower_subscriptions(subscription_id) ON DELETE CASCADE,
    master_execution_id UUID NOT NULL,
    child_order_id UUID,
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    master_fill_price NUMERIC(18, 4) NOT NULL,
    master_fill_quantity NUMERIC(18, 6) NOT NULL,
    target_child_quantity NUMERIC(18, 6) NOT NULL,
    executed_child_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    executed_child_price NUMERIC(18, 4),
    slippage_bps NUMERIC(8, 2),
    status child_order_status NOT NULL DEFAULT 'PENDING_DISPATCH',
    rejection_reason TEXT,
    fanout_latency_ms NUMERIC(6, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Weekly Profit-Share Settlements
CREATE TABLE profit_share_settlements (
    settlement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES follower_subscriptions(subscription_id) ON DELETE RESTRICT,
    cycle_start_time TIMESTAMPTZ NOT NULL,
    cycle_end_time TIMESTAMPTZ NOT NULL,
    pre_settlement_hwm NUMERIC(18, 4) NOT NULL,
    closing_nav NUMERIC(18, 4) NOT NULL,
    net_new_profit NUMERIC(18, 4) NOT NULL,
    profit_share_rate NUMERIC(5, 2) NOT NULL,
    gross_profit_share NUMERIC(18, 4) NOT NULL,
    gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 18.00,
    gst_deducted NUMERIC(18, 4) NOT NULL,
    net_master_payout NUMERIC(18, 4) NOT NULL,
    post_settlement_hwm NUMERIC(18, 4) NOT NULL,
    on_chain_tx_hash VARCHAR(66),
    settled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Circuit Breaker & Detachment Event Logs (ADR-0042, RUNBOOK-30)
CREATE TABLE circuit_breaker_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES follower_subscriptions(subscription_id) ON DELETE CASCADE,
    trigger_type VARCHAR(32) NOT NULL, -- 'DRAWDOWN_BREACH', 'MANUAL_OVERRIDE', 'STRATEGY_HALTED', 'BALANCE_DEPLETED'
    hwm_at_trigger NUMERIC(18, 4) NOT NULL,
    nav_at_trigger NUMERIC(18, 4) NOT NULL,
    drawdown_percentage NUMERIC(5, 2) NOT NULL,
    threshold_percentage NUMERIC(5, 2) NOT NULL,
    action_executed copytrading_detachment_action NOT NULL,
    unwind_order_count INT NOT NULL DEFAULT 0,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance Indexes
CREATE INDEX idx_master_strategies_user ON master_strategies(master_user_id, is_active);
CREATE INDEX idx_follower_subscriptions_strategy ON follower_subscriptions(strategy_id, status);
CREATE INDEX idx_follower_subscriptions_user ON follower_subscriptions(follower_user_id, status);
CREATE INDEX idx_replicated_orders_sub ON replicated_orders(subscription_id, created_at DESC);
CREATE INDEX idx_replicated_orders_master_exec ON replicated_orders(master_execution_id);
CREATE INDEX idx_hwm_snapshots_sub ON hwm_snapshots(subscription_id);
CREATE INDEX idx_profit_share_cycle ON profit_share_settlements(subscription_id, cycle_end_time DESC);
CREATE INDEX idx_circuit_breaker_sub ON circuit_breaker_events(subscription_id, triggered_at DESC);
```

## Security & Compliance Notes
- **SEBI Non-Advisory Compliance:** In strict adherence to SEBI regulations governing social trading, investment advisory, and algorithmic trading, the platform operates strictly as an execution-only technological infrastructure. Master traders do not hold discretionary power of attorney over follower bank accounts. Followers enter into an explicit electronic Algorithmic Replication Mandate granting rule-based software execution permissions. No guarantees of returns are made, and standardized risk disclosures (Prompt 009) are acknowledged electronically before subscription.
- **Strict Slippage Guardrails (+/- 0.5% Maximum):** Rapid market volatility or wide bid-ask spreads can disadvantage followers. Child replication orders are automatically rejected or converted to price-capped limit orders if top-of-book depth deviates by more than +/- 0.5% (50 bps) from the master execution fill price.
- **Follower Maximum Drawdown Detachment Barriers:** To protect retail capital from catastrophic strategy decay, the system continuously monitors equity. If the follower's drawdown against their personal High-Water Mark exceeds their configured tolerance (default 10%, maximum allowable 50%), the emergency circuit breaker fires immediately, detaching the account and preventing further order replication.
- **Fail-Safe Circuit Breakers (ADR-0042, RUNBOOK-30):** If a master trader executes erratic trading patterns (e.g., submitting more than 50 orders per second or exceeding portfolio position limits), the system activates ADR-0042 safety halts, pausing replication across all attached followers in less than 5 milliseconds.
- **High-Water Mark (HWM) Auditability:** Performance fees (10% to 15%) are assessed strictly on cumulative net new profits. If a strategy experiences losses, no performance fee is assessed until past losses are fully recovered and the previous peak valuation is exceeded. Historical HWM state is immutable and auditable via cryptographic Merkle roots.
- **Zero On-Chain PII Guarantee:** Under the Digital Personal Data Protection Act (DPDPA 2023) and SEBI guidelines, all data broadcast to Hyperledger Besu or Kafka public feeds uses cryptographically hashed account identifiers. No investor names, PAN numbers, or IP addresses are ever published to the ledger.

## Acceptance Criteria
- [ ] Standalone service compiles in Rust (`crates/replication-core`) and Go (`cmd/copy-trading-daemon`) without compiler warnings or unresolved dependencies.
- [ ] End-to-end child order fanout latency from master fill ingestion to child order gRPC submission is verified at $< 15\text{ms}$ (p99) for 5,000 active followers under load.
- [ ] Proportional position sizing calculation accurately accounts for fractional share lots up to 6 decimal places without floating-point drift or roundoff divergence.
- [ ] Slippage filter reliably detects market price divergence $> 0.5\%$ (50 bps) against the master execution price and safely aborts or caps child orders.
- [ ] Emergency detachment circuit breaker halts order replication within 50ms of a maximum drawdown breach and logs an auditable event to `circuit_breaker_events`.
- [ ] High-Water Mark calculation correctly handles loss-recovery cycles, charging zero performance fees until past peak NAV is exceeded.
- [ ] Weekly settlement engine computes statutory 18% GST deductions on performance fees and coordinates atomic wallet balance transfers without ledger imbalances.
- [ ] Cryptographic batch settlement root is posted to `PerformanceFeeDistributor.sol` on Hyperledger Besu with zero on-chain PII.
- [ ] Child orders utilize unique idempotency keys composed of `subscription_id + master_execution_id` to guarantee exactly-once execution.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 103 (API Design Standards)
  - Prompt 104 (Event Schema & Kafka Standards)
  - Prompt 111 (Domain Model Core Entities)
  - Prompt 112 (Idempotency & Exactly-Once Processing)
  - Prompt 203 (Wallet & Account Service)
  - Prompt 204 (Order Management & Lifecycle Service)
  - Prompt 205 (Order Matching Engine)
  - Prompt 206 (Risk & Margin Checks Service)
- **Parallel Tasks:**
  - Prompt 209 (Portfolio & Holdings Service)
  - Prompt 210 (Fee & Realized PnL Engine)
  - Prompt 211 (Notification Service)
  - Prompt 262 (Automated Liquidity Provisioning & Market Maker Engine)
- **Downstream Blockers:**
  - Prompt 509 (Mobile / Flutter Copy Trading & Strategy Hub UI)
  - Prompt 603 (Web Trading Dashboard Copy Portfolio Management)
  - Prompt 315 (Settlement Guarantee Fund & Default Waterfall Architecture)
