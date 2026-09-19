# 533 - Flutter FIDO2 Passkeys & Biometric Hardware Key Signing Flow

## Purpose
Provides phishing-resistant, passwordless investor authentication and cryptographic trade authorization using device Secure Enclaves (Apple Secure Enclave on iOS/macOS via FaceID/TouchID, Android StrongBox / Trusted Execution Environment (TEE) via BiometricPrompt, Windows Hello via TPM 2.0, and external hardware security keys via FIDO2/WebAuthn USB-C/NFC/BLE such as YubiKey 5 Series).

Traditional SMS OTPs are vulnerable to SIM swapping, SS7 interception, and telecom latency, while static passwords and MPINs are susceptible to credential stuffing and phishing attacks. In high-velocity digital asset and tokenized securities markets where real INR capital is committed to settlement contracts, authentication must guarantee non-repudiation and origin binding. The Flutter FIDO2 Passkeys and Biometric Hardware Key Signing subsystem delivers seamless public-key cryptography to retail and institutional investors. By combining Possession (hardware-bound private key) with Inherence (biometric sensor verification) into a single cryptographic assertion, it eliminates credential theft, fulfills SEBI mandatory 2FA requirements, and enables cryptographic signing for trade execution and ERC-4337 smart contract wallet transactions.

## What You Are Building
A comprehensive Flutter passkey client implementation and ceremony coordinator module located at `apps/growww_flutter/lib/features/passkeys/` along with platform-native channel adapters across Android, iOS, macOS, and Windows:
- **Passkey Registration Ceremony Coordinator (`PasskeyRegistrationCoordinator`):** Orchestrates the WebAuthn credential creation ceremony via native platform credential APIs, handling RP ID verification, challenge transport, and attestation payload packaging.
- **Passkey Authentication Ceremony Coordinator (`PasskeyAuthenticationCoordinator`):** Manages user login, session renewal, and step-up authentication ceremonies, converting native platform assertions into structured verification payloads for the backend.
- **Cryptographic Trade Signing Ceremony (`HardwareTradeSigningService`):** Computes a SHA-256 canonical order intent digest (binding ISIN, quantity, price, side, timestamp, and investor ID), requests a scoped challenge, and triggers an authenticated biometric assertion that cryptographically commits the investor to the exact trade parameters.
- **ERC-4337 Smart Account UserOperation Signer (`WebAuthnUserOpSigner`):** Unpacks the authenticator data, client data JSON, and DER-encoded ECDSA signature from the hardware assertion, extracting raw `(r, s)` coordinates over the NIST P-256 (secp256r1) curve for on-chain validation.
- **Passkey Management Screen (`PasskeyManagementScreen`):** User interface displaying enrolled credentials, device metadata (AAGUID, authenticator model, transport capabilities: internal, USB, NFC, BLE), sync status (iCloud Keychain / Google Password Manager vs hardware-bound security key), and credential revocation controls.
- **Hardware Security Key (YubiKey) Enrollment & Interaction Sheet (`SecurityKeyPromptSheet`):** Interactive modal guiding institutional and high-net-worth users through physical USB-C insertion or NFC tap interactions for FIPS 140-3 Level 3 compliant keys.
- **Biometric Trade Confirmation Sheet (`BiometricTradeSigningDialog`):** Hardened transaction authorization sheet presenting trade terms and invoking native biometric sensors with active anti-tampering guards.
- **Riverpod State Management Layer:** Predictable, reactive state controllers managing registration, authentication, trade signing ceremonies, and credential inventory caching.

