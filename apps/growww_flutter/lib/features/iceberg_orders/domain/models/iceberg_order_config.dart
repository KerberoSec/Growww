import 'iceberg_enums.dart';

class IcebergOrderConfig {
  final String orderId;
  final String symbol;
  final String side; // BUY or SELL
  final double totalQuantity;
  final double visibleTrancheQuantity;
  final IcebergVarianceMode varianceMode;
  final double priceLimit;
  final DateTime createdAt;

  const IcebergOrderConfig({
    required this.orderId,
    required this.symbol,
    required this.side,
    required this.totalQuantity,
    required this.visibleTrancheQuantity,
    required this.varianceMode,
    required this.priceLimit,
    required this.createdAt,
  });

  double get estimatedTranches => visibleTrancheQuantity > 0
      ? (totalQuantity / visibleTrancheQuantity)
      : 1.0;

  double get hiddenQuantity => (totalQuantity - visibleTrancheQuantity).clamp(0.0, totalQuantity);

  Map<String, dynamic> toJson() => {
        'order_id': orderId,
        'symbol': symbol,
        'side': side,
        'total_quantity': totalQuantity,
        'visible_tranche_quantity': visibleTrancheQuantity,
        'variance_mode': varianceMode.name,
        'price_limit': priceLimit,
        'created_at': createdAt.toIso8601String(),
      };

  factory IcebergOrderConfig.fromJson(Map<String, dynamic> json) {
    return IcebergOrderConfig(
      orderId: json['order_id'] as String,
      symbol: json['symbol'] as String,
      side: json['side'] as String,
      totalQuantity: (json['total_quantity'] as num).toDouble(),
      visibleTrancheQuantity: (json['visible_tranche_quantity'] as num).toDouble(),
      varianceMode: IcebergVarianceMode.values.byName(json['variance_mode'] as String),
      priceLimit: (json['price_limit'] as num).toDouble(),
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}
