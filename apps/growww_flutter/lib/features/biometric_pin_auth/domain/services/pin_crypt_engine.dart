import 'dart:convert';
import 'package:crypto/crypto.dart';
import '../models/auth_attempt_result.dart';
import '../models/pin_auth_enums.dart';
import '../models/pin_security_config.dart';

/// Cryptographic and security evaluation engine for 6-digit PIN pad and Biometric authentication.
class PinCryptEngine {
  /// Computes salted SHA-256 hash of the PIN.
  static String hashPin(String pin, String salt) {
    final bytes = utf8.encode('$pin:$salt:growww_enclave_v1');
    return sha256.convert(bytes).toString();
  }

  /// Checks if a proposed PIN is weak or vulnerable:
  /// - All identical digits (e.g. 000000, 111111)
  /// - Monotonically increasing sequences (e.g. 123456, 234567)
  /// - Monotonically decreasing sequences (e.g. 654321, 987654)
  static bool isWeakPin(String pin) {
    if (pin.length != 6) return true;
    if (!RegExp(r'^[0-9]{6}$').hasMatch(pin)) return true;

    // Check all same digits
    final firstChar = pin[0];
    if (pin.split('').every((c) => c == firstChar)) {
      return true;
    }

    // Check ascending sequence
    bool isAscending = true;
    for (int i = 1; i < pin.length; i++) {
      final prev = int.parse(pin[i - 1]);
      final curr = int.parse(pin[i]);
      if (curr != prev + 1) {
        isAscending = false;
        break;
      }
    }
    if (isAscending) return true;

    // Check descending sequence
    bool isDescending = true;
    for (int i = 1; i < pin.length; i++) {
      final prev = int.parse(pin[i - 1]);
      final curr = int.parse(pin[i]);
      if (curr != prev - 1) {
        isDescending = false;
        break;
      }
    }
    if (isDescending) return true;

    return false;
  }

  /// Evaluates PIN attempt against security configuration and failed attempt counter.
  static AuthAttemptResult evaluatePinAttempt({
    required String enteredPin,
    required PinSecurityConfig config,
    required int currentFailedAttempts,
    required DateTime? lockoutExpiry,
    required AuthPurpose purpose,
  }) {
    final now = DateTime.now();

    // Check if currently locked out
    if (lockoutExpiry != null && now.isBefore(lockoutExpiry)) {
      final remainingSec = lockoutExpiry.difference(now).inSeconds;
      return AuthAttemptResult.failure(
        remainingAttempts: 0,
        isLockedOut: true,
        lockoutSecondsRemaining: remainingSec > 0 ? remainingSec : 1,
        errorMessage: 'Account locked due to excessive failed attempts. Retry in ${remainingSec}s.',
      );
    }

    final computedHash = hashPin(enteredPin, config.salt);
    if (computedHash == config.hashedPin) {
      final token = generateSessionToken(
        pin: enteredPin,
        salt: config.salt,
        purpose: purpose,
      );
      return AuthAttemptResult.success(sessionToken: token);
    }

    final newFailedAttempts = currentFailedAttempts + 1;
    final remaining = config.maxFailedAttempts - newFailedAttempts;

    if (remaining <= 0) {
      return AuthAttemptResult.failure(
        remainingAttempts: 0,
        isLockedOut: true,
        lockoutSecondsRemaining: config.lockoutDurationSeconds,
        errorMessage: 'Maximum attempts exceeded. Keypad locked for ${config.lockoutDurationSeconds} seconds.',
      );
    }

    return AuthAttemptResult.failure(
      remainingAttempts: remaining,
      isLockedOut: false,
      errorMessage: 'Incorrect PIN. $remaining attempts remaining.',
    );
  }

  /// Generates a tamper-proof session authentication token.
  static String generateSessionToken({
    required String pin,
    required String salt,
    required AuthPurpose purpose,
  }) {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final raw = '$pin|$salt|${purpose.name}|$timestamp';
    final tokenBytes = utf8.encode(raw);
    final digest = sha256.convert(tokenBytes).toString();
    return 'gw_auth_${digest.substring(0, 32)}';
  }
}
