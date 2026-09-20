import 'dart:async';
import '../../domain/models/trade_confirmation_models.dart';
import '../../domain/services/haptic_trade_confirmation_service.dart';

class HapticTradeConfirmationState {
  final SwipeConfirmationState state;
  final double progress;
  final String? signature;
  final String? errorMessage;

  const HapticTradeConfirmationState({
    required this.state,
    this.progress = 0.0,
    this.signature,
    this.errorMessage,
  });

  HapticTradeConfirmationState copyWith({
    SwipeConfirmationState? state,
    double? progress,
    String? signature,
    String? errorMessage,
  }) {
    return HapticTradeConfirmationState(
      state: state ?? this.state,
      progress: progress ?? this.progress,
      signature: signature ?? this.signature,
      errorMessage: errorMessage ?? this.errorMessage,
    );
  }
}

class HapticTradeConfirmationController {
  final HapticTradeConfirmationService service;
  final TradeConfirmationIntent intent;

  HapticTradeConfirmationState _state = const HapticTradeConfirmationState(
    state: SwipeConfirmationState.idle,
    progress: 0.0,
  );

  final _stateController = StreamController<HapticTradeConfirmationState>.broadcast();
  Stream<HapticTradeConfirmationState> get stream => _stateController.stream;
  HapticTradeConfirmationState get state => _state;

  HapticTradeConfirmationController({
    required this.service,
    required this.intent,
  });

  void updateSliderProgress(double newProgress) {
    if (_state.state == SwipeConfirmationState.submitted ||
        _state.state == SwipeConfirmationState.biometricPrompted) {
      return;
    }

    final clamped = newProgress.clamp(0.0, 1.0);
    service.onSliderProgressChanged(_state.progress, clamped);

    if (service.isSwipeThresholdReached(clamped)) {
      _setState(_state.copyWith(
        progress: clamped,
        state: SwipeConfirmationState.thresholdReached,
      ));
    } else {
      _setState(_state.copyWith(
        progress: clamped,
        state: clamped > 0 ? SwipeConfirmationState.swiping : SwipeConfirmationState.idle,
      ));
    }
  }

  Future<bool> executeTradeConfirmation() async {
    if (!service.isSwipeThresholdReached(_state.progress)) {
      resetSlider();
      return false;
    }

    _setState(_state.copyWith(state: SwipeConfirmationState.biometricPrompted));

    final sig = await service.confirmAndSignOrder(
      intent: intent,
      sliderProgress: _state.progress,
    );

    if (sig != null) {
      _setState(_state.copyWith(
        state: SwipeConfirmationState.submitted,
        signature: sig,
        progress: 1.0,
      ));
      return true;
    } else {
      _setState(_state.copyWith(
        state: SwipeConfirmationState.failed,
        errorMessage: 'Biometric authorization failed or cancelled',
      ));
      return false;
    }
  }

  void resetSlider() {
    _setState(const HapticTradeConfirmationState(
      state: SwipeConfirmationState.idle,
      progress: 0.0,
    ));
  }

  void _setState(HapticTradeConfirmationState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void dispose() {
    _stateController.close();
  }
}
