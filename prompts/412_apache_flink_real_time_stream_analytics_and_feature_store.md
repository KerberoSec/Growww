# 412 - Apache Flink Real-Time Stream Analytics & Online Feature Store

## Purpose
In high-frequency 24/7 multi-asset digital exchanges and regulated equity markets, static batch processing and database polling mechanisms are inadequate for managing market risk, dynamic circuit bands, and abusive trading behavior. Peak market trading sessions generate hundreds of thousands of order cancellations, price depth modifications, and trade executions per second across equities, commodities, and tokenized assets. Downstream decision systems such as pre-trade risk gates, market surveillance engines, and value-at-risk calculators require stateful analytical aggregations computed at sub-millisecond latencies over continuous event streams.

The **Apache Flink Real-Time Stream Analytics & Online Feature Store** service (`services/flink-stream-analytics`) serves as the ultra-low-latency stateful stream processing and feature computation backbone for the platform. Operating natively on event-time streams with out-of-order event handling and stateful sliding windows, the engine continuously calculates dynamic volatility indicators, order book depth imbalances, participant behavioral features, and surveillance aggregates. The engine sinks computed feature vectors into a high-performance Feast online feature store backed by Redis, enabling downstream microservices, machine learning models, and blockchain oracle relayers to retrieve real-time state vectors in sub-millisecond lookups.

## What You Are Building
A production-grade, distributed stream processing platform and online feature store integration comprising:
- **Flink Stateful Stream Processing Pipelines (`services/flink-stream-analytics`):** Highly optimized Apache Flink 1.19+ streaming applications implemented in Java / Scala, managed natively on Kubernetes via the Flink Kubernetes Operator.
- **Embedded RocksDB Keyed State Engine:** Enterprise-grade RocksDB state backend configurations managing gigabytes of partitioned operational state across TaskManagers with incremental checkpointing to S3-compatible object storage.
- **Dynamic Volatility & Microstructure Job Graph:** High-frequency stream computation job ingesting trade fills (`matching.matches.v1`) and tick quotes to calculate rolling VWAP, Garman-Klass volatility, Parkinson volatility, and Order Book Imbalance (OBI) across multi-duration sliding windows (5s, 1m, 5m, 15m).
- **Trader Behavioral Feature Generation Job Graph:** Stateful streaming pipeline keying on participant identifiers to calculate real-time Order-to-Trade Ratios (OTR), aggressive cancellation velocities, order fill ratios, and quote stuffing burst counters.
- **Market Surveillance Aggregate Detector:** Real-time stream processor detecting market manipulation anomalies including wash trading loops, layering patterns, and momentum ignition bursts, emitting alert streams to the Market Surveillance Engine (Prompt 228).
- **Feast Online Feature Store with Redis Cluster:** Unified feature repository declaring streaming feature views, schema contracts, and direct high-throughput sink writers into Redis 7.2 Cluster with Protobuf serialization and TTL-based eviction.
- **Besu Oracle Relayer Bridge:** Streaming publisher pushing authenticated market pricing benchmarks, aggregate volatility metrics, and liquidation risk indices to off-chain relayers feeding `OracleAggregator.sol` on Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
  - Apache Flink 1.19+ streaming job graphs (Market Microstructure Pipeline, Trader Behavioral Pipeline, Surveillance Pipeline).
  - Event-time processing, watermark generation strategies (bounded out-of-orderness), and late-event side-outputs.
  - RocksDB state backend tuning: off-heap memory budgets, block cache allocations, write buffer configurations, and incremental checkpointing.
  - Integration with Feast 0.38+ online feature store and low-latency Redis 7.2 Cluster sink writers.
  - Exact mathematical formulas for streaming metrics: Garman-Klass, Parkinson, rolling EWMA volatility, OTR, OBI, and volume-weighted aggregations.
  - End-to-end exactly-once processing guarantees across Kafka sources and transactional Kafka / Redis sinks.
  - Protobuf v3 contracts for feature vectors and downstream event topics.
  - SEBI CSCRF compliance: deterministic state checkpointing, tamper-evident audit trails, and state replay capabilities.
