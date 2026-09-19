# 113 - Distributed Tracing & OpenTelemetry Microsecond Precision Standard

## Purpose
Establishes the enterprise-wide distributed tracing, OpenTelemetry (OTel) instrumentation, and microsecond-precision latency profiling standard across all tiers of the Growww trading platform. In an institutional-grade, low-latency financial brokerage executing high-throughput order matching alongside fractional security token settlements, deterministic observability across hybrid distributed systems is paramount.

Transactions traverse polyglot boundaries: starting from Flutter mobile clients, traversing Envoy API Gateways, entering Go pre-trade risk and order orchestration microservices, executing within a nanosecond-benchmarked Rust order matching engine, fanning out across Apache Kafka event streams, invoking asynchronous Go DvP trade settlement engines, and culminating in private Hyperledger Besu blockchain consensus commits. 

This specification establishes end-to-end W3C TraceContext propagation, lock-free low-overhead tracing within the matching engine hot path, high-throughput OpenTelemetry Collector topologies, and deep correlation linking off-chain trade lifecycles with on-chain Ethereum Virtual Machine (EVM) transactions and QBFT consensus blocks.

## What You Are Building
A complete architectural specification (`docs/architecture/distributed_tracing_and_opentelemetry.md`), reusable cross-language OpenTelemetry SDK libraries, collector deployment manifests, and storage schemas providing microsecond-to-nanosecond observability:
- **Standardized Polyglot OpenTelemetry SDKs:** Unified tracing initialization, lifecycle management, and span context management for Go (`go.opentelemetry.io/otel`), Rust (`tracing`, `tracing-opentelemetry`), Python (`opentelemetry-api`), and Flutter/Dart (`opentelemetry`).
- **W3C TraceContext & Baggage Propagation:** Wire-level injection and extraction standards for `traceparent` and `tracestate` across HTTP/1.1, HTTP/2 gRPC metadata, WebSocket connection parameters, and Apache Kafka record headers.
- **Zero-Allocation Rust Matching Engine Tracer:** Asynchronous, lock-free ring-buffer trace recording using thread-local cycle counters (TSC / RDTSC) to profile the sub-100µs matching engine core loop without lock contention or context switching penalties.
- **OpenTelemetry Collector Pipeline:** Two-tier collector architecture (HostPort DaemonSet agents + centralized scalable aggregators) featuring tail-based sampling, attribute scrubbing, and batch compression.
- **Trace Visualization & Analytical Storage:** Dual-backend storage utilizing Grafana Tempo (object-store-backed trace waterfall visualization) and ClickHouse (high-cardinality multi-dimensional span search and latency percentiles).
- **Blockchain Trace Correlation Engine:** Cryptographic correlation framework associating off-chain order execution traces with Hyperledger Besu EIP-712 authorization hashes, JSON-RPC relay requests, smart contract execution spans, and QBFT block commits.

## Scope Boundaries
- **In Scope:**
  - Standardized W3C TraceContext (`traceparent`, `tracestate`) injection and extraction rules across HTTP, gRPC, WebSockets, and Kafka.
  - Core OpenTelemetry SDK wrapper packages for Go, Rust, Python, and Flutter.
  - Sub-microsecond non-blocking instrumentation strategy for the Rust order matching engine.
  - OpenTelemetry Collector DaemonSet and Aggregator architectures with tail-based sampling rules.
  - Blockchain transaction correlation for Hyperledger Besu (EIP-712 signing, JSON-RPC relaying, QBFT consensus block tracing).
  - PII sanitization, data redaction, and high-cardinality attribute scrubbers complying with SEBI CSCRF and DPDP Act.
  - ClickHouse span metadata schema and Grafana Tempo storage integration.
  - Canonical span naming conventions, semantic conventions, and latency SLA budgets.
