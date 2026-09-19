# 248 - Exchange Adapter Pre-Open Auction & Asymmetric Speed Bump Guard (Go)

## Purpose
In modern electronic financial markets, the market opening transition and the continuous trading session represent periods of heightened volatility, extreme order volume, and vulnerability to predatory High-Frequency Trading (HFT) latency arbitrage. The Indian stock exchanges, National Stock Exchange of India (NSE) and Bombay Stock Exchange (BSE), conduct a structured Pre-Open Auction session between 09:00:00 and 09:15:00 IST to achieve orderly price discovery, absorb overnight order flow, and dampen opening volatility. 

Between 09:00:00 and 09:07:00 IST (with random closure up to 09:08:00 IST), market participants submit, modify, and cancel orders in the order collection phase. At 09:08:00 IST, the exchange matching engine calculates the Indicative Equilibrium Price (IEP) and Indicative Opening Volume (IOV) via a single-price multilateral batch auction algorithm, followed by order execution and a mandatory buffer period prior to the continuous market open at 09:15:00 IST.

During regular continuous trading, market makers and retail liquidity providers face substantial adverse selection risks from collocated algorithmic traders who exploit sub-millisecond market data feed discrepancies across primary exchanges and the internal 24/7 tokenized market. Without defensive mechanisms, resting quotes are sniped before liquidity providers can cancel or adjust prices in response to external price shifts.

The **Exchange Adapter Pre-Open Auction & Asymmetric Speed Bump Guard** (`services/preopen-speedbump-engine`) solves these structural challenges through a dedicated Go microservice:
1. **Pre-Open Batch Manifest Injection (09:00:00-09:07:00 IST):** Aggregates After-Market Orders (AMO), scheduled systematic investment orders, and overnight tokenized market transitions into cryptographically verified batch manifests, streaming them to NSE NEAT and BSE BOLT Direct Market Access (DMA) gateways with deterministic pacing and rate-limit compliance.
2. **Dynamic Equilibrium Collar Realignment (09:08:00 IST):** Ingests official exchange Indicative Equilibrium Prices at the conclusion of the auction matching window, dynamically recalibrating synthetic circuit bands, reference pegs, and risk collars across internal matching engines and smart order routers before continuous trading begins.
3. **500-Microsecond Asymmetric Speed Bump:** Imposes a deterministic 500-microsecond intentional latency delay on non-display aggressive (taker) orders while permitting passive liquidity updates (maker quotes, modifications, cancellations) to execute with zero latency (0 microseconds), neutralizing toxic cross-venue latency arbitrage.
4. **Pegged Volatility Band Protection:** Dynamically tracks the consolidated National Best Bid and Offer (NBBO) midpoint with adaptive volatility collars, quarantining or rejecting aggressive orders that arrive outside verified micro-price bands during high-frequency quote dislocations.

## What You Are Building
A high-throughput, sub-microsecond concurrency Go microservice located in `services/preopen-speedbump-engine/`. Concrete architectural components include:
- **Pre-Open Order Aggregator & Manifest Compiler:** Collects queued After-Market Orders (AMO) and overnight tokenized trading handoffs from Kafka, validates account balances against pre-trade risk state, deduplicates orders, and compiles optimized batch manifests structured for exchange DMA line protocols.
- **DMA Batch Manifest Injection Engine:** High-performance TCP socket pipeline that executes deterministic, sequence-controlled injection of batch manifests to NSE NEAT and BSE BOLT DMA gateways between 09:00:00 and 09:07:00 IST, enforcing strict per-line throttling (up to 5,000 orders/sec per gateway socket) to eliminate exchange throttling rejections.
- **Equilibrium Price Collector & Dynamic Collar Realignment Engine:** Captures tick broadcasts and Level-2 auction depth at 09:08:00 IST, extracts the Indicative Equilibrium Price (IEP) and Indicative Opening Volume (IOV), computes equilibrium collar adjustments, and propagates realigned price limits to the Order Matching Engine (Prompt 205) and Smart Order Router (Prompt 242).
- **Asymmetric Speed Bump Pipeline (500us Delay Buffer):** Low-overhead, lock-free ring-buffer pipeline utilizing hardware monotonic clock counters (TSC) that delays aggressive liquidity-extracting order frames by exactly 500 microseconds while bypassing passive liquidity cancellations and quote updates with zero delay.
- **Pegged Volatility Band & Toxic Flow Classifier:** Real-time risk classifier that computes continuous rolling volatility envelopes (Bollinger/Kaufman dynamic spread bands) around the consolidated NBBO midpoint, identifying and quarantining incoming toxic snipe orders that attempt execution during quote lag anomalies.
- **Colocation Latency Arbitrage Detector:** Anomaly detection engine that correlates exchange timestamp deltas with internal ingress timestamps, detecting sub-millisecond predatory cross-venue arbitrage patterns and generating compliance audit records.
- **In-Memory Redis 7.2 State & Telemetry Store:** In-memory state layer tracking active collar limits, auction state transitions, speed bump queues, and latency histograms for sub-microsecond lookup.