- **Out of Scope / Handled Elsewhere:**
  - In-memory order book FIFO matching and execution sequencing (handled in Prompt 205).
  - Pre-trade single-order credit limit and static price band validation (handled in Prompt 206).
  - Post-trade clearing and atomic Delivery-versus-Payment (DvP) settlement (handled in Prompt 208).
  - Long-term historical OLAP data warehousing and multi-year backtesting (handled in Prompt 404 and Prompt 408).
  - On-chain smart contract execution and consensus validation (handled in Prompt 328).
  - Regulatory compliance report PDF generation and filing dispatchers (handled in Prompt 216).

## Technology to Use
- **Stream Processing Framework:** **Apache Flink 1.19+** (Java 17 / 21 LTS). Flink provides true record-by-record streaming, millisecond processing latencies, rich event-time windowing semantics, and robust state fault tolerance.
- **State Storage Backend:** **Embedded RocksDB State Backend (`EmbeddedRocksDBStateBackend`)** with asynchronous incremental checkpointing, off-heap memory management via jemalloc, and Bloom filters for point lookups.
- **Orchestration & Deployment:** **Flink Kubernetes Operator 1.8+** running on Kubernetes clusters with dynamic pod autoscaling, high-availability ZooKeeper/Kubernetes lease coordinators, and reactive scaling.
- **Message Broker:** **Apache Kafka 3.7+** operating under KRaft consensus, using `KafkaSource` and transactional `KafkaSink` with two-phase commit (2PC) for end-to-end exactly-once guarantees.
- **Online Feature Store:** **Feast 0.38+** paired with **Redis 7.2+ Cluster** utilizing pipelined multiplexed writes, binary Protobuf value encoding, and configured TTL expiry policies.
- **Distributed Checkpoint Store:** **AWS S3 / MinIO** with S3 Object Lock and multi-part transactional upload support for durable checkpoint and savepoint storage.
- **Data Serialization:** **Protocol Buffers v3** for binary wire encoding of feature vectors and events; Flink TypeInformation and TypeSerializers optimized for Protobuf POJOs.
- **Observability:** **Prometheus & Grafana** scraping Flink TaskManager metrics via Flink Prometheus Reporter, monitoring checkpoint durations, backpressure ratios, watermark lag, and RocksDB cache hit rates.

## Backend / Infra Touchpoints
- **Apache Kafka Topics:**
  - Ingress: `matching.matches.v1`, `orders.events.v1`, `market.ticks.v1`, `market.depth.v1`.
  - Egress: `features.market_volatility.v1`, `features.trader_behavior.v1`, `surveillance.aggregates.v1`, `oracle.aggregates.v1`.
- **Market Surveillance Engine (Prompt 228):** Consumes `surveillance.aggregates.v1` and queries Redis feature store to execute deep fraud investigations and pattern classification.
- **Real-Time VaR Engine (Prompt 229):** Consumes `features.market_volatility.v1` to update intra-day EWMA covariance matrices and dynamic Extreme Loss Margins (ELM).
- **Pre-Trade Risk & Margin Engine (Prompt 206):** Directly queries Redis online feature store for real-time trader behavioral features (OTR, burst rates) during pre-order checks.
- **Redis 7.2 Cluster:** Online store hosting key namespaces:
  - `feat:trader:{user_id}`: Active trader behavioral vectors.
  - `feat:market:{isin}`: Real-time price volatility and liquidity metrics.
  - `feat:microstructure:{isin}`: Rolling order book depth imbalance indicators.
- **Historical TimescaleDB / ClickHouse Pipeline (Prompt 408):** Ingests rolling multi-window candle snapshots emitted by Flink for analytical persistence.
- **AWS S3 / MinIO:** Object bucket `s3://growww-flink-checkpoints/` containing incremental state rocksdb snapshots and operator savepoints.

## Blockchain Interaction
The stream processing platform interfaces directly with the permissioned Hyperledger Besu consortium ledger (QBFT consensus, 2-second block finality, 1:1 asset backing, zero PII) to provide continuous cryptographic and market integrity guarantees:

