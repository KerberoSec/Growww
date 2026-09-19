# 242 - NSE / BSE Market Data Feeds, Bhavcopy Ingestion & Smart Order Routing (SOR) Adapter (Go)

## Purpose
To provide seamless liquidity integration and regulatory-grade price discovery for Growww's hybrid exchange architecture, the platform must bridge traditional Indian primary financial markets with its continuous 24/7 tokenized blockchain trading layer. The National Stock Exchange of India (NSE) and the Bombay Stock Exchange (BSE) represent the primary venues for equity and derivative liquidity in India, operating during standard market hours (09:15 to 15:30 IST).

The **NSE / BSE Market Data Feeds, Bhavcopy Ingestion & Smart Order Routing (SOR) Adapter** serves as the mission-critical gateway connecting Growww to domestic exchange infrastructure:
1. **Exchange Market Data Feeds:** Ingests, decodes, and normalizes high-frequency tick-by-tick (TBT) and Level-2/Level-3 market data feeds from NSE (NOW / NEAT) and BSE (BOLT / BOLT Plus).
2. **Automated Bhavcopy Ingestion:** Downloads, validates, and ingests end-of-day official Bhavcopy files and security master updates published by NSE (National Clearing Limited - NCL) and BSE (Indian Clearing Corporation Limited - ICCL).
3. **Security Master Normalization:** Maintains a unified, authoritative catalog mapping domestic security tokens, scrip codes, and series codes to canonical International Securities Identification Numbers (ISINs) with continuous corporate action adjustments.
4. **Smart Order Routing (SOR):** Evaluates real-time consolidated National Best Bid and Offer (NBBO) across NSE and BSE, dynamically routing parent and child orders to achieve best execution, minimize market impact, reduce transaction costs, and maximize fill probability in accordance with SEBI Best Execution guidelines.
5. **Market Hours & 24/7 Tokenized Liquidity Bridge:** Manages the operational transition between regular domestic trading sessions (Pre-Open 09:00-09:15 IST, Regular 09:15-15:30 IST, Post-Closing 15:40-16:00 IST) and off-market 24/7 tokenized trading on Growww's permissioned Hyperledger Besu blockchain, guaranteeing continuous price integrity and collateral synchronization.

## What You Are Building
A high-throughput, low-latency Go microservice (`services/exchange-adapter`) built for concurrent socket stream ingestion, automated file ETL, and microsecond-level order routing. Concrete deliverables include:
- **Exchange Market Data Ingestion Engine:** High-performance UDP multicast and TCP socket client ingesting binary and FIX/FAST data streams from NSE NEAT/NOW and BSE BOLT/FAST protocols, parsing Level-2 depth diffs and trade ticks into normalized internal event structures.
- **Automated Bhavcopy & Master File Ingestion Pipeline:** Automated cron-based and event-driven pipeline downloading daily Bhavcopy files (CSV/ZIP), verifying SHA-256 checksums, parsing equity, derivative, and index records, and populating the historical price database and security master catalog.
- **Canonical Security Master & Identifier Translator:** In-memory, bidirectional translation index mapping NSE symbol/token, BSE scrip code, Bloomberg ticker, and internal token contract addresses to standard ISIN codes (e.g., `INE002A01018` for Reliance Industries Limited).
- **Sub-Millisecond Smart Order Router (SOR Core):** Deterministic routing engine calculating synthetic NBBO across NSE and BSE, computing effective prices after statutory charges (Securities Transaction Tax [STT], exchange turnover fees, SEBI turnover fees, stamp duty), evaluating book depth, and splitting orders across venues.
- **Primary Exchange Order Execution Gateway:** Bidirectional order adapter converting internal gRPC order commands into exchange-compliant FIX 4.2/4.4/5.0 or native binary DMA messages, tracking execution reports, partial fills, rejections, and cancellations.
- **Market Session State Machine & 24/7 Bridge Orchestrator:** Real-time state machine coordinating exchange lifecycle phases (PRE_OPEN, CONTINUOUS, POST_CLOSE, OFF_MARKET) with 24/7 tokenized market trading, publishing synthetic price reference bands and adjusting liquidity pools during off-market hours.
- **Redis & PostgreSQL Dual-Tier Storage Layer:** Ultra-fast Redis 7.2 Cluster caching live top-of-book quotes, NBBO snapshots, and token lookup tables, paired with PostgreSQL 16+ storing historical Bhavcopy records, security master metadata, and SOR execution audit logs.

