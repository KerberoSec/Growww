# 347 - On-Chain Affiliate Rebate & Commission Splitter Smart Contract (Solidity)

## Purpose
In electronic financial exchanges and retail brokerage ecosystems, affiliate referral incentives, sub-broker revenue sharing, and high-volume maker rebates serve as primary catalysts for user acquisition and market liquidity. In traditional Indian capital markets regulated by the Securities and Exchange Board of India (SEBI), intermediary revenue sharing is subject to rigorous regulatory boundaries: brokerages and clearing members are prohibited from offering speculative cash kickbacks, guaranteed profit distributions, or multi-level marketing (MLM) commission cascades. Furthermore, any promotional incentive or volume rebate must be funded strictly from the platform's proprietary revenue share rather than touching client principal, statutory clearing funds, or regulatory tax escrows.

On the Growww National Blockchain Stock Exchange (NBSE) deployed on a permissioned Hyperledger Besu consortium ledger, secondary trade settlement operates under an atomic Delivery-versus-Payment (DvP Model 1) regime. Every executed trade incurs a canonical, non-negotiable 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee. Under Growww's institutional fee distribution waterfall, this fee is programmatically allocated across three segregated channels: 60% to the Platform Operating Treasury, 25% to the Core Settlement Guarantee Fund (SGF), and 15% to the Investor Protection Fund (IPF). 

Historically, affiliate and rebate disbursements in retail finance have relied on asynchronous, off-chain batch processing cycles at monthly or weekly intervals. This legacy approach introduces severe operational vulnerabilities: opaque fee accounting, delayed payouts, disputes over trade attribution, heavy administrative overhead, and friction for institutional market makers demanding real-time rebate realization. 

This prompt specifies the **On-Chain Affiliate Rebate & Commission Splitter Smart Contract (`AffiliateCommissionSplitter.sol`, `IAffiliateCommissionSplitter.sol`)**. Deployed under `contracts/src/governance/`, this contract delivers an automated, verifiable, deterministic, and gasless on-chain distribution engine for affiliate referral commissions and volume rebates. Operating on Hyperledger Besu under QBFT consensus, the contract intercepts designated revenue allocations directly from the Platform Operating Treasury's 60% fee split at the moment of trade settlement. It computes multi-tier referral splits and maker-taker volume rebates in tokenized fiat (weINR) and institutional cross-border stablecoins (USDT), enforces strict immutable fee ceilings, implements anti-gaming wash-trading dampeners, isolates reward accounting from user principal, and empowers beneficiaries to claim accumulated earnings gaslessly via ERC-4337 Account Abstraction paymasters.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/src/governance/` comprising:
- `AffiliateCommissionSplitter.sol`: Core upgradeable contract (UUPS proxy pattern) managing affiliate registration, tiered commission splits, volume rebate schedules, anti-gaming validation, escrow cooldowns, and gasless reward claims.
- `IAffiliateCommissionSplitter.sol`: Master Solidity interface specifying all affiliate data structures, tier configurations, commission calculation models, custom error codes, regulatory event schemas, and external mutator/view function signatures.
- Multi-Tier Referral & Volume Rebate Calculation Engine: Mathematical calculation module supporting multi-tier referral splits (Tier 1 direct referrer, Tier 2 master affiliate network) and institutional trading desk volume tiers (Bronze, Silver, Gold, Platinum) operating with basis-point precision (`BPS_PRECISION = 10,000`).
- Strict Fee Boundary Enforcement: Immutable runtime invariants guaranteeing that affiliate payouts are funded strictly and exclusively from the platform Treasury operating fee split, enforcing an absolute ceiling (`MAX_AFFILIATE_SHARE_BPS = 3000`, representing a maximum of 30% of the Treasury fee portion, or 0.0018% of gross trade volume), preserving the Core SGF (25%), IPF (15%), and Section 194S e-TDS tax withholdings.
- Anti-Gaming and Anti-Sybil Protection Module: Cryptographic validation layer that rejects self-referrals via salted identity hashes (`panHash`), imposes minimum trade interval limits, tracks per-block volume velocity to neutralize wash trading, and enforces a mandatory T+1 settlement cooldown window prior to reward unlock.
- Pull-Over-Push Reward Accounting & Gasless Claims: Internal accounting ledger accumulating non-transferable reward entitlements per affiliate address, enabling users to withdraw accrued weINR/USDT rewards on-demand directly or via ERC-4337 Account Abstraction Paymasters using EIP-712 meta-transaction claim permits.
- Comprehensive Foundry Test Suite (`test/governance/AffiliateCommissionSplitter.t.sol`): Exhaustive unit, fuzz, invariant, and differential test suites verifying mathematical precision, zero-leakage conservation of funds, role-based access control, and anti-gaming protections.
- Deployment and Upgrade Scripts (`script/DeployAffiliateCommissionSplitter.s.sol`): Hardened Foundry deployment scripts supporting UUPS initialization and multi-sig ownership transfer.

