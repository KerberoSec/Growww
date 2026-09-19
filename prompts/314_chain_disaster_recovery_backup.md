# 314 - Blockchain Ledger Disaster Recovery, RocksDB Snapshotting & Node Failover

## Purpose
Financial market infrastructure mandates high-assurance business continuity planning (BCP) and disaster recovery (DR) capabilities complying with SEBI and RBI Operational Risk frameworks. A catastrophic event - such as a cloud region outage, storage corruption, ransomware attack, or physical data center failure - must never compromise the integrity of equity ownership records, settlement receipts, or proof-of-reserve history.

This prompt specifies the engineering, automation, and operational runbooks for the **Blockchain Ledger Disaster Recovery & RocksDB State Snapshotting System** (`infra/blockchain/disaster-recovery/`). The architecture delivers automated, consistent database snapshots, immutable WORM-compliant cloud archive storage, cross-region replication, and an automated Kubernetes failover operator that guarantees a **Recovery Time Objective (RTO) $\le$ 15 minutes** and a **Recovery Point Objective (RPO) = 0 blocks** (zero loss of committed settlement transactions).

## What You Are Building
An automated, enterprise disaster recovery and snapshot automation subsystem comprising:
- RocksDB Hot Snapshot Daemon (Go / Python): Periodically captures consistent, point-in-time RocksDB storage checkpoints from live Besu nodes without halting consensus block production.
- Immutable Archive Pipeline: Compresses, encrypts (AES-256-GCM), and uploads snapshots to multi-region AWS S3 / GCP Cloud Storage with Object Lock (WORM compliance) and 7-year retention policies.
- Automated DR Restoration Operator: Kubernetes operator / CLI tool that provisions replacement validator and RPC nodes, hydrates persistent NVMe volumes from the latest verified snapshot, and fast-syncs remaining blocks to catch up to the consortium head.
- Automated Daily Disaster Recovery Drill Harness: Scheduled CI/CD workflow executing end-to-end node destruction, cold restore, and block integrity verification in an isolated staging environment.

## Scope Boundaries
- **In Scope:**
 - Automated Besu node RocksDB state checkpointing.
 - Multi-region encrypted backup storage with WORM retention.
 - Automated validator node restoration and fast peer catch-up.
 - Automated ledger integrity verification (block hash chain validation from genesis).
 - Disaster recovery playbooks and automated failover runbooks.
- **Out of Scope / Handled Elsewhere:**
 - Relational PostgreSQL database backups (handled in Prompt 406).
 - Overall business continuity and organization-wide DR plans (handled in Prompt 710).
 - Infrastructure as Code provisioning (handled in Prompt 805).

## Technology to Use
- **Snapshot Engine:** Hyperledger Besu RocksDB checkpoint utility / custom Go snapshot sidecar.
  *Justification:* RocksDB provides native hard-link checkpointing (`CreateCheckpoint`), allowing instantaneous, consistent point-in-time state snapshots without pausing read/write operations or degrading consensus block proposal latency.
- **Cloud Storage:** AWS S3 with Glacier Instant Retrieval, Cross-Region Replication (CRR), and S3 Object Lock (Compliance Mode).
- **Automation / Orchestration:** Go (v1.22+), Bash, Kubernetes Operators, Velero, Terraform.
- **Encryption:** AWS KMS / HashiCorp Vault with customer-managed keys (CMK) and AES-256-GCM encryption.

## Backend / Infra Touchpoints
- **Hyperledger Besu Validator StatefulSets:** Persistent volume mounts on AWS `gp3` / `io2` storage.
- **AWS S3 / GCP Storage Buckets:** `s3://growww-ledger-backups-mumbai/` and `s3://growww-ledger-backups-hyderabad-dr/`.
- **Vault KMS:** Handles automated backup encryption and decryption keys.
- **Prometheus / Alertmanager (Prompt 310):** Monitors backup execution status, snapshot size, and recovery drill results.

## Blockchain Interaction
- **Consensus Continuity:** In the event of 1 validator failure, QBFT consensus continues unaffected ($N=4, F=1$). If 2 validators fail, consensus halts safely; the DR operator restores 1 validator from snapshot, allowing QBFT to resume producing blocks with zero data loss ($RPO = 0$).
- **State Integrity:** Restored nodes verify block header hashes, receipt roots, and QBFT extra-data signatures sequentially from the snapshot height to the current consortium chain head.
- **Zero PII in Backups:** Ledger database files contain only bytecode, encrypted states, and pseudonymous transactions.

## Step-by-Step Build Instructions
1. Scaffold disaster recovery repository structure under `infra/blockchain/disaster-recovery/` with subdirectories: `snapshot-daemon/`, `restore-operator/`, `terraform/`, `runbooks/`.
2. Configure AWS S3 backup buckets in primary region (Mumbai `ap-south-1`) and DR region (Hyderabad `ap-south-2`) using Terraform:
 - Enable S3 Object Lock in Compliance Mode (retention period: 7 years).
 - Configure bidirectional Cross-Region Replication (CRR).
 - Enable default SSE-KMS encryption with customer-managed key.
3. Implement the Go snapshot daemon (`snapshot-daemon/main.go`):
 - Trigger hourly snapshot via Besu RocksDB checkpointing.
 - Package snapshot data directory (`/var/lib/besu/data/database`) using `tar.zstd` compression.
 - Compute SHA-256 checksum and metadata manifest (`manifest.json` containing block height, block hash, timestamp, chain ID).
 - Encrypt archive with AES-256-GCM key from Vault KMS.
 - Stream compressed archive directly to S3 backup bucket.
