# 267 - SEBI SCORES 2.0 & Regulatory Grievance Gateway (Python / FastAPI)

## Purpose
Statutory compliance within Indian capital markets and fintech ecosystems mandates strict adherence to formal dispute resolution timelines set by financial authorities. Under the Securities and Exchange Board of India (SEBI) modernized framework SEBI Complaint Redress System (SCORES 2.0), governed by circular `SEBI/HO/OIAE/OIAE_IAD-1/P/CIR/2023/145` (dated August 11, 2023) and master circular `SEBI/HO/OIAE/OIAE_IAD-1/P/CIR/2023/131` (Securities Market Approach for Resolution Through ODR - SMART ODR), intermediaries must resolve investor grievances and file a comprehensive Action Taken Report (ATR) within a non-negotiable 21 calendar days from intake. Parallel regulatory mandates, including the Reserve Bank of India (RBI) Integrated Ombudsman Scheme, 2021 (administered via the RBI Complaint Management System - CMS), impose rigorous 30-calendar-day adjudication windows for payments, pre-paid instruments (PPI), and credit ledger disputes.

Manual grievance redressal workflows in fast-moving electronic brokerage environments result in catastrophic compliance vulnerabilities: untracked regulatory notices, fragmented audit trails across disparate microservices, manual data aggregation delays, missed statutory deadlines, and automatic escalation to Market Infrastructure Institutions (MIIs) or SEBI regulatory enforcement panels under Section 15C of the SEBI Act, 1992.

The **SEBI SCORES 2.0 & Regulatory Grievance Gateway** (`services/grievance-gateway`) operates as an enterprise-grade regulatory orchestration middleware. It bridges external statutory regulatory portals (SEBI SCORES 2.0, RBI Ombudsman CMS, SMART ODR) with internal trading, settlement, banking, and compliance microservices. The gateway automates real-time docket ingestion, investor identification via Permanent Account Number (PAN) and Unique Client Code (UCC), cross-service cryptographic audit trail assembly, proactive 21-day SLA countdown monitoring with multi-stage breach prevention (ADR-0046, RUNBOOK-34), Maker-Checker Action Taken Report (ATR) authoring, Class-3 Digital Signature Certificate (DSC) cryptographic signing, and immutable SHA-256 receipt anchoring on Hyperledger Besu.

## What You Are Building
A resilient, secure Python 3.11+ / FastAPI enterprise microservice (`services/grievance-gateway`) coupled with Celery distributed workers, Redis caching, and PostgreSQL 16 persistence. Core components include:

- **Regulatory Portal Ingestion Adapters:** Multi-protocol inbound connectors establishing secure, authenticated synchronization with external regulatory platforms:
  - *SEBI SCORES 2.0 REST Adapter:* Bidirectional integration utilizing Mutual TLS (mTLS) and HMAC-SHA256 authenticated webhooks and polling clients to ingest fresh dockets, investor claims, and regulatory clarifications.
  - *RBI Ombudsman CMS Ingestion Engine:* Automated connector polling and processing banking/wallet grievance dockets and statutory notices.
  - *SMART ODR Portal Bridge:* Event-driven webhook consumer for dispute notifications, conciliation filings, and arbitration hearing schedules from designated Indian MIIs.
- **Complaint Lifecycle Finite State Machine (FSM):** Deterministic, event-driven state engine managing grievances through discrete operational states: `INGESTED`, `MAPPED_TO_USER`, `DIAGNOSTIC_TRIAGE`, `INVESTIGATION_PENDING`, `AUDIT_DOSSIER_COLLECTED`, `ATR_DRAFTED_MAKER`, `ATR_APPROVED_CHECKER`, `ATR_SUBMITTED_TO_REGULATOR`, `FIRST_REVIEW_ESCALATED`, `ODR_CONCILIATION`, `RESOLVED_CLOSED`, and `REJECTED_INVALID`.
- **Automated Diagnostic Triage & Evidence Harvesting Pipeline:** Celery-driven background worker pipeline that extracts evidence packages upon ticket docketing:
  - Fetches KYC/AML profile and account status from User Service (Prompt 201).
  - Fetches order placement, execution reports, matching logs, and cancellation telemetry from Order Matching & Trade Settlement (Prompt 205 / 208).
  - Fetches fiat ledger balances, bank deposit UTRs, and payment gateway callbacks from Wallet Account Service (Prompt 203) and Payment Gateway Integration Service (Prompt 212).
  - Fetches immutable WORM event logs from Audit Log Service (Prompt 218).
- **21-Day Statutory SLA Countdown & Breach Prevention Daemon:** A high-precision monitoring scheduler backed by Redis Sorted Sets (`scores:sla:deadline:zset`) enforcing strict escalation intervals:
  - *Day 0:* Intake, automatic PAN/UCC correlation, diagnostic triage, investor acknowledgment.
  - *Day 3:* Automated evidence dossier compilation completed; docket assigned to Compliance Officer.
  - *Day 7:* Stage 1 alert to assigned investigator if ATR draft is not compiled.
  - *Day 14:* Stage 2 escalation to Principal Compliance Officer (PCO).
  - *Day 18 (P1 Critical Alert per RUNBOOK-34):* Senior Legal Lead triage, executive war-room escalation.
  - *Day 20:* Mandatory expedited ATR submission freeze.
  - *Day 21:* Hard statutory SLA breach boundary prevented before regulatory auto-escalation to Stock Exchange Designated Body or SEBI.
- **Action Taken Report (ATR) Compiler & DSC Signer:** Standardized regulatory artifact compiler assembling formal PDF/A-1b and XML resolution dossiers, integrating with Hardware Security Modules (HSM) via PKCS#11 or PyHanko to append Class-3 digital signatures.
- **Hyperledger Besu Cryptographic Receipt Relayer:** Web3 JSON-RPC client anchoring the SHA-256 digest of every regulatory docket, compiled evidence dossier, and submitted ATR receipt to the permissioned Besu audit ledger, providing mathematical non-repudiation and timestamp proof.
- **Regulatory Reporting & Export Subsystem:** Aggregation engine feeding periodic grievance statistics, turnaround times (TAT), and category-wise resolution counts to Regulatory Reporting Service (Prompt 216) for SEBI monthly annexures and public web disclosures.

