# 514 - Flutter Settings, Profile & Regulatory Demat Management UI

## Purpose
The Settings and Profile Management module serves as the regulatory and operational control center for the investor within the Growww multi-platform client. In compliance with SEBI master circulars for digital brokerages and depository participants (NSDL/CDSL), investors require unambiguous visibility into their Demat account linkage (BOID, DP ID), KYC tier status, nominee registrations, linked primary/secondary bank accounts, security controls (Biometric auth, MPIN, 2FA hardware tokens), session device history, and mandatory statutory risk disclosures.

Within the Growww fractional asset-backed ecosystem, this module also gives investors direct visibility into their whitelisted on-chain identity attributes (via ERC-3643 `IdentityRegistry`), enabling transparent verification of their compliance standing across both domestic custody ledgers and the GIFT City international gateway.

## What You Are Building
A modular, highly responsive Flutter settings sub-system located at `lib/features/settings/` and `lib/features/profile/` with dedicated sub-screens and components:
- `ProfileOverviewScreen`: Personal KYC identity details (masked PAN, Aadhaar C-KYC reference, investor classification: Retail/HNI/NRI), email/phone verification badges.
- `DematBankLinkageScreen`: Depository Participant ID, Beneficiary Owner ID (BOID), eDIS T-PIN pre-authorization status, and penny-drop verified bank mandates.
- `NomineeManagementScreen`: SEBI-mandated nominee declaration list (up to 3 nominees with percentage share allocation, guardian details for minors, e-sign opt-in/opt-out status).
- `SecurityPrivacySettingsScreen`: Biometric authentication toggle (FaceID/Fingerprint), 6-digit MPIN rotation, Active Sessions manager with remote revocation, 2FA device setup.
- `PreferencesScreen`: Theme selector (Light/Dark/System OLED), default order execution preferences, language selection (delegating to Prompt 523), local cache management.
- `StatutoryDisclosuresScreen`: Interactive offline-accessible viewer for SEBI Investor Charter, Risk Disclosure Document for Capital Markets, Privacy Policy (DPDP Act 2023 compliant), and Terms of Service.
- `AccountClosureWorkflow`: Multi-step compliance-mandated account deactivation/closure flow requiring zero-holding and zero-balance verification.

## Scope Boundaries
- **In Scope:**
 - Complete UI/UX state management for settings, profile, nominees, security, session management, and statutory disclosures.
 - Integration with local secure storage (Prompt 521) for local biometric/MPIN preferences.
 - Form validation, dynamic field masking (PAN, bank account numbers, Aadhaar VID), and e-sign callback integration.
 - Active session device list display with single and all-device termination triggers.
- **Out of Scope / Handled Elsewhere:**
 - Backend user profile persistence and KYC document validation (Prompt 201, Prompt 202).
 - Depository linkage backend adapter (Prompt 213).
 - Cryptographic key vault management (Prompt 521).
 - Multi-language localization implementation files (Prompt 523).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using Flutter Riverpod (2.5+) for state management and `freezed` / `json_serializable` for immutable state and DTOs.
- **Justification:** Riverpod provides deterministic auto-dispose lifecycle management and async notifier state machines, ensuring sensitive profile data is cleanly purged from memory upon logout or background timeout.
- **Dependencies:**
 - `flutter_riverpod: ^2.5.1`
 - `go_router: ^14.0.0`
 - `flutter_secure_storage: ^9.0.0`
 - `local_auth: ^2.2.0` (biometric capabilities)
 - `package_info_plus: ^8.0.0` (client versioning and build metadata)
 - `url_launcher: ^6.2.6` (statutory external links and regulatory grievance portals)

## Backend / Infra Touchpoints
- **User & Profile Service (Prompt 201):** `GET /api/v1/user/profile`, `PATCH /api/v1/user/preferences`, `GET /api/v1/user/sessions`, `DELETE /api/v1/user/sessions/{id}`.
- **KYC & Demat Service (Prompt 202 / 213):** `GET /api/v1/kyc/demat-details`, `POST /api/v1/kyc/nominees`, `PUT /api/v1/kyc/nominees/{id}`.
- **Auth & Session Service (Prompt 201 / 105):** `POST /api/v1/auth/mpin/change`, `POST /api/v1/auth/sessions/revoke-all`.
- **Statutory CMS / S3 CloudFront:** CDN-backed static signed Markdown/PDF legal disclosure documents.

## Blockchain Interaction
- **Permissioned Ledger Identity Check:** Reads the investor's on-chain KYC/AML verification badge status and whitelisted address hash from the `ComplianceRegistry.sol` / `IdentityRegistry.sol` on Hyperledger Besu.
- **Proof-of-Reserve & Custody Link:** Renders the cryptographic verification link connecting the investor's internal account ID to their anonymized on-chain proof-of-holding Merkle leaf.
- **Zero Raw Key Exposure:** The client only queries public on-chain identity status via the API Gateway BFF (Prompt 219); no private transaction signing keys are held within standard profile views.

