# 227 - 24/7 Institutional Dark Pool & Block Crossing Service (Rust)

## Purpose
In a continuous 24/7 equity exchange, large institutional asset managers, sovereign wealth funds, and market makers face significant market impact and information leakage when executing large block trades on a public (lit) order book.

The Institutional Dark Pool & Block Crossing Service provides an off-book, non-displayed crossing network operating continuously under SEBI and IFSCA GIFT City regulatory guidelines for Large-in-Scale (LIS) transactions and institutional liquidity provision.

Key operating principles:
1. **Zero Pre-Trade Transparency:** Orders in the dark pool are completely invisible to the public market; depth, prices, and sizes are never broadcast over public market data feeds.
2. **Midpoint & Collar Execution:** Matches execute strictly at or within the National Best Bid and Offer (NBBO) midpoint derived in real time from the lit Growww matching engine (Prompt 205), guaranteeing price improvement without slippage.
3. **Minimum Fill Size (MFS) & Anti-Toxicity Controls:** Institutions can mandate minimum execution sizes to prevent retail pinging and high-frequency toxic flow exploration.
4. **Immediate Post-Trade DvP Settlement & Tape Reporting:** All matched block trades are immediately printed to the public tape (complying with SEBI/IFSCA 15-second post-trade transparency mandates) and routed directly to the permissioned blockchain for atomic Delivery-versus-Payment (DvP) settlement on Hyperledger Besu.

## What You Are Building
A low-latency, privacy-preserving Rust microservice (`services/dark-pool`). Concrete deliverables include:
- **Non-Displayed Order Book Engine:** In-memory queue matching un-displayed buy and sell orders using Size-Time Priority or Midpoint Crossing Priority.
- **Dynamic NBBO Reference Tracker:** Real-time consumer of lit order book top-of-book quotes (`matching.depth.v1`) to dynamically anchor crossing prices to the midpoint $\frac{\text{BestBid} + \text{BestAsk}}{2}$.
- **Discrete Frequent Batch Auction (FBA) Engine:** Periodic micro-batch auction runner (every 100ms) that matches accumulated non-displayed interest simultaneously at a single clearing price to eliminate latency arbitrage.
- **Minimum Execution Size (MES/MFS) Filter:** Matching logic ensuring orders only execute if counter-liquidity satisfies the client's minimum quantity threshold.
- **Post-Trade Tape Printer:** Kafka publisher that emits post-trade execution prints to the public consolidated tape (`marketdata.tape.prints.v1`).
- **Blockchain DvP Settlement Dispatcher:** Direct atomic trade submission pipeline feeding `SettlementDvP.sol` on Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
 - Non-displayed order intake (Dark Limit, Dark Midpoint, Dark Pegged).
 - Continuous midpoint crossing and periodic discrete batch crossing.
 - Large-in-Scale (LIS) minimum block size enforcement.
 - User and liquidity tier categorization (Institutional vs. Market Maker vs. Systematic Internaliser).
 - Anti-gaming jitter delays (random 1-5ms arrival time randomization) to thwart latency snipers.
 - Post-trade public tape broadcasting and regulatory trade reporting.
- **Out of Scope / Handled Elsewhere:**
 - Public lit order book matching (Prompt 205).
 - FIX protocol gateway transport (Prompt 225).
 - Pre-trade retail risk checks (Prompt 206).
 - Cash wallet ledger double-entry storage (Prompt 203).

## Technology to Use
- **Primary Language & Runtime:** Rust 2021 Edition (stable 1.78+) for microsecond execution determinism, zero memory leaks, and GC-free low-latency concurrency.
- **Async Runtime & Locks:** `tokio`, `parking_lot::RwLock`, and `crossbeam` channels for intra-service lock-free order handoffs.
- **Data Structures:** Custom `VecDeque` and `BTreeMap` structures indexed by ISIN and side for Size-Time priority queues.
- **Persistence & Checkpoints:** `rocksdb` for persistent order book journal and snapshotting.
- **Database & Storage:** PostgreSQL 16+ for institutional trade history, fee reconciliation, and regulatory audit archives.
- **Messaging & Inter-Service RPC:** `rdkafka` for market data depth consumption and tape publishing; `tonic` / `prost` for gRPC order entry and settlement triggers.