## Scope Boundaries
- **In Scope:**
  - Ingestion and bi-directional status synchronization with SEBI SCORES 2.0 API, RBI Ombudsman CMS, and SMART ODR portals.
  - Identification and mapping of complainants against internal user records via PAN, Demat BOID, UCC, mobile number, and email.
  - Automated diagnostic triage for routine transaction inquiries (failed deposits, missing bank UTRs, execution confirmation slips).
  - Asynchronous cross-service evidence gathering from User Service (Prompt 201), Wallet Service (Prompt 203), Settlement Service (Prompt 208), and Audit Log Service (Prompt 218).
  - Dual-control Maker-Checker review workflow for Action Taken Report (ATR) authoring and sign-off.
  - Class-3 DSC digital signing of ATR packages.
  - Real-time SLA breach countdown monitoring and multi-channel alerting (Slack, PagerDuty, email).
  - Anchoring of submission hashes to Hyperledger Besu smart contracts.
  - Immutable archival of all complaint dossiers with an 8-year WORM retention policy under DPDP Act 2023 and PMLA mandates.
  - Export of consolidated grievance redressal metrics for statutory reporting.
- **Out of Scope / Handled Elsewhere:**
  - Direct end-user conversational chat interfaces or L1 ticketing helpdesk UI (handled by CRM / Customer Support Gateway).
  - General statutory XBRL/XML broker filings, holding disclosures, and FIU-IND STR/CTR reporting (handled in Prompt 216 Regulatory Reporting Service).
  - Primary KYC verification, Aadhaar e-Sign, and biometric onboarding (handled in Prompt 201 User Service and Prompt 202 KYC/AML Service).
  - Direct financial ledger adjustments, wallet crediting, or refund banking transfers (orchestrated by Prompt 203 Wallet Account Service and Prompt 212 Payment Gateway Service).
  - Secondary market trade matching, order cancellation, or position liquidation (handled in Prompt 205 Order Matching Engine and Prompt 208 Settlement Service).
  - Smart contract compilation and node deployment on Besu (handled in Prompt 301/302).

## Technology to Use
- **Core Runtime & Language:** **Python 3.11+** running on asynchronous ASGI architecture.
- **Web & API Framework:** **FastAPI 0.111+** with Pydantic v2 for data validation, dependency injection, and automatic OpenAPI 3.1 schema generation.
- **Asynchronous Task Queue & Scheduling:** **Celery 5.4+** with **Redis 7.2+ Cluster** as broker and result backend; **Celery Beat** for periodic SLA countdown monitoring, external portal polling, and scheduled report rollups.
- **Database & Object Relational Mapping:** **PostgreSQL 16+** using `asyncpg` and `SQLAlchemy 2.0 (async)` with JSONB support for unstructured regulatory payloads, strict row-level locking (`SELECT FOR UPDATE`), and table partitioning by intake month.
- **Distributed Caching & Concurrency Control:** **Redis 7.2+ Cluster** with `redis-py` using Redis Sorted Sets (ZSETs) for millisecond-precision SLA breach countdowns, distributed locking via Redlock (`aioredlock`) for docket status updates, and session caching.
- **Digital Signatures & Cryptography:** `pyHanko` and `cryptography` libraries for PKCS#7 / CMS detached signatures, PDF/A generation, and X.509 Class-3 certificate signing via HSM / PKCS#11 network interfaces.
- **Document & Evidence Rendering:** `WeasyPrint` / `ReportLab` for generating tamper-evident regulatory PDF/A dossiers; `lxml` for strict XSD schema validation of SCORES and RBI XML documents.
- **Object Storage & WORM Archival:** **MinIO / AWS S3** with Object Lock (Compliance Mode) enabling WORM (Write Once Read Many) storage for 8-year regulatory retention.
- **Blockchain Connectivity:** `web3.py` (v6+) connecting to Hyperledger Besu QBFT consortium nodes via JSON-RPC over Mutual TLS.
- **Event Streaming & Message Bus:** **Apache Kafka 3.7+** via `confluent-kafka` (librdkafka C-bindings) for asynchronous inter-service event publishing and consumption.
- **HTTP Client & Security:** `httpx` with HTTP/2 and mTLS client certificates, equipped with `tenacity` for exponential backoff and circuit-breaking.

## Backend / Infra Touchpoints
- **External Regulatory Portal Endpoints:**
  - *SEBI SCORES 2.0 API:* `https://scores.sebi.gov.in/api/v2` (Endpoints: `/complaints/inbound`, `/atr/submit`, `/status/query`, `/clarification/reply`) using mTLS 1.3 and static IP whitelisting.
  - *RBI Ombudsman CMS Gateway:* `https://cms.rbi.org.in/api/v1` (Endpoints: `/dockets/fetch`, `/response/upload`) via OAuth 2.0 mTLS.
  - *SMART ODR Portal:* `https://smartodr.in/api/v1` (Endpoints: `/disputes/webhook`, `/evidence/submit`, `/hearing/schedule`).
- **Internal Microservice Interfaces:**
  - **User Service (Prompt 201):** Synchronous gRPC / REST queries to resolve PAN, demat account numbers, registered communication contacts, and KYC audit trails.
  - **Audit Log Service (Prompt 218):** Ingests immutable system event records, matching engine sequence numbers, and administrator activity trails corresponding to contested transactions.
  - **Regulatory Reporting Service (Prompt 216):** Publishes monthly and quarterly aggregate grievance redressal data (SEBI Annexure B) and statutory disclosure summaries.
  - **Wallet Account Service (Prompt 203):** Fetches double-entry fiat ledger transactions, UTR numbers, account freeze states, and balance histories.
  - **Trade Settlement Service (Prompt 208):** Fetches official electronic contract notes (ECN), settlement obligation sheets, and depository transfer receipts.
  - **Notification Service (Prompt 211):** Dispatches P1/P2 operational alerts to compliance teams (Slack, PagerDuty) and delivers statutory resolution notices to complainants via DLT-registered SMS and registered email.
