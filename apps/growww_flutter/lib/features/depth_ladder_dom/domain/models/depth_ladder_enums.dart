/// Depth of Market (DOM) ladder and Orderbook enums.
enum LadderTickAggregation {
  point05(0.05, '0.05'),
  point1(0.10, '0.10'),
  point5(0.50, '0.50'),
  one(1.00, '1.00'),
  five(5.00, '5.00'),
  ten(10.00, '10.00');

  final double step;
  final String label;

  const LadderTickAggregation(this.step, this.label);
}

enum DomOrderSide {
  bid,
  ask;

  bool get isBid => this == DomOrderSide.bid;
  bool get isAsk => this == DomOrderSide.ask;
  String get displayName => this == DomOrderSide.bid ? 'BUY' : 'SELL';
}

enum DepthMode {
  level2Aggregated,
  level3MarketByOrder;

  String get label => this == DepthMode.level2Aggregated ? 'L2 Depth (Aggregated)' : 'L3 MBO (Individual Orders)';
}