## Scope Boundaries
- **In Scope:**
  - Ingestion, framing, and zero-allocation parsing of raw market data feeds from NSE (NOW / NEAT) and BSE (BOLT / BOLT Plus).
  - Automated downloading, parsing, schema validation, and storage of daily Bhavcopy files (NSE and BSE Cash and F&O segments).
  - Canonical Security Master management: ISIN resolution, symbol mapping, series handling (EQ, BE, BZ, SM), tick sizes, lot sizes, and daily price bands (upper/lower circuits).
  - Real-time consolidated NBBO computation across NSE and BSE venues.
  - Smart Order Routing algorithm: liquidity-based splitting, price improvement detection, fee optimization, and exchange latency compensation.
  - Outbound order formatting and inbound execution report translation for domestic exchange Direct Market Access (DMA) connectivity.
  - Session state synchronization bridging primary market trading hours (09:15-15:30 IST) with continuous 24/7 tokenized blockchain liquidity on Hyperledger Besu.
- **Out of Scope / Handled Elsewhere:**
  - Matching engine order book for the internal 24/7 tokenized market (handled in Prompt 205).
  - Pre-trade risk margin validation, user balance verification, and leverage limits (handled in Prompt 206 and 241).
  - Double-entry cash wallet ledgers and fiat banking gateways (handled in Prompt 203 and 212).
  - Direct depository integration with NSDL/CDSL for demat account settlement and share immobilization (handled in Prompt 213).
  - On-chain Delivery-versus-Payment (DvP) smart contract execution (handled in Prompt 306).
  - Retail WebSocket fan-out to Flutter and Web client applications (handled in Prompt 207).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Go provides exceptional concurrent I/O performance via lightweight goroutines, predictable memory management with sub-millisecond GC pauses, and native networking primitives for UDP multicast and TCP streams.
- **Exchange Feed & Network Protocols:**
  - Raw socket listeners using `golang.org/x/net/ipv4` for UDP multicast group subscription (NSE/BSE tick feeds).
  - `github.com/quickfixgo/quickfix` for standard FIX 4.2 / 4.4 / 5.0 SP2 institutional DMA order routing and drop-copy processing.
  - Custom binary protocol framing and zero-allocation byte decoders for NEAT and BOLT binary payloads.
- **In-Memory Cache & Fast State Store:** Redis 7.2+ Cluster for live top-of-book caching, NBBO state matrices, venue health heartbeats, and sub-millisecond token-to-ISIN index queries.
- **Relational Database:** PostgreSQL 16+ with `pgx/v5` connection pooling and `sqlc` for type-safe SQL queries, managing the canonical Security Master catalog, daily Bhavcopy archives, and SOR execution audit trails.
- **Messaging Backbone:** Apache Kafka 3.7+ (KRaft mode) using `segmentio/kafka-go` for streaming normalized ticks (`exchange.feed.nse.v1`, `exchange.feed.bse.v1`), Bhavcopy sync notifications (`exchange.bhavcopy.synced.v1`), and SOR execution events (`order.routing.executed.v1`).
- **Inter-Service Communication:** gRPC over HTTP/2 with Protocol Buffers v3 (`google.golang.org/grpc`) for ultra-low-latency inter-service routing and security master queries.
- **Compression & Parsing:** `compress/gzip`, `compress/zlib`, and `archive/zip` for handling compressed daily Bhavcopy archives from exchange servers.

