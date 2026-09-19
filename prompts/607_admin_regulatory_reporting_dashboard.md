# 607 - Admin Console: Regulatory Reporting & Audit Export Portal

## Purpose
Provides compliance officers, principal regulatory officers, and internal auditors with a dedicated back-office administration portal (`apps/growww_admin/app/(dashboard)/regulatory/`) for generating, validating, cryptographically signing, and exporting statutory compliance reports. Supports mandatory filings required by the Securities and Exchange Board of India (SEBI), Reserve Bank of India (RBI), Financial Intelligence Unit - India (FIU-IND), and International Financial Services Centres Authority (IFSCA / GIFT City), with integrated on-chain cryptographic settlement proofs embedded directly into regulatory audit packages.

## What You Are Building
- Regulatory reporting console (`apps/growww_admin/app/(dashboard)/regulatory/`) with categorized jurisdiction hubs (SEBI, RBI, FIU-IND, IFSCA).
- Statutory Report Generator Catalog:
 - SEBI Daily Transaction & Settlement Log (DvP audit trails, trade execution timestamps).
 - FIU-IND Suspicious Transaction Report (STR) & Cash Transaction Report (CTR) filing generator.
 - RBI Foreign Inward Remittance Certificate (FIRC) & Liberalised Remittance Scheme (LRS) logs.
 - SEBI Regulatory Sandbox Metrics & System Performance Report.
 - Daily Depository Demat vs On-Chain Token Reconciliation Statement.
- High-volume interactive report preview data table with virtualized scrolling and schema validation error highlighter.
- Compliance Officer PKI Digital Signature / Attestation signing modal.
- Encrypted Audit Package Export Manager bundling generated CSV/XML/XBRL reports, cryptographic checksums, and on-chain verification manifests into signed ZIP archives.

## Scope Boundaries
- **In Scope:**
 - Admin regulatory reporting UI, filing catalog, and date/entity filter controls.
 - High-performance tabular data preview and pre-submission validation engine.
 - Client-side data integrity verification and digital attestation workflows.
 - Audit package bundler (ZIP generator with SHA-256 manifest).
 - Filing history tracking and regulatory submission acknowledgment logging.
- **Out of Scope / Handled Elsewhere:**
 - Backend regulatory reporting batch generation engine (Prompt 216).
 - Direct machine-to-machine API gateway to SEBI/FIU portals (Prompt 216).
 - Transaction monitoring and automated suspicious activity detection (Prompt 704).
 - Regulatory sandbox pathway definition (Prompt 003).

## Technology to Use
- **Next.js 14 App Router, React 18/19, TypeScript 5.4+:** Provides secure, authenticated server-rendered views for internal compliance workflows.
- **`@tanstack/react-table` (v8) & `@tanstack/react-virtual`:** Renders massive regulatory data tables (10,000+ transaction rows) with fluid virtualized scrolling, multi-column sorting, and column visibility toggling.
- **`jszip` & `file-saver`:** Enables fast client-side compression and bundling of multi-file regulatory export packages with SHA-256 integrity checksum files.
- **`date-fns`:** Handles Indian financial year calendars, quarterly filing cutoff dates, and T+1 settlement window calculations.
- **shadcn/ui (Radix UI) & Tailwind CSS:** Clean enterprise interface with status alerts and form controls.

## Backend / Infra Touchpoints
- **Regulatory Reporting Microservice (Prompt 216):** Interacts via REST API (`/api/v1/admin/reports/*`) to trigger async report generation and fetch structured tabular datasets.
- **Admin Service (Prompt 217):** Manages user session permissions and filing authorization state.
- **Audit Log Service (Prompt 218):** Logs all report generation, preview, download, and export events with full analyst attribution.

## Blockchain Interaction
- **On-Chain Settlement Proof Extraction:** Extracts immutable on-chain transaction hashes, block numbers, and smart contract event logs (`SettlementDvP.SettlementCompleted`, `DigitalSecurityToken.Transfer`) to embed as cryptographically verifiable audit proofs in SEBI trade logs.
- **Attestation Registry Check:** Cross-references published `ProofOfReserveRegistry.sol` roots to validate that reported depository balances match on-chain attestations for the selected reporting period.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `SettlementDvP.sol` (Queries settlement receipts for trade reconciliation filings).
 - `ComplianceRegistry.sol` (Queries historical investor whitelist/freeze status for AML filings).
 - `ProofOfReserveRegistry.sol` (Queries daily reserve attestations for depository audit filings).

