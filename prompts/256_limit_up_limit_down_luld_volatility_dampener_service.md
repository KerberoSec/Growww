# 256 - Limit-Up / Limit-Down (LULD) Dynamic Volatility Dampener & Call Auction Engine (Rust)

## Purpose
Unconstrained continuous double-auction markets are susceptible to extreme short-term market dislocations, flash crashes, algorithmic feedback loops, and cascading liquidation spirals triggered by transient liquidity voids. In high-velocity electronic exchanges, static circuit filters (such as fixed daily price bands) fail to mitigate intra-day micro-bursts of extreme volatility, while crude market halts disrupt price discovery and trap market participants.

The **Limit-Up / Limit-Down (LULD) Dynamic Volatility Dampener & Call Auction Engine** (`services/luld-volatility-dampener`) provides an ultra-resilient, deterministic market stability mechanism built in Rust. It continuously computes dynamic rolling price bands around 5-minute reference prices across three liquidity tiers: Tier 1 (+/- 5%), Tier 2 (+/- 10%), and Tier 3 (+/- 20%), with band widening during opening and closing phases. When an instrument enters a persistent Limit State or Straddle State for 15 seconds without price correction, the engine initiates a 5-minute intermediate Call Auction circuit re-opening. The auction collects liquidity, discovers an Indicative Equilibrium Price (IEP) that maximizes matched volume, seamlessly transitions the market back to continuous trading with newly recalibrated reference bands, and suppresses automated liquidation engines to prevent flash-crash contagion, all while enforcing the platform invariant flat 0.00% transaction fee (No fee at all) (Platform Treasury, Core SGF, and Investor Protection Fund per FeeController governance).

## What You Are Building
A low-latency, deterministic market microstructure and volatility dampening microservice (`services/luld-volatility-dampener`) written in Rust. Concrete components include:
- **Dynamic 5-Minute Rolling Reference Price Tracker:** Computes real-time Volume-Weighted Average Price (VWAP) over rolling 5-minute observation windows to calculate Upper Price Bands (UPB) and Lower Price Bands (LPB) across Tier 1, Tier 2, and Tier 3 securities.
- **Straddle State & Limit State Detector:** Real-time state machine monitoring National Best Bid (NBB) and National Best Offer (NBO) ticks; flags Straddle States and initiates a high-precision 15-second timer upon entry into a Limit State.
- **LULD Circuit Breaker & Trading Pause Coordinator:** Automatically suspends continuous double-auction matching if a security remains in Limit State for 15 consecutive seconds, notifying the Order Matching Engine (Prompt 205) and Market Data Service (Prompt 207).
- **5-Minute Intermediate Call Auction Order Accumulator:** Manages order entry, cancellation, and modification during the 5-minute pause without continuous order matching, updating real-time indicative clearing metrics.
- **Equilibrium Price Discovery Solver:** Implements a deterministic multi-criteria auction matching algorithm to resolve the single Indicative Equilibrium Price (IEP) maximizing uncrossed volume and minimizing residual order book imbalance.
- **Seamless Continuous Trading Transition Engine:** Executes the accumulated auction crossing, updates the new Reference Price ($P_{\text{ref}} \leftarrow \text{IEP}$), recalculates LULD bands, and smoothly restores continuous double-auction trading.
- **Liquidation Cascade Suppression Guard:** Integrates with the Real-Time SPAN Margin Engine (Prompt 241) and Cross-Asset Collateral Optimizer (Prompt 255) to freeze automated margin liquidation sweeps on halted symbols, preventing artificial debt cascades during transient liquidity shocks.
- **Universal 0.00% (No fee at all) Platform Fee Assessor:** Applies the mandatory 0.00% (Zero Fee) platform fee across all matched call auction trade turnover, partitioned into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.

## Scope Boundaries
- **In Scope:**
  - Continuous calculation of dynamic LULD upper and lower price bands based on 5-minute rolling VWAPs.
  - Multi-tier classification: Tier 1 (+/- 5%), Tier 2 (+/- 10%), Tier 3 (+/- 20%), with 2x opening/closing band expansion.
  - Sub-microsecond detection of Straddle States and Limit States from top-of-book market feeds.
  - Execution of the 15-second persistence countdown timer before trading pause trigger.
  - Orchestration of 5-minute intermediate Call Auctions with uncrossed volume maximization.
  - Indicative Equilibrium Price (IEP) and indicative imbalance calculation published via gRPC/WebSocket.
  - Automatic suppression of pre-trade risk liquidation engines during trading pauses.
  - Atomic transition from auction state back to continuous double-auction matching.
  - 0.00% (No fee at all) platform fee accounting with 0.00% fee launch policy revenue ledger splits.