## Backend / Infra Touchpoints
- **Lit Matching Engine (Prompt 205):** Consumes real-time BBO stream to maintain valid reference price collars.
- **FIX / OUCH Gateway (Prompt 225):** Ingests institutional dark pool order messages via FIX Tag `18=M` (Midpoint) / Tag `109=DARK`.
- **Trade Settlement Service (Prompt 208):** Ingests dark match events to trigger DvP smart contract execution.
- **Market Data Service (Prompt 207):** Receives post-trade execution prints for public tape dissemination.
- **Apache Kafka Topics:** Consumes from `matching.depth.v1`; publishes to `darkpool.matches.v1` and `marketdata.tape.prints.v1`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Direct DvP Settlement:** Dark pool trades trigger direct atomic transfers on `SettlementDvP.sol` (Prompt 306). Buyer cash tokens are exchanged for seller fractional equity tokens atomically in the same QBFT block.
- **On-Chain Privacy:** While post-trade price and volume are printed to the public tape, institutional wallet addresses and trade counterparties are kept confidential using pseudonymous token IDs and zero PII.
- **Settlement Receipt Proof:** Outbound trade confirmations include the on-chain settlement transaction hash and block number confirming 1:1 asset backing on Hyperledger Besu.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Rust Project:** Initialize Cargo crate `services/dark-pool` with high-performance compiler release configurations.
2. **Define Fixed-Point Price & Quantity Types:** Reuse standardized 128-bit fixed-point representations for fractional equity lots ($10^{-6}$) and INR sub-paise ($10^{-4}$).
3. **Build Dynamic Reference Price Tracker:** Implement thread-safe cache subscribed to lit book BBO; compute reference midpoint:
   $$\text{Midpoint} = \frac{\text{BestBid} + \text{BestAsk}}{2}$$
   and enforce price collar bands ($[\text{BestBid}, \text{BestAsk}]$).
4. **Implement Dark Order Book Structure:** Build `DarkOrderBook` managing non-displayed orders partitioned by ISIN with fields for `min_fill_quantity`, `allowed_counterparty_types`, and `time_in_force`.
5. **Implement Continuous Midpoint Crossing Algorithm:** On arrival of a new dark order, check for crossing candidate orders on the opposite side where counterparty limit covers the current lit midpoint and quantity satisfies both parties' `min_fill_quantity`.
6. **Implement Frequent Batch Auction (FBA) Mode:** Build 100ms discrete auction timer: collect all resting dark orders, calculate single clearing price maximizing matched volume, and execute multi-party matches atomically.
7. **Implement Anti-Sniping Arrival Jitter:** Add randomized microsecond delay queue (1-5ms) to all incoming dark orders to prevent HFT cross-market latency arbitrage.
8. **Implement Self-Match & Counterparty Filtering:** Prevent crossing between trading desks of the same legal entity (`firm_id`) and allow institutions to opt-out of matching against retail flow or specific liquidity tiers.
9. **Build Post-Trade Tape Publisher:** Stream execution prints containing `isin`, `matched_price`, `matched_quantity`, and `timestamp` to `marketdata.tape.prints.v1` within $< 15\text{ms}$ of match execution.
10. **Build RocksDB WAL & Snapshotter:** Persist incoming dark orders and trade journals to local NVMe RocksDB for instantaneous crash recovery without state loss.
11. **Implement gRPC Dark Pool Ingress API:** Implement `tonic` gRPC service implementing `SubmitDarkOrder`, `CancelDarkOrder`, and `GetDarkOrderStatus`.
12. **Integrate Blockchain Settlement Dispatcher:** Stream dark match execution receipts directly to Trade Settlement Service (Prompt 208) for on-chain `SettlementDvP.sol` submission.
13. **Design PostgreSQL Archival Schema:** Create tables for `dark_pool_orders`, `dark_pool_executions`, and `institutional_counterparty_metrics`.
14. **Instrument Metrics & Health Checks:** Export Prometheus metrics tracking dark pool volume ratio vs. lit book, midpoint price improvement amounts in INR, and rejection rates.
15. **Execute Chaos & Stress Tests:** Simulate concurrent institutional block orders, rapid lit book NBBO shifts, and batch auction clearing correctness under heavy load.

## Interfaces / Contracts

