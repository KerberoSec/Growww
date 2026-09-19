# 521 - Local Secure Storage (Keychain, KeyStore, DPAPI & Libsecret Abstraction) for Secrets

## Purpose
Fintech trading terminals and digital investment clients store high-sensitivity secrets on client hardware: OAuth2/OIDC refresh tokens, session JWTs, device-bound asymmetric cryptographic keys, Drift database encryption master seeds, and biometric trading authorization nonces. Storing these credentials in plaintext or standard shared preferences violates SEBI cybersecurity frameworks, RBI digital lending guidelines, and ISO 27001 data protection standards.

This prompt specifies the unified, multi-platform Local Secure Storage architecture for the Growww Flutter client. It establishes an abstraction layer wrapping platform-native hardware security modules: **Android KeyStore** (with `EncryptedSharedPreferences` and MasterKey AES-256-GCM), **iOS/macOS Keychain** (with `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly`), **Windows DPAPI** (Data Protection API with machine-and-user entropy), and **Linux Libsecret** (GNOME Keyring / KWallet). It also implements in-memory key wiping, biometric cryptographic gating, and automated fallback error handling.

## What You Are Building
A high-security storage infrastructure located in `lib/core/security/storage/`:
- `SecureStorageService`: Core Dart interface and implementation providing type-safe, asynchronous read, write, delete, and contains operations for sensitive tokens and configuration flags.
- `PlatformSecureStorageAdapter`: Multi-platform native bridge wrapping `flutter_secure_storage` and custom FFI extensions for platform-specific hardware encryption.
- `BiometricGatedStorageService`: Advanced wrapper requiring immediate biometric prompt (FaceID / Fingerprint) before releasing high-privilege secrets (e.g. trade signing tokens).
- `SecureMemoryBuffer`: Zero-allocation memory structure using `dart:ffi` / typed byte buffers that explicitly overwrites byte arrays with random entropy upon garbage collection or user logout.
- `DatabaseMasterKeyManager`: Ephemeral key generator and manager that derives and securely provisions 256-bit AES keys for the Drift SQLCipher database (Prompt 515).

## Scope Boundaries
- **In Scope:**
 - Native secure storage integration across Android, iOS, Windows, Linux, and macOS.
 - Hardware-backed key isolation (Secure Enclave, StrongBox Keymaster, TPM, DPAPI).
 - Biometric authentication gating before secret retrieval.
 - Secure in-memory buffer handling and zeroization.
 - Key migration and corruption recovery strategies.
- **Out of Scope / Handled Elsewhere:**
 - Backend user authentication services (Prompt 201).
 - UI screens for biometric settings (Prompt 505, Prompt 514).
 - Network API token refresh lifecycle (Prompt 525).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_secure_storage` (v9.0+), `local_auth` (v2.2+), and `crypto` (v3.0+).
- **Justification:** `flutter_secure_storage` interfaces directly with Android KeyStore (AES-256 GCM) and Apple Keychain, while providing Windows DPAPI and Linux Libsecret desktop support with zero external binary dependencies.
- **Dependencies:**
 - `flutter_secure_storage: ^9.2.2`
 - `local_auth: ^2.2.0`
 - `cryptography: ^2.7.0`
 - `flutter_riverpod: ^2.5.1`

## Backend / Infra Touchpoints
- **Auth & Token Refresh Flow (Prompt 525):** Provides persistent storage for OAuth2 refresh tokens and client device certificates.
- **Drift Local Database (Prompt 515):** Supplies on-device master encryption key for SQLCipher local offline tables.

## Blockchain Interaction
- **Device-Bound Transaction Signer:** Securely holds client-side cryptographic ephemeral keypairs (Ed25519 or secp256k1) used to sign client intent messages before dispatching to the Hyperledger Besu JSON-RPC relayer.
- **Zero Raw Private Key Exposure:** Private keys are generated in hardware-backed storage where available and never exported as plaintext strings to logging frameworks or application memory dumps.

## Step-by-Step Build Instructions
1. Scaffold `lib/core/security/storage/` with `domain/`, `data/`, and `adapters/` folders.
2. Define `ISecureStorageService` interface with asynchronous methods: `writeString()`, `readString()`, `writeBytes()`, `readBytes()`, `delete()`, `clearAll()`, and `contains()`.
3. Configure platform-specific options for `FlutterSecureStorage`:
 - **Android:** `AndroidOptions(encryptedSharedPreferences: true, resetOnError: true, keyCipherAlgorithm: KeyCipherAlgorithm.RSA_ECB_OAEPwithSHA_256andMGF1Padding)`.
 - **iOS / macOS:** `IOSOptions(accessibility: KeychainAccessibility.first_unlock_this_device, synchronizable: false)`.
 - **Windows:** `WindowsOptions(useBackwardCompatibility: false)`.
 - **Linux:** `LinuxOptions()`.
4. Implement `SecureStorageService` class managing initialization, connection health checks, and retry mechanics upon transient I/O exceptions.
5. Implement `BiometricGatedStorageService` combining `local_auth` authentication prompts with secure storage access, returning secrets only upon successful hardware biometric validation.
6. Create `SecureMemoryBuffer` helper class utilizing `Uint8List` with an explicit `wipe()` method that fills the allocated buffer with zero bytes or crypto-random bytes before disposal.
7. Implement `DatabaseMasterKeyManager`: on first app launch, generate a cryptographically secure 256-bit random seed (`Random.secure()`), persist it to secure storage, and retrieve it on subsequent launches to unlock Drift SQLCipher.
8. Add corruption detection and automated recovery: if KeyStore/Keychain becomes corrupted (e.g., following OS upgrade), catch `PlatformException`, log an obfuscated security event, clear stale entries, and trigger re-login flow without crashing.
9. Implement Riverpod provider `secureStorageProvider` as an un-disposable singleton provider accessible across feature repositories.
10. Implement atomic key-value operations with synchronization locks to prevent race conditions during rapid concurrent read/writes.
11. Write unit tests mocking `FlutterSecureStorage` verifying key serialization and null handling.
12. Write integration tests on physical Android (KeyStore) and iOS (Keychain) devices verifying that stored values survive app restarts and are cleared upon app uninstallation.
13. Write desktop integration tests verifying Windows DPAPI and Linux Libsecret persistence.
14. Perform memory leak and heap dump analysis to confirm zero plaintext retention of tokens in long-lived Dart objects.

## Interfaces / Contracts

```dart
// lib/core/security/storage/domain/secure_storage_service.dart
abstract class ISecureStorageService {
  Future<void> writeString({required String key, required String value});
  Future<String?> readString({required String key});
  
