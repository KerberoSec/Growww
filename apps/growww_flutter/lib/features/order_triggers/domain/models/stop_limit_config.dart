import 'trigger_enums.dart';

class StopLimitConfig {
  final String orderId;
  final String symbol;
  final String side; // "BUY" or "SELL"
  final TriggerType triggerType;
  final TriggerPriceType triggerPriceType;
  final TriggerCondition triggerCondition;
  final double triggerPrice;
  final double? limitPrice; // Null if market order
  final double quantity;
  final double? trailingDeltaPercent; // For trailing stop
  final double slippageTolerancePercent;
  final double currentMarketPrice;
  final DateTime createdAt;

  const StopLimitConfig({
    required this.orderId,
    required this.symbol,
    required this.side,
    required this.triggerType,
    required this.triggerPriceType,
    required this.triggerCondition,
    required this.triggerPrice,
    this.limitPrice,
    required this.quantity,
    this.trailingDeltaPercent,
    this.slippageTolerancePercent = 0.5,
    required this.currentMarketPrice,
    required this.createdAt,
  });

  double get executionPrice => limitPrice ?? triggerPrice;
  double get estimatedTotalInr => executionPrice * quantity;
  double get triggerDistancePercent =>
      ((triggerPrice - currentMarketPrice) / currentMarketPrice) * 100.0;

  Map<String, dynamic> toJson() => {
        'order_id': orderId,
        'symbol': symbol,
        'side': side,
        'trigger_type': triggerType.name,
        'trigger_price_type': triggerPriceType.name,
        'trigger_condition': triggerCondition.name,
        'trigger_price': triggerPrice,
        'limit_price': limitPrice,
        'quantity': quantity,
        'trailing_delta_percent': trailingDeltaPercent,
        'slippage_tolerance_percent': slippageTolerancePercent,
        'current_market_price': currentMarketPrice,
        'created_at': createdAt.toIso8601String(),
      };

  factory StopLimitConfig.fromJson(Map<String, dynamic> json) {
    return StopLimitConfig(
      orderId: json['order_id'] as String,
      symbol: json['symbol'] as String,
      side: json['side'] as String,
      triggerType: TriggerType.values.byName(json['trigger_type'] as String),
      triggerPriceType: TriggerPriceType.values.byName(json['trigger_price_type'] as String),
      triggerCondition: TriggerCondition.values.byName(json['trigger_condition'] as String),
      triggerPrice: (json['trigger_price'] as num).toDouble(),
      limitPrice: (json['limit_price'] as num?)?.toDouble(),
      quantity: (json['quantity'] as num).toDouble(),
      trailingDeltaPercent: (json['trailing_delta_percent'] as num?)?.toDouble(),
      slippageTolerancePercent: (json['slippage_tolerance_percent'] as num).toDouble(),
      currentMarketPrice: (json['current_market_price'] as num).toDouble(),
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}
