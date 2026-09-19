# 228 - Real-Time Market Surveillance & Dynamic Volatility Engine (Go)

## Purpose
Operating an equity stock exchange continuously 24 hours a day, 7 days a week introduces unique market microstructure challenges: varying liquidity regimes across overnight/weekend windows, potential flash crashes, low-liquidity manipulation, and algorithmic market abuse.

The Real-Time Market Surveillance & Dynamic Volatility Engine protects exchange integrity and satisfies SEBI Integrated Market Surveillance System (IMSS) and IFSCA regulatory mandates.

The engine executes two critical real-time missions:
1. **Dynamic Volatility Circuit Breakers & 24/7 Price Bands:** Enforces static daily price limits (e.g., +/- 10%, +/- 20% from baseline reference prices) combined with rolling, dynamic 5-minute Limit-Up / Limit-Down (LULD) volatility corridors (+/- 3%, +/- 5%). When price breaches a dynamic tier, the engine automatically commands the Matching Engine (Prompt 205) to initiate a 5-minute cooling-off halt or call auction.
2. **Real-Time Algorithmic Market Abuse Detection:** Analyzes high-speed order book events, cancellations, and trade streams using sliding-window statistical detectors to flag manipulative trading patterns:
 - **Spoofing / Layering:** Placing large non-bona-fide orders to manipulate perceived book depth, followed by immediate cancellation upon execution of a small contra-order.
 - **Wash Trading / Circular Trading:** Pre-arranged trading between colluding accounts or self-matching across distinct client IDs with identical beneficial ownership.
 - **Marking the Close / Price Painting:** Aggressive aggressive buying/selling at reference price calculation windows.
 - **Pump-and-Dump & Momentum Ignition:** Rapid aggressive volume bursts across low-float securities.

## What You Are Building
A high-throughput, stream-processing Go microservice (`services/surveillance-engine`). Concrete deliverables include:
- **Dynamic Price Band & LULD Controller:** Real-time sliding window price collar manager updating reference prices, dynamic upper/lower bands, and triggering automatic symbol trading halts.
- **Microstructure Pattern Detectors:** Sliding window analytical algorithms evaluating order-to-trade ratios (OTR), cancellation speeds, depth skew anomalies, and circular trading graphs.
- **Exchange Halt Control Plane:** Direct gRPC control interface to Order Matching Engine (Prompt 205) and Order Service (Prompt 204) to freeze matching, cancel unexecutable orders, and transition symbols into Call Auction or Halting states.
- **Surveillance Alert & Case Management Pipeline:** Real-time alert generator populating database tables and alerting compliance officers via WebSocket dashboards and automated SEBI/IFSCA incident exports.
- **On-Chain Emergency Freeze Coordinator:** High-priority multisig trigger interfacing with `TransferComplianceHooks.sol` and `SettlementDvP.sol` during systemic exchange-wide emergency circuit breakers.

## Scope Boundaries
- **In Scope:**
 - Continuous calculation and enforcement of dynamic price bands and LULD corridors.
 - Automated symbol-level and index-level trading halt state transitions (Trading, Quoting-Only Call Auction, Halt, Resume).
 - High-frequency order stream anomaly detection (Spoofing, Layering, Wash Trading, Momentum Ignition).
 - Calculation of participant Order-to-Trade Ratios (OTR) and automated fine/throttling triggers.
 - Alert logging and regulatory incident export generation for SEBI IMSS.
- **Out of Scope / Handled Elsewhere:**
 - Individual pre-trade user balance and credit checks (Prompt 206).
 - In-memory order book FIFO execution (Prompt 205).
 - Post-trade tax and ledger reporting (Prompt 210, 216).
 - Admin UI presentation layer (Prompt 604, 605).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ utilizing lock-free concurrent ringbuffers and worker pools for real-time stream processing.
- **High-Velocity State Storage:** Redis 7.2 (with Redis Streams and sorted sets) for maintaining rolling 5-minute tick windows, VWAP calculations, and instantaneous OTR counters.
- **Analytical & Time-Series Database:** ClickHouse for petabyte-scale trade and order event analytics; PostgreSQL 16+ for surveillance cases, alert investigations, and regulatory audit records.
- **Messaging & Event Streaming:** `segmentio/kafka-go` consuming all inbound orders, cancellations, and trades (`order.matching.commands.v1`, `engine.matches.v1`, `matching.depth.v1`).
- **Inter-Service Communication:** `google.golang.org/grpc` for real-time control plane commands to matching engine nodes.

