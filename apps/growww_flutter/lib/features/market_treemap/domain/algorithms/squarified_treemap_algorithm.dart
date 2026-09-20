import 'dart:math' as math;
import '../models/treemap_geometry.dart';
import '../models/treemap_node.dart';

/// Pure Dart implementation of the Bruls-Huizing-van Wijk Squarified Treemap Algorithm.
/// Recursively partitions 2D rectangular viewports targeting aspect ratios near 1.0.
class SquarifiedTreemapAlgorithm {
  final double sectorHeaderHeight;
  final double sectorPadding;
  final double leafPadding;

  const SquarifiedTreemapAlgorithm({
    this.sectorHeaderHeight = 22.0,
    this.sectorPadding = 3.0,
    this.leafPadding = 1.5,
  });

  /// Computes squarified layout for the root node and all nested children
  SectorGroupNode computeLayout({
    required SectorGroupNode root,
    required TreemapRect containerRect,
  }) {
    final rootWithRect = root.copyWithRect(containerRect);
    if (root.children.isEmpty || containerRect.width <= 0 || containerRect.height <= 0) {
      return rootWithRect;
    }

    // Partition children within containerRect
    final laidOutChildren = _layoutChildren(
      children: root.children,
      bounds: containerRect.deflated(sectorPadding),
    );

    // Recursively layout any nested sector group nodes
    final recursiveChildren = <TreemapNode>[];
    for (final child in laidOutChildren) {
      if (child is SectorGroupNode) {
        // Reserve space for sector header
        final innerBounds = TreemapRect(
          left: child.rect.left + sectorPadding,
          top: child.rect.top + sectorHeaderHeight + sectorPadding,
          width: math.max(0.0, child.rect.width - sectorPadding * 2),
          height: math.max(0.0, child.rect.height - sectorHeaderHeight - sectorPadding * 2),
        );
        recursiveChildren.add(computeLayout(root: child, containerRect: innerBounds));
      } else if (child is SecurityLeafNode) {
        recursiveChildren.add(child.copyWithRect(child.rect.deflated(leafPadding)));
      } else {
        recursiveChildren.add(child);
      }
    }

    return rootWithRect.copyWithChildren(recursiveChildren);
  }

  List<TreemapNode> _layoutChildren({
    required List<TreemapNode> children,
    required TreemapRect bounds,
  }) {
    if (children.isEmpty) return const [];
    if (bounds.width <= 0 || bounds.height <= 0) {
      return children.map((c) => c.copyWithRect(TreemapRect(left: bounds.left, top: bounds.top, width: 0, height: 0))).toList();
    }

    // Sort items by weight descending
    final sorted = List<TreemapNode>.from(children)
      ..sort((a, b) => b.value.compareTo(a.value));

    final totalWeight = sorted.fold(0.0, (sum, item) => sum + math.max(0.0001, item.value));
    final containerArea = bounds.width * bounds.height;

    // Convert weights to normalized areas
    final areaMap = <TreemapNode, double>{};
    for (final node in sorted) {
      final normWeight = math.max(0.0001, node.value);
      areaMap[node] = (normWeight / totalWeight) * containerArea;
    }

    final result = <TreemapNode>[];
    _squarify(
      items: sorted,
      areaMap: areaMap,
      row: [],
      bounds: bounds,
      result: result,
    );

    return result;
  }

  void _squarify({
    required List<TreemapNode> items,
    required Map<TreemapNode, double> areaMap,
    required List<TreemapNode> row,
    required TreemapRect bounds,
    required List<TreemapNode> result,
  }) {
    if (items.isEmpty) {
      if (row.isNotEmpty) {
        _layoutRow(row, areaMap, bounds, result);
      }
      return;
    }

    if (bounds.width <= 0 || bounds.height <= 0) {
      for (final item in items) {
        result.add(item.copyWithRect(TreemapRect(left: bounds.left, top: bounds.top, width: 0, height: 0)));
      }
      return;
    }

    final side = math.min(bounds.width, bounds.height);
    final candidate = items.first;

    if (row.isEmpty) {
      _squarify(
        items: items.sublist(1),
        areaMap: areaMap,
        row: [candidate],
        bounds: bounds,
        result: result,
      );
    } else {
      final currentWorst = _worstAspectRatio(row, areaMap, side);
      final nextRow = List<TreemapNode>.from(row)..add(candidate);
      final nextWorst = _worstAspectRatio(nextRow, areaMap, side);

      if (nextWorst <= currentWorst) {
        _squarify(
          items: items.sublist(1),
          areaMap: areaMap,
          row: nextRow,
          bounds: bounds,
          result: result,
        );
      } else {
        final remainingBounds = _layoutRow(row, areaMap, bounds, result);
        _squarify(
          items: items,
          areaMap: areaMap,
          row: [],
          bounds: remainingBounds,
          result: result,
        );
      }
    }
  }

  double _worstAspectRatio(
    List<TreemapNode> row,
    Map<TreemapNode, double> areaMap,
    double sideLength,
  ) {
    if (row.isEmpty || sideLength <= 0) return double.infinity;
    double sumArea = 0.0;
    double minArea = double.infinity;
    double maxArea = 0.0;

    for (final node in row) {
      final a = areaMap[node] ?? 0.001;
      sumArea += a;
      if (a < minArea) minArea = a;
      if (a > maxArea) maxArea = a;
    }

    if (sumArea <= 0 || minArea <= 0) return double.infinity;

    final s2 = sideLength * sideLength;
    final sum2 = sumArea * sumArea;

    final ratio1 = (s2 * maxArea) / sum2;
    final ratio2 = sum2 / (s2 * minArea);

    return math.max(ratio1, ratio2);
  }

  TreemapRect _layoutRow(
    List<TreemapNode> row,
    Map<TreemapNode, double> areaMap,
    TreemapRect bounds,
    List<TreemapNode> result,
  ) {
    final rowArea = row.fold(0.0, (sum, node) => sum + (areaMap[node] ?? 0.0));
    final isHorizontal = bounds.width >= bounds.height;

    if (isHorizontal) {
      final rowWidth = rowArea / bounds.height;
      double currentY = bounds.top;

      for (final node in row) {
        final nodeArea = areaMap[node] ?? 0.0;
        final nodeHeight = rowWidth > 0 ? (nodeArea / rowWidth) : 0.0;

        final nodeRect = TreemapRect(
          left: bounds.left,
          top: currentY,
          width: rowWidth,
          height: nodeHeight,
        );
        result.add(node.copyWithRect(nodeRect));
        currentY += nodeHeight;
      }

      return TreemapRect(
        left: bounds.left + rowWidth,
        top: bounds.top,
        width: math.max(0.0, bounds.width - rowWidth),
        height: bounds.height,
      );
    } else {
      final rowHeight = rowArea / bounds.width;
      double currentX = bounds.left;

      for (final node in row) {
        final nodeArea = areaMap[node] ?? 0.0;
        final nodeWidth = rowHeight > 0 ? (nodeArea / rowHeight) : 0.0;

        final nodeRect = TreemapRect(
          left: currentX,
          top: bounds.top,
          width: nodeWidth,
          height: rowHeight,
        );
        result.add(node.copyWithRect(nodeRect));
        currentX += nodeWidth;
      }

      return TreemapRect(
        left: bounds.left,
        top: bounds.top + rowHeight,
        width: bounds.width,
        height: math.max(0.0, bounds.height - rowHeight),
      );
    }
  }
}