## Scope Boundaries
- **In Scope:**
  - Flutter client-side passkey ceremony coordination (Registration, Authentication, and Step-Up Trade Signing).
  - Platform channel bridging to Android Credential Manager API (`androidx.credentials`), iOS AuthenticationServices (`ASAuthorizationPlatformPublicKeyCredentialProvider`), Windows WebAuthn API (`webauthn.dll`), and macOS LocalAuthentication.
  - Client-side CBOR decoding of authenticator data, attestation objects, and extraction of credential public keys.
  - Conversion of ASN.1 DER-encoded ECDSA signatures to raw IEEE P1363 `(r, s)` 64-byte format and normalization for low-s Ethereum/Besu compliance.
  - Computation of client data JSON and SHA-256 order intent digests for challenge binding.
  - UI screens for credential enrollment, device inventory management, trade confirmation, and physical security key pairing.
  - Graceful fallback flows to MPIN + SMS OTP when hardware passkeys are unsupported or fail.
- **Out of Scope / Handled Elsewhere:**
  - Backend FIDO2 Relying Party (RP) server verification, challenge issuance, and public key persistence (handled in User Service Prompt 201 and Auth Architecture Prompt 105).
  - Order matching and execution engine logic (handled in Order Service Prompt 204 and Order Matching Engine Prompt 205).
  - Basic password/MPIN input styling and SMS autofill listeners (handled in Prompt 505).
  - Encrypted local storage of refresh tokens and cached user profiles (handled in Prompt 521).
  - ERC-4337 bundler, paymaster RPC services, and on-chain P-256 verifier smart contract deployment (handled in Prompt 253 and Prompt 309).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ utilizing `flutter_riverpod: ^2.5.1` for predictable, state-driven ceremony workflows.
  - *Justification:* Flutter allows unified business logic for WebAuthn challenge processing while delegating sensitive UI prompts directly to native OS security dialogs, ensuring zero touch of biometric templates by application code.
- **Native Platform Channels & Credential Providers:**
  - *Android:* Kotlin platform channel interfacing with Android Credential Manager API (`androidx.credentials:credentials:1.3.0` and `androidx.credentials:credentials-play-services-auth:1.3.0`). Supports passkeys in Google Password Manager and FIDO2 physical keys via USB/NFC.
  - *iOS & macOS:* Swift platform channel interfacing with Apple `AuthenticationServices` (`ASAuthorizationController`, `ASAuthorizationPlatformPublicKeyCredentialProvider`, and `ASAuthorizationSecurityKeyPublicKeyCredentialProvider`). Supports iCloud Keychain passkeys and external security keys.
  - *Windows:* C++ Win32 platform channel wrapper integrating directly with `webauthn.dll` (Windows WebAuthn API v6+) for Windows Hello (TPM 2.0 biometric/PIN) and external FIDO2 tokens.
- **Binary Encoding & Cryptography:**
  - `cbor: ^6.3.0`: For parsing binary CBOR structures in `attestationObject` and extracting COSE public keys.
  - `crypto: ^3.0.3`: For client-side SHA-256 hashing of trade parameters and client data structures.
  - `asn1lib: ^1.5.1`: For parsing DER-encoded ECDSA signatures output by authenticators and extracting discrete `r` and `s` integers.
- **Transport & Networking:** `dio: ^5.4.3+1` for communicating with the backend FIDO2 Relying Party endpoints.

## Backend / Infra Touchpoints
- **User Service (Prompt 201):**
  - `POST /api/v1/auth/passkey/register/options`: Retrieves `PublicKeyCredentialCreationOptions` (server challenge, RP ID, user ID, username, supported algorithms: ES256 / COSE -7, authenticator criteria).
  - `POST /api/v1/auth/passkey/register/verify`: Transmits client attestation response (`credentialId`, `clientDataJSON`, `attestationObject`, `transports`) for cryptographic verification and credential storage.
  - `POST /api/v1/auth/passkey/authenticate/options`: Retrieves `PublicKeyCredentialRequestOptions` (server challenge, RP ID, allowCredentials list, userVerification: `required`).
  - `POST /api/v1/auth/passkey/authenticate/verify`: Submits assertion response (`credentialId`, `clientDataJSON`, `authenticatorData`, `signature`, `userHandle`) to obtain authenticated session JWTs.
  - `GET /api/v1/user/passkeys`: Returns enrolled credentials list for the authenticated user, including AAGUID, creation timestamp, and last-used timestamp.
  - `DELETE /api/v1/user/passkeys/{credential_id}`: Revokes an enrolled passkey credential.