- **Out of Scope / Handled Elsewhere:**
  - Underlying Kubernetes infrastructure and MinIO/S3 object storage deployment (handled in Prompts 401 & 402).
  - Prometheus metrics instrumentation and alert rule management (handled in Prompt 803 - System Monitoring & Prometheus Metrics).
  - Centralized structured logging and Loki log aggregation (handled in Prompt 804 - Centralized Logging & Loki).
  - Synthetic end-to-end user experience probes and blackbox ping monitors (handled in Prompt 807).

## Technology to Use
- **Core Tracing Protocols:** OpenTelemetry Protocol (OTLP/gRPC on port 4317, OTLP/HTTP on port 4318), W3C TraceContext Recommendation (Level 1 / RFC 7230), W3C Baggage Specification.
- **Instrumentation SDKs & Frameworks:**
  - *Go:* `go.opentelemetry.io/otel` v1.28+, `go.opentelemetry.io/otel/trace`, `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc`, `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`.
  - *Rust:* `tracing` 0.1+, `tracing-subscriber` 0.3+, `tracing-opentelemetry` 0.25+, `opentelemetry` 0.24+, `opentelemetry-otlp` 0.24+ (with `tonic` transport), hardware timestamping via `quanta` or `minstant` (x86_64 invariant RDTSC).
  - *Python:* `opentelemetry-api` 1.25+, `opentelemetry-sdk`, `opentelemetry-instrumentation-fastapi`, `opentelemetry-instrumentation-sqlalchemy`.
  - *Flutter / Dart:* `opentelemetry` package with custom `http.BaseClient` interceptor and W3C header injection.
- **Collector Infrastructure:** OpenTelemetry Collector Contrib v0.104+ configured as Kubernetes HostPort DaemonSets forwarding to a scalable aggregrator deployment.
- **Trace Backends & Visualization:**
  - *Grafana Tempo 2.5+:* Distributed trace storage with Parquet columnar blocks on S3-compatible object storage.
  - *ClickHouse 24.3+:* Structured analytical columnar engine for multi-attribute span search, latency aggregations, and p99.9 anomaly profiling.
  - *Jaeger UI / Grafana Explore:* Interactive waterfall visualization, dependency graphing, and critical-path latency analysis.

## Backend / Infra Touchpoints
- **Envoy Edge API Gateway:** Evaluates incoming client requests; injects canonical W3C `traceparent` if absent; records TLS termination latency, edge routing duration, and downstream cluster dispatch spans.
- **Microservices Layer (Go / Python):** Extracts incoming `traceparent` via gRPC metadata or HTTP headers; establishes parent-child span hierarchy; attaches business context (order ID, symbol, entity tag); records internal database (PostgreSQL/Redis) execution spans.
- **Rust Order Matching Engine:** Receives binary orders via gRPC or shared ring buffer; utilizes zero-alloc thread-local span IDs; emits asynchronous trace events over an isolated lock-free ring buffer channel to an off-thread OTLP batch processor.
- **Apache Kafka Message Brokers:** Message producers inject `traceparent` and `tracestate` into Kafka `RecordHeaders`; consumers extract header context to instantiate continuous consumer spans without mutating JSON/Protobuf message payloads.
- **Kubernetes Node DaemonSets:** Local OpenTelemetry Collector agents bound to `nodeIP:4317` / `127.0.0.1:4317`, minimizing network latency and offloading batching, PII scrubbing, and gzip compression from application pods.
- **ClickHouse & Grafana Tempo:** Ingestion pipeline writes trace metadata into ClickHouse `telemetry_spans` tables for sub-second index filtering and dumps full trace payloads into Tempo object storage blocks.

