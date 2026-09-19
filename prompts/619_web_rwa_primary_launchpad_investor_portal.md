# 619 - Next.js 14 RWA Primary Launchpad & Auction Investor Terminal

## Purpose
In regulated primary capital markets operating under SEBI (Securities and Exchange Board of India), MCA (Ministry of Corporate Affairs), and IFSCA (GIFT City) frameworks, the initial issuance and allocation of tokenized Real-World Assets (RWAs) - such as corporate debt (Non-Convertible Debentures / NCDs), fractional commercial real estate, municipal infrastructure bonds, and commodity-backed warehouse instruments - require mathematically rigorous, transparent, and manipulation-resistant distribution mechanisms. Traditional fixed-price public or private offerings frequently suffer from severe underpricing, allocation opacity, manual book-building bottlenecks, and inequitable distribution favoring syndicate insiders.

Under ADR-0044 (RWA Launchpad Dutch Auctions & Smart Contract Linear Vesting), Growww implements an institutional and retail primary distribution terminal. This terminal combines uniform-price Dutch auction mechanics with pro-rata over-subscription pooling and second-by-second linear streaming token vesting vaults on Hyperledger Besu. Dutch auctions eliminate gas wars, prevent bot-driven front-running, and ensure market-clearing equilibrium where every winning participant pays the identical final clearing price ($P_{clear}$). Concurrently, fixed-yield tranches utilize pro-rata subscription pools with automated excess refund sweeps, while secondary market selling pressure is mitigated through continuous on-chain linear vesting vaults.

This prompt specifies the end-to-end architecture, frontend interface, real-time data streaming, state management, cryptographic signing, and smart contract interaction layer for the **Next.js 14 RWA Primary Launchpad & Auction Investor Terminal** (`apps/growww_web/launchpad`). The terminal provides accredited institutional investors, corporate treasuries, and verified retail participants with an interactive, low-latency cockpit to participate in primary offerings, place and modify auction bids, monitor cumulative demand curves, commit fiat capital via escrow holds, and claim streaming vested security tokens with zero gas overhead.

## What You Are Building
An enterprise-grade, high-performance Next.js 14 web application located in `apps/growww_web/launchpad` (and route group `apps/growww_web/app/(launchpad)/`) featuring:
- `LaunchpadDiscoveryDirectory`: Interactive catalog of primary asset offerings categorized by status (`UPCOMING`, `ACTIVE_AUCTION`, `SUBSCRIPTION_OPEN`, `SETTLING`, `COMPLETED`), asset class (Corporate Debt, Real Estate Yield, Commodity Receipts), risk rating (CRISIL/ICRA AAA to BBB), offering structure (Uniform Price Dutch Auction vs Fixed Pro-Rata), and regulatory regime (Domestic SEBI Section 42 Private Placement vs IFSCA Cross-Border).
- `DutchAuctionTerminal`: Real-time primary auction trading interface incorporating:
  - Dynamic Dutch Auction Price Decay Chart: Canvas and SVG-based rendering of the declining price curve over auction time (from initial ceiling price $P_{start}$ down to reserve floor price $P_{reserve}$), overlaying real-time clearing price ($P_{clear}$) and projected clearing time.
  - Interactive Cumulative Demand vs Supply Depth Curve: Visualizing cumulative bid volume relative to total token tranche allocation to depict auction clearing status in real time.
  - Live Dutch Auction Order Sheet: High-density order book showing pseudonymized active bids, ticket sizes, bid prices, and cutoff demarcation.
  - Bid Placement & Revision Sheet: Fast bidding modal with instant margin calculation, minimum ticket validation, and EIP-712 signature dispatch.
- `ProRataSubscriptionPoolWidget`: Book-building subscription interface for fixed-price asset tranches featuring:
  - Real-time over-subscription multiple gauge (e.g., 1.45x, 3.20x).
  - Dynamic estimated allocation calculator adjusting for green-shoe retention options.
  - Automated excess refund estimator detailing projected cash returns upon allotment.
  - Fiat reservation selector integrating Wallet & Account Ledger Service (Prompt 203) for nodal bank escrow holds.
- `LinearVestingStreamWidget`: High-frequency, second-by-second streaming token release dashboard:
  - Real-time continuous progress bar and numeric ticker rendering unlocked vs locked asset tokens.
  - Precise claimable token counter calculating real-time vesting accrual via client-side high-precision math (`requestAnimationFrame` synchronized with Besu block timestamps).
  - Gasless "One-Click Claim" modal dispatching meta-transactions via EIP-2771 forwarders to `LinearVestingVault.sol`.
  - Historical claim audit log showing timestamps, transaction hashes, and custodial depository receipts.