## Scope Boundaries
- **In Scope:**
  - Aggregation, validation, and batch manifest assembly of overnight and AMO equity/derivative orders.
  - Automated, throttled socket injection of pre-open batch manifests to NSE and BSE DMA gateways between 09:00:00 and 09:07:00 IST.
  - Ingestion and validation of Indicative Equilibrium Prices (IEP) from exchange broadcast feeds at 09:08:00 IST.
  - Dynamic collar realignment and synchronization across internal matching engines and SOR adapters prior to 09:15:00 IST.
  - Implementation of the 500-microsecond asymmetric speed bump queue for aggressive orders with zero-delay pass-through for maker cancellations and amendments.
  - Real-time pegged volatility band calculation and quote protection filtering.
  - PostgreSQL schema definitions for batch manifests, auction snapshots, speed bump audits, and volatility bands.
  - Kafka event streaming for pre-open lifecycle transitions, collar realignments, and latency arbitrage interceptions.
  - gRPC service endpoints for manual auction triggers, collar overrides, speed bump configuration, and latency metrics.
- **Out of Scope / Handled Elsewhere:**
  - Base FIX 4.2/4.4 protocol session management and drop-copy handling (handled in Prompt 225).
  - Exchange socket listeners, Bhavcopy parsing, and primary Smart Order Routing (handled in Prompt 242).
  - Continuous matching engine execution for internal 24/7 tokenized markets (handled in Prompt 205).
  - Pre-trade margin calculations, user wallet balance validation, and collateral checks (handled in Prompt 206 and 241).
  - Retail WebSocket quote distribution to client applications (handled in Prompt 207).
  - Depository settlement and share immobilization (handled in Prompt 213).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Leverages lightweight goroutines, lock-free ring buffers (`sync/atomic`, unsafe pointer casting), and deterministic memory layouts. Garbage collection is managed using ballast memory structures and `GOMEMLIMIT` tuning to guarantee sub-millisecond GC latency.
- **Precision Timestamping & Hardware Clocks:** Intel/AMD Time Stamp Counter (RDTSC) and Go `time.Now()` monotonic clock abstractions for sub-microsecond timestamping and precise 500-microsecond delay enforcement.
- **Concurrency & Lock-Free Data Structures:** Lock-free Single-Producer Multi-Consumer (SPMC) and Multi-Producer Single-Consumer (MPSC) circular ring buffers for the speed bump pipeline, eliminating mutex lock contention on the critical path.
- **In-Memory Cache & State Store:** Redis 7.2 Cluster with pipelined EVALSHA Lua scripts for atomic collar retrieval, session status verification, and sub-millisecond speed bump queue management.
- **Relational Database:** PostgreSQL 16+ utilizing `pgx/v5` connection pooling and `sqlc` for compile-time verified SQL queries, storing pre-open batch manifests, equilibrium auction records, and latency arbitrage audit logs.
- **Messaging Backbone:** Apache Kafka 3.7+ (KRaft mode) with `segmentio/kafka-go` for streaming pre-open dispatch events (`auction.preopen.manifest.dispatched.v1`), collar realignments (`auction.equilibrium.realigned.v1`), and speed bump interceptions (`guard.speedbump.intercepted.v1`).
- **Inter-Service Protocol:** gRPC over HTTP/2 with Protocol Buffers v3 (`google.golang.org/grpc`) for high-speed inter-service communication with the Order Matching Engine (Prompt 205) and Smart Order Router (Prompt 242).

## Backend / Infra Touchpoints
- **Redis 7.2 Key Patterns:**
  - `preopen:manifest:{manifest_id}`: Hash storing manifest metadata, batch status, total order count, and exchange acknowledgement metrics.
  - `auction:equilibrium:{isin}`: Hash containing Indicative Equilibrium Price (IEP), Indicative Opening Volume (IOV), total buy/sell imbalances, and capture timestamp.
  - `collar:realignment:{isin}`: Hash storing realigned dynamic price collars (upper collar, lower collar, reference price, realignment timestamp).
  - `speedbump:config:{isin}`: Hash storing asymmetric speed bump parameters (delay duration in microseconds, active status, exempt participant IDs).
  - `speedbump:queue:{shard_id}`: Redis Stream / Sorted Set tracking queued aggressive orders undergoing the 500us latency delay.
  - `volatility:band:{isin}`: Hash storing current NBBO midpoint, upper volatility threshold, lower volatility threshold, and tick update timestamp.
