import 'dart:convert';
import 'package:crypto/crypto.dart';

enum OrderSide {
  buy,
  sell;

  bool get isBuy => this == OrderSide.buy;
  bool get isSell => this == OrderSide.sell;
}

enum OrderType {
  market,
  limit,
  stopLimit,
  iceberg;
}

enum HapticPattern {
  selectionClick,
  lightTap,
  mediumImpact,
  heavyPulse,
  warningBuzz;
}

enum SwipeConfirmationState {
  idle,
  swiping,
  thresholdReached,
  biometricPrompted,
  biometricSuccess,
  orderSigned,
  submitted,
  failed;

  bool get isCompleted => this == SwipeConfirmationState.submitted;
  bool get isFailed => this == SwipeConfirmationState.failed;
}

class TradeConfirmationIntent {
  final String orderId;
  final String symbol;
  final OrderSide side;
  final OrderType orderType;
  final double quantity;
  final double price;
  final double slippageTolerancePct;
  final bool requireBiometric;
  final String? mpcKeyShardId;
  final int timestampMs;

  const TradeConfirmationIntent({
    required this.orderId,
    required this.symbol,
    required this.side,
    required this.orderType,
    required this.quantity,
    required this.price,
    this.slippageTolerancePct = 0.5,
    this.requireBiometric = true,
    this.mpcKeyShardId,
    required this.timestampMs,
  });

  double get notionalValue => quantity * price;

  String computeOrderDigest() {
    final payload = '$orderId|$symbol|${side.name}|${orderType.name}|$quantity|$price|$timestampMs';
    return sha256.convert(utf8.encode(payload)).toString();
  }
}
