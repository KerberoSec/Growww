# 410 - Debezium CDC Transactional Outbox & Event-Sourced Ledger Pipeline

## Purpose
In high-concurrency financial trading, multi-asset digital custody, and real-time ledger accounting, data consistency between relational databases and asynchronous event streaming platforms is paramount. Distributed architectures commonly suffer from dual-write anomalies: when an application writes state changes to a relational database and subsequently publishes events to Apache Kafka, network timeouts, process crashes, or broker rebalances can cause one operation to succeed while the other fails. In financial workflows such as double-entry ledger postings (Prompt 203) or Delivery-versus-Payment (DvP) trade settlements (Prompt 208), dual-write failures result in balance drift, missing settlement triggers, unreconciled custody allocations, and audit non-compliance.

This prompt defines the enterprise architecture, deployment configurations, relational schemas, and consumer guarantees for the **Debezium CDC Transactional Outbox & Event-Sourced Ledger Pipeline**. By adopting the Transactional Outbox pattern powered by Debezium Change Data Capture (CDC) and PostgreSQL native logical decoding (`pgoutput`), this pipeline guarantees 100% atomic state transitions between PostgreSQL databases and Kafka event streams. Business domain mutations and corresponding outbox event records are committed within a single local ACID transaction. Debezium continuously tails PostgreSQL write-ahead logs (WAL) asynchronously with sub-millisecond latency, transforms outbox records into standardized event envelopes, and dispatches them into targeted Kafka topics with strict partition ordering and zero operational polling overhead on primary transactional tables.

## What You Are Building
A production-grade, fault-tolerant Change Data Capture (CDC) and Transactional Outbox event infrastructure across all Growww transactional microservices, comprising:
- **Declarative PostgreSQL Outbox Schemas (`outbox_events`):** High-throughput, range-partitioned relational tables deployed across all transactional services capturing domain events, aggregate IDs, sequence counters, trace contexts, and JSONB payloads in the same database transaction as domain entities.
- **Debezium Distributed Kafka Connect Architecture:** Resilient Kafka Connect worker clusters deployed via Strimzi Kubernetes Operator running Debezium PostgreSQL Connector 2.5+ instances with non-blocking failover, continuous WAL offset tracking, and automated replication slot management.
- **Debezium Outbox Event Router Single Message Transform (SMT):** Optimized transformation pipeline converting PostgreSQL table mutations into clean, header-enriched Kafka events routed dynamically to target domain topics (e.g. `ledger.journal.posted`, `trading.order.matched`, `settlement.dvp.completed`) while routing partition keys to preserve strict domain ordering.
- **Signal-Based Incremental Snapshotting & Heartbeat Monitoring:** Dedicated PostgreSQL signaling tables and heartbeat topics enabling ad-hoc table snapshots without table locks, automated WAL progress tracking, and proactive replication slot leak detection.
- **High-Performance Consumer Inbox & Deduplication Framework:** Standardized consumer deduplication contract (`inbox_events`) enforcing exactly-once processing semantics through aggregate-level sequence verification, transactional inbox record inserts, and idempotency key constraints.
- **Secret Redaction & Security Masking Pipeline:** Pre-CDC database triggers and Kafka Connect transform filters ensuring sensitive credentials, encryption keys, and unmasked Personally Identifiable Information (PII) are scrubbed before reaching event streams.

## Scope Boundaries
- **In Scope:**
  - PostgreSQL 16+ logical replication configuration (`wal_level=logical`, `pgoutput`, replication slots, publications).
  - Universal DDL for `outbox_events` and consumer `inbox_events` with declarative time-based partitioning.
  - Debezium 2.5+ PostgreSQL connector configuration (Strimzi `KafkaConnector` CRDs and distributed REST JSON configs).
  - Debezium Outbox Event Router SMT configuration, dynamic topic routing, partition key assignment, and tombstone generation.
  - Signal table DDL and mechanics for on-demand table re-snapshotting.
  - Consumer deduplication contract, aggregate sequence numbering, and exactly-once processing mechanics.
  - WAL disk consumption monitoring, replication slot leak alerting, and automated partition cleanup.
  - Zero-PII and secret redaction rules across the CDC pipeline.
