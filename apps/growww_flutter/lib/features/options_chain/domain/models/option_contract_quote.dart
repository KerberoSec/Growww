import 'options_enums.dart';

/// Single option contract quotes and Greek sensitivities.
class OptionContractQuote {
  final String symbol; // e.g. BTC-26SEP26-64000-CE
  final double strikePrice;
  final OptionType type;
  final double ltp;
  final double priceChange;
  final double changePercent;
  final double bidPrice;
  final double bidQty;
  final double askPrice;
  final double askQty;
  final double openInterest;
  final double oiChange;
  final double volume;
  final double iv; // Implied Volatility (e.g. 0.52 for 52%)
  final double delta;
  final double gamma;
  final double theta;
  final double vega;
  final OptionMoneyness moneyness;

  const OptionContractQuote({
    required this.symbol,
    required this.strikePrice,
    required this.type,
    required this.ltp,
    this.priceChange = 0.0,
    this.changePercent = 0.0,
    required this.bidPrice,
    required this.bidQty,
    required this.askPrice,
    required this.askQty,
    required this.openInterest,
    this.oiChange = 0.0,
    required this.volume,
    required this.iv,
    required this.delta,
    required this.gamma,
    required this.theta,
    required this.vega,
    this.moneyness = OptionMoneyness.otm,
  });

  OptionContractQuote copyWith({
    String? symbol,
    double? strikePrice,
    OptionType? type,
    double? ltp,
    double? priceChange,
    double? changePercent,
    double? bidPrice,
    double? bidQty,
    double? askPrice,
    double? askQty,
    double? openInterest,
    double? oiChange,
    double? volume,
    double? iv,
    double? delta,
    double? gamma,
    double? theta,
    double? vega,
    OptionMoneyness? moneyness,
  }) {
    return OptionContractQuote(
      symbol: symbol ?? this.symbol,
      strikePrice: strikePrice ?? this.strikePrice,
      type: type ?? this.type,
      ltp: ltp ?? this.ltp,
      priceChange: priceChange ?? this.priceChange,
      changePercent: changePercent ?? this.changePercent,
      bidPrice: bidPrice ?? this.bidPrice,
      bidQty: bidQty ?? this.bidQty,
      askPrice: askPrice ?? this.askPrice,
      askQty: askQty ?? this.askQty,
      openInterest: openInterest ?? this.openInterest,
      oiChange: oiChange ?? this.oiChange,
      volume: volume ?? this.volume,
      iv: iv ?? this.iv,
      delta: delta ?? this.delta,
      gamma: gamma ?? this.gamma,
      theta: theta ?? this.theta,
      vega: vega ?? this.vega,
      moneyness: moneyness ?? this.moneyness,
    );
  }
}
