# 530 - Flutter Commodity Physical Delivery and Vault Redemption Flow (gGOLD, gSILVER, WDRA eNWR, Armored Logistics)

## Purpose
Provides a regulated, institutional-grade physical redemption interface for tokenized commodities (`gGOLD` 999/999.9 purity fine gold, `gSILVER` 999 purity fine silver) within the Growww multi-platform client (Android, iOS, Windows, macOS, Linux). Operating under the regulatory frameworks of the Warehousing Development and Regulatory Authority (WDRA), Multi Commodity Exchange of India (MCX), and Securities and Exchange Board of India (SEBI), every tokenized commodity unit held by investors is 1:1 backed by physical bullion lodged in accredited vaults and registered as electronic Negotiable Warehouse Receipts (eNWR) across National E-Repository Limited (NERL) or CDSL Commodity Repository Limited (CCRL).

When investors choose to convert digital commodity balances into physical bullion, the client application must facilitate an immutable physical redemption journey: selecting standardized bar denominations, choosing between self-pickup at WDRA-accredited vaults or insured armored doorstep courier delivery, computing transparent sub-paise fee breakdowns, executing dual-factor biometric and TOTP authorization, streaming live armored carrier logistics telemetry, and generating cryptographically signed eNWR burn delivery receipts.

## What You Are Building
A high-security, multi-screen physical redemption suite located in `apps/growww_flutter/lib/features/commodity_delivery/`:
- `CommodityRedemptionWizardScreen`: Multi-step redemption configurator allowing users to select commodity assets (`gGOLD`, `gSILVER`), pick bar denominations (e.g. 1g, 8g, 10g, 50g, 100g gold bars; 10g, 100g, 500g, 1kg silver bars), and choose delivery methods (WDRA Vault Self-Pickup vs Insured Armored Doorstep Courier).
- `WdraVaultLocatorScreen`: Interactive map and searchable list view of WDRA-accredited vault facilities across Indian bullion hubs (Mumbai, Ahmedabad, New Delhi, Jaipur, Bengaluru, Chennai, Kolkata, Hyderabad, Surat) displaying repository registration IDs, operating hours, available inventory denominations, and security protocols.
- `DoorstepDeliveryAddressScreen`: Delivery address configuration enforcing strict matching against KYC-verified Aadhaar/DigiLocker primary addresses, pincode serviceability validation, and tamper-evident courier dispatch instructions.
- `FeeBreakdownBottomSheet`: Transparent financial summary presenting sub-paise INR calculations for vault rematerialization charges, 0.00% (Zero Fee) platform settlement fee, assay packaging fees, transit insurance premium (100% replacement value coverage), and applicable statutory GST (3% on precious metals, 18% on logistics/vault handling).
- `BiometricDualFactorAuthModal`: Hardware-backed biometric challenge (`local_auth` FaceID / Fingerprint / Touch ID / Windows Hello) combined with time-based one-time password (TOTP) or SMS OTP verification for physical asset release authorization.
- `ArmoredLogisticsTrackerScreen`: Real-time shipment status portal for carrier integration (Brink's Secure Logistics, Sequel Logistics, BVC Logistics) displaying armored vehicle dispatch, GPS checkpoint timeline, tamper-evident bag seal numbers, and time-locked OTP generation for doorstep handoff.
- `DeliveryAcknowledgementScreen`: Cryptographically signed digital receipt modal presenting the on-chain eNWR token burn transaction hash, allocated physical bar serial numbers, BIS hallmark certification IDs, and dual-custody delivery completion signatures.
- `Riverpod State Layer`: Asynchronous notifiers (`CommodityRedemptionNotifier`, `VaultLocatorNotifier`, `LogisticsTrackerNotifier`, `DeliveryFeeNotifier`) managing reactive state transitions, live WebSocket tracking feeds, and offline cache synchronization.

## Scope Boundaries
- **In Scope:**
  - Multi-platform Flutter screens, bottom sheets, map overlays, and confirmation dialogs for commodity redemption.
  - Denomination selection matrices for `gGOLD` (1g, 8g, 10g, 50g, 100g) and `gSILVER` (10g, 100g, 500g, 1kg).
  - WDRA-accredited vault discovery, filtering by city and asset availability, and interactive map marker rendering.
  - KYC address matching verification against DigiLocker-linked investor profile.
  - Real-time sub-paise fee calculation engine displaying exact breakdowns (vault fees, 0.00% (Zero Fee) settlement fee, courier insurance, GST).
  - Dual-factor authentication modal integrating `local_auth` and TOTP/SMS validation.
  - Real-time carrier tracking integration with WebSocket updates and time-locked delivery OTP generation.
  - Digital delivery receipt generation with embedded QR codes, bar serials, and cryptographic hash verification.
- **Out of Scope / Handled Elsewhere:**
  - Direct repository API/SFTP integration with NERL and CCRL eNWR systems (Prompt 243).
  - Physical vault gate pass issuance, bar scanning, and armored carrier ERP integrations (Prompt 243).
  - On-chain smart contract token escrow and burning in `TokenRedemption.sol` (Prompt 304, Prompt 329).
  - Primary fiat payment gateway processing for delivery fees (Prompt 212, Prompt 511).
  - Daily Sparse Merkle Tree Proof-of-Reserve calculations (Prompt 243, Prompt 308).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+) for reactive state management.
