import 'dart:math' as math;
import '../models/drawing_point.dart';
import '../models/fibonacci_level.dart';

class DrawingGeometryEngine {
  static const List<double> standardFibRatios = [
    0.0,
    0.236,
    0.382,
    0.500,
    0.618,
    0.786,
    1.000,
    1.618,
    2.618,
  ];

  static final Map<double, int> fibColors = {
    0.000: 0xFF8B949E, // Gray
    0.236: 0xFFFF5252, // Coral Red
    0.382: 0xFFFFB142, // Amber Orange
    0.500: 0xFF2ED573, // Emerald Green
    0.618: 0xFF00E5FF, // Cyan
    0.786: 0xFF70A1FF, // Cobalt Blue
    1.000: 0xFF8B949E, // Gray
    1.618: 0xFFFF78E1, // Magenta
    2.618: 0xFFFF4757, // Crimson
  };

  /// Computes exact price levels for Fibonacci retracement
  static List<FibonacciLevel> computeFibonacciLevels({
    required double anchorPrice1,
    required double anchorPrice2,
    List<double>? customRatios,
  }) {
    final ratios = customRatios ?? standardFibRatios;
    final diff = anchorPrice2 - anchorPrice1;

    return ratios.map((ratio) {
      final price = anchorPrice1 + (diff * ratio);
      final color = fibColors[ratio] ?? 0xFF00F0A0;
      final label = '${(ratio * 100).toStringAsFixed(1)}% (₹${price.toStringAsFixed(2)})';

      return FibonacciLevel(
        ratio: ratio,
        price: (price * 100).round() / 100.0,
        colorHex: color,
        label: label,
      );
    }).toList();
  }

  /// Calculates perpendicular pixel distance from point (px, py) to line segment (p1, p2)
  static double distanceToLineSegment({
    required double px,
    required double py,
    required double x1,
    required double y1,
    required double x2,
    required double y2,
  }) {
    final l2 = (x2 - x1) * (x2 - x1) + (y2 - y1) * (y2 - y1);
    if (l2 == 0) {
      return math.sqrt((px - x1) * (px - x1) + (py - y1) * (py - y1));
    }

    // Projection factor t clamped between 0 and 1
    final t = math.max(0.0, math.min(1.0, ((px - x1) * (x2 - x1) + (py - y1) * (y2 - y1)) / l2));
    final projX = x1 + t * (x2 - x1);
    final projY = y1 + t * (y2 - y1);

    return math.sqrt((px - projX) * (px - projX) + (py - projY) * (py - projY));
  }

  /// Magnetic snapping: Snaps target price to nearest candle OHLC level if within threshold
  static double snapToOhlc({
    required double rawPrice,
    required List<double> ohlcLevels,
    double threshold = 5.0,
  }) {
    if (ohlcLevels.isEmpty) return rawPrice;
    double closest = rawPrice;
    double minDiff = threshold;

    for (final level in ohlcLevels) {
      final diff = (rawPrice - level).abs();
      if (diff < minDiff) {
        minDiff = diff;
        closest = level;
      }
    }
    return closest;
  }
}
