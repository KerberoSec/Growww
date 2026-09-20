import 'dart:async';
import '../../domain/interfaces/i_earn_vault_repository.dart';
import '../../domain/interfaces/i_staking_execution_repository.dart';
import '../../domain/models/apy_calculation_result.dart';
import '../../domain/models/earn_enums.dart';
import '../../domain/models/sub_paise_amount.dart';
import '../../domain/models/tds_withholding_estimate.dart';
import '../../domain/models/user_stake_record.dart';
import '../../domain/models/yield_vault_product.dart';
import '../../domain/services/yield_math_engine.dart';

class MockEarnVaultRepository implements IEarnVaultRepository, IStakingExecutionRepository {
  static final List<YieldVaultProduct> _vaults = [
    YieldVaultProduct(
      vaultId: 'VAULT-GSEC-718GS2033',
      vaultAddress: '0x1234567890123456789012345678901234567891',
      assetSymbol: 'gGSEC-718GS2033',
      displayName: '7.18% GS 2033 Sovereign Vault',
      assetType: YieldAssetType.gGsec,
      isin: 'IN0020230085',
      issuerName: 'Government of India / RBI',
      creditRating: 'SOV',
      baseApy: 7.18,
      tenorBonusApy: const {
        TenorType.flexible: 0.0,
        TenorType.locked30Days: 0.25,
        TenorType.locked90Days: 0.50,
        TenorType.locked180Days: 0.85,
        TenorType.locked365Days: 1.25,
      },
      minStakeAmount: SubPaiseAmount.fromInr(100.0),
      totalValueLocked: SubPaiseAmount.fromInr(125000000.0),
      maxPoolCapacity: SubPaiseAmount.fromInr(250000000.0),
      riskTier: RiskTier.sovereignRiskFree,
      bondMaturityDate: DateTime(2033, 8, 14),
      modifiedDurationYears: 6.82,
      isAcceptingStakes: true,
      proofOfReserveRootHash: '0x3a4b5c6d7e8f901234567890abcdef1234567890',
    ),
    YieldVaultProduct(
      vaultId: 'VAULT-TBILL-91D',
      vaultAddress: '0x2345678901234567890123456789012345678912',
      assetSymbol: 'gTBILL-91D',
      displayName: '91-Day Sovereign Treasury Bill Vault',
      assetType: YieldAssetType.gTbill,
      isin: 'IN002024X091',
      issuerName: 'Reserve Bank of India',
      creditRating: 'SOV',
      baseApy: 6.75,
      tenorBonusApy: const {
        TenorType.flexible: 0.0,
        TenorType.locked30Days: 0.15,
        TenorType.locked90Days: 0.40,
        TenorType.locked180Days: 0.60,
        TenorType.locked365Days: 0.90,
      },
      minStakeAmount: SubPaiseAmount.fromInr(100.0),
      totalValueLocked: SubPaiseAmount.fromInr(85000000.0),
      maxPoolCapacity: SubPaiseAmount.fromInr(100000000.0),
      riskTier: RiskTier.sovereignRiskFree,
      bondMaturityDate: DateTime.now().add(const Duration(days: 91)),
      modifiedDurationYears: 0.25,
      isAcceptingStakes: true,
      proofOfReserveRootHash: '0x4b5c6d7e8f901234567890abcdef12345678901a',
    ),
    YieldVaultProduct(
      vaultId: 'VAULT-CORP-HDFC',
      vaultAddress: '0x3456789012345678901234567890123456789123',
      assetSymbol: 'gCORP-HDFC2028',
      displayName: '8.20% HDFC Bank AAA Corporate Vault',
      assetType: YieldAssetType.gCorp,
      isin: 'INE040A08435',
      issuerName: 'HDFC Bank Limited',
      creditRating: 'CRISIL AAA',
      baseApy: 8.20,
      tenorBonusApy: const {
        TenorType.flexible: 0.0,
        TenorType.locked30Days: 0.35,
        TenorType.locked90Days: 0.70,
        TenorType.locked180Days: 1.10,
        TenorType.locked365Days: 1.65,
      },
      minStakeAmount: SubPaiseAmount.fromInr(500.0),
      totalValueLocked: SubPaiseAmount.fromInr(45000000.0),
      maxPoolCapacity: SubPaiseAmount.fromInr(50000000.0),
      riskTier: RiskTier.institutionalLowRisk,
      bondMaturityDate: DateTime(2028, 5, 20),
      modifiedDurationYears: 3.45,
      isAcceptingStakes: true,
      proofOfReserveRootHash: '0x5c6d7e8f901234567890abcdef12345678901a2b',
    ),
  ];

