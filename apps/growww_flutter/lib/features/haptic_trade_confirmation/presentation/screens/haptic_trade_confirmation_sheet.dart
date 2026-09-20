import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../domain/models/trade_confirmation_models.dart';
import '../../domain/services/haptic_trade_confirmation_service.dart';
import '../controllers/haptic_trade_confirmation_controller.dart';

class FlutterHapticDriver implements HapticDriver {
  @override
  void trigger(HapticPattern pattern) {
    switch (pattern) {
      case HapticPattern.selectionClick:
        HapticFeedback.selectionClick();
        break;
      case HapticPattern.lightTap:
        HapticFeedback.lightImpact();
        break;
      case HapticPattern.mediumImpact:
        HapticFeedback.mediumImpact();
        break;
      case HapticPattern.heavyPulse:
        HapticFeedback.heavyImpact();
        break;
      case HapticPattern.warningBuzz:
        HapticFeedback.vibrate();
        break;
    }
  }
}

class HapticTradeConfirmationSheet extends StatefulWidget {
  final TradeConfirmationIntent intent;
  final Function(String signature) onTradeSuccess;

  const HapticTradeConfirmationSheet({
    super.key,
    required this.intent,
    required this.onTradeSuccess,
  });

  @override
  State<HapticTradeConfirmationSheet> createState() => _HapticTradeConfirmationSheetState();
}

class _HapticTradeConfirmationSheetState extends State<HapticTradeConfirmationSheet> {
  late HapticTradeConfirmationController _controller;

  @override
  void initState() {
    super.initState();
    final service = HapticTradeConfirmationService(
      hapticDriver: FlutterHapticDriver(),
    );
    _controller = HapticTradeConfirmationController(
      service: service,
      intent: widget.intent,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isBuy = widget.intent.side.isBuy;
    final primaryColor = isBuy ? Colors.green : Colors.redAccent;

    return StreamBuilder<HapticTradeConfirmationState>(
      stream: _controller.stream,
      initialData: _controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data!;

        return Container(
          padding: const EdgeInsets.all(24),
          decoration: const BoxDecoration(
            color: Color(0xFF1E222D),
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[600],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Text(
                'Confirm ${widget.intent.side.name.toUpperCase()} Order',
                style: theme.textTheme.titleLarge?.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Instrument', style: TextStyle(color: Colors.grey[400])),
                  Text(widget.intent.symbol, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Quantity', style: TextStyle(color: Colors.grey[400])),
                  Text('${widget.intent.quantity}', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Price', style: TextStyle(color: Colors.grey[400])),
                  Text('₹${widget.intent.price.toStringAsFixed(2)}', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Total Value', style: TextStyle(color: Colors.grey[400])),
                  Text('₹${widget.intent.notionalValue.toStringAsFixed(2)}', style: TextStyle(color: primaryColor, fontWeight: FontWeight.bold)),
                ],
              ),
              const SizedBox(height: 24),
              // Swipe to Trade Slider Area
              GestureDetector(
                onHorizontalDragUpdate: (details) {
                  final renderBox = context.findRenderObject() as RenderBox?;
                  final width = renderBox?.size.width ?? 300.0;
                  final deltaProgress = details.delta.dx / (width - 60);
                  _controller.updateSliderProgress(_controller.state.progress + deltaProgress);
                },
                onHorizontalDragEnd: (details) async {
                  if (_controller.state.progress >= 0.85) {
                    final success = await _controller.executeTradeConfirmation();
                    if (success && _controller.state.signature != null) {
                      widget.onTradeSuccess(_controller.state.signature!);
                    }
                  } else {
                    _controller.resetSlider();
                  }
                },
                child: Container(
                  height: 56,
                  decoration: BoxDecoration(
                    color: const Color(0xFF2A2E39),
                    borderRadius: BorderRadius.circular(28),
                    border: Border.all(
                      color: primaryColor.withOpacity(0.5),
                    ),
                  ),
                  child: Stack(
                    children: [
                      Center(
                        child: Text(
                          state.state == SwipeConfirmationState.submitted
                              ? 'ORDER SUBMITTED'
                              : 'Slide to Confirm ${widget.intent.side.name.toUpperCase()}',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.8),
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      FractionallySizedBox(
                        alignment: Alignment.centerLeft,
                        widthFactor: state.progress.clamp(0.15, 1.0),
                        child: Container(
                          decoration: BoxDecoration(
                            color: primaryColor.withOpacity(0.4),
                            borderRadius: BorderRadius.circular(28),
                          ),
                        ),
                      ),
                      Positioned(
                        left: (state.progress * 240).clamp(0.0, 240.0),
                        top: 4,
                        bottom: 4,
                        child: Container(
                          width: 48,
                          height: 48,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: primaryColor,
                          ),
                          child: const Icon(
                            Icons.chevron_right,
                            color: Colors.white,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        );
      },
    );
  }
}
