# 268 - Real-Time Auto-Deleveraging (ADL) & Margin Default Resolver (Rust)

## Purpose
In extreme market gap events, high-volatility flash crashes, or sudden liquidity vacuums, forced liquidations of under-margined derivative positions cannot always be filled on the central limit order book before market price pierces the liquidation threshold. When a liquidated position's execution price drops below its Bankruptcy Price ($P_{bk}$), the trader enters negative equity ($\text{Account Equity} < 0$), generating an unbacked deficit for the clearinghouse. Under standard risk protocols (ADR-0029, RUNBOOK-19), the exchange Settlement Guarantee Fund (SGF) and liquidation penalty insurance reserves absorb this deficit. 

However, during catastrophic systemic market shocks, the cumulative bankruptcy deficit may deplete the entire available Insurance Fund and SGF liquid buffer. Without a deterministic resolution mechanism, the exchange faces central counterparty (CCP) insolvency. Socializing losses across unleveraged spot traders or halting settlements is strictly prohibited under SEBI Comprehensive Risk Management and Clearing Corporation regulations.

The **Real-Time Auto-Deleveraging (ADL) & Margin Default Resolver** operates as the exchange's ultimate solvency backstop (RUNBOOK-25). When liquidated positions produce negative balances and the SGF insurance buffer is completely exhausted, the ADL Engine steps in. It calculates real-time counterparty priority rankings across all profitable and highly leveraged accounts, selects opposing positions deterministically, and executes immediate off-book matching at the bankrupt account's Bankruptcy Price. This halts deficit accumulation instantly, guarantees central counterparty solvency, and maintains full market integrity without mutualized clawbacks on spot traders.

## What You Are Building
An ultra-low latency, mission-critical distributed Rust microservice (`services/adl-engine`) operating in the nanosecond-to-microsecond latency regime. Concrete deliverables include:

- **Real-Time ADL Ranking Core (Rust):** A lock-free, in-memory ranking pipeline continuously sorting active counterparty positions per instrument and per direction (Long/Short) based on the standardized ADL score formula:
  $$\text{ADL Score} = \text{ProfitRankingPercentile} \times \text{EffectiveLeverage}$$
- **Lock-Free Priority Queue Engine:** A high-performance concurrency architecture utilizing cache-padded skip lists (`crossbeam-skiplist`) and atomic rank trackers providing lock-free $O(\log N)$ position updates, insertions, and $O(1)$ top-rank candidate evictions during market-wide price movements.
- **Dynamic 5-Bar ADL Indicator Daemon:** A streaming telemetry calculator that maps each trader's current ADL percentile into a transparent 5-bar visual indicator (0-20%, 20-40%, 40-60%, 60-80%, 80-100%) and broadcasts priority changes over WebSockets via Redis Pub/Sub.
- **Deterministic Bankruptcy Matcher:** An execution unit that consumes unfillable liquidation bankruptcy events from Kafka, computes exact contract allocations across top-ranked ADL counterparties at $P_{bk}$, and generates atomic execution instructions for the Matching Engine (Prompt 205).
- **SGF Buffer Exhaustion Monitor:** A real-time threshold watcher tracking live SGF / Insurance Fund liquid reserves from the SGF Service (Prompt 230), switching the clearinghouse between standard insurance absorption mode and emergency ADL execution mode with sub-millisecond precision.
- **On-Chain Settlement Reconciliation Relayer:** An asynchronous bridge invoking the default waterfall smart contract (`SettlementGuaranteeFund.sol`) on Hyperledger Besu, logging immutable cryptographic proofs of negative equity absorption and ADL contract closeouts.
- **PostgreSQL Audit & Regulatory Vault:** A persistent ledger recording all ADL triggers, counterparty rankings, matched bankruptcy fills, and SEBI systemic risk compliance reports.

