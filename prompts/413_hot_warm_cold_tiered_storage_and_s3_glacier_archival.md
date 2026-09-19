# 413 - Hot-Warm-Cold Tiered Storage & S3 Glacier Immutable Regulatory Archival

## Purpose
Indian financial market participants operating under the regulatory supervision of the Securities and Exchange Board of India (SEBI), Reserve Bank of India (RBI), and Prevention of Money Laundering Act (PMLA) are statutorily required to maintain an unalterable, non-repudiable audit trail of every market tick, order lifecycle transition, trade execution, ledger journal entry, and administrative action for a minimum of 8 years (2,920 days). During peak market volatility, high-frequency equity, currency, and derivatives feeds generate hundreds of gigabytes of market ticks and operational logs daily. Maintaining petabyte-scale historical datasets on high-performance primary NVMe storage leads to runaway infrastructure costs, storage volume exhaustion, and severe operational query degradation across live relational databases.

This specification implements the automated Hot-Warm-Cold Tiered Storage & S3 Glacier Immutable Regulatory Archival system (`services/storage-archival-worker`). It automates the seamless lifecycle transition of high-throughput market tick data, double-entry ledger journals, PostgreSQL write-ahead logs (WAL), and audit event streams across three distinct storage tiers:
1. **Hot Tier (0 to 30/90 days):** Local NVMe SSDs (AWS EBS `io2`/`gp3` or instance NVMe) powering sub-millisecond operational matching, transactional ledger balancing, and interactive intraday chart rendering.
2. **Warm Tier (90 days to 365 days):** Columnar ClickHouse clusters backed by local SSD caching and AWS S3 Standard / S3 Standard-Infrequent Access (S3 Standard-IA) for cost-effective analytical querying, quantitative backtesting, and compliance investigations.
3. **Cold Tier (1 year to 8+ years):** Cryptographically sealed, Zstandard-compressed Parquet and raw payload archives stored in AWS S3 Glacier Vault Lock and Glacier Deep Archive under strict Write Once, Read Many (WORM) compliance locks, preventing record alteration or deletion by any IAM identity or root account.

The system guarantees mathematical data integrity across all lifecycle phases through SHA-256 tree hashing, client-side AWS KMS envelope encryption, and periodic on-chain cryptographic anchoring on the permissioned Hyperledger Besu blockchain network.

## What You Are Building
A production-grade, distributed data tiering daemon and regulatory archival engine containing:
- **Storage Archival Worker Daemon (`services/storage-archival-worker`):** A high-performance Go 1.22+ worker service that coordinates scheduled partition offloading, data stream compression, cryptographic hashing, client-side envelope encryption, and S3/Glacier multipart transmission.
- **ClickHouse Multi-Tier Storage Configuration:** Declarative XML storage policies (`infra/clickhouse/storage_configuration.xml`) defining multi-volume hierarchies across NVMe disks and S3 object stores, accompanied by automated background TTL policies shifting expired data parts.
- **PostgreSQL Historical Partition & WAL Archiver:** Automated partition detachment daemon extracting closed historical monthly ledger tables (`ledger.journal_entries`, `trading.trades`) and PostgreSQL WAL base backups into zstd-compressed columnar Parquet files.
- **Audit Trail & Blockchain State Snapshot Archiver:** Archival pipeline ingesting finalized audit blocks from the Audit Log Service (Prompt 218) and historical Hyperledger Besu epoch state trie snapshots, Merkle receipts, and block headers.
- **Glacier Vault Lock WORM Policy Engine:** Terraform manifests (`infra/terraform/storage/glacier_vault_lock.tf`) establishing an AWS S3 Glacier Vault configured with a locked, non-deletable Vault Lock policy enforcing SEBI 8-year mandatory retention.
- **Cryptographic Verification & Integrity CLI:** An operator CLI tool allowing compliance auditors to verify cold archive SHA-256 tree checksums directly against on-chain Hyperledger Besu state attestations without de-archiving multi-terabyte data sets.
- **Delivered Artifacts:**
  - `services/storage-archival-worker/cmd/worker/main.go` - High-throughput tiering daemon entry point.
  - `services/storage-archival-worker/internal/tiering/clickhouse_tierer.go` - ClickHouse partition movement manager.
  - `services/storage-archival-worker/internal/tiering/postgres_wal_archiver.go` - PostgreSQL partition and WAL archiver.
  - `services/storage-archival-worker/internal/worm/glacier_vault_lock.go` - AWS S3 Glacier Vault Lock integration.
  - `services/storage-archival-worker/internal/crypto/zstd_sha256.go` - Zstandard level 9-19 stream compressor and SHA-256 tree hasher.
  - `services/storage-archival-worker/internal/blockchain/besu_snapshot_archiver.go` - Besu epoch snapshot and receipt archiver.
  - `infra/terraform/storage/glacier_vault_lock.tf` - Terraform infrastructure manifests for WORM Glacier Vault.
  - `infra/clickhouse/storage_configuration.xml` - ClickHouse multi-tier disk and volume configuration.
  - `config/lifecycle_policies.json` - Tiering schedule and dataset retention definition.

