# 216 - Regulatory Reporting & Compliance Service (SEBI / RBI / IFSCA)

## Purpose
The Regulatory Reporting & Compliance Service automates the compilation, cryptographic attestation, validation, formatting, and secure electronic submission of statutory filings required by Indian and international financial regulators. Operating within the dual-entity framework of Growww (domestic regulated clearing entity under SEBI/RBI and international gateway entity under GIFT City/IFSCA), this service eliminates manual reporting errors, guarantees regulatory transparency, and ensures strict adherence to reporting timelines.

The service generates official reporting artifacts including SEBI periodic risk and holding disclosures, FIU-IND (Financial Intelligence Unit - India) Suspicious Transaction Reports (STRs) and Cash Transaction Reports (CTRs) under the Prevention of Money Laundering Act (PMLA), RBI foreign remittance returns (FETERS, Form A2), and IFSCA capital market compliance returns.

## What You Are Building
A specialized Python/FastAPI microservice (`services/reporting-service`) providing:
- **Statutory Report Generation Engine:** Renders XBRL, XML, ISO 20022, CSV, and digitally signed PDF filings.
- **FIU-IND FINnet 2.0 Gateway:** Generates and validates XML reports matching FIU-IND schema specifications.
- **Cryptographic Attestation Module:** Embeds Hyperledger Besu transaction receipts and Merkle proof hashes into regulatory submission packages.
- **Submission Scheduler & Dispatcher:** Automates secure dispatch via SFTP/API with digital signatures generated via HSM.
- **Artifacts Delivered:**
 - `services/reporting-service/app/main.py` - FastAPI service entry point.
 - `services/reporting-service/app/generators/sebi_generator.py` - SEBI periodic filing renderers.
 - `services/reporting-service/app/generators/fiu_generator.py` - FIU-IND FINnet 2.0 XML generator.
 - `services/reporting-service/app/generators/ifsca_generator.py` - IFSCA statutory return generator.
 - `proto/growww/reporting/v1/reporting.proto` - gRPC contracts for compliance reporting.

## Scope Boundaries
- **In Scope:**
 - Aggregating transactional, KYC, and holding data from internal databases for report compilation.
 - Generating XBRL/XML/PDF reports adhering to SEBI, RBI, FIU-IND, and IFSCA schemas.
 - Digitally signing filings using Class-3 digital signatures stored in HSM.
 - Maintaining an immutable WORM (Write Once Read Many) compliant archive of all generated reports.
- **Out of Scope / Handled Elsewhere:**
 - Real-time AML transaction monitoring and rule evaluation (handled by Prompt 202).
 - Real-time pre-trade risk checks (handled by Prompt 206).
 - Public proof-of-reserve portal rendering (handled by Prompt 007 / Category 6).

## Technology to Use
- **Primary Language & Framework:** Python 3.12+ with FastAPI, Celery (with Redis broker), and Jinja2 / `lxml` / `weasyprint`.
- **Justification:** Python offers the most mature ecosystem for parsing complex regulatory schemas (XBRL, XML DTD/XSD validation), template rendering, cryptographic document stamping, and PDF generation, coupled with Celery for managing long-running batch extraction jobs.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for storing report definitions, schedules, and filing audit records.
 - `lxml` for strict XSD schema validation of FIU-IND and SEBI XML payloads.
 - `python-xbrl` / `Arelle` for XBRL taxonomy handling.
 - `weasyprint` for generating high-fidelity PDF regulatory statements.
 - AWS S3 Object Lock / MinIO WORM for immutable regulatory document storage.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Core analytical read-replicas and reporting metadata tables.
- **Hyperledger Besu (QBFT):** Queries on-chain block receipts, Merkle roots, and transaction hashes for inclusion in audit reports.
- **MinIO / AWS S3 (WORM Storage):** Immutable storage for generated regulatory packages with 8-year compliance hold.
- **Kafka Topics:**
 - Subscribes: `reconciliation.eod.certified`, `kyc.aml.flagged`, `trade.settled`.
 - Publishes: `regulatory.report.generated`, `regulatory.filing.submitted`.
- **External Regulators:** FIU-IND FINnet Portal, SEBI Portal SFTP, RBI XBRL Gateway, IFSCA Compliance Portal.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **Cryptographic Audit Embedding:** The reporting service extracts verifiable cryptographic proofs from `ProofOfReserveRegistry.sol` and `SettlementDvP.sol`.
- **Report Attachment:** Embeds on-chain block numbers, transaction hashes, and Merkle tree roots into statutory filings submitted to SEBI and IFSCA as undeniable mathematical proof of real-time asset backing and trade settlement.
- **Zero PII on Ledger:** Reconciles on-chain anonymous transaction hashes with off-chain compliance databases to populate regulatory-mandated investor PAN / Passport details strictly within off-chain encrypted regulatory filings.

## Step-by-Step Build Instructions
1. Scaffold Python FastAPI service under `services/reporting-service` with Celery worker integration.
2. Define Protobuf definitions in `proto/growww/reporting/v1/reporting.proto` and generate client stubs.
3. Configure PostgreSQL schema migrations for `regulatory_reports`, `filing_schedules`, and `compliance_submissions`.
4. Implement FIU-IND XML generator validating against official FINnet 2.0 XSD schemas for STR and CTR submissions.
5. Implement SEBI monthly and quarterly holding disclosure generator with automated pool account balance rollups.
6. Build the IFSCA statutory filing module for cross-border investor funding and foreign currency exposure reports.
7. Implement digital signature engine interfacing with HSM to sign PDF and XML files with Class-3 organizational certificates.
8. Integrate MinIO / S3 Object Lock client to archive all generated reports with immutable retention policies (PMLA 8-year rule).
9. Build the regulatory dispatcher module supporting automated SFTP upload, AS2 messaging, and secure HTTPS API submission.
10. Implement Celery periodic tasks for scheduled EOD, weekly, monthly, and quarterly filing runs.
11. Add Maker-Checker compliance officer review APIs allowing authorized officers to inspect, approve, or annotate filings.
12. Build unit and integration tests verifying XML schema validity, PDF generation accuracy, and signature verification.