## Blockchain Interaction
Establishes bi-directional distributed trace correlation between off-chain microservices and permissioned Hyperledger Besu smart contracts:
- **EIP-712 Structured Data Signing Spans:** Instruments client-side and backend relayer signing pipelines. Emits span `eth.eip712.sign` capturing domain separator calculation, signature generation, and verification timing, tagging attributes `eth.eip712.domain_hash` and `eth.eip712.primary_type`.
- **JSON-RPC Relayer Spans:** Wraps Web3 JSON-RPC calls made by the Go Trade Settlement Service (`eth_sendRawTransaction`, `eth_getTransactionReceipt`, `eth_call`) with `rpc.system=ethereum`, `rpc.method=eth_sendRawTransaction`, and `eth.tx_hash`.
- **QBFT Consensus & Block Commit Correlation:**
  - When the Settlement Service submits a settlement transaction (`executeDvP`), the active trace context (`trace_id`, `span_id`) is stored in the off-chain settlement persistence record alongside the transaction hash.
  - When an asynchronous block listener or WebSocket event subscriber receives the on-chain confirmation, it retrieves the parent trace context and starts a child span:
    - Span Name: `besu.qbft.block_commit`
    - Attributes: `besu.block_number`, `besu.block_hash`, `besu.tx_index`, `besu.gas_used`, `besu.qbft.round`, `eth.contract_address`, `besu.validator_address`.
- **Smart Contract DvP Event Linking:** On-chain events emitted by `SettlementDvP.sol` (`DvPExecuted`, `TokensMinted`) carry the deterministic `settlement_id` / `trade_id`, enabling complete reconstruction of the order lifecycle from mobile tap to on-chain finality in a single unified trace view.

## Step-by-Step Build Instructions
1. Author canonical architecture specification `docs/architecture/distributed_tracing_and_opentelemetry.md` establishing trace propagation semantics, span taxonomy, latency SLA budgets, and storage topologies.
2. Formulate the W3C TraceContext and Baggage propagation standard:
   - Wire header: `traceparent: {version}-{trace_id}-{parent_id}-{trace_flags}` (e.g., `00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01`).
   - Standardize cross-boundary propagation rules across HTTP headers, gRPC metadata context, Kafka message headers, and WebSocket query parameters.
3. Build the core Go tracing library (`pkg/telemetry/tracer.go`):
   - Configure OpenTelemetry TracerProvider with OTLP gRPC exporter.
   - Attach mandatory semantic resource attributes: `service.name`, `service.version`, `service.namespace`, `deployment.environment`, `host.name`.
   - Implement gRPC unary and streaming interceptors (`pkg/telemetry/grpc_interceptor.go`).
4. Implement the zero-allocation Rust tracing engine crate (`crates/growww-tracing`):
   - Wrap `tracing` and `tracing-opentelemetry` with thread-local cycle counting (`quanta` / `minstant`).
   - Construct a lock-free ring-buffer channel (`crossbeam-channel` or `rtrb`) decoupling the matching engine's sub-100µs hot execution loop from network OTLP export threads.
   - Guarantee <= 2 microseconds maximum overhead per matching event.
5. Build the Python FastAPI & SQLAlchemy tracing package (`common/telemetry/otel.py`):
   - Integrate `FastAPIInstrumentor` and `SQLAlchemyInstrumentor`.
   - Implement automated SQL parameter scrubbing to prevent PII exposure in trace span attributes.
6. Develop the Apache Kafka TraceContext Carrier (`pkg/telemetry/kafka_carrier.go`):
   - Implement the `propagation.TextMapCarrier` interface over Kafka `[]kafka.Header`.
   - Inject `traceparent` and `tracestate` on message produce; extract and instantiate child spans on message consume.
7. Configure Envoy Gateway tracing filter (`deploy/envoy/envoy.yaml`):
   - Enable `envoy.tracers.opentelemetry` with OTLP gRPC collector endpoint.
   - Configure root trace ID generation (UUIDv7-compatible 128-bit hex format) when external requests arrive without valid `traceparent`.
   - Record ingress edge metrics: TLS negotiation duration, downstream connection reuse, and upstream cluster latency.
8. Construct Kubernetes OpenTelemetry Collector DaemonSet (`deploy/telemetry/otel-collector-daemonset.yaml`):
   - Configure `otlp` receivers on `0.0.0.0:4317` (gRPC) and `0.0.0.0:4318` (HTTP).
   - Configure `memory_limiter`, `batch`, and `k8sattributes` processors to enrich spans with pod, namespace, and container metadata.
