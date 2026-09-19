# 617 - Next.js 14 Regulatory Grievance & SEBI SCORES 2.0 Management Portal

## Purpose
Operating an equity and tokenized financial market infrastructure under the regulatory purview of the Securities and Exchange Board of India (SEBI), the Reserve Bank of India (RBI), and the International Financial Services Centres Authority (IFSCA) imposes uncompromising legal obligations regarding investor dispute resolution. Under the SEBI Master Circular on Redressal of Investor Grievances through the SEBI Complaints Redress System (SCORES 2.0), the Securities and Exchange Board of India (Alternative Dispute Resolution Mechanism) Regulations, the RBI Integrated Ombudsman Scheme (RB-IOS), and the SMART ODR (Securities Market Approach for Resolution Through ODR Portal) circulars, market infrastructure institutions (MIIs) and registered brokerages must resolve investor grievances within a strict statutory timeframe of 21 calendar days from receipt. Failure to adhere to these statutory Turnaround Times (TATs) triggers automatic systemic escalation, formal show-cause notices, regulatory fines, and administrative enforcement actions.

Managing these multi-channel regulatory mandates cannot rely on generic customer support ticketing software or fragmented spreadsheets. Exchange legal officers, designated Grievance Redressal Officers (GROs), and Principal Compliance Officers (PCOs) require a specialized, high-assurance web administration portal to ingest, triage, investigate, and adjudicate formal disputes. This portal must orchestrate comprehensive Action Taken Reports (ATRs), validate complex supporting evidence (including nanosecond-level trade logs, demat allocation receipts, and atomic on-chain settlement records), enforce dual-officer maker-checker approvals, and cryptographically anchor final dispute resolutions to the Hyperledger Besu consortium blockchain to guarantee an immutable, legally defensible chain of custody.

This prompt specifies the architecture, technical requirements, and implementation blueprint of the **Next.js 14 Regulatory Grievance & SEBI SCORES 2.0 Management Portal** (`apps/growww_admin/grievance`). Governed by Architectural Decision Record **ADR-0046** (*Regulatory Grievance Ledger & Evidentiary Chain of Custody*) and operational procedure **RUNBOOK-34** (*SCORES 2.0 & SMART ODR Statutory Incident Handling*), this high-security web console provides exchange legal teams with real-time complaint triage queues, statutory 21-day SLA countdown widgets, rich-text ATR drafting editors, client-side cryptographic evidence hashing, WebAuthn-based maker-checker verification, Besu ledger audit verification, and Section 65B certified evidence export.

## What You Are Building
An enterprise-grade, regulatory-hardened Next.js 14 web application located at `apps/growww_admin/grievance`, featuring:
- `GrievanceDashboardShell`: Institutional administrative shell providing role-based route guards (Grievance Redressal Officer, Legal Counsel, Principal Compliance Officer, Statutory Auditor), dynamic forensic screen watermarking, and active regulatory filter toggles (SEBI SCORES 2.0, RBI Ombudsman, SMART ODR, Internal Escalations).
- `ComplaintTriageQueue`: High-performance, virtualized data table rendering multi-source regulatory complaints with millisecond-precision filtering, sorting, multi-attribute tagging, and instant priority triaging.
- `SlaCountdownWidget`: Visual statutory SLA tracker computing real-time remaining calendar and business days against the 21-day SEBI legal limit, displaying multi-tiered urgency alerts (Green, Amber, Red, Flashing Statutory Critical).
- `ComplaintDetailViewer`: Comprehensive dispute investigation console displaying original complainant claims, linked investor profiles, order book executions, settlement transaction hashes, demat transfers, and historical customer communication transcripts.
- `AtrDraftingEditor`: Specialized rich-text editor conforming strictly to SEBI SCORES 2.0 structured report schemas (dispute classification, root cause analysis, factual findings, corrective actions, restitution amounts), equipped with regulatory template presets.
- `EvidenceVaultUploader`: Cryptographically secured evidence management module that performs in-browser SHA-256 digest computation, enforces strict PDF/A sanitization, verifies anti-virus scan attestations, and registers files into an S3 Write-Once-Read-Many (WORM) repository.
- `DualOfficerReviewConsole`: Maker-checker authorization workflow requiring WebAuthn/FIDO2 hardware cryptographic signatures from both the drafting legal officer (Maker) and the Principal Compliance Officer (Checker) prior to official transmission to regulatory endpoints.
- `BesuProofTimestampInspector`: Interactive Web3 verification panel querying Hyperledger Besu permissioned RPC nodes to validate block timestamps, transaction receipts, and Merkle inclusion proofs of anchored ATR documents.
- `AuditTrailExporter`: Evidence preservation engine producing cryptographically signed Section 65B (Indian Evidence Act, 1872 / Bharatiya Sakshya Adhiniyam, 2023) certified forensic ZIP archives containing raw logs, tamper manifests, and blockchain inclusion receipts.

