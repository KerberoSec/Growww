# 257 - Institutional Algorithmic Execution Engine (TWAP, VWAP, POV & Iceberg) (Rust)

## Purpose
Institutional asset managers, sovereign wealth funds, family offices, and authorized participants handling large-block transactions in tokenized real-world assets (RWAs), sovereign debt, and equities face severe execution penalties when placing wholesale parent orders directly into lit continuous order books. Naive large-block routing creates devastating market impact, informational leakage, front-running by predatory High-Frequency Trading (HFT) sniffing algorithms, and steep implementation shortfall. Furthermore, executing across 24/7 continuous RWA trading sessions alongside legacy primary exchange call auctions (such as NSE/BSE pre-open and closing sessions) demands algorithmic precision capable of transitioning between continuous volume curves and discrete auction liquidity pools.

The **Institutional Algorithmic Execution Engine** (`services/algo-execution-engine`) is an ultra-low latency, deterministic execution microservice engineered in Rust 2021. It decomposes massive parent orders into intelligently timed, sized, and priced micro-slices. The engine deploys five primary execution strategies: Time-Weighted Average Price (TWAP) with randomized Poisson intervals, Volume-Weighted Average Price (VWAP) driven by SIMD-accelerated historical volume curve interpolation, Percentage of Volume (POV) dynamic participation tracking real-time tape prints, synthetic Iceberg orders with randomized variance, and Almgren-Chriss optimal Implementation Shortfall (IS) trajectory balancing inventory risk against market impact. Operating under strict SEBI algorithmic trading regulatory mandates, the engine enforces pre-trade risk filters, Order-to-Trade Ratio (OTR) limits, and anti-spoofing constraints while collecting the platform invariant flat 0.00% transaction fee (No fee at all) (Platform Treasury, Core SGF, and Investor Protection Fund per FeeController governance).

## What You Are Building
A mission-critical, ultra-low latency Rust service (`services/algo-execution-engine`) designed for deterministic institutional order slicing and schedule dispatch. Key subsystems include:
- **Parent Order Lifecycle Manager:** Coordinates institutional parent order intake, schedule generation, state persistence, slice reconciliation, and terminal state transitions (Completed, Paused, Canceled, Expired).
- **SIMD Volume Curve Interpolator:** Ingests intraday tick distributions from TimescaleDB and ClickHouse; computes smooth cubic spline volume profiles across historical buckets using AVX2/AVX-512 SIMD vectorization to forecast intraday volume density $v(t)$.
- **Randomized TWAP Scheduler:** Slices parent volume into randomized sub-lots dispatched at non-deterministic Poisson intervals ($t_k = t_{k-1} + \Delta t + \xi$, where $\xi \sim \text{Poisson}(\lambda)$) to neutralize predatory HFT signature detection.
- **Dynamic VWAP Trajectory Tracker:** Computes real-time dynamic volume participation targets, adjusting child order aggression based on observed cumulative market volume versus historical expectations.
- **POV (Percentage of Volume) Real-Time Tape Tracker:** Ingests live trade match feeds from the Order Matching Engine (Prompt 205); dynamically calculates target participation rate $\alpha$ (e.g., 5% to 15% of tape volume) and dispatches child liquidity without exceeding the configured market footprint.
- **Iceberg Replenishment Controller:** Dispatches visible "tip" quantities with randomized display lot variance (+/- 10% to 25%) and hidden balance reserves; immediately reloads child slices upon fill confirmation via lock-free ring buffers.
- **Almgren-Chriss Implementation Shortfall Optimizer:** Solves optimal trading trajectories minimizing execution price drift relative to arrival benchmark price, balancing risk-aversion coefficient $\lambda$ against temporary and permanent market impact functions.
- **SEBI Algorithmic Pre-Trade Risk & Anti-Gaming Guard:** Enforces hard price bands, Order-to-Trade Ratio (OTR) throttling, single-slice value ceilings, and runaway algo kill switches to prevent spoofing, layering, and quote stuffing.
- **Zero-PII On-Chain Aggregator:** Bundles multiple executed child slices into aggregated execution receipts for atomic DvP settlement on Hyperledger Besu (`SettlementDvP.sol`) using blinded cryptographic execution tokens.
- **Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically logs and assesses the 0.00% (Zero Fee) platform fee across aggregated turnover, partitioned into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.