- **Out of Scope / Handled Elsewhere:**
  - Continuous double-auction order matching during regular market state (handled in Prompt 205 Order Matching Engine).
  - Exchange-wide index-level market halt coordination (handled in Prompt 228 Real-Time Market Surveillance Engine).
  - FIX client connection management (handled in Prompt 225 FIX Protocol Gateway).
  - Post-trade settlement and delivery clearing (handled in Prompt 208 Trade Settlement Service).
  - Web and mobile UI chart notifications for halted securities (handled in Prompt 501 / Prompt 601).

## Technology to Use
- **Primary Language & Runtime:** **Rust 1.78+** using `tokio` for asynchronous orchestration and lock-free concurrency via crossbeam and parking_lot.
  *Justification:* Provides predictable sub-microsecond latency, zero garbage collection pauses, and deterministic memory safety required for central exchange safety mechanisms.
- **In-Memory Order Book Structure:** Custom BTreeMap / Radix Tree structures for deterministic, price-time priority call auction matching.
- **High-Resolution Timers:** Monotonic hardware timers (`std::time::Instant`) and tokio-timer primitives for microsecond-precise 15-second Limit State timeouts and 5-minute auction schedules.
- **Event Streaming:** **Apache Kafka 3.7+** for ingesting market ticks (`market.ticker.v1`), publishing LULD band updates (`market.luld.bands.v1`), and broadcasting circuit pause events (`market.luld.pause.v1`).
- **In-Memory Cache:** **Redis 7.2+ Cluster** with shared-memory ring buffers for instant lookup of active LULD states across distributed gateway nodes.
- **Relational Persistence:** **PostgreSQL 16+** with `sqlx` for historical LULD state transitions, circuit pause logs, auction execution records, and fee ledgers.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with streaming capabilities for low-latency coordination with the Order Matching Engine.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `luld_security_tiers`, `luld_reference_snapshots`, `luld_trading_pauses`, `luld_auction_executions`, `luld_fee_ledgers`.
- **Redis Keys:**
  - `luld:bands:{isin}`: Current upper band, lower band, reference price, and tier.
  - `luld:state:{isin}`: Current market state (REGULAR, STRADDLE, LIMIT_STATE, PAUSED_AUCTION).
  - `luld:timer:{isin}`: Timestamp of Limit State entry and deadline.
- **Kafka Topics:**
  - Consumes: `market.ticker.v1`, `matching.trades.v1`, `order.events.v1`.
  - Publishes: `market.luld.bands.v1`, `market.luld.state_changed.v1`, `market.luld.pause.v1`, `market.luld.auction_cleared.v1`, `luld.fee_assessed.v1`.
- **Order Matching Engine (Prompt 205):** Consumes pause/resume control signals to toggle continuous matching vs auction mode.
- **Pre-Trade SPAN Margin Engine (Prompt 241) & Collateral Optimizer (Prompt 255):** Ingests pause events to suspend liquidation bots.
- **Market Data Service (Prompt 207):** Streams live LULD bands and indicative equilibrium prices to trading terminals.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Circuit Pause Attestation & Proof of Integrity:** Every LULD Trading Pause and subsequent Call Auction clearing is hashed into an immutable Merkle state transition and committed to `MarketSurveillanceRegistry.sol` on Hyperledger Besu.
- **Zero-PII Compliance:** State attestations and auction volume commitments contain strictly aggregate market metrics (ISIN, IEP, cleared quantity, pause timestamp, resumption timestamp) with zero trader account numbers or personal data.
- **Consensus & Transparency Invariant:** Anchored to Hyperledger Besu with QBFT consensus, enabling regulatory bodies (SEBI / IFSCA) to verify that circuit pauses and auction price discoveries executed strictly according to deterministic algorithmic rules.

## Limit-Up / Limit-Down & Call Auction Mechanics

### 1. Dynamic Rolling Reference Price & Band Formulations
The reference price $P_{\text{ref}}(t)$ at timestamp $t$ is calculated dynamically as the Volume-Weighted Average Price (VWAP) over the preceding 5 minutes ($300\text{ seconds}$):