- `KYCAndAccreditationGateBanner`: Context-aware compliance shield reading investor identity claims directly from `IdentityRegistry.sol` (Prompt 303/305) and CKYC/DigiLocker tiers (Prompt 202/602), dynamically enforcing statutory caps (e.g., max 200 investors under Section 42 of Companies Act 2013) and investor tier restrictions (QIB, HNI, Accredited Retail).
- `PrimaryPortfolioSettlementModal`: Post-auction allocation confirmation card displaying final clearing price, allotted token quantity, executed fiat debit, refund surplus credit, and demat allocation proof.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router investor application under `apps/growww_web/launchpad` and route group `apps/growww_web/app/(launchpad)/`.
  - Client-side and server-rendered Dutch auction price curves, cumulative bid depth charts, and order books.
  - Real-time WebSocket connection to Launchpad Engine (Prompt 266) for auction tick updates, bid placement confirmations, and clearing price shifts.
  - Direct integration with Wallet Service (Prompt 203) via BFF/REST to create, inspect, and release INR fiat escrow holds during bidding.
  - Web3 interaction via Viem 2.x and Wagmi with `DutchAuction.sol` and `LinearVestingVault.sol` deployed on Hyperledger Besu.
  - Client-side EIP-712 structured bid signing and EIP-2771 meta-transaction claim dispatching.
  - Micro-ticker rendering of streaming token vesting claims using 64-bit integer fixed-point arithmetic.
  - Full adherence to SEBI Private Placement lot sizes, 200-investor cap checks, and investor accreditation gating.
  - Unit, component, and Playwright end-to-end test suites.
- **Out of Scope / Handled Elsewhere:**
  - Asset origination, depository demat lien marking, and token factory deployment (handled by Issuer Portal Prompt 615 and Prompt 303).
  - Off-chain matching engine for continuous secondary limit order books (handled by Prompt 204 and Prompt 205).
  - Core banking payment gateway processing (UPI/IMPS/NEFT/RTGS) (handled by Prompt 212).
  - Legal document parsing and CKYC verification backend (handled by Prompt 202).
  - Multi-signature treasury withdrawals and custodial trustee sign-offs (handled by Prompt 217 and Prompt 307).
  - Secondary market AMM liquidity pools and derivatives (handled by Prompt 610 and Prompt 413).

## Technology to Use
- **Frontend Framework:** Next.js 14 App Router leveraging React Server Components (RSC) for initial page hydration, streaming SSR with Suspense boundaries, and React 18 Client Components for high-frequency interactive canvas/chart widgets.
- **Language:** TypeScript 5.x configured with strict mode, zero `any` assertions, and exact optional property typing.
- **Styling & Design System:** Tailwind CSS 3.4+ configured with custom tabular numeric fonts (`font-mono`, `tabular-nums`), dark/light mode switching via CSS variables, and accessible headless UI primitives from Radix UI / shadcn/ui.
- **Web3 & Blockchain SDK:** Viem 2.x and Wagmi 2.x configured for Hyperledger Besu (EVM-compatible, Chain ID 1337 / designated consortium network), handling EIP-712 structured typed data signing, contract ABI encoding/decoding, and EIP-2771 gasless meta-transaction relaying.
- **Real-Time Data Streaming:** Duplex WebSockets utilizing native browser WebSocket or STOMP/WS over TLS with automatic heartbeat reconnection, exponential backoff, and state resynchronization.
- **Data Visualization & Charting:** Lightweight Charts (TradingView) and custom HTML5 Canvas for ultra-fast rendering of the continuous Dutch auction price decay curve, stepped order book supply-demand intersection, and streaming vesting arcs.
- **High-Precision Arithmetic:** Decimal.js and `BigInt` native primitives for all financial calculations, bid valuations, pro-rata allotment allocations, and streaming token claim accumulations to prevent IEEE 754 floating-point inaccuracies.
- **Form Management & Validation:** React Hook Form combined with Zod schemas for client-side validation of bid prices, token quantities, and investment limits against minimum ticket rules.
- **State Management:** Zustand with Immer middleware for local auction state, active bids, stream progress cache, and WebSocket subscription tracking.

## Backend / Infra Touchpoints
- **Launchpad Engine (Prompt 266):**
  - Connects via authenticated REST and low-latency WebSockets (`/api/v1/launchpad/auctions/*` and `wss://api.growww.in/ws/v1/launchpad`).
  - Streams auction lifecycle state changes (`PENDING`, `OPEN`, `PRICE_DECAY`, `CLEARED`, `ALLOTTED`, `CANCELLED`).
  - Ingests investor bids, validates cryptographic signatures, updates aggregate demand, and broadcasts the current uniform clearing price.
