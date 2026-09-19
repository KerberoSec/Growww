# 258 - Real-Time National Best Bid and Offer (NBBO) Consolidated Tape Engine (Go / Rust)

## Purpose
In a modern multi-exchange capital market architecture, trading liquidity for identical financial instruments is fragmented across multiple trading venues. In the Indian financial landscape, primary equity and derivative liquidity resides on the National Stock Exchange of India (NSE) and the Bombay Stock Exchange (BSE), commodity and bullion liquidity resides on the Multi Commodity Exchange (MCX), while 24/7 asset-backed continuous token trading operates natively on the National Blockchain Stock Exchange (NBSE) order matching books.

Without an authoritative, centralized consolidation layer, fragmented quotation streams produce information asymmetry, wide synthetic spreads, stale quotation references, and predatory latency arbitrage where high-frequency trading (HFT) participants exploit cross-venue dissemination delays at the expense of retail and institutional participants.

The **Real-Time National Best Bid and Offer (NBBO) Consolidated Tape Engine** (`services/nbbo-tape-engine`) serves as the central market data spine for the Growww / NBSE trading ecosystem. It continuously ingests top-of-book (Level-1 BBO) and market depth streams across NSE, BSE, MCX, and the internal NBSE continuous 24/7 matching books. The engine synthesizes an authoritative National Best Bid and Offer (NBBO), aggregates depth across price points, detects locked and crossed market conditions, generates smart execution venue routing hints, publishes unified composite trade tapes (Tape A for NSE, Tape B for BSE, Tape C for MCX, and Tape N for NBSE), and provides decentralized oracle feeds for on-chain proof-of-reserve valuation in strict compliance with SEBI Best Execution mandates.

## What You Are Building
An ultra-low latency, deterministic consolidated tape and NBBO calculation engine (`services/nbbo-tape-engine`) implemented in a hybrid Go and Rust architecture: Rust powers the zero-allocation, sub-microsecond NBBO calculation kernel and shared-memory ring buffers, while Go manages external network ingestion, Kafka event distribution, and gRPC streaming services.

Concrete deliverables include:
- **Multi-Venue Ingestion Feed Handler:** High-throughput UDP multicast listeners and TCP stream decoders ingesting raw binary Level-1/Level-2 feeds from NSE (NOW / NEAT), BSE (BOLT Plus), MCX, and internal Aeron IPC streams from the NBSE Matching Engine (Prompt 205).
- **Lock-Free In-Memory Top-of-Book Matrix:** A cache-line-aligned, lock-free price book matrix tracking real-time bid, ask, and volume per venue for every canonical International Securities Identification Number (ISIN).
- **Sub-Microsecond NBBO Synthesizer:** Deterministic evaluation kernel computing National Best Bid (highest price across all active venues), National Best Offer (lowest price across all active venues), consolidated spread, and aggregate depth at the inside market in under 2.5 microseconds.
- **Locked & Crossed Market Detector:** Real-time state machine flagging locked markets (NBB equals NBO) and crossed markets (NBB exceeds NBO), calculating inverted spread magnitudes and emitting regulatory alert telemetry.
- **Execution Venue Routing Hints Generator:** Computes optimal venue allocation vectors based on venue depth, tick latency profiles, and exchange transaction fees to supply the Smart Order Router (Prompt 242) with actionable routing hints.
- **Unified Composite Trade Tape (Tapes A, B, C, N):** Chronologically orders multi-venue trade prints using IEEE 1588 Precision Time Protocol (PTP) nanosecond hardware timestamps, publishing an interleaved composite sales tape.
- **Shared-Memory IPC & Streaming Distribution:** Lock-free single-producer multi-consumer (SPMC) ring buffers and high-performance gRPC streaming server broadcasting consolidated quotes to internal risk, matching, and surveillance engines.
- **Redis Cluster Snapshot & Kafka Tape Publisher:** Asynchronous pipelines updating live Redis 7.2 NBBO hashes for sub-millisecond front-end caching and streaming consolidated ticks to Kafka topics (`market.consolidated.tape.v1`, `market.nbbo.quotes.v1`).
- **Decentralized Oracle Price Feed Relayer:** Signs and dispatches periodic Volume-Weighted National Midpoint Price attestations to Hyperledger Besu for on-chain Proof-of-Reserve valuation references.

