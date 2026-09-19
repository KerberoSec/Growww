# 618 - Next.js 14 P2P Dispute Arbitration & Compliance Desk

## Purpose
Peer-to-peer (P2P) fiat-to-digital asset transactions represent a critical bridge for liquidity onboarding, yet they introduce profound operational, counterparty, and regulatory risks. Under ADR-0039 (Peer-to-Peer Automated Escrow Lockbox & Dispute Arbitration) and RUNBOOK-27 (P2P Fiat Dispute Maker-Checker Arbitration), the Growww platform employs automated smart contract escrow lockboxes on `P2PEscrow.sol` to protect market participants. However, when contested trades arise (such as fraudulent claims of fiat remittance, uncredited bank transfers, third-party payment impersonation, altered bank payment receipts, or buyer payment timer expirations following off-ramp transfers), algorithmic resolution alone is insufficient. Adjudication requires human oversight executed under institutional-grade security controls.

This prompt specifies the architecture, user experience, and technical implementation of the **Next.js 14 P2P Dispute Arbitration & Compliance Desk** (`apps/growww_admin/p2p-disputes`). Operating within the administrative control perimeter, this high-security operator console empowers compliance officers, fraud investigators, and Money Laundering Reporting Officers (MLROs) to adjudicate contested P2P trades. The platform provides split-screen comparative forensic investigation, automated Open Banking / Account Aggregator Unique Transaction Reference (UTR) settlement verification, tamper-evident PDF bank statement analysis, and dynamic KYC name-matching invariance checks. Crucially, fund movement is governed by a mandatory dual-control (Maker-Checker) workflow that generates cryptographic EIP-712 authorization payloads to execute 2-of-3 multisig arbitration releases or refunds on `P2PEscrow.sol` in conjunction with the CloudHSM Key Lifecycle & Signing Daemon (Prompt 717). The console also features direct integration with Financial Intelligence Unit - India (FIU-IND) Suspicious Transaction Report (STR) drafting pipelines to uphold the Prevention of Money Laundering Act (PMLA) and Digital Personal Data Protection (DPDP) Act 2023.

## What You Are Building
A mission-critical Next.js 14 administrative application module located at `apps/growww_admin/p2p-disputes`, delivering:
- `P2PDisputeConsoleShell`: Hardened operator shell enforcing Mutual TLS (mTLS), hardware WebAuthn/FIDO2 multi-factor authentication, dynamic session collision prevention (Redis distributed locks), and anti-exfiltration forensic screen watermarking displaying officer ID, IP address, and session timestamp.
- `DisputeQueueDesk`: Real-time virtualized dispute queue streaming active P2P arbitration tickets via WebSockets, categorized by SLA countdown urgency, fiat volume, asset pair (e.g., USDT/INR, USDC/INR), risk score, and dispute reason (`PAYMENT_NOT_RECEIVED`, `NAME_MISMATCH`, `INCORRECT_AMOUNT`, `FRAUDULENT_CHARGEBACK_THREAT`, `ACCOUNT_FROZEN`).
- `SplitScreenDisputeInspector`: Side-by-side comparative inspection workspace presenting:
  - **Buyer Dossier:** Verified legal name, masked PAN/Aadhaar status, registered bank account/VPA, submitted payment proof (bank statement PDF, UPI transaction reference, raw UTR), device fingerprint, IP geolocation, and trade completion history.
  - **Seller Dossier:** Verified legal name, registered receiving bank account/VPA, dispute claims, uncredited balance attestation, bank statement export, and historical dispute frequency metrics.
- `BankStatementVerificationTools`: Automated and forensic document verification suite:
  - Optical Character Recognition (OCR) text layer visualizer highlighting transaction timestamp, UTR, remit amount, and counterparty account details.
  - Document tampering detector flagging font inconsistencies, modified PDF metadata, missing bank digital signatures, and rasterized copy-paste alterations.
  - Open Banking / Account Aggregator (AA) automated UTR lookup reconciling bank network status directly against central clearing clearinghouses (NPCI / RBI IMPS / NEFT).
  - Strict KYC Name Matching Invariant analyzer computing Levenshtein distance, phonetic matching (Double Metaphone), and PAN identity verification between the fiat sender and the buyer account.
- `MakerCheckerArbitrationWorkflow`: Dual-operator governance console enforcing strict separation of duties:
  - **Maker Flow:** Compliance Officer reviews evidence, formulates an adjudication recommendation (`RELEASE_ESCROW_TO_BUYER`, `REFUND_ESCROW_TO_SELLER`, `CANCEL_AND_HOLD`), enters statutory case notes, and signs the proposal via WebAuthn hardware key.
  - **Checker Flow:** Independent Senior Compliance Officer / MLRO inspects the adjudication package, independently validates bank verification receipts, and either approves or rejects the determination with WebAuthn hardware key co-signing.