## Scope Boundaries
- **In Scope:**
  - Ingestion of live position updates, mark price feeds, and maintenance margin ratios across all leveraged derivatives contracts.
  - Continuous mathematical calculation of Profit Ranking Percentiles and Effective Leverage per participant.
  - Maintenance of lock-free in-memory priority queues of opposing counterparties per contract symbol.
  - Real-time calculation and broadcasting of the trader 5-bar ADL risk indicator.
  - Integration with SGF Service to detect insurance pool depletion and arm/disarm the ADL execution protocol.
  - Deterministic slicing and assignment of bankrupt position volumes against top-priority ADL candidates at the exact Bankruptcy Price.
  - Publishing matched execution events to Matching Engine (Prompt 205), Risk Service (Prompt 206), and user notification streams with error code `ERR_ADL_EXECUTED`.
  - Dispatching cryptographic proof-of-ADL and negative balance reconciliation transactions to Hyperledger Besu.
- **Out of Scope / Handled Elsewhere:**
  - Initial Maintenance Margin Requirement (MMR) checks and routine pre-trade margin validation (handled in Prompt 206 and Prompt 229).
  - Primary forced liquidation order routing into the public lit order book (handled in Prompt 208 and Prompt 230).
  - Daily SGF stress testing (Cover-1 / Cover-2) and regular member contribution sizing (handled in Prompt 230).
  - Cash wallet fiat deposits and banking payment gateway reconciliations (handled in Prompt 203 and Prompt 212).
  - Unleveraged spot equity clearing and central depository settlement (handled in Prompt 213 and Prompt 260).

## Technology to Use
- **Core Language & Concurrency:** **Rust 2021 Edition** (1.78+) leveraging `tokio` multi-threaded runtime for asynchronous I/O and dedicated pinned worker threads (`core_affinity`) for core mathematical ranking loops.
- **Lock-Free Memory Structures:** `crossbeam` and `crossbeam-skiplist` for concurrent, lock-free priority queue management; atomic primitives (`std::sync::atomic::{AtomicU64, AtomicBool}`) for zero-mutex execution triggers.
- **Memory Optimization:** Fixed-size structs, zero-heap allocators where feasible, cache-line aligned structures (`#[repr(align(64))]`) to eliminate CPU false sharing across worker cores.
- **In-Memory Cache & Indicator Pub/Sub:** **Redis 7.2+ Cluster** utilizing Redis Sorted Sets (ZSETs) for external status queries and Redis Streams / Pub/Sub for high-throughput 5-bar indicator distribution.
- **Event Streaming:** **Apache Kafka** (`rdkafka` Rust driver) consuming `risk.liquidation.unfilled.v1`, `sgf.pool_balance.v1`, and `market.mark_price.v1`; publishing to `adl.executions.v1`, `adl.indicators.v1`, and `compliance.sebi_alerts.v1`.
- **Database & Persistence:** **PostgreSQL 16+** with `sqlx` (Rust async driver) for immutable audit trails, historical default snapshots, and regulatory reporting tables.
- **RPC Framework:** **gRPC (HTTP/2 with Protobuf)** via `tonic` and `prost` for synchronous low-latency query endpoints.

