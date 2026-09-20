import 'package:test/test.dart';
import 'package:growww_flutter/features/depth_ladder_dom/domain/models/depth_ladder_enums.dart';
import 'package:growww_flutter/features/depth_ladder_dom/domain/models/l3_order_entry.dart';
import 'package:growww_flutter/features/depth_ladder_dom/domain/services/depth_ladder_engine.dart';
import 'package:growww_flutter/features/depth_ladder_dom/presentation/controllers/depth_ladder_controller.dart';

void main() {
  group('Prompt 544 - Live L2/L3 Depth Ladder DOM & Orderbook Screen', () {
    late DepthLadderController controller;

    setUp(() {
      controller = DepthLadderController(
        symbol: 'BTC/USDT',
        centerPrice: 64250.0,
        aggregation: LadderTickAggregation.one,
      );
    });

    tearDown(() {
      controller.dispose();
    });

    test('DepthLadderEngine rounds prices to tick aggregations and computes spread bps', () {
      expect(DepthLadderEngine.roundToTick(64250.23, LadderTickAggregation.point1), equals(64250.2));
      expect(DepthLadderEngine.roundToTick(64250.28, LadderTickAggregation.point05), equals(64250.30));
      expect(DepthLadderEngine.roundToTick(64253.40, LadderTickAggregation.five), equals(64255.0));

      final spreadResult = DepthLadderEngine.computeSpread(
        bestBid: 64249.0,
        bestAsk: 64251.0,
      );
      expect(spreadResult.spread, equals(2.0));
      // Mid price = 64250.0. bps = (2 / 64250) * 10000 ≈ 0.31128
      expect(spreadResult.spreadBps, closeTo(0.311, 0.01));
    });

    test('DepthLadderEngine aggregates individual L3 orders into consolidated ladder levels', () {
      final now = DateTime.now();
      final l3Orders = [
        L3OrderEntry(
          orderId: 'O1',
          traderMaskedId: 'T1',
          price: 64249.0,
          quantity: 2.0,
          side: DomOrderSide.bid,
          queuePriority: 1,
          timestamp: now,
        ),
        L3OrderEntry(
          orderId: 'O2',
          traderMaskedId: 'T2',
          price: 64249.0,
          quantity: 1.5,
          side: DomOrderSide.bid,
          queuePriority: 2,
          timestamp: now,
        ),
        L3OrderEntry(
          orderId: 'O3',
          traderMaskedId: 'T3',
          price: 64248.0,
          quantity: 4.0,
          side: DomOrderSide.bid,
          queuePriority: 1,
          timestamp: now,
        ),
      ];

      final bidLevels = DepthLadderEngine.aggregateL3Orders(
        orders: l3Orders,
        side: DomOrderSide.bid,
        aggregation: LadderTickAggregation.one,
      );

      expect(bidLevels.length, equals(2));
      // Highest bid first
      expect(bidLevels[0].price, equals(64249.0));
      expect(bidLevels[0].quantity, equals(3.5)); // 2.0 + 1.5
      expect(bidLevels[0].orderCount, equals(2));
      expect(bidLevels[0].isBestPrice, isTrue);

      expect(bidLevels[1].price, equals(64248.0));
      expect(bidLevels[1].quantity, equals(4.0));
      expect(bidLevels[1].cumulativeQuantity, equals(7.5)); // 3.5 + 4.0
    });

    test('DepthLadderEngine normalizes visual depth bar percentages properly', () {
      final snapshot = DepthLadderEngine.generateSnapshot(
        symbol: 'BTC/USDT',
        centerPrice: 64250.0,
        aggregation: LadderTickAggregation.one,
        levelsPerSide: 5,
      );

      expect(snapshot.bidLevels.length, equals(5));
      expect(snapshot.askLevels.length, equals(5));
      expect(snapshot.bidLevels.first.isBestPrice, isTrue);
      expect(snapshot.askLevels.first.isBestPrice, isTrue);

      for (final lvl in snapshot.bidLevels) {
        expect(lvl.depthPercent, greaterThanOrEqualTo(0.04));
        expect(lvl.depthPercent, lessThanOrEqualTo(1.0));
      }
    });

    test('DepthLadderController manages tick aggregation switching and depth mode', () {
      expect(controller.state.aggregation, equals(LadderTickAggregation.one));

      controller.setAggregation(LadderTickAggregation.five);
      expect(controller.state.aggregation, equals(LadderTickAggregation.five));
      expect(controller.state.toastMessage, contains('Tick size set to 5.00'));

      controller.setDepthMode(DepthMode.level3MarketByOrder);
      expect(controller.state.depthMode, equals(DepthMode.level3MarketByOrder));
    });

    test('DepthLadderController handles limit order staging, quantity adjustment, and execution', () async {
      expect(controller.state.isOrderStaged, isFalse);

      // Select price level
      controller.selectPriceLevel(64245.0, DomOrderSide.bid);
      expect(controller.state.isOrderStaged, isTrue);
      expect(controller.state.selectedPrice, equals(64245.0));
      expect(controller.state.stagedSide, equals(DomOrderSide.bid));

      // Adjust quantity
      controller.setStagedQuantity(2.5);
      expect(controller.state.stagedQuantity, equals(2.5));
      expect(controller.state.estimatedNotional, equals(64245.0 * 2.5));

      // Execute staged order
      final ok = await controller.executeStagedOrder();
      expect(ok, isTrue);
      expect(controller.state.isOrderStaged, isFalse);
      expect(controller.state.selectedPrice, isNull);
      expect(controller.state.toastMessage, contains('Limit BUY placed: 2.5 BTC @ ₹64245.00'));

      // Quick order test
      final quickOk = await controller.quickOrder(DomOrderSide.ask, 64260.0, 1.0);
      expect(quickOk, isTrue);
      expect(controller.state.toastMessage, contains('Quick SELL executed'));
    });
  });
}