## Step-by-Step Build Instructions
1. Scaffold admin regulatory reporting dashboard route (`apps/growww_admin/app/(dashboard)/regulatory/`) with navigation tabs for SEBI, RBI, FIU-IND, and IFSCA.
2. Build Statutory Report Catalog card grid showcasing available report templates with filing frequency, statutory authority, and last generated date.
3. Implement Report Generation Filter Drawer with controls for Date Range (Calendar picker), Asset Category, Entity (Domestic Regulated Entity vs GIFT City Gateway), and File Format (CSV, XML, XBRL, JSON).
4. Implement Async Report Generation Trigger: Dispatches job to Reporting Service and displays live polling progress indicator.
5. Build Tabular Report Preview Component using `@tanstack/react-table` and `@tanstack/react-virtual` supporting pagination, column sorting, and instant search.
6. Implement Pre-Filing Validation Engine that parses preview rows against regulatory schema rules (e.g., verifying 10-digit PAN format, 12-digit ISIN format, positive turnover amounts) and highlights errors in red.
7. Build SEBI Regulatory Sandbox Metrics Module displaying active sandbox investor count, total trading turnover, average latency, and incident counters.
8. Build FIU-IND STR/CTR Module displaying flagged suspicious transactions with risk scores, reason codes, and analyst review attachments.
9. Implement Digital Signature & Attestation Modal allowing Principal Compliance Officer to review report summary, enter compliance statement, and apply digital signature with WebAuthn/TOTP.
10. Build Audit Package Exporter: Bundles report files, cryptographic SHA-256 manifest (`manifest.json`), on-chain settlement receipts, and signed compliance statement into an encrypted ZIP file via `jszip`.
11. Implement Filing History & Status Tracker table showing Report Name, Reporting Period, Submitting Officer, Digital Seal ID, Filing Acknowledgment Ref, and Status (Draft, Approved, Exported, Filed).
12. Write unit tests for client-side validation rules and end-to-end Playwright tests verifying report generation and ZIP download.

## Interfaces / Contracts
```typescript
export type RegulatoryAuthority = 'SEBI' | 'RBI' | 'FIU_IND' | 'IFSCA';

export interface RegulatoryReportMetadata {
  reportId: string;
  reportCode: 'SEBI_DAILY_DVP_LOG' | 'FIU_STR_MONTHLY' | 'RBI_LRS_QUARTERLY' | 'SEBI_SANDBOX_METRICS' | 'POR_DAILY_RECON';
  title: string;
  authority: RegulatoryAuthority;
  filingFrequency: 'DAILY' | 'MONTHLY' | 'QUARTERLY' | 'ANNUAL' | 'AD_HOC';
  lastGeneratedAt?: string;
  schemaVersion: string;
}

export interface ReportGenerationRequest {
  reportCode: string;
  startDate: string;
  endDate: string;
  entityScope: 'DOMESTIC' | 'GIFT_CITY' | 'CONSOLIDATED';
  format: 'CSV' | 'XML' | 'XBRL' | 'JSON';
  includeOnChainProofs: boolean;
}

export interface ReportValidationIssue {
  rowNumber: number;
  columnName: string;
  currentValue: string;
  errorMessage: string;
  severity: 'WARNING' | 'BLOCKING_ERROR';
}

export interface SebiDailyTradeLogRecord {
  tradeId: string;
  executionTimestamp: string;
  investorUcc: string; // Unique Client Code
  panMasked: string;
  isin: string;
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: string;
  priceInr: string;
  grossAmountInr: string;
  platformProfitFee: string;
  settlementDvPTxHash: `0x${string}`;
  besuBlockNumber: number;
  custodianRef: string;
}

export interface AuditPackageManifest {
  packageId: string;
  generatedAt: string;
  reportCode: string;
  files: Array<{
    fileName: string;
    sha256Checksum: string;
    fileSizeBytes: number;
  }>;
  complianceOfficer: {
    name: string;
    empId: string;
    digitalSignatureSeal: string;
  };
  onChainBlockReference: {
    startBlock: number;
    endBlock: number;
    besuChainId: number;
  };
}
```

## Security & Compliance Notes
- **SEBI Cyber Resilience Compliance:** All exported report files enforce strict encryption (AES-256) and SHA-256 integrity checksum manifests.
- **Confidentiality & Access Control:** Access restricted to authorized Compliance Officers and Internal Auditors via strict RBAC (Prompt 702).
- **Immutable Audit Trail:** Every report creation, preview, data filter modification, and file download is permanently recorded in the Audit Log Service (Prompt 218).
- **PII Redaction Rules:** Exports destined for public or external auditors adhere to mandatory Aadhaar masking and PAN redaction standards.

## Acceptance Criteria
- [ ] Regulatory reporting catalog renders all required SEBI, RBI, FIU-IND, and IFSCA report templates.
- [ ] Report generation filter triggers async backend job and previews 1,000+ data rows with smooth virtualized scrolling.
- [ ] Pre-filing validation engine accurately flags invalid records with inline visual alerts and disables export until resolved.
- [ ] On-chain DvP transaction hashes and block numbers are embedded accurately in SEBI trade log records.
- [ ] Principal Compliance Officer digital attestation captures timestamped approval and WebAuthn MFA.
- [ ] Audit Package Exporter generates a valid ZIP archive containing report data, SHA-256 checksum manifest, and signature seal.
- [ ] Filing history table logs complete submission lifecycle with filing acknowledgment tracking.
- [ ] Playwright E2E tests cover filter selection, report preview, validation error highlighting, and export bundling.

## Suggested Order / Dependencies
- **Prerequisites:** 003 (Regulatory Pathway), 216 (Reporting Service), 217 (Admin Service), 218 (Audit Log Service), 601 (Web Scaffolding), 702 (IAM).
- **Direct Successors / Parallel:** 604 (KYC Review), 605 (Approvals UI), 606 (Proof of Reserve).
