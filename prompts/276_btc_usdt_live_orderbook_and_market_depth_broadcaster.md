# 276 - BTC/USDT Real-Time Order Book & Market Depth Broadcaster (Go / Rust)

## Purpose
High-frequency financial trading systems require rapid, deterministic dissemination of market liquidity to frontend charting widgets, mobile trading terminals, and algorithmic API consumers. In the BTC/USDT spot and derivative markets, price updates occur at thousands of events per second. Transmitting every raw resting order change directly to client browsers or mobile devices saturates client network sockets, triggers garbage collection pauses, causes browser thread lockups, and degrades battery life on mobile devices.

The **BTC/USDT Real-Time Order Book & Market Depth Broadcaster** solves this fundamental market microstructure challenge. It provides high-frequency aggregation and dissemination of Level-2 (aggregated price ladders) and Level-3 (order-by-order state) BTC/USDT order books and market depth to frontend chart and terminal interfaces across Flutter mobile, desktop apps, and Next.js web trading dashboards.

Operating as a dedicated, low-latency microservice (`services/depth-broadcaster`), the system ingests microsecond order life-cycle events from the core Matching Engine (Prompt 205) and Demo Matching Engine (Prompt 273). It maintains an in-memory, lock-free order book state machine, merges internal resting orders with live reference tickers, calculates deterministic 100ms conflated depth deltas, and fans out compressed binary (Protobuf) and JSON streams to tens of thousands of concurrent WebSocket subscribers with microsecond-level internal processing latency.

---

## What You Are Building
A high-throughput, low-latency market depth broadcaster service (`services/depth-broadcaster`) engineered in Go 1.22+ or Rust 1.78+ that acts as the dedicated real-time market depth distribution layer for BTC/USDT. Core deliverables include:

- **Dual-Source Ingestion Pipeline:** High-performance stream consumers ingesting microsecond-level order placement, cancellation, modification, and fill events from the Production Matching Engine (Prompt 205) and Demo Matching Engine (Prompt 273).
- **In-Memory Lock-Free Order Book Engine:** An optimized order book state machine maintaining continuous Level-2 aggregated price ladders (top 5, 10, 20, 50, and 100 price levels) and internal Level-3 book state using cache-conscious price-indexed arrays or lock-free skip lists.
- **100ms Conflation & Delta Computation Engine:** A deterministic ticker pipeline that batches and conflates microsecond book oscillations over fixed 100ms intervals, computing exact incremental depth diffs (`depth_delta`) and periodic baseline snapshots (`depth_snapshot`).
- **High-Concurrency WebSocket Broadcast Hub:** A massively scalable WebSocket distribution server capable of maintaining 100,000+ persistent client connections with zero-allocation memory framing, supporting permessage-deflate compression and dynamic protocol switching (Protobuf binary vs. JSON).
- **Client Ring-Buffer Backpressure Supervisor:** Per-connection bounded circular ring buffers with drop-oldest semantics and proactive slow-consumer eviction, shielding the broadcaster from slow 4G/5G mobile client head-of-line blocking.
- **Sequence Continuity & Checksum Engine:** Generation of strictly monotonic sequence counters (`first_update_id`, `last_update_id`) and top-10 depth CRC32/Adler32 checksums, enabling frontend clients to detect dropped packets and trigger self-healing snapshot resynchronization.
- **Multi-Environment Routing (Live vs. Demo):** Transparent channel multiplexing allowing clients to toggle seamlessly between real exchange liquidity (`btc_usdt@depth`) and synthetic market-maker liquidity (`btc_usdt_demo@depth`) without reconnecting.

---

## Scope Boundaries

### In Scope
- Microsecond ingestion of resting order creation, cancellation, modification, and execution events for BTC/USDT from the Matching Engine (Prompt 205) and Demo Matching Engine (Prompt 273).
- Real-time maintenance of in-memory Level-2 depth state (top 5, 10, 20, 50, and 100 bids and asks) and Level-3 order queues.
- Conflating order book deltas into deterministic 100ms broadcast frames to eliminate client UI thrashing.
- Streaming initial full-depth snapshots upon client connection or channel subscription.
- Streaming incremental depth deltas with strict update ID sequences (`first_update_id`, `last_update_id`).
- Real-time computation of CRC32 / Adler32 checksums over the top-10 bid/ask ladder for continuous client-side integrity validation.
- Encoding outgoing market depth frames in both zero-copy Protobuf binary and standard JSON formats.
- Managing client WebSocket sessions, heartbeat ping/pong cycles, channel subscriptions, and backpressure eviction.
- Redis ring-buffer caching for instant snapshot retrieval and cross-pod synchronization.
- Microsecond metrics export to Prometheus for broadcast latency, connection counts, dropped frames, and memory overhead.

