# 715 - NBSE Market Surveillance & SEBI Reporting Engine (Cross-Market Manipulation, Circular Trading, Front-Running)

## Purpose
Establishes a high-throughput, institutional-grade Cross-Market Surveillance and Statutory Regulatory Reporting Engine operating across domestic and offshore trading venues. As the platform bridges the National Blockchain Stock Exchange (NBSE) 24/7 tokenized fractional equity market in GIFT City with traditional domestic execution venues including the National Stock Exchange of India (NSE), Bombay Stock Exchange (BSE), and Multi Commodity Exchange (MCX), it creates unique market microstructure surfaces susceptible to sophisticated multi-venue manipulation.

In compliance with the **SEBI (Prohibition of Fraudulent and Unfair Trade Practices relating to Securities Market) Regulations, 2003 (PFUTP)**, **SEBI (Prohibition of Insider Trading) Regulations, 2015 (PIT)**, **SEBI Master Circulars on Market Surveillance**, and **IFSCA (Market Conduct) Regulations, 2021**, this service provides continuous, real-time surveillance across synchronized multi-exchange order and trade streams. It detects and neutralizes cross-market market abuse typologies, including:
1. **Cross-Venue Spoofing and Layering:** Injecting large non-bona fide liquidity or phantom quotes on NBSE to create artificial price pressure while executing aggressive opposite-side fills on NSE/BSE/MCX, or vice versa, followed by immediate quote cancellation.
2. **Synchronized Cross-Market Wash Trading:** Executing offsetting trades across NBSE and domestic exchanges between collusive accounts or beneficial ownership clusters to fabricate synthetic trading volumes and distort benchmark reference pricing.
3. **Multi-Entity Circular Trading Loops:** Routing buy/sell sequences through closed loops of interconnected client entities across NBSE, NSE, and BSE to artificially inflate prices, bypass position limits, or manufacture fraudulent tax credits.
4. **Cross-Market Front-Running and Tailgating:** Exploiting latency differentials between NBSE order books and domestic exchange feeds to trade ahead of known institutional parent orders or anticipated cross-border arbitrage flows.
5. **Marking the Open / Marking the Close Arbitrage:** Manipulating closing settlement prices on domestic exchanges by pushing aggressive off-market quotes on NBSE during standard market close discovery windows.

When abusive patterns are identified with high statistical confidence, this service automatically enforces protective trading throttles, triggers cryptographic trading freezes on the permissioned settlement ledger, and compiles tamper-evident statutory reporting dossiers formatted for direct electronic filing with SEBI and IFSCA.

## What You Are Building
A distributed, low-latency stream processing and compliance microservice (`services/cross-market-surveillance`):
- **Multi-Venue Market Feed Synchronizer & Clock Normalizer:** Ingests high-frequency tick-by-tick order lifecycle events and L2/L3 order book updates across NBSE, NSE, BSE, and MCX, normalizing timestamps to nanosecond precision with automated IEEE 1588 Precision Time Protocol (PTP) clock-drift compensation.
- **Cross-Market Complex Event Processing (CEP) Correlator:** Stateful sliding-window correlation pipeline evaluating multi-exchange order and trade sequences across 50ms, 250ms, 1s, 5s, 1m, 15m, and end-of-day windows.
- **Beneficial Ownership & Entity Graph Resolver:** In-memory graph analytics engine mapping domestic PAN, GIFT City Legal Entity Identifiers (LEI), passport IDs, bank account hashes, hardware fingerprints, and IP/MAC subnets into unified trading clusters.
- **Multi-Typology Abuse Detection Engines:** Modular, deterministic algorithmic detectors identifying cross-market spoofing, circular trading loops, wash sales, front-running, and cross-border price divergence manipulation.
- **Composite Risk Scorer & Automated Defense Orchestrator:** Computes normalized threat scores ($0.0 - 100.0$) and issues dynamic risk multipliers, quote purge directives, or account freeze signals to the Pre-Trade Risk Engine (Prompt 206) and FIX Gateway (Prompt 225).
- **SEBI & IFSCA Statutory Dossier Generator:** Compiles standardized regulatory alert dossiers (including visual order-book depth chronologies, tick heatmaps, entity relationship subgraphs, and cryptographic evidence hashes) ready for statutory submission to SEBI IMSS and IFSCA DART systems.

