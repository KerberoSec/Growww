/// Order side representing transaction intent (BUY / SELL).
enum OrderSide {
  buy,
  sell;

  String get label => name.toUpperCase();

  static OrderSide fromString(String value) {
    return OrderSide.values.firstWhere(
      (e) => e.name.toLowerCase() == value.toLowerCase(),
      orElse: () => OrderSide.buy,
    );
  }
}

/// Order execution type supported across the matching engine.
enum OrderType {
  limit,
  market,
  stopLimit,
  stopMarket,
  trailingStop,
  iceberg,
  oco,
  twap;

  String get displayName {
    switch (this) {
      case OrderType.limit:
        return 'Limit';
      case OrderType.market:
        return 'Market';
      case OrderType.stopLimit:
        return 'Stop Limit';
      case OrderType.stopMarket:
        return 'Stop Market';
      case OrderType.trailingStop:
        return 'Trailing Stop';
      case OrderType.iceberg:
        return 'Iceberg';
      case OrderType.oco:
        return 'OCO (One-Cancels-Other)';
      case OrderType.twap:
        return 'TWAP';
    }
  }

  static OrderType fromString(String value) {
    return OrderType.values.firstWhere(
      (e) => e.name.toLowerCase() == value.replaceAll('-', '').replaceAll('_', '').toLowerCase(),
      orElse: () => OrderType.limit,
    );
  }
}

/// Time in force execution policies.
enum TimeInForce {
  gtc, // Good 'Til Cancelled
  ioc, // Immediate Or Cancel
  fok, // Fill Or Kill
  postOnly; // Maker Only (rejects if would cross spread)

  String get code {
    switch (this) {
      case TimeInForce.gtc:
        return 'GTC';
      case TimeInForce.ioc:
        return 'IOC';
      case TimeInForce.fok:
        return 'FOK';
      case TimeInForce.postOnly:
        return 'POST_ONLY';
    }
  }

  static TimeInForce fromString(String value) {
    switch (value.toUpperCase()) {
      case 'IOC':
        return TimeInForce.ioc;
      case 'FOK':
        return TimeInForce.fok;
      case 'POST_ONLY':
      case 'POSTONLY':
        return TimeInForce.postOnly;
      case 'GTC':
      default:
        return TimeInForce.gtc;
    }
  }
}

/// Order lifecycle status states.
enum OrderStatus {
  pending,
  newOrder,
  partiallyFilled,
  filled,
  cancelled,
  rejected,
  expired;

  String get label {
    switch (this) {
      case OrderStatus.pending:
        return 'Pending';
      case OrderStatus.newOrder:
        return 'Open';
      case OrderStatus.partiallyFilled:
        return 'Partially Filled';
      case OrderStatus.filled:
        return 'Filled';
      case OrderStatus.cancelled:
        return 'Cancelled';
      case OrderStatus.rejected:
        return 'Rejected';
      case OrderStatus.expired:
        return 'Expired';
    }
  }

  static OrderStatus fromString(String value) {
    switch (value.toLowerCase()) {
      case 'pending':
        return OrderStatus.pending;
      case 'new':
      case 'neworder':
      case 'open':
        return OrderStatus.newOrder;
      case 'partially_filled':
      case 'partiallyfilled':
        return OrderStatus.partiallyFilled;
      case 'filled':
        return OrderStatus.filled;
      case 'cancelled':
      case 'canceled':
        return OrderStatus.cancelled;
      case 'rejected':
        return OrderStatus.rejected;
      case 'expired':
        return OrderStatus.expired;
      default:
        return OrderStatus.newOrder;
    }
  }
}

/// Settlement channel for sovereign exchange orders.
enum SettlementMode {
  instantDvP,
  besuOnChain,
  offChainMatching;

  static SettlementMode fromString(String value) {
    switch (value.toLowerCase()) {
      case 'instantdvp':
      case 'instant_dvp':
        return SettlementMode.instantDvP;
      case 'besuonchain':
      case 'besu_on_chain':
        return SettlementMode.besuOnChain;
      default:
        return SettlementMode.offChainMatching;
    }
  }
}

