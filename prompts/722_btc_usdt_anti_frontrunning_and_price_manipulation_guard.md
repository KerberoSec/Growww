# 722 - BTC/USDT Anti-Frontrunning & Price Manipulation Surveillance Guard (Go)

## Purpose
Establishes an ultra-low-latency, real-time market integrity monitoring daemon and deterministic order gating service (`services/market-integrity-guard`) engineered specifically to safeguard the BTC/USDT trading pair. In 24/7 digital asset and cross-border tokenized spot markets, the BTC/USDT pair serves as the primary liquidity anchor, making it the highest-priority target for predatory algorithmic strategies. These include high-frequency latency arbitrage against external reference markets, quote spoofing, order book layering, wash trading, and manipulative reference-price distortions.

Operating in alignment with the **SEBI (Prohibition of Fraudulent and Unfair Trade Practices relating to Securities Market) Regulations, 2003 (PFUTP)**, the **IFSCA (Market Conduct) Regulations**, and international market integrity recommendations (IOSCO Principles for the Regulation and Supervision of Digital Asset Markets), this service maintains fair, orderly, and transparent price formation. It neutralizes unfair microsecond structural advantages by decoupling aggressive order routing from passive quote provision, verifying internal liquidity against consolidated global spot benchmarks, detecting collusive trading rings, and executing automated circuit-breaker protections.

## What You Are Building
A high-speed, concurrent Go surveillance daemon (`services/market-integrity-guard`) operating inline and alongside the execution pathway:
- **Asymmetric Speed Bump & Deterministic Order Gatekeeper:** Enforces an exact 500-microsecond asymmetric latency buffer on aggressive taker orders (market orders, marketable limit orders, IOC, FOK) while permitting passive liquidity-providing maker quotes (post-only and non-marketable limit orders) to reach the order book immediately ($0\mu\text{s}$ delay). This protects resting retail and market-maker quotes from predatory latency arbitrageurs exploiting price shifts on external global exchanges.
- **Real-Time External VWAP & Price Basis Deviation Engine:** Ingests high-frequency consolidated spot tick streams from external liquidity venues (Binance, Coinbase, OKX, Kraken) via Market Feeder (Prompt 272). Computes rolling Volume-Weighted Average Prices (VWAP) and median reference prices across 100ms, 1s, 5s, and 60s windows, calculating instantaneous basis point divergence ($\Delta_{\text{bps}}$) against the internal BTC/USDT Best Bid and Offer (BBO) midpoint.
- **Automated Volatility Halt & Circuit-Breaker Trigger:** Detects anomalous price divergence between internal order book execution levels and global reference feeds. When basis deviation exceeds calibrated regulatory and microstructural bands (e.g., $> 150\text{ bps}$ for $250\text{ms}$ or $> 300\text{ bps}$ instantaneous), the guard dispatches microsecond halt signals to the Matching Engine (Prompt 205) and Risk Service (Prompt 206) to freeze matching and initiate protective call auctions.
- **In-Memory Microstructure Pattern Detector (Spoofing, Layering, Quote Stuffing):** Reconstructs the internal Limit Order Book (LOB) depth state and evaluates order cancellation velocities, life-cycle durations, and depth imbalance skews to detect phantom liquidity, spoofing (< 200ms lifespan quotes canceled prior to fills), and multi-tier layering schemes.
- **Coordinated Wash Trading & Self-Matching Interceptor:** Intercepts potential executions in real time, cross-correlating buyer and seller identity parameters in Redis (salted PAN hashes, beneficial owner clusters, bank account hashes, and device fingerprints) to reject direct self-trades and flag collusive cyclic wash rings.
- **Audit Telemetry Streamer & Hyperledger Besu Audit Anchor:** Emits nanosecond-timestamped telemetry events to ClickHouse for regulatory audit replay and records cryptographic certificates of circuit-breaker halts on the Hyperledger Besu permissioned ledger.

