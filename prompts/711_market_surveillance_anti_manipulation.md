# 711 - Market Surveillance & Anti-Manipulation Engine (Spoofing, Layering, Wash Trading, Momentum Ignition)

## Purpose
Establishes a real-time, ultra-low-latency market surveillance and algorithmic anti-manipulation engine operating across continuous 24/7 fractional equity trading. In continuous round-the-clock trading environments, market microstructure is vulnerable to high-frequency manipulative order-book tactics, predatory liquidity games, and coordinated price distortion schemes.

This service continuously ingests full L2/L3 order book depth, order placement/modification/cancellation telemetry, and trade executions to detect, score, and neutralize abusive trading practices in sub-second timeframes. It targets four primary market abuse typologies defined under the **SEBI (Prohibition of Fraudulent and Unfair Trade Practices relating to Securities Market) Regulations, 2003 (PFUTP)** and **IFSCA (Market Conduct) Regulations**:
1. **Spoofing & Phantom Liquidity:** Placing large non-bona fide orders with the intention to cancel before execution to create a false impression of supply/demand.
2. **Layering:** Submitting multiple tiered limit orders across multiple price books away from the BBO (Best Bid/Offer) to manipulate the visible depth ladder, then executing opposite-side orders and immediately cancelling the layers.
3. **Wash Trading & Self-Matching:** Coordinated or unilateral execution of trades where there is no genuine change in beneficial ownership (direct self-trades, cross-account collusive rings, or matched orders between affiliated PAN/entity clusters).
4. **Momentum Ignition & Quote Stuffing:** Rapid bursts of aggressive market orders and micro-second cancellations intended to spike volatility, trigger stop-losses or automated liquidations, and exploit artificial price momentum.

## What You Are Building
A distributed, high-throughput stream processing and surveillance microservice (`services/market-surveillance`):
- **High-Velocity Order Book Reconstructor & Feature Extractor:** Reconstructs microsecond-level limit order book (LOB) state per ISIN from Kafka market data feeds, computing dynamic order-to-trade ratios (OTR), cancel-replace velocities, and book imbalance skews.
- **Complex Event Processing (CEP) Manipulation Detector:** Stateful stream processing pipelines evaluating pattern recognition heuristics and sliding-window statistical anomaly models across 100ms, 1s, 10s, 1m, and 15m windows.
- **Beneficial Ownership Ring Correlator:** In-memory graph cross-matcher linking order submissions across disparate client IDs sharing common beneficial owners (KYC PAN, bank accounts, device hardware fingerprints, IP subnets, or MAC addresses).
- **Automated Protective Throttle & Order Rejection Gateway:** Interfacing with the Pre-Trade Risk Engine (Prompt 206) and Matching Engine (Prompt 205) to enforce algorithmic order throttles, cancel abusive resting quotes, and trigger dynamic surveillance margin multipliers.
- **Surveillance Case Management & SEBI Alert Dossier Generator:** Generating cryptographically anchored surveillance alert packages ready for statutory escalation and compliance officer review.

## Scope Boundaries
- **In Scope:**
 - Real-time ingestion and stateful analysis of 100% of order lifecycles (`OrderCreated`, `OrderModified`, `OrderCanceled`, `OrderMatched`).
 - Microstructure metric computation: Cancel-to-Fill Ratio (CFR), Order-to-Trade Ratio (OTR), Book Depth Imbalance (BDI), and Weighted Midpoint Slippage.
 - Multi-tiered heuristic algorithms for Spoofing, Layering, Wash Trading, Marking the Close/Open, Quote Stuffing, and Momentum Ignition.
 - Automated low-latency intervention hooks: sending synthetic throttle/freeze signals to API Gateway (Prompt 220) and Risk Engine (Prompt 206).
 - Production of tamper-evident surveillance alert events and forensic playback packages.
- **Out of Scope / Handled Elsewhere:**
 - In-memory order matching and book execution (handled in Prompt 205).
 - Pre-trade fund reservation and balance ledgering (handled in Prompts 203 & 206).
 - Long-term off-chain graph analytics for insider trading/UPSI networks (handled in Prompt 712).
 - System-wide market volatility circuit breakers and halt coordination (handled in Prompt 713).