## Scope Boundaries
- **In Scope:**
  - Parent order intake, validation, strategy compilation, and dynamic slicing (TWAP, VWAP, POV, Iceberg, Implementation Shortfall).
  - High-performance SIMD historical volume curve calibration and real-time intraday progression tracking.
  - Non-deterministic microsecond slice scheduling with Poisson and Gaussian timing jitter to defeat HFT sniffing.
  - Ingestion of live BBO quotes and market trade prints from `market.ticker.v1` and `matching.trades.v1`.
  - Dispatching native child Limit and Immediate-Or-Cancel (IOC) orders to the Order Matching Engine (Prompt 205).
  - Child order fill reconciliation, cancellation routing, and dynamic trajectory recalibration.
  - SEBI algorithmic pre-trade risk filter validation, OTR surveillance, and emergency kill switches.
  - Aggregation of child fills for DvP settlement attestation on Hyperledger Besu.
  - 0.00% (No fee at all) platform fee assessment and 0.00% fee launch policy distribution ledgering.
- **Out of Scope / Handled Elsewhere:**
  - In-memory Limit Order Book matching and trade execution (handled in Prompt 205 Order Matching Engine).
  - Cash wallet balance reservation and multi-currency ledgering (handled in Prompt 203 Wallet Account Service).
  - Regulatory CKYC/KRA investor verification (handled in Prompt 202 KYC/AML Service).
  - FIX / OUCH raw protocol wire termination (handled in Prompt 225 FIX Protocol Gateway).
  - Dynamic LULD volatility band enforcement and call auction halts (handled in Prompt 256 LULD Volatility Dampener Service).
  - Retail order routing and web/mobile UI rendering (handled in Prompts 509 and 603).

## Technology to Use
- **Primary Language & Runtime:** **Rust 2021 edition (Rust 1.78+)** utilizing `tokio` multi-threaded asynchronous runtime for network I/O and dedicated pinned worker threads for deterministic math execution.
  *Justification:* Zero-cost abstractions, zero garbage collection pauses, memory safety without runtime overhead, and explicit SIMD hardware vectorization.
- **Lock-Free Concurrency Primitives:** `crossbeam-channel`, `parking_lot`, and custom cache-line padded lock-free ring buffers (`rtrb` / `disruptor-rs`) for zero-allocation inter-thread slice dispatching.
- **SIMD Hardware Acceleration:** `packed_simd_2` / `std::simd` with AVX2 and AVX-512 target features for parallel historical volume bucket interpolation and matrix math.
- **Numerical & Optimization Math:** `nalgebra` and `statrs` for probability distributions (Poisson, Gaussian) and matrix solvers for Almgren-Chriss trajectories.
- **Event Streaming & Messaging:** `rdkafka` (Rust wrapper for `librdkafka`) for ultra-low latency Kafka consumption and production with zero-copy deserialization.
- **In-Memory Caches & Profiles:** **Redis 7.2+ Cluster** with pipelined Redis strings and hashes for active parent state, runtime OTR counters, and pre-computed volume curves.
- **Time-Series & Volume Distribution Storage:** **TimescaleDB / ClickHouse** for querying granular historical 1-minute and 5-minute intraday volume curves across trading days.
- **Transactional Persistence:** **PostgreSQL 16+** with `sqlx` (asynchronous compile-time checked SQL) for persistent parent orders, slice execution history, and fee ledgers.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** (`tonic` and `prost`) for low-latency algorithmic order submission, state queries, and slice execution events.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `algo_parent_orders`, `algo_child_slices`, `algo_execution_benchmarks`, `algo_risk_parameters`, `algo_fee_ledgers`.
- **TimescaleDB / ClickHouse Tables:** `intraday_volume_profiles_1m`, `historical_turnover_distributions`, `tick_volatility_metrics`.
- **Redis 7.2 Keys:**
  - `algo:parent:{parent_order_id}`: Active parent order runtime state, filled quantity, and remaining balance.
  - `algo:otr:counter:{user_id}:{minute}`: Rolling 60-second Order-to-Trade Ratio sliding window.
  - `algo:curve:{isin}:{day_of_week}`: Normalized 390-minute intraday cumulative volume distribution curve.
  - `algo:kill_switch:global` and `algo:kill_switch:{user_id}`: Circuit breakers for immediate algo cessation.