## Scope Boundaries
- **In Scope:**
  - Ingestion and normalization of order lifecycle events (`OrderPlaced`, `OrderModified`, `OrderCancelled`, `TradeExecuted`) across NBSE, NSE, BSE, and MCX.
  - Nanosecond timestamp alignment and cross-venue clock drift logging.
  - Algorithmic detection of Cross-Venue Spoofing, Cross-Venue Layering, Synchronized Cross-Market Wash Trading, Multi-Entity Circular Trading, and Cross-Market Front-Running.
  - Continuous graph cycle detection across order routing networks to identify multi-hop circular trading rings ($k \ge 3$ counterparties).
  - Real-time calculation of cross-market metrics: Venue Price Divergence (VPD), Cross-Market Cancel-to-Fill Ratio (CCFR), and Cross-Venue Volume Correlation (CVVC).
  - Automated dispatch of protective throttle and trading suspension directives.
  - Generation and archival of SEBI PFUTP and IFSCA statutory compliance alert filings.
  - Production of cryptographic evidence receipts anchored to Hyperledger Besu.
- **Out of Scope / Handled Elsewhere:**
  - Single-venue intraday order matching and local book execution (handled in Prompt 205).
  - Volatility circuit breakers and market-wide continuous trading halts (handled in Prompt 713).
  - Off-chain corporate insider trading and Unpublished Price Sensitive Information (UPSI) social graph analysis (handled in Prompt 712).
  - General fiat AML monitoring and smurfing detection (handled in Prompts 202 and 704).
  - Physical custodian clearing and depository DvP settlements (handled in Prompts 208 and 213).

## Technology to Use
- **Core Streaming & Ingestion Layer:** Go 1.22+ for high-throughput multi-exchange feed ingestion, binary protocol parsing (FIX 5.0 SP2, ITCH/OUCH, and FAST protocol), and sub-millisecond gRPC routing.
- **Analytics, Graph & Regulatory Engine:** Python 3.12+ with FastAPI, Polars (for high-speed vectorized order book analytics), NetworkX / `rustworkx` (for real-time circular trading graph cycle detection), and Jinja2 / `weasyprint` (for statutory dossier generation).
- **Stream Processing Backbone:** Apache Flink 1.19+ with RocksDB state backend for stateful sliding-window CEP pattern matching and event-time watermark processing.
- **Graph Database & Entity Store:** Neo4j Enterprise 5+ or Memgraph Enterprise with Cypher query support for real-time traversal of beneficial ownership and trading ring networks.
- **In-Memory Microstructure Cache:** Redis 7 Cluster (or Dragonfly) for sliding-window depth state, rate-limit buckets, and sub-millisecond alert deduplication.
- **Columnar Time-Series Analytics Store:** ClickHouse 24+ for raw tick and order event archival, historical cross-market replay, and backtesting.
- **Relational Case & Audit Database:** PostgreSQL 16 with TimescaleDB extension and Row-Level Security (RLS) for alert lifecycles, investigation workflows, and regulatory audit trails.
- **Justification:** Evaluating cross-market manipulation across multiple exchanges requires processing hundreds of thousands of order events per second with sub-millisecond latency. Go provides lightweight concurrency and low memory overhead for feed ingestion, Flink ensures fault-tolerant stateful CEP streaming with out-of-order handling, and Python provides the data science and graph toolchains required for complex entity clustering and SEBI regulatory reporting.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
  - `services/order-matching-engine` (Prompt 205): Real-time NBSE order lifecycle feeds, trade execution logs, and L2/L3 order book snapshots.
  - `services/market-data-service` (Prompt 207): Real-time normalized external market feeds from NSE, BSE, and MCX.
  - `services/fix-gateway` (Prompt 225): FIX 5.0 SP2 message logs, client routing identifiers, and institutional cancel-replace streams.
  - `services/user-service` & `services/kyc-service` (Prompts 201 & 202): Domestic PAN, GIFT City LEI/Passport, hardware UUIDs, IP subnets, and KYC entity linkages.