- `MultisigArbitrationTerminal`: Web3 cryptographic signing bridge utilizing Viem 2.x to construct EIP-712 typed arbitration payloads and orchestrate 2-of-3 threshold multisig execution on `P2PEscrow.sol` (combining Maker/Checker officer signatures with the platform CloudHSM co-signer from Prompt 717).
- `FiuSuspiciousActivityFlagging`: Embedded regulatory compliance terminal enabling operators to trigger FIU-IND suspicious activity red flags, compile PMLA audit packages, and queue automated STR drafts to the Regulatory Reporting Dispatcher (Prompt 233).
- `LiveEvidenceChatTimeline`: Read-only, tamper-proof audit visualizer displaying the end-to-end P2P chat history between buyer and seller, escrow state milestones, system timers, and immutable operational audit entries.

## Scope Boundaries
- **In Scope:**
  - Complete Next.js 14 App Router portal located under `apps/growww_admin/p2p-disputes`.
  - Real-time WebSocket subscriptions for dispute ticket assignments, escrow state transitions, and operator presence locking.
  - Split-screen comparison UI rendering buyer and seller KYC profiles, payment evidence, and historical trade statistics.
  - Interactive PDF bank statement viewer with forensic metadata inspection, OCR text overlay, and zoom/pan controls.
  - Automated UTR reconciliation query visualizer interfacing with banking gateways.
  - Strict KYC name matching invariant calculation and phonetic distance badge indicators.
  - Maker-Checker multi-stage decision submission, review, rejection, and approval workflows.
  - Client-side EIP-712 structured data hashing and WebAuthn / hardware-wallet transaction co-signing via Viem.
  - FIU-IND red flag generation, SAR/STR dossier compilation, and dispatch triggers.
  - Anti-exfiltration dynamic forensic watermarking and DOM tamper-detection observers.
  - Full immutable operator telemetry and audit logging dispatched to Kafka (`audit.admin.p2p.v1`).
- **Out of Scope / Handled Elsewhere:**
  - P2P order matching engine, maker/taker order book creation, and trader-facing communication (handled by P2P Escrow Service - Prompt 269).
  - Mobile client P2P trading interface for end-user buyers and sellers (handled by Flutter apps - Prompts 500-series).
  - Direct banking Account Aggregator integration and clearinghouse network node hosting (handled by Banking Integration Service).
  - Physical CloudHSM appliance provisioning, partition initialization, and FIPS 140-2 Level 3 root key generation (handled by Prompt 717).
  - General administrative back-office user management, accounting books, and exchange-wide ledger reconciliations (handled by Prompts 217 & 218).
  - Smart contract consensus deployment and core ledger node operations (handled by Prompts 302 & 306).

## Technology to Use
- **Next.js 14 (App Router):** Server-side rendered layouts and React Server Components for authenticated data prefetching, paired with strict `'use client'` interactive inspection desks.
- **TypeScript 5.4+:** Strict type checking across domain entities, banking schemas, EIP-712 payloads, and REST/gRPC responses.
- **Tailwind CSS 3.4+ & shadcn/ui:** High-contrast institutional dark-mode UI library with accessible modals, collapsible panels, and data-dense tables built on Radix UI primitives.
- **Viem 2.x:** Type-safe Ethereum / Hyperledger Besu Web3 client for client-side EIP-712 typed data hashing, cryptographic signature validation, and multisig smart contract interactions on `P2PEscrow.sol`.
- **WebAuthn / FIDO2 API:** Cryptographic hardware token attestation (YubiKey 5 Series) required for Maker proposal submissions and Checker authorization sign-offs.
- **WebSockets (Native WS / STOMP):** Low-latency bi-directional event streaming for live dispute updates, escrow timers, and operator presence tracking.
- **PDF.js (`pdfjs-dist` / `react-pdf`):** Sandboxed, client-side PDF document visualizer supporting high-resolution zoom, text layer selection, and forensic metadata inspection.
- **TanStack Table (v8) & TanStack Virtual:** High-performance virtualized dispute queue rendering thousands of tickets with zero DOM degradation.
- **Zod:** Runtime schema validation for all incoming API payloads, adjudication actions, and EIP-712 typed data envelopes.
- **Lucide React:** Standardized institutional icons for financial and security workflows.

