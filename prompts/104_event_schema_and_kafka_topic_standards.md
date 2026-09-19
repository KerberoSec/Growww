# 104 - Event Schema & Kafka Topic Naming Standards (CloudEvents, Schema Registry & Partitioning)

## Purpose
Establishes the enterprise asynchronous event-driven messaging standards, Kafka topic taxonomy, event schema lifecycle, and partition key strategies for the Growww investment platform. As a real-time financial trading and DvP settlement system, Growww relies on Apache Kafka for high-throughput, ordered event streaming between order matching, risk evaluation, custodial sync, blockchain relayers, and regulatory audit pipelines.

This prompt provides developers with the canonical specifications for CloudEvents v1.0 compliance, Confluent Schema Registry configuration, deterministic partition key assignment to ensure strict in-order processing, Dead-Letter Queue (DLQ) retry topologies, and schema evolution compatibility rules.

## What You Are Building
A comprehensive event streaming standards specification (`docs/standards/event_and_kafka_standards.md`), base CloudEvents Protobuf schemas, Schema Registry configuration, and Kafka topic catalog:
- Kafka Topic Naming Taxonomy: Formal standard `<entity>.<domain>.<dataset>.<event-type>.<version>` (e.g., `domestic.trading.orders.matched.v1`).
- CloudEvents v1.0 Specification: Base Protobuf schema defining standard event headers (`id`, `source`, `specversion`, `type`, `time`, `datacontenttype`, `subject`) and custom Growww extension attributes (`correlationid`, `actorid`, `entityid`).
- Partitioning & Ordering Strategy: Deterministic partition key standards guaranteeing per-user and per-security FIFO event ordering.
- Schema Evolution & Compatibility Policy: Confluent Schema Registry rules enforcing `BACKWARD_TRANSITIVE` compatibility.
- Error Handling & Dead Letter Queue (DLQ) Topology: Standardized retry, exponential backoff, and poison-pill routing standards.

## Scope Boundaries
- **In Scope:**
 - Topic naming conventions, environment prefixes, and domain taxonomy.
 - Event envelope schemas (CloudEvents v1.0 in Protobuf / Avro / JSON).
 - Partition key selection matrix for ordering guarantees.
 - Schema Registry configuration, CI validation rules, and backward compatibility gates.
 - Dead Letter Queue (DLQ) architecture and poison-message handling.
- **Out of Scope / Handled Elsewhere:**
 - Synchronous gRPC / REST API design (handled in Prompt 103).
 - Kafka cluster infrastructure setup, replication, and Kubernetes operator deployment (handled in Prompt 403 & 802).
 - Specific business event payloads (handled in Category 2 & 3).

## Technology to Use
- **Messaging Backbone:** Apache Kafka 3.7+ with KRaft consensus.
- **Schema Management:** Confluent Schema Registry / Karapace with Protobuf and Avro serialization formats.
- **Event Standard:** CNCF CloudEvents Specification v1.0.
- **Client Libraries:**
 - *Go:* `confluent-kafka-go` (librdkafka wrapper) with strict idempotency and schema registry serializers.
 - *Rust:* `rdkafka` asynchronous client with zero-copy deserialization for high-throughput matching and settlement streams.
 - *Python:* `confluent-kafka-python` with FastAvro/Protobuf integration for background worker tasks and audit consumers.

## Backend / Infra Touchpoints
- **Schema Registry Cluster:** Central schema governance endpoint integrated into CI/CD pipelines.
- **Kafka Brokers:** Multi-broker, multi-AZ deployment with TLS encryption and SASL/SCRAM authentication.
- **Debezium CDC Connectors:** PostgreSQL change-data-capture streaming database changes to CDC topics (`cdc.<domain>.<table_name>`).
- **Kafka Connect / Sink Connectors:** Streaming audit topics into S3/GCS data lake for long-term cold storage.

## Blockchain Interaction
Standardizes on-chain event ingestion and off-chain streaming:
- **Blockchain Event Ingestion Topic:** `domestic.ledger.besu.events.v1` receives raw event logs emitted by smart contracts (`DvPExecuted`, `SecurityMinted`, `SecurityBurned`, `KYCClaimUpdated`, `ProofOfReservePublished`).
- **Standardized Blockchain Event Envelope:** Extracted smart contract logs are normalized into CloudEvents envelopes including `block_number`, `transaction_hash`, `log_index`, and `contract_address` as metadata extensions.
- **Reconciliation Consumers:** Financial reconciliation services consume the blockchain event stream to perform automated two-way matching between off-chain database transactions and immutable ledger states.

## Step-by-Step Build Instructions
1. Initialize directory `proto/events/v1/` and documentation file `docs/standards/event_and_kafka_standards.md`.
2. Define the formal topic naming grammar: `<env>.<entity>.<bounded-context>.<aggregate>.<event-name>.<version>`.
 - *Example:* `prod.domestic.trading.order.matched.v1`
 - *Example:* `prod.giftcity.settlement.dvp.executed.v1`
 - *Example:* `prod.domestic.compliance.investor.whitelisted.v1`
3. Create the canonical CloudEvents Protobuf schema under `proto/events/v1/cloudevent.proto`.
4. Define standard partition key strategies:
 - *User Actions (Orders, Balances, KYC):* Partition by `user_id` to guarantee per-user sequential processing.
 - *Market Actions (Order Matching, Price Updates):* Partition by `security_isin` (e.g., `INE002A01018`) to ensure deterministic matching order per instrument.
 - *Settlement Actions:* Partition by `settlement_id`.
5. Establish topic configuration standards:
 - *Core Trading & Settlement:* `min.insync.replicas=2`, `replication.factor=3`, `cleanup.policy=delete`, `retention.ms=604800000` (7 days).
 - *Audit & Ledger Events:* `cleanup.policy=compact,delete`, `retention.ms=31536000000` (365 days / indefinite with archival).