- **Wallet & Double-Entry Account Ledger Service (Prompt 203):**
  - Communicates via BFF REST endpoints (`POST /api/v1/wallet/holds/reserve` and `POST /api/v1/wallet/holds/release`).
  - Automatically locks the maximum fiat commitment in a nodal bank escrow hold when an investor places a Dutch auction bid or pro-rata pool subscription.
  - Automatically releases excess funds upon auction clearing or pro-rata allotment completion without manual intervention.
- **Permissioned Asset Token Issuance Service (Prompt 303):**
  - Interfaces with ERC-3643 `DigitalSecurityToken` contracts deployed on Hyperledger Besu.
  - Verifies token metadata (ISIN, face value, coupon structure, credit rating, trustee address).
  - Links to `IdentityRegistry.sol` to verify that the prospective bidder possesses valid KYC and accreditation claims.
- **API Gateway & BFF (Prompt 219):**
  - Enforces JWT session validation, OAuth2 bearer token authentication, request throttling, and mTLS proxying to internal gRPC microservices.
- **Audit Log Service (Prompt 218):**
  - Dispatches non-repudiable audit logs for all placed bids, cancelled bids, terms accepted, and vesting claims with IP addresses and client user agents.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu Consortium Blockchain (QBFT consensus, 2-second block intervals, deterministic finality, gas-metered but sponsored via EIP-2771 forwarder contracts).
- **Interfaced Smart Contracts:**
  - `DutchAuction.sol`:
    - Handles on-chain state verification, final uniform price settlement, and token allotment authorization.
    - Key view and mutative methods:
      ```solidity
      function getAuctionDetails(bytes32 auctionId) external view returns (
          address tokenAddress,
          uint256 totalTokensOffered,
          uint256 startPriceInr,
          uint256 reservePriceInr,
          uint256 startTime,
          uint256 endTime,
          uint256 priceDecayRate,
          uint8 status,
          uint256 currentClearingPrice,
          uint256 totalCommittedCapital
      );

      function getCurrentPrice(bytes32 auctionId) external view returns (uint256);

      function calculateClearingPrice(bytes32 auctionId) external view returns (uint256 clearingPrice, bool isFullySubscribed);

      function submitBidWithPermit(
          bytes32 auctionId,
          uint256 quantity,
          uint256 maxPriceInr,
          uint256 nonce,
          uint256 deadline,
          bytes calldata signature
      ) external returns (bytes32 bidId);

      function claimAuctionAllocation(bytes32 auctionId, address investor) external returns (uint256 allottedTokens, uint256 refundInr);
      ```
  - `LinearVestingVault.sol`:
    - Manages post-allotment linear token streaming for locked tranches.
    - Implements second-by-second continuous unlocking with cliff support.
    - Key view and mutative methods:
      ```solidity
      function getVestingSchedule(address tokenAddress, address beneficiary) external view returns (
          uint256 totalAmount,
          uint256 claimedAmount,
          uint256 startTime,
          uint256 cliffTime,
          uint256 endTime,
          bool isRevocable
      );

      function getClaimableAmount(address tokenAddress, address beneficiary) external view returns (uint256 claimable);

      function claimUnlockedTokens(address tokenAddress) external returns (uint256 amountClaimed);

      function executeClaimViaForwarder(
          address tokenAddress,
          address beneficiary,
          bytes calldata forwarderSignature
      ) external returns (uint256 amountClaimed);
      ```
  - `IdentityRegistry.sol` (Prompt 303):
    - Queried client-side via Viem to verify that the connected wallet address holds an active KYC/AML claim matching the asset's required topic (e.g., Topic 1 = Identity KYC Verified, Topic 2 = Accredited Investor).
    - Method: `isVerified(address userAddress) external view returns (bool)`.

- **Cryptographic Custody & Meta-Transactions:**
  - **Zero PII On-Chain:** Bidder addresses are pseudonymous EVM accounts mapped to off-chain KYC IDs in the relational database.
  - **EIP-712 Typed Bidding Signatures:** Investors sign off-chain bid structures specifying auction ID, quantity, maximum bid price, nonce, and expiration deadline.
  - **EIP-2771 Gasless Vesting Claims:** Investors trigger token claims via platform paymasters, ensuring retail and institutional users never need native network gas tokens.

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Launchpad Route Architecture:**
   - Initialize directory structure at `apps/growww_web/launchpad` and route group `apps/growww_web/app/(launchpad)/`.
   - Configure root layouts, top-level navigation, wallet connect provider, and Web3 modal wrappers:
     - `app/(launchpad)/layout.tsx`: Shell containing header, wallet connection status, KYC tier pill, network indicator, and real-time notification toaster.
     - `app/(launchpad)/page.tsx`: Launchpad home directory displaying active auctions, upcoming offerings, and subscription pools.
     - `app/(launchpad)/auctions/[id]/page.tsx`: Real-time Dutch auction bidding terminal and interactive price discovery canvas.
     - `app/(launchpad)/pools/[id]/page.tsx`: Fixed-price pro-rata subscription book-building interface.
     - `app/(launchpad)/vesting/page.tsx`: Streaming linear vesting portal and gasless claim terminal.
     - `app/(launchpad)/portfolio/page.tsx`: Primary issuance allocation summary, payment hold ledger, and demat credit certificates.

