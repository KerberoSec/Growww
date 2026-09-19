import 'dart:async';
import 'dart:convert';
import 'package:crypto/crypto.dart';

/// Available biometric modalities on mobile and desktop platforms.
enum BiometricType {
  fingerprint,
  faceId,
  iris,
  passkeyHardware,
  none;

  String get displayName {
    switch (this) {
      case BiometricType.fingerprint:
        return 'Fingerprint / Touch ID';
      case BiometricType.faceId:
        return 'Face ID / Windows Hello';
      case BiometricType.iris:
        return 'Iris Scanner';
      case BiometricType.passkeyHardware:
        return 'FIDO2 / YubiKey Hardware Token';
      case BiometricType.none:
        return 'None';
    }
  }
}

/// Operational status of the device biometric subsystem.
enum BiometricStatus {
  available,
  notEnrolled,
  lockedOut,
  hardwareUnavailable,
  unsupported;

  bool get isReady => this == BiometricStatus.available;
}

/// Result returned from an authentication attempt.
class BiometricAuthResult {
  final bool success;
  final String? signature;
  final String? errorMessage;
  final BiometricType typeUsed;
  final DateTime timestamp;

  const BiometricAuthResult({
    required this.success,
    this.signature,
    this.errorMessage,
    required this.typeUsed,
    required this.timestamp,
  });

  factory BiometricAuthResult.success({
    required String signature,
    required BiometricType type,
  }) {
    return BiometricAuthResult(
      success: true,
      signature: signature,
      typeUsed: type,
      timestamp: DateTime.now(),
    );
  }

  factory BiometricAuthResult.failure({
    required String errorMessage,
    BiometricType type = BiometricType.none,
  }) {
    return BiometricAuthResult(
      success: false,
      errorMessage: errorMessage,
      typeUsed: type,
      timestamp: DateTime.now(),
    );
  }

  @override
  String toString() =>
      'BiometricAuthResult(success: $success, type: ${typeUsed.displayName}, sig: $signature, err: $errorMessage)';
}

/// Abstract platform authenticator allowing native platform injection and unit test mocks.
abstract class IPlatformBiometricGateway {
  Future<BiometricStatus> getStatus();
  Future<List<BiometricType>> getAvailableBiometrics();
  Future<bool> promptBiometric({
    required String localizedReason,
    required String title,
    String cancelButtonText = 'Cancel',
  });
}

/// Default mockable platform gateway for Linux/Testing/Headless environments.
class MockPlatformBiometricGateway implements IPlatformBiometricGateway {
  BiometricStatus mockStatus;
  List<BiometricType> mockBiometrics;
  bool shouldSucceed;

  MockPlatformBiometricGateway({
    this.mockStatus = BiometricStatus.available,
    this.mockBiometrics = const [BiometricType.fingerprint, BiometricType.faceId],
    this.shouldSucceed = true,
  });

  @override
  Future<BiometricStatus> getStatus() async => mockStatus;

  @override
  Future<List<BiometricType>> getAvailableBiometrics() async => mockBiometrics;

  @override
  Future<bool> promptBiometric({
    required String localizedReason,
    required String title,
    String cancelButtonText = 'Cancel',
  }) async {
    return shouldSucceed;
  }
}

/// Biometric Authentication Service engineered for:
/// - Enforcing FaceID / Fingerprint authorization on high-value orders
/// - SEBI compliance on high-notional transactions (> ₹ 1,00,000 INR or > 1,000 USDT)
/// - Cryptographic signing of order verification nonces
/// - App-level biometric session locking
class BiometricAuthService {
  final IPlatformBiometricGateway _gateway;
  final String _secureEnclaveKey; // Simulated master enclave secret

  /// Configurable high-value order thresholds requiring mandatory biometric sign-off
  double inrHighValueThreshold;
  double usdtHighValueThreshold;

  BiometricAuthService({
    IPlatformBiometricGateway? gateway,
    String? secureEnclaveKey,
    this.inrHighValueThreshold = 100000.0, // ₹ 1,00,000 INR
    this.usdtHighValueThreshold = 1000.0,  // 1,000 USDT
  })  : _gateway = gateway ?? MockPlatformBiometricGateway(),
        _secureEnclaveKey = secureEnclaveKey ?? 'growww_secure_enclave_device_master_key_v1';

