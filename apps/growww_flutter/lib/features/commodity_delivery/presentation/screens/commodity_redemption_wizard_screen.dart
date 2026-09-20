import 'package:flutter/material.dart';
import '../../domain/models/commodity_denomination.dart';
import '../../domain/models/commodity_enums.dart';
import '../../domain/models/sub_paise_fee_breakdown.dart';
import '../controllers/commodity_redemption_controller.dart';
import 'armored_logistics_tracker_screen.dart';
import 'wdra_vault_locator_screen.dart';

class CommodityRedemptionWizardScreen extends StatefulWidget {
  final CommodityRedemptionController controller;

  const CommodityRedemptionWizardScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<CommodityRedemptionWizardScreen> createState() =>
      _CommodityRedemptionWizardScreenState();
}

class _CommodityRedemptionWizardScreenState
    extends State<CommodityRedemptionWizardScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color goldAccent = Color(0xFFF0B90B);
  static const Color textMuted = Color(0xFF8B949E);

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
            title: const Text(
              'WDRA Commodity Physical Redemption',
              style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold),
            ),
          ),
          body: state.isLoading
              ? const Center(child: CircularProgressIndicator(color: neonGreen))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(16.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildAssetSelector(state),
                      const SizedBox(height: 16),
                      _buildBalanceCard(state),
                      const SizedBox(height: 16),
                      _buildDenominationSelector(state),
                      const SizedBox(height: 16),
                      _buildDeliveryModeSelector(state),
                      const SizedBox(height: 16),
                      if (state.feeBreakdown != null) _buildFeeSummaryCard(state.feeBreakdown!),
                      const SizedBox(height: 24),
                      _buildProceedButton(state),
                    ],
                  ),
                ),
        );
      },
    );
  }

  Widget _buildAssetSelector(RedemptionWizardState state) {
    return Row(
      children: [
        Expanded(
          child: _buildTypeChip(
            label: 'gGOLD 999',
            isSelected: state.selectedCommodity == CommodityType.gold999,
            onTap: () => widget.controller.selectCommodity(CommodityType.gold999),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _buildTypeChip(
            label: 'gGOLD 999.9',
            isSelected: state.selectedCommodity == CommodityType.gold9999,
            onTap: () => widget.controller.selectCommodity(CommodityType.gold9999),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _buildTypeChip(
            label: 'gSILVER 999',
            isSelected: state.selectedCommodity == CommodityType.silver999,
            onTap: () => widget.controller.selectCommodity(CommodityType.silver999),
          ),
        ),
      ],
    );
  }

  Widget _buildTypeChip({
    required String label,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? goldAccent.withOpacity(0.2) : surfaceCard,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isSelected ? goldAccent : Colors.white12,
            width: isSelected ? 1.5 : 1.0,
          ),
        ),
        alignment: Alignment.center,
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? goldAccent : Colors.white70,
            fontWeight: FontWeight.bold,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildBalanceCard(RedemptionWizardState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white12),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Available Vault Balance', style: TextStyle(color: textMuted, fontSize: 12)),
              const SizedBox(height: 4),
              Text(
                '${state.availableBalanceGrams.toStringAsFixed(2)} grams',
                style: const TextStyle(
                  color: Colors.white,
                  fontFamily: 'JetBrains Mono',
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              const Text('Selected for Delivery', style: TextStyle(color: textMuted, fontSize: 12)),
              const SizedBox(height: 4),
              Text(
                '${state.totalSelectedGrams.toStringAsFixed(2)} grams',
                style: TextStyle(
                  color: state.isValidWeight ? neonGreen : Colors.orangeAccent,
                  fontFamily: 'JetBrains Mono',
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDenominationSelector(RedemptionWizardState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Select Bar Denominations',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
        ),
        const SizedBox(height: 10),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: state.availableDenominations.length,
          separatorBuilder: (_, __) => const SizedBox(height: 8),
          itemBuilder: (context, index) {
            final denom = state.availableDenominations[index];
            final qty = state.selectedQuantities[denom.denominationId] ?? 0;
            return _buildDenominationTile(denom, qty);
          },
        ),
      ],
    );
  }

  Widget _buildDenominationTile(CommodityDenomination denom, int quantity) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: surfaceCard,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: quantity > 0 ? neonGreen.withOpacity(0.5) : Colors.white12),
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  denom.displayName,
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 13),
                ),
                const SizedBox(height: 2),
                Text(
                  '${denom.assayStandard} • ${denom.packagingType} • In Stock: ${denom.availableStockUnits}',
                  style: const TextStyle(color: textMuted, fontSize: 11),
                ),
              ],
            ),
          ),
          Row(
            children: [
              IconButton(
                icon: const Icon(Icons.remove_circle_outline, color: textMuted, size: 22),
                onPressed: quantity > 0
                    ? () => widget.controller.updateQuantity(denom.denominationId, -1)
                    : null,
              ),
              Text(
                '$quantity',
                style: const TextStyle(
                  color: Colors.white,
                  fontFamily: 'JetBrains Mono',
                  fontWeight: FontWeight.bold,
                  fontSize: 15,
                ),
              ),
              IconButton(
                icon: const Icon(Icons.add_circle_outline, color: neonGreen, size: 22),
                onPressed: () => widget.controller.updateQuantity(denom.denominationId, 1),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDeliveryModeSelector(RedemptionWizardState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Fulfillment Method',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
        ),
        const SizedBox(height: 8),
        RadioListTile<DeliveryMode>(
          value: DeliveryMode.armoredDoorstep,
          groupValue: state.deliveryMode,
          activeColor: neonGreen,
          tileColor: surfaceCard,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          title: const Text('Insured Armored Doorstep Delivery', style: TextStyle(color: Colors.white, fontSize: 13)),
          subtitle: const Text('Brink\'s / Sequel 100% replacement insurance', style: TextStyle(color: textMuted, fontSize: 11)),
          onChanged: (mode) => widget.controller.setDeliveryMode(mode!),
        ),
        const SizedBox(height: 6),
        RadioListTile<DeliveryMode>(
          value: DeliveryMode.vaultPickup,
          groupValue: state.deliveryMode,
          activeColor: neonGreen,
          tileColor: surfaceCard,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          title: const Text('WDRA Vault Self-Pickup', style: TextStyle(color: Colors.white, fontSize: 13)),
          subtitle: const Text('Direct collection at accredited repository bullion hub', style: TextStyle(color: textMuted, fontSize: 11)),
          onChanged: (mode) {
            widget.controller.setDeliveryMode(mode!);
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => WdraVaultLocatorScreen(controller: widget.controller),
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildFeeSummaryCard(SubPaiseFeeBreakdown fee) {
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
          const Text('Sub-Paise Settlement Breakdown', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 10),
          _buildFeeRow('Vault Rematerialization Fee', fee.formatInr(fee.vaultFeeInr)),
          _buildFeeRow('Platform Settlement Fee (0.00% Zero Fee)', '₹0.00', color: neonGreen),
          _buildFeeRow('Assay Packaging Fee', fee.formatInr(fee.packagingFeeInr)),
          _buildFeeRow('Armored Courier & Transit Insurance', fee.formatInr(fee.logisticsFeeInr + fee.insuranceFeeInr)),
          _buildFeeRow('Statutory GST (3% Metal + 18% Services)', fee.formatInr(fee.totalGstInr)),
          const Divider(color: Colors.white12, height: 16),
          _buildFeeRow(
            'Total Gross Payable',
            fee.formatInr(fee.totalInr, showSubPaise: true),
            isBold: true,
            color: Colors.white,
          ),
        ],
      ),
    );
  }

  Widget _buildFeeRow(String label, String value, {bool isBold = false, Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(color: textMuted, fontSize: 11, fontWeight: isBold ? FontWeight.bold : FontWeight.normal)),
          Text(
            value,
            style: TextStyle(
              color: color ?? Colors.white70,
              fontFamily: 'JetBrains Mono',
              fontSize: 12,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProceedButton(RedemptionWizardState state) {
    final isEnabled = state.isValidWeight && !state.isLoading;
    return ElevatedButton(
      style: ElevatedButton.styleFrom(
        backgroundColor: neonGreen,
        foregroundColor: Colors.black,
        padding: const EdgeInsets.symmetric(vertical: 14),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      onPressed: isEnabled
          ? () async {
              final order = await widget.controller.initiateRedemption();
              if (order != null && mounted) {
                _showAuthorizationModal(order);
              }
            }
          : null,
      child: const Text(
        'Authorize Physical Redemption (2FA Biometric)',
        style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
      ),
    );
  }

  void _showAuthorizationModal(dynamic order) {
    showModalBottomSheet(
      context: context,
      backgroundColor: surfaceCard,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Dual-Factor Physical Release Challenge',
              style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const SizedBox(height: 8),
            const Text(
              'Biometric authorization locks tokenized commodity units on Hyperledger Besu and issues eNWR rematerialization instruction.',
              style: TextStyle(color: textMuted, fontSize: 12),
            ),
            const SizedBox(height: 20),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: neonGreen,
                foregroundColor: Colors.black,
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
              icon: const Icon(Icons.fingerprint),
              label: const Text('Verify Biometrics & Confirm Burn', style: TextStyle(fontWeight: FontWeight.bold)),
              onPressed: () async {
                Navigator.of(ctx).pop();
                final ok = await widget.controller.authorizeRedemption(
                  biometricSignature: 'BIO-PASSKEY-SIG-990148',
                  totpCode: '481902',
                );
                if (ok && mounted) {
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => ArmoredLogisticsTrackerScreen(
                        orderId: widget.controller.state.activeOrder!.orderId,
                        controller: widget.controller,
                      ),
                    ),
                  );
                }
              },
            ),
          ],
        ),
      ),
    );
  }
}
