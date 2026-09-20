import 'depth_ladder_enums.dart';

/// Single aggregated price level on the DOM Depth Ladder.
class LadderDepthLevel {
  final double price;
  final double quantity;
  final int orderCount;
  final double cumulativeQuantity;
  final double depthPercent; // 0.0 to 1.0 for visual bar filling
  final DomOrderSide side;
  final bool isBestPrice;

  const LadderDepthLevel({
    required this.price,
    required this.quantity,
    this.orderCount = 1,
    this.cumulativeQuantity = 0.0,
    this.depthPercent = 0.0,
    required this.side,
    this.isBestPrice = false,
  });

  LadderDepthLevel copyWith({
    double? price,
    double? quantity,
    int? orderCount,
    double? cumulativeQuantity,
    double? depthPercent,
    DomOrderSide? side,
    bool? isBestPrice,
  }) {
    return LadderDepthLevel(
      price: price ?? this.price,
      quantity: quantity ?? this.quantity,
      orderCount: orderCount ?? this.orderCount,
      cumulativeQuantity: cumulativeQuantity ?? this.cumulativeQuantity,
      depthPercent: depthPercent ?? this.depthPercent,
      side: side ?? this.side,
      isBestPrice: isBestPrice ?? this.isBestPrice,
    );
  }
}
