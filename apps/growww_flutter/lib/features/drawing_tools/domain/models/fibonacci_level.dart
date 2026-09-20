class FibonacciLevel {
  final double ratio;
  final double price;
  final int colorHex;
  final String label;
  final bool isVisible;

  const FibonacciLevel({
    required this.ratio,
    required this.price,
    required this.colorHex,
    required this.label,
    this.isVisible = true,
  });

  Map<String, dynamic> toJson() => {
        'ratio': ratio,
        'price': price,
        'color_hex': colorHex,
        'label': label,
        'is_visible': isVisible,
      };

  factory FibonacciLevel.fromJson(Map<String, dynamic> json) {
    return FibonacciLevel(
      ratio: (json['ratio'] as num).toDouble(),
      price: (json['price'] as num).toDouble(),
      colorHex: json['color_hex'] as int,
      label: json['label'] as String,
      isVisible: json['is_visible'] as bool,
    );
  }
}