- **Justification:** Riverpod provides deterministic asynchronous state handling (`AsyncValue`), modular dependency injection, and clean separation between UI presentation and redemption business logic.
- **Fixed-Point Financial Mathematics:** `decimal: ^2.3.3` for sub-paise INR arithmetic (0.0001 INR precision) ensuring zero floating-point calculation discrepancies.
- **Hardware Biometrics:** `local_auth: ^2.2.0` for biometric signature challenges on mobile and desktop platforms.
- **Dynamic Vector QR Codes:** `qr_flutter: ^4.1.0` for rendering crisp pickup gate passes and receipt verification QR payloads.
- **Interactive Mapping:** `google_maps_flutter: ^2.6.0` (with platform-adaptive fallback for desktop) for rendering WDRA vault locations.
- **Networking & Real-Time Streams:** `dio: ^5.4.3+1` for typed REST API communication and `web_socket_channel: ^3.0.0` for real-time logistics tracking updates.
- **Cryptographic Hashing:** `crypto: ^3.0.3` for client-side SHA-256 checksum verification of eNWR receipts and delivery certificates.

## Backend / Infra Touchpoints
- **Vault Directory API (`GET /api/v1/commodity/vaults`):** Fetches the directory of WDRA-accredited vaults, repository registrations, address coordinates, and active inventory.
- **Denominations & Availability API (`GET /api/v1/commodity/inventory/denominations`):** Retrieves supported bar sizes, live vault stock levels, and rematerialization availability per commodity.
- **Delivery Fee Estimation API (`POST /api/v1/commodity/delivery/fee-estimate`):** Computes exact sub-paise fee breakdown including vault handling, 0.00% (Zero Fee) platform fee, armored courier charges, transit insurance, and GST.
- **Pincode Serviceability & Address Verification API (`POST /api/v1/commodity/delivery/check-serviceability`):** Validates destination pincode against armored carrier service zones and verifies match against KYC record.
- **Redemption Intent API (`POST /api/v1/commodity/redemption/initiate`):** Creates a pending redemption order, reserves physical bar inventory, and issues an escrow challenge.
- **Dual-Factor Authorization API (`POST /api/v1/commodity/redemption/authorize`):** Submits signed biometric proof and TOTP token to trigger eNWR rematerialization and on-chain token lockup.
- **Real-Time Logistics Tracking WebSocket (`wss://ws.growww.in/v1/commodity/logistics/{order_id}`):** Streams live carrier status updates, GPS checkpoints, armored vehicle dispatch timestamps, and delivery milestones.
- **Delivery Acknowledgement API (`POST /api/v1/commodity/redemption/acknowledge`):** Submits signed delivery confirmation and retrieves final cryptographically signed eNWR burn receipt.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **1:1 Custody Backing and Escrow Lockup:** When a redemption intent is authorized, the client requests the backend to lock the corresponding `gGOLD` or `gSILVER` ERC-3643 tokens in `TokenRedemption.sol` on Hyperledger Besu.
- **eNWR Immobilization and Burn Proof:** Upon physical bar handover (at the vault or doorstep), the repository extinguishes the eNWR in NERL/CCRL. The backend triggers the smart contract token burn, generating an on-chain burn transaction hash linked to the eNWR rematerialization ID.
- **Zero On-Chain PII Guarantee:** No investor names, contact numbers, or delivery addresses are written to the blockchain. The ledger records only pseudonymous `wallet_address`, `redemption_id` UUID, commodity fine grams burned, and cryptographic SHA-256 digests of delivery receipts.
- **Proof-of-Reserve Reconciliation:** Decrements on-chain token supply synchronously with physical vault unencumbrance, ensuring total token supply matches remaining vaulted physical grams at all times.