9. Implement Tail-Based Sampling in the central OTel Aggregator (`deploy/telemetry/otel-collector-aggregator.yaml`):
   - 100% sampling for spans with HTTP status code >= 500 or gRPC status != `OK`.
   - 100% sampling for spans exhibiting matching engine duration >= 50ms or order execution latency >= 500ms.
   - 100% sampling for all blockchain settlement, DvP execution, and wallet debit/credit spans.
   - Deterministic 5% probabilistic hash-based sampling for routine healthy traffic.
10. Implement High-Cardinality & PII Attribute Scrubbing Processor:
    - Configure collector regex attribute processors masking Indian financial identifiers: PAN (`[A-Z]{5}[0-9]{4}[A-Z]{1}`), Aadhaar (12 digits), bank account numbers, IFSC codes, and private keys.
    - Transform sensitive user IDs into cryptographically salted hashes (`user.id.hash`).
11. Instrument Hyperledger Besu relayer client and event listeners (`pkg/blockchain/tracer.go`):
    - Create tracing wrappers for Web3 JSON-RPC client calls.
    - Instrument block listener workers to emit `besu.qbft.block_commit` spans linked via parent trace context to the initiating DvP trade settlement.
12. Design ClickHouse Analytical Span Schema and ingestion pipeline (`deploy/clickhouse/migrations/001_telemetry_spans.sql`):
    - Table `telemetry_spans`: `TraceId`, `SpanId`, `ParentSpanId`, `TraceState`, `SpanName`, `SpanKind`, `ServiceName`, `ResourceAttributes`, `SpanAttributes`, `DurationNanos`, `StatusCode`, `Timestamp`.
    - Apply `LowCardinality` dictionary encoding and Bloom filter indexes for sub-second span filtering across billions of records.
13. Configure Grafana Tempo backend (`deploy/tempo/tempo.yaml`):
    - Configure Parquet columnar block storage on MinIO/S3.
    - Configure Tempo-to-ClickHouse and Tempo-to-Loki trace-to-logs cross-navigation links.
14. Construct automated distributed trace end-to-end integration tests:
    - Execute end-to-end test placing an order through mock Flutter client -> Envoy -> Risk Check -> Rust Matching Engine -> Kafka -> Settlement -> Besu mock node.
    - Verify unbroken span DAG (Directed Acyclic Graph) containing continuous trace IDs across all 6 hops.
15. Build Grafana Distributed Tracing Dashboard (`deploy/dashboards/trace_waterfall.json`):
    - Visualize service dependency topology, latency waterfall breakdown, p50/p90/p99 latency percentiles, and SEBI CSCRF audit trail traces.

## Interfaces / Contracts

### Protobuf Trace Context Definition (`proto/telemetry/v1/trace_context.proto`)
```protobuf
syntax = "proto3";

package growww.telemetry.v1;

option go_package = "github.com/growww/platform/gen/go/telemetry/v1;telemetryv1";

// Distributed trace context carrier for internal gRPC headers and binary serialization
message TraceContext {
  // W3C compliant 128-bit trace ID represented as 32 hex characters
  string trace_id = 1;
  
  // W3C compliant 64-bit parent span ID represented as 16 hex characters
  string span_id = 2;
  
  // 8-bit trace flags represented as 2 hex characters (e.g., '01' for sampled)
  string trace_flags = 3;
  
  // Vendor-specific state key-value pairs per W3C specification
  string trace_state = 4;
  
  // Distributed baggage key-value pairs propagating business context
  map<string, string> baggage = 5;
}

// Latency profile checkpoint emitted by ultra-low-latency engines
message LatencyProfileEvent {
  string trace_id = 1;
  string span_id = 2;
  string component_name = 3;
  string event_name = 4;
  uint64 timestamp_unix_nanos = 5;
  uint64 cpu_cycle_counter = 6;
  map<string, string> attributes = 7;
}
```

