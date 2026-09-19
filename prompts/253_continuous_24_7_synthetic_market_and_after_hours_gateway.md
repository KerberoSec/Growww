# 253 - Continuous 24/7 Synthetic Market & After-Hours Liquidity Gateway (Rust / Redis / Besu)

## Purpose
Traditional Indian equity and commodity exchanges operate strictly within standard market sessions (09:15 to 15:30 IST on weekdays), leaving capital markets closed during evenings, nights, weekends, and public holidays. However, global market events, international macroeconomic data releases, overseas corporate earnings, and cross-border geopolitical shifts occur continuously 24 hours a day, 7 days a week. International investors participating via IFSC GIFT City and domestic digital asset participants require continuous hedging, price discovery, and portfolio rebalancing capabilities without being forced to absorb overnight gap risk.

The **Continuous 24/7 Synthetic Market & After-Hours Liquidity Gateway** (`services/afterhours-matching-gateway`) decouples fractional tokenized Indian asset trading from primary exchange operational schedules. The service runs a high-throughput, low-latency matching and liquidity engine that seamlessly transitions between Primary Market Hours (09:15-15:30 IST), After-Market Hours (15:30-09:00 IST), and Pre-Open Auction (09:00-09:15 IST). During after-market windows, it powers a hybrid Central Limit Order Book (CLOB) supported by a virtual Automated Market Maker (vAMM) liquidity pool, guarantees instant on-chain Delivery-versus-Payment (DvP) settlement on Hyperledger Besu, applies dynamic volatility oracle collars to prevent predatory pricing during thin liquidity, and buffers primary-market designated orders for automated injection into the 09:00 IST NSE/BSE pre-open session. Across all executions, the platform invariant fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) (0.00% fee at launch; future fee parameters governed by FeeController.sol) is strictly calculated and enforced.

## What You Are Building
A mission-critical, ultra-low latency gateway and matching microservice (`services/afterhours-matching-gateway`) built in Rust. Core components include:
- **Session Transition & State Coordinator:** Deterministically manages platform operational states: `PRIMARY_OPEN` (09:15-15:30 IST), `AFTER_HOURS_CONTINUOUS` (15:30-09:00 IST), and `PRE_OPEN_BUFFERING` (09:00-09:15 IST).
- **Hybrid After-Hours CLOB + vAMM Matching Engine:** Facilitates peer-to-peer limit and market order matching in after-hours mode. When order book depth is thin, unmatched market or marketable limit orders interact against an automated virtual constant-product liquidity pool ($x \cdot y = k$) to guarantee execution without predatory slippage.
- **Dynamic Oracle Collar & Volatility Guard:** Continuously monitors composite benchmark prices (last primary market closing price, GIFT City synthetic futures, and international ADR/GDR feeds). Clamps allowable bid/ask prices within dynamic dynamic bands (e.g. $+/- 3.00\%$ under normal conditions, expanding to $+/- 5.00\%$ during high-volatility international sessions) to prevent market manipulation.
- **Primary Market Order Queue & Pre-Open Dispatcher:** Segregates orders marked for primary exchange execution (NSE/BSE/MCX) entered during after-hours. Validates pre-trade margin, holds orders in an encrypted memory-mapped queue, and injects them deterministically into the primary exchange FIX Gateway (Prompt 225 / Prompt 248) at 09:00:00 IST sharp.
- **24/7 Instant DvP Settlement Relayer:** Dispatches atomic on-chain settlement transactions to `SettlementDvP.sol` on Hyperledger Besu immediately upon after-hours trade execution, guaranteeing instant finality and zero counterparty settlement risk.
- **Universal Fee Integration:** Calculates and records the mandatory 0.00% transaction fee (No fee at all) on all after-hours trades, with dynamic routing governed by FeeController.sol (0.00% at launch).

## Scope Boundaries
- **In Scope:**
  - 24/7 order intake, validation, and session-aware routing.
  - After-hours hybrid CLOB order matching with integrated vAMM liquidity buffer.
  - Dynamic oracle price collar evaluation and order rejection for predatory out-of-band quotes.
  - Off-hours queuing, cancellation management, and pre-open 09:00 IST injection for primary exchange orders.
  - Real-time on-chain DvP settlement dispatch to Hyperledger Besu.
  - Universal 0.00% fee (No fee at all) calculation and split accounting.
- **Out of Scope / Handled Elsewhere:**
  - Primary exchange direct binary connectivity and drop-copy processing (handled in Prompt 225 / Prompt 242).
  - Primary exchange pre-open asymmetric speed bumps (handled in Prompt 248).
  - Core daytime matching engine cluster (handled in Prompt 205 / Prompt 246).
  - SPAN margin mathematical model and liquidation triggers (handled in Prompt 241).
  - Physical depository custody ingestion and demat locking (handled in Prompt 213).