## Backend / Infra Touchpoints
- **Order Matching Engine (Prompt 205):** Ingests live match and depth streams; sends `HaltSymbol`, `ResumeSymbol`, `StartCallAuction` gRPC commands.
- **Risk Engine (Prompt 206):** Broadcasts dynamic price band updates every 10 seconds to update pre-trade rejection thresholds.
- **Admin Back Office Service (Prompt 217):** Delivers real-time suspicious activity alerts and emergency override capabilities to compliance officers.
- **Regulatory Reporting Service (Prompt 216):** Feeds verified market abuse incidents for statutory SEBI/IFSCA reporting.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Circuit Breaker Coordination:** If market surveillance detects critical exchange-wide anomalies or catastrophic market drops ($> 20\%$), the surveillance engine triggers an emergency pause hook on `TransferComplianceHooks.sol` (Prompt 305) and `SettlementDvP.sol` (Prompt 306).
- **Audit Hash Anchor:** Daily cryptographic hashes of all surveillance logs, alerts, and circuit breaker trip events are written to the Hyperledger Besu ledger via `ProofOfReserve.sol` / `AuditLogService` (Prompt 218) to guarantee tamper-proof historical record retention.
- **Zero PII:** Surveillance models analyze anonymous `user_id`, `firm_id`, and public wallet addresses (`0x...`), decrypting identity records only during formal regulatory case escalation.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/surveillance-engine` with clean architecture, Makefile, and domain-separated packages.
2. **Define Protobuf Contracts:** Create `proto/growww/surveillance/v1/surveillance_service.proto` for symbol status control, price band queries, and alert notifications.
3. **Design ClickHouse & PostgreSQL Schemas:** Create ClickHouse tables for high-velocity tick logs and PostgreSQL tables for `surveillance_alerts`, `circuit_breaker_events`, and `investigation_cases`.
4. **Implement Dynamic Price Band Engine:** 
 - Compute static daily baseline reference price $P_{\text{ref}}$ (e.g. previous midnight UTC VWAP).
 - Compute rolling 5-minute VWAP $P_{\text{5min\_vwap}}$.
 - Enforce Dynamic Upper Band: $P_{\text{upper}} = \min(P_{\text{ref}} \times 1.20, P_{\text{5min\_vwap}} \times 1.05)$.
 - Enforce Dynamic Lower Band: $P_{\text{lower}} = \max(P_{\text{ref}} \times 0.80, P_{\text{5min\_vwap}} \times 0.95)$.
5. **Implement Circuit Breaker State Machine:** Build deterministic state manager managing transitions: `CONTINUOUS_TRADING` $\rightarrow$ `VOLATILITY_PAUSE` $\rightarrow$ `CALL_AUCTION` $\rightarrow$ `RESUMED`.
6. **Implement Matching Engine Halt Control:** On circuit breaker trigger, send immediate gRPC command to Matching Engine to halt execution for the affected ISIN and broadcast public administrative notice.
7. **Implement Spoofing & Layering Detector:** Build pattern analyzer tracking rapid quote cancellation rates: if a participant places orders $> 500$ lots, cancels within $< 200\text{ms}$, and executes opposite side trade within $< 500\text{ms}$, emit `ALERT_SPOOFING`.
8. **Implement Wash Trading Detector:** Monitor matched buyer/seller IDs in `engine.matches.v1`. Flag transactions where buyer and seller share the same underlying PAN/KYC entity or clustered IP/device fingerprints.
9. **Implement Order-to-Trade Ratio (OTR) Monitor:** Compute rolling hourly OTR per participant:
   $$\text{OTR} = \frac{\text{Orders Submitted} + \text{Orders Modified} + \text{Orders Cancelled}}{\text{Trades Executed} + 1}$$
   If $\text{OTR} > 100$, trigger automated rate throttling and issue fine notifications.
10. **Implement Marking the Reference Detector:** Detect aggressive market orders placed in the final 60 seconds before daily index/NAV calculation windows that alter reference prices by $> 1.5\%$.
11. **Build Alert Dispatcher & Case Management:** Aggregate raw detection triggers, compute threat confidence score ($0.0 - 1.0$), and dispatch high-severity alerts to compliance WebSocket feeds.
12. **Implement SEBI IMSS Export Formatter:** Build scheduled exporter formatting daily alerts into standard SEBI IMSS XML/CSV reporting schemas.
13. **Integrate On-Chain Emergency Hook:** Implement automated webhook trigger that invokes multisig emergency pause on `TransferComplianceHooks.sol` upon systemic index circuit breaker trip.
14. **Instrument Prometheus Telemetry:** Export real-time gauges for active halted symbols, detected alerts by category per hour, and stream ingestion lag.
15. **Execute Microstructure Simulation Tests:** Run stress test simulating flash crash price drops and coordinated spoofing order flows, verifying detection accuracy and circuit breaker tripping within $< 10\text{ms}$.

## Interfaces / Contracts

### Protobuf Definition (`surveillance_service.proto`)
```protobuf
syntax = "proto3";

