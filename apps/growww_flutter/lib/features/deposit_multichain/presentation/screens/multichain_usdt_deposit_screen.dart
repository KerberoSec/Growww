import 'package:flutter/material.dart';
import '../../domain/models/crypto_network_info.dart';
import '../../domain/models/deposit_enums.dart';
import '../controllers/multichain_deposit_controller.dart';

class MultichainUsdtDepositScreen extends StatefulWidget {
  final MultichainDepositController controller;

  const MultichainUsdtDepositScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<MultichainUsdtDepositScreen> createState() =>
      _MultichainUsdtDepositScreenState();
}

class _MultichainUsdtDepositScreenState
    extends State<MultichainUsdtDepositScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<MultichainDepositState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;

        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: const Text(
              'FLUTTER MULTICHAIN USDT DEPOSIT MODAL WITH NE',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.0,
              ),
            ),
          ),
          body: state.isLoading
              ? const Center(child: CircularProgressIndicator(color: neonGreen))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildNetworkSelectorCard(state),
                      const SizedBox(height: 16),
                      if (state.depositAddress != null) ...[
                        _buildQrAndAddressCard(state),
                        const SizedBox(height: 16),
                        _buildNetworkInvariantsCard(state.selectedNetwork!),
                      ],
                      const SizedBox(height: 16),
                      _buildComplianceNotice(),
                      const SizedBox(height: 24),
                      _buildActionButtons(state),
                    ],
                  ),
                ),
        );
      },
    );
  }

  Widget _buildNetworkSelectorCard(MultichainDepositState state) {
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
          const Text('Select Deposit Network', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 4),
          const Text('Ensure destination network matches withdrawal platform.', style: TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(height: 12),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: state.networks.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              final net = state.networks[index];
              final isSelected = state.selectedNetwork?.networkId == net.networkId;
              return InkWell(
                onTap: () => widget.controller.selectNetwork(net),
                borderRadius: BorderRadius.circular(6),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  decoration: BoxDecoration(
                    color: isSelected ? neonGreen.withOpacity(0.12) : obsidianBackground,
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: isSelected ? neonGreen : Colors.white12),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            net.displayName,
                            style: TextStyle(color: isSelected ? neonGreen : Colors.white, fontWeight: FontWeight.bold, fontSize: 12),
                          ),
                          Text(
                            'Est. Arrival: ~${net.estimatedArrivalMinutes} min • ${net.requiredConfirmations} Confirmations',
                            style: const TextStyle(color: textMuted, fontSize: 10),
                          ),
                        ],
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(color: Colors.white10, borderRadius: BorderRadius.circular(4)),
                        child: Text(
                          net.networkType.networkTag,
                          style: const TextStyle(color: Colors.white70, fontSize: 10, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildQrAndAddressCard(MultichainDepositState state) {
    final addr = state.depositAddress!.address;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        children: [
          // QR Box Placeholder
          Container(
            width: 140,
            height: 140,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Center(
              child: Icon(Icons.qr_code_2, size: 120, color: Colors.black),
            ),
          ),
          const SizedBox(height: 14),
          const Text('USDT Segregated Deposit Address', style: TextStyle(color: textMuted, fontSize: 11)),
          const SizedBox(height: 6),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            decoration: BoxDecoration(color: obsidianBackground, borderRadius: BorderRadius.circular(4)),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    addr,
                    style: const TextStyle(color: Colors.white, fontFamily: 'JetBrains Mono', fontSize: 11, fontWeight: FontWeight.w600),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.copy, color: neonGreen, size: 18),
                  tooltip: 'Copy Address',
                  onPressed: () {
                    widget.controller.copyAddressToClipboard();
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Address copied to clipboard!'), duration: Duration(seconds: 1)),
                    );
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNetworkInvariantsCard(CryptoNetworkInfo net) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Network Invariants & Limits', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 12)),
          const SizedBox(height: 8),
          _buildInvariantRow('Min Deposit Amount', '${net.minDepositUsdt.toStringAsFixed(2)} USDT'),
          _buildInvariantRow('Deposit Fee', '₹0.00 (Zero Fee Sponsored)', color: neonGreen),
          _buildInvariantRow('Required Confirmations', '${net.requiredConfirmations} Block Confirmations'),
          _buildInvariantRow('Chain Protocol', net.chainName),
        ],
      ),
    );
  }

  Widget _buildInvariantRow(String label, String value, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: textMuted, fontSize: 11)),
          Text(
            value,
            style: TextStyle(color: color ?? Colors.white, fontSize: 11, fontFamily: 'JetBrains Mono', fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }

  Widget _buildComplianceNotice() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFF1E1A14),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.orange.withOpacity(0.3)),
      ),
      child: const Text(
        'Warning: Send only USDT to this deposit address. Sending any other digital asset will result in permanent loss.',
        style: TextStyle(color: Colors.orangeAccent, fontSize: 10),
      ),
    );
  }

  Widget _buildActionButtons(MultichainDepositState state) {
    return Row(
      children: [
        Expanded(
          child: ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: neonGreen,
              foregroundColor: Colors.black,
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
            ),
            onPressed: () {
              widget.controller.copyAddressToClipboard();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('USDT Deposit address copied!'), backgroundColor: Color(0xFF1B382B)),
              );
            },
            child: const Text('Copy Deposit Address (Neon Green)', style: TextStyle(fontWeight: FontWeight.bold)),
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
            child: const Text('Dismiss Modal'),
          ),
        ),
      ],
    );
  }
}