2. **Establish Data Models & Database Schemas:**
   - Define PostgreSQL tables via Prisma/Drizzle in `packages/database/prisma/schema.prisma`:
     - `LaunchpadOffering`: Stores offering type (`DUTCH_AUCTION`, `PRO_RATA_POOL`), asset reference ID, ISIN, total units offered, minimum lot size, price parameters, start/end timestamps, and status.
     - `AuctionBid`: Stores user ID, auction offering ID, requested quantity, maximum limit price, committed fiat amount, wallet hold ID, EIP-712 signature hash, status (`ACTIVE`, `MATCHED`, `OUTBID`, `CANCELLED`, `ALLOTTED`), and allotment quantity.
     - `ProRataSubscription`: Stores user ID, pool offering ID, committed capital, deposit hold ID, target allocation, settled allocation, and excess refund amount.
     - `VestingSchedule`: Stores token address, beneficiary address, total allocation, cliff timestamp, duration, claimed tokens, last claimed timestamp, and Besu transaction receipts.
     - `InvestorAccreditationCache`: Caches on-chain `IdentityRegistry` verification status, claim expiration, and SEBI investor classification (QIB, HNI, Retail).

3. **Configure Web3 Viem & Wagmi Client Layer:**
   - Setup `packages/web3-config/src/client.ts` configured with Hyperledger Besu RPC transport, fallback failovers, and custom chain definitions.
   - Implement typed contract definitions importing ABIs for `DutchAuction.sol`, `LinearVestingVault.sol`, and `IdentityRegistry.sol`.
   - Setup React Query client with optimized cache invalidation for blockchain reads (staleTime: 4000ms, block-synced).
   - Implement EIP-712 domain separator and structured type definition for `DutchAuctionBid(bytes32 auctionId,uint256 quantity,uint256 maxPriceInr,uint256 nonce,uint256 deadline)`.

4. **Build `LaunchpadDiscoveryDirectory` Component:**
   - Implement responsive server-rendered grid displaying primary offerings with dynamic filtering (Asset Class, Maturity, Expected Yield, Rating, Auction vs Fixed).
   - Render live badge indicators indicating auction countdown timer, time remaining, subscription progress bar, and accreditation tier required.
   - Implement optimistic search and filter state stored in URL query parameters (`searchParams`).

5. **Develop `DutchAuctionPriceChart` Canvas Component:**
   - Construct a high-performance visual component displaying the uniform Dutch auction price decay curve:
     - Calculates decay trajectory: $P(t) = P_{start} - \left(\frac{P_{start} - P_{reserve}}{T_{duration}}\right) \times (t - T_{start})$ for linear decay (or configurable exponential decay curve).
     - Renders historical decay line, projected decay line, current elapsed price point, and animated ticker.
     - Visualizes the horizontal clearing price line ($P_{clear}$) calculated from the cumulative bid book intersection.
     - Synchronizes chart time axis with Besu block timestamp received via WebSocket to avoid client clock drift.

6. **Build `CumulativeDemandDepthChart` Component:**
   - Construct stepped area chart depicting aggregate investor demand (X-axis: Cumulative Units, Y-axis: Bid Price):
     - Plots horizontal line representing Total Offering Supply.
     - Highlights the exact intersection point where Cumulative Demand equals Total Supply, marking the dynamic Clearing Price ($P_{clear}$).
     - Visually differentiates In-The-Money (winning) bids vs Out-of-The-Money (unallocated) bids using contrasting color accents.

7. **Implement `DutchAuctionOrderSheet` & Bidding Flow:**
   - Build virtualized TanStack table displaying live incoming bids with pseudonymized address masks (e.g., `0x7a3...9b2`), timestamp, unit volume, and bid price.
   - Build `BidPlacementSheet` modal:
     - Ingests desired quantity and maximum price willing to pay.
     - Validates minimum ticket size (e.g., ₹10,000 retail, ₹1,00,000 corporate/HNI).
     - Automatically calls Wallet Service (Prompt 203) to confirm available INR fiat balance.
     - Dispatches EIP-712 signature request to user's connected wallet or in-app passkey signer.
     - Transmits signed bid payload to Launchpad Engine (Prompt 266) via WebSocket, receiving immediate order receipt.