- **Kafka Topics:**
  - *Consumes:*
    - `user.account_status_changed.v1`: Updates compliance records if an investor account is suspended, frozen, or closed.
    - `wallet.deposit_settled.v1`: Real-time ingestion of bank UTR and settlement receipts for automated diagnostic triage.
    - `trade.settled.v1`: Trade execution confirmations for instant contract note evidence binding.
  - *Publishes:*
    - `grievance.complaint.ingested.v1`: Emitted upon intake and docket creation from SCORES/RBI/ODR.
    - `grievance.sla.warning.v1`: Emitted at Day 7, Day 14, and Day 18 SLA countdown thresholds.
    - `grievance.sla.breached.v1`: Emitted if a docket exceeds the statutory 21-day or 30-day resolution window.
    - `grievance.atr.submitted.v1`: Emitted when an Action Taken Report is formally submitted to the regulator.
    - `grievance.receipt.anchored.v1`: Emitted when the Besu transaction hash confirms on-chain timestamp anchoring.
- **PostgreSQL Database Tables:**
  - `grievance_complaints`: Master entity tracking regulatory dockets, source portals, complainant details, and current FSM state.
  - `grievance_dossiers`: Aggregated evidence dossiers containing references to audit logs, ledger entries, contract notes, and banking UTR receipts.
  - `grievance_action_taken_reports`: Compiled ATR documents, Maker-Checker sign-off records, DSC signatures, and regulatory acknowledgment numbers.
  - `grievance_sla_events`: Timestamped log of countdown events, threshold alerts, and escalation actions.
  - `grievance_besu_anchors`: Blockchain anchoring records storing docket IDs, SHA-256 hashes, transaction receipts, block numbers, and gas receipts.
  - `grievance_audit_logs`: Append-only tamper-evident log of all internal actions, view events, annotations, and state transitions.
- **Redis Keys & Data Structures:**
  - `scores:sla:deadline:zset`: Sorted set with score = Unix timestamp of statutory deadline; member = `complaint_id`.
  - `lock:grievance:{complaint_id}`: Distributed mutex ensuring atomic state transitions and single-flight ATR submission.
  - `cache:pan_mapping:{hashed_pan}`: In-memory cache mapping investor PAN hashes to internal `user_id` UUIDs.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu enterprise permissioned consortium network operating Istanbul Byzantine Fault Tolerant / QBFT consensus.
- **Smart Contract Interface:** `GrievanceReceiptRegistry.sol` deployed on the Besu compliance subnet:
  ```solidity
  // SPDX-License-Identifier: Apache-2.0
  pragma solidity ^0.8.24;

  interface IGrievanceReceiptRegistry {
      event ReceiptAnchored(
          bytes32 indexed complaintId,
          bytes32 indexed complaintHash,
          bytes32 atrPackageHash,
          uint64 submissionTimestamp,
          bytes32 regulatorRef,
          address indexed relayer
      );

      function anchorGrievanceReceipt(
          bytes32 complaintId,
          bytes32 complaintHash,
          bytes32 atrPackageHash,
          uint64 submissionTimestamp,
          bytes32 regulatorRef
      ) external returns (bool);

      function verifyReceipt(
          bytes32 complaintId
      ) external view returns (
          bytes32 complaintHash,
          bytes32 atrPackageHash,
          uint64 submissionTimestamp,
          bytes32 regulatorRef,
          uint256 blockTimestamp,
          address relayer
      );
  }
  ```
- **Cryptographic Anchoring Workflow:**
  1. *Intake Digest:* Upon docketing, the service computes `complaintHash = SHA-256(external_docket_id || regulator_code || intake_timestamp_iso || anonymized_user_uuid)`.
  2. *ATR Submission Digest:* Upon successful ATR generation, DSC signing, and regulatory transmission, the service computes `atrPackageHash = SHA-256(signed_atr_pdf_bytes || regulatory_ack_token || submission_timestamp_iso)`.
  3. *On-Chain Commit:* An asynchronous Web3 relayer worker signs and submits `anchorGrievanceReceipt` to `GrievanceReceiptRegistry.sol`.
  4. *Receipt Verification:* The returned Ethereum transaction hash, block number, and block timestamp are persisted in `grievance_besu_anchors`. This provides undeniable mathematical proof in court or SEBI appellate tribunals (Securities Appellate Tribunal - SAT) that the entity resolved the complaint within statutory deadlines.
- **Zero PII on Ledger Guarantee:**
  - Absolute exclusion of plain-text Personally Identifiable Information (PII) from blockchain parameters.
  - No investor names, PAN cards, mobile numbers, email addresses, or physical locations are transmitted to Besu.
  - Only cryptographic SHA-256 digests and pseudonymous internal UUIDs exist on-chain, satisfying the Digital Personal Data Protection (DPDP) Act, 2023.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Service Directory & Project Structure:** Scaffold `services/grievance-gateway` adhering to FastAPI modular architecture, with directories for `api/`, `core/`, `adapters/`, `fsm/`, `workers/`, `models/`, `schemas/`, and `migrations/`.
2. **Define Data Models & PostgreSQL Migrations:** Author Alembic/SQL DDL migrations creating `grievance_complaints`, `grievance_dossiers`, `grievance_action_taken_reports`, `grievance_sla_events`, `grievance_besu_anchors`, and `grievance_audit_logs` with proper indexes, foreign keys, and partition strategies.
3. **Implement Core Configuration & Vault Integration:** Configure environment settings via Pydantic Settings (`core/config.py`) to pull mTLS client certificates, HSM PKCS#11 credentials, API keys, and Besu RPC secrets from HashiCorp Vault or AWS Secrets Manager.
4. **Implement External Regulatory Inbound Adapters:**
   - Develop `adapters/scores_adapter.py` supporting SEBI SCORES 2.0 mTLS REST endpoints, polling workers, and webhook verification.
   - Develop `adapters/rbi_cms_adapter.py` supporting RBI Ombudsman CMS API and batch notification formats.
   - Develop `adapters/smart_odr_adapter.py` supporting SMART ODR webhook parsing and conciliation event processing.