## State Management Architecture
- **CommodityRedemptionNotifier (`AutoDisposeAsyncNotifier<RedemptionState>`):** Orchestrates multi-step redemption flow, tracking selected commodity type, chosen bar denomination quantities, delivery method (vault vs doorstep), selected address/vault, and transaction lifecycle.
- **VaultLocatorNotifier (`AutoDisposeAsyncNotifier<VaultLocatorState>`):** Manages vault discovery, distance sorting from user location, city filtering, and repository accreditation metadata.
- **DeliveryFeeNotifier (`AutoDisposeAsyncNotifier<DeliveryFeeState>`):** Reactively calculates sub-paise fee breakdowns whenever denomination quantities or delivery destinations change.
- **LogisticsTrackerNotifier (`AutoDisposeStreamNotifier<LogisticsTrackingState>`):** Connects to the real-time carrier tracking WebSocket, maintaining status timeline, GPS updates, carrier contact data, and time-locked delivery OTP.
- **DeliveryReceiptNotifier (`AutoDisposeAsyncNotifier<DeliveryReceiptState>`):** Manages retrieval, cryptographic signature verification, and PDF/image export of finalized delivery certificates.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold feature directory structure under `apps/growww_flutter/lib/features/commodity_delivery/`: `presentation/screens/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, `data/datasources/`.
2. Define domain enums and immutable data models in `domain/models/`: `CommodityType`, `DeliveryMode`, `RedemptionStatus`, `LogisticsCarrier`, `CommodityDenomination`, `WdraVaultLocation`, `DeliveryAddress`, `SubPaiseFeeBreakdown`, `RedemptionOrder`, `LogisticsTrackingEvent`, and `DeliveryReceipt`.
3. Implement `ICommodityRedemptionRepository` and `ILogisticsTrackingRepository` interfaces defining contract methods for inventory queries, fee calculations, authorization dispatch, tracking streams, and receipt verification.
4. Implement `DeliveryFeeNotifier` utilizing fixed-point `Decimal` arithmetic to compute sub-paise vault handling fees, 0.00% (Zero Fee) platform settlement fees, courier insurance premiums, and statutory GST splits.
5. Build `CommodityRedemptionWizardScreen`:
   - Asset selector toggle between `gGOLD` (Gold 999/999.9) and `gSILVER` (Silver 999).
   - Available balance card displaying total fine grams owned and redeemable balance.
   - Denomination quantity selector grid with incremental counter buttons and minimum redemption weight guards (e.g. minimum 1g gold, minimum 10g silver).
   - Total redemption weight summary with gross weight and fine weight conversion.
6. Build `DeliveryModeSelectorWidget`:
   - Dual-tab or segmented control allowing selection between "WDRA Vault Self-Pickup" and "Insured Armored Doorstep Delivery".
   - Contextual information cards detailing ID requirements for vault pickup and courier security protocols for doorstep delivery.
7. Build `WdraVaultLocatorScreen`:
   - Interactive map view displaying WDRA vault markers with custom bullion brand pin icons.
   - Searchable list view with city filter chips (Mumbai, Ahmedabad, Delhi NCR, Jaipur, etc.).
   - Vault detail card presenting repository accreditation ID (NERL/CCRL), vault operator name (e.g. Sequel, Brink's, MMTC-PAMP), operating hours, security instructions, and distance.
8. Build `DoorstepDeliveryAddressScreen`:
   - Primary KYC-registered address card fetched from investor profile with "Verified" badge.
   - Pincode serviceability checker displaying instant armored logistics coverage status and estimated transit duration.
   - Secondary contact and delivery instructions form (landmark, recipient gate pass notes).
9. Build `FeeBreakdownBottomSheet`:
   - Itemized fee list displaying Vault Rematerialization Fee, 0.00% (Zero Fee) Platform Settlement Fee, Assayed Packaging Fee, Armored Courier & 100% Insurance Premium, and GST (3% precious metal tax, 18% services tax).
   - High-precision sub-paise display formatted as standard Indian Rupee notation (e.g. ₹124.5050) with net payable amount.
   - Payment method selector for fee settlement (Demat cash ledger balance vs instant UPI).
10. Build `BiometricDualFactorAuthModal`:
    - Summary of redemption order (asset, fine grams, bar serial count, delivery destination, total fee).
    - Biometric challenge integration via `local_auth` verifying device biometric signature.
    - TOTP 2FA input field with automatic clipboard paste detection and fallback to SMS OTP with 60-second cooldown timer.
    - Idempotency key generation to prevent duplicate redemption requests.
11. Build `ArmoredLogisticsTrackerScreen`:
    - Order summary banner with real-time status badge (Order Placed -> Vault Allocated -> eNWR Rematerialized -> Armored Dispatch -> In Transit -> Out for Delivery -> Delivered).
    - Live tracking timeline rendering carrier checkpoints with timestamps, vehicle tracking status, and security escort details.
    - Tamper-evident security pouch seal number verification card.
    - Time-locked 6-digit delivery OTP generator with dynamic countdown timer for recipient handoff.
12. Build `DeliveryAcknowledgementScreen`:
    - Success animation with verified checkmark and institutional badge.
    - Digital delivery certificate showing on-chain eNWR token burn transaction hash, bar serial numbers, BIS hallmark certificate number, and carrier handoff timestamp.
    - Dynamic QR code embedding cryptographic receipt verification payload.
    - One-tap actions for downloading tamper-proof PDF receipt and opening transaction on Besu block explorer.
13. Implement comprehensive unit tests verifying sub-paise fee calculations, denomination constraints, pincode validation, and receipt SHA-256 signature verification.
14. Implement widget and integration tests for `CommodityRedemptionWizardScreen`, `WdraVaultLocatorScreen`, `FeeBreakdownBottomSheet`, and `ArmoredLogisticsTrackerScreen`.

## Interfaces / Contracts

```dart
// lib/features/commodity_delivery/domain/models/commodity_enums.dart

