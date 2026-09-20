import 'package:test/test.dart';
import 'package:growww_flutter/features/haptic_trade_confirmation/domain/models/trade_confirmation_models.dart';
import 'package:growww_flutter/features/haptic_trade_confirmation/domain/services/haptic_trade_confirmation_service.dart';
import 'package:growww_flutter/features/haptic_trade_confirmation/presentation/controllers/haptic_trade_confirmation_controller.dart';

void main() {
  group('Prompt 054 - Biometric Quick-Auth & Haptic Feedback Trade Confirmation', () {
    late MockHapticDriver mockHaptics;
    late MockBiometricAuthenticator mockBiometric;
    late HapticTradeConfirmationService service;
    late TradeConfirmationIntent testIntent;
    late HapticTradeConfirmationController controller;

    setUp(() {
      mockHaptics = MockHapticDriver();
      mockBiometric = MockBiometricAuthenticator();
      service = HapticTradeConfirmationService(
        hapticDriver: mockHaptics,
        biometricAuthenticator: mockBiometric,
        swipeThreshold: 0.85,
      );

      testIntent = const TradeConfirmationIntent(
        orderId: 'ORD_HAPTIC_8831',
        symbol: 'RELIANCE',
        side: OrderSide.buy,
        orderType: OrderType.limit,
        quantity: 50.0,
        price: 2950.0,
        requireBiometric: true,
        mpcKeyShardId: 'shard_secp256k1_client_key_01',
        timestampMs: 1720000000000,
      );

      controller = HapticTradeConfirmationController(
        service: service,
        intent: testIntent,
      );
    });

    tearDown(() {
      controller.dispose();
    });

    test('TradeConfirmationIntent computes deterministic order digest and notional value', () {
      expect(testIntent.notionalValue, equals(147500.0));
      final digest1 = testIntent.computeOrderDigest();
      final digest2 = testIntent.computeOrderDigest();
      expect(digest1, equals(digest2));
      expect(digest1.length, equals(64)); // SHA-256 hex string
    });

    test('Haptic triggers provide progressive tactile ticks and impact on threshold', () {
      mockHaptics.clear();
      // Progress from 0.0 to 0.3 should fire light tap
      service.onSliderProgressChanged(0.0, 0.3);
      expect(mockHaptics.history, contains(HapticPattern.lightTap));

      // Progress past 0.85 should fire medium impact
      mockHaptics.clear();
      service.onSliderProgressChanged(0.7, 0.9);
      expect(mockHaptics.history, contains(HapticPattern.mediumImpact));
    });

    test('Controller handles slider progress clamp and threshold transition', () {
      controller.updateSliderProgress(0.5);
      expect(controller.state.progress, equals(0.5));
      expect(controller.state.state, equals(SwipeConfirmationState.swiping));

      controller.updateSliderProgress(0.9);
      expect(controller.state.progress, equals(0.9));
      expect(controller.state.state, equals(SwipeConfirmationState.thresholdReached));
    });

    test('Successful biometric authentication signs order and generates heavy pulse', () async {
      mockBiometric.shouldSucceed = true;
      controller.updateSliderProgress(0.9);

      final result = await controller.executeTradeConfirmation();
      expect(result, isTrue);
      expect(controller.state.state, equals(SwipeConfirmationState.submitted));
      expect(controller.state.signature, isNotNull);
      expect(mockHaptics.history, contains(HapticPattern.heavyPulse));
    });

    test('Failed biometric rejects order, sets failure state, and emits warning buzz', () async {
      mockBiometric.shouldSucceed = false;
      controller.updateSliderProgress(0.9);

      final result = await controller.executeTradeConfirmation();
      expect(result, isFalse);
      expect(controller.state.state, equals(SwipeConfirmationState.failed));
      expect(mockHaptics.history, contains(HapticPattern.warningBuzz));
    });

    test('Submitting before threshold resets slider to idle', () async {
      controller.updateSliderProgress(0.4);
      final result = await controller.executeTradeConfirmation();
      expect(result, isFalse);
      expect(controller.state.state, equals(SwipeConfirmationState.idle));
      expect(controller.state.progress, equals(0.0));
    });
  });
}