## Scope Boundaries
- **In Scope:**
  - Automated on-chain splitting of trading fees allocated from the Platform Treasury operating share into affiliate and rebate reward ledgers.
  - Multi-tier commission distribution logic: Tier 1 (direct referrer), Tier 2 (master affiliate/channel partner), and institutional volume rebate tiers.
  - Multi-currency reward support: Domestic tokenized sovereign rupee (weINR, 18 decimals) and cross-border GIFT City settlement currency (USDT, 6 decimals).
  - Strict mathematical bounding: Hard maximum affiliate commission cap (`MAX_AFFILIATE_SHARE_BPS`) to eliminate the risk of fee over-allocation.
  - Anti-gaming mechanics: Rejection of self-referrals (`referrer != referee` and `panHash` inequality), wash-trading dampening, and configurable reward escrow cooldowns.
  - Non-transferable internal reward ledger: Rewards accrue as internal accounting balances claimable only by the verified beneficiary; reward balances cannot be transferred, traded, or wrapped into secondary tokens.
  - Gasless claims via ERC-4337 Paymaster integration and EIP-712 signature verification (`claimRewardsWithPermit`).
  - Emergency administration: Circuit breaker pause controls, multi-sig parameter governance, and malicious account blacklisting with reward forfeiture to the Investor Protection Fund.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain referral link generation, deep linking, click attribution, and web conversion tracking (handled in Prompt 224 and Prompt 271).
  - Determination of gross trade volume, matching engine order execution, and order book matching (handled in Prompt 204 and Prompt 205).
  - Primary 0.00% (Zero Fee) platform fee calculation and gross deduction during DvP execution (handled in Prompt 210 and Prompt 329).
  - Statutory Section 194S e-TDS tax deduction and challan remittance (handled in Prompt 223 and Prompt 336).
  - Statutory allocation to the Settlement Guarantee Fund (handled in Prompt 315 and Prompt 329).
  - Fiat bank withdrawals via UPI/IMPS/NEFT and banking gateway integration (handled in Prompt 203 and Prompt 212).
  - Institutional cross-chain liquidity bridge operations (handled in Prompt 319 and Prompt 323).

| Out-of-Scope Component | Responsible System | Relevant Prompt |
| :--- | :--- | :--- |
| Off-Chain Referral Attribution & Campaigns | Referral & Growth Service / Affiliate Engine | Prompt 224 / Prompt 271 |
| Gross Trade Execution & DvP Settlement | Settlement DvP & Matching Engine | Prompt 205 / Prompt 306 / Prompt 329 |
| Platform Fee Calculation (0.00% (Zero Fee) Fixed) | Fee & Realized PnL Engine | Prompt 210 / Prompt 006 |
| Statutory e-TDS Withholding (Section 194S) | Automated Tax Withholding & e-TDS Ledger | Prompt 223 / Prompt 336 |
| Settlement Guarantee Fund Waterfall | Settlement Guarantee Fund Contract | Prompt 230 / Prompt 315 |
| ERC-4337 Paymaster Sponsorship & Bundler | AA Bundler & Paymaster Service | Prompt 259 / Prompt 337 |
| Off-Chain Fiat Off-Ramping (UPI/RTGS) | Payment Gateway Integration Service | Prompt 212 / Prompt 203 |

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD` for gas-efficient reentrancy locks, native checked arithmetic, and custom errors).
- **Base Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`).
  - OpenZeppelin Contracts v5.0 `SafeERC20` for secure transfer of weINR and USDT.
  - OpenZeppelin Contracts v5.0 `Math.sol` for full 512-bit safe fixed-point multiplication and division (`mulDiv`).
  - OpenZeppelin Contracts v5.0 `ECDSA` and `EIP712Upgradeable` for cryptographic signature verification on gasless claim permits.
- **Token Standards:**
  - ERC-20 (`IERC20`, `SafeERC20`) for weINR (18 decimals) and USDT (6 decimals).
  - ERC-3643 (`IERC3643`, `IIdentityRegistry`) integration for validating recipient KYC/AML compliance status prior to token distribution.
- **Arithmetic Standards:**
  - Standard basis-point scaling: `BPS_PRECISION = 10_000` (where 100 bps = 1.00%, 0 bps (0.00% fee at launch) = 0.00% (Zero Fee)).
  - Treasury share basis: Payouts are computed as a pro-rata fraction of the Platform Treasury allocation (`treasuryFeeAmount`), strictly bounded by `MAX_AFFILIATE_SHARE_BPS = 3_000` (30.00% of Treasury fee).
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, gas profiling, and invariant fuzzing; `cast` for Besu JSON-RPC diagnostic interaction).
- **Static Analysis & Formal Auditing:** Slither (Trail of Bits), Mythril, and Solhint CI pipelines.