8. **Build `ProRataSubscriptionPoolWidget` Component:**
   - Construct primary book-building interface for fixed-price asset tranches:
     - Real-time gauge displaying Total Committed Capital vs Target Issue Size.
     - Dynamic calculation of current over-subscription factor ($\text{Multiple} = \frac{\text{Total Committed}}{\text{Target Size}}$).
     - Live allotment estimator: $\text{Estimated Units} = \min\left(\text{Requested Units}, \frac{\text{Requested Units}}{\text{Multiple}}\right)$.
     - Clear refund breakdown showing estimated immediate cash return if the tranche remains oversubscribed at closing.
     - One-click subscription commitment initiating atomic fiat escrow hold via Prompt 203.

9. **Develop `LinearVestingStreamWidget` Component:**
   - Construct real-time streaming token release interface:
     - Queries `LinearVestingVault.sol` via Viem for `getVestingSchedule` and `getClaimableAmount`.
     - Sets up a high-precision sub-second numeric ticker powered by `requestAnimationFrame` and `BigInt` time math:
       $$\text{Claimable}(t) = \frac{\text{TotalAmount} \times (t - \text{StartTime})}{\text{EndTime} - \text{StartTime}} - \text{ClaimedAmount}$$
     - Visualizes continuous stream progression with glowing circular arc or linear bar displaying unlocked, claimable, and remaining locked tokens.
     - Automatically accounts for cliff periods, locking claims until cliff timestamp is surpassed.

10. **Implement EIP-2771 Gasless Vesting Claim Modal:**
    - Develop transaction dispatcher for claiming unlocked tokens:
      - Checks if `claimable > 0`.
      - Prepares forwarder meta-transaction request with user address, token address, nonce, and deadline.
      - Prompts user for a lightweight off-chain signature (EIP-712 ForwardRequest).
      - Posts signed request to platform Gas Relayer / Paymaster service (`/api/v1/launchpad/vesting/relay-claim`).
      - Displays live step-by-step transaction tracker: Signature Verified -> Relayer Dispatched -> Block Included (Besu) -> Tokens Transferred to Demat/Wallet.

11. **Build `KYCAndAccreditationGateBanner` & Regulatory Restrictions:**
    - Read user's on-chain identity status from `IdentityRegistry.sol`:
      - Validates presence of verified KYC claim (`Topic 1`).
      - Validates Accredited Investor qualification (`Topic 2`) for private credit offerings.
    - If unverified, renders blocking overlay with direct deep-link to Web Onboarding & KYC Flow (Prompt 602).
    - Enforces Companies Act Section 42 check: If auction/pool participant count reaches 200, disables retail bidding and displays regulatory cap notification.

12. **Implement Real-Time WebSocket Synchronization Layer:**
    - Build React hook `useLaunchpadSocket(offeringId)`:
      - Connects to `wss://api.growww.in/ws/v1/launchpad/auctions/{id}`.
      - Subscribes to topics: `auction:ticks`, `auction:bids`, `auction:clearing`, `pool:stats`.
      - Implements heartbeat ping/pong (15s interval) and connection recovery with state catch-up via REST query.
      - Dispatches incoming ticks directly into Zustand store, updating charts without triggering unnecessary React component tree re-renders.

13. **Construct `PrimaryPortfolioSettlementModal` & Document Exporter:**
    - Build post-settlement summary dashboard:
      - Displays final auction clearing price ($P_{clear}$) vs user's maximum bid price.
      - Displays total allotted units, executed fiat consideration, and refunded excess funds.
      - Generates downloadable Cryptographic Allocation Certificate containing SHA-256 hash of the on-chain Besu settlement transaction, Demat credit reference, and Debenture Trustee attestation.

14. **End-to-End Testing, Resilience, and Verification:**
    - Write unit and component tests with Vitest and React Testing Library:
      - Test Dutch auction price decay calculation across linear and exponential curves.
      - Test pro-rata allocation math, floor rounding, and excess refund calculations.
      - Test streaming vesting accrual logic, cliff handling, and micro-second timer updates.
      - Test Zod validation schemas for minimum lot sizes and maximum investment limits.
    - Implement Playwright E2E integration suites:
      - Test wallet connection, KYC accreditation gating, and access restriction enforcement.
      - Test placing a Dutch auction bid, verifying simulated fiat escrow hold, and checking order sheet appearance.
      - Test real-time WebSocket tick receipt and dynamic chart update.
      - Test gasless linear vesting claim flow with mock EIP-2771 forwarder.

