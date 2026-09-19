# 604 - Admin Console: User Management & KYC Review Dashboard

## Purpose
Provides compliance officers, KYC analysts, and operations teams with an enterprise internal back-office administration portal (`apps/growww_admin/app/(dashboard)/kyc-review/`) to inspect investor profiles, perform maker-checker KYC document reviews, inspect automated liveness and OCR confidence scores, examine sanctions/PEP screening matches, and execute regulatory onboarding decisions. Ensures strict compliance with SEBI and RBI anti-money laundering (AML) guidelines by enforcing dual-authorization approval before an investor is registered on the permissioned blockchain compliance whitelist.

## What You Are Building
- High-performance, filterable KYC applicant queue table (`apps/growww_admin/app/(dashboard)/kyc-review/`) with real-time status badges (Pending Maker Review, Pending Checker Approval, Flagged/High Risk, Approved, Rejected).
- Split-screen KYC inspection workstation: side-by-side comparison of DigiLocker Aadhaar XML data vs User Submitted Data, OCR extracted PAN card vs NSDL database record, and WebRTC selfie image vs Aadhaar photo with synchronized zoom.
- Sanctions / Politically Exposed Person (PEP) screening alert review modal displaying match confidence scores, match attributes (DOB, aliases, country), and false-positive dismissal controls.
- Maker-checker workflow controls: Maker submits approval recommendations or rejection reasons; Checker conducts final review with WebAuthn/MFA challenge.
- Comprehensive investor audit trail timeline displaying timestamped history of all submitted documents, OCR scores, analyst notes, and status changes.

## Scope Boundaries
- **In Scope:**
 - Admin KYC review UI, filterable queue, and real-time status indicators.
 - Split-pane document comparison viewer with pan/zoom tools.
 - OCR confidence and facial biometric distance score indicators.
 - Maker-checker review action modals with standardized SEBI rejection codes.
 - On-chain whitelist initiation and transaction status tracking.
- **Out of Scope / Handled Elsewhere:**
 - Retail investor KYC submission wizard (Prompt 602).
 - Backend KYC verification and OCR pipeline (Prompt 202).
 - Admin backend service (Prompt 217).
 - Global Sanctions / PEP screening backend integration (Prompt 703).

## Technology to Use
- **Next.js 14 App Router, React 18/19, TypeScript 5.4+:** Provides secure SSR rendering for internal admin tools with robust role-based route protection.
- **`@tanstack/react-table` (v8):** Powers the high-performance KYC queue table with multi-column sorting, facet filtering, server-side pagination, and row selection.
- **`@panzoom/panzoom` / React Image Zoom:** Delivers hardware-accelerated image zoom and panning for inspecting high-resolution document scans and selfie captures.
- **shadcn/ui (Radix UI) & Tailwind CSS:** Clean, accessible enterprise interface with consistent alert badges and modal dialogs.
- **Lucide React & Sonner:** Informative iconography and actionable toast alerts.

## Backend / Infra Touchpoints
- **Admin Service (Prompt 217):** Connects via REST / gRPC-Web (`/api/v1/admin/kyc/*`) for queue fetching, profile details, and decision dispatch.
- **KYC Microservice (Prompt 202):** Fetches encrypted document URLs (short-lived pre-signed S3 links) and biometric similarity scores.
- **Audit Log Service (Prompt 218):** Logs all document views, unmasking events, and approval/rejection actions to immutable audit log.
- **Sanctions Screening Engine (Prompt 703):** Fetches detailed PEP/Sanctions match profiles from World-Check / OpenSanctions.

## Blockchain Interaction
- **On-Chain Compliance Whitelisting:** Upon final Checker approval, the admin console triggers the backend HSM relayer to execute `ComplianceRegistry.whitelistInvestor(address userWallet, bytes32 identityHash, uint8 kycTier)` on Hyperledger Besu.
- **Audit Verification:** The admin UI displays the resulting Besu transaction hash, block number, and on-chain event receipt directly in the investor's compliance record for immediate auditing.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `ComplianceRegistry.sol` (Maintains whitelist status, KYC tier, and freeze state per investor address).
- **Security Control:** Admin UI does not sign on-chain transactions directly with raw keys; it submits signed administrative requests authenticated by the officer's WebAuthn/FIDO2 session to the Admin Service, which executes through a FIPS 140-2 Level 3 HSM-backed relayer.