## Backend / Infra Touchpoints
- **Matching Engine (Prompt 205):** Ingests atomic ADL match instructions via private dedicated Unix domain sockets or ultra-fast IPC, bypassing public order books to directly update trade journals and position states.
- **Real-Time VaR & Margin Engine (Prompt 229):** Supplies continuous mark-to-market position valuations, unrealized PnL, and current portfolio equity for counterparty ranking scores.
- **SGF & Default Waterfall Service (Prompt 230):** Broadcasts real-time SGF insurance fund balance updates; triggers the ADL engine when Tier 5 SGF capital is fully exhausted.
- **Pre-Trade Risk Engine (Prompt 206):** Receives instant notifications of reduced counterparty positions to immediately decrement open exposure and release locked maintenance margins.
- **Market Data Service (Prompt 207):** Provides real-time consolidated Mark Price and Index Price tick feeds to evaluate unrealized PnL and trigger price-dependent rank rebalances.
- **Notification Service (Prompt 211):** Emits priority automated alerts (SMS, Email, In-App Push, Webhook) to de-leveraged traders explaining the exact fill price, freed margin, and regulatory reference.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain SGF Waterfall Invocation:** When the off-chain engine exhausts SGF reserves and triggers ADL, it dispatches an authorized cryptographic call to `SettlementGuaranteeFund.sol` invoking `recordAdlTrigger(bytes32 symbolHash, uint256 deficitInr, uint256 timestamp)`.
- **Negative Balance Reconciliation:** All settled ADL matches record an on-chain ledger proof calling `reconcileNegativeEquity(bytes32 defaultId, bytes32 bankruptAccountHash, uint256 deficitCoveredInr, bytes32 txReceiptHash)`. This guarantees that negative equity was resolved without unbacked minting or spot ledger tampering.
- **Immutable Attestation Root:** Every 60 seconds (or immediately following an ADL execution batch), the service commits an Ed25519-signed Merkle root of the active ADL queue and matched deleveraging records to the Hyperledger Besu state.
- **Zero On-Chain PII Guarantee:** All blockchain transactions strictly reference pseudonymous cryptographic account hashes (`keccak256(user_id)`), instrument symbol hashes, and integer amounts in micro-INR (1 INR = 1,000,000 micro-INR). Personal investor identifiers, KYC details, and PAN numbers are never written to the blockchain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Workspace & Service Layout:** Scaffold the Rust crate `services/adl-engine` with a modular architecture: `src/ranking/` (lock-free ranking queue), `src/matcher/` (deterministic bankruptcy resolution), `src/sgf_monitor/` (insurance balance watcher), `src/events/` (Kafka producers/consumers), `src/blockchain/` (Besu relayer), and `src/grpc/` (RPC server).
2. **Define Protobuf Specifications:** Create `proto/growww/adl/v1/adl_service.proto` defining endpoints for `GetAdlPriorityStatus`, `StreamAdlIndicator`, `SimulateAdlExposure`, and Kafka event schemas for bankruptcy and execution reports.
3. **Generate Rust Stubs:** Compile protobuf files using `tonic-build` and `prost` inside `services/adl-engine/build.rs`.
4. **Author PostgreSQL DDL Migrations:** Author SQL migration files setting up `adl_execution_events`, `adl_queue_snapshots`, `adl_contract_configurations`, and `adl_systemic_incident_reports`.
5. **Implement Mathematical ADL Ranking Model:** In `src/ranking/score.rs`, implement the standardized ranking algorithm:
   - Compute Unrealized Profit Ratio:
     $$\text{ProfitRatio}_{\text{Long}} = \frac{P_{\text{mark}} - P_{\text{entry}}}{P_{\text{entry}}}, \quad \text{ProfitRatio}_{\text{Short}} = \frac{P_{\text{entry}} - P_{\text{mark}}}{P_{\text{entry}}}$$
   - Filter positions to strictly profitable traders ($\text{ProfitRatio} > 0$). Unprofitable positions receive an ADL score of zero and are excluded from deleveraging.
   - Compute Effective Leverage:
     $$\text{EffectiveLeverage} = \frac{\text{Position Quantity} \times P_{\text{mark}}}{\text{Account Equity}}$$
   - Compute final ADL Score:
     $$\text{ADL Score} = \text{PercentileRank}(\text{ProfitRatio}) \times \text{EffectiveLeverage}$$
6. **Construct Lock-Free In-Memory Queue Architecture:** Implement `AdlQueueManager` using `crossbeam_skiplist::SkipMap` keyed by `(Reverse(OrderedF64(score)), AccountId)` partitioned by contract symbol and side (Long/Short). Guarantee sub-microsecond insertion, removal, and traversal.
7. **Build Real-Time 5-Bar Indicator Generator:** Implement a concurrent background task dividing the active profitable queue into 5 equal quintiles:
   - Quintile 1 ($0\% - 20\%$): 1 Bar (Low Risk)
   - Quintile 2 ($20\% - 40\%$): 2 Bars (Low-Moderate Risk)
   - Quintile 3 ($40\% - 60\%$): 3 Bars (Moderate Risk)
   - Quintile 4 ($60\% - 80\%$): 4 Bars (High Risk)
   - Quintile 5 ($80\% - 100\%$): 5 Bars (Critical ADL Risk)
   Publish live updates to Redis Sorted Sets and stream to frontends via WebSocket message bus.