- **Out of Scope / Handled Elsewhere:**
  - Microservice business domain handlers and REST/gRPC endpoints (handled in Category 2 microservices: Prompts 201-226).
  - Kafka broker cluster provisioning and KRaft consensus topology (handled in Prompt 403).
  - OLAP historical data ingestion into ClickHouse and TimescaleDB (handled in Prompts 404 and 408).
  - Redis cache synchronization and real-time WebSocket distribution (handled in Prompts 402 and 207).
  - Core double-entry ledger domain model and chart of accounts (handled in Prompt 203).

## Technology to Use
- **Change Data Capture Engine:** **Debezium 2.5+** (PostgreSQL Connector). Debezium is the industry standard for database log mining, providing low-latency, non-intrusive stream capture directly from the PostgreSQL WAL without application-level polling.
- **Database Engine & Decoding Plugin:** **PostgreSQL 16+** with native **`pgoutput`** logical decoding plugin. Native `pgoutput` operates entirely within the PostgreSQL core engine, eliminating the need for external C-based extensions (such as `decoderbufs` or `wal2json`) and minimizing CPU and memory overhead during high-volume transactions.
- **Event Streaming Broker:** **Apache Kafka 3.7+** running in KRaft metadata mode, providing high-durability commit logs, partition-level strict ordering, and multi-zone replication.
- **Integration Framework:** **Kafka Connect (Distributed Mode)** managed via the **Strimzi Kafka Operator 0.40+** on Kubernetes. Provides horizontal scalability, declarative lifecycle management through Kubernetes Custom Resource Definitions (CRDs), and automatic worker task rebalancing.
- **Message Serialization & Schema Governance:** **Apache Avro / JSON Schema** integrated with **Confluent Schema Registry / Karapace**, enforcing schema validation and strict backward-transitive compatibility.
- **Partition Maintenance:** **`pg_partman` 5.x** for automated daily/weekly range-partition management and high-speed drop-partition truncation of processed outbox events.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Primary Databases:** Primary relational databases for each service (Wallet Ledger, Settlement Service, Order Service, Custody Service) configured with `wal_level=logical`, dedicated replication user roles, and publication scopes.
- **PostgreSQL Write-Ahead Log (WAL):** High-IOPS NVMe storage volumes hosting WAL segments (`pg_wal`), monitored continuously to prevent disk saturation from stalled replication slots.
- **Kafka Connect Worker Cluster:** Dedicated 3-replica Kubernetes deployment running containerized Kafka Connect runtime with Debezium 2.5 plugins installed in `/kafka/connect/plugins`.
- **Apache Kafka Cluster:** Ingestion broker cluster hosting destination domain topics, CDC offset storage topics (`connect_offsets`), config storage topics (`connect_configs`), status storage topics (`connect_status`), and heartbeat topics (`debezium.heartbeat.*`).
- **Schema Registry Cluster:** Centralized schema registry validating event envelopes and payload schemas.
- **Wallet & Account Service (Prompt 203):** Core transactional producer writing ledger journal entries and `outbox_events` atomically during account debits, credits, and balance holds.
- **Trade Settlement Service (Prompt 208):** Core settlement engine committing Delivery-versus-Payment (DvP) trade states and dispatching settlement events via CDC outbox.
- **HashiCorp Vault / AWS Secrets Manager:** Secure storage providing dynamic, rotating database credentials for Debezium replication users.
- **Prometheus & Grafana:** Observability stack scraping Debezium JMX metrics, tracking WAL consumption, consumer group lag, and end-to-end CDC dispatch latency.

## Blockchain Interaction
While Debezium operates primarily on off-chain relational databases, it plays a vital role in synchronizing off-chain state with the permissioned Hyperledger Besu consortium blockchain (QBFT consensus, 2-second block finality, 1:1 asset backing, zero PII):

