# 615 - Next.js 14 RWA Asset Issuer & Tokenization Originator Portal

## Purpose
In modern institutional real-world asset (RWA) tokenization ecosystems operating under the regulatory supervision of SEBI (Securities and Exchange Board of India), MCA (Ministry of Corporate Affairs), and IFSCA (International Financial Services Centres Authority), primary asset origination demands an institutional-grade, self-service portal. Real-world assets such as corporate debt (Non-Convertible Debentures / NCDs), Real Estate Investment Trust (REIT) units, fractionalized commercial real estate, and agricultural or industrial commodity warehouse receipts require rigorous legal, custodial, and technical onboarding before on-chain representation is permissible.

To prevent unbacked synthetic token generation, fraudulent offerings, or regulatory non-compliance, asset originators must undergo corporate Know Your Business (KYB) checks, deposit legally binding trust deeds, secure depository credit confirmations from national depositories (CDSL, NSDL) or commodity repositories (WDRA CCRL/NERL), execute lien markings in favor of SEBI-registered Debenture Trustees, and instantiate strictly permissioned ERC-3643 security tokens on Hyperledger Besu.

This prompt specifies the architecture, user experience, data models, and technical implementation of the **Next.js 14 RWA Asset Issuer & Tokenization Originator Portal** (`apps/growww_web/issuer`). The portal empowers corporate debt issuers, REIT sponsors, commercial property syndicators, and commodity aggregators to execute complete asset onboarding lifecycles: from corporate verification and document ingestion via secure pre-signed URLs to depository escrow verification, on-chain token factory deployment, primary subscription book-building, and post-issuance corporate action servicing.

## What You Are Building
An enterprise-grade, high-availability Next.js 14 institutional origination portal featuring:
- `IssuerOnboardingKYBWizard`: Corporate entity verification portal integrating Ministry of Corporate Affairs (MCA-21) data validation, Corporate Identification Number (CIN), Permanent Account Number (PAN), Goods and Services Tax Identification Number (GSTIN), Legal Entity Identifier (LEI), Director Identification Numbers (DIN), Ultimate Beneficial Ownership (UBO) declarations, and authorized signatory board resolutions.
- `AssetOriginationSubmissionForm`: Multi-step asset origination wizard supporting three primary asset archetypes:
  1. *Corporate Debt & Commercial Paper:* Senior secured NCDs, subordinated bonds, and green debt issuances.
  2. *Real Estate & REIT Tranches:* Commercial office floors, warehousing parks, and rental yield contracts.
  3. *Commodities:* Electronic Negotiable Warehouse Receipts (eNWR) issued under WDRA-accredited repositories (CCRL / NERL) for standardized metals and agricultural commodities.
- `LegalDocumentVaultAndDematLienMarking`: High-throughput document ingestion engine leveraging direct-to-object-storage pre-signed S3/MinIO URLs for uploading Information Memorandums (IM / PAS-4), Debenture Trust Deeds (DTD), Credit Rating Rationales (CRISIL, ICRA, CARE), Title Search Reports, Valuation Certificates, and Board Resolutions, with client-side SHA-256 cryptographic fingerprinting.
- `DepositoryEscrowVerificationCard`: Real-time custodial status engine interfacing with the Depository Integration Service (Prompt 213) to verify ISIN allocation, demat credit freeze, and electronic lien marking to the designated SEBI-registered Debenture Trustee or Escrow Agent prior to token deployment.
- `TokenFactoryDeploymentWizard`: Controlled Web3 wizard interacting with `TokenFactory.sol` on Hyperledger Besu via Viem, deploying an upgradeable ERC-3643 compliant `DigitalSecurityToken` instance, linking it to the platform's `IdentityRegistry.sol`, configuring claim requirements, and establishing fractional partition parameters (18-decimal precision).
- `ComplianceRuleModuleConfigurator`: Administrative interface enabling issuers and compliance officers to enforce statutory transfer restrictions directly on-chain, including private placement investor caps (maximum 200 investors under Section 42 of Companies Act 2013), minimum investment lot sizes, and geographical / investor category whitelists.
- `PrimaryIssuanceTrancheManager`: Primary market book-building and tranche management cockpit, allowing issuers to define subscription windows, pricing bands, payment modes (RTGS, e-INR CBDC, Virtual Account Escrow), over-subscription handling, and pro-rata allotment rules.
- `SubscriptionMonitoringDashboard`: Real-time order book and cap-table dashboard providing live visibility into primary market subscription velocity, institutional investor allocations, fund clearing statuses, and allotment execution.
- `CorporateActionAndCouponDistributionTerminal`: Post-issuance lifecycle hub enabling issuers to configure coupon payment schedules, amortizing principal repayments, record dates, TDS withholding rates, and mandatory regulatory disclosure uploads.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router application structure located under `apps/growww_web/issuer` (and route group `apps/growww_web/app/(issuer)/`).
  - Multi-tenant corporate authentication, session management, and role-based access control (Issuer Admin, Issuer Maker, Issuer Checker, Debenture Trustee Signatory, Auditor).
  - Multi-step asset submission wizards for Corporate Debt, Real Estate / REITs, and Commodities (eNWR).
  - Secure chunked document uploads via AWS S3 / MinIO pre-signed URLs with client-side SHA-256 hashing.
  - Interactive integration with Custodian / Depository Integration Service (Prompt 213) to poll and display demat lock receipts and ISIN allocation status.
  - Web3 transaction dispatching and monitoring for ERC-3643 `TokenFactory` contract deployments and initial supply minting on Hyperledger Besu.
  - Real-time primary issuance subscription tracking via WebSocket and Server-Sent Events (SSE).
  - Cap-table visualization and investor categorization breakdown (Institutional, HNI, Corporate).
  - Regulatory report and filing generator producing standardized SEBI ICDR, SEBI NCS, and MCA Form CHG-9 filings.