  final List<UserStakeRecord> _activeStakes = [
    UserStakeRecord(
      stakeId: 'STK-GS-99014',
      vaultId: 'VAULT-GSEC-718GS2033',
      assetSymbol: 'gGSEC-718GS2033',
      principalAmount: SubPaiseAmount.fromInr(50000.0),
      tenorType: TenorType.locked90Days,
      agreedApy: 7.68,
      compoundingFrequency: CompoundingFrequency.daily,
      status: StakeStatus.active,
      stakedAt: DateTime.now().subtract(const Duration(days: 45)),
      lockedUntil: DateTime.now().add(const Duration(days: 45)),
      totalAccruedYield: SubPaiseAmount.fromInr(473.4215),
      unallocatedYield: SubPaiseAmount.fromInr(128.5420),
      alreadyClaimedYield: SubPaiseAmount.fromInr(344.8795),
      lastYieldCheckpoint: DateTime.now(),
      onChainTxHash: '0x7a8b9c0d1e2f3a4b5c6d7e8f901234567890abcdef',
    ),
  ];

  @override
  Future<List<YieldVaultProduct>> fetchYieldVaults({
    YieldAssetType? filterType,
  }) async {
    if (filterType == null) return _vaults;
    return _vaults.where((v) => v.assetType == filterType).toList();
  }

  @override
  Future<YieldVaultProduct> fetchVaultDetails({
    required String vaultId,
  }) async {
    return _vaults.firstWhere((v) => v.vaultId == vaultId, orElse: () => _vaults.first);
  }

  @override
  Future<List<UserStakeRecord>> fetchUserActiveStakes() async {
    return _activeStakes;
  }

  @override
  Future<ApyCalculationResult> calculateApyReturns({
    required String vaultId,
    required SubPaiseAmount principal,
    required TenorType tenor,
    required CompoundingFrequency compounding,
  }) async {
    final vault = await fetchVaultDetails(vaultId: vaultId);
    return YieldMathEngine.computeReturns(
      principal: principal,
      baseApy: vault.baseApy,
      tenor: tenor,
      compounding: compounding,
    );
  }

  @override
  Stream<SubPaiseAmount> streamUserLiveAccrual({
    required String userId,
  }) {
    int microTick = 0;
    return Stream.periodic(const Duration(milliseconds: 500), (_) {
      microTick += 1;
      // Continuous sub-second micro-accrual: 0.0001 INR increments
      return SubPaiseAmount.fromInr(128.5420 + (microTick * 0.0002));
    });
  }

  @override
  Future<TdsWithholdingEstimate> estimateTdsWithholding({
    required SubPaiseAmount interestAmount,
  }) async {
    return YieldMathEngine.estimateTds(grossInterest: interestAmount);
  }

  @override
  Future<UserStakeRecord> initiateStake({
    required String vaultId,
    required SubPaiseAmount amount,
    required TenorType tenor,
    required CompoundingFrequency compounding,
    required String biometricAuthToken,
    required String idempotencyKey,
  }) async {
    final vault = await fetchVaultDetails(vaultId: vaultId);
    final apy = vault.getEffectiveApy(tenor);

    final newStake = UserStakeRecord(
      stakeId: 'STK-${DateTime.now().millisecondsSinceEpoch}',
      vaultId: vaultId,
      assetSymbol: vault.assetSymbol,
      principalAmount: amount,
      tenorType: tenor,
      agreedApy: apy,
      compoundingFrequency: compounding,
      status: StakeStatus.active,
      stakedAt: DateTime.now(),
      lockedUntil: tenor.durationDays > 0 ? DateTime.now().add(Duration(days: tenor.durationDays)) : null,
      totalAccruedYield: SubPaiseAmount.zero(),
      unallocatedYield: SubPaiseAmount.zero(),
      alreadyClaimedYield: SubPaiseAmount.zero(),
      lastYieldCheckpoint: DateTime.now(),
      onChainTxHash: '0x8b9c0d1e2f3a4b5c6d7e8f901234567890abcdef12',
    );

    _activeStakes.add(newStake);
    return newStake;
  }

  @override
  Future<String> claimAccruedYield({
    required String stakeId,
    required ClaimDestination destination,
  }) async {
    final idx = _activeStakes.indexWhere((s) => s.stakeId == stakeId);
    if (idx != -1) {
      final s = _activeStakes[idx];
      final claimed = s.unallocatedYield;
      _activeStakes[idx] = s.copyWith(
        unallocatedYield: SubPaiseAmount.zero(),
        alreadyClaimedYield: s.alreadyClaimedYield + claimed,
        lastYieldCheckpoint: DateTime.now(),
      );
    }
    return '0xclaim99887766554433221100aabbccddeeff00112233';
  }

  @override
  Future<String> initiateUnstake({
    required String stakeId,
    required bool isEmergencyExit,
    required String biometricAuthToken,
  }) async {
    final idx = _activeStakes.indexWhere((s) => s.stakeId == stakeId);
    if (idx != -1) {
      _activeStakes[idx] = _activeStakes[idx].copyWith(
        status: StakeStatus.completed,
      );
    }
    return '0xunstake11223344556677889900aabbccddeeff001122';
  }
}
