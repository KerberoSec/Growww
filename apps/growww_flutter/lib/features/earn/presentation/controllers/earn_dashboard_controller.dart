import 'dart:async';
import '../../domain/interfaces/i_earn_vault_repository.dart';
import '../../domain/interfaces/i_staking_execution_repository.dart';
import '../../domain/models/earn_enums.dart';
import '../../domain/models/sub_paise_amount.dart';
import '../../domain/models/user_stake_record.dart';
import '../../domain/models/yield_vault_product.dart';

class EarnDashboardState {
  final YieldAssetType? selectedFilter;
  final List<YieldVaultProduct> vaults;
  final List<UserStakeRecord> activeStakes;
  final SubPaiseAmount liveAccruedYield;
  final bool isLoading;
  final String? errorMessage;
  final String? successMessage;

  const EarnDashboardState({
    this.selectedFilter,
    this.vaults = const [],
    this.activeStakes = const [],
    this.liveAccruedYield = const SubPaiseAmount(1285420), // 128.5420 INR
    this.isLoading = false,
    this.errorMessage,
    this.successMessage,
  });

  SubPaiseAmount get totalStakedPrincipal {
    int total = 0;
    for (final s in activeStakes) {
      total += s.principalAmount.subPaiseValue;
    }
    return SubPaiseAmount(total);
  }

  SubPaiseAmount get totalAccruedYield {
    int total = 0;
    for (final s in activeStakes) {
      total += s.totalAccruedYield.subPaiseValue;
    }
    return SubPaiseAmount(total);
  }

  EarnDashboardState copyWith({
    YieldAssetType? selectedFilter,
    bool clearFilter = false,
    List<YieldVaultProduct>? vaults,
    List<UserStakeRecord>? activeStakes,
    SubPaiseAmount? liveAccruedYield,
    bool? isLoading,
    String? errorMessage,
    String? successMessage,
  }) {
    return EarnDashboardState(
      selectedFilter: clearFilter ? null : (selectedFilter ?? this.selectedFilter),
      vaults: vaults ?? this.vaults,
      activeStakes: activeStakes ?? this.activeStakes,
      liveAccruedYield: liveAccruedYield ?? this.liveAccruedYield,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      successMessage: successMessage,
    );
  }
}

class EarnDashboardController {
  final IEarnVaultRepository _vaultRepo;
  final IStakingExecutionRepository _stakingRepo;

  EarnDashboardState _state = const EarnDashboardState();
  final _stateController = StreamController<EarnDashboardState>.broadcast();
  StreamSubscription<SubPaiseAmount>? _accrualSubscription;

  EarnDashboardState get state => _state;
  Stream<EarnDashboardState> get stream => _stateController.stream;

  EarnDashboardController({
    required IEarnVaultRepository vaultRepo,
    required IStakingExecutionRepository stakingRepo,
  })  : _vaultRepo = vaultRepo,
        _stakingRepo = stakingRepo {
    loadDashboard();
    _subscribeAccrual();
  }

  void _emit(EarnDashboardState newState) {
    _state = newState;
    if (!_stateController.isClosed) {
      _stateController.add(_state);
    }
  }

  Future<void> loadDashboard() async {
    _emit(_state.copyWith(isLoading: true));
    try {
      final vaults = await _vaultRepo.fetchYieldVaults(filterType: _state.selectedFilter);
      final stakes = await _vaultRepo.fetchUserActiveStakes();
      _emit(_state.copyWith(
        vaults: vaults,
        activeStakes: stakes,
        isLoading: false,
      ));
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
    }
  }

  void setFilter(YieldAssetType? filter) {
    if (filter == _state.selectedFilter) {
      _emit(_state.copyWith(clearFilter: true));
    } else {
      _emit(_state.copyWith(selectedFilter: filter));
    }
    loadDashboard();
  }

  Future<bool> claimYield(String stakeId, ClaimDestination destination) async {
    _emit(_state.copyWith(isLoading: true));
    try {
      final tx = await _stakingRepo.claimAccruedYield(stakeId: stakeId, destination: destination);
      final stakes = await _vaultRepo.fetchUserActiveStakes();
      _emit(_state.copyWith(
        activeStakes: stakes,
        isLoading: false,
        successMessage: 'Yield claimed successfully! Tx: ${tx.substring(0, 14)}...',
      ));
      return true;
    } catch (e) {
      _emit(_state.copyWith(isLoading: false, errorMessage: e.toString()));
      return false;
    }
  }

  void _subscribeAccrual() {
    _accrualSubscription?.cancel();
    _accrualSubscription = _vaultRepo.streamUserLiveAccrual(userId: 'USR-NBSE-7749').listen((accrual) {
      _emit(_state.copyWith(liveAccruedYield: accrual));
    });
  }

  void dispose() {
    _accrualSubscription?.cancel();
    _stateController.close();
  }
}