5. **Implement Complainant User Resolution Engine:** Build `core/user_resolver.py` querying User Service (Prompt 201) to automatically map incoming complaint PANs, UCCs, and contact credentials to internal user accounts. Fall back to manual unmapped compliance queue if resolution fails.
6. **Implement Automated Diagnostic Triage Service:** Construct `core/diagnostic_triage.py` executing automated rules for high-frequency complaint categories (deposit status, withdrawal processing, contract note generation, dividend credits). Automatically gather transaction receipts and generate preliminary diagnostic findings within 5 seconds of intake.
7. **Implement Asynchronous Evidence Dossier Harvesting:** Build Celery worker `workers/dossier_collector.py` to aggregate records across microservices:
   - Call User Service (Prompt 201) for KYC status and risk category.
   - Call Audit Log Service (Prompt 218) for cryptographically validated WORM system logs.
   - Call Wallet Account Service (Prompt 203) for bank UTRs, ledger statements, and transaction logs.
   - Call Trade Settlement Service (Prompt 208) for matching logs and electronic contract notes.
   - Package gathered artifacts into MinIO/S3 WORM compliance storage bucket.
8. **Build Complaint Lifecycle Finite State Machine:** Implement `fsm/state_machine.py` enforcing strict transitions between states (`INGESTED` through `RESOLVED_CLOSED`), validating prerequisites and enforcing Redlock distributed locks on `complaint_id`.
9. **Implement Redis SLA Countdown & Breach Prevention Daemon:**
   - Configure Celery Beat periodic task running every 60 seconds (`workers/sla_monitor.py`).
   - Query Redis Sorted Set `scores:sla:deadline:zset` for approaching deadlines.
   - Trigger multi-tier alerts: Day 7 Warning, Day 14 Officer Escalation, Day 18 Critical P1 Escalation (RUNBOOK-34), and Day 20 Mandatory Sign-Off Freeze.
   - Emit notifications to Kafka `grievance.sla.warning.v1` and dispatch Slack/PagerDuty messages via Notification Service (Prompt 211).
10. **Implement Maker-Checker ATR Authoring Workflow:**
    - Develop endpoints for Compliance Officers to draft Action Taken Reports (`POST /api/v1/grievance/{id}/atr/draft`).
    - Develop dual-control approval endpoints for Senior Compliance Officers / Principal Compliance Officer (`POST /api/v1/grievance/{id}/atr/approve`), strictly enforcing distinct user identity between Maker and Checker.
11. **Implement PDF/A Compiler & HSM DSC Signing Engine:**
    - Build `core/pdf_compiler.py` rendering standardized, tamper-evident PDF/A-1b Action Taken Reports embedding evidence dossiers, transaction logs, and internal findings.
    - Implement `core/dsc_signer.py` using `pyHanko` to interface with PKCS#11 HSM tokens and append a cryptographically valid Class-3 Digital Signature Certificate.
12. **Implement Regulatory ATR Dispatcher:** Build `core/atr_dispatcher.py` to transmit the signed ATR PDF and structured metadata to the relevant regulatory portal (SCORES 2.0, RBI CMS, SMART ODR), verify HTTP 200/201 response, and record the external acknowledgment reference.
13. **Implement Hyperledger Besu Cryptographic Anchoring Relayer:** Build `workers/blockchain_relayer.py` using `web3.py` to compute SHA-256 digests of the complaint and ATR package, invoke `anchorGrievanceReceipt` on `GrievanceReceiptRegistry.sol`, and persist transaction metadata in `grievance_besu_anchors`.
14. **Implement Regulatory Export & Metrics Aggregator:** Implement `api/v1/endpoints/reporting.py` aggregating grievance statistics (intake count, resolution turnaround time, category distributions, pending dockets by age) and publishing to Regulatory Reporting Service (Prompt 216).
15. **Build Comprehensive Test Suite & Chaos Scenarios:** Write pytest suites testing:
    - End-to-end docket ingestion from mock SCORES 2.0 / RBI CMS servers.
    - Automated PAN mapping and fallback workflows.
    - SLA countdown progression and Day 18 RUNBOOK-34 breach escalations.
    - Maker-Checker dual-control enforcement (ensuring Maker cannot approve own ATR).
    - DSC signature verification and Besu on-chain receipt verification.

## Interfaces / Contracts