## Scope Boundaries
- **In Scope:**
  - Automated scanning and detection of aged partitions in ClickHouse and PostgreSQL.
  - Multi-tier lifecycle transition orchestration: NVMe (hot) -> S3 Standard/IA (warm) -> S3 Glacier Vault Lock (cold).
  - Columnar conversion of raw transactional database partitions to Apache Parquet with Zstandard compression.
  - Streaming SHA-256 checksum and Merkle tree hash generation for every generated archive bundle.
  - Client-side envelope encryption using AWS KMS Customer Managed Keys (CMK) prior to network transmission.
  - Automated submission of archive objects to AWS S3 Glacier Vault Lock with WORM compliance locks.
  - Archiving past Hyperledger Besu blockchain epoch state snapshots, block receipts, and Merkle proofs.
  - On-chain notarization of archive manifests to Hyperledger Besu smart contracts.
  - Offline mathematical integrity verification tooling for statutory compliance audits.
- **Out of Scope / Handled Elsewhere:**
  - Primary matching engine transaction logging and mmap ring buffer WALs (handled in Prompt 246).
  - Live analytical query optimization and hypertable partitioning (handled in Prompt 408).
  - Real-time audit event ingestion and streaming Kafka topics (handled in Prompt 218).
  - Disaster recovery live snapshot replication and multi-region failover (handled in Prompt 406).
  - Data privacy consent tracking and DPDP customer deletion workflows (handled in Prompt 008 and Prompt 405).

## Technology to Use
The tiering and archival architecture balances high-throughput stream processing, maximum storage compression, and legally indisputable cryptographic immutability.

- **Primary Runtime Language:** Go 1.22+. Go provides predictable low memory usage, zero-overhead concurrency via goroutines for high-throughput S3 multipart uploads, and native support for the official AWS SDK for Go v2.
- **Secondary Tooling & Utilities:** Python 3.12 with PyArrow and Boto3 for auxiliary data lake format validation and batch Parquet schemas.
- **Compression Algorithm:** Zstandard (zstd) using compression levels 9 to 19. Zstandard yields a 75% to 85% compression ratio over JSON/CSV logs and uncompressed time-series ticks, while maintaining decompression throughput exceeding 1.2 GB/s per core.
- **Cold Storage Format:** Apache Parquet with Snappy or ZSTD compression for structured tabular data; tar.zst archives for RocksDB blockchain state directories and PostgreSQL WAL segments.
- **Storage Infrastructure:**
  - **Hot Tier:** Local NVMe SSDs / AWS EBS `io2` / `gp3` provisioned on primary database clusters.
  - **Warm Tier:** ClickHouse S3 Disks backed by AWS S3 Standard and S3 Standard-IA in AWS Mumbai (`ap-south-1`).
  - **Cold Tier:** AWS S3 Glacier Vault Lock and Glacier Deep Archive configured with immutable WORM policies.
- **Cryptographic Encryption & Key Management:** Client-side envelope encryption with AES-GCM-256 using AWS KMS Customer Managed Keys (CMK) with automated annual rotation.
- **On-Chain Attestation:** Permissioned Hyperledger Besu consortium network (QBFT consensus, 1:1 asset backing, zero PII).

