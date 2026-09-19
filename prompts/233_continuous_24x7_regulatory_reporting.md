# 233 - Continuous 24x7 Regulatory Statutory Reporting & Compliance Dispatcher

## Purpose
Establishes a continuous, automated statutory regulatory reporting and compliance attestation engine operating 24 hours a day, 7 days a week, 365 days a year. In a continuous 24/7 trading paradigm spanning Indian domestic markets and the GIFT City IFSCA international financial corridor, traditional batch End-of-Day (EOD) reporting is insufficient to prevent systemic risk and maintain real-time regulatory oversight.

This service transforms periodic batch filings into real-time, event-driven regulatory data pipelines. It continuously generates, validates, cryptographically seals, and electronically submits compliance streams to three distinct financial regulators:
1. **SEBI (Securities and Exchange Board of India):** Real-time market surveillance telemetry, continuous trade log feeds, 24/7 margin utilization metrics, short-collection alerts, and instant investor grievance logs.
2. **RBI (Reserve Bank of India):** Continuous Liberalised Remittance Scheme (LRS) quota tracking ($<\$250,000/\text{year}$), continuous FETERS (Foreign Exchange Transactions Electronic Reporting System) records, Form A2 validation, and Nostro/Vostro nodal balance reconciliations.
3. **IFSCA (International Financial Services Centres Authority):** Real-time cross-border capital flow tracking, foreign investor exposure limits, capital adequacy ratio (CAR) monitoring, and continuous liquidity coverage disclosures.

## What You Are Building
A specialized, resilient Python/Go distributed reporting microservice (`services/continuous-regulatory-reporting`):
- **Real-Time Regulatory Event Stream Compiler:** Consumes high-velocity trade settlements, fiat cash flows, FX conversions, and risk events from Kafka topics, continuously compiling structured filing packages.
- **Multi-Taxonomy Schema & Protocol Validator:** Enforces strict structural, semantic, and mathematical validation against SEBI XBRL taxonomies, RBI FETERS/LRS schemas, and IFSCA XML/ISO 20022 formats.
- **Hardware-Secured Digital Signature (HSM) Sealer:** Digitally signs every regulatory payload using Class-3 Organization Digital Signatures hosted inside FIPS 140-2 Level 3 Hardware Security Modules.
- **Automated Multi-Protocol Regulatory Dispatcher:** Orchestrates resilient transmission via RESTful APIs, Secure SFTP, AS2 (Applicability Statement 2) encrypted channels, and dedicated leased-line regulatory gateways with automated retry and acknowledgment tracking.
- **Continuous WORM Compliance Archive:** Stores immutable copies of all transmitted reports, cryptographic proofs, and regulatory acknowledgment receipts in AWS S3 Object Lock (Compliance Mode) with an 8-year statutory retention lock.

## Scope Boundaries
- **In Scope:**
 - Continuous 24/7 generation of trade-by-trade and aggregated statutory filings for SEBI, RBI, and IFSCA.
 - Real-time LRS limit monitoring and RBI Form A2 reporting for international fiat funding.
 - XML, XBRL, CSV, and ISO 20022 message serialization and schema validation (XSD/DTD).
 - HSM-backed cryptographic sealing and Merkle tree state proof embedding.
 - Automated dispatch and receipt reconciliation with regulatory endpoints.
- **Out of Scope / Handled Elsewhere:**
 - Internal batch tax statements (ITR Schedule FA, AIS/TIS) for individual end-users (handled in Prompt 223).
 - Pre-trade AML sanctions screening (handled in Prompt 202).
 - Real-time manipulation pattern recognition (handled in Prompt 711).