### W3C HTTP & gRPC Metadata Schema
- **HTTP Header Name:** `traceparent`
  - Format: `{version}-{trace_id}-{parent_id}-{trace_flags}`
  - Regex: `^00-[0-9a-f]{32}-[0-9a-f]{16}-[0-9a-f]{2}$`
  - Example: `00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01`
- **HTTP Baggage Header:** `baggage`
  - Format: `key1=value1,key2=value2`
  - Example: `user.entity=domestic,order.type=LIMIT,market=NSE`
- **gRPC Metadata Keys (Binary and Text):**
  - Text metadata: `traceparent` (string), `tracestate` (string)
  - Inter-service metadata propagation follows standard gRPC incoming/outgoing context maps.

### Apache Kafka RecordHeader Wire Format
Tracing headers injected into each Kafka record without modifying message payload:
| Header Key | Format / Encoding | Description |
| :--- | :--- | :--- |
| `traceparent` | UTF-8 String (ASCII) | Full W3C traceparent string |
| `tracestate` | UTF-8 String (ASCII) | Optional W3C tracestate string |
| `x-growww-span-origin` | UTF-8 String (ASCII) | Emitting microservice name and version |

### ClickHouse Span Storage DDL (`telemetry_spans`)
```sql
CREATE TABLE IF NOT EXISTS telemetry_spans (
    Timestamp DateTime64(9, 'UTC') CODEC(DoubleDelta, ZSTD(1)),
    TraceId FixedString(32) CODEC(ZSTD(1)),
    SpanId FixedString(16) CODEC(ZSTD(1)),
    ParentSpanId FixedString(16) CODEC(ZSTD(1)),
    TraceState String CODEC(ZSTD(1)),
    SpanName LowCardinality(String) CODEC(ZSTD(1)),
    SpanKind LowCardinality(String) CODEC(ZSTD(1)),
    ServiceName LowCardinality(String) CODEC(ZSTD(1)),
    DurationNanos UInt64 CODEC(T64, ZSTD(1)),
    StatusCode LowCardinality(String) CODEC(ZSTD(1)),
    StatusMessage String CODEC(ZSTD(1)),
    ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    SpanAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    Events Nested (
        Timestamp DateTime64(9, 'UTC'),
        Name LowCardinality(String),
        Attributes Map(LowCardinality(String), String)
    ) CODEC(ZSTD(1)),
    INDEX idx_trace_id TraceId TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_span_name SpanName TYPE set(100) GRANULARITY 1,
    INDEX idx_duration DurationNanos TYPE minmax GRANULARITY 1
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMMDD(Timestamp)
ORDER BY (ServiceName, SpanName, StatusCode, toUnixTimestamp(Timestamp), TraceId, SpanId)
TTL toDateTime(Timestamp) + INTERVAL 30 DAY DELETE;
```

### Canonical Span Tagging Dictionary
| Attribute Key | Type | Description / Example | Allowed Values / Constraint |
| :--- | :--- | :--- | :--- |
| `service.name` | String | Standard microservice identifier | `order-service`, `risk-engine`, `matching-engine` |
| `growww.entity` | String | Legal entity processing trade | `domestic_in`, `gift_city_ifsc` |
| `growww.order_id` | String | Unique Order UUIDv7 | `ord_01HZX89AB72K9M12P5QRSTUVWX` |
| `growww.trade_id` | String | Executed trade identifier | `trd_01HZX90BC83L0N23Q6RSTUVWXY` |
| `growww.symbol` | String | Security trading symbol | `RELIANCE`, `TCS`, `INFY` |
| `growww.market` | String | Execution exchange / ledger | `NSE`, `BSE`, `BESU_PRIVATE` |
| `eth.tx_hash` | String | Blockchain transaction hash | `0x7f9a2b8...` (66-char hex) |
| `besu.block_number` | UInt64 | Confirmed block number | `18459203` |
| `besu.qbft.round` | UInt32 | QBFT consensus commit round | `0` |
| `matching.queue_time_us` | Float64 | Ring buffer queuing duration | Latency in microseconds |
| `matching.exec_time_us` | Float64 | Pure order book match latency | Latency in microseconds |