## Backend / Infra Touchpoints
- **Redis 7.2 Key Patterns:**
  - `exchange:quote:{isin}:nse`: Hash containing NSE LTP, Best Bid/Ask, volume, and timestamp.
  - `exchange:quote:{isin}:bse`: Hash containing BSE LTP, Best Bid/Ask, volume, and timestamp.
  - `exchange:nbbo:{isin}`: Hash containing consolidated Best Bid (price, qty, venue), Best Ask (price, qty, venue), spread, and calculation timestamp.
  - `exchange:token_map:nse:{token_id}`: String mapping NSE token integer to ISIN string.
  - `exchange:token_map:bse:{scrip_code}`: String mapping BSE scrip code integer to ISIN string.
  - `exchange:circuits:{isin}`: Hash storing upper band, lower band, daily price reference, and circuit filter percentage (5%, 10%, 20%).
  - `exchange:session:status`: Hash tracking current session state (PRE_OPEN, CONTINUOUS, POST_CLOSE, OFF_MARKET) for NSE and BSE.
- **PostgreSQL 16 Tables:**
  - `exchange_security_master`: Authoritative instrument definitions, ISIN, symbol, company name, asset class, tick size, lot size, face value, status.
  - `exchange_security_identifiers`: Mapping records linking ISIN to NSE token, NSE symbol/series, BSE scrip code, BSE scrip ID, and Bloomberg/Reuters identifiers.
  - `exchange_bhavcopy_records`: Daily historical trading metrics per security (Open, High, Low, Close, VWAP, Total Volume, Total Turnover, Total Trades, Delivery Quantity, Delivery Percentage, 52-Week High/Low).
  - `exchange_market_sessions`: Market calendar, holidays, trading session transition timestamps, and settlement schedules.
  - `exchange_sor_routing_rules`: Configuration parameters for SOR routing strategies, venue priority weights, fee tables, and latency adjustments.
  - `exchange_sor_executions`: Audit trail of all smart order routing decisions, capturing pre-route NBBO quotes, selected venue(s), executed price, slippage, and routing rationale.
- **Apache Kafka Topics:**
  - Consumes:
    - `order.placement.requested.v1`: Inbound orders requiring external exchange liquidity routing.
    - `custody.asset_immobilized.v1`: Depository notification of immobilized physical shares enabling tokenized trading.
  - Publishes:
    - `exchange.feed.ticks.v1`: Normalized real-time price ticks across NSE and BSE.
    - `exchange.feed.depth.v1`: Normalized Level-2/Level-3 book depth updates from domestic exchanges.
    - `exchange.bhavcopy.ingested.v1`: Broadcast event upon complete daily Bhavcopy parsing and database load.
    - `exchange.session.state_changed.v1`: Notifications of market phase transitions (e.g., regular market open/close).
    - `order.primary_execution.v1`: Execution confirmations, partial fills, and cancel confirmations from primary exchanges.
- **Microservice Interactions:**
  - Order Management Service (Prompt 204): Routes primary market order execution requests to the adapter.
  - Order Matching Engine (Prompt 205): Synchronizes off-market price bands and reference prices with the adapter.
  - Real-Time Market Data Service (Prompt 207): Ingests normalized ticks for client broadcast and candlestick compilation.
  - Custodian & Depository Integration Service (Prompt 213): Reconciles daily Bhavcopy prices for depository portfolio valuation.
  - Real-Time Market Surveillance Engine (Prompt 228): Monitors routing compliance and venue execution quality.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Security Master Canonical Hash Registry:** The Security Master definitions and ISIN mappings generated by this service are hashed and registered in the `TokenIssuance.sol` (Prompt 303) and `ProofOfReserveRegistry.sol` (Prompt 308) smart contracts on Hyperledger Besu, ensuring immutable linking between token contracts (`DigitalSecurityToken`) and real-world ISINs.
- **Proof-of-Reserve Daily Valuation Anchor:** The official closing prices, VWAP, and total traded turnover ingested from daily Bhavcopy files provide the official external price reference used by the Proof-of-Reserve Reconciliation Engine (Prompt 215) to calculate total Asset Under Custody (AUC) valuations against custodian demat holdings.
- **Market Hours Bridge & Off-Market Liquidity Synchronization:** When domestic exchanges close at 15:30 IST, the adapter executes a reconciliation handshake:
  1. Captures final NSE/BSE official closing prices from Bhavcopy.
  2. Anchors the initial reference price for the continuous 24/7 tokenized market on Hyperledger Besu.
  3. Signals the internal matching engine (Prompt 205) to activate dynamic synthetic price collars (e.g., +/- 10% from official close) for continuous token trading.
  4. At 09:00 IST the following trading day, aggregates overnight tokenized volume and price discovery to inform pre-open session orders.