## Scope Boundaries
- **In Scope:**
  - Inline and sidecar microsecond surveillance of 100% of BTC/USDT order submissions, cancellations, modifications, and trade executions.
  - Implementation of the 500-microsecond asymmetric delay queue for aggressive taker orders prior to Matching Engine ingestion.
  - Sub-millisecond calculation of rolling VWAP and basis divergence against external spot market feeds.
  - Automated triggering and coordination of micro-volatility pauses and circuit-breaker trading halts for the BTC/USDT pair.
  - Microsecond detection heuristics for Spoofing, Layering, Quote Stuffing, and Cross-Account Wash Trading.
  - Real-time containment signaling to the Pre-Trade Risk Engine (Prompt 206) to throttle abusive accounts.
  - Nanosecond telemetry ingestion into ClickHouse with partitioned historical WORM storage.
  - On-chain cryptographic attestation of market-halt events to the Hyperledger Besu audit registry.
- **Out of Scope / Handled Elsewhere:**
  - Core order matching, trade clearing, and order book sequencing (handled in Prompt 205).
  - External exchange WebSocket ingestion, protocol parsing, and normalization (handled in Prompt 272).
  - Pre-trade investor account balance reservations and margin adequacy verification (handled in Prompt 206).
  - Multi-asset long-term macro graph analytics across weeks or months (handled in Prompt 719).
  - Off-chain fiat banking wire settlement and payment gateway clearing (handled in Prompts 212 and 215).

## Technology to Use
- **Language & Runtime:** Go 1.22+ utilizing Goroutines, lock-free ring buffers, zero-allocation byte buffers, and optimized memory management (`GOMEMLIMIT`, tuned `GOGC`) to eliminate non-deterministic Garbage Collection pauses.
- **In-Memory Microstructure & Identity Store:** Redis 7.2 Cluster utilizing Redis Streams, atomic Lua scripts, and sorted sets (`ZSET`) for tracking active quote lifetimes, rolling cancel counters, and account cluster mappings in sub-millisecond lookups.
- **Event Streaming Backbone:** Apache Kafka 3.7+ with strict single-partition ordering keyed by currency pair (`BTC_USDT`) for low-latency event transport.
- **Columnar Analytical Telemetry Store:** ClickHouse 24+ for raw nanosecond order-event telemetry, tick-by-tick VWAP deviation logging, and 8-year statutory WORM compliance retention.
- **Consortium Blockchain:** Hyperledger Besu (QBFT consensus, 1-second block time) for anchoring immutable circuit-breaker halt receipts and surveillance proof hashes.
- **Justification:** Go 1.22 provides predictable execution speed, superior native networking concurrency, and low memory overhead without JVM garbage collection spikes or C++ memory safety vulnerabilities. It guarantees that the 500-microsecond asymmetric speed bump operates with microsecond precision, while Redis, Kafka, and ClickHouse provide an end-to-end resilient architecture capable of processing over 50,000 events per second.

## Backend / Infra Touchpoints
- **Upstream Microservices & Feeds:**
  - `services/market-feeder` (Prompt 272): Delivers real-time external consolidated spot tick updates and reference prices (`market.external.ticks.btc_usdt`).
  - `services/order-service` (Prompt 204): Publishes raw inbound client orders submitted for the BTC/USDT pair (`order.inbound.btc_usdt`).
  - `services/order-matching-engine` (Prompt 205): Emits internal trade execution prints and LOB depth deltas (`market.internal.trades.btc_usdt`, `market.internal.lob.btc_usdt`).
  - `services/kyc-aml-service` (Prompt 202) & `services/user-service` (Prompt 201): Supplies updated identity cross-reference hashes, PAN hashes, and device fingerprints (`user.identity.hashes.updated`).
- **Downstream Services & Control Hooks:**
  - `services/order-matching-engine` (Prompt 205): Consumes speed-bump-delayed aggressive orders (`market.gated.orders.btc_usdt`) and halt control directives (`market.control.halt.btc_usdt`).
  - `services/risk-and-margin-checks-service` (Prompt 206): Ingests automated trader throttle directives and account suspension signals.
  - `services/api-gateway` (Prompt 219) & `services/notification-service` (Prompt 211): Receives volatility halt alerts to broadcast status banners to web and mobile clients.
  - `services/admin-back-office` (Prompt 217): Presents real-time surveillance dashboards and alert queues for compliance officer review.