## Backend / Infra Touchpoints
- **PostgreSQL 16 Enterprise Cluster:** Primary relational store for orders, trades, and double-entry ledger journals. Source for table partition detachment and WAL archiving via `pg_receivewal` and `archive_command`.
- **ClickHouse 24.x Cluster (Prompt 408):** Historical market tick datastore (`analytics.market_ticks_historical`). Ingests ticks from Kafka and shifts historical parts across storage disks via native storage policies.
- **Audit Log Service (Prompt 218):** Source of immutable audit trail records (`audit_events` and `audit_epochs`), providing RFC 6962 Merkle tree roots for cold-storage consolidation.
- **Hyperledger Besu Archive Nodes:** Consortium blockchain nodes providing full state trie dumps, block execution receipts, and event logs via RPC (`debug_dumpBlock`, `eth_getBlockReceipts`).
- **AWS KMS (ap-south-1):** Dedicated hardware security module (HSM) backed Customer Managed Key with strict IAM key policies preventing key deletion.
- **AWS S3 & Glacier Vault:** Secure buckets and vaults in Mumbai (`ap-south-1`) with S3 Object Lock and Glacier Vault Lock compliance rules.
- **AWS Glue Data Catalog & Athena:** Catalogs cold-archived Parquet datasets for serverless ad-hoc audit queries.

## Blockchain Interaction
The archival daemon interfaces directly with the permissioned Hyperledger Besu blockchain network (QBFT consensus, 1:1 custody backing, zero PII) to provide permanent mathematical verification for all archived assets:

### Detailed On-Chain Integration Mechanics:
1. **Blockchain Epoch State Snapshot Archival:**
   - Every 100,000 blocks (or 30 calendar days), Besu archive nodes generate a consistent RocksDB state trie snapshot and export complete block receipts.
   - The archival daemon packages these files into an encrypted `besu_epoch_{start}_{end}.tar.zst` archive and transmits it to the S3 Glacier Vault.
2. **Dual-Tier Merkle Root Notarization:**
   - Upon creating any cold archive bundle (market ticks, ledger journals, audit logs, or Besu snapshots), the worker computes:
     - The archive payload SHA-256 checksum.
     - The RFC 6962 Merkle tree root representing all constituent records.
   - The daemon signs an on-chain transaction calling `StorageArchivalRegistry.sol` using an HSM-backed relayer account.
   - The smart contract records:
     - `archiveId` (UUIDv4)
     - `datasetType` (e.g., `MARKET_TICKS`, `LEDGER_JOURNAL`, `AUDIT_LOG`, `BESU_SNAPSHOT`)
     - `epochPeriod` (ISO 8601 interval)
     - `recordCount` (total rows or items encapsulated)
     - `merkleRootHash` (bytes32)
     - `sha256PayloadHash` (bytes32)
     - `glacierArchiveId` (string identifier assigned by AWS Glacier)
     - `vaultArn` (string ARN of the WORM vault)
3. **Zero PII Exposure Guarantee:**
   - Only cryptographic hashes, archive identifiers, record counts, and storage locations are recorded on the public or consortium ledger. No investor names, PANs, account balances, or plaintext orders touch the smart contract.
4. **Independent Audit Verification:**
   - External statutory auditors (SEBI/RBI) can retrieve any cold archive from Glacier, recalculate its SHA-256 hash and Merkle root, and query `StorageArchivalRegistry.sol` to verify that the mathematical proof matches the record notarized 8 years prior.

## Step-by-Step Build Instructions
1. **Initialize Worker Service Structure:**
   - Scaffold Go 1.22 project under `services/storage-archival-worker` with subpackages for `tiering`, `worm`, `crypto`, and `blockchain`.
2. **Provision Immutable S3 Glacier Vault Infrastructure:**
   - Author Terraform manifests in `infra/terraform/storage/glacier_vault_lock.tf` provisioning an AWS S3 Glacier Vault (`growww-regulatory-archive-prod`).
   - Define the Vault Lock policy requiring strict WORM compliance and an 8-year (2,920-day) lock period.
   - Author the Terraform resource to initiate the Vault Lock and complete the lock within the 24-hour AWS confirmation window.
3. **Configure ClickHouse Multi-Tier Storage Policies:**
   - Author `infra/clickhouse/storage_configuration.xml` establishing a disk hierarchy:
     - `default` / `hot_disk`: Local NVMe SSD mount path `/var/lib/clickhouse/`.
     - `s3_warm_disk`: S3-backed storage bucket `growww-clickhouse-warm-ap-south-1`.
   - Configure storage policy `hot_to_warm_policy` with automatic background part movement when partition age exceeds 90 days or disk utilization exceeds 80%.
