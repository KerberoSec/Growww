import 'drawing_enums.dart';
import 'drawing_point.dart';

class DrawingElement {
  final String elementId;
  final DrawingToolType toolType;
  final List<DrawingPoint> points;
  final int strokeColorHex;
  final double strokeWidth;
  final LineStyle lineStyle;
  final bool isLocked;
  final bool isSelected;
  final String? label;

  const DrawingElement({
    required this.elementId,
    required this.toolType,
    required this.points,
    this.strokeColorHex = 0xFF00F0A0, // Neon green default
    this.strokeWidth = 2.0,
    this.lineStyle = LineStyle.solid,
    this.isLocked = false,
    this.isSelected = false,
    this.label,
  });

  DrawingElement copyWith({
    List<DrawingPoint>? points,
    int? strokeColorHex,
    double? strokeWidth,
    LineStyle? lineStyle,
    bool? isLocked,
    bool? isSelected,
    String? label,
  }) {
    return DrawingElement(
      elementId: elementId,
      toolType: toolType,
      points: points ?? this.points,
      strokeColorHex: strokeColorHex ?? this.strokeColorHex,
      strokeWidth: strokeWidth ?? this.strokeWidth,
      lineStyle: lineStyle ?? this.lineStyle,
      isLocked: isLocked ?? this.isLocked,
      isSelected: isSelected ?? this.isSelected,
      label: label ?? this.label,
    );
  }

  Map<String, dynamic> toJson() => {
        'element_id': elementId,
        'tool_type': toolType.name,
        'points': points.map((p) => p.toJson()).toList(),
        'stroke_color_hex': strokeColorHex,
        'stroke_width': strokeWidth,
        'line_style': lineStyle.name,
        'is_locked': isLocked,
        'is_selected': isSelected,
        'label': label,
      };

  factory DrawingElement.fromJson(Map<String, dynamic> json) {
    return DrawingElement(
      elementId: json['element_id'] as String,
      toolType: DrawingToolType.values.byName(json['tool_type'] as String),
      points: (json['points'] as List<dynamic>)
          .map((p) => DrawingPoint.fromJson(p as Map<String, dynamic>))
          .toList(),
      strokeColorHex: json['stroke_color_hex'] as int,
      strokeWidth: (json['stroke_width'] as num).toDouble(),
      lineStyle: LineStyle.values.byName(json['line_style'] as String),
      isLocked: json['is_locked'] as bool,
      isSelected: json['is_selected'] as bool,
      label: json['label'] as String?,
    );
  }
}