- **Kafka Topics:**
  - Consumes:
    - `market.external.ticks.btc_usdt` (External spot ticks from Binance, Coinbase, OKX, Kraken)
    - `order.inbound.btc_usdt` (Inbound client orders)
    - `market.internal.trades.btc_usdt` (Internal trade executions)
    - `market.internal.lob.btc_usdt` (Internal L2/L3 order book depth updates)
    - `user.identity.hashes.updated` (Investor identity mapping updates)
  - Publishes:
    - `market.gated.orders.btc_usdt` (Orders released from speed bump to matching engine)
    - `surveillance.alerts.btc_usdt` (Manipulation and front-running alert events)
    - `market.control.halt.btc_usdt` (Circuit-breaker halt and resumption directives)
    - `telemetry.surveillance.clickhouse` (Microsecond telemetry for ClickHouse ingestion)

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus with 1-second block finality.
- **Immutable Circuit-Breaker Halt Audit Trail:** When an automated volatility halt, circuit-breaker freeze, or severe market manipulation containment action is executed on the BTC/USDT pair, the daemon compiles an ECDSA-signed cryptographic halt certificate and logs it to `MarketIntegrityAuditRegistry.sol` via RPC.
- **Logged Event Attributes:**
  - `pairSymbol`: `BTC/USDT`
  - `haltType`: Enumerated trigger condition (`VOLATILITY_DEVIATION_HALT`, `EXTERNAL_FEED_STALE_HALT`, `COORDINATED_MANIPULATION_FREEZE`)
  - `internalPriceBps`: Internal order book midpoint at trigger time
  - `externalVwapBps`: Consolidated external reference VWAP at trigger time
  - `deviationBps`: Absolute basis point divergence
  - `triggerTimestampNs`: Nanosecond Unix timestamp of the event
  - `evidenceMerkleRoot`: SHA-256 Merkle tree root of the preceding 1,000 order book state deltas and external reference ticks
- **Zero PII on Ledger:** In compliance with DPDP Act 2023 and global privacy standards, no personally identifiable information (PII), plain-text user IDs, or IP addresses are recorded on-chain. Suspicious account actions reference only deterministic one-way cryptographic hashes (`account_hash = HMAC_SHA256(user_uuid, consortium_salt)`).
- **Statutory Evidentiary Standard:** The on-chain receipt provides non-repudiable legal proof to regulatory authorities (SEBI, IFSCA) that trading halts were initiated deterministically by the automated surveillance system in response to objective mathematical thresholds, ruling out operator bias or discretionary front-running.

## Step-by-Step Build Instructions
1. Initialize Go module `services/market-integrity-guard` adhering to clean architecture: `cmd/guard/`, `internal/speedbump/`, `internal/vwap/`, `internal/detectors/`, `internal/halt/`, `internal/blockchain/`, `internal/storage/`, and `pkg/types/`.
2. Generate Go stubs and Protobuf serializers from `surveillance_alert.proto`, `price_deviation.proto`, and `market_control.proto` using `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc`.
3. Implement the **Asymmetric Speed Bump Gating Engine (`internal/speedbump`)**:
   - Construct a lock-free, zero-allocation ring buffer queue consuming from `order.inbound.btc_usdt`.
   - Inspect order execution flags: passive liquidity makers (limit orders without immediate fill capability, post-only orders) bypass the buffer and route to `market.gated.orders.btc_usdt` with zero delay ($0\mu\text{s}$).
   - Enforce an exact 500-microsecond delay on aggressive taker orders (market orders, marketable limit orders, IOC, FOK) using a high-resolution monotonic timer wheel (`time.AfterFunc` or spin-wait scheduler).
   - If a market-wide halt or significant reference price update occurs during the 500-microsecond hold window, re-evaluate order validity before release to prevent stale resting limit orders from being sniped.
4. Implement the **Consolidated External VWAP Engine (`internal/vwap`)**:
   - Ingest normalized trade ticks and BBO feeds from `services/market-feeder` (Prompt 272) across external venues (Binance, Coinbase, OKX, Kraken).
   - Maintain sliding circular buffers to compute volume-weighted average price: $\text{VWAP} = \frac{\sum (P_i \times V_i)}{\sum V_i}$ across rolling windows of 100ms, 1s, 5s, and 60s.
   - Calculate median external reference price to discard anomalous outlier prints from individual exchange disconnects.