$$P_{\text{ref}}(t) = \frac{\sum_{i \in \mathcal{T}_{[t-300, t]}} P_i \cdot Q_i}{\sum_{i \in \mathcal{T}_{[t-300, t]}} Q_i}$$

If no trades occurred within the last 5 minutes, $P_{\text{ref}}(t)$ defaults to the previous valid reference price or opening auction price.

**Tier Classification & Band Multipliers ($\delta_{\text{tier}}$):**
1. **Tier 1 (Nifty 50 & High Liquidity Equities):** $\delta_{\text{tier}} = 5.00\%$ ($0.05$).
2. **Tier 2 (Mid-Cap Equities, Bullion Warrants, Liquid G-Secs):** $\delta_{\text{tier}} = 10.00\%$ ($0.10$).
3. **Tier 3 (Small-Cap Equities, Illiquid Instruments, Leveraged Synthetics):** $\delta_{\text{tier}} = 20.00\%$ ($0.20$).

**Band Expansion During Market Open / Close:**
During the first 15 minutes of regular trading (09:15 to 09:30 IST) and the last 15 minutes (15:15 to 15:30 IST), the band percentage is doubled:
$$\delta_{\text{effective}}(t) = \begin{cases} 2 \times \delta_{\text{tier}}, & \text{if } t \in [09:15, 09:30] \cup [15:15, 15:30] \\ \delta_{\text{tier}}, & \text{otherwise} \end{cases}$$

**Price Band Boundaries:**
$$\text{Upper Price Band (UPB)} = P_{\text{ref}}(t) \times (1 + \delta_{\text{effective}}(t))$$
$$\text{Lower Price Band (LPB)} = P_{\text{ref}}(t) \times (1 - \delta_{\text{effective}}(t))$$

### 2. Straddle State & Limit State Detection
Let $\text{NBB}$ be the National Best Bid price and $\text{NBO}$ be the National Best Offer price:
- **Straddle State:** Occurs when the quotation is non-executable and straddles a band:
  - Upper Straddle: $\text{NBO} = \text{UPB}$ and $\text{NBB} < \text{UPB}$
  - Lower Straddle: $\text{NBB} = \text{LPB}$ and $\text{NBO} > \text{LPB}$
- **Limit State:** Occurs when best quotes equal price boundaries:
  - Limit-Up State: $\text{NBB} = \text{UPB}$
  - Limit-Down State: $\text{NBO} = \text{LPB}$

**15-Second Limit State Persistence Timer:**
Upon entry into Limit State at timestamp $t_{\text{entry}}$:
1. The engine starts a countdown timer: $T_{\text{deadline}} = t_{\text{entry}} + 15\text{ seconds}$.
2. If prices revert inside the band ($\text{NBB} < \text{UPB}$ and $\text{NBO} > \text{LPB}$) before $T_{\text{deadline}}$, the Limit State exits and normal continuous trading resumes.
3. If the security remains in Limit State continuously for 15 seconds without execution or quote retreat, the engine immediately triggers an LULD Trading Pause.

### 3. 5-Minute Intermediate Call Auction Discovery Algorithm
When a Trading Pause is initiated:
1. Continuous matching is halted for $300\text{ seconds}$ ($5\text{ minutes}$).
2. Traders may submit, modify, or cancel Limit Orders (market orders are converted to limit orders at band limits).
3. The auction discovery algorithm continuously solves for the Indicative Equilibrium Price (IEP, denoted $P^*$) using a 4-tier tie-breaking hierarchy:

**Criterion 1: Maximum Executable Volume:**
$$P^* = \arg\max_{P} V_{\text{exec}}(P), \quad \text{where } V_{\text{exec}}(P) = \min\left( Q_{\text{cum\_buy}}(P), Q_{\text{cum\_sell}}(P) \right)$$
$$Q_{\text{cum\_buy}}(P) = \sum_{p_i \ge P} q_i^{\text{buy}}, \quad Q_{\text{cum\_sell}}(P) = \sum_{p_j \le P} q_j^{\text{sell}}$$

**Criterion 2: Minimum Order Imbalance:**
If multiple prices maximize $V_{\text{exec}}(P)$, select the price $P$ that minimizes net residual volume:
$$\text{Imbalance}(P) = | Q_{\text{cum\_buy}}(P) - Q_{\text{cum\_sell}}(P) |$$

