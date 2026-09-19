# 226 - Advanced Order Types & Algorithmic Trigger Engine (Go)

## Purpose
Operating a continuous 24/7 equity exchange requires sophisticated execution tooling beyond basic Limit and Market orders. Institutional traders, algorithmic desks, and retail investors need algorithmic order types to minimize market impact, protect against overnight volatility, and automate complex execution strategies without continuous manual intervention.

The Advanced Order Types & Algorithmic Trigger Engine acts as an off-book synthetic order orchestrator. It manages parent order lifecycles, tracks real-time high/low market watermarks, monitors the National Best Bid and Offer (NBBO) spread, and dynamically injects child slice orders into the core Order Matching Engine (Prompt 205) when configured market conditions or replenishment thresholds are met.

Supported advanced execution algorithms include:
1. **Iceberg Orders:** Large parent orders divided into visible "tip" quantities and hidden reserves, automatically reloading upon execution with randomized display variance.
2. **Trailing-Stop Orders:** Dynamic stop-loss and take-profit triggers that automatically ratchet upward (for Long positions) or downward (for Short positions) alongside favorable price movements.
3. **Pegged Orders:** Orders dynamically pegged to the Best Bid (Primary Peg), Best Ask (Market Peg), or Midpoint (Midpoint Peg) with customizable tick offsets.
4. **Bracket Orders & OCO (One-Cancels-the-Other):** Paired profit-target and stop-loss orders where execution of one leg immediately cancels the contingent leg.
5. **TWAP (Time-Weighted Average Price) Slices:** Deterministic periodic order slicing across pre-defined time intervals to achieve smooth volume-weighted fills over 24/7 trading sessions.

## What You Are Building
A high-throughput, fault-tolerant Go microservice (`services/algorithmic-orders`). Concrete deliverables include:
- **Synthetic Order State Engine:** Distributed order lifecycle manager tracking parent order status, child slice allocations, cumulative fills, and execution schedules.
- **Real-Time Market Tick Listener:** High-frequency event consumer listening to live trades and top-of-book quotes (`engine.matches.v1`, `matching.depth.v1`) to evaluate trigger criteria.
- **Iceberg Replenishment Controller:** Real-time watcher that detects partial/full fills of visible child tips and immediately synthesizes next child slice with optional randomized lot variance (+/- 10-20%) to prevent algorithmic reverse-engineering.
- **Trailing Watermark Tracker:** Low-latency in-memory state engine tracking high-water marks (for sell trailing stops) and low-water marks (for buy trailing stops) in Redis and local memory.
- **Pegged Order Re-pricing Engine:** Dynamic collar watcher updating resting pegged child order prices when the lit book NBBO shifts by more than $N$ ticks.
- **Pre-Trade Reservation Allocator:** Interfaces with Wallet Service (Prompt 203) and Portfolio Service (Prompt 209) to manage total parent order fund/share holds while dispatching child orders incrementally.

## Scope Boundaries
- **In Scope:**
 - Ingestion, validation, and lifecycle management of synthetic parent orders (Iceberg, Trailing-Stop, Pegged, OCO, Bracket, TWAP).
 - High-frequency trigger evaluation against live market ticks.
 - Slicing and dispatching standard Limit/Market child orders to the core Order Service (Prompt 204) and Matching Engine (Prompt 205).
 - Child fill reconciliation, parent progress tracking, and residual hold release upon completion or cancellation.
 - Safe handling of exchange volatility halts (suspending child order dispatching during LULD circuit breakers).
- **Out of Scope / Handled Elsewhere:**
 - Raw in-memory Limit Order Book matching (Prompt 205).
 - Double-entry cash balance ledgering (Prompt 203).
 - FIX / OUCH protocol framing (Prompt 225).
 - Dark pool non-displayed crossing (Prompt 227).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing goroutine worker pools and channels for concurrent symbol monitoring.