- **Out of Scope / Handled Elsewhere:**
  - Secondary market order matching, limit order books, and continuous trading (handled by Prompt 204 and Prompt 205).
  - Retail investor onboarding and trading terminal UI (handled by Prompt 601, Prompt 602, and Prompt 603).
  - Core blockchain smart contract authoring, testing, and EVM bytecode deployment (handled by Prompt 303, Prompt 305, and Prompt 307).
  - Physical inspection and warehouse assaying of commodities (handled by Prompt 235 and Prompt 611).
  - Bank clearing integration and RTGS payment gateway processing (handled by Prompt 212 and Prompt 213).
  - Real-time pre-trade risk and margin calculations (handled by Prompt 206 and Prompt 229).

## Technology to Use
- **Framework:** Next.js 14 App Router with React Server Components (RSC) for optimized initial payload rendering, Streaming SSR with Suspense, and Server Actions for secure mutations.
- **Language:** TypeScript 5.x with strict type checking, strict null checks, and zero `any` declarations.
- **Styling & UI Kit:** Tailwind CSS 3.4+ combined with shadcn/ui components (Radix UI primitives), optimized for high-density corporate workstations, dark/light theme switching, and WCAG 2.1 AA accessibility compliance.
- **Object Storage & Pre-Signed Uploads:** AWS SDK v3 (`@aws-sdk/s3-request-presigner`, `@aws-sdk/client-s3`) and MinIO SDK for generating secure pre-signed PUT URLs, facilitating direct browser-to-bucket multi-part document ingestion.
- **Relational Persistence:** PostgreSQL 16 managed via Prisma ORM / Drizzle ORM for relational persistence of issuer corporate profiles, draft origination records, tranche configurations, and legal document metadata.
- **Cache & Real-Time Transport:** Redis 7.x for draft state caching, session store, and WebSocket / SSE pub/sub bridging.
- **Blockchain Client:** Viem 2.x and Wagmi for interaction with Hyperledger Besu JSON-RPC endpoints, smart contract reads/writes, EIP-712 structured message signing, and ABI decoding.
- **Data Tables & Virtualization:** TanStack Table v8 (`@tanstack/react-table`) for high-performance virtualized rendering of large primary investor allotment registries and cap tables.
- **Form Management & Validation:** React Hook Form coupled with Zod schemas for multi-step corporate origination form validation.
- **Data Visualization:** Recharts and Tremor for primary issuance funding progress, cap-table composition donuts, and coupon amortization curves.
- **Precision Financial Math:** Decimal.js for exact financial math, coupon calculations, tax deductions, and lot-size fractionalization without IEEE 754 floating-point inaccuracies.