- **PostgreSQL 16 Tables:**
  - `preopen_batch_manifests`: Master record of all compiled pre-open auction batches, target venue, total orders, aggregated notional, SHA-256 manifest digest, and dispatch status.
  - `preopen_manifest_orders`: Granular mapping of individual parent order IDs included in each batch manifest, allocation sequence, and gateway submission timestamps.
  - `auction_equilibrium_snapshots`: Historical auction clearing prices, indicative volume, calculated price shift from previous close, and collar adjustment factors.
  - `speedbump_arbitrage_audits`: Immutable audit trail of delayed aggressive orders, intercepted toxic quote snipes, time differentials, and protection actions taken.
  - `pegged_volatility_bands`: Historical record of dynamic volatility collars, NBBO midpoint adjustments, and circuit collar realignment events.
- **Apache Kafka Topics:**
  - Consumes:
    - `order.placement.amo.v1`: Queued After-Market Orders and pre-open candidate orders.
    - `exchange.feed.ticks.v1`: Real-time tick broadcast from NSE/BSE containing auction depth and IEP updates.
    - `order.routing.requested.v1`: Inbound aggressive orders subject to speed bump evaluation.
  - Publishes:
    - `auction.preopen.manifest.dispatched.v1`: Manifest transmission events to exchange DMA lines.
    - `auction.equilibrium.realigned.v1`: Realignment events broadcast when 09:08:00 IST equilibrium prices are locked.
    - `guard.speedbump.intercepted.v1`: Interception notifications for predatory toxic orders delayed or rejected by the speed bump.
    - `guard.volatility_band.updated.v1`: Broadcast of recalibrated volatility envelopes and dynamic collars.
- **Microservice Interactions:**
  - Order Lifecycle Service (Prompt 204): Submits queued AMO orders for pre-open manifest aggregation.
  - Order Matching Engine (Prompt 205): Receives realigned price collars at 09:08:00 IST to establish opening limit bounds.
  - Pre-Trade Risk Engine (Prompt 206): Validates aggregated batch manifest limits and speed bump exemption privileges.
  - Smart Order Router (Prompt 242): Receives live speed bump routing clearances and equilibrium collar constraints.
  - Real-Time Market Surveillance Engine (Prompt 228): Ingests speed bump arbitrage audit logs for HFT manipulation analysis.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Pre-Open Auction Equilibrium Commitment:** The final Indicative Equilibrium Prices (IEP) and cumulative traded volumes established during the 09:00-09:08 IST auction window are cryptographically hashed into an epoch state root and notarized to `MarketSessionRegistry.sol` on Hyperledger Besu under QBFT consensus. This provides an immutable, tamper-evident audit record of opening price discovery.