## Backend / Infra Touchpoints
- **Affiliate Referral Engine (Prompt 271):** Core microservice managing affiliate registration, referral trees, campaign attribution, and off-chain volume aggregation. Dispatches signed tier updates and batch split transactions to Besu via relayer queues.
- **Fee Engine (Prompt 210):** Evaluates gross transaction fees on trade executions, separates the Treasury reserve share from SGF (25%) and IPF (15%), and invokes `AffiliateCommissionSplitter.sol` to record eligible commission legs.
- **Account Abstraction Bundler & Paymaster Service (Prompt 259):** Submits gasless user operations on behalf of affiliates claiming accumulated commissions, sponsoring gas fees on Besu via the `NBSEPaymaster` contract.
- **NBSE Delivery-versus-Payment Settlement & Fee Collector (Prompt 329):** The on-chain settlement gateway that atomically collects the 0.00% (Zero Fee) platform fee during trade finalization and routes the Treasury fee portion to `AffiliateCommissionSplitter.sol`.
- **Trade Settlement Service (Prompt 208):** Ingests matched trades, constructs settlement batches, verifies counterparty identity commitments, and passes referral attribution tags to the settlement transaction.
- **Real-Time Market Surveillance Engine (Prompt 228):** Continuously monitors trade graphs for wash-trading patterns, circular volume loops, or sybil referral rings; dispatches automated freeze signals to blacklist collusive affiliate accounts.
- **Blockchain Event Indexer (Prompt 309):** Ingests `CommissionSplitProcessed`, `RewardAccrued`, `RewardsClaimed`, and `AffiliateTierUpdated` events, updating PostgreSQL read-replicas within 50ms for display in affiliate dashboards.
- **Wallet & Account Service (Prompt 203):** Tracks off-chain mirror ledger balances of investor and affiliate weINR holdings for reconciliation against bank-held escrow reserves.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Block Finality:** Deployed on Hyperledger Besu private permissioned consortium network running QBFT consensus with 2-second deterministic block intervals and instant single-block transaction finality (zero chain reorganizations).
- **Direct Atomic Split from Treasury Inflow:** When `NBSEFeeCollector.sol` processes platform fee allocations during trade settlement, it routes the designated affiliate rebate allocation directly to `AffiliateCommissionSplitter.sol`. The splitter contract credits the internal ledger without executing synchronous external token transfers, minimizing gas consumption in the critical settlement loop.
- **Zero On-Chain PII Guarantee:** In accordance with the Digital Personal Data Protection Act (DPDPA 2023) and SEBI intermediary standards, no raw personal information (such as affiliate legal names, PAN, Aadhaar, email addresses, or phone numbers) is recorded on-chain. Affiliates and traders are identified solely by their EVM wallet addresses (`address affiliateAddress`), salted identity commitments (`bytes32 panHash = keccak256(abi.encodePacked(rawPAN, salt))`), and hashed referral codes (`bytes32 referralCodeHash`).
- **Multi-Currency Token Support:** 
  - weINR: 18-decimal wrapped electronic Indian Rupee representing 1:1 escrowed fiat in scheduled commercial banks.
  - USDT: 6-decimal regulated stablecoin utilized in GIFT City / IFSCA foreign investor settlement flows.
