# 537 - Flutter RWA Launchpad & Primary Subscription Screen

## Purpose
Provides institutional and accredited retail investors with a compliant, high-performance primary token sale participation interface for tokenized Real-World Assets (RWAs). High-value asset classes such as commercial real estate (Grade-A office towers, logistics hubs), structured private debt (senior secured credit, invoice factoring pools), and venture-backed startup equity require sophisticated primary price discovery mechanisms rather than simplistic fixed-price purchase forms. Traditional capital formation suffers from opaque order books, allocation favouritism, and manual post-raise token distribution.

This specification details the architecture and implementation of the Flutter primary launchpad module (ADR-0044). The screen integrates interactive Dutch auction price decay curves, pro-rata subscription escrow pools, live order book commitment depth, and continuous second-by-second linear vesting claim streams. Investors navigate primary offerings, inspect certified valuation documents, pass accredited investor verification gates, place biometrically authorized bids, and monitor on-chain settlement transitions from capital escrow to streaming token distribution.

## What You Are Building
A responsive, multi-platform Flutter feature suite located in `apps/growww_flutter/lib/features/launchpad/` and exposed via `lib/screens/launchpad/`:
- `RwaLaunchpadScreen`: The primary container widget displaying active offerings, upcoming price discovery auctions, and active vesting vaults with dynamic tab navigation, search, and asset-class filters (Real Estate, Private Debt, Startup Equity).
- `OfferingCardWidget`: High-density summary card featuring issuer branding, asset category tags, funding target progress gauge, live price or current Dutch auction clearing estimate, time remaining countdown, and accreditation requirement badges.
- `DutchAuctionDetailScreen`: Deep-dive price discovery view displaying the downward price decay curve using a high-frequency `CustomPainter`, real-time clearing price calculator, aggregate subscription demand gauge, and real-time bid depth chart.
- `BiddingSheetModal`: Interactive bottom sheet enabling investors to specify bid parameters (maximum bid price, committed capital, token allocation quantity), real-time escrow balance verification, accredited investor tier validation, and biometric order authorization.
- `ProRataSubscriptionModal`: Over-subscription commitment modal for capped private debt and startup equity sales displaying real-time allocation percentage, oversubscription multiples (e.g. 3.4x), maximum capital caps, and automated refund estimates.
- `LinearVestingStreamWidget`: Real-time streaming claim component featuring an animated continuous ticker visualizing second-by-second token release, claimable balances, locked vault balances, and an instant gasless one-tap claim action.
- `AccreditedInvestorBadgeWidget`: Contextual compliance banner and lock-out guard verifying SEBI Accredited Investor or IFSCA Qualified Market Participant credential validity before enabling bidding controls.
- `OfferingDocumentVaultBottomSheet`: Secure PDF and document preview sheet displaying audited property valuations, SEBI/IFSCA regulatory filings, issuer corporate structures, and smart contract audit attestations.

## Scope Boundaries
- **In Scope:**
  - Responsive Flutter UI components and screens for RWA offerings, Dutch auction live charts, pro-rata subscription modals, and vesting streams.
  - Interactive mathematical rendering of linear and exponential Dutch auction decay curves ($P(t)$) with animated current-time cursors.
  - Continuous second-by-second linear vesting animations driven by high-resolution display tickers.
  - Riverpod state management for auction telemetry, bid parameter calculations, biometric verification, and stream updates.
  - Real-time WebSocket feed subscription for live bid depth, auction clearing price updates, and sale status transitions.
  - Biometric authentication (`local_auth`) integration for signing and submitting bids.
  - EIP-712 typed data hashing and UserOperation packaging for gasless Paymaster transaction execution.
  - SEBI / IFSCA accreditation credential evaluation and interface gating.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain Launchpad Matching & Allocation Engine (Prompt 266).
  - Smart contract implementation and EVM compilation for `DutchAuction.sol` and `LinearVestingVault.sol` (Prompt 303, Prompt 306).
  - ERC-4337 Account Abstraction Bundler and Paymaster sponsorship infrastructure (Prompt 259).
  - Fiat bank transfer, UPI settlement, and multi-currency custody ledger (Prompt 203, Prompt 212, Prompt 511).
  - Secondary order book matching engine and trading interface post-listing (Prompt 204, Prompt 205, Prompt 509).
  - Investor identity document ingestion, OCR, and DigiLocker verification (Prompt 202, Prompt 504).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ targeting mobile (Android, iOS) and desktop (macOS, Windows, Linux).
