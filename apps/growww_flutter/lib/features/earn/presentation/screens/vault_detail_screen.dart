import 'package:flutter/material.dart';
import '../../domain/models/earn_enums.dart';
import '../../domain/models/sub_paise_amount.dart';
import '../../domain/models/yield_vault_product.dart';
import '../controllers/apy_calculator_controller.dart';

class VaultDetailScreen extends StatefulWidget {
  final YieldVaultProduct vault;

  const VaultDetailScreen({
    Key? key,
    required this.vault,
  }) : super(key: key);

  @override
  State<VaultDetailScreen> createState() => _VaultDetailScreenState();
}

class _VaultDetailScreenState extends State<VaultDetailScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  late final ApyCalculatorController _calculatorController;

  @override
  void initState() {
    super.initState();
    _calculatorController = ApyCalculatorController(vault: widget.vault);
  }

  @override
  void dispose() {
    _calculatorController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<ApyCalculatorState>(
      stream: _calculatorController.stream,
      initialData: _calculatorController.state,
      builder: (context, snapshot) {
        final calcState = snapshot.data ?? _calculatorController.state;
        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: Text(widget.vault.assetSymbol, style: const TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold)),
          ),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _buildHeaderCard(),
                const SizedBox(height: 16),
                _buildBondDetailsCard(),
                const SizedBox(height: 16),
                _buildApyCalculatorCard(calcState),
                const SizedBox(height: 16),
                _buildStatutoryNoticeCard(),
                const SizedBox(height: 24),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: neonGreen,
                    foregroundColor: Colors.black,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  onPressed: () => _showStakeModal(context, calcState),
                  child: const Text('Stake Now (Gasless ERC-4337)', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildHeaderCard() {
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
              Text(widget.vault.displayName, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 15)),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: neonGreen.withOpacity(0.15), borderRadius: BorderRadius.circular(4)),
                child: Text('${widget.vault.baseApy.toStringAsFixed(2)}% Base APY', style: const TextStyle(color: neonGreen, fontSize: 11, fontWeight: FontWeight.bold, fontFamily: 'JetBrains Mono')),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text('${widget.vault.issuerName} • ISIN: ${widget.vault.isin}', style: const TextStyle(color: textMuted, fontSize: 11)),
        ],
      ),
    );
  }

  Widget _buildBondDetailsCard() {
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
          const Text('Bond & Custody Invariants', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 10),
          _buildParamRow('Depository Rating', widget.vault.creditRating),
          _buildParamRow('Modified Duration', '${widget.vault.modifiedDurationYears} Years'),
          _buildParamRow('Proof-of-Reserve Hash', widget.vault.proofOfReserveRootHash.substring(0, 16) + '...', isMonospace: true),
          _buildParamRow('TVL Pool Capacity', '${widget.vault.totalValueLocked.formatInr()} / ${widget.vault.maxPoolCapacity.formatInr()}'),
        ],
      ),
    );
  }

  Widget _buildApyCalculatorCard(ApyCalculatorState state) {
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
          const Text('Interactive Return Calculator', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13)),
          const SizedBox(height: 12),
          // Principal quick chips
          Row(
            children: [5000, 25000, 100000].map((amt) {
              final subPaise = SubPaiseAmount.fromInr(amt.toDouble());
              final isSelected = state.principal == subPaise;
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: ChoiceChip(
                  label: Text('₹$amt', style: TextStyle(color: isSelected ? Colors.black : Colors.white, fontSize: 11)),
                  selected: isSelected,
                  selectedColor: neonGreen,
                  backgroundColor: obsidianBackground,
                  onSelected: (_) => _calculatorController.setPrincipal(subPaise),
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
          // Tenor chips
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: TenorType.values.map((t) {
                final isSelected = state.tenor == t;
                return Padding(
                  padding: const EdgeInsets.only(right: 6),
                  child: ChoiceChip(
                    label: Text(t.displayName, style: TextStyle(color: isSelected ? Colors.black : Colors.white, fontSize: 11)),
                    selected: isSelected,
                    selectedColor: neonGreen,
                    backgroundColor: obsidianBackground,
                    onSelected: (_) => _calculatorController.setTenor(t),
                  ),
                );
              }).toList(),
            ),
          ),
          const SizedBox(height: 16),
          // Returns summary
          _buildParamRow('Effective APY', '${state.result.effectiveApy.toStringAsFixed(2)}%', color: neonGreen),
          _buildParamRow('Projected Daily Yield', state.result.projectedDailyYield.formatInr(showSubPaise: true), isMonospace: true),
          _buildParamRow('Projected Monthly Yield', state.result.projectedMonthlyYield.formatInr(), isMonospace: true),
          _buildParamRow('Total Maturity Yield', state.result.projectedTotalYield.formatInr(), isMonospace: true),
          _buildParamRow('Section 194A TDS', '-${state.result.estimatedTdsDeduction.formatInr()}', isMonospace: true, color: Colors.orangeAccent),
          const Divider(color: Colors.white12, height: 16),
          _buildParamRow('Net Maturity Amount', state.result.netMaturityAmount.formatInr(), isMonospace: true, color: Colors.white, isBold: true),
        ],
      ),
    );
  }

  Widget _buildStatutoryNoticeCard() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFF1E1A14),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.orange.withOpacity(0.3)),
      ),
      child: const Text(
        'Statutory SEBI & IT Dept Disclosure: Yields reflect underlying bond coupon rates and market YTM. Subject to Section 194A TDS withholding (10%) on annual interest exceeding ₹40,000.',
        style: TextStyle(color: Colors.orangeAccent, fontSize: 10),
      ),
    );
  }

  Widget _buildParamRow(String label, String value, {bool isMonospace = false, Color? color, bool isBold = false}) {
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
              fontFamily: isMonospace ? 'JetBrains Mono' : null,
              fontSize: 11,
              fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  void _showStakeModal(BuildContext context, ApyCalculatorState state) {
    showModalBottomSheet(
      context: context,
      backgroundColor: surfaceCard,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(16))),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text('Confirm On-Chain Staking Deposit', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 15)),
            const SizedBox(height: 8),
            Text('Principal: ${state.principal.formatInr()} • Tenor: ${state.tenor.displayName}', style: const TextStyle(color: textMuted, fontSize: 12)),
            const SizedBox(height: 20),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(backgroundColor: neonGreen, foregroundColor: Colors.black, padding: const EdgeInsets.symmetric(vertical: 12)),
              icon: const Icon(Icons.fingerprint),
              label: const Text('Biometric Signature Authorization', style: TextStyle(fontWeight: FontWeight.bold)),
              onPressed: () {
                Navigator.of(ctx).pop();
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Staking transaction sponsored by Growww Paymaster and confirmed!'),
                    backgroundColor: Color(0xFF1B382B),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