## Scope Boundaries
- **In Scope:**
  - Complete Next.js 14 App Router application structure within `apps/growww_admin/grievance`.
  - Responsive, high-density compliance UI using TypeScript 5.4, Tailwind CSS 3.4, and shadcn/ui.
  - Virtualized triage tables capable of rendering thousands of concurrent regulatory complaints with TanStack Table v8.
  - Real-time 21-day statutory SLA countdown calculation, warning thresholds (Day 7, Day 14, Day 19, Day 21), and escalation triggers.
  - Rich-text Action Taken Report (ATR) drafting editor supporting SEBI-mandated categorization and field validations.
  - Client-side Web Crypto SHA-256 calculation for all evidentiary uploads and PDF dossier packages.
  - Dual-officer maker-checker approval pipeline guarded by WebAuthn hardware token authentication.
  - Blockchain verification widget connecting to Hyperledger Besu RPC to confirm ATR timestamp hashes on `GrievanceLedger.sol` / `AuditAnchorRegistry.sol`.
  - Bidirectional integration with internal Grievance Gateway (Prompt 267), User Service (Prompt 201), and Audit Log Service (Prompt 218).
  - DPDP Act 2023 compliant PII redaction and dynamic forensic screen watermarking to prevent data exfiltration.
  - Generation of Section 65B certified legal evidence packages with SHA-256 checksum manifests.
- **Out of Scope / Handled Elsewhere:**
  - Direct upstream SOAP/REST protocol adapters connecting to SEBI SCORES 2.0, RBI Ombudsman, and SMART ODR servers (handled by the Go/Python Grievance Gateway in Prompt 267).
  - Public retail investor support ticket submission and live chat interfaces (handled in Prompt 608 and Prompt 910).
  - Core trade execution, matching engine operations, and real-time order cancellation (handled in Prompts 204 and 205).
  - Permissioned blockchain validator deployment, QBFT consensus configuration, and smart contract compilations (handled in Prompts 301 and 302).
  - Direct banking payment gateway dispute settlement disbursement and NACH mandate revocations (handled in Prompt 209).

## Technology to Use
- **Next.js 14 (App Router):** Server Components for secure data pre-fetching and strict layout boundary separation; Client Components (`'use client'`) for interactive triage queues, rich-text drafting, and cryptographic hashing.
- **TypeScript 5.4+:** Enforces complete compile-time type safety across regulatory schemas, complaint states, API responses, and Web3 event logs.
- **Tailwind CSS 3.4+ & shadcn/ui:** Institutional, high-contrast dark/light theme interface optimized for dense regulatory data layouts, compliant modals, and alert badges.
- **TanStack Table v8 & TanStack Virtual:** High-performance virtualized grid rendering thousands of historical and active dispute records without DOM degradation.
- **TipTap / ProseMirror:** Headless rich-text editor customized for legal drafting, enforcing strict formatting constraints, character counters, and SEBI-mandated structural sections.
- **PostgreSQL 16 & Prisma ORM:** Internal relational storage for local ATR drafts, officer assignments, ticket locking mechanisms, and SLA escalation tracking.
- **Viem 2.x:** Lightweight, type-safe Web3 client for interacting with Hyperledger Besu permissioned JSON-RPC nodes to read transaction receipts and verify on-chain anchors.
- **Web Crypto API:** Native in-browser cryptographic library used to calculate SHA-256 digests of uploaded evidence and exported ATR PDFs before transmission.
- **WebAuthn / FIDO2 API:** Hardware security token authentication (YubiKey 5 Series) enforced for high-privilege checker approvals and ATR submission seals.
- **Zod 3.23+:** Schema validation library enforcing strict input boundaries on ATR forms, assignment payloads, and external webhook structures.
- **jszip & file-saver:** Client-side generation of cryptographically signed compliance ZIP archives containing evidentiary dossiers and cryptographic manifests.