## Interfaces / Contracts

### Protobuf Definition: Launchpad Engine & Auction Service
```protobuf
syntax = "proto3";

package growww.launchpad.v1;

import "google/protobuf/timestamp.proto";

enum OfferingType {
  OFFERING_TYPE_UNSPECIFIED = 0;
  OFFERING_TYPE_DUTCH_AUCTION = 1;
  OFFERING_TYPE_PRO_RATA_SUBSCRIPTION = 2;
  OFFERING_TYPE_FIXED_PRICE_FIFO = 3;
}

enum OfferingStatus {
  OFFERING_STATUS_UNSPECIFIED = 0;
  OFFERING_STATUS_UPCOMING = 1;
  OFFERING_STATUS_OPEN = 2;
  OFFERING_STATUS_PAUSED = 3;
  OFFERING_STATUS_PRICE_CLEARED = 4;
  OFFERING_STATUS_SETTLED_ALLOTTED = 5;
  OFFERING_STATUS_CANCELLED = 6;
}

enum BidStatus {
  BID_STATUS_UNSPECIFIED = 0;
  BID_STATUS_PENDING_HOLD = 1;
  BID_STATUS_ACTIVE = 2;
  BID_STATUS_OUTBID = 3;
  BID_STATUS_MATCHED_WINNING = 4;
  BID_STATUS_SETTLED = 5;
  BID_STATUS_REFUNDED = 6;
  BID_STATUS_CANCELLED = 7;
}

message LaunchpadOfferingSummary {
  string offering_id = 1;
  string asset_id = 2;
  string isin = 3;
  string token_name = 4;
  string token_symbol = 5;
  string token_address = 6;
  OfferingType offering_type = 7;
  OfferingStatus status = 8;
  string total_units_offered = 9;
  string start_price_inr = 10;
  string reserve_price_inr = 11;
  string current_price_inr = 12;
  string clearing_price_inr = 13;
  string total_committed_capital_inr = 14;
  double oversubscription_multiple = 15;
  int32 total_participants_count = 16;
  int32 max_participants_limit = 17;
  string min_ticket_inr = 18;
  string max_ticket_inr = 19;
  google.protobuf.Timestamp start_time = 20;
  google.protobuf.Timestamp end_time = 21;
  google.protobuf.Timestamp vesting_start_time = 22;
  google.protobuf.Timestamp vesting_end_time = 23;
  uint64 vesting_cliff_seconds = 24;
}

message PlaceDutchBidRequest {
  string offering_id = 1;
  string investor_ucc = 2;
  string wallet_address = 3;
  string requested_units = 4;
  string max_limit_price_inr = 5;
  uint64 nonce = 6;
  uint64 deadline = 7;
  bytes eip712_signature = 8;
}

message PlaceDutchBidResponse {
  string bid_id = 1;
  string offering_id = 2;
  BidStatus status = 3;
  string committed_fiat_inr = 4;
  string hold_reference_id = 5;
  google.protobuf.Timestamp created_at = 6;
}

message VestingScheduleResponse {
  string token_address = 1;
  string beneficiary_address = 2;
  string total_granted_tokens = 3;
  string claimed_tokens = 4;
  string claimable_tokens = 5;
  string locked_tokens = 6;
  google.protobuf.Timestamp start_time = 7;
  google.protobuf.Timestamp cliff_time = 8;
  google.protobuf.Timestamp end_time = 9;
  bool is_claim_active = 10;
}
```

