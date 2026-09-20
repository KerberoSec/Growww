import 'package:test/test.dart';
import 'package:growww_flutter/features/biometric_pin_auth/domain/models/pin_auth_enums.dart';
import 'package:growww_flutter/features/biometric_pin_auth/domain/models/pin_security_config.dart';
import 'package:growww_flutter/features/biometric_pin_auth/domain/services/pin_crypt_engine.dart';
import 'package:growww_flutter/features/biometric_pin_auth/presentation/controllers/biometric_pin_auth_controller.dart';

void main() {
  group('Prompt 505 - Biometric Authentication & 6-Digit PIN Screen', () {
    const salt = 'gw_salt_098273';
    final hashedPin = PinCryptEngine.hashPin('847291', salt);

    late PinSecurityConfig config;
    late BiometricPinAuthController controller;

    setUp(() {
      config = PinSecurityConfig(
        salt: salt,
        hashedPin: hashedPin,
        maxFailedAttempts: 3,
        lockoutDurationSeconds: 15,
      );
      controller = BiometricPinAuthController(securityConfig: config);
    });

    tearDown(() {
      controller.dispose();
    });

    test('PinCryptEngine identifies weak pins and validates salted SHA-256 hashes', () {
      // Trivial repeated digits
      expect(PinCryptEngine.isWeakPin('111111'), isTrue);
      expect(PinCryptEngine.isWeakPin('000000'), isTrue);

      // Sequential digits
      expect(PinCryptEngine.isWeakPin('123456'), isTrue);
      expect(PinCryptEngine.isWeakPin('654321'), isTrue);
      expect(PinCryptEngine.isWeakPin('234567'), isTrue);

      // Wrong lengths
      expect(PinCryptEngine.isWeakPin('12345'), isTrue);
      expect(PinCryptEngine.isWeakPin('1234567'), isTrue);
      expect(PinCryptEngine.isWeakPin('abcdef'), isTrue);

      // Strong / valid non-sequential PINs
      expect(PinCryptEngine.isWeakPin('847291'), isFalse);
      expect(PinCryptEngine.isWeakPin('481903'), isFalse);

      // Salted hash consistency
      final h1 = PinCryptEngine.hashPin('847291', salt);
      final h2 = PinCryptEngine.hashPin('847291', salt);
      expect(h1, equals(h2));
      expect(h1.length, equals(64));
    });

    test('PinCryptEngine throttles failed attempts and enforces lockout duration', () {
      // 1st failed attempt
      final res1 = PinCryptEngine.evaluatePinAttempt(
        enteredPin: '111222',
        config: config,
        currentFailedAttempts: 0,
        lockoutExpiry: null,
        purpose: AuthPurpose.appUnlock,
      );
      expect(res1.isSuccess, isFalse);
      expect(res1.remainingAttempts, equals(2));
      expect(res1.isLockedOut, isFalse);

      // 2nd failed attempt
      final res2 = PinCryptEngine.evaluatePinAttempt(
        enteredPin: '333444',
        config: config,
        currentFailedAttempts: 1,
        lockoutExpiry: null,
        purpose: AuthPurpose.appUnlock,
      );
      expect(res2.isSuccess, isFalse);
      expect(res2.remainingAttempts, equals(1));
      expect(res2.isLockedOut, isFalse);

      // 3rd failed attempt => Lockout
      final res3 = PinCryptEngine.evaluatePinAttempt(
        enteredPin: '555666',
        config: config,
        currentFailedAttempts: 2,
        lockoutExpiry: null,
        purpose: AuthPurpose.appUnlock,
      );
      expect(res3.isSuccess, isFalse);
      expect(res3.remainingAttempts, equals(0));
      expect(res3.isLockedOut, isTrue);
      expect(res3.lockoutSecondsRemaining, equals(15));
    });

    test('BiometricPinAuthController manages keypad input, backspace, and clear', () {
      expect(controller.state.filledDots, equals(0));

      controller.pressDigit(8);
      controller.pressDigit(4);
      controller.pressDigit(7);
      expect(controller.state.filledDots, equals(3));
      expect(controller.state.enteredDigits, equals([8, 4, 7]));

      controller.pressBackspace();
      expect(controller.state.filledDots, equals(2));
      expect(controller.state.enteredDigits, equals([8, 4]));

      controller.clearPin();
      expect(controller.state.filledDots, equals(0));
      expect(controller.state.enteredDigits.isEmpty, isTrue);
    });

    test('BiometricPinAuthController completes verification on 6th correct digit', () {
      // Default correct PIN is 847291
      controller.pressDigit(8);
      controller.pressDigit(4);
      controller.pressDigit(7);
      controller.pressDigit(2);
      controller.pressDigit(9);
      expect(controller.state.isAuthenticated, isFalse);

      controller.pressDigit(1); // 6th digit auto-verifies
      expect(controller.state.isAuthenticated, isTrue);
      expect(controller.state.entryState, equals(PinEntryState.authenticated));
      expect(controller.state.sessionToken, isNotNull);
      expect(controller.state.sessionToken, startsWith('gw_auth_'));
    });

    test('BiometricPinAuthController triggers biometric authorization modal', () async {
      final bioSuccess = await controller.triggerBiometricAuth(simulateSuccess: true);
      expect(bioSuccess, isTrue);
      expect(controller.state.isAuthenticated, isTrue);
      expect(controller.state.sessionToken, isNotNull);
      expect(controller.state.sessionToken, contains('gw_bio_'));
    });

    test('BiometricPinAuthController handles brute-force lockout on repeated bad PINs', () {
      // 3 failed attempts (since maxFailedAttempts=3)
      // Attempt 1:
      for (final d in [1, 2, 3, 4, 5, 0]) {
        controller.pressDigit(d);
      }
      expect(controller.state.entryState, equals(PinEntryState.error));
      expect(controller.state.failedAttempts, equals(1));

      // Attempt 2:
      for (final d in [1, 2, 3, 4, 5, 0]) {
        controller.pressDigit(d);
      }
      expect(controller.state.entryState, equals(PinEntryState.error));
      expect(controller.state.failedAttempts, equals(2));

      // Attempt 3 => Triggers lockout:
      for (final d in [1, 2, 3, 4, 5, 0]) {
        controller.pressDigit(d);
      }
      expect(controller.state.entryState, equals(PinEntryState.lockedOut));
      expect(controller.state.isLocked, isTrue);
      expect(controller.state.lockoutRemainingSeconds, equals(15));
    });
  });
}
