# 272 - BTC/USDT Real-Time External Market Data Feeder & Aggregator (Go)

## Purpose
The NBSE platform operates continuous, highly reliable institutional and retail trading infrastructure across both domestic and international financial jurisdictions (such as GIFT City under IFSCA). To support spot, fractional, and synthetic digital asset trading-alongside risk simulation in both Testnet (Demo) and Mainnet production environments-the platform requires an ultra-low-latency, tamper-resistant, and high-frequency external market data feed for BTC/USDT.

Relying on a single external cryptocurrency exchange exposes trading systems to catastrophic failure modes: exchange API outages, regional network routing disruptions, localized flash crashes, artificial spoofing, and malicious wash trading. A localized anomaly on an isolated exchange must never be permitted to trigger cascading liquidations, distorted index pricing, or unfair fills on NBSE.

The **BTC/USDT Real-Time External Market Data Feeder & Aggregator** (`services/market-feeder`) is a mission-critical Go microservice engineered to:
1. Ingest redundant, high-frequency real-time WebSocket trade feeds from top-tier spot venues: Binance, Coinbase, Kraken, and OKX.
2. Interrogate the on-chain Chainlink BTC/USD Data Feed oracle deployed on the permissioned Hyperledger Besu ledger to establish an immutable, manipulation-resistant reference anchor price.
3. Apply a mathematical Outlier Price Rejection engine based on a dynamic Trimmed Median and Median Absolute Deviation (MAD) algorithm to filter corrupted, stale, or manipulative ticks.
4. Calculate a unified, volume-weighted Fair Index Price and consolidated ticker updated in sub-millisecond intervals.
5. Compute multi-timeframe rolling OHLCV (Open, High, Low, Close, Volume) candlestick bars (1s, 1m, 5m, 15m, 1h, 1d) with zero tick loss and deterministic time-boundary rollover.
6. Disseminate standardized Protobuf-encoded ticks, tickers, and candlestick updates to internal platform consumers with sub-5ms fanout over Apache Kafka and Redis Pub/Sub.
7. Seamlessly feed both the live production Matching Engine (Prompt 205) and the Demo Trading Engine (Prompt 273), enabling realistic paper trading with real-time live market dynamics.

---

## What You Are Building
A standalone, high-performance Go 1.22+ streaming service located in the repository at `services/market-feeder`. Core architectural modules include:

- **Multi-Exchange WebSocket Ingestion Adapters:** Resilient, concurrent WebSocket client workers maintaining dedicated persistent connections to:
  - *Binance:* `<wss://stream.binance.com:9443/ws/btcusdt@trade>`
  - *Coinbase Exchange:* `<wss://ws-feed.exchange.coinbase.com>` (`matches` channel for `BTC-USDT` / `BTC-USD`)
  - *Kraken:* `<wss://ws.kraken.com/v2>` (`trade` channel for `BTC/USDT` / `BTC/USD`)
  - *OKX:* `<wss://ws.okx.com:8443/ws/v5/public>` (`trades` channel for `BTC-USDT`)
- **Resilient Connection Supervisor:** Manages independent WebSocket lifecycles, active ping/pong heartbeats, socket read deadlines, TCP keep-alives, and automatic reconnection using exponential backoff with decorrelated full jitter.
- **Chainlink On-Chain Oracle Monitor:** Periodically queries the `AggregatorV3Interface` contract on Hyperledger Besu via high-speed JSON-RPC/IPC to fetch the canonical BTC/USD anchor price, verifying round timestamps, answer ranges, and heartbeat freshness.
- **Outlier Rejection & Fair Index Engine:** Normalizes incoming ticks into standardized fixed-point decimal structures (`shopspring/decimal`), discards ticks violating exchange sanity bands, and executes a real-time trimmed median algorithm across all active exchange feeds anchored against the Chainlink oracle reference price.
- **Real-Time OHLCV Candlestick Aggregator:** In-memory sliding-window bucket engine maintaining continuous multi-resolution candles (1-second, 1-minute, 5-minute, 15-minute, 1-hour, 1-day), updating partial bars on every valid tick and sealing completed intervals with millisecond precision.
- **Ultra-Low-Latency Message Fanout Layer:** Dispatches consolidated tickers, raw normalized ticks, and candlestick updates via dual channels:
  - *Kafka Topics:* High-durability pub/sub for downstream analytical services, historical ingestion, and the Demo Trading Engine.
  - *Redis Pub/Sub:* Sub-millisecond in-memory distribution for the Market Data Service (Prompt 207) and Frontend WebSocket Gateways.