5. Implement the **Price Basis Deviation & Arbitrage Tracker**:
   - Ingest internal execution ticks and L1 BBO midpoint from `services/order-matching-engine` (Prompt 205).
   - Compute instantaneous basis point divergence:
     $$\Delta_{\text{bps}} = \left| \frac{\text{Price}_{\text{internal}} - \text{VWAP}_{\text{external}}}{\text{VWAP}_{\text{external}}} \right| \times 10000$$
   - Emit `PriceDeviationEvent` records to Kafka when $\Delta_{\text{bps}} \ge 50\text{ bps}$.
6. Implement the **Automated Volatility Halt Controller (`internal/halt`)**:
   - Multi-tier circuit-breaker rules:
     - Level 1 (Warning / Pre-Halt): $\Delta_{\text{bps}} > 100\text{ bps}$ sustained for $500\text{ms} \implies$ expand taker speed bump to $1000\mu\text{s}$ and emit operator alert.
     - Level 2 (Micro-Volatility Pause): $\Delta_{\text{bps}} > 150\text{ bps}$ sustained for $250\text{ms} \implies$ trigger automated 60-second execution pause (`VOLATILITY_PAUSE`), allowing order cancellations while rejecting aggressive fills.
     - Level 3 (Circuit-Breaker Halt): $\Delta_{\text{bps}} > 300\text{ bps}$ instantaneous $\implies$ trigger system-wide trading halt (`CIRCUIT_BREAKER_HALT`), suspending matching and transitioning order book to randomized pre-resumption call auction.
   - Publish `market.control.halt.btc_usdt` directly to the Matching Engine over ultra-low-latency IPC/gRPC.
7. Implement the **Anti-Spoofing Detection Engine (`internal/detectors/spoofing`)**:
   - Track depth additions at top 5 book levels. Flag large quote injections ($\ge 3\times$ average depth size) with active resting lifetimes $< 200\text{ms}$ that are canceled without receiving fills.
   - Detect "quote flashing" where high-volume quotes are placed and canceled within single-digit milliseconds to create phantom depth.
8. Implement the **Anti-Layering Detection Engine (`internal/detectors/layering`)**:
   - Detect tiered limit order submissions ($\ge 3$ price levels away from BBO on one side) followed by aggressive market order executions on the opposite side, immediately followed by bulk cancellation of the tiered layers within $\le 200\text{ms}$.
9. Implement the **Coordinated Wash Trading & Self-Trade Engine (`internal/detectors/washtrading`)**:
   - Intercept pre-match buyer and seller IDs. Query Redis in-memory identity clusters populated from KYC/AML Service (Prompt 202).
   - Reject executions where buyer and seller share identical PAN hashes, beneficial owner IDs, bank account hashes, hardware UUIDs, or primary IP subnets.
   - Detect cyclic wash trading rings ($A \to B \to C \to A$) within rolling 60-second windows where net beneficial ownership transfer is zero.
10. Implement the **Pre-Trade Risk Containment Dispatcher**:
    - When an account's aggregate manipulation risk score exceeds 75/100, emit gRPC containment directives to `services/risk-and-margin-checks-service` (Prompt 206).
    - Automatically enforce restrictive measures: cancel resting quotes, force post-only execution, increase initial margin requirements to $300\%$, or temporarily freeze order submission.
11. Implement the **Hyperledger Besu On-Chain Logging Subsystem (`internal/blockchain`)**:
    - Implement a resilient worker queue using `go-ethereum/ethclient` communicating with the local Besu node over JSON-RPC.
    - Format and sign transactions calling `recordHaltEvent` on `MarketIntegrityAuditRegistry.sol`.
    - Implement automatic nonce management, gas price bumping, and receipt confirmation handling without blocking hot-path order surveillance.
12. Implement the **ClickHouse Telemetry Streamer (`internal/storage`)**:
    - Buffer raw order lifecycle events, basis deviation metrics, and alert triggers in memory, flushing batches to ClickHouse via native TCP client every 100ms or 1,000 records.
    - Persist to tables `surveillance_order_telemetry`, `btc_usdt_vwap_deviation_ticks`, and `surveillance_alerts_archive`.