## Security & Compliance Notes
- **SEBI CSCRF (Cybersecurity and Cyber Resilience Framework) Compliance:** Distributed traces serve as an immutable forensic audit log. Spans capturing user-initiated trades must link the originating client IP (hashed via HMAC-SHA-256), session identifier, order ID, exchange execution ID, and blockchain transaction hash with nanosecond-synchronized timestamps (PTP / NTP stratum-1).
- **DPDP Act & GDPR Zero-PII Enforcement:** Under no circumstances may raw Personally Identifiable Information (PII) appear in span names, attributes, or baggage. Prohibited fields include: Permanent Account Number (PAN), Aadhaar number, bank account details, investor real names, passwords, and private keys. All sensitive identifiers must be sanitized or pseudonymized at the SDK interceptor layer prior to export.
- **High-Cardinality Attribute Scrubbing:** Unbounded parameters (e.g., arbitrary user text search queries, raw request bodies, unindexed timestamps in tag keys) must be stripped by the OpenTelemetry Collector attribute processors to protect ClickHouse and Tempo from memory exhaustion.
- **mTLS OTLP Telemetry Transport:** All span payloads transmitted from application pods to local Node DaemonSets, and from DaemonSets to the central Collector cluster, must use Mutual TLS (mTLS) with SPIFFE/SPIRE x509 workload identities or internal CA certificates, preventing eavesdropping or injection of fraudulent telemetry data.
- **Trace Sampling Audit Integrity:** While probabilistic sampling (5%) applies to standard read-only traffic, financial state-changing transactions (orders, wallet debits, DvP settlements, smart contract commits) and all error states (HTTP 5xx, gRPC error codes) must enforce 100% deterministic capture for regulatory compliance.

## Acceptance Criteria
- [ ] Published comprehensive architecture document `docs/architecture/distributed_tracing_and_opentelemetry.md` detailing end-to-end trace topologies, propagation protocols, and sampling rules.
- [ ] Go, Rust, and Python OpenTelemetry SDK packages implemented with automated W3C `traceparent` injection/extraction interceptors.
- [ ] Rust matching engine instrumentation verified to introduce <= 2 microseconds overhead on the critical order book matching loop through asynchronous ring-buffer export.
- [ ] Apache Kafka producer and consumer interceptors successfully propagate `traceparent` across topics without payload mutation.
- [ ] Hyperledger Besu relayer client and block listener workers emit spans linking off-chain trade settlement records with on-chain QBFT transaction hashes and block numbers.
- [ ] OpenTelemetry Collector DaemonSet and Aggregator configurations deployed with active PII scrubbing and tail-based sampling rules.
- [ ] ClickHouse `telemetry_spans` schema and Grafana Tempo backend deployed and verified to perform multi-attribute span searches in < 1 second across 100M+ spans.
- [ ] End-to-end integration test verifies an unbroken span DAG linking mobile client -> Envoy Gateway -> Risk Service -> Rust Engine -> Kafka -> Settlement Service -> Besu Ledger.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture Overview), Prompt 103 (API Design Standards), Prompt 104 (Event Schema & Kafka Topic Standards), Prompt 107 (Polyglot Coding Standards).
- **Parallel Work:** Prompt 105 (Authentication & Authorization Architecture), Prompt 108 (Configuration Management), Prompt 112 (Idempotency & Exactly-Once Processing).
- **Blocks:** Prompt 204 (Order Service), Prompt 205 (Order Matching Engine), Prompt 206 (Risk & Margin Checks), Prompt 207 (Market Data Service), Prompt 208 (Trade Settlement Service), Prompt 306 (Settlement DvP Smart Contract), Prompt 803 (System Monitoring & Prometheus Metrics), Prompt 804 (Centralized Logging & Loki).