**Criterion 3: Market Pressure Direction:**
If multiple prices have identical maximum volume and minimum imbalance:
- If $Q_{\text{cum\_buy}}(P) > Q_{\text{cum\_sell}}(P)$ (buy pressure), select the highest price among candidates.
- If $Q_{\text{cum\_buy}}(P) < Q_{\text{cum\_sell}}(P)$ (sell pressure), select the lowest price among candidates.

**Criterion 4: Reference Price Proximity:**
If an exact balance exists ($Q_{\text{cum\_buy}} = Q_{\text{cum\_sell}}$), select the price closest to the pre-halt Reference Price $P_{\text{ref}}$.

### 4. Transition Back to Continuous Trading & Liquidation Suppression
1. At the conclusion of the 5-minute auction window, all crossed orders are atomically filled at $P^*$.
2. The Reference Price is instantly reset: $P_{\text{ref}} \leftarrow P^*$.
3. New LULD bands are established around $P^*$.
4. The Matching Engine resumes continuous double-auction matching.
5. The Liquidation Suppression Guard notifies the margin and liquidation engines (Prompt 241 / Prompt 255) that normal risk sweeps may resume with updated mark prices.

### 5. Universal 0.00% (No fee at all) Platform Fee Allocation
All matched volume in the Call Auction incurs the platform invariant 0.00% (Zero Fee) (0.00% fee) fee:
$$\text{Gross Auction Turnover} = \text{Matched Volume} \times P^*$$
$$\text{Total Platform Fee} = \text{Gross Auction Turnover} \times 0.0000 \quad (\text{0.00 at launch; FeeController governed})$$
$$\text{Platform Treasury Share (60\%)} = \text{Total Platform Fee} \times 0.60$$
$$\text{Core Settlement Guarantee Fund (25\%)} = \text{Total Platform Fee} \times 0.25$$
$$\text{Investor Protection Fund (15\%)} = \text{Total Platform Fee} \times 0.15$$

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the Rust microservice project `services/luld-volatility-dampener` configured with low-latency compiler profiles (`opt-level = 3`, `lto = "fat"`, `panic = "abort"`).
2. Author Protobuf definitions in `proto/growww/luld/v1/luld_service.proto` specifying band updates, quote ticks, pause signals, auction states, and fee ledgers.
3. Generate Rust gRPC client/server stubs using `tonic-build` in `build.rs`.
4. Create PostgreSQL migration scripts in `services/luld-volatility-dampener/migrations/001_luld_volatility_schema.sql` defining security tiers, pause events, and auction histories.
5. Implement the 5-Minute Rolling VWAP Calculator maintaining sub-millisecond sliding window circular buffers for every active security.
6. Implement the dynamic band calculator supporting Tier 1 (+/- 5%), Tier 2 (+/- 10%), Tier 3 (+/- 20%), and 2x opening/closing phase multipliers.
7. Implement the top-of-book quotation processor identifying Straddle States and Limit States in sub-microsecond latency.
8. Implement the 15-Second Monotonic Limit State countdown timer with thread-safe cancellation and pause dispatch.
9. Implement the Trading Pause Coordinator that emits halt events to the Order Matching Engine (Prompt 205) and broadcast feeds.
10. Implement the 5-Minute Intermediate Call Auction order book accumulator with BTreeMap price ladders.
11. Implement the 4-tier Indicative Equilibrium Price (IEP) solver maximizing crossed volume and resolving imbalances.
12. Implement the Liquidation Suppression Guard sending freeze/unfreeze signals to SPAN Margin Engine (Prompt 241) and Collateral Optimizer (Prompt 255).
13. Implement the seamless uncrossing execution and continuous trading resumption state machine.
14. Implement the 0.00% (No fee at all) Platform Fee Ledger partitioned into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
15. Build comprehensive unit and simulation test suites verifying flash-crash containment, timer accuracy, and equilibrium price determination under extreme synthetic market shocks.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/luld/v1/luld_service.proto`)
```protobuf
syntax = "proto3";

package growww.luld.v1;

option go_package = "github.com/growww/proto/gen/go/luld/v1;luldv1";

