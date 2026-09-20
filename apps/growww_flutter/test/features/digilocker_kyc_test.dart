import 'package:test/test.dart';
import 'package:growww_flutter/features/digilocker_kyc/domain/models/kyc_enums.dart';
import 'package:growww_flutter/features/digilocker_kyc/domain/models/pan_verification_payload.dart';
import 'package:growww_flutter/features/digilocker_kyc/domain/services/kyc_validation_engine.dart';
import 'package:growww_flutter/features/digilocker_kyc/presentation/controllers/digilocker_kyc_controller.dart';

void main() {
  group('Prompt 504 - DigiLocker Onboarding KYC Flow', () {
    late DigiLockerKycController controller;

    setUp(() {
      controller = DigiLockerKycController();
    });

    tearDown(() {
      controller.dispose();
    });

    test('KycValidationEngine verifies UIDAI Verhoeff algorithm and Aadhaar constraints', () {
      const prefix = '28471928374';
      final checkDigit = KycValidationEngine.generateVerhoeffCheckDigit(prefix);
      final validAadhaar = '$prefix$checkDigit';

      final isVerhoeffValid = KycValidationEngine.validateVerhoeff(validAadhaar);
      expect(isVerhoeffValid, isTrue);

      // Full validateAadhaar check
      expect(KycValidationEngine.validateAadhaar(validAadhaar), isTrue);

      // Rejects Aadhaar starting with 0 or 1
      expect(KycValidationEngine.validateAadhaar('08471928374$checkDigit'), isFalse);
      expect(KycValidationEngine.validateAadhaar('18471928374$checkDigit'), isFalse);

      // Rejects Aadhaar with corrupted checksum
      final corrupted = '$prefix${(checkDigit + 1) % 10}';
      expect(KycValidationEngine.validateAadhaar(corrupted), isFalse);

      // Rejects short/long inputs
      expect(KycValidationEngine.validateAadhaar('28471928374'), isFalse);
      expect(KycValidationEngine.validateAadhaar('2847192837419'), isFalse);
      expect(KycValidationEngine.validateAadhaar('letters12345'), isFalse);
    });

    test('KycValidationEngine validates PAN syntax and identifies entity classifications', () {
      // Individual PAN
      expect(KycValidationEngine.validatePan('ABCPS1234D'), isTrue);
      // Company PAN
      expect(KycValidationEngine.validatePan('ABCCP9876E'), isTrue);

      // Invalid PANs
      expect(KycValidationEngine.validatePan('ABC1234D'), isFalse); // too short
      expect(KycValidationEngine.validatePan('12345ABCDE'), isFalse); // numbers first
      expect(KycValidationEngine.validatePan('ABCPS12345'), isFalse); // ends in digit

      // Entity classification in PanVerificationPayload
      const individualPan = PanVerificationPayload(rawPan: 'ABCPS1234D');
      expect(individualPan.entityClassification, equals('Individual / Person'));
      expect(individualPan.maskedPan, equals('ABCPS****D'));

      const companyPan = PanVerificationPayload(rawPan: 'XYZCR9999M');
      expect(companyPan.entityClassification, equals('Company / Corporate'));

      const hufPan = PanVerificationPayload(rawPan: 'AAAHK1111A');
      expect(hufPan.entityClassification, equals('Hindu Undivided Family (HUF)'));
    });

    test('KycValidationEngine validates OTP and generates cryptographic consent hash', () {
      expect(KycValidationEngine.validateOtp('123456'), isTrue);
      expect(KycValidationEngine.validateOtp('12345'), isFalse);
      expect(KycValidationEngine.validateOtp('1234567'), isFalse);
      expect(KycValidationEngine.validateOtp('12a456'), isFalse);

      final hash1 = KycValidationEngine.computeConsentHash(
        userDid: 'did:growww:user:101',
        timestampIso: '2026-09-20T10:00:00Z',
        scopes: ['AADHAAR_EKYC', 'PAN_VERIFICATION'],
      );
      final hash2 = KycValidationEngine.computeConsentHash(
        userDid: 'did:growww:user:101',
        timestampIso: '2026-09-20T10:00:00Z',
        scopes: ['AADHAAR_EKYC', 'PAN_VERIFICATION'],
      );
      expect(hash1, equals(hash2));
      expect(hash1.length, equals(64)); // SHA-256 hex string
    });

    test('DigiLockerKycController handles full Aadhaar OTP request and PAN validation flow', () {
      expect(controller.state.step, equals(DigiLockerKycStep.aadhaarEntry));

      // Attempt invalid Aadhaar
      controller.setAadhaarInput('123456789012');
      final reqFailed = controller.requestAadhaarOtp();
      expect(reqFailed, isFalse);
      expect(controller.state.errorMessage, contains('Invalid 12-digit Aadhaar'));

      // Valid Aadhaar
      final checkDigit = KycValidationEngine.generateVerhoeffCheckDigit('28471928374');
      controller.setAadhaarInput('2847 1928 374$checkDigit');
      final reqOk = controller.requestAadhaarOtp();
      expect(reqOk, isTrue);
      expect(controller.state.isOtpSent, isTrue);
      expect(controller.state.step, equals(DigiLockerKycStep.aadhaarOtpVerification));
      expect(controller.state.otpCooldown, equals(60));

      // Attempt invalid OTP
      final otpFailed = controller.verifyAadhaarOtp('999');
      expect(otpFailed, isFalse);

      // Valid OTP
      final otpOk = controller.verifyAadhaarOtp('123456');
      expect(otpOk, isTrue);
      expect(controller.state.step, equals(DigiLockerKycStep.panVerification));
      expect(controller.state.aadhaarPayload?.isVerified, isTrue);
      expect(controller.state.aadhaarPayload?.fullName, equals('AARAV SHARMA'));

      // Validate PAN step
      controller.setPanInput('ABCPS1234D');
      final panOk = controller.verifyPan();
      expect(panOk, isTrue);
      expect(controller.state.step, equals(DigiLockerKycStep.completed));
      expect(controller.state.kycTier, equals(KycTier.tier2Sovereign));
      expect(controller.state.onChainAttestationTx, isNotNull);
    });

    test('DigiLockerKycController executes instant DigiLocker fast-track sync', () async {
      final syncOk = await controller.triggerDigiLockerSync();
      expect(syncOk, isTrue);
      expect(controller.state.digiLockerStatus, equals(DigiLockerAuthStatus.verified));
      expect(controller.state.step, equals(DigiLockerKycStep.completed));
      expect(controller.state.kycTier, equals(KycTier.tier2Sovereign));
      expect(controller.state.consentRecord, isNotNull);
      expect(controller.state.consentRecord?.isValid, isTrue);
      expect(controller.state.onChainAttestationTx, isNotNull);
    });
  });
}
