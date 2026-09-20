import 'deposit_enums.dart';

class CryptoNetworkInfo {
  final String networkId;
  final DepositNetworkType networkType;
  final String displayName;
  final String chainName;
  final int requiredConfirmations;
  final int estimatedArrivalMinutes;
  final double depositFeeUsdt;
  final double minDepositUsdt;
  final String contractAddress;
  final String explorerUrl;
  final NetworkCongestionLevel congestionLevel;
  final bool isEnabled;

  const CryptoNetworkInfo({
    required this.networkId,
    required this.networkType,
    required this.displayName,
    required this.chainName,
    required this.requiredConfirmations,
    required this.estimatedArrivalMinutes,
    required this.depositFeeUsdt,
    required this.minDepositUsdt,
    required this.contractAddress,
    required this.explorerUrl,
    required this.congestionLevel,
    required this.isEnabled,
  });

  Map<String, dynamic> toJson() => {
        'network_id': networkId,
        'network_type': networkType.name,
        'display_name': displayName,
        'chain_name': chainName,
        'required_confirmations': requiredConfirmations,
        'estimated_arrival_minutes': estimatedArrivalMinutes,
        'deposit_fee_usdt': depositFeeUsdt,
        'min_deposit_usdt': minDepositUsdt,
        'contract_address': contractAddress,
        'explorer_url': explorerUrl,
        'congestion_level': congestionLevel.name,
        'is_enabled': isEnabled,
      };

  factory CryptoNetworkInfo.fromJson(Map<String, dynamic> json) {
    return CryptoNetworkInfo(
      networkId: json['network_id'] as String,
      networkType: DepositNetworkType.values.byName(json['network_type'] as String),
      displayName: json['display_name'] as String,
      chainName: json['chain_name'] as String,
      requiredConfirmations: json['required_confirmations'] as int,
      estimatedArrivalMinutes: json['estimated_arrival_minutes'] as int,
      depositFeeUsdt: (json['deposit_fee_usdt'] as num).toDouble(),
      minDepositUsdt: (json['min_deposit_usdt'] as num).toDouble(),
      contractAddress: json['contract_address'] as String,
      explorerUrl: json['explorer_url'] as String,
      congestionLevel: NetworkCongestionLevel.values.byName(json['congestion_level'] as String),
      isEnabled: json['is_enabled'] as bool,
    );
  }
}