- **Gasless Claim Architecture:** Beneficiaries claim their accumulated balances either directly via standard transaction or gaslessly via ERC-4337 UserOperations. Alternatively, affiliates sign an EIP-712 typed digest (`ClaimPermit`), allowing an authorized relayer or paymaster to execute `claimRewardsWithPermit` on their behalf, with rewards transferred directly to the affiliate's registered wallet.
- **Multi-Sig Governance Integration:** Contract parameters (such as tier rates, cooldown durations, and paymaster authorizers) are governed by `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 Hardware Security Modules (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & Directory Layout:** Create Foundry project layout under `contracts/`:
   - `contracts/src/governance/AffiliateCommissionSplitter.sol`
   - `contracts/src/interfaces/IAffiliateCommissionSplitter.sol`
   - `test/governance/AffiliateCommissionSplitter.t.sol`
   - `test/mocks/MockERC20.sol`
   - `test/mocks/MockFeeCollector.sol`
   - `script/DeployAffiliateCommissionSplitter.s.sol`
2. **Define Data Structures, Enums, and Custom Errors:** Implement `IAffiliateCommissionSplitter.sol` specifying `AffiliateTier` (`STANDARD`, `BRONZE`, `SILVER`, `GOLD`, `PLATINUM`, `INSTITUTIONAL`), `CommissionType` (`DIRECT_REFERRAL`, `MASTER_AFFILIATE`, `VOLUME_REBATE`), `TierConfig`, `AffiliateProfile`, `RewardBalance`, `ClaimPermit`, and descriptive custom errors.
3. **Configure Storage Layout & ERC-7201 Namespaces:** Implement upgrade-safe storage layout using ERC-7201 standard to prevent storage collisions across proxy implementations:
   - Base slot: `keccak256(abi.encode(uint256(keccak256("growww.storage.AffiliateCommissionSplitter")) - 1)) & ~bytes32(uint256(0xff))`
4. **Implement Role-Based Access Control:** Integrate OpenZeppelin `AccessControlUpgradeable` configuring four enterprise roles:
   - `DEFAULT_ADMIN_ROLE`: Multi-sig governance root for contract upgrades and role administration.
   - `COMMISSION_OPERATOR_ROLE`: Authorized operational service (Fee Engine / Settlement DvP) permitted to submit commission splits.
   - `AFFILIATE_ADMIN_ROLE`: Growth operations role managing affiliate tier assignments, referral code registrations, and parameter updates.
   - `SURVEILLANCE_GUARDIAN_ROLE`: Automated market surveillance account capable of blacklisting collusive addresses and executing clawbacks.
5. **Implement Mathematical Calculation Engine:** Build fixed-point arithmetic routines utilizing OpenZeppelin `Math.mulDiv`:
   - `calculateReferralSplit(uint256 treasuryFeeAmount, uint16 tierBps)`: Calculates commission share with `Rounding.Floor`.
   - Enforce invariant: `tier0 bps (0.00% fee at launch) + tier2Bps <= MAX_AFFILIATE_SHARE_BPS` (where `MAX_AFFILIATE_SHARE_BPS = 3000`, representing 30% of Treasury fee).
6. **Implement Affiliate Registration & Profile Management:** Build `registerAffiliate` and `batchRegisterAffiliates` storing hashed referral codes, tier classification, beneficiary payout address, and salted identity hash (`panHash`). Verify that beneficiary address is non-zero and not already registered under another active profile.
7. **Implement Anti-Gaming & Self-Referral Validation:** Implement internal check `_validateReferralIntegrity`:
   - Reject transaction if `traderAddress == referrerAddress`.
   - Reject transaction if `traderPanHash == referrerPanHash` (anti-sybil identity matching).
   - Enforce velocity limit: assert that cumulative split volume for the referrer within the current block/epoch does not exceed `MAX_EPOCH_VOLUME_CAP`.
8. **Implement Single and Batched Commission Split Ingestion:** Implement `splitCommission` and `splitCommissionBatch` callable strictly by `COMMISSION_OPERATOR_ROLE`:
   - Verify input token is whitelisted (weINR or USDT).
   - Deduct commission amount from incoming Treasury fee allocation.
   - Credit beneficiary's escrowed reward balance with a timestamp lock (`unlockTimestamp = block.timestamp + COOLDOWN_PERIOD`).
   - Emit `CommissionSplitProcessed` and `RewardAccrued` events with zero PII.
9. **Implement Institutional Volume Rebate Logic:** Implement `creditVolumeRebate` to reward market-making desks and institutional participants based on monthly trailing volume thresholds verified against the exchange trade ledger.
10. **Implement Timelocked Escrow & Cooldown Processing:** Implement balance transition logic separating `pendingEscrowBalance` from `claimableBalance`. Rewards automatically transition to `claimableBalance` once `block.timestamp >= unlockTimestamp`.
11. **Implement Direct and Gasless Reward Claims:** 
    - Implement `claimRewards(address token)`: Transfers unlocked claimable balance to beneficiary via `SafeERC20.safeTransfer`.
    - Implement `claimRewardsWithPermit(...)`: Verifies EIP-712 signature over `ClaimPermit(address affiliate, address token, uint256 amount, uint256 nonce, uint256 deadline)` and transfers tokens to the beneficiary's registered payout address, sponsored by an authorized paymaster.
12. **Implement Surveillance Blacklisting & Forfeiture:** Build `freezeAndForfeitRewards` restricted to `SURVEILLANCE_GUARDIAN_ROLE`:
    - Freezes flagged affiliate account upon detection of wash trading.
    - Confiscates pending and claimable balances, transferring forfeited tokens directly to the Investor Protection Fund (IPF) address.
    - Emits `AffiliateBlacklisted` and `RewardsForfeited` audit events.
13. **Build Comprehensive Foundry Mock & Unit Test Harness:** Implement `MockERC20` (weINR/USDT) and `MockFeeCollector`. Write unit tests verifying registration, split calculations, tier transitions, cooldown expirations, and claim permits.
14. **Implement Fuzz & Invariant Testing Suite:** Write property-based tests asserting:
    - Zero token leakage: Total tokens deposited by Fee Collector strictly equals sum of all claimable balances + pending escrow + forfeited balances + claimed tokens.
    - Bound invariant: At no point can affiliate commission exceed 30% of the input Treasury fee.
    - Self-referral impossibility: Invariant fuzzer generates arbitrary address pairs and asserts rejection on identity match.
15. **Conduct Slither Static Analysis & Gas Profiling:** Execute Slither analysis ensuring zero high or medium findings. Profile gas costs to guarantee `splitCommission` consumes $< 65,000$ gas per trade and batched splits consume $< 35,000$ gas per allocation.

## Interfaces / Contracts

### 1. Affiliate Commission Splitter Interface (`IAffiliateCommissionSplitter.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IAffiliateCommissionSplitter
 * @author Growww National Blockchain Stock Exchange (NBSE)
 * @notice Interface for on-chain affiliate commission distribution, volume rebates,
 *         and gasless reward claims funded from Platform Treasury fee splits.
 */
interface IAffiliateCommissionSplitter {
    // --- Enums ---

