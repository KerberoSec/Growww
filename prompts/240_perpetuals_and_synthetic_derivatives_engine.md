# 240 - Perpetuals & Synthetic Derivatives Engine (Rust)

## Purpose
The Perpetuals & Synthetic Derivatives Engine is the ultra-low-latency, 24/7 computational core for non-expiring perpetual futures contracts, tokenized synthetic equities, and cross-collateralized commodity derivatives on the Growww platform. While spot equity trading operates under strict asset-backed fractional custody, continuous international markets, 24/7 global synthetic equities (such as fractional US tech giants and global indices), and continuous commodity contracts (such as Gold, Silver, and Brent Crude) require a deterministic, non-stop derivatives clearinghouse.

This engine executes continuous risk evaluation, hybrid order execution combining an off-chain Central Limit Order Book (CLOB) with a Virtual Automated Market Maker (vAMM) liquidity backstop, dynamic mark price formulation from composite oracle feeds, continuous 8-hour funding rate calculation via premium index time-weighted average price (TWAP), multi-tier position leverage scaling from 1x to 20x, and deterministic Auto-Deleveraging (ADL) ranking to guarantee exchange solvency under extreme market volatility without socialized loss ambiguity.

## What You Are Building
A standalone, high-concurrency, bare-metal optimized Rust microservice (`services/perpetuals-engine`). Concrete deliverables include:
- **In-Memory Hybrid Matching Engine:** Dual-layer matching mechanism executing trades against an off-chain Central Limit Order Book (CLOB) with automatic fall-through routing to a deterministic Virtual Automated Market Maker (vAMM) constant product liquidity curve ($x \cdot y = k$) when CLOB depth is insufficient.
- **Continuous 8-Hour Funding Rate Engine:** Rolling continuous TWAP calculation of the Premium Index between perp contract mid-prices and underlying composite spot index prices, enforcing clamped 8-hour funding rate intervals with continuous per-second or per-block cash transfers between longs and shorts.
- **Multi-Tier Leverage & Margin Controller:** Dynamic initial and maintenance margin scaling engine supporting leverage from 1x up to 20x, adjusting collateralization requirements based on aggregate position notional brackets.
- **Composite Oracle Mark Price Aggregator:** Robust mark price derivation engine synthesizing multi-source spot feeds, calculating 30-minute basis moving averages, and filtering price anomalies to prevent market manipulation and unwarranted liquidations.
- **Liquidation & ADL Ranking Engine:** High-frequency liquidation monitoring pipeline that triggers stepped margin calls, routes unviable positions to the platform Insurance Fund, and deterministically computes Auto-Deleveraging (ADL) quantile rankings based on profit percentage and effective leverage to offload bankrupt positions against high-ranking counterparty accounts.
- **Snapshotting & State Recovery Engine:** Microsecond-latency state checkpointing using RocksDB on NVMe and Kafka Write-Ahead Log (WAL) offset tracking for cold recovery within $< 2$ seconds.

## Scope Boundaries
- **In Scope:**
  - 24/7/365 continuous trading and lifecycle management for perpetual contracts, synthetic equities, and commodity derivatives.
  - Hybrid CLOB and vAMM matching logic with fractional contract sizes (down to 8 decimal places).
  - Continuous 8-hour funding rate calculations and TWAP premium tracking.
  - Multi-tier initial and maintenance margin requirements based on bracketed position notional.
  - Composite oracle feed ingestion, basis calculations, and robust mark price generation.
  - Stepped partial liquidation execution and Insurance Fund intervention.
  - Auto-Deleveraging (ADL) ranking calculations and deterministic deleveraging execution.
  - Real-time unrealized PnL (uPnL) and realized PnL (rPnL) calculations using fixed-point arithmetic.
- **Out of Scope / Handled Elsewhere:**
  - Fiat currency on-ramps and payment gateway processing (Prompt 212).
  - Spot equity token issuance and physical depository custody (Prompt 213, Prompt 303).
  - Global user identity verification and international sanctions onboarding (Prompt 202).
  - Client-facing WebSocket message fan-out distribution (Prompt 207).
  - Front-end Flutter derivative trading interfaces (Prompt 508, Prompt 509).

