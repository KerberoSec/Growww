import 'dart:math' as math;

/// Single price-volume rung in the Level 2 orderbook ladder.
class DepthEntry {
  final double price;
  final double quantity;
  final int orderCount;
  final double cumulativeQuantity;

  const DepthEntry({
    required this.price,
    required this.quantity,
    this.orderCount = 1,
    this.cumulativeQuantity = 0.0,
  });

  DepthEntry copyWith({
    double? price,
    double? quantity,
    int? orderCount,
    double? cumulativeQuantity,
  }) {
    return DepthEntry(
      price: price ?? this.price,
      quantity: quantity ?? this.quantity,
      orderCount: orderCount ?? this.orderCount,
      cumulativeQuantity: cumulativeQuantity ?? this.cumulativeQuantity,
    );
  }

  Map<String, dynamic> toJson() => {
        'price': price,
        'quantity': quantity,
        'order_count': orderCount,
        'cumulative_quantity': cumulativeQuantity,
      };

  factory DepthEntry.fromJson(dynamic json) {
    if (json is List) {
      final p = (json[0] as num).toDouble();
      final q = (json[1] as num).toDouble();
      final count = json.length > 2 ? (json[2] as num).toInt() : 1;
      return DepthEntry(price: p, quantity: q, orderCount: count);
    }
    final map = json as Map<String, dynamic>;
    return DepthEntry(
      price: (map['price'] as num?)?.toDouble() ?? 0.0,
      quantity: (map['quantity'] as num?)?.toDouble() ?? 0.0,
      orderCount: (map['order_count'] as num?)?.toInt() ?? 1,
      cumulativeQuantity: (map['cumulative_quantity'] as num?)?.toDouble() ?? 0.0,
    );
  }

  @override
  String toString() => 'DepthEntry(price: $price, qty: $quantity, cum: $cumulativeQuantity)';
}

/// Full Level 2 Orderbook snapshot with bids (descending) and asks (ascending).
class OrderBookDepth {
  final String symbol;
  final int sequence;
  final DateTime timestamp;
  final List<DepthEntry> bids;
  final List<DepthEntry> asks;

  const OrderBookDepth({
    required this.symbol,
    required this.sequence,
    required this.timestamp,
    required this.bids,
    required this.asks,
  });

  DepthEntry? get bestBid => bids.isNotEmpty ? bids.first : null;
  DepthEntry? get bestAsk => asks.isNotEmpty ? asks.first : null;

  double get spread {
    if (bestAsk != null && bestBid != null) {
      return math.max(0.0, bestAsk!.price - bestBid!.price);
    }
    return 0.0;
  }

  double get spreadPercentage {
    if (bestBid != null && bestBid!.price > 0 && bestAsk != null) {
      return (spread / bestBid!.price) * 100.0;
    }
    return 0.0;
  }

  double get midPrice {
    if (bestAsk != null && bestBid != null) {
      return (bestAsk!.price + bestBid!.price) / 2.0;
    }
    return bestBid?.price ?? bestAsk?.price ?? 0.0;
  }

  /// Calculates cumulative quantities along both ladders.
  static OrderBookDepth withCumulatives({
    required String symbol,
    required int sequence,
    required DateTime timestamp,
    required List<DepthEntry> rawBids,
    required List<DepthEntry> rawAsks,
  }) {
    final sortedBids = List<DepthEntry>.from(rawBids)
      ..sort((a, b) => b.price.compareTo(a.price));
    final sortedAsks = List<DepthEntry>.from(rawAsks)
      ..sort((a, b) => a.price.compareTo(b.price));

    double runningBidTotal = 0.0;
    final cumBids = <DepthEntry>[];
    for (final bid in sortedBids) {
      runningBidTotal += bid.quantity;
      cumBids.add(bid.copyWith(cumulativeQuantity: runningBidTotal));
    }

    double runningAskTotal = 0.0;
    final cumAsks = <DepthEntry>[];
    for (final ask in sortedAsks) {
      runningAskTotal += ask.quantity;
      cumAsks.add(ask.copyWith(cumulativeQuantity: runningAskTotal));
    }

    return OrderBookDepth(
      symbol: symbol,
      sequence: sequence,
      timestamp: timestamp,
      bids: cumBids,
      asks: cumAsks,
    );
  }