    /// @notice Operational tier defining the affiliate revenue share rate
    enum AffiliateTier {
        STANDARD,        // Base tier: standard retail referrer
        BRONZE,          // Verified active referrer
        SILVER,          // Power affiliate / financial educator
        GOLD,            // Institutional partner / regional broker
        PLATINUM,        // Master affiliate network
        INSTITUTIONAL    // Designated market maker volume rebate tier
    }

    /// @notice Classification of commission revenue leg
    enum CommissionType {
        DIRECT_REFERRAL, // Tier 1: Direct introducing referrer
        MASTER_AFFILIATE,// Tier 2: Master partner / affiliate network
        VOLUME_REBATE    // Institutional liquidity provider / maker rebate
    }

    /// @notice Status of an affiliate account
    enum AccountStatus {
        ACTIVE,          // Fully operational and eligible for splits
        PAUSED,          // Temporarily paused for compliance review
        BLACKLISTED      // Permanently banned for wash trading or fraud
    }

    // --- Structs ---

    /// @notice Configuration parameters for an affiliate tier
    struct TierConfig {
        AffiliateTier tier;
        uint16 tier0 bps (0.00% fee at launch);             // Basis points of Treasury fee for Tier 1 (e.g., 1500 = 15%)
        uint16 tier2Bps;             // Basis points of Treasury fee for Tier 2 (e.g., 500 = 5%)
        uint32 minVolumeThreshold;   // Minimum 30-day volume in INR (paise) for qualification
        bool isActive;               // Whether this tier is currently active
    }

    /// @notice On-chain affiliate profile metadata (Zero PII)
    struct AffiliateProfile {
        address beneficiaryAddress;  // Registered payout destination wallet
        bytes32 referralCodeHash;    // keccak256 hash of promotional referral code
        bytes32 panHash;             // Salted identity hash for anti-sybil enforcement
        address masterAffiliate;     // Parent affiliate address for Tier 2 split (optional)
        AffiliateTier tier;          // Current active tier
        AccountStatus status;        // Current operational status
        uint64 registeredAt;         // Registration timestamp
        uint64 lastActiveAt;         // Timestamp of latest commission event
    }

    /// @notice Reward accounting ledger per asset
    struct RewardBalance {
        uint256 pendingEscrow;       // Balance locked under cooldown period
        uint256 claimable;           // Balance unlocked and available for withdrawal
        uint256 totalClaimed;        // Cumulative historical rewards claimed
        uint256 totalForfeited;      // Cumulative rewards forfeited due to violations
        uint64 lastUnlockTimestamp;  // Timestamp when pending escrow matures
    }

    /// @notice Input payload for processing a trade commission split
    struct SplitRequest {
        bytes32 tradeId;             // Unique identifier of settled trade
        address traderAddress;       // Wallet address of the executing trader
        bytes32 traderPanHash;       // Salted PAN hash of trader for anti-sybil check
        address referrerAddress;     // Direct introducing affiliate (Tier 1)
        address tokenAddress;        // Settlement asset (weINR or USDT)
        uint256 treasuryFeeAmount;   // Treasury reserve fee portion from trade execution
    }

    /// @notice EIP-712 Permit structure for gasless reward claims
    struct ClaimPermit {
        address affiliate;           // Target affiliate claiming rewards
        address token;               // Asset being claimed (weINR or USDT)
        uint256 amount;              // Amount to claim (0 for maximum claimable)
        uint256 nonce;               // Anti-replay nonce
        uint256 deadline;            // Unix expiration timestamp
    }

    // --- Events ---

    event AffiliateRegistered(
        address indexed affiliateAddress,
        address indexed beneficiaryAddress,
        bytes32 indexed referralCodeHash,
        address masterAffiliate,
        AffiliateTier tier
    );

    event AffiliateTierUpdated(
        address indexed affiliateAddress,
        AffiliateTier previousTier,
        AffiliateTier newTier,
        address updatedBy
    );

    event AffiliateStatusChanged(
        address indexed affiliateAddress,
        AccountStatus previousStatus,
        AccountStatus newStatus,
        bytes32 reasonHash,
        address updatedBy
    );

    event TierConfigUpdated(
        AffiliateTier indexed tier,
        uint16 tier0 bps (0.00% fee at launch),
        uint16 tier2Bps,
        uint32 minVolumeThreshold
    );

    event CommissionSplitProcessed(
        bytes32 indexed tradeId,
        address indexed tokenAddress,
        address indexed referrerTier1,
        address referrerTier2,
        uint256 tier1Amount,
        uint256 tier2Amount,
        uint256 retainedTreasuryAmount
    );

    event VolumeRebateCredited(
        address indexed institutionalDesk,
        address indexed tokenAddress,
        uint256 rebateAmount,
        uint32 monthlyVolumeINR
    );

    event RewardsClaimed(
        address indexed affiliateAddress,
        address indexed beneficiaryAddress,
        address indexed tokenAddress,
        uint256 amountClaimed,
        address claimedBy
    );

    event RewardsEscrowMatured(
        address indexed affiliateAddress,
        address indexed tokenAddress,
        uint256 unlockedAmount
    );

