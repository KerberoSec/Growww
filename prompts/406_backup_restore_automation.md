# 406 - Backup & Restore Automation for All Datastores

## Purpose
In a regulated securities and fractional investment platform, data resilience is paramount. Hardware failure, data corruption, ransomware, or regional cloud outages must never result in financial ledger discrepancies, lost trade orders, or unrecoverable investor balances. The SEBI Cyber Security and Cyber Resilience Framework mandates rigorous data recovery capabilities, continuous data replication, and scheduled disaster recovery drills.

This prompt implements the automated backup, continuous Write-Ahead Log (WAL) archiving, Point-in-Time Recovery (PITR), and automated restore verification pipeline across all Growww data tiers (PostgreSQL, Redis, Kafka, ClickHouse, and Hyperledger Besu blockchain nodes). It guarantees a Recovery Point Objective (RPO) of < 1 minute and a Recovery Time Objective (RTO) of < 15 minutes, fulfilling the core non-functional requirements defined in Prompt 010.

## What You Are Building
An enterprise-grade, automated multi-datastore backup and recovery platform containing:
- **PostgreSQL pgBackRest Infrastructure:** Multi-threaded backup and continuous WAL archiving configuration (`infra/database/pgbackrest/`) streaming transaction logs directly to AWS S3.
- **Point-in-Time Recovery (PITR) Engine:** Automated CLI tools and Kubernetes jobs to restore PostgreSQL to any arbitrary second within the past 30 days.
- **Redis Snapshot Automation:** Automated scheduled RDB/AOF snapshot jobs shipping Redis state to isolated S3 backup storage.
- **Kafka Topic & Offset Snapshotting:** Backup tooling capturing Kafka consumer group offsets and critical compacted topic states (`market.securities_master`, `compliance.investor_status`).
- **ClickHouse Backup Automation:** `clickhouse-backup` tool integration for freeze, compression, and S3 upload of columnar analytical partitions.
- **Blockchain Node Snapshotting:** Coordinated data directory snapshotting for Hyperledger Besu validator and RPC nodes.
- **Automated Restore Drill Test Suite:** A weekly automated Kubernetes CronJob (`infra/drills/automated_restore_drill.sh`) that spins up isolated ephemeral databases, restores the latest backup, runs financial consistency checks, and alerts on failure.
- **Prometheus Backup Monitoring:** Metrics exporters and alerting rules for backup failures, WAL archive lag, and failed restore drills.

## Scope Boundaries
- **In Scope:** pgBackRest configuration, continuous WAL streaming, PITR restoration scripts, Redis snapshotting, ClickHouse backups, Besu data directory snapshotting, automated restore drills, and backup health telemetry.
- **Out of Scope / Handled Elsewhere:**
 - High-level business Disaster Recovery Plan document (handled in Prompt 710).
 - Hyperledger Besu consensus recovery and validator re-keying (handled in Prompt 314).
 - PostgreSQL live schema migrations (handled in Prompt 401).
 - AWS Infrastructure provisioning (handled in Prompt 802).

## Technology to Use
pgBackRest 2.50+ is selected as the primary backup solution for PostgreSQL. pgBackRest provides multi-threaded backup/restore performance, delta restores, asynchronous continuous WAL archiving, page checksum validation, and native AWS S3 integration. For Kubernetes persistent volumes, Velero 1.13+ is used. For ClickHouse, `clickhouse-backup` is used for freeze/unfreeze partition shipping.

- **PostgreSQL Backup:** pgBackRest 2.50+ with Zstandard (ZSTD) compression.
- **ClickHouse Backup:** `clickhouse-backup` v2.5+.
- **Kubernetes PVC Backup:** Velero 1.13+ with CSI snapshot plugin.
- **Storage Target:** AWS S3 (Primary: `ap-south-1` Mumbai, Replicated: `ap-south-2` Hyderabad via S3 Cross-Region Replication).
- **Automation / Orchestration:** Kubernetes CronJobs, Bash 5.2, Python 3.12 (Boto3).

## Backend / Infra Touchpoints
- **PostgreSQL Cluster:** Primary and standby nodes streaming WAL segments to pgBackRest repository.
- **Redis Cluster:** Multi-AZ nodes generating scheduled RDB dumps.
- **ClickHouse Cluster:** 3-node cluster executing scheduled table partition backups.
- **Hyperledger Besu Nodes:** Validator and archive nodes creating filesystem snapshots.
- **AWS S3 Backup Buckets:** Dedicated backup buckets with KMS encryption and Cross-Region Replication (CRR).
- **Prometheus & Alertmanager:** Scraping `pgbackrest-exporter` and restore drill execution status.

## Blockchain Interaction
Coordinated multi-tier backups ensure that off-chain PostgreSQL state remains in exact mathematical harmony with the permissioned Hyperledger Besu blockchain network.

### Detailed On-Chain Integration Mechanics:
- **Synchronized Snapshot Timestamps:** Backup jobs record the exact Hyperledger Besu block height and block timestamp corresponding to every PostgreSQL full backup and ClickHouse freeze.
- **Besu Data Directory Snapshots:** Nightly automated volume snapshots of Besu QBFT state directories (`/var/lib/besu/data`) allow rapid node spin-up without re-syncing from genesis block #0.
- **Post-Restore Ledger Reconciliation:** Following a database PITR restore, the system executes Prompt 215 (Reconciliation Service) to compare restored PostgreSQL token balances against live `DigitalSecurityToken.sol` on-chain balances and `ProofOfReserveRegistry.sol` Merkle roots, flagging any post-recovery discrepancies for resolution.