- **Authentication & Authorization Architecture (Prompt 105):**
  - Validates cryptographic proof of possession, tracks device hardware security levels, and issues hardware-bound session claims.
- **Order Service (Prompt 204):**
  - `POST /api/v1/orders/signing-challenge`: Submits an order draft to obtain a cryptographically bound challenge containing the canonical SHA-256 order intent digest.
  - `POST /api/v1/orders/execute-with-passkey`: Submits the finalized order accompanied by the FIDO2 assertion signature for verification prior to routing to the matching engine.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **ERC-4337 Account Abstraction & Secp256r1 (NIST P-256) Verification:**
  - Standard Ethereum accounts use secp256k1 keys. Mobile Secure Enclaves and FIDO2 authenticators generate signatures over the NIST P-256 (secp256r1) curve.
  - Growww investor wallets on Hyperledger Besu are deployed as ERC-4337 compliant smart contract accounts (`InvestorSmartAccount.sol`).
  - The Flutter client maps the FIDO2 assertion to an ERC-4337 `UserOperation`. The signature field is formatted with the WebAuthn assertion payload:
    $$\text{UserOp.signature} = \text{abi.encode}(authenticatorData, clientDataJSON, challengeLocation, responseTypeLocation, r, s)$$
  - Hyperledger Besu validates the signature using the RIP-7212 P-256 precompile (at address `0x0000000000000000000000000000000000000100`) or an optimized on-chain P-256 verifier, confirming non-repudiable trade authorization directly on the ledger.
- **Session Key Delegation for High-Frequency Actions:**
  - For rapid portfolio rebalancing or active order cancellations, the passkey signs an on-chain delegation transaction that authorizes a temporary, scoped session key stored in the device secure storage (`Prompt 521`), bounded by time (e.g., 60 minutes) and maximum cumulative INR spending limits.
- **Zero On-Chain PII Guarantee:**
  - Public keys ($Q_x, Q_y$ coordinates) and credential identifiers contain zero investor personal data (no PAN, Aadhaar, email, or name). All identity bindings remain strictly off-chain within the SEBI-regulated User Service (Prompt 201).

## Passkeys & Hardware Key Signing Architecture & Flow

### 1. Registration Ceremony (Passkey Enrollment)
```
+---------------+              +--------------------+              +--------------------+
|  Growww App   |              | Native OS Enclave  |              | Backend User Svc   |
| (Flutter UI)  |              | (Passkey Provider) |              |  (FIDO2 RP Server) |
+-------+-------+              +---------+----------+              +---------+----------+
        |                                |                                   |
        | 1. Request Register Options    |                                   |
        +------------------------------------------------------------------->|
        |                                |                                   |
        | 2. Return CreationOptions (Challenge, RP ID, User ID, Params)     |
        |<-------------------------------------------------------------------+
        |                                |                                   |
        | 3. Invoke Platform Channel     |                                   |
        |    (createCredentialRequest)   |                                   |
        +------------------------------->|                                   |
        |                                |                                   |
        |                                | 4. Prompt Biometric (Face/Touch)  |
        |                                |    Generate P-256 Keypair         |
        |                                |    Sign Attestation Object        |
        |                                |                                   |
        | 5. Return Credential Attestation                                  |
        |<-------------------------------+                                   |
        |                                                                    |
        | 6. Submit Attestation (clientDataJSON, attestationObject)          |
        +------------------------------------------------------------------->|
        |                                                                    |
        |                                7. Verify Attestation, Register Key |
        | 8. Registration Confirmed (200 OK)                                |
        |<-------------------------------------------------------------------+
```

