import 'depth_ladder_enums.dart';
import 'l3_order_entry.dart';
import 'ladder_depth_level.dart';

/// Full Depth of Market (DOM) state snapshot.
class LadderDomSnapshot {
  final String symbol; // e.g. BTC/USDT or ETH/INR
  final double lastTradedPrice;
  final double priceChangePercent24h;
  final double bestBidPrice;
  final double bestAskPrice;
  final double spread;
  final double spreadBps; // Basis points: (spread / mid) * 10,000
  final List<LadderDepthLevel> bidLevels; // Sorted descending by price
  final List<LadderDepthLevel> askLevels; // Sorted ascending by price
  final List<L3OrderEntry> l3Orders;
  final double maxCumulativeVolume;
  final LadderTickAggregation aggregation;
  final DepthMode depthMode;

  const LadderDomSnapshot({
    required this.symbol,
    required this.lastTradedPrice,
    this.priceChangePercent24h = 0.0,
    required this.bestBidPrice,
    required this.bestAskPrice,
    required this.spread,
    required this.spreadBps,
    required this.bidLevels,
    required this.askLevels,
    this.l3Orders = const [],
    required this.maxCumulativeVolume,
    this.aggregation = LadderTickAggregation.one,
    this.depthMode = DepthMode.level2Aggregated,
  });

  double get midPrice => (bestBidPrice + bestAskPrice) / 2.0;

  LadderDomSnapshot copyWith({
    String? symbol,
    double? lastTradedPrice,
    double? priceChangePercent24h,
    double? bestBidPrice,
    double? bestAskPrice,
    double? spread,
    double? spreadBps,
    List<LadderDepthLevel>? bidLevels,
    List<LadderDepthLevel>? askLevels,
    List<L3OrderEntry>? l3Orders,
    double? maxCumulativeVolume,
    LadderTickAggregation? aggregation,
    DepthMode? depthMode,
  }) {
    return LadderDomSnapshot(
      symbol: symbol ?? this.symbol,
      lastTradedPrice: lastTradedPrice ?? this.lastTradedPrice,
      priceChangePercent24h: priceChangePercent24h ?? this.priceChangePercent24h,
      bestBidPrice: bestBidPrice ?? this.bestBidPrice,
      bestAskPrice: bestAskPrice ?? this.bestAskPrice,
      spread: spread ?? this.spread,
      spreadBps: spreadBps ?? this.spreadBps,
      bidLevels: bidLevels ?? this.bidLevels,
      askLevels: askLevels ?? this.askLevels,
      l3Orders: l3Orders ?? this.l3Orders,
      maxCumulativeVolume: maxCumulativeVolume ?? this.maxCumulativeVolume,
      aggregation: aggregation ?? this.aggregation,
      depthMode: depthMode ?? this.depthMode,
    );
  }
}
