/// PAN Verification Model matching NSDL Income Tax Department standard.
class PanVerificationPayload {
  final String rawPan;
  final String? registeredName;
  final String? dateOfBirth;
  final String? panStatus; // e.g., 'ACTIVE_OPERATIVE'
  final String? entityType; // e.g., 'INDIVIDUAL', 'COMPANY'
  final bool isVerified;
  final bool aadhaarPanLinked;
  final DateTime? verifiedAt;

  const PanVerificationPayload({
    required this.rawPan,
    this.registeredName,
    this.dateOfBirth,
    this.panStatus,
    this.entityType,
    this.isVerified = false,
    this.aadhaarPanLinked = true,
    this.verifiedAt,
  });

  /// Formatted masked PAN e.g. "ABCDE****F"
  String get maskedPan {
    final clean = rawPan.trim().toUpperCase();
    if (clean.length == 10) {
      return '${clean.substring(0, 5)}****${clean.substring(9, 10)}';
    }
    return '**********';
  }

  /// 4th character denotes entity status: P=Individual, C=Company, H=HUF, F=Firm, T=Trust, etc.
  String get entityClassification {
    final clean = rawPan.trim().toUpperCase();
    if (clean.length >= 4) {
      final code = clean[3];
      switch (code) {
        case 'P':
          return 'Individual / Person';
        case 'C':
          return 'Company / Corporate';
        case 'H':
          return 'Hindu Undivided Family (HUF)';
        case 'F':
          return 'Partnership Firm / LLP';
        case 'T':
          return 'Trust / Foundation';
        case 'A':
          return 'Association of Persons (AOP)';
        default:
          return 'Other Entity ($code)';
      }
    }
    return 'Unknown';
  }

  PanVerificationPayload copyWith({
    String? rawPan,
    String? registeredName,
    String? dateOfBirth,
    String? panStatus,
    String? entityType,
    bool? isVerified,
    bool? aadhaarPanLinked,
    DateTime? verifiedAt,
  }) {
    return PanVerificationPayload(
      rawPan: rawPan ?? this.rawPan,
      registeredName: registeredName ?? this.registeredName,
      dateOfBirth: dateOfBirth ?? this.dateOfBirth,
      panStatus: panStatus ?? this.panStatus,
      entityType: entityType ?? this.entityType,
      isVerified: isVerified ?? this.isVerified,
      aadhaarPanLinked: aadhaarPanLinked ?? this.aadhaarPanLinked,
      verifiedAt: verifiedAt ?? this.verifiedAt,
    );
  }
}