## Technology to Use
- **Primary Language & Toolchain:** Rust 2021 Edition (stable 1.78+). Rust guarantees zero garbage collection latency spikes, deterministic memory layouts, compile-time thread safety, and sub-15-microsecond execution on the hot matching path.
- **Async Runtime & Thread Architecture:** `tokio` for I/O multiplexing and gRPC server layers; dedicated pinned OS worker threads communicating over lock-free `crossbeam-channel` ring buffers for matching and risk accounting.
- **Fixed-Point Financial Math:** Fixed-point 128-bit integer arithmetic (`fixed` crate or custom fixed-point structs with $10^{18}$ precision for collateral and $10^8$ precision for contract quantities and price fractions) to completely eliminate IEEE 754 floating-point rounding errors.
- **Local NVMe Storage & Persistence:** `rocksdb` for ultra-fast local state snapshotting, position state indices, and WAL checkpointing.
- **Messaging & RPC:** `rdkafka` (librdkafka C bindings) for high-throughput stream ingestion and broadcasting; `tonic` and `prost` for low-latency Protobuf gRPC interfaces.

## Backend / Infra Touchpoints
- **Apache Kafka Ingestion Topics:** Consumes order commands from `derivatives.orders.commands.v1`, oracle price ticks from `market.oracle.ticks.v1`, and collateral deposit notifications from `wallet.collateral.events.v1`.
- **Apache Kafka Publication Topics:** Emits fill events to `derivatives.matches.v1`, position updates to `derivatives.positions.v1`, funding payments to `derivatives.funding.v1`, liquidation alerts to `derivatives.liquidations.v1`, and ADL execution events to `derivatives.adl.v1`.
- **Pre-Trade Risk Engine (Prompt 206):** Verifies global user limits, account collateral status, and circuit breaker states before dispatching derivative orders.
- **Real-Time VaR Margin Engine (Prompt 229):** Exchanges portfolio margin metrics and cross-collateral stress-test requirements.
- **Settlement Guarantee Fund (Prompt 230 / Prompt 315):** Capitalizes the Derivatives Insurance Fund and absorbs residual liquidation deficits prior to activating Auto-Deleveraging.
- **Local NVMe Storage:** Fast local NVMe path `/var/data/growww/perpetuals/snapshots/` for interval RocksDB state dumps.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Collateral Attestation & Vault Proofs:** Perpetual derivatives margin balances (backed by cash collateral, tokenized treasury bills, or 1:1 custodial stable assets) are linked to cryptographic state roots anchored on Hyperledger Besu. The engine reads verified balance allocations from `CollateralVault.sol` under QBFT consensus.
- **Insurance Fund State Anchoring:** Periodic snapshots of the Derivatives Insurance Fund balance, cumulative funding rate settlements, and liquidation receipts are batched and anchored to `DerivativesSettlementRegistry.sol` to provide transparent, verifiable proof of exchange solvency.
- **Cryptographic Execution Attestation:** Match events, liquidation logs, and ADL settlement receipts are signed with the engine's SECP256k1 operational key and recorded with sequence numbers, enabling tamper-evident trade verification without revealing individual trader identities.
- **Zero On-Chain PII:** The engine operates strictly on pseudonymous identifiers (`account_id`, `subaccount_id`, `ledger_address`), ensuring complete privacy of trading strategies and positions.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Cargo Workspace & Structure:** Scaffold `services/perpetuals-engine` with modular crates: `core-math`, `oracle-aggregator`, `clob-book`, `vamm-engine`, `margin-engine`, `funding-calculator`, `liquidation-engine`, and `persistence`. Configure optimized release profile (`opt-level = 3`, `lto = "fat"`, `codegen-units = 1`, `panic = "abort"`).
2. **Implement Fixed-Point Financial Mathematics:** Build 128-bit fixed-point data structures for prices, quantities, margin ratios, and funding percentages. Implement checked and saturating arithmetic operations with strict unit testing against edge cases.
3. **Build Composite Oracle Ingestion & Mark Price Engine:** Create an oracle pipeline that ingests price ticks across multiple providers, computes median spot indices, calculates a 30-minute exponential moving average basis ($P_{\text{basis}} = P_{\text{clob\_mid}} - P_{\text{spot\_index}}$), and emits Mark Price = $P_{\text{spot\_index}} + \text{TWAP}(P_{\text{basis}})$.
4. **Implement Central Limit Order Book (CLOB):** Implement an in-memory limit order book using `BTreeMap` for sorted price levels and intrusive linked lists for FIFO order execution. Support Limit, Market, Post-Only, and Immediate-or-Cancel (IOC) orders.
5. **Implement Virtual Automated Market Maker (vAMM):** Construct a deterministic vAMM liquidity pool based on the constant product invariant $(x \cdot y = k)$ to serve as continuous backstop liquidity. Configure slippage guardrails and virtual reserve adjustments.
6. **Construct Hybrid Matching Router:** Develop the hybrid routing layer that attempts matching against resting CLOB limit orders first; if depth is exhausted, routes the unfilled residual volume to the vAMM within configured maximum slippage tolerances.
7. **Implement Multi-Tier Margin & Leverage Module:** Create leverage configuration tables mapping position notional tiers (1x up to 20x) to Initial Margin Ratio (IMR) and Maintenance Margin Ratio (MMR). Enforce dynamic margin checks before order acceptance.
8. **Build Continuous 8-Hour Funding Rate Engine:** Implement a rolling 8-hour TWAP calculator for the Premium Index: $P_{\text{index}} = \frac{\max(0, P_{\text{impact\_bid}} - P_{\text{index}}) - \max(0, P_{\text{index}} - P_{\text{impact\_ask}})}{P_{\text{index}}}$. Calculate clamped funding rates and execute continuous funding fee transfers between long and short positions.
9. **Build Unrealized and Realized PnL Calculator:** Implement real-time position accounting evaluating mark-to-market valuations, entry prices, cumulative funding payments, and fee deductions.
10. **Implement Stepped Liquidation Pipeline:** Build an ultra-low-latency liquidation evaluator checking account equity against Maintenance Margin. If account equity falls below MMR, initiate partial liquidation orders against the CLOB/vAMM or transfer positions to the Insurance Fund.
11. **Implement Auto-Deleveraging (ADL) Ranking Queue:** Construct an ADL engine that sorts opposing profitable positions by ADL quantile score: $\text{Score}_{\text{ADL}} = \text{PnL\%} \times \text{Effective Leverage}$. If the Insurance Fund cannot absorb a bankrupt liquidation, match the position against top-ranked profitable accounts at the bankruptcy price.
12. **Implement Persistence & RocksDB Snapshotting:** Build state serialization and background snapshotting routines that flush memory structures to NVMe storage at regular sequence intervals without blocking the main matching thread.
13. **Build Kafka Command Ingestor & Event Publisher:** Implement zero-copy Kafka command ingestion with per-market pinned channels and batch event broadcasting for fills, liquidations, and funding events.
14. **Implement High-Performance gRPC Admin & Query Server:** Expose gRPC endpoints using `tonic` for real-time market stats, position lookups, oracle feeds, manual risk controls, and emergency market circuit halts.
15. **Execute Benchmarking, Invariant Verification, and Stress Testing:** Write Criterion benchmarks verifying sub-15-microsecond match latency; build fuzz tests verifying mathematical invariants, zero balance leakage, and 100% deterministic replay from Kafka WAL logs.

