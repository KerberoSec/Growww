# 408 - Historical Market Data Pipeline & Time-Series Analytics (TimescaleDB & ClickHouse)

## Purpose
Indian equity, currency, and commodity markets across the National Stock Exchange (NSE), Bombay Stock Exchange (BSE), and Multi Commodity Exchange (MCX) generate massive streams of trade executions and price ticks during every trading session. A modern financial platform must reliably capture, validate, aggregate, and store this high-frequency market data to power interactive technical charts, quantitative research, risk analytics, and statutory regulatory compliance.

The Historical Market Data Pipeline and Time-Series Analytics Service (`services/market-data-analytics`) is the high-throughput analytical backbone for market time series. It establishes a specialized dual-datastore architecture combining TimescaleDB and ClickHouse. TimescaleDB serves high-concurrency, low-latency operational queries for recent intraday candles (1s, 1m, 5m, 15m, 1h, 1d) powering TradingView Lightweight Charts on Flutter and Web. Simultaneously, ClickHouse serves petabyte-scale tick archival, multi-year backtesting queries, and official exchange Bhavcopy daily settlement record preservation.

## What You Are Building
A production-grade, distributed market data ingestion and time-series analytical engine containing:
- **Market Data Analytics Microservice (`services/market-data-analytics`):** A high-performance Go 1.22+ service exposing gRPC, REST, and WebSocket interfaces optimized for TradingView charting components.
- **High-Frequency Kafka Ingestion Workers:** Resilient consumer groups ingesting continuous tick feeds (`market.ticks.nse`, `market.ticks.bse`, `market.ticks.mcx`, `trading.trades.executed`) with batch writing into time-series storage.
- **TimescaleDB Time-Series Engine:** Optimized hypertables with native continuous aggregates, automated rollups (1s to 1m, 5m, 15m, 1h, 1d), chunk compression policies, and data retention windows.
- **ClickHouse Multi-Year Data Lake:** Columnar storage cluster utilizing `MergeTree`, `AggregatingMergeTree`, and `ReplacingMergeTree` engines for tick-by-tick archival and cross-sectional market analytics with tiered S3 cold storage.
- **Daily Bhavcopy Ingestion & Archival Engine:** Automated scheduled workers downloading, validating, deduplicating, and indexing official end-of-day Bhavcopy files from NSE, BSE, and MCX.
- **TradingView UDF-Compatible REST & WebSocket API:** Low-latency REST endpoints (`/v1/market/history`, `/v1/market/bhavcopy`) and WebSocket bar subscription streams delivering sub-50ms candlestick queries to client apps.
- **In-Memory Cache Layer:** Redis 7.2 caching active trading day candle windows to offload operational database queries during peak market volatility.

## Scope Boundaries
- **In Scope:**
  - Ingestion of high-frequency market ticks and executed trade streams from Kafka topics.
  - TimescaleDB hypertable DDLs, continuous aggregate policies, and chunk compression settings.
  - ClickHouse columnar tables, Kafka engine tables, and continuous materialized view pipelines.
  - Multi-timeframe OHLCV bar calculation (1s, 1m, 5m, 15m, 1h, 1d) with volume, VWAP, and trade count.
  - Daily Bhavcopy ETL jobs for NSE, BSE, and MCX with schema normalization and checksum verification.
  - Low-latency query serving API (gRPC, REST UDF, WebSocket streaming) for Flutter (Prompt 508) and Web (Prompt 603).
  - Tiered storage lifecycle policies migrating data from hot NVMe to cold encrypted S3 object storage.
- **Out of Scope / Handled Elsewhere:**
  - Real-time sub-millisecond Level-2 order book depth aggregation and raw tick broadcasting (handled in Prompt 207).
  - Direct exchange binary line-rate feed handlers and FIX/FAST protocol decoders (handled in Prompt 225 and Prompt 242).
  - Canonical security definitions, ISIN master registry, and corporate actions tracking (handled in Prompt 407).
  - Client-side chart rendering and UI gesture handling in Flutter (Prompt 508) or Web (Prompt 603).
  - Transactional relational database schemas for users, balances, and orders (handled in Prompt 401).

## Technology to Use
The analytical pipeline employs a hybrid time-series and columnar architecture to balance sub-50ms chart query response times with multi-year cost-efficient storage.

