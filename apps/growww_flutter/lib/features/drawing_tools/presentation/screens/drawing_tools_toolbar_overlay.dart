import 'package:flutter/material.dart';
import '../../domain/models/drawing_enums.dart';
import '../../domain/models/drawing_point.dart';
import '../controllers/drawing_tools_controller.dart';

class DrawingToolsToolbarOverlay extends StatefulWidget {
  final DrawingToolsController controller;

  const DrawingToolsToolbarOverlay({
    Key? key,
    required this.controller,
  }) : super(key: key);

  @override
  State<DrawingToolsToolbarOverlay> createState() =>
      _DrawingToolsToolbarOverlayState();
}

class _DrawingToolsToolbarOverlayState
    extends State<DrawingToolsToolbarOverlay> {
  static const Color obsidianBackground = Color(0xFF0B0E14);
  static const Color surfaceCard = Color(0xFF141923);
  static const Color neonGreen = Color(0xFF00F0A0);
  static const Color textMuted = Color(0xFF8B949E);

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<DrawingToolsState>(
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
              'FLUTTER DRAWING TOOLS TOOLBAR (TRENDLINES, FI',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.0,
              ),
            ),
          ),
          body: Column(
            children: [
              _buildTopToolbar(state),
              Expanded(
                child: GestureDetector(
                  onTapUp: (details) {
                    final pt = DrawingPoint(
                      timestampMs: DateTime.now().millisecondsSinceEpoch,
                      price: 64200.0 - (details.localPosition.dy * 10),
                      screenPixelX: details.localPosition.dx,
                      screenPixelY: details.localPosition.dy,
                    );
                    widget.controller.addPointToDrawing(pt);
                  },
                  child: Container(
                    margin: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: surfaceCard,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.white12),
                    ),
                    child: Stack(
                      children: [
                        _buildChartGridPlaceholder(),
                        _buildFibonacciOverlay(state),
                        _buildElementsOverlay(state),
                        _buildInstructionBadge(state),
                      ],
                    ),
                  ),
                ),
              ),
              _buildBottomActionDock(state),
            ],
          ),
        );
      },
    );
  }

  Widget _buildTopToolbar(DrawingToolsState state) {
    return Container(
      height: 52,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      color: surfaceCard,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: [
          for (final tool in DrawingToolType.values) ...[
            Padding(
              padding: const EdgeInsets.only(right: 6),
              child: IconButton(
                icon: _getToolIcon(tool),
                color: state.activeTool == tool ? neonGreen : textMuted,
                tooltip: tool.displayName,
                onPressed: () => widget.controller.setActiveTool(tool),
              ),
            ),
          ],
          const VerticalDivider(color: Colors.white12, width: 16),
          IconButton(
            icon: const Icon(Icons.undo, color: textMuted),
            tooltip: 'Undo',
            onPressed: state.undoStack.isNotEmpty ? () => widget.controller.undo() : null,
          ),
          IconButton(
            icon: const Icon(Icons.redo, color: textMuted),
            tooltip: 'Redo',
            onPressed: state.redoStack.isNotEmpty ? () => widget.controller.redo() : null,
          ),
          IconButton(
            icon: const Icon(Icons.delete_sweep, color: Colors.redAccent),
            tooltip: 'Clear All',
            onPressed: () => widget.controller.clearAll(),
          ),
        ],
      ),
    );
  }

  Icon _getToolIcon(DrawingToolType tool) {
    switch (tool) {
      case DrawingToolType.cursor:
        return const Icon(Icons.near_me);
      case DrawingToolType.trendline:
        return const Icon(Icons.timeline);
      case DrawingToolType.horizontalLine:
        return const Icon(Icons.horizontal_rule);
      case DrawingToolType.verticalLine:
        return const Icon(Icons.more_vert);
      case DrawingToolType.fibonacciRetracement:
        return const Icon(Icons.format_line_spacing);
      case DrawingToolType.parallelChannel:
        return const Icon(Icons.view_headline);
      case DrawingToolType.priceRange:
        return const Icon(Icons.aspect_ratio);
      case DrawingToolType.textNote:
        return const Icon(Icons.text_fields);
    }
  }

  Widget _buildChartGridPlaceholder() {
    return Positioned.fill(
      child: Center(
        child: Text(
          '[ Interactive Candlestick Canvas & Grid ]\nTap to place anchor points',
          textAlign: TextAlign.center,
          style: const TextStyle(color: Colors.white24, fontSize: 13, fontFamily: 'JetBrains Mono'),
        ),
      ),
    );
  }

  Widget _buildFibonacciOverlay(DrawingToolsState state) {
    if (state.activeFibonacciLevels.isEmpty) return const SizedBox();
    return Positioned(
      left: 20,
      top: 40,
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: obsidianBackground.withOpacity(0.85),
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: neonGreen.withOpacity(0.3)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Fibonacci Retracement Grid', style: TextStyle(color: neonGreen, fontSize: 11, fontWeight: FontWeight.bold)),
            const SizedBox(height: 6),
            for (final fib in state.activeFibonacciLevels) ...[
              Text(
                fib.label,
                style: TextStyle(
                  color: Color(fib.colorHex),
                  fontFamily: 'JetBrains Mono',
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildElementsOverlay(DrawingToolsState state) {
    return Positioned(
      right: 16,
      top: 16,
      child: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(color: obsidianBackground, borderRadius: BorderRadius.circular(4)),
        child: Text(
          'Active Overlays: ${state.elements.length}',
          style: const TextStyle(color: Colors.white70, fontSize: 11, fontFamily: 'JetBrains Mono'),
        ),
      ),
    );
  }

  Widget _buildInstructionBadge(DrawingToolsState state) {
    final tool = state.activeTool;
    final remaining = tool.requiredPoints - state.inProgressPoints.length;

    return Positioned(
      left: 16,
      bottom: 16,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        decoration: BoxDecoration(color: obsidianBackground, borderRadius: BorderRadius.circular(4), border: Border.all(color: Colors.white12)),
        child: Text(
          'Tool: ${tool.displayName} ${remaining > 0 ? '(Tap $remaining more points)' : ''}',
          style: const TextStyle(color: neonGreen, fontSize: 11, fontWeight: FontWeight.bold),
        ),
      ),
    );
  }

  Widget _buildBottomActionDock(DrawingToolsState state) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Row(
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
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Chart template with ${state.elements.length} drawings saved!'),
                    backgroundColor: const Color(0xFF1B382B),
                  ),
                );
              },
              child: const Text('Save Chart Layout (Neon Green)', style: TextStyle(fontWeight: FontWeight.bold)),
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
              child: const Text('Close Toolbar'),
            ),
          ),
        ],
      ),
    );
  }
}