### Detailed On-Chain Integration Mechanics:
- **Streaming Risk & Price Aggregates to Besu Oracle Relayers:** The Flink engine continuously calculates Volume Weighted Average Price (VWAP) benchmarks and dynamic volatility corridors across rolling windows. High-frequency price and volatility snapshots are published to Kafka topic `oracle.aggregates.v1`. The Decentralized Oracle Aggregator Relayer (Prompt 328) consumes this topic, batches cryptographically signed price feeds using AWS CloudHSM / Vault keys, and submits oracle updates to `OracleAggregator.sol` on Hyperledger Besu.
- **Dynamic On-Chain Collateral & Haircut Parameters:** Hyperledger Besu smart contracts governing tokenized collateral liquidations and synthetic margin mechanisms (`PerpetualsVault.sol`, `SettlementDvP.sol`) consume the streaming volatility indices produced by Flink. During extreme market turbulence, on-chain haircuts and liquidation penalties are automatically scaled based on the authenticated volatility metrics relayed from Flink.
- **Zero-PII On-Chain Policy:** All event streams, state keys, and egress feature vectors processed by Flink utilize pseudonymous internal identifiers (`user_id` UUIDs, instrument ISINs, ledger addresses `0x...`). No investor names, tax identifiers (PAN/Aadhaar), bank account numbers, or plain identity attributes are ingested into stream state or emitted to blockchain oracle feeds.

## Stream Processing Architecture & Feature Store Mechanics

### 1. Flink Streaming Topology & Job Graph Architecture
The Flink processing cluster executes three decoupled streaming job graphs to ensure computational isolation and independent scaling:

```
[Kafka Ingress: matching.matches.v1 / market.depth.v1 / orders.events.v1]
                               |
                               v
            +------------------------------------+
            | Flink KafkaSource (Event-Time)     |
            | Watermarks: BoundedOutOfOrderness  |
            +------------------------------------+
                               |
        +----------------------+----------------------+
        |                                             |
        v                                             v
+-------------------------------+             +-------------------------------+
| Market Microstructure Job     |             | Trader Behavioral Job         |
| - KeyBy(isin)                 |             | - KeyBy(user_id)              |
| - Sliding Windows (1m, 5m)    |             | - Sliding Windows (10s, 60s)  |
| - RocksDB State: Price/Volume |             | - RocksDB State: Order/Cancel |
| - Compute: GK Vol, OBI, VWAP  |             | - Compute: OTR, Burst Rate    |
+-------------------------------+             +-------------------------------+
        |                                             |
        +----------------------+----------------------+
                               |
                               v
            +------------------------------------+
            | Market Surveillance Aggregator     |
            | - Pattern Match (CEP) & Cross-Key  |
            | - Wash Trade & Layering Detection  |
            +------------------------------------+
                               |
        +----------------------+----------------------+
        |                                             |
        v                                             v
+-------------------------------+             +-------------------------------+
| Redis Online Store Sink       |             | Kafka Transactional Sink      |
| - Feast Feature Format        |             | - Exact-Once 2PC Delivery     |
| - Pipeline Batch Writes       |             | - Egress: features.*.v1       |
+-------------------------------+             +-------------------------------+
```

### 2. Streaming Volatility & Market Microstructure Algorithms
The Market Microstructure Job executes mathematically rigorous streaming calculations over high-velocity order and trade events:

- **Rolling Volume-Weighted Average Price (VWAP):**
  Computed over a sliding window of length $W$ with event timestamp $t_i$:
  $$\text{VWAP}_W = \frac{\sum_{i \in W} P_i \cdot V_i}{\sum_{i \in W} V_i}$$
  Maintained in RocksDB keyed state as an accumulator tuple `(sum_price_volume, sum_volume)`.

- **Garman-Klass Volatility Indicator:**
  Measures intraday price volatility utilizing open, high, low, and close prices over sliding windows:
  $$\sigma_{GK}^2 = \frac{1}{N} \sum_{i=1}^N \left[ 0.5 \left( \ln\frac{H_i}{L_i} \right)^2 - (2\ln 2 - 1) \left( \ln\frac{C_i}{O_i} \right)^2 \right]$$
  Where $H_i$, $L_i$, $O_i$, $C_i$ represent highest, lowest, opening, and closing trade prices in sub-window slice $i$.

- **Parkinson High-Low Volatility:**
  Estimates asset variance based purely on extreme price movements:
  $$\sigma_P^2 = \frac{1}{4 \ln 2 \cdot N} \sum_{i=1}^N \left( \ln\frac{H_i}{L_i} \right)^2$$

