# 611 - Web Commodity Physical Delivery Portal & Institutional Vault Management

## Purpose
Democratizing access to institutional-grade physical commodities and bullion redemption requires a secure, high-integrity web interface tailored for institutional treasury desks, family offices, and certified bullion dealers. Operating in strict alignment with SEBI Electronic Gold Receipt (EGR) frameworks, WDRA (Warehousing Development and Regulatory Authority) Act 2007 guidelines, MCX physical delivery bylaws, and IFSCA bullion custody standards, institutional investors require seamless capability to inspect vaulted inventory, request physical redemption of digital commodity tokens (such as Gold 999.9/995.0 and Silver 999.0), verify NABL-accredited assay certificates, authorize high-value custody releases via multi-signature governance, and track armored transport telemetry in real time.

This prompt specifies the architecture and implementation of the Next.js 14 Institutional Web Commodity Delivery Portal and Vault Management Workstation (`apps/growww_web/app/(app)/commodities/delivery/`, `apps/growww_web/app/(app)/commodities/vault/`, and `apps/growww_web/app/(app)/commodities/transit/`). The module delivers an interactive geospatial vault explorer across WDRA-accredited depositories (CCRL and NERL), a multi-step physical redemption wizard with automated 0.00% (Zero Fee) platform fee and insured freight calculation, an institutional multi-signature custody release authorization system for large lots (>100g Gold / >1kg Silver), an NABL assay certificate cryptographic verifier, and a real-time secure armored transit telemetry tracking dashboard integrated with Brink's and Sequel logistics feeds.

## What You Are Building
A high-performance Next.js 14 institutional physical commodity redemption and vault management interface featuring:
- `VaultInventoryMap`: Interactive geospatial Mapbox/Leaflet vault explorer displaying WDRA-accredited depository locations across India (Mumbai, Ahmedabad, GIFT City, New Delhi, Bengaluru, Chennai, Hyderabad, Kolkata), showing live bullion stock levels, depository type (CCRL vs NERL), available bar denominations (10g, 50g, 100g, 1kg Gold; 100g, 500g, 1kg, 30kg Silver), and vault accreditation status.
- `CommodityRedemptionWizard`: Multi-step institutional delivery redemption workflow guiding users through depository selection, physical bar denomination selection, fee estimation, destination address registration, and settlement escrow locking.
- `MultiSigCustodyReleaseModal`: Multi-signature custody release authorization modal for institutional lots (>100g Gold / >1kg Silver) enforcing Maker-Checker institutional governance, Web3 MPC/Hardware key (Trezor, Ledger, MetaMask Institutional) dual-signature approvals, and corporate treasury limits.
- `AssayCertificateVerifier`: Digital inspection interface rendering bar serial number allocations, BIS hallmark badges, and downloadable NABL-accredited assay certificates with client-side SHA-256 and Keccak-256 hash validation against on-chain records and IPFS document registries.
- `FeeAndFreightCalculator`: Precision pricing engine computing the 0.00% (Zero Fee) platform redemption fee, dynamic insured freight shipping costs based on distance and armored courier rate cards, comprehensive transit insurance premiums, and applicable GST breakdown.
- `ArmoredTransitTelemetryDashboard`: Real-time secure logistics tracker integrating Brink's and Sequel GPS telemetry streams, dynamic route maps, lock tamper status indicators, dual-custody vault exit/entry timestamps, and OTP/dual-biometric handover verification on final delivery.
- `VaultAttestationReportViewer`: Institutional depository audit viewer displaying daily electronic Negotiable Warehouse Receipt (e-NWR) balances, vault custodian signatures, and physical audit inspection logs.