enum CommodityType {
  gold999,
  gold9999,
  silver999,
}

enum DeliveryMode {
  vaultPickup,
  armoredDoorstep,
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

enum LogisticsCarrier {
  sequelLogistics,
  brinksSecure,
  bvcLogistics,
  malcaAmit,
}

// lib/features/commodity_delivery/domain/models/commodity_denomination.dart

class CommodityDenomination {
  final String denominationId;
  final CommodityType commodityType;
  final double weightGrams;
  final double fineness;
  final String displayName;
  final String barFormFactor; // e.g. "Minted Bar", "Cast Bar", "Coin"
  final int availableStockUnits;
  final String assayStandard; // e.g. "BIS 1417:2016", "LBMA Good Delivery"
  final String packagingType; // e.g. "CertiCard Blister", "Tamper-Evident Capsule"

  const CommodityDenomination({
    required this.denominationId,
    required this.commodityType,
    required this.weightGrams,
    required this.fineness,
    required this.displayName,
    required this.barFormFactor,
    required this.availableStockUnits,
    required this.assayStandard,
    required this.packagingType,
  });
}

// lib/features/commodity_delivery/domain/models/wdra_vault_location.dart

class WdraVaultLocation {
  final String vaultId;
  final String repositoryName; // e.g. "National E-Repository Limited (NERL)"
  final String repositoryRegistrationNo;
  final String operatorName; // e.g. "MMTC-PAMP India Pvt Ltd", "Sequel Vaulting"
  final String facilityName;
  final String addressLine;
  final String city;
  final String state;
  final String pincode;
  final double latitude;
  final double longitude;
  final String operatingHours;
  final List<CommodityType> supportedCommodities;
  final List<String> pickupRequirements; // e.g. "Original PAN Card", "Aadhaar Card"
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
}

// lib/features/commodity_delivery/domain/models/delivery_address.dart

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
  final String kycSource; // e.g. "DigiLocker / Aadhaar"

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
}

// lib/features/commodity_delivery/domain/models/sub_paise_fee_breakdown.dart

class SubPaiseFeeBreakdown {
  final int vaultRematerializationFeeSubPaise; // 1 INR = 10000 sub-paise
  final int settlementFeeSubPaise; // 0.00% (Zero Fee) platform settlement fee in sub-paise
  final int assayPackagingFeeSubPaise;
  final int armoredLogisticsFeeSubPaise;
  final int transitInsuranceFeeSubPaise;
  final int preciousMetalGstSubPaise; // 3% GST
  final int handlingServicesGstSubPaise; // 18% GST on services
  final int totalGrossPayableSubPaise;

