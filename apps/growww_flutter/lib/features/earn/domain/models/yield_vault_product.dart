import 'earn_enums.dart';
import 'sub_paise_amount.dart';

class YieldVaultProduct {
  final String vaultId;
  final String vaultAddress; // Smart contract address on Besu
  final String assetSymbol; // "gGSEC-718GS2033"
  final String displayName; // "7.18% GS 2033 Sovereign Vault"
  final YieldAssetType assetType;
  final String isin;
  final String issuerName; // "Government of India / RBI"
  final String creditRating; // "SOV", "CRISIL AAA"
  final double baseApy; // 7.18%
  final Map<TenorType, double> tenorBonusApy;
  final SubPaiseAmount minStakeAmount;
  final SubPaiseAmount totalValueLocked;
  final SubPaiseAmount maxPoolCapacity;
  final RiskTier riskTier;
  final DateTime bondMaturityDate;
  final double modifiedDurationYears;
  final bool isAcceptingStakes;
  final String proofOfReserveRootHash;

  const YieldVaultProduct({
    required this.vaultId,
    required this.vaultAddress,
    required this.assetSymbol,
    required this.displayName,
    required this.assetType,
    required this.isin,
    required this.issuerName,
    required this.creditRating,
    required this.baseApy,
    required this.tenorBonusApy,
    required this.minStakeAmount,
    required this.totalValueLocked,
    required this.maxPoolCapacity,
    required this.riskTier,
    required this.bondMaturityDate,
    required this.modifiedDurationYears,
    required this.isAcceptingStakes,
    required this.proofOfReserveRootHash,
  });

  double getEffectiveApy(TenorType tenor) {
    final bonus = tenorBonusApy[tenor] ?? tenor.defaultBonusApy;
    return baseApy + bonus;
  }

  double get tvlFillRatio {
    if (maxPoolCapacity.subPaiseValue <= 0) return 0.0;
    return (totalValueLocked.subPaiseValue / maxPoolCapacity.subPaiseValue).clamp(0.0, 1.0);
  }

  Map<String, dynamic> toJson() => {
        'vault_id': vaultId,
        'vault_address': vaultAddress,
        'asset_symbol': assetSymbol,
        'display_name': displayName,
        'asset_type': assetType.name,
        'isin': isin,
        'issuer_name': issuerName,
        'credit_rating': creditRating,
        'base_apy': baseApy,
        'tenor_bonus_apy': tenorBonusApy.map((k, v) => MapEntry(k.name, v)),
        'min_stake_amount': minStakeAmount.toJson(),
        'total_value_locked': totalValueLocked.toJson(),
        'max_pool_capacity': maxPoolCapacity.toJson(),
        'risk_tier': riskTier.name,
        'bond_maturity_date': bondMaturityDate.toIso8601String(),
        'modified_duration_years': modifiedDurationYears,
        'is_accepting_stakes': isAcceptingStakes,
        'proof_of_reserve_root_hash': proofOfReserveRootHash,
      };

  factory YieldVaultProduct.fromJson(Map<String, dynamic> json) {
    final rawBonus = json['tenor_bonus_apy'] as Map<String, dynamic>;
    final bonusMap = rawBonus.map(
      (k, v) => MapEntry(TenorType.values.byName(k), (v as num).toDouble()),
    );

    return YieldVaultProduct(
      vaultId: json['vault_id'] as String,
      vaultAddress: json['vault_address'] as String,
      assetSymbol: json['asset_symbol'] as String,
      displayName: json['display_name'] as String,
      assetType: YieldAssetType.values.byName(json['asset_type'] as String),
      isin: json['isin'] as String,
      issuerName: json['issuer_name'] as String,
      creditRating: json['credit_rating'] as String,
      baseApy: (json['base_apy'] as num).toDouble(),
      tenorBonusApy: bonusMap,
      minStakeAmount: SubPaiseAmount.fromJson(json['min_stake_amount'] as Map<String, dynamic>),
      totalValueLocked: SubPaiseAmount.fromJson(json['total_value_locked'] as Map<String, dynamic>),
      maxPoolCapacity: SubPaiseAmount.fromJson(json['max_pool_capacity'] as Map<String, dynamic>),
      riskTier: RiskTier.values.byName(json['risk_tier'] as String),
      bondMaturityDate: DateTime.parse(json['bond_maturity_date'] as String),
      modifiedDurationYears: (json['modified_duration_years'] as num).toDouble(),
      isAcceptingStakes: json['is_accepting_stakes'] as bool,
      proofOfReserveRootHash: json['proof_of_reserve_root_hash'] as String,
    );
  }
}
