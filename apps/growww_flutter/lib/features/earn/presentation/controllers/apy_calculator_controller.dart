import 'dart:async';
import '../../domain/interfaces/i_earn_vault_repository.dart';
import '../../domain/models/apy_calculation_result.dart';
import '../../domain/models/earn_enums.dart';
import '../../domain/models/sub_paise_amount.dart';
import '../../domain/models/yield_vault_product.dart';
import '../../domain/services/yield_math_engine.dart';

class ApyCalculatorState {
  final YieldVaultProduct vault;
  final SubPaiseAmount principal;
  final TenorType tenor;
  final CompoundingFrequency compounding;
  final bool isSeniorCitizen;
  final bool hasForm15G15H;
  final ApyCalculationResult result;

  const ApyCalculatorState({
    required this.vault,
    required this.principal,
    required this.tenor,
    required this.compounding,
    required this.isSeniorCitizen,
    required this.hasForm15G15H,
    required this.result,
  });

  ApyCalculatorState copyWith({
    YieldVaultProduct? vault,
    SubPaiseAmount? principal,
    TenorType? tenor,
    CompoundingFrequency? compounding,
    bool? isSeniorCitizen,
    bool? hasForm15G15H,
    ApyCalculationResult? result,
  }) {
    return ApyCalculatorState(
      vault: vault ?? this.vault,
      principal: principal ?? this.principal,
      tenor: tenor ?? this.tenor,
      compounding: compounding ?? this.compounding,
      isSeniorCitizen: isSeniorCitizen ?? this.isSeniorCitizen,
      hasForm15G15H: hasForm15G15H ?? this.hasForm15G15H,
      result: result ?? this.result,
    );
  }
}

class ApyCalculatorController {
  final YieldVaultProduct vault;

  late ApyCalculatorState _state;
  final _stateController = StreamController<ApyCalculatorState>.broadcast();

  ApyCalculatorState get state => _state;
  Stream<ApyCalculatorState> get stream => _stateController.stream;

  ApyCalculatorController({required this.vault}) {
    final defaultPrincipal = SubPaiseAmount.fromInr(25000.0);
    const defaultTenor = TenorType.locked90Days;
    const defaultCompounding = CompoundingFrequency.daily;

    final initialResult = YieldMathEngine.computeReturns(
      principal: defaultPrincipal,
      baseApy: vault.baseApy,
      tenor: defaultTenor,
      compounding: defaultCompounding,
    );

    _state = ApyCalculatorState(
      vault: vault,
      principal: defaultPrincipal,
      tenor: defaultTenor,
      compounding: defaultCompounding,
      isSeniorCitizen: false,
      hasForm15G15H: false,
      result: initialResult,
    );
  }

  void _recalculate() {
    final res = YieldMathEngine.computeReturns(
      principal: _state.principal,
      baseApy: _state.vault.baseApy,
      tenor: _state.tenor,
      compounding: _state.compounding,
      isSeniorCitizen: _state.isSeniorCitizen,
      hasForm15G15H: _state.hasForm15G15H,
    );
    _state = _state.copyWith(result: res);
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  void setPrincipal(SubPaiseAmount amount) {
    _state = _state.copyWith(principal: amount);
    _recalculate();
  }

  void setTenor(TenorType tenor) {
    _state = _state.copyWith(tenor: tenor);
    _recalculate();
  }

  void setCompounding(CompoundingFrequency freq) {
    _state = _state.copyWith(compounding: freq);
    _recalculate();
  }

  void toggleSeniorCitizen(bool isSenior) {
    _state = _state.copyWith(isSeniorCitizen: isSenior);
    _recalculate();
  }

  void toggleForm15G15H(bool hasForm) {
    _state = _state.copyWith(hasForm15G15H: hasForm);
    _recalculate();
  }

  void dispose() {
    _stateController.close();
  }
}