- **Dynamic Collar State Hash Anchoring:** Daily dynamic price collar realignments are anchored on-chain prior to 09:15:00 IST, ensuring that synthetic derivatives and tokenized asset trading pools enforce mathematical price boundaries consistent with primary exchange opening values.
- **Latency Arbitrage & Speed Bump Audit Proofs:** Aggregated daily summaries of speed bump interceptions, rejected predatory snipe attempts, and microsecond latency arbitrage events are batched into cryptographic Merkle trees and anchored to `SurveillanceAuditRegistry.sol` for regulatory compliance verification under SEBI algorithmic trading mandates.
- **Zero PII on Wire and On-Chain:** All pre-open batch manifests, speed bump event streams, and blockchain notarization payloads contain strictly pseudonymous order IDs, ISIN identifiers, token codes, numerical price levels, and microsecond execution timestamps. No client personal data (PAN, Aadhaar, name, demat account number) is ever processed or persisted.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Go Module & Workspace Architecture:** Scaffold `services/preopen-speedbump-engine` following standard clean architecture conventions (`cmd/engine/`, `internal/manifest/`, `internal/injector/`, `internal/equilibrium/`, `internal/speedbump/`, `internal/volatility/`, `internal/storage/`, `pkg/models/`). Configure high-performance runtime flags and compiler optimizations.
2. **Define Protocol Buffers & gRPC Stubs:** Create Protobuf v3 definitions in `preopen_speedbump.proto` specifying services for `PreOpenAuctionService`, `SpeedBumpGuardService`, and `CollarRealignmentService`. Generate Go server and client stubs using `protoc-gen-go` and `protoc-gen-go-grpc`.
3. **Implement PostgreSQL Database Schema & Migrations:** Write database migration scripts establishing `preopen_batch_manifests`, `preopen_manifest_orders`, `auction_equilibrium_snapshots`, `speedbump_arbitrage_audits`, and `pegged_volatility_bands` with B-Tree and BRIN indexes optimized for time-series queries.
4. **Build Pre-Open Order Aggregator & Manifest Compiler:** Construct the order collector that consumes `order.placement.amo.v1` events from Kafka between 18:00:00 (previous evening) and 08:55:00 IST, groups orders by ISIN and venue (NSE/BSE), validates integrity checksums, and compiles structured binary batch manifests.
5. **Implement DMA Manifest Injection Engine:** Develop the deterministic socket dispatcher connected to NSE NEAT and BSE BOLT DMA gateways. Implement token-bucket pacing to inject manifest batches at exactly 09:00:00 IST up to 09:07:00 IST, enforcing per-socket rate limits (5,000 orders/sec) with microsecond pacing intervals.
6. **Implement Pre-Open Modification & Cancellation Handler:** Build concurrent handling for order modifications and cancellations arriving between 09:00:00 and 09:07:00 IST, routing high-priority cancellation frames to DMA lines ahead of queued batches.
7. **Construct Equilibrium Price Collector (09:08:00 IST):** Create the market feed listener that monitors Level-2/Level-3 exchange broadcasts, detects the conclusion of the order collection phase, extracts official Indicative Equilibrium Prices (IEP) and Indicative Opening Volumes (IOV), and persists snapshots into PostgreSQL and Redis.
8. **Build Dynamic Collar Realignment Engine:** Implement the mathematical collar calculator that evaluates IEP against previous closing prices, calculates expanded/contracted dynamic circuit bands (e.g., +/- 5%, +/- 10%, or dynamic volatility collars), and publishes `auction.equilibrium.realigned.v1` events to Kafka.
9. **Implement Lock-Free Asymmetric Speed Bump Pipeline:** Construct a high-performance lock-free ring-buffer pipeline in Go. Routes incoming aggressive (liquidity-taking) order commands into a 500-microsecond delay buffer using monotonic hardware timestamping, while immediately passing through passive liquidity modifications and cancellations with 0-microsecond delay.
10. **Build Pegged Volatility Band Guard:** Develop the real-time volatility filter that calculates dynamic price bands around the consolidated NBBO midpoint ($P_{\text{mid}} \pm k \cdot \sigma_{\text{rolling}}$). Flags incoming orders arriving during wide-spread dislocations or stale quote conditions, preventing latency arbitrage exploitation.
11. **Implement Colocation Latency Arbitrage Detector:** Build the telemetry collector that measures the delta between exchange outbound timestamps, public feed delivery timestamps, and internal order arrival times. Emits `guard.speedbump.intercepted.v1` events and writes audit records when predatory latency sniping is identified.
12. **Implement gRPC Admin & Telemetry Services:** Expose gRPC management endpoints for operational session monitoring, manual auction injection triggers, dynamic collar adjustments, speed bump parameter tuning (e.g., adjusting delay from 500us to 1000us), and latency profiling.
13. **Integrate Prometheus Metrics & OpenTelemetry Spans:** Instrument the service with Prometheus metric counters and histograms (`preopen_manifest_orders_total`, `dma_injection_latency_micros`, `speedbump_delayed_orders_total`, `speedbump_delay_duration_micros`, `collar_realignment_duration_ms`) and OpenTelemetry distributed tracing context.
14. **Build Unit, Conformance & Simulation Test Suites:** Develop synthetic exchange auction replay fixtures, stress-test fixtures for 500us delay ring buffers, DMA socket disconnect recovery tests, and simulated multi-venue HFT latency arbitrage attack tests.
15. **Execute Microsecond Latency Benchmarks & Stress Tests:** Benchmark the engine under continuous loads of 50,000 orders/sec in the speed bump pipeline and 100,000 orders/sec during pre-open manifest injection, verifying that speed bump delay jitter remains below 5 microseconds and zero memory allocations occur on the critical routing path.

## Interfaces / Contracts