- **TimescaleDB Historical Archival Worker:** Asynchronously persists sealed 1-minute and 1-day candlestick records and high-resolution tick aggregates to PostgreSQL/TimescaleDB hypertable storage.
- **gRPC Inspection & Control Server:** Implements `MarketFeederService` to provide instant snapshots of active exchange feed health, latencies, current weights, and manual administrative overrides.

---

## Scope Boundaries

### In Scope
- Establishing and supervising persistent TLS WebSocket connections to Binance, Coinbase, Kraken, and OKX.
- Parsing, validating, and normalizing disparate exchange wire formats into a canonical internal tick representation.
- In-memory outlier detection, statistical filtering (trimmed median, MAD), and composite fair index price derivation.
- Querying the permissioned Besu network for Chainlink BTC/USD reference rounds to guard against external exchange flash loans or coordinated manipulation.
- Building sliding-window OHLCV candles across standard intervals (1s, 1m, 5m, 15m, 1h, 1d).
- Sub-5ms publication of Protobuf-serialized ticker and candle events to Kafka and Redis Pub/Sub.
- TimescaleDB batch persistence for continuous historical charting data.
- Dual-environment mode dispatching live feeds to both Mainnet infrastructure and the Demo Trading Engine (Prompt 273).
- Structured Prometheus metrics, health probes, and feed degradation alerting.

### Out of Scope / Handled Elsewhere
- End-user WebSocket client connection termination and authentication (handled in Real-Time Market Data Service, Prompt 207, and API Gateway, Prompt 219).
- Matching retail/institutional BTC/USDT orders against internal order books (handled in Order Matching Engine, Prompt 205).
- Simulated balance tracking, virtual fills, and paper trading portfolios (handled in Demo Trading Engine, Prompt 273).
- User interface chart rendering in Flutter or React (handled in Prompt 508 and Prompt 603).
- Custodial vault cold storage and Bitcoin blockchain UTXO ingress (handled in Bitcoin Lightning & Taproot Ingress Service, Prompt 234, and MPC Vault Custody, Prompt 237).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Go provides superior low-overhead concurrency via goroutines and channels, deterministic garbage collection latency (sub-millisecond GC pauses), and efficient binary serialization.
- **WebSocket Client:** `github.com/gorilla/websocket` with dialed TLS configurations, zero-copy read buffers, and custom TCP keep-alive dialers.
- **Fixed-Point Financial Mathematics:** `github.com/shopspring/decimal` to eliminate floating-point imprecision across price, volume, and index calculations, maintaining 18 decimal places of internal precision.
- **In-Memory Messaging & Caching:** Redis 7.2+ Cluster via `github.com/redis/go-redis/v9` for sub-millisecond Pub/Sub broadcast and sliding tick cache.
- **Event Streaming:** Apache Kafka 3.7+ via `github.com/segmentio/kafka-go`, utilizing snappy compression, batching linger of 2ms, and partition hashing on trading symbol (`BTCUSDT`).
- **Time-Series Storage:** PostgreSQL 16+ with TimescaleDB extension, utilizing `jackc/pgx/v5` connection pooling and hypertable partitioning on time intervals.
- **Blockchain Client:** `github.com/ethereum/go-ethereum/ethclient` communicating with Hyperledger Besu nodes over mTLS-secured JSON-RPC/WebSocket.
- **Serialization & RPC:** Protocol Buffers v3 (`google.golang.org/protobuf`) and gRPC (`google.golang.org/grpc`) for strongly typed, minimal-overhead payloads.
- **Observability:** Prometheus metrics client (`prometheus/client_golang`) and OpenTelemetry Go SDK (`go.opentelemetry.io/otel`).

---