### 2. Trade Signing Ceremony (Cryptographic Order Authorization)
```
+---------------+              +--------------------+              +--------------------+
|  Growww App   |              | Native OS Enclave  |              | Backend Order Svc  |
| (Flutter UI)  |              | (Passkey Provider) |              | (Challenge & Exec) |
+-------+-------+              +---------+----------+              +---------+----------+
        |                                |                                   |
        | 1. Submit Order Draft for Signing Challenge                        |
        +------------------------------------------------------------------->|
        |                                |                                   |
        |                                | 2. Calculate Order Intent Hash:   |
        |                                |    H = SHA256(ISIN|Qty|Price|...) |
        |                                |    Generate Scoped Challenge      |
        |                                |                                   |
        | 3. Return RequestOptions (Challenge = H, AllowCredentials)         |
        |<-------------------------------------------------------------------+
        |                                |                                   |
        | 4. Invoke Platform Channel     |                                   |
        |    (getCredentialRequest)      |                                   |
        +------------------------------->|                                   |
        |                                |                                   |
        |                                | 5. Prompt Biometric Verification  |
        |                                |    Sign Challenge with P-256 Key  |
        |                                |    Assert User Presence & Verify  |
        |                                |                                   |
        | 6. Return Assertion (authenticatorData, clientDataJSON, signature) |
        |<-------------------------------+                                   |
        |                                                                    |
        | 7. Submit Order with Hardware Assertion Signature                  |
        +------------------------------------------------------------------->|
        |                                                                    |
        |                                8. Verify P-256 Signature over H    |
        |                                   Submit to Matching / Besu Ledger |
        | 9. Order Accepted & Dispatched (201 Created)                       |
        |<-------------------------------------------------------------------+
```

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:** Create `apps/growww_flutter/lib/features/passkeys/` with subdirectories `presentation/screens/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, `data/datasources/`, and `data/platform/`.
2. **Implement Android Native Platform Channel (`PasskeyChannelPlugin.kt`):** Author platform channel handler interfacing with Android `CredentialManager`. Implement `createPasskey` using `CreatePublicKeyCredentialRequest` and `getPasskeyAssertion` using `GetPublicKeyCredentialRequest`, serializing JSON responses back across the MethodChannel.
3. **Implement iOS/macOS Native Platform Channel (`PasskeyChannelPlugin.swift`):** Author platform channel handler utilizing Apple `AuthenticationServices`. Implement `ASAuthorizationControllerDelegate` to handle `ASAuthorizationPlatformPublicKeyCredentialRegistration` and `ASAuthorizationPlatformPublicKeyCredentialAssertion`, formatting credential outputs as JSON dictionaries.
4. **Implement Windows Native Platform Channel (`webauthn_plugin.cpp`):** Author C++ Win32 wrapper linking against `webauthn.dll`. Implement `WebAuthNAuthenticatorMakeCredential` and `WebAuthNAuthenticatorGetAssertion`, exposing them to Dart via platform channel or FFI.
5. **Author Domain Models & Data Transfer Objects:** Define immutable Dart domain models in `domain/models/`: `PasskeyCredential`, `PasskeyCreationOptions`, `PasskeyRequestOptions`, `AuthenticatorAttestation`, `AuthenticatorAssertion`, `COSEPublicKey`, and `PasskeyDeviceMetadata`.
6. **Implement CBOR & Cryptographic Utilities:** Author `CborDecoderUtil` and `DerSignatureUtil` to parse authenticator data binary fields, extract flags (`UP` User Present, `UV` User Verified, `AT` Attested Credential Data), and convert ASN.1 DER ECDSA signatures to standard 64-byte raw `(r, s)` values.
7. **Implement Low-s Signature Normalization:** Implement mathematical curve order check: if $s > \frac{N}{2}$ where $N$ is the NIST P-256 curve order, normalize $s' = N - s$ to ensure signature validity across both backend verifiers and on-chain ERC-4337 smart contracts.
8. **Build Passkey Repository (`PasskeyRepository`):** Implement API client methods connecting to User Service (Prompt 201) to fetch creation options, verify attestation, fetch authentication options, submit assertions, and list or delete registered credentials.
9. **Implement Registration Ceremony Notifier (`PasskeyRegistrationNotifier`):** Author Riverpod `StateNotifier` or `AsyncNotifier` orchestrating the passkey registration lifecycle through discrete states: `idle`, `requestingOptions`, `waitingForBiometric`, `submittingAttestation`, `registered`, and `error`.
10. **Implement Authentication Ceremony Notifier (`PasskeyAuthenticationNotifier`):** Author Riverpod state machine managing passkey login and session re-authentication.
11. **Implement Trade Signing Service (`HardwareTradeSigningService`):** Build the trade signing coordinator that takes an order draft, formats canonical JSON, requests a challenge from Order Service (Prompt 204), invokes the hardware passkey prompt, and returns the signed trade payload.
12. **Implement ERC-4337 UserOperation Formatter (`WebAuthnUserOpSigner`):** Construct ABI-encoded signature payloads containing `authenticatorData`, `clientDataJSON`, challenge byte index, and normalized `(r, s)` coordinates for Hyperledger Besu smart account execution.
13. **Build Passkey Management Screen (`PasskeyManagementScreen`):** Build responsive settings view displaying enrolled passkeys, hardware badges (TouchID, FaceID, Windows Hello, Security Key), creation dates, backup sync indicators, and credential revocation confirmation dialogs.
14. **Build Biometric Trade Signing Dialog (`BiometricTradeSigningDialog`):** Implement transaction authorization bottom sheet / modal featuring order parameter summary, biometric trigger button, security audit badges, and fallback button.
15. **Build Physical Security Key (YubiKey) Prompt Sheet (`SecurityKeyPromptSheet`):** Author dedicated modal guiding users through USB-C insertion or NFC contact, including animated connection indicators and FIDO2 PIN entry prompts.
16. **Write Comprehensive Test Suite:** Author unit tests for CBOR parsing, DER signature conversion, and low-s normalization. Write mock platform channel tests verifying cancellation and error handling, and Golden UI tests for trade confirmation and management screens.

## Interfaces / Contracts

### Protobuf Service Definition (`passkey_service.proto`)
```protobuf
syntax = "proto3";