- **Database & Storage:** PostgreSQL 16+ with `pgx/v5` for persistent parent/child order records; Redis 7.2 for ultra-fast watermark lookups, active parent order index caches, and distributed lease locking.
- **SQL Generation:** `sqlc` for compile-time verified, high-performance database access.
- **Messaging & Streaming:** `segmentio/kafka-go` or `confluent-kafka-go` consuming `matching.depth.v1` and `engine.matches.v1`, publishing child orders to `order.matching.commands.v1`.
- **Inter-Service Communication:** `google.golang.org/grpc` for low-latency calls to Order Service, Risk Engine, and Wallet Service.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `synthetic_parent_orders`, `synthetic_order_slices`, `trailing_stop_watermarks`.
- **Redis 7.2:** Key-value hashes for active trigger monitors (`algo:watermark:{order_id}`, `algo:pegged:{order_id}`).
- **Market Data Stream (Prompt 207 / 205):** Consumes real-time BBO depth and last-traded prices.
- **Order Service (Prompt 204):** Ingests child slice orders via gRPC or direct Kafka submission.
- **Risk Engine (Prompt 206):** Validates total parent order risk limits upon submission and re-verifies child order limits before injection.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Aggregated Settlement Linking:** Parent orders aggregate multiple micro-trades executed across child slices. Each executed child slice maps to an on-chain `match_id` settled via `SettlementDvP.sol` (Prompt 306).
- **Parent Order Traceability:** The parent synthetic order ID is cryptographically hashed and included in the metadata fields of internal execution receipts, enabling end-to-end DvP verification on Hyperledger Besu without exposing algorithmic intent on-chain.
- **Zero PII:** Only pseudonymous `user_id` and public Ethereum-compatible addresses (`0x...`) are processed.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/algorithmic-orders` with domain-driven design structure, strict linting, and dependency injection.
2. **Define Protobuf Contracts:** Create `proto/growww/algo/v1/algo_service.proto` defining endpoints for submitting, cancelling, and querying synthetic parent orders.
3. **Design PostgreSQL Schema:** Write database migrations for `synthetic_parent_orders`, `synthetic_order_slices`, and `algorithmic_triggers` with composite indexes on `(status, isin, trigger_type)`.
4. **Configure `sqlc` Queries:** Generate type-safe Go query models for atomic parent order creation, slice status updates, and parent completion calculations.
5. **Implement Pre-Trade Parent Reservation Coordinator:** Reserve the total maximum potential fund amount (for buy orders) or full fractional share quantity (for sell orders) via Wallet Service and Portfolio Service upon parent order intake.
6. **Implement Iceberg Slicing Engine:** Build algorithm that computes initial visible tip and hidden remainder; when a child slice fill event is consumed, dynamically generates the next slice with optional random jitter ($+/- 15\%$).
7. **Implement Trailing-Stop Monitor:** Build in-memory tick evaluator:
 - For Sell Trailing-Stop: update `peak_price = max(peak_price, current_price)`; trigger market/limit sell when `current_price <= peak_price - trail_amount` (or `peak_price * (1 - trail_percent)`).
 - For Buy Trailing-Stop: update `trough_price = min(trough_price, current_price)`; trigger buy when `current_price >= trough_price + trail_amount`.
8. **Implement Pegged Order Manager:** Monitor top-of-book BBO updates. When BBO moves beyond tick threshold, atomically cancel existing resting child slice and place new slice at updated pegged price level.
9. **Implement Bracket & OCO Coordinator:** Maintain mutual cancellation links between Take-Profit and Stop-Loss child orders; upon receipt of `FILLED` status for one leg, immediately dispatch cancellation for the counterpart leg.
10. **Implement TWAP Schedule Dispatcher:** Construct time-interval scheduler dividing total order quantity across $N$ discrete time windows, dispatching randomized sub-slices.
11. **Implement Circuit Breaker Interceptor:** Subscribe to market status events from Surveillance Engine (Prompt 228); automatically pause child slice generation during symbol trading halts.
12. **Build Failure Recovery & State Reconciliation:** On service startup, reload all active synthetic orders from PostgreSQL, rebuild Redis watermarks, and synchronize child order states with Order Service.
13. **Expose gRPC & REST APIs:** Implement gRPC handlers and REST gateway endpoints with structured validation and error handling.
14. **Instrument Prometheus Metrics & Tracing:** Track active synthetic orders count by type, trigger-to-dispatch latency histograms, and slice fill ratios.
15. **Write Comprehensive Test Suite:** Implement unit tests for all trigger algorithms, and integration tests using `testcontainers-go` simulating volatile market tick sequences and child order execution workflows.

## Interfaces / Contracts

### Protobuf Definition (`algo_service.proto`)
```protobuf
syntax = "proto3";