    event AffiliateBlacklisted(
        address indexed affiliateAddress,
        bytes32 reasonHash,
        address indexed guardian
    );

    event RewardsForfeited(
        address indexed affiliateAddress,
        address indexed tokenAddress,
        uint256 forfeitedAmount,
        address indexed ipfDestination
    );

    // --- Custom Errors ---

    error ZeroAddressDetected();
    error ZeroAmountDetected();
    error UnauthorizedCaller(address caller, bytes32 requiredRole);
    error AffiliateAlreadyRegistered(address affiliateAddress);
    error AffiliateNotRegistered(address affiliateAddress);
    error ReferralCodeCollision(bytes32 referralCodeHash);
    error InvalidTierConfiguration(uint16 totalBps, uint16 maxAllowedBps);
    error SelfReferralDetected(address trader, address referrer);
    error SybilIdentityDetected(bytes32 traderPanHash, bytes32 referrerPanHash);
    error AccountNotActive(address affiliateAddress, AccountStatus status);
    error TokenNotWhitelisted(address tokenAddress);
    error ExcessiveCommissionShare(uint256 requestedBps, uint256 maxBps);
    error InsufficientClaimableBalance(uint256 requested, uint256 available);
    error EscrowCooldownActive(uint64 matureTimestamp, uint64 currentTimestamp);
    error PermitExpired(uint256 deadline, uint256 currentTimestamp);
    error InvalidPermitSignature();
    error InvalidNonce(uint256 expectedNonce, uint256 providedNonce);
    error WashTradingVelocityExceeded(address affiliate, uint256 currentEpochVolume);
    error ArrayLengthMismatch();

    // --- View Methods ---

    function getAffiliateProfile(address affiliateAddress) external view returns (AffiliateProfile memory);
    function getTierConfig(AffiliateTier tier) external view returns (TierConfig memory);
    function getRewardBalance(address affiliateAddress, address tokenAddress) external view returns (RewardBalance memory);
    function getClaimableAmount(address affiliateAddress, address tokenAddress) external view returns (uint256);
    function isTokenWhitelisted(address tokenAddress) external view returns (bool);
    function getAffiliateNonce(address affiliateAddress) external view returns (uint256);

    // --- State-Modifying Methods ---

    function registerAffiliate(
        address affiliateAddress,
        address beneficiaryAddress,
        bytes32 referralCodeHash,
        bytes32 panHash,
        address masterAffiliate,
        AffiliateTier tier
    ) external;

    function setAffiliateTier(address affiliateAddress, AffiliateTier newTier) external;

    function setTierConfig(
        AffiliateTier tier,
        uint16 tier0 bps (0.00% fee at launch),
        uint16 tier2Bps,
        uint32 minVolumeThreshold
    ) external;

    function splitCommission(SplitRequest calldata request) external returns (uint256 totalCommissionDeducted);

    function splitCommissionBatch(SplitRequest[] calldata requests) external returns (uint256 totalCommissionsDeducted);

    function creditVolumeRebate(
        address deskAddress,
        address tokenAddress,
        uint256 rebateAmount,
        uint32 monthlyVolumeINR
    ) external;

    function unlockMaturedEscrow(address affiliateAddress, address tokenAddress) external returns (uint256 maturedAmount);

    function claimRewards(address tokenAddress, uint256 amount) external returns (uint256 claimedAmount);

    function claimRewardsWithPermit(
        ClaimPermit calldata permit,
        bytes calldata signature
    ) external returns (uint256 claimedAmount);

    function freezeAndForfeitRewards(address affiliateAddress, bytes32 reasonHash) external;
}
```

### 2. Integration Pattern with NBSE Fee Collector (`NBSEFeeCollector.sol` snippet)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {IAffiliateCommissionSplitter} from "./IAffiliateCommissionSplitter.sol";

/**
 * @title TreasuryFeeRoutingSnippet
 * @notice Demonstrates how the canonical Treasury reserve fee allocation interacts
 *         with the AffiliateCommissionSplitter during post-trade settlement.
 */
abstract contract TreasuryFeeRoutingSnippet {
    using SafeERC20 for IERC20;

    IAffiliateCommissionSplitter public affiliateSplitter;
    address public treasuryVault;
    address public coreSGF;
    address public ipfVault;

    // Platform Fee: 0 bps (0.00% fee at launch) (0.00% (Zero Fee)) of trade volume
    // Waterfall: Treasury reserve, Core SGF, Investor Protection Fund per FeeController governance
    uint16 public constant TREASURY_BPS = 6000;
    uint16 public constant SGF_BPS = 2500;
    uint16 public constant IPF_BPS = 1500;
    uint16 public constant BPS_DIVISOR = 10000;

    function routeTradeFee(
        bytes32 tradeId,
        address token,
        uint256 grossFeeAmount,
        address trader,
        bytes32 traderPanHash,
        address referrer
    ) internal {
        // 1. Calculate statutory splits
        uint256 sgfShare = (grossFeeAmount * SGF_BPS) / BPS_DIVISOR;
        uint256 ipfShare = (grossFeeAmount * IPF_BPS) / BPS_DIVISOR;
        uint256 treasuryShare = grossFeeAmount - sgfShare - ipfShare;

        // 2. Transfer statutory reserves immediately
        IERC20(token).safeTransfer(coreSGF, sgfShare);
        IERC20(token).safeTransfer(ipfVault, ipfShare);

        // 3. Process affiliate rebate from Treasury share if referrer is present
        uint256 affiliateDeduction = 0;
        if (referrer != address(0) && address(affiliateSplitter) != address(0)) {
            IERC20(token).forceApprove(address(affiliateSplitter), treasuryShare);

            IAffiliateCommissionSplitter.SplitRequest memory request = IAffiliateCommissionSplitter.SplitRequest({
                tradeId: tradeId,
                traderAddress: trader,
                traderPanHash: traderPanHash,
                referrerAddress: referrer,
                tokenAddress: token,
                treasuryFeeAmount: treasuryShare
            });

            affiliateDeduction = affiliateSplitter.splitCommission(request);
        }

        // 4. Send remaining net Treasury share to platform operational reserve
        uint256 netTreasuryAmount = treasuryShare - affiliateDeduction;
        if (netTreasuryAmount > 0) {
            IERC20(token).safeTransfer(treasuryVault, netTreasuryAmount);
        }
    }
}
```

