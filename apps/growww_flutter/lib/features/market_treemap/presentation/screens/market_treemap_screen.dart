import 'package:flutter/material.dart';
import '../../domain/models/treemap_enums.dart';
import '../../domain/models/treemap_node.dart';
import '../../domain/services/heatmap_color_engine.dart';
import '../controllers/treemap_controller.dart';

class MarketTreemapScreen extends StatefulWidget {
  final TreemapController controller;

  const MarketTreemapScreen({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<MarketTreemapScreen> createState() => _MarketTreemapScreenState();
}

class _MarketTreemapScreenState extends State<MarketTreemapScreen> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<TreemapState>(
      stream: widget.controller.stream,
      initialData: widget.controller.state,
      builder: (context, snapshot) {
        final state = snapshot.data ?? widget.controller.state;
        return Scaffold(
          backgroundColor: obsidianBackground,
          appBar: AppBar(
            backgroundColor: obsidianBackground,
            elevation: 0,
            title: Text(
              '${state.selectedIndex} Market Heatmap',
              style: const TextStyle(color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.palette_outlined, color: Colors.white70),
                tooltip: 'Toggle SEBI Color-Blind Accessibility',
                onPressed: () {
                  final nextMode = state.colorMode == HeatmapColorMode.standardRedGreen
                      ? HeatmapColorMode.accessibleBlueOrange
                      : (state.colorMode == HeatmapColorMode.accessibleBlueOrange
                          ? HeatmapColorMode.highContrastMonochrome
                          : HeatmapColorMode.standardRedGreen);
                  widget.controller.setColorMode(nextMode);
                },
              ),
            ],
          ),
          body: Column(
            children: [
              _buildMetricsToolbar(state),
              _buildBreadcrumbBar(state),
              Expanded(
                child: LayoutBuilder(
                  builder: (context, constraints) {
                    // Update layout dimensions dynamically
                    widget.controller.updateViewportDimensions(
                      constraints.maxWidth,
                      constraints.maxHeight,
                    );

                    if (state.isLoading || state.activeViewNode == null) {
                      return const Center(child: CircularProgressIndicator(color: neonGreen));
                    }

                    return Stack(
                      children: _renderTreemapNodes(state.activeViewNode!, state),
                    );
                  },
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildMetricsToolbar(TreemapState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      color: surfaceCard,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Sizing Metric Dropdown
          DropdownButton<TreemapMetricType>(
            value: state.selectedMetric,
            dropdownColor: surfaceCard,
            underline: const SizedBox(),
            style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
            icon: const Icon(Icons.arrow_drop_down, color: neonGreen),
            items: TreemapMetricType.values.map((m) {
              return DropdownMenuItem(
                value: m,
                child: Text(m.displayName),
              );
            }).toList(),
            onChanged: (m) => widget.controller.setMetric(m!),
          ),
          // Time Horizon Chips
          Row(
            children: TreemapTimeHorizon.values.map((h) {
              final isSelected = h == state.selectedHorizon;
              return Padding(
                padding: const EdgeInsets.only(left: 4),
                child: InkWell(
                  onTap: () => widget.controller.setTimeHorizon(h),
                  borderRadius: BorderRadius.circular(4),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: isSelected ? neonGreen : Colors.transparent,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      h.label,
                      style: TextStyle(
                        color: isSelected ? Colors.black : textMuted,
                        fontWeight: FontWeight.bold,
                        fontSize: 11,
                      ),
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildBreadcrumbBar(TreemapState state) {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      alignment: Alignment.centerLeft,
      color: obsidianBackground,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: state.breadcrumbTrail.length,
        separatorBuilder: (_, __) => const Icon(Icons.chevron_right, size: 16, color: textMuted),
        itemBuilder: (context, index) {
          final node = state.breadcrumbTrail[index];
          final isLast = index == state.breadcrumbTrail.length - 1;
          return InkWell(
            onTap: () => widget.controller.navigateBreadcrumb(index),
            child: Center(
              child: Text(
                node.name,
                style: TextStyle(
                  color: isLast ? neonGreen : Colors.white70,
                  fontWeight: isLast ? FontWeight.bold : FontWeight.normal,
                  fontSize: 12,
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  List<Widget> _renderTreemapNodes(SectorGroupNode root, TreemapState state) {
    final widgets = <Widget>[];

    void visit(TreemapNode node) {
      if (node is SectorGroupNode) {
        // Render sector border & header if not root
        if (node != root && node.rect.width > 20 && node.rect.height > 20) {
          widgets.add(
            Positioned(
              left: node.rect.left,
              top: node.rect.top,
              width: node.rect.width,
              height: node.rect.height,
              child: InkWell(
                onDoubleTap: () => widget.controller.drillIntoSector(node),
                child: Container(
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.white24, width: 1),
                  ),
                  alignment: Alignment.topLeft,
                  padding: const EdgeInsets.all(4),
                  child: Text(
                    '${node.name} (${node.cumulativeChangePercentage >= 0 ? '+' : ''}${node.cumulativeChangePercentage.toStringAsFixed(2)}%)',
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(color: textMuted, fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ),
          );
        }
        for (final child in node.children) {
          visit(child);
        }
      } else if (node is SecurityLeafNode) {
        if (node.rect.width > 10 && node.rect.height > 10) {
          final bgColor = HeatmapColorEngine.getColor(
            changePercentage: node.changePercentage,
            mode: state.colorMode,
          );
          final textColor = HeatmapColorEngine.getAccessibleTextColor(bgColor);

          widgets.add(
            Positioned(
              left: node.rect.left,
              top: node.rect.top,
              width: node.rect.width,
              height: node.rect.height,
              child: InkWell(
                onTap: () => _showSecurityPreview(node),
                child: Container(
                  decoration: BoxDecoration(
                    color: Color(bgColor.value),
                    borderRadius: BorderRadius.circular(2),
                    border: Border.all(color: Colors.black26, width: 0.5),
                  ),
                  padding: const EdgeInsets.all(4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Flexible(
                            child: Text(
                              node.symbol,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                color: Color(textColor.value),
                                fontWeight: FontWeight.bold,
                                fontSize: node.rect.width > 80 ? 12 : 10,
                              ),
                            ),
                          ),
                          if (node.isTokenizedRwa)
                            Padding(
                              padding: const EdgeInsets.only(left: 2),
                              child: Icon(Icons.shield, size: 10, color: Color(textColor.value)),
                            ),
                        ],
                      ),
                      if (node.rect.height > 35) ...[
                        const SizedBox(height: 2),
                        Text(
                          '₹${node.ltp.toStringAsFixed(2)}',
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: Color(textColor.value),
                            fontFamily: 'JetBrains Mono',
                            fontSize: node.rect.width > 80 ? 11 : 9,
                          ),
                        ),
                        Text(
                          '${node.changePercentage >= 0 ? '+' : ''}${node.changePercentage.toStringAsFixed(2)}%',
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: Color(textColor.value),
                            fontFamily: 'JetBrains Mono',
                            fontWeight: FontWeight.bold,
                            fontSize: node.rect.width > 80 ? 11 : 9,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          );
        }
      }
    }

    visit(root);
    return widgets;
  }

  void _showSecurityPreview(SecurityLeafNode node) {
    widget.controller.selectSecurity(node);
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
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(node.name, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
                    Text('${node.symbol} • ${node.isin}', style: const TextStyle(color: textMuted, fontSize: 12)),
                  ],
                ),
                if (node.isTokenizedRwa)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(color: neonGreen.withOpacity(0.15), borderRadius: BorderRadius.circular(4)),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.shield, size: 12, color: neonGreen),
                        SizedBox(width: 4),
                        Text('Proof-of-Reserve', style: TextStyle(color: neonGreen, fontSize: 10, fontWeight: FontWeight.bold)),
                      ],
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _buildModalMetric('Last Price', '₹${node.ltp.toStringAsFixed(2)}'),
                _buildModalMetric(
                  '24h Return',
                  '${node.changePercentage >= 0 ? '+' : ''}${node.changePercentage.toStringAsFixed(2)}%',
                  color: node.changePercentage >= 0 ? neonGreen : Colors.redAccent,
                ),
                _buildModalMetric('Turnover', '₹${(node.turnoverInr / 10000000).toStringAsFixed(2)} Cr'),
              ],
            ),
            if (node.isTokenizedRwa && node.rwaProofOfReserveHash != null) ...[
              const SizedBox(height: 12),
              Text(
                'Besu PoR Merkle Root: ${node.rwaProofOfReserveHash!.substring(0, 18)}...',
                style: const TextStyle(color: Colors.white54, fontSize: 10, fontFamily: 'JetBrains Mono'),
              ),
            ],
            const SizedBox(height: 20),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: neonGreen,
                      foregroundColor: Colors.black,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    onPressed: () => Navigator.of(ctx).pop(),
                    child: const Text('Quick Buy', style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: OutlinedButton(
                    style: OutlinedButton.styleFrom(
                      foregroundColor: Colors.redAccent,
                      side: const BorderSide(color: Colors.redAccent),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    onPressed: () => Navigator.of(ctx).pop(),
                    child: const Text('Quick Sell', style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildModalMetric(String label, String value, {Color? color}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(color: textMuted, fontSize: 11)),
        const SizedBox(height: 2),
        Text(
          value,
          style: TextStyle(
            color: color ?? Colors.white,
            fontFamily: 'JetBrains Mono',
            fontWeight: FontWeight.bold,
            fontSize: 14,
          ),
        ),
      ],
    );
  }
}