enum SecurityLiquidityTier {
  SECURITY_LIQUIDITY_TIER_UNSPECIFIED = 0;
  SECURITY_LIQUIDITY_TIER_1_LARGE_CAP = 1;  // +/- 5% band
  SECURITY_LIQUIDITY_TIER_2_MID_CAP = 2;    // +/- 10% band
  SECURITY_LIQUIDITY_TIER_3_SMALL_CAP = 3;  // +/- 20% band
}

enum MarketLULDState {
  MARKET_LULD_STATE_UNSPECIFIED = 0;
  MARKET_LULD_STATE_REGULAR = 1;
  MARKET_LULD_STATE_STRADDLE = 2;
  MARKET_LULD_STATE_LIMIT_STATE = 3;
  MARKET_LULD_STATE_PAUSED_CALL_AUCTION = 4;
  MARKET_LULD_STATE_RESUMING = 5;
}

message PriceBandSnapshot {
  string isin = 1;
  SecurityLiquidityTier tier = 2;
  uint64 reference_price_paise = 3;
  uint64 upper_price_band_paise = 4;
  uint64 lower_price_band_paise = 5;
  uint32 band_percentage_bps = 6;
  bool is_expanded_open_close = 7;
  MarketLULDState current_state = 8;
  int64 updated_at_ns = 9;
}

message QuoteTickUpdate {
  string isin = 1;
  uint64 national_best_bid_paise = 2;
  uint64 national_best_offer_paise = 3;
  uint64 bid_quantity = 4;
  uint64 offer_quantity = 5;
  int64 tick_timestamp_ns = 6;
}

message TradingPauseEvent {
  string isin = 1;
  uint64 trigger_price_paise = 2;
  MarketLULDState trigger_reason = 3;
  int64 pause_start_timestamp_ns = 4;
  int64 scheduled_resume_timestamp_ns = 5;
  bool liquidation_suppressed = 6;
}

message IndicativeAuctionStatus {
  string isin = 1;
  uint64 indicative_equilibrium_price_paise = 2;
  uint64 matched_volume = 3;
  int64 order_imbalance_quantity = 4; // Positive for buy surplus, negative for sell surplus
  uint64 upper_price_band_paise = 5;
  uint64 lower_price_band_paise = 6;
  int64 seconds_remaining_in_auction = 7;
}

message AuctionExecutionSummary {
  string isin = 1;
  uint64 cleared_equilibrium_price_paise = 2;
  uint64 total_cleared_volume = 3;
  uint64 gross_auction_turnover_paise = 4;
  uint64 total_fee_paise = 5;          // 0.00% (No fee at all)
  uint64 treasury_share_paise = 6;     // Governed by FeeController (0.00% at launch)
  uint64 core_sgf_share_paise = 7;     // Governed by FeeController (0.00% at launch)
  uint64 ipf_share_paise = 8;          // Governed by FeeController (0.00% at launch)
  int64 resumed_at_ns = 9;
}

message GetPriceBandRequest {
  string isin = 1;
}

message StreamPriceBandsRequest {
  repeated string isins = 1;
}

