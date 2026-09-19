# 713 - Continuous Market Circuit Breaker & Resumption Coordinator

## Purpose
Establishes a mission-critical, ultra-low-latency Circuit Breaker and Volatility Control Coordinator designed specifically for continuous 24/7 fractional equity trading. In continuous round-the-clock trading without traditional market opening and closing bells, structural volatility risks, algorithmic runaway feedback loops, flash crashes, and cross-border liquidity fragmentation require continuous, deterministic price-band stabilization mechanisms.

Operating in full alignment with **SEBI Master Circular on Market Surveillance & Volatility Controls** and **IFSCA Market Infrastructure Regulations**, this service continuously monitors market-wide index levels, sector baskets, and individual stock price trajectories. It orchestrates dynamic volatility collars (DVC), multi-tiered circuit breaker halts (10%, 15%, 20%), automated call-auction price discovery intervals, and safe, transparent trading resumptions across both the Domestic Regulated Entity and the GIFT City International Gateway.

## What You Are Building
A fault-tolerant, high-performance distributed state machine and volatility governor (`services/circuit-breaker-coordinator`):
- **Continuous Dynamic Volatility Collar (DVC) Engine:** Evaluates rolling price movement bands (e.g. $\pm 3\%$ within 60 seconds or $\pm 5\%$ within 5 minutes) against dynamic reference prices (Volume-Weighted Average Price - VWAP) to trigger micro-volatility pauses.
- **System-Wide & Single-Stock Circuit Breaker State Machine:** Manages multi-level market-wide circuit breakers (10%, 15%, 20% index drops) and security-specific circuit limits (5%, 10%, 20% static/dynamic bands) with automated state transitions: `NORMAL_TRADING` $\to$ `VOLATILITY_INTERRUPT` $\to$ `CANCELLATION_ONLY` $\to$ `CALL_AUCTION_DISCOVERY` $\to$ `NORMAL_TRADING`.
- **Order Book State & Resumption Coordinator:** Directly controls Matching Engine (Prompt 205) partitions, purging synthetic resting quote spam, freezing trade execution while permitting order cancellations during cool-off phases.
- **Automated Call Auction Engine for Price Discovery:** Conducts randomized-duration pre-resumption call auctions (5 to 15 minutes) with non-continuous clearing to establish unmanipulated equilibrium opening prices.
- **Cross-Entity Synchronized Halt Broadcaster:** Enforces simultaneous trading halts across Domestic Regulated Entity and GIFT City IFSCA Ledger to prevent cross-border regulatory arbitrage and price divergence.

## Scope Boundaries
- **In Scope:**
 - Real-time price feed and index computation tracking for 100% of listed equities and composite fractional indices.
 - Multi-tier dynamic price band calculations (updated dynamically based on rolling 30-day volatility profiles).
 - Micro-second event publishing to freeze/unfreeze matching engine partitions.
 - Call auction order accumulation, uncrossing algorithm execution, and equilibrium price determination.
 - Multi-channel market halt alerts to all connected Flutter mobile, desktop, and web clients over WebSockets.
- **Out of Scope / Handled Elsewhere:**
 - Individual investor pre-trade margin limits (handled in Prompt 206).
 - Algorithmic manipulation pattern detection like spoofing/layering (handled in Prompt 711).
 - Off-chain physical share custodian settlement (handled in Prompt 208).

## Technology to Use
- **Core Engine & State Machine:** Rust for zero-cost abstractions, deterministic sub-millisecond execution, and memory safety without garbage collection pauses.
- **In-Memory Volatility Matrix & State Cache:** Redis 7 (Cluster) utilizing atomic Lua scripts and pub/sub channels for instantaneous state broadcasting across all matching engine nodes.
- **Consensus & Coordinated State Store:** Apache Kafka (with strict single-partition keying per ISIN) and Raft-backed distributed coordinator (etcd / embedded Raft) to guarantee strictly ordered, non-conflicting halt/resumption transitions.
- **Relational History & Regulatory Log:** PostgreSQL 16 for recording exact microsecond timestamps, trigger conditions, and auction metrics for every halt event.
- **Justification:** Flash crash mitigation requires deterministic execution in $< 1\text{ms}$. Rust guarantees zero GC pauses, ensuring that circuit breaker halts are executed before catastrophic cascade liquidations occur in downstream matching engines.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
 - `services/order-matching-engine` (Prompt 205): Stream of high-speed trade prices and LOB BBO quotes.
 - `services/market-data-service` (Prompt 207): Real-time fractional index calculations and VWAP feeds.