  /// Applies an incremental L2 diff snapshot and returns a re-computed depth book.
  OrderBookDepth applyDiff(OrderBookDiff diff) {
    final bidMap = <double, DepthEntry>{for (var b in bids) b.price: b};
    final askMap = <double, DepthEntry>{for (var a in asks) a.price: a};

    for (final b in diff.bids) {
      if (b.quantity <= 0) {
        bidMap.remove(b.price);
      } else {
        bidMap[b.price] = b;
      }
    }

    for (final a in diff.asks) {
      if (a.quantity <= 0) {
        askMap.remove(a.price);
      } else {
        askMap[a.price] = a;
      }
    }

    return OrderBookDepth.withCumulatives(
      symbol: symbol,
      sequence: diff.lastSequence,
      timestamp: diff.timestamp,
      rawBids: bidMap.values.toList(),
      rawAsks: askMap.values.toList(),
    );
  }

  Map<String, dynamic> toJson() => {
        'symbol': symbol,
        'sequence': sequence,
        'timestamp': timestamp.toIso8601String(),
        'bids': bids.map((b) => b.toJson()).toList(),
        'asks': asks.map((a) => a.toJson()).toList(),
      };

  factory OrderBookDepth.fromJson(Map<String, dynamic> json) {
    final rawBids = (json['bids'] as List? ?? [])
        .map((e) => DepthEntry.fromJson(e))
        .toList();
    final rawAsks = (json['asks'] as List? ?? [])
        .map((e) => DepthEntry.fromJson(e))
        .toList();

    return OrderBookDepth.withCumulatives(
      symbol: json['symbol'] as String? ?? '',
      sequence: (json['sequence'] as num?)?.toInt() ?? 0,
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
      rawBids: rawBids,
      rawAsks: rawAsks,
    );
  }
}

/// Incremental delta update for Level 2 orderbook streaming.
class OrderBookDiff {
  final String symbol;
  final int firstSequence;
  final int lastSequence;
  final List<DepthEntry> bids;
  final List<DepthEntry> asks;
  final DateTime timestamp;

  const OrderBookDiff({
    required this.symbol,
    required this.firstSequence,
    required this.lastSequence,
    required this.bids,
    required this.asks,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
        'symbol': symbol,
        'first_sequence': firstSequence,
        'last_sequence': lastSequence,
        'bids': bids.map((b) => b.toJson()).toList(),
        'asks': asks.map((a) => a.toJson()).toList(),
        'timestamp': timestamp.toIso8601String(),
      };

  factory OrderBookDiff.fromJson(Map<String, dynamic> json) {
    return OrderBookDiff(
      symbol: json['symbol'] as String? ?? '',
      firstSequence: (json['first_sequence'] as num?)?.toInt() ?? 0,
      lastSequence: (json['last_sequence'] as num?)?.toInt() ?? 0,
      bids: (json['bids'] as List? ?? [])
          .map((e) => DepthEntry.fromJson(e))
          .toList(),
      asks: (json['asks'] as List? ?? [])
          .map((e) => DepthEntry.fromJson(e))
          .toList(),
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
    );
  }
}

/// Real-time 24h market ticker statistics.
class Ticker {
  final String symbol;
  final double lastPrice;
  final double priceChange24h;
  final double priceChangePercent24h;
  final double highPrice24h;
  final double lowPrice24h;
  final double baseVolume24h;
  final double quoteVolume24h;
  final double openPrice24h;
  final double bidPrice;
  final double askPrice;
  final DateTime timestamp;

  const Ticker({
    required this.symbol,
    required this.lastPrice,
    required this.priceChange24h,
    required this.priceChangePercent24h,
    required this.highPrice24h,
    required this.lowPrice24h,
    required this.baseVolume24h,
    required this.quoteVolume24h,
    required this.openPrice24h,
    required this.bidPrice,
    required this.askPrice,
    required this.timestamp,
  });

  bool get isBullish => priceChangePercent24h >= 0.0;

  Map<String, dynamic> toJson() => {
        'symbol': symbol,
        'last_price': lastPrice,
        'price_change_24h': priceChange24h,
        'price_change_percent_24h': priceChangePercent24h,
        'high_price_24h': highPrice24h,
        'low_price_24h': lowPrice24h,
        'base_volume_24h': baseVolume24h,
        'quote_volume_24h': quoteVolume24h,
        'open_price_24h': openPrice24h,
        'bid_price': bidPrice,
        'ask_price': askPrice,
        'timestamp': timestamp.toIso8601String(),
      };

