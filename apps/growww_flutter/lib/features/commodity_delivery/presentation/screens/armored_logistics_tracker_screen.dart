import 'package:flutter/material.dart';
import '../../domain/models/commodity_enums.dart';
import '../../domain/models/logistics_tracking_event.dart';
import '../controllers/commodity_redemption_controller.dart';
import 'delivery_acknowledgement_screen.dart';

class ArmoredLogisticsTrackerScreen extends StatefulWidget {
  final String orderId;
  final CommodityRedemptionController controller;

  const ArmoredLogisticsTrackerScreen({
    Key? key,
    required this.orderId,
    required this.controller,
  }) : super(key: key);

  @override
  State<ArmoredLogisticsTrackerScreen> createState() =>
      _ArmoredLogisticsTrackerScreenState();
}

class _ArmoredLogisticsTrackerScreenState
    extends State<ArmoredLogisticsTrackerScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  void initState() {
    super.initState();
    widget.controller.loadTrackingInfo(widget.orderId);
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<RedemptionWizardState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: const Text('Armored Carrier Telemetry', style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold)),
          ),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _buildCarrierBanner(state),
                const SizedBox(height: 16),
                _buildOtpCard(state),
                const SizedBox(height: 16),
                _buildTimeline(state.trackingEvents),
                const SizedBox(height: 24),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: neonGreen,
                    foregroundColor: Colors.black,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  onPressed: () async {
                    await widget.controller.fetchReceipt(widget.orderId);
                    if (mounted) {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => DeliveryAcknowledgementScreen(
                            orderId: widget.orderId,
                            controller: widget.controller,
                          ),
                        ),
                      );
                    }
                  },
                  child: const Text('View Cryptographic Delivery Receipt', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildCarrierBanner(RedemptionWizardState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Sequel Secure Armored Logistics', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14)),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: neonGreen.withOpacity(0.15), borderRadius: BorderRadius.circular(4)),
                child: const Text('100% Insured', style: TextStyle(color: neonGreen, fontSize: 10, fontWeight: FontWeight.bold)),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text('Order ID: ${widget.orderId}', style: const TextStyle(color: textMuted, fontSize: 11, fontFamily: 'JetBrains Mono')),
          const SizedBox(height: 2),
          const Text('Security Seal #: SEAL-SEC-994102', style: TextStyle(color: Colors.white70, fontSize: 11, fontFamily: 'JetBrains Mono')),
        ],
      ),
    );
  }

  Widget _buildOtpCard(RedemptionWizardState state) {
    final otp = state.currentOtp ?? '748291';
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF1B382B),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: neonGreen.withOpacity(0.4)),
      ),
      child: Column(
        children: [
          const Text('Time-Locked Handover OTP', style: TextStyle(color: neonGreen, fontWeight: FontWeight.bold, fontSize: 12)),
          const SizedBox(height: 6),
          Text(
            otp,
            style: const TextStyle(
              color: Colors.white,
              fontFamily: 'JetBrains Mono',
              fontSize: 28,
              letterSpacing: 6,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          const Text('Share this code with the armored courier upon seal verification', style: TextStyle(color: Colors.white70, fontSize: 11)),
        ],
      ),
    );
  }

  Widget _buildTimeline(List<LogisticsTrackingEvent> events) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Chain-of-Custody Checkpoints', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
        const SizedBox(height: 12),
        for (int i = 0; i < events.length; i++) ...[
          _buildTimelineItem(events[i], isLast: i == events.length - 1),
        ],
      ],
    );
  }

  Widget _buildTimelineItem(LogisticsTrackingEvent event, {required bool isLast}) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Column(
          children: [
            Container(
              width: 12,
              height: 12,
              decoration: const BoxDecoration(
                shape: BoxShape.circle,
                color: neonGreen,
              ),
            ),
            if (!isLast)
              Container(
                width: 2,
                height: 40,
                color: Colors.white24,
              ),
          ],
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                event.status.displayName,
                style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 12),
              ),
              const SizedBox(height: 2),
              Text(event.statusDescription, style: const TextStyle(color: textMuted, fontSize: 11)),
              const SizedBox(height: 2),
              Text('${event.locationCity} • ${event.timestamp.hour}:${event.timestamp.minute.toString().padLeft(2, '0')}',
                  style: const TextStyle(color: Colors.white54, fontSize: 10)),
              const SizedBox(height: 12),
            ],
          ),
        ),
      ],
    );
  }
}