- **Zero PII on Wire and On-Chain:** All market data ticks, Bhavcopy archives, and outbound exchange routing packets contain strictly institutional identifiers, ISINs, venue codes, and pseudonymous account identifiers. No personal investor data (PAN, Aadhaar, name, bank account) is ever transmitted over exchange feeds or published to the blockchain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Directory & Architecture:** Initialize the Go workspace `services/exchange-adapter` following standard Go layout conventions (`cmd/adapter/`, `internal/feed/`, `internal/bhavcopy/`, `internal/master/`, `internal/sor/`, `internal/gateway/`, `internal/session/`, `pkg/models/`).
2. **Define Protocol Buffers & Interfaces:** Write `.proto` definitions for `ExchangeAdapterService`, `SmartOrderRoutingService`, and `SecurityMasterService`, generating Go stubs with protoc and gRPC plugins.
3. **Implement Database Schema & Migration:** Write PostgreSQL migration scripts creating `exchange_security_master`, `exchange_security_identifiers`, `exchange_bhavcopy_records`, `exchange_market_sessions`, `exchange_sor_routing_rules`, and `exchange_sor_executions` tables with optimized indexes on `isin`, `trading_symbol`, `exchange_token`, and `record_date`.
4. **Implement NSE / BSE Socket Ingestion Engine:** Build asynchronous UDP multicast and TCP socket listeners with ringbuffer pools to ingest raw tick and Level-2 depth broadcasts from NSE NEAT/NOW and BSE BOLT protocols, implementing zero-allocation byte parsers.
5. **Build Normalized Feed Dissector & Kafka Broadcaster:** Translate raw exchange packet structures into unified `MarketTick` and `OrderBookDepth` structs, publishing them to `exchange.feed.ticks.v1` and `exchange.feed.depth.v1` via Kafka and updating Redis quote hashes.
6. **Implement Automated Bhavcopy Ingestion Engine:** Build an automated scheduler and HTTPS/FTP scraper to fetch daily Bhavcopy files (NSE Cash `cm<date>bhav.csv.zip`, BSE Cash `Equity_<date>.csv`, NSE F&O `fo<date>bhav.csv.zip`), verify cryptographic SHA-256 hashes, decompress archives, parse row streams, and batch-insert records into PostgreSQL.
7. **Build Security Master Synchronization Pipeline:** Implement pipeline to parse NSE `security.gz` and BSE `scrip_master.csv` daily master files, resolving corporate actions (splits, bonuses, symbol changes), updating circuit filter bands, and maintaining in-memory lookup maps in Redis.
8. **Develop In-Memory Synthetic NBBO Engine:** Create a concurrent quotation aggregator that continuously monitors real-time bids and asks from NSE and BSE for every active ISIN, computing the consolidated National Best Bid and Offer (NBBO) and tracking spread compression in sub-millisecond cycles.
9. **Implement Smart Order Routing (SOR) Algorithm:** Build the SOR core implementing venue selection logic:
   - Evaluates best available price across NSE and BSE.
   - Calculates effective net realization after venue-specific transaction fees, STT, and stamp duties.
   - Evaluates book depth at best price level; if order size exceeds top-of-book, calculates optimal multi-venue split ratios (e.g., 60% NSE, 40% BSE) to minimize market impact.
   - Selects route with lowest historical execution latency and highest fill probability.