8. **Implement SGF Insurance Fund Depletion Watcher:** Build a high-availability listener in `src/sgf_monitor/` tracking real-time liquidity on Kafka topic `sgf.pool_balance.v1`. Maintain an atomic flag `IS_ADL_ACTIVE`. When available SGF liquid reserves reach 0 INR and unhedged bankruptcy deficits occur, instantly activate the ADL matching pipeline.
9. **Build Bankruptcy Price ($P_{bk}$) Calculation Engine:** Implement precise rational arithmetic computing the exact price where the bankrupt trader's remaining balance is zero:
   - For Long position: $P_{bk} = P_{\text{entry}} \times \left(1 - \frac{\text{Initial Margin}}{\text{Position Value}}\right)$
   - For Short position: $P_{bk} = P_{\text{entry}} \times \left(1 + \frac{\text{Initial Margin}}{\text{Position Value}}\right)$
10. **Implement Deterministic Bankruptcy Matching Logic:** In `src/matcher/engine.rs`, when a bankrupt liquidation event arrives:
    - Take the bankrupt position's remaining unsatisfied volume $Q_{\text{deficit}}$.
    - Pop top-priority counterparties from the opposing ADL SkipMap.
    - Match quantity $\Delta Q = \min(Q_{\text{counterparty}}, Q_{\text{remaining}})$.
    - Generate fill pairs at $P_{bk}$ with execution code `ERR_ADL_EXECUTED`.
    - Repeat until $Q_{\text{deficit}} == 0$.
11. **Construct Low-Latency Dispatcher to Matching Engine:** Format match instructions into zero-copy binary packets and transmit over high-speed IPC/Unix Domain Socket to Matching Engine (Prompt 205) to execute position adjustments atomically.
12. **Implement Hyperledger Besu Blockchain Relayer:** Build `src/blockchain/relayer.rs` using `ethers-rs` to submit gas-sponsored Besu transactions recording the default ID, bankrupt account hash, deficit covered, and counterparty closeout count to `SettlementGuaranteeFund.sol`.
13. **Build Notification & Regulatory Dispatcher:** Stream detailed liquidation events to Kafka topic `adl.executions.v1` and generate automated SEBI Systemic Incident reports to `compliance.sebi_alerts.v1`.
14. **Instrument Comprehensive Metrics & Tracing:** Expose Prometheus metrics: `adl_queue_depth_total`, `adl_score_recomputation_latency_nanos`, `adl_executions_total`, `adl_volume_deleveraged_notional_inr`, and `sgf_exhaustion_events_total`.
15. **Formulate End-to-End Stress Test Suite:** Write automated integration tests simulating:
    - Flash crash of 40% in underlying index within 500ms.
    - SGF balance drawdown to zero.
    - Deterministic ADL triggering across top 50 profitable accounts.
    - Strict verification of the Zero Spot Loss Invariant and exact balance sheet equality.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/adl/v1/adl_service.proto`)
```protobuf
syntax = "proto3";

package growww.adl.v1;

option go_package = "growww/adl/v1;adlv1";

// Service managing real-time Auto-Deleveraging priority ranking and execution
service AdlService {
  // Query the current ADL indicator and queue priority for a specific account and contract
  rpc GetAdlPriorityStatus (AdlPriorityStatusRequest) returns (AdlPriorityStatusResponse);

  // Stream real-time ADL 5-bar indicator updates for an active trader session
  rpc StreamAdlIndicator (StreamAdlIndicatorRequest) returns (stream AdlIndicatorUpdate);

  // Simulate what-if ADL exposure if price moves by delta or rank increases
  rpc SimulateAdlExposure (SimulateAdlExposureRequest) returns (SimulateAdlExposureResponse);

  // Administrative endpoint to inspect the top N queue candidates for an instrument
  rpc GetTopAdlCandidates (GetTopAdlCandidatesRequest) returns (GetTopAdlCandidatesResponse);
}

enum PositionSide {
  POSITION_SIDE_UNSPECIFIED = 0;
  POSITION_SIDE_LONG = 1;
  POSITION_SIDE_SHORT = 2;
}

enum AdlRiskBarLevel {
  ADL_RISK_BAR_LEVEL_UNSPECIFIED = 0;
  ADL_RISK_BAR_LEVEL_ONE = 1;    // 0% - 20% (Lowest Priority)
  ADL_RISK_BAR_LEVEL_TWO = 2;    // 20% - 40%
  ADL_RISK_BAR_LEVEL_THREE = 3;  // 40% - 60%
  ADL_RISK_BAR_LEVEL_FOUR = 4;   // 60% - 80%
  ADL_RISK_BAR_LEVEL_FIVE = 5;   // 80% - 100% (Immediate Deleveraging Candidate)
}