- **Kafka Topics:**
  - Consumes: `market.ticker.v1`, `matching.trades.v1`, `market.luld.state_changed.v1`, `risk.limits_updated.v1`.
  - Publishes: `algo.orders.submitted.v1`, `order.matching.commands.v1`, `algo.slices.dispatched.v1`, `algo.executions.aggregated.v1`, `algo.fee_assessed.v1`.
- **Order Matching Engine (Prompt 205):** Ingests child orders (`order.matching.commands.v1`) and emits execution trade reports.
- **Pre-Trade Risk Engine (Prompt 206):** Validates aggregate parent order buying power before strategy instantiation.
- **Market Data Service (Prompt 207):** Ingests live consolidated tape and top-of-book depth for dynamic POV adjustments.
- **Settlement DvP Service (Prompt 208):** Ingests aggregated multi-fill execution vouchers for on-chain settlement.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Aggregated Multi-Fill Batch Settlement:** To minimize on-chain transaction bloat and prevent gas exhaustion, hundreds of micro-slices executed by a single parent order are deterministically aggregated into a single signed settlement voucher:
  $$\text{SettlementVoucher} = \mathcal{H}\Big(\text{ParentOrderID} \parallel \text{ISIN} \parallel \sum Q_{\text{slice}} \parallel \bar{P}_{\text{VWAP}} \parallel \text{Nonce} \parallel \text{Timestamp}\Big)$$
- **DvP Smart Contract Clearing:** The aggregated voucher is notarized and submitted to `SettlementDvP.sol` on Hyperledger Besu under QBFT consensus (2.0s block time, single-block finality).
- **Zero-PII Execution Tokens:** Institutional clients execute algorithmic orders under rotating cryptographic pseudonym tokens (`0x...` ephemeral execution addresses). All Kafka events, Protobuf messages, and on-chain logs reference strictly the token hash, ensuring zero personally identifiable information (PII) is exposed to public or node operators, complying with DPDP Act 2023 and SEBI mandates.
- **Cryptographic Audit Attestation:** The execution benchmark report (Arrival Price, Realized VWAP, Slippage, Participation Curve) is hashed into a Merkle tree leaf and anchored to `AlgoAuditRegistry.sol` for tamper-proof SEBI regulatory verification.

## Algorithmic Execution Mechanics & Mathematical Formulations

### 1. TWAP (Time-Weighted Average Price) with Non-Deterministic Poisson Slicing
Naive TWAP divides the total order quantity $Q_{\text{total}}$ evenly across $N$ intervals of fixed duration $\Delta t$, resulting in predictable execution footprints that predatory HFT algorithms exploit. The engine introduces stochastic Poisson interval jitter and Gaussian slice volume variance:
- Target slice count over total duration $T$:
  $$N = \left\lfloor \frac{T}{\Delta t_{\text{nominal}}} \right\rfloor$$
- Next slice dispatch timestamp $t_k$:
  $$t_k = t_{k-1} + \Delta t_{\text{nominal}} + \xi_k, \quad \xi_k \sim \text{Poisson}(\lambda) - \lambda$$
- Slice volume $q_k$ with randomized variance $\sigma_v$:
  $$q_k = \frac{Q_{\text{remaining}}}{N - k + 1} \times (1 + \epsilon_k), \quad \epsilon_k \sim \mathcal{N}(0, \sigma_v^2), \quad \text{subject to } \sum_{i=1}^N q_i = Q_{\text{total}}$$

### 2. VWAP (Volume-Weighted Average Price) with SIMD Volume Curve Interpolation
VWAP aims to match or beat the market volume-weighted average price across the execution horizon by weighting slice sizes according to the expected intraday volume curve $V(t)$:
- Let $u_m \in [0, 1]$ be the historical fraction of daily volume traded up to minute $m \in [1, 390]$:
  $$u_m = \frac{\sum_{i=1}^m \bar{v}_i}{V_{\text{day}}}$$
- The engine interpolates the continuous cumulative volume curve $U(t)$ using SIMD-vectorized cubic Hermite splines across 5-minute historical nodes.
- For an execution window $[t_{\text{start}}, t_{\text{end}}]$, the scheduled child slice volume $q_k$ for interval $[t_{k-1}, t_k]$ is:
  $$q_k = Q_{\text{total}} \times \left( \frac{U(t_k) - U(t_{k-1})}{U(t_{\text{end}}) - U(t_{\text{start}})} \right)$$