### Protobuf Definition (`dark_pool_service.proto`)
```protobuf
syntax = "proto3";

package growww.darkpool.v1;

option go_package = "growww/darkpool/v1;darkpoolv1";

service DarkPoolService {
  rpc SubmitDarkOrder (SubmitDarkOrderRequest) returns (SubmitDarkOrderResponse);
  rpc CancelDarkOrder (CancelDarkOrderRequest) returns (CancelDarkOrderResponse);
  rpc GetDarkOrderStatus (GetDarkOrderStatusRequest) returns (GetDarkOrderStatusResponse);
}

enum DarkCrossMode {
  DARK_CROSS_CONTINUOUS_MIDPOINT = 0;
  DARK_CROSS_DISCRETE_BATCH_AUCTION = 1;
}

enum LiquidityTier {
  TIER_INSTITUTIONAL = 0;
  TIER_MARKET_MAKER = 1;
  TIER_SYSTEMATIC_INTERNALISER = 2;
}

message SubmitDarkOrderRequest {
  string client_order_id = 1;
  string firm_id = 2;
  string isin = 3;
  string side = 4; // "BUY" or "SELL"
  string micro_quantity = 5; // e.g. "5000000000" (5,000 shares in micro-units)
  string limit_price_sub_paise = 6; // Worst-case limit price collar
  string min_fill_quantity = 7; // Minimum Fill Size (MFS)
  DarkCrossMode cross_mode = 8;
  LiquidityTier allowed_counterparty_tier = 9;
  int64 expiry_unix_seconds = 10;
}

message SubmitDarkOrderResponse {
  string dark_order_id = 1;
  string status = 2; // "ACCEPTED", "RESTING", "FILLED", "REJECTED"
  string initial_matched_quantity = 3;
  int64 timestamp_ns = 4;
}

message CancelDarkOrderRequest {
  string dark_order_id = 1;
  string firm_id = 2;
  string reason = 3;
}

message CancelDarkOrderResponse {
  string dark_order_id = 1;
  bool cancelled = 2;
  string remaining_quantity = 3;
}

message GetDarkOrderStatusRequest {
  string dark_order_id = 1;
  string firm_id = 2;
}

message GetDarkOrderStatusResponse {
  string dark_order_id = 1;
  string isin = 2;
  string side = 3;
  string original_quantity = 4;
  string filled_quantity = 5;
  string remaining_quantity = 6;
  string average_fill_price = 7;
  string status = 8;
  int64 created_at_unix = 9;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE dark_order_status_enum AS ENUM (
    'ACCEPTED', 
    'RESTING', 
    'PARTIALLY_FILLED', 
    'FILLED', 
    'CANCELLED', 
    'EXPIRED', 
    'REJECTED'
);

CREATE TABLE dark_pool_orders (
    dark_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firm_id VARCHAR(64) NOT NULL,
    client_order_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    status dark_order_status_enum NOT NULL DEFAULT 'ACCEPTED',
    
    total_quantity NUMERIC(18, 6) NOT NULL CHECK (total_quantity > 0),
    min_fill_quantity NUMERIC(18, 6) NOT NULL DEFAULT 1.000000,
    filled_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    remaining_quantity NUMERIC(18, 6) NOT NULL,
    
    limit_price NUMERIC(18, 4), -- Optional collar
    cross_mode VARCHAR(32) NOT NULL DEFAULT 'CONTINUOUS_MIDPOINT',
    allowed_counterparty_tier VARCHAR(32) NOT NULL DEFAULT 'ANY',
    
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE TABLE dark_pool_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dark_order_id UUID NOT NULL REFERENCES dark_pool_orders(dark_order_id),
    counterparty_order_id UUID NOT NULL,
    isin VARCHAR(12) NOT NULL,
    matched_price NUMERIC(18, 4) NOT NULL, -- Fixed midpoint price
    matched_quantity NUMERIC(18, 6) NOT NULL,
    lit_best_bid NUMERIC(18, 4) NOT NULL,
    lit_best_ask NUMERIC(18, 4) NOT NULL,
    price_improvement_inr NUMERIC(18, 4) NOT NULL,
    tape_print_id VARCHAR(64) NOT NULL,
    onchain_settlement_tx VARCHAR(66),
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dark_orders_firm ON dark_pool_orders(firm_id, status);
CREATE INDEX idx_dark_orders_isin ON dark_pool_orders(isin, status);
CREATE INDEX idx_dark_executions_isin ON dark_pool_executions(isin, executed_at);
```

## Security & Compliance Notes
- **SEBI/IFSCA Large-in-Scale (LIS) Compliance:** Dark crossing enforces minimum ticket size thresholds (e.g., minimum INR 20 Lakhs / USD 25,000 equivalent) to ensure retail order flow is not improperly internalized without regulatory disclosure.
- **Reference Price Collar Enforcement:** Orders cannot execute outside the lit order book's National Best Bid and Offer (NBBO) spread, preventing off-market price dislocation or fraudulent transfer pricing.
- **Mandatory 15-Second Tape Disclosure:** Every executed dark match publishes price, size, and timestamp to the public market tape within 15 seconds of matching to satisfy post-trade transparency rules.
- **Anti-Toxicity & Non-Discriminatory Access:** All participants matching criteria operate under deterministic code rules with complete cryptographic audit trails.

## Acceptance Criteria
- [ ] Dark orders execute strictly at or within the live lit order book NBBO midpoint.
- [ ] Minimum Fill Size (MFS) constraints are rigorously enforced; orders never execute for less than their configured minimum quantity.
- [ ] Non-displayed orders are 100% invisible on public L2/L3 ITCH and WebSocket market depth feeds.
- [ ] Executed matches immediately emit post-trade tape prints to `marketdata.tape.prints.v1`.
- [ ] Matches trigger atomic DvP smart contract settlement on Hyperledger Besu with verified transaction hash linkages.
- [ ] Periodic discrete batch auctions (FBA) compute volume-maximizing clearing prices deterministically without price leakage.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (Architecture Overview), 205 (Order Matching Engine), 206 (Risk Engine), 207 (Market Data Service), 208 (Settlement Service).
- **Parallel Tasks:** 225 (FIX Protocol Gateway), 226 (Advanced Order Types Engine).
- **Downstream Blockers:** 228 (Real-Time Market Surveillance Engine), 216 (Regulatory Reporting Service).
