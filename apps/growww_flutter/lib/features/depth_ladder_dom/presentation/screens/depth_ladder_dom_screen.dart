import 'package:flutter/material.dart';
import '../../domain/models/depth_ladder_enums.dart';
import '../../domain/models/ladder_depth_level.dart';
import '../../domain/models/ladder_dom_snapshot.dart';
import '../controllers/depth_ladder_controller.dart';

/// Flutter Live L2/L3 Depth Ladder DOM & Orderbook Screen
/// Institutional Depth of Market (DOM) workstation featuring bidirectional depth volume bars,
/// L2/L3 MBO inspection, tick aggregation selector, and one-tap click-to-trade limit order execution.
class DepthLadderDomScreen extends StatefulWidget {
  final DepthLadderController controller;

  const DepthLadderDomScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<DepthLadderDomScreen> createState() => _DepthLadderDomScreenState();
}

class _DepthLadderDomScreenState extends State<DepthLadderDomScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color neonCyan = Color(0xFF00E5FF);
  static const Color neonYellow = Color(0xFFFFD166);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<DepthLadderState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        final snap = state.snapshot;

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            leading: IconButton(
              icon: const Icon(Icons.arrow_back_ios, color: Colors.white, size: 18),
              onPressed: () => Navigator.of(context).maybePop(),
            ),
            title: Row(
              children: [
                Text(
                  snap.symbol,
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: neonGreen.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '₹${snap.lastTradedPrice.toStringAsFixed(2)}',
                    style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontSize: 12, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.vertical_align_center, color: neonCyan, size: 20),
                tooltip: 'Center Ladder',
                onPressed: () => widget.controller.centerOnLastPrice(),
              ),
            ],
            bottom: PreferredSize(
              preferredSize: const Size.fromHeight(80),
              child: Column(
                children: [
                  _buildSpreadAndMetricsBar(snap),
                  _buildAggregationControls(state),
                ],
              ),
            ),
          ),
          body: Column(
            children: [
              _buildLadderHeader(state),
              Expanded(
                child: ListView(
                  children: [
                    // Asks ladder (inverted so lowest ask is near spread)
                    ...snap.askLevels.reversed.map((lvl) => _buildLadderRow(lvl, state)),
                    // Spread Divider
                    _buildSpreadCenterRow(snap),
                    // Bids ladder (highest bid near spread)
                    ...snap.bidLevels.map((lvl) => _buildLadderRow(lvl, state)),
                  ],
                ),
              ),
              if (state.isOrderStaged)
                _buildOrderStagingBar(state),
            ],
          ),
        );
      },
    );
  }

  Widget _buildSpreadAndMetricsBar(LadderDomSnapshot snap) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      color: surfaceCard,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              const Text('Spread: ', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '₹${snap.spread.toStringAsFixed(2)} (${snap.spreadBps.toStringAsFixed(1)} bps)',
                style: const TextStyle(color: neonYellow, fontFamily: 'JetBrains Mono', fontSize: 11, fontWeight: FontWeight.bold),
              ),
            ],
          ),
          Row(
            children: [
              const Text('24h Change: ', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '+${snap.priceChangePercent24h.toStringAsFixed(2)}%',
                style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontSize: 11, fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAggregationControls(DepthLadderState state) {
    return Container(
      height: 44,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: const BoxDecoration(
        color: obsidianBackground,
        border: Border(bottom: BorderSide(color: Colors.white12)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              const Text('Tick: ', style: TextStyle(color: textMuted, fontSize: 11)),
              ...LadderTickAggregation.values.map((agg) {
                final isSel = state.aggregation == agg;
                return InkWell(
                  onTap: () => widget.controller.setAggregation(agg),
                  child: Container(
                    margin: const EdgeInsets.symmetric(horizontal: 2),
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                    decoration: BoxDecoration(
                      color: isSel ? neonCyan.withOpacity(0.2) : surfaceCard,
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(color: isSel ? neonCyan : Colors.white12),
                    ),
                    child: Text(
                      agg.label,
                      style: TextStyle(
                        color: isSel ? neonCyan : Colors.white70,
                        fontSize: 10,
                        fontFamily: 'JetBrains Mono',
                        fontWeight: isSel ? FontWeight.bold : FontWeight.normal,
                      ),
                    ),
                  ),
                );
              }),
            ],
          ),
          InkWell(
            onTap: () {
              final nextMode = state.depthMode == DepthMode.level2Aggregated
                  ? DepthMode.level3MarketByOrder
                  : DepthMode.level2Aggregated;
              widget.controller.setDepthMode(nextMode);
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
              decoration: BoxDecoration(
                color: surfaceCard,
                borderRadius: BorderRadius.circular(4),
                border: Border.all(color: Colors.white24),
              ),
              child: Text(
                state.depthMode == DepthMode.level2Aggregated ? 'L2' : 'L3 MBO',
                style: const TextStyle(color: neonCyan, fontSize: 10, fontWeight: FontWeight.bold),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLadderHeader(DepthLadderState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      color: surfaceCard,
      child: Row(
        children: const [
          Expanded(
            flex: 3,
            child: Text('BID SIZE (CUMULATIVE)', style: TextStyle(color: neonGreen, fontSize: 10, fontWeight: FontWeight.bold)),
          ),
          Expanded(
            flex: 2,
            child: Center(
              child: Text('PRICE LADDER', style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
            ),
          ),
          Expanded(
            flex: 3,
            child: Text('ASK SIZE (CUMULATIVE)', textAlign: TextAlign.right, style: TextStyle(color: neonRed, fontSize: 10, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  Widget _buildSpreadCenterRow(LadderDomSnapshot snap) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 4),
      color: surfaceCard.withOpacity(0.5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: obsidianBackground,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: neonYellow.withOpacity(0.5)),
            ),
            child: Text(
              'SPREAD ₹${snap.spread.toStringAsFixed(2)}',
              style: const TextStyle(color: neonYellow, fontSize: 10, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLadderRow(LadderDepthLevel lvl, DepthLadderState state) {
    final isBid = lvl.side.isBid;
    final isSelected = state.selectedPrice == lvl.price;

    return InkWell(
      onTap: () => widget.controller.selectPriceLevel(lvl.price, lvl.side),
      child: Container(
        height: 28,
        decoration: BoxDecoration(
          color: isSelected ? neonCyan.withOpacity(0.15) : Colors.transparent,
          border: Border(
            bottom: const BorderSide(color: Colors.white10),
            left: isSelected ? const BorderSide(color: neonCyan, width: 2) : BorderSide.none,
          ),
        ),
        child: Row(
          children: [
            // Bid Side Bar & Text
            Expanded(
              flex: 3,
              child: isBid
                  ? Stack(
                      children: [
                        Align(
                          alignment: Alignment.centerLeft,
                          child: FractionallySizedBox(
                            widthFactor: lvl.depthPercent,
                            child: Container(color: neonGreen.withOpacity(0.18)),
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                          child: Align(
                            alignment: Alignment.centerLeft,
                            child: Text(
                              '${lvl.quantity.toStringAsFixed(2)} (${lvl.orderCount})',
                              style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontSize: 11),
                            ),
                          ),
                        ),
                      ],
                    )
                  : const SizedBox(),
            ),

            // Center Price Pill
            Expanded(
              flex: 2,
              child: Container(
                color: surfaceCard,
                alignment: Alignment.center,
                child: Text(
                  lvl.price.toStringAsFixed(2),
                  style: TextStyle(
                    color: lvl.isBestPrice ? (isBid ? neonGreen : neonRed) : Colors.white,
                    fontFamily: 'JetBrains Mono',
                    fontWeight: lvl.isBestPrice ? FontWeight.bold : FontWeight.w500,
                    fontSize: 11,
                  ),
                ),
              ),
            ),

            // Ask Side Bar & Text
            Expanded(
              flex: 3,
              child: !isBid
                  ? Stack(
                      children: [
                        Align(
                          alignment: Alignment.centerRight,
                          child: FractionallySizedBox(
                            widthFactor: lvl.depthPercent,
                            child: Container(color: neonRed.withOpacity(0.18)),
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                          child: Align(
                            alignment: Alignment.centerRight,
                            child: Text(
                              '(${lvl.orderCount}) ${lvl.quantity.toStringAsFixed(2)}',
                              style: const TextStyle(color: neonRed, fontFamily: 'JetBrains Mono', fontSize: 11),
                            ),
                          ),
                        ),
                      ],
                    )
                  : const SizedBox(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildOrderStagingBar(DepthLadderState state) {
    final isBid = state.stagedSide?.isBid ?? true;

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: surfaceCard,
        border: const Border(top: BorderSide(color: neonCyan, width: 1.5)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: isBid ? neonGreen.withOpacity(0.2) : neonRed.withOpacity(0.2),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              isBid ? 'BUY LIMIT' : 'SELL LIMIT',
              style: TextStyle(color: isBid ? neonGreen : neonRed, fontWeight: FontWeight.bold, fontSize: 11),
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '@ ₹${state.selectedPrice?.toStringAsFixed(2)}',
            style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13),
          ),
          const Spacer(),
          IconButton(
            icon: const Icon(Icons.remove_circle_outline, color: textMuted, size: 18),
            onPressed: () => widget.controller.setStagedQuantity(state.stagedQuantity - 0.5),
          ),
          Text(
            '${state.stagedQuantity.toStringAsFixed(1)}',
            style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 12),
          ),
          IconButton(
            icon: const Icon(Icons.add_circle_outline, color: neonGreen, size: 18),
            onPressed: () => widget.controller.setStagedQuantity(state.stagedQuantity + 0.5),
          ),
          const SizedBox(width: 8),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: isBid ? neonGreen : neonRed,
              foregroundColor: isBid ? Colors.black : Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () => widget.controller.executeStagedOrder(),
            child: const Text('Place Limit', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)),
          ),
          IconButton(
            icon: const Icon(Icons.close, color: textMuted, size: 18),
            onPressed: () => widget.controller.clearStagedOrder(),
          ),
        ],
      ),
    );
  }
}