- **Downstream Microservices:**
 - `services/order-matching-engine` (Prompt 205): Ingests halt and call-auction control directives.
 - `services/order-service` (Prompt 204): Rejects aggressive market orders and queues limit orders during auctions.
 - `services/api-gateway` (Prompt 219) & `services/notification-service` (Prompt 211): Pushes market-wide halt banners and push notifications to all users.
- **Messaging Topics (Apache Kafka):**
 - Consumes: `market.trades.executed`, `market.index.ticks`, `market.vwap.updated`.
 - Publishes: `market.circuit_breaker.halt_triggered`, `market.circuit_breaker.auction_started`, `market.circuit_breaker.resumed`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **On-Chain Emergency Trading Pause Hook:** When a market-wide 20% circuit breaker is triggered, the coordinator emits an automated high-priority transaction to `MarketGovernanceHub.sol` or `DigitalSecurityTokenRegistry.sol`, pausing on-chain peer-to-peer transfers and smart contract settlement hooks (`pauseContract(bytes32 reasonCode, bytes signature)`).
- **Cryptographic Halt & Resumption Proofs:** The coordinator generates an ECDSA signed attestation for every halt and resumption event, containing the trigger price, index level, exact timestamp, and call-auction uncrossing price. This receipt is anchored to `ProofOfReserveRegistry.sol` for verifiable regulatory auditability.
- **Zero PII on Ledger:** All circuit breaker transactions are purely protocol-level and reference publicly visible market parameters (ISIN, Index Code, Circuit Level, Timestamps) with zero investor identity or PII data.

## Step-by-Step Build Instructions
1. Scaffold the `services/circuit-breaker-coordinator` project using Rust (with Tokio runtime and Actix/gRPC framework).
2. Define Protobuf definitions in `proto/growww/circuit_breaker/v1/circuit_breaker.proto` and build client libraries.
3. Implement the **Dynamic Volatility Collar (DVC) Engine**:
 - Maintain sliding window microsecond price ticks and compute 1-minute and 5-minute rolling VWAPs per security.
 - Calculate static price bands (e.g. $\pm 10\%$, $\pm 20\%$ of previous settlement base price) and dynamic bands ($\pm 3\%$ of rolling 5-minute VWAP).
4. Implement the **Market-Wide Circuit Breaker (MWCB) Engine**:
 - Monitor fractional equity benchmark indices (e.g. NIFTY-50 composite).
 - Enforce tiered circuit triggers:
 - **Level 1 (10% movement):** 45-minute trading halt $\to$ 15-minute call auction.
 - **Level 2 (15% movement):** 1-hour 45-minute trading halt $\to$ 15-minute call auction.
 - **Level 3 (20% movement):** Market suspension for remaining cycle $\to$ manual regulatory review.
5. Implement the **Single-Stock Circuit Breaker State Machine**:
 - Manage states per ISIN: `TRADING_ACTIVE`, `VOLATILITY_PAUSED`, `ORDER_CANCELLATION_ONLY`, `CALL_AUCTION_OPEN`, `CALL_AUCTION_MATCHING`, `RESUMPTION_BUFFER`.
6. Implement the **Order Book Control Dispatcher**:
 - On volatility trigger: Emit immediate `HaltCommand` over dedicated shared-memory/gRPC channel to the Matching Engine.
 - Matching Engine immediately suspends trade execution and updates order book state.
7. Implement the **Call Auction Equilibrium Price Calculator**:
 - Accumulate limit orders during the call auction window without continuous matching.
 - Execute maximum executable volume algorithm to determine uncrossing equilibrium price:
     $$\text{Equilibrium Price} = \arg\max_P \min\left(\sum_{p \ge P} Q_{\text{buy}}(p), \sum_{p \le P} Q_{\text{sell}}(p)\right)$$
 - Execute crossed orders at the determined equilibrium single price before transitioning back to normal trading.