  factory Ticker.fromJson(Map<String, dynamic> json) {
    return Ticker(
      symbol: json['symbol'] as String? ?? '',
      lastPrice: (json['last_price'] as num?)?.toDouble() ?? 0.0,
      priceChange24h: (json['price_change_24h'] as num?)?.toDouble() ?? 0.0,
      priceChangePercent24h:
          (json['price_change_percent_24h'] as num?)?.toDouble() ?? 0.0,
      highPrice24h: (json['high_price_24h'] as num?)?.toDouble() ?? 0.0,
      lowPrice24h: (json['low_price_24h'] as num?)?.toDouble() ?? 0.0,
      baseVolume24h: (json['base_volume_24h'] as num?)?.toDouble() ?? 0.0,
      quoteVolume24h: (json['quote_volume_24h'] as num?)?.toDouble() ?? 0.0,
      openPrice24h: (json['open_price_24h'] as num?)?.toDouble() ?? 0.0,
      bidPrice: (json['bid_price'] as num?)?.toDouble() ?? 0.0,
      askPrice: (json['ask_price'] as num?)?.toDouble() ?? 0.0,
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
    );
  }
}

/// Candlestick chart interval.
enum CandleInterval {
  m1('1m'),
  m5('5m'),
  m15('15m'),
  h1('1h'),
  h4('4h'),
  d1('1D');

  final String label;
  const CandleInterval(this.label);
}

/// Candlestick OHLCV data point.
class CandlestickData {
  final DateTime timestamp;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;

  const CandlestickData({
    required this.timestamp,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
  });

  bool get isBullish => close >= open;

  Map<String, dynamic> toJson() => {
        'timestamp': timestamp.toIso8601String(),
        'open': open,
        'high': high,
        'low': low,
        'close': close,
        'volume': volume,
      };

