// Iceberg Order Slicing Enums (Prompt 552)

enum IcebergVarianceMode {
  fixedTranche,
  randomizedVariance5to10,
  randomizedVariance10to20,
}

extension IcebergVarianceModeExtension on IcebergVarianceMode {
  String get displayName {
    switch (this) {
      case IcebergVarianceMode.fixedTranche:
        return 'Fixed Tranche Size (0% Variance)';
      case IcebergVarianceMode.randomizedVariance5to10:
        return 'Randomized Display (±5% - 10% Variance)';
      case IcebergVarianceMode.randomizedVariance10to20:
        return 'High Anti-Detection (±10% - 20% Variance)';
    }
  }

  double get minVariancePercent {
    switch (this) {
      case IcebergVarianceMode.fixedTranche:
        return 0.0;
      case IcebergVarianceMode.randomizedVariance5to10:
        return 0.05;
      case IcebergVarianceMode.randomizedVariance10to20:
        return 0.10;
    }
  }

  double get maxVariancePercent {
    switch (this) {
      case IcebergVarianceMode.fixedTranche:
        return 0.0;
      case IcebergVarianceMode.randomizedVariance5to10:
        return 0.10;
      case IcebergVarianceMode.randomizedVariance10to20:
        return 0.20;
    }
  }
}

enum TrancheStatus {
  pending,
  active,
  partiallyFilled,
  filled,
  cancelled,
}
