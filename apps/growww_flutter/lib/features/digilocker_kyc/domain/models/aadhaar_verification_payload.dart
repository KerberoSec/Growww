/// Payload representing Aadhaar e-KYC transaction and verification state.
class AadhaarVerificationPayload {
  final String rawAadhaarNumber;
  final String transactionId;
  final String? otpCode;
  final bool isVerified;
  final String? maskedUid;
  final String? fullName;
  final String? dateOfBirth;
  final String? gender;
  final String? addressLine;
  final String? pincode;
  final DateTime requestedAt;
  final DateTime? verifiedAt;

  const AadhaarVerificationPayload({
    required this.rawAadhaarNumber,
    required this.transactionId,
    this.otpCode,
    this.isVerified = false,
    this.maskedUid,
    this.fullName,
    this.dateOfBirth,
    this.gender,
    this.addressLine,
    this.pincode,
    required this.requestedAt,
    this.verifiedAt,
  });

  /// Returns masked format: "XXXX XXXX 1234"
  String get formattedMaskedNumber {
    if (maskedUid != null) return maskedUid!;
    final clean = rawAadhaarNumber.replaceAll(' ', '');
    if (clean.length == 12) {
      return 'XXXX XXXX ${clean.substring(8, 12)}';
    }
    return 'XXXX XXXX XXXX';
  }

  AadhaarVerificationPayload copyWith({
    String? rawAadhaarNumber,
    String? transactionId,
    String? otpCode,
    bool? isVerified,
    String? maskedUid,
    String? fullName,
    String? dateOfBirth,
    String? gender,
    String? addressLine,
    String? pincode,
    DateTime? requestedAt,
    DateTime? verifiedAt,
  }) {
    return AadhaarVerificationPayload(
      rawAadhaarNumber: rawAadhaarNumber ?? this.rawAadhaarNumber,
      transactionId: transactionId ?? this.transactionId,
      otpCode: otpCode ?? this.otpCode,
      isVerified: isVerified ?? this.isVerified,
      maskedUid: maskedUid ?? this.maskedUid,
      fullName: fullName ?? this.fullName,
      dateOfBirth: dateOfBirth ?? this.dateOfBirth,
      gender: gender ?? this.gender,
      addressLine: addressLine ?? this.addressLine,
      pincode: pincode ?? this.pincode,
      requestedAt: requestedAt ?? this.requestedAt,
      verifiedAt: verifiedAt ?? this.verifiedAt,
    );
  }
}
