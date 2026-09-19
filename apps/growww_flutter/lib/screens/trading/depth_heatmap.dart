import 'package:flutter/material.dart';

class DepthPriceLevel {
  final double price;
  final double volume;
  final double cumulativeVolume;

  DepthPriceLevel({required this.price, required this.volume, required this.cumulativeVolume});
}

class DepthHeatmapWidget extends StatelessWidget {
  final List<DepthPriceLevel> bids;
  final List<DepthPriceLevel> asks;

  const DepthHeatmapWidget({Key? key, required this.bids, required this.asks}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    double maxCumulative = 1.0;
    if (bids.isNotEmpty && bids.last.cumulativeVolume > maxCumulative) {
      maxCumulative = bids.last.cumulativeVolume;
    }
    if (asks.isNotEmpty && asks.last.cumulativeVolume > maxCumulative) {
      maxCumulative = asks.last.cumulativeVolume;
    }

    return Container(
      color: const Color(0xFF08090C),
      padding: const EdgeInsets.all(8),
      child: Column(
        children: [
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 4.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('PRICE (USDT)', style: TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontWeight: FontWeight.bold)),
                Text('SIZE', style: TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontWeight: FontWeight.bold)),
                Text('TOTAL', style: TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                // Bids Side (Green Heatmap)
                Expanded(
                  child: ListView.builder(
                    itemCount: bids.length,
                    itemBuilder: (ctx, i) {
                      final b = bids[i];
                      final ratio = (b.cumulativeVolume / maxCumulative).clamp(0.0, 1.0);
                      return Stack(
                        children: [
                          Positioned.fill(
                            child: Align(
                              alignment: Alignment.centerRight,
                              child: FractionallySizedBox(
                                widthFactor: ratio,
                                child: Container(color: const Color(0xFF00E676).withOpacity(0.15)),
                              ),
                            ),
                          ),
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 2.0, horizontal: 4.0),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(b.price.toStringAsFixed(2), style: const TextStyle(color: Color(0xFF00E676), fontSize: 12, fontFamily: 'monospace')),
                                Text(b.volume.toStringAsFixed(4), style: const TextStyle(color: Colors.white, fontSize: 12, fontFamily: 'monospace')),
                              ],
                            ),
                          ),
                        ],
                      );
                    },
                  ),
                ),
                const SizedBox(width: 4),
                // Asks Side (Red Heatmap)
                Expanded(
                  child: ListView.builder(
                    itemCount: asks.length,
                    itemBuilder: (ctx, i) {
                      final a = asks[i];
                      final ratio = (a.cumulativeVolume / maxCumulative).clamp(0.0, 1.0);
                      return Stack(
                        children: [
                          Positioned.fill(
                            child: Align(
                              alignment: Alignment.centerLeft,
                              child: FractionallySizedBox(
                                widthFactor: ratio,
                                child: Container(color: const Color(0xFFFF1744).withOpacity(0.15)),
                              ),
                            ),
                          ),
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 2.0, horizontal: 4.0),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(a.volume.toStringAsFixed(4), style: const TextStyle(color: Colors.white, fontSize: 12, fontFamily: 'monospace')),
                                Text(a.price.toStringAsFixed(2), style: const TextStyle(color: Color(0xFFFF1744), fontSize: 12, fontFamily: 'monospace')),
                              ],
                            ),
                          ),
                        ],
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
