import 'package:flutter/material.dart';
import '../../domain/models/iceberg_enums.dart';
import '../controllers/iceberg_order_controller.dart';

class IcebergOrderSlicingScreen extends StatefulWidget {
  final IcebergOrderController controller;

  const IcebergOrderSlicingScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<IcebergOrderSlicingScreen> createState() =>
      _IcebergOrderSlicingScreenState();
}

class _IcebergOrderSlicingScreenState extends State<IcebergOrderSlicingScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<IcebergOrderState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        final isBuy = state.side.toUpperCase() == 'BUY';

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: const Text(
              'FLUTTER ICEBERG ORDER SLICING & HIDDEN SIZE Q',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.0,
              ),
            ),
          ),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _buildSideToggle(isBuy),
                const SizedBox(height: 16),
                _buildInputsCard(state),
                const SizedBox(height: 16),
                _buildVarianceSelector(state),
                const SizedBox(height: 16),
                _buildTrancheBreakdownCard(state),
                const SizedBox(height: 16),
                if (!state.validation.isValid)
                  _buildErrorAlert(state.validation.errorMessage!),
                const SizedBox(height: 24),
                _buildActionButtons(state, isBuy),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildSideToggle(bool isBuy) {
    return Row(
      children: [
        Expanded(
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: isBuy ? neonGreen : surfaceCard,
              foregroundColor: isBuy ? Colors.black : textMuted,
              padding: const EdgeInsets.symmetric(vertical: 12),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () => widget.controller.setSide('BUY'),
            child: const Text('BUY ICEBERG', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: !isBuy ? neonRed : surfaceCard,
              foregroundColor: !isBuy ? Colors.white : textMuted,
              padding: const EdgeInsets.symmetric(vertical: 12),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () => widget.controller.setSide('SELL'),
            child: const Text('SELL ICEBERG', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ),
      ],
    );
  }

  Widget _buildInputsCard(IcebergOrderState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildNumericRow(
            label: 'Total Order Quantity (${state.symbol.split('/').first})',
            value: state.totalQuantity,
            onChanged: (v) => widget.controller.setTotalQuantity(v),
          ),
          const SizedBox(height: 12),
          _buildNumericRow(
            label: 'Visible Tranche Size (${state.symbol.split('/').first})',
            value: state.visibleTrancheQuantity,
            onChanged: (v) => widget.controller.setVisibleTrancheQuantity(v),
          ),
          const SizedBox(height: 12),
          _buildNumericRow(
            label: 'Limit Execution Price (₹)',
            value: state.priceLimit,
            onChanged: (v) => widget.controller.setPriceLimit(v),
          ),
          const Divider(color: Colors.white12, height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Hidden Size In Book', style: TextStyle(color: textMuted, fontSize: 12)),
              Text(
                '${state.hiddenQuantity.toStringAsFixed(2)} ${state.symbol.split('/').first}',
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Total Order Value', style: TextStyle(color: textMuted, fontSize: 12)),
              Text(
                '₹${state.estimatedTotalInr.toStringAsFixed(2)}',
                style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildNumericRow({
    required String label,
    required double value,
    required ValueChanged<double> onChanged,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Flexible(child: Text(label, style: const TextStyle(color: Colors.white70, fontSize: 12))),
        Row(
          children: [
            IconButton(
              icon: const Icon(Icons.remove, color: textMuted, size: 18),
              onPressed: () => onChanged((value - (value > 10 ? 10.0 : 0.5)).clamp(0.1, 1000000.0)),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
              decoration: BoxDecoration(color: obsidianBackground, borderRadius: BorderRadius.circular(4)),
              child: Text(
                value.toStringAsFixed(2),
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.add, color: neonGreen, size: 18),
              onPressed: () => onChanged(value + (value >= 10 ? 10.0 : 0.5)),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildVarianceSelector(IcebergOrderState state) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Anti-Detection Display Variance', style: TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(height: 6),
          for (final mode in IcebergVarianceMode.values) ...[
            RadioListTile<IcebergVarianceMode>(
              value: mode,
              groupValue: state.varianceMode,
              activeColor: neonGreen,
              dense: true,
              contentPadding: EdgeInsets.zero,
              title: Text(mode.displayName, style: const TextStyle(color: Colors.white, fontSize: 12)),
              onChanged: (m) => widget.controller.setVarianceMode(m!),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildTrancheBreakdownCard(IcebergOrderState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Allocated Tranche Slices', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
              Text(
                '${state.simulatedTranches.length} Tranches',
                style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontSize: 12, fontWeight: FontWeight.bold),
              ),
            ],
          ),
          const SizedBox(height: 10),
          for (final t in state.simulatedTranches) ...[
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 3.0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Tranche #${t.trancheIndex + 1}', style: const TextStyle(color: textMuted, fontSize: 11)),
                  Text(
                    'Visible: ${t.visibleQuantity.toStringAsFixed(4)} (Hidden: ${t.hiddenRemaining.toStringAsFixed(4)})',
                    style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 11),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildErrorAlert(String message) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: neonRed.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: neonRed),
      ),
      child: Text(message, style: const TextStyle(color: neonRed, fontSize: 11)),
    );
  }

  Widget _buildActionButtons(IcebergOrderState state, bool isBuy) {
    return Row(
      children: [
        Expanded(
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: isBuy ? neonGreen : neonRed,
              foregroundColor: isBuy ? Colors.black : Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: state.validation.isValid && !state.isSubmitting
                ? () async {
                    final ok = await widget.controller.submitIcebergOrder();
                    if (ok && mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(widget.controller.state.toastMessage ?? 'Iceberg Order Dispatched!'),
                          backgroundColor: const Color(0xFF1B382B),
                        ),
                      );
                    }
                  }
                : null,
            child: const Text('Dispatch Iceberg (Neon Green)', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: OutlinedButton(
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.white,
              side: const BorderSide(color: Colors.white30),
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () => Navigator.of(context).maybePop(),
            child: const Text('Cancel Slicing'),
          ),
        ),
      ],
    );
  }
}