package growww.algo.v1;

option go_package = "growww/algo/v1;algov1";

service AlgoOrderService {
  rpc SubmitAlgoOrder (SubmitAlgoOrderRequest) returns (SubmitAlgoOrderResponse);
  rpc CancelAlgoOrder (CancelAlgoOrderRequest) returns (CancelAlgoOrderResponse);
  rpc GetAlgoOrder (GetAlgoOrderRequest) returns (GetAlgoOrderResponse);
  rpc ListActiveAlgoOrders (ListActiveAlgoOrdersRequest) returns (ListActiveAlgoOrdersResponse);
}

enum AlgoType {
  ALGO_TYPE_UNSPECIFIED = 0;
  ALGO_TYPE_ICEBERG = 1;
  ALGO_TYPE_TRAILING_STOP = 2;
  ALGO_TYPE_PEGGED = 3;
  ALGO_TYPE_BRACKET = 4;
  ALGO_TYPE_OCO = 5;
  ALGO_TYPE_TWAP = 6;
}

enum PegOffsetType {
  PEG_OFFSET_PRIMARY = 0;   // Peg to same side (Bid for Buy, Ask for Sell)
  PEG_OFFSET_MARKET = 1;    // Peg to opposite side
  PEG_OFFSET_MIDPOINT = 2;  // Peg to (Bid + Ask) / 2
}

message SubmitAlgoOrderRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string isin = 3;
  string side = 4; // "BUY" or "SELL"
  AlgoType algo_type = 5;
  string total_quantity = 6; // e.g. "100.000000"
  
  // Iceberg Parameters
  string display_quantity = 7; // e.g. "10.000000"
  bool randomize_display = 8;
  
  // Trailing-Stop Parameters
  string trail_amount = 9;      // Fixed INR trail (e.g. "5.00")
  string trail_percent = 10;    // Percentage trail (e.g. "0.02" for 2%)
  string activation_price = 11; // Optional activation threshold
  
  // Pegged Parameters
  PegOffsetType peg_type = 12;
  string peg_offset_paise = 13; // Offset in INR paise (+/-)
  string limit_price_cap = 14;  // Worst-case limit price collar
  
  // Bracket / OCO Parameters
  string take_profit_price = 15;
  string stop_loss_price = 16;
  
  // TWAP Parameters
  int64 twap_duration_seconds = 17;
  int32 twap_slices_count = 18;
}

message SubmitAlgoOrderResponse {
  string parent_order_id = 1;
  AlgoType algo_type = 2;
  string status = 3; // "ACTIVE", "PAUSED", "COMPLETED", "REJECTED"
  int64 created_at_unix = 4;
}

message CancelAlgoOrderRequest {
  string parent_order_id = 1;
  string user_id = 2;
  string reason = 3;
}

message CancelAlgoOrderResponse {
  string parent_order_id = 1;
  bool cancelled = 2;
  string final_status = 3;
}

message GetAlgoOrderRequest {
  string parent_order_id = 1;
  string user_id = 2;
}

message GetAlgoOrderResponse {
  string parent_order_id = 1;
  string user_id = 2;
  string isin = 3;
  AlgoType algo_type = 4;
  string status = 5;
  string total_quantity = 6;
  string filled_quantity = 7;
  string remaining_quantity = 8;
  string average_fill_price = 9;
  int32 active_slices_count = 10;
  int64 created_at_unix = 11;
  int64 updated_at_unix = 12;
}