## Scope Boundaries
- **In Scope:**
  - Responsive Next.js 14 web interface for commodity physical delivery redemption, vault exploration, and transit monitoring.
  - Interactive geospatial vault map with dynamic filtering by commodity type, depository (CCRL/NERL), and bar denomination availability.
  - Client-side multi-signature signature collection and institutional Maker-Checker approval workflow management.
  - PDF viewer for NABL assay certificates with client-side cryptographic SHA-256 hash verification against on-chain contract state.
  - Real-time fee calculation engine (0.00% (Zero Fee) platform fee, armored freight, transit insurance, GST).
  - Armored logistics telemetry dashboard visualizing real-time GPS coordinates, vehicle lock status, and milestone progression.
  - Delivery receipt attestation flow with dual-party cryptographic verification.
- **Out of Scope / Handled Elsewhere:**
  - Commodity token smart contracts and vault registry (`PhysicalVaultRegistry.sol`, `CommodityToken.sol` - Prompt 330).
  - MCX commodity and warehouse repository adapter backend service (`services/commodity-vault-adapter` - Prompt 243).
  - Physical vault audit and inspector microservice (`services/vault-audit-portal` - Prompt 716).
  - Double-entry wallet ledger and fiat settlement service (Prompt 203, Prompt 214).
  - Core order matching engine for commodity spot and derivatives (Prompt 205, Prompt 240).
  - Sanctions and AML screening microservice (Prompt 703).

## Technology to Use
- **Next.js 14 App Router (React Server Components & Client Components):** Delivers server-rendered initial shell and streaming client components for dynamic vault pricing, real-time telemetry, and cryptographic verification.
- **Tailwind CSS 3.4+ & shadcn/ui (Radix UI primitives):** Provides an institutional dark-themed interface with precision tables, badges, dialogs, progress bars, and accessible form controls.
- **Mapbox GL JS / React-Map-GL:** Powers interactive geospatial rendering of WDRA vault facilities, depository clusters, and live armored vehicle transit trajectories with custom bullion marker pins.
- **Lucide React & Framer Motion:** Delivers smooth transit telemetry milestone transitions, vehicle animation, lock-tamper status pulses, and security badge indicators.
- **Viem 2.x & Wagmi 2.x:** Handles EIP-712 typed data signing for multi-sig custody release approvals and executes direct read queries against Hyperledger Besu smart contracts.
- **PDF.js / `@react-pdf-viewer/core`:** Provides high-fidelity in-browser rendering of NABL assay certificates with digital signature stamp overlays.
- **`@noble/hashes` (SHA-256) & Viem (`keccak256`):** Executes client-side cryptographic hashing on downloaded PDF certificates to verify document immutability against on-chain registered hashes.
- **Decimal.js:** Executes exact multi-denomination weight conversions (grams, kilograms, troy ounces), 0.00% (Zero Fee) platform fee calculations, and GST itemizations without floating-point inaccuracies.
- **Zustand with Immer:** Ephemeral state store managing active delivery drafts, selected bar serial allocations, multi-sig approval states, and real-time transit telemetry streams.

## Backend / Infra Touchpoints
- **MCX Commodity & Warehouse Receipt Adapter (Prompt 243):** Connects to `https://api.growww.in/api/v1/commodities/delivery` and `wss://api.growww.in/ws/v1/commodities/telemetry` for depository metadata (CCRL/NERL), bar allocation, and armored transit tracking.
- **Physical Vault Audit & Inspector Portal (Prompt 716):** Queries `GET /api/v1/vault-audit/bars/{serialNumber}` for NABL assay certificate metadata and vault inspection records.
- **Hyperledger Besu JSON-RPC Node (Prompt 302, Prompt 330):** Queries `PhysicalVaultRegistry.sol` and `CommodityToken.sol` for on-chain lot reservations, token burns, and custody release events.
- **Double-Entry Wallet Account Ledger Service (Prompt 203):** Validates investor digital commodity token balances and reserves tokens during the redemption escrow phase.
- **Secure Logistics Gateway (Brink's / Sequel Adapter):** Ingests webhook updates and WebSocket telemetry feeds for armored transport vehicle GPS coordinates, sensor readings, and custody handovers.
- **IPFS / Decentralized Object Storage (Prompt 401):** Resolves immutable NABL assay certificate PDFs and vault inspection reports via decentralized content identifiers (CIDs).
- **Audit Log Service (Prompt 218):** Ingests non-repudiable audit trails of multi-sig custody release authorizations, Maker-Checker actions, and final delivery acknowledgments.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus and deterministic 2-second block finality).
- **Core Smart Contracts Interfaced:**
  - `PhysicalVaultRegistry.sol`: Maintains vault registrations, accredited WDRA statuses, bar serial numbers, e-NWR IDs, assay certificate hashes, and custody lifecycle states (`VAULTED`, `RELEASE_REQUESTED`, `RELEASE_APPROVED`, `IN_TRANSIT`, `DELIVERED`).
  - `CommodityToken.sol` (e.g., `GoldToken.sol` 999.9/995.0, `SilverToken.sol` 999.0): Locks digital commodity tokens in escrow upon redemption request and executes permanent token burns once physical custody is released.
  - `InstitutionalMultiSigAuth.sol`: Validates M-of-N multi-signature authorizations on-chain for physical redemptions exceeding institutional lot thresholds (>100g Gold / >1kg Silver).