### Detailed On-Chain Integration Mechanics:
- **Mirroring On-Chain Commits to Outbox:** When the Blockchain Event Indexer (Prompt 309) receives verified block events from Hyperledger Besu smart contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ProofOfReserveRegistry.sol`), it writes the parsed state change and the block receipt into PostgreSQL within a single ACID transaction that also inserts an event into `outbox_events`.
- **Unified Audit Trail Generation:** By committing on-chain receipts (transaction hash, block height, log index, validator signatures) to the transactional outbox, Debezium streams these immutable blockchain confirmations into the Kafka event backbone. Downstream microservices (such as Prompt 215 Reconciliation Service and Prompt 216 Regulatory Reporting Service) consume a unified, ordered event stream that correlates off-chain fiat operations with on-chain cryptographic settlement without establishing direct RPC connections to Besu nodes.
- **Two-Way Settlement Choreography:**
  1. Off-chain trade matching engine executes an order and writes trade state plus `TradeMatched` outbox event to PostgreSQL.
  2. Debezium streams `TradeMatched` to Kafka topic `trading.orders.matched`.
  3. DvP Settlement Relayer consumes the event, constructs an on-chain settlement transaction, and submits it to `SettlementDvP.sol`.
  4. Upon block commit, the Besu Indexer records the settlement execution in PostgreSQL and writes `DvPExecuted` to `outbox_events`.
  5. Debezium streams `DvPExecuted` to Kafka topic `settlement.dvp.executed`, notifying Wallet Ledger (Prompt 203) to finalize fiat balance transfers.
- **Zero-PII On-Chain Guarantee:** All events routed through the CDC outbox pipeline replace real investor identities with pseudonymous UUIDs and deterministic Ethereum addresses (`0x...`). No names, PANs, bank accounts, or unencrypted identity attributes are streamed through Kafka or committed to blockchain logs.

## Pipeline Architecture & CDC Outbox Mechanics

### 1. The Dual-Write Problem and the Transactional Outbox Pattern
In distributed financial architectures, writing state changes to a primary database and publishing an event to a message broker in separate steps causes dual-write inconsistencies:
```
[Application Layer]
     |
     +---> 1. UPDATE accounts SET balance = balance - 100 ... (SUCCESS)
     |
     +---> 2. kafkaProducer.send("ledger.events", payload) ... (NETWORK FAILURE / CRASH)
Result: Database updated, but downstream systems never receive the event!
```
The Transactional Outbox pattern resolves this problem by storing the event inside the database itself:
```
BEGIN TRANSACTION;
  UPDATE ledger.accounts SET balance = balance - 100 WHERE account_id = 'ACC-123';
  INSERT INTO ledger.journal_entries (...) VALUES (...);
  INSERT INTO ledger.outbox_events (
    id, aggregate_type, aggregate_id, event_type, payload, headers, sequence_number
  ) VALUES (
    '018e5f2a-7b3c-789a-bcde-0123456789ab', 'Account', 'ACC-123',
    'AccountDebited', '{"amount": "100.00", "currency": "INR"}',
    '{"trace_id": "tr-98765"}', 42
  );
COMMIT;
```
Because both operations execute inside the same ACID boundary, either both succeed or both roll back.

### 2. Debezium CDC Engine with PostgreSQL `pgoutput`
Rather than polling the `outbox_events` table using heavy `SELECT ... FOR UPDATE SKIP LOCKED` queries that cause CPU spikes and table bloat, Debezium attaches to PostgreSQL as a logical replication consumer:
```
+---------------------+
|  PostgreSQL 16 DB   |
|  +---------------+  |
|  | outbox_events |  |
|  +---------------+  |
|         |           |
|     (WAL Log)       |
|         |           |
|  [Logical Dec.]     |
|     (pgoutput)      |
+---------+-----------+
          |  Replication Protocol (TCP)
          v
+-------------------------------------------------------------+
|               Kafka Connect Worker (Debezium 2.5)           |
|  +-------------------------------------------------------+  |
|  | PostgreSQL Connector: tail WAL stream                 |  |
|  +-------------------------------------------------------+  |
|                             |                               |
|                             v                               |
|  +-------------------------------------------------------+  |
|  | Outbox Event Router (Single Message Transform - SMT)  |  |
|  | - Extracts payload & headers                          |  |
|  | - Determines target topic: {route_by_field}           |  |
|  | - Assigns partition key: aggregate_id                 |  |
|  | - Emits clean CloudEvents-compliant message           |  |
|  +-------------------------------------------------------+  |
+-----------------------------+-------------------------------+
                              |
                              v Produce
