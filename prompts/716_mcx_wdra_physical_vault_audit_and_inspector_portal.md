# 716 - MCX & WDRA Physical Vault Audit & Inspector Portal (Assayer Attestations, Bar Weighment, e-NWR Verification)

## Purpose
Establishes a zero-trust, institutional-grade security, governance, and physical audit microservice (`services/vault-audit-portal`) for physical commodity backing verification. Operating in strict adherence to **SEBI Commodity Derivatives Regulations**, **MCX Physical Delivery and Vault Accreditation Norms**, **WDRA (Warehousing Development and Regulatory Authority) Act, 2007 Guidelines**, and **IFSCA Bullion Custody Frameworks**, this service provides end-to-end cryptographic verification that all fractional commodity tokens (such as physical Gold 999.9, Silver 999, and precious metals) minted on the permissioned settlement ledger are 100% backed by verified, segregated, and unencumbered physical bars stored in accredited vaults.

The service provides a dedicated, tamper-evident portal and secure API layer for certified physical vault inspectors, NABL-accredited assayers, and accredited vault custodians (such as Brink's, Sequel, Malca-Amit, and MMTC-PAMP). It ingests IoT calibrated scale weighment telemetry, verifies electronic Negotiable Warehouse Receipts (e-NWRs) issued by WDRA repositories (NERL and CCRL), validates assay spectroscopic purity certificates, and computes cryptographic Proof-of-Reserve (PoR) Merkle trees for immutable settlement on Hyperledger Besu.

## What You Are Building
A high-integrity, multi-party governance and audit microservice (`services/vault-audit-portal`):
- **Inspector & Assayer Multi-Role Attestation Engine:** Authenticated portal and gRPC API for accredited physical inspectors and assaying laboratories to submit cryptographic assay certificates, spectroscopic purity readings, ultrasonic density logs, and optical inspection data.
- **Calibrated Scale & IoT Weighment Telemetry Ingestor:** Digitally signed weighment telemetry ingestion from certified, tamper-sealed Mettler Toledo and Sartorius precision digital scales with microgram resolution, verifying gross weight, net weight, and tare weight against manifest declarations.
- **WDRA e-NWR / Repository Integrity Validator:** Real-time connector to NERL (National E-Repository Limited) and CCRL (CDSL Commodity Repository Limited) to verify electronic Negotiable Warehouse Receipts (e-NWRs), lock/lien status, quality parameters, and unencumbered ownership.
- **Physical-to-Digital Reconciliation & Merkle Root Builder:** Reconciles physical vault bar inventory against on-chain token supply, constructing deterministic SHA-256 Merkle trees of all certified vault reserves.
- **Quorum-Based Multi-Signature Attestation State Machine:** Requires M-of-N threshold cryptographic signatures from accredited inspectors, assayers, and vault managers before attestation packages can update on-chain reserve registries.
- **Vault Discrepancy & Circuit Breaker Dispatcher:** Detects weight variances ($> 0.001\%$), bar serial collisions, purity deviations ($< 999.9$ fine gold), or missing physical audit cycles, instantly publishing emergency mint freeze directives to the on-chain Compliance Registry and Risk Engine.

## Scope Boundaries
- **In Scope:**
  - Role-based attestation workflows for accredited Vault Inspectors, NABL-certified Assayers, and Vault Custodians.
  - Bar-level registry tracking: serial number, refinery brand (e.g. MMTC-PAMP, Valcambi, Argor-Heraeus, Rand Refinery), melt lot ID, fineness/purity grade, gross/net weight.
  - Direct integration with WDRA repositories (NERL and CCRL) for e-NWR lifecycle verification, pledge validation, and lien-free status checks.
  - Cryptographic digital signing (ECDSA secp256k1 / Ed25519) of inspection reports, weighment slips, and assay certificates.
  - Automated comparison between physical vault audit records and on-chain Proof-of-Reserve balances.
  - Audit case management, physical seal tracking, and discrepancy quarantine workflows.
  - Publishing Proof-of-Reserve attestation hashes and audit events to Kafka and Hyperledger Besu.
- **Out of Scope / Handled Elsewhere:**
  - Smart contract token minting and burning execution (handled in Prompt 301 and Prompt 303).
  - High-frequency order matching and secondary trading of commodity tokens (handled in Prompt 205).
  - Fiat payments and banking settlements for commodity delivery (handled in Prompt 203 and Prompt 214).
  - General customer KYC and investor AML screening (handled in Prompt 202 and Prompt 714).
  - Armored truck transport logistics and physical delivery dispatch (handled in logistics and transit management service).

## Technology to Use
- **Core Microservice:** Go 1.22+ for secure, low-latency attestation processing, cryptographic signature validation, and high-concurrency external API orchestration.
- **Relational Audit & Vault Registry Database:** PostgreSQL 16 with Row-Level Security (RLS) and PostGIS for vault geofencing and inspector location verification.
- **Time-Series Telemetry & Audit Trail:** TimescaleDB / ClickHouse 24+ for high-volume IoT scale weight streams, environmental telemetry (vault temperature/humidity), and tamper sensor logs.
- **In-Memory State & Attestation Staging:** Redis 7 Cluster for caching active inspection sessions, OTP/MFA states, and pending multi-sig attestation quorum tracking.
- **Stream Ingestion & Event Bus:** Apache Kafka for consuming scale weighment events, e-NWR status updates, and publishing compliance screening decisions.
- **Immutable Object & Document Storage:** S3-compatible immutable object store (MinIO / AWS S3 with Object Lock / WORM compliance) for tamper-evident storage of PDF assay certificates, XRF spectrum curves, and high-resolution bar photographs.
- **Key Custody & Hardware Security:** Hardware Security Modules (HSM) / HashiCorp Vault for inspector key management, X.509 PKI certificate authority, and digital signature validation.
- **Justification:** Physical asset backing verification requires zero-trust cryptographic guarantees, multi-party signature aggregation, and resilient connections to external statutory repositories (WDRA/NERL/CCRL). Go provides high concurrency, strict memory safety, and native cryptographic primitives for institutional attestation workflows.

## Backend / Infra Touchpoints
- **Upstream Microservices & Systems:**
  - `services/custodian-adapter` (Prompt 213): Ingests institutional depository balances and vault holding manifests.
  - External Assayer Systems / NABL Labs: Inbound API/portal submissions of chemical assay certificates and XRF scans.
  - External WDRA Repositories: NERL and CCRL API gateways for e-NWR status verification.
  - Physical Vault IoT Gateways: Digitally signed weighment telemetry from calibrated scales.
- **Downstream Microservices:**
  - `services/proof-of-reserve-relayer` (Prompt 303 / 319): Consumes attestation Merkle roots to update on-chain reserve registries.
  - `services/risk-engine` (Prompt 206): Ingests vault discrepancy alerts to freeze token minting or trigger margin adjustments.
  - `services/regulatory-reporting` (Prompts 216 & 233): Consumes statutory audit reports for SEBI, MCX, and WDRA compliance filings.
  - `services/admin-back-office` (Prompts 217, 604 & 605): Feeds compliance officer dashboard and auditor workbench.
- **Messaging Topics (Apache Kafka):**
  - Consumes: `vault.weighment.telemetry.v1`, `wdra.enwr.update_received.v1`.
  - Publishes: `vault.audit.attestation_completed.v1`, `vault.audit.discrepancy_flagged.v1`, `vault.audit.por_proof_ready.v1`, `compliance.vault.mint_freeze_command.v1`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **Physical Asset Vault Registry (`PhysicalAssetVaultRegistry.sol` / `ProofOfReserveRegistry.sol`):**
  - Periodic and on-demand anchoring of vault inventory Merkle root:
    $$\text{VaultInventoryMerkleRoot} = \text{MerkleRoot}(\text{SHA256}(\text{BarID} \parallel \text{GrossWeight} \parallel \text{FinenessBps} \parallel \text{VaultCode}))$$
  - Cryptographic Multi-Party Attestation: The smart contract enforces an M-of-N threshold signature requirement from accredited inspector, assayer, and custodian public keys before updating reserve balances.
  - Automated Mint Circuit Breaker: If an audit attestation reveals a negative weight variance, purity failure, or missing physical custody confirmation, `VaultAuditPortal` dispatches a high-priority call to `ProofOfReserveRegistry.emergencyPauseMinting(bytes32 vaultId, bytes32 discrepancyReasonHash)`.
  - Zero PII on Ledger: Only anonymized bar hashes, refinery codes, certified gross/fine weights, vault identifier hashes, and cryptographic signature digests are recorded on-chain.

## Step-by-Step Build Instructions
1. Scaffold the `services/vault-audit-portal` repository in Go 1.22+ following the standard workspace layout (Prompt 106) with linting, formatting, and strict type checking (Prompt 107).
2. Generate Go and Python gRPC stubs from the Protobuf definitions in `proto/growww/vault_audit/v1/vault_audit.proto`.
3. Design and execute PostgreSQL 16 database migrations for tables `vault_facilities`, `vault_bar_inventory`, `vault_weighment_receipts`, `assay_attestation_certificates`, `wdra_enwr_records`, `vault_audit_sessions`, and `vault_audit_discrepancies`.
4. Configure TimescaleDB / ClickHouse schema and table engines for high-frequency IoT scale weighment streams and environmental telemetry.
5. Implement the **PKI & Inspector Identity Module**:
   - Validate X.509 client certificates and hardware-backed ECDSA/Ed25519 digital signatures for all registered inspectors, assayers, and vault managers.
   - Enforce WebAuthn / FIDO2 Multi-Factor Authentication for web portal sessions.
6. Implement the **Assayer Certificate & Quality Verification Engine**:
   - Parse standardized assay reports (NABL accredited laboratory format, XRF spectrometer readings, ultrasonic density logs).
   - Validate purity thresholds (minimum 999.9 fine gold basis points = 9999, minimum 999.0 silver = 9990).
   - Verify refinery hallmark against London Bullion Market Association (LBMA) and Bureau of Indian Standards (BIS) recognized refinery lists.
7. Implement the **Calibrated Scale IoT Telemetry Ingestor**:
   - Ingest cryptographically signed payloads from calibrated precision scales via mTLS / MQTT over TLS.
   - Verify scale calibration certificate validity dates, scale serial numbers, and tamper seal status.
   - Perform automated tare weight validation and gross-to-net weight calculation with microgram accuracy.
8. Implement the **WDRA e-NWR Repository Connector**:
   - Connect to NERL and CCRL repository gateways using mutual TLS and signed JSON-RPC / REST APIs.
   - Synchronize electronic Negotiable Warehouse Receipt (e-NWR) states, verifying unencumbered ownership, valid validity dates, and zero pledge/lien encumbrance.
   - Flag receipts nearing expiry ($< 30$ days) or flagged with legal disputes.
9. Implement the **Physical-to-Digital Reconciliation Engine**:
   - Compare physical bar inventory (total fine grams) against on-chain tokenized supply (`services/proof-of-reserve-relayer`).
   - Flag any negative variance ($\text{Physical Weight} < \text{Tokenized Supply}$) or unassigned bar serial numbers.
10. Implement the **Cryptographic Attestation & Merkle Proof Generator**:
    - Aggregate verified bar inventory into canonical sorted order by bar hash.
    - Compute SHA-256 Merkle tree root and generate individual Merkle inclusion proofs for each bar.
    - Assemble attestation package containing Merkle root, timestamp, session ID, and M-of-N inspector signatures.
11. Implement the **Discrepancy & Anomaly Workflow State Machine**:
    - Transition flagged items into `UNDER_INVESTIGATION`, `QUARANTINED`, or `RESOLVED` states.
    - Publish `vault.audit.discrepancy_flagged.v1` and trigger `compliance.vault.mint_freeze_command.v1` when severity is `CRITICAL`.
12. Build the **WORM Document Archive Integrator**:
    - Store immutable PDF certificates, XRF raw scan data, scale calibration logs, and bar macro photos in S3 with Object Lock enabled.
    - Store SHA-256 content hashes in PostgreSQL audit tables.
13. Configure Prometheus metrics (`vault_audit_attestations_total`, `vault_audit_discrepancy_count`, `vault_scale_weighment_latency_ms`, `vault_enwr_sync_status`), OpenTelemetry distributed tracing, and structured JSON audit logging.

## Interfaces / Contracts

### Protobuf Definition (`vault_audit.proto`)
```protobuf
syntax = "proto3";

package growww.vault_audit.v1;

option go_package = "github.com/growww/services/vault-audit-portal/gen/v1;vaultauditv1";

service VaultAuditService {
  rpc SubmitBarWeighment (SubmitBarWeighmentRequest) returns (SubmitBarWeighmentResponse);
  rpc SubmitAssayAttestation (SubmitAssayAttestationRequest) returns (SubmitAssayAttestationResponse);
  rpc VerifyEnwrReceipt (VerifyEnwrReceiptRequest) returns (VerifyEnwrReceiptResponse);
  rpc CreateAuditSession (CreateAuditSessionRequest) returns (CreateAuditSessionResponse);
  rpc SubmitVaultAuditReport (SubmitVaultAuditReportRequest) returns (SubmitVaultAuditReportResponse);
  rpc GetVaultInventory (GetVaultInventoryRequest) returns (GetVaultInventoryResponse);
  rpc GetProofOfReserveSnapshot (GetProofOfReserveSnapshotRequest) returns (GetProofOfReserveSnapshotResponse);
  rpc ResolveDiscrepancy (ResolveDiscrepancyRequest) returns (ResolveDiscrepancyResponse);
}

enum CommodityType {
  COMMODITY_TYPE_UNSPECIFIED = 0;
  GOLD_9999 = 1;
  GOLD_9950 = 2;
  SILVER_9990 = 3;
  PLATINUM_9995 = 4;
}

enum AuditType {
  AUDIT_TYPE_UNSPECIFIED = 0;
  INBOUND_INGRESS = 1;
  PERIODIC_CYCLE_COUNT = 2;
  RANDOM_SAMPLE_ASSAY = 3;
  OUTBOUND_EGRESS = 4;
  REGULATORY_SURPRISE_AUDIT = 5;
}

enum AttestationStatus {
  ATTESTATION_STATUS_UNSPECIFIED = 0;
  DRAFT = 1;
  PENDING_ASSAYER_SIGNATURE = 2;
  PENDING_INSPECTOR_SIGNATURE = 3;
  PENDING_CUSTODIAN_SIGNATURE = 4;
  ATTESTATION_COMPLETED = 5;
  ATTESTATION_REJECTED = 6;
  QUARANTINED = 7;
}

enum DiscrepancySeverity {
  SEVERITY_UNSPECIFIED = 0;
  LOW_DOCUMENTATION = 1;
  MEDIUM_SEAL_MISMATCH = 2;
  HIGH_WEIGHT_VARIANCE = 3;
  CRITICAL_PURITY_FAILURE = 4;
  CRITICAL_BAR_MISSING = 5;
}

enum EnwrRepository {
  REPOSITORY_UNSPECIFIED = 0;
  NERL = 1;
  CCRL = 2;
}

message SubmitBarWeighmentRequest {
  string session_id = 1;
  string vault_id = 2;
  string bar_serial_number = 3;
  CommodityType commodity_type = 4;
  string scale_device_id = 5;
  string scale_calibration_cert_id = 6;
  double gross_weight_grams = 7;
  double tare_weight_grams = 8;
  double net_weight_grams = 9;
  string weighment_raw_telemetry = 10;
  string scale_digital_signature = 11;
  int64 weighment_timestamp_utc = 12;
}

message SubmitBarWeighmentResponse {
  string weighment_receipt_id = 1;
  bool weight_variance_acceptable = 2;
  double manifest_variance_percentage = 3;
  string weighment_hash = 4;
  int64 recorded_at_utc = 5;
}

message SubmitAssayAttestationRequest {
  string session_id = 1;
  string vault_id = 2;
  string bar_serial_number = 3;
  string refinery_name = 4;
  string refinery_batch_number = 5;
  string assay_laboratory_id = 6;
  string nabl_accreditation_number = 7;
  CommodityType commodity_type = 8;
  int32 purity_fineness_bps = 9; // e.g. 9999 for 99.99%
  double ultrasonic_velocity_m_per_s = 10;
  double xrf_purity_reading_pct = 11;
  string assay_certificate_document_s3_uri = 12;
  string assay_certificate_sha256 = 13;
  string assayer_public_key = 14;
  string assayer_signature = 15;
  int64 assay_timestamp_utc = 16;
}

message SubmitAssayAttestationResponse {
  string attestation_id = 1;
  bool purity_acceptable = 2;
  string attestation_hash = 3;
  int64 recorded_at_utc = 4;
}

message VerifyEnwrReceiptRequest {
  EnwrRepository repository = 1;
  string enwr_number = 2;
  string vault_id = 3;
  string beneficiary_client_code = 4;
}

message VerifyEnwrReceiptResponse {
  string enwr_id = 1;
  bool is_valid = 2;
  bool is_pledged_or_encumbered = 3;
  string commodity_description = 4;
  double total_weight_grams = 5;
  int32 purity_basis_points = 6;
  string warehouse_receipt_status = 7;
  int64 validity_start_utc = 8;
  int64 validity_end_utc = 9;
  string repository_sync_hash = 10;
}

message CreateAuditSessionRequest {
  string vault_id = 1;
  AuditType audit_type = 2;
  string lead_inspector_id = 3;
  string assayer_id = 4;
  string custodian_representative_id = 5;
  repeated string expected_bar_serials = 6;
  string audit_mandate_reference = 7;
}

message CreateAuditSessionResponse {
  string session_id = 1;
  string vault_id = 2;
  int64 opened_at_utc = 3;
}

message SubmitVaultAuditReportRequest {
  string session_id = 1;
  string lead_inspector_id = 2;
  int32 total_bars_audited = 3;
  double total_gross_weight_grams = 4;
  double total_fine_weight_grams = 5;
  int32 discrepancies_count = 6;
  string inspection_notes = 7;
  string report_document_s3_uri = 8;
  string report_sha256 = 9;
  string inspector_signature = 10;
  string custodian_signature = 11;
  string assayer_signature = 12;
}

message SubmitVaultAuditReportResponse {
  string report_id = 1;
  string session_id = 2;
  AttestationStatus status = 3;
  string inventory_merkle_root = 4;
  string on_chain_attestation_tx_hash = 5;
  int64 finalized_at_utc = 6;
}

message GetVaultInventoryRequest {
  string vault_id = 1;
  CommodityType commodity_type = 2;
  int32 page_size = 3;
  string page_token = 4;
}

message VaultBarSummary {
  string bar_serial_number = 1;
  string refinery_name = 2;
  CommodityType commodity_type = 3;
  int32 purity_fineness_bps = 4;
  double gross_weight_grams = 5;
  double fine_weight_grams = 6;
  string enwr_number = 7;
  string last_weighed_receipt_id = 8;
  string last_assay_attestation_id = 9;
  bool is_quarantined = 10;
}

message GetVaultInventoryResponse {
  string vault_id = 1;
  repeated VaultBarSummary bars = 2;
  double total_gross_weight_grams = 3;
  double total_fine_weight_grams = 4;
  string next_page_token = 5;
}

message GetProofOfReserveSnapshotRequest {
  string vault_id = 1;
  CommodityType commodity_type = 2;
}

message GetProofOfReserveSnapshotResponse {
  string snapshot_id = 1;
  string vault_id = 2;
  CommodityType commodity_type = 3;
  int32 total_physical_bars = 4;
  double total_physical_fine_weight_grams = 5;
  double tokenized_circulating_supply_grams = 6;
  double reserve_ratio = 7; // e.g. 1.0000 for exact 100% backing
  string inventory_merkle_root = 8;
  string besu_block_hash = 9;
  int64 block_number = 10;
  int64 timestamp_utc = 11;
}

message ResolveDiscrepancyRequest {
  string discrepancy_id = 1;
  string resolution_officer_id = 2;
  string resolution_action = 3; // "REWEIGHED_ACCEPTED", "BAR_REPLACED", "VAULT_ADJUSTMENT"
  string resolution_justification = 4;
  string supporting_document_s3_uri = 5;
  string mfa_token = 6;
}

message ResolveDiscrepancyResponse {
  string discrepancy_id = 1;
  bool is_resolved = 2;
  int64 resolved_at_utc = 3;
}
```

### PostgreSQL Database Schema
```sql
CREATE TYPE commodity_type_enum AS ENUM (
    'GOLD_9999',
    'GOLD_9950',
    'SILVER_9990',
    'PLATINUM_9995'
);

CREATE TYPE audit_type_enum AS ENUM (
    'INBOUND_INGRESS',
    'PERIODIC_CYCLE_COUNT',
    'RANDOM_SAMPLE_ASSAY',
    'OUTBOUND_EGRESS',
    'REGULATORY_SURPRISE_AUDIT'
);

CREATE TYPE attestation_status_enum AS ENUM (
    'DRAFT',
    'PENDING_ASSAYER_SIGNATURE',
    'PENDING_INSPECTOR_SIGNATURE',
    'PENDING_CUSTODIAN_SIGNATURE',
    'ATTESTATION_COMPLETED',
    'ATTESTATION_REJECTED',
    'QUARANTINED'
);

CREATE TYPE discrepancy_severity_enum AS ENUM (
    'LOW_DOCUMENTATION',
    'MEDIUM_SEAL_MISMATCH',
    'HIGH_WEIGHT_VARIANCE',
    'CRITICAL_PURITY_FAILURE',
    'CRITICAL_BAR_MISSING'
);

CREATE TYPE enwr_repository_enum AS ENUM (
    'NERL',
    'CCRL'
);

CREATE TABLE vault_facilities (
    vault_id VARCHAR(64) PRIMARY KEY,
    vault_name VARCHAR(128) NOT NULL,
    custodian_company_name VARCHAR(128) NOT NULL, -- e.g. "Brink's India", "Sequel Logistics"
    mcx_accreditation_number VARCHAR(64) NOT NULL UNIQUE,
    wdra_registration_number VARCHAR(64) NOT NULL UNIQUE,
    facility_address TEXT NOT NULL,
    latitude NUMERIC(9, 6) NOT NULL,
    longitude NUMERIC(9, 6) NOT NULL,
    geofence_radius_meters INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_audit_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_facilities(vault_id),
    audit_type audit_type_enum NOT NULL,
    lead_inspector_id UUID NOT NULL,
    assayer_id UUID NOT NULL,
    custodian_representative_id UUID NOT NULL,
    audit_mandate_reference VARCHAR(128) NOT NULL,
    status attestation_status_enum NOT NULL DEFAULT 'DRAFT',
    total_bars_audited INT NOT NULL DEFAULT 0,
    total_gross_weight_grams NUMERIC(14, 4) NOT NULL DEFAULT 0.0000,
    total_fine_weight_grams NUMERIC(14, 4) NOT NULL DEFAULT 0.0000,
    inventory_merkle_root CHAR(64),
    on_chain_attestation_tx_hash VARCHAR(66),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ
);

CREATE TABLE vault_bar_inventory (
    bar_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_facilities(vault_id),
    bar_serial_number VARCHAR(64) NOT NULL,
    refinery_name VARCHAR(128) NOT NULL,
    refinery_batch_number VARCHAR(64) NOT NULL,
    commodity_type commodity_type_enum NOT NULL,
    purity_fineness_bps INT NOT NULL, -- e.g. 9999
    gross_weight_grams NUMERIC(14, 4) NOT NULL,
    fine_weight_grams NUMERIC(14, 4) NOT NULL,
    enwr_number VARCHAR(64),
    tamper_bag_seal_number VARCHAR(64) NOT NULL,
    is_quarantined BOOLEAN NOT NULL DEFAULT FALSE,
    last_audit_session_id UUID REFERENCES vault_audit_sessions(session_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_vault_bar_serial UNIQUE (refinery_name, bar_serial_number)
);

CREATE TABLE vault_weighment_receipts (
    weighment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES vault_audit_sessions(session_id),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_facilities(vault_id),
    bar_id UUID NOT NULL REFERENCES vault_bar_inventory(bar_id),
    scale_device_id VARCHAR(64) NOT NULL,
    scale_calibration_cert_id VARCHAR(64) NOT NULL,
    gross_weight_grams NUMERIC(14, 4) NOT NULL,
    tare_weight_grams NUMERIC(14, 4) NOT NULL,
    net_weight_grams NUMERIC(14, 4) NOT NULL,
    manifest_weight_grams NUMERIC(14, 4) NOT NULL,
    variance_percentage NUMERIC(7, 5) NOT NULL,
    scale_digital_signature TEXT NOT NULL,
    weighment_hash CHAR(64) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assay_attestation_certificates (
    certificate_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES vault_audit_sessions(session_id),
    bar_id UUID NOT NULL REFERENCES vault_bar_inventory(bar_id),
    assay_laboratory_id UUID NOT NULL,
    nabl_accreditation_number VARCHAR(64) NOT NULL,
    measured_fineness_bps INT NOT NULL,
    ultrasonic_velocity_m_per_s NUMERIC(8, 2) NOT NULL,
    xrf_purity_reading_pct NUMERIC(6, 3) NOT NULL,
    document_s3_uri VARCHAR(256) NOT NULL,
    document_sha256 CHAR(64) NOT NULL,
    assayer_signature TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wdra_enwr_records (
    enwr_record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enwr_number VARCHAR(64) NOT NULL UNIQUE,
    repository enwr_repository_enum NOT NULL,
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_facilities(vault_id),
    beneficiary_client_code VARCHAR(64) NOT NULL,
    commodity_type commodity_type_enum NOT NULL,
    total_weight_grams NUMERIC(14, 4) NOT NULL,
    purity_bps INT NOT NULL,
    is_pledged BOOLEAN NOT NULL DEFAULT FALSE,
    is_lien_marked BOOLEAN NOT NULL DEFAULT FALSE,
    receipt_status VARCHAR(32) NOT NULL,
    validity_start TIMESTAMPTZ NOT NULL,
    validity_end TIMESTAMPTZ NOT NULL,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_audit_discrepancies (
    discrepancy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES vault_audit_sessions(session_id),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_facilities(vault_id),
    bar_id UUID REFERENCES vault_bar_inventory(bar_id),
    severity discrepancy_severity_enum NOT NULL,
    discrepancy_type VARCHAR(64) NOT NULL, -- "WEIGHT_MISMATCH", "PURITY_DEVIATION", "SEAL_BROKEN", "ENWR_PLEDGED"
    discrepancy_details JSONB NOT NULL,
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolution_officer_id UUID,
    resolution_action VARCHAR(64),
    resolution_justification TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vbi_vault_commodity ON vault_bar_inventory (vault_id, commodity_type);
CREATE INDEX idx_vbi_serial ON vault_bar_inventory (bar_serial_number);
CREATE INDEX idx_vas_status ON vault_audit_sessions (status, opened_at DESC);
CREATE INDEX idx_vwr_session ON vault_weighment_receipts (session_id);
CREATE INDEX idx_enwr_num ON wdra_enwr_records (enwr_number);
CREATE INDEX idx_vad_unresolved ON vault_audit_discrepancies (vault_id, is_resolved) WHERE NOT is_resolved;
```

### Kafka Event Schemas

#### Topic: `vault.audit.attestation_completed.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "VaultAuditAttestationCompletedEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "session_id": { "type": "string", "format": "uuid" },
    "vault_id": { "type": "string" },
    "audit_type": { "type": "string", "enum": ["INBOUND_INGRESS", "PERIODIC_CYCLE_COUNT", "RANDOM_SAMPLE_ASSAY", "OUTBOUND_EGRESS", "REGULATORY_SURPRISE_AUDIT"] },
    "commodity_type": { "type": "string", "enum": ["GOLD_9999", "GOLD_9950", "SILVER_9990", "PLATINUM_9995"] },
    "total_bars_audited": { "type": "integer", "minimum": 1 },
    "total_gross_weight_grams": { "type": "number", "minimum": 0.0001 },
    "total_fine_weight_grams": { "type": "number", "minimum": 0.0001 },
    "inventory_merkle_root": { "type": "string", "minLength": 64, "maxLength": 64 },
    "on_chain_tx_hash": { "type": "string" },
    "lead_inspector_id": { "type": "string", "format": "uuid" },
    "assayer_id": { "type": "string", "format": "uuid" },
    "custodian_id": { "type": "string", "format": "uuid" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "session_id",
    "vault_id",
    "audit_type",
    "commodity_type",
    "total_bars_audited",
    "total_gross_weight_grams",
    "total_fine_weight_grams",
    "inventory_merkle_root",
    "lead_inspector_id",
    "assayer_id",
    "custodian_id",
    "timestamp_utc"
  ]
}
```

#### Topic: `vault.audit.discrepancy_flagged.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "VaultAuditDiscrepancyFlaggedEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "discrepancy_id": { "type": "string", "format": "uuid" },
    "session_id": { "type": "string", "format": "uuid" },
    "vault_id": { "type": "string" },
    "bar_serial_number": { "type": "string" },
    "refinery_name": { "type": "string" },
    "severity": { "type": "string", "enum": ["LOW_DOCUMENTATION", "MEDIUM_SEAL_MISMATCH", "HIGH_WEIGHT_VARIANCE", "CRITICAL_PURITY_FAILURE", "CRITICAL_BAR_MISSING"] },
    "discrepancy_type": { "type": "string" },
    "expected_value": { "type": "string" },
    "measured_value": { "type": "string" },
    "variance_percentage": { "type": "number" },
    "emergency_mint_halt_triggered": { "type": "boolean" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "discrepancy_id",
    "session_id",
    "vault_id",
    "severity",
    "discrepancy_type",
    "emergency_mint_halt_triggered",
    "timestamp_utc"
  ]
}
```

## Security & Compliance Notes
- **SEBI & MCX Vault Accreditation Compliance:** Enforces full compliance with MCX physical delivery standards, requiring physical inspections by certified independent auditors at least once every quarter and random spot-checks without advance notice.
- **WDRA e-NWR Electronic Ownership Safeguards:** Ensures that no commodity token can be issued on-chain without an active, unencumbered, and verified electronic Negotiable Warehouse Receipt (e-NWR) registered with NERL or CCRL.
- **Hardware-Enforced IoT Weighment Integrity:** Scales are paired with IoT edge cryptographic modules utilizing hardware secure elements (ATECC608A / TPM 2.0) that digitally sign raw serial weighment streams. Any attempt to spoof scale data or inject simulated weight readings fails digital signature verification.
- **NABL Laboratory Assaying Purity Mandate:** Assayers must hold active NABL (National Accreditation Board for Testing and Calibration Laboratories) accreditation. Assay reports must include dual spectroscopic (XRF) and ultrasonic velocity measurements to detect gold-plated tungsten bars.
- **Multi-Signature Quorum for Reserve Updates:** No single entity (inspector, assayer, or vault manager) can unilaterally approve reserve balances. Updating on-chain reserves requires an M-of-N quorum (typically 2 of 3 or 3 of 3) with separate private keys stored in HSMs.
- **Zero PII on Public and Ledger Channels:** Bar serials and vault codes are hashed or recorded with asset-only technical metadata. No investor names, PAN numbers, or beneficial ownership identifiers are ever committed to the ledger or exposed on inspector portals.

## Acceptance Criteria
- [ ] End-to-end audit session lifecycle (session creation, scale weighment ingestion, assay certification, and Merkle root generation) completes with p99 API latency $< 1,000\text{ms}$.
- [ ] Successfully validates 100% of digitally signed weighment packets from simulated Mettler Toledo and Sartorius IoT scale endpoints.
- [ ] Rejects any bar weighment exhibit where gross weight differs from manifest declaration by $> 0.001\%$, automatically opening a `HIGH_WEIGHT_VARIANCE` discrepancy.
- [ ] Successfully connects to NERL and CCRL mock repository endpoints and automatically flags pledged or lien-marked e-NWRs.
- [ ] Rejects any assay certificate submission with gold fineness $< 999.9$ basis points ($< 9999$), immediately triggering a `CRITICAL_PURITY_FAILURE` alert.
- [ ] Constructs a deterministic SHA-256 Merkle tree of 10,000 vault bars in $< 500\text{ms}$ and verifies arbitrary inclusion proofs in $< 1\text{ms}$.
- [ ] Automatically dispatches `compliance.vault.mint_freeze_command.v1` and triggers on-chain mint pause on Hyperledger Besu when a critical discrepancy is flagged.
- [ ] Stores 100% of PDF assay certificates and scale calibration files in S3 with WORM retention locks and SHA-256 integrity digests.
- [ ] Code coverage for unit, integration, and mock physical vault hardware tests meets or exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 213 (Custodian Depository Integration), Prompt 303 (Proof of Reserve Contracts), Prompt 305 (Transfer Compliance Hooks).
- **Parallel Tasks:** Prompt 214 (Foreign Investor Funding & FX Service), Prompt 319 (Institutional Custody Bridge), Prompt 714 (Cross-Chain AML Screener).
- **Downstream Dependents:** Prompt 216 (Regulatory Reporting Service), Prompt 233 (Continuous 24x7 Regulatory Reporting), Prompt 604 / 605 (Admin & Compliance Review Consoles).