## Scope Boundaries
- **In Scope:**
  - Ingestion and normalization of multi-venue Level-1 and Level-2 tick feeds across NSE, BSE, MCX, and internal NBSE continuous 24/7 books.
  - Sub-microsecond computation of authoritative consolidated NBBO (National Best Bid, National Best Offer, aggregate sizes, and venue attributions).
  - Detection, logging, and dispatching of locked and crossed market states.
  - Calculation of depth-at-NBBO and consolidated spread metrics (quoted spread, effective spread).
  - Generation of execution venue routing hints for downstream order routing.
  - Low-latency distribution via Aeron IPC, POSIX shared memory, and gRPC streaming.
  - High-throughput distribution via Apache Kafka and Redis 7.2 Cluster.
  - Hardware timestamp sequencing using IEEE 1588 PTP nanosecond clocks.
  - Publication of decentralized oracle price feeds to Hyperledger Besu for Proof-of-Reserve valuation.
- **Out of Scope / Handled Elsewhere:**
  - Order execution and internal order book matching (handled in Prompt 205 Order Matching Engine).
  - Parent/child order splitting, FIX routing, and DMA order transmission to external venues (handled in Prompt 225 FIX Gateway and Prompt 242 SOR Adapter).
  - Daily historical Bhavcopy ETL and security master database updates (handled in Prompt 242).
  - Retail WebSocket distribution to Flutter and Next.js client applications (handled in Prompt 207 Market Data Service).
  - Pre-trade risk, margin calculations, and client balance validation (handled in Prompt 206 and Prompt 241).
  - On-chain token settlement and DvP smart contract execution (handled in Prompt 306).

## Technology to Use
- **Primary Languages & Runtimes:**
  - **Rust 1.78+:** Core NBBO synthesizer, lock-free venue price matrix, Aeron IPC consumers, and SIMD-optimized math kernel. Compiled with `opt-level = 3`, `lto = "fat"`, `codegen-units = 1`, and `panic = "abort"` for deterministic sub-microsecond latency.
  - **Go 1.22+:** Network feed orchestration, external gRPC streaming services, Kafka producers, Redis cluster synchronization, and Besu blockchain relayer.
- **Low-Latency Messaging & Ring Buffers:**
  - **Aeron Messaging:** High-throughput, low-latency IPC (`aeron:ipc`) and UDP unicast/multicast transport for inter-service communication between the matching engine and tape engine.
  - **Lock-Free Ring Buffers:** POSIX shared-memory memory-mapped queues with cache-line padding (64-byte alignment) to eliminate false sharing.
- **Network Ingestion & Sockets:**
  - Linux `epoll` and `io_uring` for high-concurrency UDP multicast socket ingestion of exchange feeds with kernel-bypass capabilities (`AF_XDP`).
- **In-Memory Cache & Snapshot Store:**
  - **Redis 7.2+ Cluster:** In-memory key-value store for microsecond retrieval of active NBBO snapshots and venue health heartbeats.
- **Event Streaming Backbone:**
  - **Apache Kafka 3.7+ (KRaft mode):** High-throughput log partitions for `market.consolidated.tape.v1` and `market.nbbo.quotes.v1`.
- **Time-Series Persistence:**
  - **PostgreSQL 16+ with TimescaleDB / ClickHouse:** Columnar analytical storage for tick-by-tick NBBO snapshots, spread analysis, and SEBI best execution compliance audits.
- **Inter-Service Communication:**
  - **gRPC over HTTP/2 with Protocol Buffers v3:** Zero-copy binary streaming APIs for consumption by the Smart Order Router and Risk Engines.
- **Time Synchronization:**
  - **IEEE 1588 PTP (Precision Time Protocol):** Sub-microsecond hardware timestamping across all venue ingestion network interface cards (NICs).

