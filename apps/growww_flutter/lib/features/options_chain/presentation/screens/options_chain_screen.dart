import 'package:flutter/material.dart';
import '../../domain/models/option_contract_quote.dart';
import '../../domain/models/options_chain_snapshot.dart';
import '../../domain/models/options_enums.dart';
import '../../domain/models/options_strike_row.dart';
import '../controllers/options_chain_controller.dart';

/// Flutter Options Chain Matrix & Strike Selector Screen
/// Institutional cyber-aesthetic options matrix with live ITM/ATM shading,
/// Put-Call Ratio (PCR), Max Pain analytics, Greeks toggle, and click-to-trade order drawer.
class OptionsChainScreen extends StatefulWidget {
  final OptionsChainController controller;

  const OptionsChainScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<OptionsChainScreen> createState() => _OptionsChainScreenState();
}

class _OptionsChainScreenState extends State<OptionsChainScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color neonYellow = Color(0xFFFFD166);
  static const Color neonCyan = Color(0xFF00E5FF);
  static const Color itmCallShade = Color(0x1A00F0A0);
  static const Color itmPutShade = Color(0x1AFF3B56);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<OptionsChainState>(
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
            title: Row(
              children: [
                Text(
                  snap.underlyingSymbol,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                    fontSize: 15,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: neonGreen.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '₹${snap.spotPrice.toStringAsFixed(2)}',
                    style: const TextStyle(
                      color: neonGreen,
                      fontFamily: 'JetBrains Mono',
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            actions: [
              IconButton(
                icon: Icon(
                  state.showGreeks ? Icons.analytics : Icons.analytics_outlined,
                  color: state.showGreeks ? neonCyan : textMuted,
                ),
                tooltip: 'Toggle Greeks',
                onPressed: () => widget.controller.toggleGreeks(),
              ),
            ],
            bottom: PreferredSize(
              preferredSize: const Size.fromHeight(88),
              child: Column(
                children: [
                  _buildMarketMetricsHeader(snap),
                  _buildExpirySelector(snap),
                ],
              ),
            ),
          ),
          body: Column(
            children: [
              _buildFilterBar(state),
              _buildMatrixTableHeader(state),
              Expanded(
                child: ListView.builder(
                  itemCount: snap.strikeRows.length,
                  itemBuilder: (context, index) {
                    final row = snap.strikeRows[index];
                    return _buildStrikeRowItem(row, state);
                  },
                ),
              ),
              if (state.selectedContract != null)
                _buildOrderStagingDrawer(state),
            ],
          ),
        );
      },
    );
  }

  Widget _buildMarketMetricsHeader(OptionsChainSnapshot snap) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      color: surfaceCard,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              const Text('PCR: ', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                snap.overallPcr.toStringAsFixed(2),
                style: TextStyle(
                  color: snap.overallPcr >= 1.0 ? neonGreen : neonRed,
                  fontFamily: 'JetBrains Mono',
                  fontWeight: FontWeight.bold,
                  fontSize: 11,
                ),
              ),
            ],
          ),
          Row(
            children: [
              const Text('Max Pain: ', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '₹${snap.maxPainStrike.toStringAsFixed(0)}',
                style: const TextStyle(
                  color: neonYellow,
                  fontFamily: 'JetBrains Mono',
                  fontWeight: FontWeight.bold,
                  fontSize: 11,
                ),
              ),
            ],
          ),
          Row(
            children: [
              const Text('Total Call OI: ', style: TextStyle(color: textMuted, fontSize: 11)),
              Text(
                '${(snap.totalCallOi / 1000).toStringAsFixed(1)}k',
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 11),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildExpirySelector(OptionsChainSnapshot snap) {
    return Container(
      height: 40,
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        itemCount: snap.availableExpiries.length,
        itemBuilder: (context, index) {
          final exp = snap.availableExpiries[index];
          final isSelected = exp == snap.selectedExpiry;

          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: ChoiceChip(
              label: Text(
                exp,
                style: TextStyle(
                  color: isSelected ? Colors.black : Colors.white70,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
              selected: isSelected,
              selectedColor: neonCyan,
              backgroundColor: surfaceCard,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
              onSelected: (_) => widget.controller.selectExpiry(exp),
            ),
          );
        },
      ),
    );
  }

  Widget _buildFilterBar(OptionsChainState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: const BoxDecoration(
        color: obsidianBackground,
        border: Border(bottom: BorderSide(color: Colors.white12)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: OptionsViewFilter.values.map((f) {
              final isSel = state.filter == f;
              return InkWell(
                onTap: () => widget.controller.setFilter(f),
                child: Container(
                  margin: const EdgeInsets.only(right: 6),
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: isSel ? Colors.white12 : Colors.transparent,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    f.label,
                    style: TextStyle(
                      color: isSel ? neonGreen : textMuted,
                      fontSize: 11,
                      fontWeight: isSel ? FontWeight.bold : FontWeight.normal,
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildMatrixTableHeader(OptionsChainState state) {
    final showCalls = state.filter != OptionsViewFilter.putsOnly;
    final showPuts = state.filter != OptionsViewFilter.callsOnly;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      color: surfaceCard,
      child: Row(
        children: [
          if (showCalls)
            Expanded(
              flex: 4,
              child: Text(
                state.showGreeks ? 'CALLS (Delta | IV | LTP)' : 'CALLS (OI | Vol | LTP)',
                textAlign: TextAlign.left,
                style: const TextStyle(color: neonGreen, fontSize: 10, fontWeight: FontWeight.bold),
              ),
            ),
          Container(
            width: 76,
            alignment: Alignment.center,
            child: const Text(
              'STRIKE',
              style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
            ),
          ),
          if (showPuts)
            Expanded(
              flex: 4,
              child: Text(
                state.showGreeks ? 'PUTS (LTP | IV | Delta)' : 'PUTS (LTP | Vol | OI)',
                textAlign: TextAlign.right,
                style: const TextStyle(color: neonRed, fontSize: 10, fontWeight: FontWeight.bold),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildStrikeRowItem(OptionsStrikeRow row, OptionsChainState state) {
    final isAtm = row.isAtm;
    final showCalls = state.filter != OptionsViewFilter.putsOnly;
    final showPuts = state.filter != OptionsViewFilter.callsOnly;

    final isCallItm = row.call.moneyness.isInTheMoney;
    final isPutItm = row.put.moneyness.isInTheMoney;

    return Container(
      decoration: BoxDecoration(
        color: isAtm ? neonYellow.withOpacity(0.08) : Colors.transparent,
        border: Border(
          bottom: const BorderSide(color: Colors.white10),
          top: isAtm ? const BorderSide(color: neonYellow, width: 1.5) : BorderSide.none,
        ),
      ),
      child: Row(
        children: [
          // Calls Section
          if (showCalls)
            Expanded(
              flex: 4,
              child: InkWell(
                onTap: () => widget.controller.selectContract(row.call, 'BUY'),
                child: Container(
                  color: isCallItm ? itmCallShade : Colors.transparent,
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        state.showGreeks
                            ? '${row.call.delta.toStringAsFixed(2)} | ${(row.call.iv * 100).toStringAsFixed(0)}%'
                            : '${row.call.openInterest.toInt()} | ${row.call.volume.toInt()}',
                        style: const TextStyle(color: textMuted, fontFamily: 'JetBrains Mono', fontSize: 10),
                      ),
                      Text(
                        '₹${row.call.ltp.toStringAsFixed(1)}',
                        style: const TextStyle(
                          color: Colors.white,
                          fontFamily: 'JetBrains Mono',
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),

          // Center Strike Price Pill
          Container(
            width: 76,
            padding: const EdgeInsets.symmetric(vertical: 8),
            color: surfaceCard,
            child: Center(
              child: Text(
                row.strikePrice.toStringAsFixed(0),
                style: TextStyle(
                  color: isAtm ? neonYellow : Colors.white,
                  fontFamily: 'JetBrains Mono',
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ),
          ),

          // Puts Section
          if (showPuts)
            Expanded(
              flex: 4,
              child: InkWell(
                onTap: () => widget.controller.selectContract(row.put, 'BUY'),
                child: Container(
                  color: isPutItm ? itmPutShade : Colors.transparent,
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        '₹${row.put.ltp.toStringAsFixed(1)}',
                        style: const TextStyle(
                          color: Colors.white,
                          fontFamily: 'JetBrains Mono',
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        state.showGreeks
                            ? '${(row.put.iv * 100).toStringAsFixed(0)}% | ${row.put.delta.toStringAsFixed(2)}'
                            : '${row.put.volume.toInt()} | ${row.put.openInterest.toInt()}',
                        style: const TextStyle(color: textMuted, fontFamily: 'JetBrains Mono', fontSize: 10),
                      ),
                    ],
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildOrderStagingDrawer(OptionsChainState state) {
    final contract = state.selectedContract!;
    final isBuy = state.stagedSide == 'BUY';

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        border: const Border(top: BorderSide(color: neonCyan, width: 1.5)),
        boxShadow: const [BoxShadow(color: Colors.black54, blurRadius: 10)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    contract.symbol,
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'LTP: ₹${contract.ltp.toStringAsFixed(2)} • IV: ${(contract.iv * 100).toStringAsFixed(1)}% • Delta: ${contract.delta.toStringAsFixed(2)}',
                    style: const TextStyle(color: textMuted, fontSize: 11),
                  ),
                ],
              ),
              IconButton(
                icon: const Icon(Icons.close, color: textMuted, size: 20),
                onPressed: () => widget.controller.clearSelection(),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: isBuy ? neonGreen : Colors.white10,
                    foregroundColor: isBuy ? Colors.black : Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                  ),
                  onPressed: () => widget.controller.selectContract(contract, 'BUY'),
                  child: const Text('BUY', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: !isBuy ? neonRed : Colors.white10,
                    foregroundColor: !isBuy ? Colors.white : Colors.white70,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                  ),
                  onPressed: () => widget.controller.selectContract(contract, 'SELL'),
                  child: const Text('SELL', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ),
              const SizedBox(width: 12),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: neonCyan,
                  foregroundColor: Colors.black,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
                ),
                onPressed: () => widget.controller.placeStagedOrder(),
                child: const Text('Execute Order', style: TextStyle(fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