- **State Management:** `flutter_riverpod: ^2.5.1` with code-generated `AsyncNotifier`, `Notifier`, and `StreamNotifier` primitives for deterministic state mutations.
- **Visual Rendering & Curves:** Flutter Impeller / Skia engine utilizing `CustomPainter` with `Canvas.drawPath` and cubic Bézier smoothing for mathematical decay curve rendering.
- **Continuous Animations:** Flutter `AnimationController` and `Ticker` managing 60/120 FPS high-refresh linear vesting increments.
- **Real-Time Streaming:** `web_socket_channel: ^3.0.0` with JSON and binary Protocol Buffer decoding for sub-50ms auction depth ticks.
- **Cryptographic Signing & Biometrics:** `local_auth: ^2.2.0` for biometric challenges, combined with `web3dart: ^2.7.3` and Dart `crypto: ^3.0.3` for EIP-712 typed signature serialization.
- **Secure Storage:** `flutter_secure_storage: ^9.2.2` for storing session credentials and ephemeral signing nonces.
- **Document Rendering:** `flutter_pdfview: ^1.3.2` or native platform views for viewing offering prospectuses and valuation reports.

## Backend / Infra Touchpoints
- **Launchpad Engine (Prompt 266):**
  - `GET /api/v1/launchpad/offerings`: Retrieves paginated list of active, upcoming, and closed offerings with asset classification, funding metrics, and auction parameters.
  - `GET /api/v1/launchpad/offerings/{id}`: Detailed offering dossier including financial prospectus URLs, valuation models, tokenomics partitions, and contract addresses.
  - `GET /api/v1/launchpad/offerings/{id}/auction-state`: Snapshot of Dutch auction parameters (start price, reserve price, decay rate, start time, end time, current clearing price, total bids committed).
  - `POST /api/v1/launchpad/offerings/{id}/bids`: Submits signed bid intent containing investor wallet, maximum price, capital commitment, and biometric proof token.
  - `GET /api/v1/launchpad/offerings/{id}/allocations/me`: Fetches investor specific allotment status, escrow refund amount, and vesting vault contract mapping.
  - `WSS /ws/v1/launchpad/offerings/{id}/stream`: WebSocket channel streaming live auction price decay, order book bid updates, total committed capital, and clearing notifications.
- **Wallet & Ledger Service (Prompt 203):**
  - `GET /api/v1/wallet/balances`: Queries investor available funds in INR cash ledger, e-Rupee (CBDC), or permitted stablecoins (USDC) for escrow commitment.
  - `POST /api/v1/wallet/escrow/lock`: Places cryptographic hold on investor balances during active auction or subscription period.
- **Token Issuance & Contracts (Prompt 303):**
  - Provides permissioned ERC-3643 asset token addresses, identity registry verification, and compliance check endpoint.
- **Account Abstraction Bundler & Paymaster (Prompt 259):**
  - `POST /api/v1/bundler/paymaster/sponsor`: Obtains Paymaster `paymasterAndData` sponsorship signature for gasless bid submission and vesting claim execution.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Network Topology:**
  - Permissioned Hyperledger Besu enterprise network operating Istanbul/QBFT consensus with 1-second block times and deterministic finality.
  - Zero gas fees incurred by investors: All transactions utilize ERC-4337 Account Abstraction sponsored by institutional Paymasters (`GrowwwPaymaster.sol`).