6. Design the Dead Letter Queue (DLQ) topology:
 - Main topic: `<topic-name>`
 - Retry topic: `<topic-name>.retry` (with exponential backoff header tracking)
 - Dead Letter topic: `<topic-name>.dlq` (for poison pills requiring human/ops intervention)
7. Author `proto/events/v1/dlq_envelope.proto` containing original payload, failure reason, stack trace, attempt count, and timestamp.
8. Configure Confluent Schema Registry compatibility mode to `BACKWARD_TRANSITIVE`, preventing any schema change that breaks existing consumers.
9. Define CI validation check using Buf and Schema Registry CLI to verify new event schemas during pull requests before merge.
10. Formulate consumer group naming conventions: `<service-name>.<bounded-context>.<purpose>` (e.g., `settlement-service.trading.dvp-processor`).
11. Document producer and consumer configuration requirements: `enable.idempotence=true`, `acks=all`, `max.in.flight.requests.per.connection=5`.
12. Review and publish `docs/standards/event_and_kafka_standards.md` to the architecture repository.

## Interfaces / Contracts

### Base CloudEvents Protobuf Schema (`proto/events/v1/cloudevent.proto`)
```protobuf
syntax = "proto3";

package growww.events.v1;

import "google/protobuf/any.proto";
import "google/protobuf/timestamp.proto";

option go_package = "github.com/growww/proto/gen/go/events/v1;eventsv1";

message CloudEvent {
  // CloudEvents Core Attributes (v1.0)
  string id = 1;                     // Unique event UUIDv7
  string source = 2;                 // URI identifying producer service (e.g., "/services/matching-engine")
  string spec_version = 3;           // Must be "1.0"
  string type = 4;                   // Event type URI (e.g., "growww.trading.order.matched.v1")
  google.protobuf.Timestamp time = 5;// Event generation timestamp
  string datacontenttype = 6;        // e.g., "application/x-protobuf" or "application/json"
  string subject = 7;                // Target resource ID (e.g., "order_01HZX89AB...")

  // Growww Enterprise Extension Attributes
  string correlation_id = 8;         // Distributed tracing identifier
  string actor_id = 9;               // User ID or Service Account triggering the event
  string entity_id = 10;             // "DOMESTIC" or "GIFT_CITY"
  string idempotency_key = 11;       // Unique deduplication key

  // Business Payload
  google.protobuf.Any data = 12;
}
```

### Dead Letter Queue Envelope (`proto/events/v1/dlq_envelope.proto`)
```protobuf
syntax = "proto3";

package growww.events.v1;

import "google/protobuf/timestamp.proto";
import "proto/events/v1/cloudevent.proto";

message DeadLetterQueueEnvelope {
  string dlq_id = 1;
  CloudEvent original_event = 2;
  string consumer_group = 3;
  string error_message = 4;
  string error_code = 5;
  string error_stack_trace = 6;
  int32 retry_count = 7;
  google.protobuf.Timestamp failed_at = 8;
  map<string, string> diagnostic_headers = 9;
}
```

### Topic Taxonomy Matrix
| Domain / Context | Event Name | Topic Name | Partition Key | Retention |
|---|---|---|---|---|
| **Trading** | `OrderPlaced` | `domestic.trading.orders.placed.v1` | `user_id` | 7 Days |
| **Trading** | `OrderMatched` | `domestic.trading.orders.matched.v1` | `security_isin` | 7 Days |
| **Settlement** | `DvPInitiated` | `domestic.settlement.dvp.initiated.v1` | `settlement_id` | 30 Days |
| **Settlement** | `DvPCompleted` | `domestic.settlement.dvp.completed.v1` | `settlement_id` | 365 Days |
| **Blockchain** | `LedgerEvent` | `domestic.ledger.besu.events.v1` | `tx_hash` | Indefinite (Archival) |
| **Custody** | `CustodySynced` | `domestic.custody.positions.synced.v1` | `security_isin` | Indefinite |
| **Compliance** | `KYCApproved` | `domestic.compliance.kyc.approved.v1` | `user_id` | 365 Days |

## Security & Compliance Notes
- **PII Protection & Crypto-Shredding:** No sensitive investor PII (Aadhaar, PAN, bank account numbers) is serialized into event payloads. If user-identifiable data is required, it must be encrypted with a dedicated per-user encryption key stored in KMS; deleting the key renders historical event data unreadable (crypto-shredding compliance under DPDP Act).
- **Topic-Level Access Control (ACLs):** Kafka brokers enforce strict mTLS certificates and ACLs preventing unauthorized microservices from publishing or subscribing to restricted financial topics.
- **Audit Logging Immutability:** Event streams relating to orders, trades, and settlements are mirrored to an immutable WORM (Write Once Read Many) compliance S3 bucket for SEBI regulatory inspection.

## Acceptance Criteria
- [ ] Comprehensive event standards document (`docs/standards/event_and_kafka_standards.md`) published and approved.
- [ ] Canonical `cloudevent.proto` and `dlq_envelope.proto` schemas authored and verified with Buf.
- [ ] Complete topic taxonomy and partition key strategy documented for all 10 bounded contexts.
- [ ] Schema Registry compatibility rules (`BACKWARD_TRANSITIVE`) configured and automated CI schema check established.
- [ ] DLQ retry and poison-pill routing pattern formalized with code examples in Go and Python.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 103 (API Standards).
- **Parallel Work:** Prompt 105 (Auth Architecture), Prompt 111 (Domain Model), Prompt 112 (Idempotency).
- **Blocks:** Prompt 403 (Kafka Cluster Design), Category 2 (All Event Producers & Consumers), Category 3 (Event Indexer).
