import 'package:test/test.dart';
import 'package:growww_flutter/features/drawing_tools/domain/algorithms/drawing_geometry_engine.dart';
import 'package:growww_flutter/features/drawing_tools/domain/models/drawing_enums.dart';
import 'package:growww_flutter/features/drawing_tools/domain/models/drawing_point.dart';
import 'package:growww_flutter/features/drawing_tools/presentation/controllers/drawing_tools_controller.dart';

void main() {
  group('Prompt 558 - Chart Drawing Tools and Fibonacci Toolbar', () {
    late DrawingToolsController controller;

    setUp(() {
      controller = DrawingToolsController();
    });

    tearDown(() {
      controller.dispose();
    });

    test('DrawingGeometryEngine calculates exact Fibonacci retracement ratios and prices', () {
      // Swing Low at 60,000, Swing High at 70,000 (diff = 10,000)
      final levels = DrawingGeometryEngine.computeFibonacciLevels(
        anchorPrice1: 60000.0,
        anchorPrice2: 70000.0,
      );

      expect(levels.length, equals(9));

      final fib0 = levels.firstWhere((l) => l.ratio == 0.0);
      expect(fib0.price, equals(60000.0));

      final fib236 = levels.firstWhere((l) => l.ratio == 0.236);
      expect(fib236.price, equals(62360.0));

      final fib382 = levels.firstWhere((l) => l.ratio == 0.382);
      expect(fib382.price, equals(63820.0));

      final fib500 = levels.firstWhere((l) => l.ratio == 0.500);
      expect(fib500.price, equals(65000.0));

      final fib618 = levels.firstWhere((l) => l.ratio == 0.618);
      expect(fib618.price, equals(66180.0));

      final fib100 = levels.firstWhere((l) => l.ratio == 1.0);
      expect(fib100.price, equals(70000.0));

      final fib1618 = levels.firstWhere((l) => l.ratio == 1.618);
      expect(fib1618.price, equals(76180.0));
    });

    test('DrawingGeometryEngine calculates perpendicular point-to-line touch distance', () {
      // Line from (0, 0) to (100, 0) (horizontal line along X axis)
      final d1 = DrawingGeometryEngine.distanceToLineSegment(
        px: 50,
        py: 10,
        x1: 0,
        y1: 0,
        x2: 100,
        y2: 0,
      );
      expect(d1, equals(10.0));

      // Point past line end
      final d2 = DrawingGeometryEngine.distanceToLineSegment(
        px: 110,
        py: 0,
        x1: 0,
        y1: 0,
        x2: 100,
        y2: 0,
      );
      expect(d2, equals(10.0));
    });

    test('DrawingGeometryEngine snaps target price to nearest candle OHLC level', () {
      final ohlc = [64150.0, 64280.0, 64100.0, 64220.0];

      final snapped1 = DrawingGeometryEngine.snapToOhlc(
        rawPrice: 64222.0,
        ohlcLevels: ohlc,
        threshold: 5.0,
      );
      expect(snapped1, equals(64220.0)); // Snapped to close

      final notSnapped = DrawingGeometryEngine.snapToOhlc(
        rawPrice: 64500.0,
        ohlcLevels: ohlc,
        threshold: 5.0,
      );
      expect(notSnapped, equals(64500.0)); // Beyond threshold
    });

    test('DrawingToolsController manages point collection, undo/redo stack, and element creation', () {
      controller.setActiveTool(DrawingToolType.trendline);
      expect(controller.state.activeTool, equals(DrawingToolType.trendline));

      // Trendline requires 2 points
      const p1 = DrawingPoint(timestampMs: 1000, price: 64000.0, screenPixelX: 50.0, screenPixelY: 100.0);
      controller.addPointToDrawing(p1);
      expect(controller.state.inProgressPoints.length, equals(1));
      expect(controller.state.elements.isEmpty, isTrue);

      const p2 = DrawingPoint(timestampMs: 2000, price: 65000.0, screenPixelX: 150.0, screenPixelY: 80.0);
      controller.addPointToDrawing(p2);
      expect(controller.state.elements.length, equals(1));
      expect(controller.state.inProgressPoints.isEmpty, isTrue);

      // Test Undo
      controller.undo();
      expect(controller.state.elements.isEmpty, isTrue);
      expect(controller.state.redoStack.length, equals(1));

      // Test Redo
      controller.redo();
      expect(controller.state.elements.length, equals(1));
    });
  });
}
