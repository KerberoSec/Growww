import 'package:test/test.dart';
import 'package:growww_flutter/features/order_triggers/domain/models/stop_limit_config.dart';
import 'package:growww_flutter/features/order_triggers/domain/models/trigger_enums.dart';
import 'package:growww_flutter/features/order_triggers/domain/services/trigger_validation_engine.dart';
import 'package:growww_flutter/features/order_triggers/presentation/controllers/stop_limit_controller.dart';

void main() {
  group('Prompt 549 - Stop-Limit and Conditional Trigger Configuration', () {
    late StopLimitController controller;

    setUp(() {
      controller = StopLimitController();
    });

    tearDown(() {
      controller.dispose();
    });

    test('TriggerValidationEngine enforces directional logic for Buy Stop and Sell Stop orders', () {
      // Buy Stop: Trigger price must be >= current market price
      final validBuyStop = StopLimitConfig(
        orderId: 'T1',
        symbol: 'BTC/USDT',
        side: 'BUY',
        triggerType: TriggerType.stopLimit,
        triggerPriceType: TriggerPriceType.lastPrice,
        triggerCondition: TriggerCondition.greaterOrEqual,
        triggerPrice: 65000.0, // Above current 64000
        limitPrice: 65100.0,
        quantity: 1.0,
        currentMarketPrice: 64000.0,
        createdAt: DateTime.now(),
      );
      final res1 = TriggerValidationEngine.validate(validBuyStop);
      expect(res1.isValid, isTrue);

      final invalidBuyStop = StopLimitConfig(
        orderId: 'T2',
        symbol: 'BTC/USDT',
        side: 'BUY',
        triggerType: TriggerType.stopLimit,
        triggerPriceType: TriggerPriceType.lastPrice,
        triggerCondition: TriggerCondition.greaterOrEqual,
        triggerPrice: 63000.0, // Invalid: below current 64000
        limitPrice: 63100.0,
        quantity: 1.0,
        currentMarketPrice: 64000.0,
        createdAt: DateTime.now(),
      );
      final res2 = TriggerValidationEngine.validate(invalidBuyStop);
      expect(res2.isValid, isFalse);
      expect(res2.errorMessage, contains('must be higher than or equal to current price'));

      // Sell Stop: Trigger price must be <= current market price
      final validSellStop = StopLimitConfig(
        orderId: 'T3',
        symbol: 'BTC/USDT',
        side: 'SELL',
        triggerType: TriggerType.stopMarket,
        triggerPriceType: TriggerPriceType.markPrice,
        triggerCondition: TriggerCondition.lessOrEqual,
        triggerPrice: 62000.0, // Below current 64000
        quantity: 1.0,
        currentMarketPrice: 64000.0,
        createdAt: DateTime.now(),
      );
      final res3 = TriggerValidationEngine.validate(validSellStop);
      expect(res3.isValid, isTrue);
    });

    test('TriggerValidationEngine checks limit vs trigger price slippage buffer', () {
      // Buy Stop with Limit below Trigger (unfavorable fill warning)
      final buyStopWarning = StopLimitConfig(
        orderId: 'T4',
        symbol: 'BTC/USDT',
        side: 'BUY',
        triggerType: TriggerType.stopLimit,
        triggerPriceType: TriggerPriceType.lastPrice,
        triggerCondition: TriggerCondition.greaterOrEqual,
        triggerPrice: 65000.0,
        limitPrice: 64900.0, // Below trigger
        quantity: 1.0,
        currentMarketPrice: 64000.0,
        createdAt: DateTime.now(),
      );
      final res = TriggerValidationEngine.validate(buyStopWarning);
      expect(res.isValid, isTrue);
      expect(res.warningMessage, isNotNull);
      expect(res.warningMessage, contains('May not execute immediately'));
    });

    test('TriggerValidationEngine rejects misaligned price ticks', () {
      final invalidTick = StopLimitConfig(
        orderId: 'T5',
        symbol: 'BTC/USDT',
        side: 'BUY',
        triggerType: TriggerType.stopMarket,
        triggerPriceType: TriggerPriceType.lastPrice,
        triggerCondition: TriggerCondition.greaterOrEqual,
        triggerPrice: 65000.03, // Invalid tick (not multiple of 0.05)
        quantity: 1.0,
        currentMarketPrice: 64000.0,
        createdAt: DateTime.now(),
      );
      final res = TriggerValidationEngine.validate(invalidTick);
      expect(res.isValid, isFalse);
      expect(res.errorMessage, contains('₹0.05 tick size'));
    });

    test('StopLimitController manages interactive configuration and order dispatch', () async {
      expect(controller.state.side, equals('SELL'));
      expect(controller.state.validation.isValid, isTrue);

      controller.setSide('BUY');
      controller.setTriggerPrice(66000.0);
      controller.setLimitPrice(66050.0);
      controller.setQuantity(2.0);

      expect(controller.state.estimatedTotalInr, equals(132100.0));
      expect(controller.state.validation.isValid, isTrue);

      final submitted = await controller.submitOrder();
      expect(submitted, isTrue);
      expect(controller.state.submittedOrderId, isNotNull);
      expect(controller.state.toastMessage, contains('Conditional Stop Limit armed successfully'));
    });
  });
}