## Backend / Infra Touchpoints
- **Token Issuance Service (Prompt 303):** Communicates via gRPC and authenticated REST endpoints (`POST /api/v1/issuance/tokens`) to trigger smart contract factory deployments, register token parameters, and monitor on-chain transaction execution receipts.
- **Custodian & Depository Integration Service (Prompt 213):** Connects via REST (`/api/v1/depository/escrow/verify`) and Webhooks to verify ISIN creation, CDSL/NSDL demat credit records, CCRL/NERL warehouse receipt escrow locks, and lien marking to the Debenture Trustee.
- **KYC & AML Service (Prompt 202):** Interfaces with corporate KYB endpoints (`/api/v1/kyc/corporate/verify`) to validate CIN, PAN, GSTIN, LEI, and CKYC records for directors and authorized signatories.
- **Audit Log Service (Prompt 218):** Dispatches immutable, hash-chained audit events for every draft creation, document upload, maker-checker authorization, token deployment, and corporate action publication.
- **Notification Service (Prompt 211):** Emits automated event triggers (email, SMS, WhatsApp, and Webhooks) notifying trustees, rating agencies, compliance officers, and issuers of workflow state changes.
- **API Gateway & BFF (Prompt 219):** Terminates external TLS, performs JWT bearer token validation, enforces mTLS for institutional connectors, and applies rate-limiting policies.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu (Enterprise Permissioned Consortium Ledger, QBFT Consensus, 2-second deterministic block intervals, zero public gas volatility).
- **Interfaced Smart Contracts:**
  - `TokenFactory.sol` (Prompt 303): Factory contract orchestrating the deployment of ERC-3643 `DigitalSecurityToken` proxies. The portal constructs initialization arguments and dispatches:
    ```solidity
    function createToken(
        string memory name,
        string memory symbol,
        uint8 decimals,
        address identityRegistry,
        address compliance,
        bytes32 isinHash,
        bytes32 dematLockReceiptHash
    ) external returns (address tokenAddress);
    ```
  - `IdentityRegistry.sol` (Prompt 303, Prompt 305): Validates that subscribers hold valid, unexpired verified identity claims before primary issuance allotment. The portal queries:
    ```solidity
    function isVerified(address userAddress) external view returns (bool);
    function identity(address userAddress) external view returns (address);
    ```
  - `Compliance.sol` (Prompt 305): Modular compliance contract enforcing statutory limits. The portal configures and audits:
    ```solidity
    function setMaxInvestors(uint256 maxInvestors) external;
    function setCountryRestriction(uint16 countryCode, bool isAllowed) external;
    function canTransfer(address from, address to, uint256 value) external view returns (bool);
    ```
  - `DigitalSecurityToken.sol` (Prompt 303): Core token contract. The portal triggers primary minting to authorized escrow and allotment contracts upon depository lock confirmation:
    ```solidity
    function mintWithCustodyProof(
        address to,
        uint256 amount,
        bytes32 depositoryBatchId,
        bytes32 dematTxRefHash
    ) external;
    ```
- **Cryptographic Custody Linking & Zero PII:**
  - Every token issuance is cryptographically bound to an underlying demat lock slip or warehouse receipt by hashing the depository credit receipt (`dematTxRefHash = SHA256(ISIN + DepositoryRef + Units + Timestamp)`).
  - Zero Personally Identifiable Information (PII) is published on-chain. Issuer and investor identities are stored as pseudonymous `bytes32 identityId` references resolved strictly off-chain by the KYC service.

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Issuer Route Architecture:**
   - Initialize the directory structure at `apps/growww_web/issuer` and route group `apps/growww_web/app/(issuer)/`.
   - Create root layouts with navigation sidebars, session breadcrumbs, environment status badges, and institutional context providers:
     - `app/(issuer)/dashboard/page.tsx`: Executive dashboard displaying active issuances, pipeline status, and primary subscription tallies.
     - `app/(issuer)/originations/new/page.tsx`: Step-by-step asset onboarding wizard.
     - `app/(issuer)/originations/[id]/page.tsx`: Detailed asset workspace, document repository, and verification audit.
     - `app/(issuer)/originations/[id]/token-deployment/page.tsx`: On-chain ERC-3643 deployment terminal.
     - `app/(issuer)/originations/[id]/subscriptions/page.tsx`: Real-time order book and primary allotment tracker.
     - `app/(issuer)/originations/[id]/corporate-actions/page.tsx`: Coupon distribution, amortization, and compliance disclosure portal.
     - `app/(issuer)/kyb/page.tsx`: Corporate profile, UBO declarations, and authorized signatory manager.

2. **Establish Data Models & Database Migrations:**
   - Define PostgreSQL tables using Prisma in `packages/database/prisma/schema.prisma`:
     - `IssuerEntity`: Stores legal name, CIN, LEI, PAN, GSTIN, registered office, incorporation date, and KYB status.
     - `AuthorizedSignatory`: Stores signatory name, DIN/PAN, designation, digital signature certificate (DSC) fingerprint, and board resolution reference.
     - `AssetOrigination`: Core table storing asset category, title, currency, target amount, minimum subscription, tenor, coupon structure, credit rating, and status (`DRAFT`, `KYB_PENDING`, `DOCUMENTS_UPLOADED`, `ESCROW_VERIFIED`, `TOKEN_DEPLOYED`, `SUBSCRIPTION_OPEN`, `ALLOTTED`, `MATURED`).
     - `OriginationDocument`: Tracks document type, S3 object key, client SHA-256 hash, MIME type, file size, upload timestamp, and verification status.
     - `DematEscrowReceipt`: Stores ISIN, depository code (CDSL/NSDL/CCRL), beneficiary DP ID, client ID, locked quantity, debenture trustee pledge reference, and depository API confirmation timestamp.
     - `TokenDeploymentRecord`: Stores Besu contract address, factory transaction hash, block number, IdentityRegistry address, Compliance contract address, and initial minted supply.
     - `SubscriptionOrder`: Records investor ID, allocated units, gross amount, payment method, virtual account reference, KYC status, and settlement status.

