# 345 - Fixed-Income Tokenized Yield & Staking Vault Smart Contract (Solidity)

## Purpose
In traditional fixed-income markets, retail and institutional investors seeking yield from sovereign debt (Government Securities - G-Secs, Treasury Bills - T-Bills, State Development Loans - SDLs) and high-grade corporate bonds face substantial operational friction. Coupon payments are disbursed semi-annually or annually into fiat bank accounts, requiring manual reconciliation, active reinvestment to capture compounding returns, and fragmented tax accounting for Tax Deducted at Source (TDS). Furthermore, illiquidity in secondary bond markets often traps investor capital until bond maturity.

On the Growww National Blockchain Stock Exchange (NBSE) deployed on Hyperledger Besu (QBFT consensus), tokenized debt securities (represented via ERC-3643 compliant security tokens such as `TokenizedGSec.sol`) possess 1:1 backing in RBI Constituent Subsidiary General Ledger (CSGL) accounts or NSDL/CDSL depository demat accounts. To unlock institutional-grade yield optimization for all market participants, this prompt specifies the smart contract architecture for **Fixed-Income Tokenized Yield & Staking Vaults (`BondYieldVault.sol` and `YieldDistributor.sol`)** under `contracts/src/yield/`.

Governed by Architectural Decision Record **ADR-0040 (Crypto Earn & PoS Staking Yield Vaults)** and **ADR-0005 (ERC-3643 Permissioned Security Tokens)**, these vaults implement the standardized **ERC-4626 Tokenized Vault Standard**, adapted to operate within a regulated, permissioned financial ecosystem. The vault suite aggregates underlying fixed-income debt tokens, automatically compounds periodic coupon cash flows into vault share appreciation, supports fixed-term staking tiers with boosted yields (e.g., 30-day, 90-day, 180-day, 365-day locks), enforces early withdrawal penalty mechanics, eliminates ERC-4626 first-depositor inflation attacks using virtual share offsets, and verifies automated TDS tax withholding before net yield distribution.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite and Foundry test architecture under `contracts/src/yield/` comprising:
- `BondYieldVault.sol`: Core ERC-4626 tokenized vault adapted for ERC-3643 permissioned fixed-income debt tokens (`TokenizedGSec.sol`, `TokenizedCorporateBond.sol`). It manages share-to-asset conversion, deposit/mint and withdraw/redeem lifecycles, fixed-term staking lockup tiers, early withdrawal penalties, and virtual share offsets ($10^3$ virtual shares and $10^3$ virtual assets) to prevent first-depositor inflation attacks.
- `YieldDistributor.sol`: Continuous yield collection and streaming engine. It ingests periodic coupon cash flows (in e-Rupee CBDC or INR fiat stable escrow) from `CouponDistributor.sol` (Prompt 333) and primary bond issuers, applies statutory Section 193/194A TDS withholding verification, streams net interest into the vault reserve, and updates the vault exchange rate monotonically.
- High-Precision Q64.96 Fixed-Point Math Engine (`FullMath.sol`): Mathematical module executing continuous compounding and discrete periodic share-to-asset calculations without precision loss or numerical overflow.
- `IBondYieldVault.sol` and `IYieldDistributor.sol`: Comprehensive Solidity interfaces specifying all data structures, staking tier configurations, custom errors, events, view functions, and state-altering entrypoints.
- Universal Flat Fee Integration: Enforces Growww's canonical 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform transaction fee on vault share transfers and secondary redemptions, splitting proceeds Platform Treasury, Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF) per FeeController governance.
- Comprehensive Foundry Test Harness (`test/yield/BondYieldVault.t.sol`, `test/yield/YieldDistributor.t.sol`): Invariant fuzzing, inflation attack simulation, multi-tier lockup expiration, TDS tax withholding assertions, and reentrancy resistance tests.

