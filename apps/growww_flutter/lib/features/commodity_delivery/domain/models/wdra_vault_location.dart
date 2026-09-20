import 'commodity_enums.dart';

class WdraVaultLocation {
  final String vaultId;
  final String repositoryName; // "National E-Repository Limited (NERL)"
  final String repositoryRegistrationNo;
  final String operatorName; // "MMTC-PAMP India Pvt Ltd", "Sequel Vaulting"
  final String facilityName;
  final String addressLine;
  final String city;
  final String state;
  final String pincode;
  final double latitude;
  final double longitude;
  final String operatingHours;
  final List<CommodityType> supportedCommodities;
  final List<String> pickupRequirements;
  final bool isOperational;

  const WdraVaultLocation({
    required this.vaultId,
    required this.repositoryName,
    required this.repositoryRegistrationNo,
    required this.operatorName,
    required this.facilityName,
    required this.addressLine,
    required this.city,
    required this.state,
    required this.pincode,
    required this.latitude,
    required this.longitude,
    required this.operatingHours,
    required this.supportedCommodities,
    required this.pickupRequirements,
    required this.isOperational,
  });

  Map<String, dynamic> toJson() => {
        'vault_id': vaultId,
        'repository_name': repositoryName,
        'repository_registration_no': repositoryRegistrationNo,
        'operator_name': operatorName,
        'facility_name': facilityName,
        'address_line': addressLine,
        'city': city,
        'state': state,
        'pincode': pincode,
        'latitude': latitude,
        'longitude': longitude,
        'operating_hours': operatingHours,
        'supported_commodities': supportedCommodities.map((e) => e.name).toList(),
        'pickup_requirements': pickupRequirements,
        'is_operational': isOperational,
      };

  factory WdraVaultLocation.fromJson(Map<String, dynamic> json) {
    return WdraVaultLocation(
      vaultId: json['vault_id'] as String,
      repositoryName: json['repository_name'] as String,
      repositoryRegistrationNo: json['repository_registration_no'] as String,
      operatorName: json['operator_name'] as String,
      facilityName: json['facility_name'] as String,
      addressLine: json['address_line'] as String,
      city: json['city'] as String,
      state: json['state'] as String,
      pincode: json['pincode'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      operatingHours: json['operating_hours'] as String,
      supportedCommodities: (json['supported_commodities'] as List<dynamic>)
          .map((e) => CommodityType.values.byName(e as String))
          .toList(),
      pickupRequirements: List<String>.from(json['pickup_requirements'] as List),
      isOperational: json['is_operational'] as bool,
    );
  }
}