### Out of Scope / Handled Elsewhere
- Matching engine trade execution, limit order book matching algorithms, and trade clearing (handled in Prompt 205 and Prompt 273).
- General market ticker synthesis (24h high, low, volume, VWAP) and OHLCV candlestick aggregation (handled in Market Data Service Prompt 207).
- Network perimeter TLS termination, edge DDoS mitigation, and global client authentication (handled in API Gateway Prompt 219).
- Frontend chart rendering, WebGL canvas painting, and DOM updates (handled in Flutter Trading UI Prompt 508 / Prompt 509 and Web Trading Dashboard Prompt 603).
- Fiat/crypto balance verification, margin requirements, and wallet ledger debits/credits (handled in Prompt 203 and Prompt 206).
- Direct on-chain transaction submission or block indexing (handled in blockchain settlement adapters).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ or Rust 1.78+.
  - *Go option:* Utilizes lightweight goroutines, channel multiplexing, `sync.Pool` for zero-allocation byte buffers, and `github.com/coder/websocket` for high-concurrency connection handling.
  - *Rust option:* Utilizes Tokio async runtime, `tokio-tungstenite`, crossbeam lock-free channels, and zero-cost abstractions to guarantee sub-millisecond end-to-end fanout without garbage collection pauses.
- **WebSocket Transport Framework:** `github.com/coder/websocket` (Go) or `tokio-tungstenite` (Rust) configured with custom TCP read/write buffer tuning (`SO_RCVBUF`, `SO_SNDBUF`), TCP_NODELAY enabled, and permessage-deflate support.
- **Binary Serialization:** Google Protocol Buffers v3 (`google.golang.org/protobuf` or `prost` in Rust) for compact, zero-copy binary depth frames.
- **Fast JSON Serialization:** `github.com/bytedance/sonic` or `simdjson-go` (Go) or `serde_json` (Rust) for ultra-fast JSON serialization for browser clients unable to process Protobuf binary frames.
- **In-Memory Caching & Ring Buffers:** Redis 7.2+ Cluster using Redis Hashes for latest depth snapshots and Redis Pub/Sub or Streams for inter-pod distribution and state hydration.
- **Inter-Service Communication:** Apache Kafka 3.7+ (`segmentio/kafka-go` or `rdkafka`) and gRPC over HTTP/2 for receiving order book change feeds directly from matching engines.
- **High-Precision Fixed-Point Math:** 64-bit integer arithmetic scaling BTC satoshis ($10^8$) and micro-USDT ($10^6$), eliminating floating-point rounding inaccuracies in price-ladder aggregation.

---

## Backend / Infra Touchpoints
- **Matching Engine (Prompt 205):** Ingests live production resting order modifications, trade executions, and cancellations via dedicated low-latency gRPC streaming or shared-memory Kafka topics (`matching.orders.btc_usdt.v1`, `matching.book_events.btc_usdt.v1`).
- **Demo Matching Engine (Prompt 273):** Ingests paper-trading order book events and synthetic market-maker liquidity updates via topic `demo.matching.book_events.btc_usdt.v1`.
- **Market Data Service (Prompt 207):** Coordinates with general market data streaming; the Market Data Service delegates specialized L2/L3 order book WebSocket requests to this dedicated depth broadcaster.
- **API Gateway & BFF (Prompt 219):** Authenticates client WebSocket handshake requests, extracts user JWT / session scope, validates rate limits, and proxies WebSocket connections to the depth broadcaster cluster via sticky session routing.
- **Redis Cluster:** Maintains the latest 100-level order book snapshots for BTC/USDT and BTC/USDT-DEMO, enabling instant connection hydration without querying the matching engine.
- **Prometheus & Grafana:** Ingests real-time telemetry including:
  - `depth_broadcaster_connected_clients`: Active WebSocket connections.
  - `depth_broadcaster_conflation_duration_micros`: Time required to conflate 100ms book state.
  - `depth_broadcaster_broadcast_latency_micros`: Time from conflation completion to socket write.
  - `depth_broadcaster_dropped_messages_total`: Count of frames dropped due to slow consumer ring buffers.
  - `depth_broadcaster_client_evictions_total`: Disconnections triggered by unrecoverable client backpressure.