### OpenAPI 3.1 REST Specification (`openapi.yaml`)
```yaml
openapi: 3.1.0
info:
  title: SEBI SCORES 2.0 & Regulatory Grievance Gateway API
  description: Enterprise gateway managing regulatory grievances, automated evidence collection, SLA countdowns, and ATR submissions.
  version: 1.0.0
servers:
  - url: https://grievance-gateway.internal.growww.in
    description: Internal Production Gateway
paths:
  /api/v1/grievance/scores/webhook:
    post:
      summary: Ingest complaint from SEBI SCORES 2.0
      operationId: ingestScoresComplaint
      security:
        - MutualTLS: []
        - HmacSignature: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ScoresWebhookPayload'
      responses:
        '201':
          description: Docket created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GrievanceDocketResponse'
        '400':
          description: Invalid payload or signature verification failed
        '409':
          description: Duplicate complaint docket already ingested

  /api/v1/grievance/{complaint_id}/dossier/collect:
    post:
      summary: Trigger automated cross-service evidence harvesting
      operationId: triggerDossierCollection
      parameters:
        - name: complaint_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '202':
          description: Evidence harvesting initiated in background
        '404':
          description: Grievance docket not found

  /api/v1/grievance/{complaint_id}/atr/draft:
    post:
      summary: Draft an Action Taken Report (Maker action)
      operationId: draftAtr
      parameters:
        - name: complaint_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AtrDraftRequest'
      responses:
        '200':
          description: ATR drafted and moved to PENDING_CHECKER_APPROVAL
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AtrDetailsResponse'
        '403':
          description: Insufficient permissions for Maker role

  /api/v1/grievance/{complaint_id}/atr/approve:
    post:
      summary: Approve Action Taken Report (Checker action)
      operationId: approveAtr
      parameters:
        - name: complaint_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AtrApprovalRequest'
      responses:
        '200':
          description: ATR approved and queued for DSC signing and regulatory dispatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AtrDetailsResponse'
        '400':
          description: Dual-control violation (Maker cannot approve own draft)

  /api/v1/grievance/{complaint_id}/atr/submit:
    post:
      summary: Formally sign and dispatch ATR to regulatory portal
      operationId: submitAtrToRegulator
      parameters:
        - name: complaint_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: ATR transmitted, acknowledged by regulator, and anchored to Besu
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AtrSubmissionResponse'

  /api/v1/grievance/sla/dashboard:
    get:
      summary: Query real-time SLA countdown and active escalation tiers
      operationId: getSlaDashboard
      parameters:
        - name: regulator
          in: query
          required: false
          schema:
            type: string
            enum: [SEBI_SCORES, RBI_CMS, SMART_ODR]
      responses:
        '200':
          description: Real-time SLA monitoring statistics
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SlaDashboardResponse'

components:
  securitySchemes:
    MutualTLS:
      type: mutualTLS
    HmacSignature:
      type: apiKey
      name: X-Scores-Signature-256
      in: header

  schemas:
    ScoresWebhookPayload:
      type: object
      required:
        - scores_registration_number
        - complainant_name
        - complainant_pan
        - complaint_category
        - complaint_description
        - receipt_date
      properties:
        scores_registration_number:
          type: string
          example: "SEBIP/2026/0014529"
        complainant_name:
          type: string
          example: "Rahul Sharma"
        complainant_pan:
          type: string
          pattern: "^[A-Z]{5}[0-9]{4}[A-Z]{1}$"
          example: "ABCDE1234F"
        complainant_email:
          type: string
          format: email
        complainant_mobile:
          type: string
          example: "+919876543210"
        complaint_category:
          type: string
          example: "NON_CREDIT_OF_FUNDS"
        sub_category:
          type: string
          example: "UPI_DEPOSIT_DELAY"
        complaint_description:
          type: string
          example: "Funds debited from bank account via UPI but not credited to trading wallet."
        receipt_date:
          type: string
          format: date-time
        disputed_amount:
          type: number
          format: double
          example: 50000.00
        demat_boid:
          type: string
          example: "1208160012345678"

    GrievanceDocketResponse:
      type: object
      required:
        - complaint_id
        - external_registration_number
        - regulator
        - status
        - statutory_deadline
        - days_remaining
      properties:
        complaint_id:
          type: string
          format: uuid
        external_registration_number:
          type: string
        regulator:
          type: string
          enum: [SEBI_SCORES, RBI_CMS, SMART_ODR]
        status:
          type: string
        matched_user_id:
          type: string
          format: uuid
          nullable: true
        statutory_deadline:
          type: string
          format: date-time
        days_remaining:
          type: integer
        intake_timestamp:
          type: string
          format: date-time

    AtrDraftRequest:
      type: object
      required:
        - resolution_category
        - action_taken_summary
        - redressal_details
      properties:
        resolution_category:
          type: string
          enum: [RESOLVED_SATISFIED, RESOLVED_WITH_REFUND, REJECTED_UNFOUNDED, CLARIFICATION_PROVIDED]
        action_taken_summary:
          type: string
          maxLength: 1000
        redressal_details:
          type: string
        refund_amount:
          type: number
          format: double
        bank_utr:
          type: string
        supporting_document_ids:
          type: array
          items:
            type: string
            format: uuid

    AtrApprovalRequest:
      type: object
      required:
        - checker_notes
        - approved
      properties:
        checker_notes:
          type: string
        approved:
          type: boolean

    AtrDetailsResponse:
      type: object
      required:
        - atr_id
        - complaint_id
        - status
        - maker_user_id
        - draft_content
      properties:
        atr_id:
          type: string
          format: uuid
        complaint_id:
          type: string
          format: uuid
        status:
          type: string
        maker_user_id:
          type: string
          format: uuid
        checker_user_id:
          type: string
          format: uuid
          nullable: true
        draft_content:
          type: object
        created_at:
          type: string
          format: date-time

    AtrSubmissionResponse:
      type: object
      required:
        - atr_id
        - complaint_id
        - regulatory_ack_number
        - submitted_at
        - besu_transaction_hash
        - dsc_signature_valid
      properties:
        atr_id:
          type: string
          format: uuid
        complaint_id:
          type: string
          format: uuid
        regulatory_ack_number:
          type: string
          example: "SEBI-ATR-ACK-2026-98124"
        submitted_at:
          type: string
          format: date-time
        besu_transaction_hash:
          type: string
          example: "0x7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069"
        dsc_signature_valid:
          type: boolean

    SlaDashboardResponse:
      type: object
      required:
        - total_active_dockets
        - critical_breach_risk_count
        - dockets_by_tier
      properties:
        total_active_dockets:
          type: integer
        critical_breach_risk_count:
          type: integer
        dockets_by_tier:
          type: object
          properties:
            stage_0_normal:
              type: integer
            stage_1_day_7_warning:
              type: integer
            stage_2_day_14_escalation:
              type: integer
            stage_3_day_18_p1_critical:
              type: integer
            stage_4_day_20_freeze:
              type: integer
```

