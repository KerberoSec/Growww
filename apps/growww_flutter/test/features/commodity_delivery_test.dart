import 'package:test/test.dart';
import 'package:growww_flutter/features/commodity_delivery/data/repositories/mock_commodity_redemption_repository.dart';
import 'package:growww_flutter/features/commodity_delivery/domain/models/commodity_denomination.dart';
import 'package:growww_flutter/features/commodity_delivery/domain/models/commodity_enums.dart';
import 'package:growww_flutter/features/commodity_delivery/domain/models/delivery_receipt.dart';
import 'package:growww_flutter/features/commodity_delivery/domain/models/sub_paise_fee_breakdown.dart';
import 'package:growww_flutter/features/commodity_delivery/presentation/controllers/commodity_redemption_controller.dart';

void main() {
  group('Prompt 530 - Commodity Physical Delivery and Vault Redemption Flow', () {
    late MockCommodityRedemptionRepository repository;
    late CommodityRedemptionController controller;

    setUp(() {
      repository = MockCommodityRedemptionRepository();
      controller = CommodityRedemptionController(
        redemptionRepo: repository,
        trackingRepo: repository,
      );
    });

    tearDown(() {
      controller.dispose();
    });

    test('CommodityType extensions return valid display names and minimum grams', () {
      expect(CommodityType.gold999.displayName, contains('999 Fine Gold'));
      expect(CommodityType.gold9999.displayName, contains('999.9 Minted Gold'));
      expect(CommodityType.silver999.displayName, contains('999 Fine Silver'));

      expect(CommodityType.gold999.minimumRedemptionGrams, equals(1.0));
      expect(CommodityType.silver999.minimumRedemptionGrams, equals(10.0));
    });

    test('SubPaiseFeeBreakdown calculation enforces 0.00% Zero Fee and exact GST splits', () {
      final fee = SubPaiseFeeBreakdown.calculate(
        totalWeightGrams: 10.0,
        indicativeMetalValueInr: 72000.0,
        isDoorstep: true,
        itemsCount: 1,
      );

      // Vault fee: ₹100 base + ₹50 (for 10g) = ₹150.00 -> 1,500,000 sub-paise
      expect(fee.vaultRematerializationFeeSubPaise, equals(1500000));
      expect(fee.vaultFeeInr, equals(150.0));

      // 0.00% Platform settlement fee invariant
      expect(fee.settlementFeeSubPaise, equals(0));
      expect(fee.settlementFeeInr, equals(0.0));

      // Packaging fee: 1 * 75 = ₹75 -> 750,000 sub-paise
      expect(fee.assayPackagingFeeSubPaise, equals(750000));
      expect(fee.packagingFeeInr, equals(75.0));

      // Armored logistics: ₹450 -> 4,500,000 sub-paise
      expect(fee.armoredLogisticsFeeSubPaise, equals(4500000));
      expect(fee.logisticsFeeInr, equals(450.0));

      // Transit insurance: 72,000 * 0.15% = ₹108.00 -> 1,080,000 sub-paise
      expect(fee.transitInsuranceFeeSubPaise, equals(1080000));
      expect(fee.insuranceFeeInr, equals(108.0));

      // 3% GST on vault: 150 * 0.03 = ₹4.50 -> 45,000 sub-paise
      expect(fee.preciousMetalGstSubPaise, equals(45000));
      expect(fee.preciousMetalGstInr, equals(4.50));

      // 18% GST on services (75 + 450 + 108 = 633): 633 * 0.18 = ₹113.94 -> 1,139,400 sub-paise
      expect(fee.handlingServicesGstSubPaise, equals(1139400));
      expect(fee.handlingServicesGstInr, equals(113.94));

      // Total Gross: 150 + 75 + 450 + 108 + 4.5 + 113.94 = ₹901.44
      expect(fee.totalInr, closeTo(901.44, 0.01));
    });

    test('WDRA Vault Locator filters vaults by city and supported commodity', () async {
      final mumbaiVaults = await repository.fetchWdraVaultLocations(
        city: 'Mumbai',
        commodityType: CommodityType.gold999,
      );

      expect(mumbaiVaults.isNotEmpty, isTrue);
      expect(mumbaiVaults.first.operatorName, contains('Sequel'));
      expect(mumbaiVaults.first.repositoryRegistrationNo, contains('WDRA'));
      expect(mumbaiVaults.first.pickupRequirements, contains('Aadhaar Biometric Check'));

      final nonExistent = await repository.fetchWdraVaultLocations(city: 'Kochi');
      expect(nonExistent.isEmpty, isTrue);
    });

    test('Pincode serviceability correctly validates delivery coverage', () async {
      expect(await repository.checkPincodeServiceability(pincode: '400051'), isTrue);
      expect(await repository.checkPincodeServiceability(pincode: '560001'), isTrue);
      expect(await repository.checkPincodeServiceability(pincode: '999999'), isFalse);
    });

    test('DeliveryReceipt SHA-256 cryptographic checksum and integrity verification', () async {
      final receipt = await repository.fetchDeliveryReceipt(orderId: 'RDM-TEST-001');

      expect(receipt.receiptId, contains('RDM-TEST-001'));
      expect(receipt.onChainBurnTxHash, startsWith('0x'));
      expect(receipt.deliveredBars.isNotEmpty, isTrue);
      expect(receipt.deliveredBars.first.bisHallmarkNumber, contains('BIS-HM'));

      final checksum = receipt.computeChecksum();
      expect(checksum.length, equals(64)); // SHA-256 hex string
      expect(receipt.verifyIntegrity(), isTrue);
    });

    test('CommodityRedemptionController manages full redemption wizard lifecycle', () async {
      await controller.initialize();
      expect(controller.state.availableDenominations.isNotEmpty, isTrue);

      // Select gold 10g bar
      final gold10g = controller.state.availableDenominations
          .firstWhere((d) => d.denominationId == 'GOLD-10G-MINT');
      await controller.updateQuantity(gold10g.denominationId, 1);

      expect(controller.state.totalSelectedGrams, equals(10.0));
      expect(controller.state.isValidWeight, isTrue);
      expect(controller.state.feeBreakdown, isNotNull);

      // Initiate order
      final order = await controller.initiateRedemption();
      expect(order, isNotNull);
      expect(order!.totalFineGrams, equals(10.0));
      expect(controller.state.activeOrder, isNotNull);

      // Authorize with biometric signature
      final authorized = await controller.authorizeRedemption(
        biometricSignature: 'BIO-PASSKEY-AUTH-TEST',
        totpCode: '123456',
      );
      expect(authorized, isTrue);
      expect(controller.state.activeOrder?.status, equals(RedemptionStatus.authorized));
      expect(controller.state.activeOrder?.enwrNumber, contains('ENWR-NERL'));
      expect(controller.state.activeOrder?.onChainBurnTxHash, startsWith('0x'));

      // Load tracking and receipt
      await controller.loadTrackingInfo(order.orderId);
      expect(controller.state.trackingEvents.isNotEmpty, isTrue);
      expect(controller.state.currentOtp, isNotNull);
      expect(controller.state.currentOtp!.length, equals(6));

      await controller.fetchReceipt(order.orderId);
      expect(controller.state.deliveryReceipt, isNotNull);
      expect(controller.state.deliveryReceipt!.verifyIntegrity(), isTrue);
    });
  });
}
