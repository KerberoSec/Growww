# 218 - Immutable Audit Log Service (High-Throughput Cryptographic Trail)

## Purpose
The Immutable Audit Log Service provides a tamper-evident, cryptographically chained, and legally non-repudiable record of every financial, administrative, operational, and authentication event occurring within the Growww ecosystem. In a regulated financial institution combining off-chain microservices with a permissioned ledger, regulatory bodies (SEBI, RBI, IFSCA) and statutory auditors require conclusive mathematical proof that historical logs have not been altered, truncated, or injected after the fact.

This service ingests millions of events per day across all distributed services, constructs continuous SHA-256 Merkle trees and hash chains (RFC 6962), archives raw payloads to WORM (Write Once Read Many) object storage, and periodically anchors hourly Merkle roots directly onto the permissioned Hyperledger Besu blockchain.

## What You Are Building
A high-throughput, high-assurance Rust / Go microservice (`services/audit-log-service`) providing:
- **Streaming Event Ingestion Engine:** Consumes audit events across all Kafka topics at sub-millisecond latency.
- **RFC 6962 Merkle Tree Chaining:** Computes continuous cryptographic hash chains over structured event blocks.
- **On-Chain Merkle Root Anchoring Relayer:** Periodically writes cryptographic epoch commitments to `AuditAnchorRegistry.sol` on Hyperledger Besu.
- **Tamper-Evidence & Verification API:** Allows internal compliance officers and external regulators to cryptographically verify any historical event against on-chain block receipts.
- **Artifacts Delivered:**
 - `services/audit-log-service/src/main.rs` (or `cmd/server/main.go`) - High-throughput service entry point.
 - `services/audit-log-service/src/merkle/tree.rs` - RFC 6962 Merkle tree builder and proof generator.
 - `services/audit-log-service/src/chain/anchor.rs` - Hyperledger Besu contract anchoring client.
 - `services/audit-log-service/src/storage/worm.rs` - S3 Object Lock / WORM archival client.
 - `proto/growww/audit/v1/audit.proto` - gRPC ingestion and verification contracts.

## Scope Boundaries
- **In Scope:**
 - Ingestion, validation, and schema enforcement of all microservice audit events.
 - Cryptographic hash chaining and Merkle tree generation.
 - Periodic on-chain root anchoring on Hyperledger Besu.
 - Providing Merkle inclusion proofs for specific audit event queries.
 - Long-term WORM archival integration (8-year retention).
- **Out of Scope / Handled Elsewhere:**
 - Application performance monitoring and metrics (handled by Prometheus/Grafana, Prompt 806).
 - Business intelligence / OLAP analytics (handled by Data Warehouse, Prompt 404).

## Technology to Use
- **Primary Language & Framework:** Rust (or Go 1.22+) using Tokio, `rdkafka`, `sha2`, and `alloy-rs` (or `go-ethereum`).
- **Justification:** Rust is ideal for the Audit Log Service due to its zero-cost abstractions, predictable memory footprint under sustained high throughput, absolute memory safety, and top-tier cryptographic hashing performance required to compute Merkle trees over tens of thousands of events per second without garbage collection pauses.
- **Dependencies & Libraries:**
 - ClickHouse (for high-speed analytical querying of billions of audit events) + PostgreSQL 16 (for anchoring epoch metadata).
 - Apache Kafka 3.7+ for consuming all cluster audit topics.
 - AWS S3 Object Lock / MinIO WORM storage.
 - `sha2` / `merkle-tree` cryptographic libraries.

## Backend / Infra Touchpoints
- **ClickHouse:** Columnar store indexing structured event attributes (actor, action, resource, timestamp, hash).
- **PostgreSQL 16:** Epoch metadata, anchor transaction receipts, and verification certificates.
- **MinIO / AWS S3 WORM:** Compressed, encrypted raw event archives locked against deletion.
- **Kafka Topics:**
 - Subscribes to: `*.audit`, `user.*`, `order.*`, `trade.*`, `admin.*`, `custody.*`.
- **Hyperledger Besu (QBFT):** Writes hourly Merkle root commitments to `AuditAnchorRegistry.sol`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium ledger running QBFT consensus.
- **Anchoring Contract:** Interacts with `AuditAnchorRegistry.sol` via HSM-signed transactions.
- **Epoch Commitment Mechanism:**
  1. Every 60 minutes (or every 100,000 events), an epoch is finalized.
  2. The service computes the root hash: $\text{Root} = \text{MerkleTree}(\text{Hash}(E_1), \dots, \text{Hash}(E_n))$.
  3. The service submits `anchorEpochRoot(uint256 epochId, bytes32 merkleRoot, uint64 eventCount)` to Besu.
- **Zero PII Guarantee:** Only SHA-256 cryptographic hashes and Merkle roots are committed to the public ledger. Raw payloads containing pseudonymized IDs remain in encrypted off-chain storage.

