# 602 - Web Onboarding, DigiLocker & Camera KYC Flow

## Purpose
Delivers a streamlined, regulatory-compliant web onboarding and KYC verification wizard for domestic Indian retail and high-net-worth investors (`apps/growww_web/app/(auth)/onboarding/`). Integrates DigiLocker OAuth2 redirect/modal flow, instant Income Tax / NSDL PAN verification, bank penny-drop verification, and browser-based WebRTC camera selfie capture with passive liveness detection and face alignment guidance. Fully compliant with SEBI master KYC guidelines, Prevention of Money Laundering Act (PMLA), and RBI digital lending/onboarding norms to ensure only authenticated, verified investors can access fractional equity tokens.

## What You Are Building
- Multi-step KYC onboarding wizard (`apps/growww_web/app/(auth)/onboarding/`) with persistent progress recovery.
- Step 1: Investor profile creation & phone/email OTP verification.
- Step 2: Instant PAN input form with real-time format validation and NSDL verification lookup.
- Step 3: DigiLocker integration client (OAuth pop-up / redirect handling Aadhaar XML extraction).
- Step 4: WebRTC browser camera capture module with real-time facial guide overlay, lighting check, and passive liveness capture.
- Step 5: Bank account details input with IFSC auto-fetch (branch/bank name lookup) and automated penny-drop verification trigger.
- Step 6: Mandatory SEBI declarations (tax residency, PEP status, FATCA/CRS, income bracket).
- Step 7: Final review summary screen with digital signature capture (canvas signature pad) and CKYC consent checkbox.

## Scope Boundaries
- **In Scope:**
 - Client-side multi-step wizard UI, state management, and step transition validation.
 - WebRTC video stream capture, canvas frame extraction, and client-side face alignment guidance.
 - DigiLocker OAuth client flow and popup message listener.
 - Form validation with Zod schemas and React Hook Form.
 - Client-side document image preview, compression, and error recovery modals.
- **Out of Scope / Handled Elsewhere:**
 - Server-side KYC processing engine, OCR, and face-match algorithms (Prompt 202).
 - Admin manual KYC review dashboard (Prompt 604).
 - Sanctions and PEP screening backend services (Prompt 703).
 - Foreign investor onboarding via GIFT City (Prompts 005, 214).

## Technology to Use
- **Next.js 14 App Router & React 18/19:** Powers client-side state transitions with server actions for secure document upload token generation.
- **React Hook Form & Zod:** Provides high-performance, zero-re-render form state management with strict TypeScript schema validation. Justification: Zod schemas ensure compile-time and runtime type safety matching backend API contracts.
- **WebRTC `MediaDevices.getUserMedia()` & HTML5 Canvas API:** Captures high-resolution webcam video streams directly in modern browsers without third-party plugins.
- **Face-api.js / Lightweight WebAssembly Face Detector:** Runs lightweight client-side bounding box and head pose checks to provide real-time visual feedback ("Center your face", "Increase lighting") before capture.
- **Lucide React & Tailwind CSS:** Clean, accessible UI components with animated progress steppers.

## Backend / Infra Touchpoints
- **KYC Backend Microservice (Prompt 202):** Connects via REST (`/api/v1/kyc/*`) for PAN verification, DigiLocker token exchange, selfie upload, and liveness scoring.
- **Payment Gateway Integration Service (Prompt 212):** Triggers automated penny-drop bank account validation.
- **User & Identity Service (Prompt 201):** Updates user onboarding state upon milestone completions.
- **S3 / MinIO Object Storage (Prompt 401):** Uploads encrypted document blobs using short-lived pre-signed PUT URLs.

## Blockchain Interaction
- **Compliance Registry Awareness:** Displays live on-chain investor compliance status indicator ("KYC Whitelist Pending / Active").
- **On-Chain Whitelist Listener:** On KYC approval by compliance officers, the web client listens for the `ComplianceRegistry.InvestorWhitelisted` event on Hyperledger Besu to unlock trading capabilities dynamically without requiring full page refresh.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `ComplianceRegistry.sol` (Queries `isWhitelisted(address userAddress)` and `getKycTier(address userAddress)`).
- **Cryptographic Attestation:** The hash of the verified identity package (`identity_hash = SHA256(pan + aadhaar_ref + timestamp)`) is registered on-chain by the KYC relayer upon approval, establishing an immutable compliance record without exposing raw PII on-chain.