4. **Implement ClickHouse Partition Tiering Daemon:**
   - In `internal/tiering/clickhouse_tierer.go`, implement a scheduled scanner querying `system.parts` for partitions older than 90 days.
   - Execute `ALTER TABLE analytics.market_ticks_historical MOVE PARTITION <partition_id> TO DISK 's3_warm_disk'` to gracefully transition data without query downtime.
5. **Implement PostgreSQL Partition Detachment Pipeline:**
   - In `internal/tiering/postgres_wal_archiver.go`, implement a database worker that scans for closed monthly partitions in `ledger.journal_entries` and `trading.trades` older than 90 days.
   - Execute `ALTER TABLE ... DETACH PARTITION` concurrently to decouple the historical partition from the live operational query planner.
6. **Implement Streaming Parquet Exporter:**
   - Stream rows from detached PostgreSQL partitions into columnar Apache Parquet files chunked at 512 MB boundaries.
   - Apply Zstandard level 12 compression and enforce strict decimal precision types for financial balances.
7. **Implement PostgreSQL WAL Segment Archiver:**
   - Configure PostgreSQL `archive_command` to push completed 16 MB WAL segments to a local staging buffer.
   - The worker packages batches of 1,024 WAL segments into compressed `pg_wal_batch_{timestamp}.tar.zst` bundles with accompanying LSN range manifests.
8. **Build Audit Log Consolidation Pipeline:**
   - Integrate with Audit Log Service (Prompt 218) to read sealed hourly audit epochs.
   - Aggregate daily audit blocks, verify their internal RFC 6962 hash chains, and export consolidated archives for cold storage.
9. **Build Hyperledger Besu Epoch Snapshot Extractor:**
   - In `internal/blockchain/besu_snapshot_archiver.go`, connect to Besu archive node RPC APIs.
   - Extract block headers, transaction receipts, and state trie dumps for finalized 100,000-block intervals.
   - Bundle exported receipts and state snapshots into tar.zst packages.
10. **Implement Zstandard Stream Compression & SHA-256 Hasher:**
    - In `internal/crypto/zstd_sha256.go`, build an `io.WriteCloser` wrapper that compresses byte streams via `github.com/klauspost/compress/zstd` while simultaneously calculating the SHA-256 checksum and Merkle leaf tree hashes.
11. **Implement AWS KMS Client-Side Envelope Encryption:**
    - Generate an ephemeral 256-bit AES-GCM data key via `kms:GenerateDataKey` using the regulatory Customer Managed Key (CMK).
    - Encrypt the compressed payload stream locally using AES-GCM before emitting network bytes.
    - Prepend the encrypted envelope data key and initialization vector (IV) to the archive file header.
12. **Implement Glacier Multipart Vault Uploader:**
    - In `internal/worm/glacier_vault_lock.go`, implement a resilient multipart uploader using the AWS SDK v2 Glacier client.
    - Stream encrypted payloads in 128 MB chunks, computing Glacier tree-hash checksums per chunk and assembling the final archive.
    - Capture the returned AWS Glacier Archive ID and record creation metadata.
13. **Deploy On-Chain Notarization Relayer:**
    - Develop smart contract client calling `StorageArchivalRegistry.sol` on Hyperledger Besu.
    - Submit the cryptographic manifest (archive ID, dataset type, record count, SHA-256 hash, Merkle root, Glacier archive ID) via an HSM-signed transaction.
14. **Implement Drop & Purge Safety Mechanism:**
    - Validate that:
      - The Glacier upload returned HTTP 201 Created with verified checksums.
      - The Besu notarization transaction has achieved minimum 12-block finality.
      - The archive manifest has been verified by a test read.
    - Once validated, drop the detached PostgreSQL partition and remove local staging files.
15. **Build Administrative Verification CLI & Health Dashboards:**
    - Create CLI command `storage-archival-worker verify --archive-id <uuid>` to compare local/Glacier file hashes against the on-chain Besu registry.
    - Export Prometheus metrics (`storage_tiering_bytes_migrated_total`, `storage_archival_duration_seconds`, `glacier_upload_errors_total`) and configure Alertmanager alarms.

## Interfaces / Contracts