- **Dynamic Real-Time Tracking Correction:** The engine monitors actual cumulative market volume $V_{\text{actual}}(t)$ against expected volume $V_{\text{expected}}(t)$. If the market is running ahead of historical volume, the participation pace is accelerated:
  $$\beta(t) = \frac{V_{\text{actual}}(t)}{V_{\text{expected}}(t)}, \quad q_k^{\text{adjusted}} = q_k \times \beta(t_k)^{\gamma}$$
  where $\gamma \in [0.5, 1.0]$ is an aggression tuning parameter.

### 3. Percentage of Volume (POV) Dynamic Tape Tracking
POV executes child orders in direct proportion to real-time market turnover without a fixed end time:
- Configured participation rate: $\alpha \in (0, 0.30]$ (e.g., 10% of the tape).
- Let $\Delta V_{\text{market}}(t_{k-1}, t_k)$ be the lit market volume executed across all exchange participants during the observation window.
- The engine dispatches child order size:
  $$q_k = \min\left( Q_{\text{remaining}}, \; \frac{\alpha}{1 - \alpha} \times \Delta V_{\text{market}}(t_{k-1}, t_k) \right)$$
- If market liquidity dries up, the engine automatically throttles down child order injection to eliminate artificial price inflation.

### 4. Synthetic Iceberg Replenishment Controller
Large limit orders are posted to the book displaying only a small visible fraction ("tip"), keeping the bulk hidden in synthetic reserve:
- Display quantity: $Q_{\text{visible}} = Q_{\text{nominal\_tip}} \times (1 + \delta_k)$, where $\delta_k \sim \text{Uniform}(-\delta_{\text{max}}, +\delta_{\text{max}})$, typically $\delta_{\text{max}} = 0.20$ (20% display jitter).
- When a child slice fill event is consumed from Kafka (`matching.trades.v1`), the replenishment controller generates the subsequent slice immediately via lock-free ring buffer:
  $$Q_{\text{next}} = \min\left( Q_{\text{visible}}, \; Q_{\text{hidden\_remaining}} \right)$$
- Hidden reserve is held strictly off-book in in-memory memory buffers, preventing depth-of-book sniffing.

### 5. Almgren-Chriss Optimal Implementation Shortfall Trajectory
Balances market impact against price volatility risk over trading duration $T = M \tau$:
- Total holdings trajectory $x_k$ at time step $k$:
  $$x_k = \frac{\sinh(\kappa (T - t_k))}{\sinh(\kappa T)} X_0$$
  where $X_0$ is total initial order size, and $\kappa$ is the velocity parameter:
  $$\kappa \approx \sqrt{\frac{\lambda \sigma^2}{\eta}}$$
  with $\lambda$ being institutional risk aversion, $\sigma$ intraday asset volatility, and $\eta$ temporary market impact parameter.
- Slices $n_k = x_{k-1} - x_k$ are calculated and dispatched adaptively based on real-time price deviation from Arrival Price $P_0$.

