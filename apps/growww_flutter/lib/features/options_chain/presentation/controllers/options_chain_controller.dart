import 'dart:async';
import '../../domain/models/option_contract_quote.dart';
import '../../domain/models/options_chain_snapshot.dart';
import '../../domain/models/options_enums.dart';
import '../../domain/services/options_analytics_engine.dart';

class OptionsChainState {
  final OptionsChainSnapshot snapshot;
  final OptionsViewFilter filter;
  final bool showGreeks;
  final double? selectedStrike;
  final OptionContractQuote? selectedContract;
  final String? stagedSide; // 'BUY' or 'SELL'
  final double stagedQuantity;
  final String? toastMessage;
  final bool isSubmittingOrder;

  const OptionsChainState({
    required this.snapshot,
    this.filter = OptionsViewFilter.all,
    this.showGreeks = false,
    this.selectedStrike,
    this.selectedContract,
    this.stagedSide,
    this.stagedQuantity = 1.0,
    this.toastMessage,
    this.isSubmittingOrder = false,
  });

  double get estimatedOrderValue =>
      (selectedContract != null) ? selectedContract!.ltp * stagedQuantity : 0.0;

  OptionsChainState copyWith({
    OptionsChainSnapshot? snapshot,
    OptionsViewFilter? filter,
    bool? showGreeks,
    double? selectedStrike,
    OptionContractQuote? selectedContract,
    String? stagedSide,
    double? stagedQuantity,
    String? toastMessage,
    bool? isSubmittingOrder,
    bool clearSelectedContract = false,
  }) {
    return OptionsChainState(
      snapshot: snapshot ?? this.snapshot,
      filter: filter ?? this.filter,
      showGreeks: showGreeks ?? this.showGreeks,
      selectedStrike: clearSelectedContract ? null : (selectedStrike ?? this.selectedStrike),
      selectedContract: clearSelectedContract ? null : (selectedContract ?? this.selectedContract),
      stagedSide: clearSelectedContract ? null : (stagedSide ?? this.stagedSide),
      stagedQuantity: stagedQuantity ?? this.stagedQuantity,
      toastMessage: toastMessage,
      isSubmittingOrder: isSubmittingOrder ?? this.isSubmittingOrder,
    );
  }
}

class OptionsChainController {
  late OptionsChainState _state;
  final _stateController = StreamController<OptionsChainState>.broadcast();

  OptionsChainState get state => _state;
  Stream<OptionsChainState> get stream => _stateController.stream;

  OptionsChainController({
    String symbol = 'BTC/USDT',
    double initialSpot = 64250.0,
  }) {
    final expiries = ['26 SEP 2026', '03 OCT 2026', '31 OCT 2026', '25 DEC 2026'];
    final initialSnapshot = OptionsAnalyticsEngine.generateMockChain(
      symbol: symbol,
      spotPrice: initialSpot,
      selectedExpiry: expiries.first,
      expiries: expiries,
    );

    _state = OptionsChainState(snapshot: initialSnapshot);
    _emit(_state);
  }

  void _emit(OptionsChainState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void selectExpiry(String expiry) {
    if (expiry == _state.snapshot.selectedExpiry) return;

    final updatedSnapshot = OptionsAnalyticsEngine.generateMockChain(
      symbol: _state.snapshot.underlyingSymbol,
      spotPrice: _state.snapshot.spotPrice,
      selectedExpiry: expiry,
      expiries: _state.snapshot.availableExpiries,
    );

    _emit(_state.copyWith(
      snapshot: updatedSnapshot,
      clearSelectedContract: true,
      toastMessage: 'Loaded expiry $expiry',
    ));
  }

  void setFilter(OptionsViewFilter filter) {
    _emit(_state.copyWith(filter: filter));
  }

  void toggleGreeks() {
    _emit(_state.copyWith(showGreeks: !_state.showGreeks));
  }

  void selectContract(OptionContractQuote contract, String side) {
    _emit(_state.copyWith(
      selectedContract: contract,
      selectedStrike: contract.strikePrice,
      stagedSide: side.toUpperCase(),
      stagedQuantity: 1.0,
      toastMessage: 'Staged $side order for ${contract.symbol}',
    ));
  }

  void setStagedQuantity(double qty) {
    if (qty <= 0) return;
    _emit(_state.copyWith(stagedQuantity: qty));
  }

  void clearSelection() {
    _emit(_state.copyWith(
      clearSelectedContract: true,
      toastMessage: null,
    ));
  }

  void updateSpotPrice(double newSpot) {
    final updated = OptionsAnalyticsEngine.generateMockChain(
      symbol: _state.snapshot.underlyingSymbol,
      spotPrice: newSpot,
      selectedExpiry: _state.snapshot.selectedExpiry,
      expiries: _state.snapshot.availableExpiries,
    );
    _emit(_state.copyWith(snapshot: updated));
  }

  Future<bool> placeStagedOrder() async {
    if (_state.selectedContract == null || _state.stagedSide == null) return false;

    _emit(_state.copyWith(isSubmittingOrder: true));
    await Future.delayed(const Duration(milliseconds: 50));

    final orderSummary =
        '${_state.stagedSide} ${_state.stagedQuantity}x ${_state.selectedContract!.symbol} filled at ₹${_state.selectedContract!.ltp.toStringAsFixed(2)}';

    _emit(_state.copyWith(
      isSubmittingOrder: false,
      clearSelectedContract: true,
      toastMessage: orderSummary,
    ));

    return true;
  }

  void dispose() {
    _stateController.close();
  }
}