### 1. Lifecycle Policies Configuration (`config/lifecycle_policies.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "version": "1.0.0",
  "retention_jurisdiction": "SEBI_INDIA",
  "compliance_retention_days": 2920,
  "policies": [
    {
      "dataset_name": "market.ticks",
      "source_system": "ClickHouse",
      "source_table": "analytics.market_ticks_historical",
      "hot_tier": {
        "storage_medium": "LOCAL_NVME",
        "retention_days": 90,
        "action": "MOVE_TO_WARM_DISK"
      },
      "warm_tier": {
        "storage_medium": "S3_STANDARD_IA",
        "retention_days": 275,
        "action": "EXPORT_PARQUET_TO_COLD"
      },
      "cold_tier": {
        "storage_medium": "AWS_GLACIER_VAULT_LOCK",
        "worm_mode": "COMPLIANCE",
        "retention_days": 2555,
        "compression": "ZSTD_LEVEL_19",
        "encryption": "KMS_CMK_AES_GCM_256"
      }
    },
    {
      "dataset_name": "ledger.journals",
      "source_system": "PostgreSQL",
      "source_table": "ledger.journal_entries",
      "hot_tier": {
        "storage_medium": "EBS_GP3",
        "retention_days": 90,
        "action": "DETACH_AND_EXPORT_PARQUET"
      },
      "warm_tier": {
        "storage_medium": "S3_STANDARD",
        "retention_days": 275,
        "action": "TRANSITION_TO_COLD"
      },
      "cold_tier": {
        "storage_medium": "AWS_GLACIER_VAULT_LOCK",
        "worm_mode": "COMPLIANCE",
        "retention_days": 2555,
        "compression": "ZSTD_LEVEL_15",
        "encryption": "KMS_CMK_AES_GCM_256"
      }
    },
    {
      "dataset_name": "postgres.wal",
      "source_system": "PostgreSQL_WAL",
      "source_path": "/var/lib/postgresql/wal_archive",
      "hot_tier": {
        "storage_medium": "LOCAL_NVME",
        "retention_days": 7,
        "action": "BUNDLE_AND_COMPRESS"
      },
      "warm_tier": {
        "storage_medium": "S3_STANDARD",
        "retention_days": 83,
        "action": "MOVE_TO_COLD"
      },
      "cold_tier": {
        "storage_medium": "AWS_GLACIER_VAULT_LOCK",
        "worm_mode": "COMPLIANCE",
        "retention_days": 2830,
        "compression": "TAR_ZSTD_LEVEL_12",
        "encryption": "KMS_CMK_AES_GCM_256"
      }
    },
    {
      "dataset_name": "blockchain.besu_snapshots",
      "source_system": "Hyperledger_Besu",
      "source_path": "/var/lib/besu/database",
      "hot_tier": {
        "storage_medium": "LOCAL_NVME",
        "retention_days": 30,
        "action": "EXPORT_EPOCH_STATE"
      },
      "warm_tier": {
        "storage_medium": "S3_STANDARD",
        "retention_days": 60,
        "action": "MOVE_TO_COLD"
      },
      "cold_tier": {
        "storage_medium": "AWS_GLACIER_VAULT_LOCK",
        "worm_mode": "COMPLIANCE",
        "retention_days": 2830,
        "compression": "TAR_ZSTD_LEVEL_16",
        "encryption": "KMS_CMK_AES_GCM_256"
      }
    }
  ]
}
```

### 2. ClickHouse Storage Configuration (`infra/clickhouse/storage_configuration.xml`)
```xml
<clickhouse>
    <storage_configuration>
        <disks>
            <!-- Hot Tier: High-speed local NVMe SSD -->
            <hot_nvme_disk>
                <type>local</type>
                <path>/var/lib/clickhouse/data/hot/</path>
            </hot_nvme_disk>
            
            <!-- Warm Tier: S3 Object Storage Bucket in Mumbai -->
            <s3_warm_disk>
                <type>s3</type>
                <endpoint>https://growww-clickhouse-warm-prod.s3.ap-south-1.amazonaws.com/data/</endpoint>
                <access_key_id from_env="AWS_ACCESS_KEY_ID"/>
                <secret_access_key from_env="AWS_SECRET_ACCESS_KEY"/>
                <metadata_path>/var/lib/clickhouse/disks/s3_warm_metadata/</metadata_path>
                <cache_enabled>true</cache_enabled>
                <cache_path>/var/lib/clickhouse/disks/s3_warm_cache/</cache_path>
                <cache_size>107374182400</cache_size> <!-- 100 GB Local Read Cache -->
            </s3_warm_disk>
        </disks>
        
        <policies>
            <hot_to_warm_policy>
                <volumes>
                    <hot_volume>
                        <default>true</default>
                        <disk>hot_nvme_disk</disk>
                        <max_data_part_size_bytes>10737418240</max_data_part_size_bytes> <!-- 10 GB -->
                    </hot_volume>
                    <warm_volume>
                        <disk>s3_warm_disk</disk>
                    </warm_volume>
                </volumes>
                <move_factor>0.2</move_factor>
            </hot_to_warm_policy>
        </policies>
    </storage_configuration>