package growww.auth.v1;

option go_package = "growww/auth/v1;authv1";
option java_multiple_files = true;
option java_package = "com.growww.auth.v1";

service PasskeyService {
  rpc GetRegisterOptions(PasskeyRegisterOptionsRequest) returns (PasskeyRegisterOptionsResponse);
  rpc VerifyRegistration(PasskeyRegisterVerifyRequest) returns (PasskeyRegisterVerifyResponse);
  rpc GetAuthOptions(PasskeyAuthOptionsRequest) returns (PasskeyAuthOptionsResponse);
  rpc VerifyAuthentication(PasskeyAuthVerifyRequest) returns (PasskeyAuthVerifyResponse);
  rpc ListUserPasskeys(ListPasskeysRequest) returns (ListPasskeysResponse);
  rpc RevokePasskey(RevokePasskeyRequest) returns (RevokePasskeyResponse);
}

message PasskeyRegisterOptionsRequest {
  string user_id = 1;
  string display_name = 2;
  string authenticator_attachment = 3; // "platform" or "cross-platform"
}

message PasskeyRegisterOptionsResponse {
  string challenge = 1;
  string rp_id = 2;
  string rp_name = 3;
  string user_id = 4;
  string user_name = 5;
  repeated int32 supported_algorithms = 6; // e.g. -7 for ES256
  int64 timeout_ms = 7;
  string attestation_preference = 8;
  repeated string exclude_credential_ids = 9;
}