## Scope Boundaries
- **In Scope:**
  - Implementation of ERC-4626 tokenized vault standard adapted for ERC-3643 permissioned underlying assets.
  - Multi-tier staking lockups: Flexible (0-day), 30-day, 90-day, 180-day, and 365-day fixed terms with yield bonus multipliers.
  - Early withdrawal penalty engine: Configurable penalty schedules for premature lockup exits, transferring forfeited interest to vault reserves and platform insurance pools.
  - Virtual share and asset offset ($10^3$ virtual shares, $10^3$ virtual assets) mitigating ERC-4626 first-depositor share inflation attacks.
  - Continuous yield streaming and daily coupon ingestion from `CouponDistributor.sol`.
  - Share-to-asset and asset-to-share mathematical conversions with explicit rounding direction control (favoring the vault against arbitrageurs).
  - Integration with `IIdentityRegistry` (Prompt 303 / Prompt 305) to enforce KYC/AML compliance on vault deposits, share transfers, and redemptions.
  - Statutory TDS tax withholding verification hooks ensuring tax compliance before net yield distribution.
  - 0.00% (No fee at all) platform fee assessment on secondary turnover with standard 0.00% fee at launch (governed by FeeController.sol) revenue split.
  - Circuit breaker and emergency halt integration with `CircuitBreakerHalt.sol` (Prompt 341).
  - Role-Based Access Control (RBAC) via OpenZeppelin `AccessControlUpgradeable` and ERC-1967 UUPS upgradeability.
