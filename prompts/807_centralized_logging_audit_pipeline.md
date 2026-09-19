# 807 - Centralized Logging, Immutable Audit Trail & Regulatory SIEM Pipeline

## Purpose
Securities and financial infrastructure operations in India are governed by stringent regulatory mandates from SEBI, RBI, and the Digital Personal Data Protection (DPDP) Act. These regulations require complete, tamper-proof, immutable audit trails of all system access, administrative overrides, KYC verifications, order placements, trade executions, and fund/token transfers, preserved for a mandatory minimum of 7 to 8 years.

This prompt establishes Growww's centralized logging, immutable audit lake, and Security Information and Event Management (SIEM) ingestion pipeline. By combining high-throughput log collection via Vector / Fluent Bit, an isolated Kafka audit event bus, cold storage in WORM (Write-Once-Read-Many) compliant object vaults (AWS S3 Glacier Object Lock in Compliance Mode), and periodic on-chain cryptographic Merkle root anchoring on the Hyperledger Besu ledger, the system guarantees mathematical proof of log integrity and zero data tampering.

## What You Are Building
An enterprise-scale, tamper-evident audit logging and SIEM ingestion pipeline:
- `deployments/k8s/logging/vector-agent.yaml`: DaemonSet configuration for Vector log collection agents capturing stdout/stderr from all pods with metadata enrichment (K8s pod, namespace, container, commit SHA).
- `deployments/k8s/logging/vector-aggregator.yaml`: Scalable Vector aggregator pipeline parsing structured JSON logs, enforcing field-level PII encryption, and routing streams to OpenSearch (hot), SIEM (real-time), and WORM S3 (immutable archive).
- `services/audit-anchoring-worker/`: Go-based background daemon that aggregates hourly batches of audit logs, computes a SHA-256 Merkle tree root, and writes the Merkle root to the on-chain `ProofOfReserveRegistry.sol` / `AuditAnchorContract.sol`.
- `infra/terraform/modules/worm-storage/`: Terraform module provisioning AWS S3 Glacier buckets with Object Lock in `COMPLIANCE` mode (7-year retention period with strict legal hold capabilities).
- `deployments/siem/wazuh-rules/`: SIEM detection rules and alerting patterns for detecting unauthorized admin actions, repeated failed authentications, and abnormal token minting attempts.

## Scope Boundaries
- **In Scope:**
 - Node-level and cluster-level log collection across all Kubernetes namespaces.
 - Dedicated, isolated Kafka topics for compliance and security events (`audit.system.v1`, `audit.security.v1`, `audit.trading.v1`).
 - Field-level encryption (FLE) for sensitive investor data before storage.
 - Immutable WORM storage provisioning on AWS S3 Glacier Object Lock (Compliance Mode).
 - Cryptographic hourly Merkle root generation and on-chain ledger anchoring.
 - Real-time SIEM integration (Wazuh / Splunk / Elastic Security).
- **Out of Scope / Handled Elsewhere:**
 - Operational application metrics and trace visualization (Prompt 806).
 - Microservice audit log event emission interfaces (Prompt 218).
 - Business P&L calculation and tax statement generation (Prompt 210, 223).

## Technology to Use
- **Vector (v0.38+)**: High-performance, memory-safe observability data pipeline written in Rust. Justification: Consumes 10x less memory and CPU than traditional Java/Ruby log agents (e.g. Logstash/Fluentd), offers native Vector Remap Language (VRL) for complex transformation and field-level encryption, and guarantees zero data loss via disk-backed buffers.
- **AWS S3 Glacier Object Lock (Compliance Mode)**: Immutable WORM storage. Justification: Legally certified compliance storage preventing object deletion or modification even by the AWS root account during the 7-year retention window.
- **OpenSearch 2.13+**: Scalable distributed search and analytics engine for operational hot log analysis (30-day retention).
- **Wazuh SIEM**: Open-source enterprise security monitoring and compliance platform.
- **Go (v1.22+)**: Used for the lightweight cryptographic Merkle anchoring worker.

## Backend / Infra Touchpoints
- **Kubernetes Pod Logs**: Captured via `/var/log/pods` mount points on all nodes.
- **Kafka Cluster**: Dedicated audit cluster with mTLS client authentication.
- **AWS S3 Object Lock**: Target storage bucket in Mumbai region (`growww-prod-immutable-audit-logs`).
- **OpenSearch Cluster**: Target hot/warm search cluster.
- **Hyperledger Besu RPC**: Target blockchain endpoint for on-chain Merkle root commits.

## Blockchain Interaction
The audit anchoring worker permanently anchors off-chain audit logs to the permissioned Hyperledger Besu blockchain:
- **Hourly Merkle Root Anchoring**: Every hour, the worker reads all finalized audit log chunks stored in S3, computes a SHA-256 Merkle tree root representing all log entries in that hour window, and invokes `anchorAuditBatch(uint256 timestamp, bytes32 merkleRoot, string s3Uri, uint256 logCount)` on the on-chain `AuditAnchorContract.sol`.
- **Tamper Verification**: External auditors or SEBI inspectors can verify any individual log entry by generating its Merkle proof and verifying it against the immutable on-chain root recorded on the permissioned ledger.
- **Transaction Hash Logging**: The resulting Ethereum transaction hash (`tx_hash`) and block number are saved back to the metadata manifest associated with the archived S3 log bundle.