## Technology to Use
- **Stream Processing Core:** Rust / Go with Apache Flink 1.19+ / Rust `timely-dataflow` for deterministic, stateful event stream processing with out-of-order event handling via event-time watermarks.
- **Columnar Analytical Engine:** ClickHouse 24+ for raw tick and order-event time-series storage, enabling microsecond LOB historical replay and backtesting.
- **In-Memory Microstructure State:** Redis 7 Cluster (Redis Enterprise / DragonFly) utilizing Sliding Window aggregations, Circular Ring Buffers, and HyperLogLog structures for high-speed counter tracking.
- **Relational Surveillance Case Store:** PostgreSQL 16 with TimescaleDB extension for operational alerts, regulatory investigation logs, and compliance audit states.
- **Justification:** Detecting layering and spoofing requires processing hundreds of thousands of order events per second with sub-millisecond evaluation latency. Rust and Apache Flink provide deterministic exactly-once stream processing and memory safety, while ClickHouse enables sub-second analytical queries across billions of historical order-lifecycle records.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
 - `services/order-service` (Prompt 204): Stream of inbound order requests.
 - `services/order-matching-engine` (Prompt 205): Internal tick execution stream, order book depth snapshots, and cancellation confirmations.
 - `services/user-service` & `services/kyc-service` (Prompts 201 & 202): Investor PAN, beneficial ownership, device metadata, and geolocation attributes.
- **Downstream Microservices:**
 - `services/risk-engine` (Prompt 206): Real-time pre-trade risk multiplier updates and user-level order submission blocks.
 - `services/api-gateway` (Prompt 219 & 220): IP/Device rate-limiting and connection termination.
 - `services/admin-back-office` (Prompt 217 & 604): Surveillance analyst workbench and investigation queue.
- **Messaging Topics (Apache Kafka):**
 - Consumes: `market.orders.lifecycle`, `market.trades.executed`, `market.lob.snapshot_10ms`, `user.auth.session_fingerprint`.
 - Publishes: `surveillance.alerts.manipulation`, `surveillance.actions.throttle_trader`, `surveillance.case.opened`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **On-Chain Forensic Evidence Anchoring:** When manipulative patterns meet high-conviction severity thresholds ($\text{Confidence} \ge 95\%$), the service computes a cryptographic SHA-256 Merkle tree root of the entire order-book lifecycle snapshot (including timestamps, message sequence numbers, and anonymized trader addresses) and writes it to `MarketSurveillanceRegistry.sol`.
- **On-Chain Coordinated Freezing:** For flagrant systemic market abuse (e.g., automated high-frequency wash-trading rings attempting to siphon liquidity or manipulate token reference prices), the surveillance engine dispatches a high-priority gRPC call to the smart contract relayer to trigger `ComplianceRegistry.freezeTradingAccount(address, reasonCode, proofHash)` on-chain.
- **Zero PII on Ledger:** All blockchain events reference deterministic cryptographic hashes of investor identifiers (`investor_account_hash = HMAC-SHA256(pan_number, salt)`), ensuring zero PII enters the immutable ledger.

## Step-by-Step Build Instructions
1. Scaffold the `services/market-surveillance` repository with modular architecture: `ingestor/`, `reconstructor/`, `detectors/`, `scoring/`, `actions/`, and `storage/`.
2. Define Protobuf definitions in `proto/growww/surveillance/v1/surveillance.proto` and generate gRPC stubs.
3. Configure Kafka stream ingestion pipelines for `market.orders.lifecycle` and `market.trades.executed` with strict ISIN-keyed partition ordering.
4. Implement the **LOB Microstructure Reconstructor**:
 - Maintain sliding in-memory L2/L3 order book state per security.
 - Compute real-time Order-to-Trade Ratio (OTR) and Cancel-to-Fill Ratio (CFR) over 1s, 10s, 60s, and 300s rolling windows.
5. Implement the **Spoofing Detection Engine**:
 - Track large quote injections ($\ge 5\times$ average depth size at level 1-3) followed by opposite-side trades and subsequent cancellation within $< 500\text{ms}$.
 - Flag non-bona fide quote persistence where average lifetime of quotes placed at or outside BBO is $< 100\text{ms}$ with zero fill intent.
6. Implement the **Layering Detection Engine**:
 - Detect multiple limit orders ($> 3$ layers) submitted on one side of the order book, followed by an aggressive market order on the opposite side, immediately followed by bulk cancellation ($> 90\%$ volume) of the layered orders within $\le 250\text{ms}$.
7. Implement the **Wash Trading & Self-Trade Engine**:
 - Cross-correlate buyer and seller identifiers in real time using the in-memory PAN and Beneficial Ownership graph cache.
 - Detect matching trades between different account IDs sharing identical PAN, primary bank account, hardware device UUID, or shared network session.
