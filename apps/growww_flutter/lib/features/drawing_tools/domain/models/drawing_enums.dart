// Drawing Tools Enums (Prompt 558)

enum DrawingToolType {
  cursor,
  trendline,
  horizontalLine,
  verticalLine,
  fibonacciRetracement,
  parallelChannel,
  priceRange,
  textNote,
}

extension DrawingToolTypeExtension on DrawingToolType {
  String get displayName {
    switch (this) {
      case DrawingToolType.cursor:
        return 'Cursor / Select';
      case DrawingToolType.trendline:
        return 'Trendline';
      case DrawingToolType.horizontalLine:
        return 'Horizontal Ray';
      case DrawingToolType.verticalLine:
        return 'Vertical Time Line';
      case DrawingToolType.fibonacciRetracement:
        return 'Fibonacci Retracement';
      case DrawingToolType.parallelChannel:
        return 'Parallel Channel';
      case DrawingToolType.priceRange:
        return 'Price & Date Range';
      case DrawingToolType.textNote:
        return 'Text Annotation';
    }
  }

  int get requiredPoints {
    switch (this) {
      case DrawingToolType.cursor:
        return 0;
      case DrawingToolType.horizontalLine:
      case DrawingToolType.verticalLine:
      case DrawingToolType.textNote:
        return 1;
      case DrawingToolType.trendline:
      case DrawingToolType.fibonacciRetracement:
      case DrawingToolType.priceRange:
        return 2;
      case DrawingToolType.parallelChannel:
        return 3;
    }
  }
}

enum LineStyle {
  solid,
  dashed,
  dotted,
}

enum MagnetMode {
  none,
  weak, // Snaps within 15px
  strong, // Snaps within 30px
}
