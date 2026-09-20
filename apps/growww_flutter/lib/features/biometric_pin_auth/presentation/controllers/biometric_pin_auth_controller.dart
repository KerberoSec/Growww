import 'dart:async';
import '../../domain/models/auth_attempt_result.dart';
import '../../domain/models/pin_auth_enums.dart';
import '../../domain/models/pin_security_config.dart';
import '../../domain/services/pin_crypt_engine.dart';

class BiometricPinAuthState {
  final List<int> enteredDigits;
  final int pinLength;
  final AuthPurpose authPurpose;
  final BiometricPromptType biometricType;
  final PinEntryState entryState;
  final int failedAttempts;
  final int lockoutRemainingSeconds;
  final bool isBiometricSupported;
  final String? sessionToken;
  final String? errorMessage;
  final String? successMessage;

  const BiometricPinAuthState({
    this.enteredDigits = const [],
    this.pinLength = 6,
    this.authPurpose = AuthPurpose.appUnlock,
    this.biometricType = BiometricPromptType.faceId,
    this.entryState = PinEntryState.entering,
    this.failedAttempts = 0,
    this.lockoutRemainingSeconds = 0,
    this.isBiometricSupported = true,
    this.sessionToken,
    this.errorMessage,
    this.successMessage,
  });

  int get filledDots => enteredDigits.length;
  bool get isComplete => enteredDigits.length == pinLength;
  bool get isLocked => entryState == PinEntryState.lockedOut;
  bool get isAuthenticated => entryState == PinEntryState.authenticated;

  BiometricPinAuthState copyWith({
    List<int>? enteredDigits,
    int? pinLength,
    AuthPurpose? authPurpose,
    BiometricPromptType? biometricType,
    PinEntryState? entryState,
    int? failedAttempts,
    int? lockoutRemainingSeconds,
    bool? isBiometricSupported,
    String? sessionToken,
    String? errorMessage,
    String? successMessage,
  }) {
    return BiometricPinAuthState(
      enteredDigits: enteredDigits ?? this.enteredDigits,
      pinLength: pinLength ?? this.pinLength,
      authPurpose: authPurpose ?? this.authPurpose,
      biometricType: biometricType ?? this.biometricType,
      entryState: entryState ?? this.entryState,
      failedAttempts: failedAttempts ?? this.failedAttempts,
      lockoutRemainingSeconds: lockoutRemainingSeconds ?? this.lockoutRemainingSeconds,
      isBiometricSupported: isBiometricSupported ?? this.isBiometricSupported,
      sessionToken: sessionToken ?? this.sessionToken,
      errorMessage: errorMessage,
      successMessage: successMessage,
    );
  }
}

class BiometricPinAuthController {
  final PinSecurityConfig config;
  BiometricPinAuthState _state;
  final _stateController = StreamController<BiometricPinAuthState>.broadcast();
  DateTime? _lockoutExpiry;
  Timer? _lockoutTimer;

  BiometricPinAuthState get state => _state;
  Stream<BiometricPinAuthState> get stream => _stateController.stream;

  BiometricPinAuthController({
    PinSecurityConfig? securityConfig,
    AuthPurpose purpose = AuthPurpose.appUnlock,
    BiometricPromptType biometric = BiometricPromptType.faceId,
  })  : config = securityConfig ??
            PinSecurityConfig(
              salt: 'gw_salt_098273',
              // Pre-hashed default PIN '847291'
              hashedPin: PinCryptEngine.hashPin('847291', 'gw_salt_098273'),
            ),
        _state = BiometricPinAuthState(
          authPurpose: purpose,
          biometricType: biometric,
        ) {
    _emit(_state);
  }