### 6. Universal 0.00% (No fee at all) Platform Fee Model
Every executed child slice and aggregated batch turnover incurs the non-negotiable 0.00% (Zero Fee) platform fee:
$$\text{Gross Turnover} = \sum_{i=1}^M (P_i \times Q_i)$$
$$\text{Total Platform Fee} = \text{Gross Turnover} \times 0.0000 = 0$$
$$\text{Platform Treasury (60%)} = \text{Total Platform Fee} \times 0.60$$
$$\text{Core Settlement Guarantee Fund (25%)} = \text{Total Platform Fee} \times 0.25$$
$$\text{Investor Protection Fund (15%)} = \text{Total Platform Fee} \times 0.15$$

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Infrastructure:** Create Rust 2021 crate `services/algo-execution-engine` with modular directory structure (`src/domain`, `src/engine`, `src/simd`, `src/risk`, `src/grpc`, `src/kafka`, `src/storage`). Configure Cargo workspace dependencies (`tokio`, `tonic`, `rdkafka`, `sqlx`, `crossbeam`, `nalgebra`).
2. **Define Protobuf API Contracts:** Author `proto/growww/algo/v1/algo_execution_engine.proto` specifying gRPC services for parent order intake, slice cancellation, state streaming, and execution analytics. Compile Rust bindings via `tonic-build`.
3. **Design PostgreSQL & TimescaleDB Schemas:** Create migration scripts for relational parent order storage, slice audit records, and TimescaleDB 1-minute volume profile hyper-tables. Implement composite indexing on `(user_id, status)` and `(isin, timestamp)`.
4. **Implement SIMD Historical Volume Curve Module:** Write AVX2/AVX-512 SIMD kernels in `src/simd/volume_spline.rs` to compute cubic spline interpolation across 390-minute historical trading bins in under 200 nanoseconds.
5. **Implement Pre-Trade SEBI Risk Filter:** Build the risk validation module in `src/risk/pre_trade.rs`. Verify maximum order value, price collar limits (+/- 5% of NBBO), maximum participation caps, and rolling 60-second Order-to-Trade Ratio (OTR <= 50:1).
6. **Construct Lock-Free Dispatch Core:** Implement cache-line aligned lock-free ring buffers (`src/engine/ring_buffer.rs`) for passing slice instructions between strategy calculation threads and network dispatch worker pools without lock contention.
7. **Implement TWAP Execution Engine:** Build `src/engine/twap.rs` incorporating Poisson interval scheduling and Gaussian volume randomization. Implement dynamic clock timers using `tokio::time::sleep_until` aligned to monotonic hardware clocks.
8. **Implement VWAP Execution Engine:** Build `src/engine/vwap.rs` integrating SIMD historical volume profiles with real-time participation pace correction factors ($\beta(t)$) reacting to live market volume updates.
9. **Implement POV (Percentage of Volume) Engine:** Build `src/engine/pov.rs` tracking real-time market trade volume from `matching.trades.v1`. Compute proportional child slice sizes with upper limit caps to prevent market chasing.
10. **Implement Iceberg Replenishment Controller:** Build `src/engine/iceberg.rs` with randomized display lot sizing (+/- 20%) and immediate sub-millisecond child order replenishment upon fill receipt.
11. **Implement Almgren-Chriss Implementation Shortfall Solver:** Build `src/engine/implementation_shortfall.rs` using `nalgebra` matrix algebra to compute optimal liquidation/accumulation trajectories based on user risk aversion.
12. **Build Kafka Stream Consumers & Producers:** Implement Kafka handlers in `src/kafka/` consuming `market.ticker.v1` and `matching.trades.v1`, while publishing child orders to `order.matching.commands.v1` with zero-copy serialization.
13. **Integrate Hyperledger Besu Voucher Notarization:** Implement cryptographic aggregation in `src/blockchain/voucher.rs` hashing completed parent order fills into an EIP-712 DvP settlement payload with zero-PII execution tokens.
14. **Enforce Universal 0.00% fee (No fee at all) Assessor:** Implement fee deduction and tripartite split calculation (0.00% fee at launch; future fee parameters governed by FeeController.sol) on every child fill, emitting `algo.fee_assessed.v1`.
15. **Develop Unit, SIMD, and Integration Test Suite:** Write property-based tests for volume curve interpolation, concurrency benchmarks asserting <15 microsecond internal scheduling latency, and integration tests with simulated order books.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/algo/v1/algo_execution_engine.proto`)
```protobuf
syntax = "proto3";

package growww.algo.v1;

option go_package = "growww/algo/v1;algov1";

service AlgoExecutionEngineService {
  rpc SubmitAlgoOrder (SubmitAlgoOrderRequest) returns (SubmitAlgoOrderResponse);
  rpc CancelAlgoOrder (CancelAlgoOrderRequest) returns (CancelAlgoOrderResponse);
  rpc PauseAlgoOrder (PauseAlgoOrderRequest) returns (PauseAlgoOrderResponse);
  rpc ResumeAlgoOrder (ResumeAlgoOrderRequest) returns (ResumeAlgoOrderResponse);
  rpc GetAlgoOrderStatus (GetAlgoOrderStatusRequest) returns (AlgoOrderStatusSnapshot);
  rpc StreamAlgoExecutionEvents (StreamAlgoExecutionEventsRequest) returns (stream AlgoExecutionEvent);
}