13. Implement the **Supervisory gRPC & Health Monitoring Interface**:
    - Expose gRPC service `MarketIntegrityService` allowing compliance officers to inspect real-time metrics, review active alerts, and authorize post-halt auction resumption.
    - Export Prometheus metrics (`guard_speedbump_latency_micros`, `guard_basis_deviation_bps`, `guard_alerts_total`, `guard_halt_events_total`).

## Interfaces / Contracts

### Protobuf Definitions (`surveillance_alert.proto`)
```protobuf
syntax = "proto3";

package growww.surveillance.v1;

option go_package = "github.com/growww/services/market-integrity-guard/gen/v1;guardv1";

enum SurveillanceAlertSeverity {
  SEVERITY_UNSPECIFIED = 0;
  SEVERITY_INFO = 1;
  SEVERITY_WARNING = 2;
  SEVERITY_HIGH = 3;
  SEVERITY_CRITICAL = 4;
}

enum ManipulationType {
  MANIPULATION_UNSPECIFIED = 0;
  SPOOFING = 1;
  LAYERING = 2;
  WASH_TRADING_DIRECT = 3;
  WASH_TRADING_COLLUSIVE_RING = 4;
  PREDATORY_LATENCY_ARBITRAGE = 5;
  QUOTE_STUFFING = 6;
  REFERENCE_PRICE_DISTORTION = 7;
}

enum HaltAction {
  HALT_ACTION_UNSPECIFIED = 0;
  HALT_ACTION_SPEEDBUMP_WIDEN = 1;
  HALT_ACTION_VOLATILITY_PAUSE = 2;
  HALT_ACTION_CIRCUIT_BREAKER_HALT = 3;
  HALT_ACTION_CALL_AUCTION_RESUME = 4;
  HALT_ACTION_FULL_RESUME = 5;
}

message SurveillanceAlert {
  string alert_id = 1;
  int64 timestamp_ns = 2;
  string symbol = 3; // "BTC/USDT"
  ManipulationType manipulation_type = 4;
  SurveillanceAlertSeverity severity = 5;
  double confidence_score = 6; // 0.0 to 100.0
  repeated string suspect_account_hashes = 7;
  repeated string related_order_ids = 8;
  repeated string executed_trade_ids = 9;
  double basis_deviation_bps = 10;
  string evidence_snapshot_hash = 11;
  string narrative = 12;
}

message PriceDeviationEvent {
  string event_id = 1;
  int64 timestamp_ns = 2;
  string symbol = 3;
  double internal_midpoint_price = 4;
  double external_consolidated_vwap = 5;
  double deviation_bps = 6;
  int64 external_tick_timestamp_ns = 7;
  repeated string contributing_venues = 8;
}

message MarketControlDirective {
  string directive_id = 1;
  int64 timestamp_ns = 2;
  string symbol = 3;
  HaltAction action = 4;
  int32 duration_seconds = 5;
  double trigger_deviation_bps = 6;
  string reason_code = 7;
  string on_chain_tx_hash = 8;
}

service MarketIntegrityService {
  rpc StreamSurveillanceAlerts (StreamAlertsRequest) returns (stream SurveillanceAlert);
  rpc GetPriceDeviationMetrics (PriceDeviationRequest) returns (PriceDeviationResponse);
  rpc ExecuteMarketHalt (MarketControlDirective) returns (MarketControlResponse);
  rpc AuthorizeResumption (ResumptionRequest) returns (ResumptionResponse);
}

message StreamAlertsRequest {
  string symbol = 1;
  SurveillanceAlertSeverity min_severity = 2;
}

message PriceDeviationRequest {
  string symbol = 1;
  int64 window_ms = 2;
}

message PriceDeviationResponse {
  string symbol = 1;
  double current_deviation_bps = 2;
  double rolling_vwap_100ms = 3;
  double rolling_vwap_1s = 4;
  double rolling_vwap_60s = 5;
  double internal_bbo_midpoint = 6;
}

message MarketControlResponse {
  bool accepted = 1;
  string status_message = 2;
}

message ResumptionRequest {
  string symbol = 1;
  string authorized_officer_id = 2;
  string cryptographic_signature = 3;
  string resumption_mode = 4; // "CALL_AUCTION" or "CONTINUOUS"
}

message ResumptionResponse {
  bool success = 1;
  string message = 2;
}
```