8. Implement the **Momentum Ignition & Quote Stuffing Engine**:
 - Identify bursts of rapid order placement and cancellation ($> 50$ events/second by a single entity/cluster) that cause artificial price moves of $> 1.5\%$ within 5 seconds without external fundamental news flow.
9. Build the **Composite Manipulation Risk Scorer**:
 - Aggregate multi-typology signals into a unified anomaly score ($0.0 - 100.0$) using weighted statistical deviation ($Z\text{-score} > 3.5$) and historical behavioral baselines.
10. Implement the **Automated Protective Action Dispatcher**:
 - If Risk Score $\ge 75$: Emit throttle directive to `services/risk-engine` (increase pre-trade margin requirement by $200\%$, restrict aggressive order types).
 - If Risk Score $\ge 90$: Automatically cancel all active resting quotes of the offender, suspend order submission capabilities, and trigger compliance alert.
11. Build the **ClickHouse Ingestion & Snapshot Archival Pipeline**:
 - Persist all raw order events, reconstructed LOB states, and detection markers into partitioned ClickHouse tables with 8-year WORM retention policies.
12. Integrate the **SEBI PFUTP Export Module**:
 - Format alerts into standardized regulatory reporting dossiers containing order chronologies, depth ladder heatmaps, counterparty breakdowns, and cryptographic evidence hashes.

## Interfaces / Contracts

### Protobuf Definition (`surveillance.proto`)
```protobuf
syntax = "proto3";

package growww.surveillance.v1;

option go_package = "github.com/growww/services/market-surveillance/gen/v1;surveillancev1";

service MarketSurveillanceService {
  rpc EvaluateOrderEvent (EvaluateOrderEventRequest) returns (EvaluateOrderEventResponse);
  rpc GetSurveillanceAlert (GetSurveillanceAlertRequest) returns (GetSurveillanceAlertResponse);
  rpc ListActiveAlerts (ListActiveAlertsRequest) returns (ListActiveAlertsResponse);
  rpc OverrideSurveillanceAction (OverrideActionRequest) returns (OverrideActionResponse);
}

enum ManipulationPattern {
  PATTERN_UNSPECIFIED = 0;
  SPOOFING = 1;
  LAYERING = 2;
  WASH_TRADING_DIRECT = 3;
  WASH_TRADING_COLLUSIVE = 4;
  MOMENTUM_IGNITION = 5;
  QUOTE_STUFFING = 6;
  MARKING_THE_CLOSE = 7;
}

enum ActionLevel {
  ACTION_NONE = 0;
  ACTION_FLAG_ALERT = 1;
  ACTION_THROTTLE_RATE = 2;
  ACTION_RESTRICT_AGGRESSIVE_ORDERS = 3;
  ACTION_CANCEL_RESTING_QUOTES = 4;
  ACTION_FREEZE_ACCOUNT = 5;
}

message OrderLifecycleEvent {
  string event_id = 1;
  int64 timestamp_ns = 2;
  string order_id = 3;
  string investor_id = 4;
  string pan_hash = 5;
  string isin = 6;
  string side = 7; // BUY / SELL
  string order_type = 8; // LIMIT / MARKET / IOC
  double price = 9;
  double quantity = 10;
  string event_type = 11; // NEW / AMEND / CANCEL / FILL
  string device_fingerprint = 12;
  string client_ip = 13;
}

message SurveillanceAlert {
  string alert_id = 1;
  int64 detected_at_ns = 2;
  string isin = 3;
  ManipulationPattern pattern = 4;
  double confidence_score = 5; // 0.0 to 100.0
  repeated string involved_trader_hashes = 6;
  repeated string related_order_ids = 7;
  repeated string executed_trade_ids = 8;
  ActionLevel enforced_action = 9;
  string evidence_snapshot_hash = 10;
  string summary_narrative = 11;
}

message EvaluateOrderEventRequest {
  OrderLifecycleEvent event = 1;
}

message EvaluateOrderEventResponse {
  bool is_flagged = 1;
  ActionLevel recommended_action = 2;
  double current_risk_score = 3;
}

message GetSurveillanceAlertRequest {
  string alert_id = 1;
}

message GetSurveillanceAlertResponse {
  SurveillanceAlert alert = 1;
}

message ListActiveAlertsRequest {
  string isin = 1;
  ManipulationPattern pattern_filter = 2;
  int64 start_time = 3;
  int64 end_time = 4;
}

message ListActiveAlertsResponse {
  repeated SurveillanceAlert alerts = 1;
}

message OverrideActionRequest {
  string alert_id = 1;
  string compliance_officer_id = 2;
  string justification = 3;
  ActionLevel new_action_level = 4;
}

message OverrideActionResponse {
  bool success = 1;
  string updated_status = 2;
}
```