## Interfaces / Contracts

### Protobuf Service & Message Definitions (`perpetuals_engine.proto`)
```protobuf
syntax = "proto3";

package growww.perpetuals.v1;

option go_package = "growww/perpetuals/v1;perpetualsv1";

service PerpetualsEngineService {
  rpc SubmitOrder (SubmitDerivativeOrderRequest) returns (SubmitDerivativeOrderResponse);
  rpc CancelOrder (CancelDerivativeOrderRequest) returns (CancelDerivativeOrderResponse);
  rpc GetPosition (GetPositionRequest) returns (GetPositionResponse);
  rpc GetMarketSummary (GetMarketSummaryRequest) returns (GetMarketSummaryResponse);
  rpc UpdateOraclePrice (UpdateOraclePriceRequest) returns (UpdateOraclePriceResponse);
  rpc TriggerEmergencyHalt (EmergencyHaltRequest) returns (EmergencyHaltResponse);
}

enum DerivativeAssetType {
  DERIVATIVE_ASSET_TYPE_UNSPECIFIED = 0;
  DERIVATIVE_ASSET_TYPE_PERPETUAL_FUTURE = 1;
  DERIVATIVE_ASSET_TYPE_SYNTHETIC_EQUITY = 2;
  DERIVATIVE_ASSET_TYPE_COMMODITY = 3;
}

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BUY = 1;
  ORDER_SIDE_SELL = 2;
}

enum OrderExecutionType {
  ORDER_EXECUTION_TYPE_UNSPECIFIED = 0;
  ORDER_EXECUTION_TYPE_LIMIT = 1;
  ORDER_EXECUTION_TYPE_MARKET = 2;
  ORDER_EXECUTION_TYPE_POST_ONLY = 3;
  ORDER_EXECUTION_TYPE_IOC = 4;
  ORDER_EXECUTION_TYPE_FOK = 5;
}

enum ExecutionVenue {
  EXECUTION_VENUE_UNSPECIFIED = 0;
  EXECUTION_VENUE_CLOB = 1;
  EXECUTION_VENUE_VAMM = 2;
  EXECUTION_VENUE_HYBRID = 3;
}

message SubmitDerivativeOrderRequest {
  string order_id = 1;
  string account_id = 2;
  string market_symbol = 3;
  DerivativeAssetType asset_type = 4;
  OrderSide side = 5;
  OrderExecutionType execution_type = 6;
  uint64 price_e8 = 7;
  uint64 quantity_e8 = 8;
  uint32 requested_leverage = 9;
  bool is_reduce_only = 10;
  int64 client_timestamp_ns = 11;
}

message SubmitDerivativeOrderResponse {
  bool accepted = 1;
  string order_id = 2;
  string rejection_reason = 3;
  uint64 sequence_number = 4;
  int64 processed_at_ns = 5;
}

message CancelDerivativeOrderRequest {
  string order_id = 1;
  string account_id = 2;
  string market_symbol = 3;
}

message CancelDerivativeOrderResponse {
  bool success = 1;
  string order_id = 2;
  string error_message = 3;
}

message GetPositionRequest {
  string account_id = 1;
  string market_symbol = 2;
}

message PositionDto {
  string account_id = 1;
  string market_symbol = 2;
  OrderSide side = 3;
  uint64 size_e8 = 4;
  uint64 entry_price_e8 = 5;
  uint64 mark_price_e8 = 6;
  uint64 liquidation_price_e8 = 7;
  uint64 collateral_amount_e18 = 8;
  uint32 effective_leverage = 9;
  int64 unrealized_pnl_e18 = 10;
  int64 cumulative_funding_e18 = 11;
  uint32 adl_quantile = 12;
  int64 updated_at_ns = 13;
}

message GetPositionResponse {
  PositionDto position = 1;
}

message GetMarketSummaryRequest {
  string market_symbol = 1;
}

message GetMarketSummaryResponse {
  string market_symbol = 1;
  uint64 index_price_e8 = 2;
  uint64 mark_price_e8 = 3;
  int64 funding_rate_e8 = 4;
  int64 next_funding_timestamp_ns = 5;
  uint64 open_interest_long_e8 = 6;
  uint64 open_interest_short_e8 = 7;
  uint64 insurance_fund_balance_e18 = 8;
  uint64 vamm_reserve_base_e8 = 9;
  uint64 vamm_reserve_quote_e18 = 10;
  int64 last_match_timestamp_ns = 11;
}

message UpdateOraclePriceRequest {
  string market_symbol = 1;
  string oracle_source = 2;
  uint64 spot_price_e8 = 3;
  uint64 confidence_interval_e8 = 4;
  int64 oracle_timestamp_ns = 5;
}

message UpdateOraclePriceResponse {
  bool accepted = 1;
  uint64 consolidated_index_price_e8 = 2;
  uint64 derived_mark_price_e8 = 3;
}

message EmergencyHaltRequest {
  string market_symbol = 1;
  string reason = 2;
  string operator_signature = 3;
}

message EmergencyHaltResponse {
  bool halted = 1;
  int64 halted_at_ns = 2;
}
```

