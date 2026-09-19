# 504 - Flutter Onboarding & KYC Flow UI

## Purpose
Implements the multi-step regulatory onboarding and Know-Your-Customer (KYC) user experience for domestic Indian investors (SEBI/KRA, Aadhaar OTP/DigiLocker, PAN verification, live selfie) and international investors onboarding via the GIFT City IFSC gateway (Passport, international proof of address). This flow ensures 100% compliance with SEBI and IFSCA anti-money laundering mandates while maintaining a frictionless, high-conversion mobile and desktop onboarding journey.

## What You Are Building
A comprehensive, secure onboarding module located under `apps/growww_flutter/lib/features/onboarding/` and `features/kyc/` including:
- **Investor Type Selection:** Toggle between Domestic Indian Resident (PAN/Aadhaar) and International / NRI (GIFT City gateway).
- **DigiLocker / Aadhaar Webview & OTP Flow:** Embedded secure browser bridge for direct C-KYC and DigiLocker identity document retrieval.
- **Document Capture with Auto-Edge Detection:** Native camera-based capture for PAN cards, Passports, and bank statements with client-side cropping and blur detection.
- **Biometric Liveness & Selfie Capture:** Interactive camera interface guiding the user with head-turn / blink detection to prevent spoofing.
- **Bank Account & Penny Drop Verification UI:** Bank IFSC lookup, account number input, and instant penny-drop validation feedback.
- **On-Chain Identity Whitelist Status:** Real-time indicator displaying the user's progress toward on-chain permissioning in `ComplianceRegistry.sol`.

## Scope Boundaries
- **In Scope:**
 - Client-side onboarding wizards, camera preview widgets, OCR document crop review, DigiLocker OAuth webview integration, selfie liveness capture UI, and form state validation.
- **Out of Scope / Handled Elsewhere:**
 - Backend KYC OCR, Aadhaar e-Sign, and C-KYC fetching services (handled in Prompt 202).
 - Third-party sanctions and PEP screening pipelines (handled in Prompt 703).
 - Local secure token storage (handled in Prompt 521).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` state management.
 - *Justification:* Flutter's native camera plugin and canvas clipping provide hardware-accelerated video frame analysis and responsive camera overlays across both mobile (iOS/Android) and desktop platforms.
- **Camera & Scanning:** `camera` plugin for live camera feed with real-time face overlay painter; `image_picker` for manual document selection fallback.
- **Secure Webview Bridge:** `webview_flutter` for isolated DigiLocker OAuth2 authentication.
- **Image Processing & Compression:** `flutter_image_compress` to compress captured documents under 500KB before upload.

## Backend / Infra Touchpoints
- **KYC Microservice:** REST API endpoints (`/api/v1/kyc/pan-verify`, `/api/v1/kyc/digilocker-url`, `/api/v1/kyc/liveness`, `/api/v1/kyc/bank-verify`) from Prompt 202.
- **Direct S3/MinIO Secure Upload:** Uses temporary presigned URLs to upload encrypted document images directly to secure object storage.

## Blockchain Interaction
- **Investor Whitelist Registration Display:** Shows the user their decentralized KYC status badge. Once KYC is verified by the backend compliance engine, the client displays their assigned on-chain identity hash registered in `ComplianceRegistry.sol` (ERC-3643 identity registry), qualifying the user for permissioned token transfers.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/kyc/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/repositories/`.
2. Define KYC domain state model `KycState` (`step`, `investorType`, `panStatus`, `aadhaarStatus`, `livenessStatus`, `bankStatus`, `complianceRegistryStatus`).
3. Build `InvestorTypeSelectionScreen` allowing the user to select Indian Resident vs International Investor.
4. Implement `PanVerificationScreen` with real-time regex validation (`[A-Z]{5}[0-9]{4}[A-Z]{1}`) and automated name matching feedback.
5. Build `DigiLockerWebViewScreen` with secure SSL navigation delegates, intercepting OAuth redirect callbacks and extracting temporary auth codes.
6. Implement `DocumentScannerScreen` using Flutter `camera` controller, drawing an optical guide frame for ID cards and checking image sharpness.
7. Build `LivenessCaptureScreen` rendering a circular face mask, providing visual cues ("Look straight", "Blink now", "Turn head slightly") and capturing the verification frame.
8. Implement image compression step adding a transparent watermark overlay: *"FOR GROWWW KYC VERIFICATION ONLY - [DATE]"*.
9. Build `BankDetailsScreen` with real-time RBI IFSC lookup displaying the corresponding bank name and branch.
10. Build `KycStatusPollingScreen` listening to server-sent events / WebSockets for asynchronous KYC approval updates.
11. Implement on-chain identity attestation visualizer showing the investor's verified cryptographic credential hash.
12. Write widget tests verifying form validation rules, camera permission error handling, and state progression.

## Interfaces / Contracts
```dart
// lib/features/kyc/domain/models/kyc_step_model.dart
enum KycStep { investorType, panVerify, digilocker, documentUpload, liveness, bankDetails, pendingApproval, approved }

enum InvestorType { domesticResident, internationalGiftCity }

class KycSubmissionPayload {
  final InvestorType investorType;
  final String panNumber;
  final String? digilockerTransactionId;
  final String selfieImagePresignedKey;
  final String? documentImagePresignedKey;
  final String bankAccountNumber;
  final String bankIfsc;

  const KycSubmissionPayload({
    required this.investorType,
    required this.panNumber,
    this.digilockerTransactionId,
    required this.selfieImagePresignedKey,
    this.documentImagePresignedKey,
    required this.bankAccountNumber,
    required this.bankIfsc,
  });

  Map<String, dynamic> toJson() => {
    'investor_type': investorType.name,
    'pan_number': panNumber,
    'digilocker_tx_id': digilockerTransactionId,
    'selfie_key': selfieImagePresignedKey,
    'doc_key': documentImagePresignedKey,
    'account_number': bankAccountNumber,
    'ifsc_code': bankIfsc,
  };
}
```

## Security & Compliance Notes
- **Zero Unencrypted Local Storage:** Raw captured photos of PAN, Aadhaar, Passports, or selfies MUST NEVER be written to unencrypted local device storage. Images are stored in volatile RAM or encrypted temporary cache and purged immediately after upload.
- **Client-Side Image Watermarking:** All captured identity documents are stamped with a mandatory anti-reuse watermark before transmission.
- **DPDP Act Compliance:** Masked Aadhaar display (only the last 4 digits visible) in compliance with UIDAI regulations.

## Acceptance Criteria
- [ ] PAN validation accepts only valid alphanumeric formats and integrates with backend verification.
- [ ] DigiLocker webview launches, completes OAuth handoff, and returns cleanly without certificate errors.
- [ ] Camera capture provides responsive framing and compresses images to <500KB within 200ms.
- [ ] Liveness selfie interface reliably guides user with intuitive visual prompts.
- [ ] Screen transitions maintain state across app backgrounding / orientation changes.
- [ ] Successful completion displays on-chain `ComplianceRegistry` whitelisting confirmation.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 202 (KYC/AML Service).
- **Enables:** Prompt 505 (Authentication UI) and Prompt 506 (Home Dashboard).