  /// Queries whether device hardware supports biometric authentication and user is enrolled.
  Future<bool> isBiometricAvailable() async {
    final status = await _gateway.getStatus();
    return status == BiometricStatus.available;
  }

  /// Lists available biometric modalities (FaceID, Fingerprint, etc.).
  Future<List<BiometricType>> getAvailableBiometrics() async {
    return await _gateway.getAvailableBiometrics();
  }

  /// Evaluates whether an order qualifies as high-value requiring biometric signing.
  bool requiresBiometricAuth({
    required double notionalValue,
    required String currency,
  }) {
    final cur = currency.toUpperCase();
    if (cur == 'INR' && notionalValue >= inrHighValueThreshold) {
      return true;
    }
    if ((cur == 'USDT' || cur == 'USD' || cur == 'USDC') &&
        notionalValue >= usdtHighValueThreshold) {
      return true;
    }
    return false;
  }

  /// Prompts user for biometric verification (Face ID or Fingerprint) and generates
  /// a tamper-proof cryptographic challenge proof token for order submission.
  Future<BiometricAuthResult> authenticateForOrder({
    required String clientOrderId,
    required String symbol,
    required double notionalValue,
    required String currency,
  }) async {
    final isAvailable = await isBiometricAvailable();
    if (!isAvailable) {
      return BiometricAuthResult.failure(
        errorMessage: 'Biometric hardware unavailable or credentials not enrolled.',
      );
    }

    final biometrics = await getAvailableBiometrics();
    final primaryType = biometrics.isNotEmpty ? biometrics.first : BiometricType.fingerprint;

    final reason = 'Authorize high-value order of $notionalValue $currency for $symbol';
    final title = 'Confirm Sovereign Execution';

    final authenticated = await _gateway.promptBiometric(
      localizedReason: reason,
      title: title,
    );

    if (!authenticated) {
      return BiometricAuthResult.failure(
        errorMessage: 'Biometric authentication cancelled or failed.',
        type: primaryType,
      );
    }

    // Generate cryptographic HMAC-SHA256 signature representing biometric approval
    final signature = generateOrderBiometricSignature(
      clientOrderId: clientOrderId,
      symbol: symbol,
      notionalValue: notionalValue,
      currency: currency,
      timestampMs: DateTime.now().millisecondsSinceEpoch,
    );

    return BiometricAuthResult.success(
      signature: signature,
      type: primaryType,
    );
  }

  /// Intercepts order placement: requires and validates biometric auth if order is high-value.
  Future<BiometricAuthResult?> verifyOrderIfHighValue({
    required String clientOrderId,
    required String symbol,
    required double notionalValue,
    required String currency,
  }) async {
    if (!requiresBiometricAuth(notionalValue: notionalValue, currency: currency)) {
      // Normal value order; no biometric step required
      return null;
    }

    return await authenticateForOrder(
      clientOrderId: clientOrderId,
      symbol: symbol,
      notionalValue: notionalValue,
      currency: currency,
    );
  }

  /// Deterministic cryptographic challenge signer over order metadata.
  String generateOrderBiometricSignature({
    required String clientOrderId,
    required String symbol,
    required double notionalValue,
    required String currency,
    required int timestampMs,
  }) {
    final message = '$clientOrderId|$symbol|$notionalValue|$currency|$timestampMs';
    final keyBytes = utf8.encode(_secureEnclaveKey);
    final messageBytes = utf8.encode(message);

    final hmacSha256 = Hmac(sha256, keyBytes);
    final digest = hmacSha256.convert(messageBytes);

    return 'bio_${digest.toString().substring(0, 32)}';
  }

  /// Verifies validity of an order biometric proof token.
  bool verifySignature({
    required String signature,
    required String clientOrderId,
    required String symbol,
    required double notionalValue,
    required String currency,
    required int timestampMs,
  }) {
    final expected = generateOrderBiometricSignature(
      clientOrderId: clientOrderId,
      symbol: symbol,
      notionalValue: notionalValue,
      currency: currency,
      timestampMs: timestampMs,
    );
    return signature == expected;
  }
}
