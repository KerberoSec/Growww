// Fixed-Income & Staking Vault Enums (Prompt 539)

enum YieldAssetType {
  gGsec,
  gTbill,
  gSdl,
  gCorp,
}

extension YieldAssetTypeExtension on YieldAssetType {
  String get displayName {
    switch (this) {
      case YieldAssetType.gGsec:
        return 'Sovereign G-Secs';
      case YieldAssetType.gTbill:
        return 'Treasury Bills';
      case YieldAssetType.gSdl:
        return 'State Development Loans';
      case YieldAssetType.gCorp:
        return 'AAA Corporate Bonds';
    }
  }

  String get badgeText {
    switch (this) {
      case YieldAssetType.gGsec:
        return 'GOI SOVEREIGN';
      case YieldAssetType.gTbill:
        return 'RBI T-BILL';
      case YieldAssetType.gSdl:
        return 'STATE SOVEREIGN';
      case YieldAssetType.gCorp:
        return 'CRISIL AAA';
    }
  }
}

enum TenorType {
  flexible,
  locked30Days,
  locked90Days,
  locked180Days,
  locked365Days,
}

extension TenorTypeExtension on TenorType {
  String get displayName {
    switch (this) {
      case TenorType.flexible:
        return 'Flexible (T+0)';
      case TenorType.locked30Days:
        return '30 Days Lock';
      case TenorType.locked90Days:
        return '90 Days Lock';
      case TenorType.locked180Days:
        return '180 Days Lock';
      case TenorType.locked365Days:
        return '365 Days Lock';
    }
  }

  int get durationDays {
    switch (this) {
      case TenorType.flexible:
        return 0;
      case TenorType.locked30Days:
        return 30;
      case TenorType.locked90Days:
        return 90;
      case TenorType.locked180Days:
        return 180;
      case TenorType.locked365Days:
        return 365;
    }
  }

  double get defaultBonusApy {
    switch (this) {
      case TenorType.flexible:
        return 0.0;
      case TenorType.locked30Days:
        return 0.25;
      case TenorType.locked90Days:
        return 0.50;
      case TenorType.locked180Days:
        return 0.85;
      case TenorType.locked365Days:
        return 1.25;
    }
  }
}

enum StakeStatus {
  active,
  pendingClaim,
  unbonding,
  matured,
  completed,
  cancelled,
}

enum RiskTier {
  sovereignRiskFree,
  institutionalLowRisk,
  corporateModerateRisk,
}

extension RiskTierExtension on RiskTier {
  String get label {
    switch (this) {
      case RiskTier.sovereignRiskFree:
        return 'Sovereign (Risk-Free)';
      case RiskTier.institutionalLowRisk:
        return 'Institutional Low Risk';
      case RiskTier.corporateModerateRisk:
        return 'Corporate Moderate Risk';
    }
  }
}

enum CompoundingFrequency {
  daily,
  atMaturity,
}

enum ClaimDestination {
  dematCashLedger,
  autoCompoundReinvest,
}
