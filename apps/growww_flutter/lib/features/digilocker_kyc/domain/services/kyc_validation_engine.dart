import 'dart:convert';
import 'package:crypto/crypto.dart';

/// KYC Validation Engine implementing UIDAI Aadhaar Verhoeff algorithm,
/// NSDL PAN format verification, and cryptographic consent integrity.
class KycValidationEngine {
  // Verhoeff algorithm multiplication table
  static const List<List<int>> _verhoeffD = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    [1, 2, 3, 4, 0, 6, 7, 8, 9, 5],
    [2, 3, 4, 0, 1, 7, 8, 9, 5, 6],
    [3, 4, 0, 1, 2, 8, 9, 5, 6, 7],
    [4, 0, 1, 2, 3, 9, 5, 6, 7, 8],
    [5, 9, 8, 7, 6, 0, 4, 3, 2, 1],
    [6, 5, 9, 8, 7, 1, 0, 4, 3, 2],
    [7, 6, 5, 9, 8, 2, 1, 0, 4, 3],
    [8, 7, 6, 5, 9, 3, 2, 1, 0, 4],
    [9, 8, 7, 6, 5, 4, 3, 2, 1, 0],
  ];

  // Verhoeff algorithm permutation table
  static const List<List<int>> _verhoeffP = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    [1, 5, 7, 6, 2, 8, 3, 0, 9, 4],
    [5, 8, 0, 3, 7, 9, 6, 1, 4, 2],
    [8, 9, 1, 6, 0, 4, 3, 5, 2, 7],
    [9, 4, 5, 3, 1, 2, 6, 8, 7, 0],
    [4, 2, 8, 6, 5, 7, 3, 9, 0, 1],
    [2, 7, 9, 3, 8, 0, 6, 4, 1, 5],
    [7, 0, 4, 6, 9, 1, 3, 2, 5, 8],
  ];

  // Verhoeff inverse table
  static const List<int> _inv = [0, 4, 3, 2, 1, 5, 6, 7, 8, 9];

  /// Computes the Verhoeff check digit for an 11-digit or n-digit numerical prefix.
  static int generateVerhoeffCheckDigit(String num) {
    int c = 0;
    final reversed = num.replaceAll(' ', '').split('').reversed.toList();
    for (int i = 0; i < reversed.length; i++) {
      final digit = int.parse(reversed[i]);
      c = _verhoeffD[c][_verhoeffP[(i + 1) % 8][digit]];
    }
    return _inv[c];
  }

  /// Validates Aadhaar format and Verhoeff checksum.
  /// Aadhaar must:
  /// 1. Be 12 digits long
  /// 2. Not start with '0' or '1'
  /// 3. Pass Verhoeff checksum algorithm
  static bool validateAadhaar(String input) {
    final clean = input.replaceAll(' ', '').trim();
    if (clean.length != 12) return false;
    if (!RegExp(r'^[2-9][0-9]{11}$').hasMatch(clean)) return false;

    return validateVerhoeff(clean);
  }

  /// Verhoeff algorithm checksum validator.
  static bool validateVerhoeff(String num) {
    final clean = num.replaceAll(' ', '').trim();
    if (clean.isEmpty) return false;
    int c = 0;
    final reversed = clean.split('').reversed.toList();
    for (int i = 0; i < reversed.length; i++) {
      final digit = int.tryParse(reversed[i]);
      if (digit == null) return false;
      c = _verhoeffD[c][_verhoeffP[i % 8][digit]];
    }
    return c == 0;
  }

  /// Strict PAN format validation adhering to Indian Income Tax Department:
  /// - Exactly 10 characters
  /// - 5 uppercase letters (first 3: alphabetic series, 4th: entity type, 5th: surname initial)
  /// - 4 numeric digits
  /// - 1 uppercase alphabetic check character
  static bool validatePan(String input) {
    final clean = input.trim().toUpperCase();
    if (clean.length != 10) return false;
    final panRegex = RegExp(r'^[A-Z]{5}[0-9]{4}[A-Z]$');
    return panRegex.hasMatch(clean);
  }

  /// Validates 6-digit numeric OTP.
  static bool validateOtp(String input) {
    final clean = input.trim();
    return RegExp(r'^[0-9]{6}$').hasMatch(clean);
  }

  /// Computes SHA-256 hash of consent terms for DigiLocker audit trail.
  static String computeConsentHash({
    required String userDid,
    required String timestampIso,
    required List<String> scopes,
  }) {
    final raw = '$userDid|$timestampIso|${scopes.join(',')}';
    final bytes = utf8.encode(raw);
    return sha256.convert(bytes).toString();
  }

  /// Masks Aadhaar string into standard format "XXXX XXXX 1234"
  static String maskAadhaar(String raw) {
    final clean = raw.replaceAll(' ', '').trim();
    if (clean.length == 12) {
      return 'XXXX XXXX ${clean.substring(8, 12)}';
    }
    return clean;
  }

  /// Validates age compliance (must be >= 18 for Sovereign KYC)
  static bool isAgeEligible(DateTime dateOfBirth, {int minimumAge = 18}) {
    final now = DateTime.now();
    int age = now.year - dateOfBirth.year;
    if (now.month < dateOfBirth.month ||
        (now.month == dateOfBirth.month && now.day < dateOfBirth.day)) {
      age--;
    }
    return age >= minimumAge;
  }
}
