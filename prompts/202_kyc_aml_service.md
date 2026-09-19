# 202 - KYC & Sanctions Screening Service (FastAPI / Celery)

## Purpose
The KYC & Sanctions Screening Service orchestrates the regulatory onboarding pipeline for both domestic Indian investors and international non-resident/foreign investors entering via the GIFT City IFSCA gateway. In compliance with SEBI Master Circulars on AML/CFT, RBI KYC Directions, and Prevention of Money Laundering Act (PMLA) rules, this service validates identity documents (Aadhaar offline XML / DigiLocker, PAN-Aadhaar linkage via NSDL/Income Tax Department, CKYCR central registry lookup), conducts passive/active liveness detection, and runs automated real-time screening against global sanctions and Politically Exposed Persons (PEP) watchlists.

This service is the regulatory trust anchor of Growww: an investor cannot deposit INR or trade tokenized fractional equities until this service verifies their legal identity and authorizes their cryptographic ledger address on the permissioned blockchain.

## What You Are Building
A high-throughput Python FastAPI application paired with Celery distributed task workers (`services/kyc-service`). Deliverables include:
- REST endpoints for document upload (encrypted multipart/form-data), DigiLocker OAuth2 redirect handling, and real-time liveness session validation.
- Asynchronous Celery worker pipelines for Optical Character Recognition (OCR), facial embedding extraction, and document anti-tampering verification.
- Sanctions & PEP screening engine executing fuzzy matching (Levenshtein / Jaro-Winkler) against UN Security Council, OFAC, EU, and Indian domestic MHA watchlists.
- CKYCR (Central KYC Records Registry) client adapter for automated retrieval of existing KYC identifiers.
- Encrypted document storage interface with AWS S3 / MinIO utilizing envelope encryption.
- Kafka publisher streaming KYC status changes (`kyc.submitted.v1`, `kyc.approved.v1`, `kyc.rejected.v1`, `kyc.sanctions_flagged.v1`).

## Scope Boundaries
- **In Scope:**
 - Domestic KYC: PAN validation, Aadhaar Paperless Offline e-KYC (XML with share code), DigiLocker integration.
 - Foreign / NRI KYC: Passport OCR, proof of address, FATCA / CRS declaration capture.
 - Biometric face-match scoring (selfie vs ID photo) and passive liveness challenge evaluation.
 - Automated screening against Consolidated Sanctions Lists and PEP databases.
 - Risk tier calculation (Low, Medium, High risk categorization).
- **Out of Scope / Handled Elsewhere:**
 - User profile and authentication session management (Prompt 201).
 - On-chain smart contract relayer submitting whitelist transactions (Prompt 305).
 - Admin/compliance officer manual document review dashboard UI (Prompt 604).
 - Post-trade transaction surveillance and structuring detection (Prompt 704).