### Protobuf Service Contract (`proto/growww/grievance/v1/grievance_service.proto`)
```protobuf
syntax = "proto3";

package growww.grievance.v1;

option go_package = "growww/grievance/v1;grievancev1";
option py_generic_services = true;

// Internal gRPC service for inter-microservice grievance queries and automated triage
service GrievanceGatewayService {
  rpc GetComplaintDetails (GetComplaintRequest) returns (GetComplaintResponse);
  rpc TriggerAutomatedTriage (TriggerTriageRequest) returns (TriggerTriageResponse);
  rpc IngestEvidenceArtifact (IngestEvidenceRequest) returns (IngestEvidenceResponse);
  rpc QueryUserGrievanceHistory (UserHistoryRequest) returns (UserHistoryResponse);
  rpc GetSlaStatus (GetSlaStatusRequest) returns (GetSlaStatusResponse);
}

enum RegulatorType {
  REGULATOR_TYPE_UNSPECIFIED = 0;
  REGULATOR_TYPE_SEBI_SCORES = 1;
  REGULATOR_TYPE_RBI_CMS = 2;
  REGULATOR_TYPE_SMART_ODR = 3;
}

enum GrievanceState {
  GRIEVANCE_STATE_UNSPECIFIED = 0;
  GRIEVANCE_STATE_INGESTED = 1;
  GRIEVANCE_STATE_MAPPED_TO_USER = 2;
  GRIEVANCE_STATE_DIAGNOSTIC_TRIAGE = 3;
  GRIEVANCE_STATE_INVESTIGATION_PENDING = 4;
  GRIEVANCE_STATE_AUDIT_DOSSIER_COLLECTED = 5;
  GRIEVANCE_STATE_ATR_DRAFTED_MAKER = 6;
  GRIEVANCE_STATE_ATR_APPROVED_CHECKER = 7;
  GRIEVANCE_STATE_ATR_SUBMITTED = 8;
  GRIEVANCE_STATE_FIRST_REVIEW_ESCALATED = 9;
  GRIEVANCE_STATE_ODR_CONCILIATION = 10;
  GRIEVANCE_STATE_RESOLVED_CLOSED = 11;
  GRIEVANCE_STATE_REJECTED_INVALID = 12;
}

message GetComplaintRequest {
  string complaint_id = 1;
}

message GetComplaintResponse {
  string complaint_id = 1;
  string external_reg_number = 2;
  RegulatorType regulator = 3;
  GrievanceState state = 4;
  string user_id = 5;
  string complainant_pan_hash = 6;
  string category = 7;
  string description = 8;
  int64 intake_timestamp = 9;
  int64 statutory_deadline = 10;
  int32 days_remaining = 11;
  string besu_tx_hash = 12;
}

message TriggerTriageRequest {
  string complaint_id = 1;
}

message TriggerTriageResponse {
  string complaint_id = 1;
  bool automated_resolution_possible = 2;
  string diagnostic_summary = 3;
  string evidence_dossier_id = 4;
}

message IngestEvidenceRequest {
  string complaint_id = 1;
  string source_service = 2; // user-service, wallet-service, audit-log-service
  string artifact_type = 3;   // BANK_UTR, CONTRACT_NOTE, AUDIT_LOG_EXCERPT
  string document_uri = 4;
  string document_sha256 = 5;
}

message IngestEvidenceResponse {
  string evidence_id = 1;
  bool successfully_attached = 2;
}

message UserHistoryRequest {
  string user_id = 1;
  string pan_hash = 2;
}

message UserHistoryResponse {
  repeated GetComplaintResponse complaints = 1;
  int32 total_historical_complaints = 2;
  int32 active_complaints = 3;
}

message GetSlaStatusRequest {
  string complaint_id = 1;
}

message GetSlaStatusResponse {
  string complaint_id = 1;
  int64 intake_timestamp = 2;
  int64 deadline_timestamp = 3;
  int32 total_sla_days = 4; // 21 for SCORES, 30 for RBI
  int32 days_elapsed = 5;
  int32 days_remaining = 6;
  int32 escalation_stage = 7; // 0=Normal, 1=Day 7, 2=Day 14, 3=Day 18 (P1), 4=Day 20
  bool is_breached = 8;
}
```