## Security & Compliance Notes
- **Reentrancy Protection & Pull-Over-Push Accounting:**
  - The contract strictly enforces OpenZeppelin's `ReentrancyGuardUpgradeable` across all state-mutating functions.
  - To eliminate denial-of-service (DoS) and reentrancy attack vectors during high-throughput secondary market trade settlements, commission distribution follows the Pull-Over-Push design pattern. 
  - When `splitCommission` is called by the Fee Collector, tokens are transferred into the `AffiliateCommissionSplitter` vault and recorded in internal balances. No external calls or token transfers to affiliate wallets occur during trade settlement. Affiliates must actively invoke `claimRewards` or use `claimRewardsWithPermit` to withdraw funds.
- **Strict Invariant: Hard-Capped Referral Percentages & Statutory Isolation:**
  - Affiliate commissions are bounded by an immutable constant: `MAX_AFFILIATE_SHARE_BPS = 3000` (representing a hard ceiling of 30.00% of the Treasury fee portion).
  - Because the Treasury receives 60% of the 0.00% (Zero Fee) platform fee, the absolute maximum affiliate commission is strictly:
    $$\text{Max Commission} = \text{Trade Volume} \times \text{feeRate} \times \text{treasuryRate} \times \text{rebateRate} \quad (\text{0.00 at launch})$$
  - The contract mathematically guarantees that at no time can affiliate splits touch, dilute, or deduct from:
    1. Client trading principal or collateral margin.
    2. The Core Settlement Guarantee Fund (SGF, 25%).
    3. The Investor Protection Fund (IPF, 15%).
    4. Section 194S e-TDS withholding balances (Prompt 336).
- **Non-Transferable Reward Balances & Anti-Secondary Speculation:**
  - All accrued rewards are stored in internal ledger mappings (`mapping(address => mapping(address => RewardBalance))`).
  - No ERC-20 reward points, loyalty tokens, or tradeable debt vouchers are minted. 
  - Unclaimed reward balances cannot be transferred between addresses, pledged as loan collateral, or assigned to third parties. Tokens can only be withdrawn directly to the affiliate's registered beneficiary wallet upon verified KYC credential checks.
- **Anti-Gaming, Anti-Sybil, and Self-Referral Prevention:**
  - The contract programmatically enforces two layers of self-referral rejection:
    1. Direct address check: `traderAddress != referrerAddress`.
    2. Salted identity hash check: `traderPanHash != referrerPanHash`. This prevents traders from creating alternate Ethereum accounts to farm rebates on their own personal trades.
  - Rate limiting and volume velocity limits: If an affiliate's attributed volume exceeds `MAX_EPOCH_VOLUME_CAP` within a rolling window, excess splits are automatically diverted to pending escrow for manual surveillance review.
- **T+1 Settlement Cooldown & Fraud Clawback:**
  - Commission splits are credited to `pendingEscrow` with a mandatory cooldown period (`COOLDOWN_PERIOD = 86,400 seconds` / 24 hours), matching the NBSE T+1 equity settlement cycle.
  - If the Real-Time Market Surveillance Engine (Prompt 228) flags an affiliate for circular trading, wash trading, or market manipulation during this window, the `SURVEILLANCE_GUARDIAN_ROLE` can invoke `freezeAndForfeitRewards`.
  - Forfeited funds are routed directly to the Investor Protection Fund (`ipfDestination`), permanently depriving malicious actors of financial gain.
