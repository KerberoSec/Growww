class DrawingPoint {
  final int timestampMs;
  final double price;
  final double screenPixelX;
  final double screenPixelY;

  const DrawingPoint({
    required this.timestampMs,
    required this.price,
    required this.screenPixelX,
    required this.screenPixelY,
  });

  DrawingPoint copyWith({
    int? timestampMs,
    double? price,
    double? screenPixelX,
    double? screenPixelY,
  }) {
    return DrawingPoint(
      timestampMs: timestampMs ?? this.timestampMs,
      price: price ?? this.price,
      screenPixelX: screenPixelX ?? this.screenPixelX,
      screenPixelY: screenPixelY ?? this.screenPixelY,
    );
  }

  Map<String, dynamic> toJson() => {
        'timestamp_ms': timestampMs,
        'price': price,
        'screen_pixel_x': screenPixelX,
        'screen_pixel_y': screenPixelY,
      };

  factory DrawingPoint.fromJson(Map<String, dynamic> json) {
    return DrawingPoint(
      timestampMs: json['timestamp_ms'] as int,
      price: (json['price'] as num).toDouble(),
      screenPixelX: (json['screen_pixel_x'] as num).toDouble(),
      screenPixelY: (json['screen_pixel_y'] as num).toDouble(),
    );
  }
}