- **Order Book Imbalance (OBI):**
  Evaluates short-term buying/selling pressure from Level-2 depth streams:
  $$\text{OBI}_t = \frac{V_t^{\text{bid}} - V_t^{\text{ask}}}{V_t^{\text{bid}} + V_t^{\text{ask}}}$$
  Where $V_t^{\text{bid}}$ and $V_t^{\text{ask}}$ represent cumulative volume across the top 5 price levels.

### 3. Trader Behavioral Feature Computation
The Trader Behavioral Job aggregates orders and executions keyed by `user_id` to build operational risk features:

- **Order-to-Trade Ratio (OTR):**
  Calculated across rolling 60-second windows to detect quote stuffing and algorithmic order spam:
  $$\text{OTR} = \frac{N_{\text{new\_orders}} + N_{\text{modifications}} + N_{\text{cancellations}}}{\max(1, N_{\text{executed\_trades}})}$$
  If $\text{OTR} > \theta_{\text{regulatory}}$ (e.g. 500:1) during active market periods, an alert flag is asserted.

- **Order Cancellation Latency & Aggressive Ratio:**
  Tracks the median duration between order creation and cancellation. Orders cancelled within $< 10\text{ms}$ of placement contribute to the high-frequency spoofing score.

### 4. Feast Online Feature Store Architecture
The online feature store provides sub-millisecond point lookups for pre-trade risk and surveillance models:
- **Entity Definitions:** Primary entities declared in Feast are `trader` (`user_id`) and `instrument` (`isin`).
- **Feature Views:** Stream feature views consume from Kafka egress topics and map incoming Protobuf structures directly to Redis hash keys.
- **Key Serialization Scheme:**
  - Redis Key Format: `growww:feast:{entity_type}:{entity_id}`
  - Redis Field Format: Hash fields containing individual feature values or serialized binary Protobuf feature payloads.
- **TTL Strategy:** Real-time feature hashes are configured with sliding TTLs (e.g. 86400 seconds for trader profiles, 3600 seconds for intraday volatility) to prevent unconstrained memory expansion.

### 5. Exactly-Once State Semantics & RocksDB Checkpointing
- **Checkpoint Algorithm:** Flink utilizes the Chandy-Lamport variant (asynchronous barrier snapshotting). Checkpoint barriers are injected into Kafka source partitions and flow through the operator graph.
- **RocksDB State Backend Tuning:**
  - State writes are handled off-heap in RocksDB memtables.
  - Periodic flushing creates SST files on local NVMe disk.
  - Checkpoint creation utilizes hard links and asynchronous upload of newly created SST files to S3, achieving incremental checkpoint completion times $< 1000\text{ms}$.
- **Kafka Two-Phase Commit:** Egress Kafka sinks utilize `DeliveryGuarantee.EXACTLY_ONCE` with Kafka transactions, committing offsets only when Flink confirms checkpoint completion.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project Repository:** Scaffold Maven / Gradle multi-module project `services/flink-stream-analytics` containing modules: `common-contracts`, `market-analytics-job`, `trader-analytics-job`, `surveillance-job`, and `feature-store-connector`.
2. **Define Protocol Buffer Schemas:** Author Protobuf schemas for market ticks, trade matches, trader actions, market feature vectors, and trader behavioral vectors in `common-contracts`.
3. **Configure Flink Runtime & Dependencies:** Configure `pom.xml` with Apache Flink 1.19 dependencies, `flink-connector-kafka`, `flink-statebackend-rocksdb`, Protobuf serializer libraries, and JUnit 5 test harnesses.
4. **Implement Event-Time Watermark Generators:** Build customized `WatermarkStrategy` using `BoundedOutOfOrderness` (max delay 200ms) with idempotent timestamp extractors reading event headers.
5. **Implement Market Microstructure Streaming Job:**
   - Create Flink `DataStream` reading `matching.matches.v1` and `market.depth.v1`.
   - Implement `KeyedProcessFunction` keyed by ISIN maintaining rolling accumulators in RocksDB `ValueState` and `ListState`.
   - Implement sliding window aggregators (5s slide, 1m size and 15m size) computing VWAP, Garman-Klass, and Parkinson volatility.
6. **Implement Order Book Imbalance (OBI) Operator:** Process Level-2 order book depth events to compute instantaneous and 5-second moving average OBI indicators.
7. **Implement Trader Behavioral Streaming Job:**
   - Consume order lifecycle events (`orders.events.v1`).
   - Key stream by `user_id` and compute rolling 10-second and 60-second OTR counters, cancellation speeds, and burst intensities.
