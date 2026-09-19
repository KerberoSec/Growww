import 'package:flutter/material.dart';

class CandlestickPoint {
  final int timestampMs;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;

  CandlestickPoint({
    required this.timestampMs,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
  });

  bool get isBullish => close >= open;
}

class TradingViewChartWidget extends StatefulWidget {
  final String symbol;
  final String interval; // 1m, 5m, 15m, 1h, 1d
  final List<CandlestickPoint> candles;

  const TradingViewChartWidget({
    Key? key,
    required this.symbol,
    this.interval = '15m',
    required this.candles,
  }) : super(key: key);

  @override
  State<TradingViewChartWidget> createState() => _TradingViewChartWidgetState();
}

class _TradingViewChartWidgetState extends State<TradingViewChartWidget> {
  String _selectedInterval = '15m';

  @override
  void initState() {
    super.initState();
    _selectedInterval = widget.interval;
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF12141A),
      child: Column(
        children: [
          // Interval Toolbar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: Color(0xFF232732))),
            ),
            child: Row(
              children: [
                Text(
                  widget.symbol,
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
                ),
                const SizedBox(width: 16),
                for (final iv in ['1m', '5m', '15m', '1h', '4h', '1D'])
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: InkWell(
                      onTap: () => setState(() => _selectedInterval = iv),
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: _selectedInterval == iv ? const Color(0xFF222630) : Colors.transparent,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          iv,
                          style: TextStyle(
                            color: _selectedInterval == iv ? const Color(0xFF00E676) : const Color(0xFF9096A2),
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ),
                  ),
                const Spacer(),
                const Icon(Icons.show_chart, color: Color(0xFF9096A2), size: 18),
                const SizedBox(width: 12),
                const Icon(Icons.fullscreen, color: Color(0xFF9096A2), size: 18),
              ],
            ),
          ),
          // Chart Canvas / Candlestick Renderer
          Expanded(
            child: CustomPaint(
              size: Size.infinite,
              painter: CandlestickPainter(candles: widget.candles),
            ),
          ),
        ],
      ),
    );
  }
}

class CandlestickPainter extends CustomPainter {
  final List<CandlestickPoint> candles;

  CandlestickPainter({required this.candles});

  @override
  void paint(Canvas canvas, Size size) {
    if (candles.isEmpty) return;

    final bullPaint = Paint()..color = const Color(0xFF00E676)..style = PaintingStyle.fill;
    final bearPaint = Paint()..color = const Color(0xFFFF1744)..style = PaintingStyle.fill;
    final wickPaint = Paint()..strokeWidth = 1.2;

    double minPrice = candles.map((c) => c.low).reduce((a, b) => a < b ? a : b);
    double maxPrice = candles.map((c) => c.high).reduce((a, b) => a > b ? a : b);
    if (maxPrice == minPrice) maxPrice += 1.0;

    double candleWidth = (size.width / candles.length).clamp(2.0, 16.0);

    for (int i = 0; i < candles.length; i++) {
      final c = candles[i];
      final x = i * (size.width / candles.length) + (size.width / candles.length) / 2;

      final yOpen = size.height - ((c.open - minPrice) / (maxPrice - minPrice)) * size.height;
      final yClose = size.height - ((c.close - minPrice) / (maxPrice - minPrice)) * size.height;
      final yHigh = size.height - ((c.high - minPrice) / (maxPrice - minPrice)) * size.height;
      final yLow = size.height - ((c.low - minPrice) / (maxPrice - minPrice)) * size.height;

      final paint = c.isBullish ? bullPaint : bearPaint;
      wickPaint.color = c.isBullish ? const Color(0xFF00E676) : const Color(0xFFFF1744);

      // Draw wick
      canvas.drawLine(Offset(x, yHigh), Offset(x, yLow), wickPaint);

      // Draw candle body
      final top = yOpen < yClose ? yOpen : yClose;
      final bottom = yOpen > yClose ? yOpen : yClose;
      final height = (bottom - top).clamp(1.0, size.height);

      canvas.drawRect(
        Rect.fromCenter(center: Offset(x, top + height / 2), width: candleWidth * 0.75, height: height),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}
