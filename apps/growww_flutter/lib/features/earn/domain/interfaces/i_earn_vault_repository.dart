import '../models/apy_calculation_result.dart';
import '../models/earn_enums.dart';
import '../models/sub_paise_amount.dart';
import '../models/tds_withholding_estimate.dart';
import '../models/user_stake_record.dart';
import '../models/yield_vault_product.dart';

abstract class IEarnVaultRepository {
  Future<List<YieldVaultProduct>> fetchYieldVaults({
    YieldAssetType? filterType,
  });

  Future<YieldVaultProduct> fetchVaultDetails({
    required String vaultId,
  });

  Future<List<UserStakeRecord>> fetchUserActiveStakes();

  Future<ApyCalculationResult> calculateApyReturns({
    required String vaultId,
    required SubPaiseAmount principal,
    required TenorType tenor,
    required CompoundingFrequency compounding,
  });

  Stream<SubPaiseAmount> streamUserLiveAccrual({
    required String userId,
  });

  Future<TdsWithholdingEstimate> estimateTdsWithholding({
    required SubPaiseAmount interestAmount,
  });
}