- **Downstream Microservices:**
  - `services/risk-engine` (Prompt 206): Ingests dynamic margin multipliers and participant throttling commands.
  - `services/regulatory-reporting` (Prompts 216 & 233): Consumes formatted statutory dossiers for automated SEBI / IFSCA portal dispatch.
  - `services/admin-back-office` (Prompts 217, 604 & 605): Populates the compliance officer surveillance workbench with interactive order-book replay timelines.
  - `services/notification-service` (Prompt 211): Delivers high-priority compliance alert notifications to compliance officers.
- **Messaging Topics (Apache Kafka):**
  - Consumes: `market.nbse.orders.lifecycle.v1`, `market.nbse.trades.v1`, `market.external.nse.ticks.v1`, `market.external.bse.ticks.v1`, `market.external.mcx.ticks.v1`, `user.auth.session_fingerprint.v1`.
  - Publishes: `surveillance.cross_market.alert.v1`, `surveillance.cross_market.throttle.v1`, `surveillance.cross_market.freeze_command.v1`, `regulatory.filing.sebi_alert.v1`, `regulatory.filing.ifsca_alert.v1`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **On-Chain Evidence Anchoring:** For every cross-market alert with confidence $\ge 90\%$, the service computes a deterministic SHA-256 Merkle root of the multi-exchange order logs, execution ticks, entity graph topology, and detector metrics. The hash is anchored to `MarketSurveillanceRegistry.sol` on Hyperledger Besu, creating an immutable, legally defensible audit trail.
- **Automated Trading Account Freezing:** When high-conviction cross-market manipulation is confirmed ($\text{Risk Score} \ge 95$), the service dispatches a gRPC command to the smart contract relayer to execute `ComplianceRegistry.freezeTradingAccount(address account, bytes32 reasonCode, bytes32 proofHash)` on-chain. This immediately halts all tokenized equity transfers and matching engine settlements for the offending entity.
- **Zero PII on Ledger:** All blockchain records strictly reference pseudonymized investor account hashes (`account_hash = HMAC-SHA256(pan_or_lei, platform_salt)`), external exchange order reference hashes, and evidence Merkle roots. Zero investor names, tax IDs, or personal data are ever committed to the ledger.

## Step-by-Step Build Instructions
1. Scaffold the `services/cross-market-surveillance` monorepo directory structure:
   - `cmd/ingestor/` (Go 1.22+ feed ingesters and protocol decoders).
   - `cmd/correlator/` (Go/Flink CEP engine).
   - `pkg/detectors/` (Modular manipulation detection algorithms).
   - `pkg/graph/` (Entity link analysis and graph cycle detection).
   - `pkg/dossier/` (Python 3.12+ SEBI / IFSCA statutory report generator).
2. Generate Go and Python gRPC / Protobuf bindings from `proto/growww/cross_market_surveillance/v1/cross_market_surveillance.proto`.
3. Execute PostgreSQL 16 database migrations for tables: `cross_market_surveillance_alerts`, `cross_market_entities`, `circular_trading_rings`, `regulatory_filing_dossiers`, and `venue_clock_drift_logs`.
4. Configure ClickHouse schema and table partitions for high-throughput multi-exchange tick ingestion and 8-year WORM compliance retention.
5. Implement the **Multi-Venue Feed Normalizer**:
   - Ingest NBSE internal order lifecycle stream (`market.nbse.orders.lifecycle.v1`).
   - Ingest external exchange tick feeds from NSE, BSE, and MCX via FIX/FAST feed adapters.
   - Normalize order fields (ISIN, side, price, volume, venue, timestamp, order ID, participant ID).
   - Compute running clock drift metrics against PTP reference clocks, recording clock drift logs when offset exceeds $\pm 1\text{ms}$.
6. Implement the **Cross-Venue Spoofing & Layering Detector**:
   - Track large quote injections on NBSE ($\ge 4\times$ average depth size at top-3 price levels) with short lifetime ($< 500\text{ms}$) combined with opposite-side market order fills on NSE/BSE for the same underlying ISIN.
   - Track opposite scenario: large quote injections on NSE/BSE followed by aggressive fills on NBSE tokenized fractional shares.
   - Calculate Cross-Market Cancel-to-Fill Ratio (CCFR):
     $$\text{CCFR}_{\text{venue}_A} = \frac{\text{Cancelled Volume}_{\text{venue}_A}}{\text{Filled Volume}_{\text{venue}_A} + 1}$$
   - Flag events where $\text{CCFR} > 20.0$ on one venue while fill volume on the paired venue exceeds $70\%$ within a 1,000ms window.
