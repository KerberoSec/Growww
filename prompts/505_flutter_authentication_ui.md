# 505 - Flutter Authentication UI (Biometrics, MPIN, OTP, Session)

## Purpose
Provides bank-grade, multi-factor authentication (MFA) and seamless session management across mobile and desktop platforms. Because Growww manages real-money INR balances and regulated digital securities custody, the authentication experience must deliver uncompromising security - combining phone/email OTP, a 6-digit cryptographic MPIN, platform-native biometrics (FaceID, TouchID, Android BiometricPrompt, Windows Hello, macOS Touch ID), and automatic background session timeouts.

## What You Are Building
A complete authentication and session security module in `apps/growww_flutter/lib/features/auth/` containing:
- **Phone / Email Login Screen:** Form with country code selection, input formatting, and SMS/WhatsApp OTP delivery request.
- **OTP Verification Screen:** 6-digit numeric input with auto-fill (SMS Retriever API on Android, SMS autofill on iOS), countdown timer, and resend mechanics.
- **MPIN Setup & Unlock Screen:** Custom scrambled numeric pinpad with masked dots, biometric shortcut trigger, and brute-force attempt throttling.
- **Biometric Integration Manager:** Seamless wrapper across iOS FaceID/TouchID, Android BiometricPrompt (Class 3 strong biometrics), Windows Hello, and macOS Touch ID.
- **Session Lifecycle & Inactivity Guard:** Background listener that locks the UI after 5 minutes of inactivity or upon returning from background state.
- **Device Management & Trusted Hardware Binding:** UI list of registered active sessions with one-tap remote logout.

## Scope Boundaries
- **In Scope:**
 - Client-side login forms, OTP auto-fill integration, custom MPIN pinpad widget, biometric auth triggers, auto-lock overlay dialogs, and token refresh UI coordinators.
- **Out of Scope / Handled Elsewhere:**
 - Backend OAuth2/OIDC token generation, JWT issuance, and rate limiting (handled in Prompts 105, 201, 220).
 - Platform-native secure storage primitives (Keychain, Keystore, DPAPI) (handled in Prompt 521).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Riverpod for reactive authentication state.
 - *Justification:* Flutter's custom canvas rendering allows building a hardened, non-standard numeric keypad for MPIN entry, eliminating third-party keyboard keylogger vulnerabilities.
- **Biometrics:** `local_auth` plugin configured for biometric-only strong authentication.
- **SMS Autofill:** `pinput` and `sms_autofill` for instant, frictionless OTP capture.
- **Secure Persistence:** `flutter_secure_storage` interface for local hardware-backed refresh token storage.

## Backend / Infra Touchpoints
- **User Service:** REST endpoints (`/api/v1/auth/otp/send`, `/api/v1/auth/otp/verify`, `/api/v1/auth/mpin/login`, `/api/v1/auth/token/refresh`) from Prompt 201.
- **Notification Service:** Delivers SMS/WhatsApp OTP via backend gateway (Prompt 211).

## Blockchain Interaction
- **Wallet Address & Delegation Key Derivation:** Upon successful MPIN/biometric unlock, the client unlocks the user's localized signing key or session delegation token used to sign non-custodial inquiries and request DvP trade executions against `SettlementDvP.sol`.

## Step-by-Step Build Instructions
1. Scaffold auth directory: `lib/features/auth/presentation/`, `application/`, `domain/`, `data/`.
2. Define auth domain models: `AuthSession`, `UserCredentials`, `MpinValidationResult`, and `BiometricCapability`.
3. Implement `AuthNotifier` managing discrete state transitions: `unauthenticated`, `otpPending`, `mpinRequired`, `biometricPrompting`, `authenticated`, `locked`.
4. Build `LoginScreen` with country-code selector (+91 default) and phone/email input validation.
5. Implement `OtpVerificationScreen` using `Pinput` with paste detection, SMS autofill listeners, and 60-second cooldown timer.
6. Build custom `MpinKeypadWidget` with optional randomized button layout (to defeat shoulder-surfing) and tactile haptic feedback on every tap.
7. Implement `BiometricAuthService` querying available hardware biometrics and executing authentication with fallback to MPIN.
8. Build `SessionInactivityListener` wrapping the app root widget, detecting user tap/scroll events and triggering the lock overlay after 300 seconds of inactivity.
9. Implement `AppLockOverlayScreen` that covers the active UI with a blur effect whenever the app enters background or the inactivity timer fires.
10. Implement token refresh interceptor hook that seamlessly requests a new access token or prompts for MPIN re-authentication if the refresh token expires.
11. Build `ActiveSessionsScreen` allowing users to view and revoke active devices (IP address, device model, last active timestamp).
12. Write unit and widget tests for PIN validation, lock timers, and biometric fallback.

## Interfaces / Contracts
```dart
// lib/features/auth/domain/models/auth_state.dart
import 'package:freezed_annotation/freezed_annotation.dart';
part 'auth_state.freezed.dart';

@freezed
class AuthState with _$AuthState {
  const factory AuthState.initial() = _Initial;
  const factory AuthState.unauthenticated() = _Unauthenticated;
  const factory AuthState.otpSent({required String phoneNumber, required int expirySeconds}) = _OtpSent;
  const factory AuthState.mpinRequired({required String userId, required bool biometricAvailable}) = _MpinRequired;
  const factory AuthState.authenticated({required String userId, required String accessToken, required String userRole}) = _Authenticated;
  const factory AuthState.locked({required String userId}) = _Locked;
}

abstract class IBiometricService {
  Future<bool> isBiometricAvailable();
  Future<bool> authenticate({required String localizedReason});
}
```

## Security & Compliance Notes
- **Anti-Screen Capture / Blur:** Sensitive authentication screens (MPIN entry, OTP verification) must enable `FLAG_SECURE` on Android and apply a Gaussian blur overlay on iOS/macOS when the app is switched to the multitasking switcher.
- **Lockout on Failed Attempts:** After 3 consecutive failed MPIN attempts, enforce a mandatory 5-minute cooldown and notify the user via SMS/Email.
- **Zero In-Memory Storage of Raw MPIN:** The MPIN is immediately hashed with a salt or passed directly to secure enclaves; plain text strings are wiped from RAM.

## Acceptance Criteria
- [ ] Phone OTP login successfully sends, auto-fills SMS OTP, and advances to MPIN/Biometric setup.
- [ ] MPIN entry with custom scrambled keypad registers and validates inputs seamlessly.
- [ ] FaceID, TouchID, and Windows Hello biometric prompts execute and authenticate properly.
- [ ] App automatically locks after 5 minutes of inactivity or when resumed from background.
- [ ] Screenshots are prevented/blacked out on sensitive authentication views.
- [ ] Session refresh happens silently in the background without disturbing active user tasks.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System), Prompt 521 (Secure Storage).
- **Backend Dependency:** Prompt 201 (User Service), Prompt 105 (Auth Architecture).
- **Enables:** Prompts 506-515 (Authenticated user feature screens).