- **Out of Scope / Handled Elsewhere:**
  - Real-time yield curve generation, Nelson-Siegel-Svensson fitting, and dirty price calculations (handled in Prompt 254 Bond Yield Curve & Dirty Price Calculator Service).
  - Physical depository custody settlement at RBI CSGL or NSDL/CDSL (handled in Prompt 213 Custodian Depository Integration Service).
  - Central bank Digital Rupee (e₹) CBDC payment rail adapters (handled in Prompt 232 / Prompt 212).
  - Off-chain tax statement generation, PAN-level Form 16A/26AS generation (handled in Prompt 223 Tax Reporting & Statement Service).
  - Secondary order matching and CLOB trading of vault receipt tokens (handled in Prompt 205 Order Matching Engine).
  - Proof of solvency zk-SNARK aggregation (handled in Prompt 317 ZK Proof of Solvency Verifier).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD`, and custom errors).
- **Token & Vault Standards:**
  - **ERC-4626:** Yield-Bearing Tokenized Vault Standard adapted for permissioned digital debt securities.
  - **ERC-3643 (T-REX):** Permissioned security token interface ensuring on-chain identity verification via `IIdentityRegistry` and transfer compliance via `ICompliance`.
  - **ERC-20:** Standard fungible share token interface for vault receipt tokens.
- **Fixed-Point Arithmetic:** FullMath Q64.96 and OpenZeppelin `Math.mulDiv` with explicit `Rounding.Floor` and `Rounding.Ceil` modes to eliminate precision truncation and prevent rounding arbitrage.
- **Security & Upgradeability Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, property-based fuzzing, and invariant testing; `cast` for Besu JSON-RPC contract interactions).
- **Static Analysis & Verification:** Slither, Mythril, and Solhint automated CI analysis pipelines.
- **Consortium Ledger:** Hyperledger Besu enterprise blockchain running QBFT consensus with 2-second block finality, 1:1 custodial backing, and zero on-chain PII.

## Backend / Infra Touchpoints
- **Bond Yield Curve & Dirty Price Calculator Service (Prompt 254):** Ingests vault share pricing, underlying bond dirty prices, and historical yield streaming data to publish real-time Annual Percentage Yield (APY) metrics and duration metrics.
- **Custodian Depository Integration Service (Prompt 213):** Reconciles underlying tokenized G-Sec and bond balances in the vault against official RBI CSGL and NSDL/CDSL demat holding statements.
- **Fee and Realized PnL Engine (Prompt 210):** Audits and records the 0.00% (Zero Fee) platform transaction fee allocations (Treasury reserve, Core SGF, Investor Protection Fund per FeeController governance) and early exit penalty deductions.
- **Corporate Actions Service (Prompt 222) & Coupon Distributor (Prompt 333):** Notifies `YieldDistributor.sol` when periodic coupon disbursements occur, orchestrating the deposit of interest funds into the vault asset pool.
- **Tax Reporting & Statement Service (Prompt 223):** Ingests on-chain `TdsDeducted` events, mapping pseudonymous investor wallet addresses to off-chain PAN records for quarterly 26Q TDS returns and Form 16A generation.
- **Blockchain Event Indexer (Prompt 309):** Listens for `DepositWithLockup`, `WithdrawWithLockup`, `YieldHarvested`, and `EarlyExitPenaltyAssessed` events, updating off-chain PostgreSQL portfolio ledgers within 50ms.
- **Circuit Breaker & Emergency Halt Contract (Prompt 341):** Monitors single-ISIN and sector-level circuit filters, pausing vault deposits and redemptions in the event of credit rating downgrades, default notices, or market-wide circuit breaker trips.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Deployment Topology:** Deployed as ERC-1967 UUPS upgradeable proxies on Hyperledger Besu enterprise consortium network utilizing QBFT consensus with 2-second deterministic finality.
- **Continuous Compound Interest Accrual:**
  Underlying bonds generate periodic coupon cash flows. `YieldDistributor.sol` accepts coupon payments and streams them linearly or compounds them discretely into `BondYieldVault.sol`. The total asset backing grows over time while share supply remains constant (until new deposits/withdrawals occur), causing the share-to-asset exchange rate to increase continuously:
  $$\text{Exchange Rate}(t) = \frac{\text{TotalAssets}(t) + 10^3}{\text{TotalShares}(t) + 10^3}$$
- **Share-to-Asset Conversion Math (ERC-4626 with Virtual Offset):**
  To completely eradicate the first-depositor inflation attack (where an attacker deposits 1 wei of asset, donates assets directly to the vault, and dilutes subsequent depositors due to integer division truncation), the vault maintains virtual shares and virtual assets offsets ($10^3$):
  $$\text{Shares to Mint} = \frac{\text{Deposited Assets} \times (\text{TotalShares} + 10^3)}{\text{TotalAssets} + 10^3}$$
  $$\text{Assets to Redeem} = \frac{\text{Burned Shares} \times (\text{TotalAssets} + 10^3)}{\text{TotalShares} + 10^3}$$
  - Conversions for `deposit` and `mint` round down shares minted, while conversions for `withdraw` and `redeem` round up shares burned or round down assets paid out, strictly favoring the vault.
- **Fixed-Term Lockup Tiers & Penalty Mechanics:**
  Investors choose a staking tier upon deposit:
  - `FLEXIBLE` (0-day lock): Standard base yield, instant zero-penalty redemption.
  - `TIER_30_DAYS`: 30-day lock, 1.05x yield multiplier, 1.00% early withdrawal penalty on principal.
  - `TIER_90_DAYS`: 90-day lock, 1.15x yield multiplier, 2.00% early withdrawal penalty on principal.
  - `TIER_180_DAYS`: 180-day lock, 1.25x yield multiplier, 3.50% early withdrawal penalty on principal.
  - `TIER_365_DAYS`: 365-day lock, 1.40x yield multiplier, 5.00% early withdrawal penalty on principal.
  If an investor exits prior to the lockup expiration timestamp, the early exit penalty is assessed. The penalty is deducted from redeemed principal: 50% is added directly to vault reserves (benefiting remaining locked stakers) and 50% is routed to the Core Settlement Guarantee Fund (SGF).
- **Zero On-Chain PII Guarantee:**
  All staking positions, claim records, and TDS withholdings reference strictly pseudonymous Ethereum addresses (`address indexed user`), position IDs (`bytes32 positionId`), and security identifiers (`bytes32 isinHash`). No investor names, PAN numbers, or physical bank accounts are recorded on-chain.
- **Compliance Gating via ERC-3643:**
  Deposits, withdrawals, and share transfers invoke `IIdentityRegistry.isVerified(user)` on the underlying security token's identity registry, ensuring that only KYC-cleared investors can hold vault shares or receive yield.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & Directory Scaffolding:** Set up Foundry directory structure under `contracts/src/yield/`, `contracts/src/interfaces/yield/`, `contracts/src/libraries/`, `test/yield/`, and `script/yield/`.
2. **Implement Q64.96 Fixed-Point Math Library:** Create `contracts/src/libraries/FullMath.sol` providing high-precision 512-bit intermediate multiplication and division (`mulDiv`, `mulDivRoundingUp`) to support continuous interest calculations without overflow.
3. **Define Interfaces & Data Structures:** Implement `contracts/src/interfaces/yield/IBondYieldVault.sol` and `IYieldDistributor.sol` specifying enums (`LockupTier`, `VaultState`), structs (`StakePosition`, `TierConfig`, `TdsDeductionRecord`), custom errors, and events.
4. **Implement Base Upgradeable ERC-4626 Architecture:** Build `BondYieldVault.sol` inheriting from OpenZeppelin `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, and `ERC20Upgradeable`.
5. **Incorporate Virtual Shares Offset Defense:** Configure the internal asset-to-share and share-to-asset conversion logic in `BondYieldVault.sol` to enforce an immutable $10^3$ virtual shares and $10^3$ virtual assets offset:
   - `_convertToShares(uint256 assets, Math.Rounding rounding)`
   - `_convertToAssets(uint256 shares, Math.Rounding rounding)`
