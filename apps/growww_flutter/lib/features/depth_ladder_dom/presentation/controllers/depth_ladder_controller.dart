import 'dart:async';
import '../../domain/models/depth_ladder_enums.dart';
import '../../domain/models/ladder_dom_snapshot.dart';
import '../../domain/services/depth_ladder_engine.dart';

class DepthLadderState {
  final LadderDomSnapshot snapshot;
  final LadderTickAggregation aggregation;
  final DepthMode depthMode;
  final double? selectedPrice;
  final DomOrderSide? stagedSide;
  final double stagedQuantity;
  final bool isOrderStaged;
  final bool isExecuting;
  final String? toastMessage;

  const DepthLadderState({
    required this.snapshot,
    this.aggregation = LadderTickAggregation.one,
    this.depthMode = DepthMode.level2Aggregated,
    this.selectedPrice,
    this.stagedSide,
    this.stagedQuantity = 1.0,
    this.isOrderStaged = false,
    this.isExecuting = false,
    this.toastMessage,
  });

  double get estimatedNotional =>
      (selectedPrice != null) ? selectedPrice! * stagedQuantity : 0.0;

  DepthLadderState copyWith({
    LadderDomSnapshot? snapshot,
    LadderTickAggregation? aggregation,
    DepthMode? depthMode,
    double? selectedPrice,
    DomOrderSide? stagedSide,
    double? stagedQuantity,
    bool? isOrderStaged,
    bool? isExecuting,
    String? toastMessage,
    bool clearStaged = false,
  }) {
    return DepthLadderState(
      snapshot: snapshot ?? this.snapshot,
      aggregation: aggregation ?? this.aggregation,
      depthMode: depthMode ?? this.depthMode,
      selectedPrice: clearStaged ? null : (selectedPrice ?? this.selectedPrice),
      stagedSide: clearStaged ? null : (stagedSide ?? this.stagedSide),
      stagedQuantity: stagedQuantity ?? this.stagedQuantity,
      isOrderStaged: clearStaged ? false : (isOrderStaged ?? this.isOrderStaged),
      isExecuting: isExecuting ?? this.isExecuting,
      toastMessage: toastMessage,
    );
  }
}

class DepthLadderController {
  late DepthLadderState _state;
  final _stateController = StreamController<DepthLadderState>.broadcast();

  DepthLadderState get state => _state;
  Stream<DepthLadderState> get stream => _stateController.stream;

  DepthLadderController({
    String symbol = 'BTC/USDT',
    double centerPrice = 64250.0,
    LadderTickAggregation aggregation = LadderTickAggregation.one,
  }) {
    final initialSnapshot = DepthLadderEngine.generateSnapshot(
      symbol: symbol,
      centerPrice: centerPrice,
      aggregation: aggregation,
    );

    _state = DepthLadderState(
      snapshot: initialSnapshot,
      aggregation: aggregation,
    );
    _emit(_state);
  }

  void _emit(DepthLadderState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void setAggregation(LadderTickAggregation agg) {
    if (agg == _state.aggregation) return;

    final updatedSnapshot = DepthLadderEngine.generateSnapshot(
      symbol: _state.snapshot.symbol,
      centerPrice: _state.snapshot.lastTradedPrice,
      aggregation: agg,
      depthMode: _state.depthMode,
    );

    _emit(_state.copyWith(
      aggregation: agg,
      snapshot: updatedSnapshot,
      clearStaged: true,
      toastMessage: 'Tick size set to ${agg.label}',
    ));
  }

  void setDepthMode(DepthMode mode) {
    if (mode == _state.depthMode) return;

    final updatedSnapshot = DepthLadderEngine.generateSnapshot(
      symbol: _state.snapshot.symbol,
      centerPrice: _state.snapshot.lastTradedPrice,
      aggregation: _state.aggregation,
      depthMode: mode,
    );

    _emit(_state.copyWith(
      depthMode: mode,
      snapshot: updatedSnapshot,
      toastMessage: null,
    ));
  }

  void selectPriceLevel(double price, DomOrderSide side) {
    _emit(_state.copyWith(
      selectedPrice: price,
      stagedSide: side,
      isOrderStaged: true,
      toastMessage: 'Staged ${side.displayName} Limit @ ₹${price.toStringAsFixed(2)}',
    ));
  }

  void setStagedQuantity(double qty) {
    if (qty <= 0) return;
    _emit(_state.copyWith(stagedQuantity: qty));
  }

  void clearStagedOrder() {
    _emit(_state.copyWith(clearStaged: true, toastMessage: null));
  }

  Future<bool> executeStagedOrder() async {
    if (!_state.isOrderStaged || _state.selectedPrice == null || _state.stagedSide == null) {
      return false;
    }

    _emit(_state.copyWith(isExecuting: true));
    await Future.delayed(const Duration(milliseconds: 50));

    final sideStr = _state.stagedSide!.displayName;
    final p = _state.selectedPrice!.toStringAsFixed(2);
    final q = _state.stagedQuantity;
    final summary = 'Limit $sideStr placed: $q ${_state.snapshot.symbol.split('/').first} @ ₹$p';

    _emit(_state.copyWith(
      isExecuting: false,
      clearStaged: true,
      toastMessage: summary,
    ));

    return true;
  }

  /// One-tap quick place order from ladder action rows.
  Future<bool> quickOrder(DomOrderSide side, double price, double quantity) async {
    final sideStr = side.displayName;
    final p = price.toStringAsFixed(2);
    final summary = 'Quick $sideStr executed: $quantity @ ₹$p';

    _emit(_state.copyWith(toastMessage: summary));
    return true;
  }

  void centerOnLastPrice() {
    final updated = DepthLadderEngine.generateSnapshot(
      symbol: _state.snapshot.symbol,
      centerPrice: _state.snapshot.lastTradedPrice,
      aggregation: _state.aggregation,
      depthMode: _state.depthMode,
    );
    _emit(_state.copyWith(snapshot: updated, toastMessage: 'Centered on last price'));
  }

  void dispose() {
    _stateController.close();
  }
}