+-------------------------------------------------------------+
|                     Apache Kafka Broker                     |
|  Topic: ledger.account.events (Key: 'ACC-123')              |
|  [Offset 101] [Offset 102] [Offset 103] ...                 |
+-------------------------------------------------------------+
```

### 3. Debezium Outbox Event Router SMT Configuration
The Debezium Outbox Event Router (`io.debezium.transforms.outbox.EventRouter`) transforms row-level change events from the `outbox_events` table into domain-specific Kafka records:
- **Topic Routing:** The target Kafka topic name is determined dynamically based on the event metadata. For instance, `route.by.field=destination_topic` routes directly to the topic specified in the outbox row, or `route.topic.replacement=${routedByValue}` maps `event_type` to a standardized domain topic.
- **Payload Extraction:** The SMT extracts the `payload` JSONB column and sets it as the Kafka message value, unwrapping it from the CDC metadata envelope (`before`, `after`, `source`, `op`).
- **Partition Key Assignment:** The SMT extracts `aggregate_id` and sets it as the Kafka message key, guaranteeing that all events for the same entity (e.g. account ID or order ID) land in the same Kafka partition, preserving chronological ordering.
- **Header Mapping:** Columns such as `id`, `event_type`, `aggregate_type`, `trace_parent`, and `schema_version` are mapped directly to Kafka record headers.
- **Tombstone Emission:** When configured, Debezium can automatically emit a tombstone record (null payload with identical key) after an entity deletion event to support Kafka log-compacted topics.

### 4. Outbox Table Partitioning and Storage Lifecycle
Because high-volume services generate tens of millions of outbox events daily, the `outbox_events` table must not grow indefinitely:
- **Declarative Range Partitioning:** Tables are partitioned by `created_at` in daily or weekly intervals using PostgreSQL 16 declarative partitioning and `pg_partman`.
- **Zero WAL Replication Bloat:** Debezium reads rows from the WAL in near real time. Once WAL segments are processed, physical database rows are no longer needed by Debezium.
- **Partition Drop Truncation:** Older partitions (e.g. older than 7 days) are dropped via `DROP TABLE` or `pg_partman` retention policies. Dropping partitions executes instantly with zero row-by-row `DELETE` overhead, zero vacuum bloat, and zero WAL amplification.

### 5. Consumer Deduplication Architecture (The Inbox Pattern)
Kafka provides at-least-once delivery semantics. During broker rebalances or connector restarts, duplicate events may be delivered to consumer microservices. To achieve exactly-once domain processing, consumers implement the **Inbox Pattern**:
```
Consumer receives Kafka Record (Key: aggregate_id, Headers: event_id, sequence_number)
                         |
                         v
               BEGIN DB TRANSACTION;
                         |
      +------------------+------------------+
      |                                     |
      v                                     v
INSERT INTO inbox_events (            Verify aggregate sequence:
  event_id, consumer_group,             IF sequence_number != expected_sequence:
  processed_at                          ABORT OR BUFFER OUT-OF-ORDER;
) VALUES (
  '018e5f2a-...', 'wallet-worker', NOW()
);
ON CONFLICT (event_id, consumer_group) DO NOTHING;
      |
      +---> IF 0 rows inserted: Duplicate event detected!
      |     COMMIT (Acknowledge Kafka offset without re-processing).
      |
      +---> IF 1 row inserted: First time seen!
            Execute business logic (e.g. update user balance);
            COMMIT;