3. **Implement S3 / MinIO Direct Pre-Signed Upload Pipeline:**
   - Create secure Server Action `generatePresignedUploadUrl(filename, fileType, checksumSha256, documentCategory, originationId)`:
     - Enforces strict MIME type whitelisting (`application/pdf`, `application/xml`, `image/tiff`).
     - Limits maximum file size (50MB for general documents, 250MB for high-resolution property valuation / deed scans).
     - Issues pre-signed `PUT` URLs containing metadata tags and expiration windows (maximum 15 minutes).
   - Build client-side chunked uploader component `DocumentDropzone.tsx`:
     - Computes browser-side SHA-256 hash using the Web Crypto API (`crypto.subtle.digest`) prior to transmission.
     - Performs direct multipart upload to S3/MinIO with progress tracking.
     - On completion, calls server confirmation endpoint `/api/v1/issuer/documents/confirm` which validates the uploaded object hash against the client-declared SHA-256 checksum.

4. **Build `IssuerOnboardingKYBWizard` Component:**
   - Implement multi-step onboarding wizard for legal entities:
     - *Step 1 (Entity Verification):* Ingest CIN and trigger MCA-21 verification API, auto-populating corporate name, incorporation date, registered address, authorized capital, and active director list.
     - *Step 2 (Tax & Financial Identification):* Ingest and verify company PAN, GSTIN, and 20-digit Legal Entity Identifier (LEI).
     - *Step 3 (Beneficial Ownership & Directors):* Capture Ultimate Beneficial Owners (individuals holding >= 10% voting rights or shares) in compliance with Prevention of Money Laundering (PMLA) rules.
     - *Step 4 (Board Resolution & Signatories):* Ingest Board Resolution under Section 179/180 of Companies Act 2013 authorizing borrowing and asset tokenization; link authorized signatories with digital signature verification.

5. **Develop `AssetOriginationSubmissionForm` Component:**
   - Implement dynamic, asset-specific origination forms powered by React Hook Form and Zod:
     - *Corporate Debt / NCD Flow:* Face value per debenture (e.g. ₹1,00,000 or ₹10,000 for retail bond windows), coupon rate type (Fixed, Floating, Zero Coupon), coupon frequency (Monthly, Quarterly, Annual), day-count convention (Actual/Actual, 30/360), issue size, green-shoe option limit, credit rating details, and appointment of Debenture Trustee.
     - *Real Estate / REIT Flow:* Underlying commercial asset address, land registry parcel number, gross leasable area (GLA), occupancy percentage, anchor tenants, weighted average lease expiry (WALE), net operating income (NOI), projected yield, and title search certificate.
     - *Commodity eNWR Flow:* Electronic warehouse receipt number, repository (CCRL / NERL), warehouse location, WDRA accreditation code, commodity grade, net weight / volume, moisture content, assaying certificate, and crop season.

6. **Build `LegalDocumentVaultAndDematLienMarking` Component:**
   - Construct institutional document vault UI categorizing uploads by regulatory mandate:
     - Statutory Offer Documents (SEBI Private Placement Offer Letter PAS-4, Information Memorandum).
     - Custodial & Legal Deeds (Debenture Trust Deed Form SH-12, Escrow Agreement, Mortgage Deed / Hypothecation Deed).
     - Expert Reports (Credit Rating Agency Rationale, Independent Chartered Accountant Net-Worth Certificate, Legal Title Search Report, Assayer Certificate).
   - Provide inline PDF previewer with cryptographic SHA-256 fingerprint verification badge and timestamped audit watermark.

7. **Implement `DepositoryEscrowVerificationCard` Component:**
   - Construct real-time depository integration card:
     - Interfaces with Custodian / Depository Integration Service (Prompt 213) via REST API `GET /api/v1/depository/escrow/status/{originationId}`.
     - Displays verification progress milestones:
       1. Demat Account Setup & ISIN Creation Confirmation.
       2. Corporate Action Credit into Issuer Demat Account.
       3. Execution of Lien / Pledge in favor of Debenture Trustee (CDSL Form 14 / NSDL equivalent).
       4. Depository Free-Balance Debit and Escrow Freeze Confirmation.
     - Displays cryptographic hash of the depository credit slip and renders an immutable verification seal when all custodial conditions are satisfied.

8. **Build `TokenFactoryDeploymentWizard` Component:**
   - Implement Web3 deployment cockpit leveraging Viem and Besu RPC:
     - Form to configure ERC-3643 token parameters: Token Name (e.g., "Growww L&T NCD 2029 Series I"), Symbol (e.g., "GW-LT-NCD1"), Decimals (standard 18), and Initial Supply.
     - Automatically binds system `IdentityRegistry` and `Compliance` contract addresses.
     - Inserts the immutable `isinHash` and `dematLockReceiptHash` into the constructor parameters.
     - Executes transaction via Viem, displays live block confirmation ticker, extracts the deployed proxy address from the `TokenCreated` event log, and commits the address to PostgreSQL.