- **Primary Service Language:** Go 1.22+. Go provides superior concurrency characteristics, minimal garbage collection overhead, and high-throughput network I/O for streaming market data.
- **Operational Time-Series Datastore:** PostgreSQL 16 with TimescaleDB 2.14+ extension. TimescaleDB hypertables provide automatic time-based partitioning, real-time continuous aggregates with watermark tracking, and 90%+ columnar compression on completed chunks.
- **Historical OLAP Columnar Datastore:** ClickHouse 24.x with ClickHouse Keeper. ClickHouse provides vector execution, rapid parallel aggregations across billions of historical ticks, and native integration with S3-backed storage tiers.
- **Streaming Message Broker:** Apache Kafka 3.7+ with snappy compression, key-partitioned by ISIN and market symbol.
- **In-Memory Cache:** Redis 7.2 Cluster for caching hot intraday candle sets and symbol session stats.
- **Storage Tiering:** Local NVMe SSDs (`gp3`/`io2`) for hot data partitions (0-90 days), transitioning automatically to AWS S3 (`ap-south-1`) for long-term historical records (up to 8 years).
- **Communication Protocols:** gRPC / Protobuf v3 for internal microservice RPCs; HTTP/2 REST (UDF compatible) and WebSockets for client applications.

## Backend / Infra Touchpoints
- **Apache Kafka:** Topics `market.ticks.nse`, `market.ticks.bse`, `market.ticks.mcx`, `trading.trades.executed`.
- **TimescaleDB Cluster:** Hypertables `market_ticks`, `ohlcv_1s`, `ohlcv_1m`, `ohlcv_5m`, `ohlcv_15m`, `ohlcv_1h`, `ohlcv_1d`.
- **ClickHouse Cluster:** Tables `analytics.market_ticks_historical`, `analytics.bhavcopy_daily`, `analytics.mv_ohlcv_1m`.
- **AWS S3 / MinIO:** Cold storage tier for ClickHouse partition data and immutable raw exchange Bhavcopy file archives.
- **Redis Cluster:** Key namespace `market:candles:{isin}:{timeframe}` for fast lookups.
- **Client Interfaces:** Flutter Mobile/Desktop (Prompt 508) and Next.js Web (Prompt 603) via API Gateway (Prompt 219).

## Blockchain Interaction
While market data analytics executes off-chain to achieve sub-second latency, it directly interfaces with the permissioned Hyperledger Besu blockchain network (QBFT consensus, 1:1 asset backing, zero PII) in several critical operational areas:

### Detailed On-Chain Integration Mechanics:
- **Historical Valuation Attestation:** The pipeline provides official Volume Weighted Average Price (VWAP) and daily closing prices used by `ProofOfReserveRegistry.sol` to record daily mark-to-market valuations for tokenized equities.
- **DvP Execution Verification:** During trade dispute audits, historical tick sequences from ClickHouse are correlated with on-chain settlement transactions (`SettlementDvP.sol`) to verify execution price validity against prevailing exchange circuit bands.
- **Corporate Action Price Adjustments:** Historical price series are adjusted based on corporate action records committed to on-chain governance contracts (`CorporateActionsRegistry.sol`), maintaining continuous adjusted historical charts.
- **Zero PII Exposure:** All stored time-series data contains purely market-level quotes, timestamps, volumes, and instrument identifiers (ISINs). No investor account numbers, PANs, or personal identities are present in the time-series datastores.