4. Implement snapshot cleanup policy: retain hourly snapshots for 7 days, daily snapshots for 90 days, monthly snapshots for 7 years.
5. Implement the Restore Operator CLI (`restore-operator/main.go`):
 - Query S3 bucket for the latest valid snapshot manifest.
 - Download and verify SHA-256 checksum against manifest.
 - Decrypt archive using KMS key.
 - Extract RocksDB database files into the target Persistent Volume Claim (PVC) mount point.
 - Update Besu configuration to point to active consortium bootnodes.
6. Configure Kubernetes Job templates (`templates/restore_validator_job.yaml`) executing the automated restore operator before launching the Besu validator container.
7. Implement node health and state verification script (`scripts/verify_node_state.sh`):
 - Connect to local Besu JSON-RPC endpoint.
 - Call `eth_blockNumber` and compare with consortium peers.
 - Verify block hash integrity: check `eth_getBlockByNumber` matches consortium median block hash.
 - Call `qbft_getValidatorsByBlockNumber` to verify validator set membership.
8. Set up automated daily DR test in staging:
 - A scheduled GitHub Actions / Argo Workflows job terminates a running validator node and deletes its persistent storage volume.
 - The restore operator provisions a new volume from S3, hydrates data, starts the node, and verifies block sync within 10 minutes.
 - Job logs metrics (`dr_restore_duration_seconds`, `dr_restored_block_height`) to Prometheus.
9. Configure Prometheus alert rules:
 - `LedgerSnapshotFailed`: Alert if no valid snapshot is uploaded for $>2$ hours.
 - `LedgerDRDrillFailed`: Alert if automated daily DR restoration test fails.
10. Write step-by-step operational Disaster Recovery Runbook (`docs/ops/dr_runbook_validator_failure.md`):
 - Procedure A: Single Validator Node Failure & Quick Volume Swap.
 - Procedure B: Total Cloud Region Outage & Cross-Region Failover to Hyderabad.
 - Procedure C: Ledger State Corruption & Coordinated Network Rollback.
11. Test and document key restoration protocols ensuring validator HSM signing keys are restored in the DR region (Prompt 311).
12. Conduct full manual tabletop DR drill with DevOps and Security team members; obtain compliance audit sign-off.

## Interfaces / Contracts

### Snapshot Manifest JSON Schema (`manifest.json`)
```json
{
  "manifest_version": "1.0",
  "snapshot_id": "SNAP-BESU-20260918-120000",
  "chain_id": 13370,
  "consensus_engine": "QBFT",
  "block_height": 4829100,
  "block_hash": "0x4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b",
  "timestamp": "2026-09-18T12:00:00Z",
  "storage_engine": "RocksDB",
  "uncompressed_size_bytes": 142857142857,
  "compressed_size_bytes": 38472918273,
  "archive_file": "snap_besu_4829100.tar.zst.enc",
  "sha256_checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "kms_key_id": "arn:aws:kms:ap-south-1:123456789012:key/growww-ledger-backup-key",
  "encryption_algorithm": "AES-256-GCM"
}
```

### Restore CLI Command Interface
```bash
# Automated Disaster Recovery Restore Command
growww-ledger-dr restore \
  --snapshot-id latest \
  --s3-bucket growww-ledger-backups-mumbai \
  --target-dir /var/lib/besu/data \
  --verify-checksum=true \
  --kms-key-arn arn:aws:kms:ap-south-1:123456789012:key/growww-ledger-backup-key \
  --log-level info
```

## Security & Compliance Notes
- **SEBI & RBI Regulatory Retention (7 Years WORM):** Snapshot archives are protected by S3 Object Lock in Compliance Mode. Even AWS account root users cannot delete or modify ledger backup archives during the 7-year retention window.
- **Client-Side Encryption:** All snapshots are encrypted with AES-256-GCM keys managed in HSM/KMS before leaving the Kubernetes cluster boundaries.
- **Zero Data Loss Guarantee ($RPO = 0$):** Because QBFT consensus provides instant finality and the consortium state is replicated across multiple live validator nodes, restoring a node from an hourly snapshot allows it to rapidly pull the few remaining blocks from peer validators, achieving exact $RPO = 0$ blocks.

## Acceptance Criteria
- [ ] Automated snapshot daemon successfully generates consistent RocksDB checkpoints hourly and uploads encrypted archives to S3.
- [ ] Restore operator successfully downloads, verifies, decrypts, and restores a 100GB+ snapshot in $\le 8\text{ minutes}$.
- [ ] Restored Besu node reconnects to peers and catches up to the chain head within 2 minutes ($RTO < 15\text{ minutes}$, $RPO = 0\text{ blocks}$).
- [ ] S3 Object Lock and Cross-Region Replication configured and verified in Terraform.
- [ ] Automated daily DR drill executes in CI/CD pipeline and passes with zero manual intervention.
- [ ] Comprehensive DR runbook approved by Chief Risk Officer and Head of Security.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `302` (Network Topology & Validator Setup), Prompt `310` (Chain Node Monitoring), Prompt `710` (Business Continuity & DR Plan).
- **Parallel Tasks:** Prompt `406` (Backup & Restore Automation for Datastores), Prompt `805` (Infrastructure as Code).
- **Subsequent Prompts Enabled:** Prompt `904` (Chaos Engineering & Resilience Testing), Prompt `908` (Production Launch Runbook).