### PostgreSQL Database Schema (DDL)
```sql
-- Pre-Open Batch Manifests Table
CREATE TABLE preopen_batch_manifests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_code VARCHAR(64) NOT NULL UNIQUE,
    exchange_code VARCHAR(8) NOT NULL, -- 'NSE', 'BSE'
    session_date DATE NOT NULL,
    total_orders INT NOT NULL,
    total_quantity NUMERIC(18, 6) NOT NULL,
    total_notional_inr NUMERIC(18, 4) NOT NULL,
    manifest_checksum_sha256 VARCHAR(64) NOT NULL,
    dispatch_status VARCHAR(24) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'DISPATCHING', 'DISPATCHED', 'FAILED', 'PARTIAL_ACK'
    scheduled_dispatch_time TIMESTAMPTZ NOT NULL,
    actual_dispatch_start_time TIMESTAMPTZ,
    actual_dispatch_end_time TIMESTAMPTZ,
    acknowledged_orders INT NOT NULL DEFAULT 0,
    rejected_orders INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pre-Open Individual Manifest Orders Mapping
CREATE TABLE preopen_manifest_orders (
    id BIGSERIAL PRIMARY KEY,
    manifest_id UUID NOT NULL REFERENCES preopen_batch_manifests(id) ON DELETE CASCADE,
    parent_order_id VARCHAR(36) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    trading_symbol VARCHAR(32) NOT NULL,
    side VARCHAR(4) NOT NULL, -- 'BUY', 'SELL'
    order_type VARCHAR(16) NOT NULL, -- 'LIMIT', 'MARKET'
    quantity NUMERIC(18, 6) NOT NULL,
    limit_price NUMERIC(14, 4),
    sequence_in_batch INT NOT NULL,
    exchange_assigned_order_id VARCHAR(64),
    submission_timestamp TIMESTAMPTZ,
    acknowledgement_timestamp TIMESTAMPTZ,
    status VARCHAR(24) NOT NULL DEFAULT 'QUEUED', -- 'QUEUED', 'SUBMITTED', 'ACKNOWLEDGED', 'REJECTED'
    rejection_reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indicative Equilibrium Price (IEP) Snapshots (09:08:00 IST)
CREATE TABLE auction_equilibrium_snapshots (
    id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL,
    trading_symbol VARCHAR(32) NOT NULL,
    exchange_code VARCHAR(8) NOT NULL, -- 'NSE', 'BSE'
    session_date DATE NOT NULL,
    indicative_equilibrium_price NUMERIC(14, 4) NOT NULL,
    indicative_opening_volume BIGINT NOT NULL,
    previous_closing_price NUMERIC(14, 4) NOT NULL,
    price_change_percent NUMERIC(8, 4) NOT NULL,
    total_buy_quantity BIGINT NOT NULL,
    total_sell_quantity BIGINT NOT NULL,
    unmatched_quantity BIGINT NOT NULL,
    auction_clearing_timestamp TIMESTAMPTZ NOT NULL,
    realigned_upper_collar NUMERIC(14, 4) NOT NULL,
    realigned_lower_collar NUMERIC(14, 4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_auction_equilibrium_record UNIQUE (session_date, exchange_code, isin)
);

-- Pegged Volatility Bands & Collar Configuration Table
CREATE TABLE pegged_volatility_bands (
    id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL,
    exchange_code VARCHAR(8) NOT NULL,
    session_date DATE NOT NULL,
    base_reference_price NUMERIC(14, 4) NOT NULL,
    upper_circuit_limit NUMERIC(14, 4) NOT NULL,
    lower_circuit_limit NUMERIC(14, 4) NOT NULL,
    volatility_band_multiplier NUMERIC(6, 4) NOT NULL DEFAULT 1.0000,
    band_expansion_factor NUMERIC(6, 4) NOT NULL DEFAULT 0.0500,
    is_speedbump_active BOOLEAN NOT NULL DEFAULT TRUE,
    speedbump_delay_micros INT NOT NULL DEFAULT 500,
    last_realignment_timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_pegged_volatility_band UNIQUE (session_date, exchange_code, isin)
);

-- Speed Bump Latency Arbitrage Audit Trail
CREATE TABLE speedbump_arbitrage_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(36) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    participant_id VARCHAR(32) NOT NULL,
    side VARCHAR(4) NOT NULL,
    order_type VARCHAR(16) NOT NULL,
    price NUMERIC(14, 4) NOT NULL,
    quantity NUMERIC(18, 6) NOT NULL,
    nbbo_bid_at_ingress NUMERIC(14, 4) NOT NULL,
    nbbo_ask_at_ingress NUMERIC(14, 4) NOT NULL,
    nbbo_midpoint_at_ingress NUMERIC(14, 4) NOT NULL,
    delay_applied_micros INT NOT NULL DEFAULT 500,
    ingress_timestamp_nanos BIGINT NOT NULL,
    egress_timestamp_nanos BIGINT NOT NULL,
    action_taken VARCHAR(24) NOT NULL, -- 'DELAYED_PASSED', 'PRICE_RECALIBRATED', 'SNIPE_INTERCEPTED', 'REJECTED'
    arbitrage_detected BOOLEAN NOT NULL DEFAULT FALSE,
    estimated_adverse_selection_paisa NUMERIC(10, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create High-Performance B-Tree Indexes
CREATE INDEX idx_preopen_manifest_orders_parent ON preopen_manifest_orders(parent_order_id);
CREATE INDEX idx_preopen_manifest_orders_isin ON preopen_manifest_orders(isin, sequence_in_batch);
CREATE INDEX idx_auction_equilibrium_snapshots_lookup ON auction_equilibrium_snapshots(session_date, isin);
CREATE INDEX idx_speedbump_arbitrage_audits_order ON speedbump_arbitrage_audits(order_id);
CREATE INDEX idx_speedbump_arbitrage_audits_time ON speedbump_arbitrage_audits(created_at, isin);
```