## Step-by-Step Build Instructions
1. Scaffold project repository: `services/market-data-analytics/` containing `cmd/server/`, `internal/ingest/`, `internal/timescale/`, `internal/clickhouse/`, `internal/bhavcopy/`, `internal/api/`, `proto/`, and `migrations/`.
2. Configure TimescaleDB PostgreSQL migrations creating the primary `market_ticks` hypertable partitioned by 1-day time chunks.
3. Define TimescaleDB continuous aggregate views for multi-timeframe candles: `ohlcv_1s`, `ohlcv_1m`, `ohlcv_5m`, `ohlcv_15m`, `ohlcv_1h`, `ohlcv_1d`.
4. Configure TimescaleDB continuous aggregate refresh policies and chunk compression policies (compress chunks older than 3 days; retain uncompressed hot data for intraday writes).
5. Implement ClickHouse DDL schemas for raw tick historical archive (`analytics.market_ticks_historical`), daily Bhavcopy archive (`analytics.bhavcopy_daily`), and Kafka engine ingestion tables.
6. Configure ClickHouse storage policies defining automatic tiering from local NVMe storage to encrypted AWS S3 buckets for partitions older than 90 days.
7. Build high-throughput Kafka ingestion workers in Go utilizing batch writes (10,000 ticks or 100ms flush intervals) to stream incoming market data into TimescaleDB and ClickHouse.
8. Implement automated daily Bhavcopy ETL worker scheduled to download, parse, validate, and persist official end-of-day market summaries from NSE, BSE, and MCX portals.
9. Implement Redis caching layer to store active-day completed 1-minute and 5-minute candles with 24-hour TTLs to minimize database load.
10. Define Protobuf definitions (`market_data_analytics.proto`) and implement gRPC server providing historical bar extraction and Bhavcopy lookups.
11. Implement TradingView UDF (Universal Data Feed) compatible REST endpoints (`/v1/market/history`, `/v1/market/symbols`, `/v1/market/time`) returning formatted JSON candle arrays (`t`, `o`, `h`, `l`, `c`, `v`).
12. Implement WebSocket streaming hub delivering real-time newly closed candle bars and in-progress live bar updates to connected Flutter and Web clients.
13. Implement comprehensive unit and integration tests verifying tick-to-candle math, continuous aggregate rollups, Bhavcopy checksum validation, and out-of-order tick handling.
14. Benchmark query performance under realistic load: verify historical 1-minute candle queries over 30 days return in under 30ms for 1,000 concurrent requests.

## Interfaces / Contracts

### 1. Protocol Buffers: `market_data_analytics.proto`
```protobuf
syntax = "proto3";

package growww.marketdata.analytics.v1;

option go_package = "github.com/growww/nbse/services/market-data-analytics/proto/v1;analyticsv1";

service MarketDataAnalyticsService {
  rpc GetHistoricalCandles (GetHistoricalCandlesRequest) returns (GetHistoricalCandlesResponse);
  rpc GetBhavcopyRecord (GetBhavcopyRecordRequest) returns (GetBhavcopyRecordResponse);
  rpc GetDailySummary (GetDailySummaryRequest) returns (GetDailySummaryResponse);
  rpc StreamCandles (StreamCandlesRequest) returns (stream CandleUpdate);
}

enum Timeframe {
  TIMEFRAME_UNSPECIFIED = 0;
  TIMEFRAME_1S = 1;
  TIMEFRAME_1M = 2;
  TIMEFRAME_5M = 3;
  TIMEFRAME_15M = 4;
  TIMEFRAME_1H = 5;
  TIMEFRAME_1D = 6;
  TIMEFRAME_1W = 7;
}

message GetHistoricalCandlesRequest {
  string isin = 1;
  string exchange = 2; // "NSE", "BSE", "MCX"
  Timeframe timeframe = 3;
  int64 from_timestamp = 4; // Unix epoch seconds
  int64 to_timestamp = 5;   // Unix epoch seconds
  int32 limit = 6;          // Max number of bars (default 500, max 5000)
}

message CandleBar {
  int64 timestamp = 1;      // Bucket start timestamp (Unix epoch seconds)
  string open = 2;          // Decimal formatted string (e.g. "2450.50")
  string high = 3;
  string low = 4;
  string close = 5;
  string volume = 6;        // Total traded volume in fractional units
  string vwap = 7;          // Volume Weighted Average Price
  int64 trade_count = 8;    // Number of executed trades in bucket
  string turnover = 9;      // Total traded turnover value in INR
}

message GetHistoricalCandlesResponse {
  string isin = 1;
  string symbol = 2;
  Timeframe timeframe = 3;
  repeated CandleBar bars = 4;
}

message GetBhavcopyRecordRequest {
  string isin = 1;
  string trade_date = 2;    // Format: "YYYY-MM-DD"
  string exchange = 3;
}

message GetBhavcopyRecordResponse {
  string isin = 1;
  string symbol = 2;
  string trade_date = 3;
  string exchange = 4;
  string open_price = 5;
  string high_price = 6;
  string low_price = 7;
  string close_price = 8;
  string last_price = 9;
  string prev_close = 10;
  string total_traded_qty = 11;
  string total_traded_val = 12;
  int64 total_trades = 13;
  string delivery_qty = 14;
  string delivery_percent = 15;
  string open_interest = 16;
}

message GetDailySummaryRequest {
  string isin = 1;
  string exchange = 2;
}

message GetDailySummaryResponse {
  string isin = 1;
  string symbol = 2;
  string open = 3;
  string high = 4;
  string low = 5;
  string close = 6;
  string ltp = 7;
  string volume = 8;
  string turnover = 9;
  string vwap = 10;
  string change = 11;
  string change_percent = 12;
  int64 timestamp = 13;
}

message StreamCandlesRequest {
  string isin = 1;
  Timeframe timeframe = 2;
}

message CandleUpdate {
  string isin = 1;
  Timeframe timeframe = 2;
  CandleBar bar = 3;
  bool is_closed = 4;
}
```