---

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Zero Direct Blockchain Interaction:** The Market Depth Broadcaster operates entirely off-chain at microsecond to millisecond cadences. Querying or writing to the Hyperledger Besu distributed ledger during real-time 100ms broadcast cycles is strictly prohibited to avoid latency spikes and blocking I/O.
- **Representation of Ledger-Backed Assets:** The BTC/USDT order book reflects trades of Bitcoin and USDT holdings that are 1:1 physically collateralized and audited via Proof-of-Reserve contracts (Prompt 327) and multi-sig vault registries on Hyperledger Besu under QBFT consensus.
- **Zero PII Invariant:** All order book broadcasts are completely anonymous. Broadcast frames contain only aggregated price levels, resting quantities, and order counts. No user IDs, wallet addresses, KYC identifiers, or account numbers are ever embedded in public or authenticated market depth channels.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Broadcaster Service Repository:** Initialize the project module at `services/depth-broadcaster` with modular directory structure:
   - `cmd/server/`: Entry point, configuration loader, OS signal trap, and graceful shutdown handlers.
   - `internal/book/`: In-memory Level-2/Level-3 order book data structures, lock-free price ladders, and delta calculation logic.
   - `internal/conflator/`: 100ms ticker wheel, delta accumulator, and checksum computation.
   - `internal/ingest/`: High-speed Kafka consumer and gRPC client adapters for Matching Engine and Demo Matching Engine.
   - `internal/ws/`: WebSocket connection manager, subscription router, framing, and client session registry.
   - `internal/buffer/`: Bounded circular ring buffer with slow-consumer detection and drop-oldest policies.
   - `proto/`: Protobuf message definitions and generated Go/Rust stubs.
2. **Define Protobuf Binary Specifications:** Author `proto/growww/depth/v1/market_depth.proto` declaring messages for `OrderBookSnapshot`, `OrderBookDelta`, `PriceLevel`, `DepthSubscriptionRequest`, and `DepthSubscriptionResponse`. Compile stubs using `buf` or `protoc`.
3. **Implement In-Memory Level-2 Order Book State Machine:** Build a high-performance order book aggregator in `internal/book/`. Maintain separate sorted structures for bids (descending) and asks (ascending):
   - Store prices and quantities as fixed-point integers (`int64`).
   - Implement methods `ApplyOrderAdd()`, `ApplyOrderModify()`, `ApplyOrderCancel()`, and `ApplyTradeFill()`.
   - Provide atomic snapshot extraction: `ExtractSnapshot(depth int) -> OrderBookSnapshot`.
4. **Implement 100ms Conflation Engine:** Build a periodic conflation worker triggered by a microsecond-resolution ticker every 100ms:
   - Accumulate micro-updates occurring within the 100ms window into a single consolidated delta frame per symbol.
   - Assign monotonically increasing sequence counters: `first_update_id` (sequence of first book modification in batch) and `last_update_id` (sequence of last modification in batch).
   - If no book changes occur within the 100ms window, suppress broadcast to conserve network bandwidth.
5. **Implement Real-Time Checksum Calculator:** Author an optimized Adler32/CRC32 checksum generator over the top-10 bids and top-10 asks:
   - Concatenate price-quantity pairs in canonical order: `bid_p0:bid_q0:bid_p1:bid_q1:...:ask_p0:ask_q0:...`.
   - Compute CRC32 integer hash and attach to both `OrderBookSnapshot` and `OrderBookDelta` frames for client validation.
6. **Implement Ingestion Adapters for Live and Demo Matching Engines:**
   - Implement gRPC streaming consumer reading from Matching Engine (Prompt 205).
   - Implement Kafka consumer group reading from Demo Matching Engine (Prompt 273) topic `demo.matching.book_events.btc_usdt.v1`.
   - Normalize ingested raw order events into internal `BookEvent` structs and route to the appropriate order book instance (`btc_usdt_live` vs `btc_usdt_demo`).