</clickhouse>
```

### 3. PostgreSQL Partition Detachment Script (`scripts/detach_partitions.sql`)
```sql
-- Automated PostgreSQL Historical Partition Detachment Function
-- Executes safely under low locking priority to prevent blocking operational trades

CREATE OR REPLACE FUNCTION maintenance.detach_aged_partitions(
    p_schema_name TEXT,
    p_parent_table TEXT,
    p_cutoff_date DATE
) RETURNS TABLE(detached_partition_name TEXT, status TEXT) AS $$
DECLARE
    rec RECORD;
    v_partition_name TEXT;
    v_query TEXT;
BEGIN
    FOR rec IN 
        SELECT 
            c.relname AS partition_name,
            pg_get_expr(c.relpartbound, c.oid) AS partition_expression
        FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        JOIN pg_class p ON p.oid = i.inhparent
        JOIN pg_namespace n ON n.oid = p.relnamespace
        WHERE n.nspname = p_schema_name 
          AND p.relname = p_parent_table
          AND c.relname ~ ('^' || p_parent_table || '_[0-9]{4}_[0-9]{2}$')
        ORDER BY c.relname ASC
    LOOP
        -- Check if partition boundary precedes the cutoff date
        IF rec.partition_name < (p_parent_table || '_' || to_char(p_cutoff_date, 'YYYY_MM')) THEN
            v_partition_name := quote_ident(p_schema_name) || '.' || quote_ident(rec.partition_name);
            
            -- Attempt non-blocking concurrent detach
            BEGIN
                EXECUTE format('ALTER TABLE %I.%I DETACH PARTITION %s CONCURRENTLY;', 
                               p_schema_name, p_parent_table, v_partition_name);
                
                detached_partition_name := rec.partition_name;
                status := 'SUCCESSFULLY_DETACHED';
                RETURN NEXT;
            EXCEPTION WHEN OTHERS THEN
                detached_partition_name := rec.partition_name;
                status := 'DETACH_FAILED: ' || SQLERRM;
                RETURN NEXT;
            END;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### 4. AWS S3 Glacier Vault Lock Policy Specification (`vault_lock_policy.json`)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EnforceSEBIEightYearWORMRetention",
      "Effect": "Deny",
      "Principal": "*",
      "Action": [
        "glacier:DeleteArchive",
        "glacier:DeleteVault",
        "glacier:AbortVaultLock"
      ],
      "Resource": "arn:aws:glacier:ap-south-1:123456789012:vaults/growww-regulatory-archive-prod",
      "Condition": {
        "NumericLessThan": {
          "glacier:ArchiveAgeInDays": "2920"
        }
      }
    }
  ]
}
```

### 5. Archival Manifest Schema (`archive_manifest.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "RegulatoryArchivalManifest",
  "type": "object",
  "properties": {
    "manifestVersion": { "type": "string", "enum": ["1.0.0"] },
    "archiveId": { "type": "string", "format": "uuid" },
    "datasetName": { 
      "type": "string", 
      "enum": ["market.ticks", "ledger.journals", "audit.events", "postgres.wal", "blockchain.besu_snapshots"] 
    },
    "epochInterval": {
      "type": "object",
      "properties": {
        "startTime": { "type": "string", "format": "date-time" },
        "endTime": { "type": "string", "format": "date-time" }
      },
      "required": ["startTime", "endTime"]
    },
    "recordCount": { "type": "integer", "minimum": 0 },
    "uncompressedBytes": { "type": "integer", "minimum": 0 },
    "compressedBytes": { "type": "integer", "minimum": 0 },
    "compressionCodec": { "type": "string", "enum": ["ZSTD_12", "ZSTD_15", "ZSTD_19", "TAR_ZSTD"] },
    "encryptionMetadata": {
      "type": "object",
      "properties": {
        "algorithm": { "type": "string", "enum": ["AES-GCM-256"] },
        "kmsKeyArn": { "type": "string" },
        "encryptedDataKeyBase64": { "type": "string" },
        "ivBase64": { "type": "string" }
      },
      "required": ["algorithm", "kmsKeyArn", "encryptedDataKeyBase64", "ivBase64"]
    },
    "cryptographicProof": {
      "type": "object",
      "properties": {
        "sha256PayloadHash": { "type": "string", "pattern": "^[a-f0-9]{64}$" },
        "merkleTreeRoot": { "type": "string", "pattern": "^0x[a-f0-9]{64}$" },
        "merkleTreeDepth": { "type": "integer", "minimum": 1 },
        "glacierTreeHash": { "type": "string", "pattern": "^[a-f0-9]{64}$" }
      },
      "required": ["sha256PayloadHash", "merkleTreeRoot", "glacierTreeHash"]
    },
    "storageLocation": {
      "type": "object",
      "properties": {
        "vaultArn": { "type": "string" },
        "glacierArchiveId": { "type": "string" },
        "wormMode": { "type": "string", "enum": ["COMPLIANCE"] },
        "retentionLockedUntil": { "type": "string", "format": "date-time" }
      },
      "required": ["vaultArn", "glacierArchiveId", "wormMode", "retentionLockedUntil"]
    },
    "blockchainAttestation": {
      "type": "object",
      "properties": {
        "network": { "type": "string", "enum": ["BESU_MAINNET_CONSORTIUM"] },
        "contractAddress": { "type": "string", "pattern": "^0x[a-fA-F0-9]{40}$" },
        "transactionHash": { "type": "string", "pattern": "^0x[a-fA-F0-9]{64}$" },
        "blockNumber": { "type": "integer" }
      },
      "required": ["network", "contractAddress", "transactionHash", "blockNumber"]
    }
  },
  "required": [
    "manifestVersion",
    "archiveId",
    "datasetName",
    "epochInterval",
    "recordCount",
    "cryptographicProof",
    "storageLocation",
    "blockchainAttestation"
  ]
}
```

### 6. Hyperledger Besu Notarization Interface (`IStorageArchivalRegistry.sol`)
```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