## Backend / Infra Touchpoints
- **P2P Escrow Service (Prompt 269):** Primary upstream service providing dispute metadata, locking disputed escrows, serving buyer/seller payment proofs, and receiving arbitration resolution webhooks (`/api/v1/p2p/disputes/{id}`).
- **Admin Service (Prompt 217):** Manages administrative authentication, role-based access control (RBAC), Maker-Checker task lifecycle, and operator presence locks (`/api/v1/admin/maker-checker/tasks`).
- **CloudHSM Key Lifecycle & Signing Daemon (Prompt 717):** Provides the platform co-signing service for 2-of-3 multisig arbitration execution via REST/mTLS endpoint (`/api/v1/signer/p2p-arbitration`).
- **Banking Aggregator & UTR Gateway:** Supplies live bank clearance data, UTR timestamp confirmation, remitter bank name, and remitter account number matches.
- **Regulatory Reporting Dispatcher (Prompt 233):** Ingests FIU-IND STR alerts and structured compliance dossiers for statutory regulatory filing.
- **Immutable Audit Log Service (Prompt 218):** Commits all operator actions, inspection durations, evidence downloads, and decision payloads to an immutable audit ledger.
- **Kafka Topics:**
  - `p2p.disputes.v1`: Ingestion of dispute state transitions (`DISPUTE_OPENED`, `EVIDENCE_UPLOADED`, `RESOLVED`, `CANCELLED`).
  - `admin.maker_checker.p2p.v1`: Maker proposal creations, checker rejections, and checker approvals.
  - `fiu.str.flags.v1`: High-priority suspicious activity indicators requiring MLRO intervention.
  - `audit.admin.p2p.v1`: Detailed clickstream and verification audit logs.
- **Redis 7.x Cluster:** Ephemeral operator distributed locking (preventing two compliance officers from working the same dispute simultaneously) and real-time operator heartbeat tracking.
- **AWS S3 / WORM Object Storage:** Cryptographically secured, encrypted storage for buyer/seller payment receipts, bank statement PDFs, and dispute attachments.

## Blockchain Interaction
- **Smart Contract Target:** `P2PEscrow.sol` deployed on the Hyperledger Besu consortium ledger (or EVM L2 settlement network).
- **Multisig Arbitration Scheme:** 2-of-3 threshold signature validation:
  - **Key 1 (Buyer/Seller):** Normal non-contested trade completion requires buyer or seller authorization.
  - **Key 2 (Operator Maker):** Authorized Compliance Officer ECDSA signing key (or WebAuthn hardware-bound key).
  - **Key 3 (Operator Checker or Platform CloudHSM):** Authorized Supervisor / MLRO ECDSA key or the automated FIPS 140-2 Level 3 CloudHSM key (Prompt 717).
- **On-Chain Dispute Execution Methods:**
  - `executeArbitrationRelease(bytes32 escrowId, address recipient, uint256 amount, bytes32 disputeHash, bytes[] signatures)`: Unlocks collateral from the escrow lockbox and transfers it to the buyer when payment has been definitively verified.
  - `executeArbitrationRefund(bytes32 escrowId, address seller, bytes32 disputeHash, bytes[] signatures)`: Returns locked collateral to the seller when the buyer fails to prove payment or submitted fraudulent receipts.
  - `executeSplitForfeiture(bytes32 escrowId, address buyer, address seller, uint256 buyerAmount, uint256 sellerAmount, uint256 penaltyAmount, bytes32 disputeHash, bytes[] signatures)`: Resolves complex disputes with partial settlement and penalty deductions for terms-of-service violations.
- **EIP-712 Structured Data Domain:**
  ```text
  Domain: {
    name: "GrowwwP2PEscrow",
    version: "1",
    chainId: 1337,
    verifyingContract: "0x3F8B...Escrow"
  }
  Type: ArbitrationResolution(
    bytes32 escrowId,
    address recipient,
    uint256 amount,
    uint256 nonce,
    uint256 deadline,
    bytes32 disputeHash,
    bytes32 justificationHash
  )
  ```
- **Zero PII on Blockchain:** The smart contract stores only anonymous cryptographic identifiers: `escrowId`, Ethereum wallet addresses (`0x...`), token amounts, and SHA-256 hashes of the adjudication dossier (`disputeHash`, `justificationHash`). No fiat bank account numbers, UTRs, names, or contact info are committed to the distributed ledger.

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Route Architecture & Layout:**
   - Create root dashboard layout at `apps/growww_admin/p2p-disputes/layout.tsx` incorporating institutional dark theme, header navigation, operator status badge, and global forensic watermarking.
   - Configure modular sub-routes:
     - `queue/page.tsx`: Real-time virtualized dispute ticket queue.
     - `case/[id]/page.tsx`: Comprehensive split-screen dispute inspection and adjudication console.
     - `checker/page.tsx`: Checker queue for pending adjudication approvals.
     - `fiu-str/page.tsx`: FIU-IND suspicious activity reporting and audit terminal.
     - `audit/page.tsx`: Immutable dispute resolution log and on-chain transaction verifier.