7. Implement the **Synchronized Cross-Market Wash Trading Detector**:
   - Ingest execution streams across all venues.
   - Query the in-memory entity graph to identify trades where buyer and seller share identical beneficial ownership (same PAN, LEI, primary bank account, or device UUID) across different venues within $\Delta t \le 200\text{ms}$.
   - Flag cross-venue self-matching and zero-economic-risk trades executed at off-market prices ($> 0.5\%$ divergence from prevailing NBSE/NSE BBO).
8. Implement the **Multi-Entity Circular Trading Ring Detector**:
   - Build a directed dynamic transaction graph $G = (V, E)$ where vertices $V$ represent trading participant accounts and directed edges $E = (u, v, w, t)$ represent share transfers/sales from account $u$ to account $v$ with volume $w$ at timestamp $t$.
   - Run Johnson's cycle finding algorithm or DFS cycle detection with depth $3 \le k \le 8$ over sliding 15-minute and 1-day windows.
   - Detect closed volume loops where:
     $$\sum_{e \in \text{cycle}} \text{Volume}(e) \approx \text{Constant}, \quad \text{and} \quad \text{Price}(e_{\text{end}}) > \text{Price}(e_{\text{start}})$$
   - Calculate volume round-trip percentage ($> 85\%$ of original volume returning to origin entity cluster).
9. Implement the **Cross-Market Front-Running & Tailgating Detector**:
   - Ingest institutional large-inbound-order queue telemetry from `services/fix-gateway`.
   - Detect pattern where retail or proprietary accounts submit aggressive orders on NBSE within $10\text{ms} - 500\text{ms}$ prior to the arrival or execution of large institutional parent orders on domestic exchanges (NSE/BSE), followed by rapid unwinding within $< 60\text{s}$.
10. Implement the **Composite Manipulation Risk Scorer**:
    - Aggregate anomaly signals from all active detectors into a unified composite score:
      $$\text{Score} = \min\left(100.0, \; \sum_{i=1}^{N} w_i \cdot S_i \cdot \mathbb{I}(\text{Confidence}_i \ge \theta_i)\right)$$
    - Categorize severity levels: `LOW` ($0-49$), `MEDIUM` ($50-74$), `HIGH` ($75-89$), `CRITICAL` ($90-100$).
11. Implement the **Automated Action & On-Chain Freeze Dispatcher**:
    - If `HIGH`: Emit throttle event to `services/risk-engine` (impose $300\%$ margin multiplier and disable market orders).
    - If `CRITICAL`: Cancel all resting quotes across NBSE, emit `surveillance.cross_market.freeze_command.v1` to invoke `ComplianceRegistry.sol` on Hyperledger Besu, and open a P1 regulatory compliance case.
12. Build the **SEBI & IFSCA Statutory Dossier Generator**:
    - Format alerts into machine-readable JSON / XBRL and tamper-evident PDF dossiers.
    - Include full multi-exchange LOB depth snapshots, trade execution timelines, counterparty attribution graphs, and cryptographic Merkle evidence hashes.
    - Configure automated dispatch to `services/regulatory-reporting` for filing with SEBI and IFSCA.

## Interfaces / Contracts