## Technology to Use
- **Primary Language & Framework:** Python 3.12, FastAPI (0.111+), Celery (5.4+) with Redis message broker. Python is selected due to its unparalleled ecosystem for computer vision (OpenCV, ONNX Runtime), high-performance text fuzzy matching (`rapidfuzz`), and rich cryptographic validation toolkits.
- **Computer Vision & ML:** `onnxruntime` executing lightweight FaceNet / MobileFaceNet models for 512-dimensional facial embedding generation; `opencv-python-headless` for image pre-processing and blur/glare detection.
- **Database & Storage:** PostgreSQL 16+ for structured KYC records; AWS S3 or MinIO with server-side KMS encryption for identity documents.
- **External Integration:** `httpx` (async HTTP/2) for government API gateways (NSDL, UIDAI, CERSAI CKYC); `cryptography` library for Aadhaar XML digital signature verification.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `kyc_applications`, `kyc_documents`, `sanctions_screening_hits`, `kyc_audit_trail`.
- **Redis 7.2:** Celery task queue, broker, result backend, and transient liveness session state store.
- **S3 / MinIO Object Storage:** Encrypted bucket `growww-kyc-vault-prod` with strict bucket policies and immutable object locking (WORM).
- **Apache Kafka:** Publishes to topic `kyc.events.v1` partitioned by `user_id`.
- **User Service (Prompt 201):** Updates user status to `ACTIVE` upon KYC approval.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Zero On-Chain PII Guarantee:** Under no circumstances is any investor name, PAN, Aadhaar number, passport number, or document image stored on or transmitted to the Hyperledger Besu blockchain.
- **On-Chain Whitelist Trigger:** Upon KYC/AML approval, this service publishes a `kyc.approved.v1` event containing the user's pseudonymized ledger address (`0x...`), jurisdiction flag (`DOMESTIC_RETAIL`, `GIFT_CITY_ACCREDITED`), and an expiry timestamp.
- **Compliance Registry Interaction:** The compliance relayer service (Prompt 305) listens to this event and calls `ComplianceRegistry.setUserComplianceStatus(address, status, jurisdictionExpiry)` on Hyperledger Besu, enabling the address to receive and transfer ERC-3643 compliant digital security tokens.
- **Consensus Compatibility:** Ledger updates are finalized via QBFT consensus with 2-second block finality.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Create `services/kyc-service` structure with FastAPI, Celery worker definitions, and configuration management.
2. **Define Database Schemas & Migrations:** Implement SQLAlchemy 2.0 models and Alembic migrations for `kyc_applications`, `kyc_documents`, and `sanctions_screening_hits`.
3. **Configure Encrypted Object Storage:** Implement S3/MinIO client with AES-256 client-side envelope encryption using AWS KMS / HashiCorp Vault.
4. **Implement PAN Verification Client:** Build client adapter interfacing with NSDL/Protean and Income Tax Department APIs to verify PAN validity and PAN-Aadhaar seeding status.
5. **Implement Aadhaar Offline XML & DigiLocker Verifier:** Build XML parser, signature validator using UIDAI public certificates, and 4-digit share code decryption.
6. **Implement Document Masking Utility:** Implement automated image processing to permanently mask the first 8 digits of Aadhaar cards before persisting images to object storage.
7. **Implement Facial Liveness & Match Pipeline:** Integrate ONNX Runtime with MobileFaceNet to compute cosine similarity between selfie and ID photo (threshold $\ge 0.85$).
8. **Build Sanctions & PEP Screening Engine:** Build automated task that downloads consolidated sanctions feeds (UNSC, OFAC SDN, EU, MHA) and runs fuzzy search using `rapidfuzz` token-sort ratio with threshold $\ge 88$.
9. **Implement CKYCR Integration:** Build integration with CERSAI CKYC registry to auto-fetch existing KYC records for faster domestic onboarding.
10. **Implement KYC State Machine:** Enforce strict state transitions (`DRAFT` $\rightarrow$ `DOCUMENTS_UPLOADED` $\rightarrow$ `PROCESSING` $\rightarrow$ `APPROVED` / `REJECTED` / `MANUAL_REVIEW_REQUIRED`).
11. **Implement Kafka Event Publisher:** Emit structured Avro/JSON events to `kyc.events.v1` upon state transitions.
12. **Expose gRPC Service for Internal Verification:** Implement `KYCService` gRPC server to allow Order Service and Wallet Service to verify real-time KYC compliance status.
13. **Configure Prometheus Metrics & Tracing:** Instrument Celery worker latency, OCR error rates, sanctions hit counts, and facial match confidence distributions.
14. **Write End-to-End Test Suite:** Create integration test suite mocking government gateways, testing valid KYC approvals, rejected low-liveness selfies, and flagged sanctions names.

## Interfaces / Contracts

