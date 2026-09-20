class TreemapTickUpdate {
  final String symbol;
  final double ltp;
  final double changeInr;
  final double changePercentage;
  final double turnoverInr;
  final double totalBidVolume;
  final double totalAskVolume;
  final DateTime timestamp;

  const TreemapTickUpdate({
    required this.symbol,
    required this.ltp,
    required this.changeInr,
    required this.changePercentage,
    required this.turnoverInr,
    required this.totalBidVolume,
    required this.totalAskVolume,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
        'symbol': symbol,
        'ltp': ltp,
        'change_inr': changeInr,
        'change_percentage': changePercentage,
        'turnover_inr': turnoverInr,
        'total_bid_volume': totalBidVolume,
        'total_ask_volume': totalAskVolume,
        'timestamp': timestamp.toIso8601String(),
      };

  factory TreemapTickUpdate.fromJson(Map<String, dynamic> json) {
    return TreemapTickUpdate(
      symbol: json['symbol'] as String,
      ltp: (json['ltp'] as num).toDouble(),
      changeInr: (json['change_inr'] as num).toDouble(),
      changePercentage: (json['change_percentage'] as num).toDouble(),
      turnoverInr: (json['turnover_inr'] as num).toDouble(),
      totalBidVolume: (json['total_bid_volume'] as num).toDouble(),
      totalAskVolume: (json['total_ask_volume'] as num).toDouble(),
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }
}