  factory CandlestickData.fromJson(dynamic json) {
    if (json is List) {
      return CandlestickData(
        timestamp: DateTime.fromMillisecondsSinceEpoch((json[0] as num).toInt()),
        open: (json[1] as num).toDouble(),
        high: (json[2] as num).toDouble(),
        low: (json[3] as num).toDouble(),
        close: (json[4] as num).toDouble(),
        volume: (json[5] as num).toDouble(),
      );
    }
    final map = json as Map<String, dynamic>;
    return CandlestickData(
      timestamp: map['timestamp'] != null
          ? DateTime.parse(map['timestamp'] as String)
          : DateTime.now(),
      open: (map['open'] as num?)?.toDouble() ?? 0.0,
      high: (map['high'] as num?)?.toDouble() ?? 0.0,
      low: (map['low'] as num?)?.toDouble() ?? 0.0,
      close: (map['close'] as num?)?.toDouble() ?? 0.0,
      volume: (map['volume'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

/// Classification of financial asset classes on the platform.
enum AssetClass {
  equityInr,
  cryptoVda,
  fiatInr,
  stablecoinUsdt;

  String get displayName {
    switch (this) {
      case AssetClass.equityInr:
        return 'NSE/BSE Equities';
      case AssetClass.cryptoVda:
        return 'Crypto VDA';
      case AssetClass.fiatInr:
        return 'INR Cash Balance';
      case AssetClass.stablecoinUsdt:
        return 'USDT Stablecoin';
    }
  }
}

/// Portfolio holding item combining traditional Demat and crypto VDA assets.
class PortfolioAsset {
  final String symbol;
  final String assetName;
  final AssetClass assetClass;
  final double balance;
  final double lockedBalance;
  final double averageCostBasis;
  final double currentPrice; // INR value
  final double tdsAccrued; // Section 194S 1% TDS on VDA
  final String? onChainDvPTokenAddress;

  const PortfolioAsset({
    required this.symbol,
    required this.assetName,
    required this.assetClass,
    required this.balance,
    this.lockedBalance = 0.0,
    required this.averageCostBasis,
    required this.currentPrice,
    this.tdsAccrued = 0.0,
    this.onChainDvPTokenAddress,
  });

  double get totalBalance => balance + lockedBalance;
  double get investedValue => totalBalance * averageCostBasis;
  double get currentValueInr => totalBalance * currentPrice;
  double get unrealizedPnL => currentValueInr - investedValue;
  double get unrealizedPnLPercent =>
      investedValue > 0 ? (unrealizedPnL / investedValue) * 100.0 : 0.0;

  /// Section 115BBH 30% flat tax estimation on net positive VDA gains.
  double get estimatedVdaGainTax {
    if (assetClass == AssetClass.cryptoVda && unrealizedPnL > 0) {
      return unrealizedPnL * 0.30;
    }
    return 0.0;
  }

  Map<String, dynamic> toJson() => {
        'symbol': symbol,
        'asset_name': assetName,
        'asset_class': assetClass.name,
        'balance': balance,
        'locked_balance': lockedBalance,
        'average_cost_basis': averageCostBasis,
        'current_price': currentPrice,
        'tds_accrued': tdsAccrued,
        'on_chain_dvp_token_address': onChainDvPTokenAddress,
      };

  factory PortfolioAsset.fromJson(Map<String, dynamic> json) {
    return PortfolioAsset(
      symbol: json['symbol'] as String? ?? '',
      assetName: json['asset_name'] as String? ?? '',
      assetClass: AssetClass.values.firstWhere(
        (e) => e.name == json['asset_class'],
        orElse: () => AssetClass.cryptoVda,
      ),
      balance: (json['balance'] as num?)?.toDouble() ?? 0.0,
      lockedBalance: (json['locked_balance'] as num?)?.toDouble() ?? 0.0,
      averageCostBasis: (json['average_cost_basis'] as num?)?.toDouble() ?? 0.0,
      currentPrice: (json['current_price'] as num?)?.toDouble() ?? 0.0,
      tdsAccrued: (json['tds_accrued'] as num?)?.toDouble() ?? 0.0,
      onChainDvPTokenAddress: json['on_chain_dvp_token_address'] as String?,
    );
  }
}

/// Consolidated portfolio summary combining INR Demat equities and crypto VDAs.
class PortfolioSummary {
  final double unallocatedCashInr;
  final List<PortfolioAsset> assets;
  final DateTime lastUpdated;
  final String proofOfReserveRootHash;

  const PortfolioSummary({
    required this.unallocatedCashInr,
    required this.assets,
    required this.lastUpdated,
    this.proofOfReserveRootHash = '0x9a8f2e7b1c4d3e5a6f8b0c2d4e6f8a0b1c3d5e7f',
  });

  double get totalEquityInr => assets
      .where((a) => a.assetClass == AssetClass.equityInr)
      .fold(0.0, (sum, a) => sum + a.currentValueInr);

  double get totalCryptoVdaInr => assets
      .where((a) => a.assetClass == AssetClass.cryptoVda || a.assetClass == AssetClass.stablecoinUsdt)
      .fold(0.0, (sum, a) => sum + a.currentValueInr);

  double get totalNetWorthInr => totalEquityInr + totalCryptoVdaInr + unallocatedCashInr;

  double get totalInvestedInr => assets.fold(0.0, (sum, a) => sum + a.investedValue);

  double get totalUnrealizedPnLInr => assets.fold(0.0, (sum, a) => sum + a.unrealizedPnL);

  double get totalUnrealizedPnLPercent =>
      totalInvestedInr > 0 ? (totalUnrealizedPnLInr / totalInvestedInr) * 100.0 : 0.0;

  double get totalTdsAccruedInr => assets.fold(0.0, (sum, a) => sum + a.tdsAccrued);

  double get totalSection115bbhTaxEstimate =>
      assets.fold(0.0, (sum, a) => sum + a.estimatedVdaGainTax);

  double get equityAllocationPercent =>
      totalNetWorthInr > 0 ? (totalEquityInr / totalNetWorthInr) * 100.0 : 0.0;

  double get cryptoAllocationPercent =>
      totalNetWorthInr > 0 ? (totalCryptoVdaInr / totalNetWorthInr) * 100.0 : 0.0;

  double get cashAllocationPercent =>
      totalNetWorthInr > 0 ? (unallocatedCashInr / totalNetWorthInr) * 100.0 : 0.0;

  Map<String, dynamic> toJson() => {
        'unallocated_cash_inr': unallocatedCashInr,
        'assets': assets.map((a) => a.toJson()).toList(),
        'last_updated': lastUpdated.toIso8601String(),
        'proof_of_reserve_root_hash': proofOfReserveRootHash,
      };

  factory PortfolioSummary.fromJson(Map<String, dynamic> json) {
    return PortfolioSummary(
      unallocatedCashInr: (json['unallocated_cash_inr'] as num?)?.toDouble() ?? 0.0,
      assets: (json['assets'] as List? ?? [])
          .map((e) => PortfolioAsset.fromJson(e as Map<String, dynamic>))
          .toList(),
      lastUpdated: json['last_updated'] != null
          ? DateTime.parse(json['last_updated'] as String)
          : DateTime.now(),
      proofOfReserveRootHash: json['proof_of_reserve_root_hash'] as String? ?? '',
    );
  }
}

/// Thrown when a WebSocket sequence gap is identified during live orderbook streaming.
class SequenceGapException implements Exception {
  final String symbol;
  final int expectedSequence;
  final int receivedSequence;
  final String message;

  SequenceGapException({
    required this.symbol,
    required this.expectedSequence,
    required this.receivedSequence,
  }) : message =
            'WebSocket Sequence Gap detected for $symbol: expected $expectedSequence, received $receivedSequence. Initiating REST snapshot resync.';

  @override
  String toString() => 'SequenceGapException: $message';
}