package growww.surveillance.v1;

option go_package = "growww/surveillance/v1;surveillancev1";

service SurveillanceService {
  rpc GetSymbolMarketState (GetSymbolMarketStateRequest) returns (GetSymbolMarketStateResponse);
  rpc TriggerManualHalt (TriggerManualHaltRequest) returns (TriggerManualHaltResponse);
  rpc ResumeSymbolTrading (ResumeSymbolTradingRequest) returns (ResumeSymbolTradingResponse);
  rpc ListSurveillanceAlerts (ListSurveillanceAlertsRequest) returns (ListSurveillanceAlertsResponse);
}

enum MarketTradingStatus {
  STATUS_CONTINUOUS_TRADING = 0;
  STATUS_VOLATILITY_PAUSE = 1;
  STATUS_CALL_AUCTION = 2;
  STATUS_HALTED = 3;
  STATUS_EMERGENCY_STOP = 4;
}

enum AlertCategory {
  ALERT_CATEGORY_UNSPECIFIED = 0;
  ALERT_CATEGORY_SPOOFING = 1;
  ALERT_CATEGORY_LAYERING = 2;
  ALERT_CATEGORY_WASH_TRADING = 3;
  ALERT_CATEGORY_EXCESSIVE_OTR = 4;
  ALERT_CATEGORY_CIRCUIT_BREAKER_TRIP = 5;
  ALERT_CATEGORY_PRICE_BAND_VIOLATION = 6;
  ALERT_CATEGORY_MARKING_THE_CLOSE = 7;
}

message GetSymbolMarketStateRequest {
  string isin = 1;
}

message GetSymbolMarketStateResponse {
  string isin = 1;
  MarketTradingStatus status = 2;
  string static_upper_band = 3;
  string static_lower_band = 4;
  string dynamic_upper_band = 5;
  string dynamic_lower_band = 6;
  string reference_price = 7;
  string current_5min_vwap = 8;
  int64 halt_start_time_unix = 9;
  int64 halt_duration_seconds = 10;
}

message TriggerManualHaltRequest {
  string isin = 1;
  string reason = 2;
  string admin_user_id = 3;
  int64 duration_seconds = 4;
}

message TriggerManualHaltResponse {
  string isin = 1;
  MarketTradingStatus new_status = 2;
  bool halt_enacted = 3;
}

message ResumeSymbolTradingRequest {
  string isin = 1;
  string admin_user_id = 2;
  bool start_with_call_auction = 3;
}

message ResumeSymbolTradingResponse {
  string isin = 1;
  MarketTradingStatus new_status = 2;
  bool resumed = 3;
}

message ListSurveillanceAlertsRequest {
  string isin = 1;
  AlertCategory category = 2;
  int64 start_time_unix = 3;
  int64 end_time_unix = 4;
  int32 page_size = 5;
}