enum AlgoStrategyType {
  ALGO_STRATEGY_UNSPECIFIED = 0;
  ALGO_STRATEGY_TWAP = 1;
  ALGO_STRATEGY_VWAP = 2;
  ALGO_STRATEGY_POV = 3;
  ALGO_STRATEGY_ICEBERG = 4;
  ALGO_STRATEGY_IMPLEMENTATION_SHORTFALL = 5;
}

enum AlgoOrderStatus {
  ALGO_STATUS_UNSPECIFIED = 0;
  ALGO_STATUS_PENDING = 1;
  ALGO_STATUS_ACTIVE = 2;
  ALGO_STATUS_PAUSED = 3;
  ALGO_STATUS_COMPLETED = 4;
  ALGO_STATUS_CANCELLED = 5;
  ALGO_STATUS_REJECTED = 6;
  ALGO_STATUS_EXPIRED = 7;
}

message SubmitAlgoOrderRequest {
  string idempotency_key = 1;
  string execution_token = 2; // Blinded Zero-PII institutional token
  string isin = 3;
  string side = 4;            // "BUY" or "SELL"
  uint64 total_quantity = 5;  // Total units (with 6 decimal places implied, e.g. 100000000 = 100.000000)
  AlgoStrategyType strategy_type = 6;
  
  // Execution Schedule Limits
  int64 start_time_unix_ns = 7;
  int64 end_time_unix_ns = 8;
  uint64 limit_price_cap_paise = 9; // Worst-case limit price collar
  
  // TWAP Parameters
  int64 twap_interval_seconds = 10;
  bool enable_poisson_jitter = 11;
  
  // VWAP Parameters
  string historical_profile_tag = 12; // e.g. "DEFAULT_30D", "EARNINGS_DAY"
  double vwap_aggression_gamma = 13;  // Tuning coefficient (0.5 to 1.0)
  
  // POV Parameters
  double target_participation_rate = 14; // e.g. 0.10 for 10%
  double max_participation_rate = 15;    // Hard cap e.g. 0.15 for 15%
  
  // Iceberg Parameters
  uint64 visible_tip_quantity = 16;
  double display_variance_pct = 17;     // e.g. 0.20 for +/-20%
  
  // Implementation Shortfall Parameters
  double risk_aversion_lambda = 18;
  double expected_daily_volatility = 19;
}

message SubmitAlgoOrderResponse {
  string parent_order_id = 1;
  AlgoStrategyType strategy_type = 2;
  AlgoOrderStatus status = 3;
  int64 accepted_at_unix_ns = 4;
  string rejection_reason = 5;
}

message CancelAlgoOrderRequest {
  string parent_order_id = 1;
  string execution_token = 2;
  string reason = 3;
}

message CancelAlgoOrderResponse {
  string parent_order_id = 1;
  bool acknowledged = 2;
  AlgoOrderStatus final_status = 3;
}

message PauseAlgoOrderRequest {
  string parent_order_id = 1;
  string execution_token = 2;
}

message PauseAlgoOrderResponse {
  string parent_order_id = 1;
  bool is_paused = 2;
}

message ResumeAlgoOrderRequest {
  string parent_order_id = 1;
  string execution_token = 2;
}

message ResumeAlgoOrderResponse {
  string parent_order_id = 1;
  bool is_resumed = 2;
}

message GetAlgoOrderStatusRequest {
  string parent_order_id = 1;
  string execution_token = 2;
}

message AlgoOrderStatusSnapshot {
  string parent_order_id = 1;
  string isin = 2;
  string side = 3;
  AlgoStrategyType strategy_type = 4;
  AlgoOrderStatus status = 5;
  uint64 total_quantity = 6;
  uint64 executed_quantity = 7;
  uint64 remaining_quantity = 8;
  uint64 volume_weighted_average_price_paise = 9;
  uint64 arrival_price_paise = 10;
  int64 slippage_bps = 11;
  uint32 slices_dispatched_count = 12;
  uint32 slices_filled_count = 13;
  int64 updated_at_unix_ns = 14;
}

message StreamAlgoExecutionEventsRequest {
  string execution_token = 1;
  string parent_order_id = 2; // Optional, empty for all orders belonging to token
}