/**
 * @title IStorageArchivalRegistry
 * @notice Notarizes cold-storage regulatory archives to guarantee tamper evidence under SEBI/RBI audits.
 */
interface IStorageArchivalRegistry {
    enum DatasetType {
        MARKET_TICKS,
        LEDGER_JOURNALS,
        AUDIT_EVENTS,
        POSTGRES_WAL,
        BESU_SNAPSHOTS
    }

    struct ArchiveRecord {
        bytes32 archiveId;
        DatasetType datasetType;
        uint64 startTimestamp;
        uint64 endTimestamp;
        uint64 recordCount;
        bytes32 sha256PayloadHash;
        bytes32 merkleTreeRoot;
        string glacierArchiveId;
        string vaultArn;
        uint256 blockTimestamp;
    }

    event ArchiveNotarized(
        bytes32 indexed archiveId,
        DatasetType indexed datasetType,
        bytes32 indexed merkleTreeRoot,
        bytes32 sha256PayloadHash,
        string glacierArchiveId,
        uint64 recordCount
    );

    /**
     * @notice Commits an immutable archive manifest record on-chain.
     * @dev Callable only by authorized HSM-backed Archival Worker relayer address.
     */
    function notarizeArchive(
        bytes32 archiveId,
        DatasetType datasetType,
        uint64 startTimestamp,
        uint64 endTimestamp,
        uint64 recordCount,
        bytes32 sha256PayloadHash,
        bytes32 merkleTreeRoot,
        string calldata glacierArchiveId,
        string calldata vaultArn
    ) external returns (bool);