## Backend / Infra Touchpoints
- **Real-Time Market Data Service (Prompt 207):** Ingests the unified fair index price and aggregated candlestick streams from Redis Pub/Sub channels (`market:feed:ticker:BTCUSDT`, `market:feed:candles:BTCUSDT:1m`) to broadcast updates to hundreds of thousands of retail Flutter and web clients.
- **Order Matching Engine (Prompt 205):** Ingests the index price feed via Kafka (`market.external.btc_usdt.ticker.v1`) to compute dynamic circuit breakers, maximum order price bands, and initial margin validations.
- **Demo Trading Engine (Prompt 273):** Directly consumes the high-frequency tick stream from Kafka (`market.external.btc_usdt.ticks.v1`) to drive paper trading fills, limit order matches, and realistic slippage simulation for retail onboarding.
- **Frontend WebSocket Gateway (Prompt 219):** Receives ticker and L1 depth broadcasts for instant quote updates in client applications.
- **Redis Cluster:** Relays in-memory pub/sub topics across service pods and caches the last 1,000 raw ticks for instant bootstrap queries.
- **TimescaleDB Cluster:** Stores hypertable partitions `external_btc_ticks` and `external_btc_candles` with continuous retention policies.

---

## Blockchain Interaction (Permissioned Hyperledger Besu Ledger with 1:1 Custody Backing, Zero PII, QBFT)
- **Consortium Network Environment:** Hyperledger Besu private permissioned network running Istanbul Byzantine Fault Tolerant (QBFT) consensus with 1-second block finality.
- **Chainlink Oracle Contract Interrogation:**
  - The feeder connects to the verified on-chain Chainlink `AggregatorV3Interface` contract deployed on Besu:
    ```solidity
    interface AggregatorV3Interface {
        function latestRoundData() external view returns (
            uint80 roundId,
            int256 answer,
            uint256 startedAt,
            uint256 updatedAt,
            uint80 answeredInRound
        );
    }
    ```
  - Queries `latestRoundData()` at a configurable frequency (e.g., every 5 seconds) to obtain the canonical BTC/USD price anchor (scaled to 8 decimal places).
- **Oracle Sanity Validation:**
  - *Heartbeat Check:* Asserts that `block.timestamp - updatedAt <= 120 seconds`. If the oracle has not updated within its maximum heartbeat, the feeder flags the oracle as stale and relies temporarily on the multi-exchange trimmed median.
  - *Completeness Check:* Asserts `answeredInRound >= roundId` and `answer > 0`.
  - *Anchor Corridor:* Compares the live exchange median against `answer`. If the off-chain composite deviates from the Chainlink anchor by more than a configurable corridor (e.g., $\pm 3.0\%$), an immediate critical alert is emitted, and circuit breaker dampening is applied to prevent malicious flash crashes from propagating to the matching engine.
- **Zero-PII Ledger Invariant:** The feeder performs read-only contract calls against the public oracle state. It produces no user-identifying transactions or state changes on the ledger. All market data feeds remain entirely anonymous and public.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Layout:** Initialize the Go module `services/market-feeder` with a clean architecture directory structure:
   - `cmd/feeder/main.go`: Entrypoint, configuration loader, service lifecycle.
   - `internal/adapters/`: Exchange-specific WebSocket clients (Binance, Coinbase, Kraken, OKX).
   - `internal/oracle/`: Besu Chainlink contract client and heartbeat validator.
   - `internal/aggregator/`: Trimmed median algorithm, MAD filter, and index pricing engine.
   - `internal/candle/`: Sliding-window multi-resolution OHLCV accumulator.
   - `internal/publisher/`: Kafka and Redis message dispatchers.
   - `internal/storage/`: TimescaleDB batch persistence worker.
   - `internal/server/`: gRPC health and inspection server.
   - `proto/`: Protocol Buffer definitions.
2. **Define Protobuf Schemas:** Author `proto/market_feeder.proto` defining `ExchangeTickMessage`, `TickerMessage`, `CandleMessage`, and `FairIndexPriceMessage`. Generate Go bindings via `protoc-gen-go` and `protoc-gen-go-grpc`.
3. **Implement Resilient WebSocket Client Core:** Build a reusable, thread-safe WebSocket connection wrapper (`internal/adapters/ws_client.go`) managing TLS handshakes, socket read/write deadlines, ping/pong handlers, read pumps, and exponential backoff reconnection with decorrelated jitter:
   $$T_{\text{reconnect}} = \min(T_{\text{max}},\, T_{\text{base}} \times 2^{\text{attempt}}) \pm \text{jitter}$$
