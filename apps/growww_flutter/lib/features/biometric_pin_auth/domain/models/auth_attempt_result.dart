import 'pin_auth_enums.dart';

/// Result of a PIN or biometric authentication evaluation.
class AuthAttemptResult {
  final bool isSuccess;
  final int remainingAttempts;
  final bool isLockedOut;
  final int lockoutSecondsRemaining;
  final String? sessionToken;
  final String? errorMessage;
  final BiometricPromptType? biometricTypeUsed;

  const AuthAttemptResult({
    required this.isSuccess,
    this.remainingAttempts = 5,
    this.isLockedOut = false,
    this.lockoutSecondsRemaining = 0,
    this.sessionToken,
    this.errorMessage,
    this.biometricTypeUsed,
  });

  factory AuthAttemptResult.success({
    required String sessionToken,
    BiometricPromptType? biometricType,
  }) {
    return AuthAttemptResult(
      isSuccess: true,
      sessionToken: sessionToken,
      biometricTypeUsed: biometricType,
      remainingAttempts: 5,
    );
  }

  factory AuthAttemptResult.failure({
    required int remainingAttempts,
    required String errorMessage,
    bool isLockedOut = false,
    int lockoutSecondsRemaining = 0,
  }) {
    return AuthAttemptResult(
      isSuccess: false,
      remainingAttempts: remainingAttempts,
      isLockedOut: isLockedOut,
      lockoutSecondsRemaining: lockoutSecondsRemaining,
      errorMessage: errorMessage,
    );
  }
}
