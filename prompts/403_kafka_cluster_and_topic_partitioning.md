# 403 - Kafka Cluster Design & Topic Partitioning Strategy

## Purpose
The event-driven financial architecture of Growww requires a fault-tolerant, horizontally scalable, and distributed messaging backbone. Apache Kafka serves as the central nervous system, decoupling synchronous microservices and choreographing critical workflows such as order lifecycle events, trade matching feeds, Delivery-versus-Payment (DvP) settlements, ledger postings, blockchain event fan-out, and regulatory audit streams.

This prompt defines the Kafka cluster topology, KRaft consensus metadata configuration, partition key routing strategies, retention policies, Confluent/Apicurio Schema Registry integration, and Dead Letter Queue (DLQ) patterns. It ensures strict partition-level event ordering - critical for maintaining deterministic market books per ISIN and immutable user account balances - while supporting multi-thousand messages per second with sub-10ms delivery latency.

## What You Are Building
A production-grade Apache Kafka 3.7+ infrastructure and topic governance framework containing:
- **KRaft Cluster Manifests:** Strimzi Kafka Operator manifests (`infra/kafka/cluster/`) deploying a 5-broker Kafka cluster with KRaft metadata mode across 3 Availability Zones.
- **Declarative Topic Specifications:** Kubernetes Custom Resources (`KafkaTopic`) specifying partitions, replication factor (RF=3, min.insync.replicas=2), log compaction, and tiered retention across all bounded contexts.
- **Partition Key Routing Matrix:** Explicit partitioning standard assigning partition keys based on domain invariants (e.g. `isin` for market trades, `user_id` for ledger journals).
- **Schema Registry Deployment & Governance:** Automated Schema Registry infrastructure enforcing strict Avro / Protobuf schema validation with `BACKWARD_TRANSITIVE` compatibility rules.
- **DLQ & Error Retry Pipeline:** Standardized Dead Letter Queue (`<topic>.dlq`) and exponential backoff retry topic topology (`<topic>.retry.1m`, `<topic>.retry.5m`).
- **Kafka Exporter & Observability:** Prometheus JMX Exporter configurations tracking consumer lag, under-replicated partitions, leader election rates, and byte throughput.

## Scope Boundaries
- **In Scope:** Kafka broker cluster configuration, KRaft metadata quorum, Strimzi CRDs, topic topologies, partition count formulas, retention rules, Schema Registry setup, and DLQ architecture.
- **Out of Scope / Handled Elsewhere:**
 - Standard event naming conventions and envelope schemas (handled in Prompt 104).
 - Microservice-specific producer and consumer implementations (handled in Category 2 microservices).
 - Real-time Redis streaming for client WebSockets (handled in Prompt 402 and Prompt 207).
 - ClickHouse Kafka Connect ingestion pipelines (handled in Prompt 404).

## Technology to Use
Apache Kafka 3.7+ in KRaft (Kafka Raft) metadata mode managed via the Strimzi Kubernetes Operator is selected. Kafka provides distributed log durability, linear read/write scalability through partitioning, high sequential I/O throughput, and rock-solid consumer group offset management. KRaft mode eliminates ZooKeeper dependency, providing faster metadata recovery, simplified cluster operations, and support for millions of partitions.

- **Messaging Engine:** Apache Kafka 3.7+ (KRaft consensus mode).
- **Kubernetes Operator:** Strimzi Kafka Operator v0.40+.
- **Schema Governance:** Confluent Schema Registry / Karapace v7.5+.
- **Storage Tiering:** NVMe / AWS EBS `gp3` storage with IOPS tuning for commit logs.
- **Client Protocol:** SASL_SSL over TLS 1.3 with SCRAM-SHA-512 authentication.

## Backend / Infra Touchpoints
- **Kubernetes Cluster:** Dedicated Kafka worker node pool with anti-affinity rules spreading brokers across 3 Availability Zones.
- **Schema Registry:** High-availability cluster backed by the internal `_schemas` Kafka topic.
- **HashiCorp Vault:** Dynamic credential generation for Kafka user ACLs and SASL/SCRAM secrets.
- **Prometheus & Alertmanager:** JMX exporter scraping broker JVM metrics and consumer group lag.