### TypeScript Data Models & Client Interfaces
```typescript
export type OfferingType = 'DUTCH_AUCTION' | 'PRO_RATA_SUBSCRIPTION' | 'FIXED_PRICE_FIFO';

export type OfferingStatus = 
  | 'UPCOMING'
  | 'OPEN'
  | 'PAUSED'
  | 'PRICE_CLEARED'
  | 'SETTLED_ALLOTTED'
  | 'CANCELLED';

export interface LaunchpadOffering {
  offeringId: string;
  assetId: string;
  isin: string;
  tokenName: string;
  tokenSymbol: string;
  tokenAddress: `0x${string}`;
  offeringType: OfferingType;
  status: OfferingStatus;
  totalUnitsOffered: string; // 18-decimal string
  startPriceInr: string;      // Fixed-point string (2 decimals)
  reservePriceInr: string;
  currentPriceInr: string;
  clearingPriceInr?: string;
  totalCommittedCapitalInr: string;
  oversubscriptionMultiple: number;
  totalParticipantsCount: number;
  maxParticipantsLimit: number; // 200 for Section 42
  minTicketInr: string;
  maxTicketInr: string;
  startTime: string; // ISO-8601
  endTime: string;
  vestingConfig: {
    startTime: string;
    endTime: string;
    cliffSeconds: number;
    isStreaming: boolean;
  };
}

export interface DutchAuctionBid {
  bidId: string;
  offeringId: string;
  investorUcc: string;
  walletAddress: `0x${string}`;
  units: string;
  maxPriceInr: string;
  committedCapitalInr: string;
  holdReferenceId: string;
  status: 'ACTIVE' | 'MATCHED' | 'OUTBID' | 'ALLOTTED' | 'REFUNDED';
  allottedUnits?: string;
  finalPriceInr?: string;
  refundInr?: string;
  timestamp: string;
}

export interface VestingStreamState {
  tokenAddress: `0x${string}`;
  tokenSymbol: string;
  beneficiary: `0x${string}`;
  totalGranted: bigint;
  claimed: bigint;
  claimable: bigint;
  remainingLocked: bigint;
  startTime: number; // Unix seconds
  cliffTime: number;
  endTime: number;
  unlockRatePerSecond: bigint; // Tokens per second (18 decimals)
}

export interface EIP712BidPayload {
  domain: {
    name: string;
    version: string;
    chainId: number;
    verifyingContract: `0x${string}`;
  };
  types: {
    DutchAuctionBid: [
      { name: 'auctionId'; type: 'bytes32' },
      { name: 'investor'; type: 'address' },
      { name: 'quantity'; type: 'uint256' },
      { name: 'maxPriceInr'; type: 'uint256' },
      { name: 'nonce'; type: 'uint256' },
      { name: 'deadline'; type: 'uint256' }
    ];
  };
  primaryType: 'DutchAuctionBid';
  message: {
    auctionId: `0x${string}`;
    investor: `0x${string}`;
    quantity: bigint;
    maxPriceInr: bigint;
    nonce: bigint;
    deadline: bigint;
  };
}
```

### REST & WebSocket API Endpoints
- `GET /api/v1/launchpad/offerings`:
  - Query: `?status=OPEN&type=DUTCH_AUCTION&limit=20&page=1`
  - Returns paginated list of active and upcoming launchpad offerings.
- `GET /api/v1/launchpad/offerings/{id}`:
  - Returns complete offering details, price decay configuration, depository lien proof, and vesting parameters.
- `POST /api/v1/launchpad/offerings/{id}/bids`:
  - Request: `{ requestedUnits: string, maxLimitPriceInr: string, eip712Signature: string, nonce: number, deadline: number }`
  - Validates balance with Wallet Service (Prompt 203), locks INR escrow hold, registers bid with Launchpad Engine (Prompt 266), and returns bid confirmation.
- `DELETE /api/v1/launchpad/offerings/{id}/bids/{bidId}`:
  - Cancels active bid prior to auction clearing and initiates immediate release of fiat escrow hold.
- `POST /api/v1/launchpad/offerings/{id}/subscribe-pool`:
  - Initiates fixed-price pro-rata pool subscription commitment, locking funds via Wallet Service.
- `GET /api/v1/launchpad/vesting/schedules`:
  - Returns all active and historical vesting schedules for the authenticated user across all subscribed token issuances.
- `POST /api/v1/launchpad/vesting/relay-claim`:
  - Relays signed EIP-2771 forwarder transaction to execute gasless token claim from `LinearVestingVault.sol`.
- `WSS wss://api.growww.in/ws/v1/launchpad/auctions/{id}`:
  - Duplex WebSocket connection streaming real-time auction state:
    - Channel: `auction:ticks`: Emits current decaying price, elapsed time, and remaining duration.
    - Channel: `auction:bids`: Emits new pseudonymized bids and depth changes.
    - Channel: `auction:clearing`: Emits live calculated clearing price ($P_{clear}$) and over-subscription metrics.

## Security & Compliance Notes
- **SEBI Private Placement Cap (Companies Act Section 42):**
  - Section 42 of the Companies Act 2013 strictly restricts private placements to a maximum of 200 distinct investors in an aggregate financial year per security class.
  - The Launchpad application enforces this limit both in frontend UI (disabling bidding when confirmed participants reach 200) and through the backend Launchpad Engine. The on-chain `Compliance.sol` module acts as the ultimate immutable arbiter, reverting any mint or transfer that would expand the token holder registry beyond 200 holders.