## Backend / Infra Touchpoints
- **Grievance Gateway (Prompt 267):** Primary backend service handling bidirectional synchronization with external regulatory platforms. Exposes REST endpoints (`/api/v1/grievance/*`) and WebSocket streams (`wss://grievance.growww.in/ws/v1/tickets`) for real-time ticket ingestion, state propagation, and ATR payload dispatch.
- **User Service (Prompt 201):** Interfaced via internal REST (`https://user.growww.in/api/v1/users/{id}`) to retrieve investor account profiles, KYC verification status, linked Demat account numbers, and masked PAN records.
- **Immutable Audit Log Service (Prompt 218):** Interfaced via gRPC and REST (`https://audit.growww.in/api/v1/audit/events`) to stream all ticket interactions, officer assignments, ATR revisions, and export downloads, receiving RFC 6962 Merkle inclusion proofs.
- **Hyperledger Besu Consortium Nodes (Prompt 302):** Connects via JSON-RPC (`https://rpc.besu.growww.in`) to inspect `GrievanceLedger.sol` and `AuditAnchorRegistry.sol` for verifying transaction receipts, block numbers, and validator consensus attestations.
- **S3-Compatible WORM Object Store (MinIO / AWS S3 with Object Lock):** Stores encrypted ATR PDF artifacts and evidentiary files in compliance mode (retention period: 8 years) to satisfy regulatory data retention mandates.
- **Redis 7.2 Cluster:** Used for distributed ticket locking (preventing dual-officer race conditions during drafting) and ephemeral caching of real-time SLA countdown states.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu private consortium network running QBFT (Quorum Byzantine Fault Tolerant) consensus with deterministic 2-second block times.
- **Smart Contract Interfaces:**
  - `GrievanceLedger.sol`:
    - `anchorAtrRecord(bytes32 complaintIdHash, bytes32 atrDocumentHash, uint64 regulatoryPortalCode, uint256 resolutionTimestamp) external returns (bytes32 txHash)`
    - `verifyAtrRecord(bytes32 complaintIdHash) external view returns (bytes32 atrDocumentHash, uint256 blockNumber, uint256 blockTimestamp, address anchorOfficer)`
  - `AuditAnchorRegistry.sol`:
    - `verifyMerkleProof(bytes32 root, bytes32 leaf, bytes32[] calldata proof) external view returns (bool)`
- **Cryptographic Evidentiary Flow:**
  1. Once the Action Taken Report is finalized by the Maker and approved by the Checker, the portal client compiles the report and attachments into an archival PDF/A-1b format.
  2. The browser computes the cryptographic digest:
     $$\text{atrDocumentHash} = \text{SHA-256}(\text{PDF\_A\_Bytes})$$
  3. The Checker co-signs the digest using their WebAuthn hardware token.
  4. The Grievance Gateway (Prompt 267) anchors the hash to `GrievanceLedger.sol` on Hyperledger Besu.
  5. The Next.js portal queries the Besu node using Viem, retrieves the transaction receipt, verifies block finality, and displays an interactive cryptographic badge showing block height, timestamp, and validator signatures.
- **Zero PII on Blockchain:** Under the Digital Personal Data Protection (DPDP) Act 2023, no personal information (complainant names, PAN, Aadhaar references, phone numbers, or monetary balances) is ever committed to the blockchain. All on-chain records consist strictly of blinded SHA-256 hashes (`bytes32`).

## Step-by-Step Build Instructions

1. **Scaffold Grievance Administration Route Structure & Layout:**
   - Create root directory `apps/growww_admin/grievance` using Next.js 14 App Router.
   - Configure root layout `layout.tsx` incorporating institutional header, navigation sidebar, active portal status bar, and forensic watermark shell.
   - Establish dedicated sub-routes:
     - `queue/page.tsx`: Central complaint triage and filtering queue.
     - `ticket/[id]/page.tsx`: Comprehensive dispute dossier and investigation view.
     - `ticket/[id]/draft-atr/page.tsx`: Rich-text SEBI SCORES 2.0 ATR drafting console.
     - `ticket/[id]/evidence/page.tsx`: Cryptographic evidence vault and document manager.
     - `ticket/[id]/review/page.tsx`: Dual-officer maker-checker approval console.
     - `analytics/sla/page.tsx`: Executive 21-day SLA monitoring and escalation heatmaps.
     - `audit/export/page.tsx`: Section 65B compliance dossier generation and export.

