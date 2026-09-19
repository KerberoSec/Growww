# 404 - Data Warehouse / Analytics Pipeline (For Compliance & BI)

## Purpose
Financial regulatory compliance under SEBI, RBI, and IFSCA demands deep, real-time auditability, market surveillance, trade reconstitution, and financial reporting. Traditional transactional OLTP databases cannot sustain complex aggregations over billions of historical trade records, market ticks, and ledger postings without degrading operational performance.

The ClickHouse-based Data Warehouse and Analytics Pipeline provides a real-time columnar analytical backbone for Growww. It ingests continuous event streams from Kafka and Debezium CDC, providing sub-second analytical query capabilities for regulatory surveillance (e.g. spoofing, front-running, and wash trading detection), executive BI dashboards, proof-of-reserve historical audit trails, investor tax reports, and real-time OHLCV candlestick aggregation.

## What You Are Building
A production-grade, real-time analytical data warehouse pipeline containing:
- **ClickHouse Cluster Architecture:** Kubernetes manifests and configuration (`infra/clickhouse/`) deploying a 3-node ClickHouse cluster with ClickHouse Keeper for replication.
- **Columnar Analytical Schemas:** Optimized DDL definitions using `MergeTree`, `ReplacingMergeTree`, and `AggregatingMergeTree` table engines for trades, orders, ledger postings, market ticks, and on-chain blockchain transactions.
- **Streaming Kafka Ingestion Pipeline:** ClickHouse Kafka Engine tables and consumer pipelines consuming directly from Kafka topics (`trading.trades.executed`, `ledger.journal.postings`, `blockchain.besu.events`).
- **Debezium CDC Pipeline:** Debezium PostgreSQL connector streaming transactional mutations (`identity`, `custody`, `compliance`) into the data warehouse.
- **Materialized Views for Real-Time Aggregations:** Continuous materialized views computing OHLCV (Open-High-Low-Close-Volume) candlestick timeframes (1m, 5m, 1h, 1d) and rolling profit-fee analytics.
- **BI & Surveillance Integration:** Apache Superset dashboards and automated surveillance SQL alert queries for market manipulation detection.

## Scope Boundaries
- **In Scope:** ClickHouse cluster topology, columnar DDLs, Kafka Engine connectors, Materialized Views, Debezium CDC configuration, tiered S3 storage policies, and query optimization.
- **Out of Scope / Handled Elsewhere:**
 - OLTP transactional database design (handled in Prompt 401).
 - Upstream Kafka cluster provisioning (handled in Prompt 403).
 - Business microservices executing trades or orders (handled in Category 2).
 - Client-facing reporting microservice (handled in Prompt 216).

## Technology to Use
ClickHouse 24.x is selected as the primary OLAP columnar database engine. ClickHouse delivers exceptional vector execution speeds, data compression rates exceeding 80%, native Kafka streaming table engines, and scalable storage tiering (local NVMe hot storage migrating to S3 object storage). Compared to Snowflake or BigQuery, ClickHouse allows on-premise/in-region data residency (satisfying RBI localization mandates) with zero external egress latency and predictable infrastructure costs.

- **Analytical Engine:** ClickHouse 24.x with ClickHouse Keeper (Zookeeper-free).
- **Streaming CDC:** Debezium 2.5+ PostgreSQL Connector.
- **BI Visualization:** Apache Superset 3.1+.
- **Storage Tiering:** Hot local EBS `gp3` disks with automatic cold partition migration to AWS S3 (Mumbai `ap-south-1`).
- **Query Protocol:** ClickHouse Native Protocol / HTTP REST Interface.

## Backend / Infra Touchpoints
- **Kafka Cluster:** Ingests high-volume event streams from topics (`trading.trades.executed`, `blockchain.besu.events`).
- **PostgreSQL 16:** Source database captured via Debezium CDC.
- **AWS S3 / MinIO:** Object storage data lake tier for archival and multi-year regulatory partition retention.
- **Apache Superset:** Connected to ClickHouse for executive and compliance reporting dashboards.