## Technology to Use
- **Primary Service Framework:** Python 3.12+ with FastAPI and Celery/Temporal for workflow orchestration, paired with Go for high-throughput stream formatting.
- **Regulatory Schema & Document Tools:** `lxml` (for ultra-fast XSD validation), `Arelle` / `python-xbrl` for XBRL taxonomy processing, and `cryptography` / PKCS#11 libraries for HSM interaction.
- **Stream Ingestion & Event Storage:** Apache Kafka with Kafka Connect for real-time transactional event capture.
- **Transactional & Audit Store:** PostgreSQL 16 with TimescaleDB for filing metadata and submission acknowledgments.
- **Immutable Storage:** AWS S3 / MinIO configured with Object Lock (WORM - Write Once Read Many) in Compliance Mode.
- **Justification:** Python provides standard-setting libraries for complex financial XML and XBRL taxonomies, while Go provides high concurrency for streaming thousands of continuous regulatory telemetry payloads per minute without memory bloat.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
 - `services/trade-settlement-service` (Prompt 208): Continuous trade execution and DvP settlement stream.
 - `services/foreign-investor-funding-service` (Prompt 214): Real-time USD/INR FX conversions and cross-border bank transfers.
 - `services/risk-engine` (Prompt 206): Real-time margin utilization, peak margins, and circuit breaker activations.
 - `services/reconciliation-service` (Prompt 215): Continuous 3-way reconciliation balances (Demat vs Tokens vs Fiat).
- **Downstream Microservices:**
 - `services/admin-back-office` (Prompt 217 & 607): Regulatory reporting dashboard for compliance officers.
 - `services/audit-log-service` (Prompt 218): Immutable recording of all filing transmissions.
- **External Regulators & Gateways:**
 - SEBI Electronic Reporting Gateway (SERG / SFTP).
 - RBI XBRL & FETERS Reporting Portal.
 - IFSCA Regulatory Portal & FINnet 2.0 Gateway.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **Cryptographic Merkle Proof Attestation:** For every continuous statutory filing package, the reporting service extracts the latest block state from `SettlementDvP.sol` and `ProofOfReserveRegistry.sol`. It computes a SHA-256 Merkle tree root of all trades included in the reporting batch and embeds the Besu block hash, transaction receipts, and validator signatures directly into the regulatory XML/XBRL payload.
- **Denial-Proof Regulatory Verification:** Regulators operating observer nodes on the Besu consortium can instantaneously verify that the off-chain filed trades match the on-chain cryptographic settlement state byte-for-byte in real time.
- **Zero PII on Ledger:** All personal identification details (PAN, Aadhaar Vault Reference, Passport Number) remain strictly within encrypted off-chain regulatory files; on-chain references consist entirely of cryptographic account hashes.

## Step-by-Step Build Instructions
1. Scaffold the `services/continuous-regulatory-reporting` service repository with `ingestor/`, `compilers/`, `validators/`, `signer/`, `dispatcher/`, and `archive/` packages.
2. Define Protobuf contracts in `proto/growww/regulatory_reporting/v1/continuous_reporting.proto` and generate gRPC client stubs.
3. Configure PostgreSQL tables for continuous filing metadata, regulatory endpoints, transmission logs, and acknowledgment receipts.
4. Implement the **SEBI Continuous Telemetry Compiler**:
 - Stream real-time trade logs (`isin`, `timestamp_ns`, `price`, `quantity`, `buyer_category`, `seller_category`).
 - Compile hourly rolling margin collection files and peak-margin violation reports.
5. Implement the **RBI 24/7 Cross-Border & LRS Compiler**:
 - Monitor real-time cross-border remittances; track individual investor cumulative 12-month LRS utilization against the $\$250,000$ statutory ceiling.
 - Generate automated FETERS XML and Form A2 returns upon each cross-border funding or repatriation event.
6. Implement the **IFSCA 24/7 Prudential Reporting Compiler**:
 - Compute real-time Capital Adequacy Ratio (CAR), Net Capital Balance, and client asset segregation metrics.
 - Format ISO 20022 compliant messages for foreign currency settlement records.
7. Implement the **XSD & XBRL Strict Schema Validation Engine**:
 - Validate 100% of generated XML/XBRL payloads against official regulator schemas before signing.
 - If schema validation fails, immediately raise high-priority alert to on-call compliance engineers.
