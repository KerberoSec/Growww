# 405 - Data Retention & Archival Policy Implementation

## Purpose
Financial institutions operating under SEBI (Securities and Exchange Board of India), RBI (Reserve Bank of India), and PMLA (Prevention of Money Laundering Act) regulations are legally mandated to retain all transactional, order book, customer identity verification, and communication records for a minimum of 8 years in an immutable, tamper-evident format. At the same time, privacy regulations such as India's Digital Personal Data Protection (DPDP) Act 2023 mandate strict data minimization and right-to-erasure for non-statutory personal data.

This prompt implements the automated data retention, partition offloading, cold-storage compression, and immutable WORM (Write Once, Read Many) archival policy engine. It automatically migrates historical transactional and ledger partitions from hot operational databases into cost-effective, cryptographically sealed S3 Glacier Deep Archive storage while enabling federated on-demand querying via AWS Athena and DuckDB for regulatory audits.

## What You Are Building
A production-grade, automated data retention and archival platform containing:
- **Archival Daemon Service:** A containerized worker (`services/archival-daemon/`) that orchestrates scheduled table partition detachment, Parquet conversion, encryption, and cold-storage offloading.
- **WORM Object Storage Infrastructure:** Terraform manifests (`infra/terraform/storage/archival_vault.tf`) provisioning AWS S3 Buckets with S3 Object Lock in strict `COMPLIANCE` mode with an 8-year retention lock.
- **High-Performance Parquet Exporter:** Python/PyArrow conversion pipeline transforming detached PostgreSQL / ClickHouse partitions into zstd-compressed, columnar Parquet files with SHA-256 manifest generation.
- **Cryptographic Seal Engine:** Client-side envelope encryption using AWS KMS / HashiCorp Vault ensuring data is encrypted prior to transport and storage.
- **Federated Query Interface:** AWS Athena / DuckDB configuration scripts enabling compliance officers to query cold archived Parquet datasets using standard SQL without restoring database backups.
- **DPDP Data Purging Engine:** Automated purge routines for transient, non-financial data (expired OTPs, temporary device sessions, unverified KYC drafts older than 30 days).

## Scope Boundaries
- **In Scope:** Partition lifecycle automation, S3 Glacier Object Lock policies, Parquet conversion, client-side encryption, integrity checksum verification, Athena schema cataloging, and DPDP non-financial data purging.
- **Out of Scope / Handled Elsewhere:**
 - Active PostgreSQL live table partitioning (handled in Prompt 401).
 - Disaster recovery snapshot backups (handled in Prompt 406).
 - ClickHouse live analytical querying (handled in Prompt 404).
 - Privacy policy and DPDP legal framework definitions (handled in Prompt 008).

## Technology to Use
Python 3.12 with PyArrow and AWS SDK (Boto3) / Go 1.22 is selected for the archival daemon. PyArrow provides high-throughput streaming export to Parquet format with superior ZSTD compression ratios (reducing raw transactional data size by up to 85%), while preserving strict column typing and decimal precision. AWS S3 Object Lock in COMPLIANCE mode guarantees that archived records cannot be overwritten, modified, or deleted by any user - including AWS root accounts - for the entire 8-year statutory period.

- **Daemon Runtime:** Python 3.12 (AsyncIO, PyArrow, Boto3) / Go 1.22.
- **Storage Format:** Apache Parquet with Zstandard (ZSTD) compression level 9.
- **Cold Storage Engine:** AWS S3 Glacier Flexible / Deep Archive with S3 Object Lock (Compliance Mode).
- **Federated SQL Engine:** AWS Athena / DuckDB for ad-hoc audit queries.
- **Infrastructure as Code:** Terraform 1.7+ for immutable vault provisioning.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Source transactional database for detached historical partitions (`trading.trades`, `ledger.journal_entries`).
- **ClickHouse 24.x:** Source OLAP cluster for detached tick and analytical data partitions.
- **AWS S3 Object Lock:** WORM-compliant storage in AWS Mumbai (`ap-south-1`).
- **AWS KMS / Vault Transit:** Envelope encryption keys managed under separate compliance security policies.
- **AWS Glue Data Catalog:** Catalogs archived Parquet partitions for Athena queries.