## Blockchain Interaction
The analytics pipeline provides real-time and historical analytics across all transactions occurring on the permissioned Hyperledger Besu blockchain network (QBFT consensus).

### Detailed On-Chain Integration Mechanics:
- **Ledger Ingestion:** Consumes parsed smart contract events emitted by `DigitalSecurityToken.sol`, `SettlementDvP.sol`, and `ProofOfReserveRegistry.sol` via Kafka.
- **Proof-of-Reserve Historical Audit:** Stores daily on-chain Merkle root commitments alongside physical NSDL/CDSL depository share balances to generate verifiable proof-of-reserve time series graphs.
- **DvP Settlement Latency Analytics:** Calculates latency distributions between trade matching off-chain and final on-chain DvP settlement confirmation block timestamps.
- **Zero PII Storage:** ClickHouse tables strictly record pseudonymous wallet addresses (`0x...`), ISINs, and token contract addresses. No investor PII is stored in the analytics warehouse.

## Step-by-Step Build Instructions
1. Scaffold repository structure: `analytics/clickhouse/schemas/`, `analytics/kafka_connect/`, `analytics/superset/`.
2. Deploy ClickHouse 3-node cluster with ClickHouse Keeper on Kubernetes across 3 AZs.
3. Configure storage tiering policy in ClickHouse (`hot_disk` local storage for 90 days, `s3_cold` bucket for 8 years).
4. Implement ClickHouse core analytical tables: `analytics.trades`, `analytics.orders`, `analytics.ledger_postings`, `analytics.blockchain_events`.
5. Implement ClickHouse Kafka Engine integration tables to ingest JSON/Avro messages directly from Kafka topics.
6. Create Materialized Views linking Kafka Engine tables to target `MergeTree` tables for continuous, lock-free ingestion.
7. Implement Materialized Views for automated OHLCV candlestick aggregation (`analytics.ohlcv_1m`, `analytics.ohlcv_1d`).
8. Deploy Debezium PostgreSQL Connector to stream CDC changes from `identity.users` and `custody.allocations` to Kafka.
9. Implement ClickHouse `ReplacingMergeTree` tables to maintain CDC state updates with deduplication on version timestamps.
10. Implement market surveillance detection queries (e.g. wash trading queries identifying simultaneous buy/sell between linked addresses within 500ms).
11. Deploy Apache Superset and configure data sources, semantic datasets, and pre-built compliance dashboards.
12. Benchmark analytical query latency: verify aggregate queries over 100M rows execute in < 250ms.

## Interfaces / Contracts

