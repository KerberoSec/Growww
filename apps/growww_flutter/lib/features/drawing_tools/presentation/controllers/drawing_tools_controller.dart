import 'dart:async';
import '../../domain/algorithms/drawing_geometry_engine.dart';
import '../../domain/models/drawing_element.dart';
import '../../domain/models/drawing_enums.dart';
import '../../domain/models/drawing_point.dart';
import '../../domain/models/fibonacci_level.dart';

class DrawingToolsState {
  final DrawingToolType activeTool;
  final MagnetMode magnetMode;
  final int activeColorHex;
  final double activeStrokeWidth;
  final LineStyle activeLineStyle;
  final List<DrawingElement> elements;
  final List<List<DrawingElement>> undoStack;
  final List<List<DrawingElement>> redoStack;
  final String? selectedElementId;
  final List<DrawingPoint> inProgressPoints;
  final List<FibonacciLevel> activeFibonacciLevels;

  const DrawingToolsState({
    this.activeTool = DrawingToolType.cursor,
    this.magnetMode = MagnetMode.none,
    this.activeColorHex = 0xFF00F0A0,
    this.activeStrokeWidth = 2.0,
    this.activeLineStyle = LineStyle.solid,
    this.elements = const [],
    this.undoStack = const [],
    this.redoStack = const [],
    this.selectedElementId,
    this.inProgressPoints = const [],
    this.activeFibonacciLevels = const [],
  });

  DrawingToolsState copyWith({
    DrawingToolType? activeTool,
    MagnetMode? magnetMode,
    int? activeColorHex,
    double? activeStrokeWidth,
    LineStyle? activeLineStyle,
    List<DrawingElement>? elements,
    List<List<DrawingElement>>? undoStack,
    List<List<DrawingElement>>? redoStack,
    String? selectedElementId,
    bool clearSelected = false,
    List<DrawingPoint>? inProgressPoints,
    List<FibonacciLevel>? activeFibonacciLevels,
  }) {
    return DrawingToolsState(
      activeTool: activeTool ?? this.activeTool,
      magnetMode: magnetMode ?? this.magnetMode,
      activeColorHex: activeColorHex ?? this.activeColorHex,
      activeStrokeWidth: activeStrokeWidth ?? this.activeStrokeWidth,
      activeLineStyle: activeLineStyle ?? this.activeLineStyle,
      elements: elements ?? this.elements,
      undoStack: undoStack ?? this.undoStack,
      redoStack: redoStack ?? this.redoStack,
      selectedElementId: clearSelected ? null : (selectedElementId ?? this.selectedElementId),
      inProgressPoints: inProgressPoints ?? this.inProgressPoints,
      activeFibonacciLevels: activeFibonacciLevels ?? this.activeFibonacciLevels,
    );
  }
}

class DrawingToolsController {
  DrawingToolsState _state = const DrawingToolsState();
  final _stateController = StreamController<DrawingToolsState>.broadcast();

  DrawingToolsState get state => _state;
  Stream<DrawingToolsState> get stream => _stateController.stream;

  void _emit(DrawingToolsState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void setActiveTool(DrawingToolType tool) {
    _emit(_state.copyWith(
      activeTool: tool,
      inProgressPoints: [],
      clearSelected: true,
    ));
  }

  void setMagnetMode(MagnetMode mode) {
    _emit(_state.copyWith(magnetMode: mode));
  }

  void setColor(int colorHex) {
    _emit(_state.copyWith(activeColorHex: colorHex));
  }

  void setStrokeWidth(double width) {
    _emit(_state.copyWith(activeStrokeWidth: width));
  }

  void addPointToDrawing(DrawingPoint point) {
    if (_state.activeTool == DrawingToolType.cursor) return;

    final updatedPoints = List<DrawingPoint>.from(_state.inProgressPoints)..add(point);
    final required = _state.activeTool.requiredPoints;

    if (updatedPoints.length >= required) {
      final newElement = DrawingElement(
        elementId: 'DRAW-${DateTime.now().millisecondsSinceEpoch}',
        toolType: _state.activeTool,
        points: updatedPoints,
        strokeColorHex: _state.activeColorHex,
        strokeWidth: _state.activeStrokeWidth,
        lineStyle: _state.activeLineStyle,
      );

      final newUndo = List<List<DrawingElement>>.from(_state.undoStack)..add(_state.elements);
      final newElements = List<DrawingElement>.from(_state.elements)..add(newElement);

      List<FibonacciLevel> fibLevels = _state.activeFibonacciLevels;
      if (_state.activeTool == DrawingToolType.fibonacciRetracement && updatedPoints.length >= 2) {
        fibLevels = DrawingGeometryEngine.computeFibonacciLevels(
          anchorPrice1: updatedPoints[0].price,
          anchorPrice2: updatedPoints[1].price,
        );
      }

      _emit(_state.copyWith(
        elements: newElements,
        undoStack: newUndo,
        redoStack: [],
        inProgressPoints: [],
        activeFibonacciLevels: fibLevels,
      ));
    } else {
      _emit(_state.copyWith(inProgressPoints: updatedPoints));
    }
  }

  void undo() {
    if (_state.undoStack.isEmpty) return;
    final previous = _state.undoStack.last;
    final newUndo = List<List<DrawingElement>>.from(_state.undoStack)..removeLast();
    final newRedo = List<List<DrawingElement>>.from(_state.redoStack)..add(_state.elements);

    _emit(_state.copyWith(
      elements: previous,
      undoStack: newUndo,
      redoStack: newRedo,
      clearSelected: true,
    ));
  }

  void redo() {
    if (_state.redoStack.isEmpty) return;
    final next = _state.redoStack.last;
    final newRedo = List<List<DrawingElement>>.from(_state.redoStack)..removeLast();
    final newUndo = List<List<DrawingElement>>.from(_state.undoStack)..add(_state.elements);

    _emit(_state.copyWith(
      elements: next,
      undoStack: newUndo,
      redoStack: newRedo,
    ));
  }

  void deleteElement(String elementId) {
    final newUndo = List<List<DrawingElement>>.from(_state.undoStack)..add(_state.elements);
    final updated = _state.elements.where((e) => e.elementId != elementId).toList();
    _emit(_state.copyWith(
      elements: updated,
      undoStack: newUndo,
      clearSelected: true,
    ));
  }

  void clearAll() {
    if (_state.elements.isEmpty) return;
    final newUndo = List<List<DrawingElement>>.from(_state.undoStack)..add(_state.elements);
    _emit(_state.copyWith(
      elements: [],
      undoStack: newUndo,
      redoStack: [],
      clearSelected: true,
      activeFibonacciLevels: [],
    ));
  }

  void dispose() {
    _stateController.close();
  }
}