10. **Implement Primary Exchange DMA Gateway:** Build outbound FIX 4.2/4.4 and native exchange session handlers managing outbound `NewOrderSingle`, `OrderCancelRequest`, `OrderCancelReplaceRequest`, and processing inbound `ExecutionReport` messages with monotonic sequence number tracking and state recovery.
11. **Implement Market Session State Machine & 24/7 Bridge:** Implement deterministic state manager tracking Indian market phases (PRE_OPEN, CONTINUOUS, POST_CLOSE, OFF_MARKET), managing pre-open auction transitions, anchoring daily closing prices for the 24/7 tokenized blockchain matching engine, and syncing overnight token discovery at market open.
12. **Implement gRPC & REST Administration Endpoints:** Expose gRPC service endpoints for real-time NBBO queries, manual routing overrides, Bhavcopy re-ingestion triggers, and security master searches.
13. **Add Prometheus Metrics & OpenTelemetry Tracing:** Instrument the service with Prometheus metrics (`exchange_feed_messages_total`, `exchange_feed_latency_microseconds`, `sor_routing_decisions_total`, `sor_price_improvement_paisa`, `bhavcopy_ingestion_duration_seconds`) and distributed OpenTelemetry spans.
14. **Build Unit, Conformance & Replay Test Suites:** Develop comprehensive test fixtures including synthetic UDP multicast stream generators, historical Bhavcopy sample archives, multi-venue order routing simulation tests, and network disconnect recovery tests.
15. **Execute High-Throughput Latency & Stress Benchmarks:** Benchmark the SOR engine and feed parser under loads of 100,000 incoming ticks/sec and 10,000 order routing evaluations/sec, verifying p99 internal processing latency under 50 microseconds and zero memory leaks.

## Interfaces / Contracts

