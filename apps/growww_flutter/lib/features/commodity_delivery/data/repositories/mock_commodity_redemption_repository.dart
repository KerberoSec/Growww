import 'dart:async';
import 'dart:convert';
import 'package:crypto/crypto.dart';

import '../../domain/interfaces/i_commodity_redemption_repository.dart';
import '../../domain/interfaces/i_logistics_tracking_repository.dart';
import '../../domain/models/commodity_denomination.dart';
import '../../domain/models/commodity_enums.dart';
import '../../domain/models/delivery_address.dart';
import '../../domain/models/delivery_receipt.dart';
import '../../domain/models/logistics_tracking_event.dart';
import '../../domain/models/redemption_order.dart';
import '../../domain/models/sub_paise_fee_breakdown.dart';
import '../../domain/models/wdra_vault_location.dart';

class MockCommodityRedemptionRepository
    implements ICommodityRedemptionRepository, ILogisticsTrackingRepository {
  final Map<String, RedemptionOrder> _orders = {};

  static final List<CommodityDenomination> _denominations = [
    // gGOLD Denominations
    const CommodityDenomination(
      denominationId: 'GOLD-1G-MINT',
      commodityType: CommodityType.gold999,
      weightGrams: 1.0,
      fineness: 999.0,
      displayName: '1g Gold Bar 999',
      barFormFactor: 'Minted Bar',
      availableStockUnits: 450,
      assayStandard: 'BIS 1417:2016',
      packagingType: 'CertiCard Blister',
    ),
    const CommodityDenomination(
      denominationId: 'GOLD-8G-COIN',
      commodityType: CommodityType.gold999,
      weightGrams: 8.0,
      fineness: 999.0,
      displayName: '8g Guinea Gold Coin 999',
      barFormFactor: 'Coin',
      availableStockUnits: 120,
      assayStandard: 'BIS 1417:2016',
      packagingType: 'Tamper-Evident Capsule',
    ),
    const CommodityDenomination(
      denominationId: 'GOLD-10G-MINT',
      commodityType: CommodityType.gold999,
      weightGrams: 10.0,
      fineness: 999.0,
      displayName: '10g Gold Bar 999',
      barFormFactor: 'Minted Bar',
      availableStockUnits: 800,
      assayStandard: 'BIS 1417:2016',
      packagingType: 'CertiCard Blister',
    ),
    const CommodityDenomination(
      denominationId: 'GOLD-50G-CAST',
      commodityType: CommodityType.gold999,
      weightGrams: 50.0,
      fineness: 999.0,
      displayName: '50g Gold Bar 999',
      barFormFactor: 'Cast Bar',
      availableStockUnits: 65,
      assayStandard: 'LBMA Good Delivery',
      packagingType: 'Tamper-Evident Pouch',
    ),
    const CommodityDenomination(
      denominationId: 'GOLD-100G-CAST',
      commodityType: CommodityType.gold999,
      weightGrams: 100.0,
      fineness: 999.0,
      displayName: '100g Gold Bar 999',
      barFormFactor: 'Cast Bar',
      availableStockUnits: 40,
      assayStandard: 'LBMA Good Delivery',
      packagingType: 'Tamper-Evident Pouch',
    ),
    // gSILVER Denominations
    const CommodityDenomination(
      denominationId: 'SILVER-10G-COIN',
      commodityType: CommodityType.silver999,
      weightGrams: 10.0,
      fineness: 999.0,
      displayName: '10g Silver Coin 999',
      barFormFactor: 'Coin',
      availableStockUnits: 1500,
      assayStandard: 'BIS 2112:2014',
      packagingType: 'Tamper-Evident Capsule',
    ),
    const CommodityDenomination(
      denominationId: 'SILVER-100G-BAR',
      commodityType: CommodityType.silver999,
      weightGrams: 100.0,
      fineness: 999.0,
      displayName: '100g Silver Bar 999',
      barFormFactor: 'Minted Bar',
      availableStockUnits: 600,
      assayStandard: 'BIS 2112:2014',
      packagingType: 'CertiCard Blister',
    ),
    const CommodityDenomination(
      denominationId: 'SILVER-500G-BAR',
      commodityType: CommodityType.silver999,
      weightGrams: 500.0,
      fineness: 999.0,
      displayName: '500g Silver Bar 999',
      barFormFactor: 'Cast Bar',
      availableStockUnits: 220,
      assayStandard: 'BIS 2112:2014',
      packagingType: 'Tamper-Evident Pouch',
    ),
    const CommodityDenomination(
      denominationId: 'SILVER-1KG-BAR',
      commodityType: CommodityType.silver999,
      weightGrams: 1000.0,
      fineness: 999.0,
      displayName: '1kg Silver Bar 999',
      barFormFactor: 'Cast Bar',
      availableStockUnits: 95,
      assayStandard: 'LBMA Good Delivery',
      packagingType: 'Tamper-Evident Pouch',
    ),
  ];

  static final List<WdraVaultLocation> _vaults = [
    const WdraVaultLocation(
      vaultId: 'VAULT-MUM-001',
      repositoryName: 'National E-Repository Limited (NERL)',
      repositoryRegistrationNo: 'WDRA/REG/2021/MUM/048',
      operatorName: 'Sequel Vaulting Logistics Pvt Ltd',
      facilityName: 'Bandra-Kurla Complex Bullion Hub',
      addressLine: 'Plot C-59, G-Block, BKC, Bandra East',
      city: 'Mumbai',
      state: 'Maharashtra',
      pincode: '400051',
      latitude: 19.0657,
      longitude: 72.8687,
      operatingHours: '10:00 AM - 05:00 PM (Mon-Fri)',
      supportedCommodities: [CommodityType.gold999, CommodityType.gold9999, CommodityType.silver999],
      pickupRequirements: ['Original PAN Card', 'Aadhaar Biometric Check', 'DigiLocker Gate Pass'],
      isOperational: true,
    ),
    const WdraVaultLocation(
      vaultId: 'VAULT-DEL-002',
      repositoryName: 'CDSL Commodity Repository Limited (CCRL)',
      repositoryRegistrationNo: 'WDRA/REG/2020/DEL/012',
      operatorName: 'MMTC-PAMP India Pvt Ltd',
      facilityName: 'Connaught Place Central Depository',
      addressLine: 'Barakhamba Road, Statesman House, Ground Floor',
      city: 'New Delhi',
      state: 'Delhi',
      pincode: '110001',
      latitude: 28.6289,
      longitude: 77.2285,
      operatingHours: '09:30 AM - 04:30 PM (Mon-Fri)',
      supportedCommodities: [CommodityType.gold999, CommodityType.gold9999, CommodityType.silver999],
      pickupRequirements: ['Original PAN Card', 'Aadhaar Biometric Check', 'Gate Pass OTP'],
      isOperational: true,
    ),
    const WdraVaultLocation(
      vaultId: 'VAULT-AHM-003',
      repositoryName: 'National E-Repository Limited (NERL)',
      repositoryRegistrationNo: 'WDRA/REG/2022/GUJ/089',
      operatorName: 'Brink\'s India Secure Logistics',
      facilityName: 'GIFT City Bullion Vault',
      addressLine: 'Zone 1, GIFT SEZ, Gandhinagar',
      city: 'Ahmedabad',
      state: 'Gujarat',
      pincode: '382355',
      latitude: 23.1601,
      longitude: 72.6841,
      operatingHours: '10:00 AM - 06:00 PM (Mon-Fri)',
      supportedCommodities: [CommodityType.gold999, CommodityType.gold9999],
      pickupRequirements: ['Original PAN Card', 'SEZ Visitor Pass', 'DigiLocker QR'],
      isOperational: true,
    ),
    const WdraVaultLocation(
      vaultId: 'VAULT-BLR-004',
      repositoryName: 'National E-Repository Limited (NERL)',
      repositoryRegistrationNo: 'WDRA/REG/2023/KAR/104',
      operatorName: 'BVC Secure Logistics Pvt Ltd',
      facilityName: 'MG Road Depository Center',
      addressLine: '45 MG Road, Ashok Nagar',
      city: 'Bengaluru',
      state: 'Karnataka',
      pincode: '560001',
      latitude: 12.9756,
      longitude: 77.6066,
      operatingHours: '10:00 AM - 05:00 PM (Mon-Sat)',
      supportedCommodities: [CommodityType.gold999, CommodityType.silver999],
      pickupRequirements: ['Original PAN Card', 'Aadhaar Card', 'Delivery Token'],
      isOperational: true,
    ),
  ];

  static const List<String> _serviceablePincodes = [
    '400051', '400001', '400050', '110001', '110020',
    '382355', '380009', '560001', '560025', '600001',
    '700001', '500001', '302001', '395003'
  ];

  @override
  Future<List<CommodityDenomination>> fetchAvailableDenominations({
    required CommodityType commodityType,
  }) async {
    return _denominations.where((d) => d.commodityType == commodityType).toList();
  }

  @override
  Future<List<WdraVaultLocation>> fetchWdraVaultLocations({
    CommodityType? commodityType,
    String? city,
  }) async {
    return _vaults.where((v) {
      if (commodityType != null && !v.supportedCommodities.contains(commodityType)) {
        return false;
      }
      if (city != null && city.isNotEmpty && v.city.toLowerCase() != city.toLowerCase()) {
        return false;
      }
      return true;
    }).toList();
  }

  @override
  Future<SubPaiseFeeBreakdown> calculateDeliveryFee({
    required CommodityType commodityType,
    required List<DenominationOrderItem> items,
    required DeliveryMode deliveryMode,
    String? destinationPincode,
    String? vaultId,
  }) async {
    final double totalWeight = items.fold(0.0, (sum, i) => sum + i.totalWeightGrams);
    final int totalCount = items.fold(0, (sum, i) => sum + i.quantity);

    // Indicative metal spot price: ₹7,200/g for Gold, ₹85/g for Silver
    final double ratePerGram = commodityType == CommodityType.silver999 ? 85.0 : 7200.0;
    final double metalValue = totalWeight * ratePerGram;

    return SubPaiseFeeBreakdown.calculate(
      totalWeightGrams: totalWeight,
      indicativeMetalValueInr: metalValue,
      isDoorstep: deliveryMode == DeliveryMode.armoredDoorstep,
      itemsCount: totalCount,
    );
  }

  @override
  Future<bool> checkPincodeServiceability({
    required String pincode,
  }) async {
    return _serviceablePincodes.contains(pincode.trim());
  }

  @override
  Future<RedemptionOrder> initiateRedemptionIntent({
    required CommodityType commodityType,
    required List<DenominationOrderItem> items,
    required DeliveryMode deliveryMode,
    String? vaultId,
    String? addressId,
    required String idempotencyKey,
  }) async {
    final double totalGrams = items.fold(0.0, (sum, i) => sum + i.totalWeightGrams);
    final fee = await calculateDeliveryFee(
      commodityType: commodityType,
      items: items,
      deliveryMode: deliveryMode,
    );

    WdraVaultLocation? vault;
    if (vaultId != null) {
      vault = _vaults.firstWhere((v) => v.vaultId == vaultId, orElse: () => _vaults.first);
    }

    DeliveryAddress? address;
    if (addressId != null) {
      address = const DeliveryAddress(
        addressId: 'ADDR-KYC-001',
        recipientName: 'ARUN KUMAR SHARMA',
        recipientPhone: '+91 98765 43210',
        addressLine1: 'Flat 402, Lotus Towers, 14th Main',
        addressLine2: 'Indiranagar 2nd Stage',
        landmark: 'Near Indiranagar Metro Station',
        city: 'Bengaluru',
        state: 'Karnataka',
        pincode: '560001',
        isKycMatched: true,
        kycSource: 'DigiLocker / Aadhaar',
      );
    }

    final orderId = 'RDM-NERL-${DateTime.now().millisecondsSinceEpoch}';
    final order = RedemptionOrder(
      orderId: orderId,
      userId: 'USR-NBSE-7749',
      commodityType: commodityType,
      items: items,
      totalFineGrams: totalGrams,
      deliveryMode: deliveryMode,
      selectedVault: vault,
      deliveryAddress: address,
      feeBreakdown: fee,
      status: RedemptionStatus.draft,
      idempotencyKey: idempotencyKey,
      createdAt: DateTime.now(),
      estimatedDeliveryDate: DateTime.now().add(const Duration(days: 3)),
    );

    _orders[orderId] = order;
    return order;
  }

  @override
  Future<RedemptionOrder> authorizeRedemption({
    required String orderId,
    required String biometricSignature,
    required String totpCode,
  }) async {
    final existing = _orders[orderId] ??
        RedemptionOrder(
          orderId: orderId,
          userId: 'USR-NBSE-7749',
          commodityType: CommodityType.gold999,
          items: const [],
          totalFineGrams: 10.0,
          deliveryMode: DeliveryMode.armoredDoorstep,
          feeBreakdown: SubPaiseFeeBreakdown.calculate(
            totalWeightGrams: 10.0,
            indicativeMetalValueInr: 72000.0,
            isDoorstep: true,
          ),
          status: RedemptionStatus.draft,
          createdAt: DateTime.now(),
        );

    final eNwrNumber = 'ENWR-NERL-2026-${orderId.substring(orderId.length - 6)}';
    final burnTx = '0x${sha256.convert(utf8.encode(orderId + biometricSignature)).toString().substring(0, 40)}';

    final updated = existing.copyWith(
      status: RedemptionStatus.authorized,
      enwrNumber: eNwrNumber,
      onChainBurnTxHash: burnTx,
    );
    _orders[orderId] = updated;
    return updated;
  }

  @override
  Future<DeliveryReceipt> fetchDeliveryReceipt({
    required String orderId,
  }) async {
    final order = _orders[orderId];
    final fineGrams = order?.totalFineGrams ?? 10.0;
    final enwr = order?.enwrNumber ?? 'ENWR-NERL-2026-990142';
    final burnTx = order?.onChainBurnTxHash ?? '0x9a8f4c2e1b6d7a8e5f3c1b2a4d6e8f0a2c4e6b8d';
    final now = DateTime.now();

    final bars = [
      DeliveredBarDetail(
        barSerialNumber: 'MMTC-GOLD-999-2026-004812',
        grossWeightGrams: fineGrams,
        fineness: 999.0,
        refineryName: 'MMTC-PAMP India Pvt Ltd (NABL Accredited)',
        bisHallmarkNumber: 'BIS-HM-999-748920',
        assayCertificateUrl: 'https://cert.growww.in/assay/MMTC-GOLD-999-2026-004812.pdf',
      ),
    ];

    final receiptId = 'RCP-$orderId';
    final payload = '$receiptId:$orderId:$enwr:$burnTx:$fineGrams:${now.toIso8601String()}';
    final sig = sha256.convert(utf8.encode(payload)).toString();

    return DeliveryReceipt(
      receiptId: receiptId,
      orderId: orderId,
      enwrNumber: enwr,
      repositoryName: 'National E-Repository Limited (NERL)',
      onChainBurnTxHash: burnTx,
      deliveredBars: bars,
      totalFineGramsDelivered: fineGrams,
      recipientName: 'ARUN KUMAR SHARMA',
      verificationSignature: sig,
      deliveryQrPayload: 'growww://verify/receipt?id=$receiptId&sig=${sig.substring(0, 16)}',
      completedAt: now,
    );
  }

  @override
  Future<List<LogisticsTrackingEvent>> fetchTrackingHistory({
    required String orderId,
  }) async {
    final now = DateTime.now();
    return [
      LogisticsTrackingEvent(
        eventId: 'EVT-01',
        orderId: orderId,
        carrier: LogisticsCarrier.sequelLogistics,
        trackingNumber: 'SEQL-ARM-8840192',
        status: RedemptionStatus.vaultAllocated,
        statusDescription: 'Physical fine bullion allocated and packaged in tamper-evident seal.',
        locationCity: 'Mumbai BKC Vault',
        securityBagSealNumber: 'SEAL-SEC-994102',
        armoredEscortId: 'ESCORT-ARM-04',
        timestamp: now.subtract(const Duration(hours: 4)),
      ),
      LogisticsTrackingEvent(
        eventId: 'EVT-02',
        orderId: orderId,
        carrier: LogisticsCarrier.sequelLogistics,
        trackingNumber: 'SEQL-ARM-8840192',
        status: RedemptionStatus.inTransit,
        statusDescription: 'Armored transit vehicle dispatched with armed escort under GPS telemetry.',
        locationCity: 'Bengaluru Logistics Hub',
        currentLatitude: 12.9716,
        currentLongitude: 77.5946,
        securityBagSealNumber: 'SEAL-SEC-994102',
        armoredEscortId: 'ESCORT-ARM-04',
        timestamp: now.subtract(const Duration(hours: 1)),
      ),
      LogisticsTrackingEvent(
        eventId: 'EVT-03',
        orderId: orderId,
        carrier: LogisticsCarrier.sequelLogistics,
        trackingNumber: 'SEQL-ARM-8840192',
        status: RedemptionStatus.outForDelivery,
        statusDescription: 'Out for doorstep delivery. Prepare government photo ID and handoff OTP.',
        locationCity: 'Bengaluru Indiranagar',
        currentLatitude: 12.9784,
        currentLongitude: 77.6408,
        securityBagSealNumber: 'SEAL-SEC-994102',
        armoredEscortId: 'ESCORT-ARM-04',
        timestamp: now,
      ),
    ];
  }

  @override
  Stream<LogisticsTrackingEvent> streamLiveTrackingEvents({
    required String orderId,
  }) {
    return Stream.periodic(const Duration(seconds: 10), (count) {
      final now = DateTime.now();
      return LogisticsTrackingEvent(
        eventId: 'EVT-LIVE-$count',
        orderId: orderId,
        carrier: LogisticsCarrier.sequelLogistics,
        trackingNumber: 'SEQL-ARM-8840192',
        status: count >= 2 ? RedemptionStatus.delivered : RedemptionStatus.outForDelivery,
        statusDescription: count >= 2
            ? 'Delivery successfully completed and verified.'
            : 'Armored vehicle approaching recipient destination.',
        locationCity: 'Bengaluru Indiranagar',
        currentLatitude: 12.9784 + (count * 0.0001),
        currentLongitude: 77.6408 + (count * 0.0001),
        securityBagSealNumber: 'SEAL-SEC-994102',
        armoredEscortId: 'ESCORT-ARM-04',
        timestamp: now,
      );
    });
  }

  @override
  Future<String> generateTimeLockedDeliveryOtp({
    required String orderId,
  }) async {
    final seed = '${orderId}_${DateTime.now().millisecondsSinceEpoch ~/ 30000}';
    final hash = sha256.convert(utf8.encode(seed)).toString();
    final digits = int.parse(hash.substring(0, 6), radix: 16) % 1000000;
    return digits.toString().padLeft(6, '7');
  }
}