2. **Implement Operator Authentication, RBAC, and Redis Collision Lock:**
   - Enforce Mutual TLS (mTLS) and Next.js middleware validating administrative JWTs with roles: `P2P_DISPUTE_MAKER`, `P2P_DISPUTE_CHECKER`, `P2P_DISPUTE_SUPERVISOR`, `MLRO`.
   - Implement an operator collision lock mechanism via Redis (`POST /api/v1/p2p-disputes/{id}/lock`): when an operator opens a case, acquire an exclusive 15-minute lease with a 60-second heartbeat.
   - Display active operator presence indicators on the case view and prevent dual concurrent modifications.

3. **Construct Dynamic Forensic Screen Watermark (`ForensicWatermark.tsx`):**
   - Implement an immutable SVG overlay rendered across the entire viewport displaying: Operator ID, Role, Client Workstation IP, Active Case ID, UTC Timestamp, and Ephemeral Session Hash.
   - Attach a DOM MutationObserver to detect element hiding, opacity modification, or DOM removal, triggering an immediate UI freeze and security incident alert.

4. **Build Virtualized Real-Time Dispute Queue Desk (`DisputeQueueDesk.tsx`):**
   - Implement TanStack Table v8 with TanStack Virtual for smooth scrolling across thousands of active disputes.
   - Connect to WebSocket feed (`wss://admin.growww.in/ws/v1/p2p-disputes`) to stream live queue updates and state transitions.
   - Implement priority sorting based on SLA Countdown (e.g., <30m remaining = critical red highlight), fiat value bracket, dispute reason, and counterparty fraud risk tier.

5. **Construct Split-Screen Dispute Inspection Viewport (`SplitScreenDisputeInspector.tsx`):**
   - Design a responsive two-column grid: Left pane dedicated to Buyer Dossier, Right pane dedicated to Seller Dossier.
   - Ingest KYC identity verification data (PAN name match status, account age, total trades, dispute rate).
   - Display banking details: Registered IFSC, VPA, Account Number (masked with toggle for authorized compliance roles), and bank name.

6. **Develop Sandboxed PDF Bank Statement Viewer (`BankStatementViewer.tsx`):**
   - Integrate `react-pdf` / `pdfjs-dist` within a secure, sandboxed client canvas (disabling external script execution and link navigation).
   - Implement zoom, pan, rotation, high-contrast inversion, and side-by-side comparison controls.
   - Render an interactive OCR text layer overlay highlighting key extracted fields: Transaction Date, Counterparty Name, UTR / Reference ID, Transaction Amount, and Closing Balance.

7. **Implement Bank UTR Gateway & KYC Name-Matching Analyzer (`UtrReconciliationTool.tsx`):**
   - Create an automated reconciliation panel querying the Banking Aggregator API (`GET /api/v1/banking/utr/{utrNumber}`).
   - Display live clearance status: `SETTLED`, `PENDING_CLEARANCE`, `REVERSED`, `INVALID_UTR`, or `NOT_FOUND`.
   - Compute strict KYC Name-Matching Invariant:
     - Execute exact string comparison, Levenshtein distance score, and Double Metaphone phonetic similarity between the remitter name from the bank and the registered buyer KYC name.
     - Display a color-coded confidence badge: `100% MATCH (PASSED)`, `PHONETIC_MATCH (REVIEW)`, or `CRITICAL_MISMATCH (THIRD_PARTY_SUSPECTED)`.

8. **Build Dispute Chat Timeline & Evidence Log (`DisputeChatTimeline.tsx`):**
   - Construct a chronological audit log displaying the buyer-seller negotiation chat history, uploaded attachment thumbnails, payment timer expirations, and dispute initiation triggers.
   - Render unalterable timestamps with millisecond precision and cryptographic SHA-256 hashes of all submitted images/documents.

9. **Build Maker Adjudication Proposal Terminal (`MakerProposalForm.tsx`):**
   - Implement adjudication action selectors: `RELEASE_ESCROW_TO_BUYER`, `REFUND_ESCROW_TO_SELLER`, `PARTIAL_SPLIT_SETTLEMENT`.
   - Require mandatory structured adjudication inputs: Primary Justification Category, Detailed Legal Case Notes (minimum 100 characters), Bank Verification Reference, and Evidence Attachment IDs.
   - Integrate WebAuthn / FIDO2 challenge requiring the Maker officer to authenticate with a physical hardware key (YubiKey) to sign and dispatch the proposal (`POST /api/v1/p2p-disputes/{id}/maker-proposal`).

10. **Implement Checker Review & Dual-Control Approval Panel (`CheckerReviewPanel.tsx`):**
    - Construct an air-gapped review interface displaying the Maker's determination, evidence links, and bank verification outputs.
    - Prevent the Maker officer from acting as Checker on the same case (strict dual-control identity check: `makerId !== checkerId`).
    - Provide `APPROVE_AND_EXECUTE` and `REJECT_WITH_REMARKS` actions.
    - Requiring Checker WebAuthn / FIDO2 hardware token signing upon approval.