message AdlPriorityStatusRequest {
  string user_id = 1;
  string symbol = 2;
}

message AdlPriorityStatusResponse {
  string user_id = 1;
  string symbol = 2;
  PositionSide side = 3;
  string position_quantity = 4;
  string entry_price = 5;
  string mark_price = 6;
  string unrealized_pnl_inr = 7;
  string effective_leverage = 8;
  string profit_percentile = 9;
  string adl_score = 10;
  AdlRiskBarLevel risk_bar_level = 11;
  uint32 rank_in_queue = 12;
  uint32 total_profitable_counterparties = 13;
  int64 timestamp_ns = 14;
}

message StreamAdlIndicatorRequest {
  string user_id = 1;
  repeated string symbols = 2;
}

message AdlIndicatorUpdate {
  string user_id = 1;
  string symbol = 2;
  AdlRiskBarLevel risk_bar_level = 3;
  string adl_score = 4;
  uint32 rank_in_queue = 5;
  int64 updated_at_ns = 6;
}

message SimulateAdlExposureRequest {
  string user_id = 1;
  string symbol = 2;
  string simulated_mark_price = 3;
}

message SimulateAdlExposureResponse {
  string simulated_adl_score = 1;
  AdlRiskBarLevel simulated_risk_bar_level = 2;
  string estimated_bankruptcy_pnl_inr = 3;
}

message GetTopAdlCandidatesRequest {
  string symbol = 1;
  PositionSide side = 2;
  uint32 limit = 3;
}

message AdlCandidateSummary {
  string account_hash = 1;
  string position_quantity = 2;
  string effective_leverage = 3;
  string profit_percentile = 4;
  string adl_score = 5;
  uint32 rank = 6;
}

message GetTopAdlCandidatesResponse {
  string symbol = 1;
  PositionSide side = 2;
  repeated AdlCandidateSummary candidates = 3;
  int64 snapshot_timestamp_ns = 4;
}
```

### Kafka Event Schemas

#### 1. Inbound Unfilled Liquidation Bankruptcy Event (`growww.risk.bankruptcy.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "UnfilledLiquidationBankruptcyEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "liquidation_id": { "type": "string", "format": "uuid" },
    "bankrupt_user_id": { "type": "string", "format": "uuid" },
    "symbol": { "type": "string" },
    "position_side": { "type": "string", "enum": ["LONG", "SHORT"] },
    "unfilled_quantity": { "type": "string" },
    "bankruptcy_price": { "type": "string" },
    "maintenance_margin_deficit_inr": { "type": "string" },
    "sgf_available_balance_inr": { "type": "string" },
    "timestamp_ns": { "type": "integer" }
  },
  "required": [
    "event_id",
    "liquidation_id",
    "bankrupt_user_id",
    "symbol",
    "position_side",
    "unfilled_quantity",
    "bankruptcy_price",
    "sgf_available_balance_inr",
    "timestamp_ns"
  ]
}
```

#### 2. Outbound ADL Execution Event (`growww.adl.executions.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AdlExecutionEvent",
  "type": "object",
  "properties": {
    "adl_execution_id": { "type": "string", "format": "uuid" },
    "liquidation_id": { "type": "string", "format": "uuid" },
    "symbol": { "type": "string" },
    "bankrupt_user_id": { "type": "string", "format": "uuid" },
    "counterparty_user_id": { "type": "string", "format": "uuid" },
    "counterparty_side": { "type": "string", "enum": ["LONG", "SHORT"] },
    "executed_quantity": { "type": "string" },
    "bankruptcy_price": { "type": "string" },
    "counterparty_adl_score": { "type": "string" },
    "counterparty_rank_at_execution": { "type": "integer" },
    "freed_margin_inr": { "type": "string" },
    "realized_pnl_inr": { "type": "string" },
    "matching_engine_tx_id": { "type": "string" },
    "reason_code": { "type": "string", "enum": ["ERR_ADL_EXECUTED"] },
    "timestamp_ns": { "type": "integer" }
  },
  "required": [
    "adl_execution_id",
    "liquidation_id",
    "symbol",
    "bankrupt_user_id",
    "counterparty_user_id",
    "counterparty_side",
    "executed_quantity",
    "bankruptcy_price",
    "reason_code",
    "timestamp_ns"
  ]
}
```

### PostgreSQL DDL Schema (`services/adl-engine/migrations/001_initial_adl_schema.sql`)
```sql
CREATE TYPE adl_position_side_enum AS ENUM ('LONG', 'SHORT');
CREATE TYPE adl_risk_tier_enum AS ENUM ('ONE_BAR', 'TWO_BAR', 'THREE_BAR', 'FOUR_BAR', 'FIVE_BAR');
CREATE TYPE adl_execution_status_enum AS ENUM ('SUBMITTED', 'MATCHED_MATCHING_ENGINE', 'FAILED_ME', 'ON_CHAIN_RECONCILED');