4. **Implement Exchange-Specific Adapters:**
   - *Binance Adapter:* Connect to trade stream, subscribe to `btcusdt@trade`, parse price, quantity, trade ID, and trade execution timestamp.
   - *Coinbase Adapter:* Subscribe to `matches` channel for product `BTC-USDT` (or fallback to `BTC-USD`), unpack ISO-8601 timestamps and decimal fields.
   - *Kraken Adapter:* Subscribe to `trade` channel for `BTC/USDT`, handle Kraken sequence counters and array-based payload layouts.
   - *OKX Adapter:* Subscribe to `trades` channel for instId `BTC-USDT`, process batched trade objects.
5. **Implement Normalized Internal Tick Pipeline:** Direct all raw exchange trades into a buffered, lock-free ring buffer or bounded Go channel (`chan *ExchangeTickMessage`, capacity 100,000) to isolate network I/O from statistical processing.
6. **Implement Chainlink On-Chain Oracle Client:** Integrate `ethclient` targeting the Besu RPC URL. Implement `GetLatestAnchorPrice()` invoking the `AggregatorV3Interface` contract on Besu. Enforce heartbeat checks and staleness boundaries.
7. **Build Trimmed Median & MAD Outlier Rejection Engine:**
   - Maintain a sliding 1-second window of latest prices across all active exchanges.
   - For every calculation tick, sort prices $P = [p_1, p_2, \dots, p_n]$:
     - If $n \ge 4$, compute the trimmed median by discarding the lowest and highest values and averaging the central elements.
     - Calculate Median Absolute Deviation: $\text{MAD} = \text{median}(|p_i - \text{median}(P)|)$.
     - Discard any exchange tick where $|p_i - \text{median}(P)| > 3 \times \text{MAD} \times 1.4826$.
     - Verify resulting fair index against Chainlink anchor price within $\pm 3.0\%$ tolerance.
8. **Build Real-Time Ticker Synthesizer:** Ingest validated ticks to compute rolling 24-hour high, 24-hour low, cumulative volume, Last Traded Price (LTP), 24-hour percentage change, and current exchange weight distribution.
9. **Implement In-Memory Candlestick Engine:** Construct sliding-window aggregators for 1s, 1m, 5m, 15m, 1h, and 1d intervals. On each valid tick, update the current bar's High, Low, Close, Volume, and VWAP. At each boundary (e.g., minute transition), seal the completed candle and initialize the subsequent candle with $O = C_{\text{prev}}$.
10. **Build Low-Latency Redis Pub/Sub Fanout Worker:** Establish pooled Redis connections and publish serialized Protobuf payloads to designated channels (`market:feed:ticker:BTCUSDT`, `market:feed:candles:BTCUSDT:1m`) within $< 1\text{ms}$ of tick arrival.
11. **Build Kafka Producer Pipeline:** Implement a high-throughput, non-blocking Kafka writer emitting to `market.external.btc_usdt.ticks.v1`, `market.external.btc_usdt.ticker.v1`, and `market.external.btc_usdt.candles.v1` with symbol partitioning for sequential ordering.
12. **Implement TimescaleDB Batch Persistence:** Create background batch worker aggregating completed candles and micro-sampled ticks into slices, executing atomic PostgreSQL `COPY` or parameterized batch `INSERT` queries every 5 seconds.
13. **Implement Heartbeat Monitoring & Failover Sentinel:** Monitor individual exchange tick intervals. If an exchange fails to emit a tick within 3,000ms, mark its health status as degraded, exclude its weight from the trimmed median calculation, log an alert, and trigger a background reconnect probe.
14. **Implement Prometheus Instrumentation:** Expose `/metrics` on port 9090 tracking tick ingestion rate per exchange, outlier rejection count, end-to-end processing latency histograms, active WebSocket connection status, and Besu oracle discrepancy delta.
15. **Execute Fault-Tolerance & Flash-Crash Simulation Tests:** Run integration tests injecting simulated flash crashes (-20% sudden spike on a single exchange) to confirm 100% rejection by the MAD engine with zero corruption of the index price or candlestick streams.

---

## Interfaces / Contracts