## Blockchain Interaction
Kafka acts as the high-throughput bridge connecting the permissioned Hyperledger Besu consortium blockchain to off-chain microservices.

### Detailed On-Chain Integration Mechanics:
- **On-Chain Event Ingestion:** The Hyperledger Besu Event Indexer (Prompt 309) listens to QBFT block logs and smart contract events (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`) and publishes them directly to Kafka topic `blockchain.besu.events`.
- **Fan-Out & Reconciliation:** Specialized consumers (Prompt 215 Reconciliation Service, Prompt 208 DvP Settlement Service) subscribe to `blockchain.besu.events` to reconcile token mint/burn events with physical NSDL/CDSL depository receipts.
- **Tx Relayer Queue:** High-priority topic `blockchain.besu.tx_submission` buffers transactions destined for HSM signing and submission to Besu validator RPC nodes, preventing node saturation during high volatility.
- **Zero PII Guarantee:** All event payloads passing through Kafka contain zero unencrypted investor PII. Payloads reference only pseudonymous Ethereum addresses (`0x...`), ISINs, and UUIDs.

## Step-by-Step Build Instructions
1. Deploy Strimzi Kafka Operator into the Kubernetes cluster (`kafka` namespace).
2. Configure 5-node Kafka cluster CR (`Kafka` CRD) with 3 KRaft controllers and 5 combined broker nodes spread across 3 AZs.
3. Configure persistent volume claims with NVMe/gp3 storage and XFS filesystem for optimal commit log write throughput.
4. Enable TLS 1.3 listener with SASL/SCRAM authentication for internal microservice communication.
5. Deploy Confluent Schema Registry / Karapace with multi-replica high availability.
6. Define and apply core domain `KafkaTopic` CRDs with calculated partition counts and retention policies.
7. Configure log compaction on stateful snapshot topics (`market.securities_master`, `compliance.investor_status`).
8. Establish tiered retention: 7 days on high-throughput operational topics (`market.orders.intake`), 30 days on financial settlement topics (`trading.settlement.dvp`).
9. Implement standardized Dead Letter Queue (DLQ) and retry topic hierarchy (`<topic>.retry.<interval>`, `<topic>.dlq`).
10. Configure Kafka ACLs via `KafkaUser` CRDs granting least-privilege topic read/write permissions per microservice.
11. Deploy Prometheus Kafka Exporter and JMX Exporter sidecars.
12. Configure Grafana dashboards and Alertmanager rules for consumer group lag exceeding 1,000 records and under-replicated partition count > 0.
13. Execute chaos test validating zero message loss and automatic partition leader failover during simulated broker node termination.

## Interfaces / Contracts

```yaml
# Strimzi Kafka Cluster CR Manifest (Snippet)
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: growww-kafka-cluster
  namespace: kafka
spec:
  kafka:
    version: 3.7.0
    replicas: 5
    listeners:
 - name: tls
        port: 9093
        type: internal
        tls: true
        authentication:
          type: scram-sha-512
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      default.replication.factor: 3
      min.insync.replicas: 2
      compression.type: zstd
      log.cleanup.policy: delete
      unclean.leader.election.enable: false
      auto.create.topics.enable: false
    storage:
      type: persistent-claim
      size: 500Gi
      class: gp3-nvme
```

```yaml
# Topic Configuration Specifications & Partition Key Matrix
Topics:
 - name: "trading.orders.placed"
    partitions: 32
    replicationFactor: 3
    minIsr: 2
    partitionKey: "isin" # Ensures strict price-time order sequence per security
    cleanupPolicy: "delete"
    retentionMs: 604800000 # 7 days

 - name: "trading.trades.executed"
    partitions: 32
    replicationFactor: 3
    minIsr: 2
    partitionKey: "isin" # Matches order book partition key
    cleanupPolicy: "delete"
    retentionMs: 2592000000 # 30 days

 - name: "ledger.journal.postings"
    partitions: 64
    replicationFactor: 3
    minIsr: 2
    partitionKey: "user_id" # Guarantees strict sequential balance updates per user
    cleanupPolicy: "delete"
    retentionMs: 2592000000 # 30 days

 - name: "blockchain.besu.events"
    partitions: 16
    replicationFactor: 3
    minIsr: 2
    partitionKey: "contract_address"
    cleanupPolicy: "delete"
    retentionMs: 2592000000 # 30 days

 - name: "compliance.investor.status"
    partitions: 16
    replicationFactor: 3
    minIsr: 2
    partitionKey: "blockchain_address"
    cleanupPolicy: "compact" # Stateful latest snapshot per investor
```

```json
// Sample Avro Schema: TradeExecutedEvent.avsc
{
  "type": "record",
  "name": "TradeExecutedEvent",
  "namespace": "in.growww.trading.events",
  "doc": "Emitted by matching engine upon trade execution",
  "fields": [
    {"name": "eventId", "type": "string", "logicalType": "uuid"},
    {"name": "tradeId", "type": "string", "logicalType": "uuid"},
    {"name": "isin", "type": "string"},
    {"name": "symbol", "type": "string"},
    {"name": "tokenAddress", "type": "string"},
    {"name": "buyerOrderId", "type": "string", "logicalType": "uuid"},
    {"name": "sellerOrderId", "type": "string", "logicalType": "uuid"},
    {"name": "buyerAddress", "type": "string"},
    {"name": "sellerAddress", "type": "string"},
    {"name": "fractionalUnits", "type": "string", "doc": "Stringified decimal (up to 6 decimals)"},
    {"name": "pricePerUnitINR", "type": "string", "doc": "Stringified decimal (up to 4 decimals)"},
    {"name": "grossAmountINR", "type": "string"},
    {"name": "feeAmountINR", "type": "string"},
    {"name": "executedAt", "type": "long", "logicalType": "timestamp-millis"}
  ]
}
```

## Security & Compliance Notes
- **Mutual TLS & SASL/SCRAM:** All inter-broker and client connections mandate TLS 1.3 with SASL/SCRAM-SHA-512 authentication. Plaintext listeners (`PLAINTEXT`) are strictly disabled.
- **Zero PII Policy:** Investor identification data (names, PAN, Aadhaar) must NEVER be placed in Kafka message headers, keys, or bodies. Only pseudonymous UUIDs and wallet addresses are permitted.
- **Unclean Leader Election Disabled:** `unclean.leader.election.enable` is hardcoded to `false` to guarantee zero message loss during broker failures, satisfying financial transaction audit rules.
- **Egress & Ingress Network Policies:** Kubernetes NetworkPolicies restrict Kafka broker ports (9092, 9093) to authenticated microservice pods only.

## Acceptance Criteria
- [ ] 5-node Apache Kafka cluster running on KRaft mode successfully deployed via Strimzi Operator on Kubernetes.
- [ ] Automated topic provisioning verified with `min.insync.replicas=2` and `unclean.leader.election.enable=false`.
- [ ] Confluent / Karapace Schema Registry operational with Avro schema validation and `BACKWARD_TRANSITIVE` evolution rules.
- [ ] Partition key routing tested: all messages for a specific `isin` route deterministically to the same partition.
- [ ] Standardized DLQ and exponential retry pipeline verified with simulated consumer processing exceptions.
- [ ] High-throughput benchmark achieves >25,000 msg/sec with end-to-end publish-to-consume latency < 15ms.
- [ ] Simulated broker node outage causes zero message loss and seamless client reconnection.
- [ ] Prometheus metrics export consumer lag and partition ISR health with active Alertmanager alerts.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Kafka Standards), Prompt 112 (Idempotency), Prompt 802 (Kubernetes Infrastructure).
- **Parallel Tasks:** Prompt 401 (PostgreSQL), Prompt 402 (Redis), Prompt 404 (Data Warehouse / ClickHouse).