- **Dutch Auction Smart Contract (`DutchAuction.sol`):**
  - Mathematical decay parameters: Start Price ($P_0$), Reserve Price ($P_{floor}$), Start Time ($T_0$), and End Time ($T_{end}$).
  - Price decay formula executed on-chain:
    $$P(t) = P_{floor} + (P_0 - P_{floor}) \times \max\left(0, \frac{T_{end} - t}{T_{end} - T_0}\right)$$
  - Bidders submit commitments with a maximum price cap: `submitBid(uint256 maxPrice, uint256 tokenAmount)`.
  - When aggregate committed value matches or exceeds remaining token supply at current price $P(t)$, auction clears at $P^* = P(t)$. All winning participants settle uniformly at clearing price $P^*$, and excess capital is refunded immediately.
- **Linear Vesting Vault Contract (`LinearVestingVault.sol`):**
  - Governed by ADR-0044 standards for continuous token release without discrete cliff lockups.
  - Token vesting mathematical formula:
    $$Claimable(t) = TotalAllocated \times \min\left(1.0, \frac{\max(0, t - T_{start})}{T_{end} - T_start}\right) - TotalClaimed$$
  - User claims tokens by calling `claimVestedTokens(address beneficiary)`. The call transfers unlocked ERC-3643 security tokens to the investor verified smart account.
- **Zero On-Chain PII Guarantee:**
  - Smart contracts register only anonymous EVM addresses (`0x...`), ONCHAINID claims hashes, token counts, and cryptographic commitment receipts.
  - Personal investor identities, PAN numbers, and demat accounts are isolated in secure off-chain relational vaults (Prompt 201, Prompt 202).

## State Management Architecture & Subscription Lifecycle
- **Riverpod State Provider Hierarchy:**
  - `LaunchpadCatalogNotifier` (`AutoDisposeAsyncNotifier<LaunchpadCatalogState>`): Manages active, upcoming, and completed offering cards with filtering by asset category (Real Estate, Debt, Equity) and funding status.
  - `DutchAuctionLiveNotifier` (`AutoDisposeAsyncNotifier<DutchAuctionLiveState>`): Manages high-frequency auction telemetry, interpolates current decay price based on server synchronized time, and tracks live clearing price.
  - `AuctionDecayCurveController` (`ChangeNotifier`): Directly interfaces with Flutter `AnimationController` and `CustomPainter` to repaint auction canvas at 60 FPS without triggering full Riverpod widget rebuilds.
  - `BiddingModalNotifier` (`AutoDisposeNotifier<BiddingModalState>`): Validates bid amount inputs against investor wallet balance, verifies accreditation limits, tracks biometric authentication progress, and dispatches API payloads.
  - `VestingStreamNotifier` (`AutoDisposeAsyncNotifier<VestingStreamState>`): Consumes on-chain vault metrics, drives a local continuous `Ticker` computing second-by-second unlocked token fractions, and executes gasless claim operations.
- **Offering Lifecycle State Machine:**
  - `UPCOMING`: Information published, documents accessible, investor whitelisting and accreditation checks active.
  - `AUCTION_ACTIVE`: Dutch auction live, downward price decay in progress, bids accepted over WebSocket and REST.
  - `PRO_RATA_ESCROW`: Fixed-price subscription open, capital collected into smart contract escrow pool.
  - `SETTLING`: Auction concluded or subscription closed, allocation engine computing clearing price and pro-rata shares.
  - `ALLOTMENT_FINALIZED`: Allocations confirmed, excess escrow capital refunded to wallet balances.
  - `VESTING_STREAMING`: Tokens deposited into `LinearVestingVault.sol`, second-by-second linear claims active.
  - `FULLY_VESTED`: 100% of allocation claimed, vault lifecycle complete.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold feature directories under `apps/growww_flutter/lib/features/launchpad/`:
   `domain/models/`, `domain/interfaces/`, `data/datasources/`, `data/repositories/`, `presentation/controllers/`, `presentation/painters/`, `presentation/screens/`, `presentation/widgets/`, and `presentation/modals/`.
