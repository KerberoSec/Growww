import 'package:flutter/material.dart';
import '../../domain/models/delivery_receipt.dart';
import '../controllers/commodity_redemption_controller.dart';

class DeliveryAcknowledgementScreen extends StatelessWidget {
  final String orderId;
  final CommodityRedemptionController controller;

  const DeliveryAcknowledgementScreen({
    Key? key,
    required this.orderId,
    required this.controller,
  }) : super(key: key);

  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    final receipt = controller.state.deliveryReceipt;

    return Scaffold(
      backgroundColor: obsidianBackground,
      appBar: AppBar(
        backgroundColor: obsidianBackground,
        elevation: 0,
        title: const Text('eNWR Token Burn Receipt', style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold)),
      ),
      body: receipt == null
          ? const Center(child: CircularProgressIndicator(color: neonGreen))
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Center(
                    child: Container(
                      width: 64,
                      height: 64,
                      decoration: const BoxDecoration(
                        color: Color(0xFF1B382B),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.verified_outlined, color: neonGreen, size: 36),
                    ),
                  ),
                  const SizedBox(height: 12),
                  const Center(
                    child: Text(
                      'Physical Bullion Rematerialized',
                      style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                    ),
                  ),
                  const SizedBox(height: 4),
                  const Center(
                    child: Text(
                      'On-chain token units extinguished on Hyperledger Besu',
                      style: TextStyle(color: textMuted, fontSize: 12),
                    ),
                  ),
                  const SizedBox(height: 20),
                  _buildReceiptCard(receipt),
                  const SizedBox(height: 16),
                  _buildBarSerialsCard(receipt),
                  const SizedBox(height: 24),
                  ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: surfaceCard,
                      foregroundColor: Colors.white,
                      side: const BorderSide(color: Colors.white24),
                      padding: const EdgeInsets.symmetric(vertical: 14),
                    ),
                    onPressed: () => Navigator.of(context).popUntil((route) => route.isFirst),
                    child: const Text('Return to Portfolio'),
                  ),
                ],
              ),
            ),
    );
  }

  Widget _buildReceiptCard(DeliveryReceipt receipt) {
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
          const Text('WDRA eNWR Delivery Certificate', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 10),
          _buildDetailRow('Receipt ID', receipt.receiptId),
          _buildDetailRow('Repository', receipt.repositoryName),
          _buildDetailRow('eNWR Number', receipt.enwrNumber),
          _buildDetailRow('Fine Delivered', '${receipt.totalFineGramsDelivered.toStringAsFixed(2)} g'),
          _buildDetailRow('Besu Burn Tx', receipt.onChainBurnTxHash, isMonospace: true),
          _buildDetailRow('SHA-256 Digest', receipt.verificationSignature.substring(0, 20) + '...', isMonospace: true),
        ],
      ),
    );
  }

  Widget _buildBarSerialsCard(DeliveryReceipt receipt) {
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
          const Text('Allocated Physical Bar Serials', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 10),
          for (final bar in receipt.deliveredBars) ...[
            Text('Serial: ${bar.barSerialNumber}', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 12, fontFamily: 'JetBrains Mono')),
            const SizedBox(height: 2),
            Text('Hallmark: ${bar.bisHallmarkNumber} • ${bar.refineryName}', style: const TextStyle(color: textMuted, fontSize: 11)),
            const SizedBox(height: 8),
          ],
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value, {bool isMonospace = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(width: 8),
          Flexible(
            child: Text(
              value,
              textAlign: TextAlign.right,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: Colors.white,
                fontFamily: isMonospace ? 'JetBrains Mono' : null,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