## Step-by-Step Build Instructions
1. Create onboarding route structure (`app/(auth)/onboarding/page.tsx`, `components/stepper.tsx`, `hooks/use-kyc-state.ts`).
2. Implement persistent onboarding state store using Zustand with `localStorage` fallback to allow users to resume interrupted sessions.
3. Build Step 1 (Profile & OTP): Phone and Email input forms with 6-digit OTP countdown timer and auto-focus input cells.
4. Build Step 2 (PAN Verification): Form with Zod regex validation (`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`), auto-capitalization, and instant validation status badge showing registered taxpayer name.
5. Build Step 3 (DigiLocker Integration): Implement "Fetch via DigiLocker" button opening an authenticated OAuth popup window with `window.postMessage` listener to handle redirect callbacks and extract verified Aadhaar details.
6. Build Step 4 (WebRTC Camera Selfie Capture): Request user media permissions (`video: { width: 1280, height: 720, facingMode: 'user' }`) with error handling for denied permissions.
7. Implement canvas overlay rendering an oval alignment guide, real-time lighting luminance check, and head posture feedback.
8. Implement burst frame capture capturing 3 sequential frames upon stable liveness detection, compressing to WebP format (<500KB) on client.
9. Build Step 5 (Bank Account Verification): Bank account number and IFSC code form with automated bank name/branch auto-population and penny-drop trigger button displaying live verification status.
10. Build Step 6 (SEBI Declarations): Checkbox forms for Politically Exposed Person (PEP) status, Indian Tax Residency, Gross Annual Income bracket, and Occupation.
11. Build Step 7 (Review & Digital Signature): Summary card displaying extracted details, HTML5 canvas signature pad for investor digital signature, and CKYC consent checkbox.
12. Implement document upload mechanism requesting pre-signed S3 URLs from backend and uploading encrypted blobs with SHA-256 integrity checksums.
13. Write comprehensive end-to-end Playwright tests simulating webcam stream with virtual video device and mocking DigiLocker OAuth flow.

## Interfaces / Contracts
```typescript
import { z } from 'zod';

export const PanSchema = z.object({
  panNumber: z.string().regex(/^[A-Z]{5}[0-9]{4}[A-Z]{1}$/, 'Invalid PAN format'),
  fullNameAsPerPan: z.string().min(2, 'Name is required'),
  dateOfBirth: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'YYYY-MM-DD required'),
});

export const BankDetailsSchema = z.object({
  accountNumber: z.string().min(9).max(18),
  confirmAccountNumber: z.string(),
  ifscCode: z.string().regex(/^[A-Z]{4}0[A-Z0-9]{6}$/, 'Invalid IFSC format'),
  accountType: z.enum(['SAVINGS', 'CURRENT']),
}).refine(data => data.accountNumber === data.confirmAccountNumber, {
  message: "Account numbers don't match",
  path: ['confirmAccountNumber'],
});

export interface KycWizardState {
  currentStep: number;
  completedSteps: number[];
  panData?: z.infer<typeof PanSchema>;
  digiLockerSuccess: boolean;
  aadhaarLastFour?: string;
  selfieBlobUri?: string;
  livenessConfidence?: number;
  bankDetails?: z.infer<typeof BankDetailsSchema>;
  bankVerifiedName?: string;
  pennyDropSuccess: boolean;
  declarations: {
    isPep: boolean;
    taxResidentIndiaOnly: boolean;
    incomeBracket: string;
    occupation: string;
  };
  signatureImageBase64?: string;
}

export interface KycSubmissionPayload {
  userId: string;
  pan: string;
  nameAsPerPan: string;
  dob: string;
  digiLockerTransactionId: string;
  selfieStorageKey: string;
  livenessScore: number;
  bankAccountNumber: string;
  ifsc: string;
  pennyDropRefId: string;
  declarations: KycWizardState['declarations'];
  digitalSignatureStorageKey: string;
  clientChecksum: string;
}
```

## Security & Compliance Notes
- **Aadhaar Masking:** Strictly displays only the last 4 digits of Aadhaar (UIDAI compliance regulations).
- **Ephemeral Camera Data:** Video stream frames processed purely in memory; raw video stream terminated immediately upon capture.
- **Client-Side SHA-256 Checksums:** All uploaded document blobs calculate SHA-256 checksums on the client to guarantee payload integrity during upload.
- **Encrypted Transmission:** All KYC data transmitted over TLS 1.3 with anti-CSRF headers and session-bound JWT tokens.

## Acceptance Criteria
- [ ] Multi-step KYC wizard successfully advances and persists state across page reloads.
- [ ] PAN input strictly enforces format validation and resolves registered name via API.
- [ ] DigiLocker popup flow initiates, captures OAuth token via `postMessage`, and populates address details.
- [ ] WebRTC camera component streams webcam, displays face alignment guide, and captures 3-frame burst.
- [ ] Bank account step verifies IFSC and displays penny-drop confirmation within 5 seconds.
- [ ] SEBI declarations and digital signature pad capture valid inputs and block submission if incomplete.
- [ ] All uploaded files include pre-signed S3 URL upload and client-side SHA-256 integrity checks.
- [ ] E2E Playwright test suite passes for complete onboarding flow with mocked APIs.

## Suggested Order / Dependencies
- **Prerequisites:** 201 (User Service), 202 (KYC Service), 212 (Payment Gateway), 601 (Web Scaffolding).
- **Direct Successors / Parallel:** 504 (Flutter KYC Flow), 603 (Web Trading Dashboard), 604 (Admin KYC Review).