### Protobuf Definition (`proto/market_feeder.proto`)
```protobuf
syntax = "proto3";

package growww.marketfeeder.v1;

option go_package = "growww/marketfeeder/v1;marketfeederv1";

enum ExchangeVenue {
  EXCHANGE_VENUE_UNSPECIFIED = 0;
  EXCHANGE_VENUE_BINANCE = 1;
  EXCHANGE_VENUE_COINBASE = 2;
  EXCHANGE_VENUE_KRAKEN = 3;
  EXCHANGE_VENUE_OKX = 4;
}

message ExchangeTickMessage {
  string trade_id = 1;
  ExchangeVenue venue = 2;
  string symbol = 3;             // e.g., "BTCUSDT"
  string price = 4;              // Fixed-point decimal string
  string quantity = 5;           // Fixed-point decimal string
  int64 exchange_timestamp_ns = 6;
  int64 ingested_timestamp_ns = 7;
  bool is_buyer_maker = 8;
}

message FairIndexPriceMessage {
  string symbol = 1;             // "BTCUSDT"
  string index_price = 2;        // Computed fair trimmed median
  string chainlink_price = 3;    // On-chain anchor price
  int64 timestamp_ns = 4;
  uint32 active_venues_count = 5;
  repeated ExchangeVenue participating_venues = 6;
  bool anchor_divergence_warning = 7;
}

message TickerMessage {
  string symbol = 1;             // "BTCUSDT"
  string ltp = 2;                // Last Traded Price (index or last validated tick)
  string fair_index_price = 3;   // Real-time trimmed median
  string high_24h = 4;
  string low_24h = 5;
  string volume_24h = 6;
  string turnover_24h = 7;       // 24h quote volume in USDT
  string price_change_24h = 8;
  string price_change_percent_24h = 9;
  int64 timestamp_ms = 10;
  uint64 sequence_number = 11;
}

message CandleMessage {
  string symbol = 1;             // "BTCUSDT"
  string interval = 2;           // "1s", "1m", "5m", "15m", "1h", "1d"
  int64 start_time_ms = 3;
  int64 close_time_ms = 4;
  string open = 5;
  string high = 6;
  string low = 7;
  string close = 8;
  string volume = 9;             // Total BTC volume
  string turnover = 10;          // Total USDT turnover
  string vwap = 11;              // Volume-Weighted Average Price
  uint64 trade_count = 12;
  bool is_closed = 13;           // True if candle interval has sealed
}

service MarketFeederService {
  rpc GetCurrentIndex (GetCurrentIndexRequest) returns (FairIndexPriceMessage);
  rpc StreamTicks (StreamTicksRequest) returns (stream ExchangeTickMessage);
  rpc StreamIndex (StreamIndexRequest) returns (stream FairIndexPriceMessage);
}

message GetCurrentIndexRequest {
  string symbol = 1;
}

message StreamTicksRequest {
  string symbol = 1;
  repeated ExchangeVenue venues = 2;
}

message StreamIndexRequest {
  string symbol = 1;
}
```

### Redis Channel Topology
| Channel Pattern | Content Type | Publishing Frequency | Consumer Services | Description |
| :--- | :--- | :--- | :--- | :--- |
| `market:feed:ticker:BTCUSDT` | Protobuf (`TickerMessage`) | On every index update / 100ms throttle | Market Data Service (Prompt 207), Gateway | High-frequency live ticker broadcast |
| `market:feed:index:BTCUSDT` | Protobuf (`FairIndexPriceMessage`) | Real-time on every tick batch | Matching Engine (Prompt 205), Margin Engine | Reference mark price & circuit breaker checks |
| `market:feed:candles:BTCUSDT:1s` | Protobuf (`CandleMessage`) | 1-second interval | Real-Time Charting Worker, Analytics | Micro-candlestick stream for live visualizers |
| `market:feed:candles:BTCUSDT:1m` | Protobuf (`CandleMessage`) | 1-minute interval / real-time updates | Market Data Service (Prompt 207) | 1-minute standard trading charts |
| `market:feed:candles:BTCUSDT:1d` | Protobuf (`CandleMessage`) | 1-day interval / real-time updates | Portfolio Service (Prompt 209), Daily Stats | 24-hour daily chart reference |

### Kafka Topic Standards
| Topic Name | Key Format | Value Serialization | Partitions | Retention Policy | Purpose |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `market.external.btc_usdt.ticks.v1` | `BTCUSDT` | Protobuf (`ExchangeTickMessage`) | 6 | 3 days (compact/delete) | Raw normalized tick stream for Demo Engine & Replay |
| `market.external.btc_usdt.ticker.v1` | `BTCUSDT` | Protobuf (`TickerMessage`) | 3 | 7 days (delete) | Downstream financial services & analytics |
| `market.external.btc_usdt.candles.v1` | `BTCUSDT:{interval}` | Protobuf (`CandleMessage`) | 3 | 30 days (compact) | Historical bar persistence & backup streaming |

