import 'package:flutter/material.dart';
import '../../domain/models/trigger_enums.dart';
import '../controllers/stop_limit_controller.dart';

class StopLimitTriggerConfigurationScreen extends StatefulWidget {
  final StopLimitController controller;

  const StopLimitTriggerConfigurationScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<StopLimitTriggerConfigurationScreen> createState() =>
      _StopLimitTriggerConfigurationScreenState();
}

class _StopLimitTriggerConfigurationScreenState
    extends State<StopLimitTriggerConfigurationScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color neonRed = Color(0xFFFF3B56);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<StopLimitState>(
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
              'FLUTTER STOP-LIMIT & STOP-MARKET CONDITIONAL',
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
                _buildOrderSideToggle(isBuy),
                const SizedBox(height: 16),
                _buildTriggerTypeSelector(state),
                const SizedBox(height: 16),
                _buildPriceTypeRadios(state),
                const SizedBox(height: 16),
                _buildInputsCard(state),
                const SizedBox(height: 16),
                if (!state.validation.isValid)
                  _buildErrorAlert(state.validation.errorMessage!)
                else if (state.validation.warningMessage != null)
                  _buildWarningAlert(state.validation.warningMessage!),
                const SizedBox(height: 24),
                _buildActionButtons(state, isBuy),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildOrderSideToggle(bool isBuy) {
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
            child: const Text('BUY / LONG', style: TextStyle(fontWeight: FontWeight.bold)),
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
            child: const Text('SELL / SHORT', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ),
      ],
    );
  }

  Widget _buildTriggerTypeSelector(StopLimitState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.white12),
      ),
      child: DropdownButton<TriggerType>(
        value: state.triggerType,
        dropdownColor: surfaceCard,
        isExpanded: true,
        underline: const SizedBox(),
        style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
        icon: const Icon(Icons.arrow_drop_down, color: neonGreen),
        items: TriggerType.values.map((t) {
          return DropdownMenuItem(
            value: t,
            child: Text(t.displayName),
          );
        }).toList(),
        onChanged: (t) => widget.controller.setTriggerType(t!),
      ),
    );
  }

  Widget _buildPriceTypeRadios(StopLimitState state) {
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
          const Text('Trigger Price Reference', style: TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(height: 6),
          Row(
            children: TriggerPriceType.values.map((pt) {
              final isSelected = state.triggerPriceType == pt;
              return Expanded(
                child: InkWell(
                  onTap: () => widget.controller.setPriceType(pt),
                  child: Row(
                    children: [
                      Radio<TriggerPriceType>(
                        value: pt,
                        groupValue: state.triggerPriceType,
                        activeColor: neonGreen,
                        onChanged: (val) => widget.controller.setPriceType(val!),
                      ),
                      Text(pt.shortCode, style: TextStyle(color: isSelected ? Colors.white : textMuted, fontSize: 11, fontWeight: FontWeight.bold)),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildInputsCard(StopLimitState state) {
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
              const Text('Current Reference Price', style: TextStyle(color: textMuted, fontSize: 12)),
              Text(
                '₹${state.referencePrice.toStringAsFixed(2)}',
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
          const Divider(color: Colors.white12, height: 20),
          _buildNumericField(
            label: 'Stop Trigger Price',
            value: state.triggerPrice,
            onChanged: (val) => widget.controller.setTriggerPrice(val),
          ),
          if (state.triggerType.isLimitOrder) ...[
            const SizedBox(height: 12),
            _buildNumericField(
              label: 'Order Limit Price',
              value: state.limitPrice,
              onChanged: (val) => widget.controller.setLimitPrice(val),
            ),
          ],
          const SizedBox(height: 12),
          _buildNumericField(
            label: 'Quantity (${state.symbol.split('/').first})',
            value: state.quantity,
            onChanged: (val) => widget.controller.setQuantity(val),
          ),
          const Divider(color: Colors.white12, height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Estimated Order Value', style: TextStyle(color: textMuted, fontSize: 12)),
              Text(
                '₹${state.estimatedTotalInr.toStringAsFixed(2)}',
                style: const TextStyle(color: neonGreen, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildNumericField({
    required String label,
    required double value,
    required ValueChanged<double> onChanged,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: const TextStyle(color: Colors.white70, fontSize: 12)),
        Row(
          children: [
            IconButton(
              icon: const Icon(Icons.remove, color: textMuted, size: 18),
              onPressed: () => onChanged(value - 50.0),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(color: obsidianBackground, borderRadius: BorderRadius.circular(4)),
              child: Text(
                value.toStringAsFixed(2),
                style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold, fontSize: 13),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.add, color: neonGreen, size: 18),
              onPressed: () => onChanged(value + 50.0),
            ),
          ],
        ),
      ],
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

  Widget _buildWarningAlert(String message) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.orangeAccent.withOpacity(0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.orangeAccent),
      ),
      child: Text(message, style: const TextStyle(color: Colors.orangeAccent, fontSize: 11)),
    );
  }

  Widget _buildActionButtons(StopLimitState state, bool isBuy) {
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
                    final ok = await widget.controller.submitOrder();
                    if (ok && mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(widget.controller.state.toastMessage ?? 'Trigger Order Armed!'),
                          backgroundColor: const Color(0xFF1B382B),
                        ),
                      );
                    }
                  }
                : null,
            child: Text(
              '${isBuy ? 'Buy' : 'Sell'} ${state.triggerType.displayName} (Neon Green)',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
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
            child: const Text('Cancel Setup'),
          ),
        ),
      ],
    );
  }
}
