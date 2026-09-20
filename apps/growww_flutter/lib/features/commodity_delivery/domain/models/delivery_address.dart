class DeliveryAddress {
  final String addressId;
  final String recipientName;
  final String recipientPhone;
  final String addressLine1;
  final String addressLine2;
  final String landmark;
  final String city;
  final String state;
  final String pincode;
  final bool isKycMatched;
  final String kycSource; // "DigiLocker / Aadhaar"

  const DeliveryAddress({
    required this.addressId,
    required this.recipientName,
    required this.recipientPhone,
    required this.addressLine1,
    required this.addressLine2,
    required this.landmark,
    required this.city,
    required this.state,
    required this.pincode,
    required this.isKycMatched,
    required this.kycSource,
  });

  Map<String, dynamic> toJson() => {
        'address_id': addressId,
        'recipient_name': recipientName,
        'recipient_phone': recipientPhone,
        'address_line_1': addressLine1,
        'address_line_2': addressLine2,
        'landmark': landmark,
        'city': city,
        'state': state,
        'pincode': pincode,
        'is_kyc_matched': isKycMatched,
        'kyc_source': kycSource,
      };

  factory DeliveryAddress.fromJson(Map<String, dynamic> json) {
    return DeliveryAddress(
      addressId: json['address_id'] as String,
      recipientName: json['recipient_name'] as String,
      recipientPhone: json['recipient_phone'] as String,
      addressLine1: json['address_line_1'] as String,
      addressLine2: json['address_line_2'] as String,
      landmark: json['landmark'] as String,
      city: json['city'] as String,
      state: json['state'] as String,
      pincode: json['pincode'] as String,
      isKycMatched: json['is_kyc_matched'] as bool,
      kycSource: json['kyc_source'] as String,
    );
  }
}
