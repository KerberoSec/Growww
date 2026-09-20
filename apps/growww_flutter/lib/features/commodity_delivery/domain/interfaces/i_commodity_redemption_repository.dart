import '../models/commodity_denomination.dart';
import '../models/commodity_enums.dart';
import '../models/delivery_receipt.dart';
import '../models/redemption_order.dart';
import '../models/sub_paise_fee_breakdown.dart';
import '../models/wdra_vault_location.dart';

abstract class ICommodityRedemptionRepository {
  Future<List<CommodityDenomination>> fetchAvailableDenominations({
    required CommodityType commodityType,
  });

  Future<List<WdraVaultLocation>> fetchWdraVaultLocations({
    CommodityType? commodityType,
    String? city,
  });

  Future<SubPaiseFeeBreakdown> calculateDeliveryFee({
    required CommodityType commodityType,
    required List<DenominationOrderItem> items,
    required DeliveryMode deliveryMode,
    String? destinationPincode,
    String? vaultId,
  });

  Future<bool> checkPincodeServiceability({
    required String pincode,
  });

  Future<RedemptionOrder> initiateRedemptionIntent({
    required CommodityType commodityType,
    required List<DenominationOrderItem> items,
    required DeliveryMode deliveryMode,
    String? vaultId,
    String? addressId,
    required String idempotencyKey,
  });

  Future<RedemptionOrder> authorizeRedemption({
    required String orderId,
    required String biometricSignature,
    required String totpCode,
  });

  Future<DeliveryReceipt> fetchDeliveryReceipt({
    required String orderId,
  });
}