- **On-Chain Lifecycle & Cryptographic Verification:**
  - Step 1: Investor initiates redemption request, locking digital commodity tokens in `PhysicalVaultRegistry.sol` escrow.
  - Step 2: For institutional lots (>100g Gold / >1kg Silver), the web interface collects M-of-N EIP-712 signatures from authorized corporate signers and submits `authorizeCustodyRelease(...)`.
  - Step 3: Vault custodian allocates physical bar serial numbers. The web portal fetches the associated NABL assay certificate PDF from IPFS, computes its SHA-256 hash client-side, and verifies that it matches `assayCertificateHash` stored on-chain.
  - Step 4: Upon dispatch by armored carrier (Brink's/Sequel), the carrier issues an on-chain dispatch attestation with tracking number and dispatch commitment hash.
  - Step 5: On physical receipt, the client verifies delivery with dual-custody OTP and biometric signature, triggering the on-chain token burn and final delivery settlement event.
- **Zero PII Invariant:** Delivery recipient physical street addresses, contact numbers, and armored carrier driver identities are end-to-end encrypted with the carrier public key and never published in plaintext on the blockchain.

## Step-by-Step Build Instructions
1. Scaffold commodity physical delivery routes under `apps/growww_web/app/(app)/commodities/`:
   - `delivery/page.tsx`: Main physical delivery redemption wizard and active delivery order list.
   - `delivery/[orderId]/page.tsx`: Detailed delivery order status, bar allocation inspection, and NABL assay verifier.
   - `vault/page.tsx`: Interactive WDRA-accredited depository explorer and live vault stock map.
   - `transit/[deliveryId]/page.tsx`: Real-time Brink's / Sequel secure transit telemetry tracking dashboard.
2. Implement `VaultInventoryMap` component using Mapbox GL JS:
   - Render interactive map pins for accredited vault facilities in Mumbai, Ahmedabad, GIFT City, New Delhi, Bengaluru, Chennai, Hyderabad, and Kolkata.
   - Provide visual markers distinguishing CCRL (CDSL Commodity Repository Limited) and NERL (National E-Repository Limited) facilities.
   - Implement popup cards displaying vault capacity, active stock (Gold 999.9, Gold 995.0, Silver 999.0), WDRA accreditation validity, security grade, and operational hours.
   - Add filter controls for commodity type, purity grade, minimum available weight, and depository type.
3. Build `CommodityRedemptionWizard` step-by-step workflow:
   - Step 1 (Asset & Lot Selection): Select commodity asset, target redemption weight, and preferred physical bar denominations (e.g., 10x 10g Gold bars vs 1x 100g Gold bar).
   - Step 2 (Depository & Vault Selection): Choose source vault location based on proximity, stock availability, or CCRL/NERL preference.
   - Step 3 (Delivery Method & Address): Choose between secure armored door delivery (Brink's / Sequel) or in-person vault pickup with KYC verification.
   - Step 4 (Fee & Freight Calculation): Display dynamic cost breakdown (0.00% (Zero Fee) platform fee, insured armored transport fee, transit insurance premium, 18% GST).
   - Step 5 (Review & Token Escrow): Confirm redemption request and execute on-chain token escrow reservation via Viem.
4. Implement `MultiSigCustodyReleaseModal` component:
   - Automatically detect if redemption weight exceeds institutional threshold (>100g Gold or >1kg Silver).
   - Render Maker-Checker approval workflow showing required signatures (e.g., 2-of-3 or 3-of-5 signers).
   - Integrate Viem/Wagmi EIP-712 structured typed data signing modal for institutional signers using hardware wallets (Ledger, Trezor) or MPC wallets.
   - Display live signature status checklist showing each authorized signer, approval timestamp, public key, and cryptographic signature hash.
5. Build `AssayCertificateVerifier` component:
   - Render allocated bar serial numbers with refiner details (e.g., MMTC-PAMP, Valcambi, Rand Refinery, Bangalore Refinery), gross weight, fineness, and BIS hallmark ID.
   - Embed PDF viewer loading the official NABL-accredited assay certificate from decentralized storage (IPFS).
   - Execute client-side SHA-256 hash computation over the downloaded PDF buffer using `@noble/hashes`.
   - Compare computed hash against the immutable `assayCertificateHash` retrieved from `PhysicalVaultRegistry.sol`.
   - Display visual cryptographic verification badge ("Cryptographically Verified with On-Chain Ledger" with green checkmark, or "Hash Mismatch Alert" with red warning).
6. Implement `FeeAndFreightCalculator` utility and UI component:
   - Calculate 0.00% (Zero Fee) platform redemption fee: `platformFeeInr = 0.00 (0.00% fee at launch)`.
   - Estimate insured freight cost using origin vault pincode, destination pincode, weight tier, and courier rate matrix.
   - Calculate transit insurance premium based on declared bullion value and distance bracket.
   - Compute applicable GST itemization (18% on platform and transport services) and total invoice payable.
7. Build `ArmoredTransitTelemetryDashboard` component:
   - Connect to WebSocket streaming endpoint (`wss://api.growww.in/ws/v1/commodities/telemetry`) for live transit updates.
   - Render dynamic route map displaying source vault origin, real-time armored vehicle GPS coordinates, planned transit route, and destination.
   - Build security status monitor displaying lock tamper sensor status (Normal / Alert), temperature/vibration telemetry, escort team count, and vehicle ignition status.
   - Implement milestone stepper tracking transit lifecycle stages: Vault Lodgement -> Custody Handover -> In Transit -> Hub Arrival -> Out for Delivery -> Handover Verification.
8. Implement Secure Handover Verification Flow:
   - Generate time-based one-time delivery authorization code (TOTP) and dual-biometric confirmation prompt for delivery agent and recipient.
   - Provide digital delivery acknowledgment signing with recipient Web3 key.
   - Submit completion proof to backend and update on-chain status to `DELIVERED`.
9. Build Vault Attestation & e-NWR Document Center:
   - Provide downloadable electronic Negotiable Warehouse Receipts (e-NWR) registered with CCRL/NERL.
   - Render historical vault audit inspection reports with auditor sign-offs and Merkle proof attestations.
10. Integrate comprehensive unit tests using Vitest and end-to-end institutional user journey tests with Playwright.

## Interfaces / Contracts

### Vault & Depository Schemas
```typescript
export type CommodityType = 'GOLD_9999' | 'GOLD_9950' | 'SILVER_9990' | 'PLATINUM_9995';

export type DepositoryType = 'CCRL' | 'NERL';

export type VaultSecurityRating = 'GRADE_AAA' | 'GRADE_AA' | 'GRADE_A';

export interface WDRAVaultDepository {
  vaultId: string;
  vaultCode: string;
  vaultName: string;
  operatorName: string; // e.g. "Brink's India", "Sequel Logistics", "MMTC-PAMP"
  depositoryType: DepositoryType;
  wdraRegistrationNumber: string;
  wdraAccreditationExpiry: string; // ISO date string
  securityRating: VaultSecurityRating;
  city: string;
  state: string;
  pincode: string;
  latitude: number;
  longitude: number;
  contactEmail: string;
  operatingHours: string;
  isActive: boolean;
  supportedCommodities: CommodityType[];
}

export interface VaultInventoryItem {
  vaultId: string;
  commodityType: CommodityType;
  totalWeightGrams: string;
  availableWeightGrams: string;
  lockedInRedemptionGrams: string;
  barDenominations: Array<{
    denominationGrams: number;
    availableCount: number;
    allocatedCount: number;
  }>;
  lastAuditedAt: string;
  auditorName: string;
}
```

### Commodity Bar & NABL Assay Schemas
```typescript
export interface AllocatedCommodityBar {
  barSerialNumber: string;
  commodityType: CommodityType;
  vaultId: string;
  refinerName: string; // e.g. "MMTC-PAMP India", "Valcambi SA", "Rand Refinery"
  refinerBarId: string;
  grossWeightGrams: number;
  netPureWeightGrams: number;
  finenessPurity: string; // e.g. "999.9", "995.0", "999.0"
  bisHallmarkNumber: string;
  enwrReceiptNumber: string;
  assayCertificateId: string;
  assayReportUrl: string;
  assayCertificateHash: `0x${string}`; // Keccak-256 or SHA-256 hash
  vaultSlotCoordinates: string; // e.g. "VAULT-MUM-01:RACK-4:SHELF-B:BIN-12"
  allocatedAt: string;
}

export interface NABLAssayCertificate {
  certificateId: string;
  barSerialNumber: string;
  laboratoryName: string; // e.g. "NABL Accredited Bullion Testing Center"
  nablAccreditationNumber: string;
  assayMethod: 'FIRE_ASSAY' | 'XRF_SPECTROMETRY' | 'ICP_OES';
  testedPurityPercentage: number;
  testedWeightGrams: number;
  leadAssayerName: string;
  assayDate: string;
  digitalSignatureStamp: string;
  ipfsDocumentCid: string;
  sha256Hash: `0x${string}`;
}

export interface BarAssayVerificationResult {
  barSerialNumber: string;
  calculatedSha256Hash: `0x${string}`;
  onChainRegisteredHash: `0x${string}`;
  isHashMatching: boolean;
  nablAccreditationValid: boolean;
  verificationTimestamp: string;
  onChainVerificationBlock: number;
}
```

### Physical Delivery Redemption & Fee Schemas
```typescript
export type DeliveryMethod = 'ARMORED_DOOR_DELIVERY' | 'VAULT_IN_PERSON_PICKUP';

export type DeliveryStatus =
  | 'DRAFT'
  | 'ESCROW_PENDING'
  | 'ESCROW_LOCKED'
  | 'MULTISIG_PENDING'
  | 'MULTISIG_APPROVED'
  | 'BARS_ALLOCATED'
  | 'ASSAY_VERIFIED'
  | 'DISPATCHED'
  | 'IN_TRANSIT'
  | 'OUT_FOR_DELIVERY'
  | 'DELIVERED'
  | 'CANCELLED'
  | 'FAILED';

export interface DeliveryAddress {
  recipientLegalName: string;
  recipientGstn?: string;
  addressLine1: string;
  addressLine2?: string;
  city: string;
  state: string;
  pincode: string;
  country: string;
  contactPersonName: string;
  contactPhone: string;
  encryptedPayload?: string; // Carrier encrypted address payload
}

export interface DeliveryRedemptionQuoteRequest {
  investorBesuAddress: `0x${string}`;
  commodityType: CommodityType;
  totalWeightGrams: number;
  selectedVaultId: string;
  deliveryMethod: DeliveryMethod;
  destinationPincode: string;
  requestedDenominations: Array<{
    denominationGrams: number;
    quantity: number;
  }>;
}

export interface DeliveryRedemptionQuoteResponse {
  quoteId: string;
  commodityType: CommodityType;
  totalWeightGrams: number;
  spotPricePerGramInr: string;
  totalBullionValueInr: string;
  platformFeePercentage: number; // 0.00% (Zero Fee)
  platformFeeInr: string;
  freightCostInr: string;
  transitInsuranceInr: string;
  subtotalFeesInr: string;
  gstRatePercentage: number; // 18.0%
  gstAmountInr: string;
  totalDeliveryCostInr: string;
  requiredTokenEscrowAmount: string;
  estimatedTransitDays: number;
  requiresMultiSigApproval: boolean; // true if Gold > 100g or Silver > 1kg
  quoteExpiresAt: string;
}

export interface CommodityDeliveryOrder {
  deliveryId: string;
  investorId: string;
  investorBesuAddress: `0x${string}`;
  commodityType: CommodityType;
  totalWeightGrams: number;
  vaultId: string;
  deliveryMethod: DeliveryMethod;
  deliveryAddress: DeliveryAddress;
  status: DeliveryStatus;
  quote: DeliveryRedemptionQuoteResponse;
  allocatedBars: AllocatedCommodityBar[];
  escrowTxHash?: `0x${string}`;
  multiSigAuthId?: string;
  trackingNumber?: string;
  carrierName?: 'BRINKS_INDIA' | 'SEQUEL_LOGISTICS';
  createdAt: string;
  updatedAt: string;
  dispatchedAt?: string;
  deliveredAt?: string;
}
```

### Multi-Sig Custody Release Schemas
```typescript
export interface InstitutionalApproverSignature {
  approverId: string;
  approverName: string;
  approverRole: string; // e.g. "Treasury Lead", "Compliance Officer", "Risk Manager"
  signerAddress: `0x${string}`;
  signature: `0x${string}`;
  signedAt: string;
  digestHash: `0x${string}`;
}

export interface MultiSigApprovalRequest {
  approvalId: string;
  deliveryId: string;
  commodityType: CommodityType;
  totalWeightGrams: number;
  bullionValueInr: string;
  thresholdRequired: number; // e.g. 2 for 2-of-3
  totalApprovers: number;
  signatures: InstitutionalApproverSignature[];
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'EXPIRED';
  expiresAt: string;
}

export interface CustodyReleaseApprovalPayload {
  deliveryId: string;
  vaultId: string;
  commodityType: CommodityType;
  totalWeightGrams: number;
  barSerialNumbers: string[];
  investorAddress: `0x${string}`;
  nonce: number;
  deadline: number;
}
```

### Armored Transit Telemetry Schemas
```typescript
export type ArmoredCarrierType = 'BRINKS_INDIA' | 'SEQUEL_LOGISTICS';

export type SecuritySensorStatus = 'NORMAL' | 'TAMPER_ALERT' | 'COMMUNICATION_LOSS';

export interface TransitTelemetryUpdate {
  deliveryId: string;
  carrier: ArmoredCarrierType;
  vehicleRegistrationNumber: string;
  currentLatitude: number;
  currentLongitude: number;
  speedKmh: number;
  headingDegrees: number;
  lockStatus: 'SECURED' | 'UNLOCKED' | 'TAMPER_DETECTED';
  cargoCompartmentTemperatureCelsius: number;
  sensorHealth: SecuritySensorStatus;
  batteryLevelPercentage: number;
  lastGpsPingTimestamp: string;
}

export interface TransitMilestone {
  milestoneId: string;
  deliveryId: string;
  milestoneName:
    | 'VAULT_PICKUP_INITIATED'
    | 'CUSTODY_HANDOVER_TO_CARRIER'
    | 'EN_ROUTE_HUB'
    | 'INTERMEDIATE_HUB_SCAN'
    | 'OUT_FOR_FINAL_DELIVERY'
    | 'DESTINATION_ARRIVAL'
    | 'OTP_BIOMETRIC_VERIFIED'
    | 'FINAL_CUSTODY_DELIVERED';
  locationName: string;
  latitude: number;
  longitude: number;
  completedAt: string;
  verifiedByActor: string;
  proofDocumentUrl?: string;
}

export interface TransitSecurityAlert {
  alertId: string;
  deliveryId: string;
  alertType: 'UNSCHEDULED_ROUTE_DEVIATION' | 'VAULT_LOCK_TAMPER' | 'EXTENDED_GPS_SIGNAL_LOSS';
  severity: 'WARNING' | 'CRITICAL';
  occurredAt: string;
  description: string;
  resolved: boolean;
  resolvedAt?: string;
}
```

## Security & Compliance Notes
- **Multi-Signature Custody Governance:** In strict adherence to institutional treasury controls and SEBI risk standards, any physical delivery redemption exceeding 100 grams of Gold or 1 kilogram of Silver mandates on-chain M-of-N multi-signature approval from verified institutional corporate officers before vault custodian release orders are issued.
- **WDRA & e-NWR Depository Alignment:** Physical commodity releases must reconcile against electronic Negotiable Warehouse Receipts (e-NWR) registered with CCRL or NERL. No digital commodity token may be burned without verified electronic depository de-registration.
- **NABL Cryptographic Integrity Checks:** Every physical bar allocated for redemption is paired with an official NABL assay certificate. The web interface independently calculates the SHA-256 hash of the certificate PDF and verifies mathematical equality with the immutable hash registered in `PhysicalVaultRegistry.sol` before enabling release approvals.
- **Armored Transit Security Standards:** High-value bullion logistics are restricted to accredited security couriers (Brink's India, Sequel Logistics) operating GPS-tracked armored vehicles equipped with dual-custody electronic vault locks, panic telemetry, and route deviation alerting.
- **Zero PII on Ledger:** Delivery recipient street addresses, contact phone numbers, and driver identity documents are encrypted using asymmetric recipient-carrier key exchanges. Only non-identifiable delivery hashes and status milestones are published to the Hyperledger Besu consortium ledger.
- **Idempotency & Replay Protection:** Every delivery quote, token escrow lock, multi-sig approval, and receipt confirmation uses a unique client-generated UUIDv4 idempotency key and EIP-712 nonce to prevent duplicate execution during network interruptions.

## Acceptance Criteria
- [ ] `VaultInventoryMap` renders interactive Mapbox GL map pins for all WDRA-accredited vaults across India with live stock breakdowns.
- [ ] Depository selector clearly differentiates CCRL and NERL facilities with accreditation validity and security ratings.
- [ ] `CommodityRedemptionWizard` guides users through denomination selection, depository picking, and destination address input.
- [ ] `FeeAndFreightCalculator` accurately computes the 0.00% (Zero Fee) platform redemption fee, insured freight estimate, transit insurance, and GST.
- [ ] `MultiSigCustodyReleaseModal` automatically activates for orders exceeding 100g Gold or 1kg Silver, capturing M-of-N EIP-712 signatures.
- [ ] `AssayCertificateVerifier` displays NABL assay certificates and computes client-side SHA-256 hashes against on-chain contract records.
- [ ] Cryptographic verification badge displays green on matching assay hash and flags red on tampering or mismatch.
- [ ] `ArmoredTransitTelemetryDashboard` visualizes real-time vehicle GPS coordinates, milestone progression, and lock security status.
- [ ] Handover verification flow supports OTP validation and dual-signature delivery confirmation.
- [ ] Playwright end-to-end tests verify vault exploration, redemption quote creation, multi-sig signing, and telemetry tracking.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 203 (Double-Entry Wallet Account Ledger), Prompt 243 (MCX Commodity & Warehouse Receipt Adapter), Prompt 330 (MCX Commodity Token & Vault Registry), Prompt 601 (Next.js Investor Web App Scaffolding).
- **Parallel Tasks:** Prompt 603 (Web Trading Dashboard), Prompt 716 (MCX/WDRA Physical Vault Audit & Inspector Portal), Prompt 703 (Sanctions Screening & AML Service).
- **Downstream Blockers:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch).