### Protobuf Definition (`cross_market_surveillance.proto`)
```protobuf
syntax = "proto3";

package growww.cross_market_surveillance.v1;

option go_package = "github.com/growww/services/cross-market-surveillance/gen/v1;crossmarketsurveillancev1";

service CrossMarketSurveillanceService {
  rpc EvaluateCrossMarketEvent (EvaluateCrossMarketEventRequest) returns (EvaluateCrossMarketEventResponse);
  rpc GetSurveillanceAlert (GetSurveillanceAlertRequest) returns (GetSurveillanceAlertResponse);
  rpc ListActiveAlerts (ListActiveAlertsRequest) returns (ListActiveAlertsResponse);
  rpc DetectCircularTradingRings (DetectCircularTradingRingsRequest) returns (DetectCircularTradingRingsResponse);
  rpc GenerateRegulatoryDossier (GenerateRegulatoryDossierRequest) returns (GenerateRegulatoryDossierResponse);
  rpc AcknowledgeAlert (AcknowledgeAlertRequest) returns (AcknowledgeAlertResponse);
}

enum ExchangeVenue {
  VENUE_UNSPECIFIED = 0;
  VENUE_NBSE = 1;
  VENUE_NSE = 2;
  VENUE_BSE = 3;
  VENUE_MCX = 4;
}

enum ManipulationTypology {
  TYPOLOGY_UNSPECIFIED = 0;
  TYPOLOGY_CROSS_VENUE_SPOOFING = 1;
  TYPOLOGY_CROSS_VENUE_LAYERING = 2;
  TYPOLOGY_CROSS_MARKET_WASH_TRADING = 3;
  TYPOLOGY_CIRCULAR_TRADING_RING = 4;
  TYPOLOGY_CROSS_MARKET_FRONT_RUNNING = 5;
  TYPOLOGY_MARKING_THE_CLOSE = 6;
  TYPOLOGY_PRICE_DIVERGENCE_ARBITRAGE = 7;
}

enum AlertSeverity {
  SEVERITY_UNSPECIFIED = 0;
  SEVERITY_LOW = 1;
  SEVERITY_MEDIUM = 2;
  SEVERITY_HIGH = 3;
  SEVERITY_CRITICAL = 4;
}

enum AlertStatus {
  STATUS_UNSPECIFIED = 0;
  STATUS_OPEN = 1;
  STATUS_UNDER_INVESTIGATION = 2;
  STATUS_ESCALATED_SEBI = 3;
  STATUS_ESCALATED_IFSCA = 4;
  STATUS_RESOLVED_FALSE_POSITIVE = 5;
  STATUS_CLOSED_CONFIRMED = 6;
}

message NormalizedOrderEvent {
  string event_id = 1;
  ExchangeVenue venue = 2;
  string isin = 3;
  string symbol = 4;
  string order_id = 5;
  string client_id = 6;
  string entity_cluster_id = 7;
  string side = 8; // "BUY" or "SELL"
  string order_type = 9; // "LIMIT", "MARKET", "STOP"
  double price = 10;
  int64 quantity = 11;
  int64 filled_quantity = 12;
  int64 cancelled_quantity = 13;
  string lifecycle_state = 14; // "NEW", "MODIFIED", "CANCELLED", "FILLED"
  int64 event_timestamp_ns = 15;
  int64 ingested_timestamp_ns = 16;
}

message EvaluateCrossMarketEventRequest {
  NormalizedOrderEvent order_event = 1;
}

message EvaluateCrossMarketEventResponse {
  bool alert_triggered = 1;
  string alert_id = 2;
  AlertSeverity severity = 3;
  double composite_risk_score = 4;
  repeated ManipulationTypology detected_typologies = 5;
  bool throttle_dispatched = 6;
  bool on_chain_freeze_dispatched = 7;
}

message GetSurveillanceAlertRequest {
  string alert_id = 1;
}

message GetSurveillanceAlertResponse {
  string alert_id = 1;
  string isin = 2;
  string symbol = 3;
  AlertSeverity severity = 4;
  AlertStatus status = 5;
  double composite_risk_score = 6;
  repeated ManipulationTypology typologies = 7;
  repeated ExchangeVenue involved_venues = 8;
  repeated string involved_account_hashes = 9;
  string primary_entity_cluster_id = 10;
  string evidence_merkle_root = 11;
  string on_chain_tx_hash = 12;
  string regulatory_dossier_url = 13;
  int64 triggered_at_utc = 14;
  int64 updated_at_utc = 15;
}

message ListActiveAlertsRequest {
  AlertSeverity min_severity = 1;
  AlertStatus status = 2;
  string isin = 3;
  int32 limit = 4;
  int32 offset = 5;
}

message ListActiveAlertsResponse {
  repeated GetSurveillanceAlertResponse alerts = 1;
  int64 total_count = 2;
}

message DetectCircularTradingRingsRequest {
  string isin = 1;
  int64 start_time_utc = 2;
  int64 end_time_utc = 3;
  int32 min_ring_size = 4;
  int32 max_ring_size = 5;
  double min_volume_threshold = 6;
}

message TradingRingHop {
  string from_account_hash = 1;
  string to_account_hash = 2;
  ExchangeVenue venue = 3;
  int64 trade_volume = 4;
  double trade_price = 5;
  int64 execution_timestamp_ns = 6;
}

message CircularTradingRing {
  string ring_id = 1;
  string isin = 2;
  int32 hop_count = 3;
  repeated TradingRingHop hops = 4;
  int64 total_recycled_volume = 5;
  double price_inflation_pct = 6;
  double confidence_score = 7;
}

message DetectCircularTradingRingsResponse {
  repeated CircularTradingRing detected_rings = 1;
  int64 evaluated_trades_count = 2;
  int64 execution_duration_ms = 3;
}

message GenerateRegulatoryDossierRequest {
  string alert_id = 1;
  string target_regulator = 2; // "SEBI", "IFSCA", "BOTH"
  string compliance_officer_id = 3;
  string officer_notes = 4;
}

message GenerateRegulatoryDossierResponse {
  string dossier_id = 1;
  string alert_id = 2;
  string target_regulator = 3;
  string statutory_reference_number = 4;
  string document_download_url = 5;
  string evidence_hash = 6;
  int64 generated_at_utc = 7;
}

message AcknowledgeAlertRequest {
  string alert_id = 1;
  string compliance_officer_id = 2;
  AlertStatus new_status = 3;
  string notes = 4;
  string mfa_token = 5;
}

message AcknowledgeAlertResponse {
  bool success = 1;
  string alert_id = 2;
  AlertStatus current_status = 3;
  int64 updated_at_utc = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TYPE exchange_venue_enum AS ENUM (
    'VENUE_NBSE',
    'VENUE_NSE',
    'VENUE_BSE',
    'VENUE_MCX'
);

CREATE TYPE manipulation_typology_enum AS ENUM (
    'TYPOLOGY_CROSS_VENUE_SPOOFING',
    'TYPOLOGY_CROSS_VENUE_LAYERING',
    'TYPOLOGY_CROSS_MARKET_WASH_TRADING',
    'TYPOLOGY_CIRCULAR_TRADING_RING',
    'TYPOLOGY_CROSS_MARKET_FRONT_RUNNING',
    'TYPOLOGY_MARKING_THE_CLOSE',
    'TYPOLOGY_PRICE_DIVERGENCE_ARBITRAGE'
);

CREATE TYPE alert_severity_enum AS ENUM (
    'SEVERITY_LOW',
    'SEVERITY_MEDIUM',
    'SEVERITY_HIGH',
    'SEVERITY_CRITICAL'
);

CREATE TYPE alert_status_enum AS ENUM (
    'STATUS_OPEN',
    'STATUS_UNDER_INVESTIGATION',
    'STATUS_ESCALATED_SEBI',
    'STATUS_ESCALATED_IFSCA',
    'STATUS_RESOLVED_FALSE_POSITIVE',
    'STATUS_CLOSED_CONFIRMED'
);

CREATE TABLE cross_market_surveillance_alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    severity alert_severity_enum NOT NULL,
    status alert_status_enum NOT NULL DEFAULT 'STATUS_OPEN',
    composite_risk_score NUMERIC(5, 2) NOT NULL,
    primary_typology manipulation_typology_enum NOT NULL,
    secondary_typologies manipulation_typology_enum[] NOT NULL DEFAULT '{}',
    involved_venues exchange_venue_enum[] NOT NULL,
    involved_account_hashes TEXT[] NOT NULL DEFAULT '{}',
    entity_cluster_id VARCHAR(64),
    evidence_merkle_root CHAR(64) NOT NULL,
    on_chain_tx_hash VARCHAR(66),
    on_chain_freeze_tx_hash VARCHAR(66),
    detection_metadata JSONB NOT NULL,
    compliance_officer_id UUID,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cross_market_entities (
    entity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_cluster_id VARCHAR(64) NOT NULL UNIQUE,
    pan_hash CHAR(64) NOT NULL,
    lei_code VARCHAR(20),
    passport_hash CHAR(64),
    primary_bank_account_hash CHAR(64),
    known_account_hashes TEXT[] NOT NULL DEFAULT '{}',
    hardware_fingerprints TEXT[] NOT NULL DEFAULT '{}',
    ip_subnets TEXT[] NOT NULL DEFAULT '{}',
    risk_score NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    is_frozen BOOLEAN NOT NULL DEFAULT FALSE,
    freeze_reason VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE circular_trading_rings (
    ring_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES cross_market_surveillance_alerts(alert_id),
    isin VARCHAR(12) NOT NULL,
    hop_count INT NOT NULL,
    total_recycled_volume NUMERIC(28, 6) NOT NULL,
    price_inflation_pct NUMERIC(6, 2) NOT NULL,
    involved_accounts TEXT[] NOT NULL,
    involved_venues exchange_venue_enum[] NOT NULL,
    ring_structure JSONB NOT NULL,
    confidence_score NUMERIC(5, 2) NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE regulatory_filing_dossiers (
    dossier_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES cross_market_surveillance_alerts(alert_id),
    target_regulator VARCHAR(16) NOT NULL, -- 'SEBI', 'IFSCA', 'BOTH'
    statutory_reference_number VARCHAR(64) NOT NULL UNIQUE,
    statutory_act_reference VARCHAR(128) NOT NULL, -- e.g. 'SEBI PFUTP Regulation 4(2)'
    dossier_payload JSONB NOT NULL,
    dossier_pdf_s3_url VARCHAR(512),
    evidence_merkle_root CHAR(64) NOT NULL,
    submitted_by UUID NOT NULL,
    submission_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    acknowledgement_reference VARCHAR(128),
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE venue_clock_drift_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venue exchange_venue_enum NOT NULL,
    reference_clock_source VARCHAR(64) NOT NULL, -- 'PTP_IEEE_1588_MASTER'
    measured_drift_nanoseconds BIGINT NOT NULL,
    drift_status VARCHAR(16) NOT NULL, -- 'NORMAL', 'WARNING', 'CRITICAL'
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cmsa_isin ON cross_market_surveillance_alerts (isin, created_at DESC);
CREATE INDEX idx_cmsa_status ON cross_market_surveillance_alerts (status, severity);
CREATE INDEX idx_cmsa_cluster ON cross_market_surveillance_alerts (entity_cluster_id);
CREATE INDEX idx_cme_pan ON cross_market_entities (pan_hash);
CREATE INDEX idx_ctr_alert ON circular_trading_rings (alert_id);
CREATE INDEX idx_rfd_alert ON regulatory_filing_dossiers (alert_id);
CREATE INDEX idx_vcdl_venue ON venue_clock_drift_logs (venue, recorded_at DESC);
```