message PasskeyRegisterVerifyRequest {
  string user_id = 1;
  string credential_id = 2;
  string raw_id = 3;
  string client_data_json = 4;
  string attestation_object = 5;
  repeated string transports = 6;
  string client_device_name = 7;
}

message PasskeyRegisterVerifyResponse {
  bool success = 1;
  string credential_id = 2;
  string aaguid = 3;
  string registered_at = 4;
}

message PasskeyAuthOptionsRequest {
  string user_id = 1; // Optional for discoverable credentials
  string purpose = 2; // "LOGIN", "STEP_UP", "TRADE_CONFIRMATION"
  string order_intent_digest = 3; // Optional SHA-256 hash of trade payload
}

message PasskeyAuthOptionsResponse {
  string challenge = 1;
  string rp_id = 2;
  int64 timeout_ms = 3;
  string user_verification = 4; // "required"
  repeated string allowed_credential_ids = 5;
}

message PasskeyAuthVerifyRequest {
  string credential_id = 1;
  string raw_id = 2;
  string client_data_json = 3;
  string authenticator_data = 4;
  string signature = 5;
  string user_handle = 6;
  string order_intent_digest = 7;
}

message PasskeyAuthVerifyResponse {
  bool authenticated = 1;
  string session_token = 2;
  string trade_authorization_token = 3;
  int64 expires_at_unix = 4;
}

message ListPasskeysRequest {
  string user_id = 1;
}

message PasskeyDeviceItem {
  string credential_id = 1;
  string friendly_name = 2;
  string aaguid = 3;
  string platform_type = 4; // "APPLE_SECURE_ENCLAVE", "ANDROID_STRONGBOX", "WINDOWS_HELLO", "YUBIKEY"
  bool is_synced = 5;
  int64 created_at_unix = 6;
  int64 last_used_at_unix = 7;
}

message ListPasskeysResponse {
  repeated PasskeyDeviceItem passkeys = 1;
}

message RevokePasskeyRequest {
  string credential_id = 1;
}

message RevokePasskeyResponse {
  bool revoked = 1;
  string revoked_at = 2;
}
```

### Dart Domain Models & State Contracts (`lib/features/passkeys/domain/models/passkey_models.dart`)
```dart
enum PasskeyCeremonyType {
  registration,
  authentication,
  tradeSigning,
}

enum HardwareAuthenticatorType {
  appleSecureEnclave,
  androidStrongBoxOrTee,
  windowsHello,
  hardwareSecurityKey, // YubiKey, Nitrokey
  unknown,
}

class PasskeyCredentialInfo {
  final String credentialId;
  final String friendlyName;
  final String aaguid;
  final HardwareAuthenticatorType authenticatorType;
  final bool isSynced;
  final DateTime createdAt;
  final DateTime? lastUsedAt;

  const PasskeyCredentialInfo({
    required this.credentialId,
    required this.friendlyName,
    required this.aaguid,
    required this.authenticatorType,
    required this.isSynced,
    required this.createdAt,
    this.lastUsedAt,
  });
}

class WebAuthnAssertionPayload {
  final String credentialId;
  final String rawIdBase64;
  final String clientDataJsonBase64;
  final String authenticatorDataBase64;
  final String signatureBase64;
  final String? userHandleBase64;
  final List<int> rCoordinate;
  final List<int> sCoordinate;

  const WebAuthnAssertionPayload({
    required this.credentialId,
    required this.rawIdBase64,
    required this.clientDataJsonBase64,
    required this.authenticatorDataBase64,
    required this.signatureBase64,
    this.userHandleBase64,
    required this.rCoordinate,
    required this.sCoordinate,
  });
}

class OrderIntentDigest {
  final String isin;
  final String side; // "BUY" or "SELL"
  final double quantity;
  final double priceInr;
  final int timestampNonce;
  final String sha256DigestHex;

  const OrderIntentDigest({
    required this.isin,
    required this.side,
    required this.quantity,
    required this.priceInr,
    required this.timestampNonce,
    required this.sha256DigestHex,
  });
}