message AlgoExecutionEvent {
  string event_id = 1;
  string parent_order_id = 2;
  string slice_order_id = 3;
  string isin = 4;
  string side = 5;
  uint64 slice_quantity = 6;
  uint64 executed_price_paise = 7;
  int64 execution_timestamp_ns = 8;
  uint64 platform_fee_paise = 9;
  uint64 treasury_fee_paise = 10;
  uint64 sgf_fee_paise = 11;
  uint64 ipf_fee_paise = 12;
  string besu_settlement_voucher_hash = 13;
}
```

### 2. PostgreSQL & TimescaleDB Database Schema (`services/algo-execution-engine/migrations/001_algo_execution_schema.sql`)
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE algo_strategy_type_enum AS ENUM (
    'TWAP',
    'VWAP',
    'POV',
    'ICEBERG',
    'IMPLEMENTATION_SHORTFALL'
);

CREATE TYPE algo_order_status_enum AS ENUM (
    'PENDING',
    'ACTIVE',
    'PAUSED',
    'COMPLETED',
    'CANCELLED',
    'REJECTED',
    'EXPIRED'
);

-- Parent Algorithmic Orders Table
CREATE TABLE algo_parent_orders (
    parent_order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    execution_token VARCHAR(66) NOT NULL, -- Cryptographic pseudonym, zero PII
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    strategy_type algo_strategy_type_enum NOT NULL,
    status algo_order_status_enum NOT NULL DEFAULT 'PENDING',
    
    total_quantity BIGINT NOT NULL CHECK (total_quantity > 0),
    executed_quantity BIGINT NOT NULL DEFAULT 0,
    remaining_quantity BIGINT NOT NULL,
    
    limit_price_cap_paise BIGINT NOT NULL,
    arrival_price_paise BIGINT,
    vwap_achieved_paise BIGINT,
    
    strategy_parameters JSONB NOT NULL,
    
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Child Slice Orders Table
CREATE TABLE algo_child_slices (
    slice_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_order_id UUID NOT NULL REFERENCES algo_parent_orders(parent_order_id) ON DELETE CASCADE,
    matching_order_id UUID NOT NULL, -- Child order submitted to Order Matching Engine
    slice_sequence_num INT NOT NULL,
    target_quantity BIGINT NOT NULL,
    executed_quantity BIGINT NOT NULL DEFAULT 0,
    limit_price_paise BIGINT NOT NULL,
    executed_price_paise BIGINT,
    slice_status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED',
    scheduled_dispatch_time TIMESTAMPTZ NOT NULL,
    actual_dispatched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT
);

-- TimescaleDB Intraday Historical Volume Profiles Hypertable
CREATE TABLE intraday_volume_profiles_1m (
    isin VARCHAR(12) NOT NULL,
    bucket_minute TIMESTAMPTZ NOT NULL,
    day_of_week INT NOT NULL,
    average_volume BIGINT NOT NULL,
    median_volume BIGINT NOT NULL,
    volume_standard_deviation NUMERIC(18, 4),
    cumulative_volume_ratio NUMERIC(8, 6) NOT NULL,
    PRIMARY KEY (isin, bucket_minute)
);

-- Multi-Fill Aggregated DvP Vouchers Table
CREATE TABLE algo_settlement_vouchers (
    voucher_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_order_id UUID NOT NULL REFERENCES algo_parent_orders(parent_order_id),
    execution_token VARCHAR(66) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL,
    total_aggregated_quantity BIGINT NOT NULL,
    gross_turnover_paise BIGINT NOT NULL,
    effective_vwap_paise BIGINT NOT NULL,
    merkle_root_hash VARCHAR(66) NOT NULL,
    eip712_signature VARCHAR(132) NOT NULL,
    besu_tx_hash VARCHAR(66),
    settled_at TIMESTAMPTZ
);

-- 0.00% (No fee at all) Platform Fee Ledgers
CREATE TABLE algo_fee_ledgers (
    ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_order_id UUID NOT NULL REFERENCES algo_parent_orders(parent_order_id),
    slice_id UUID NOT NULL REFERENCES algo_child_slices(slice_id),
    gross_slice_turnover_paise BIGINT NOT NULL,
    platform_fee_paise BIGINT NOT NULL,   -- 0.00% (No fee at all)
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_algo_parent_active ON algo_parent_orders(execution_token, status) WHERE status IN ('ACTIVE', 'PENDING');
CREATE INDEX idx_algo_parent_isin ON algo_parent_orders(isin, status);
CREATE INDEX idx_algo_slices_parent ON algo_child_slices(parent_order_id, slice_sequence_num);
CREATE INDEX idx_algo_vouchers_parent ON algo_settlement_vouchers(parent_order_id);
CREATE INDEX idx_algo_fees_parent ON algo_fee_ledgers(parent_order_id);
```