2. **Configure PostgreSQL Schema & Prisma ORM Layer:**
   - Define database models in `prisma/schema.prisma`: `GrievanceTicket`, `AtrDraft`, `EvidenceAttachment`, `SlaHistory`, `OfficerAssignment`, and `BlockchainAnchorRecord`.
   - Implement PostgreSQL database migration scripts ensuring foreign key constraints, composite indices on `(portalSource, status, slaDeadline)`, and full-text search indices on complaint summaries.
   - Configure connection pooling and edge-compatible Prisma client initialization.

3. **Implement Dynamic Anti-Exfiltration Forensic Watermarking:**
   - Build `GrievanceWatermark.tsx` rendering a tamper-evident, non-intrusive repeating canvas overlay across all pages.
   - Watermark tiles must dynamically display: Officer Badge ID, Workstation Public/Private IP, Active Session Nonce, and High-Resolution UTC Timestamp.
   - Attach DOM MutationObservers that immediately blank the screen and emit a security violation event if the watermark container is hidden, edited, or deleted via browser developer tools.

4. **Construct Virtualized Complaint Triage Queue (`ComplaintTriageQueue.tsx`):**
   - Integrate TanStack Table v8 with `@tanstack/react-virtual` to manage large complaint datasets efficiently.
   - Implement multi-column sorting, faceted filtering by `PortalSource` (SEBI SCORES, RBI, SMART ODR), `DisputeCategory`, `SlaUrgency`, and `AssignmentStatus`.
   - Build batch-triage actions: Batch Assignment, Severity Tagging, and Regulatory Category Remapping.
   - Provide visual indicators for ticket locks currently held by active legal officers.

5. **Develop Statutory 21-Day SLA Countdown Engine (`SlaCountdownWidget.tsx`):**
   - Implement dynamic SLA calculation logic tracking remaining calendar and business days against SEBI statutory 21-day limits.
   - Define state-driven color thresholds:
     - **Green (Normal):** $> 14$ days remaining.
     - **Amber (Warning):** $7 \text{ to } 14$ days remaining (triggers operational reminder).
     - **Red (Urgent):** $2 \text{ to } 7$ days remaining (triggers Legal Head escalation alert).
     - **Flashing Critical (Statutory Breach Imminent):** $< 48$ hours remaining (escalates directly to Principal Compliance Officer and Executive Board).
   - Display visual progress bars with milestone flags for Day 7 (Internal Assessment), Day 14 (Evidence Assembly), and Day 19 (Maker-Checker Review).

6. **Create Comprehensive Dispute Investigation Dossier (`ComplaintDetailViewer.tsx`):**
   - Build modular investigation panels presenting complainant details (masked per DPDP Act), initial grievance text, and uploaded external documents.
   - Integrate with User Service (Prompt 201) to display investor Demat holdings, trading account history, and historical customer support interactions.
   - Implement linked trade transaction inspector pulling execution logs, order book timestamps, and clearing settlement hashes corresponding to the disputed trades.

7. **Construct SEBI SCORES 2.0 Compliant ATR Editor (`AtrDraftingEditor.tsx`):**
   - Build a structured rich-text editor using TipTap/ProseMirror enforcing mandatory regulatory sections:
     - Section A: Nature of Grievance & Summary of Dispute.
     - Section B: Chronology of Factual Events & Internal Examination.
     - Section C: Factual Findings & Root Cause Analysis.
     - Section D: Corrective / Remedial Action Taken (including refund/restitution details).
     - Section E: Investor Communication & Grievance Redressal Status.
   - Implement real-time character count validation, mandatory field checks, and auto-save functionality to local PostgreSQL storage with distributed Redis locking.