  void _emit(BiometricPinAuthState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  /// Appends a digit (0-9) to the current PIN entry.
  void pressDigit(int digit) {
    if (_state.isLocked || _state.isAuthenticated) return;
    if (_state.enteredDigits.length >= _state.pinLength) return;

    final updated = List<int>.from(_state.enteredDigits)..add(digit);
    _emit(_state.copyWith(
      enteredDigits: updated,
      entryState: PinEntryState.entering,
      errorMessage: null,
    ));

    // Auto-verify when all 6 digits are typed
    if (updated.length == _state.pinLength) {
      _verifyPin(updated.join(''));
    }
  }

  /// Removes the last entered digit.
  void pressBackspace() {
    if (_state.isLocked || _state.isAuthenticated) return;
    if (_state.enteredDigits.isEmpty) return;

    final updated = List<int>.from(_state.enteredDigits)..removeLast();
    _emit(_state.copyWith(
      enteredDigits: updated,
      entryState: PinEntryState.entering,
      errorMessage: null,
    ));
  }

  /// Clears the entire PIN pad.
  void clearPin() {
    if (_state.isLocked || _state.isAuthenticated) return;
    _emit(_state.copyWith(
      enteredDigits: [],
      entryState: PinEntryState.entering,
      errorMessage: null,
    ));
  }

  /// Evaluates entered 6-digit PIN against security rules.
  void _verifyPin(String pin) {
    _emit(_state.copyWith(entryState: PinEntryState.verifying));

    final result = PinCryptEngine.evaluatePinAttempt(
      enteredPin: pin,
      config: config,
      currentFailedAttempts: _state.failedAttempts,
      lockoutExpiry: _lockoutExpiry,
      purpose: _state.authPurpose,
    );

    if (result.isSuccess) {
      _emit(_state.copyWith(
        entryState: PinEntryState.authenticated,
        sessionToken: result.sessionToken,
        failedAttempts: 0,
        successMessage: 'PIN authentication verified.',
      ));
    } else if (result.isLockedOut) {
      _lockoutExpiry = DateTime.now().add(Duration(seconds: result.lockoutSecondsRemaining));
      _emit(_state.copyWith(
        enteredDigits: [],
        entryState: PinEntryState.lockedOut,
        failedAttempts: _state.failedAttempts + 1,
        lockoutRemainingSeconds: result.lockoutSecondsRemaining,
        errorMessage: result.errorMessage,
      ));
      _startLockoutCountdown();
    } else {
      _emit(_state.copyWith(
        enteredDigits: [],
        entryState: PinEntryState.error,
        failedAttempts: _state.failedAttempts + 1,
        errorMessage: result.errorMessage,
      ));
    }
  }

  void _startLockoutCountdown() {
    _lockoutTimer?.cancel();
    _lockoutTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_state.lockoutRemainingSeconds > 1) {
        _emit(_state.copyWith(
          lockoutRemainingSeconds: _state.lockoutRemainingSeconds - 1,
        ));
      } else {
        timer.cancel();
        _lockoutExpiry = null;
        _emit(_state.copyWith(
          entryState: PinEntryState.entering,
          lockoutRemainingSeconds: 0,
          errorMessage: null,
        ));
      }
    });
  }

  /// Triggers biometric authentication dialog (FaceID or Fingerprint).
  Future<bool> triggerBiometricAuth({bool simulateSuccess = true}) async {
    if (_state.isLocked) return false;

    _emit(_state.copyWith(entryState: PinEntryState.verifying));

    if (simulateSuccess) {
      final token = 'gw_bio_${_state.biometricType.name}_${DateTime.now().millisecondsSinceEpoch}';
      _emit(_state.copyWith(
        entryState: PinEntryState.authenticated,
        sessionToken: token,
        failedAttempts: 0,
        successMessage: '${_state.biometricType.displayName} authorized.',
      ));
      return true;
    } else {
      _emit(_state.copyWith(
        entryState: PinEntryState.error,
        errorMessage: 'Biometric scan failed or cancelled. Please enter PIN.',
      ));
      return false;
    }
  }

  void reset() {
    _lockoutTimer?.cancel();
    _lockoutExpiry = null;
    _emit(BiometricPinAuthState(
      authPurpose: _state.authPurpose,
      biometricType: _state.biometricType,
    ));
  }

  void dispose() {
    _lockoutTimer?.cancel();
    _stateController.close();
  }
}
