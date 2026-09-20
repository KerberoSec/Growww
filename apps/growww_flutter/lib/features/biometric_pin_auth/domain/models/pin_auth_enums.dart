/// Biometric and PIN authentication enums.
enum BiometricPromptType {
  fingerprint,
  faceId,
  iris,
  none;

  String get displayName {
    switch (this) {
      case BiometricPromptType.fingerprint:
        return 'Fingerprint / Touch ID';
      case BiometricPromptType.faceId:
        return 'Face ID / Windows Hello';
      case BiometricPromptType.iris:
        return 'Iris Scanner';
      case BiometricPromptType.none:
        return 'None';
    }
  }

  String get iconName {
    switch (this) {
      case BiometricPromptType.fingerprint:
        return 'fingerprint';
      case BiometricPromptType.faceId:
        return 'face';
      case BiometricPromptType.iris:
        return 'remove_red_eye';
      case BiometricPromptType.none:
        return 'lock';
    }
  }
}

enum PinEntryState {
  entering,
  verifying,
  authenticated,
  lockedOut,
  error;

  bool get isLocked => this == PinEntryState.lockedOut;
  bool get isSuccess => this == PinEntryState.authenticated;
}

enum AuthPurpose {
  appUnlock,
  orderExecution,
  vaultWithdrawal,
  apiSecretAccess;

  String get title {
    switch (this) {
      case AuthPurpose.appUnlock:
        return 'Unlock Sovereign Workstation';
      case AuthPurpose.orderExecution:
        return 'Authorize Institutional Order';
      case AuthPurpose.vaultWithdrawal:
        return 'Authorize Vault Transfer';
      case AuthPurpose.apiSecretAccess:
        return 'Reveal Institutional API Keys';
    }
  }

  String get subtitle {
    switch (this) {
      case AuthPurpose.appUnlock:
        return 'Enter your 6-digit security PIN or scan biometrics.';
      case AuthPurpose.orderExecution:
        return 'High-value transaction requires biometric or PIN signature.';
      case AuthPurpose.vaultWithdrawal:
        return 'Digital asset custody withdrawal requires multi-factor approval.';
      case AuthPurpose.apiSecretAccess:
        return 'Authenticate to view encrypted API secret secrets.';
    }
  }
}