## Backend / Infra Touchpoints
- **External Feeds:**
  - NSE UDP Multicast (NEAT TBT / Level-2 5-depth stream).
  - BSE UDP Multicast (BOLT Plus Level-2 stream).
  - MCX Multicast Feed (Commodity tick streams).
- **Internal Feeds:**
  - NBSE Matching Engine (Prompt 205) Aeron IPC stream (`aeron:ipc?stream=1001`).
  - Synthetic 24/7 Market Gateway (Prompt 253) Kafka tick stream (`market.synthetic.ticks.v1`).
- **Redis 7.2 Key Patterns:**
  - `tape:nbbo:{isin}`: Hash containing `best_bid_price`, `best_bid_qty`, `best_bid_venue`, `best_ask_price`, `best_ask_qty`, `best_ask_venue`, `spread_paise`, `spread_bps`, `timestamp_ns`.
  - `tape:venue:{isin}:{venue}`: Hash containing venue-specific Level-1 top-of-book and latency metrics.
  - `tape:routing_hint:{isin}`: Hash containing primary venue, secondary venue, and split ratio vector.
  - `tape:market_condition:{isin}`: String flag indicating `NORMAL`, `LOCKED`, or `CROSSED`.
- **Kafka Topics:**
  - Consumes: `exchange.feed.nse.v1`, `exchange.feed.bse.v1`, `exchange.feed.mcx.v1`, `matching.trades.v1`, `matching.quotes.v1`.
  - Publishes: `market.consolidated.tape.v1`, `market.nbbo.quotes.v1`, `market.nbbo.locked_crossed.v1`, `oracle.price.reference.v1`.
- **PostgreSQL / TimescaleDB Tables:**
  - `nbbo_consolidated_ticks`: Granular tick-by-tick tape history.
  - `nbbo_spread_metrics`: Rolling 1-second and 1-minute spread and liquidity analytics.
  - `nbbo_locked_crossed_events`: Audit log of market abnormalities and resolution latencies.
  - `nbbo_oracle_attestations`: Historical record of price oracle attestations submitted on-chain.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Decentralized Oracle Price Feed Publishing:** The engine calculates the authoritative 1-second Volume-Weighted National Midpoint Price (VWNMP) and submits cryptographic attestations to `OraclePriceFeedRegistry.sol` on Hyperledger Besu.
