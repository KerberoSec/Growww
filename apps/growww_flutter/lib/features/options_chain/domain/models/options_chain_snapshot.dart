import 'options_strike_row.dart';

/// Full options chain matrix snapshot for an underlying asset.
class OptionsChainSnapshot {
  final String underlyingSymbol; // e.g. BTC/USDT or NIFTY
  final double spotPrice;
  final double spotPriceChangePercent;
  final String selectedExpiry;
  final List<String> availableExpiries;
  final List<OptionsStrikeRow> strikeRows;
  final double totalCallOi;
  final double totalPutOi;
  final double overallPcr;
  final double maxPainStrike;

  const OptionsChainSnapshot({
    required this.underlyingSymbol,
    required this.spotPrice,
    this.spotPriceChangePercent = 0.0,
    required this.selectedExpiry,
    required this.availableExpiries,
    required this.strikeRows,
    required this.totalCallOi,
    required this.totalPutOi,
    required this.overallPcr,
    required this.maxPainStrike,
  });

  OptionsStrikeRow? get atmRow {
    try {
      return strikeRows.firstWhere((r) => r.isAtm);
    } catch (_) {
      return null;
    }
  }

  OptionsChainSnapshot copyWith({
    String? underlyingSymbol,
    double? spotPrice,
    double? spotPriceChangePercent,
    String? selectedExpiry,
    List<String>? availableExpiries,
    List<OptionsStrikeRow>? strikeRows,
    double? totalCallOi,
    double? totalPutOi,
    double? overallPcr,
    double? maxPainStrike,
  }) {
    return OptionsChainSnapshot(
      underlyingSymbol: underlyingSymbol ?? this.underlyingSymbol,
      spotPrice: spotPrice ?? this.spotPrice,
      spotPriceChangePercent: spotPriceChangePercent ?? this.spotPriceChangePercent,
      selectedExpiry: selectedExpiry ?? this.selectedExpiry,
      availableExpiries: availableExpiries ?? this.availableExpiries,
      strikeRows: strikeRows ?? this.strikeRows,
      totalCallOi: totalCallOi ?? this.totalCallOi,
      totalPutOi: totalPutOi ?? this.totalPutOi,
      overallPcr: overallPcr ?? this.overallPcr,
      maxPainStrike: maxPainStrike ?? this.maxPainStrike,
    );
  }
}
