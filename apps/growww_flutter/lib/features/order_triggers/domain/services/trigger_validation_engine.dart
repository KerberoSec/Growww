import '../models/stop_limit_config.dart';
import '../models/trigger_enums.dart';

class TriggerValidationResult {
  final bool isValid;
  final String? errorMessage;
  final String? warningMessage;

  const TriggerValidationResult({
    required this.isValid,
    this.errorMessage,
    this.warningMessage,
  });

  factory TriggerValidationResult.valid({String? warning}) =>
      TriggerValidationResult(isValid: true, warningMessage: warning);

  factory TriggerValidationResult.invalid(String error) =>
      TriggerValidationResult(isValid: false, errorMessage: error);
}

class TriggerValidationEngine {
  static const double minTickSize = 0.05;

  static TriggerValidationResult validate(StopLimitConfig config) {
    if (config.quantity <= 0) {
      return TriggerValidationResult.invalid('Order quantity must be greater than zero.');
    }

    if (config.triggerPrice <= 0) {
      return TriggerValidationResult.invalid('Trigger price must be greater than zero.');
    }

    // Tick alignment check (within float epsilon)
    final tickRemainder = (config.triggerPrice / minTickSize).roundToDouble() * minTickSize;
    if ((config.triggerPrice - tickRemainder).abs() > 0.0001) {
      return TriggerValidationResult.invalid('Trigger price must align to ₹0.05 tick size.');
    }

    final isBuy = config.side.toUpperCase() == 'BUY';
    final cur = config.currentMarketPrice;
    final trig = config.triggerPrice;

    // Trigger directional consistency check
    switch (config.triggerType) {
      case TriggerType.stopLimit:
      case TriggerType.stopMarket:
        if (isBuy && trig < cur) {
          return TriggerValidationResult.invalid(
            'For a Buy Stop, trigger price (₹${trig.toStringAsFixed(2)}) must be higher than or equal to current price (₹${cur.toStringAsFixed(2)}).',
          );
        }
        if (!isBuy && trig > cur) {
          return TriggerValidationResult.invalid(
            'For a Sell Stop, trigger price (₹${trig.toStringAsFixed(2)}) must be lower than or equal to current price (₹${cur.toStringAsFixed(2)}).',
          );
        }
        break;

      case TriggerType.takeProfitLimit:
      case TriggerType.takeProfitMarket:
        if (isBuy && trig > cur) {
          return TriggerValidationResult.invalid(
            'For a Buy Take-Profit, trigger price must be lower than or equal to current price.',
          );
        }
        if (!isBuy && trig < cur) {
          return TriggerValidationResult.invalid(
            'For a Sell Take-Profit, trigger price must be higher than or equal to current price.',
          );
        }
        break;

      case TriggerType.trailingStop:
        if (config.trailingDeltaPercent == null || config.trailingDeltaPercent! <= 0 || config.trailingDeltaPercent! > 50.0) {
          return TriggerValidationResult.invalid('Trailing delta must be between 0.1% and 50.0%.');
        }
        break;
    }

    // Slippage buffer checks for Limit Orders
    if (config.triggerType.isLimitOrder) {
      if (config.limitPrice == null || config.limitPrice! <= 0) {
        return TriggerValidationResult.invalid('Limit price is required for Stop Limit orders.');
      }

      final lim = config.limitPrice!;
      if (isBuy && lim < trig) {
        return TriggerValidationResult.valid(
          warning: 'Buy Limit price (₹${lim.toStringAsFixed(2)}) is below Trigger (₹${trig.toStringAsFixed(2)}). May not execute immediately.',
        );
      }
      if (!isBuy && lim > trig) {
        return TriggerValidationResult.valid(
          warning: 'Sell Limit price (₹${lim.toStringAsFixed(2)}) is above Trigger (₹${trig.toStringAsFixed(2)}). May not execute immediately.',
        );
      }
    }

    return TriggerValidationResult.valid();
  }
}