9. **Construct `ComplianceRuleModuleConfigurator` Component:**
   - Interface for setting on-chain compliance parameters via the deployed `Compliance.sol` contract:
     - *Investor Limit:* Configure private placement cap (maximum 200 investors under Companies Act 2013 Section 42).
     - *Minimum Lot Size:* Enforce minimum ticket size (e.g., ₹1,00,000 for standard private placement or ₹10,000 for electronic book-building platforms).
     - *Jurisdiction Rules:* Specify allowable ISO 3166 country codes (e.g., India 356, GIFT City IFSC designated foreign jurisdictions).
     - *Investor Categorization:* Restrict token holdings strictly to KYC-verified identities bearing Accredited Investor or QIB claims from the `ClaimTopicsRegistry`.

10. **Implement `PrimaryIssuanceTrancheManager` Component:**
    - Develop issuance lifecycle management interface:
      - Configure subscription opening date, closing date, and settlement cut-off times.
      - Establish tranche types: Firm Allotment (Institutional Anchors), Book Building, or Fixed Price Offer.
      - Configure Escrow Bank Accounts (virtual accounts generated per applicant via ICICI / Axis Bank host-to-host banking or RBI e-Kuber e-INR smart contract escrow).
      - Set green-shoe over-subscription retention percentages and automated refund triggers if the 75% minimum subscription threshold (under SEBI regulations) is not achieved.

11. **Build `SubscriptionMonitoringDashboard` Component:**
    - Real-time institutional order book monitor streaming via WebSocket:
      - Total committed capital vs target issue size with percentage progress bar.
      - Live subscription velocity ticker (orders per hour, average ticket size).
      - Virtualized TanStack table of all incoming bids showing Investor UCC, Category (QIB, Corporate, HNI), Bid Units, Bid Price, Payment Confirmation Status, and Allocation Status.
      - Allotment execution button enabling Maker-Checker dual confirmation to finalize allotments, trigger token distribution on Besu, and release escrowed funds to the issuer's operating account.

12. **Build `CorporateActionAndCouponDistributionTerminal` Component:**
    - Lifecycle servicing interface for active tokens:
      - Automated coupon calculation engine: Computes upcoming coupon payout based on issue face value, coupon rate, day-count convention, and active cap-table snapshot.
      - Record Date Announcement: Schedules automated cap-table snapshots on Besu at 23:59:59 IST on the record date.
      - Tax Deduction at Source (TDS) Manager: Calculates Section 193 TDS withholding based on investor residency and PAN verification.
      - Payout Execution Gateway: Coordinates with Payment Gateway (Prompt 212) to dispatch net coupon payments via RTGS/NEFT/e-INR, and updates the on-chain corporate action registry.

13. **Construct Regulatory Filing Export & Audit Trail:**
    - Build automated PDF / XML filing generator compiling:
      - SEBI Form A / Form B due diligence submission packages.
      - MCA Form CHG-9 (Creation of Charge for Debentures) with auto-filled asset mortgage schedules.
      - Private Placement Offer Letter (PAS-4) with embedded SHA-256 document checksums.
    - Provide complete, tamper-evident audit trail viewer pulling from Audit Log Service (Prompt 218).

14. **End-to-End Testing, Mocking, and Verification:**
    - Write unit tests using Vitest covering coupon calculation math, Zod origination form validations, and S3 checksum verification logic.
    - Implement Playwright E2E integration test suites validating:
      - Full KYB onboarding flow with mock MCA-21 API responses.
      - Multi-step asset submission and document upload using mock S3 pre-signed endpoints.
      - Depository escrow verification card state transitions.
      - Simulated Besu token deployment and event extraction using mock Viem client.
      - Primary subscription allotment execution and cap-table update.

## Interfaces / Contracts

