import 'deposit_enums.dart';

class DepositAddressInfo {
  final DepositNetworkType networkType;
  final String address;
  final String? memoOrTag;
  final String qrPayload;
  final bool isSegregatedVault;
  final DateTime generatedAt;

  const DepositAddressInfo({
    required this.networkType,
    required this.address,
    this.memoOrTag,
    required this.qrPayload,
    required this.isSegregatedVault,
    required this.generatedAt,
  });

  Map<String, dynamic> toJson() => {
        'network_type': networkType.name,
        'address': address,
        'memo_or_tag': memoOrTag,
        'qr_payload': qrPayload,
        'is_segregated_vault': isSegregatedVault,
        'generated_at': generatedAt.toIso8601String(),
      };

  factory DepositAddressInfo.fromJson(Map<String, dynamic> json) {
    return DepositAddressInfo(
      networkType: DepositNetworkType.values.byName(json['network_type'] as String),
      address: json['address'] as String,
      memoOrTag: json['memo_or_tag'] as String?,
      qrPayload: json['qr_payload'] as String,
      isSegregatedVault: json['is_segregated_vault'] as bool,
      generatedAt: DateTime.parse(json['generated_at'] as String),
    );
  }
}