## Step-by-Step Build Instructions
1. Scaffold admin KYC review route under `apps/growww_admin/app/(dashboard)/kyc-review/` with sub-routes for the queue list and individual applicant detail view (`[id]/page.tsx`).
2. Implement TanStack Table for the KYC queue with columns: Applicant ID, Full Name, PAN, Submission Timestamp, Risk Score (Low/Medium/High), Stage (Maker/Checker), and Status.
3. Build facet filter bar allowing compliance officers to filter by Risk Level, Date Range, Rejection Category, and Assigned Analyst.
4. Add WebSocket subscription to auto-refresh queue counts and flash newly submitted applications in real time.
5. Build Split-Screen Inspection Workstation layout with adjustable split-pane resizing.
6. Build Left Panel: Identity & Bank Data Card displaying Applicant PAN, NSDL Name, Aadhaar Name, DOB, Address, Bank Name, Account Number, IFSC, and Penny-Drop Match status.
7. Build Right Panel: Biometric & Document Comparator featuring dual synchronized viewports displaying the WebRTC selfie photo and the Aadhaar/PAN photo with side-by-side zoom/pan.
8. Implement OCR Confidence & Biometric Match Widget displaying face-match similarity percentage (e.g. 96.4%) and highlighting any text mismatch between form input and OCR text.
9. Implement Sanctions & PEP Match Drawer showing matching watchlist entries (UNSC, OFAC, SEBI Debarred, Indian Terrorist List) with matched fields and risk severity.
10. Build Maker Action Modal with options: "Recommend Approval" (passes to Checker), "Request Clarification" (sends SMS/email to user for re-upload), or "Reject" (with standardized dropdown rejection codes: e.g. "Blurry Document", "Name Mismatch", "Liveness Failed").
11. Build Checker Authorization Modal requiring a second senior compliance officer to review the Maker's notes and confirm with WebAuthn/FIDO2 hardware key or TOTP MFA.
12. Build Compliance History Timeline displaying complete audit logs with timestamps, analyst names, and on-chain transaction hash upon whitelisting.
13. Write comprehensive unit and integration tests using Vitest and Playwright covering maker-checker approval transitions.

## Interfaces / Contracts
```typescript
export type KycReviewStatus =
  | 'PENDING_MAKER'
  | 'PENDING_CHECKER'
  | 'FLAGGED_HIGH_RISK'
  | 'APPROVED'
  | 'REJECTED'
  | 'CLARIFICATION_REQUESTED';

export interface KycQueueItem {
  applicationId: string;
  userId: string;
  fullName: string;
  pan: string;
  riskScore: 'LOW' | 'MEDIUM' | 'HIGH';
  sanctionsMatchCount: number;
  submittedAt: string;
  status: KycReviewStatus;
  assignedMaker?: string;
  assignedChecker?: string;
}

export interface KycApplicantDetail {
  applicationId: string;
  userId: string;
  personalInfo: {
    fullName: string;
    dob: string;
    panNumber: string;
    nsdlRegisteredName: string;
    panNsdlMatchPercent: number;
    aadhaarLastFour: string;
    aadhaarAddress: string;
  };
  bankVerification: {
    accountNumber: string;
    ifsc: string;
    bankName: string;
    registeredName: string;
    pennyDropStatus: 'VERIFIED' | 'FAILED' | 'NAME_MISMATCH';
  };
  biometrics: {
    selfieUrl: string;
    aadhaarPhotoUrl: string;
    panPhotoUrl: string;
    faceMatchScore: number; // 0.0 to 1.0
    livenessScore: number;
    livenessStatus: 'PASSED' | 'FAILED';
  };
  sanctionsMatches: Array<{
    listName: string;
    entityName: string;
    matchScore: number;
    matchDetails: string;
    isDismissed: boolean;
    dismissalReason?: string;
  }>;
  onChainStatus?: {
    isWhitelisted: boolean;
    txHash?: `0x${string}`;
    blockNumber?: number;
  };
}

export interface KycDecisionRequest {
  applicationId: string;
  action: 'MAKER_RECOMMEND_APPROVE' | 'MAKER_REJECT' | 'CHECKER_APPROVE' | 'CHECKER_REJECT' | 'REQUEST_CLARIFICATION';
  rejectionCode?: string;
  notes: string;
  mfaToken?: string;
}
```

## Security & Compliance Notes
- **Mandatory PII Masking:** Bank account numbers, PANs, and addresses are masked by default (e.g. `XXXX-XXXX-1234`); clicking "Unmask" logs an explicit unmask event in the immutable audit log (Prompt 218).
- **Strict Role-Based Access Control (RBAC):** Maker and Checker cannot be the same user for any single application (strict four-eyes principle).
- **Session Protection & Inactivity Lock:** Admin console automatically locks screen after 15 minutes of inactivity; re-authentication via WebAuthn required.
- **SEBI Audit Trail Compliance:** Retains immutable record of all reviewer actions, notes, and document checksums for a minimum of 8 years.

## Acceptance Criteria
- [ ] KYC queue table loads 500+ records with fluid filtering by risk score, date, and review stage.
- [ ] Split-screen document viewer enables synchronized zoom and pan across selfie and identity document photos.
- [ ] Face similarity score and OCR match percentage display prominently with color-coded warning thresholds.
- [ ] Maker review action successfully transitions status to "PENDING_CHECKER" and notifies designated checker.
- [ ] Checker approval requires MFA verification and cannot be performed by the initial Maker.
- [ ] Upon final approval, on-chain whitelisting transaction is initiated and the resulting Besu transaction hash is recorded in the timeline.
- [ ] Unmasking sensitive PII triggers an immediate audit log entry with analyst ID and timestamp.
- [ ] Automated Playwright tests cover queue filtering, document inspection, and dual-authorization approval.

## Suggested Order / Dependencies
- **Prerequisites:** 202 (KYC Service), 217 (Admin Service), 218 (Audit Log Service), 601 (Web Scaffolding), 702 (IAM).
- **Direct Successors / Parallel:** 605 (Multi-Party Approval UI), 606 (Proof of Reserve Dashboard), 607 (Regulatory Reporting).
