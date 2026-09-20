// Treemap Enums (Prompt 535)

enum TreemapMetricType {
  marketCap,
  volume24h,
  marketDepthLiquidity,
  openInterest,
}

extension TreemapMetricTypeExtension on TreemapMetricType {
  String get displayName {
    switch (this) {
      case TreemapMetricType.marketCap:
        return 'Market Capitalization';
      case TreemapMetricType.volume24h:
        return '24h Turnover';
      case TreemapMetricType.marketDepthLiquidity:
        return 'Consolidated Depth';
      case TreemapMetricType.openInterest:
        return 'Open Interest';
    }
  }

  String get shortName {
    switch (this) {
      case TreemapMetricType.marketCap:
        return 'MCap';
      case TreemapMetricType.volume24h:
        return 'Turnover';
      case TreemapMetricType.marketDepthLiquidity:
        return 'Depth';
      case TreemapMetricType.openInterest:
        return 'OI';
    }
  }
}

enum HeatmapColorMode {
  standardRedGreen,
  accessibleBlueOrange,
  highContrastMonochrome,
}

extension HeatmapColorModeExtension on HeatmapColorMode {
  String get displayName {
    switch (this) {
      case HeatmapColorMode.standardRedGreen:
        return 'Standard (Red / Green)';
      case HeatmapColorMode.accessibleBlueOrange:
        return 'Accessible (Blue / Orange)';
      case HeatmapColorMode.highContrastMonochrome:
        return 'High-Contrast Monochrome';
    }
  }
}

enum TreemapTimeHorizon {
  oneDay,
  oneWeek,
  oneMonth,
  oneYear,
}

extension TreemapTimeHorizonExtension on TreemapTimeHorizon {
  String get label {
    switch (this) {
      case TreemapTimeHorizon.oneDay:
        return '1D';
      case TreemapTimeHorizon.oneWeek:
        return '1W';
      case TreemapTimeHorizon.oneMonth:
        return '1M';
      case TreemapTimeHorizon.oneYear:
        return '1Y';
    }
  }
}