---

## Security & Compliance Notes
- **Protection Against Oracle & Flash-Loan Manipulation:** In decentralized and centralized venues, flash loans or low-liquidity attacks can skew spot prices by millions of dollars within a single block. The feeder guards against this by requiring a minimum quorum of at least 3 independent Tier-1 spot exchanges plus the Besu on-chain Chainlink anchor. If fewer than 2 external exchanges are connected, the system immediately flags the index as degraded and widens matching engine risk collars.
- **Flash Crash & Spike Filtering:** Any tick deviating by more than $3 \times \text{MAD}$ from the current median or exceeding a $\pm 1.5\%$ instantaneous price jump within a 500ms sliding window is discarded as an aberrant spike. Discarded ticks are logged with venue ID, timestamp, and deviation magnitude for audit forensic analysis.
- **WebSocket Reconnection & Backoff Architecture:** All external exchange connections employ exponential backoff with full jitter to avoid the "thundering herd" problem during public internet routing flaps. Sockets utilize TLS 1.3 with strict server certificate validation and separate read/write goroutines protected by channel-based lifecycle management to prevent memory leaks or hung goroutines.
- **Heartbeat & Zombie Connection Detection:** Public exchange WebSockets frequently stall silently without sending TCP FIN packets. The feeder sets a 5-second socket read deadline that is refreshed exclusively upon receiving valid WebSocket frames or pong replies. If the deadline expires, the socket is immediately terminated and recycled.
- **Zero-PII Invariant:** Market data ingestion, index processing, and broadcast streams contain zero personally identifiable information (PII) or user accounts. All processing is public market telemetry.
- **Auditability & Regulatory Verification:** Under IFSCA and SEBI algorithmic market guidelines, all computed index prices record the participating exchange venue bitmask, the discarded venue IDs, and the exact nanosecond calculation timestamp, guaranteeing full forensic determinism.

---

## Acceptance Criteria
- [ ] Concurrently maintains persistent TLS WebSocket connections to Binance, Coinbase, Kraken, and OKX with $< 0.1\%$ dropped frames during continuous operation.
- [ ] Automatic reconnection triggers within 1,000ms of socket disruption and restores normal ingestion within 5 seconds without crashing the service.
- [ ] Ingests and validates the Chainlink BTC/USD oracle round from Hyperledger Besu every 5 seconds, verifying timestamp freshness within 120 seconds.
- [ ] Outlier detection engine correctly isolates and rejects intentional simulated single-exchange flash spikes ($> 2.5\%$ deviation) in $< 1\text{ms}$ without altering the composite index price.
- [ ] Trimmed median Fair Index Price computation and fanout to Redis Pub/Sub completes within $< 5\text{ms}$ of receiving external trade ticks (p99 latency).
- [ ] In-memory candlestick engine aggregates 1s, 1m, 5m, 15m, 1h, and 1d bars with zero missed ticks, correct VWAP computation, and consistent boundary transitions.
- [ ] Kafka producer successfully publishes all normalized ticks and sealed candles with partition ordering preserved.
- [ ] Demo Trading Engine (Prompt 273) successfully ingests `market.external.btc_usdt.ticks.v1` and simulates paper trading fills at live market prices.
- [ ] Prometheus endpoint exposes exchange tick rates, outlier rejection counts, and processing latency histograms with zero deadlocks under heavy market loads (10,000 ticks/sec).
- [ ] Service memory footprint remains stable ($< 150\text{MB}$) over a 24-hour continuous burn-in test with zero goroutine leaks.

---

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Design Standards), 104 (Kafka Event Schemas), 207 (Real-Time Market Data Service), 302 (Network Topology & Besu Setup).
- **Parallel Tasks:** 205 (Order Matching Engine), 206 (Pre-Trade Risk & Margin Checks).
- **Downstream Blockers:** 273 (Demo Trading & Paper Simulation Engine), 508 (Flutter BTC Detail & Charting Screen), 603 (Web Trading Dashboard BTC/USDT Terminal).