## Technology to Use
- **Primary Language & Runtime:** **Rust 1.78+** using `tokio` multi-threaded asynchronous runtime and `tonic` for sub-millisecond gRPC services. Rust provides memory safety, deterministic execution, and zero garbage collection pauses required for continuous trading.
- **In-Memory Cache & State Store:** **Redis Cluster** with memory-mapped replication for ultra-fast session flags, order collar parameters, and vAMM pool curve invariants.
- **Message Broker:** **Apache Kafka** for low-latency event ingestion (`order.afterhours.submit.v1`) and distribution of execution events (`trade.afterhours.matched.v1`, `order.primary_queued.v1`).
- **Relational Persistence:** **PostgreSQL 16+** with `sqlx` for audit logging of after-hours executions, oracle price collar history, and primary market pre-open dispatch receipts.
- **Consortium Blockchain:** **Hyperledger Besu (QBFT)** for atomic 24/7 DvP settlement and cryptographic proof-of-trade notarization.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `afterhours_orders`, `vamm_pool_states`, `oracle_price_collars`, `primary_queued_orders`, `afterhours_trade_executions`, `afterhours_fee_distributions`.
- **Kafka Topics:** Consumes `order.afterhours.submit.v1`, `marketdata.closing_price.v1`, `oracle.synthetic_adr.v1`; publishes `order.afterhours.matched.v1`, `order.primary_queued.v1`, `settlement.dvp_requested.v1`, `risk.collar_breach.v1`.
- **Matching Engine (Prompt 205):** Coordinates graceful session handover at 15:30:00 IST and 09:00:00 IST.
- **Exchange Adapter & Speed Bump Guard (Prompt 248):** Receives queued primary market orders from the gateway for pre-open auction injection.
- **Real-Time Market Surveillance (Prompt 228 / Prompt 715):** Ingests after-hours tick streams to detect spoofing, wash trading, or abnormal volatility patterns during thin trading windows.
- **Fixed Fee & Revenue Distribution Engine (Prompt 244):** Consumes after-hours fee ledgers to execute automatic ledger distributions.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Instant 24/7 On-Chain DvP Settlement:** While traditional clearing corporations suspend settlement overnight, the after-hours gateway immediately submits executed trades to `SettlementDvP.sol`. Tokenized shares and Digital Rupee (e₹) CBDC / synthetic stablecoin balances transfer atomically within single QBFT block times (< 2 seconds).
- **Oracle Attestation Verification:** Price collar thresholds are cross-referenced with signed EIP-712 price feeds submitted by authorized consortium oracle nodes to `OracleAggregator.sol`.
- **Zero PII Exposure:** The gateway operates entirely on pseudonymous internal account identifiers and on-chain wallet addresses. No investor PII (such as PAN, Aadhaar, names, or contact details) is ever written to the ledger or Kafka payloads.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize the Rust workspace project `services/afterhours-matching-gateway` with dependencies: `tokio`, `tonic`, `prost`, `rdkafka`, `redis`, `sqlx`, `rust_decimal`, and `tracing`.
2. **Define Protobuf Schemas:** Create `proto/growww/afterhours/v1/afterhours_service.proto` specifying RPC methods: `SubmitAfterHoursOrder`, `CancelAfterHoursOrder`, `GetAfterHoursMarketDepth`, `GetVammPoolState`, and `QueryQueuedPrimaryOrders`.
3. **Generate gRPC Stubs:** Use `tonic-build` in `build.rs` to generate client and server stubs for Rust services.
4. **Design PostgreSQL Schema:** Write database migration scripts creating `afterhours_orders`, `vamm_pool_states`, `oracle_price_collars`, `primary_queued_orders`, and `afterhours_trade_executions` tables.
5. **Implement Session Transition State Machine:** Build a timer-driven state machine coordinating market modes: `PRIMARY_OPEN` (09:15-15:30 IST), `AFTER_HOURS_CONTINUOUS` (15:30-09:00 IST), and `PRE_OPEN_BUFFERING` (09:00-09:15 IST).
6. **Implement Dynamic Oracle Collar Engine:** Ingest closing prices and international ADR/GDR synthetic prices. Compute hard dynamic price collars (e.g. $[P_{\text{ref}} \cdot (1 - \text{collar}), P_{\text{ref}} \cdot (1 + \text{collar})]$) and reject non-compliant order submissions.
7. **Implement In-Memory After-Hours CLOB:** Build a high-performance price-time priority limit order book tailored for after-hours trading with sub-10 microsecond matching latency.
8. **Implement Virtual AMM (vAMM) Liquidity Pool:** Construct a constant-product virtual liquidity pool ($x \cdot y = k$) to provide automated market-maker backing for marketable orders when the CLOB spread exceeds predefined liquidity tolerances.
9. **Implement Primary Market Queue Buffer:** Implement an encrypted, persistent queue that buffers orders designated with `EXECUTION_VENUE_PRIMARY` received outside primary hours, ensuring FIFO ordering.
10. **Implement Pre-Open Auction Dispatcher:** Build the automated 09:00:00 IST scheduler that streams buffered primary orders to the exchange FIX gateway (Prompt 225 / Prompt 248) with rate-governed burst controls.
11. **Implement 24/7 On-Chain DvP Settlement Relayer:** Connect to the Hyperledger Besu relayer pool (Prompt 245) to dispatch atomic settlement payloads to `SettlementDvP.sol` immediately upon trade fill.
12. **Implement fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Calculator:** Assess 0.00% (Zero Fee) platform fee on all executed volume and publish fee events (0.00% at launch; FeeController governed) to Kafka.
13. **Implement Redis State Caching:** Maintain live order book top-of-book, vAMM pool reserves, and session statuses in Redis Cluster with sub-millisecond read access.
14. **Write Integration & Invariant Test Suite:** Build automated test harnesses verifying: (a) collar rejection of out-of-band quotes, (b) seamless vAMM fallback execution, (c) 100% pre-open delivery at 09:00:00 IST, and (d) mathematical precision of 0.00% fee (No fee at all) splits.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/afterhours/v1/afterhours_service.proto`)
```protobuf
syntax = "proto3";