7. **Build Redis Snapshot Hydration Worker:**
   - Persist latest 50-level and 100-level snapshots to Redis Hashes every 1000ms (`depth:snapshot:btc_usdt`, `depth:snapshot:btc_usdt_demo`).
   - On broadcaster pod restart, immediately hydrate local in-memory order books from Redis before opening consumer streams, ensuring zero cold-start delay.
8. **Build High-Concurrency WebSocket Server:**
   - Implement HTTP upgrade handler at `/v1/market/depth/stream` supporting both WebSocket and secure WSS protocols.
   - Configure socket options: disable Nagle's algorithm (`TCP_NODELAY`), configure 64KB read/write buffers, and enable permessage-deflate compression with threshold > 512 bytes.
   - Support protocol negotiation: default to Protobuf binary for native mobile/desktop apps, fall back to JSON for standard web clients.
9. **Implement Client Subscription Router:**
   - Parse inbound text frames: `{"action": "subscribe", "channel": "btc_usdt@depth20@100ms", "format": "protobuf"}`.
   - Register clients in lock-free subscription maps (`sync.Map` or partitioned read-write mutex maps).
   - Immediately dispatch an initial `OrderBookSnapshot` to the subscriber so the client establishes a valid baseline before processing deltas.
10. **Implement Client Bounded Ring Buffer & Backpressure Supervisor:**
    - Allocate a bounded circular ring buffer (capacity: 256 messages, ~128KB) for each connected WebSocket client.
    - If a client socket is blocked and the buffer fills to 100%:
      - Strategy 1 (Transient lag): Drop older delta updates while logging a backpressure warning metric.
      - Strategy 2 (Critical lag > 2000ms): Disconnect the client with WebSocket close code `4008` (Policy Violation / Buffer Overflow), forcing the client to reconnect and request a fresh snapshot.
11. **Implement Dual-Mode Live/Demo Routing Engine:**
    - Inspect subscription channel names: route `btc_usdt@depth` to the production matching engine book, and route `btc_usdt_demo@depth` to the demo paper-trading book.
    - Ensure complete memory and channel isolation between live and demo state trees to prevent synthetic mock liquidity from leaking into production books.
12. **Implement Heartbeat Ping/Pong Supervisor:**
    - Schedule periodic WebSocket `PING` frames every 30 seconds.
    - Terminate zombie connections if a corresponding `PONG` response is not received within a 10-second timeout window.
13. **Add Telemetry, Tracing, and Prometheus Metrics:**
    - Register Prometheus metrics for fanout throughput, broadcast latency, active subscriptions by symbol, and conflation cycle time.
    - Integrate OpenTelemetry context propagation to track order-to-broadcast propagation delay across the distributed system.
14. **Author Comprehensive Unit & Benchmarking Test Suite:**
    - Benchmark order book insertion and snapshot extraction under synthetic loads of 50,000 events/sec.
    - Verify strict monotonic ordering of `first_update_id` and `last_update_id`.
    - Verify CRC32 checksum calculation correctness across edge cases (empty book, inverted spreads, fractional quantities).
    - Test slow-consumer ring buffer eviction under simulated network throttling.

---

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/depth/v1/market_depth.proto`)
```protobuf
syntax = "proto3";

package growww.depth.v1;

option go_package = "growww/depth/v1;depthv1";

// Individual price level in the order book ladder
message PriceLevel {
  // Price scaled to fixed-point integer (e.g., satoshis or micro-units)
  int64 price_raw = 1;
  // Floating-point string representation for direct UI rendering
  string price = 2;
  // Aggregated volume available at this price level
  int64 quantity_raw = 3;
  string quantity = 4;
  // Number of resting orders contributing to this price level (L2 depth)
  int32 order_count = 5;
}

// Full snapshot of the order book ladder
message OrderBookSnapshot {
  string symbol = 1; // e.g. "BTC/USDT" or "BTC/USDT-DEMO"
  int64 last_update_id = 2; // Monotonically increasing sequence ID
  int64 timestamp_ms = 3; // Timestamp of snapshot generation
  repeated PriceLevel bids = 4; // Sorted descending by price
  repeated PriceLevel asks = 5; // Sorted ascending by price
  uint32 checksum = 6; // CRC32 checksum of top 10 bids/asks
  bool is_demo = 7; // Flag indicating production vs sandbox liquidity
}