abstract class IPasskeyPlatformChannel {
  Future<bool> isPasskeySupported();
  Future<bool> isHardwareSecurityKeySupported();
  
  Future<String> createPasskeyCredential({
    required String requestJson,
  });

  Future<String> getPasskeyAssertion({
    required String requestJson,
  });
}

abstract class IPasskeyRepository {
  Future<Map<String, dynamic>> fetchRegisterOptions({
    required String authenticatorAttachment,
  });

  Future<bool> submitRegisterAttestation({
    required Map<String, dynamic> attestationData,
    required String friendlyDeviceName,
  });

  Future<Map<String, dynamic>> fetchAuthOptions({
    String? orderIntentDigest,
    required String purpose,
  });

  Future<WebAuthnAssertionPayload> parseAndNormalizeAssertion({
    required String rawPlatformResponseJson,
  });

  Future<String> submitTradeAssertion({
    required WebAuthnAssertionPayload assertion,
    required OrderIntentDigest orderDigest,
  });

  Future<List<PasskeyCredentialInfo>> listUserPasskeys();
  Future<bool> revokePasskey(String credentialId);
}
```

### Riverpod State Definition (`lib/features/passkeys/presentation/controllers/passkey_states.dart`)
```dart
class PasskeyRegistrationState {
  final bool isLoading;
  final String? activeStatusMessage;
  final String? errorMessage;
  final PasskeyCredentialInfo? createdCredential;

  const PasskeyRegistrationState({
    this.isLoading = false,
    this.activeStatusMessage,
    this.errorMessage,
    this.createdCredential,
  });

  PasskeyRegistrationState copyWith({
    bool? isLoading,
    String? activeStatusMessage,
    String? errorMessage,
    PasskeyCredentialInfo? createdCredential,
  }) {
    return PasskeyRegistrationState(
      isLoading: isLoading ?? this.isLoading,
      activeStatusMessage: activeStatusMessage,
      errorMessage: errorMessage,
      createdCredential: createdCredential ?? this.createdCredential,
    );
  }
}

class TradeSigningState {
  final bool isSigning;
  final OrderIntentDigest? orderDigest;
  final String? signedAuthToken;
  final String? errorMessage;
  final bool isSuccess;

  const TradeSigningState({
    this.isSigning = false,
    this.orderDigest,
    this.signedAuthToken,
    this.errorMessage,
    this.isSuccess = false,
  });