### ClickHouse Data Schemas
```sql
-- Nanosecond Raw Order Telemetry Table
CREATE TABLE default.surveillance_order_telemetry (
    event_time DateTime64(9, 'UTC'),
    event_date Date DEFAULT toDate(event_time),
    symbol LowCardinality(String),
    order_id UUID,
    account_hash FixedString(64),
    side Enum8('BUY' = 1, 'SELL' = 2),
    order_type Enum8('LIMIT' = 1, 'MARKET' = 2, 'IOC' = 3, 'FOK' = 4),
    is_aggressive UInt8,
    speedbump_delay_micros UInt32,
    price Decimal64(4),
    quantity Decimal64(8),
    action_type Enum8('NEW' = 1, 'AMEND' = 2, 'CANCEL' = 3, 'FILL' = 4),
    lifetime_micros UInt64,
    bbo_spread_bps Int32,
    client_ip_hash FixedString(64),
    device_uuid_hash FixedString(64)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (symbol, event_time, order_id)
TTL event_date + INTERVAL 8 YEAR;

-- Nanosecond External VWAP & Price Basis Deviation Table
CREATE TABLE default.btc_usdt_vwap_deviation_ticks (
    tick_time DateTime64(9, 'UTC'),
    tick_date Date DEFAULT toDate(tick_time),
    symbol LowCardinality(String),
    internal_mid_price Decimal64(4),
    external_vwap_100ms Decimal64(4),
    external_vwap_1s Decimal64(4),
    external_vwap_60s Decimal64(4),
    deviation_bps Decimal32(2),
    active_speedbump_micros UInt32,
    market_state Enum8('NORMAL' = 1, 'WARNING' = 2, 'VOLATILITY_PAUSE' = 3, 'HALTED' = 4)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(tick_date)
ORDER BY (symbol, tick_time)
TTL tick_date + INTERVAL 8 YEAR;

-- Surveillance Alerts Archive Table
CREATE TABLE default.surveillance_alerts_archive (
    alert_time DateTime64(6, 'UTC'),
    alert_date Date DEFAULT toDate(alert_time),
    alert_id UUID,
    symbol LowCardinality(String),
    manipulation_type LowCardinality(String),
    severity LowCardinality(String),
    confidence_score Float32,
    basis_deviation_bps Float32,
    suspect_hashes Array(FixedString(64)),
    evidence_merkle_root FixedString(64),
    besu_tx_hash String,
    narrative String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(alert_date)
ORDER BY (symbol, alert_time, alert_id)
TTL alert_date + INTERVAL 8 YEAR;
```