### 2. TimescaleDB Schema & Continuous Aggregates DDL
```sql
-- TimescaleDB Extensions & Base Table
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE SCHEMA IF NOT EXISTS market_data;

CREATE TABLE market_data.market_ticks (
    time TIMESTAMPTZ NOT NULL,
    isin CHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    exchange VARCHAR(8) NOT NULL,
    price NUMERIC(18, 4) NOT NULL,
    volume NUMERIC(18, 6) NOT NULL,
    buyer_order_id UUID,
    seller_order_id UUID,
    trade_id UUID NOT NULL
);

-- Convert to Hypertable partitioned by 1 day chunks
SELECT create_hypertable('market_data.market_ticks', 'time', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_market_ticks_isin_time ON market_data.market_ticks (isin, time DESC);

-- 1-Minute Continuous Aggregate View
CREATE MATERIALIZED VIEW market_data.ohlcv_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    isin,
    exchange,
    FIRST(price, time) AS open,
    MAX(price) AS high,
    MIN(price) AS low,
    LAST(price, time) AS close,
    SUM(volume) AS volume,
    SUM(price * volume) / NULLIF(SUM(volume), 0) AS vwap,
    COUNT(*) AS trade_count,
    SUM(price * volume) AS turnover
FROM market_data.market_ticks
GROUP BY bucket, isin, exchange
WITH NO DATA;

-- 1-Hour Continuous Aggregate View (Rolled up from 1-Minute View)
CREATE MATERIALIZED VIEW market_data.ohlcv_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    isin,
    exchange,
    FIRST(open, bucket) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, bucket) AS close,
    SUM(volume) AS volume,
    SUM(turnover) / NULLIF(SUM(volume), 0) AS vwap,
    SUM(trade_count) AS trade_count,
    SUM(turnover) AS turnover
FROM market_data.ohlcv_1m
GROUP BY time_bucket('1 hour', bucket), isin, exchange
WITH NO DATA;

-- Continuous Aggregate Refresh Policies
SELECT add_continuous_aggregate_policy('market_data.ohlcv_1m',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

SELECT add_continuous_aggregate_policy('market_data.ohlcv_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- Hypertable Compression Policy (Compress raw ticks older than 3 days)
ALTER TABLE market_data.market_ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'isin, exchange',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('market_data.market_ticks', INTERVAL '3 days');

-- Retention Policy (Drop raw ticks from TimescaleDB after 90 days; long-term kept in ClickHouse)
SELECT add_retention_policy('market_data.market_ticks', INTERVAL '90 days');
```

### 3. ClickHouse Historical Analytical DDL
```sql
CREATE DATABASE IF NOT EXISTS analytics;

-- Long-Term Tick Archive Table (Multi-Year)
CREATE TABLE analytics.market_ticks_historical (
    trade_date Date,
    executed_at DateTime64(3, 'Asia/Kolkata'),
    isin LowCardinality(String),
    symbol LowCardinality(String),
    exchange LowCardinality(String),
    price Decimal64(4),
    volume Decimal64(6),
    turnover Decimal64(4),
    buyer_order_id UUID,
    seller_order_id UUID,
    trade_id UUID
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(trade_date)
ORDER BY (isin, exchange, executed_at, trade_id)
SETTINGS index_granularity = 8192;

-- Daily Bhavcopy Archive Table
CREATE TABLE analytics.bhavcopy_daily (
    trade_date Date,
    exchange LowCardinality(String),
    isin LowCardinality(String),
    symbol LowCardinality(String),
    series LowCardinality(String),
    open_price Decimal64(4),
    high_price Decimal64(4),
    low_price Decimal64(4),
    close_price Decimal64(4),
    last_price Decimal64(4),
    prev_close Decimal64(4),
    total_traded_qty Decimal64(6),
    total_traded_val Decimal64(4),
    total_trades UInt64,
    delivery_qty Decimal64(6),
    delivery_percent Decimal64(2),
    open_interest UInt64,
    ingested_at DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(ingested_at)
PARTITION BY toYYYYMM(trade_date)
ORDER BY (trade_date, exchange, isin);
```