## Blockchain Interaction
The archival engine interfaces with the permissioned Hyperledger Besu ledger to provide mathematical proof of archive integrity.

### Detailed On-Chain Integration Mechanics:
- **Archive Merkle Root Attestation:** Upon generating an archived monthly Parquet dataset, the daemon calculates a Merkle root hash of all included records and transactions.
- **On-Chain Notarization:** This Merkle root, along with partition identifier, start timestamp, end timestamp, and S3 URI, is committed to the `ProofOfReserveRegistry.sol` or `ComplianceRegistry.sol` smart contract via an HSM-signed transaction.
- **Audit Verification:** During SEBI regulatory inspections, auditors can independently re-hash the S3 Parquet archive and verify that the calculated Merkle root matches the immutable hash committed on Hyperledger Besu years prior.
- **Zero PII on Ledger:** Only cryptographic file hashes, record counts, and partition metadata are submitted on-chain.

## Step-by-Step Build Instructions
1. Scaffold project directory: `services/archival-daemon/` and `infra/terraform/storage/`.
2. Author Terraform manifests provisioning the S3 Archival Bucket with Object Lock in `COMPLIANCE` mode, default retention of 2,920 days (8 years), and server-side KMS encryption.
3. Configure S3 Lifecycle rules transitioning uploaded archive objects to Glacier Flexible Archive after 30 days and Glacier Deep Archive after 180 days.
4. Implement PostgreSQL partition offloader in Python/Go: identify partitions older than 90 days, detach them using `ALTER TABLE ... DETACH PARTITION`.
5. Implement streaming data extractor reading detached partition rows and writing them to columnar Parquet files with ZSTD compression.
6. Compute SHA-256 checksum and Merkle tree root of the exported Parquet file.
7. Encrypt the Parquet archive using AWS KMS envelope encryption (AES-GCM-256).
8. Upload encrypted archive and checksum manifest to the S3 Object Lock bucket with `x-amz-object-lock-mode: COMPLIANCE`.
9. Submit the archive Merkle root and S3 metadata to the Hyperledger Besu `ProofOfReserveRegistry.sol` contract.
10. Register the new partition location in AWS Glue Data Catalog for instant SQL querying via AWS Athena.
11. Drop the detached PostgreSQL table partition after verifying successful S3 upload, checksum confirmation, and on-chain receipt.
12. Implement DPDP purging job: run daily cron deleting expired non-financial records (sessions, OTP logs, unverified onboarding drafts > 30 days).
13. Author end-to-end integration test validating export, encryption, WORM upload, on-chain commitment, and Athena query retrieval.

## Interfaces / Contracts

```yaml
# Data Retention Policy Matrix
Policies:
 - dataset: "trading.trades"
    hot_storage_days: 90 # PostgreSQL / ClickHouse NVMe
    warm_storage_days: 365 # S3 Standard-IA
    cold_glacier_years: 8 # S3 Glacier Deep Archive (WORM locked)
    purge_action: "LEGAL_HOLD_OR_PURGE_AFTER_8_YEARS"
    statutory_basis: "SEBI Stock Brokers Regulations 1992 & PMLA 2002"

 - dataset: "ledger.journal_entries"
    hot_storage_days: 180
    warm_storage_days: 365
    cold_glacier_years: 8
    purge_action: "LEGAL_HOLD_OR_PURGE_AFTER_8_YEARS"
    statutory_basis: "Companies Act 2013 & SEBI Regulations"

 - dataset: "identity.kyc_audit_logs"
    hot_storage_days: 365
    warm_storage_days: 730
    cold_glacier_years: 8
    purge_action: "PURGE_AFTER_8_YEARS_UNLESS_ACTIVE_ACCOUNT"
    statutory_basis: "RBI Master Direction - KYC & PMLA 2002"

 - dataset: "auth.session_tokens_and_otp"
    hot_storage_days: 30
    warm_storage_days: 0
    cold_glacier_years: 0
    purge_action: "HARD_DELETE_AFTER_30_DAYS"
    statutory_basis: "DPDP Act 2023 (Data Minimization)"
```