8. **Build Evidence Vault & Client-Side Cryptographic Hashing (`EvidenceVaultUploader.tsx`):**
   - Implement secure drag-and-drop file upload supporting PDF, CSV, and PNG/JPEG formats up to 50 MB.
   - Compute the SHA-256 digest of each file directly in the browser using the Web Crypto API prior to transmission:
     $$\text{fileHash} = \text{Array.from}(\text{new Uint8Array}(\text{await crypto.subtle.digest('SHA-256', buffer})))$$
   - Request pre-signed S3 upload URLs from Grievance Gateway (Prompt 267) and upload directly to WORM-compliant storage.
   - Enforce client-side PDF sanitization (stripping JavaScript macros, embedded actions, and unapproved metadata).

9. **Construct Dual-Officer Maker-Checker Review Console (`DualOfficerReviewConsole.tsx`):**
   - Enforce strict maker-checker segregation of duties: the officer who drafts the ATR (Maker) cannot approve or submit it.
   - Render side-by-side comparison view displaying the original complaint text, attached evidentiary records, and the drafted ATR.
   - Checker console features: "Approve for Regulatory Submission", "Reject with Modification Comments", and "Request Additional Evidence".
   - Approval requires WebAuthn / FIDO2 hardware token biometric/PIN attestation, generating an unforgeable cryptographic approval signature.

10. **Build Hyperledger Besu Blockchain Proof Inspector (`BesuProofTimestampInspector.tsx`):**
    - Integrate Viem to query the consortium Besu RPC endpoint (`https://rpc.besu.growww.in`).
    - Query `GrievanceLedger.sol` using `verifyAtrRecord(complaintIdHash)`.
    - Render cryptographic verification card: Transaction Hash, Block Number, QBFT Block Timestamp, Signer Address, and Merkle Leaf Verification Status.
    - Provide deep links to the internal consortium Block Explorer.

11. **Develop Section 65B Legal Evidence Dossier Exporter (`AuditTrailExporter.tsx`):**
    - Implement export generator bundling the complete case dossier:
      - Certified Action Taken Report (PDF/A-1b).
      - Original complaint metadata and regulatory portal ingestion timestamps.
      - All uploaded supporting evidentiary files.
      - Besu blockchain transaction receipts and cryptographic inclusion proofs.
      - Immutable audit logs retrieved from Prompt 218.
    - Package files into a sealed ZIP archive containing a cryptographically signed `manifest.sha256` and Section 65B Indian Evidence Act certification statement.

12. **Configure Security Hardening, Content Security Policy, and Anti-Tampering:**
    - Apply strict HTTP response headers via Next.js middleware:
      - `Content-Security-Policy`: Disallow inline scripts (nonce enforced), restrict `connect-src` strictly to internal APIs and Besu RPC, disallow `frame-ancestors`.
      - `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`.
    - Implement print and screen-capture deterrence: apply CSS media query overrides (`@media print { body { display: none !important; } }`) and intercept copy/cut clipboard operations with warning modals and audit logging.

13. **Implement Automated Vitest Unit & Playwright E2E Test Suite:**
    - Vitest unit tests verifying SLA countdown calculations, SEBI ATR form schema validation, Web Crypto SHA-256 computations, and Merkle proof traversal.
    - Playwright end-to-end tests validating:
      - Ingestion and triage of complaints across multi-portal queues.
      - Accurate SLA urgency state transitions (Green -> Amber -> Red).
      - Maker drafting, file upload, and draft submission.
      - Checker review rejection with comments and subsequent approval with WebAuthn mock signatures.
      - Successful generation and verification of Section 65B ZIP export bundles.

## Interfaces / Contracts