  Future<void> writeBytes({required String key, required List<int> bytes});
  Future<List<int>?> readBytes({required String key});

  Future<bool> containsKey({required String key});
  Future<void> delete({required String key});
  Future<void> clearAll();
}

// Biometric Gated Storage Interface
abstract class IBiometricGatedStorageService {
  Future<String?> readSecretWithBiometric({
    required String key,
    required String promptMessage,
  });

  Future<void> writeSecretWithBiometric({
    required String key,
    required String secret,
    required String promptMessage,
  });
}

// Memory Buffer Zeroization Helper
class SecureMemoryBuffer {
  final Uint8List _bytes;
  bool _isDisposed = false;

  SecureMemoryBuffer(this._bytes);

  Uint8List get bytes {
    if (_isDisposed) throw StateError('Memory buffer has already been wiped and disposed.');
    return _bytes;
  }

  void wipe() {
    if (!_isDisposed) {
      _bytes.fillRange(0, _bytes.length, 0);
      _isDisposed = true;
    }
  }
}
```

```dart
// lib/core/security/storage/data/secure_storage_impl.dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../domain/secure_storage_service.dart';

class SecureStorageServiceImpl implements ISecureStorageService {
  final FlutterSecureStorage _storage;

  SecureStorageServiceImpl()
      : _storage = const FlutterSecureStorage(
          aOptions: AndroidOptions(
            encryptedSharedPreferences: true,
            resetOnError: true,
          ),
          iOptions: IOSOptions(
            accessibility: KeychainAccessibility.first_unlock_this_device,
            synchronizable: false,
          ),
          mOptions: MacOsOptions(
            accessibility: KeychainAccessibility.first_unlock_this_device,
            synchronizable: false,
          ),
          wOptions: WindowsOptions(),
          lOptions: LinuxOptions(),
        );

  @override
  Future<void> writeString({required String key, required String value}) async {
    await _storage.write(key: key, value: value);
  }

  @override
  Future<String?> readString({required String key}) async {
    return await _storage.read(key: key);
  }

  @override
  Future<void> delete({required String key}) async {
    await _storage.delete(key: key);
  }

  @override
  Future<void> clearAll() async {
    await _storage.deleteAll();
  }

  @override
  Future<bool> containsKey({required String key}) async {
    return await _storage.containsKey(key: key);
  }
  
  @override
  Future<void> writeBytes({required String key, required List<int> bytes}) async {
    final base64Val = base64Encode(bytes);
    await _storage.write(key: key, value: base64Val);
  }

  @override
  Future<List<int>?> readBytes({required String key}) async {
    final base64Val = await _storage.read(key: key);
    if (base64Val == null) return null;
    return base64Decode(base64Val);
  }
}
```

## Security & Compliance Notes
- **Hardware Isolation:** On iOS, items are flagged `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly` to prevent inclusion in unencrypted iTunes/iCloud backups. On Android, keys are anchored to StrongBox or TEE (Trusted Execution Environment).
- **Anti-Forensics & Wiping:** During user logout or security trigger (device jailbreak/root detected), `clearAll()` must be invoked immediately, followed by memory buffer zeroization.
- **DPAPI Protection on Windows:** Secrets are encrypted via `CryptProtectData` with user credential binding; other Windows user accounts on the same machine cannot decrypt the data.
- **Zero Logging:** Under no circumstances should secrets stored or retrieved from secure storage be passed to console print statements, analytics breadcrumbs, or error telemetry.

## Acceptance Criteria
- [ ] Read and write operations succeed reliably across Android (KeyStore), iOS (Keychain), Windows (DPAPI), Linux (Libsecret), and macOS (Keychain).
- [ ] Database encryption master key generates once, persists securely, and unlocks the Drift database on subsequent restarts.
- [ ] Biometric-gated operations correctly present FaceID / Fingerprint prompts and reject retrieval if biometric check fails or is canceled.
- [ ] Key corruption or OS upgrade exceptions are handled gracefully without application crashes.
- [ ] Unit and integration tests verify full CRUD lifecycle of secure storage tokens.
- [ ] Memory analysis confirms wiped buffers contain only zero bytes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 105 (Auth Architecture), Prompt 109 (Secrets Management).
- **Parallel Tasks:** Prompt 514 (Settings & Profile UI), Prompt 515 (Offline Queued Orders), Prompt 525 (API Client Layer).