### 4. REST TradingView UDF Endpoint Contract
```json
// GET /v1/market/history?symbol=INE002A01018&resolution=1&from=1726650000&to=1726653600
// Response (TradingView UDF Format):
{
  "s": "ok",
  "t": [1726650000, 1726650060, 1726650120, 1726650180],
  "o": [2450.50, 2451.00, 2449.80, 2452.10],
  "h": [2452.00, 2453.50, 2451.20, 2455.00],
  "l": [2449.00, 2450.00, 2448.50, 2451.50],
  "c": [2451.00, 2449.80, 2452.10, 2454.20],
  "v": [1240.500000, 980.200000, 1530.000000, 2100.800000]
}
```

## Security & Compliance Notes
- **SEBI 8-Year Data Retention Compliance:** In accordance with SEBI regulations for stock brokers and depositories, all historical market ticks and daily Bhavcopy records are immutably archived for a minimum of 8 years using Write Once Read Many (WORM) storage policies on AWS S3.
- **Data Integrity & Checksum Verification:** Automated Bhavcopy download jobs compute SHA-256 checksums on official exchange files before ingestion. Any checksum mismatch or record count disparity immediately triggers a compliance alert and halts ingestion.
- **Strict Data Residency:** All TimescaleDB instances, ClickHouse nodes, Redis clusters, and S3 backup buckets are hosted exclusively in the AWS Asia Pacific (Mumbai) region (`ap-south-1`), complying with Indian data localization requirements.
- **Rate Limiting & Denial of Service Protection:** The public-facing REST and WebSocket query endpoints enforce granular token-bucket rate limits per client IP and authenticated session to prevent scraping or denial-of-service degradation during market peaks.
- **Zero PII Storage:** The market data analytics datastores store purely market trade data and security identifiers. No customer Personally Identifiable Information (PII) is captured or retained in these systems.

## Acceptance Criteria
- [ ] TimescaleDB hypertable `market_data.market_ticks` and continuous aggregate views (`ohlcv_1m`, `ohlcv_1h`) created and verified with test tick streams.
- [ ] Continuous aggregate policies automatically refresh rolling candlestick windows with no manual intervention.
- [ ] TimescaleDB chunk compression policy achieves >85% storage reduction on raw tick chunks older than 3 days.
- [ ] ClickHouse 24.x tables `analytics.market_ticks_historical` and `analytics.bhavcopy_daily` deployed with S3 tiering policies.
- [ ] High-throughput Go Kafka ingestion worker consumes ticks and executes batch inserts at >50,000 ticks/sec with zero data loss.
- [ ] Automated Bhavcopy ingestion pipeline successfully downloads, validates, and stores daily official files for NSE, BSE, and MCX.
- [ ] gRPC `MarketDataAnalyticsService` server implemented and tested against all RPC methods.
- [ ] REST API provides TradingView UDF-compliant candlestick history (`/v1/market/history`) with response latency < 30ms for 30-day 1-minute ranges.
- [ ] WebSocket streaming hub broadcasts real-time candle bar updates to subscribed clients within 50ms of bar close.
- [ ] Redis caching layer successfully reduces database query volume by >80% during simulated active trading market hours.
- [ ] SHA-256 checksum verification prevents ingestion of corrupted or tampered Bhavcopy files.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 104 (Kafka Standards), Prompt 401 (PostgreSQL Schema), Prompt 403 (Kafka Cluster), Prompt 404 (ClickHouse Architecture), Prompt 407 (Master Data Management).
- **Parallel Tasks:** Prompt 207 (Real-Time Market Data Service), Prompt 242 (NSE/BSE Market Data Adapter), Prompt 243 (MCX Commodity Adapter).
- **Downstream Consumers:** Prompt 508 (Flutter Security Detail & Charts), Prompt 603 (Web Trading & Charting Interface), Prompt 215 (Reconciliation Service), Prompt 216 (Regulatory Reporting Service).
