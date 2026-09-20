import 'dart:math' as math;
import '../models/iceberg_enums.dart';
import '../models/iceberg_order_config.dart';
import '../models/iceberg_tranche.dart';

class IcebergSlicingValidationResult {
  final bool isValid;
  final String? errorMessage;

  const IcebergSlicingValidationResult({
    required this.isValid,
    this.errorMessage,
  });

  factory IcebergSlicingValidationResult.valid() =>
      const IcebergSlicingValidationResult(isValid: true);
  factory IcebergSlicingValidationResult.invalid(String error) =>
      IcebergSlicingValidationResult(isValid: false, errorMessage: error);
}

class IcebergSlicingAllocator {
  static IcebergSlicingValidationResult validateConfig(IcebergOrderConfig config) {
    if (config.totalQuantity <= 0) {
      return IcebergSlicingValidationResult.invalid('Total order quantity must be greater than zero.');
    }
    if (config.visibleTrancheQuantity <= 0) {
      return IcebergSlicingValidationResult.invalid('Visible tranche size must be greater than zero.');
    }
    if (config.visibleTrancheQuantity > config.totalQuantity) {
      return IcebergSlicingValidationResult.invalid('Visible tranche size cannot exceed total order quantity.');
    }
    if (config.totalQuantity < config.visibleTrancheQuantity * 1.5) {
      return IcebergSlicingValidationResult.invalid('Iceberg requires total quantity to be at least 1.5x visible tranche size.');
    }
    return IcebergSlicingValidationResult.valid();
  }

  /// Allocates full sequence of tranches with optional pseudorandom variance
  static List<IcebergTranche> allocateTranches(
    IcebergOrderConfig config, {
    int? randomSeed,
  }) {
    final validation = validateConfig(config);
    if (!validation.isValid) {
      return [];
    }

    final rand = randomSeed != null ? math.Random(randomSeed) : math.Random(42);
    final tranches = <IcebergTranche>[];
    double remaining = config.totalQuantity;
    int index = 0;

    final minVar = config.varianceMode.minVariancePercent;
    final maxVar = config.varianceMode.maxVariancePercent;

    while (remaining > 0.000001) {
      double targetSize = config.visibleTrancheQuantity;

      if (config.varianceMode != IcebergVarianceMode.fixedTranche) {
        // Random multiplier between (1.0 - maxVar) and (1.0 + maxVar)
        final range = maxVar - minVar;
        final delta = (rand.nextDouble() * range * 2) - range;
        final factor = 1.0 + (delta >= 0 ? delta + minVar : delta - minVar);
        targetSize = config.visibleTrancheQuantity * factor;
      }

      // Clamp target size
      double trancheSize = math.min(remaining, targetSize);

      // If remaining after this tranche is tiny sliver (< 20% base tranche), absorb into current
      if ((remaining - trancheSize) > 0 && (remaining - trancheSize) < (config.visibleTrancheQuantity * 0.2)) {
        trancheSize = remaining;
      }

      // Round to 4 decimal places for precision
      trancheSize = (trancheSize * 10000).round() / 10000.0;
      if (trancheSize > remaining) {
        trancheSize = remaining;
      }

      final hiddenAfterTranche = math.max(0.0, remaining - trancheSize);

      tranches.add(IcebergTranche(
        trancheIndex: index,
        visibleQuantity: trancheSize,
        hiddenRemaining: (hiddenAfterTranche * 10000).round() / 10000.0,
        executedQuantity: 0.0,
        price: config.priceLimit,
        status: index == 0 ? TrancheStatus.active : TrancheStatus.pending,
      ));

      remaining = (remaining - trancheSize);
      remaining = (remaining * 10000).round() / 10000.0;
      index++;
    }

    return tranches;
  }
}
