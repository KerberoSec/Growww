import 'depth_ladder_enums.dart';

/// Individual orderbook entry for Level 3 Market-By-Order (MBO) order queue inspection.
class L3OrderEntry {
  final String orderId;
  final String traderMaskedId;
  final double price;
  final double quantity;
  final DomOrderSide side;
  final int queuePriority;
  final DateTime timestamp;

  const L3OrderEntry({
    required this.orderId,
    required this.traderMaskedId,
    required this.price,
    required this.quantity,
    required this.side,
    required this.queuePriority,
    required this.timestamp,
  });

  L3OrderEntry copyWith({
    String? orderId,
    String? traderMaskedId,
    double? price,
    double? quantity,
    DomOrderSide? side,
    int? queuePriority,
    DateTime? timestamp,
  }) {
    return L3OrderEntry(
      orderId: orderId ?? this.orderId,
      traderMaskedId: traderMaskedId ?? this.traderMaskedId,
      price: price ?? this.price,
      quantity: quantity ?? this.quantity,
      side: side ?? this.side,
      queuePriority: queuePriority ?? this.queuePriority,
      timestamp: timestamp ?? this.timestamp,
    );
  }
}
