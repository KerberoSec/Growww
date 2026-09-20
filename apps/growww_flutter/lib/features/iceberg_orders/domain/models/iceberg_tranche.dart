import 'iceberg_enums.dart';

class IcebergTranche {
  final int trancheIndex;
  final double visibleQuantity;
  final double hiddenRemaining;
  final double executedQuantity;
  final double price;
  final TrancheStatus status;
  final DateTime? filledAt;

  const IcebergTranche({
    required this.trancheIndex,
    required this.visibleQuantity,
    required this.hiddenRemaining,
    required this.executedQuantity,
    required this.price,
    required this.status,
    this.filledAt,
  });

  bool get isFilled => status == TrancheStatus.filled;

  IcebergTranche copyWith({
    double? executedQuantity,
    TrancheStatus? status,
    DateTime? filledAt,
  }) {
    return IcebergTranche(
      trancheIndex: trancheIndex,
      visibleQuantity: visibleQuantity,
      hiddenRemaining: hiddenRemaining,
      executedQuantity: executedQuantity ?? this.executedQuantity,
      price: price,
      status: status ?? this.status,
      filledAt: filledAt ?? this.filledAt,
    );
  }

  Map<String, dynamic> toJson() => {
        'tranche_index': trancheIndex,
        'visible_quantity': visibleQuantity,
        'hidden_remaining': hiddenRemaining,
        'executed_quantity': executedQuantity,
        'price': price,
        'status': status.name,
        'filled_at': filledAt?.toIso8601String(),
      };

  factory IcebergTranche.fromJson(Map<String, dynamic> json) {
    return IcebergTranche(
      trancheIndex: json['tranche_index'] as int,
      visibleQuantity: (json['visible_quantity'] as num).toDouble(),
      hiddenRemaining: (json['hidden_remaining'] as num).toDouble(),
      executedQuantity: (json['executed_quantity'] as num).toDouble(),
      price: (json['price'] as num).toDouble(),
      status: TrancheStatus.values.byName(json['status'] as String),
      filledAt: json['filled_at'] != null ? DateTime.parse(json['filled_at'] as String) : null,
    );
  }
}