-- Configuration per derivative symbol for ADL parameters
CREATE TABLE adl_contract_configurations (
    symbol VARCHAR(32) PRIMARY KEY,
    is_adl_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    min_leverage_threshold NUMERIC(6, 2) NOT NULL DEFAULT 1.00,
    max_single_counterparty_deleveraging_pct NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    circuit_breaker_cooloff_seconds INT NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Real-time snapshots of ADL rankings for legal audit and compliance
CREATE TABLE adl_queue_snapshots (
    snapshot_id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL REFERENCES adl_contract_configurations(symbol),
    side adl_position_side_enum NOT NULL,
    user_id UUID NOT NULL,
    rank_in_queue INT NOT NULL,
    profit_percentile NUMERIC(8, 6) NOT NULL,
    effective_leverage NUMERIC(8, 4) NOT NULL,
    adl_score NUMERIC(12, 6) NOT NULL,
    risk_tier adl_risk_tier_enum NOT NULL,
    position_quantity NUMERIC(24, 8) NOT NULL,
    mark_price NUMERIC(18, 4) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Immutable record of executed auto-deleveraging matches
CREATE TABLE adl_execution_events (
    adl_execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    liquidation_id UUID NOT NULL,
    symbol VARCHAR(32) NOT NULL REFERENCES adl_contract_configurations(symbol),
    bankrupt_user_id UUID NOT NULL,
    counterparty_user_id UUID NOT NULL,
    counterparty_side adl_position_side_enum NOT NULL,
    executed_quantity NUMERIC(24, 8) NOT NULL,
    bankruptcy_price NUMERIC(18, 4) NOT NULL,
    counterparty_adl_score NUMERIC(12, 6) NOT NULL,
    counterparty_rank_at_execution INT NOT NULL,
    freed_margin_inr NUMERIC(18, 4) NOT NULL,
    realized_pnl_inr NUMERIC(18, 4) NOT NULL,
    matching_engine_order_id VARCHAR(64) NOT NULL UNIQUE,
    status adl_execution_status_enum NOT NULL DEFAULT 'SUBMITTED',
    
    -- Blockchain proof & audit trail
    besu_tx_hash VARCHAR(66),
    besu_block_number BIGINT,
    reconciled_on_chain_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Formal incident ledger for SEBI systemic risk reports
CREATE TABLE adl_systemic_incident_reports (
    incident_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(32) NOT NULL,
    trigger_timestamp TIMESTAMPTZ NOT NULL,
    total_bankrupt_accounts INT NOT NULL,
    total_deficit_inr NUMERIC(18, 4) NOT NULL,
    sgf_buffer_drawn_inr NUMERIC(18, 4) NOT NULL,
    total_adl_volume_deleveraged NUMERIC(24, 8) NOT NULL,
    total_counterparties_affected INT NOT NULL,
    sebi_incident_ref_number VARCHAR(64) NOT NULL UNIQUE,
    cro_signoff_user_id UUID,
    report_dispatched_to_regulator_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_adl_queue_symbol_side ON adl_queue_snapshots(symbol, side, recorded_at DESC);
CREATE INDEX idx_adl_execution_counterparty ON adl_execution_events(counterparty_user_id, created_at DESC);
CREATE INDEX idx_adl_execution_symbol ON adl_execution_events(symbol, created_at DESC);
```

## Security & Compliance Notes
- **Zero Socialized Loss Clawback Invariant:** Under SEBI CCP norms and the NBSE Risk Charter, losses incurred on leveraged derivatives books can NEVER be socialized across cash spot traders, unleveraged investors, or unrelated contract pools. Deleveraging is strictly confined to profitable opposing traders within the specific distressed contract symbol.
- **Fair and Transparent 5-Bar Queue Indicator:** To prevent investor panic and uphold regulatory transparency, all derivative accounts maintain a real-time 5-bar visual indicator. Traders have full real-time awareness of their queue priority and can voluntarily lower leverage or close positions to decrease ADL probability.
- **Deterministic Math & Non-Discretionary Matching:** ADL priority rankings and volume matches are derived purely by the deterministic scoring formula. No operator intervention, discretionary sorting, or front-running of VIP market makers is technically possible.
- **Mandatory SEBI Systemic Incident Notifications:** Any invocation of the ADL Engine constitutes a Tier-1 CCP Systemic Risk Incident. The system automatically drafts a standardized regulatory incident payload dispatched to the SEBI Market Surveillance and Risk Division within 2 hours of invocation.
- **Strict Execution at Bankruptcy Price ($P_{bk}$):** Counterparty positions are closed strictly at $P_{bk}$. The counterparty retains all profits earned up to $P_{bk}$ and pays zero penalty fees. The bankrupt trader exits with exactly zero equity.
- **Zero On-Chain PII Guarantee:** All state submissions to Hyperledger Besu contain strictly SHA-256 / Keccak-256 hashes of account identifiers, symbol tickers, integer satoshis/micro-INR, and execution hashes.

## Acceptance Criteria
- [ ] ADL priority scoring algorithm computes updated rankings for 50,000 active derivative positions across 50 contract symbols in $< 10\text{ms}$ upon Mark Price tick arrival.
- [ ] Lock-free `SkipMap` provides thread-safe $O(\log N)$ position updates without acquiring mutexes on the critical path.
- [ ] 5-Bar indicator accurately partitions profitable traders into five equal 20% risk quintiles and streams updates via Redis within 50ms of rank change.
- [ ] SGF depletion monitor switches state from `SGF_SOLVENT` to `ADL_ACTIVE` within 500 microseconds of receiving an unbacked deficit notification.
- [ ] Deterministic bankruptcy matcher correctly satisfies 100% of the bankrupt volume by slicing across top-ranked ADL counterparties at exact $P_{bk}$.
- [ ] Generated execution instructions successfully pass to the Matching Engine with reason code `ERR_ADL_EXECUTED` without order book disruption.
- [ ] Unleveraged spot accounts and traders with zero or negative unrealized PnL are strictly excluded from the ADL candidate queue ($100\%$ validation).
- [ ] Cryptographic reconciliation proofs post reliably to `SettlementGuaranteeFund.sol` on Hyperledger Besu with verified transaction receipts.
- [ ] System recovery from crash maintains consistent ADL priority queues by rehydrating state from Redis and Kafka without missed deleveraging matches.
- [ ] Standardized SEBI systemic incident report generates with precise mathematical deficit metrics and CRO digital audit signatures.

## Suggested Order / Dependencies
- **Prerequisites:**
  - 103 (API Design Standards)
  - 104 (Event Schema and Kafka Topic Standards)
  - 205 (High-Performance Matching Engine in Rust)
  - 206 (Pre-Trade Risk Engine)
  - 207 (Market Data Service)
  - 229 (Real-Time VaR & ELM Engine)
  - 230 (Settlement Guarantee Fund & Default Waterfall Service)
  - ADR-0029 (Forced Liquidation 1.0% Penalty Fee & Insurance Fund Capitalization)
  - RUNBOOK-19 (Forced Liquidation & Insurance Fund Deficit Injection)
  - RUNBOOK-25 (Auto-Deleveraging Protocol Execution & Triage)
- **Parallel Tasks:**
  - 315 (Settlement Guarantee Fund Smart Contract on Besu)
  - 326 (Perpetual Futures Clearing Smart Contract)
- **Downstream Blockers:**
  - 529 (Mobile Options Chain & Derivatives Screen 5-Bar ADL UI)
  - 604 (Web Institutional Risk & Collateral Desk)