2. Define domain models in `domain/models/`: `RwaOfferingDetail`, `DutchAuctionState`, `VestingVaultState`, `BidCommitment`, `AccreditationStatus`, and `OfferingDocument`.
3. Implement `DutchAuctionDecayPainter` in `presentation/painters/dutch_auction_decay_painter.dart` extending `CustomPainter` to render the downward price curve, reserve price baseline, elapsed time shading, and real-time falling price cursor.
4. Implement `LinearVestingStreamWidget` in `presentation/widgets/linear_vesting_stream_widget.dart` utilizing a Flutter `Ticker` to increment unlocked token digits smoothly at sub-second intervals.
5. Create `IRwaLaunchpadRepository` and concrete `RwaLaunchpadRepository` in `data/repositories/` coordinating REST snapshot calls and WebSocket telemetry subscriptions.
6. Implement `LaunchpadCatalogNotifier` using Riverpod code-generation to fetch, filter, and cache primary RWA offerings across Real Estate, Private Debt, and Startup Equity.
7. Implement `DutchAuctionLiveNotifier` managing sub-second price decay synchronization, time-offset drift compensation, and bid book depth merging.
8. Build `OfferingCardWidget` displaying offering category badges, asset valuation, funding progress bar, countdown timer, and accreditation tier indicator.
9. Build `RwaLaunchpadScreen` containing responsive header, asset-class pill filters, active auction banner, and scrollable catalog of offering cards.
10. Build `DutchAuctionDetailScreen` integrating the custom decay curve canvas, live clearing price metric box, cumulative commitment gauge, and bid placement trigger.
11. Build `BiddingSheetModal` with interactive investment slider, maximum price input, token quantity conversion preview, escrow balance check, and biometric auth confirmation.
12. Build `ProRataSubscriptionModal` displaying current oversubscription ratio (e.g. 215%), projected allotment percentage, escrow lockup duration, and refund conditions.
13. Integrate `local_auth` biometric challenge flow within bid dispatch routines, generating signed authorization payloads for the Launchpad Engine.
14. Build `OfferingDocumentVaultBottomSheet` integrating PDF viewer and cryptographic hash verification for independent asset valuation reports and legal prospectuses.
15. Implement gasless vesting claim button linking to Paymaster sponsorship endpoints, updating local state from pending claim to finalized token balance.
16. Write unit tests for Dutch auction price interpolation and vesting calculation logic, and write widget tests verifying bidding sheet input validation and decay curve rendering.

## Interfaces / Contracts

### Domain Enums & Core Identifiers
```dart
// lib/features/launchpad/domain/models/launchpad_enums.dart

enum RwaAssetClass {
  commercialRealEstate,
  privateDebtCreditPool,
  startupEquity,
  infrastructureYield,
}

enum OfferingLifecycleState {
  upcoming,
  whitelisting,
  auctionActive,
  proRataEscrowActive,
  settling,
  allotmentFinalized,
  vestingStreaming,
  closed,
}

enum OfferingPricingModel {
  dutchAuctionDecay,
  proRataSubscriptionEscrow,
  fixedPriceCapped,
}

enum InvestorAccreditationTier {
  retailStandard,
  sebiAccreditedInvestor,
  ifscaQualifiedMarketParticipant,
  institutionalEligible,
}
```

