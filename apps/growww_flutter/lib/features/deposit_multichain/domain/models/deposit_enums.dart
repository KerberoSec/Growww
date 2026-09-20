// Multichain USDT Deposit Enums (Prompt 565)

enum DepositNetworkType {
  ethereumErc20,
  tronTrc20,
  polygon,
  solana,
  arbitrumOne,
  bscBep20,
}

extension DepositNetworkTypeExtension on DepositNetworkType {
  String get displayName {
    switch (this) {
      case DepositNetworkType.ethereumErc20:
        return 'Ethereum (ERC-20)';
      case DepositNetworkType.tronTrc20:
        return 'Tron (TRC-20)';
      case DepositNetworkType.polygon:
        return 'Polygon PoS (POL)';
      case DepositNetworkType.solana:
        return 'Solana (SPL)';
      case DepositNetworkType.arbitrumOne:
        return 'Arbitrum One (Rollup)';
      case DepositNetworkType.bscBep20:
        return 'BNB Smart Chain (BEP-20)';
    }
  }

  String get networkTag {
    switch (this) {
      case DepositNetworkType.ethereumErc20:
        return 'ERC20';
      case DepositNetworkType.tronTrc20:
        return 'TRC20';
      case DepositNetworkType.polygon:
        return 'POLYGON';
      case DepositNetworkType.solana:
        return 'SOLANA';
      case DepositNetworkType.arbitrumOne:
        return 'ARB';
      case DepositNetworkType.bscBep20:
        return 'BEP20';
    }
  }
}

enum NetworkCongestionLevel {
  low,
  medium,
  high,
}

extension NetworkCongestionLevelExtension on NetworkCongestionLevel {
  String get label {
    switch (this) {
      case NetworkCongestionLevel.low:
        return 'Low Congestion (Fast)';
      case NetworkCongestionLevel.medium:
        return 'Moderate Congestion';
      case NetworkCongestionLevel.high:
        return 'High Gas Congestion';
    }
  }
}

enum DepositStatus {
  detecting,
  confirming,
  credited,
  failed,
}
