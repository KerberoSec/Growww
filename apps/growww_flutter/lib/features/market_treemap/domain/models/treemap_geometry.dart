class TreemapRect {
  final double left;
  final double top;
  final double width;
  final double height;

  const TreemapRect({
    required this.left,
    required this.top,
    required this.width,
    required this.height,
  });

  double get right => left + width;
  double get bottom => top + height;
  double get area => width * height;
  double get aspectRatio => (width > 0 && height > 0)
      ? (width > height ? width / height : height / width)
      : 1.0;

  bool contains(double x, double y) {
    return x >= left && x <= right && y >= top && y <= bottom;
  }

  TreemapRect copyWith({
    double? left,
    double? top,
    double? width,
    double? height,
  }) {
    return TreemapRect(
      left: left ?? this.left,
      top: top ?? this.top,
      width: width ?? this.width,
      height: height ?? this.height,
    );
  }

  TreemapRect deflated(double padding) {
    final newWidth = (width - padding * 2).clamp(0.0, double.infinity);
    final newHeight = (height - padding * 2).clamp(0.0, double.infinity);
    return TreemapRect(
      left: left + padding,
      top: top + padding,
      width: newWidth,
      height: newHeight,
    );
  }

  Map<String, dynamic> toJson() => {
        'left': left,
        'top': top,
        'width': width,
        'height': height,
      };

  factory TreemapRect.fromJson(Map<String, dynamic> json) {
    return TreemapRect(
      left: (json['left'] as num).toDouble(),
      top: (json['top'] as num).toDouble(),
      width: (json['width'] as num).toDouble(),
      height: (json['height'] as num).toDouble(),
    );
  }
}