### Protobuf Definition: Asset Origination & Tokenization
```protobuf
syntax = "proto3";

package growww.issuer.v1;

import "google/protobuf/timestamp.proto";

enum AssetCategory {
  ASSET_CATEGORY_UNSPECIFIED = 0;
  ASSET_CATEGORY_CORPORATE_BOND_NCD = 1;
  ASSET_CATEGORY_COMMERCIAL_PAPER = 2;
  ASSET_CATEGORY_REIT_UNIT = 3;
  ASSET_CATEGORY_REAL_ESTATE_FRACTIONAL = 4;
  ASSET_CATEGORY_COMMODITY_ENWR = 5;
}

enum OriginationStatus {
  ORIGINATION_STATUS_UNSPECIFIED = 0;
  ORIGINATION_STATUS_DRAFT = 1;
  ORIGINATION_STATUS_KYB_VERIFIED = 2;
  ORIGINATION_STATUS_DOCUMENTS_PENDING = 3;
  ORIGINATION_STATUS_DOCUMENTS_APPROVED = 4;
  ORIGINATION_STATUS_DEPOSITORY_ESCROW_LOCKED = 5;
  ORIGINATION_STATUS_TRUSTEE_SIGN_OFF = 6;
  ORIGINATION_STATUS_TOKEN_DEPLOYED = 7;
  ORIGINATION_STATUS_SUBSCRIPTION_OPEN = 8;
  ORIGINATION_STATUS_SUBSCRIPTION_CLOSED = 9;
  ORIGINATION_STATUS_ALLOTTED = 10;
  ORIGINATION_STATUS_MATURED_OR_REDEEMED = 11;
}

enum DepositoryType {
  DEPOSITORY_TYPE_UNSPECIFIED = 0;
  DEPOSITORY_TYPE_CDSL = 1;
  DEPOSITORY_TYPE_NSDL = 2;
  DEPOSITORY_TYPE_CCRL_WDRA = 3;
  DEPOSITORY_TYPE_NERL_WDRA = 4;
}

message AssetOriginationRequest {
  string origination_id = 1;
  string issuer_entity_id = 2;
  AssetCategory asset_category = 3;
  string asset_title = 4;
  string isin = 5;
  
  double total_face_value_inr = 6;
  double price_per_token_inr = 7;
  uint64 total_token_supply = 8;
  uint32 minimum_lot_size = 9;
  
  // Debt Specific Attributes
  double annual_coupon_rate_bps = 10; // e.g. 950 for 9.50%
  string coupon_frequency = 11;       // "MONTHLY", "QUARTERLY", "ANNUAL", "CUMULATIVE"
  string day_count_convention = 12;   // "ACTUAL_ACTUAL", "THIRTY_THREE_SIXTY"
  google.protobuf.Timestamp maturity_date = 13;
  string debenture_trustee_name = 14;
  string credit_rating = 15;          // e.g. "CRISIL AA+ Stable"
  
  // Real Estate Specific Attributes
  string property_address = 16;
  double gross_leasable_area_sqft = 17;
  double projected_rental_yield_bps = 18;
  
  // Commodity Specific Attributes
  string enwr_receipt_number = 19;
  string warehouse_accreditation_id = 20;
  double commodity_quantity_metric_tons = 21;
}

message DematEscrowVerificationResponse {
  string origination_id = 1;
  string isin = 2;
  DepositoryType depository = 3;
  string issuer_demat_account = 4;
  string escrow_dp_id = 5;
  uint64 locked_units = 6;
  string lien_reference_number = 7;
  bool is_verified = 8;
  string demat_tx_ref_hash = 9;
  google.protobuf.Timestamp locked_at = 10;
}

message TokenDeploymentSubmission {
  string origination_id = 1;
  string token_name = 2;
  string token_symbol = 3;
  uint32 decimals = 4; // Always 18
  string identity_registry_address = 5;
  string compliance_module_address = 6;
  string demat_tx_ref_hash = 7;
  uint64 max_investor_limit = 8; // e.g. 200 for Private Placement
}

message TokenDeploymentReceipt {
  string origination_id = 1;
  string contract_address = 2;
  string deployment_tx_hash = 3;
  uint64 block_number = 4;
  google.protobuf.Timestamp deployed_at = 5;
  bool is_verified_on_chain = 6;
}
```

### TypeScript Domain Interfaces
```typescript
export type AssetCategory = 
  | 'CORPORATE_BOND_NCD' 
  | 'COMMERCIAL_PAPER' 
  | 'REIT_UNIT' 
  | 'REAL_ESTATE_FRACTIONAL' 
  | 'COMMODITY_ENWR';

export type OriginationStatus = 
  | 'DRAFT' 
  | 'KYB_VERIFIED' 
  | 'DOCUMENTS_PENDING' 
  | 'DOCUMENTS_APPROVED' 
  | 'DEPOSITORY_ESCROW_LOCKED' 
  | 'TRUSTEE_SIGN_OFF' 
  | 'TOKEN_DEPLOYED' 
  | 'SUBSCRIPTION_OPEN' 
  | 'SUBSCRIPTION_CLOSED' 
  | 'ALLOTTED' 
  | 'MATURED_OR_REDEEMED';

export interface KYBEntityProfile {
  issuerId: string;
  legalEntityName: string;
  cin: string;
  pan: string;
  gstin: string;
  lei: string;
  registeredAddress: string;
  incorporationDate: string;
  mcaStatus: 'ACTIVE' | 'STRUCK_OFF' | 'UNDER_LIQUIDATION';
  authorizedSignatories: {
    signatoryId: string;
    fullName: string;
    dinOrPan: string;
    designation: string;
    isAuthorizedByBoard: boolean;
  }[];
  isKybApproved: boolean;
}

export interface LegalDocumentMetadata {
  documentId: string;
  originationId: string;
  documentCategory: 
    | 'INFORMATION_MEMORANDUM_PAS4'
    | 'DEBENTURE_TRUST_DEED'
    | 'CREDIT_RATING_RATIONALE'
    | 'TITLE_SEARCH_REPORT'
    | 'VALUATION_CERTIFICATE'
    | 'BOARD_RESOLUTION'
    | 'ENWR_RECEIPT';
  fileName: string;
  s3ObjectKey: string;
  sha256Checksum: string;
  fileSizeBytes: number;
  mimeType: string;
  uploadedAt: string;
  verifiedByTrustee: boolean;
  trusteeSignatureHash?: string;
}

export interface DematLockReceipt {
  originationId: string;
  isin: string;
  depository: 'CDSL' | 'NSDL' | 'CCRL_WDRA' | 'NERL_WDRA';
  depositoryReference: string;
  lockedQuantity: string;
  dematAccountBeneficiary: string;
  lienMarkedTo: string; // e.g. "Catalyst Trusteeship Ltd" or "IDBI Trusteeship Services"
  dematTxRefHash: `0x${string}`;
  isLocked: boolean;
  verifiedTimestamp: string;
}

export interface ERC3643DeploymentConfig {
  originationId: string;
  tokenName: string;
  tokenSymbol: string;
  decimals: 18;
  targetSupply: string;
  identityRegistryAddress: `0x${string}`;
  complianceAddress: `0x${string}`;
  maxInvestorsLimit: number;
  allowedCountryCodes: number[]; // e.g. [356] for India
  isinHash: `0x${string}`;
  dematTxRefHash: `0x${string}`;
}

export interface PrimarySubscriptionItem {
  orderId: string;
  investorUcc: string;
  investorCategory: 'QIB' | 'CORPORATE_TREASURY' | 'HNI' | 'RETAIL';
  requestedUnits: string;
  bidPricePerUnit: string;
  totalCommitmentInr: string;
  paymentMode: 'RTGS' | 'VIRTUAL_ACCOUNT' | 'E_INR_ESCROW';
  paymentStatus: 'AWAITING_FUNDS' | 'CLEARED' | 'ESCROW_LOCKED' | 'REFUNDED';
  identityVerified: boolean;
  allottedUnits?: string;
  settlementTxHash?: `0x${string}`;
  timestamp: string;
}
```