```

## Step-by-Step Build Instructions
1. **Configure PostgreSQL WAL Engine:** Update PostgreSQL server configuration (`postgresql.conf`) with `wal_level = logical`, `max_replication_slots = 20`, `max_wal_senders = 20`, `wal_keep_size = 16384MB`, and `track_commit_timestamp = on` across all primary database instances.
2. **Provision Dedicated CDC Database Roles:** Execute security DDL creating a dedicated, least-privilege PostgreSQL replication user (`debezium_cdc`) with `REPLICATION` permissions and read-only `SELECT` access strictly scoped to the `outbox_events` and signal tables.
3. **Deploy Declarative Outbox Tables:** Deploy migration scripts defining the partitioned `outbox_events` table schema, composite indexes, and check constraints across all transactional service schemas (`ledger`, `trading`, `settlement`, `custody`).
4. **Deploy Consumer Inbox Deduplication Tables:** Deploy migration scripts defining the `inbox_events` schema across all consumer databases to enforce idempotency and prevent duplicate processing.
5. **Configure PostgreSQL Logical Publications:** Create selective logical publications (`CREATE PUBLICATION debezium_outbox_pub FOR TABLE outbox_events, debezium_signal;`) ensuring only outbox and signaling tables are tracked by the replication stream.
6. **Deploy Strimzi Kafka Connect Cluster:** Deploy the Strimzi `KafkaConnect` Custom Resource into the Kubernetes `kafka` namespace with 3 replicas, anti-affinity rules, and persistent volume storage for connector state.
7. **Configure Debezium PostgreSQL Connectors:** Author and apply Strimzi `KafkaConnector` manifests configuring the Debezium 2.5+ PostgreSQL connector with `plugin.name=pgoutput`, `publication.autocreate.mode=disabled`, and `slot.name=debezium_<service>_slot`.
8. **Configure Outbox Event Router SMT:** Configure the `io.debezium.transforms.outbox.EventRouter` SMT within the connector configuration, setting up dynamic topic routing, partition key extraction from `aggregate_id`, and header mapping.
9. **Implement Signaling and Ad-Hoc Snapshotting:** Deploy the `debezium_signal` table and configure the connector to listen for snapshot signals (`signal.data.collection`), enabling zero-downtime ad-hoc data re-synchronization.
10. **Configure Replication Slot Heartbeat & Lag Monitoring:** Enable heartbeat generation (`heartbeat.interval.ms = 10000`, `heartbeat.topics.prefix = __debezium-heartbeat`) to ensure logical replication slots advance position even during periods of low transactional activity.
11. **Implement Automated Partition Management:** Configure `pg_partman` to manage daily outbox table partitions with automated 7-day retention drops, preventing operational database disk exhaustion.
12. **Implement Universal Consumer Idempotency Middleware:** Package a shared Go / Java middleware library wrapping event consumption with transactional `inbox_events` deduplication checks and sequence validation.
13. **Deploy Monitoring, Alerting, and Dead Letter Queues (DLQ):** Configure Kafka Connect Dead Letter Queues (`connect.dlq.*`), Prometheus JMX Exporter metrics for replication lag (`debezium_metrics_MilliSecondsBehindSource`), and PagerDuty alerts for unconsumed WAL growth.

## Interfaces / Contracts

### 1. PostgreSQL Outbox Table Schema DDL (`outbox_events`)
```sql
CREATE SCHEMA IF NOT EXISTS eventing;

CREATE TABLE eventing.outbox_events (
    id UUID NOT NULL,                             -- UUIDv7 containing millisecond timestamp
    aggregate_type VARCHAR(64) NOT NULL,          -- e.g. 'Account', 'Order', 'SettlementDvP'
    aggregate_id VARCHAR(128) NOT NULL,          -- e.g. 'ACC-987213', 'ORD-2026-9812'
    event_type VARCHAR(128) NOT NULL,             -- e.g. 'AccountDebited', 'TradeSettled'
    destination_topic VARCHAR(128) NOT NULL,      -- Target Kafka topic: 'ledger.account.events'
    payload JSONB NOT NULL,                       -- Domain event payload conforming to schema
    headers JSONB DEFAULT '{}'::jsonb NOT NULL,   -- Distributed tracing context, producer identity
    sequence_number BIGINT NOT NULL,              -- Monotonically increasing aggregate sequence
    schema_version VARCHAR(16) NOT NULL DEFAULT 'v1',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

-- Composite index to support immediate local lookups and sequence verification
CREATE INDEX idx_outbox_aggregate_seq 
    ON eventing.outbox_events (aggregate_type, aggregate_id, sequence_number ASC);

-- Set replica identity to DEFAULT (pgoutput decodes row key and attributes)
ALTER TABLE eventing.outbox_events REPLICA IDENTITY DEFAULT;

-- Initial daily partitions (managed thereafter via pg_partman)
CREATE TABLE eventing.outbox_events_y2026m09d18 
    PARTITION OF eventing.outbox_events
    FOR VALUES FROM ('2026-09-18 00:00:00+00') TO ('2026-09-19 00:00:00+00');

CREATE TABLE eventing.outbox_events_y2026m09d19 
    PARTITION OF eventing.outbox_events
    FOR VALUES FROM ('2026-09-19 00:00:00+00') TO ('2026-09-20 00:00:00+00');
```

### 2. PostgreSQL Debezium Signal Table DDL (`debezium_signal`)
```sql
CREATE TABLE eventing.debezium_signal (
    id VARCHAR(64) PRIMARY KEY,
    type VARCHAR(32) NOT NULL,
    data VARCHAR(2048) NULL
);

-- Example signal to trigger an ad-hoc incremental snapshot:
-- INSERT INTO eventing.debezium_signal (id, type, data) 
-- VALUES ('sig-001', 'execute-snapshot', '{"data-collections": ["eventing.outbox_events"], "type": "INCREMENTAL"}');
```

### 3. PostgreSQL Consumer Inbox Table Schema DDL (`inbox_events`)
```sql
CREATE TABLE eventing.inbox_events (
    event_id UUID NOT NULL,
    consumer_group VARCHAR(64) NOT NULL,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL,
    sequence_number BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, consumer_group)
);

