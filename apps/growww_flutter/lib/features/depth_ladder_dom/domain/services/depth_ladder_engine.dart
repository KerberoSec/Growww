import '../models/depth_ladder_enums.dart';
import '../models/l3_order_entry.dart';
import '../models/ladder_depth_level.dart';
import '../models/ladder_dom_snapshot.dart';

/// Calculation, aggregation, and market depth engine for DOM Depth Ladder.
class DepthLadderEngine {
  /// Rounds price according to chosen tick aggregation step.
  static double roundToTick(double price, LadderTickAggregation aggregation) {
    final step = aggregation.step;
    final factor = (price / step).round();
    final res = factor * step;
    return double.parse(res.toStringAsFixed(4));
  }

  /// Calculates spread in currency units and in basis points (bps).
  static ({double spread, double spreadBps}) computeSpread({
    required double bestBid,
    required double bestAsk,
  }) {
    if (bestBid <= 0 || bestAsk <= 0 || bestAsk < bestBid) {
      return (spread: 0.0, spreadBps: 0.0);
    }
    final spread = bestAsk - bestBid;
    final mid = (bestBid + bestAsk) / 2.0;
    final bps = (spread / mid) * 10000.0;
    return (spread: spread, spreadBps: bps);
  }

  /// Aggregates individual L3 orders into consolidated L2 ladder levels.
  static List<LadderDepthLevel> aggregateL3Orders({
    required List<L3OrderEntry> orders,
    required DomOrderSide side,
    required LadderTickAggregation aggregation,
  }) {
    final filtered = orders.where((o) => o.side == side).toList();
    if (filtered.isEmpty) return const [];

    final Map<double, ({double qty, int count})> priceMap = {};
    for (final o in filtered) {
      final aggPrice = roundToTick(o.price, aggregation);
      final existing = priceMap[aggPrice];
      if (existing == null) {
        priceMap[aggPrice] = (qty: o.quantity, count: 1);
      } else {
        priceMap[aggPrice] = (qty: existing.qty + o.quantity, count: existing.count + 1);
      }
    }

    final sortedPrices = priceMap.keys.toList();
    if (side.isBid) {
      sortedPrices.sort((a, b) => b.compareTo(a)); // Bids desc
    } else {
      sortedPrices.sort((a, b) => a.compareTo(b)); // Asks asc
    }

    double runningTotal = 0.0;
    final levels = <LadderDepthLevel>[];

    for (int i = 0; i < sortedPrices.length; i++) {
      final p = sortedPrices[i];
      final data = priceMap[p]!;
      runningTotal += data.qty;

      levels.add(LadderDepthLevel(
        price: p,
        quantity: data.qty,
        orderCount: data.count,
        cumulativeQuantity: runningTotal,
        side: side,
        isBestPrice: (i == 0),
      ));
    }

    return levels;
  }

  /// Computes cumulative volume and normalizes visual depth bar percentages (0.0 to 1.0).
  static List<LadderDepthLevel> computeCumulativeAndPercentages({
    required List<LadderDepthLevel> levels,
    required double maxDepth,
  }) {
    if (levels.isEmpty) return const [];
    final effectiveMax = maxDepth > 0 ? maxDepth : 1.0;

    double cumulative = 0.0;
    final result = <LadderDepthLevel>[];

    for (int i = 0; i < levels.length; i++) {
      final lvl = levels[i];
      cumulative += lvl.quantity;
      final pct = (cumulative / effectiveMax).clamp(0.04, 1.0);

      result.add(lvl.copyWith(
        cumulativeQuantity: cumulative,
        depthPercent: pct,
        isBestPrice: (i == 0),
      ));
    }

    return result;
  }

  /// Generates a realistic live Depth Ladder DOM snapshot.
  static LadderDomSnapshot generateSnapshot({
    required String symbol,
    required double centerPrice,
    LadderTickAggregation aggregation = LadderTickAggregation.one,
    DepthMode depthMode = DepthMode.level2Aggregated,
    int levelsPerSide = 10,
  }) {
    final step = aggregation.step;
    final roundedCenter = roundToTick(centerPrice, aggregation);
    final bestBid = roundedCenter - step;
    final bestAsk = roundedCenter + step;

    final rawBids = <LadderDepthLevel>[];
    final rawAsks = <LadderDepthLevel>[];
    final l3List = <L3OrderEntry>[];

    // Generate bids
    for (int i = 0; i < levelsPerSide; i++) {
      final p = double.parse((bestBid - (i * step)).toStringAsFixed(4));
      final qty = 1.2 + (i * 0.45);
      final count = 2 + (i % 4);

      rawBids.add(LadderDepthLevel(
        price: p,
        quantity: qty,
        orderCount: count,
        side: DomOrderSide.bid,
      ));

      l3List.add(L3OrderEntry(
        orderId: 'ORD-B-$i',
        traderMaskedId: '0x94...b${i}a',
        price: p,
        quantity: qty,
        side: DomOrderSide.bid,
        queuePriority: i + 1,
        timestamp: DateTime.now(),
      ));
    }

    // Generate asks
    for (int i = 0; i < levelsPerSide; i++) {
      final p = double.parse((bestAsk + (i * step)).toStringAsFixed(4));
      final qty = 1.1 + (i * 0.5);
      final count = 1 + (i % 5);

      rawAsks.add(LadderDepthLevel(
        price: p,
        quantity: qty,
        orderCount: count,
        side: DomOrderSide.ask,
      ));

      l3List.add(L3OrderEntry(
        orderId: 'ORD-A-$i',
        traderMaskedId: '0x32...e${i}f',
        price: p,
        quantity: qty,
        side: DomOrderSide.ask,
        queuePriority: i + 1,
        timestamp: DateTime.now(),
      ));
    }

    // Determine max cumulative volume for normalization
    final totalBidVol = rawBids.fold<double>(0.0, (sum, b) => sum + b.quantity);
    final totalAskVol = rawAsks.fold<double>(0.0, (sum, a) => sum + a.quantity);
    final maxVol = totalBidVol > totalAskVol ? totalBidVol : totalAskVol;

    final normalizedBids = computeCumulativeAndPercentages(levels: rawBids, maxDepth: maxVol);
    final normalizedAsks = computeCumulativeAndPercentages(levels: rawAsks, maxDepth: maxVol);

    final spreadMetrics = computeSpread(bestBid: bestBid, bestAsk: bestAsk);

    return LadderDomSnapshot(
      symbol: symbol,
      lastTradedPrice: centerPrice,
      priceChangePercent24h: 2.14,
      bestBidPrice: bestBid,
      bestAskPrice: bestAsk,
      spread: spreadMetrics.spread,
      spreadBps: spreadMetrics.spreadBps,
      bidLevels: normalizedBids,
      askLevels: normalizedAsks,
      l3Orders: l3List,
      maxCumulativeVolume: maxVol,
      aggregation: aggregation,
      depthMode: depthMode,
    );
  }
}
