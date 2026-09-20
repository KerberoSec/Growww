// Commodity Enums for Regulated Bullion Physical Delivery (Prompt 530)

enum CommodityType {
  gold999,
  gold9999,
  silver999,
}

extension CommodityTypeExtension on CommodityType {
  String get displayName {
    switch (this) {
      case CommodityType.gold999:
        return 'gGOLD (999 Fine Gold)';
      case CommodityType.gold9999:
        return 'gGOLD (999.9 Minted Gold)';
      case CommodityType.silver999:
        return 'gSILVER (999 Fine Silver)';
    }
  }

  String get code {
    switch (this) {
      case CommodityType.gold999:
        return 'gGOLD-999';
      case CommodityType.gold9999:
        return 'gGOLD-999.9';
      case CommodityType.silver999:
        return 'gSILVER-999';
    }
  }

  double get minimumRedemptionGrams {
    switch (this) {
      case CommodityType.gold999:
      case CommodityType.gold9999:
        return 1.0;
      case CommodityType.silver999:
        return 10.0;
    }
  }
}

enum DeliveryMode {
  vaultPickup,
  armoredDoorstep,
}

extension DeliveryModeExtension on DeliveryMode {
  String get displayName {
    switch (this) {
      case DeliveryMode.vaultPickup:
        return 'WDRA Vault Self-Pickup';
      case DeliveryMode.armoredDoorstep:
        return 'Insured Armored Doorstep Courier';
    }
  }

  String get description {
    switch (this) {
      case DeliveryMode.vaultPickup:
        return 'Collect physical bullion directly at WDRA-accredited vault with KYC verification.';
      case DeliveryMode.armoredDoorstep:
        return '100% insured armored transit (Brink\'s/Sequel/BVC) delivered to your verified address.';
    }
  }
}

enum RedemptionStatus {
  draft,
  feePending,
  authorized,
  enwrRematerializing,
  vaultAllocated,
  inTransit,
  outForDelivery,
  delivered,
  completed,
  cancelled,
  failed,
}

extension RedemptionStatusExtension on RedemptionStatus {
  String get displayName {
    switch (this) {
      case RedemptionStatus.draft:
        return 'Draft Order';
      case RedemptionStatus.feePending:
        return 'Awaiting Fee Authorization';
      case RedemptionStatus.authorized:
        return 'Authorized & Locked';
      case RedemptionStatus.enwrRematerializing:
        return 'eNWR Rematerialization';
      case RedemptionStatus.vaultAllocated:
        return 'Vault Custody Allocated';
      case RedemptionStatus.inTransit:
        return 'In Armored Transit';
      case RedemptionStatus.outForDelivery:
        return 'Out for Handover';
      case RedemptionStatus.delivered:
        return 'Delivered & Verified';
      case RedemptionStatus.completed:
        return 'Settled & Burned';
      case RedemptionStatus.cancelled:
        return 'Cancelled';
      case RedemptionStatus.failed:
        return 'Failed';
    }
  }

  bool get isFinal =>
      this == RedemptionStatus.completed ||
      this == RedemptionStatus.cancelled ||
      this == RedemptionStatus.failed;
}

enum LogisticsCarrier {
  sequelLogistics,
  brinksSecure,
  bvcLogistics,
  malcaAmit,
}

extension LogisticsCarrierExtension on LogisticsCarrier {
  String get carrierName {
    switch (this) {
      case LogisticsCarrier.sequelLogistics:
        return 'Sequel Secure Logistics';
      case LogisticsCarrier.brinksSecure:
        return 'Brink\'s India Armored Logistics';
      case LogisticsCarrier.bvcLogistics:
        return 'BVC Secure Logistics';
      case LogisticsCarrier.malcaAmit:
        return 'Malca-Amit Vaults & Logistics';
    }
  }
}
