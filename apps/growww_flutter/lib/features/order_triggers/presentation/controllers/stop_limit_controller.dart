import 'dart:async';
import '../../domain/models/stop_limit_config.dart';
import '../../domain/models/trigger_enums.dart';
import '../../domain/services/trigger_validation_engine.dart';

class StopLimitState {
  final String symbol;
  final String side; // BUY or SELL
  final double currentMarketPrice;
  final double markPrice;
  final TriggerType triggerType;
  final TriggerPriceType triggerPriceType;
  final double triggerPrice;
  final double limitPrice;
  final double quantity;
  final double? trailingDeltaPercent;
  final TriggerValidationResult validation;
  final bool isSubmitting;
  final String? submittedOrderId;
  final String? toastMessage;

  const StopLimitState({
    this.symbol = 'BTC/USDT',
    this.side = 'SELL',
    this.currentMarketPrice = 64200.0,
    this.markPrice = 64210.5,
    this.triggerType = TriggerType.stopLimit,
    this.triggerPriceType = TriggerPriceType.lastPrice,
    this.triggerPrice = 62000.0,
    this.limitPrice = 61950.0,
    this.quantity = 0.5,
    this.trailingDeltaPercent,
    this.validation = const TriggerValidationResult(isValid: true),
    this.isSubmitting = false,
    this.submittedOrderId,
    this.toastMessage,
  });

  double get referencePrice =>
      triggerPriceType == TriggerPriceType.markPrice ? markPrice : currentMarketPrice;

  double get estimatedTotalInr =>
      (triggerType.isLimitOrder ? limitPrice : triggerPrice) * quantity;

  StopLimitConfig toConfig() {
    final isBuy = side.toUpperCase() == 'BUY';
    final condition = isBuy
        ? (triggerType == TriggerType.stopLimit || triggerType == TriggerType.stopMarket
            ? TriggerCondition.greaterOrEqual
            : TriggerCondition.lessOrEqual)
        : (triggerType == TriggerType.stopLimit || triggerType == TriggerType.stopMarket
            ? TriggerCondition.lessOrEqual
            : TriggerCondition.greaterOrEqual);

    return StopLimitConfig(
      orderId: 'ORD-TRIG-${DateTime.now().millisecondsSinceEpoch}',
      symbol: symbol,
      side: side,
      triggerType: triggerType,
      triggerPriceType: triggerPriceType,
      triggerCondition: condition,
      triggerPrice: triggerPrice,
      limitPrice: triggerType.isLimitOrder ? limitPrice : null,
      quantity: quantity,
      trailingDeltaPercent: trailingDeltaPercent,
      currentMarketPrice: referencePrice,
      createdAt: DateTime.now(),
    );
  }

  StopLimitState copyWith({
    String? symbol,
    String? side,
    double? currentMarketPrice,
    double? markPrice,
    TriggerType? triggerType,
    TriggerPriceType? triggerPriceType,
    double? triggerPrice,
    double? limitPrice,
    double? quantity,
    double? trailingDeltaPercent,
    TriggerValidationResult? validation,
    bool? isSubmitting,
    String? submittedOrderId,
    String? toastMessage,
  }) {
    return StopLimitState(
      symbol: symbol ?? this.symbol,
      side: side ?? this.side,
      currentMarketPrice: currentMarketPrice ?? this.currentMarketPrice,
      markPrice: markPrice ?? this.markPrice,
      triggerType: triggerType ?? this.triggerType,
      triggerPriceType: triggerPriceType ?? this.triggerPriceType,
      triggerPrice: triggerPrice ?? this.triggerPrice,
      limitPrice: limitPrice ?? this.limitPrice,
      quantity: quantity ?? this.quantity,
      trailingDeltaPercent: trailingDeltaPercent ?? this.trailingDeltaPercent,
      validation: validation ?? this.validation,
      isSubmitting: isSubmitting ?? this.isSubmitting,
      submittedOrderId: submittedOrderId,
      toastMessage: toastMessage,
    );
  }
}

class StopLimitController {
  StopLimitState _state = const StopLimitState();
  final _stateController = StreamController<StopLimitState>.broadcast();

  StopLimitState get state => _state;
  Stream<StopLimitState> get stream => _stateController.stream;

  StopLimitController() {
    _validateState();
  }

  void _emit(StopLimitState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void _validateState() {
    final config = _state.toConfig();
    final result = TriggerValidationEngine.validate(config);
    _emit(_state.copyWith(validation: result));
  }

  void setSide(String side) {
    _emit(_state.copyWith(side: side));
    _validateState();
  }

  void setTriggerType(TriggerType type) {
    _emit(_state.copyWith(triggerType: type));
    _validateState();
  }

  void setPriceType(TriggerPriceType priceType) {
    _emit(_state.copyWith(triggerPriceType: priceType));
    _validateState();
  }

  void setTriggerPrice(double price) {
    _emit(_state.copyWith(triggerPrice: price));
    _validateState();
  }

  void setLimitPrice(double price) {
    _emit(_state.copyWith(limitPrice: price));
    _validateState();
  }

  void setQuantity(double qty) {
    _emit(_state.copyWith(quantity: qty));
    _validateState();
  }

  Future<bool> submitOrder() async {
    _validateState();
    if (!_state.validation.isValid) return false;

    _emit(_state.copyWith(isSubmitting: true));
    await Future.delayed(const Duration(milliseconds: 300));

    final orderId = 'ORD-TRIG-${DateTime.now().millisecondsSinceEpoch}';
    _emit(_state.copyWith(
      isSubmitting: false,
      submittedOrderId: orderId,
      toastMessage: 'Conditional ${_state.triggerType.displayName} armed successfully!',
    ));
    return true;
  }

  void dispose() {
    _stateController.close();
  }
}