### PostgreSQL Database Schema (`services/grievance-gateway/migrations/001_initial_schema.sql`)
```sql
-- PostgreSQL 16 Migration for SEBI SCORES 2.0 & Regulatory Grievance Gateway

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums for Regulatory Portals and FSM States
CREATE TYPE regulator_portal_type AS ENUM (
    'SEBI_SCORES',
    'RBI_CMS',
    'SMART_ODR'
);

CREATE TYPE grievance_fsm_state AS ENUM (
    'INGESTED',
    'MAPPED_TO_USER',
    'DIAGNOSTIC_TRIAGE',
    'INVESTIGATION_PENDING',
    'AUDIT_DOSSIER_COLLECTED',
    'ATR_DRAFTED_MAKER',
    'ATR_APPROVED_CHECKER',
    'ATR_SUBMITTED_TO_REGULATOR',
    'FIRST_REVIEW_ESCALATED',
    'ODR_CONCILIATION',
    'RESOLVED_CLOSED',
    'REJECTED_INVALID'
);

CREATE TYPE resolution_disposition AS ENUM (
    'RESOLVED_SATISFIED',
    'RESOLVED_WITH_REFUND',
    'REJECTED_UNFOUNDED',
    'CLARIFICATION_PROVIDED',
    'SETTLED_VIA_ODR'
);

-- Master Table: Regulatory Grievance Complaints
CREATE TABLE grievance_complaints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_registration_number VARCHAR(64) NOT NULL UNIQUE,
    regulator regulator_portal_type NOT NULL,
    current_state grievance_fsm_state NOT NULL DEFAULT 'INGESTED',
    
    -- Anonymized complainant mapping (DPDP Act 2023 compliant)
    complainant_pan_hash CHAR(64) NOT NULL,
    user_id UUID, -- Foreign reference to User Service (nullable if unmapped)
    demat_boid_hash CHAR(64),
    
    -- Encrypted PII fields (AES-256-GCM encrypted off-chain)
    encrypted_complainant_name BYTEA NOT NULL,
    encrypted_contact_email BYTEA,
    encrypted_contact_phone BYTEA,
    
    -- Complaint Details
    category VARCHAR(64) NOT NULL,
    sub_category VARCHAR(64),
    disputed_amount NUMERIC(18, 4) DEFAULT 0.0000,
    complaint_description TEXT NOT NULL,
    raw_payload JSONB NOT NULL,
    
    -- Statutory SLA Timestamps
    statutory_sla_days INT NOT NULL DEFAULT 21, -- 21 for SEBI, 30 for RBI
    intake_timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    statutory_deadline TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    
    -- Escalation tracking
    current_escalation_stage INT NOT NULL DEFAULT 0, -- 0=Normal, 1=Day7, 2=Day14, 3=Day18 (P1), 4=Day20
    is_sla_breached BOOLEAN NOT NULL DEFAULT FALSE,
    assigned_compliance_officer_id UUID,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grievance_complaints_regulator_state ON grievance_complaints (regulator, current_state);
CREATE INDEX idx_grievance_complaints_deadline ON grievance_complaints (statutory_deadline) WHERE current_state NOT IN ('RESOLVED_CLOSED', 'REJECTED_INVALID');
CREATE INDEX idx_grievance_complaints_pan_hash ON grievance_complaints (complainant_pan_hash);
CREATE INDEX idx_grievance_complaints_user_id ON grievance_complaints (user_id);

-- Evidence Dossier Harvesting Table
CREATE TABLE grievance_dossiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    complaint_id UUID NOT NULL REFERENCES grievance_complaints(id) ON DELETE CASCADE,
    
    user_service_profile_snapshot JSONB,
    wallet_ledger_statement JSONB,
    trade_contract_notes JSONB,
    audit_log_event_ids JSONB,
    
    -- Immutable WORM Storage Reference
    worm_storage_bucket VARCHAR(64) NOT NULL,
    worm_storage_key VARCHAR(256) NOT NULL,
    dossier_sha256 CHAR(64) NOT NULL,
    
    harvesting_completed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_grievance_dossiers_complaint_id ON grievance_dossiers (complaint_id);

-- Action Taken Reports (ATR) with Maker-Checker Dual-Control
CREATE TABLE grievance_action_taken_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    complaint_id UUID NOT NULL REFERENCES grievance_complaints(id) ON DELETE CASCADE,
    
    disposition resolution_disposition NOT NULL,
    action_taken_summary TEXT NOT NULL,
    detailed_redressal_text TEXT NOT NULL,
    refund_amount NUMERIC(18, 4) DEFAULT 0.0000,
    settlement_bank_utr VARCHAR(64),
    
    -- Dual-Control Maker-Checker Tracking
    maker_user_id UUID NOT NULL,
    maker_drafted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    checker_user_id UUID,
    checker_approved_at TIMESTAMPTZ,
    checker_notes TEXT,
    
    -- Artifacts & Signing
    generated_pdf_worm_key VARCHAR(256),
    pdf_sha256 CHAR(64),
    is_dsc_signed BOOLEAN NOT NULL DEFAULT FALSE,
    dsc_signer_dn VARCHAR(256),
    dsc_signed_at TIMESTAMPTZ,
    
    -- Regulatory Transmission
    regulatory_submission_status VARCHAR(32) NOT NULL DEFAULT 'DRAFT', -- DRAFT, APPROVED, SUBMITTED, ACKNOWLEDGED, FAILED
    regulatory_ack_number VARCHAR(128),
    regulatory_ack_timestamp TIMESTAMPTZ,
    raw_regulatory_response JSONB,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_dual_control CHECK (maker_user_id <> checker_user_id)
);

CREATE INDEX idx_grievance_atr_complaint ON grievance_action_taken_reports (complaint_id);

-- SLA Lifecycle and Alert Tracking
CREATE TABLE grievance_sla_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    complaint_id UUID NOT NULL REFERENCES grievance_complaints(id) ON DELETE CASCADE,
    escalation_stage INT NOT NULL,
    days_elapsed INT NOT NULL,
    trigger_type VARCHAR(32) NOT NULL, -- DAY_7_WARNING, DAY_14_ESCALATION, DAY_18_P1_CRITICAL, DAY_20_FREEZE, SLA_BREACH
    alert_channel VARCHAR(32) NOT NULL, -- SLACK, PAGERDUTY, EMAIL, SMS
    alert_payload JSONB NOT NULL,
    dispatched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grievance_sla_events_complaint ON grievance_sla_events (complaint_id);

-- Hyperledger Besu On-Chain Anchoring Table
CREATE TABLE grievance_besu_anchors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    complaint_id UUID NOT NULL REFERENCES grievance_complaints(id) ON DELETE CASCADE,
    
    complaint_hash CHAR(64) NOT NULL,
    atr_package_hash CHAR(64) NOT NULL,
    regulator_ref_bytes32 CHAR(66) NOT NULL,
    
    besu_tx_hash CHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    gas_used BIGINT NOT NULL,
    relayer_address CHAR(42) NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_besu_anchors_complaint_id ON grievance_besu_anchors (complaint_id);
CREATE INDEX idx_besu_anchors_tx_hash ON grievance_besu_anchors (besu_tx_hash);

-- Immutable Append-Only Audit Trail Table
CREATE TABLE grievance_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    complaint_id UUID NOT NULL REFERENCES grievance_complaints(id) ON DELETE CASCADE,
    action_type VARCHAR(64) NOT NULL,
    actor_id UUID NOT NULL,
    actor_role VARCHAR(32) NOT NULL,
    previous_state grievance_fsm_state,
    new_state grievance_fsm_state,
    metadata JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    logged_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grievance_audit_logs_complaint ON grievance_audit_logs (complaint_id, logged_at);
```

## Security & Compliance Notes
- **SEBI Circular `SEBI/HO/OIAE/OIAE_IAD-1/P/CIR/2023/145` Mandate:**
  - Enforces mandatory resolution within 21 calendar days from docket intake.
  - Intermediaries must maintain automated, straight-through routing directly to designated compliance departments.
  - Failure to file an ATR within 21 calendar days triggers automatic escalation to the First Review stage (Stock Exchanges and Clearing Corporations) and Second Review stage (SEBI Regulatory Panels), carrying statutory penalties under Section 15C of the SEBI Act, 1992.