/// Production-grade Order entity representing resting and completed orders.
class Order {
  final String id;
  final String clientOrderId;
  final String symbol;
  final OrderSide side;
  final OrderType type;
  final TimeInForce timeInForce;
  final double price;
  final double quantity;
  final double filledQuantity;
  final double averageFillPrice;
  final OrderStatus status;
  final double? stopPrice;
  final double? icebergVisibleQty;
  final double? trailingDeltaPercent;
  final double fee; // Zero-fee presentation invariant: 0.00%
  final String feeCurrency;
  final bool isBiometricVerified;
  final String? biometricSignature;
  final SettlementMode settlementMode;
  final DateTime createdAt;
  final DateTime updatedAt;
  final String? onChainTxHash;

  const Order({
    required this.id,
    required this.clientOrderId,
    required this.symbol,
    required this.side,
    required this.type,
    required this.timeInForce,
    required this.price,
    required this.quantity,
    this.filledQuantity = 0.0,
    this.averageFillPrice = 0.0,
    required this.status,
    this.stopPrice,
    this.icebergVisibleQty,
    this.trailingDeltaPercent,
    this.fee = 0.0,
    this.feeCurrency = 'USDT',
    this.isBiometricVerified = false,
    this.biometricSignature,
    this.settlementMode = SettlementMode.instantDvP,
    required this.createdAt,
    required this.updatedAt,
    this.onChainTxHash,
  });

  /// Remaining unfilled quantity.
  double get remainingQuantity => (quantity - filledQuantity).clamp(0.0, quantity);

  /// Fill percentage from 0.0% to 100.0%.
  double get fillPercentage => quantity > 0 ? ((filledQuantity / quantity) * 100.0).clamp(0.0, 100.0) : 0.0;

  /// Total notional value at order price.
  double get notionalValue => price * quantity;

  /// Total filled notional value.
  double get filledValue => averageFillPrice * filledQuantity;

  /// True if terminal state reached.
  bool get isDone =>
      status == OrderStatus.filled ||
      status == OrderStatus.cancelled ||
      status == OrderStatus.rejected ||
      status == OrderStatus.expired;

  /// True if order can be cancelled.
  bool get isCancellable =>
      status == OrderStatus.pending ||
      status == OrderStatus.newOrder ||
      status == OrderStatus.partiallyFilled;

  Order copyWith({
    String? id,
    String? clientOrderId,
    String? symbol,
    OrderSide? side,
    OrderType? type,
    TimeInForce? timeInForce,
    double? price,
    double? quantity,
    double? filledQuantity,
    double? averageFillPrice,
    OrderStatus? status,
    double? stopPrice,
    double? icebergVisibleQty,
    double? trailingDeltaPercent,
    double? fee,
    String? feeCurrency,
    bool? isBiometricVerified,
    String? biometricSignature,
    SettlementMode? settlementMode,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? onChainTxHash,
  }) {
    return Order(
      id: id ?? this.id,
      clientOrderId: clientOrderId ?? this.clientOrderId,
      symbol: symbol ?? this.symbol,
      side: side ?? this.side,
      type: type ?? this.type,
      timeInForce: timeInForce ?? this.timeInForce,
      price: price ?? this.price,
      quantity: quantity ?? this.quantity,
      filledQuantity: filledQuantity ?? this.filledQuantity,
      averageFillPrice: averageFillPrice ?? this.averageFillPrice,
      status: status ?? this.status,
      stopPrice: stopPrice ?? this.stopPrice,
      icebergVisibleQty: icebergVisibleQty ?? this.icebergVisibleQty,
      trailingDeltaPercent: trailingDeltaPercent ?? this.trailingDeltaPercent,
      fee: fee ?? this.fee,
      feeCurrency: feeCurrency ?? this.feeCurrency,
      isBiometricVerified: isBiometricVerified ?? this.isBiometricVerified,
      biometricSignature: biometricSignature ?? this.biometricSignature,
      settlementMode: settlementMode ?? this.settlementMode,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      onChainTxHash: onChainTxHash ?? this.onChainTxHash,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'client_order_id': clientOrderId,
      'symbol': symbol,
      'side': side.name,
      'type': type.name,
      'time_in_force': timeInForce.code,
      'price': price,
      'quantity': quantity,
      'filled_quantity': filledQuantity,
      'average_fill_price': averageFillPrice,
      'status': status.name,
      'stop_price': stopPrice,
      'iceberg_visible_qty': icebergVisibleQty,
      'trailing_delta_percent': trailingDeltaPercent,
      'fee': fee,
      'fee_currency': feeCurrency,
      'is_biometric_verified': isBiometricVerified,
      'biometric_signature': biometricSignature,
      'settlement_mode': settlementMode.name,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'on_chain_tx_hash': onChainTxHash,
    };
  }