8. Implement the **Cross-Entity Multi-Market Synchronizer**:
 - Broadcast simultaneous halt signals via authenticated mTLS gRPC to GIFT City IFSC matching nodes, guaranteeing synchronized state transitions across both jurisdictions.
9. Build the **WebSocket Market Status Broadcaster**:
 - Stream instantaneous market halt notifications, resumption countdown timers, and indicative call-auction prices to client frontend gateways.
10. Integrate **Hyperledger Besu On-Chain Pausing**:
 - Build automated relayer hook to invoke smart contract pause functions on Level 3 halts or emergency black-swan volatility events.
11. Build the **SEBI & IFSCA Statutory Halt Incident Logger**:
 - Automatically log every halt event, duration, trigger price, and auction statistics into PostgreSQL and dispatch alerts to compliance officers.
12. Build a high-throughput load and chaos testing harness verifying halt execution latency $< 1\text{ms}$ under 100,000 orders/second stress conditions.

## Interfaces / Contracts

### Protobuf Definition (`circuit_breaker.proto`)
```protobuf
syntax = "proto3";

package growww.circuit_breaker.v1;

option go_package = "github.com/growww/services/circuit-breaker-coordinator/gen/v1;circuitbreakerv1";

service CircuitBreakerService {
  rpc GetMarketStatus (MarketStatusRequest) returns (MarketStatusResponse);
  rpc GetSecurityCircuitBands (SecurityCircuitRequest) returns (SecurityCircuitResponse);
  rpc TriggerManualHalt (ManualHaltRequest) returns (ManualHaltResponse);
  rpc StreamCircuitBreakerEvents (StreamEventsRequest) returns (stream CircuitBreakerEvent);
}

enum MarketHaltState {
  STATE_NORMAL_TRADING = 0;
  STATE_VOLATILITY_INTERRUPT = 1;
  STATE_CANCELLATION_ONLY = 2;
  STATE_CALL_AUCTION = 3;
  STATE_SUSPENDED_LEVEL_3 = 4;
}

enum CircuitTriggerType {
  TRIGGER_NONE = 0;
  TRIGGER_DYNAMIC_COLLAR_SINGLE_STOCK = 1;
  TRIGGER_STATIC_BAND_SINGLE_STOCK = 2;
  TRIGGER_MARKET_WIDE_LEVEL_1_10_PCT = 3;
  TRIGGER_MARKET_WIDE_LEVEL_2_15_PCT = 4;
  TRIGGER_MARKET_WIDE_LEVEL_3_20_PCT = 5;
  TRIGGER_MANUAL_REGULATORY_INTERVENTION = 6;
}

message CircuitBreakerEvent {
  string event_id = 1;
  int64 timestamp_ns = 2;
  string isin = 3; // Empty for market-wide halts
  MarketHaltState previous_state = 4;
  MarketHaltState current_state = 5;
  CircuitTriggerType trigger_type = 6;
  double trigger_price = 7;
  double reference_vwap = 8;
  int64 estimated_resumption_utc = 9;
  string cryptographic_attestation_hash = 10;
}

message MarketStatusRequest {
  string market_entity = 1; // DOMESTIC / GIFT_CITY
}

message MarketStatusResponse {
  MarketHaltState overall_status = 1;
  int64 active_halts_count = 2;
  int64 timestamp_utc = 3;
}

message SecurityCircuitRequest {
  string isin = 1;
}

message SecurityCircuitResponse {
  string isin = 1;
  MarketHaltState state = 2;
  double lower_circuit_limit = 3;
  double upper_circuit_limit = 4;
  double dynamic_collar_lower = 5;
  double dynamic_collar_upper = 6;
  double current_vwap = 7;
  double last_traded_price = 8;
  int64 last_state_change_utc = 9;
}

message ManualHaltRequest {
  string isin = 1;
  string compliance_officer_id = 2;
  string justification = 3;
  int32 halt_duration_minutes = 4;
  string mfa_token = 5;
}

message ManualHaltResponse {
  bool success = 1;
  string event_id = 2;
  int64 halt_start_utc = 3;
}

message StreamEventsRequest {
  string market_entity = 1;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE circuit_breaker_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_entity VARCHAR(32) NOT NULL CHECK (market_entity IN ('DOMESTIC', 'GIFT_CITY', 'ALL')),
    isin VARCHAR(12), -- NULL indicates market-wide halt
    trigger_type VARCHAR(64) NOT NULL,
    trigger_price NUMERIC(18, 4) NOT NULL,
    reference_vwap NUMERIC(18, 4) NOT NULL,
    percentage_deviation NUMERIC(6, 3) NOT NULL,
    halt_started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    auction_started_at TIMESTAMPTZ,
    resumed_at TIMESTAMPTZ,
    uncrossing_price NUMERIC(18, 4),
    uncrossed_volume NUMERIC(18, 4),
    blockchain_tx_hash VARCHAR(66),
    initiated_by VARCHAR(64) NOT NULL DEFAULT 'SYSTEM_AUTOMATED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE security_volatility_parameters (
    isin VARCHAR(12) PRIMARY KEY,
    base_price NUMERIC(18, 4) NOT NULL,
    static_band_pct NUMERIC(5, 2) NOT NULL DEFAULT 10.00,
    dynamic_collar_pct NUMERIC(5, 2) NOT NULL DEFAULT 3.00,
    auction_duration_seconds INT NOT NULL DEFAULT 600,
    is_halted BOOLEAN NOT NULL DEFAULT FALSE,
    current_state VARCHAR(32) NOT NULL DEFAULT 'STATE_NORMAL_TRADING',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cb_events_isin ON circuit_breaker_events (isin, halt_started_at DESC);
CREATE INDEX idx_cb_events_started ON circuit_breaker_events (halt_started_at DESC);
```