11. **Develop Viem EIP-712 Signing Engine & Multisig Bridge (`MultisigArbitrationEngine.ts`):**
    - Construct the EIP-712 typed data payload matching `ArbitrationResolution` on `P2PEscrow.sol`.
    - Gather the Maker signature, Checker signature, and request the platform CloudHSM signature from Prompt 717 via internal mTLS service (`POST /api/v1/signer/p2p-arbitration`).
    - Execute the on-chain arbitration transaction using Viem 2.x, submitting the 2-of-3 threshold signatures to `P2PEscrow.sol`.
    - Display on-chain transaction receipt, block number, gas utilized, and Besu transaction hash with a block explorer link.

12. **Build FIU-IND Suspicious Activity Flagging Console (`FiuFlaggingPanel.tsx`):**
    - Provide a one-click regulatory escalation panel to flag suspicious parties under PMLA guidelines.
    - Include predefined suspicious indicator tags: `MULE_ACCOUNT_INDICATOR`, `RAPID_FUND_MOVEMENT`, `REPEATED_DISPUTE_DEFENDANT`, `THIRD_PARTY_NAME_MISMATCH`, `COUNTERFEIT_BANK_STATEMENT`.
    - Compile a standardized FIU-IND STR JSON payload and queue it to Kafka topic `fiu.str.flags.v1` for downstream dispatch by Prompt 233.

13. **Implement Full Operator Telemetry & Kafka Audit Logger (`AuditLogger.ts`):**
    - Track all operator interactions: Case Opened, PDF Page Viewed, OCR Field Clicked, UTR Verified, Proposal Created, Proposal Rejected, Proposal Approved.
    - Synchronously emit structured audit payloads to Kafka topic `audit.admin.p2p.v1` with cryptographic hashing for tamper resistance.

14. **Configure Security Hardening, Content Security Policy, and Anti-Clickjacking:**
    - Set strict Next.js security headers: `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`.
    - Enforce stringent Content Security Policy (CSP) blocking external script execution, data exfiltration, and unsafe eval.
    - Disable client-side printing and screen capture shortcuts via CSS print media blanking.

15. **Implement Comprehensive Automated Unit, Cryptographic, and E2E Tests:**
    - Vitest unit tests for Levenshtein name-matching algorithms, EIP-712 hash generation, and Zod schema validations.
    - Playwright end-to-end tests validating: Operator collision lock acquisition, split-screen PDF inspection, Maker proposal submission, Checker approval, and multisig dispatch.

## Interfaces / Contracts