### REST & WebSocket API Contracts
- `POST /api/v1/issuer/kyb/verify`:
  - Validates corporate entity CIN, PAN, and GSTIN via MCA-21 and tax gateways.
- `POST /api/v1/issuer/documents/presigned-url`:
  - Request: `{ filename: string, mimeType: string, fileSizeBytes: number, sha256Checksum: string, documentCategory: string, originationId: string }`
  - Response: `{ uploadUrl: string, objectKey: string, expiresAt: string }`
- `POST /api/v1/issuer/documents/confirm`:
  - Validates S3 upload completion, verifies server-side SHA-256 against client checksum, and persists metadata.
- `POST /api/v1/issuer/originations`:
  - Creates a new draft asset origination with category-specific parameters.
- `GET /api/v1/issuer/originations/{id}`:
  - Retrieves asset details, document vault records, depository status, and token address.
- `POST /api/v1/issuer/originations/{id}/depository-escrow/verify`:
  - Triggers verification query to Depository Integration Service (Prompt 213) to confirm ISIN demat freeze.
- `POST /api/v1/issuer/originations/{id}/deploy-token`:
  - Dispatches ERC-3643 Token deployment via Viem to Besu RPC node; returns transaction hash and deployed proxy address.
- `GET /api/v1/issuer/originations/{id}/subscriptions`:
  - Returns paginated list of primary investor subscriptions, payment statuses, and cap-table distribution.
- `POST /api/v1/issuer/originations/{id}/allotment/execute`:
  - Maker-Checker endpoint to execute primary tranche allotment and trigger on-chain minting.
- `POST /api/v1/issuer/originations/{id}/corporate-actions/coupon`:
  - Schedules upcoming coupon payment, announces record date, and calculates TDS withholding.
- `WSS wss://api.growww.in/ws/v1/issuer/originations/{id}/stream`:
  - WebSocket subscription streaming live events:
    - `SUB origination:depository-status:{id}`
    - `SUB origination:deployment-progress:{id}`
    - `SUB origination:subscription-feed:{id}`

## Security & Compliance Notes
- **Companies Act 2013 Compliance:**
  - *Section 42 (Private Placement):* The platform strictly enforces a ceiling of no more than 200 investors in an aggregate financial year per security class. The on-chain `Compliance.sol` module programmatically blocks any minting or transfer exceeding the 200-investor threshold.
  - *Section 71 (Debenture Issuance & Trusteeship):* For debt issuances, appointment of a SEBI-registered Debenture Trustee is mandatory before issue opening. Trust deed Form SH-12 must be executed and uploaded, and Debenture Redemption Reserve (DRR) accounts tracked.
  - *Form CHG-9 Filing:* Within 30 days of charge creation on company assets, the platform generates the mandatory RoC Form CHG-9 XML/PDF packet to ensure legal enforceability of security interests.