## Step-by-Step Build Instructions
1. Scaffold directory `deployments/k8s/logging/`, `services/audit-anchoring-worker/`, and `infra/terraform/modules/worm-storage/`.
2. Write Terraform manifests provisioning the S3 bucket with Object Lock enabled in `COMPLIANCE` mode with a retention duration of 2,555 days (7 years).
3. Deploy Vector agents as a Kubernetes `DaemonSet` on every node with host mounts to `/var/log/pods`.
4. Deploy Vector aggregator as an auto-scaling `Deployment` with persistent NVMe disk buffers (`data_dir`) to prevent log drops during traffic surges.
5. Write Vector Remap Language (VRL) transformation scripts to parse structured JSON, tag ISO-8601 timestamps, extract `tenant_id`, and mask/encrypt investor PII fields using AES-256-GCM keys managed in HashiCorp Vault.
6. Configure Vector sinks: Sink 1 (Kafka `audit.events.v1`), Sink 2 (OpenSearch hot index), Sink 3 (AWS S3 Glacier WORM bucket with gzip compression and SHA-256 manifest generation).
7. Deploy Wazuh SIEM forwarder subscribing to Kafka `audit.security.v1` topic.
8. Implement Wazuh SIEM detection rules for privilege escalation, administrative token adjustments, and out-of-hours database access.
9. Implement the Go `audit-anchoring-worker` service: queries S3 bucket manifests hourly, builds the cryptographic Merkle tree, and submits the root hash to `AuditAnchorContract.sol` via HSM relayer.
10. Implement verification CLI tool `cmd/verify-audit-log` allowing auditors to supply a raw log line, generate a Merkle inclusion proof, and verify against the Besu blockchain RPC.
11. Test the WORM configuration by attempting an explicit `aws s3 rm` command using admin credentials, verifying the AWS Object Lock rejects the deletion with `AccessDenied`.
12. Simulate high-throughput log ingestion (50,000 events/sec) to verify Vector disk buffers prevent any dropped records.
13. Document compliance verification procedures for SEBI regulatory inspections.

## Interfaces / Contracts
```json
// Structured Audit Log JSON Schema
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "GrowwwAuditLogEvent",
  "type": "object",
  "required": ["event_id", "timestamp", "actor", "action", "resource", "status", "context"],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "timestamp": { "type": "string", "format": "date-time" },
    "actor": {
      "type": "object",
      "required": ["id", "type", "ip_address"],
      "properties": {
        "id": { "type": "string" },
        "type": { "type": "string", "enum": ["USER", "ADMIN", "SYSTEM_WORKER", "CUSTODIAN_GATEWAY"] },
        "ip_address": { "type": "string" },
        "session_id": { "type": "string" }
      }
    },
    "action": { "type": "string" },
    "resource": {
      "type": "object",
      "required": ["type", "id"],
      "properties": {
        "type": { "type": "string" },
        "id": { "type": "string" }
      }
    },
    "status": { "type": "string", "enum": ["SUCCESS", "FAILURE", "BLOCKED"] },
    "context": { "type": "object" }
  }
}
```

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IAuditAnchorContract {
    event AuditBatchAnchored(
        uint256 indexed batchId,
        uint256 indexed timestamp,
        bytes32 indexed merkleRoot,
        string s3Uri,
        uint256 logCount
    );

    function anchorAuditBatch(
        uint256 timestamp,
        bytes32 merkleRoot,
        string calldata s3Uri,
        uint256 logCount
    ) external returns (uint256 batchId);

    function verifyLogEntry(
        uint256 batchId,
        bytes32 leafHash,
        bytes32[] calldata merkleProof
    ) external view returns (bool);
}
```

## Security & Compliance Notes
- 7-Year Immutable Retention: Mandated by SEBI regulations for all financial transactions and trading logs; AWS S3 Object Lock in Compliance Mode strictly prohibits deletion or lifecycle reduction by any IAM entity.
- Field-Level Encryption (FLE): Investor PII is encrypted before being sent to Elasticsearch/OpenSearch; only authorized compliance officers possessing Vault decryption tokens can view decrypted identity details.
- Tamper-Proof Anchoring: Anchoring Merkle roots on the permissioned Hyperledger Besu consortium ledger prevents retrofitted or rewritten logs from being accepted during regulatory audits.

## Acceptance Criteria
- [ ] Vector agents capture 100% of container stdout/stderr logs and route them to Kafka, OpenSearch, and S3.
- [ ] AWS S3 Glacier Object Lock in Compliance Mode successfully prevents deletion of test audit files even with admin credentials.
- [ ] Sensitive PII fields (Aadhaar, PAN, Bank Details) are verified encrypted or redacted before storage.
- [ ] Go audit anchoring worker computes hourly Merkle tree roots and writes them to `AuditAnchorContract.sol` on the Besu ledger.
- [ ] Verification tool `cmd/verify-audit-log` mathematically verifies an individual log line against the on-chain Merkle root with a valid cryptographic proof.
- [ ] Wazuh SIEM triggers real-time alerts upon simulated administrative security violations.

## Suggested Order / Dependencies
- Prerequisites: Prompt 008 (Data Protection Policy), Prompt 104 (Kafka Standards), Prompt 218 (Audit Log Service), Prompt 802 (Kubernetes Architecture).
- Parallel Tasks: Prompt 806 (Observability), Prompt 808 (Alerting & On-Call).