message ListActiveAlgoOrdersRequest {
  string user_id = 1;
  string isin = 2;
}

message ListActiveAlgoOrdersResponse {
  repeated GetAlgoOrderResponse orders = 1;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE algo_order_type_enum AS ENUM (
    'ICEBERG', 
    'TRAILING_STOP', 
    'PEGGED', 
    'BRACKET', 
    'OCO', 
    'TWAP'
);

CREATE TYPE algo_order_status_enum AS ENUM (
    'PENDING', 
    'ACTIVE', 
    'PAUSED', 
    'COMPLETED', 
    'CANCELLED', 
    'REJECTED'
);

CREATE TABLE synthetic_parent_orders (
    parent_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    algo_type algo_order_type_enum NOT NULL,
    status algo_order_status_enum NOT NULL DEFAULT 'PENDING',
    
    total_quantity NUMERIC(18, 6) NOT NULL CHECK (total_quantity > 0),
    filled_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    remaining_quantity NUMERIC(18, 6) NOT NULL,
    
    -- Specific configuration parameters in JSONB
    algo_params JSONB NOT NULL,
    
    -- Active state tracking
    current_watermark NUMERIC(18, 4),
    active_child_order_id UUID,
    
    hold_id UUID NOT NULL, -- Total funds/shares reservation hold ID
    rejection_reason TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE synthetic_order_slices (
    slice_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_order_id UUID NOT NULL REFERENCES synthetic_parent_orders(parent_order_id) ON DELETE CASCADE,
    engine_order_id UUID NOT NULL, -- Child order submitted to Order Service
    slice_sequence INT NOT NULL,
    quantity NUMERIC(18, 6) NOT NULL,
    price NUMERIC(18, 4),
    status VARCHAR(32) NOT NULL, -- 'SUBMITTED', 'FILLED', 'PARTIALLY_FILLED', 'CANCELLED'
    filled_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    dispatched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_parent_orders_active ON synthetic_parent_orders(isin, status) WHERE status IN ('ACTIVE', 'PENDING');
CREATE INDEX idx_parent_orders_user ON synthetic_parent_orders(user_id, status);
CREATE INDEX idx_slices_parent ON synthetic_order_slices(parent_order_id);
```

## Security & Compliance Notes
- **Pre-Trade Hold Safety:** Synthetic parent orders lock 100% of the required funds or fractional shares upfront in Wallet/Portfolio service, preventing execution failures during child order replenishment.
- **Anti-Sniping Randomization:** Iceberg slices support cryptographically randomized size and interval variances to prevent predatory HFT algorithms from detecting hidden volume.
- **Runaway Algorithmic Protection:** Engine enforces hard caps on slice generation frequency (max 10 slices/sec per parent order) and maximum lifetime duration (24 hours max before mandatory renewal).
- **Surveillance Integration:** Every child slice links back to the parent `parent_order_id`, providing complete transparent audit trails for SEBI/IFSCA trade surveillance.

## Acceptance Criteria
- [ ] Iceberg orders correctly replenish visible tip upon partial/full execution until total quantity is exhausted.
- [ ] Trailing-stop sell triggers correctly update high-water mark and fire market/limit sell when price drops below the defined trailing threshold.
- [ ] Pegged orders dynamically adjust child order limit prices when the lit book NBBO shifts, honoring price caps.
- [ ] Bracket/OCO orders atomically cancel the opposite leg within 50ms of the first leg fill confirmation.
- [ ] Total fund holds are strictly reserved upfront and any residual unspent funds are released back to the wallet upon completion/cancellation.
- [ ] Service recovers active state deterministically following an abrupt process termination and restarts trigger monitors without double-placing child orders.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 104 (Kafka Standards), 203 (Wallet Service), 204 (Order Service), 205 (Order Matching Engine), 206 (Risk Engine).
- **Parallel Tasks:** 225 (FIX Protocol Gateway), 227 (Institutional Dark Pool Service).
- **Downstream Blockers:** 228 (Surveillance Engine), 509 (Flutter Order Placement Flow), 603 (Web Trading Terminal).
