import 'dart:math' as math;
import '../models/treemap_enums.dart';

class TreemapColor {
  final int value;
  const TreemapColor(this.value);

  int get alpha => (0xFF000000 & value) != 0 ? ((value >> 24) & 0xFF) : 0xFF;
  int get red => (value >> 16) & 0xFF;
  int get green => (value >> 8) & 0xFF;
  int get blue => value & 0xFF;

  static const TreemapColor white = TreemapColor(0xFFFFFFFF);
  static const TreemapColor black = TreemapColor(0xFF000000);

  static TreemapColor lerp(TreemapColor a, TreemapColor b, double t) {
    final clampedT = t.clamp(0.0, 1.0);
    final alpha = (a.alpha + (b.alpha - a.alpha) * clampedT).round();
    final red = (a.red + (b.red - a.red) * clampedT).round();
    final green = (a.green + (b.green - a.green) * clampedT).round();
    final blue = (a.blue + (b.blue - a.blue) * clampedT).round();

    final val = (alpha << 24) | (red << 16) | (green << 8) | blue;
    return TreemapColor(val);
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is TreemapColor && runtimeType == other.runtimeType && value == other.value;

  @override
  int get hashCode => value.hashCode;

  @override
  String toString() => 'TreemapColor(0x${value.toRadixString(16).padLeft(8, '0')})';
}

/// Service responsible for color interpolation and SEBI digital accessibility compliance.
class HeatmapColorEngine {
  // Standard Financial Palette
  static const TreemapColor standardBull = TreemapColor(0xFF00E676); // Neon Emerald Green
  static const TreemapColor standardBear = TreemapColor(0xFFFF1744); // Crimson Red
  static const TreemapColor neutralBg = TreemapColor(0xFF161922); // Charcoal Neutral

  // SEBI Accessible Color-Blind Palette (Deuteranopia / Protanopia friendly)
  static const TreemapColor accessibleBull = TreemapColor(0xFF2979FF); // Cobalt Blue
  static const TreemapColor accessibleBear = TreemapColor(0xFFFF9100); // Amber Orange

  // High-Contrast Monochrome Palette
  static const TreemapColor monoBull = TreemapColor(0xFFE2E8F0); // Light Silver
  static const TreemapColor monoBear = TreemapColor(0xFF334155); // Dark Slate Gray

  /// Computes tile background color according to return percentage (-3% to +3% linear clamp)
  static TreemapColor getColor({
    required double changePercentage,
    required HeatmapColorMode mode,
    double maxThreshold = 3.0,
  }) {
    final clamped = (changePercentage / maxThreshold).clamp(-1.0, 1.0);
    final absRatio = clamped.abs();

    switch (mode) {
      case HeatmapColorMode.standardRedGreen:
        if (clamped > 0) {
          return TreemapColor.lerp(neutralBg, standardBull, absRatio);
        } else if (clamped < 0) {
          return TreemapColor.lerp(neutralBg, standardBear, absRatio);
        }
        return neutralBg;

      case HeatmapColorMode.accessibleBlueOrange:
        if (clamped > 0) {
          return TreemapColor.lerp(neutralBg, accessibleBull, absRatio);
        } else if (clamped < 0) {
          return TreemapColor.lerp(neutralBg, accessibleBear, absRatio);
        }
        return neutralBg;

      case HeatmapColorMode.highContrastMonochrome:
        if (clamped > 0) {
          return TreemapColor.lerp(neutralBg, monoBull, absRatio);
        } else if (clamped < 0) {
          return TreemapColor.lerp(neutralBg, monoBear, absRatio);
        }
        return neutralBg;
    }
  }

  /// Calculates relative luminance compliant with WCAG 2.1 specifications
  static double calculateLuminance(TreemapColor color) {
    double transform(double channel) {
      final s = channel / 255.0;
      return s <= 0.03928 ? s / 12.92 : math.pow((s + 0.055) / 1.055, 2.4).toDouble();
    }

    final r = transform(color.red.toDouble());
    final g = transform(color.green.toDouble());
    final b = transform(color.blue.toDouble());

    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  }

  /// Computes WCAG 2.1 contrast ratio between two colors (Target: >= 4.5:1 for standard text)
  static double calculateContrastRatio(TreemapColor color1, TreemapColor color2) {
    final l1 = calculateLuminance(color1);
    final l2 = calculateLuminance(color2);
    final lighter = math.max(l1, l2);
    final darker = math.min(l1, l2);
    return (lighter + 0.05) / (darker + 0.05);
  }

  /// Selects optimal text color (White or Black) to ensure WCAG 2.1 AA compliance (>= 4.5:1)
  static TreemapColor getAccessibleTextColor(TreemapColor backgroundColor) {
    final contrastWithWhite = calculateContrastRatio(backgroundColor, TreemapColor.white);
    final contrastWithBlack = calculateContrastRatio(backgroundColor, TreemapColor.black);
    return contrastWithWhite >= contrastWithBlack ? TreemapColor.white : TreemapColor.black;
  }
}