6. **Implement Multi-Tier Staking Lockup Storage:** Implement a position registry `mapping(bytes32 => StakePosition)` and `mapping(address => bytes32[])` tracking individual user staking positions, locked shares, deposit timestamp, unlock timestamp, tier identifier, and yield boost weighting.
7. **Implement Compliant Deposit & Mint Functions:** Build `depositWithLockup(uint256 assets, address receiver, LockupTier tier)` and `mintWithLockup(uint256 shares, address receiver, LockupTier tier)`:
   - Verify caller and receiver via `IIdentityRegistry(identityRegistry).isVerified(receiver)`.
   - Verify underlying asset transfer via `SafeERC20.safeTransferFrom`.
   - Calculate shares using virtual offset.
   - Record `StakePosition` with deterministic position ID: `keccak256(abi.encode(receiver, tier, block.timestamp, nonce))`.
   - Emit `DepositWithLockup` event.
8. **Implement Early Exit Penalty & Redemption Mechanics:** Build `withdrawWithLockup(bytes32 positionId, uint256 assets, address receiver, address owner)` and `redeemWithLockup(bytes32 positionId, uint256 shares, address receiver, address owner)`:
   - Validate caller is owner or approved spender.
   - Check if `block.timestamp < position.unlockTimestamp`.
   - If prematurely exited, calculate penalty based on `TierConfig.penaltyBps`. Deduct penalty: route 50% to vault reserve and 50% to Core SGF.
   - Burn vault shares, transfer net assets to receiver, and update position state.
   - Emit `WithdrawWithLockup` and `EarlyExitPenaltyAssessed` events.
9. **Implement Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Collection:** On secondary vault share transfers or redemptions, assess the canonical 0.00% (Zero Fee) platform transaction fee on turnover, splitting proceeds Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance via `ITokenizedGSecFeeCollector.sol`.
10. **Implement Yield Distributor Contract:** Build `contracts/src/yield/YieldDistributor.sol`:
    - Store authorized disburser roles and target vault addresses.
    - Implement `depositCouponYield(bytes32 isinHash, uint256 grossAmount, uint256 tdsWithheldAmount)` callable by `CouponDistributor.sol` (Prompt 333) or authorized treasury relayers.
    - Validate that `grossAmount >= tdsWithheldAmount`.
    - Record `TdsDeductionRecord` for audit and indexer ingestion.
    - Transfer net coupon assets into `BondYieldVault.sol` via `vault.harvestYield(netAmount)`.
11. **Implement Continuous Yield Streaming Buffer:** In `YieldDistributor.sol`, implement linear yield streaming over an epoch duration (e.g., 24-hour linear release) to prevent sudden flash-deposit sandwich attacks around coupon payout blocks.
12. **Integrate Circuit Breaker & Emergency Controls:** Implement `whenNotHalted` modifiers connected to `ICircuitBreakerHalt(circuitBreaker).isHalted(isinHash)` (Prompt 341). Allow `EMERGENCY_GUARDIAN_ROLE` to pause deposits/withdrawals during extreme volatility or depository reconciliation pauses.
13. **Write Comprehensive Foundry Unit & Fuzz Tests:**
    - `test/yield/BondYieldVault.t.sol`: Verify ERC-4626 roundtrip conversions (`convertToShares(convertToAssets(x)) == x`).
    - Test inflation attack resilience: Simulate 1-wei deposit followed by 1,000,000 token donation; assert subsequent depositor receives fair share allocation.
    - Test lockup tier maturation: Verify zero penalty after unlock timestamp, and exact penalty deduction prior to unlock timestamp.
    - Test TDS deduction balance conservation: Gross coupon equals Net yield + TDS withheld.
