/// DigiLocker and Tier-2 Sovereign KYC Enums.
enum DigiLockerKycStep {
  aadhaarEntry,
  aadhaarOtpVerification,
  digilockerConsent,
  panVerification,
  completed;

  String get displayName {
    switch (this) {
      case DigiLockerKycStep.aadhaarEntry:
        return 'Aadhaar Identification';
      case DigiLockerKycStep.aadhaarOtpVerification:
        return 'UIDAI OTP Authentication';
      case DigiLockerKycStep.digilockerConsent:
        return 'DigiLocker Consent Fast-Track';
      case DigiLockerKycStep.panVerification:
        return 'NSDL PAN Registry Validation';
      case DigiLockerKycStep.completed:
        return 'Sovereign KYC Attested';
    }
  }

  int get stepIndex {
    switch (this) {
      case DigiLockerKycStep.aadhaarEntry:
      case DigiLockerKycStep.aadhaarOtpVerification:
      case DigiLockerKycStep.digilockerConsent:
        return 1;
      case DigiLockerKycStep.panVerification:
        return 2;
      case DigiLockerKycStep.completed:
        return 3;
    }
  }
}

enum AadhaarVerificationMethod {
  directUidaiOtp,
  digilockerConsent;

  String get label => this == directUidaiOtp ? 'UIDAI OTP Verification' : 'DigiLocker Fast-Track Sync';
}

enum DigiLockerAuthStatus {
  idle,
  authorizing,
  consented,
  fetchingDocuments,
  verified,
  failed;

  bool get isSuccessful => this == DigiLockerAuthStatus.verified;
}

enum PanVerificationStatus {
  unverified,
  scanningOcr,
  validatingNsdl,
  verified,
  invalid;

  bool get isApproved => this == PanVerificationStatus.verified;
}

enum KycTier {
  tier1Basic, // Email + Phone verified, crypto deposit only
  tier2Sovereign; // UIDAI Aadhaar + NSDL PAN + Besu DID Attestation

  String get displayName => this == tier1Basic ? 'Tier 1 - Crypto Only' : 'Tier 2 - Sovereign Institutional';
}
