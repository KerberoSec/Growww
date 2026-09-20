import 'dart:async';
import '../../domain/algorithms/iceberg_slicing_allocator.dart';
import '../../domain/models/iceberg_enums.dart';
import '../../domain/models/iceberg_order_config.dart';
import '../../domain/models/iceberg_tranche.dart';

class IcebergOrderState {
  final String symbol;
  final String side; // BUY or SELL
  final double totalQuantity;
  final double visibleTrancheQuantity;
  final IcebergVarianceMode varianceMode;
  final double priceLimit;
  final List<IcebergTranche> simulatedTranches;
  final IcebergSlicingValidationResult validation;
  final bool isSubmitting;
  final String? submittedOrderId;
  final String? toastMessage;

  const IcebergOrderState({
    this.symbol = 'BTC/USDT',
    this.side = 'BUY',
    this.totalQuantity = 10.0,
    this.visibleTrancheQuantity = 1.0,
    this.varianceMode = IcebergVarianceMode.randomizedVariance5to10,
    this.priceLimit = 64250.0,
    this.simulatedTranches = const [],
    this.validation = const IcebergSlicingValidationResult(isValid: true),
    this.isSubmitting = false,
    this.submittedOrderId,
    this.toastMessage,
  });

  double get estimatedTotalInr => totalQuantity * priceLimit;
  double get hiddenQuantity => (totalQuantity - visibleTrancheQuantity).clamp(0.0, totalQuantity);

  IcebergOrderConfig toConfig() {
    return IcebergOrderConfig(
      orderId: 'ORD-ICE-${DateTime.now().millisecondsSinceEpoch}',
      symbol: symbol,
      side: side,
      totalQuantity: totalQuantity,
      visibleTrancheQuantity: visibleTrancheQuantity,
      varianceMode: varianceMode,
      priceLimit: priceLimit,
      createdAt: DateTime.now(),
    );
  }

  IcebergOrderState copyWith({
    String? symbol,
    String? side,
    double? totalQuantity,
    double? visibleTrancheQuantity,
    IcebergVarianceMode? varianceMode,
    double? priceLimit,
    List<IcebergTranche>? simulatedTranches,
    IcebergSlicingValidationResult? validation,
    bool? isSubmitting,
    String? submittedOrderId,
    String? toastMessage,
  }) {
    return IcebergOrderState(
      symbol: symbol ?? this.symbol,
      side: side ?? this.side,
      totalQuantity: totalQuantity ?? this.totalQuantity,
      visibleTrancheQuantity: visibleTrancheQuantity ?? this.visibleTrancheQuantity,
      varianceMode: varianceMode ?? this.varianceMode,
      priceLimit: priceLimit ?? this.priceLimit,
      simulatedTranches: simulatedTranches ?? this.simulatedTranches,
      validation: validation ?? this.validation,
      isSubmitting: isSubmitting ?? this.isSubmitting,
      submittedOrderId: submittedOrderId,
      toastMessage: toastMessage,
    );
  }
}

class IcebergOrderController {
  IcebergOrderState _state = const IcebergOrderState();
  final _stateController = StreamController<IcebergOrderState>.broadcast();

  IcebergOrderState get state => _state;
  Stream<IcebergOrderState> get stream => _stateController.stream;

  IcebergOrderController() {
    _recalculateTranches();
  }

  void _emit(IcebergOrderState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void _recalculateTranches() {
    final config = _state.toConfig();
    final validation = IcebergSlicingAllocator.validateConfig(config);
    final tranches = validation.isValid
        ? IcebergSlicingAllocator.allocateTranches(config)
        : <IcebergTranche>[];

    _emit(_state.copyWith(
      validation: validation,
      simulatedTranches: tranches,
    ));
  }

  void setSide(String side) {
    _emit(_state.copyWith(side: side));
    _recalculateTranches();
  }

  void setTotalQuantity(double qty) {
    _emit(_state.copyWith(totalQuantity: qty));
    _recalculateTranches();
  }

  void setVisibleTrancheQuantity(double trancheQty) {
    _emit(_state.copyWith(visibleTrancheQuantity: trancheQty));
    _recalculateTranches();
  }

  void setVarianceMode(IcebergVarianceMode mode) {
    _emit(_state.copyWith(varianceMode: mode));
    _recalculateTranches();
  }

  void setPriceLimit(double price) {
    _emit(_state.copyWith(priceLimit: price));
    _recalculateTranches();
  }

  Future<bool> submitIcebergOrder() async {
    _recalculateTranches();
    if (!_state.validation.isValid) return false;

    _emit(_state.copyWith(isSubmitting: true));
    await Future.delayed(const Duration(milliseconds: 300));

    final orderId = 'ORD-ICE-${DateTime.now().millisecondsSinceEpoch}';
    _emit(_state.copyWith(
      isSubmitting: false,
      submittedOrderId: orderId,
      toastMessage: 'Algorithmic Iceberg Order Dispatched (${_state.simulatedTranches.length} Tranches)',
    ));
    return true;
  }

  void dispose() {
    _stateController.close();
  }
}