### ClickHouse & PostgreSQL Data Schemas
```sql
-- ClickHouse OLAP Table for Raw Order Stream Replay (ClickHouse 24+)
CREATE TABLE default.surveillance_order_events (
    event_time DateTime64(6, 'UTC'),
    event_date Date DEFAULT toDate(event_time),
    isin LowCardinality(String),
    order_id UUID,
    investor_pan_hash FixedString(64),
    side Enum8('BUY' = 1, 'SELL' = 2),
    event_type Enum8('NEW' = 1, 'AMEND' = 2, 'CANCEL' = 3, 'FILL' = 4),
    price Decimal64(4),
    quantity Decimal64(4),
    lifetime_ms UInt32,
    distance_from_bbo_bps Int32,
    device_fingerprint_hash FixedString(64),
    ip_subnet FixedString(16)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (isin, event_time, order_id)
TTL event_date + INTERVAL 8 YEAR;

-- PostgreSQL Surveillance Operational Case Management Table
CREATE TABLE surveillance_investigation_cases (
    case_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL,
    isin VARCHAR(12) NOT NULL,
    pattern_type VARCHAR(64) NOT NULL,
    severity_score NUMERIC(5, 2) NOT NULL,
    primary_suspect_hash VARCHAR(64) NOT NULL,
    all_related_accounts JSONB NOT NULL,
    evidence_package_uri VARCHAR(256) NOT NULL,
    blockchain_proof_hash VARCHAR(64),
    case_status VARCHAR(32) NOT NULL DEFAULT 'OPEN', -- OPEN / UNDER_REVIEW / SEBI_ESCALATED / CLOSED_FALSE_POSITIVE
    assigned_analyst_id VARCHAR(64),
    analyst_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_surveillance_cases_isin ON surveillance_investigation_cases (isin, created_at DESC);
CREATE INDEX idx_surveillance_cases_status ON surveillance_investigation_cases (case_status);
```

## Security & Compliance Notes
- **SEBI PFUTP Regulations 2003:** Algorithms must strictly enforce detection rules specified under SEBI Circulars for Algorithmic Trading and Market Surveillance, ensuring automated audit preservation of manipulative patterns.
- **DPDP Act 2023 & Zero PII Protection:** Offender identification within analytical streams and blockchain proofs must strictly utilize salted cryptographic hashes (`pan_hash`, `device_fingerprint_hash`), preserving customer privacy while retaining investigative cross-correlation.
- **WORM Storage & Anti-Tampering:** Raw market tick feeds and surveillance alert snapshots must be committed to S3 Object Lock storage with 8-year retention lock matching statutory regulatory audit mandates.
- **Segregation of Duties & Maker-Checker Overrides:** Automated surveillance actions (such as account trading bans or throttle relief) can only be overridden by two authorized compliance officers through authenticated dual-key administrative signatures.

## Acceptance Criteria
- [ ] Ingests 100% of order lifecycle events (`NEW`, `AMEND`, `CANCEL`, `FILL`) with an end-to-end detection latency of $< 15\text{ms}$ at 50,000 events/sec throughput.
- [ ] Correctly identifies Spoofing patterns in test suite with $> 98\%$ accuracy and $< 0.5\%$ false positive rate.
- [ ] Accurately flags Layering patterns (3+ tiered quotes canceled within 250ms of opposite fill) across multi-account synthetic test books.
- [ ] Detects 100% of Wash Trading scenarios between accounts sharing identical PAN, primary bank IFSC/account, or hardware device hash.
- [ ] Automatically issues pre-trade risk throttle signals to `services/risk-engine` within $< 5\text{ms}$ of threshold violation.
- [ ] Generates valid SEBI-compliant forensic audit packages with verifiable Merkle tree cryptographic proof hashes anchored to Hyperledger Besu.
- [ ] Unit and integration test coverage across all pattern detectors exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Event Schemas & Kafka), Prompt 204 (Order Service), Prompt 205 (Order Matching Engine), Prompt 206 (Risk & Margin Checks).
- **Subsequent / Parallel Tasks:** Prompt 712 (Insider Trading & UPSI Graph Analytics), Prompt 713 (Continuous Market Circuit Breaker), Prompt 233 (Continuous 24x7 Regulatory Reporting).
