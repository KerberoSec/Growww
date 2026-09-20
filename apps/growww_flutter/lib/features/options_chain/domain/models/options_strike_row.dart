import 'option_contract_quote.dart';

/// Single horizontal row in the Options Chain Matrix containing Call quote, Strike, and Put quote.
class OptionsStrikeRow {
  final double strikePrice;
  final bool isAtm;
  final OptionContractQuote call;
  final OptionContractQuote put;

  const OptionsStrikeRow({
    required this.strikePrice,
    required this.isAtm,
    required this.call,
    required this.put,
  });

  /// Put-Call Ratio for this specific strike price (Put OI / Call OI)
  double get strikePcr {
    if (call.openInterest <= 0) return 0.0;
    return put.openInterest / call.openInterest;
  }

  OptionsStrikeRow copyWith({
    double? strikePrice,
    bool? isAtm,
    OptionContractQuote? call,
    OptionContractQuote? put,
  }) {
    return OptionsStrikeRow(
      strikePrice: strikePrice ?? this.strikePrice,
      isAtm: isAtm ?? this.isAtm,
      call: call ?? this.call,
      put: put ?? this.put,
    );
  }
}
