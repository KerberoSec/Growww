import '../models/option_contract_quote.dart';
import '../models/options_chain_snapshot.dart';
import '../models/options_enums.dart';
import '../models/options_strike_row.dart';

/// Financial analytics and quantitative derivatives engine for options chain matrix.
class OptionsAnalyticsEngine {
  /// Resolves the At-The-Money (ATM) strike closest to current spot price.
  static double findAtmStrike(double spotPrice, List<double> strikes) {
    if (strikes.isEmpty) return spotPrice;
    double closest = strikes.first;
    double minDiff = (closest - spotPrice).abs();

    for (final strike in strikes) {
      final diff = (strike - spotPrice).abs();
      if (diff < minDiff) {
        minDiff = diff;
        closest = strike;
      }
    }
    return closest;
  }

  /// Classifies moneyness for Call or Put options.
  static OptionMoneyness classifyMoneyness({
    required double strike,
    required double spotPrice,
    required OptionType type,
    double thresholdPct = 0.04, // 4% threshold for Deep ITM / Deep OTM
  }) {
    final diffPct = (strike - spotPrice) / spotPrice;

    if (type == OptionType.call) {
      if (diffPct.abs() < 0.005) {
        return OptionMoneyness.atm;
      } else if (strike < spotPrice * (1.0 - thresholdPct)) {
        return OptionMoneyness.deepItm;
      } else if (strike < spotPrice) {
        return OptionMoneyness.itm;
      } else if (strike > spotPrice * (1.0 + thresholdPct)) {
        return OptionMoneyness.deepOtm;
      } else {
        return OptionMoneyness.otm;
      }
    } else {
      // Put option
      if (diffPct.abs() < 0.005) {
        return OptionMoneyness.atm;
      } else if (strike > spotPrice * (1.0 + thresholdPct)) {
        return OptionMoneyness.deepItm;
      } else if (strike > spotPrice) {
        return OptionMoneyness.itm;
      } else if (strike < spotPrice * (1.0 - thresholdPct)) {
        return OptionMoneyness.deepOtm;
      } else {
        return OptionMoneyness.otm;
      }
    }
  }

  /// Computes Put-Call Ratio (PCR).
  static double computePcr({
    required double totalPutOi,
    required double totalCallOi,
  }) {
    if (totalCallOi <= 0.0) return 0.0;
    return totalPutOi / totalCallOi;
  }

  /// Computes Max Pain strike across all strikes in the chain.
  /// The strike where cumulative option writer payout is minimized.
  static double computeMaxPainStrike(List<OptionsStrikeRow> rows) {
    if (rows.isEmpty) return 0.0;

    double minPayout = double.infinity;
    double maxPainStrike = rows.first.strikePrice;

    for (final candidate in rows) {
      final s = candidate.strikePrice;
      double totalPayoutAtS = 0.0;

      for (final row in rows) {
        final k = row.strikePrice;
        // Call option payout if settlement is s: max(0, s - k) * Call_OI
        final callPayout = (s > k) ? (s - k) * row.call.openInterest : 0.0;
        // Put option payout if settlement is s: max(0, k - s) * Put_OI
        final putPayout = (k > s) ? (k - s) * row.put.openInterest : 0.0;

        totalPayoutAtS += (callPayout + putPayout);
      }

      if (totalPayoutAtS < minPayout) {
        minPayout = totalPayoutAtS;
        maxPainStrike = s;
      }
    }

    return maxPainStrike;
  }

  /// Generates a realistic mock options chain snapshot around a spot price.
  static OptionsChainSnapshot generateMockChain({
    required String symbol,
    required double spotPrice,
    required String selectedExpiry,
    required List<String> expiries,
    double strikeInterval = 500.0,
    int strikesAboveBelow = 4,
  }) {
    final baseStrike = (spotPrice / strikeInterval).round() * strikeInterval;
    final strikes = <double>[];

    for (int i = -strikesAboveBelow; i <= strikesAboveBelow; i++) {
      strikes.add(baseStrike + (i * strikeInterval));
    }

    final atmStrike = findAtmStrike(spotPrice, strikes);
    final rows = <OptionsStrikeRow>[];
    double totalCallOi = 0.0;
    double totalPutOi = 0.0;

    for (final strike in strikes) {
      final isAtm = (strike == atmStrike);
      final callMoneyness = classifyMoneyness(strike: strike, spotPrice: spotPrice, type: OptionType.call);
      final putMoneyness = classifyMoneyness(strike: strike, spotPrice: spotPrice, type: OptionType.put);

      // Synthetic pricing: intrinsic + extrinsic
      final callIntrinsic = (spotPrice - strike > 0) ? (spotPrice - strike) : 0.0;
      final putIntrinsic = (strike - spotPrice > 0) ? (strike - spotPrice) : 0.0;
      const timeVal = 320.0;

      final callLtp = callIntrinsic + timeVal;
      final putLtp = putIntrinsic + timeVal;

      // Realistic OI distribution (high OI near ATM)
      final distFromAtm = (strike - atmStrike).abs() / strikeInterval;
      final callOi = (1200.0 - distFromAtm * 150.0).clamp(100.0, 2000.0);
      final putOi = (1100.0 - distFromAtm * 130.0).clamp(100.0, 2000.0);

      totalCallOi += callOi;
      totalPutOi += putOi;

      // Delta approximations:
      final callDelta = isAtm
          ? 0.50
          : (strike < spotPrice ? 0.70 : 0.30);
      final putDelta = callDelta - 1.0;

      final callQuote = OptionContractQuote(
        symbol: '$symbol-${selectedExpiry.replaceAll(' ', '')}-${strike.toInt()}-CE',
        strikePrice: strike,
        type: OptionType.call,
        ltp: callLtp,
        priceChange: 15.0,
        changePercent: 2.8,
        bidPrice: callLtp - 2.0,
        bidQty: 10.0,
        askPrice: callLtp + 2.0,
        askQty: 8.0,
        openInterest: callOi,
        volume: callOi * 2.5,
        iv: 0.52,
        delta: callDelta,
        gamma: 0.00012,
        theta: -24.5,
        vega: 18.2,
        moneyness: callMoneyness,
      );

      final putQuote = OptionContractQuote(
        symbol: '$symbol-${selectedExpiry.replaceAll(' ', '')}-${strike.toInt()}-PE',
        strikePrice: strike,
        type: OptionType.put,
        ltp: putLtp,
        priceChange: -12.0,
        changePercent: -2.1,
        bidPrice: putLtp - 2.0,
        bidQty: 12.0,
        askPrice: putLtp + 2.0,
        askQty: 9.0,
        openInterest: putOi,
        volume: putOi * 2.3,
        iv: 0.54,
        delta: putDelta,
        gamma: 0.00012,
        theta: -23.8,
        vega: 18.0,
        moneyness: putMoneyness,
      );

      rows.add(OptionsStrikeRow(
        strikePrice: strike,
        isAtm: isAtm,
        call: callQuote,
        put: putQuote,
      ));
    }

    final pcr = computePcr(totalPutOi: totalPutOi, totalCallOi: totalCallOi);
    final maxPain = computeMaxPainStrike(rows);

    return OptionsChainSnapshot(
      underlyingSymbol: symbol,
      spotPrice: spotPrice,
      spotPriceChangePercent: 1.45,
      selectedExpiry: selectedExpiry,
      availableExpiries: expiries,
      strikeRows: rows,
      totalCallOi: totalCallOi,
      totalPutOi: totalPutOi,
      overallPcr: pcr,
      maxPainStrike: maxPain,
    );
  }
}