### P2P Dispute Domain Schemas (TypeScript)
```typescript
export type P2PDisputeStatus =
  | 'PENDING_ASSIGNMENT'
  | 'IN_REVIEW_MAKER'
  | 'PENDING_CHECKER_APPROVAL'
  | 'RESOLVED_RELEASED_TO_BUYER'
  | 'RESOLVED_REFUNDED_TO_SELLER'
  | 'RESOLVED_PARTIAL_SPLIT'
  | 'ESCALATED_LEGAL';

export type DisputeReason =
  | 'PAYMENT_NOT_RECEIVED'
  | 'PAYMENT_AMOUNT_MISMATCH'
  | 'THIRD_PARTY_PAYMENT_SUSPECTED'
  | 'FRAUDULENT_RECEIPT_SUBMITTED'
  | 'BANK_ACCOUNT_FROZEN'
  | 'BUYER_MARKED_PAID_WITHOUT_TRANSFER';

export type AdjudicationDecision =
  | 'RELEASE_ESCROW_TO_BUYER'
  | 'REFUND_ESCROW_TO_SELLER'
  | 'PARTIAL_SPLIT_SETTLEMENT';

export interface BankAccountDetails {
  bankName: string;
  accountNumberMasked: string;
  ifscCode: string;
  accountHolderName: string;
  upiVpa?: string;
  verificationStatus: 'PENNY_DROP_VERIFIED' | 'MANUAL_VERIFIED' | 'UNVERIFIED';
}

export interface DisputePartyProfile {
  userId: string;
  legalKycName: string;
  panMasked: string;
  kycTier: 'TIER_1' | 'TIER_2' | 'ENTERPRISE';
  walletAddress: `0x${string}`;
  bankAccount: BankAccountDetails;
  metrics: {
    totalP2POrders: number;
    completionRate: number;
    totalDisputes: number;
    disputeLossRate: number;
    reputationScore: number;
  };
  deviceFingerprint: string;
  ipAddress: string;
}

export interface PaymentEvidence {
  evidenceId: string;
  uploadedBy: 'BUYER' | 'SELLER';
  documentType: 'BANK_STATEMENT_PDF' | 'UPI_SCREENSHOT' | 'PAYMENT_SLIP' | 'OTHER';
  s3ObjectKey: string;
  fileSizeBytes: number;
  mimeType: string;
  sha256Hash: string;
  uploadedAt: string;
  ocrExtractedData?: {
    utrNumber?: string;
    remitterName?: string;
    beneficiaryName?: string;
    amount?: string;
    transactionTimestamp?: string;
    ocrConfidenceScore: number;
    tamperFlags: string[];
  };
}

export interface UtrVerificationResult {
  utrNumber: string;
  status: 'SETTLED' | 'PENDING_CLEARANCE' | 'REVERSED' | 'INVALID_UTR' | 'NOT_FOUND';
  remitterBank: string;
  beneficiaryBank: string;
  remitterNameRaw: string;
  amount: string;
  clearingTimestamp: string;
  nameMatchScore: {
    exactMatch: boolean;
    levenshteinDistance: number;
    phoneticSimilarity: number;
    verdict: 'PASS' | 'FLAG_MISMATCH' | 'MANUAL_REVIEW';
  };
}

export interface P2PEscrowDisputeCase {
  disputeId: string;
  escrowId: `0x${string}`;
  orderId: string;
  tokenSymbol: 'USDT' | 'USDC' | 'BTC' | 'ETH';
  tokenAmount: string;
  fiatCurrency: 'INR';
  fiatAmount: string;
  createdAt: string;
  slaExpiresAt: string;
  status: P2PDisputeStatus;
  reason: DisputeReason;
  buyer: DisputePartyProfile;
  seller: DisputePartyProfile;
  evidenceList: PaymentEvidence[];
  utrVerification?: UtrVerificationResult;
  activeOperatorLock?: {
    operatorId: string;
    operatorName: string;
    lockedAt: string;
    expiresAt: string;
  };
}
```

### Maker-Checker & Cryptographic Multisig Schemas
```typescript
export interface MakerProposalPayload {
  disputeId: string;
  escrowId: `0x${string}`;
  makerOfficerId: string;
  decision: AdjudicationDecision;
  recipientAddress: `0x${string}`;
  releaseAmount: string;
  penaltyDeductionAmount?: string;
  justificationCategory: string;
  detailedCaseNotes: string;
  utrReferenceValidated: string;
  webauthnAttestation: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
  proposedAt: string;
}

export interface CheckerReviewPayload {
  disputeId: string;
  proposalId: string;
  checkerOfficerId: string;
  action: 'APPROVE' | 'REJECT';
  rejectionReason?: string;
  independentNotes: string;
  webauthnAttestation: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
  reviewedAt: string;
}

export interface EIP712ArbitrationPayload {
  types: {
    EIP712Domain: [
      { name: 'name'; type: 'string' },
      { name: 'version'; type: 'string' },
      { name: 'chainId'; type: 'uint256' },
      { name: 'verifyingContract'; type: 'address' }
    ];
    ArbitrationResolution: [
      { name: 'escrowId'; type: 'bytes32' },
      { name: 'recipient'; type: 'address' },
      { name: 'amount'; type: 'uint256' },
      { name: 'nonce'; type: 'uint256' },
      { name: 'deadline'; type: 'uint256' },
      { name: 'disputeHash'; type: 'bytes32' },
      { name: 'justificationHash'; type: 'bytes32' }
    ];
  };
  primaryType: 'ArbitrationResolution';
  domain: {
    name: 'GrowwwP2PEscrow';
    version: '1';
    chainId: number;
    verifyingContract: `0x${string}`;
  };
  message: {
    escrowId: `0x${string}`;
    recipient: `0x${string}`;
    amount: bigint;
    nonce: bigint;
    deadline: bigint;
    disputeHash: `0x${string}`;
    justificationHash: `0x${string}`;
  };
}

export interface MultisigExecutionReceipt {
  escrowId: `0x${string}`;
  txHash: `0x${string}`;
  blockNumber: bigint;
  gasUsed: bigint;
  status: 'SUCCESS' | 'REVERTED';
  executedAt: string;
  arbitrationSigners: `0x${string}`[];
}
```

### FIU-IND Suspicious Transaction Report (STR) Schemas
```typescript
export interface FiuStrFlagPayload {
  caseId: string;
  disputeId: string;
  flaggedUserId: string;
  flaggedWalletAddress: `0x${string}`;
  reportingOfficerId: string;
  mlroApprovalRequired: boolean;
  suspicionIndicators: Array<
    | 'MULE_ACCOUNT_PATTERN'
    | 'RAPID_PASS_THROUGH'
    | 'THIRD_PARTY_FIAT_DEPOSIT'
    | 'DOCUMENT_FORGERY_DETECTED'
    | 'STRUCTURING_SMURFING'
    | 'REPEATED_HIGH_RISK_DISPUTES'
  >;
  fiatAmount: string;
  assetVolume: string;
  narrativeSummary: string;
  supportingDocumentHashes: string[];
  submissionTimestamp: string;
}
```