- **DPDP Act 2023 & Zero PII Leakage Architecture:**
  - Compliance with the Digital Personal Data Protection Act, 2023 requires strict purpose limitation, data minimization, and encryption.
  - Complainant PII (legal name, mobile number, email, residential address) is encrypted at rest within PostgreSQL using AES-256-GCM envelopes, with data keys managed via AWS KMS / HashiCorp Vault.
  - Zero PII is committed to Hyperledger Besu; on-chain contracts accept strictly SHA-256 hash digests and pseudonymous UUIDs.
  - Access to raw PII fields is restricted by Role-Based Access Control (RBAC) to verified Compliance Officers possessing legitimate operational need.
- **Statutory WORM Document Retention (8-Year Rule):**
  - Under the Prevention of Money Laundering Act (PMLA) 2002 and SEBI Broker Regulations, all regulatory dockets, internal diagnostic logs, evidence files, and signed ATR packages must be archived for a minimum of eight (8) years.
  - Evidence dossiers and generated PDF/A ATRs are pushed to MinIO / AWS S3 equipped with Object Lock in Compliance Mode (WORM - Write Once Read Many), preventing alteration or premature deletion even by root system administrators.
- **Maker-Checker Dual-Control Enforcement:**
  - Strict separation of duty: The compliance personnel drafting an Action Taken Report (`maker_user_id`) cannot approve or authorize submission of the same ATR (`checker_user_id`).
  - The PostgreSQL database enforces this constraint at the engine level (`CONSTRAINT chk_dual_control CHECK (maker_user_id <> checker_user_id)`).
- **Cryptographic Class-3 Digital Signatures (DSC):**
  - ATR PDFs must be signed using an organizational Class-3 Digital Signature Certificate (DSC) issued by a licensed Certifying Authority (e.g., eMudhra, (n)Code Solutions).
  - Private keys remain securely housed in FIPS 140-2 Level 3 Hardware Security Modules (HSMs) accessible strictly over PKCS#11 network protocols.
- **Mutual TLS & Perimeter Defense:**
  - All communication with external regulatory gateways (SEBI SCORES 2.0, RBI CMS) executes over Mutual TLS (mTLS 1.3) with pinned certificates and dedicated leased line / static IP egress tunnels.
  - Webhooks received from SMART ODR must validate HMAC-SHA256 request signatures before payload processing.

## Acceptance Criteria
- **SLA Breach Prevention:** 100% of valid SEBI SCORES 2.0 complaints are resolved and have ATRs filed in $< 21\text{ calendar days}$; 100% of RBI Ombudsman dockets are resolved in $< 30\text{ calendar days}$.
- **Automated Evidence Aggregation:** Cross-service evidence dossier harvesting (User Service, Wallet Service, Trade Settlement, Audit Log Service) completes in $< 60\text{ seconds}$ from docket intake.
- **Diagnostic Triage Speed:** Self-service automated diagnostic triage resolves high-frequency deposit/contract-note inquiries in $< 5\text{ seconds}$.
- **Maker-Checker Integrity:** Attempting to approve an ATR with identical `maker_user_id` and `checker_user_id` returns HTTP 400 and database constraint violation.
- **On-Chain Anchoring Latency:** 100% of submitted ATR packages are cryptographically anchored to `GrievanceReceiptRegistry.sol` on Hyperledger Besu within 120 seconds of regulatory submission acknowledgment.
- **Zero PII Exposure:** Automated static code analysis and transaction inspection confirm that zero unhashed PANs, phone numbers, or complainant names are emitted in Kafka topics or committed to Besu smart contracts.
- **WORM Storage Conformance:** All generated PDF/A-1b ATR files and evidence packages are stored in MinIO/S3 buckets with Object Lock retention set to $\ge 8\text{ years}$.
- **Test Coverage:** Comprehensive unit, integration, and chaos test suites achieve $\ge 90\%$ line and branch coverage across all adapter, FSM, and worker modules.

## Suggested Order / Dependencies
1. **Prerequisites (Upstream Services):**
   - **User Service (Prompt 201):** User profile, PAN hash mapping, and KYC state endpoints must be operational.
   - **Wallet Account Service (Prompt 203):** Double-entry ledger query endpoints and UTR retrieval must be functional.
   - **Trade Settlement Service (Prompt 208):** Contract note retrieval and settlement verification APIs must be live.
   - **Audit Log Service (Prompt 218):** WORM audit log retrieval endpoints must be accessible.
   - **Notification Service (Prompt 211):** Kafka consumers and multi-channel dispatch APIs (Slack, PagerDuty, SMS) must be active.
2. **Phase 1 - Database Schema & Data Models:** Deploy PostgreSQL 16 migrations (`001_initial_schema.sql`) and configure Redis clusters.
3. **Phase 2 - Regulatory Ingestion Adapters & State Machine:** Implement SCORES 2.0, RBI CMS, and SMART ODR adapters, along with the complaint lifecycle FSM.
4. **Phase 3 - Evidence Harvesting & Diagnostic Triage:** Build Celery worker pipelines connecting to User, Wallet, Settlement, and Audit Log services.
5. **Phase 4 - SLA Monitor & Alerting Daemon:** Implement the 21-day countdown daemon and multi-stage breach escalation (RUNBOOK-34).
6. **Phase 5 - Maker-Checker Workflow & DSC Signing:** Build the dual-control ATR authoring interface, PDF/A-1b rendering engine, and HSM PKCS#11 signing bridge.
7. **Phase 6 - Besu Blockchain Relayer & Receipt Anchoring:** Implement Web3 JSON-RPC anchoring to `GrievanceReceiptRegistry.sol`.
8. **Phase 7 - Regulatory Reporting Export (Downstream):** Expose consolidated analytics endpoints to feed Regulatory Reporting Service (Prompt 216).