## Security & Compliance Notes
- **SEBI Algorithmic Trading Compliance:** In strict accordance with SEBI Master Circular on Algorithmic Trading, every algorithmic strategy deployed in this engine incorporates:
  - System-level single-order quantity caps and value ceilings.
  - Mandatory pre-trade price collar validation preventing orders outside exchange operating bands (+/- 5% of NBBO).
  - Hard Order-to-Trade Ratio (OTR) limits enforced at 50:1 on a rolling 60-second window. Any trading algorithm exceeding OTR triggers automated throttling and alert notification to the Surveillance Engine (Prompt 228).
- **Runaway Algo Kill Switches:** A global and per-token panic kill switch instantly revokes all active parent orders and dispatches mass-cancellations to the Order Matching Engine (Prompt 205) within 2 milliseconds of invocation.
- **Spoofing, Layering & Quote Stuffing Prevention:** The engine validates that Iceberg and POV child slices reflect genuine economic intent:
  - Micro-slices are governed by minimum lot sizes (never below 1 token unit).
  - Artificial order cancellation loops designed to probe book depth without intention to trade are rejected by deterministic rate limiters.
- **Zero-PII Cryptographic Isolation:** To comply with the Digital Personal Data Protection (DPDP) Act 2023 and IFSCA data regulations, algorithmic trading orders use rotating zero-PII execution tokens (`0x...`). Real investor identities remain strictly sealed within the secure enclave of the KYC/AML Service (Prompt 202).
- **Audit Logging & Regulatory Replay:** All internal algorithmic decisions (scheduled vs actual slice times, volume curve parameters, pricing offsets, slippage benchmarks) are preserved immutably with microsecond hardware timestamps for 8-year statutory SEBI regulatory replay.

## Acceptance Criteria
- [ ] Protobuf service contracts compile cleanly with `tonic-build` producing typed Rust stubs with zero compiler warnings.
- [ ] AVX2/AVX-512 SIMD cubic spline kernel interpolates 390-minute volume curves in under 200 nanoseconds per evaluation.
- [ ] TWAP strategy generates non-deterministic slice dispatch schedules adhering to Poisson interval jitter ($\lambda$) without periodic signature repetition.
- [ ] VWAP strategy dynamically adjusts child slice volume in response to real-time market volume deviation ($\beta(t)$) within +/- 3% of benchmark curve.
- [ ] POV engine maintains target market volume participation rate $\alpha$ within +/- 1.5% without exceeding the hard participation ceiling.
- [ ] Iceberg engine reloads visible tips with randomized lot variance (+/- 20%) within 500 microseconds of consuming a fill event.
- [ ] Pre-trade risk filters reject parent and child orders violating price collars, single-order size limits, or the 50:1 OTR threshold.
- [ ] Global and per-token kill switches cancel all active child orders across the matching engine within 5 milliseconds.
- [ ] Multiple child slice fills aggregate into an EIP-712 DvP settlement voucher verified and accepted by `SettlementDvP.sol` on Hyperledger Besu.
- [ ] Universal 0.00% (No fee at all) platform fee is accurately computed and distributed with integer precision into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance ledgers.
- [ ] Adheres strictly to the 12 mandatory sections with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `203` (Wallet Account Service), Prompt `204` (Order Service), Prompt `205` (Order Matching Engine), Prompt `206` (Risk & Margin Checks Service), Prompt `207` (Market Data Service).
- **Parallel Work:** Prompt `225` (FIX Protocol Gateway), Prompt `226` (Advanced Order Types & Algorithmic Trigger Engine), Prompt `228` (Real-Time Market Surveillance Engine), Prompt `256` (Limit-Up / Limit-Down Dynamic Volatility Dampener Service).
- **Subsequent Prompts Enabled:** Prompt `306` (SettlementDvP Smart Contract), Prompt `509` (Flutter Institutional Order Slicing UI), Prompt `603` (Institutional Web Trading Terminal).