## Step-by-Step Build Instructions
1. Scaffold the feature directory structure under `lib/features/settings/` and `lib/features/profile/` containing `presentation/`, `domain/`, `data/`, and `controllers/`.
2. Define immutable domain models using Freezed: `UserProfile`, `DematDetails`, `Nominee`, `ActiveSession`, `SecurityPreferences`, and `AppThemeMode`.
3. Implement `ProfileRepository` interface and its Dio-backed data source with typed error handling for network failures and session timeouts.
4. Build `ProfileController` and `SettingsController` using `AsyncNotifier` to orchestrate fetching, optimistic state updating, and cache invalidation.
5. Create `ProfileOverviewScreen` featuring user avatar upload/picker, verified badges for KYC (PAN/Aadhaar verified), and masked tax identity fields.
6. Create `DematBankLinkageScreen` showing CDSL/NSDL 16-digit Demat Account number, DP Name, depository participant status, and linked bank accounts with primary badge.
7. Implement `NomineeManagementScreen` featuring dynamic allocation percentage sliders (totaling exactly 100%), minor guardian detail forms, and SEBI e-sign initiation CTA.
8. Build `SecurityPrivacySettingsScreen` with Biometric Login switch (integrating `local_auth`), Change MPIN wizard, and Two-Factor Authentication (TOTP/SMS) management.
9. Construct `ActiveSessionsListWidget` displaying OS/Device icon, IP location, last active timestamp, and a "Revoke Other Sessions" button.
10. Build `StatutoryDisclosuresScreen` with tabbed navigation for Investor Charter, Risk Disclosure, SEBI SCORES grievance link, and DPDP Act Privacy Notice.
11. Implement `AccountClosureFlow` with pre-flight checklist validation (zero INR balance, zero active token holdings, no open orders) before allowing submission.
12. Integrate theme mode switching (System / Light / Dark / High Contrast) with immediate reactive app-wide redraw.
13. Write comprehensive unit tests for `NomineeAllocationValidator` (enforcing 100% sum rule and minor guardian validation) and `SettingsController`.
14. Write golden widget tests across responsive mobile (Android/iOS) and desktop (Windows/Linux/macOS) layout widths (360px to 1440px).

## Interfaces / Contracts

```dart
// lib/features/profile/domain/models/user_profile.dart
import 'package:freezed_annotation/freezed_annotation.dart';

part 'user_profile.freezed.dart';
part 'user_profile.g.dart';

@freezed
class UserProfile with _$UserProfile {
  const factory UserProfile({
    required String userId,
    required String fullName,
    required String email,
    required String maskedPhone,
    required String maskedPan,
    required String kycTier, // 'TIER_1_DOMESTIC', 'TIER_2_GIFT_CITY', 'PENDING'
    required bool isKycVerified,
    required String onChainIdentityHash,
    required DematInfo dematInfo,
    required List<LinkedBankAccount> linkedBanks,
    required List<Nominee> nominees,
    required DateTime createdAt,
  }) = _UserProfile;

  factory UserProfile.fromJson(Map<String, dynamic> json) => _$UserProfileFromJson(json);
}

@freezed
class DematInfo with _$DematInfo {
  const factory DematInfo({
    required String depository, // 'NSDL' | 'CDSL'
    required String dpId,
    required String clientId,
    required String boid,
    required bool isEdisPreAuthorized,
  }) = _DematInfo;

  factory DematInfo.fromJson(Map<String, dynamic> json) => _$DematInfoFromJson(json);
}

@freezed
class Nominee with _$Nominee {
  const factory Nominee({
    String? id,
    required String name,
    required String relationship,
    required DateTime dateOfBirth,
    required double allocationPercentage, // Must sum to 100 across nominees
    required bool isMinor,
    String? guardianName,
    String? guardianPan,
  }) = _Nominee;

  factory Nominee.fromJson(Map<String, dynamic> json) => _$NomineeFromJson(json);
}

// Controller Contract
abstract class ProfileControllerContract {
  Future<void> refreshProfile();
  Future<void> updateNominees(List<Nominee> nominees);
  Future<void> updateSecurityPreferences({required bool biometricEnabled, required bool requireMpinForTrades});
  Future<void> revokeSession(String sessionId);
  Future<void> revokeAllOtherSessions();
}
```

## Security & Compliance Notes
- **SEBI & Depository Regulations:** Mask PAN (display only first 2 and last 2 characters e.g. `AB****123F`), mask Bank Account Number (show only last 4 digits), and enforce mandatory nominee declaration.
- **DPDP Act (Digital Personal Data Protection Act 2023):** User must be able to export profile data summary and initiate verified account closure with complete data lifecycle tracking.
- **Local Biometric Security:** Biometric state toggles must strictly invoke the device's hardware enclave / secure enclave through `local_auth` and persist encryption keys only in Keychain/KeyStore/DPAPI (Prompt 521).
- **Session Hygiene:** Token revocation requests must trigger immediate cache wipe of user PII from memory and local SQLite/Isar caches.

## Acceptance Criteria
- [ ] User profile renders all KYC, Demat, and masked tax identifiers correctly from backend payload.
- [ ] Nominee management enforces strict allocation percentage rules (totaling 100%) and valid guardian details for minors.
- [ ] Biometric and MPIN security settings persist state securely to platform secure storage and require biometric re-authentication to disable.
- [ ] Active sessions screen accurately lists current and other active sessions with working real-time revocation.
- [ ] Statutory disclosure screens render all mandatory SEBI/RBI investor charter and risk disclosure documents offline and online.
- [ ] Account closure wizard validates zero-balance and zero-holding requirements before submitting.
- [ ] Responsive UI adapts seamlessly across phone, tablet, and desktop (Windows/Linux/macOS) layout form factors.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 201 (User Service), Prompt 502 (App Architecture & State Management), Prompt 503 (Design System & Theming), Prompt 521 (Local Secure Storage).
- **Parallel Tasks:** Prompt 511 (Wallet & Funds Screen), Prompt 523 (Accessibility & Localization), Prompt 525 (API Client Layer).
