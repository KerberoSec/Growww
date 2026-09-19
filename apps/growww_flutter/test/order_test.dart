import 'package:test/test.dart';
import '../lib/models/order.dart';

void main() {
  group('Order Model Tests', () {
    test('OrderSide parses strings accurately', () {
      expect(OrderSide.fromString('BUY'), equals(OrderSide.buy));
      expect(OrderSide.fromString('buy'), equals(OrderSide.buy));
      expect(OrderSide.fromString('sell'), equals(OrderSide.sell));
      expect(OrderSide.buy.label, equals('BUY'));
      expect(OrderSide.sell.label, equals('SELL'));
    });

    test('OrderType enum values and display names', () {
      expect(OrderType.fromString('limit'), equals(OrderType.limit));
      expect(OrderType.fromString('market'), equals(OrderType.market));
      expect(OrderType.fromString('stop_limit'), equals(OrderType.stopLimit));
      expect(OrderType.fromString('trailing-stop'), equals(OrderType.trailingStop));
      expect(OrderType.iceberg.displayName, equals('Iceberg'));
      expect(OrderType.oco.displayName, contains('OCO'));
    });

    test('TimeInForce parses codes correctly', () {
      expect(TimeInForce.fromString('GTC'), equals(TimeInForce.gtc));
      expect(TimeInForce.fromString('IOC'), equals(TimeInForce.ioc));
      expect(TimeInForce.fromString('FOK'), equals(TimeInForce.fok));
      expect(TimeInForce.fromString('POST_ONLY'), equals(TimeInForce.postOnly));
      expect(TimeInForce.gtc.code, equals('GTC'));
    });

    test('OrderStatus lifecycle checks and isDone / isCancellable getters', () {
      final openOrder = Order(
        id: 'ord_1',
        clientOrderId: 'cl_1',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.limit,
        timeInForce: TimeInForce.gtc,
        price: 64000.0,
        quantity: 2.0,
        filledQuantity: 0.5,
        status: OrderStatus.partiallyFilled,
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
      );

      expect(openOrder.remainingQuantity, equals(1.5));
      expect(openOrder.fillPercentage, equals(25.0));
      expect(openOrder.notionalValue, equals(128000.0));
      expect(openOrder.isDone, isFalse);
      expect(openOrder.isCancellable, isTrue);

      final filledOrder = openOrder.copyWith(
        status: OrderStatus.filled,
        filledQuantity: 2.0,
        averageFillPrice: 63980.0,
      );

      expect(filledOrder.remainingQuantity, equals(0.0));
      expect(filledOrder.fillPercentage, equals(100.0));
      expect(filledOrder.filledValue, equals(127960.0));
      expect(filledOrder.isDone, isTrue);
      expect(filledOrder.isCancellable, isFalse);
    });

    test('Order JSON serialization and deserialization roundtrip', () {
      final now = DateTime.utc(2026, 9, 20, 10, 0, 0);
      final original = Order(
        id: 'ord_99',
        clientOrderId: 'cl_99',
        symbol: 'ETH/USDT',
        side: OrderSide.sell,
        type: OrderType.stopLimit,
        timeInForce: TimeInForce.postOnly,
        price: 3450.0,
        quantity: 10.0,
        filledQuantity: 5.0,
        averageFillPrice: 3450.0,
        status: OrderStatus.partiallyFilled,
        stopPrice: 3460.0,
        fee: 0.0,
        feeCurrency: 'USDT',
        isBiometricVerified: true,
        biometricSignature: 'bio_sig_test_123',
        settlementMode: SettlementMode.besuOnChain,
        createdAt: now,
        updatedAt: now,
        onChainTxHash: '0xabc123',
      );

      final jsonMap = original.toJson();
      final reconstituted = Order.fromJson(jsonMap);

      expect(reconstituted.id, equals(original.id));
      expect(reconstituted.clientOrderId, equals(original.clientOrderId));
      expect(reconstituted.symbol, equals(original.symbol));
      expect(reconstituted.side, equals(OrderSide.sell));
      expect(reconstituted.type, equals(OrderType.stopLimit));
      expect(reconstituted.timeInForce, equals(TimeInForce.postOnly));
      expect(reconstituted.price, equals(3450.0));
      expect(reconstituted.quantity, equals(10.0));
      expect(reconstituted.stopPrice, equals(3460.0));
      expect(reconstituted.fee, equals(0.0)); // Zero-fee presentation invariant
      expect(reconstituted.isBiometricVerified, isTrue);
      expect(reconstituted.biometricSignature, equals('bio_sig_test_123'));
      expect(reconstituted.settlementMode, equals(SettlementMode.besuOnChain));
      expect(reconstituted.onChainTxHash, equals('0xabc123'));
    });

    test('OrderPlacementRequest validation logic', () {
      // Valid limit order
      const validReq = OrderPlacementRequest(
        clientOrderId: 'cl_req_1',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.limit,
        price: 64000.0,
        quantity: 1.0,
      );
      expect(validReq.validate(), isNull);

      // Empty symbol
      const emptySymbolReq = OrderPlacementRequest(
        clientOrderId: 'cl_req_2',
        symbol: '',
        side: OrderSide.buy,
        type: OrderType.limit,
        price: 64000.0,
        quantity: 1.0,
      );
      expect(emptySymbolReq.validate(), contains('Symbol cannot be empty'));

      // Zero quantity
      const zeroQtyReq = OrderPlacementRequest(
        clientOrderId: 'cl_req_3',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.limit,
        price: 64000.0,
        quantity: 0.0,
      );
      expect(zeroQtyReq.validate(), contains('Quantity must be greater than zero'));

      // Missing price on limit order
      const missingPriceReq = OrderPlacementRequest(
        clientOrderId: 'cl_req_4',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.limit,
        quantity: 1.0,
      );
      expect(missingPriceReq.validate(), contains('Price must be greater than zero'));

      // Missing stop price on stop-limit
      const missingStopReq = OrderPlacementRequest(
        clientOrderId: 'cl_req_5',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.stopLimit,
        price: 64000.0,
        quantity: 1.0,
      );
      expect(missingStopReq.validate(), contains('Stop price must be specified'));

      // Iceberg visible quantity violation
      const invalidIceberg = OrderPlacementRequest(
        clientOrderId: 'cl_req_6',
        symbol: 'BTC/USDT',
        side: OrderSide.buy,
        type: OrderType.limit,
        price: 64000.0,
        quantity: 1.0,
        icebergVisibleQty: 2.0,
      );
      expect(invalidIceberg.validate(), contains('Iceberg visible quantity must be less'));
    });

    test('TradeExecution roundtrip', () {
      final now = DateTime.utc(2026, 9, 20, 12, 0, 0);
      final trade = TradeExecution(
        id: 'tx_501',
        orderId: 'ord_501',
        clientOrderId: 'cl_501',
        symbol: 'SOL/USDT',
        side: OrderSide.buy,
        price: 152.50,
        quantity: 10.0,
        quoteAmount: 1525.0,
        fee: 0.0,
        isMaker: true,
        timestamp: now,
        blockNumber: 184520,
        txHash: '0xbesu_trade_999',
      );

      final json = trade.toJson();
      final roundtrip = TradeExecution.fromJson(json);

      expect(roundtrip.id, equals('tx_501'));
      expect(roundtrip.price, equals(152.50));
      expect(roundtrip.quantity, equals(10.0));
      expect(roundtrip.quoteAmount, equals(1525.0));
      expect(roundtrip.fee, equals(0.0));
      expect(roundtrip.isMaker, isTrue);
      expect(roundtrip.blockNumber, equals(184520));
      expect(roundtrip.txHash, equals('0xbesu_trade_999'));
    });
  });
}