### PostgreSQL Database Schema (DDL)
```sql
-- Canonical Security Master Table
CREATE TABLE exchange_security_master (
    id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL UNIQUE,
    trading_symbol VARCHAR(32) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    asset_class VARCHAR(16) NOT NULL DEFAULT 'EQUITY', -- EQUITY, ETF, PREFERENCE, DEBT
    series VARCHAR(8) NOT NULL DEFAULT 'EQ', -- EQ, BE, BZ, SM
    face_value NUMERIC(12, 4) NOT NULL DEFAULT 10.0000,
    lot_size INT NOT NULL DEFAULT 1,
    tick_size NUMERIC(8, 4) NOT NULL DEFAULT 0.0500,
    upper_circuit_limit NUMERIC(14, 4) NOT NULL,
    lower_circuit_limit NUMERIC(14, 4) NOT NULL,
    circuit_band_percent NUMERIC(5, 2) NOT NULL DEFAULT 20.00,
    is_trade_for_trade BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_tokenized_247_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Exchange Specific Identifier Mapping Table
CREATE TABLE exchange_security_identifiers (
    id BIGSERIAL PRIMARY KEY,
    security_master_id BIGINT NOT NULL REFERENCES exchange_security_master(id) ON DELETE CASCADE,
    isin VARCHAR(12) NOT NULL,
    exchange_code VARCHAR(8) NOT NULL, -- NSE, BSE
    exchange_token_id VARCHAR(32) NOT NULL,
    exchange_symbol VARCHAR(32) NOT NULL,
    exchange_series VARCHAR(8) NOT NULL,
    exchange_group VARCHAR(8), -- BSE Group: A, B, T, Z
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_exchange_identifier UNIQUE (exchange_code, exchange_token_id),
    CONSTRAINT uq_exchange_symbol_series UNIQUE (exchange_code, exchange_symbol, exchange_series)
);

-- Daily Bhavcopy Ingestion Table
CREATE TABLE exchange_bhavcopy_records (
    id BIGSERIAL PRIMARY KEY,
    record_date DATE NOT NULL,
    exchange_code VARCHAR(8) NOT NULL, -- NSE, BSE
    isin VARCHAR(12) NOT NULL,
    trading_symbol VARCHAR(32) NOT NULL,
    series VARCHAR(8) NOT NULL,
    open_price NUMERIC(14, 4) NOT NULL,
    high_price NUMERIC(14, 4) NOT NULL,
    low_price NUMERIC(14, 4) NOT NULL,
    close_price NUMERIC(14, 4) NOT NULL,
    last_traded_price NUMERIC(14, 4) NOT NULL,
    previous_close NUMERIC(14, 4) NOT NULL,
    vwap NUMERIC(14, 4) NOT NULL,
    total_traded_quantity BIGINT NOT NULL,
    total_traded_value NUMERIC(18, 4) NOT NULL,
    total_trades BIGINT NOT NULL,
    deliverable_quantity BIGINT,
    deliverable_percent NUMERIC(6, 2),
    fifty_two_week_high NUMERIC(14, 4),
    fifty_two_week_low NUMERIC(14, 4),
    raw_checksum_sha256 VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_bhavcopy_daily_record UNIQUE (record_date, exchange_code, isin)
);

-- Smart Order Routing (SOR) Executions Audit Table
CREATE TABLE exchange_sor_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_order_id VARCHAR(36) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    side VARCHAR(4) NOT NULL, -- BUY, SELL
    order_type VARCHAR(16) NOT NULL, -- LIMIT, MARKET, IOC
    total_quantity NUMERIC(16, 6) NOT NULL,
    limit_price NUMERIC(14, 4),
    routing_strategy VARCHAR(32) NOT NULL, -- BEST_PRICE, LIQUIDITY_SPLIT, LOWEST_FEE
    nse_best_bid NUMERIC(14, 4),
    nse_best_ask NUMERIC(14, 4),
    nse_bid_qty NUMERIC(16, 6),
    nse_ask_qty NUMERIC(16, 6),
    bse_best_bid NUMERIC(14, 4),
    bse_best_ask NUMERIC(14, 4),
    bse_bid_qty NUMERIC(16, 6),
    bse_ask_qty NUMERIC(16, 6),
    routed_venue VARCHAR(8) NOT NULL, -- NSE, BSE, SPLIT
    nse_allocated_qty NUMERIC(16, 6) NOT NULL DEFAULT 0,
    bse_allocated_qty NUMERIC(16, 6) NOT NULL DEFAULT 0,
    executed_weighted_price NUMERIC(14, 4),
    estimated_price_improvement_paisa NUMERIC(10, 4) NOT NULL DEFAULT 0,
    execution_status VARCHAR(16) NOT NULL, -- SUBMITTED, PARTIALLY_FILLED, FILLED, REJECTED
    routing_duration_micros BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Market Sessions Management Table
CREATE TABLE exchange_market_sessions (
    id SERIAL PRIMARY KEY,
    session_date DATE NOT NULL UNIQUE,
    is_trading_day BOOLEAN NOT NULL DEFAULT TRUE,
    pre_open_start_time TIMESTAMPTZ NOT NULL,
    pre_open_end_time TIMESTAMPTZ NOT NULL,
    regular_trading_start_time TIMESTAMPTZ NOT NULL,
    regular_trading_end_time TIMESTAMPTZ NOT NULL,
    post_closing_start_time TIMESTAMPTZ NOT NULL,
    post_closing_end_time TIMESTAMPTZ NOT NULL,
    current_phase VARCHAR(16) NOT NULL DEFAULT 'OFF_MARKET', -- PRE_OPEN, CONTINUOUS, POST_CLOSE, OFF_MARKET
    holiday_description VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Protocol Buffers Definition (`exchange_adapter.proto`)
```protobuf
syntax = "proto3";

package growww.exchangeadapter.v1;

option go_package = "growww/exchangeadapter/v1;exchangeadapterv1";

service ExchangeAdapterService {
  rpc GetNationalBestBidOffer (GetNBBORequest) returns (GetNBBOResponse);
  rpc RouteSmartOrder (RouteSmartOrderRequest) returns (RouteSmartOrderResponse);
  rpc GetSecurityMasterByISIN (GetSecurityMasterRequest) returns (GetSecurityMasterResponse);
  rpc IngestBhavcopyManual (IngestBhavcopyManualRequest) returns (IngestBhavcopyManualResponse);
  rpc GetMarketSessionStatus (GetMarketSessionStatusRequest) returns (GetMarketSessionStatusResponse);
}

message QuoteLevel {
  string price = 1;      // Formatted in INR with 4 decimal places
  string quantity = 2;   // Formatted to 6 decimal places (micro-shares)
  int32 order_count = 3;
}

message VenueQuote {
  string venue = 1;       // "NSE" or "BSE"
  QuoteLevel best_bid = 2;
  QuoteLevel best_ask = 3;
  string ltp = 4;
  string vwap = 5;
  int64 total_volume = 6;
  int64 timestamp_unix_nanos = 7;
}