### Protocol Buffers Definition (`preopen_speedbump.proto`)
```protobuf
syntax = "proto3";

package growww.preopen.speedbump.v1;

option go_package = "growww/preopen/speedbump/v1;speedbumpv1";

service PreOpenAuctionService {
  rpc CompilePreOpenManifest (CompilePreOpenManifestRequest) returns (CompilePreOpenManifestResponse);
  rpc DispatchPreOpenManifest (DispatchPreOpenManifestRequest) returns (DispatchPreOpenManifestResponse);
  rpc GetManifestStatus (GetManifestStatusRequest) returns (GetManifestStatusResponse);
}

service CollarRealignmentService {
  rpc IngestEquilibriumSnapshot (IngestEquilibriumSnapshotRequest) returns (IngestEquilibriumSnapshotResponse);
  rpc GetDynamicCollars (GetDynamicCollarsRequest) returns (GetDynamicCollarsResponse);
  rpc OverrideDynamicCollar (OverrideDynamicCollarRequest) returns (OverrideDynamicCollarResponse);
}

service SpeedBumpGuardService {
  rpc EvaluateOrderSpeedBump (EvaluateOrderSpeedBumpRequest) returns (EvaluateOrderSpeedBumpResponse);
  rpc UpdateSpeedBumpConfig (UpdateSpeedBumpConfigRequest) returns (UpdateSpeedBumpConfigResponse);
  rpc GetSpeedBumpMetrics (GetSpeedBumpMetricsRequest) returns (GetSpeedBumpMetricsResponse);
}

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BUY = 1;
  ORDER_SIDE_SELL = 2;
}

enum GuardAction {
  GUARD_ACTION_UNSPECIFIED = 0;
  GUARD_ACTION_PASS_IMMEDIATE = 1;
  GUARD_ACTION_DELAY_500_MICROS = 2;
  GUARD_ACTION_RECALIBRATE_PRICE = 3;
  GUARD_ACTION_REJECT_TOXIC_SNIPE = 4;
}

message CompilePreOpenManifestRequest {
  string session_date = 1; // "YYYY-MM-DD"
  string exchange_code = 2; // "NSE" or "BSE"
  repeated string order_ids = 3;
  bool force_recompile = 4;
}

message CompilePreOpenManifestResponse {
  string manifest_id = 1;
  string manifest_code = 2;
  string exchange_code = 3;
  int32 total_orders = 4;
  string total_quantity = 5;
  string total_notional_inr = 6;
  string manifest_checksum_sha256 = 7;
  string compilation_status = 8;
  int64 compiled_at_unix_nanos = 9;
}

message DispatchPreOpenManifestRequest {
  string manifest_id = 1;
  string target_dma_line_id = 2;
  int32 orders_per_second_throttle = 3; // e.g., 5000
}

message DispatchPreOpenManifestResponse {
  string manifest_id = 1;
  string dispatch_status = 2; // "DISPATCHING", "DISPATCHED", "FAILED"
  int64 dispatch_start_unix_nanos = 3;
  int32 dispatched_order_count = 4;
}

message GetManifestStatusRequest {
  string manifest_id = 1;
}

message GetManifestStatusResponse {
  string manifest_id = 1;
  string manifest_code = 2;
  string dispatch_status = 3;
  int32 total_orders = 4;
  int32 acknowledged_orders = 5;
  int32 rejected_orders = 6;
  int64 dispatch_duration_micros = 7;
}

message IngestEquilibriumSnapshotRequest {
  string isin = 1;
  string trading_symbol = 2;
  string exchange_code = 3;
  string session_date = 4;
  string indicative_equilibrium_price = 5;
  int64 indicative_opening_volume = 6;
  string previous_closing_price = 7;
  int64 total_buy_quantity = 8;
  int64 total_sell_quantity = 9;
  int64 auction_clearing_timestamp_nanos = 10;
}

message IngestEquilibriumSnapshotResponse {
  string isin = 1;
  string realigned_upper_collar = 2;
  string realigned_lower_collar = 3;
  string price_shift_percent = 4;
  bool realignment_applied = 5;
}

message GetDynamicCollarsRequest {
  string isin = 1;
  string exchange_code = 2;
}

message GetDynamicCollarsResponse {
  string isin = 1;
  string exchange_code = 2;
  string base_reference_price = 3;
  string dynamic_upper_collar = 4;
  string dynamic_lower_collar = 5;
  string circuit_band_percent = 6;
  bool is_speedbump_enabled = 7;
  int64 last_updated_unix_nanos = 8;
}

message OverrideDynamicCollarRequest {
  string isin = 1;
  string exchange_code = 2;
  string manual_upper_collar = 3;
  string manual_lower_collar = 4;
  string override_reason = 5;
  string authorized_by_admin_id = 6;
}

message OverrideDynamicCollarResponse {
  string isin = 1;
  bool override_applied = 2;
  string active_upper_collar = 3;
  string active_lower_collar = 4;
  int64 updated_at_unix_nanos = 5;
}

message EvaluateOrderSpeedBumpRequest {
  string order_id = 1;
  string isin = 2;
  string participant_id = 3;
  OrderSide side = 4;
  string order_type = 5; // "LIMIT", "MARKET", "IOC"
  string price = 6;
  string quantity = 7;
  bool is_liquidity_taking = 8; // true = taker/aggressive, false = maker/passive
  bool is_cancellation_or_amend = 9;
  int64 ingress_timestamp_nanos = 10;
}

message EvaluateOrderSpeedBumpResponse {
  string order_id = 1;
  GuardAction action_taken = 2;
  int32 delay_applied_micros = 3; // 0 for maker cancel/replace, 500 for taker
  bool is_toxic_snipe_flagged = 4;
  string effective_price = 5;
  int64 egress_timestamp_nanos = 6;
  string guard_decision_reason = 7;
}

message UpdateSpeedBumpConfigRequest {
  string isin = 1;
  bool is_enabled = 2;
  int32 delay_duration_micros = 3; // default 500
  repeated string exempt_participant_ids = 4;
}

message UpdateSpeedBumpConfigResponse {
  string isin = 1;
  bool is_enabled = 2;
  int32 active_delay_micros = 3;
  int64 updated_at_unix_nanos = 4;
}

message GetSpeedBumpMetricsRequest {
  string isin = 1;
  int64 window_duration_seconds = 2;
}

message GetSpeedBumpMetricsResponse {
  string isin = 1;
  int64 total_orders_evaluated = 2;
  int64 orders_delayed_count = 3;
  int64 orders_bypassed_count = 4;
  int64 toxic_snipes_intercepted_count = 5;
  double average_delay_micros = 6;
  string estimated_saved_slippage_inr = 7;
}
```