CREATE INDEX idx_inbox_aggregate_seq 
    ON eventing.inbox_events (consumer_group, aggregate_type, aggregate_id, sequence_number DESC);
```

### 4. Debezium PostgreSQL Kafka Connector JSON Configuration
```json
{
  "name": "debezium-postgres-ledger-outbox-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "tasks.max": "1",
    "plugin.name": "pgoutput",
    "database.hostname": "postgres-ledger-primary.db.internal",
    "database.port": "5432",
    "database.user": "debezium_cdc",
    "database.password": "${file:/secrets/db-credentials.properties:debezium_password}",
    "database.dbname": "growww_ledger",
    "database.server.name": "ledger_service",
    "topic.prefix": "cdc.ledger",
    "table.include.list": "eventing.outbox_events,eventing.debezium_signal",
    "publication.name": "debezium_outbox_pub",
    "publication.autocreate.mode": "disabled",
    "slot.name": "debezium_ledger_outbox_slot",
    "slot.drop.on.stop": "false",
    "snapshot.mode": "never",
    "tombstones.on.delete": "false",
    "decimal.handling.mode": "double",
    
    "signal.data.collection": "eventing.debezium_signal",
    
    "heartbeat.interval.ms": "10000",
    "heartbeat.topics.prefix": "__debezium-heartbeat",
    
    "transforms": "outbox",
    "transforms.outbox.type": "io.debezium.transforms.outbox.EventRouter",
    "transforms.outbox.table.fields.additional.placement": "event_type:header:eventType,aggregate_type:header:aggregateType,sequence_number:header:sequenceNumber,schema_version:header:schemaVersion,created_at:header:eventTimestamp",
    "transforms.outbox.route.by.field": "destination_topic",
    "transforms.outbox.route.topic.replacement": "${routedByValue}",
    "transforms.outbox.table.field.event.id": "id",
    "transforms.outbox.table.field.event.key": "aggregate_id",
    "transforms.outbox.table.field.event.payload": "payload",
    "transforms.outbox.table.field.event.payload.id": "id",
    "transforms.outbox.table.expand.json.payload": "true",
    
    "key.converter": "org.apache.kafka.connect.storage.StringConverter",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false",
    
    "errors.tolerance": "none",
    "errors.log.enable": "true",
    "errors.log.include.messages": "true",
    "errors.deadletterqueue.topic.name": "cdc.ledger.dlq",
    "errors.deadletterqueue.topic.replication.factor": "3",
    "errors.deadletterqueue.context.headers.enable": "true"
  }
}
```

### 5. Strimzi Kubernetes `KafkaConnector` Custom Resource Manifest
```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: debezium-postgres-settlement-outbox
  namespace: kafka
  labels:
    strimzi.io/cluster: growww-kafka-connect