package growww.afterhours.v1;

option go_package = "github.com/growww/proto/gen/go/afterhours/v1;afterhoursv1";

enum MarketSessionState {
  MARKET_SESSION_STATE_UNSPECIFIED = 0;
  MARKET_SESSION_STATE_PRIMARY_OPEN = 1;        // 09:15 - 15:30 IST
  MARKET_SESSION_STATE_AFTER_HOURS = 2;         // 15:30 - 09:00 IST
  MARKET_SESSION_STATE_PRE_OPEN_BUFFERING = 3;  // 09:00 - 09:15 IST
  MARKET_SESSION_STATE_HALTED = 4;
}

enum ExecutionVenuePreference {
  EXECUTION_VENUE_PREFERENCE_UNSPECIFIED = 0;
  EXECUTION_VENUE_PREFERENCE_AFTER_HOURS_SYNTHETIC = 1; // Match 24/7 on NBSE CLOB + vAMM
  EXECUTION_VENUE_PREFERENCE_PRIMARY_QUEUE = 2;         // Buffer for next 09:00 IST pre-open
  EXECUTION_VENUE_PREFERENCE_HYBRID_BEST_EFFORT = 3;    // Try after-hours, queue remainder for primary
}

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BUY = 1;
  ORDER_SIDE_SELL = 2;
}

message SubmitAfterHoursOrderRequest {
  string order_id = 1;
  string account_id = 2;
  string isin = 3;
  OrderSide side = 4;
  uint64 quantity_e6 = 5;               // Fractional quantity (6 decimal places)
  uint64 limit_price_paise = 6;         // Price in paise (INR 0.01 = 1 paise)
  ExecutionVenuePreference venue_pref = 7;
  int64 client_timestamp_ns = 8;
  string signature = 9;
}

message SubmitAfterHoursOrderResponse {
  string order_id = 1;
  string status = 2;                    // "ACCEPTED_MATCHED", "ACCEPTED_QUEUED", "REJECTED_COLLAR_BREACH"
  uint64 executed_quantity_e6 = 3;
  uint64 executed_price_paise = 4;
  uint64 total_fee_paise = 5;           // 0.00% (No fee at all) platform fee
  uint64 treasury_fee_paise = 6;        // Governed by FeeController (0.00% at launch)
  uint64 sgf_fee_paise = 7;             // Governed by FeeController (0.00% at launch)
  uint64 ipf_fee_paise = 8;             // Governed by FeeController (0.00% at launch)
  string tx_hash = 9;                   // Besu DvP settlement transaction hash
  int64 processed_at_ns = 10;
}

message GetVammPoolStateRequest {
  string isin = 1;
}

message GetVammPoolStateResponse {
  string isin = 1;
  uint64 virtual_token_reserve_e6 = 2;
  uint64 virtual_quote_reserve_paise = 3;
  uint64 spot_price_paise = 4;
  uint32 dynamic_collar_bps = 5;        // Collar width in basis points
  uint64 min_collar_price_paise = 6;
  uint64 max_collar_price_paise = 7;
  int64 updated_at_ns = 8;
}