## Security & Compliance Notes
- **SEBI Volatility Management Mandate:** Ensures 100% adherence to SEBI norms for dynamic price bands, preventing runaway automated trading algorithms from causing catastrophic cascading liquidations.
- **Fail-Safe Conservative Default:** If communication breaks down between the Circuit Breaker Coordinator and any Matching Engine partition, that partition automatically falls back to `STATE_CANCELLATION_ONLY` to protect investor capital.
- **Immutable Timestamping & Attestation:** All state transitions must be stamped with synchronized PTP/NTP clocks ($< 10\mu\text{s}$ drift) and anchored to the permissioned blockchain.
- **Strict Maker-Checker Manual Override:** Emergency manual trading halts or early resumptions require cryptographically validated dual-approvals from Compliance and Risk Heads.

## Acceptance Criteria
- [ ] Sub-millisecond volatility trigger detection: executes `HaltCommand` to matching engine within $< 1\text{ms}$ of price breaching collar boundary.
- [ ] Accurately computes rolling 1-minute and 5-minute VWAP with zero tick drops across 50,000 trades/second.
- [ ] Successfully manages state transitions from `NORMAL_TRADING` $\to$ `VOLATILITY_INTERRUPT` $\to$ `CALL_AUCTION` $\to$ `RESUMED_TRADING` across synthetic chaos scenarios.
- [ ] Call auction uncrossing engine correctly identifies the maximum-executable volume equilibrium price in 100% of test book scenarios.
- [ ] Broadcasts synchronized halt signals to GIFT City IFSC node within $< 10\text{ms}$, preventing cross-jurisdiction price divergence.
- [ ] Triggers automated on-chain smart contract pause on Hyperledger Besu upon Level 3 market-wide circuit breaker.
- [ ] Integration test suite achieves $\ge 92\%$ code coverage across all state machine edge cases.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Event Schemas), Prompt 205 (Order Matching Engine), Prompt 206 (Risk & Margin Checks), Prompt 207 (Market Data Service).
- **Subsequent / Parallel Tasks:** Prompt 711 (Market Surveillance Engine), Prompt 712 (Insider Trading Analytics), Prompt 233 (Continuous 24x7 Regulatory Reporting).