8. Implement the **FIPS 140-2 Level 3 HSM Digital Signing Engine**:
 - Interface with AWS CloudHSM / SafeNet Luna HSM via PKCS#11 protocol.
 - Apply digital signatures (PAdES for PDF, XMLDSig for XML/XBRL) using registered Class-3 digital certificates.
9. Implement the **Automated Multi-Protocol Dispatcher**:
 - Build resilient dispatch pipelines supporting HTTPS REST, SFTP with mutual TLS, and AS2 with MDN (Message Disposition Notification) acknowledgment receipts.
 - Implement exponential backoff, circuit-breaking, and dead-letter queues for unreachable regulatory servers.
10. Integrate **Hyperledger Besu On-Chain Settlement Proofs**:
 - Query smart contract settlement state receipts and embed on-chain Merkle roots into regulatory headers.
11. Build the **WORM S3 Compliance Archival System**:
 - Automatically upload all raw files, signed payloads, and regulatory receipts to S3 Object Lock buckets with an 8-year compliance hold.
12. Establish automated monitoring, end-to-end filing verification tests, and schema compatibility regression test suites.

## Interfaces / Contracts

### Protobuf Definition (`continuous_reporting.proto`)
```protobuf
syntax = "proto3";

package growww.regulatory_reporting.v1;

option go_package = "github.com/growww/services/continuous-regulatory-reporting/gen/v1;reportingv1";

service ContinuousRegulatoryReportingService {
  rpc TriggerContinuousFiling (TriggerFilingRequest) returns (TriggerFilingResponse);
  rpc GetFilingStatus (FilingStatusRequest) returns (FilingStatusResponse);
  rpc VerifyRegulatoryPayload (VerifyPayloadRequest) returns (VerifyPayloadResponse);
  rpc GetLrsQuotaStatus (LrsQuotaRequest) returns (LrsQuotaResponse);
}

enum RegulatorTarget {
  REGULATOR_UNSPECIFIED = 0;
  REGULATOR_SEBI = 1;
  REGULATOR_RBI = 2;
  REGULATOR_IFSCA = 3;
}

enum FilingFrequency {
  FREQ_STREAMING_REALTIME = 0;
  FREQ_HOURLY = 1;
  FREQ_DAILY_ROLLING = 2;
  FREQ_MONTHLY_STATUTORY = 3;
}

message TriggerFilingRequest {
  RegulatorTarget regulator = 1;
  string report_code = 2; // e.g., SEBI_TRADE_FEED_V1, RBI_FETERS_24X7, IFSCA_CAR_HOURLY
  int64 window_start_utc = 3;
  int64 window_end_utc = 4;
  bool is_automated = 5;
}

message TriggerFilingResponse {
  string filing_id = 1;
  string status = 2; // COMPILING / SIGNED / DISPATCHED / ACKNOWLEDGED
  int64 records_included = 3;
  string tracking_number = 4;
}

message FilingStatusRequest {
  string filing_id = 1;
}

message FilingStatusResponse {
  string filing_id = 1;
  RegulatorTarget regulator = 2;
  string report_code = 3;
  string status = 4;
  string payload_sha256 = 5;
  string digital_signature_thumbprint = 6;
  string blockchain_merkle_root = 7;
  string regulatory_ack_receipt = 8;
  int64 submitted_at_utc = 9;
  int64 acknowledged_at_utc = 10;
  string worm_storage_uri = 11;
}

message VerifyPayloadRequest {
  string raw_payload = 1;
  string schema_type = 2; // XML / XBRL / ISO20022
}

message VerifyPayloadResponse {
  bool is_valid = 1;
  repeated string validation_errors = 2;
}

message LrsQuotaRequest {
  string investor_pan_hash = 1;
  int32 fiscal_year = 2;
}

message LrsQuotaResponse {
  string investor_pan_hash = 1;
  double total_utilized_usd = 2;
  double remaining_quota_usd = 3;
  double statutory_limit_usd = 4; // 250,000.00
  bool is_quota_exceeded = 5;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE continuous_regulatory_filings (
    filing_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regulator VARCHAR(16) NOT NULL CHECK (regulator IN ('SEBI', 'RBI', 'IFSCA')),
    report_code VARCHAR(64) NOT NULL,
    filing_frequency VARCHAR(32) NOT NULL,
    window_start_utc TIMESTAMPTZ NOT NULL,
    window_end_utc TIMESTAMPTZ NOT NULL,
    total_records_count INT NOT NULL,
    file_format VARCHAR(16) NOT NULL CHECK (file_format IN ('XML', 'XBRL', 'ISO20022', 'CSV', 'JSON_LD')),
    payload_sha256 VARCHAR(64) NOT NULL,
    hsm_signature_thumbprint VARCHAR(128) NOT NULL,
    blockchain_merkle_root VARCHAR(64) NOT NULL,
    blockchain_block_number BIGINT NOT NULL,
    worm_s3_uri VARCHAR(256) NOT NULL,
    transmission_protocol VARCHAR(16) NOT NULL CHECK (transmission_protocol IN ('HTTPS_REST', 'SFTP_MTLS', 'AS2')),
    transmission_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    regulatory_ack_receipt VARCHAR(256),
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transmitted_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ
);

CREATE TABLE rbi_lrs_utilization_tracking (
    tracking_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_pan_hash VARCHAR(64) NOT NULL,
    fiscal_year INT NOT NULL,
    transaction_id UUID NOT NULL,
    remittance_direction VARCHAR(8) NOT NULL CHECK (remittance_direction IN ('OUTFLOW', 'INFLOW')),
    amount_inr NUMERIC(18, 2) NOT NULL,
    fx_rate NUMERIC(12, 6) NOT NULL,
    amount_usd NUMERIC(18, 2) NOT NULL,
    cumulative_utilized_usd NUMERIC(18, 2) NOT NULL,
    form_a2_ref_number VARCHAR(64) NOT NULL,
    feters_purpose_code VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reg_filings_regulator ON continuous_regulatory_filings (regulator, created_at DESC);
CREATE INDEX idx_lrs_pan_fy ON rbi_lrs_utilization_tracking (investor_pan_hash, fiscal_year);
```

