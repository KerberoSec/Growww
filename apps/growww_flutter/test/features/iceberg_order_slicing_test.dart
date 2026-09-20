import 'package:test/test.dart';
import 'package:growww_flutter/features/iceberg_orders/domain/algorithms/iceberg_slicing_allocator.dart';
import 'package:growww_flutter/features/iceberg_orders/domain/models/iceberg_enums.dart';
import 'package:growww_flutter/features/iceberg_orders/domain/models/iceberg_order_config.dart';
import 'package:growww_flutter/features/iceberg_orders/presentation/controllers/iceberg_order_controller.dart';

void main() {
  group('Prompt 552 - Iceberg Order Slicing and Hidden Size Allocator', () {
    late IcebergOrderController controller;

    setUp(() {
      controller = IcebergOrderController();
    });

    tearDown(() {
      controller.dispose();
    });

    test('IcebergSlicingAllocator validates configuration constraints', () {
      final invalid1 = IcebergOrderConfig(
        orderId: 'IC-1',
        symbol: 'BTC/USDT',
        side: 'BUY',
        totalQuantity: 10.0,
        visibleTrancheQuantity: 12.0, // Exceeds total
        varianceMode: IcebergVarianceMode.fixedTranche,
        priceLimit: 64000.0,
        createdAt: DateTime.now(),
      );
      final res1 = IcebergSlicingAllocator.validateConfig(invalid1);
      expect(res1.isValid, isFalse);
      expect(res1.errorMessage, contains('cannot exceed total order quantity'));

      final invalid2 = IcebergOrderConfig(
        orderId: 'IC-2',
        symbol: 'BTC/USDT',
        side: 'BUY',
        totalQuantity: 10.0,
        visibleTrancheQuantity: 8.0, // Less than 1.5x
        varianceMode: IcebergVarianceMode.fixedTranche,
        priceLimit: 64000.0,
        createdAt: DateTime.now(),
      );
      final res2 = IcebergSlicingAllocator.validateConfig(invalid2);
      expect(res2.isValid, isFalse);
      expect(res2.errorMessage, contains('at least 1.5x'));
    });

    test('IcebergSlicingAllocator generates fixed tranches with exact sum invariant', () {
      final config = IcebergOrderConfig(
        orderId: 'IC-3',
        symbol: 'BTC/USDT',
        side: 'BUY',
        totalQuantity: 10.0,
        visibleTrancheQuantity: 2.0,
        varianceMode: IcebergVarianceMode.fixedTranche,
        priceLimit: 64000.0,
        createdAt: DateTime.now(),
      );

      final tranches = IcebergSlicingAllocator.allocateTranches(config);
      expect(tranches.length, equals(5));

      double sum = 0.0;
      for (final t in tranches) {
        sum += t.visibleQuantity;
        expect(t.visibleQuantity, equals(2.0));
      }
      expect(sum, equals(10.0));

      expect(tranches.first.status, equals(TrancheStatus.active));
      expect(tranches[1].status, equals(TrancheStatus.pending));
      expect(tranches.last.hiddenRemaining, equals(0.0));
    });

    test('IcebergSlicingAllocator randomizes tranche sizes while preserving total order quantity', () {
      final config = IcebergOrderConfig(
        orderId: 'IC-4',
        symbol: 'ETH/USDT',
        side: 'SELL',
        totalQuantity: 100.0,
        visibleTrancheQuantity: 10.0,
        varianceMode: IcebergVarianceMode.randomizedVariance10to20,
        priceLimit: 3400.0,
        createdAt: DateTime.now(),
      );

      final tranches = IcebergSlicingAllocator.allocateTranches(config, randomSeed: 99);
      expect(tranches.isNotEmpty, isTrue);

      double sum = 0.0;
      for (final t in tranches) {
        sum += t.visibleQuantity;
      }
      expect(sum, closeTo(100.0, 0.001));

      // Check that some tranches vary from base 10.0
      final hasVariation = tranches.any((t) => (t.visibleQuantity - 10.0).abs() > 0.1);
      expect(hasVariation, isTrue);
    });

    test('IcebergOrderController reacts dynamically to user inputs and dispatches order', () async {
      expect(controller.state.simulatedTranches.isNotEmpty, isTrue);
      expect(controller.state.validation.isValid, isTrue);

      controller.setTotalQuantity(50.0);
      controller.setVisibleTrancheQuantity(5.0);
      controller.setVarianceMode(IcebergVarianceMode.fixedTranche);

      expect(controller.state.simulatedTranches.length, equals(10));
      expect(controller.state.hiddenQuantity, equals(45.0));

      final submitted = await controller.submitIcebergOrder();
      expect(submitted, isTrue);
      expect(controller.state.submittedOrderId, isNotNull);
      expect(controller.state.toastMessage, contains('10 Tranches'));
    });
  });
}