### Complaint & Regulatory Portal Schemas
```typescript
export type RegulatoryPortalSource = 
  | 'SEBI_SCORES_2_0' 
  | 'RBI_OMBUDSMAN' 
  | 'SMART_ODR' 
  | 'INTERNAL_ESCALATION';

export type ComplaintStatus = 
  | 'NEW_UNASSIGNED' 
  | 'TRIAGED_IN_PROGRESS' 
  | 'ATR_DRAFTED' 
  | 'PENDING_CHECKER_APPROVAL' 
  | 'SUBMITTED_TO_REGULATOR' 
  | 'RESOLVED_CLOSED' 
  | 'REGULATOR_REOPENED' 
  | 'STATUTORY_ESCALATED';

export type DisputeCategory = 
  | 'UNAUTHORIZED_TRANSACTION' 
  | 'SETTLEMENT_DELAY' 
  | 'DIVIDEND_NON_CREDIT' 
  | 'REJECTED_KYC' 
  | 'ORDER_EXECUTION_FAILURE' 
  | 'MARGIN_SHORTFALL_PENALTY' 
  | 'TOKENIZED_ASSET_REDEMPTION' 
  | 'OTHER_REGULATORY_DISPUTE';

export type SlaUrgencyLevel = 
  | 'NORMAL_GREEN' 
  | 'WARNING_AMBER' 
  | 'CRITICAL_RED' 
  | 'BREACHED_STATUTORY';

export interface SlaStatusRecord {
  statutoryDaysTotal: 21;
  calendarDaysElapsed: number;
  calendarDaysRemaining: number;
  businessDaysRemaining: number;
  receivedAtUtc: string;
  statutoryDeadlineUtc: string;
  urgencyLevel: SlaUrgencyLevel;
  isEscalatedToPco: boolean;
}

export interface ComplaintRecord {
  complaintId: string;
  externalReferenceNo: string; // e.g., "SEBIE/MH24/0001294/1"
  portalSource: RegulatoryPortalSource;
  complainantNameMasked: string; // "R**** K****"
  panMasked: string; // "ABCDE****F"
  dematAccountNumberMasked: string; // "12081600********"
  registeredEmailMasked: string; // "r****@gmail.com"
  disputeCategory: DisputeCategory;
  claimAmountPaise: bigint; // Stored in smallest currency unit
  disputedTransactionIds: string[];
  dateReceived: string;
  sla: SlaStatusRecord;
  assignedOfficerId?: string;
  assignedOfficerName?: string;
  status: ComplaintStatus;
  isLockedForEditing: boolean;
  lockedByOfficerId?: string;
  lockExpiresAt?: string;
  createdAt: string;
  updatedAt: string;
}
```

### Action Taken Report (ATR) & Evidence Schemas
```typescript
export interface EvidenceFileRecord {
  fileId: string;
  complaintId: string;
  fileName: string;
  fileSizeBytes: number;
  mimeType: 'application/pdf' | 'text/csv' | 'image/png' | 'image/jpeg';
  sha256Hash: string; // 64-character lowercase hex
  s3WormUri: string;
  antiVirusScanStatus: 'CLEAN' | 'INFECTED' | 'SCANNING';
  uploadedAt: string;
  uploadedByOfficerBadgeId: string;
}

export interface AtrDraftPayload {
  complaintId: string;
  disputeCategory: DisputeCategory;
  rootCauseAnalysis: string;
  factualChronology: string;
  investigationFindings: string;
  correctiveActionTaken: string;
  financialRestitutionPaise: bigint;
  restitutionTransactionRef?: string;
  investorCommunicationDate: string;
  isRedressedInFull: boolean;
  attachedEvidenceIds: string[];
  makerNotes?: string;
}

export interface AtrRecord {
  atrId: string;
  complaintId: string;
  version: number;
  content: AtrDraftPayload;
  compiledPdfSha256: string;
  makerOfficerBadgeId: string;
  makerSubmittedAt: string;
  checkerOfficerBadgeId?: string;
  checkerReviewedAt?: string;
  checkerSignature?: string; // WebAuthn assertion signature
  status: 'DRAFT' | 'PENDING_APPROVAL' | 'APPROVED' | 'REJECTED';
  rejectionComments?: string;
}

export interface BlockchainAnchorRecord {
  complaintIdHash: `0x${string}`;
  atrDocumentHash: `0x${string}`;
  besuTxHash: `0x${string}`;
  besuBlockNumber: number;
  besuBlockTimestamp: number;
  qbftValidatorCount: number;
  anchoredByOfficerAddress: `0x${string}`;
  isVerifiedOnChain: boolean;
}
```