14. **Write Stateful Invariant Fuzzing Suites:** Invariant test that `vault.totalAssets()` strictly equals the underlying token balance held in the vault contract, and that total active position shares equal `vault.totalSupply()`.
15. **Run Static Analysis & Deployment Script:** Execute Slither analysis ensuring zero high or medium findings. Construct `DeployBondYieldVault.s.sol` configuring multi-sig admin roles for Hyperledger Besu deployment.

## Interfaces / Contracts

### 1. Bond Yield Vault Interface (`IBondYieldVault.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/interfaces/IERC4626.sol";

/// @title IBondYieldVault
/// @notice Interface for Fixed-Income Tokenized Yield & Staking Vaults adapted for ERC-3643 compliant debt tokens
interface IBondYieldVault is IERC4626 {
    /// @notice Staking lockup tiers offering yield boosts and early exit penalties
    enum LockupTier {
        FLEXIBLE,       // 0-day lock: base yield, zero penalty
        TIER_30_DAYS,   // 30-day lock: 1.05x boost, 1.00% early exit penalty
        TIER_90_DAYS,   // 90-day lock: 1.15x boost, 2.00% early exit penalty
        TIER_180_DAYS,  // 180-day lock: 1.25x boost, 3.50% early exit penalty
        TIER_365_DAYS   // 365-day lock: 1.40x boost, 5.00% early exit penalty
    }

    /// @notice Operational lifecycle state of the vault
    enum VaultState {
        ACTIVE,
        DEPOSITS_PAUSED,
        EMERGENCY_HALTED,
        MATURED_CLOSED
    }

    /// @notice Configuration parameters for each staking tier
    struct TierConfig {
        LockupTier tier;
        uint32 durationSeconds;     // Lockup duration in seconds
        uint16 yieldMultiplierBps;  // Multiplier in basis points (10000 = 1.0x, 10500 = 1.05x)
        uint16 penaltyBps;          // Early withdrawal penalty in basis points (100 = 1.0%)
        bool isActive;              // Whether tier is open for new deposits
    }

    /// @notice Individual user staking position record
    struct StakePosition {
        bytes32 positionId;         // Unique position identifier
        address owner;              // Owner wallet address
        LockupTier tier;            // Selected staking tier
        uint256 lockedShares;       // Amount of vault shares locked
        uint256 assetAmountAtStake; // Nominal asset amount deposited
        uint64 depositTimestamp;    // Timestamp when deposit occurred
        uint64 unlockTimestamp;     // Expiration timestamp of lockup
        bool isWithdrawn;           // Whether position has been closed
    }

    /// @notice Audit record of tax withholding during yield distribution
    struct TdsAuditRecord {
        bytes32 distributionId;
        bytes32 isinHash;
        uint256 grossYieldPaise;
        uint256 netYieldPaise;
        uint256 tdsWithheldPaise;
        uint64 timestamp;
    }

    // Events
    event DepositWithLockup(
        bytes32 indexed positionId,
        address indexed caller,
        address indexed owner,
        uint256 assets,
        uint256 shares,
        LockupTier tier,
        uint64 unlockTimestamp
    );

    event WithdrawWithLockup(
        bytes32 indexed positionId,
        address indexed receiver,
        address indexed owner,
        uint256 assets,
        uint256 shares,
        bool earlyExit
    );

    event EarlyExitPenaltyAssessed(
        bytes32 indexed positionId,
        address indexed owner,
        uint256 grossAssets,
        uint256 penaltyAssets,
        uint256 netAssetsToUser,
        uint256 reserveShare,
        uint256 sgfShare
    );

    event YieldHarvested(
        address indexed caller,
        uint256 grossAssets,
        uint256 netAssetsAdded,
        uint256 tdsWithheld,
        uint256 newTotalAssets
    );

    event TierConfigUpdated(
        LockupTier indexed tier,
        uint32 durationSeconds,
        uint16 yieldMultiplierBps,
        uint16 penaltyBps,
        bool isActive
    );

    event VaultStateChanged(VaultState indexed oldState, VaultState indexed newState);