### Kafka Event Schemas (JSON Schemas)

#### 1. Pre-Open Batch Manifest Dispatched (`auction.preopen.manifest.dispatched.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "PreOpenManifestDispatchedEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "manifestId": { "type": "string", "format": "uuid" },
    "manifestCode": { "type": "string" },
    "exchangeCode": { "type": "string", "enum": ["NSE", "BSE"] },
    "sessionDate": { "type": "string", "format": "date" },
    "totalOrders": { "type": "integer", "minimum": 1 },
    "totalQuantity": { "type": "string" },
    "totalNotionalInr": { "type": "string" },
    "manifestChecksumSha256": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "dmaLineId": { "type": "string" },
    "dispatchStartTimeNanos": { "type": "integer" },
    "dispatchEndTimeNanos": { "type": "integer" },
    "status": { "type": "string", "enum": ["DISPATCHED", "FAILED", "PARTIAL"] }
  },
  "required": [
    "eventId",
    "manifestId",
    "manifestCode",
    "exchangeCode",
    "sessionDate",
    "totalOrders",
    "totalQuantity",
    "totalNotionalInr",
    "manifestChecksumSha256",
    "dispatchStartTimeNanos",
    "status"
  ]
}
```

#### 2. Dynamic Equilibrium Collar Realigned (`auction.equilibrium.realigned.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AuctionEquilibriumRealignedEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "tradingSymbol": { "type": "string" },
    "exchangeCode": { "type": "string", "enum": ["NSE", "BSE"] },
    "sessionDate": { "type": "string", "format": "date" },
    "indicativeEquilibriumPrice": { "type": "string" },
    "indicativeOpeningVolume": { "type": "integer" },
    "previousClosingPrice": { "type": "string" },
    "priceShiftPercent": { "type": "string" },
    "realignedUpperCollar": { "type": "string" },
    "realignedLowerCollar": { "type": "string" },
    "circuitBandPercent": { "type": "string" },
    "equilibriumTimestampNanos": { "type": "integer" },
    "realignmentAppliedAt": { "type": "string", "format": "date-time" }
  },
  "required": [
    "eventId",
    "isin",
    "tradingSymbol",
    "exchangeCode",
    "sessionDate",
    "indicativeEquilibriumPrice",
    "indicativeOpeningVolume",
    "realignedUpperCollar",
    "realignedLowerCollar",
    "equilibriumTimestampNanos",
    "realignmentAppliedAt"
  ]
}
```

#### 3. Speed Bump Intercepted / Toxic Flow Flagged (`guard.speedbump.intercepted.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SpeedBumpInterceptedEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "orderId": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "participantId": { "type": "string" },
    "side": { "type": "string", "enum": ["BUY", "SELL"] },
    "orderType": { "type": "string", "enum": ["LIMIT", "MARKET", "IOC"] },
    "orderPrice": { "type": "string" },
    "orderQuantity": { "type": "string" },
    "nbboMidpointAtIngress": { "type": "string" },
    "delayAppliedMicros": { "type": "integer" },
    "ingressTimestampNanos": { "type": "integer" },
    "egressTimestampNanos": { "type": "integer" },
    "guardAction": { "type": "string", "enum": ["DELAYED_PASSED", "PRICE_RECALIBRATED", "SNIPE_INTERCEPTED", "REJECTED"] },
    "arbitrageDetected": { "type": "boolean" },
    "adverseSelectionPaisa": { "type": "string" },
    "timestamp": { "type": "string", "format": "date-time" }
  },
  "required": [
    "eventId",
    "orderId",
    "isin",
    "participantId",
    "side",
    "orderPrice",
    "orderQuantity",
    "delayAppliedMicros",
    "ingressTimestampNanos",
    "guardAction",
    "arbitrageDetected",
    "timestamp"
  ]
}
```

## Security & Compliance Notes
- **Direct Market Access (DMA) Pacing & SEBI Throttling Compliance:** SEBI algorithmic trading guidelines mandate strict throttling across Direct Market Access connections to prevent infrastructure saturation. The manifest injector strictly meters outbound socket writes to a maximum of 5,000 orders/sec per exchange DMA line, utilizing token-bucket rate controllers with microsecond pacing.
- **Asymmetric Latency Fairness & Non-Discriminatory Access:** The 500-microsecond speed bump operates deterministically across all incoming non-display aggressive orders. Market makers and retail liquidity providers receive zero-delay priority solely for quote cancellations and price amendments to protect resting passive depth against latency arbitrage, conforming to fair market access standards.
- **Pre-Open Batch Integrity & Checksum Verification:** Prior to socket injection at 09:00:00 IST, batch manifests are sealed with SHA-256 cryptographic hashes. Any in-flight tampering, malformed packet structure, or sequence discrepancy aborts the dispatch sequence, triggering an immediate security alert to the market surveillance team.
- **Dynamic Circuit Collar Compliance:** The equilibrium realignment engine strictly enforces SEBI-mandated circuit filters (e.g., +/- 5%, +/- 10%, or +/- 20% bands) adjusted from the 09:08:00 IST Indicative Equilibrium Price. Orders crossing dynamic collar thresholds are rejected immediately at the gateway ingress.
- **Hardware-Isolated Signing & DMA Credential Protection:** All DMA session keys, encryption certificates, and exchange login credentials are stored within AWS CloudHSM or HashiCorp Vault memory enclaves, never written to disk or logged in plain text.
- **Zero PII Exposure Guarantee:** Network sockets, gRPC interfaces, Kafka topics, and blockchain notarization payloads process strictly pseudonymized participant IDs, order UUIDs, ISIN codes, prices, and quantities. No personally identifiable investor data (PAN, Aadhaar, investor name, bank details) is ever exposed on the execution path.

## Acceptance Criteria
- [ ] Pre-Open order collector aggregates and compiles all queued After-Market Orders (AMO) into verified batch manifests with SHA-256 checksums prior to 08:58:00 IST.
- [ ] DMA manifest injector streams pre-open batches to NSE and BSE gateways between 09:00:00 and 09:07:00 IST at a paced rate of 5,000 orders/sec per line with zero gateway throttling errors.
- [ ] Pre-open order cancellations and modifications arriving during 09:00:00-09:07:00 IST are prioritized ahead of queued batch injection frames within $< 100$ microseconds.
- [ ] Indicative Equilibrium Price (IEP) and Indicative Opening Volume (IOV) are extracted from exchange Level-2 feeds at 09:08:00 IST and persisted within $< 5$ milliseconds of feed broadcast.
- [ ] Dynamic equilibrium collars are recalculated and published to the Order Matching Engine and SOR within $< 10$ milliseconds of IEP capture.
- [ ] Asymmetric speed bump applies a precise 500-microsecond delay (+/- 5us jitter) to aggressive taker orders while permitting maker cancellations and quote amendments with 0-microsecond delay.
- [ ] Pegged volatility band guard successfully identifies and flags simulated toxic latency snipe orders arriving during stale quote intervals.
- [ ] PostgreSQL database schema and indexes handle 1,000,000+ daily pre-open and speed bump audit rows with sub-10ms query times for audit retrieval.
- [ ] Service sustains 50,000 orders/sec concurrent throughput in the speed bump pipeline with p99 internal processing latency under 30 microseconds (excluding the intentional 500us delay).
- [ ] Zero memory allocations occur on the critical lock-free ring-buffer execution path.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (System Architecture Overview), 103 (API Standards), 104 (Kafka Topic Standards), 111 (Core Domain Models), 204 (Order Service), 242 (NSE/BSE Market Data & SOR Adapter), 401 (PostgreSQL Master Schema), 402 (Redis Patterns).
- **Parallel Tasks:** 205 (Order Matching Engine), 206 (Risk & Margin Checks Service), 225 (FIX Protocol Gateway), 228 (Real-Time Market Surveillance Engine).
- **Downstream Blockers:** 208 (Trade Settlement Service), 215 (Reconciliation Service), 230 (Settlement Guarantee Fund), 912 (Crosschain Bridge & Derivatives Stress Testing Plan).