### Offering Dossier & State Models
```dart
// lib/features/launchpad/domain/models/offering_models.dart

class RwaOfferingSummary {
  final String offeringId;
  final String issuerName;
  final String assetTitle;
  final RwaAssetClass assetClass;
  final OfferingPricingModel pricingModel;
  final OfferingLifecycleState lifecycleState;
  final String bannerImageUrl;
  final String tokenSymbol;
  final double targetCapitalInr;
  final double committedCapitalInr;
  final double progressPercentage;
  final DateTime startTime;
  final DateTime endTime;
  final InvestorAccreditationTier requiredTier;
  final bool isUserWhitelisted;

  const RwaOfferingSummary({
    required this.offeringId,
    required this.issuerName,
    required this.assetTitle,
    required this.assetClass,
    required this.pricingModel,
    required this.lifecycleState,
    required this.bannerImageUrl,
    required this.tokenSymbol,
    required this.targetCapitalInr,
    required this.committedCapitalInr,
    required this.progressPercentage,
    required this.startTime,
    required this.endTime,
    required this.requiredTier,
    required this.isUserWhitelisted,
  });
}

class DutchAuctionLiveState {
  final String offeringId;
  final double startPriceInr;
  final double reserveFloorPriceInr;
  final double currentDecayPriceInr;
  final double estimatedClearingPriceInr;
  final double totalTokenSupply;
  final double committedTokensDemand;
  final double totalCapitalCommittedInr;
  final DateTime auctionStartTime;
  final DateTime auctionEndTime;
  final int totalBidsSubmitted;
  final bool isTargetReached;
  final List<AuctionBidDepthLevel> bidDepthBook;

  const DutchAuctionLiveState({
    required this.offeringId,
    required this.startPriceInr,
    required this.reserveFloorPriceInr,
    required this.currentDecayPriceInr,
    required this.estimatedClearingPriceInr,
    required this.totalTokenSupply,
    required this.committedTokensDemand,
    required this.totalCapitalCommittedInr,
    required this.auctionStartTime,
    required this.auctionEndTime,
    required this.totalBidsSubmitted,
    required this.isTargetReached,
    required this.bidDepthBook,
  });

  double get timeElapsedFraction {
    final now = DateTime.now();
    if (now.isBefore(auctionStartTime)) return 0.0;
    if (now.isAfter(auctionEndTime)) return 1.0;
    final totalDuration = auctionEndTime.difference(auctionStartTime).inMilliseconds;
    final elapsed = now.difference(auctionStartTime).inMilliseconds;
    return (elapsed / totalDuration).clamp(0.0, 1.0);
  }
}

class AuctionBidDepthLevel {
  final double priceLevelInr;
  final double volumeTokens;
  final double cumulativeCapitalInr;

  const AuctionBidDepthLevel({
    required this.priceLevelInr,
    required this.volumeTokens,
    required this.cumulativeCapitalInr,
  });
}

class VestingVaultState {
  final String vaultAddress;
  final String tokenSymbol;
  final int tokenDecimals;
  final double totalAllocatedTokens;
  final double totalClaimedTokens;
  final DateTime vestingStartTime;
  final DateTime vestingEndTime;
  final DateTime lastClaimTimestamp;
  final bool isRevocable;

  const VestingVaultState({
    required this.vaultAddress,
    required this.tokenSymbol,
    required this.tokenDecimals,
    required this.totalAllocatedTokens,
    required this.totalClaimedTokens,
    required this.vestingStartTime,
    required this.vestingEndTime,
    required this.lastClaimTimestamp,
    required this.isRevocable,
  });

  double calculateUnlockedTokens(DateTime currentTimestamp) {
    if (currentTimestamp.isBefore(vestingStartTime)) return 0.0;
    if (currentTimestamp.isAfter(vestingEndTime)) return totalAllocatedTokens;
    final totalDurationSeconds = vestingEndTime.difference(vestingStartTime).inSeconds;
    final elapsedSeconds = currentTimestamp.difference(vestingStartTime).inSeconds;
    final fraction = (elapsedSeconds / totalDurationSeconds).clamp(0.0, 1.0);
    return totalAllocatedTokens * fraction;
  }

  double calculateClaimableTokens(DateTime currentTimestamp) {
    final unlocked = calculateUnlockedTokens(currentTimestamp);
    return (unlocked - totalClaimedTokens).clamp(0.0, totalAllocatedTokens);
  }
}
```