```json
// Archival Manifest Specification: archive_manifest.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "GrowwwArchiveManifest",
  "type": "object",
  "properties": {
    "archiveId": {"type": "string", "format": "uuid"},
    "datasetName": {"type": "string", "enum": ["trading.trades", "ledger.journal_entries", "custody.allocations"]},
    "partitionPeriod": {"type": "string", "pattern": "^[0-9]{4}-[0-9]{2}$"},
    "recordCount": {"type": "integer", "minimum": 1},
    "s3Uri": {"type": "string", "pattern": "^s3://.*\\.parquet\\.enc$"},
    "sha256Checksum": {"type": "string", "pattern": "^[a-f0-9]{64}$"},
    "merkleRootHash": {"type": "string", "pattern": "^0x[a-f0-9]{64}$"},
    "onChainTxHash": {"type": "string", "pattern": "^0x[a-f0-9]{64}$"},
    "onChainBlockNumber": {"type": "integer"},
    "kmsKeyId": {"type": "string"},
    "compressionCodec": {"type": "string", "enum": ["zstd"]},
    "retentionLockUntil": {"type": "string", "format": "date-time"},
    "archivedAt": {"type": "string", "format": "date-time"}
  },
  "required": [
    "archiveId", "datasetName", "partitionPeriod", "recordCount",
    "s3Uri", "sha256Checksum", "merkleRootHash", "onChainTxHash", "retentionLockUntil"
  ]
}
```

```hcl
# Terraform S3 WORM Compliance Vault Manifest (Snippet)
resource "aws_s3_bucket" "compliance_archive" {
  bucket = "growww-compliance-archive-prod-ap-south-1"

  object_lock_configuration {
    object_lock_enabled = "Enabled"
    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 2920 # 8 Years
      }
    }
  }
}
```

## Security & Compliance Notes
- **WORM Immutability:** S3 Object Lock configured strictly in `COMPLIANCE` mode prevents deletion or alteration by any IAM identity, including administrative or root credentials, guaranteeing full compliance with SEBI and PMLA requirements.
- **Client-Side Envelope Encryption:** Data is encrypted locally before transmission using AWS KMS keys with restricted IAM key policies. Unencrypted financial records never touch network transit or storage layers.
- **Data Sovereignty:** All archives, KMS keys, and Glacier storage repositories reside strictly within AWS Mumbai (`ap-south-1`) in compliance with RBI data localization mandates.
- **DPDP Act Compliance:** Non-statutory user telemetry, unverified onboarding drafts, and temporary auth artifacts are purged automatically via daily cleanup jobs to prevent unauthorized retention.

## Acceptance Criteria
- [ ] Terraform manifests provision S3 Bucket with Object Lock in `COMPLIANCE` mode for 8-year (2,920-day) retention.
- [ ] Archival daemon successfully detaches historical PostgreSQL partitions and exports rows to ZSTD-compressed Parquet.
- [ ] SHA-256 and Merkle root calculation verified against source database rows with 100% mathematical consistency.
- [ ] Client-side KMS envelope encryption implemented and verified before S3 upload.
- [ ] On-chain transaction successfully commits archive Merkle root to `ProofOfReserveRegistry.sol` on Hyperledger Besu.
- [ ] AWS Glue Data Catalog registers Parquet partitions; AWS Athena executes analytical SQL query successfully against cold Parquet files.
- [ ] Attempted deletion of WORM-locked archive object rejected with HTTP 403 Access Denied by AWS S3.
- [ ] DPDP daily purge job successfully removes expired non-financial records older than 30 days.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 008 (Data Protection Policy), Prompt 401 (PostgreSQL Schema), Prompt 404 (Data Warehouse), Prompt 707 (Data Encryption).
- **Parallel Tasks:** Prompt 406 (Backup & Restore Automation), Prompt 216 (Reporting Service).
