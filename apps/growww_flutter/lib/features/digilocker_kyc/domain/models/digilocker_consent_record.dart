/// Cryptographic consent record for DigiLocker gateway data sharing.
class DigiLockerConsentRecord {
  final String consentId;
  final String userDid;
  final List<String> requestedDocTypes; // e.g. ['AADHAAR_EKYC', 'PAN_VERIFICATION']
  final String consentArtifactHash; // SHA-256 hash of consent terms
  final DateTime grantedAt;
  final DateTime expiresAt;
  final bool isRevoked;

  const DigiLockerConsentRecord({
    required this.consentId,
    required this.userDid,
    required this.requestedDocTypes,
    required this.consentArtifactHash,
    required this.grantedAt,
    required this.expiresAt,
    this.isRevoked = false,
  });

  bool get isValid => !isRevoked && DateTime.now().isBefore(expiresAt);

  DigiLockerConsentRecord copyWith({
    String? consentId,
    String? userDid,
    List<String>? requestedDocTypes,
    String? consentArtifactHash,
    DateTime? grantedAt,
    DateTime? expiresAt,
    bool? isRevoked,
  }) {
    return DigiLockerConsentRecord(
      consentId: consentId ?? this.consentId,
      userDid: userDid ?? this.userDid,
      requestedDocTypes: requestedDocTypes ?? this.requestedDocTypes,
      consentArtifactHash: consentArtifactHash ?? this.consentArtifactHash,
      grantedAt: grantedAt ?? this.grantedAt,
      expiresAt: expiresAt ?? this.expiresAt,
      isRevoked: isRevoked ?? this.isRevoked,
    );
  }
}
