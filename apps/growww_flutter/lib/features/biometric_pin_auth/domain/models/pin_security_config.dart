/// Security configuration for 6-digit PIN authentication.
class PinSecurityConfig {
  final int pinLength;
  final int maxFailedAttempts;
  final int lockoutDurationSeconds;
  final String salt;
  final String hashedPin; // SHA-256(pin + salt)

  const PinSecurityConfig({
    this.pinLength = 6,
    this.maxFailedAttempts = 5,
    this.lockoutDurationSeconds = 30,
    required this.salt,
    required this.hashedPin,
  });

  PinSecurityConfig copyWith({
    int? pinLength,
    int? maxFailedAttempts,
    int? lockoutDurationSeconds,
    String? salt,
    String? hashedPin,
  }) {
    return PinSecurityConfig(
      pinLength: pinLength ?? this.pinLength,
      maxFailedAttempts: maxFailedAttempts ?? this.maxFailedAttempts,
      lockoutDurationSeconds: lockoutDurationSeconds ?? this.lockoutDurationSeconds,
      salt: salt ?? this.salt,
      hashedPin: hashedPin ?? this.hashedPin,
    );
  }
}