```sql
-- ClickHouse DDL: Trades Analytical Table
CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE analytics.trades (
    trade_id UUID,
    isin LowCardinality(String),
    symbol LowCardinality(String),
    buyer_order_id UUID,
    seller_order_id UUID,
    buyer_address FixedString(42),
    seller_address FixedString(42),
    fractional_units Decimal64(6),
    price_per_unit_inr Decimal64(4),
    gross_amount_inr Decimal64(4),
    fee_amount_inr Decimal64(4),
    settlement_status LowCardinality(String),
    on_chain_tx_hash FixedString(66),
    on_chain_block_num UInt64,
    executed_at DateTime64(3, 'Asia/Kolkata')
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(executed_at)
ORDER BY (isin, executed_at, trade_id)
SETTINGS index_granularity = 8192;

-- Kafka Engine Ingestion Table
CREATE TABLE analytics.kafka_trades_stream (
    tradeId UUID,
    isin String,
    symbol String,
    buyerOrderId UUID,
    sellerOrderId UUID,
    buyerAddress String,
    sellerAddress String,
    fractionalUnits String,
    pricePerUnitINR String,
    grossAmountINR String,
    feeAmountINR String,
    executedAt Int64
) ENGINE = Kafka
SETTINGS kafka_broker_list = 'kafka-cluster-kafka-bootstrap.kafka.svc:9092',
         kafka_topic_list = 'trading.trades.executed',
         kafka_group_name = 'clickhouse_trades_consumer',
         kafka_format = 'JSONEachRow';

-- Materialized View Ingesting Stream into Persistent Storage
CREATE MATERIALIZED VIEW analytics.mv_kafka_trades TO analytics.trades AS
SELECT
    tradeId AS trade_id,
    isin,
    symbol,
    buyerOrderId AS buyer_order_id,
    sellerOrderId AS seller_order_id,
    buyerAddress AS buyer_address,
    sellerAddress AS seller_address,
    toDecimal64(fractionalUnits, 6) AS fractional_units,
    toDecimal64(pricePerUnitINR, 4) AS price_per_unit_inr,
    toDecimal64(grossAmountINR, 4) AS gross_amount_inr,
    toDecimal64(feeAmountINR, 4) AS fee_amount_inr,
    'SETTLED_DVP' AS settlement_status,
    '' AS on_chain_tx_hash,
    0 AS on_chain_block_num,
    toDateTime64(executedAt / 1000, 3, 'Asia/Kolkata') AS executed_at
FROM analytics.kafka_trades_stream;

-- Materialized View for 1-Minute OHLCV Candlesticks
CREATE TABLE analytics.ohlcv_1m (
    isin LowCardinality(String),
    window_start DateTime('Asia/Kolkata'),
    open Decimal64(4),
    high Decimal64(4),
    low Decimal64(4),
    close Decimal64(4),
    volume Decimal64(6),
    trade_count UInt32
) ENGINE = SummingMergeTree((volume, trade_count))
PRIMARY KEY (isin, window_start);

-- Proof-of-Reserve Historical Audit Table
CREATE TABLE analytics.proof_of_reserve_history (
    snapshot_date Date,
    isin LowCardinality(String),
    depository LowCardinality(String),
    physical_shares_held Decimal64(6),
    on_chain_tokens_minted Decimal64(6),
    merkle_root_hash FixedString(66),
    on_chain_tx_hash FixedString(66),
    reconciliation_delta Decimal64(6),
    reconciled_at DateTime('Asia/Kolkata')
) ENGINE = ReplacingMergeTree(reconciled_at)
PARTITION BY toYYYYMM(snapshot_date)
ORDER BY (snapshot_date, isin, depository);
```

## Security & Compliance Notes
- **SEBI 8-Year Audit Retention:** Data tiering policies ensure all trade, order, and ledger posting records are preserved in WORM (Write Once Read Many) compliant S3 object storage for at least 8 years.
- **Market Surveillance Algorithms:** Scheduled ClickHouse queries continuously scan for illegal market patterns: circular trading, pump-and-dump coordination, and wash trading between related entities.
- **Strict Role-Based Access Control (RBAC):** BI analysts have read-only permissions with column-level masking applied to sensitive identifiers.
- **Data Sovereignty:** All ClickHouse nodes and S3 backup buckets reside strictly within AWS ap-south-1 (Mumbai), conforming to RBI guidelines on payment and banking transaction data storage.

## Acceptance Criteria
- [ ] 3-node ClickHouse cluster deployed with ClickHouse Keeper and multi-AZ replication.
- [ ] Kafka Engine tables and Materialized Views stream live trade events from Kafka with zero data drop.
- [ ] Debezium PostgreSQL CDC pipeline successfully captures table updates from `identity` and `custody` schemas into ClickHouse.
- [ ] OHLCV 1-minute and 1-day candlestick materialized views compute real-time price bars accurately.
- [ ] Analytical benchmark queries (`SELECT AVG(price_per_unit_inr), SUM(volume) FROM trades WHERE isin = ? GROUP BY toStartOfHour(executed_at)`) complete in < 150ms over 50M records.
- [ ] Storage tiering policy successfully shifts partitions older than 90 days from NVMe to encrypted S3 storage.
- [ ] Market surveillance automated SQL query successfully detects simulated wash trading scenarios.
- [ ] Apache Superset connected with executive BI dashboards and SEBI compliance audit views.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Kafka Standards), Prompt 401 (PostgreSQL Schema), Prompt 403 (Kafka Cluster).
- **Parallel Tasks:** Prompt 405 (Data Retention & Archival), Prompt 216 (Reporting Service).