## Security & Compliance Notes
- **Continuous Compliance & Zero Tipping-Off:** All STR and regulatory intelligence files are generated in an isolated zero-trust subnet; access is restricted strictly to designated compliance officers.
- **FIPS 140-2 Level 3 HSM Key Security:** Organization signing keys are non-exportable and stored within hardened HSM partitions. Every signing request requires mTLS authentication and tokenized authorization.
- **WORM Retention Mandate:** In compliance with PMLA (Section 12) and SEBI intermediary regulations, all filings, schemas, logs, and acknowledgments are locked under WORM compliance for a minimum of 8 years.
- **DPDP Act & Anonymization:** Data sent to international regulatory observers is strictly stripped of domestic PII, using verifiable one-way cryptographic tokens.

## Acceptance Criteria
- [ ] Automatically compiles and validates 100% of SEBI, RBI, and IFSCA continuous filings against their respective XSD/XBRL taxonomies.
- [ ] Successfully signs all payloads using HSM-managed Class-3 certificates with valid cryptographic digest verification.
- [ ] Dispatches real-time filings over HTTPS/SFTP/AS2 with automated acknowledgment capture and exponential retry handling.
- [ ] Enforces real-time RBI LRS limit tracking, rejecting any cross-border funding request exceeding $\$250,000$ cumulative annual quota.
- [ ] Embeds verifiable Hyperledger Besu Merkle tree roots and transaction receipts in 100% of statutory trade reports.
- [ ] Automatically persists all transmitted files to S3 Object Lock storage with immutable 8-year compliance hold.
- [ ] Unit, schema validation, and end-to-end integration test coverage exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 003 (Regulatory Pathway), Prompt 208 (Trade Settlement), Prompt 214 (Foreign Investor Funding), Prompt 215 (Reconciliation Service), Prompt 216 (Regulatory Reporting Service).
- **Subsequent / Parallel Tasks:** Prompt 711 (Market Surveillance Engine), Prompt 712 (Insider Trading Analytics), Prompt 713 (Continuous Circuit Breaker Coordinator).
