import 'commodity_denomination.dart';
import 'commodity_enums.dart';
import 'delivery_address.dart';
import 'sub_paise_fee_breakdown.dart';
import 'wdra_vault_location.dart';

class DenominationOrderItem {
  final CommodityDenomination denomination;
  final int quantity;

  const DenominationOrderItem({
    required this.denomination,
    required this.quantity,
  });

  double get totalWeightGrams => denomination.weightGrams * quantity;

  Map<String, dynamic> toJson() => {
        'denomination': denomination.toJson(),
        'quantity': quantity,
      };

  factory DenominationOrderItem.fromJson(Map<String, dynamic> json) {
    return DenominationOrderItem(
      denomination: CommodityDenomination.fromJson(json['denomination'] as Map<String, dynamic>),
      quantity: json['quantity'] as int,
    );
  }
}

class RedemptionOrder {
  final String orderId;
  final String userId;
  final CommodityType commodityType;
  final List<DenominationOrderItem> items;
  final double totalFineGrams;
  final DeliveryMode deliveryMode;
  final WdraVaultLocation? selectedVault;
  final DeliveryAddress? deliveryAddress;
  final SubPaiseFeeBreakdown feeBreakdown;
  final RedemptionStatus status;
  final String? enwrNumber;
  final String? onChainBurnTxHash;
  final String? idempotencyKey;
  final DateTime createdAt;
  final DateTime? estimatedDeliveryDate;

  const RedemptionOrder({
    required this.orderId,
    required this.userId,
    required this.commodityType,
    required this.items,
    required this.totalFineGrams,
    required this.deliveryMode,
    this.selectedVault,
    this.deliveryAddress,
    required this.feeBreakdown,
    required this.status,
    this.enwrNumber,
    this.onChainBurnTxHash,
    this.idempotencyKey,
    required this.createdAt,
    this.estimatedDeliveryDate,
  });

  RedemptionOrder copyWith({
    RedemptionStatus? status,
    String? enwrNumber,
    String? onChainBurnTxHash,
    DateTime? estimatedDeliveryDate,
  }) {
    return RedemptionOrder(
      orderId: orderId,
      userId: userId,
      commodityType: commodityType,
      items: items,
      totalFineGrams: totalFineGrams,
      deliveryMode: deliveryMode,
      selectedVault: selectedVault,
      deliveryAddress: deliveryAddress,
      feeBreakdown: feeBreakdown,
      status: status ?? this.status,
      enwrNumber: enwrNumber ?? this.enwrNumber,
      onChainBurnTxHash: onChainBurnTxHash ?? this.onChainBurnTxHash,
      idempotencyKey: idempotencyKey,
      createdAt: createdAt,
      estimatedDeliveryDate: estimatedDeliveryDate ?? this.estimatedDeliveryDate,
    );
  }

  Map<String, dynamic> toJson() => {
        'order_id': orderId,
        'user_id': userId,
        'commodity_type': commodityType.name,
        'items': items.map((e) => e.toJson()).toList(),
        'total_fine_grams': totalFineGrams,
        'delivery_mode': deliveryMode.name,
        'selected_vault': selectedVault?.toJson(),
        'delivery_address': deliveryAddress?.toJson(),
        'fee_breakdown': feeBreakdown.toJson(),
        'status': status.name,
        'enwr_number': enwrNumber,
        'on_chain_burn_tx_hash': onChainBurnTxHash,
        'idempotency_key': idempotencyKey,
        'created_at': createdAt.toIso8601String(),
        'estimated_delivery_date': estimatedDeliveryDate?.toIso8601String(),
      };

  factory RedemptionOrder.fromJson(Map<String, dynamic> json) {
    return RedemptionOrder(
      orderId: json['order_id'] as String,
      userId: json['user_id'] as String,
      commodityType: CommodityType.values.byName(json['commodity_type'] as String),
      items: (json['items'] as List<dynamic>)
          .map((e) => DenominationOrderItem.fromJson(e as Map<String, dynamic>))
          .toList(),
      totalFineGrams: (json['total_fine_grams'] as num).toDouble(),
      deliveryMode: DeliveryMode.values.byName(json['delivery_mode'] as String),
      selectedVault: json['selected_vault'] != null
          ? WdraVaultLocation.fromJson(json['selected_vault'] as Map<String, dynamic>)
          : null,
      deliveryAddress: json['delivery_address'] != null
          ? DeliveryAddress.fromJson(json['delivery_address'] as Map<String, dynamic>)
          : null,
      feeBreakdown: SubPaiseFeeBreakdown.fromJson(json['fee_breakdown'] as Map<String, dynamic>),
      status: RedemptionStatus.values.byName(json['status'] as String),
      enwrNumber: json['enwr_number'] as String?,
      onChainBurnTxHash: json['on_chain_burn_tx_hash'] as String?,
      idempotencyKey: json['idempotency_key'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      estimatedDeliveryDate: json['estimated_delivery_date'] != null
          ? DateTime.parse(json['estimated_delivery_date'] as String)
          : null,
    );
  }
}