8. **Implement Complex Event Processing (CEP) for Surveillance:** Utilize Flink CEP library to detect suspicious multi-event patterns: order placement followed by cancellation within 20ms and counter-fill on opposite side (spoofing).
9. **Configure RocksDB State Backend:** Configure `EmbeddedRocksDBStateBackend` with incremental checkpointing enabled, block cache configured to 2GB off-heap, and Bloom filters enabled for all column families.
10. **Implement Redis Pipeline Feature Store Sink:** Build custom `RichSinkFunction` or use Redis async client (Lettuce) to batch feature vector updates into Redis Cluster using pipelined writes with configurable micro-batch intervals (50ms).
11. **Configure Feast Feature Repository:** Define `feature_store.yaml`, declare `Entity` definitions, `BatchFeatureView`, and `StreamFeatureView` resources linking Kafka topics and Redis online storage.
12. **Configure Kafka Exactly-Once Egress Sink:** Set up `KafkaSink` with `DeliveryGuarantee.EXACTLY_ONCE`, transactional ID prefix, and two-phase commit coordinators for risk and oracle topics.
13. **Write Unit & State Machine Tests:** Implement integration tests using `MiniClusterWithClientExtension` and `TestHarness` to verify window calculations, out-of-order event handling, and RocksDB state restoration from savepoints.
14. **Create Kubernetes Deployment Manifests:** Author Flink Kubernetes Operator `FlinkDeployment` Custom Resource Definitions (CRDs) with TaskManager replica counts, CPU/memory limits, and S3 checkpoint URI paths.

## Interfaces / Contracts

### 1. Flink Deployment Descriptor (`flink-deployment.yaml`)
```yaml
apiVersion: flink.apache.org/v1beta1
kind: FlinkDeployment
metadata:
  name: flink-market-analytics
  namespace: analytics
spec:
  image: growww/flink-stream-analytics:1.19.0-v1
  flinkVersion: v1_19
  flinkConfiguration:
    taskmanager.numberOfTaskSlots: "4"
    state.backend: rocksdb
    state.backend.incremental: "true"
    state.checkpoints.dir: s3://growww-flink-checkpoints/market-analytics/checkpoints
    state.savepoints.dir: s3://growww-flink-checkpoints/market-analytics/savepoints
    execution.checkpointing.interval: 3000ms
    execution.checkpointing.min-pause: 1000ms
    execution.checkpointing.timeout: 30000ms
    execution.checkpointing.max-concurrent-checkpoints: "1"
    execution.checkpointing.mode: EXACTLY_ONCE
    state.backend.rocksdb.memory.managed: "true"
    state.backend.rocksdb.block.cache-size: 2147483648b
    state.backend.rocksdb.use-bloom-filter: "true"
    restart-strategy: fixed-delay
    restart-strategy.fixed-delay.attempts: "5"
    restart-strategy.fixed-delay.delay: 5000ms
  serviceAccount: flink-service-account
  jobManager:
    resource:
      memory: "4096m"
      cpu: 2.0
  taskManager:
    resource:
      memory: "8192m"
      cpu: 4.0
    replicas: 6
  job:
    jarURI: local:///opt/flink/usrlib/market-analytics-job.jar
    entryClass: com.growww.analytics.market.MarketAnalyticsJob
    parallelism: 24
    upgradeMode: savepoint
    state: running
```

### 2. Protocol Buffers: Market Volatility & Microstructure Features (`market_features.proto`)
```protobuf
syntax = "proto3";

package growww.analytics.features.v1;

option java_multiple_files = true;
option java_package = "com.growww.analytics.features.v1";
option go_package = "growww/analytics/features/v1;featuresv1";

message MarketVolatilityFeatureVector {
  string isin = 1;
  int64 window_timestamp_epoch_ms = 2;
  int64 calculation_timestamp_epoch_ms = 3;
  
  double vwap_1m = 4;
  double vwap_5m = 5;
  double vwap_15m = 6;
  
  double garman_klass_volatility_5m = 7;
  double parkinson_volatility_5m = 8;
  double realized_volatility_15m = 9;
  
  double price_high_5m = 10;
  double price_low_5m = 11;
  double price_open_5m = 12;
  double price_close_5m = 13;
  
  int64 total_trades_5m = 14;
  double total_volume_5m = 15;
  
  double order_book_imbalance_top5 = 16;
  double order_book_imbalance_top10 = 17;
  
  bool is_circuit_corridor_breached = 18;
  double dynamic_upper_band = 19;
  double dynamic_lower_band = 20;
}

message OracleMarketAggregateEvent {
  string isin = 1;
  int64 timestamp_epoch_ms = 2;
  double reference_price = 3;
  double vwap_15m = 4;
  double annualized_volatility_bps = 5;
  int64 total_trades_sample = 6;
  string source_fingerprint = 7;
  bytes cryptographic_signature = 8;
}
```