## Step-by-Step Build Instructions
1. Deploy dedicated S3 backup buckets in AWS Mumbai (`ap-south-1`) and configure Cross-Region Replication to AWS Hyderabad (`ap-south-2`).
2. Deploy pgBackRest dedicated repository host/container within the database VPC subnet.
3. Configure `pgbackrest.conf` on PostgreSQL primary and replica instances for continuous `archive-push` and daily full / incremental backups.
4. Configure S3 storage settings with SSE-KMS encryption and multi-threaded compression (`process-max=8`, `compress-type=zst`).
5. Deploy `clickhouse-backup` DaemonSet/CronJob to execute daily full snapshots of ClickHouse analytical tables.
6. Configure scheduled Redis background saves (`BGSAVE`) with automated S3 shipping via sidecar container.
7. Configure Velero with AWS CSI volume snapshotting for Kubernetes stateful workloads.
8. Author automated Point-in-Time Recovery script (`scripts/restore_postgres_pitr.sh`) supporting target timestamp arguments (`--target-time="2026-09-18 10:30:00+05:30"`).
9. Implement automated weekly restore drill Kubernetes CronJob: spin up ephemeral Postgres and ClickHouse pods, restore yesterday's backup, verify schema integrity, validate double-entry balance zero-sum invariants, and terminate test pods.
10. Deploy `pgbackrest-exporter` to export backup metrics (last backup age, backup size, WAL archive error count) to Prometheus.
11. Configure Alertmanager rules alerting if PostgreSQL full backup > 24 hours old or WAL archive lag > 5 minutes.
12. Execute full simulated DR drill: simulate database host destruction, execute PITR recovery to specific target time, and verify complete restoration within 15 minutes.

## Interfaces / Contracts

```ini
# pgBackRest Production Configuration: /etc/pgbackrest/pgbackrest.conf

[global]
repo1-type=s3
repo1-s3-bucket=growww-db-backups-prod-ap-south-1
repo1-s3-endpoint=s3.ap-south-1.amazonaws.com
repo1-s3-region=ap-south-1
repo1-s3-kms-key-id=arn:aws:kms:ap-south-1:123456789012:key/backup-key-uuid
repo1-path=/pgbackrest
repo1-retention-full=30
repo1-retention-diff=14
repo1-cipher-type=aes-256-cbc
process-max=8
compress-type=zst
compress-level=6
log-level-console=info
log-level-file=detail
start-fast=y
checksum=y

[growww-pg-cluster]
pg1-path=/var/lib/postgresql/16/main
pg1-user=postgres
pg1-port=5432
pg1-socket-path=/var/run/postgresql
```

```yaml
# Prometheus Alerting Rules for Backup Infrastructure
groups:
 - name: growww_backup_alerts
    rules:
 - alert: PostgresBackupMissing
        expr: pgbackrest_backup_full_age_seconds{stanza="growww-pg-cluster"} > 90000 # 25 Hours
        for: 15m
        labels:
          severity: critical
          team: data-infra
        annotations:
          summary: "PostgreSQL full backup is older than 25 hours"
          description: "pgBackRest has not completed a successful full backup in >25h. Check repository health."

 - alert: PostgresWALArchivalLag
        expr: pgbackrest_wal_archive_lag_seconds{stanza="growww-pg-cluster"} > 300 # 5 Minutes
        for: 5m
        labels:
          severity: critical
          team: data-infra
        annotations:
          summary: "PostgreSQL WAL archiving is lagging behind"
          description: "WAL segments are not being pushed to S3 within 5 minutes. RPO target (<1m) is in jeopardy."

 - alert: RestoreDrillFailed
        expr: growww_restore_drill_status == 0
        for: 1m
        labels:
          severity: critical
          team: compliance-ops
        annotations:
          summary: "Automated weekly restore drill failed"
          description: "Automated restore verification failed schema integrity or double-entry balance check."
```

## Security & Compliance Notes
- **SEBI Cyber Security Framework:** Mandates offline or air-gapped backup copies, continuous replication to an alternate seismic zone (Mumbai to Hyderabad), and periodic simulated recovery drills.
- **KMS Envelope Encryption:** All backup snapshots and WAL archives are encrypted at rest using AWS KMS Customer Managed Keys (CMK) with strict separation of duties (database admins cannot modify KMS key policies).
- **Access Control & Immutability:** S3 backup buckets utilize S3 Object Lock and IAM bucket policies preventing unauthorized deletion of backup objects.
- **Zero Plaintext PII in Backups:** Sensitive PII is already column-encrypted in PostgreSQL; backup files contain no unencrypted investor identity records.

## Acceptance Criteria
- [ ] pgBackRest continuous WAL archiving deployed with WAL segment shipping latency < 60 seconds (satisfying RPO < 1 min).
- [ ] Automated daily full backups and hourly differential backups configured and shipping to S3 `ap-south-1` with CRR to `ap-south-2`.
- [ ] Point-in-Time Recovery (PITR) script successfully restores PostgreSQL to a specified second within 15 minutes (satisfying RTO < 15 min).
- [ ] `clickhouse-backup` executes automated daily analytical partition snapshots to S3 without interrupting live queries.
- [ ] Weekly automated restore drill Kubernetes CronJob provisions ephemeral instance, validates ledger integrity, and reports success to Prometheus.
- [ ] Simulated failover and restore drill demonstrates zero data loss and 100% double-entry ledger balance consistency.
- [ ] Prometheus metrics and Alertmanager rules operational for backup age, WAL lag, and drill failures.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 010 (Non-Functional Requirements), Prompt 401 (PostgreSQL Schema), Prompt 402 (Redis), Prompt 403 (Kafka), Prompt 404 (ClickHouse).
- **Parallel Tasks:** Prompt 314 (Blockchain Disaster Recovery), Prompt 710 (Disaster Recovery Architecture).