### Maker-Checker & Evidence Export Schemas
```typescript
export interface MakerSubmitRequest {
  complaintId: string;
  atrContent: AtrDraftPayload;
  compiledPdfBase64: string;
  evidenceFileIds: string[];
}

export interface CheckerReviewRequest {
  complaintId: string;
  action: 'APPROVE_AND_SUBMIT' | 'REJECT_FOR_REVISION';
  rejectionReason?: string;
  webAuthnAssertion: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
}

export interface Section65BCertificate {
  certificateId: string;
  complaintId: string;
  systemName: string;
  serverWorkstationIp: string;
  hashAlgorithm: 'SHA-256';
  compiledArchiveSha256: string;
  certifyingOfficerBadgeId: string;
  certifiedTimestampUtc: string;
  legalDeclarationText: string;
}

export interface EvidenceDossierExportResponse {
  complaintId: string;
  exportArchiveUrl: string;
  archiveSha256: string;
  section65BCertificate: Section65BCertificate;
  blockchainProof: BlockchainAnchorRecord;
}
```

### REST API Endpoints
```text
# Complaint Triage & Queue Management
GET    /api/v1/grievance/complaints                    (List complaints with filtering, sorting, pagination)
GET    /api/v1/grievance/complaints/{id}               (Fetch comprehensive complaint dossier)
POST   /api/v1/grievance/complaints/{id}/assign        (Assign complaint to legal officer)
POST   /api/v1/grievance/complaints/{id}/lock          (Acquire editing lock on complaint)
DELETE /api/v1/grievance/complaints/{id}/lock          (Release editing lock)

# Evidence Vault & Cryptographic Verification
POST   /api/v1/grievance/complaints/{id}/evidence/presigned-url  (Request pre-signed S3 WORM upload URL)
POST   /api/v1/grievance/complaints/{id}/evidence/confirm        (Verify SHA-256 and confirm upload)
GET    /api/v1/grievance/complaints/{id}/evidence                (List attached evidence records)

# Action Taken Report (ATR) Drafting & Dual-Officer Review
GET    /api/v1/grievance/complaints/{id}/atr/draft     (Retrieve active ATR draft)
PUT    /api/v1/grievance/complaints/{id}/atr/draft     (Update and auto-save ATR draft)
POST   /api/v1/grievance/complaints/{id}/maker/submit  (Maker submits completed ATR for review)
POST   /api/v1/grievance/complaints/{id}/checker/review (Checker approves/rejects with WebAuthn)

# Blockchain Verification & Compliance Export
GET    /api/v1/grievance/complaints/{id}/blockchain-proof (Query Hyperledger Besu anchor receipt)
POST   /api/v1/grievance/complaints/{id}/export-dossier    (Generate Section 65B certified ZIP archive)
GET    /api/v1/grievance/analytics/sla-metrics            (Aggregate 21-day statutory SLA analytics)
```

## Security & Compliance Notes
- **SEBI Statutory 21-Day Escalation Mandate:** The 21-day SLA is an absolute statutory ceiling governed by SEBI circular `SEBI/HO/OIAE/OIAE_IAD-1/P/CIR/2023/0000000163`. The portal incorporates hard automated alert triggers at:
  - Day 7: Operational notification dispatched to GRO for preliminary review.
  - Day 14: Mid-term alert issued to Legal Counsel demanding evidence assembly.
  - Day 19: High-priority statutory escalation to Principal Compliance Officer for mandatory checker sign-off.
  - Day 21: Red alert and automatic dispatch of regulatory risk escalation notification to the Board of Directors and Audit Committee.
- **Dual-Officer Maker-Checker Enforcement:** Separation of duties is strictly enforced. The legal officer who authors the ATR (`makerOfficerBadgeId`) cannot act as the approving officer (`checkerOfficerBadgeId`). Checker approvals mandate hardware token authentication via WebAuthn/FIDO2 (YubiKey 5 Series). Software OTPs, passwords, or single-officer overrides are strictly rejected.
- **Digital Personal Data Protection (DPDP) Act 2023 Compliance:** Retail investor PII (PAN, Demat account, phone number, physical address, and bank credentials) is masked by default on all dashboard views. Unmasked inspection requires a logged, time-limited justification recorded directly in the Immutable Audit Log Service (Prompt 218). No PII is ever written to the Hyperledger Besu consortium blockchain.
- **Judicial Proof Preservation (Section 65B Indian Evidence Act / Section 63 BSA 2023):** All evidence uploaded to the portal is immediately hashed client-side with SHA-256, transmitted over TLS 1.3, and committed to S3 WORM storage with Object Lock enabled. Any export produces a cryptographically sealed ZIP archive containing a formal Section 65B digital evidence certificate, the blockchain anchor transaction receipt, and full audit logs establishing an unbroken chain of custody.
- **Dynamic Forensic Anti-Exfiltration Watermarking:** To prevent unauthorized leaks via physical screen photography or digital screen capture, all grievance views feature a semi-transparent, non-intrusive Canvas watermark containing the officer's badge ID, client IP address, workstation timestamp, and session hash. Screen tampering via CSS or DOM deletion immediately locks the viewport and records a security incident.
- **Content Security Policy (CSP) & Reverse Proxy Hardening:**
  ```http
  Content-Security-Policy: default-src 'none'; script-src 'self' 'nonce-{RANDOM}'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' https://grievance.growww.in wss://grievance.growww.in https://audit.growww.in https://rpc.besu.growww.in; font-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'none'; form-action 'self';
  ```
