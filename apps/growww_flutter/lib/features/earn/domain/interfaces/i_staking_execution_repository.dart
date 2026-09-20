import '../models/earn_enums.dart';
import '../models/sub_paise_amount.dart';
import '../models/user_stake_record.dart';

abstract class IStakingExecutionRepository {
  Future<UserStakeRecord> initiateStake({
    required String vaultId,
    required SubPaiseAmount amount,
    required TenorType tenor,
    required CompoundingFrequency compounding,
    required String biometricAuthToken,
    required String idempotencyKey,
  });

  Future<String> claimAccruedYield({
    required String stakeId,
    required ClaimDestination destination,
  });

  Future<String> initiateUnstake({
    required String stakeId,
    required bool isEmergencyExit,
    required String biometricAuthToken,
  });
}