### Event Streaming Schemas (Kafka)

```protobuf
message DerivativeMatchEvent {
  string match_id = 1;
  string market_symbol = 2;
  string maker_order_id = 3;
  string taker_order_id = 4;
  string maker_account_id = 5;
  string taker_account_id = 6;
  OrderSide taker_side = 7;
  ExecutionVenue venue = 8;
  uint64 matched_price_e8 = 9;
  uint64 matched_quantity_e8 = 10;
  uint64 trade_fee_bps = 11; // Universal Zero-Fee Model (0.00% fee - No fee at all) (0 bps (0.00% fee at launch))
  uint64 fee_amount_e18 = 12; // 0.0000 * notional (0.00% fee at launch) (0.00% fee at launch; future fee parameters governed by FeeController.sol)
  uint64 sequence_number = 13;
  int64 executed_at_ns = 14;
}

message FundingRateSettlementEvent {
  string settlement_id = 1;
  string market_symbol = 2;
  int64 funding_rate_e8 = 3;
  uint64 mark_price_e8 = 4;
  uint64 index_price_e8 = 5;
  int64 total_long_payment_e18 = 6;
  int64 total_short_payment_e18 = 7;
  uint64 open_interest_e8 = 8;
  int64 settlement_timestamp_ns = 9;
}

message LiquidationExecutionEvent {
  string liquidation_id = 1;
  string account_id = 2;
  string market_symbol = 3;
  OrderSide liquidated_side = 4;
  uint64 liquidated_size_e8 = 5;
  uint64 bankruptcy_price_e8 = 6;
  uint64 execution_price_e8 = 7;
  int64 insurance_fund_delta_e18 = 8;
  bool adl_triggered = 9;
  int64 liquidated_at_ns = 10;
}

message AdlExecutionEvent {
  string adl_id = 1;
  string target_account_id = 2;
  string counterparty_account_id = 3;
  string market_symbol = 4;
  OrderSide closed_side = 5;
  uint64 closed_size_e8 = 6;
  uint64 execution_price_e8 = 7;
  int64 realized_pnl_e18 = 8;
  int64 executed_at_ns = 9;
}
```