// Incremental 100ms conflated delta update
message OrderBookDelta {
  string symbol = 1;
  int64 first_update_id = 2; // Sequence ID of first update in batch
  int64 last_update_id = 3; // Sequence ID of last update in batch
  int64 timestamp_ms = 4;
  repeated PriceLevel updated_bids = 5; // Quantity = 0 indicates price level removal
  repeated PriceLevel updated_asks = 6;
  uint32 checksum = 7; // CRC32 checksum of resulting top 10 bids/asks
  bool is_demo = 8;
}

// Client WebSocket subscription message
message DepthSubscriptionRequest {
  enum Action {
    ACTION_UNSPECIFIED = 0;
    SUBSCRIBE = 1;
    UNSUBSCRIBE = 2;
  }
  enum Format {
    FORMAT_UNSPECIFIED = 0;
    PROTOBUF_BINARY = 1;
    JSON_TEXT = 2;
  }
  Action action = 1;
  string channel = 2; // e.g. "btc_usdt@depth20@100ms"
  Format preferred_format = 3;
  int32 depth_levels = 4; // Requested depth: 5, 10, 20, 50, or 100
}

// Server response acknowledging subscription
message DepthSubscriptionResponse {
  bool success = 1;
  string channel = 2;
  string message = 3;
  int64 timestamp_ms = 4;
}
```

### JSON WebSocket Message Schemas

#### 1. Client Subscription Frame (Inbound)
```json
{
  "action": "subscribe",
  "channel": "btc_usdt@depth20@100ms",
  "depth": 20,
  "format": "json"
}
```

#### 2. Order Book Full Snapshot Frame (Outbound Initial Response)
```json
{
  "event": "depth_snapshot",
  "symbol": "BTC/USDT",
  "last_update_id": 18492048201,
  "timestamp_ms": 1726732800102,
  "bids": [
    ["64250.50", "3.42150000", 12],
    ["64250.00", "5.10500000", 8],
    ["64249.50", "1.20000000", 4]
  ],
  "asks": [
    ["64251.00", "2.15000000", 9],
    ["64251.50", "4.89000000", 15],
    ["64252.00", "0.95000000", 2]
  ],
  "checksum": 2948102931,
  "is_demo": false
}
```
*Note on array format:* Each price level in JSON is serialized as a 3-element tuple: `[price_string, quantity_string, order_count]`.

#### 3. Incremental 100ms Delta Frame (Outbound Stream)
```json
{
  "event": "depth_delta",
  "symbol": "BTC/USDT",
  "first_update_id": 18492048202,
  "last_update_id": 18492048215,
  "timestamp_ms": 1726732800200,
  "bids": [
    ["64250.50", "4.10000000", 14],
    ["64249.00", "0.00000000", 0]
  ],
  "asks": [
    ["64251.00", "1.85000000", 7]
  ],
  "checksum": 3819402842,
  "is_demo": false
}
```
*Note:* A quantity of `"0.00000000"` signals that the price level has been completely removed from the order book ladder.

### Checksum Algorithm Specification
The client and server calculate a CRC32 checksum over the top 10 bids and top 10 asks after applying each snapshot or delta update:
1. Extract top 10 bids sorted descending by price, and top 10 asks sorted ascending by price.
2. Format each level as `<price>:<quantity>` (e.g. `64250.50:4.10000000`).
3. Construct a single colon-separated canonical string: `<bid_0>:<ask_0>:<bid_1>:<ask_1>:...:<bid_9>:<ask_9>`.
4. Calculate IEEE 802.3 CRC32 integer hash over the UTF-8 bytes of this string.
5. If client checksum differs from server checksum, the client discards local book state and requests a fresh `depth_snapshot`.

---

## Security & Compliance Notes

- **WebSocket Backpressure Management & DoS Protection:**
  - Fast-producing broadcast systems risk unbounded memory growth if slow clients (e.g., mobile devices on weak cellular connections) fail to read TCP frames fast enough.
  - The broadcaster maintains a bounded circular ring buffer per client (maximum 256 messages). If the buffer capacity is exceeded, the service drops oldest delta frames and sets a client resync flag.
  - If a client's write buffer remains saturated for $> 2000\text{ms}$, the server terminates the connection with WebSocket close code `4008` (Policy Violation), freeing memory buffers and preventing cascading socket pool exhaustion.
- **Connection Rate Limiting & Concurrency Quotas:**
  - IP-based connection limits enforced via API Gateway (Prompt 219) and broadcaster middleware: maximum 10 simultaneous WebSocket connections per IP and maximum 5 active channel subscriptions per connection.
  - Strict handshake verification: rejection of unauthorized origin headers to prevent Cross-Site WebSocket Hijacking (CSWSH).
- **Sequence Continuity & Anti-Desynchronization:**
  - To prevent displaying stale or corrupted market depth to traders, all delta frames contain strict sequence counters: for consecutive frames $n$ and $n+1$, the condition $\text{first\_update\_id}_{n+1} = \text{last\_update\_id}_n + 1$ must strictly hold.
  - Frontend SDKs must verify sequence continuity; any sequence gap indicates network loss, triggering an immediate snapshot re-synchronization.
- **Fair Dissemination & Latency Arbitrage Neutralization:**
  - Market depth conflation is fixed at 100ms ticks synchronized to the server monotonic clock.
  - Conflated delta frames are broadcast concurrently across all active client ring buffers using non-blocking fanout workers, ensuring institutional API clients, web terminals, and mobile users receive market data updates within the same microsecond transmission wave without preferential data leaks.
- **Zero PII Leak Prevention:**
  - The broadcaster completely filters internal order metadata: trader user IDs, account numbers, order IDs, and execution timestamps are stripped before frames are constructed. Only anonymized aggregated price levels and order counts are transmitted.

---

## Acceptance Criteria

- [ ] Low-latency broadcaster service initializes and runs within `services/depth-broadcaster` without memory leaks or race conditions.
- [ ] Successfully ingests real-time order life-cycle events from both Matching Engine (Prompt 205) and Demo Matching Engine (Prompt 273).
- [ ] In-memory order book maintains accurate Level-2 depth (up to 100 levels) for both production BTC/USDT and synthetic demo BTC/USDT-DEMO pairs.
- [ ] 100ms conflation ticker correctly batches high-frequency order book oscillations, emitting exactly one delta update per active 100ms window when book modifications occur.
- [ ] Outgoing messages support dual encoding formats: compact Protocol Buffers v3 binary frames and clean, standards-compliant JSON frames.
- [ ] Sequence numbers (`first_update_id`, `last_update_id`) increment monotonically without gaps, allowing clients to detect dropped or reordered network packets.
- [ ] Top-10 bid/ask CRC32 checksums match client-side calculations across all snapshot and delta updates.
- [ ] Slow-consumer backpressure supervisor successfully detects lagging network connections, drops stale deltas, and terminates unrecoverable zombie sockets without impacting healthy subscribers.
- [ ] Broadcaster supports 50,000+ concurrent simulated WebSocket connections on a standard cluster node with sub-5ms fanout latency.
- [ ] Zero PII or private trader identity data is exposed in public or authenticated market depth streams.
- [ ] Broadcaster exposes Prometheus metrics capturing connected clients, broadcast latency, conflation execution time, and dropped frame counts.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: WebSocket protocol standards, JSON framing conventions, and gRPC contracts.
  - `104_event_schema_and_kafka_topic_standards.md`: Kafka topic naming and serialization standards.
  - `111_domain_model_core_entities.md`: Canonical currency, symbol, and precision definitions.
  - `205_order_matching_engine.md`: Core matching engine producing resting order book modification events.
  - `207_market_data_service.md`: Core market data architecture and ticker streaming baseline.
- **Parallel Tasks:**
  - `219_api_gateway_and_bff.md`: Reverse proxy routing, WebSocket upgrade validation, and client rate limiting.
  - `270_rfq_and_instant_convert_swap_service.md`: Instant conversion pricing consuming top-of-book market depth.
  - `273_crypto_demo_trading_paper_matching_engine.md`: Demo matching engine producing synthetic BTC/USDT liquidity events.
- **Downstream Blockers:**
  - `508_flutter_real_time_candlestick_and_depth_chart.md`: Flutter mobile market depth ladder and order book UI.
  - `509_flutter_order_placement_and_trading_interface.md`: Mobile 1-click trading interface displaying live bid/ask spreads.
  - `603_web_trading_dashboard_and_order_entry.md`: Professional web trading terminal displaying real-time 50-level depth DOM (Depth of Market) ladders and live order book visualizers.