### Protobuf Definition (`kyc_service.proto`)
```protobuf
syntax = "proto3";

package growww.kyc.v1;

option go_package = "growww/kyc/v1;kycv1";

service KYCService {
  rpc GetKYCStatus (GetKYCStatusRequest) returns (GetKYCStatusResponse);
  rpc InitiateKYC (InitiateKYCRequest) returns (InitiateKYCResponse);
  rpc VerifyComplianceStatus (VerifyComplianceStatusRequest) returns (VerifyComplianceStatusResponse);
}

message GetKYCStatusRequest {
  string user_id = 1;
}

message GetKYCStatusResponse {
  string kyc_id = 1;
  string user_id = 2;
  string status = 3; // PENDING, PROCESSING, APPROVED, REJECTED, MANUAL_REVIEW
  string risk_tier = 4; // LOW, MEDIUM, HIGH
  string rejection_reason = 5;
  int64 verified_at_unix = 6;
  int64 expiry_at_unix = 7;
}

message InitiateKYCRequest {
  string user_id = 1;
  string jurisdiction = 2; // DOMESTIC_RESIDENT, GIFT_CITY_NRI
}

message InitiateKYCResponse {
  string kyc_id = 1;
  string digilocker_url = 2;
  string upload_session_token = 3;
}

message VerifyComplianceStatusRequest {
  string user_id = 1;
  string ledger_address = 2;
}

message VerifyComplianceStatusResponse {
  bool is_compliant = 1;
  bool is_sanctioned = 2;
  bool is_pep = 3;
  string jurisdiction = 4;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE kyc_status_enum AS ENUM ('DRAFT', 'DOCUMENTS_UPLOADED', 'PROCESSING', 'APPROVED', 'REJECTED', 'MANUAL_REVIEW');
CREATE TYPE kyc_risk_tier_enum AS ENUM ('LOW', 'MEDIUM', 'HIGH');

CREATE TABLE kyc_applications (
    kyc_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    jurisdiction VARCHAR(50) NOT NULL,
    status kyc_status_enum NOT NULL DEFAULT 'DRAFT',
    risk_tier kyc_risk_tier_enum NOT NULL DEFAULT 'LOW',
    pan_hash VARCHAR(64), -- SHA-256 of PAN for duplicate prevention
    aadhaar_reference_id VARCHAR(100),
    face_match_score NUMERIC(5, 4),
    is_pep BOOLEAN NOT NULL DEFAULT FALSE,
    is_sanctioned BOOLEAN NOT NULL DEFAULT FALSE,
    rejection_reason TEXT,
    verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE kyc_documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_id UUID NOT NULL REFERENCES kyc_applications(kyc_id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- PAN_CARD, AADHAAR_XML, PASSPORT, SELFIE
    storage_s3_key VARCHAR(512) NOT NULL,
    kms_key_id VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size_bytes INTEGER NOT NULL,
    is_masked BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sanctions_screening_hits (
    hit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_id UUID NOT NULL REFERENCES kyc_applications(kyc_id) ON DELETE CASCADE,
    watchlist_source VARCHAR(100) NOT NULL, -- UNSC, OFAC_SDN, EU_FSF, MHA_INDIA
    matched_name VARCHAR(255) NOT NULL,
    match_confidence NUMERIC(5, 4) NOT NULL,
    resolution_status VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, FALSE_POSITIVE, CONFIRMED
    reviewed_by UUID,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_status ON kyc_applications(status);
CREATE INDEX idx_kyc_pan_hash ON kyc_applications(pan_hash);
```

## Security & Compliance Notes
- **Aadhaar Masking Mandate:** UIDAI regulations strictly prohibit storing unmasked 12-digit Aadhaar numbers. All uploaded document images must undergo automated redaction of the first 8 digits before disk persistence.
- **Encryption at Rest:** All identity artifacts in S3/MinIO are encrypted using individual envelope encryption keys rotated via AWS KMS / HashiCorp Vault.
- **PMLA & Record Retention:** Under Section 12 of PMLA, all KYC verification logs, audit records, and raw responses must be immutably preserved for a minimum of 5 years following account closure.
- **Zero On-Chain PII:** The blockchain ledger only receives a boolean compliance attestation for the user's public address.

## Acceptance Criteria
- [ ] KYC application pipeline processes Aadhaar XML / DigiLocker documents and validates digital signatures against UIDAI root certificates.
- [ ] PAN-Aadhaar linkage check verifies full match with Income Tax Department records.
- [ ] Facial verification model computes cosine similarity and rejects selfies with match score $< 0.85$ or failed liveness.
- [ ] Sanctions screening engine flags exact and fuzzy matches ($\ge 88\%$ score) across OFAC, UNSC, and MHA lists.
- [ ] Approved applications trigger `kyc.approved.v1` Kafka event to enable on-chain whitelisting.
- [ ] Aadhaar numbers in images are verified to be 100% masked before S3 persistence.
- [ ] Celery task workers handle asynchronous document processing with average completion time $< 5$ seconds.

## Suggested Order / Dependencies
- **Prerequisites:** 004 (Domestic KYC Policy), 005 (Foreign KYC Policy), 105 (Auth Architecture), 201 (User Service), 703 (Sanctions Screening Policy).
- **Parallel Tasks:** 203 (Wallet Service), 211 (Notification Service).
- **Downstream Blockers:** 204 (Order Service), 305 (Compliance Registry Smart Contract), 604 (Admin KYC Review Dashboard).