  factory Order.fromJson(Map<String, dynamic> json) {
    return Order(
      id: json['id'] as String? ?? '',
      clientOrderId: json['client_order_id'] as String? ?? '',
      symbol: json['symbol'] as String? ?? '',
      side: OrderSide.fromString(json['side'] as String? ?? 'buy'),
      type: OrderType.fromString(json['type'] as String? ?? 'limit'),
      timeInForce: TimeInForce.fromString(json['time_in_force'] as String? ?? 'GTC'),
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      quantity: (json['quantity'] as num?)?.toDouble() ?? 0.0,
      filledQuantity: (json['filled_quantity'] as num?)?.toDouble() ?? 0.0,
      averageFillPrice: (json['average_fill_price'] as num?)?.toDouble() ?? 0.0,
      status: OrderStatus.fromString(json['status'] as String? ?? 'new'),
      stopPrice: (json['stop_price'] as num?)?.toDouble(),
      icebergVisibleQty: (json['iceberg_visible_qty'] as num?)?.toDouble(),
      trailingDeltaPercent: (json['trailing_delta_percent'] as num?)?.toDouble(),
      fee: (json['fee'] as num?)?.toDouble() ?? 0.0,
      feeCurrency: json['fee_currency'] as String? ?? 'USDT',
      isBiometricVerified: json['is_biometric_verified'] as bool? ?? false,
      biometricSignature: json['biometric_signature'] as String?,
      settlementMode: SettlementMode.fromString(json['settlement_mode'] as String? ?? 'instantDvP'),
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : DateTime.now(),
      onChainTxHash: json['on_chain_tx_hash'] as String?,
    );
  }

  @override
  String toString() =>
      'Order(id: $id, symbol: $symbol, side: ${side.label}, type: ${type.displayName}, price: $price, qty: $quantity, status: ${status.label})';
}

/// Executed trade fill record.
class TradeExecution {
  final String id;
  final String orderId;
  final String clientOrderId;
  final String symbol;
  final OrderSide side;
  final double price;
  final double quantity;
  final double quoteAmount;
  final double fee; // 0.00 zero-fee presentation
  final String feeCurrency;
  final bool isMaker;
  final DateTime timestamp;
  final int? blockNumber;
  final String? txHash;