message GetNBBORequest {
  string isin = 1;
}

message GetNBBOResponse {
  string isin = 1;
  string trading_symbol = 2;
  VenueQuote nse_quote = 3;
  VenueQuote bse_quote = 4;
  QuoteLevel national_best_bid = 5;
  string national_best_bid_venue = 6;
  QuoteLevel national_best_ask = 7;
  string national_best_ask_venue = 8;
  string spread = 9;
  int64 computed_at_unix_nanos = 10;
}

message RouteSmartOrderRequest {
  string parent_order_id = 1;
  string isin = 2;
  string side = 3; // "BUY" or "SELL"
  string order_type = 4; // "LIMIT", "MARKET", "IOC"
  string quantity = 5;
  string limit_price = 6;
  string time_in_force = 7; // "DAY", "IOC", "FOK"
  string routing_preference = 8; // "BEST_PRICE", "FASTEST_FILL", "MINIMUM_FEE"
}

message VenueOrderAllocation {
  string venue = 1; // "NSE" or "BSE"
  string allocated_quantity = 2;
  string allocated_limit_price = 3;
  string child_order_id = 4;
  string exchange_order_id = 5;
  string status = 6; // "PENDING", "FILLED", "REJECTED"
}

message RouteSmartOrderResponse {
  string parent_order_id = 1;
  string isin = 2;
  string routing_strategy_applied = 3;
  repeated VenueOrderAllocation allocations = 4;
  string executed_weighted_average_price = 5;
  string price_improvement_paisa = 6;
  int64 execution_latency_micros = 7;
  string overall_status = 8;
}

message GetSecurityMasterRequest {
  string isin = 1;
}

message GetSecurityMasterResponse {
  string isin = 1;
  string trading_symbol = 2;
  string company_name = 3;
  string series = 4;
  string tick_size = 5;
  int32 lot_size = 6;
  string upper_circuit_limit = 7;
  string lower_circuit_limit = 8;
  string nse_token_id = 9;
  string bse_scrip_code = 10;
  bool is_trade_for_trade = 11;
  bool is_tokenized_247_active = 12;
}

message IngestBhavcopyManualRequest {
  string exchange_code = 1; // "NSE" or "BSE"
  string record_date = 2;   // "YYYY-MM-DD"
  bool force_reprocess = 3;
}

message IngestBhavcopyManualResponse {
  string exchange_code = 1;
  string record_date = 2;
  int64 total_records_ingested = 3;
  string file_checksum_sha256 = 4;
  int64 processing_time_ms = 5;
  string status = 6;
}

message GetMarketSessionStatusRequest {
  string exchange_code = 1; // "NSE", "BSE", or "ALL"
}