- **Proof-of-Reserve Valuation References:** On-chain Proof-of-Reserve contracts (Prompt 308) and the Reconciliation Service (Prompt 215) consume these verified midpoint price feeds to compute the mark-to-market value of physical depository holdings (NSDL/CDSL demat accounts and bullion vaults) versus issued ERC-3643 tokens.
- **Zero-PII Compliance:** All on-chain price attestations are purely financial telemetry containing ISIN, timestamp, midpoint price, aggregate multi-venue volume, and cryptographic validator signature. No investor identities, account numbers, or individual order details are ever published to the ledger.
- **Deterministic Consensus Finality:** Attestations are committed via QBFT consensus (2.0s block time), ensuring tamper-proof historical price records for regulatory auditability.

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the monorepo workspace `services/nbbo-tape-engine` containing the Rust core calculation crate (`crates/nbbo-core`) and the Go orchestration service (`cmd/nbbo-service`).
2. Author Protobuf definitions in `proto/growww/tape/v1/tape_service.proto` defining `ConsolidatedQuote`, `ConsolidatedTrade`, `NBBOUpdate`, `VenueRoutingHint`, and gRPC streaming service interfaces.
3. Generate Go and Rust stubs using `buf` and `tonic-build` configured in their respective build pipelines.
4. Create PostgreSQL database migration scripts in `services/nbbo-tape-engine/migrations/001_nbbo_tape_schema.sql` defining relational tables for tick archives, spread metrics, and market condition events.
5. Implement zero-allocation binary protocol decoders in Rust for raw UDP multicast packets from NSE NEAT, BSE BOLT Plus, and MCX feeds.
6. Implement the lock-free Aeron IPC consumer in Rust subscribing to internal NBSE matching engine ticks from Prompt 205.
7. Implement the IEEE 1588 PTP Hardware Timestamp Synchronizer mapping packet arrival times to unified UTC nanosecond monotonic time.
8. Build the lock-free In-Memory Venue Quote Matrix using cache-line-aligned atomics to maintain per-venue bid/ask prices and quantities per ISIN.
9. Implement the Sub-Microsecond NBBO Synthesizer calculating National Best Bid, National Best Offer, aggregate sizes, and venue attributions with volume-weighted tie-breaking.
10. Implement the Locked and Crossed Market Detector, calculating inverted spread magnitude and generating immediate alerts when $P_{NBB} \ge P_{NBO}$.
11. Implement the Execution Venue Routing Hints Engine evaluating fee-adjusted net execution price and available liquidity depth per venue.
12. Implement the Composite Tick Tape Publisher streaming chronologically ordered trade prints across Tapes A (NSE), B (BSE), C (MCX), and N (NBSE).
13. Implement the Redis Cluster pipeline updating live `tape:nbbo:{isin}` hashes with asynchronous zero-copy writes.
14. Implement the Kafka event publisher streaming consolidated ticks to `market.consolidated.tape.v1` and `market.nbbo.quotes.v1`.
15. Implement the Besu Oracle Price Relayer packaging 1-second VWAP midpoint updates into EIP-712 signed transactions for `OraclePriceFeedRegistry.sol`.
16. Build comprehensive automated test suites including synthetic cross-venue tick interleaving, microsecond race condition harnesses, locked-market resolution benchmarks, and regulatory replay tests.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/tape/v1/tape_service.proto`)
```protobuf
syntax = "proto3";

package growww.tape.v1;

option go_package = "github.com/growww/proto/gen/go/tape/v1;tapev1";

enum MarketVenue {
  MARKET_VENUE_UNSPECIFIED = 0;
  MARKET_VENUE_NSE = 1;
  MARKET_VENUE_BSE = 2;
  MARKET_VENUE_MCX = 3;
  MARKET_VENUE_NBSE = 4;
}

enum TapeIdentifier {
  TAPE_IDENTIFIER_UNSPECIFIED = 0;
  TAPE_IDENTIFIER_A_NSE = 1;
  TAPE_IDENTIFIER_B_BSE = 2;
  TAPE_IDENTIFIER_C_MCX = 3;
  TAPE_IDENTIFIER_N_NBSE = 4;
}

enum MarketCondition {
  MARKET_CONDITION_NORMAL = 0;
  MARKET_CONDITION_LOCKED = 1;
  MARKET_CONDITION_CROSSED = 2;
  MARKET_CONDITION_HALTED = 3;
}

message VenueQuote {
  MarketVenue venue = 1;
  uint64 bid_price_paise = 2;
  uint64 bid_quantity = 3;
  uint64 ask_price_paise = 4;
  uint64 ask_quantity = 5;
  int64 venue_timestamp_ns = 6;
  int64 ingestion_timestamp_ns = 7;
}

message ConsolidatedQuote {
  string isin = 1;
  uint64 national_best_bid_paise = 2;
  uint64 national_best_bid_quantity = 3;
  MarketVenue national_best_bid_venue = 4;
  uint64 national_best_offer_paise = 5;
  uint64 national_best_offer_quantity = 6;
  MarketVenue national_best_offer_venue = 7;
  int64 spread_paise = 8;
  uint32 spread_bps = 9;
  MarketCondition condition = 10;
  repeated VenueQuote participating_venues = 11;
  int64 calculated_at_ns = 12;
  uint64 sequence_number = 13;
}

message ConsolidatedTrade {
  string isin = 1;
  TapeIdentifier tape = 2;
  MarketVenue venue = 3;
  string venue_trade_id = 4;
  uint64 trade_price_paise = 5;
  uint64 trade_quantity = 6;
  int64 trade_timestamp_ns = 7;
  int64 tape_sequence_number = 8;
  bool is_cross_trade = 9;
}