    // Custom Errors
    error LockupPeriodStillActive(bytes32 positionId, uint64 currentTimestamp, uint64 unlockTimestamp);
    error PositionAlreadyClosed(bytes32 positionId);
    error PositionNotFound(bytes32 positionId);
    error InvalidLockupTier(LockupTier tier);
    error IdentityNotVerified(address user);
    error ComplianceCheckFailed(address user, uint256 amount);
    error InsufficientVaultBalance(uint256 available, uint256 requested);
    error ZeroDepositNotAllowed();
    error ExceedsMaxVaultCapacity(uint256 currentAssets, uint256 maxCapacity);
    error CircuitBreakerActive(bytes32 isinHash);
    error UnauthorizedCaller(address caller);
    error MathRoundingError();

    // View Functions
    function underlyingISIN() external view returns (bytes32);
    function identityRegistry() external view returns (address);
    function circuitBreaker() external view returns (address);
    function getTierConfig(LockupTier tier) external view returns (TierConfig memory);
    function getPosition(bytes32 positionId) external view returns (StakePosition memory);
    function getUserPositions(address user) external view returns (bytes32[] memory);
    function virtualOffset() external pure returns (uint256 virtualShares, uint256 virtualAssets);
    function calculateEarlyExitPenalty(bytes32 positionId) external view returns (uint256 penaltyAmount, uint256 netAmount);

    // State-Altering Functions
    function depositWithLockup(
        uint256 assets,
        address receiver,
        LockupTier tier
    ) external returns (bytes32 positionId, uint256 shares);

    function mintWithLockup(
        uint256 shares,
        address receiver,
        LockupTier tier
    ) external returns (bytes32 positionId, uint256 assets);

    function withdrawWithLockup(
        bytes32 positionId,
        uint256 assets,
        address receiver,
        address owner
    ) external returns (uint256 shares);

    function redeemWithLockup(
        bytes32 positionId,
        uint256 shares,
        address receiver,
        address owner
    ) external returns (uint256 assets);

    function harvestYield(
        uint256 netAssetsAdded,
        uint256 tdsWithheld,
        bytes32 distributionId
    ) external;

    function setTierConfig(
        LockupTier tier,
        uint32 durationSeconds,
        uint16 yieldMultiplierBps,
        uint16 penaltyBps,
        bool isActive
    ) external;

    function setVaultState(VaultState newState) external;
}
```

### 2. Yield Distributor Interface (`IYieldDistributor.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// @title IYieldDistributor
/// @notice Interface for automated coupon collection and yield streaming into tokenized vaults
interface IYieldDistributor {
    struct CouponDistributionBatch {
        bytes32 distributionId;
        bytes32 isinHash;
        address targetVault;
        uint256 grossAmountPaise;
        uint256 tdsRateBps;
        uint256 tdsWithheldPaise;
        uint256 netAmountPaise;
        uint64 streamingEpochSeconds;
        uint64 distributionTimestamp;
        bool isProcessed;
    }

    event CouponYieldIngested(
        bytes32 indexed distributionId,
        bytes32 indexed isinHash,
        address indexed targetVault,
        uint256 grossAmountPaise,
        uint256 tdsWithheldPaise,
        uint256 netAmountPaise
    );

    event YieldStreamedToVault(
        bytes32 indexed distributionId,
        address indexed targetVault,
        uint256 amountStreamed,
        uint256 remainingToStream
    );

    error DistributionAlreadyProcessed(bytes32 distributionId);
    error InvalidTdsAmount(uint256 grossAmount, uint256 tdsAmount);
    error TargetVaultMismatch(address provided, address expected);
    error UnauthorizedDisburser(address caller);

    function ingestCouponDistribution(
        bytes32 distributionId,
        bytes32 isinHash,
        address targetVault,
        uint256 grossAmountPaise,
        uint256 tdsRateBps,
        uint64 streamingEpochSeconds
    ) external returns (uint256 netYieldPaise);

    function processYieldStream(bytes32 distributionId) external returns (uint256 streamedAmount);