message GetMarketSessionStatusResponse {
  string current_phase = 1; // "PRE_OPEN", "CONTINUOUS", "POST_CLOSE", "OFF_MARKET"
  bool is_regular_market_open = 2;
  bool is_tokenized_247_active = 3;
  int64 next_phase_transition_unix = 4;
  string active_reference_price_source = 5; // "LIVE_NBBO" or "BHAVCOPY_CLOSE"
}
```

### Kafka Event Schemas (JSON Schemas)

#### 1. Real-Time Market Tick (`exchange.feed.ticks.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "ExchangeMarketTickEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "exchange": { "type": "string", "enum": ["NSE", "BSE"] },
    "isin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "symbol": { "type": "string" },
    "token": { "type": "string" },
    "ltp": { "type": "string" },
    "ltq": { "type": "string" },
    "totalVolume": { "type": "integer" },
    "totalValue": { "type": "string" },
    "vwap": { "type": "string" },
    "openPrice": { "type": "string" },
    "highPrice": { "type": "string" },
    "lowPrice": { "type": "string" },
    "closePrice": { "type": "string" },
    "exchangeTimestampNanos": { "type": "integer" },
    "ingestTimestampNanos": { "type": "integer" }
  },
  "required": [
    "eventId",
    "exchange",
    "isin",
    "symbol",
    "token",
    "ltp",
    "totalVolume",
    "exchangeTimestampNanos",
    "ingestTimestampNanos"
  ]
}
```

#### 2. Bhavcopy Ingestion Completed (`exchange.bhavcopy.ingested.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "ExchangeBhavcopyIngestedEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "recordDate": { "type": "string", "format": "date" },
    "exchange": { "type": "string", "enum": ["NSE", "BSE"] },
    "segment": { "type": "string", "enum": ["EQUITY", "FNO", "INDEX"] },
    "totalRecordsProcessed": { "type": "integer" },
    "sourceFileUrl": { "type": "string" },
    "fileChecksumSha256": { "type": "string" },
    "ingestionDurationMs": { "type": "integer" },
    "completedAt": { "type": "string", "format": "date-time" }
  },
  "required": [
    "eventId",
    "recordDate",
    "exchange",
    "segment",
    "totalRecordsProcessed",
    "fileChecksumSha256",
    "completedAt"
  ]
}
```

## Security & Compliance Notes
- **Direct Market Access (DMA) Credential Isolation:** Exchange API keys, member CompIDs, encryption tokens, and VPN gateway certificates are securely loaded from HashiCorp Vault into volatile memory enclaves, preventing disk persistence or environment variable leaks.
- **SEBI Best Execution & Traceability Compliance:** Every Smart Order Routing decision produces an immutable audit record in `exchange_sor_executions` capturing the microsecond quote snapshot of both NSE and BSE books at decision time, verifying compliance with SEBI Best Execution standards.
- **Bhavcopy File Integrity & Cryptographic Validation:** All automated file downloads require SHA-256 cryptographic checksum matching against exchange-published verification digests prior to extraction and database staging. Corrupted or mismatched archives trigger immediate compliance alerts and halt master table updates.
- **Dynamic Circuit Breaker & Price Collar Enforcement:** The adapter enforces exchange-mandated daily circuit filters (5%, 10%, 20%) and trade-for-trade (TFT) price bounds. Any inbound routing attempt that violates daily exchange collars is rejected immediately at the adapter boundary.
- **Zero PII Transmission Guarantee:** Exchange network interfaces and Kafka topics handle only pseudonymous broker-dealer participant IDs, ISIN codes, scrip identifiers, and numerical quantities. No client personal data (PAN, KYC identifiers, bank accounts) is ever exposed to external exchange pipes or published over market feeds.

## Acceptance Criteria
- [ ] UDP multicast and TCP socket listeners ingest and decode NSE and BSE feed packets with sub-50 microsecond parsing latency under continuous load.
- [ ] Automated daily Bhavcopy pipeline downloads, verifies, parses, and loads 100% of NSE and BSE equity records into PostgreSQL within 60 seconds of file availability.
- [ ] In-memory Security Master maintains accurate bidirectional token/ISIN mappings, tick size tables, lot sizes, and daily circuit bounds for 5,000+ listed securities.
- [ ] In-memory NBBO aggregator updates consolidated best bid/ask across venues within 10 microseconds of tick arrival.
- [ ] Smart Order Router (SOR) evaluates liquidity depth, price improvement, and fee structures, executing multi-venue splits with measurable positive price improvement.
- [ ] Market session state machine coordinates pre-open, regular continuous, post-close, and off-market 24/7 tokenized blockchain phases without price jumps or race conditions.
- [ ] Complete SOR execution audit records are persisted in PostgreSQL with millisecond timestamps and full book snapshot telemetry.
- [ ] System handles continuous load of 50,000 incoming feed ticks/sec and 5,000 routing evaluations/sec with zero memory leaks and CPU utilization below 60%.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (Architecture Overview), 103 (API Standards), 104 (Kafka Topic Standards), 111 (Core Domain Models), 401 (PostgreSQL Schema), 402 (Redis Patterns).
- **Parallel Tasks:** 207 (Real-Time Market Data Service), 213 (Custodian Depository Integration), 225 (FIX Gateway), 407 (Master Data Management).
- **Downstream Blockers:** 204 (Order Lifecycle Service), 205 (Order Matching Engine), 206 (Risk Engine), 228 (Real-Time Market Surveillance).