### REST / WebSocket Operator Endpoints
```text
# Real-Time Queue & Operator Presence
GET    /api/v1/p2p-disputes/queue                     (Fetch Virtualized Queue)
POST   /api/v1/p2p-disputes/{id}/lock                 (Acquire 15-Minute Operator Lease)
POST   /api/v1/p2p-disputes/{id}/heartbeat            (Renew Operator Lease)
POST   /api/v1/p2p-disputes/{id}/unlock               (Release Operator Lease)

# Case Dossier & Evidence Verification
GET    /api/v1/p2p-disputes/{id}                      (Fetch Complete Split-Screen Case Dossier)
GET    /api/v1/p2p-disputes/{id}/evidence/{evidenceId}(Fetch Presigned Encrypted Evidence Stream)
POST   /api/v1/p2p-disputes/{id}/verify-utr           (Trigger Live Open Banking UTR Lookup)

# Maker-Checker Adjudication Workflow
POST   /api/v1/p2p-disputes/{id}/maker-proposal       (Submit Maker Adjudication Package)
GET    /api/v1/p2p-disputes/checker/pending           (Fetch Pending Checker Review Tasks)
POST   /api/v1/p2p-disputes/{id}/checker-decision     (Checker Approve or Reject Proposal)

# Cryptographic Multisig Arbitration & Blockchain
POST   /api/v1/p2p-disputes/{id}/generate-eip712      (Generate Structured Arbitration Hash)
POST   /api/v1/p2p-disputes/{id}/request-hsm-sign     (Request CloudHSM Co-Signature via Prompt 717)
POST   /api/v1/p2p-disputes/{id}/execute-multisig     (Submit 2-of-3 Multisig Tx to P2PEscrow.sol)

# Regulatory Compliance & Audit
POST   /api/v1/p2p-disputes/{id}/fiu-flag             (Trigger FIU-IND STR Compliance Draft)
POST   /api/v1/p2p-disputes/{id}/audit-log            (Emit Detailed Operator Telemetry to Kafka)

# WebSocket Real-Time Stream
WS     /ws/v1/p2p-disputes/live                       (Stream Ticket Status, Escalations, Locks)
```

## Security & Compliance Notes
- **Dual-Control (Maker-Checker) Mandate (RUNBOOK-27):** Under no circumstance can a single compliance officer unilaterally seize, release, or forfeit escrowed assets. Every adjudication requires a Maker proposal and an independent Checker sign-off. The system enforces `makerOfficerId !== checkerOfficerId` at the database, backend, and smart contract verification layers.
- **Hardware Security Key Requirement (WebAuthn / FIDO2):** Both Maker proposal submissions and Checker release approvals mandate biometric or physical pin verification via FIPS 140-2 Level 3 hardware security keys (e.g., YubiKey 5 Series). Software OTPs, passwords, and browser-cached keys are strictly prohibited for arbitration sign-offs.
- **2-of-3 Multisig Threshold on `P2PEscrow.sol`:** Escrow settlement transactions during dispute arbitration require two independent cryptographic signatures: either (Maker + Checker) or (Authorized Operator + CloudHSM Key Daemon from Prompt 717). This ensures zero single point of failure or rogue operator risk.
- **Prevention of Money Laundering Act (PMLA) & FIU-IND Reporting:** All adjudications involving suspected third-party bank accounts, counterfeit bank statement PDFs, or structured smurfing trigger mandatory FIU-IND red flags. The console automatically compiles an STR dossier (including verified KYC profile, device fingerprints, transaction hashes, and bank statement checksums) dispatched to Kafka topic `fiu.str.flags.v1`.
- **Digital Personal Data Protection (DPDP) Act 2023 Compliance:** Bank account numbers, UPI VPAs, and PAN/Aadhaar references are masked by default across all viewports (`XXXX-XXXX-1234`). Only compliance officers with the elevated role `P2P_DISPUTE_INVESTIGATOR` can temporarily unmask sensitive records for up to 5 minutes, with every unmask action logged immutably.
- **Sandboxed PDF Inspection:** Uploaded buyer and seller bank statements are processed in an isolated Web Worker canvas sandbox. JavaScript execution, active PDF form actions, and external hyperlinks embedded in PDFs are completely stripped to prevent cross-site scripting (XSS) or browser exploit attacks.
- **Anti-Exfiltration Forensic Watermarking:** The console renders a tamper-resistant SVG watermark across all dispute viewports, displaying the active officer ID, IP address, workstation timestamp, and session hash. Screen capture attempts or DOM manipulation trigger instant session termination.
- **Operator Collision Locking:** Redis distributed locks ensure that two compliance officers cannot work on or modify the same dispute ticket concurrently, eliminating duplicate reviews and race conditions.