- **SEBI ICDR & SEBI NCS Regulations:**
  - Mandatory credit rating from at least one SEBI-registered Credit Rating Agency (CRISIL, ICRA, CARE, India Ratings) for public and private debt placements.
  - Strict minimum application size enforcement (e.g. ₹1 Lakh for debt securities or ₹10,000 for Electronic Book Provider / EBP issues).
  - Minimum subscription condition: If the issue fails to achieve the statutory 75% subscription threshold by the closing date, the platform automatically triggers an escrow refund workflow.
- **Demat Lock Invariant & Proof of Reserve:**
  - Strict 1:1 invariant between deposited demat units / warehouse receipts and on-chain minted tokens.
  - Smart contract minting is impossible without a valid, cryptographic `dematTxRefHash` verified by the custodian microservice.
  - Depository balance statements are re-reconciled every 24 hours against on-chain total supply via the Reconciliation Engine (Prompt 215).
- **Zero PII on Distributed Ledger:**
  - In compliance with the Digital Personal Data Protection (DPDP) Act 2023, investor PAN, Aadhaar, bank details, and director identity numbers are strictly excluded from the blockchain.
  - On-chain accounts represent anonymized EVM addresses linked to hashed identity claims (`bytes32 identityId`) verified exclusively through off-chain Zero-Knowledge or secure signature claims.
- **Document Integrity & Storage Security:**
  - All legal documents, prospectuses, and title deeds uploaded via S3 pre-signed URLs are encrypted at rest with AES-256-GCM.
  - SHA-256 client and server checksum matching ensures zero document tampering in transit.
  - Multi-tenancy isolation: Issuers can access only their designated originations, strictly enforced at the database row level and API gateway layer.
- **Maker-Checker Corporate Governance:**
  - Critical actions (Asset Submission, Document Approval, Token Deployment, Allotment Finalization, and Coupon Dispatch) mandate Maker-Checker dual authorization within the issuer entity.

## Acceptance Criteria
- [ ] Next.js 14 application boots successfully under `apps/growww_web/issuer` with zero TypeScript build errors.
- [ ] `IssuerOnboardingKYBWizard` verifies corporate CIN, PAN, and GSTIN via mock MCA-21 and tax services, correctly extracting director lists.
- [ ] Direct-to-S3/MinIO chunked document uploader computes client-side SHA-256 checksums and successfully uploads large PDF files using pre-signed PUT URLs.
- [ ] Server verification endpoint rejects uploaded documents whose server-computed SHA-256 hash does not match the client-declared checksum.
- [ ] `AssetOriginationSubmissionForm` provides dedicated, validated input fields for Corporate Bonds (NCDs), Real Estate (REITs), and Commodities (eNWR).
- [ ] `DepositoryEscrowVerificationCard` connects to Custodian/Depository service and accurately displays ISIN lock status, lien marking reference, and `dematTxRefHash`.
- [ ] `TokenFactoryDeploymentWizard` successfully submits token creation parameters to `TokenFactory.sol` on Hyperledger Besu, correctly populating `name`, `symbol`, `decimals` (18), and custody hashes.
- [ ] On-chain `Compliance` module is configured with the mandatory private placement cap (maximum 200 investors) and verified via Viem contract call.
- [ ] Primary issuance book-building monitor streams real-time investor bids via WebSocket and updates cap-table visualization without full page reloads.
- [ ] Primary allotment execution enforces Maker-Checker dual approval before dispatching on-chain minting transactions.
- [ ] `CorporateActionAndCouponDistributionTerminal` calculates accurate coupon amounts using standard day-count conventions (Actual/Actual, 30/360) and computes statutory TDS withholding.
- [ ] Platform generates compliant filing packages for MCA Form CHG-9 and SEBI Private Placement Offer Letter (PAS-4).
- [ ] Vitest test suite covers financial math calculations (coupon rates, lot sizes, TDS) and Zod schema validations with 100% pass rate.
- [ ] Playwright E2E tests validate complete origination flow: from KYB onboarding to document upload, depository lock verification, token deployment, and allotment.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 202: KYC & Corporate AML Microservice (for corporate verification and UBO screening).
  - Prompt 213: Custodian & Depository Integration Service (for NSDL/CDSL/WDRA escrow verification).
  - Prompt 303: ERC-3643 Digital Security Token & TokenFactory Smart Contracts.
  - Prompt 305: On-Chain Compliance Engine & Transfer Rule Modules.
  - Prompt 601: Next.js Investor & Admin Web Scaffolding (monorepo layout, shared UI components).
- **Parallel Tasks:**
  - Prompt 604: Admin User KYC & KYB Review Dashboard.
  - Prompt 605: Admin Risk Exception & Multi-Party Approval Workstation.
  - Prompt 606: Admin Proof of Reserve Reconciliation Workstation.
- **Downstream Blockers:**
  - Prompt 610: Web Options Chain & Cross-Chain Primary Deposit Portal.
  - Prompt 611: Web Commodity Physical Delivery & Assay Portal.
  - Prompt 906: Comprehensive UAT Plan & Regulatory Sandbox Scenarios.
  - Prompt 907: Regulatory Sandbox Pilot Launch Plan.