### Mathematical Specifications & Formulation Reference

#### 1. Multi-Tier Leverage & Margin Brackets
| Tier | Position Notional Range (USD Value) | Max Leverage | Initial Margin Ratio (IMR) | Maintenance Margin Ratio (MMR) |
|---|---|---|---|---|
| Tier 1 | 0 to 50,000 | 20x | 5.0% | 2.5% |
| Tier 2 | 50,001 to 250,000 | 10x | 10.0% | 5.0% |
| Tier 3 | 250,001 to 1,000,000 | 5x | 20.0% | 10.0% |
| Tier 4 | 1,000,001 to 5,000,000 | 2x | 50.0% | 25.0% |
| Tier 5 | > 5,000,000 | 1x | 100.0% | 50.0% |

#### 2. Composite Mark Price Formulation
$$\text{Spot Index} = \text{Median}(P_{\text{oracle\_1}}, P_{\text{oracle\_2}}, P_{\text{oracle\_3}})$$
$$\text{Basis} = \text{CLOB Mid Price} - \text{Spot Index}$$
$$\text{Mark Price} = \text{Spot Index} + \text{TWAP}_{30\text{m}}(\text{Basis})$$

#### 3. Continuous 8-Hour Funding Rate Formula
$$\text{Premium Index (P)} = \frac{\max(0, P_{\text{impact\_bid}} - \text{Index Price}) - \max(0, \text{Index Price} - P_{\text{impact\_ask}})}{\text{Index Price}}$$
$$F = \text{Clamp}(\text{TWAP}_{8\text{h}}(P) + \text{Clamp}(I - \text{TWAP}_{8\text{h}}(P), -0.05\%, 0.05\%), -0.75\%, 0.75\%)$$
where $I$ is the benchmark interest rate component (default $0.01\%$ per 8 hours).