- **Gasless EIP-712 Claim Permit Security:**
  - Gasless claim operations utilize EIP-712 structured data hashing to prevent signature replay attacks across different chains or contract deployments:
    $$\text{DOMAIN\_SEPARATOR} = \text{keccak256}(\text{abi.encode}(\dots))$$
  - The permit struct enforces sequential nonces (`nonces[affiliate]++`) and an absolute deadline timestamp (`deadline >= block.timestamp`).
  - Even when submitted by a third-party relayer or paymaster, rewards are transferred strictly to the registered `beneficiaryAddress`, preventing fee theft or diversion by malicious relayers.
- **SEBI Advertising Code & Regulatory Compliance:**
  - In alignment with SEBI's Code of Conduct for Stock Brokers and Intermediaries, this contract does not implement multi-level marketing (MLM) pyramid structures. The commission topology is strictly restricted to two tiers: Tier 1 (direct introducing partner) and Tier 2 (licensed master affiliate / corporate distributor).
  - All tier configurations and maximum sharing ratios are auditable on-chain by regulatory authorities via public view methods.

## Acceptance Criteria
- [ ] Production Solidity contracts `AffiliateCommissionSplitter.sol` and `IAffiliateCommissionSplitter.sol` compiled under Solidity ^0.8.24 with zero compiler warnings and strict adherence to ERC-7201 upgradeable storage layout.
- [ ] Slither static analysis runs with zero high or medium severity warnings, verifying absence of unchecked external calls, reentrancy vulnerabilities, and integer precision loss.
- [ ] Multi-Tier Commission Logic verified: Contract correctly calculates and credits Tier 1 and Tier 2 splits using fixed-point `Math.mulDiv` at basis-point precision (`BPS_PRECISION = 10,000`).
- [ ] Fee Ceiling Invariant enforced: Any attempt by governance or operators to configure or execute splits exceeding `MAX_AFFILIATE_SHARE_BPS` (3,000 bps / 30% of Treasury fee) reverts with `ExcessiveCommissionShare`.
- [ ] Dual-Token Compatibility verified: Contract correctly handles both 18-decimal weINR and 6-decimal USDT token transfers and internal balance accounting without precision truncation.
- [ ] Anti-Gaming Safeguards verified:
  - Direct self-referrals (`traderAddress == referrerAddress`) revert with `SelfReferralDetected`.
  - Sybil self-referrals (`traderPanHash == referrerPanHash`) revert with `SybilIdentityDetected`.
- [ ] Cooldown Escrow Lifecycle verified: Rewards credited to `pendingEscrow` remain locked until `block.timestamp >= unlockTimestamp`, transitioning to `claimable` upon maturity.
- [ ] Gasless Claim Permitted: Affiliates can withdraw unlocked balances via `claimRewardsWithPermit` using an EIP-712 signature; permits revert on expired deadline or invalid nonce.
- [ ] Malicious Account Blacklisting: `SURVEILLANCE_GUARDIAN_ROLE` can freeze flagged accounts and forfeit pending/claimable balances to the Investor Protection Fund.
- [ ] Invariant Fuzz Testing: Foundry invariant test suite runs $\ge 10,000$ iterations asserting that total assets held by the contract equal the exact sum of all pending, claimable, and forfeited balances across all assets (Zero-Balance Leakage Invariant).
- [ ] Gas Efficiency Benchmark: Single trade split execution (`splitCommission`) consumes $< 65,000$ gas on Hyperledger Besu; batched split execution consumes $< 35,000$ gas per allocation.
- [ ] Foundry test harness achieves $\ge 95\%$ line and branch test coverage across all core execution paths, custom error reverts, and boundary conditions.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301` (Permissioned Blockchain Platform Evaluation & Selection - Hyperledger Besu)
  - Prompt `302` (Network Topology & QBFT Validator Infrastructure Setup)
  - Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract - SettlementDvP.sol)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
  - Prompt `329` (NBSE Delivery-versus-Payment Settlement & Automated Fee Collector)
  - Prompt `006` (Fee Model Specification: Fixed Fee Structure & Treasury Allocation)
- **Parallel Tasks:**
  - Prompt `271` (Affiliate Referral Engine & Commission Pipeline)
  - Prompt `210` (Transaction Fee & Realized PnL Calculation Engine)
  - Prompt `224` (Referral & Growth Service - SEBI Advertising Code Compliant)
  - Prompt `259` (ERC-4337 Account Abstraction Bundler & Gasless Paymaster Service)
  - Prompt `337` (ERC-4337 Modular Smart Account & Ephemeral Session Keys Contract)
  - Prompt `309` (Blockchain Event Indexing & Query Service)
- **Subsequent Prompts Enabled:**
  - Prompt `315` (On-Chain Settlement Guarantee Fund & Default Waterfall Smart Contract)
  - Prompt `336` (Automated Tax Withholding and e-TDS Compliance Ledger Smart Contract)
  - Prompt `515` (Mobile Affiliate & Referral Growth Dashboard)
  - Prompt `612` (Web Institutional Trading Desk & Volume Rebate Portal)
  - Prompt `708` (Market Manipulation & Anti-Wash Trading Detection Pipeline)
  - Prompt `811` (Testnet vs Mainnet Dual Environment CI/CD Automation)