### Bidding & WebSocket Payloads
```dart
// lib/features/launchpad/domain/models/bid_payloads.dart

class BidPlacementRequest {
  final String offeringId;
  final String investorSmartAccount;
  final double maxPriceCapInr;
  final double committedAmountInr;
  final double tokenQuantityRequested;
  final String biometricAuthSignature;
  final String idempotencyKey;
  final DateTime timestamp;

  const BidPlacementRequest({
    required this.offeringId,
    required this.investorSmartAccount,
    required this.maxPriceCapInr,
    required this.committedAmountInr,
    required this.tokenQuantityRequested,
    required this.biometricAuthSignature,
    required this.idempotencyKey,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
    'offeringId': offeringId,
    'investorSmartAccount': investorSmartAccount,
    'maxPriceCapInr': maxPriceCapInr,
    'committedAmountInr': committedAmountInr,
    'tokenQuantityRequested': tokenQuantityRequested,
    'biometricAuthSignature': biometricAuthSignature,
    'idempotencyKey': idempotencyKey,
    'timestamp': timestamp.toIso8601String(),
  };
}

class BidPlacementResponse {
  final String bidId;
  final String offeringId;
  final String status;
  final double confirmedAmountInr;
  final String transactionHash;
  final DateTime executedAt;

  const BidPlacementResponse({
    required this.bidId,
    required this.offeringId,
    required this.status,
    required this.confirmedAmountInr,
    required this.transactionHash,
    required this.executedAt,
  });

  factory BidPlacementResponse.fromJson(Map<String, dynamic> json) {
    return BidPlacementResponse(
      bidId: json['bidId'] as String,
      offeringId: json['offeringId'] as String,
      status: json['status'] as String,
      confirmedAmountInr: (json['confirmedAmountInr'] as num).toDouble(),
      transactionHash: json['transactionHash'] as String,
      executedAt: DateTime.parse(json['executedAt'] as String),
    );
  }
}

class AuctionWebSocketTick {
  final String offeringId;
  final double currentPriceInr;
  final double estimatedClearingPriceInr;
  final double cumulativeCapitalCommittedInr;
  final double cumulativeTokensDemanded;
  final int totalBidsCount;
  final DateTime serverTimestamp;

  const AuctionWebSocketTick({
    required this.offeringId,
    required this.currentPriceInr,
    required this.estimatedClearingPriceInr,
    required this.cumulativeCapitalCommittedInr,
    required this.cumulativeTokensDemanded,
    required this.totalBidsCount,
    required this.serverTimestamp,
  });

  factory AuctionWebSocketTick.fromJson(Map<String, dynamic> json) {
    return AuctionWebSocketTick(
      offeringId: json['offeringId'] as String,
      currentPriceInr: (json['currentPriceInr'] as num).toDouble(),
      estimatedClearingPriceInr: (json['estimatedClearingPriceInr'] as num).toDouble(),
      cumulativeCapitalCommittedInr: (json['cumulativeCapitalCommittedInr'] as num).toDouble(),
      cumulativeTokensDemanded: (json['cumulativeTokensDemanded'] as num).toDouble(),
      totalBidsCount: json['totalBidsCount'] as int,
      serverTimestamp: DateTime.parse(json['serverTimestamp'] as String),
    );
  }
}
```

### Repository Contracts
```dart
// lib/features/launchpad/domain/interfaces/i_rwa_launchpad_repository.dart

abstract class IRwaLaunchpadRepository {
  Future<List<RwaOfferingSummary>> fetchOfferings({
    RwaAssetClass? assetClassFilter,
    OfferingLifecycleState? stateFilter,
  });

  Future<RwaOfferingSummary> fetchOfferingDetails(String offeringId);

  Future<DutchAuctionLiveState> fetchAuctionState(String offeringId);

  Stream<AuctionWebSocketTick> subscribeAuctionTicks(String offeringId);

  Future<BidPlacementResponse> submitDutchAuctionBid(BidPlacementRequest request);

  Future<VestingVaultState> fetchVestingVaultState(String vaultAddress);

  Future<String> executeGaslessVestingClaim({
    required String vaultAddress,
    required String beneficiarySmartAccount,
  });
}
```