#### 4. Auto-Deleveraging (ADL) Quantile Ranking
$$\text{Profit Margin Ratio} = \frac{\text{Unrealized PnL}}{\text{Position Initial Margin}}$$
$$\text{Effective Leverage} = \frac{\text{Position Notional Value}}{\text{Account Equity}}$$
$$\text{ADL Score} = \begin{cases} \text{Profit Margin Ratio} \times \text{Effective Leverage} & \text{if Unrealized PnL} > 0 \\ \frac{\text{Profit Margin Ratio}}{\text{Effective Leverage}} & \text{if Unrealized PnL} \le 0 \end{cases}$$

## Security & Compliance Notes
- **Zero IEEE 754 Floating Point Math:** All financial accounting, fees, funding payments, and liquidation margins are strictly evaluated using 128-bit fixed-point arithmetic with explicit overflow and underflow checks.
- **Pre-Allocated Memory on Hot Path:** The matching and risk loop operates without runtime dynamic heap allocations, utilizing pre-allocated circular buffers, object pools, and slab allocators to prevent allocator locks and memory fragmentation.
- **Self-Trade Prevention (STP):** Orders from the same beneficial owner or linked subaccounts are automatically prevented from self-matching to prevent wash trading and artificial volume generation.
- **Oracle Manipulation & Stale Price Protection:** If an oracle price feed deviates by more than 5% from peer feeds or fails to update within 10 seconds, it is automatically disqualified from the median index. If fewer than two valid oracle feeds exist, the engine halts new position openings and transitions to reduce-only mode.
- **Deterministic Replay Guarantee:** Thread pinning and sequential Kafka WAL offsets guarantee 100% deterministic reproducibility of order matching, funding transfers, and liquidations upon system restart.

## Acceptance Criteria
- [ ] Rust microservice compiles cleanly with zero warnings under `cargo clippy --all-targets -- -D warnings`.
- [ ] Hybrid matching engine correctly matches orders against resting CLOB depth and routes residual volume to the vAMM within configured slippage limits.
- [ ] Continuous 8-hour funding rate calculation correctly evaluates Premium Index TWAP and applies funding rate clamping within $[-0.75\%, +0.75\%]$.
- [ ] Multi-tier leverage module dynamically applies initial margin (5% to 100%) and maintenance margin (2.5% to 50%) based on position size brackets.
- [ ] Mark price engine derives accurate composite prices from multiple oracle streams and handles single-oracle failure or outlier price feeds seamlessly.
- [ ] Stepped liquidation mechanism successfully liquidates under-collateralized positions and routes deficit balances to the Insurance Fund.
- [ ] ADL ranking engine correctly scores profitable counterparties and executes deterministic deleveraging when Insurance Fund balance is zero.
- [ ] Fixed-point 128-bit arithmetic produces zero precision loss across 10,000,000 simulated trades and funding cycles.
- [ ] $p99$ order matching latency remains under $20\mu\text{s}$ under an ingestion load of 50,000 orders per second.
- [ ] State snapshotting to RocksDB and WAL recovery restores identical order book, position state, and sequence numbers within 2 seconds.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 010 (NFR Targets), Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 205 (Order Matching Engine), Prompt 206 (Risk Engine), Prompt 407 (Master Data Management).
- **Parallel Tasks:** Prompt 207 (Market Data Service), Prompt 226 (Advanced Order Types Engine), Prompt 229 (Real-Time VaR Margin Engine).
- **Downstream Blockers:** Prompt 230 (Settlement Guarantee Fund Service), Prompt 315 (Settlement Guarantee Fund Contract), Prompt 508 (Security Detail Screen), Prompt 903 (Load Testing Matching Engine).