message VenueRoutingAllocation {
  MarketVenue venue = 1;
  uint32 allocation_percentage_bps = 2; // e.g. 6000 = 60.00%
  uint64 available_depth = 3;
  uint32 estimated_latency_micros = 4;
  uint32 venue_fee_bps = 5;
}

message VenueRoutingHint {
  string isin = 1;
  MarketVenue primary_venue = 2;
  MarketVenue secondary_venue = 3;
  repeated VenueRoutingAllocation allocations = 4;
  int64 generated_at_ns = 5;
}

message StreamNBBORequest {
  repeated string isins = 1;
}

message StreamConsolidatedTapeRequest {
  repeated string isins = 1;
  repeated TapeIdentifier tapes = 2;
}

message GetNBBORequest {
  string isin = 1;
}

message OraclePriceAttestation {
  string isin = 1;
  uint64 vwap_midpoint_price_paise = 2;
  uint64 aggregate_volume_shares = 3;
  int64 window_start_ns = 4;
  int64 window_end_ns = 5;
  bytes signature = 6;
  string attestation_hash = 7;
}

service NBBOConsolidatedTapeService {
  rpc GetNBBO (GetNBBORequest) returns (ConsolidatedQuote);
  rpc StreamNBBO (StreamNBBORequest) returns (stream ConsolidatedQuote);
  rpc StreamConsolidatedTape (StreamConsolidatedTapeRequest) returns (stream ConsolidatedTrade);
  rpc GetRoutingHint (GetNBBORequest) returns (VenueRoutingHint);
}
```

### 2. PostgreSQL Database Schema (`services/nbbo-tape-engine/migrations/001_nbbo_tape_schema.sql`)
```sql
CREATE TABLE nbbo_consolidated_ticks (
    tick_id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL,
    sequence_number BIGINT NOT NULL,
    best_bid_paise BIGINT NOT NULL,
    best_bid_qty BIGINT NOT NULL,
    best_bid_venue VARCHAR(16) NOT NULL,
    best_offer_paise BIGINT NOT NULL,
    best_offer_qty BIGINT NOT NULL,
    best_offer_venue VARCHAR(16) NOT NULL,
    spread_paise BIGINT NOT NULL,
    spread_bps INT NOT NULL,
    market_condition VARCHAR(16) NOT NULL DEFAULT 'NORMAL',
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nbbo_spread_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    window_interval VARCHAR(16) NOT NULL, -- 1s, 1m, 5m
    average_spread_paise NUMERIC(12, 2) NOT NULL,
    average_spread_bps NUMERIC(8, 2) NOT NULL,
    max_spread_paise BIGINT NOT NULL,
    min_spread_paise BIGINT NOT NULL,
    locked_state_duration_ms INT NOT NULL DEFAULT 0,
    crossed_state_duration_ms INT NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nbbo_locked_crossed_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    condition VARCHAR(16) NOT NULL, -- LOCKED, CROSSED
    bid_price_paise BIGINT NOT NULL,
    bid_venue VARCHAR(16) NOT NULL,
    offer_price_paise BIGINT NOT NULL,
    offer_venue VARCHAR(16) NOT NULL,
    inversion_magnitude_paise BIGINT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    duration_micros BIGINT
);