  const SubPaiseFeeBreakdown({
    required this.vaultRematerializationFeeSubPaise,
    required this.settlementFeeSubPaise,
    required this.assayPackagingFeeSubPaise,
    required this.armoredLogisticsFeeSubPaise,
    required this.transitInsuranceFeeSubPaise,
    required this.preciousMetalGstSubPaise,
    required this.handlingServicesGstSubPaise,
    required this.totalGrossPayableSubPaise,
  });

  double get totalInr => totalGrossPayableSubPaise / 10000.0;
  double get settlementFeeInr => settlementFeeSubPaise / 10000.0;
  double get vaultFeeInr => vaultRematerializationFeeSubPaise / 10000.0;
  double get logisticsFeeInr => armoredLogisticsFeeSubPaise / 10000.0;
  double get insuranceFeeInr => transitInsuranceFeeSubPaise / 10000.0;
  double get totalGstInr => (preciousMetalGstSubPaise + handlingServicesGstSubPaise) / 10000.0;
}

// lib/features/commodity_delivery/domain/models/redemption_order.dart

class DenominationOrderItem {
  final CommodityDenomination denomination;
  final int quantity;

  const DenominationOrderItem({
    required this.denomination,
    required this.quantity,
  });

  double get totalWeightGrams => denomination.weightGrams * quantity;
}

class RedemptionOrder {
  final String orderId;
  final String userId;
  final CommodityType commodityType;
  final List<DenominationOrderItem> items;
  final double totalFineGrams;
  final DeliveryMode deliveryMode;
  final WdraVaultLocation? selectedVault;
  final DeliveryAddress? deliveryAddress;
  final SubPaiseFeeBreakdown feeBreakdown;
  final RedemptionStatus status;
  final String? enwrNumber;
  final String? onChainBurnTxHash;
  final String? idempotencyKey;
  final DateTime createdAt;
  final DateTime? estimatedDeliveryDate;

  const RedemptionOrder({
    required this.orderId,
    required this.userId,
    required this.commodityType,
    required this.items,
    required this.totalFineGrams,
    required this.deliveryMode,
    this.selectedVault,
    this.deliveryAddress,
    required this.feeBreakdown,
    required this.status,
    this.enwrNumber,
    this.onChainBurnTxHash,
    this.idempotencyKey,
    required this.createdAt,
    this.estimatedDeliveryDate,
  });
}

// lib/features/commodity_delivery/domain/models/logistics_tracking_event.dart

class LogisticsTrackingEvent {
  final String eventId;
  final String orderId;
  final LogisticsCarrier carrier;
  final String trackingNumber;
  final RedemptionStatus status;
  final String statusDescription;
  final String locationCity;
  final double? currentLatitude;
  final double? currentLongitude;
  final String securityBagSealNumber;
  final String? armoredEscortId;
  final DateTime timestamp;

  const LogisticsTrackingEvent({
    required this.eventId,
    required this.orderId,
    required this.carrier,
    required this.trackingNumber,
    required this.status,
    required this.statusDescription,
    required this.locationCity,
    this.currentLatitude,
    this.currentLongitude,
    required this.securityBagSealNumber,
    this.armoredEscortId,
    required this.timestamp,
  });
}

// lib/features/commodity_delivery/domain/models/delivery_receipt.dart

class DeliveredBarDetail {
  final String barSerialNumber;
  final double grossWeightGrams;
  final double fineness;
  final String refineryName;
  final String bisHallmarkNumber;
  final String assayCertificateUrl;

  const DeliveredBarDetail({
    required this.barSerialNumber,
    required this.grossWeightGrams,
    required this.fineness,
    required this.refineryName,
    required this.bisHallmarkNumber,
    required this.assayCertificateUrl,
  });
}

class DeliveryReceipt {
  final String receiptId;
  final String orderId;
  final String enwrNumber;
  final String repositoryName;
  final String onChainBurnTxHash;
  final List<DeliveredBarDetail> deliveredBars;
  final double totalFineGramsDelivered;
  final String recipientName;
  final String verificationSignature;
  final String deliveryQrPayload;
  final DateTime completedAt;

  const DeliveryReceipt({
    required this.receiptId,
    required this.orderId,
    required this.enwrNumber,
    required this.repositoryName,
    required this.onChainBurnTxHash,
    required this.deliveredBars,
    required this.totalFineGramsDelivered,
    required this.recipientName,
    required this.verificationSignature,
    required this.deliveryQrPayload,
    required this.completedAt,
  });
}