### Kafka Event Schemas

#### Topic: `surveillance.cross_market.alert.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CrossMarketSurveillanceAlertEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "alert_id": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "symbol": { "type": "string" },
    "severity": { "type": "string", "enum": ["SEVERITY_LOW", "SEVERITY_MEDIUM", "SEVERITY_HIGH", "SEVERITY_CRITICAL"] },
    "composite_risk_score": { "type": "number", "minimum": 0.0, "maximum": 100.0 },
    "primary_typology": {
      "type": "string",
      "enum": [
        "TYPOLOGY_CROSS_VENUE_SPOOFING",
        "TYPOLOGY_CROSS_VENUE_LAYERING",
        "TYPOLOGY_CROSS_MARKET_WASH_TRADING",
        "TYPOLOGY_CIRCULAR_TRADING_RING",
        "TYPOLOGY_CROSS_MARKET_FRONT_RUNNING",
        "TYPOLOGY_MARKING_THE_CLOSE",
        "TYPOLOGY_PRICE_DIVERGENCE_ARBITRAGE"
      ]
    },
    "involved_venues": {
      "type": "array",
      "items": { "type": "string", "enum": ["VENUE_NBSE", "VENUE_NSE", "VENUE_BSE", "VENUE_MCX"] }
    },
    "entity_cluster_id": { "type": "string" },
    "evidence_merkle_root": { "type": "string", "pattern": "^[0-9a-fA-F]{64}$" },
    "on_chain_freeze_dispatched": { "type": "boolean" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "alert_id",
    "isin",
    "symbol",
    "severity",
    "composite_risk_score",
    "primary_typology",
    "involved_venues",
    "evidence_merkle_root",
    "on_chain_freeze_dispatched",
    "timestamp_utc"
  ]
}
```

#### Topic: `regulatory.filing.sebi_alert.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "RegulatoryFilingSebiAlertEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "dossier_id": { "type": "string", "format": "uuid" },
    "alert_id": { "type": "string", "format": "uuid" },
    "target_regulator": { "type": "string", "enum": ["SEBI", "IFSCA", "BOTH"] },
    "statutory_reference_number": { "type": "string" },
    "statutory_violation": { "type": "string" },
    "isin": { "type": "string" },
    "symbol": { "type": "string" },
    "participant_pan_hashes": { "type": "array", "items": { "type": "string" } },
    "estimated_illicit_gain_inr": { "type": "number" },
    "evidence_merkle_root": { "type": "string" },
    "dossier_pdf_url": { "type": "string" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "dossier_id",
    "alert_id",
    "target_regulator",
    "statutory_reference_number",
    "statutory_violation",
    "isin",
    "symbol",
    "evidence_merkle_root",
    "timestamp_utc"
  ]
}
```

## Security & Compliance Notes
- **SEBI PFUTP Statutory Enforcement:** Enforces strict adherence to SEBI PFUTP Regulations 3 and 4, ensuring zero tolerance for fraudulent, deceptive, or manipulative cross-venue trading practices across domestic and tokenized equity markets.
- **IFSCA Market Conduct Regulations Compliance:** Fully satisfies IFSCA Market Conduct rules governing capital markets operating in GIFT City, providing bi-directional surveillance between domestic rupee markets and foreign currency tokenized equity markets.
- **Clock Drift & Nanosecond Synchronization:** All surveillance nodes synchronize via IEEE 1588 PTP master clocks. If clock drift between NBSE nodes and external exchange feed taps exceeds $\pm 1\text{ms}$, the system automatically generates an operations alert and marks recorded events with a temporal uncertainty flag.
- **Immutable WORM Compliance Storage:** All raw order lifecycle events, reconstructed order books, and detection telemetry are stored in ClickHouse and S3 with Write-Once-Read-Many (WORM) retention locks for 8 years, satisfying SEBI and IFSCA record-keeping mandates.
- **Strict Zero-PII Ledger Governance:** On-chain records only contain salted HMAC hashes of investor identifiers and cryptographic Merkle roots of evidence. Plaintext investor names, PAN numbers, bank accounts, and addresses are strictly maintained off-chain inside the encrypted PostgreSQL database.
- **Dual-Control Alert Disposition:** Escalating, de-escalating, or closing a `CRITICAL` surveillance alert requires dual-control verification from the Chief Compliance Officer and the Principal Surveillance Officer with hardware token MFA (FIDO2 / WebAuthn).

## Acceptance Criteria
- [ ] Ingests and normalizes order lifecycle streams across NBSE, NSE, BSE, and MCX at sustained throughput $\ge 200,000$ events/second with p99 processing latency $< 5\text{ms}$.
- [ ] Successfully detects 100% of synthetic Cross-Venue Spoofing test scenarios where quotes injected on NBSE are cancelled within $< 500\text{ms}$ after opposite fills execute on NSE.
- [ ] Successfully detects closed circular trading rings ($3 \le k \le 8$ hops) across multi-account clusters within $< 1,000\text{ms}$ of loop completion.
- [ ] Identifies synchronized wash trades between accounts sharing common PAN, LEI, or device hardware IDs across venues with $0\%$ false dismissal on synthetic test vectors.
- [ ] Automatically dispatches `surveillance.cross_market.freeze_command.v1` and triggers on-chain account freeze on Hyperledger Besu for manipulation events with risk score $\ge 95.0$.
- [ ] Generates valid, cryptographically verified SEBI PFUTP and IFSCA statutory filing dossiers with complete order book depth reconstructions and evidence Merkle roots.
- [ ] PTP clock synchronization drift across all ingestor nodes remains within $\pm 500\mu\text{s}$ under full load.
- [ ] Code coverage for unit, integration, and mock cross-market surveillance test suites exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 104 (Kafka Event Schemas), Prompt 205 (Order Matching Engine), Prompt 207 (Market Data Service), Prompt 225 (FIX Gateway), Prompt 711 (Market Surveillance Anti-Manipulation).
- **Parallel Tasks:** Prompt 712 (Insider Trading Graph Analytics), Prompt 713 (Continuous Market Circuit Breaker), Prompt 714 (Cross-Chain AML & Forensics Screener).
- **Downstream Dependents:** Prompt 216 (Regulatory Reporting Service), Prompt 233 (Continuous 24x7 Regulatory Reporting), Prompt 604 / 605 (Admin & Compliance Consoles).