  TradeSigningState copyWith({
    bool? isSigning,
    OrderIntentDigest? orderDigest,
    String? signedAuthToken,
    String? errorMessage,
    bool? isSuccess,
  }) {
    return TradeSigningState(
      isSigning: isSigning ?? this.isSigning,
      orderDigest: orderDigest ?? this.orderDigest,
      signedAuthToken: signedAuthToken ?? this.signedAuthToken,
      errorMessage: errorMessage,
      isSuccess: isSuccess ?? this.isSuccess,
    );
  }
}
```

## Security & Compliance Notes
- **FIPS 140-2 / FIPS 140-3 Hardware Key Protection:** All private keys generated during the registration ceremony are created within tamper-resistant hardware modules (Apple Secure Enclave, Android StrongBox/TEE, or external FIPS 140-3 Level 3 YubiKeys). Private key bits can never be exported, read, or modified by the Growww app or the underlying operating system.
- **Cryptographic Phishing Resistance:** WebAuthn guarantees origin binding. The client platform calculates the clientDataJSON hash containing the active Relying Party ID (`growww.in`). A spoofed phishing website or man-in-the-middle proxy cannot replay assertions because the cryptographic signature covers the authentic origin domain.
- **SEBI Mandatory 2FA Adherence:** Under SEBI trading guidelines, two distinct authentication factors are mandatory for stock broker access. Passkeys provide both the Possession Factor (hardware private key bound to device) and the Inherence Factor (biometric face/fingerprint validation) in an indivisible cryptographic proof. User Verification (`UV=1` flag in authenticator data) is strictly verified on every trade assertion.
- **Order Parameter Tamper-Proofing (Order Intent Binding):** For trade authorizations, the challenge issued by the Order Service is not an arbitrary random string, but a SHA-256 digest of canonical order attributes:
  $$\text{Digest} = \text{SHA-256}(\text{ISIN} \parallel \text{Side} \parallel \text{Quantity} \parallel \text{Price} \parallel \text{Nonce} \parallel \text{Timestamp})$$
  Because this digest is embedded in the signed clientDataJSON, any modification of order price or quantity invalidates the cryptographic signature.
- **Signature Malleability & Low-s Enforcement:** In standard ECDSA signatures over NIST P-256, $(r, s)$ and $(r, N - s)$ are both mathematically valid signatures for curve order $N$. To prevent transaction malleability attacks on Hyperledger Besu ERC-4337 smart contracts, the client normalizes all signatures to their lower $s$-value ($s \le \frac{N}{2}$) before submission.
- **Biometric Spoof & Liveness Detection:** On Android, `BiometricPrompt` is configured with `BIOMETRIC_STRONG` (Class 3), requiring hardware-level liveness verification and preventing 2D photo spoofing. On iOS, `LocalAuthentication` relies on TrueDepth 3D infrared dot projection.
- **Anti-Screen Capture & Memory Zeroing:** The passkey management and biometric signing screens enforce `FLAG_SECURE` on Android and display privacy blurs on iOS task switchers. Any decoded signature coordinate byte arrays in Dart memory are explicitly zeroed out immediately following transmission.

## Acceptance Criteria
- [ ] Android native platform channel successfully invokes Android Credential Manager API and completes passkey creation and assertion.
- [ ] iOS native platform channel successfully invokes `ASAuthorizationController` and returns valid P-256 attestation and assertion payloads.
- [ ] Windows native channel interfaces with `webauthn.dll` to execute Windows Hello TPM 2.0 biometric signing.
- [ ] External FIDO2 hardware security keys (YubiKey) are detected and usable via NFC and USB-C across supported platforms.
- [ ] CBOR parser accurately extracts AAGUID, credential ID, and COSE public key from binary attestation objects.
- [ ] ASN.1 DER-encoded signatures are correctly unpacked into 32-byte `r` and 32-byte `s` coordinates and normalized to low-s format ($s \le \frac{N}{2}$).
- [ ] Trade signing ceremony computes canonical SHA-256 order intent digest and binds it within the assertion challenge.
- [ ] ERC-4337 `UserOperation.signature` formatting complies with Hyperledger Besu on-chain P-256 precompile requirements.
- [ ] Passkey management screen lists active credentials, displays hardware authenticator badges, and enables one-tap credential revocation.
- [ ] Biometric trade authorization modal renders trade parameters clearly and triggers native biometric prompt on tap.
- [ ] Offline or failed biometric ceremonies fall back gracefully to MPIN + SMS OTP with appropriate security warnings.
- [ ] Specification strictly adheres to the 12-section template, contains zero raw application code outside specifications, and uses standard ASCII hyphens exclusively.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 501 (Flutter Project Scaffolding)
  - Prompt 502 (Flutter State Management Architecture)
  - Prompt 503 (Design System & Theming)
  - Prompt 505 (Authentication UI)
  - Prompt 521 (Local Secure Storage)
- **Backend Dependencies:**
  - Prompt 105 (Authentication & Authorization Architecture)
  - Prompt 201 (User Service)
  - Prompt 204 (Order Service)
- **Downstream Feature Dependents:**
  - Prompt 509 (Flutter Order Placement Flow - triggers passkey trade signing)
  - Prompt 514 (Flutter Settings & Profile UI - hosts passkey management view)
  - Prompt 528 (Flutter Multichain Crypto & Asset Deposit Screen - high-value withdrawal signing)