- **Accreditation & KYC Gating:**
  - Every prospective bidder must possess an active, unexpired claim on `IdentityRegistry.sol` verified by a SEBI-registered KYC Registration Agency (KRA / CKYC).
  - High-yield unsecured debentures and structured private credit offerings enforce Accredited Investor (AI) or Qualified Institutional Buyer (QIB) claim requirements before the bidding interface unlocks.
- **Transaction Signing Security & Anti-Phishing:**
  - All auction bids require EIP-712 typed structured data signing. The domain separator binds the signature to the specific chain ID, verifying contract address, and auction ID, completely preventing cross-chain replay or cross-auction injection attacks.
  - Nonces and strict deadlines (maximum 10-minute validity) guarantee that stale signatures cannot be submitted after price moves.
- **Nodal Escrow Holds & Double-Spend Prevention:**
  - Fiat capital is never deposited into unregulated hot wallets. When a bid is placed, the Wallet Service (Prompt 203) executes a pessimistic row lock (`SELECT ... FOR UPDATE`) in PostgreSQL, establishing a dedicated fund hold in the nodal bank escrow account.
  - If a bid is outbid or cancelled, the hold is released immediately back to the available cash balance.
- **EIP-2771 Meta-Transaction Paymaster Controls:**
  - The gasless vesting claim relayer enforces strict rate-limiting (maximum 1 claim per 60 seconds per beneficiary address) to prevent relayer denial-of-service or spam transactions on Hyperledger Besu.
  - The paymaster contract validates the forwarder signature against the beneficiary's registered address, preventing unauthorized third-party token diversion.
- **Zero PII on Distributed Ledger:**
  - In compliance with the Digital Personal Data Protection (DPDP) Act 2023, investor names, PANs, phone numbers, and bank account numbers are never written to the blockchain. All on-chain events and contract storage reference pseudonymized EVM addresses tied to off-chain identity mappings.

## Acceptance Criteria
- [ ] Next.js 14 application compiles and starts successfully under `apps/growww_web/launchpad` with zero TypeScript build errors or lint warnings.
- [ ] `LaunchpadDiscoveryDirectory` correctly displays active, upcoming, and settled primary offerings with accurate status badges and filtering.
- [ ] `DutchAuctionPriceChart` renders a smooth price decay curve synchronized with Besu block timestamps, accurately reflecting price changes over time.
- [ ] `CumulativeDemandDepthChart` plots cumulative demand against total offering supply and accurately highlights the dynamic uniform clearing price ($P_{clear}$).
- [ ] Dutch auction bidding modal successfully validates minimum ticket sizes, checks fiat wallet balances via Prompt 203, prompts for EIP-712 signature, and registers bid via WebSocket.
- [ ] Cancelling a bid releases the fiat escrow hold immediately and updates the order book in real time.
- [ ] `ProRataSubscriptionPoolWidget` accurately calculates over-subscription multiples and provides exact estimated token allocations and refund amounts.
- [ ] `LinearVestingStreamWidget` updates the claimable token count smoothly using sub-second precision via `requestAnimationFrame` and `BigInt` time calculations.
- [ ] Gasless vesting claim executes successfully via EIP-2771 meta-transaction forwarder without requiring native gas tokens in the user's wallet.
- [ ] Connected wallets lacking verified identity claims on `IdentityRegistry.sol` are blocked by `KYCAndAccreditationGateBanner` with clear accreditation instructions.
- [ ] Bidding is programmatically disabled when an offering reaches the statutory 200-investor private placement threshold.
- [ ] Vitest unit test suite verifies Dutch auction decay formulas, pro-rata allocation math, and streaming vesting calculations with 100% pass rate.
- [ ] Playwright E2E tests validate complete investor journey: browsing launchpad -> inspecting auction curve -> placing bid with escrow hold -> auction clearing -> streaming token vesting claim.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 203: Wallet & Double-Entry Account Ledger Service (for nodal bank escrow fund holds).
  - Prompt 266: Launchpad Engine Microservice (for auction orchestration and order book state).
  - Prompt 303: Permissioned Asset Token Issuance Smart Contract (for ERC-3643 tokens and Identity Registry).
  - Prompt 601: Next.js 14 Investor Web App Scaffolding (for shared UI layout and theme providers).
  - Prompt 602: Web Onboarding, DigiLocker & Camera KYC Flow (for investor identity onboarding).
- **Parallel Tasks:**
  - Prompt 610: Web Options Chain, Strategy Builder & Multi-Chain Web3 Deposit Portal.
  - Prompt 615: Next.js 14 RWA Asset Issuer & Tokenization Originator Portal.
- **Downstream Blockers:**
  - Prompt 906: Comprehensive UAT Plan & Regulatory Sandbox Scenarios.
  - Prompt 907: Regulatory Sandbox Pilot Launch Plan.