## Acceptance Criteria
- [ ] Next.js 14 dispute console initializes under route `apps/growww_admin/p2p-disputes` with strict mTLS and RBAC role validation.
- [ ] Unauthorized access without valid administrative credentials or insufficient role permissions is blocked with HTTP 403 Forbidden.
- [ ] Real-time dispute queue virtualizes and renders 1,000+ active dispute cases with zero scroll latency, streaming status updates over WebSockets.
- [ ] Redis distributed locking successfully acquires a 15-minute exclusive lease when an operator opens a case, preventing other operators from modifying the ticket.
- [ ] Split-screen dispute inspector displays Buyer Dossier and Seller Dossier side-by-side with complete KYC status, bank account details, and trade history metrics.
- [ ] Sandboxed PDF viewer loads bank statements securely, provides zoom/pan/rotate controls, and overlays OCR extracted text layers.
- [ ] OCR engine automatically parses UTR number, remitter name, beneficiary name, transaction amount, and timestamp from uploaded statements.
- [ ] UTR reconciliation tool queries banking aggregator APIs and accurately returns settlement status (`SETTLED`, `PENDING_CLEARANCE`, `NOT_FOUND`).
- [ ] KYC Name Matching Invariant tool calculates Levenshtein and phonetic similarity, correctly flagging third-party payments where sender name does not match KYC.
- [ ] Dispute chat timeline renders historical buyer-seller communication, system countdown timers, and document upload hashes with millisecond accuracy.
- [ ] Maker adjudication proposal form validates required fields, enforces minimum 100-character case notes, and requires WebAuthn hardware token signing.
- [ ] Checker review panel strictly enforces identity separation (`makerOfficerId !== checkerOfficerId`) and allows Checker approval or rejection with hardware token signing.
- [ ] Viem EIP-712 engine generates valid `ArbitrationResolution` structured data hashes adhering to `P2PEscrow.sol` domain specifications.
- [ ] CloudHSM bridge (Prompt 717) successfully co-signs arbitration payloads, completing the 2-of-3 multisig threshold.
- [ ] Multisig release transaction successfully broadcasts to the Besu ledger, releasing or refunding escrowed assets and emitting on-chain event receipts.
- [ ] FIU-IND flagging console compiles structured STR packages with suspicion indicators and dispatches them to Kafka topic `fiu.str.flags.v1`.
- [ ] Dynamic forensic screen watermark renders continuously; DOM tampering or removal triggers an immediate session lock.
- [ ] All operator actions (page view, unmask, OCR inspect, maker submit, checker sign) are immutably emitted to Kafka topic `audit.admin.p2p.v1`.
- [ ] Vitest unit test suite passes with >90% code coverage across name-matching algorithms, EIP-712 hashing, and Zod schemas.
- [ ] Playwright E2E tests validate complete end-to-end adjudication flow from queue assignment to on-chain multisig execution.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 217 (Admin & Back-Office Service: Role-Based Access Control & Maker-Checker Task Engine).
  - Prompt 218 (Immutable Audit Log Service: Event commit hooks & WORM storage).
  - Prompt 269 (P2P Escrow & Order Lifecycle Service: Core dispute state machine and escrow locking).
  - Prompt 302 (Hyperledger Besu Consortium Network: Ledger RPC connectivity and node consensus).
  - Prompt 306 (Settlement DvP & Atomic Clearing Smart Contracts).
  - Prompt 601 (Next.js 14 Investor Web App Scaffolding & Shared Admin Component Library).
  - Prompt 717 (HSM & CloudHSM Key Lifecycle and Cryptographic Signing Daemon: FIPS 140-2 Level 3 co-signing).
- **Parallel Tasks:**
  - Prompt 607 (Admin Console: Regulatory Reporting & Compliance Export Portal).
  - Prompt 613 (Next.js 14 Regulatory Audit & SEBI/IFSCA Supervisory Dashboard).
  - Prompt 614 (Next.js 14 Clearing Member & Broker Capital Adequacy Portal).
  - Prompt 806 (Observability Stack: Telemetry, Distributed Tracing, and Alerting).
- **Downstream Blockers:**
  - Prompt 906 (UAT Plan: End-to-End P2P Fiat Escrow Dispute Resolution Testing).
  - Prompt 907 (Operational Readiness Review: P2P Operations Desk Runbook Certification).