spec:
  class: io.debezium.connector.postgresql.PostgresConnector
  tasksMax: 1
  config:
    plugin.name: pgoutput
    database.hostname: postgres-settlement-primary.db.internal
    database.port: 5432
    database.user: debezium_cdc
    database.password: ${secrets:vault/settlement-cdc#password}
    database.dbname: growww_settlement
    topic.prefix: cdc.settlement
    table.include.list: eventing.outbox_events,eventing.debezium_signal
    publication.name: debezium_settlement_pub
    publication.autocreate.mode: disabled
    slot.name: debezium_settlement_outbox_slot
    slot.drop.on.stop: false
    snapshot.mode: never
    tombstones.on.delete: false
    decimal.handling.mode: double
    
    signal.data.collection: eventing.debezium_signal
    heartbeat.interval.ms: 10000
    heartbeat.topics.prefix: __debezium-heartbeat
    
    transforms: outbox
    transforms.outbox.type: io.debezium.transforms.outbox.EventRouter
    transforms.outbox.table.fields.additional.placement: >-
      event_type:header:eventType,aggregate_type:header:aggregateType,sequence_number:header:sequenceNumber,schema_version:header:schemaVersion,created_at:header:eventTimestamp
    transforms.outbox.route.by.field: destination_topic
    transforms.outbox.route.topic.replacement: ${routedByValue}
    transforms.outbox.table.field.event.id: id
    transforms.outbox.table.field.event.key: aggregate_id
    transforms.outbox.table.field.event.payload: payload
    transforms.outbox.table.expand.json.payload: true
    
    key.converter: org.apache.kafka.connect.storage.StringConverter
    value.converter: org.apache.kafka.connect.json.JsonConverter
    value.converter.schemas.enable: false
    
    errors.tolerance: none
    errors.deadletterqueue.topic.name: cdc.settlement.dlq
    errors.deadletterqueue.topic.replication.factor: 3
    errors.deadletterqueue.context.headers.enable: true
```

### 6. Standardized Kafka Event Envelope Contract
Events routed out of Debezium onto target domain topics conform to the following JSON structure:
```json
{
  "eventId": "018e5f2a-7b3c-789a-bcde-0123456789ab",
  "eventType": "AccountDebited",
  "aggregateType": "Account",
  "aggregateId": "ACC-987213",
  "sequenceNumber": 14208,
  "schemaVersion": "v1",
  "timestamp": 1726685400000,
  "payload": {
    "accountId": "ACC-987213",
    "journalEntryId": "JRN-2026-0918-00451",
    "amount": "150000.00",
    "currency": "INR",
    "balanceBefore": "450000.00",
    "balanceAfter": "300000.00",
    "reason": "TRADE_MARGIN_LOCK",
    "orderId": "ORD-2026-88120"
  },
  "metadata": {
    "traceId": "c4b2a890ef123456",
    "spanId": "a1b2c3d4e5f67890",
    "userId": "usr_9982410a",
    "ipAddress": "103.21.124.50"
  }
}
```

## Security & Compliance Notes
- **Redaction of Sensitive Secrets Prior to CDC Streaming:** Database write layers strictly enforce that plaintext bank account credentials, PAN numbers, Aadhaar identities, cryptographic private keys, or symmetric token encryption keys are never written to `outbox_events.payload`. All sensitive attributes must either be tokenized, masked, or encrypted using envelope encryption (AWS KMS / HashiCorp Vault) before insertion into the outbox table.
- **Idempotent Consumer Verification with Sequence Numbers:** To guarantee aggregate-level consistency under at-least-once delivery, every event includes a strictly increasing `sequence_number` scoped to `(aggregate_type, aggregate_id)`. Consumer handlers record the last processed sequence number in `inbox_events`. If an arriving event sequence is less than or equal to the recorded sequence, it is rejected as a duplicate; if a gap is detected, the event is held in a dead-letter retry buffer to prevent state divergence.
- **Principle of Least Privilege for PostgreSQL Replication User:** The `debezium_cdc` database user is granted replication privileges solely for logical decoding and `SELECT` access strictly on `eventing.outbox_events` and `eventing.debezium_signal`. It has zero read permissions on primary user, trading, identity, or custody tables, preventing unauthorized data exfiltration via replication protocols.
- **Replication Slot Saturation & WAL Leak Prevention:** An unconsumed PostgreSQL replication slot prevents PostgreSQL from truncating WAL logs, risking database disk exhaustion. Automated Prometheus alerts trigger if `pg_replication_slots.active` is false for longer than 3 minutes, or if `wal_keep_size` usage exceeds 70% threshold. Automated failover safeguards prevent database storage denial-of-service.
- **SEBI 8-Year Audit Trail & Event Immutability:** In compliance with SEBI operational guidelines and PMLA audit mandates, events produced by Debezium CDC into financial topics (`ledger.*`, `trading.*`, `settlement.*`) are mirrored to S3 cold storage via Kafka Connect S3 Sink using WORM (Write Once Read Many) Object Lock retention for a minimum of 8 years.
- **DPDP Act 2023 Compliance & Data Residency:** All PostgreSQL databases, Kafka Connect workers, Kafka brokers, and schema registries reside strictly within the AWS Asia Pacific (Mumbai) region (`ap-south-1`). Event payloads contain only pseudonymous internal identifiers (`user_id`, `account_id`, `wallet_id`), ensuring zero unencrypted customer PII is leaked across distributed event topics.

## Acceptance Criteria
- [ ] PostgreSQL 16 server configuration successfully enables logical replication with `wal_level = logical`, `max_replication_slots >= 20`, and `pgoutput` plugin.
- [ ] Universal `eventing.outbox_events` table schema successfully created with range partitioning on `created_at` and optimized composite index on `(aggregate_type, aggregate_id, sequence_number)`.
- [ ] Consumer `eventing.inbox_events` schema deployed with primary key `(event_id, consumer_group)` to enforce atomic message deduplication.
- [ ] Debezium 2.5+ PostgreSQL Connector deployed on Strimzi Kafka Connect cluster with active status and zero task failures.
- [ ] Logical publication `debezium_outbox_pub` exclusively captures `outbox_events` and `debezium_signal`, strictly isolating primary domain tables.
- [ ] Debezium Outbox Event Router SMT correctly extracts payload, routes events to dynamic destination topics (`destination_topic`), and assigns `aggregate_id` as Kafka partition key.
- [ ] Event headers correctly populated with `eventType`, `aggregateType`, `sequenceNumber`, `schemaVersion`, and `eventTimestamp`.
- [ ] Dedicated replication user `debezium_cdc` configured with least-privilege permissions, lacking access to core identity or business tables.
- [ ] Heartbeat mechanism generates periodic events on `__debezium-heartbeat` to advance replication slot LSN during low transactional periods.
- [ ] Signaling mechanism successfully executes ad-hoc incremental snapshots via `debezium_signal` table inserts without database locks.
- [ ] End-to-end CDC dispatch latency from local PostgreSQL commit to Kafka topic landing verified at $< 10\text{ms}$ under standard load.
- [ ] Consumer idempotency middleware successfully drops duplicate event deliveries while maintaining strict sequence ordering per aggregate.
- [ ] Automated daily partition maintenance via `pg_partman` successfully drops expired outbox partitions older than 7 days with zero vacuum locks.
- [ ] Replication slot monitoring alert fires within 60 seconds if a replication slot becomes inactive or WAL lag exceeds 500MB.
- [ ] Specification contains zero em dashes or en dashes throughout the document.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Event Schema and Kafka Topic Standards), Prompt 112 (Idempotency and Exactly-Once Processing), Prompt 401 (PostgreSQL Schema Design), Prompt 403 (Kafka Cluster and Topic Partitioning Strategy).
- **Core Producing Services:** Prompt 203 (Wallet & Account Service), Prompt 204 (Order Service), Prompt 208 (Trade Settlement Service), Prompt 212 (Payment Gateway Integration Service).
- **Core Consuming Services:** Prompt 209 (Portfolio and Holdings Service), Prompt 210 (Fee & Realized PnL Engine), Prompt 215 (Reconciliation Service), Prompt 216 (Regulatory Reporting Service), Prompt 218 (Immutable Audit Log Service).
- **Downstream Infrastructure Prompts:** Prompt 309 (Blockchain Event Indexer & State Sync), Prompt 404 (Data Warehouse & ClickHouse Analytics Pipeline).