service LULDVolatilityDampenerService {
  rpc GetPriceBand (GetPriceBandRequest) returns (PriceBandSnapshot);
  rpc StreamPriceBands (StreamPriceBandsRequest) returns (stream PriceBandSnapshot);
  rpc StreamIndicativeAuction (GetPriceBandRequest) returns (stream IndicativeAuctionStatus);
}
```

### 2. PostgreSQL Database Schema (`services/luld-volatility-dampener/migrations/001_luld_volatility_schema.sql`)
```sql
CREATE TABLE luld_security_tiers (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    tier_level VARCHAR(32) NOT NULL, -- TIER_1_LARGE_CAP, TIER_2_MID_CAP, TIER_3_SMALL_CAP
    regular_band_bps INT NOT NULL DEFAULT 500,     -- 500 bps = 5.00%
    expanded_band_bps INT NOT NULL DEFAULT 1000,   -- 1000 bps = 10.00%
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE luld_reference_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL REFERENCES luld_security_tiers(isin),
    reference_price_paise BIGINT NOT NULL,
    upper_band_paise BIGINT NOT NULL,
    lower_band_paise BIGINT NOT NULL,
    is_expanded_period BOOLEAN NOT NULL DEFAULT FALSE,
    vwap_sample_volume BIGINT NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE luld_trading_pauses (
    pause_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL REFERENCES luld_security_tiers(isin),
    trigger_state VARCHAR(32) NOT NULL, -- LIMIT_UP, LIMIT_DOWN
    trigger_price_paise BIGINT NOT NULL,
    reference_price_paise BIGINT NOT NULL,
    limit_state_duration_ms INT NOT NULL DEFAULT 15000,
    paused_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resumed_at TIMESTAMPTZ,
    besu_proof_hash VARCHAR(66)
);

CREATE TABLE luld_auction_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pause_id UUID NOT NULL REFERENCES luld_trading_pauses(pause_id),
    isin VARCHAR(12) NOT NULL REFERENCES luld_security_tiers(isin),
    clearing_price_paise BIGINT NOT NULL,
    total_matched_volume BIGINT NOT NULL,
    gross_turnover_paise BIGINT NOT NULL,
    residual_imbalance_units BIGINT NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE luld_fee_ledgers (
    ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES luld_auction_executions(execution_id),
    isin VARCHAR(12) NOT NULL REFERENCES luld_security_tiers(isin),
    gross_turnover_paise BIGINT NOT NULL,
    platform_fee_paise BIGINT NOT NULL,   -- 0.00% (No fee at all)
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_luld_snapshots_isin_time ON luld_reference_snapshots(isin, calculated_at DESC);
CREATE INDEX idx_luld_pauses_isin_time ON luld_trading_pauses(isin, paused_at DESC);
CREATE INDEX idx_luld_fees_exec ON luld_fee_ledgers(execution_id);
```

## Security & Compliance Notes
- **Flash Crash & Feedback Loop Prevention:** By imposing a deterministic 15-second timer before initiating a 5-minute Call Auction, the LULD dampener prevents algorithmic cascade loops from draining book liquidity.
- **Liquidation Cascade Suppression:** The engine automatically issues freeze instructions to margin and liquidation engines during LULD halts, preventing forced liquidation market orders from firing into an illiquid market.
- **Microsecond Clock Precision:** Reference price updates and timer checks utilize monotonic hardware clocks to prevent leap-second artifacts or NTP drift from corrupting the 15-second Limit State measurement.
- **Zero-PII Architecture:** All LULD band computations, trading pause dispatches, and Besu notarizations process exclusively financial and security identifiers (ISINs, price paise, volume counts) with zero investor Personally Identifiable Information.
- **Deterministic Fee Partitioning:** The 0.00% (Zero Fee) platform fee on all call auction turnover is mathematically partitioned with exact integer precision: $\text{Fee} = \text{Treasury} + \text{Core SGF} + \text{IPF}$.

## Acceptance Criteria
- [ ] Protobuf service contracts compile cleanly with `tonic-build` producing typed Rust stubs with zero compiler warnings.
- [ ] Dynamic 5-minute rolling VWAP reference prices update accurately on every tick within 1 microsecond.
- [ ] Tier 1 (+/- 5%), Tier 2 (+/- 10%), Tier 3 (+/- 20%), and 2x opening/closing phase multipliers calculate correct price boundaries.
- [ ] Straddle States and Limit States are detected in under 100 nanoseconds from incoming quote updates.
- [ ] 15-second Limit State timer initiates, holds, and triggers a Trading Pause if price does not revert inside bands.
- [ ] Intermediate 5-minute Call Auction aggregates limit orders and resolves Indicative Equilibrium Price (IEP) maximizing matched volume.
- [ ] Order Matching Engine smoothly transitions from Call Auction uncrossing back into continuous double-auction trading.
- [ ] Pre-trade margin liquidation daemons are suppressed during trading pauses and re-enabled post-auction.
- [ ] 0.00% (No fee at all) platform fee is assessed on all executed auction trades and split into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] Adheres strictly to the 12 mandatory sections with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `205` (Order Matching Engine), Prompt `206` (Pre-Trade Risk & Margin Engine), Prompt `207` (Real-Time Market Data Service).
- **Parallel Work:** Prompt `228` (Real-Time Market Surveillance & Dynamic Volatility Engine), Prompt `241` (Real-Time SPAN Portfolio Margin Engine), Prompt `255` (Unified Cross-Asset Portfolio Margining & Dynamic Collateral Optimization Service).
- **Subsequent Prompts Enabled:** Prompt `501` (Flutter Mobile Trading Interface), Prompt `601` (Web Trading Terminal).