message SurveillanceAlert {
  string alert_id = 1;
  string isin = 2;
  AlertCategory category = 3;
  string severity = 4; // "LOW", "MEDIUM", "HIGH", "CRITICAL"
  string participant_id = 5;
  string description = 6;
  double confidence_score = 7;
  int64 triggered_at_unix = 8;
}

message ListSurveillanceAlertsResponse {
  repeated SurveillanceAlert alerts = 1;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE alert_severity_enum AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE alert_status_enum AS ENUM ('NEW', 'UNDER_INVESTIGATION', 'DISMISSED', 'ESCALATED_SEBI', 'SANCTIONED');

CREATE TABLE circuit_breaker_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    trigger_price NUMERIC(18, 4) NOT NULL,
    reference_price NUMERIC(18, 4) NOT NULL,
    band_type VARCHAR(32) NOT NULL, -- 'DYNAMIC_5MIN_UPPER', 'DYNAMIC_5MIN_LOWER', 'STATIC_DAILY_LIMIT'
    halt_type VARCHAR(32) NOT NULL, -- 'LULD_PAUSE', 'CALL_AUCTION', 'MARKET_WIDE_HALT'
    duration_seconds INT NOT NULL DEFAULT 300,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resumed_at TIMESTAMPTZ
);

CREATE TABLE surveillance_alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    participant_id VARCHAR(64) NOT NULL,
    category VARCHAR(64) NOT NULL,
    severity alert_severity_enum NOT NULL DEFAULT 'MEDIUM',
    status alert_status_enum NOT NULL DEFAULT 'NEW',
    
    confidence_score NUMERIC(5, 4) NOT NULL CHECK (confidence_score BETWEEN 0 AND 1),
    evidence_payload JSONB NOT NULL,
    
    assigned_compliance_officer UUID,
    resolution_notes TEXT,
    
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE participant_otr_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id VARCHAR(64) NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    order_count INT NOT NULL DEFAULT 0,
    cancel_count INT NOT NULL DEFAULT 0,
    trade_count INT NOT NULL DEFAULT 0,
    otr_ratio NUMERIC(10, 2) NOT NULL,
    throttled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_circuit_breakers_isin ON circuit_breaker_events(isin, triggered_at);
CREATE INDEX idx_surveillance_alerts_status ON surveillance_alerts(status, severity, triggered_at);
CREATE INDEX idx_participant_otr ON participant_otr_metrics(participant_id, window_start);
```

## Security & Compliance Notes
- **SEBI IMSS Compliance:** Surveillance rules and alert parameters strictly adhere to SEBI Master Circular on Market Surveillance and IFSCA Trading Rules.
- **Tamper-Evident Incident Records:** Every triggered alert, investigation note, and circuit breaker trip is immutably timestamped and archived into cold storage with cryptographic hashes anchored to Hyperledger Besu.
- **Automated Fail-Safe Protection:** The engine operates with automated fail-safe halts; if market data feeds experience anomalous sequence loss ($> 100$ dropped packets), the affected symbol is automatically placed into quoting-only state until sequence reconciliation completes.

## Acceptance Criteria
- [ ] Real-time dynamic volatility price bands (+/- 3% 5-minute rolling LULD, +/- 20% daily static) are recalculated and updated continuously.
- [ ] Breaching a price band automatically commands the Matching Engine to halt matching and initiate a 5-minute call auction within $< 10\text{ms}$.
- [ ] Spoofing and layering detector correctly flags order placement/cancellation patterns matching manipulation criteria with zero false positives on normal market maker quotes.
- [ ] Wash trade detector catches 100% of trades executing between accounts with matching PAN/KYC identifiers.
- [ ] Order-to-Trade Ratio (OTR) calculations correctly throttle non-compliant algorithmic trading sessions exceeding threshold limits.
- [ ] Daily alert archives and circuit breaker logs format correctly for automated SEBI IMSS compliance export.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (Architecture Overview), 104 (Kafka Standards), 205 (Order Matching Engine), 206 (Risk Engine), 207 (Market Data Service).
- **Parallel Tasks:** 225 (FIX Gateway), 226 (Advanced Order Types), 227 (Institutional Dark Pool).
- **Downstream Blockers:** 216 (Regulatory Reporting Service), 217 (Admin Back Office Service).