### 3. Protocol Buffers: Trader Behavioral Features (`trader_features.proto`)
```protobuf
syntax = "proto3";

package growww.analytics.features.v1;

option java_multiple_files = true;
option java_package = "com.growww.analytics.features.v1";
option go_package = "growww/analytics/features/v1;featuresv1";

message TraderBehavioralFeatureVector {
  string user_id = 1;
  int64 timestamp_epoch_ms = 2;
  
  int32 orders_placed_10s = 3;
  int32 orders_cancelled_10s = 4;
  int32 orders_modified_10s = 5;
  int32 trades_executed_10s = 6;
  
  double order_to_trade_ratio_60s = 7;
  double median_cancellation_latency_ms_60s = 8;
  double rapid_cancellation_ratio = 9; // Percentage of orders cancelled in < 20ms
  
  double aggressive_buy_volume_ratio_5m = 10;
  double aggressive_sell_volume_ratio_5m = 11;
  
  double gross_notional_traded_5m = 12;
  double net_exposure_delta_5m = 13;
  
  int32 unique_symbols_traded_15m = 14;
  bool is_rate_limit_warning_active = 15;
  double spoofing_risk_score = 16;
}
```

### 4. Feast Feature Store Definition (`feature_store.yaml`)
```yaml
project: growww_feature_store
registry: s3://growww-feature-store-registry/registry.pb
provider: local
online_store:
  type: redis
  connection_string: rediss://default:${REDIS_AUTH_TOKEN}@redis-cluster.analytics.internal:6379
  key_ttl: 86400
  ssl: true
offline_store:
  type: file
entity_key_serialization_version: 2
```

### 5. Feast Feature View Declarations (Python DSL Specification)
```python
# Feature store schema contracts for Feast integration
from datetime import timedelta
from feast import Entity, Field, FeatureView, KafkaSource, StreamFeatureView
from feast.types import Array, Bool, Float64, Int64, String

trader_entity = Entity(
    name="user_id",
    value_type=Entity.ValueType.STRING,
    description="Unique investor identifier UUID"
)

instrument_entity = Entity(
    name="isin",
    value_type=Entity.ValueType.STRING,
    description="International Securities Identification Number"
)

market_features_view = StreamFeatureView(
    name="market_volatility_features",
    entities=[instrument_entity],
    ttl=timedelta(hours=2),
    schema=[
        Field(name="vwap_1m", dtype=Float64),
        Field(name="vwap_5m", dtype=Float64),
        Field(name="garman_klass_volatility_5m", dtype=Float64),
        Field(name="parkinson_volatility_5m", dtype=Float64),
        Field(name="order_book_imbalance_top5", dtype=Float64),
        Field(name="dynamic_upper_band", dtype=Float64),
        Field(name="dynamic_lower_band", dtype=Float64),
    ],
    source=KafkaSource(
        name="market_volatility_stream_source",
        kafka_bootstrap_servers="kafka-broker.internal:9092",
        topic="features.market_volatility.v1",
        timestamp_extractor="calculation_timestamp_epoch_ms",
        message_format="protobuf",
    ),
    online=True
)
```