- **Strict Anti-Tampering & Clickjacking Protection:** `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, and strict `Referrer-Policy: strict-origin-when-cross-origin` are enforced on all routes.

## Acceptance Criteria
- [ ] Next.js 14 grievance management portal scaffolds cleanly under `apps/growww_admin/grievance` using App Router.
- [ ] Role-based access control (RBAC) restricts access strictly to authorized compliance, legal, and GRO officers.
- [ ] Dynamic forensic watermark renders across all viewports; DOM modification triggers immediate screen blanking and an audit alert.
- [ ] Complaint triage queue virtualizes thousands of complaints with sub-second filtering by source (SEBI SCORES 2.0, RBI Ombudsman, SMART ODR), urgency, and status.
- [ ] Statutory SLA countdown widget accurately tracks elapsed and remaining calendar and business days against the 21-day legal limit.
- [ ] Urgency indicators dynamically transition through Green (>14d), Amber (7-14d), Red (<7d), and Flashing Critical (<48h).
- [ ] Dispute dossier viewer displays linked investor profiles with masked PII and correlates disputed trades with matching engine execution IDs.
- [ ] SEBI SCORES 2.0 rich-text ATR editor enforces all mandated sections (A through E) and provides real-time character count and validation checks.
- [ ] Evidence vault calculates SHA-256 digests in-browser using Web Crypto API prior to uploading to S3 WORM storage.
- [ ] PDF uploads are sanitized, validated against malicious payloads, and linked to their cryptographic digests.
- [ ] Maker-checker workflow strictly bars self-approval and enforces WebAuthn FIDO2 hardware key authentication on checker approvals.
- [ ] Finalized ATR document hash (`bytes32`) is anchored to `GrievanceLedger.sol` on Hyperledger Besu.
- [ ] Web3 verification widget queries Besu JSON-RPC and confirms on-chain block receipts, block timestamps, and Merkle validity.
- [ ] Section 65B certified compliance package generator outputs a sealed ZIP archive containing the certified ATR, raw evidence, SHA-256 manifest, and on-chain proofs.
- [ ] All user interactions (ticket views, draft revisions, evidence downloads, maker/checker actions) are logged to Audit Log Service (Prompt 218).
- [ ] Vitest unit tests achieve $>90\%$ code coverage across SLA countdown calculations, ATR schema validation, and SHA-256 hashing routines.
- [ ] Playwright E2E tests validate complete complaint lifecycle: ingestion -> triage -> ATR drafting -> maker submission -> checker WebAuthn approval -> Besu verification -> Section 65B export.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 201 (User Service & Identity Management).
  - Prompt 218 (Immutable Audit Log Service).
  - Prompt 267 (Regulatory Grievance Gateway & SCORES 2.0 Protocol Adapter).
  - Prompt 302 (Hyperledger Besu Private Consortium Network Architecture).
  - Prompt 601 (Next.js Investor Web App Scaffolding & Foundation).
- **Parallel Tasks:**
  - Prompt 607 (Admin Console: Regulatory Reporting & Audit Export Portal).
  - Prompt 613 (Next.js 14 Regulatory Audit & SEBI/IFSCA Supervisory Dashboard).
  - Prompt 910 (Customer Support, Dispute Resolution & Grievance Redressal).
- **Downstream Blockers:**
  - Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios).
  - Prompt 907 (Regulatory Sandbox Pilot Launch & Supervisory Handover).