### Hyperledger Besu Audit Registry Interface
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IMarketIntegrityAuditRegistry {
    enum HaltType {
        VOLATILITY_DEVIATION_HALT,
        EXTERNAL_FEED_STALE_HALT,
        COORDINATED_MANIPULATION_FREEZE
    }

    event MarketHaltRecorded(
        bytes32 indexed symbolHash,
        HaltType indexed haltType,
        uint256 internalPriceBps,
        uint256 externalVwapBps,
        int256 deviationBps,
        uint256 triggerTimestampNs,
        bytes32 evidenceMerkleRoot,
        address indexed guardSigner
    );

    event AccountContainmentRecorded(
        bytes32 indexed accountHash,
        bytes32 indexed symbolHash,
        uint8 restrictionLevel,
        uint256 timestampNs,
        bytes32 evidenceHash
    );

    function recordHaltEvent(
        string calldata pairSymbol,
        HaltType haltType,
        uint256 internalPriceBps,
        uint256 externalVwapBps,
        int256 deviationBps,
        uint256 triggerTimestampNs,
        bytes32 evidenceMerkleRoot
    ) external returns (bytes32 eventId);

    function recordAccountContainment(
        bytes32 accountHash,
        string calldata pairSymbol,
        uint8 restrictionLevel,
        uint256 timestampNs,
        bytes32 evidenceHash
    ) external returns (bool success);
}
```

## Security & Compliance Notes
- **SEBI PFUTP Regulations 2003:** Algorithms must strictly enforce detection and prevention mechanisms against fraudulent, manipulative, and unfair trade practices, with comprehensive audit preservation for all non-bona fide orders.
- **SEBI Algorithmic Trading Order-to-Trade Ratio (OTR) Mandate:** Continuous tracking of OTR per user account; users exceeding permitted OTR thresholds ($> 50:1$ within short intervals) are subject to automated rate throttling and risk surcharge multipliers.
- **IFSCA Market Conduct Guidelines (GIFT City):** Continuous monitoring to ensure local tokenized spot order books do not deviate abnormally from primary international reference venues, mitigating cross-border price exploitation.
- **Detection of Coordinated Wash Trading Across Account Rings:** Beneficial ownership resolution must transcend individual account IDs. The guard evaluates composite identity clusters linking PAN hashes, nominee relationships, shared residential IP CIDRs, browser WebGL hashes, and common deposit wallet addresses to prevent volume inflation schemes.
- **Zero PII Protection & DPDP Act 2023:** All investor identifiers processed in surveillance queues, stored in analytical tables, or committed to blockchain ledgers must strictly utilize keyed HMAC-SHA256 hashes, ensuring complete cryptographic anonymity while retaining forensic linkage.
- **Dual-Key Maker-Checker Resumption:** Resuming trading following a Level 3 Circuit-Breaker Halt cannot be performed unilaterally. The guard enforces dual-signature administrative authorization from two independent compliance officers before the Matching Engine uncrosses call auction orders.

## Acceptance Criteria
- [ ] 500-microsecond asymmetric speed bump introduces deterministic latency exclusively to aggressive taker orders ($\pm 25\mu\text{s}$ tolerance) while passive maker orders bypass with zero artificial delay.
- [ ] Real-time external VWAP engine computes sliding 100ms, 1s, 5s, and 60s reference prices with update latency $< 1\text{ms}$ upon receipt of external market feeder ticks.
- [ ] Basis point deviation tracker identifies internal-versus-external price divergence and triggers Level 2 Volatility Pause within $< 5\text{ms}$ of sustained $> 150\text{ bps}$ breach.
- [ ] Instantaneous $> 300\text{ bps}$ price divergence triggers Level 3 Circuit-Breaker Halt to the Matching Engine in $< 2\text{ms}$.
- [ ] Spoofing detection accurately flags quote insertions canceled within $< 200\text{ms}$ without fill intent across synthetic benchmark test cases with $> 99\%$ recall.
- [ ] Layering detection accurately identifies 3-tier depth spoofing followed by opposite execution and mass cancellation with zero false negatives in test suites.
- [ ] Wash trading interceptor blocks 100% of direct self-matches and detects multi-account circular wash trading rings sharing identical PAN, bank account, or device fingerprints.
- [ ] Every circuit-breaker halt event compiles a valid Merkle root and successfully anchors a transaction on the Hyperledger Besu test ledger within 3 seconds.
- [ ] ClickHouse telemetry streaming pipeline sustains $\ge 50,000$ events/sec with zero dropped records.
- [ ] Unit and integration test coverage across all guard components exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 104: Event Schema & Kafka Topic Standards (Standardized order and market event formats).
  - Prompt 204: Order Service (Order submission pipeline and routing).
  - Prompt 205: Order Matching Engine (Matching pipeline, LOB depth snapshots, halt hooks).
  - Prompt 206: Risk & Margin Checks Service (Pre-trade margin controls and account throttle hooks).
  - Prompt 272: Market Feeder Service (External spot ticks from Binance, Coinbase, OKX, Kraken).
- **Subsequent / Dependent Tasks:**
  - Prompt 711: Market Surveillance & Anti-Manipulation Engine (Broader multi-asset surveillance).
  - Prompt 713: Continuous Market Circuit Breaker & Resumption Coordinator (Exchange-wide circuit breakers).
  - Prompt 715: NBSE Market Surveillance & SEBI Reporting Engine (Statutory regulatory dossier generation).
  - Prompt 719: Sybil Ring & Cross-Account Collusion Graph Surveillance Engine (Deep multi-hop network forensics).