## Security & Compliance Notes
- **SEBI CSCRF Audit Trails & State Determinism:** In strict compliance with SEBI Cybersecurity and Cyber Resilience Framework (CSCRF) regulations, Flink streaming state checkpoints and savepoints are recorded with cryptographic SHA-256 integrity digests. All checkpoint manifests, state storage logs, and pipeline configuration revisions are retained in WORM-compliant storage (Write Once Read Many) for 8 years to guarantee historical replayability during regulatory inspections.
- **Deterministic Checkpoint Storage & Encryption:** RocksDB incremental state snapshots committed to AWS S3 / MinIO are encrypted at rest using envelope encryption backed by AWS KMS or HashiCorp Vault. All inter-node data exchanges between TaskManagers and JobManagers require mutual TLS (mTLS) with TLS 1.3.
- **Exactly-Once Processing Semantics:** To prevent financial distortion in risk calculations and surveillance metrics, the pipeline enforces strict end-to-end exactly-once semantics using Kafka two-phase commit transactions and Flink transactional state operators. Duplicate messages originating from Kafka consumer rebalances are deduplicated via event IDs and monotonically increasing sequence counters.
- **Zero-PII Storage Policy:** All streaming state maintained within RocksDB and published to Redis contains solely pseudonymized keys (`user_id`, `isin`, `account_id`). Personal identifiers, investor names, Aadhaar details, and permanent account numbers (PANs) are strictly excluded from stream processing state and feature store keys in compliance with the Digital Personal Data Protection (DPDP) Act 2023.
- **Backpressure Protection & Resource Isolation:** Flink TaskManagers are provisioned with dedicated cgroup CPU and off-heap memory reservations. Network buffer pools and credit-based flow control prevent stream backpressure from cascading into upstream order matching systems. Market data streaming jobs and trader behavioral jobs run in isolated TaskManager worker pools to prevent memory contention.

## Acceptance Criteria
- [ ] Flink 1.19+ cluster deploys successfully on Kubernetes via Flink Kubernetes Operator with all JobManager and TaskManager pods entering `Running` status.
- [ ] Market Microstructure streaming job achieves sub-5ms processing latency from Kafka trade match ingestion to output feature calculation under standard load (50,000 events/sec).
- [ ] Embedded RocksDB state backend correctly performs asynchronous incremental checkpoints to S3 at 3-second intervals with zero TaskManager thread blocking.
- [ ] Garman-Klass, Parkinson, and rolling VWAP algorithms produce mathematically validated outputs verified against static analytical test vectors.
- [ ] Order Book Imbalance (OBI) operator calculates depth imbalances within $< 2\text{ms}$ of Level-2 quote arrival.
- [ ] Trader Behavioral job accurately updates rolling Order-to-Trade Ratios (OTR) and asserts abuse flags when thresholds exceed configured regulatory limits.
- [ ] Feast online feature store integration writes feature updates into Redis Cluster with pipelined latency $< 1\text{ms}$ per micro-batch.
- [ ] Downstream Pre-Trade Risk Engine (Prompt 206) and Real-Time VaR Engine (Prompt 229) successfully query Redis online features with sub-millisecond round-trip response times.
- [ ] Kafka transactional egress sinks guarantee zero message loss and zero message duplicates upon simulated TaskManager pod termination and failover.
- [ ] Job failure recovery from an S3 incremental checkpoint successfully restores full state and resumes processing within $< 30\text{s}$ (RTO $< 30\text{s}$, RPO $= 0$).
- [ ] Oracle aggregate pipeline emits cryptographically signed price and volatility updates to topic `oracle.aggregates.v1` consumed by Besu relayer nodes.
- [ ] Checkpoint manifests and operator logs are retained with tamper-evident audit logs adhering to SEBI CSCRF mandates.
- [ ] Specification contains zero em dashes or en dashes throughout the document, adhering strictly to ASCII hyphen standards.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Event Schema and Kafka Topic Standards), Prompt 112 (Idempotency and Exactly-Once Processing), Prompt 402 (Redis Patterns and Caching), Prompt 403 (Kafka Cluster and Topic Partitioning).
- **Upstream Producing Microservices:** Prompt 204 (Order Service), Prompt 205 (Order Matching Engine), Prompt 207 (Market Data Service).
- **Downstream Consuming Microservices:** Prompt 206 (Pre-Trade Risk & Margin Engine), Prompt 228 (Real-Time Market Surveillance Engine), Prompt 229 (Real-Time VaR & Margin Engine).
- **Downstream Blockchain & Storage Infrastructure:** Prompt 328 (Decentralized Multi-Source Oracle Aggregator Smart Contract), Prompt 408 (Historical Market Data TimescaleDB & ClickHouse Pipeline).