service AfterHoursGatewayService {
  rpc SubmitAfterHoursOrder (SubmitAfterHoursOrderRequest) returns (SubmitAfterHoursOrderResponse);
  rpc GetVammPoolState (GetVammPoolStateRequest) returns (GetVammPoolStateResponse);
}
```

### PostgreSQL Database Schema (`services/afterhours-matching-gateway/migrations/001_afterhours_schema.sql`)
```sql
CREATE TABLE market_session_states (
    session_id VARCHAR(32) PRIMARY KEY,
    current_state VARCHAR(32) NOT NULL DEFAULT 'PRIMARY_OPEN',
    primary_open_time_ist TIME NOT NULL DEFAULT '09:15:00',
    primary_close_time_ist TIME NOT NULL DEFAULT '15:30:00',
    pre_open_time_ist TIME NOT NULL DEFAULT '09:00:00',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE oracle_price_collars (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    reference_close_price_paise BIGINT NOT NULL,
    synthetic_adr_price_paise BIGINT,
    collar_bandwidth_bps INT NOT NULL DEFAULT 300, -- 3.00% standard collar
    min_allowed_price_paise BIGINT NOT NULL,
    max_allowed_price_paise BIGINT NOT NULL,
    last_oracle_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vamm_pool_states (
    isin VARCHAR(12) PRIMARY KEY,
    virtual_token_reserve NUMERIC(28, 6) NOT NULL,
    virtual_quote_reserve NUMERIC(28, 2) NOT NULL, -- in paise
    k_invariant NUMERIC(56, 8) NOT NULL,
    last_rebalance_price_paise BIGINT NOT NULL,
    total_liquidity_paise BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE primary_queued_orders (
    queued_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    order_side VARCHAR(4) NOT NULL,
    quantity NUMERIC(18, 6) NOT NULL,
    limit_price_paise BIGINT NOT NULL,
    queue_status VARCHAR(24) NOT NULL DEFAULT 'PENDING_PRE_OPEN', -- PENDING_PRE_OPEN, DISPATCHED, CANCELLED
    dispatched_at TIMESTAMPTZ,
    exchange_ack_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE afterhours_trade_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    buyer_account_id VARCHAR(64) NOT NULL,
    seller_account_id VARCHAR(64) NOT NULL,
    execution_venue VARCHAR(16) NOT NULL, -- 'CLOB' or 'VAMM'
    executed_quantity NUMERIC(18, 6) NOT NULL,
    executed_price_paise BIGINT NOT NULL,
    gross_notional_paise BIGINT NOT NULL,
    platform_fee_paise BIGINT NOT NULL,   -- 0.00% (No fee at all)
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    besu_tx_hash VARCHAR(66) NOT NULL,
    settlement_status VARCHAR(24) NOT NULL DEFAULT 'SETTLED',
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_queued_orders_status ON primary_queued_orders(queue_status, created_at);
CREATE INDEX idx_afterhours_trades_isin ON afterhours_trade_executions(isin, executed_at);
```

## Security & Compliance Notes
- **Dynamic Collar Hard-Stop:** Any order with a limit price outside the configured dynamic oracle collar ($[P_{\text{min}}, P_{\text{max}}]$) is rejected immediately at the gateway boundary, preventing predatory off-market price manipulation during illiquid overnight hours.
- **Fixed Fee Invariant:** The 0.00% transaction fee (No fee at all) is hardcoded into the execution pipeline and split strictly into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance, ensuring regulatory compliance and solvency backing.
- **Atomic Double-Entry Isolation:** Queued primary market orders lock collateral immediately upon acceptance, preventing double-spend attacks across parallel after-hours and pre-open venues.
- **Zero PII Leakage:** All messages across gRPC, Kafka, PostgreSQL, and Hyperledger Besu store only pseudonymous account hashes and standard financial parameters.

## Acceptance Criteria
- [ ] Gateway smoothly transitions operational states across `PRIMARY_OPEN`, `AFTER_HOURS_CONTINUOUS`, and `PRE_OPEN_BUFFERING` based on IST clock triggers.
- [ ] After-hours CLOB matches peer orders with sub-millisecond latency.
- [ ] Fallback to vAMM pool executes deterministically when CLOB liquidity is exhausted or spreads widen past threshold.
- [ ] Dynamic oracle collars reject bids and asks breaching $+/- 3.00\%$ (or dynamic setting) of reference prices.
- [ ] 100% of buffered primary market orders are injected into the pre-open auction at 09:00:00 IST sharp with valid exchange acknowledgments.
- [ ] Instant 24/7 on-chain DvP settlements finalize on Hyperledger Besu in under 2 seconds per trade.
- [ ] 0.00% (No fee at all) platform fee is assessed on all executed notionals and split accurately (0.00% fee at launch (governed by FeeController.sol)).

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 204 (Order Service), Prompt 205 (Order Matching Engine), Prompt 306 (Settlement DvP).
- **Parallel Work:** Prompt 225 (FIX Protocol Gateway), Prompt 244 (Fixed Fee Engine), Prompt 248 (Exchange Adapter & Pre-Open Speed Bump Guard).
- **Enables:** 24/7 continuous trading for global international investors and round-the-clock risk management.