CREATE TABLE nbbo_oracle_attestations (
    attestation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    vwap_midpoint_price_paise BIGINT NOT NULL,
    aggregate_volume BIGINT NOT NULL,
    window_start_ns BIGINT NOT NULL,
    window_end_ns BIGINT NOT NULL,
    besu_tx_hash VARCHAR(66),
    attestation_signature TEXT NOT NULL,
    committed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nbbo_ticks_isin_seq ON nbbo_consolidated_ticks(isin, sequence_number DESC);
CREATE INDEX idx_nbbo_ticks_time ON nbbo_consolidated_ticks(calculated_at DESC);
CREATE INDEX idx_nbbo_events_isin_time ON nbbo_locked_crossed_events(isin, started_at DESC);
CREATE INDEX idx_nbbo_oracle_isin_time ON nbbo_oracle_attestations(isin, committed_at DESC);
```

## Security & Compliance Notes
- **SEBI Best Execution Guidelines:** SEBI requires broker-dealers and execution platforms to systematically achieve best execution for client orders. The NBBO tape engine establishes an immutable, tick-level reference benchmark against which every execution fill is verified for price improvement or slippage.
- **Latency Arbitrage Mitigation:** HFT algorithms exploit cross-exchange network latency variations to front-run quote changes across venues. The engine timestamps every packet using IEEE 1588 PTP hardware timestamps on the ingestion NIC, establishing a strict relativistic ordering and flagging stale quotes exceeding a 50-millisecond drift threshold.
- **Locked & Crossed Market Safeguards:** In fragmented electronic markets, rapid quote cancellations and asynchronous matching can temporarily cause locked or crossed quotes. The engine identifies these anomalies in sub-microsecond time and supplies protective flags to the Smart Order Router, preventing destructive circular execution loops.
- **Zero-PII Data Architecture:** The consolidated tape engine processes exclusively market microstructure data (ISINs, prices in paise, quantities, venue identifiers, timestamps). Zero trader account numbers, KYC identifiers, or personal data pass through or reside within the tape engine.
- **Hardware Clock Synchronization & Leap Second Resilience:** UTC clock synchronization is sustained via dual redundant PTP Grandmaster clocks using GPS disciplining, operating strictly on International Atomic Time (TAI) offsets to prevent NTP clock steps or leap-second corruption.

## Acceptance Criteria
- [ ] Protobuf service contracts compile cleanly with `buf` and `tonic-build` producing typed Go and Rust stubs with zero compiler warnings.
- [ ] Ingests UDP multicast packets and Aeron IPC messages at sustained rates exceeding 1,000,000 quotes per second with zero buffer overflows or dropped packets.
- [ ] End-to-end NBBO calculation latency from packet arrival to ring buffer dispatch is under 2.5 microseconds at the 99th percentile (p99).
- [ ] Correctly resolves consolidated Best Bid and Best Offer across NSE, BSE, MCX, and internal NBSE books with deterministic volume and time tie-breaking.
- [ ] Detects and logs locked ($P_{NBB} = P_{NBO}$) and crossed ($P_{NBB} > P_{NBO}$) market states within 500 nanoseconds of quote matrix update.
- [ ] Composite tick tape correctly interleaves trades across Tapes A, B, C, and N into a strictly monotonic chronological stream.
- [ ] Generates dynamic execution venue routing hints reflecting real-time depth, venue latency, and statutory transaction fees.
- [ ] Redis Cluster snapshot updates maintain sub-millisecond data freshness for all active ISINs.
- [ ] Kafka topics `market.consolidated.tape.v1` and `market.nbbo.quotes.v1` receive continuous partitioned tick events without lag.
- [ ] Submits periodic 1-second VWAP midpoint price attestations to `OraclePriceFeedRegistry.sol` on Hyperledger Besu with verified cryptographic signatures.
- [ ] Adheres strictly to the 12-section template with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `205` (Order Matching Engine), Prompt `242` (NSE / BSE Market Data Feeds & SOR Adapter).
- **Parallel Work:** Prompt `207` (Real-Time Market Data & WebSocket Streaming Service), Prompt `225` (FIX Protocol Gateway), Prompt `243` (MCX Commodity & Warehouse Receipt Adapter), Prompt `253` (Continuous 24/7 Synthetic Market & After-Hours Gateway), Prompt `256` (Limit-Up / Limit-Down LULD Volatility Dampener).
- **Subsequent Prompts Enabled:** Prompt `242` (Smart Order Routing Core Execution), Prompt `308` (On-Chain Proof-of-Reserve Attestation Registry), Prompt `508` (Flutter Security Detail Screen), Prompt `603` (Next.js Web Trading Dashboard).