## Interfaces / Contracts

### Protobuf Definition (`reporting.proto`)
```protobuf
syntax = "proto3";

package growww.reporting.v1;

option go_package = "github.com/growww/services/reporting-service/gen/v1;reportingv1";

service RegulatoryReportingService {
  rpc GenerateReport (GenerateReportRequest) returns (GenerateReportResponse);
  rpc GetReportStatus (ReportStatusRequest) returns (ReportStatusResponse);
  rpc ApproveFiling (ApproveFilingRequest) returns (ApproveFilingResponse);
  rpc ListPendingFilings (ListPendingFilingsRequest) returns (ListPendingFilingsResponse);
}

message GenerateReportRequest {
  string regulator = 1; // SEBI / RBI / FIU_IND / IFSCA
  string report_type = 2; // SEBI_HOLDING_MONTHLY / FIU_STR / RBI_FETERS / IFSCA_QTR
  string period_start = 3; // YYYY-MM-DD
  string period_end = 4;   // YYYY-MM-DD
  bool dry_run = 5;
}

message GenerateReportResponse {
  string report_id = 1;
  string status = 2; // GENERATING / READY_FOR_REVIEW
  string file_format = 3; // XML / XBRL / PDF
  int64 generated_at = 4;
}

message ReportStatusRequest {
  string report_id = 1;
}

message ReportStatusResponse {
  string report_id = 1;
  string regulator = 2;
  string report_type = 3;
  string status = 4; // PENDING_APPROVAL / APPROVED / SUBMITTED / REJECTED
  string download_url = 5;
  string sha256_checksum = 6;
  string blockchain_merkle_root = 7;
  int64 created_at = 8;
  int64 submitted_at = 9;
}

message ApproveFilingRequest {
  string report_id = 1;
  string compliance_officer_id = 2;
  string digital_signature_token = 3;
}

message ApproveFilingResponse {
  string report_id = 1;
  bool submitted = 2;
  string acknowledgment_receipt = 3;
  int64 submission_timestamp = 4;
}

message ListPendingFilingsRequest {
  string regulator = 1;
}

message ListPendingFilingsResponse {
  repeated ReportSummary reports = 1;
}

message ReportSummary {
  string report_id = 1;
  string report_type = 2;
  string period = 3;
  string created_at = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE regulatory_reports (
    report_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regulator VARCHAR(32) NOT NULL CHECK (regulator IN ('SEBI', 'RBI', 'FIU_IND', 'IFSCA')),
    report_type VARCHAR(64) NOT NULL,
    reporting_period_start DATE NOT NULL,
    reporting_period_end DATE NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'GENERATED',
    file_path VARCHAR(256) NOT NULL,
    file_format VARCHAR(16) NOT NULL,
    sha256_hash VARCHAR(64) NOT NULL,
    blockchain_merkle_root VARCHAR(64),
    generated_by VARCHAR(64) NOT NULL,
    approved_by VARCHAR(64),
    submission_ack_ref VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at TIMESTAMPTZ
);

CREATE TABLE compliance_audit_filings (
    filing_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES regulatory_reports(report_id),
    action VARCHAR(32) NOT NULL,
    performed_by VARCHAR(64) NOT NULL,
    signature_fingerprint VARCHAR(128),
    details_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **PMLA (Prevention of Money Laundering Act) Compliance:** FIU-IND STR/CTR filings are treated with highest classification confidentiality; unauthorized disclosure (tipping off) is strictly prevented via RBAC.
- **WORM Document Retention:** All regulatory filings are stored in S3 Object Lock compliance mode with an 8-year retention lock, making deletion or alteration mathematically impossible.
- **HSM-Backed Digital Signatures:** Submissions are sealed using Class-3 digital signatures securely held inside FIPS 140-2 Level 3 HSMs.
- **Audit Logging:** Every view, generation, and approval of regulatory filings is recorded in the immutable audit log service.

## Acceptance Criteria
- [ ] Python FastAPI service and Celery workers run cleanly with automated health checks.
- [ ] FIU-IND STR/CTR XML generator passes 100% of FINnet 2.0 schema validation test vectors.
- [ ] SEBI holding disclosure reports accurately aggregate pool accounts and match physical Demat snapshots.
- [ ] Digital signature engine signs documents using HSM keys and outputs verifiable PKCS#7 / PAdES signatures.
- [ ] Generated reports are successfully uploaded to WORM S3 buckets with immutable object locks.
- [ ] Maker-Checker approval workflow prevents unauthorized dispatch of regulatory filings.
- [ ] Test coverage exceeds >=85% across all generators and dispatch modules.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 003 (Regulatory Pathway), Prompt 004 (Domestic KYC Policy), Prompt 202 (KYC/AML Service), Prompt 215 (Reconciliation Service).
- **Subsequent / Parallel Tasks:** Prompt 217 (Admin Service), Prompt 218 (Audit Log Service).
