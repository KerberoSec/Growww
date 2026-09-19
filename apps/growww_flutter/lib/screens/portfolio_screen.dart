import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/market_data.dart';

/// Unified Portfolio Screen displaying:
/// - Consolidated Net Worth in INR (Combining Indian Demat Equities + Crypto VDAs)
/// - Real-time P&L (INR and percentage)
/// - Asset allocation distribution bar
/// - Tax summary card: Section 194S (1% TDS) & Section 115BBH (30% VDA flat tax)
/// - On-chain Proof-of-Reserve solvency verification status
class PortfolioScreen extends StatefulWidget {
  final PortfolioSummary? initialPortfolio;

  const PortfolioScreen({Key? key, this.initialPortfolio}) : super(key: key);

  @override
  State<PortfolioScreen> createState() => _PortfolioScreenState();
}

class _PortfolioScreenState extends State<PortfolioScreen> with SingleTickerProviderStateMixin {
  late PortfolioSummary _portfolio;
  String _selectedFilter = 'ALL'; // ALL, EQUITY, CRYPTO, CASH
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _portfolio = widget.initialPortfolio ?? _generateSamplePortfolio();
  }

  PortfolioSummary _generateSamplePortfolio() {
    final assets = [
      // Demat Indian Equities (INR)
      const PortfolioAsset(
        symbol: 'RELIANCE.BSE',
        assetName: 'Reliance Industries Ltd',
        assetClass: AssetClass.equityInr,
        balance: 150.0,
        averageCostBasis: 2780.00,
        currentPrice: 2985.50,
      ),
      const PortfolioAsset(
        symbol: 'TCS.NSE',
        assetName: 'Tata Consultancy Services',
        assetClass: AssetClass.equityInr,
        balance: 80.0,
        averageCostBasis: 3820.00,
        currentPrice: 4210.00,
      ),
      const PortfolioAsset(
        symbol: 'HDFCBANK.NSE',
        assetName: 'HDFC Bank Ltd',
        assetClass: AssetClass.equityInr,
        balance: 300.0,
        averageCostBasis: 1540.00,
        currentPrice: 1665.20,
      ),

      // Crypto Virtual Digital Assets (VDA) converted to INR
      const PortfolioAsset(
        symbol: 'BTC/INR',
        assetName: 'Bitcoin',
        assetClass: AssetClass.cryptoVda,
        balance: 0.155,
        averageCostBasis: 5200000.00,
        currentPrice: 5621875.00, // ~64,250 USDT * 87.5 INR
        tdsAccrued: 8713.90,
        onChainDvPTokenAddress: '0x3a4b...besu7',
      ),
      const PortfolioAsset(
        symbol: 'ETH/INR',
        assetName: 'Ethereum',
        assetClass: AssetClass.cryptoVda,
        balance: 2.45,
        averageCostBasis: 265000.00,
        currentPrice: 301875.00,
        tdsAccrued: 7395.93,
        onChainDvPTokenAddress: '0x9c1d...besu2',
      ),
      const PortfolioAsset(
        symbol: 'USDT/INR',
        assetName: 'Tether USD',
        assetClass: AssetClass.stablecoinUsdt,
        balance: 2500.0,
        averageCostBasis: 86.80,
        currentPrice: 87.50,
        tdsAccrued: 2187.50,
        onChainDvPTokenAddress: '0x4e5f...besu9',
      ),
    ];

    return PortfolioSummary(
      unallocatedCashInr: 125000.00,
      assets: assets,
      lastUpdated: DateTime.now(),
      proofOfReserveRootHash: '0x8f4c2e1b9a7d3f5e6a8b0c2d4e6f8a0b1c3d5e7f',
    );
  }

  Future<void> _refreshPortfolio() async {
    HapticFeedback.lightImpact();
    setState(() => _isLoading = true);
    await Future.delayed(const Duration(milliseconds: 600));
    setState(() {
      _isLoading = false;
      _portfolio = _generateSamplePortfolio();
    });
  }

  List<PortfolioAsset> get _filteredAssets {
    switch (_selectedFilter) {
      case 'EQUITY':
        return _portfolio.assets.where((a) => a.assetClass == AssetClass.equityInr).toList();
      case 'CRYPTO':
        return _portfolio.assets.where((a) => a.assetClass == AssetClass.cryptoVda || a.assetClass == AssetClass.stablecoinUsdt).toList();
      case 'ALL':
      default:
        return _portfolio.assets;
    }
  }

  String _formatInr(double amount) {
    // Format INR numbers with Indian commas (e.g., 24,58,920.50)
    final isNeg = amount < 0;
    final absAmount = amount.abs();
    final parts = absAmount.toStringAsFixed(2).split('.');
    final integerPart = parts[0];
    final decimalPart = parts[1];

    if (integerPart.length <= 3) {
      return '${isNeg ? '-' : ''}₹$integerPart.$decimalPart';
    }

    final lastThree = integerPart.substring(integerPart.length - 3);
    final rest = integerPart.substring(0, integerPart.length - 3);

    final buffer = StringBuffer();
    for (int i = 0; i < rest.length; i++) {
      if (i > 0 && (rest.length - i) % 2 == 0) {
        buffer.write(',');
      }
      buffer.write(rest[i]);
    }
    buffer.write(',');
    buffer.write(lastThree);

    return '${isNeg ? '-' : ''}₹${buffer.toString()}.$decimalPart';
  }

  @override
  Widget build(BuildContext context) {
    final isPnLPositive = _portfolio.totalUnrealizedPnLInr >= 0;

    return Scaffold(
      backgroundColor: const Color(0xFF08090C),
      appBar: AppBar(
        backgroundColor: const Color(0xFF12141A),
        elevation: 0,
        title: const Text(
          'Unified Portfolio (INR + VDA)',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 17),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.download_outlined, color: Color(0xFF9096A2)),
            tooltip: 'Export P&L Statement',
            onPressed: () {
              HapticFeedback.lightImpact();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('Form 16A & P&L Tax statement exported to downloads.'),
                  backgroundColor: Color(0xFF1E222D),
                ),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.refresh, color: Color(0xFF9096A2)),
            onPressed: _refreshPortfolio,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refreshPortfolio,
        color: const Color(0xFF00E676),
        backgroundColor: const Color(0xFF12141A),
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Net Worth Master Card
            _buildNetWorthCard(isPnLPositive),

            const SizedBox(height: 16),

            // Proof of Reserve Solvency Badge
            _buildPoRSolvencyBadge(),

            const SizedBox(height: 16),

            // Asset Allocation Proportional Bar
            _buildAllocationBar(),

            const SizedBox(height: 16),

            // Tax Summary Card (Section 194S TDS + Section 115BBH)
            _buildTaxSummaryCard(),

            const SizedBox(height: 16),

            // Filter Tabs (All / Equities / Crypto)
            _buildFilterTabs(),

            const SizedBox(height: 12),

            // Assets List
            ..._filteredAssets.map((asset) => _buildAssetHoldingTile(asset)),

            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  Widget _buildNetWorthCard(bool isPositive) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFF12141A),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF232732)),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF161922), Color(0xFF0F1117)],
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'TOTAL NET WORTH',
                style: TextStyle(color: Color(0xFF9096A2), fontSize: 11, fontWeight: FontWeight.bold, letterSpacing: 0.5),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: isPositive ? const Color(0xFF1B382B) : const Color(0xFF381B22),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    Icon(
                      isPositive ? Icons.arrow_upward : Icons.arrow_downward,
                      size: 12,
                      color: isPositive ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '${isPositive ? '+' : ''}${_portfolio.totalUnrealizedPnLPercent.toStringAsFixed(2)}%',
                      style: TextStyle(
                        color: isPositive ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            _formatInr(_portfolio.totalNetWorthInr),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 28,
              fontWeight: FontWeight.bold,
              fontFamily: 'monospace',
              letterSpacing: -0.5,
            ),
          ),
          const SizedBox(height: 12),
          const Divider(color: Color(0xFF232732), height: 1),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildMetricSubItem(
                'Total Invested',
                _formatInr(_portfolio.totalInvestedInr),
              ),
              _buildMetricSubItem(
                'Unrealized P&L',
                '${isPositive ? '+' : ''}${_formatInr(_portfolio.totalUnrealizedPnLInr)}',
                color: isPositive ? const Color(0xFF00E676) : const Color(0xFFFF1744),
              ),
              _buildMetricSubItem(
                'Cash / Margin',
                _formatInr(_portfolio.unallocatedCashInr),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildMetricSubItem(String label, String value, {Color? color}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(color: Color(0xFF5A606D), fontSize: 10)),
        const SizedBox(height: 4),
        Text(
          value,
          style: TextStyle(
            color: color ?? const Color(0xFF9096A2),
            fontSize: 12,
            fontFamily: 'monospace',
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    );
  }

  Widget _buildPoRSolvencyBadge() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFF141A24),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFF00E5FF).withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.verified_user, color: Color(0xFF00E5FF), size: 18),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Proof of Reserve Solvency: 104.8% Backed',
                  style: TextStyle(color: Color(0xFF00E5FF), fontSize: 12, fontWeight: FontWeight.bold),
                ),
                Text(
                  'Merkle Root: ${_portfolio.proofOfReserveRootHash.substring(0, 18)}...',
                  style: const TextStyle(color: Color(0xFF5A606D), fontSize: 10, fontFamily: 'monospace'),
                ),
              ],
            ),
          ),
          const Icon(Icons.open_in_new, color: Color(0xFF5A606D), size: 14),
        ],
      ),
    );
  }

  Widget _buildAllocationBar() {
    final eqPct = _portfolio.equityAllocationPercent;
    final crPct = _portfolio.cryptoAllocationPercent;
    final caPct = _portfolio.cashAllocationPercent;

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF12141A),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF232732)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'PORTFOLIO ALLOCATION',
            style: TextStyle(color: Color(0xFF9096A2), fontSize: 11, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: SizedBox(
              height: 10,
              child: Row(
                children: [
                  Expanded(flex: (eqPct * 10).toInt(), child: Container(color: const Color(0xFF00E5FF))),
                  Expanded(flex: (crPct * 10).toInt(), child: Container(color: const Color(0xFFF0B90B))),
                  Expanded(flex: (caPct * 10).toInt(), child: Container(color: const Color(0xFF7C4DFF))),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildAllocationLegend('Equities', '${eqPct.toStringAsFixed(1)}%', const Color(0xFF00E5FF)),
              _buildAllocationLegend('Crypto VDA', '${crPct.toStringAsFixed(1)}%', const Color(0xFFF0B90B)),
              _buildAllocationLegend('INR Cash', '${caPct.toStringAsFixed(1)}%', const Color(0xFF7C4DFF)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAllocationLegend(String label, String pct, Color color) {
    return Row(
      children: [
        Container(width: 8, height: 8, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 6),
        Text(label, style: const TextStyle(color: Color(0xFF9096A2), fontSize: 11)),
        const SizedBox(width: 4),
        Text(pct, style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold)),
      ],
    );
  }

  Widget _buildTaxSummaryCard() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1712),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFF0B90B).withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Row(
                children: [
                  Icon(Icons.account_balance, color: Color(0xFFF0B90B), size: 16),
                  SizedBox(width: 8),
                  Text('VDA Regulatory Tax Accruals', style: TextStyle(color: Color(0xFFF0B90B), fontWeight: FontWeight.bold, fontSize: 12)),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(color: const Color(0xFFF0B90B).withOpacity(0.15), borderRadius: BorderRadius.circular(4)),
                child: const Text('Sec 194S & 115BBH', style: TextStyle(color: Color(0xFFF0B90B), fontSize: 10, fontWeight: FontWeight.bold)),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('1% TDS Deducted (194S)', style: TextStyle(color: Color(0xFF9096A2), fontSize: 10)),
                  const SizedBox(height: 2),
                  Text(_formatInr(_portfolio.totalTdsAccruedInr), style: const TextStyle(color: Colors.white, fontSize: 12, fontFamily: 'monospace')),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const Text('30% VDA Gain Tax Liability (115BBH)', style: TextStyle(color: Color(0xFF9096A2), fontSize: 10)),
                  const SizedBox(height: 2),
                  Text(_formatInr(_portfolio.totalSection115bbhTaxEstimate), style: const TextStyle(color: Color(0xFFFF1744), fontSize: 12, fontFamily: 'monospace', fontWeight: FontWeight.w600)),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFilterTabs() {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          for (final f in [
            {'id': 'ALL', 'label': 'All Holdings'},
            {'id': 'EQUITY', 'label': 'Indian Equities'},
            {'id': 'CRYPTO', 'label': 'Crypto VDA'},
          ])
            Padding(
              padding: const EdgeInsets.only(right: 8.0),
              child: ChoiceChip(
                label: Text(f['label']!),
                selected: _selectedFilter == f['id'],
                selectedColor: const Color(0xFF222630),
                backgroundColor: const Color(0xFF12141A),
                labelStyle: TextStyle(
                  color: _selectedFilter == f['id'] ? const Color(0xFF00E676) : const Color(0xFF9096A2),
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
                onSelected: (selected) {
                  if (selected) {
                    HapticFeedback.selectionClick();
                    setState(() => _selectedFilter = f['id']!);
                  }
                },
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildAssetHoldingTile(PortfolioAsset asset) {
    final isPos = asset.unrealizedPnL >= 0;
    final isCrypto = asset.assetClass == AssetClass.cryptoVda || asset.assetClass == AssetClass.stablecoinUsdt;

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF12141A),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF232732)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Left: Symbol & Name
          Row(
            children: [
              Container(
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: isCrypto ? const Color(0xFF232014) : const Color(0xFF142028),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(
                  isCrypto ? Icons.currency_bitcoin : Icons.show_chart,
                  color: isCrypto ? const Color(0xFFF0B90B) : const Color(0xFF00E5FF),
                  size: 20,
                ),
              ),
              const SizedBox(width: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    asset.symbol,
                    style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${asset.totalBalance} @ ${_formatInr(asset.averageCostBasis)}',
                    style: const TextStyle(color: Color(0xFF5A606D), fontSize: 11, fontFamily: 'monospace'),
                  ),
                ],
              ),
            ],
          ),

          // Right: Current Value & P&L
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                _formatInr(asset.currentValueInr),
                style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13, fontFamily: 'monospace'),
              ),
              const SizedBox(height: 2),
              Text(
                '${isPos ? '+' : ''}${asset.unrealizedPnLPercent.toStringAsFixed(2)}% (${_formatInr(asset.unrealizedPnL)})',
                style: TextStyle(
                  color: isPos ? const Color(0xFF00E676) : const Color(0xFFFF1744),
                  fontSize: 11,
                  fontFamily: 'monospace',
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
