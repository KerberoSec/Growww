import 'earn_enums.dart';
import 'sub_paise_amount.dart';

class UserStakeRecord {
  final String stakeId;
  final String vaultId;
  final String assetSymbol;
  final SubPaiseAmount principalAmount;
  final TenorType tenorType;
  final double agreedApy;
  final CompoundingFrequency compoundingFrequency;
  final StakeStatus status;
  final DateTime stakedAt;
  final DateTime? lockedUntil;
  final SubPaiseAmount totalAccruedYield;
  final SubPaiseAmount unallocatedYield;
  final SubPaiseAmount alreadyClaimedYield;
  final DateTime lastYieldCheckpoint;
  final String onChainTxHash;

  const UserStakeRecord({
    required this.stakeId,
    required this.vaultId,
    required this.assetSymbol,
    required this.principalAmount,
    required this.tenorType,
    required this.agreedApy,
    required this.compoundingFrequency,
    required this.status,
    required this.stakedAt,
    this.lockedUntil,
    required this.totalAccruedYield,
    required this.unallocatedYield,
    required this.alreadyClaimedYield,
    required this.lastYieldCheckpoint,
    required this.onChainTxHash,
  });

  bool get isLocked =>
      tenorType != TenorType.flexible &&
      lockedUntil != null &&
      DateTime.now().isBefore(lockedUntil!);

  UserStakeRecord copyWith({
    SubPaiseAmount? totalAccruedYield,
    SubPaiseAmount? unallocatedYield,
    SubPaiseAmount? alreadyClaimedYield,
    StakeStatus? status,
    DateTime? lastYieldCheckpoint,
  }) {
    return UserStakeRecord(
      stakeId: stakeId,
      vaultId: vaultId,
      assetSymbol: assetSymbol,
      principalAmount: principalAmount,
      tenorType: tenorType,
      agreedApy: agreedApy,
      compoundingFrequency: compoundingFrequency,
      status: status ?? this.status,
      stakedAt: stakedAt,
      lockedUntil: lockedUntil,
      totalAccruedYield: totalAccruedYield ?? this.totalAccruedYield,
      unallocatedYield: unallocatedYield ?? this.unallocatedYield,
      alreadyClaimedYield: alreadyClaimedYield ?? this.alreadyClaimedYield,
      lastYieldCheckpoint: lastYieldCheckpoint ?? this.lastYieldCheckpoint,
      onChainTxHash: onChainTxHash,
    );
  }

  Map<String, dynamic> toJson() => {
        'stake_id': stakeId,
        'vault_id': vaultId,
        'asset_symbol': assetSymbol,
        'principal_amount': principalAmount.toJson(),
        'tenor_type': tenorType.name,
        'agreed_apy': agreedApy,
        'compounding_frequency': compoundingFrequency.name,
        'status': status.name,
        'staked_at': stakedAt.toIso8601String(),
        'locked_until': lockedUntil?.toIso8601String(),
        'total_accrued_yield': totalAccruedYield.toJson(),
        'unallocated_yield': unallocatedYield.toJson(),
        'already_claimed_yield': alreadyClaimedYield.toJson(),
        'last_yield_checkpoint': lastYieldCheckpoint.toIso8601String(),
        'on_chain_tx_hash': onChainTxHash,
      };

  factory UserStakeRecord.fromJson(Map<String, dynamic> json) {
    return UserStakeRecord(
      stakeId: json['stake_id'] as String,
      vaultId: json['vault_id'] as String,
      assetSymbol: json['asset_symbol'] as String,
      principalAmount: SubPaiseAmount.fromJson(json['principal_amount'] as Map<String, dynamic>),
      tenorType: TenorType.values.byName(json['tenor_type'] as String),
      agreedApy: (json['agreed_apy'] as num).toDouble(),
      compoundingFrequency: CompoundingFrequency.values.byName(json['compounding_frequency'] as String),
      status: StakeStatus.values.byName(json['status'] as String),
      stakedAt: DateTime.parse(json['staked_at'] as String),
      lockedUntil: json['locked_until'] != null ? DateTime.parse(json['locked_until'] as String) : null,
      totalAccruedYield: SubPaiseAmount.fromJson(json['total_accrued_yield'] as Map<String, dynamic>),
      unallocatedYield: SubPaiseAmount.fromJson(json['unallocated_yield'] as Map<String, dynamic>),
      alreadyClaimedYield: SubPaiseAmount.fromJson(json['already_claimed_yield'] as Map<String, dynamic>),
      lastYieldCheckpoint: DateTime.parse(json['last_yield_checkpoint'] as String),
      onChainTxHash: json['on_chain_tx_hash'] as String,
    );
  }
}