    function getDistributionBatch(bytes32 distributionId) external view returns (CouponDistributionBatch memory);
}
```

## Security & Compliance Notes
- **Prevention of ERC-4626 Inflation Attacks (Virtual Offset):** Standard ERC-4626 implementations are susceptible to first-depositor inflation attacks, where an attacker deposits 1 wei of assets, receives 1 share, and subsequently donates a massive amount of assets to the vault directly. Subsequent depositors who deposit standard sums suffer 100% rounding losses due to integer division truncation. `BondYieldVault.sol` eliminates this attack vector by incorporating $10^3$ virtual shares and $10^3$ virtual assets into all conversion calculations, ensuring that share value cannot be artificially inflated.
- **Section 193 / 194A TDS Withholding Compliance:** In accordance with the Indian Income Tax Act (1961), interest payments on corporate debt and sovereign securities disbursed to resident and non-resident investors are subject to statutory Tax Deducted at Source (TDS). The vault infrastructure verifies that the gross coupon payment is segregated into net yield (credited to vault assets) and TDS withheld (credited to the government tax withholding escrow ledger), emitting immutable audit events for reconciliation with Form 26AS/16A.
- **1:1 Custodial Asset Backing Invariant:** Vault assets consist strictly of tokenized debt securities (`TokenizedGSec.sol`) or cash settlement assets holding certified 1:1 custody in RBI CSGL or NSDL/CDSL accounts. The contract enforces that total vault assets never exceed audited depository limits.
- **Identity & Transfer Compliance via ERC-3643:** All interactions (`deposit`, `mint`, `transfer`, `withdraw`, `redeem`) execute synchronous verification calls against `IIdentityRegistry(identityRegistry).isVerified(investorAddress)`. Non-KYC or sanctioned addresses cannot deposit assets or receive vault shares.
- **Reentrancy Protection & Checks-Effects-Interactions:** All asset movements and share burns follow the strict Checks-Effects-Interactions pattern and are protected by OpenZeppelin's `ReentrancyGuardUpgradeable`.
- **Circuit Breaker Integration:** The vault natively queries `CircuitBreakerHalt.sol` (Prompt 341) via `whenNotHalted(underlyingISIN)`. If a market-wide or single-ISIN circuit breaker trips, all vault deposits and redemptions are halted instantly to protect investor capital.

## Acceptance Criteria
- [ ] `IBondYieldVault.sol` and `IYieldDistributor.sol` compile cleanly with Solidity `^0.8.24` with zero warnings.
- [ ] ERC-4626 implementation fully passes standard ERC-4626 property tests, correctly maintaining conversion mathematical symmetries.
- [ ] Virtual share and asset offset ($10^3$) successfully withstands simulated first-depositor inflation attacks in Foundry fuzz tests with zero asset dilution.
- [ ] Fixed-term staking tiers (`FLEXIBLE`, `TIER_30_DAYS`, `TIER_90_DAYS`, `TIER_180_DAYS`, `TIER_365_DAYS`) accurately track unlock timestamps and enforce exact early withdrawal penalties.
- [ ] Early withdrawal penalties are mathematically split: exactly 50% retained in vault reserves for locked holders and 50% routed to the Core SGF.
- [ ] Non-KYC or unverified addresses are unconditionally rejected by `depositWithLockup` and `mintWithLockup` via `IdentityRegistry`.
- [ ] Daily coupon ingestion through `YieldDistributor.sol` correctly applies TDS deductions without balance leakage.
- [ ] 0.00% (No fee at all) platform fee is accurately assessed on secondary turnover and routed Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] Circuit breaker halts trigger instantaneous deposit and withdrawal freezes across affected ISIN vaults.
- [ ] 100% Foundry test coverage achieved across unit, fuzz, and invariant test suites.
- [ ] Slither static analysis returns zero high, medium, or reentrancy issues.
- [ ] Document strictly complies with all 12 mandatory sections, containing zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `303` (Asset Token Issuance Smart Contract - ERC-3643).
  - Prompt `305` (Transfer Compliance Hooks Smart Contract).
  - Prompt `307` (MultiSig Governance Smart Contract).
  - Prompt `333` (Tokenized G-Sec Bonds & Coupon Accrual Smart Contracts).
  - Prompt `341` (On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract).
- **Parallel Tasks:**
  - Prompt `254` (Bond Yield Curve & Dirty Price Calculator Service).
  - Prompt `213` (Custodian Depository Integration Service).
  - Prompt `210` (Fee and Realized PnL Engine).
  - Prompt `222` (Corporate Actions Service).
- **Subsequent Prompts Enabled:**
  - Prompt `531` (Flutter Sovereign Debt & Fixed-Income Staking Portal).
  - Prompt `609` (Fixed-Income Earn Vaults Web Dashboard).
  - Prompt `346` (Cross-Chain Yield Aggregator Bridge).