## Step-by-Step Build Instructions
1. Initialize Rust project under `services/audit-log-service` with Tokio async runtime and multi-worker thread pools.
2. Define Protobuf definitions in `proto/growww/audit/v1/audit.proto` and generate Rust / Go interfaces.
3. Configure ClickHouse table schemas for `audit_events` and PostgreSQL migrations for `audit_epochs`.
4. Implement the Kafka event consumer supporting high-throughput multi-partition fan-out.
5. Implement the PII Redaction & HMAC Anonymization pipeline stripping sensitive personal data before hashing.
6. Build the RFC 6962 compliant Merkle Tree engine capable of fast parallel leaf hashing using SIMD SHA-256.
7. Build the S3 / MinIO Object Lock storage adapter that uploads compressed hourly batch blocks with legal hold locks.
8. Implement the Hyperledger Besu blockchain relayer that submits epoch roots to `AuditAnchorRegistry.sol` via HSM.
9. Implement the Merkle Inclusion Proof generator API allowing verification of individual events ($\mathcal{O}(\log N)$ proof size).
10. Build the Tamper Verification background scanner that periodically verifies database records against on-chain anchored roots.
11. Add Prometheus metrics (`audit_events_ingested_total`, `merkle_tree_build_ms`, `anchor_tx_latency_ms`).
12. Construct end-to-end integration tests validating that bit-level payload tampering is immediately detected by the verification engine.

## Interfaces / Contracts

### Protobuf Definition (`audit.proto`)
```protobuf
syntax = "proto3";

package growww.audit.v1;

option go_package = "github.com/growww/services/audit-log/gen/v1;auditv1";

service AuditLogService {
  rpc IngestAuditEvent (AuditEventRequest) returns (AuditEventResponse);
  rpc GetEventInclusionProof (InclusionProofRequest) returns (InclusionProofResponse);
  rpc VerifyEventIntegrity (VerifyEventRequest) returns (VerifyEventResponse);
  rpc GetEpochStatus (EpochStatusRequest) returns (EpochStatusResponse);
}

message AuditEventRequest {
  string event_id = 1;
  string source_service = 2;
  string actor_id = 3; // Pseudonymized UUID or System Agent
  string action = 4; // e.g., ORDER_PLACED, ADMIN_FREEZE, SETTLEMENT_EXECUTED
  string resource_type = 5;
  string resource_id = 6;
  string payload_json = 7;
  int64 timestamp_ns = 8;
}

message AuditEventResponse {
  string event_id = 1;
  string event_hash = 2;
  bool queued = 3;
}

message InclusionProofRequest {
  string event_id = 1;
}

message InclusionProofResponse {
  string event_id = 1;
  string event_hash = 2;
  uint64 epoch_id = 3;
  string merkle_root = 4;
  repeated string audit_path = 5;
  uint64 leaf_index = 6;
  string on_chain_tx_hash = 7;
  uint64 block_number = 8;
}

message VerifyEventRequest {
  string event_id = 1;
  string payload_json = 2;
}

message VerifyEventResponse {
  bool is_valid = 1;
  string verification_message = 2;
  int64 on_chain_timestamp = 3;
}

message EpochStatusRequest {
  uint64 epoch_id = 1;
}

message EpochStatusResponse {
  uint64 epoch_id = 1;
  string merkle_root = 2;
  uint64 event_count = 3;
  string status = 4; // OPEN / SEALED / ANCHORED
  string on_chain_tx_hash = 5;
}
```

### ClickHouse & PostgreSQL Database Schemas
```sql
-- ClickHouse Schema for Audit Events
CREATE TABLE IF NOT EXISTS audit_events (
    event_id UUID,
    epoch_id UInt64,
    source_service LowCardinality(String),
    actor_id String,
    action LowCardinality(String),
    resource_type LowCardinality(String),
    resource_id String,
    event_hash FixedString(64),
    payload_json String,
    created_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
ORDER BY (source_service, action, created_at, event_id);

-- PostgreSQL Schema for Anchored Epochs
CREATE TABLE audit_epochs (
    epoch_id BIGSERIAL PRIMARY KEY,
    merkle_root VARCHAR(64) NOT NULL,
    event_count BIGINT NOT NULL,
    s3_archive_path VARCHAR(256) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'SEALED',
    on_chain_tx_hash VARCHAR(66),
    block_number BIGINT,
    sealed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    anchored_at TIMESTAMPTZ
);
```

## Security & Compliance Notes
- **Tamper-Evident Guarantees:** Once an epoch Merkle root is committed to Hyperledger Besu with QBFT consensus finality, altering any single historic event produces a mathematical hash mismatch.
- **DPDP Act (Digital Personal Data Protection) Privacy Compliance:** No plain-text investor PII is stored in audit logs. User identifiers are transformed into cryptographic HMAC pseudonyms using isolated HSM master salt keys.
- **WORM Storage Enforcement:** MinIO/S3 buckets operate under Strict Compliance Mode preventing object deletion or modification even by root cloud credentials for 8 years.
- **Non-Repudiation for Regulators:** Merkle proofs generated by this service are admissible as certified electronic evidence under Section 65B of the Indian Evidence Act.

## Acceptance Criteria
- [ ] Service compiles cleanly in release mode with zero memory leaks and zero data race warnings.
- [ ] Successfully ingests 50,000+ Kafka audit events per second on standard 4-core worker nodes.
- [ ] Merkle tree calculation strictly conforms to RFC 6962 standards.
- [ ] Epoch roots are automatically committed to `AuditAnchorRegistry.sol` on test Besu ledger every epoch interval.
- [ ] Verification API successfully validates inclusion proofs against live blockchain receipts in sub-10ms.
- [ ] Injected corrupted event payload immediately fails cryptographic verification.
- [ ] Automated test suite achieves >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 104 (Kafka Standards), Prompt 109 (Secrets & HSM), Prompt 302 (Besu Node Setup).
- **Subsequent / Parallel Tasks:** Prompt 216 (Reporting Service), Prompt 217 (Admin Service).
