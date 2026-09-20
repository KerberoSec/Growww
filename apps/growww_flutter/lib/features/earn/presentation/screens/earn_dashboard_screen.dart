import 'package:flutter/material.dart';
import '../../domain/models/earn_enums.dart';
import '../../domain/models/yield_vault_product.dart';
import '../controllers/earn_dashboard_controller.dart';
import '../widgets/daily_accrued_interest_ticker.dart';
import 'vault_detail_screen.dart';

class EarnDashboardScreen extends StatefulWidget {
  final EarnDashboardController controller;

  const EarnDashboardScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<EarnDashboardScreen> createState() => _EarnDashboardScreenState();
}

class _EarnDashboardScreenState extends State<EarnDashboardScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<EarnDashboardState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: const Text('Fixed-Income Yield & Staking', style: TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold)),
          ),
          body: state.isLoading
              ? const Center(child: CircularProgressIndicator(color: neonGreen))
              : RefreshIndicator(
                  color: neonGreen,
                  onRefresh: () => widget.controller.loadDashboard(),
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        _buildPortfolioSummaryCard(state),
                        const SizedBox(height: 16),
                        _buildCategoryFilterChips(state),
                        const SizedBox(height: 16),
                        _buildVaultDirectory(state),
                      ],
                    ),
                  ),
                ),
        );
      },
    );
  }

  Widget _buildPortfolioSummaryCard(EarnDashboardState state) {
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
              const Text('Active Staked Principal', style: TextStyle(color: textMuted, fontSize: 12)),
              DailyAccruedInterestTicker(amount: state.liveAccruedYield),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            state.totalStakedPrincipal.formatInr(),
            style: const TextStyle(
              color: Colors.white,
              fontFamily: 'JetBrains Mono',
              fontSize: 22,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Total Accrued Yield: ${state.totalAccruedYield.formatInr(showSubPaise: true)}', style: const TextStyle(color: neonGreen, fontSize: 11, fontFamily: 'JetBrains Mono')),
              const Text('Gasless ERC-4337 Sponsored', style: TextStyle(color: textMuted, fontSize: 10)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryFilterChips(EarnDashboardState state) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          _buildFilterChip('All Vaults', state.selectedFilter == null, () => widget.controller.setFilter(null)),
          const SizedBox(width: 8),
          for (final type in YieldAssetType.values) ...[
            _buildFilterChip(type.displayName, state.selectedFilter == type, () => widget.controller.setFilter(type)),
            const SizedBox(width: 8),
          ],
        ],
      ),
    );
  }

  Widget _buildFilterChip(String label, bool isSelected, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? neonGreen : surfaceCard,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: isSelected ? neonGreen : Colors.white12),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.black : Colors.white,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildVaultDirectory(EarnDashboardState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Featured Fixed-Income Vaults', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14)),
        const SizedBox(height: 10),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: state.vaults.length,
          separatorBuilder: (_, __) => const SizedBox(height: 10),
          itemBuilder: (context, index) {
            final vault = state.vaults[index];
            return _buildVaultTile(vault);
          },
        ),
      ],
    );
  }

  Widget _buildVaultTile(YieldVaultProduct vault) {
    return InkWell(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => VaultDetailScreen(vault: vault),
          ),
        );
      },
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(14),
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
                Expanded(
                  child: Text(
                    vault.displayName,
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: neonGreen.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '${vault.baseApy.toStringAsFixed(2)}% APY',
                    style: const TextStyle(color: neonGreen, fontWeight: FontWeight.bold, fontSize: 12, fontFamily: 'JetBrains Mono'),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Text('${vault.issuerName} • ${vault.creditRating}', style: const TextStyle(color: textMuted, fontSize: 11)),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('TVL: ${vault.totalValueLocked.formatInr()}', style: const TextStyle(color: Colors.white70, fontSize: 11, fontFamily: 'JetBrains Mono')),
                Text('Min Stake: ${vault.minStakeAmount.formatInr()}', style: const TextStyle(color: textMuted, fontSize: 11)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