    /**
     * @notice Verifies whether a given SHA-256 and Merkle root match the notarized archive.
     */
    function verifyArchiveIntegrity(
        bytes32 archiveId,
        bytes32 sha256PayloadHash,
        bytes32 merkleTreeRoot
    ) external view returns (bool isValid, uint256 blockTimestamp);
}
```

## Security & Compliance Notes
- **SEBI 8-Year WORM Mandate:** Under SEBI circular SEBI/HO/MIRSD/CIR/P/2018/144 and Stock Brokers Regulations 1992, records must remain fully unalterable and non-deletable for 8 continuous years. S3 Glacier Vault Lock in `COMPLIANCE` mode enforces this constraint at the physical AWS storage plane; even root account credentials cannot terminate, delete, or modify a locked vault archive during the retention window.
- **S3 Glacier Vault Lock vs S3 Object Lock:** While S3 Object Lock enforces WORM at the individual bucket object layer, S3 Glacier Vault Lock enforces hardware-backed compliance across the entire vault archive container. The vault policy is locked via a two-step handshake: `InitiateVaultLock` generates a Lock ID, which must be verified and locked via `CompleteVaultLock` within 24 hours. Once locked, the policy cannot be deleted or modified by any entity.
- **KMS Customer Managed Keys (CMK) Rotation & Protection:**
  - Encryption is performed using dedicated AWS KMS CMKs with annual automatic key rotation enabled.
  - The KMS key policy strictly forbids `kms:ScheduleKeyDeletion` and `kms:DeleteImportedKeyMaterial` for compliance keys.
  - Envelope encryption prevents plaintext data from ever touching network interfaces or S3 storage buffers.
- **Data Sovereignty & RBI Localization:** In adherence to RBI Master Directions on Storage of Payment System Data, all storage repositories, KMS hardware security modules, and compute workers are strictly physically localized in AWS Mumbai (`ap-south-1`) with disaster recovery replicas restricted to AWS Hyderabad (`ap-south-2`).
- **Zero Plaintext PII on Public/Consortium Ledger:** All on-chain notarization transactions strictly transmit irreversible SHA-256 hashes, Merkle roots, record counts, and storage identifiers. No customer Personally Identifiable Information (PII) or plaintext financial figures are ever written to the blockchain.

## Acceptance Criteria
- [ ] Terraform manifests provision an AWS S3 Glacier Vault with an initiated and locked WORM Vault Lock policy enforcing 2,920 days (8 years) mandatory retention.
- [ ] Attempted deletion of any Glacier archive or vault within the 8-year window is rejected by AWS with HTTP 403 Access Denied.
- [ ] ClickHouse `storage_configuration.xml` defines `hot_nvme_disk` and `s3_warm_disk` with automated `hot_to_warm_policy` operational without restart.
- [ ] Background ClickHouse daemon successfully migrates partitions older than 90 days from NVMe to S3 warm storage with zero disruption to active SELECT queries.
- [ ] PostgreSQL partition detachment script detaches closed monthly partitions concurrently without causing lock timeouts on live trade execution tables.
- [ ] Streaming Parquet exporter converts detached PostgreSQL tables to Zstandard-compressed Parquet with 100% precision preservation for monetary numeric columns.
- [ ] PostgreSQL WAL segment archiver packages 1,024-segment batches into compressed `.tar.zst` files with validated LSN continuous ranges.
- [ ] Besu snapshot worker exports state trie checkpoints and block receipts for 100,000-block intervals with verified Merkle proof generation.
- [ ] Zstandard compression achieves a minimum 70% compression ratio on tick data and ledger journal records.
- [ ] AWS KMS client-side envelope encryption applies unique AES-GCM-256 data keys per archive bundle before upload.
- [ ] Glacier multipart uploader streams multi-gigabyte archives with validated Glacier tree-hash checksums matching AWS calculations.
- [ ] Smart contract `StorageArchivalRegistry.sol` successfully records archive manifest proofs on Hyperledger Besu with emitted `ArchiveNotarized` events.
- [ ] Administrative CLI `storage-archival-worker verify` re-computes SHA-256 and Merkle roots and confirms 100% match with on-chain records.
- [ ] Primary database partitions are only dropped after successful Glacier receipt verification and minimum 12-block Besu transaction confirmation.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 218: Immutable Audit Log Service (source for audit event blocks and Merkle roots).
  - Prompt 401: PostgreSQL Schema Design (source for partitioned ledger and trade tables).
  - Prompt 405: Data Retention and Archival Policy (defines statutory regulatory retention windows).
  - Prompt 408: Historical Market Data Timescale & ClickHouse Pipeline (source for tick data multi-tier storage).
- **Parallel Tasks:**
  - Prompt 406: Backup & Restore Automation (coordinates snapshot backups alongside continuous cold tiering).
  - Prompt 409: Proof of Reserve Sparse Merkle Tree Store (coordinates cryptographic state attestations).
- **Downstream Tasks:**
  - Prompt 614: Clearing Member and Broker Capital Adequacy Portal (consumes archived records for compliance audits).
  - Prompt 714: Crosschain AML and Blockchain Forensics Screener (leverages archived Besu epoch snapshots).