// lib/features/commodity_delivery/domain/interfaces/i_commodity_redemption_repository.dart

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

// lib/features/commodity_delivery/domain/interfaces/i_logistics_tracking_repository.dart

abstract class ILogisticsTrackingRepository {
  Future<List<LogisticsTrackingEvent>> fetchTrackingHistory({
    required String orderId,
  });

  Stream<LogisticsTrackingEvent> streamLiveTrackingEvents({
    required String orderId,
  });

  Future<String> generateTimeLockedDeliveryOtp({
    required String orderId,
  });
}
```

## Security & Compliance Notes
- **WDRA & SEBI Electronic Commodity Regulatory Compliance:** Physical redemption strictly interfaces with WDRA-accredited repositories (NERL/CCRL). No unencumbered bullion is released without verifiable electronic rematerialization instructions.
- **KYC Address Matching Policy:** To prevent fraudulent diversion of high-value bullion, doorstep deliveries are strictly restricted to primary residential addresses verified through DigiLocker / Aadhaar during user onboarding (Prompt 504). Address modifications trigger re-KYC and a 48-hour cooling period.
- **Sub-Paise Precision & Financial Math:** All fee calculations are handled in sub-paise integer units (1 INR = 10,000 sub-paise) to prevent fractional rounding errors across vault charges, 0.00% (Zero Fee) settlement fees, and GST splits.
- **Biometric Dual-Factor Physical Release:** Authorizing physical bullion redemption requires hardware biometric verification (`local_auth`) paired with a dynamic TOTP / SMS factor, preventing unauthorized withdrawal in the event of device compromise.
- **Armored Logistics Chain-of-Custody:** Physical transit uses insured armored logistics partners (Brink's, Sequel, BVC) with dual-custody vault handoff, tamper-evident numbered seals, vehicle telemetry, and secure time-locked OTP verification at doorstep handover.
- **Zero PII on Distributed Ledger:** Smart contract redemption records (`TokenRedemption.sol`) store only pseudonymous wallet identifiers, order UUIDs, fine weights, and cryptographic receipt hashes, maintaining absolute user privacy on-chain.

## Acceptance Criteria
- [ ] Commodity selector seamlessly switches between `gGOLD` and `gSILVER` with live balance updates.
- [ ] Denomination matrix allows picking valid bar quantities with real-time weight summation and minimum redemption threshold enforcement.
- [ ] WDRA vault locator renders interactive map markers and searchable list filtered by city and repository accreditation.
- [ ] Pincode serviceability check validates armored logistics coverage and enforces KYC address match.
- [ ] `FeeBreakdownBottomSheet` displays transparent sub-paise INR breakdowns including 0.00% (Zero Fee) settlement fee, vault fees, insurance, and GST.
- [ ] `BiometricDualFactorAuthModal` challenges the user with device biometrics and TOTP before submitting authorization.
- [ ] Real-time carrier tracking screen streams status updates and GPS checkpoints over WebSocket.
- [ ] Tamper-evident seal number and dynamic time-locked OTP generator render accurately during delivery handoff.
- [ ] `DeliveryAcknowledgementScreen` renders on-chain token burn hash, individual bar serials, BIS hallmark numbers, and verifiable QR payload.
- [ ] Specification contains zero application implementation code (only domain contracts, interfaces, and architecture).
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Scaffolding), Prompt 502 (Riverpod Architecture), Prompt 503 (Design System & Theming), Prompt 504 (Onboarding & KYC Flow), Prompt 505 (Authentication UI), Prompt 521 (Secure Storage), Prompt 525 (API Client Layer).
- **Backend Dependencies:** Prompt 202 (KYC & Sanctions Screening Service), Prompt 203 (Wallet & Ledger Service), Prompt 210 (Fee & PnL Engine), Prompt 211 (Notification Service), Prompt 243 (MCX Commodity & Warehouse Receipt Adapter).
- **Blockchain Dependencies:** Prompt 304 (Token Redemption Service), Prompt 308 (Proof-of-Reserve Registry Contract), Prompt 329 (NBSE Settlement DvP & Fee Collector).
- **Downstream Blockers:** Prompt 510 (Portfolio Holdings Screen), Prompt 512 (Transaction History Screen), Prompt 902 (End-to-End Integration Testing).