## Security & Compliance Notes
- **SEBI & IFSCA Accreditation Verification:** Primary issuances of fractional real estate and private credit require accredited status under SEBI (Alternative Investment Funds) Regulations or IFSCA (Fund Management) Regulations. The client validates the presence of an unexpired `AccreditedInvestorBadge` verified by the backend compliance service (Prompt 202). Users lacking accreditation are restricted to view-only mode with bidding buttons locked.
- **Biometric Challenge & Idempotency Safeguards:** Every bid commitment requires biometric confirmation via `local_auth` (FaceID, TouchID, or platform biometric prompt). Bidding payloads are sealed with a client-generated UUIDv4 `idempotencyKey` and timestamped nonce to eliminate duplicate bid submissions caused by network retries.
- **Escrow Balance Pre-Authorization:** To avoid phantom bidding, the client validates available investor ledger balances via the Wallet Service (Prompt 203) before opening the biometric challenge. The backend atomically locks the committed INR or e-Rupee in an institutional escrow account upon bid acceptance.
- **Gasless Transaction Sponsoring via Paymaster:** End users never hold raw native gas tokens on the permissioned Hyperledger Besu network. Vesting claims and bid settlements are wrapped inside ERC-4337 UserOperations and relayed through `GrowwwPaymaster.sol` (Prompt 259) using EIP-712 typed data signatures, preventing gas exhaustion failures.
- **Front-Running & Information Asymmetry Mitigation:** In Dutch auctions, real-time bid depth updates are streamed symmetrically to all connected clients over TLS 1.3 WebSockets. Bids are submitted directly to the off-chain matching coordinator or permissioned mempool, eliminating MEV searcher exploitation common on public blockchains.
- **Zero On-Chain PII Policy:** All smart contract state variables reference only pseudonymous smart contract accounts (`0x...`). Demat account details, PAN credentials, and investor names remain strictly quarantined within ISO-27001 certified backend databases.

## Acceptance Criteria
- [ ] `RwaLaunchpadScreen` renders active, upcoming, and completed primary offerings with responsive category filtering across mobile and desktop viewports.
- [ ] `OfferingCardWidget` accurately reflects asset class badges, funding progress percentage, countdown clocks, and accreditation requirement status.
- [ ] `DutchAuctionDetailScreen` visualizes the downward price decay curve using a smooth `CustomPainter`, updating the current-time falling price indicator continuously.
- [ ] Real-time WebSocket connection to `WSS /ws/v1/launchpad/offerings/{id}/stream` updates current price, clearing price estimate, and total committed demand without UI stutter.
- [ ] `BiddingSheetModal` dynamically computes token quantities based on user capital input and current decay price, enforcing minimum investment thresholds.
- [ ] Bidding flow requires biometric authentication (`local_auth`) and enforces wallet balance sufficiency before triggering bid dispatch.
- [ ] Submitting a bid sends a valid `BidPlacementRequest` with an `idempotencyKey` and displays an animated confirmation receipt with transaction hash.
- [ ] `ProRataSubscriptionModal` displays live oversubscription multiples (e.g. 3.2x), calculates projected allocation proration, and specifies escrow refund policies.
- [ ] `LinearVestingStreamWidget` features a second-by-second animated ticker displaying continuously increasing unlocked token fractions at 60 FPS.
- [ ] Tapping "Claim Unlocked Tokens" executes a gasless Paymaster UserOperation and updates claimed vs claimable token metrics upon on-chain confirmation.
- [ ] Document vault modal allows users to view and download verified valuation reports, SEBI/IFSCA filings, and contract audit documents.
- [ ] Screen readers (TalkBack and VoiceOver) announce offering names, current auction price, remaining time, and vesting totals with semantic contrast compliance.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Scaffolding), Prompt 502 (App Architecture & Riverpod), Prompt 503 (Design System & Theming), Prompt 505 (Authentication UI), Prompt 521 (Secure Storage), Prompt 525 (API Client Layer).
- **Backend Dependencies:** Prompt 266 (Launchpad Engine & Allocation Coordinator), Prompt 203 (Wallet & Account Service), Prompt 202 (KYC, AML & Investor Accreditation), Prompt 259 (ERC-4337 Bundler & Paymaster Service).
- **Blockchain Dependencies:** Prompt 303 (Permissioned Asset Token Issuance Smart Contract), Prompt 306 (ERC-3643 Compliant Security Token Engine), `DutchAuction.sol`, `LinearVestingVault.sol`.
- **Downstream Blockers:** Prompt 510 (Portfolio & Holdings Screen), Prompt 511 (Wallet & Funds Management), Prompt 512 (Transaction History Screen), Prompt 902 (Cross-Platform End-to-End Test Suite).