  const TradeExecution({
    required this.id,
    required this.orderId,
    required this.clientOrderId,
    required this.symbol,
    required this.side,
    required this.price,
    required this.quantity,
    required this.quoteAmount,
    this.fee = 0.0,
    this.feeCurrency = 'USDT',
    this.isMaker = false,
    required this.timestamp,
    this.blockNumber,
    this.txHash,
  });

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'order_id': orderId,
      'client_order_id': clientOrderId,
      'symbol': symbol,
      'side': side.name,
      'price': price,
      'quantity': quantity,
      'quote_amount': quoteAmount,
      'fee': fee,
      'fee_currency': feeCurrency,
      'is_maker': isMaker,
      'timestamp': timestamp.toIso8601String(),
      'block_number': blockNumber,
      'tx_hash': txHash,
    };
  }

  factory TradeExecution.fromJson(Map<String, dynamic> json) {
    return TradeExecution(
      id: json['id'] as String? ?? '',
      orderId: json['order_id'] as String? ?? '',
      clientOrderId: json['client_order_id'] as String? ?? '',
      symbol: json['symbol'] as String? ?? '',
      side: OrderSide.fromString(json['side'] as String? ?? 'buy'),
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      quantity: (json['quantity'] as num?)?.toDouble() ?? 0.0,
      quoteAmount: (json['quote_amount'] as num?)?.toDouble() ?? 0.0,
      fee: (json['fee'] as num?)?.toDouble() ?? 0.0,
      feeCurrency: json['fee_currency'] as String? ?? 'USDT',
      isMaker: json['is_maker'] as bool? ?? false,
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
      blockNumber: (json['block_number'] as num?)?.toInt(),
      txHash: json['tx_hash'] as String?,
    );
  }

  @override
  String toString() =>
      'TradeExecution(id: $id, symbol: $symbol, side: ${side.label}, price: $price, qty: $quantity)';
}

/// Request payload to submit an order to the matching engine.
class OrderPlacementRequest {
  final String clientOrderId;
  final String symbol;
  final OrderSide side;
  final OrderType type;
  final double? price;
  final double quantity;
  final TimeInForce timeInForce;
  final double? stopPrice;
  final double? icebergVisibleQty;
  final bool postOnly;
  final String? biometricSignature;

  const OrderPlacementRequest({
    required this.clientOrderId,
    required this.symbol,
    required this.side,
    required this.type,
    this.price,
    required this.quantity,
    this.timeInForce = TimeInForce.gtc,
    this.stopPrice,
    this.icebergVisibleQty,
    this.postOnly = false,
    this.biometricSignature,
  });

  /// Validates domain integrity of the order request.
  String? validate() {
    if (symbol.isEmpty) return 'Symbol cannot be empty';
    if (quantity <= 0) return 'Quantity must be greater than zero';
    if (type == OrderType.limit && (price == null || price! <= 0)) {
      return 'Price must be greater than zero for limit orders';
    }
    if ((type == OrderType.stopLimit || type == OrderType.stopMarket) &&
        (stopPrice == null || stopPrice! <= 0)) {
      return 'Stop price must be specified for trigger orders';
    }
    if (icebergVisibleQty != null && icebergVisibleQty! >= quantity) {
      return 'Iceberg visible quantity must be less than total quantity';
    }
    return null;
  }

  Map<String, dynamic> toJson() {
    return {
      'client_order_id': clientOrderId,
      'symbol': symbol,
      'side': side.name,
      'type': type.name,
      'price': price,
      'quantity': quantity,
      'time_in_force': timeInForce.code,
      'stop_price': stopPrice,
      'iceberg_visible_qty': icebergVisibleQty,
      'post_only': postOnly,
      'biometric_signature': biometricSignature,
    };
  }

  factory OrderPlacementRequest.fromJson(Map<String, dynamic> json) {
    return OrderPlacementRequest(
      clientOrderId: json['client_order_id'] as String? ?? '',
      symbol: json['symbol'] as String? ?? '',
      side: OrderSide.fromString(json['side'] as String? ?? 'buy'),
      type: OrderType.fromString(json['type'] as String? ?? 'limit'),
      price: (json['price'] as num?)?.toDouble(),
      quantity: (json['quantity'] as num?)?.toDouble() ?? 0.0,
      timeInForce: TimeInForce.fromString(json['time_in_force'] as String? ?? 'GTC'),
      stopPrice: (json['stop_price'] as num?)?.toDouble(),
      icebergVisibleQty: (json['iceberg_visible_qty'] as num?)?.toDouble(),
      postOnly: json['post_only'] as bool? ?? false,
      biometricSignature: json['biometric_signature'] as String?,
    );
  }
}
