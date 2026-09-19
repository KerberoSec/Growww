# 344 - RWA Dutch Auction & On-Chain Linear Vesting Smart Contract (Solidity)

## Purpose
Primary issuance of Real-World Asset (RWA) tokens (such as tokenized commercial real estate, corporate debt debentures, infrastructure yields, and fractional private equity) on the Growww National Blockchain Stock Exchange (NBSE) requires deterministic, fair, and manipulation-resistant price discovery. Traditional English auctions, fixed-price initial offerings, and off-chain order matching suffer from gas wars, front-running, predatory bot sniping, winner's curse, and artificial illiquidity. Furthermore, unchecked secondary market trading immediately following token generation events (TGE) often leads to speculative volatility and rapid dumping by large initial allocators, destabilizing market confidence and impairing long-term capital formation.

To resolve these primary market challenges, this specification defines the smart contract architecture for the **RWA Dutch Auction & On-Chain Linear Vesting Smart Contract Suite (`DutchAuction.sol`, `LinearVestingVault.sol`, `IDutchAuction.sol`, `ILinearVestingVault.sol`)** deployed on Hyperledger Besu under Architectural Decision Record ADR-0044 and Operational Runbook RUNBOOK-32.

The Dutch auction model employs a descending price schedule where the unit price decreases linearly over time from a predefined start price ($P_{start}$) down to a reserve floor price ($P_{reserve}$). Investors place capital bids at any point during the auction window. The auction terminates when total cumulative demand equals or exceeds the token supply available for sale, or when the auction timeframe expires. Crucially, the auction executes a **uniform clearing price mechanism**: all winning bidders pay the exact same final clearing price ($P_{clear}$), eliminating the winner's curse and incentivizing bidders to commit bids at their true valuation without fear of overpaying. Any surplus capital deposited by bidders whose bid price exceeded the clearing price is immediately unlocked for pro-rata refund.

Upon auction settlement, allocated tokens are not disbursed in a single vulnerable lump-sum. Instead, they are programmatically transferred directly into the `LinearVestingVault.sol` streaming smart contract. The vault streams token claims second-by-second based on continuous block timestamps (`block.timestamp`), enforcing custom cliff durations, linear vesting schedules, and compliance whitelisting via ERC-3643 transfer hooks. The system integrates natively with the ERC-4337 / EIP-2771 Account Abstraction Paymaster (Prompt 259), guaranteeing gasless bidding and claim interactions for retail and institutional investors. The smart contracts strictly enforce the statutory 200-investor private placement cap mandated under Section 42 of the Indian Companies Act, 2013, and assess Growww's canonical 0.00% (Zero Fee) platform fee with strictly 0.00% fees at launch (future fee adjustments governed by FeeController.sol).

## What You Are Building
A production-grade, modular, upgradeable Solidity smart contract suite located under `contracts/src/launchpad/` and `contracts/interfaces/launchpad/` comprising:
- `DutchAuction.sol`: Enterprise primary market Dutch auction contract operating on Hyperledger Besu. Manages auction initialization, linear descending price calculation, deposit escrow, uniform clearing price determination, over-subscription pro-rata allocation, and automated refund claims.
- `LinearVestingVault.sol`: Continuous streaming token vault managing on-chain second-by-second linear vesting schedules. Supports custom cliff windows, revocable and irrevocable schedules, gasless claims via meta-transactions, and administrative synchronization per RUNBOOK-32.
- `IDutchAuction.sol`: Formal Solidity interface declaring auction configurations, state enums, bid records, clearing data structures, custom errors, events, and external function signatures.
- `ILinearVestingVault.sol`: Formal Solidity interface specifying vesting schedules, streaming claim records, revocability parameters, custom errors, events, and external function signatures.
- Mathematical Price Decay & Pro-Rata Scaling Engine: High-precision fixed-point arithmetic module implementing continuous price reduction and proportional token allocation with zero rounding leakage or integer overflow risk.
- Companies Act 2013 Private Placement Cap Enforcer: On-chain investor count tracker restricting total unique participating bidders to a maximum of 200 qualified investors per auction instance.
- ERC-3643 Permissioned Transfer Compliance Integration: Native compliance checking against `ICompliance` and `IIdentityRegistry` hooks on both bidding deposits and streaming claim releases.
- EIP-2771 / ERC-4337 Gasless Paymaster Compatibility: Built-in trusted forwarder support enabling end users to sign bids and claim vested tokens with zero native gas tokens (ETH/BESU) required.
- Platform Fee & Multi-Vault Splitter: Automatic 0.00% (Zero Fee) (0.00% fee) fee calculation on gross auction proceeds, with dynamic routing governed by FeeController.sol (0.00% at launch).
- Comprehensive Foundry Test Suite (`test/launchpad/DutchAuction.t.sol`, `test/launchpad/LinearVestingVault.t.sol`): Full test coverage with fuzzing, multi-bidder clearing invariants, edge-case cliff calculations, and gas profiling.

## Scope Boundaries
- **In Scope:**
  - Solidity ^0.8.24 contracts `DutchAuction.sol` and `LinearVestingVault.sol` implementing interfaces `IDutchAuction.sol` and `ILinearVestingVault.sol`.
  - Linear descending price decay formulation:
    $$P(t) = P_{start} - \frac{(P_{start} - P_{reserve}) \times (t - t_{start})}{t_{end} - t_{start}}$$
  - Uniform clearing price calculation: determining the single price where cumulative committed demand equals or exceeds total offered supply.
  - Pro-rata over-subscription scaling for marginal tier bids and automated excess capital refunding.
  - Second-by-second streaming linear vesting calculation:
    $$Vested(t) = TotalAllocation \times \frac{t - t_{start}}{t_{end} - t_{start}} \quad (\text{for } t \ge t_{cliff})$$
  - On-chain private placement investor cap validation (maximum 200 unique bidders per private placement offering under Section 42 Companies Act 2013).
  - Escrow and accounting of payment tokens (`paymentToken`, e.g., eINR, USDC, or authorized stablecoins) and offered RWA tokens (`tokenToSell`, ERC-20 / ERC-3643).
  - Gasless meta-transaction execution via `ERC2771ContextUpgradeable` compatible with Growww Paymaster (Prompt 259).
  - Administrative control, emergency pause (`PausableUpgradeable`), and proxy upgradeability (`UUPSUpgradeable`).
  - Strict fee deduction of 0.00% (Zero Fee) on total cleared capital and split across Treasury (60%), Core SGF (25%), and IPF (15%).
- **Out of Scope / Handled Elsewhere:**
  - Off-chain legal prospectus preparation, credit rating publishing, and real-world asset origination (handled in Launchpad Engine, Prompt 266).
  - User identity verification, KYC/AML attestation, and ONCHAINID claim issuance (handled in Prompt 202, Prompt 305, and Prompt 703).
  - Depository reconciliation with physical asset deeds, debenture trustees, or NSDL/CDSL custody ledgers (handled in Depository Adapter, Prompt 213).
  - Off-chain UserOperation bundling, mempool sequencing, and paymaster gas treasury sponsorship (handled in Prompt 259).
  - Secondary market CLOB matching and off-hours synthetic AMM liquidity pools (handled in Prompt 205 and Prompt 339).
  - Fiat banking on/off-ramps and payment gateway processing (handled in Prompt 212).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD` for reentrancy guards, and custom errors for minimal bytecode footprint).
- **Base Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `ERC2771ContextUpgradeable`).
  - OpenZeppelin `SafeERC20` for reliable ERC-20 token transfers across non-standard implementations.
  - OpenZeppelin `Math.sol` for full 512-bit intermediate multiplication and division (`mulDiv`) with explicit rounding modes (`Rounding.Floor`, `Rounding.Ceil`).
- **Security Token Standards:**
  - ERC-3643 (T-REX) Permissioned Token Standard (`IERC3643`, `IIdentityRegistry`, `ICompliance`).
- **Development & Testing Toolchain:**
  - **Foundry** (`forge` for compilation, property-based fuzz testing, and gas benchmarks; `cast` for Besu JSON-RPC interaction).
  - **Static Analysis:** Slither, Mythril, and Solhint linters integrated into automated verification pipelines.

## Backend / Infra Touchpoints
- **Launchpad Engine (Prompt 266):**
  - Primary microservice coordinating issuer onboarding, offering prospectus approval, and auction deployment.
  - Submits on-chain transaction to invoke `createAuction()` on `DutchAuction.sol` with certified offering parameters.
  - Monitors auction lifecycle states and triggers `finalizeAuction()` when clearing conditions are met.
- **Account Abstraction Bundler & Paymaster Service (Prompt 259):**
  - Ingests user meta-transactions for bids and vesting claims, bundles them into `UserOperation` structs, and sponsors gas fees on Hyperledger Besu.
  - Monitors Paymaster gas balance and pending transaction queues to prevent distribution stalls per RUNBOOK-32.
- **Custodian Depository Integration Service (Prompt 213):**
  - Verifies that underlying real-world assets (e.g., land registry title deeds, commercial mortgage notes, corporate debentures) are legally locked in demat or trustee custody prior to auction creation.
  - Verifies 1:1 asset backing before the issuer is permitted to transfer `tokenToSell` into `DutchAuction.sol` escrow.
- **Blockchain Event Indexer (Prompt 309):**
  - Listens to emitted events: `AuctionCreated`, `BidPlaced`, `AuctionCleared`, `RefundClaimed`, `VestingScheduleCreated`, and `TokensClaimed`.
  - Ingests event logs into PostgreSQL database with sub-50ms latency, powering real-time web and mobile investor dashboards.
- **Settlement Guarantee Fund & Treasury Vaults (Prompt 210 / Prompt 315):**
  - Operates with 0.00% fees at launch under the Universal Zero-Fee Model.
- **Automated Tax Withholding & eTDS Ledger (Prompt 336):**
  - Captures on-chain clearing events to record investor allotment prices, cost basis, and applicable stamp duty / tax withholding commitments.

## Blockchain Interaction (Permissioned Hyperledger Besu Ledger with 1:1 Custody Backing, Zero PII, QBFT)
- **Consensus & Finality:** Deployed on the Growww NBSE Hyperledger Besu permissioned consortium network operating QBFT consensus with 2-second deterministic block times and instant single-block transaction finality (zero chain reorganizations or uncle blocks).
- **Atomic Fund Deposit & Escrow:** When placing a bid, the investor's payment tokens (`paymentToken`, e.g., eINR or authorized stablecoins) are pulled atomically into `DutchAuction.sol` escrow via `SafeERC20.safeTransferFrom`. Funds remain locked until auction clearing or cancellation.
- **Uniform Clearing Price Mechanism:** Bidders specify the maximum price they are willing to pay ($P_{bid}$) and the capital amount deposited. As time advances and price decays, the clearing engine determines the precise clearing price ($P_{clear}$) at which total accumulated demand equals total available supply. All successful bidders acquire tokens at exactly $P_{clear}$, regardless of whether their bid was entered at a higher price.
- **Second-by-Second Streaming Linear Vesting:** Upon auction finalization, token allocations are created inside `LinearVestingVault.sol`. The vault dynamically computes claimable tokens on-the-fly using:
  $$\text{claimable} = \frac{\text{totalAmount} \times (\text{block.timestamp} - \text{startTime})}{\text{endTime} - \text{startTime}} - \text{claimedAmount}$$
  Tokens stream continuously every single second. Investors can claim any accrued tokens at any time without waiting for periodic batch unlock cycles.
- **Zero On-Chain PII Invariant:** The smart contracts store strictly pseudonymous data: Ethereum addresses, token balances, auction IDs, timestamps, and cryptographic proof hashes. No personal identifiable information (PAN, Aadhaar, name, email, residential address) is ever written to storage or emitted in event logs.
- **Multi-Signature Governance:** Administrative functions (deploying offerings, updating safety parameters, pausing auctions, or updating trusted forwarders) require 3-of-5 multi-signature authorization (`MultiSigGovernance.sol`, Prompt 307) backed by FIPS 140-2 Level 3 HSM hardware keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:**
   - Create directories `contracts/src/launchpad/`, `contracts/interfaces/launchpad/`, and `test/launchpad/`.
   - Scaffold files: `contracts/interfaces/launchpad/IDutchAuction.sol`, `contracts/interfaces/launchpad/ILinearVestingVault.sol`, `contracts/src/launchpad/DutchAuction.sol`, and `contracts/src/launchpad/LinearVestingVault.sol`.
2. **Declare Complete Interface Definitions (`IDutchAuction.sol`, `ILinearVestingVault.sol`):**
   - Define all lifecycle enums: `AuctionState` (`PENDING`, `ACTIVE`, `CLEARED`, `CANCELLED`, `FINALIZED`), `VestingState` (`ACTIVE`, `REVOKED`, `COMPLETED`).
   - Define core data structures: `AuctionConfig`, `Bid`, `ClearingResult`, `VestingSchedule`, `FeeDistribution`.
   - Declare all custom errors (avoiding generic string reverts) and typed event schemas.
3. **Initialize Base Contracts & Upgradeability Setup:**
   - Inherit `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, and `ERC2771ContextUpgradeable` across both contracts.
   - Configure role constants: `DEFAULT_ADMIN_ROLE`, `LAUNCHPAD_ADMIN_ROLE`, `OPERATOR_ROLE`, and `EMERGENCY_ROLE`.
   - Initialize trusted forwarder address for EIP-2771 meta-transaction support.
4. **Implement Linear Price Decay Calculation:**
   - Implement `getCurrentPrice(uint256 auctionId)` returning the instantaneous token price in payment token units.
   - Use `Math.mulDiv` to ensure full precision and zero overflow:
     $$\Delta P = \frac{(P_{start} - P_{reserve}) \times (\text{block.timestamp} - t_{start})}{t_{end} - t_{start}}$$
     $$P(t) = P_{start} - \Delta P$$
   - Ensure price stays strictly at $P_{reserve}$ if $\text{block.timestamp} \ge t_{end}$ and auction has not yet cleared.
5. **Implement Auction Creation & Asset Escrow (`createAuction`):**
   - Validate auction parameters: $P_{start} > P_{reserve} > 0$, $t_{start} \ge \text{block.timestamp}$, $t_{end} > t_{start}$, `totalTokens > 0`.
   - Transfer `totalTokens` of `tokenToSell` from issuer into contract escrow using `SafeERC20.safeTransferFrom`.
   - If `tokenToSell` is an ERC-3643 token, verify compliance checks pass for the escrow deposit.
   - Initialize auction state to `ACTIVE` (or `PENDING` if $t_{start} > \text{block.timestamp}$) and emit `AuctionCreated`.
6. **Implement Bidding Engine with Statutory 200-Investor Cap (`placeBid`):**
   - Restrict bids to active auction window ($\text{block.timestamp} \in [t_{start}, t_{end}]$).
   - Enforce investor accreditation: verify bidder holds valid ERC-3643 identity claim if compliance is enabled.
   - Enforce Section 42 Companies Act 2013 cap: maintain `investorCount`; if `msg.sender` has not bid previously in this auction, verify `investorCount < 200`. Increment counter on new investor.
   - Transfer payment tokens from bidder into contract escrow. Record `Bid` struct and emit `BidPlaced`.
7. **Implement Uniform Price Clearing Algorithm (`finalizeAuction`):**
   - Check termination condition: either $\text{block.timestamp} \ge t_{end}$, or total committed demand at current price meets or exceeds total tokens offered.
   - Determine uniform clearing price $P_{clear}$:
     - If auction expires without reaching supply cap, $P_{clear} = P_{reserve}$.
     - If auction cleared during active decay, calculate the exact price where cumulative demand matches supply.
   - Calculate total gross proceeds in payment tokens: $\text{grossProceeds} = \text{clearedTokens} \times P_{clear}$.
   - Transition auction state to `CLEARED`.
8. **Implement Platform Fee Deduction & Revenue Split:**
   - Calculate Growww 0.00% (Zero Fee) platform fee: $\text{feeAmount} = (\text{grossProceeds} \times 1) / 10000$.
   - Route per FeeController governance (0.00% at launch).
   - Transfer fee portions to designated vault addresses and remit remaining net proceeds to the asset issuer.
   - Emit `PlatformFeeDistributed`.
9. **Implement Pro-Rata Allocation & Refund Claims (`claimRefundAndVesting`):**
   - For each bidder, compute final token allocation:
     - If bidder's maximum bid price $< P_{clear}$, bidder was unsuccessful; allocation is 0, full refund of deposit is unlocked.
     - If bidder's maximum bid price $\ge P_{clear}$, bidder is successful; token allocation is calculated at $P_{clear}$.
     - If marginal clearing tier is over-subscribed, apply pro-rata scaling factor.
   - Calculate excess capital refund: $\text{refund} = \text{depositedAmount} - (\text{allocatedTokens} \times P_{clear})$.
   - If refund $> 0$, transfer payment tokens back to bidder.
   - Deposit `allocatedTokens` directly into `LinearVestingVault.sol` for the bidder.
10. **Implement Linear Vesting Vault Schedule Setup (`createVestingSchedule`):**
    - Callable by `DutchAuction.sol` or authorized `LAUNCHPAD_ADMIN_ROLE`.
    - Accept parameters: `beneficiary`, `token`, `totalAmount`, `startTime`, `cliffDuration`, `vestingDuration`, `revocable`.
    - Validate parameters: `totalAmount > 0`, `vestingDuration > 0`, `cliffDuration <= vestingDuration`.
    - Store `VestingSchedule` record keyed by `bytes32 scheduleId = keccak256(abi.encode(beneficiary, token, startTime))`.
    - Escrow RWA tokens in `LinearVestingVault.sol` and emit `VestingScheduleCreated`.
11. **Implement Second-by-Second Streaming Token Claim (`claimVestedTokens`):**
    - Calculate vested token entitlement based on current `block.timestamp`:
      - If $\text{block.timestamp} < \text{startTime} + \text{cliffDuration}$, vested amount is 0.
      - If $\text{block.timestamp} \ge \text{startTime} + \text{vestingDuration}$, vested amount is 100% of `totalAmount`.
      - Otherwise, $\text{vested} = \text{Math.mulDiv}(\text{totalAmount}, \text{block.timestamp} - \text{startTime}, \text{vestingDuration})$.
    - Determine claimable amount: $\text{claimable} = \text{vested} - \text{alreadyClaimed}$.
    - Require `claimable > 0`.
    - Update `alreadyClaimed += claimable`.
    - Transfer `claimable` tokens to beneficiary using `SafeERC20.safeTransfer`.
    - Check ERC-3643 compliance hooks to ensure beneficiary remains authorized to receive securities tokens.
    - Emit `TokensClaimed`.
12. **Implement Revocability & Recovery Operations (RUNBOOK-32 Compliance):**
    - Implement `revokeVestingSchedule(bytes32 scheduleId)` restricted to multi-sig governance.
    - If schedule is revocable, compute vested amount up to revocation timestamp, pay out unreleased vested tokens to beneficiary, and return unvested balance to issuer/treasury.
    - Implement `syncVestingSchedule(bytes32 scheduleId)` allowing administrative state synchronization in the event of relayer stalls per RUNBOOK-32.
13. **Construct Comprehensive Foundry Unit & Fuzz Tests (`test/launchpad/`):**
    - Unit tests: single-bidder clear, multiple-bidder uniform clearing, price decay precision, cliff enforcement, second-by-second linear release.
    - Invariant test: sum of all token allocations plus unsold tokens must exactly equal original escrowed tokens.
    - Invariant test: sum of all refunds plus gross cleared proceeds must exactly equal total deposited payment tokens.
    - Fuzz test: arbitrary bid amounts, timing offsets, and decay intervals asserting zero arithmetic panics.
14. **Construct Boundary & Cap Constraint Tests:**
    - Test Companies Act 200-investor cap: verify that 200 unique bidders succeed and the 201st distinct bidder reverts with `InvestorCapExceeded`.
    - Test over-subscription pro-rata distribution with non-divisible token quantities, asserting remainder dust conservation.
    - Test reentrancy attack vectors on `placeBid`, `claimRefundAndVesting`, and `claimVestedTokens`.
15. **Execute Static Analysis & Gas Benchmarking:**
    - Execute Slither and Mythril static analysis; verify zero high or medium severity findings.
    - Profile gas costs for `placeBid` and `claimVestedTokens` to ensure smooth execution within Besu gas limits and Paymaster subsidies.

## Interfaces / Contracts

### 1. Dutch Auction Interface (`contracts/interfaces/launchpad/IDutchAuction.sol`)

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IDutchAuction
 * @notice Interface for the RWA Dutch Auction Smart Contract on Hyperledger Besu.
 * @dev Manages primary market descending price discovery, uniform clearing, and over-subscription pro-rata refunds.
 */
interface IDutchAuction {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum AuctionState {
        PENDING,     // Auction created, waiting for start timestamp
        ACTIVE,      // Auction is currently open for bids
        CLEARED,     // Auction cleared and finalized at uniform price
        CANCELLED    // Auction cancelled by issuer or admin prior to clearing
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct AuctionConfig {
        address issuer;               // Asset issuer offering tokens
        address tokenToSell;          // Address of ERC-20 / ERC-3643 RWA token
        address paymentToken;         // Address of settlement token (e.g. eINR, stablecoin)
        address vestingVault;         // LinearVestingVault contract address
        uint256 totalTokens;          // Total quantity of tokens offered for sale
        uint256 startPrice;           // Starting auction price (in payment token units)
        uint256 reservePrice;         // Floor price below which auction will not clear
        uint64 startTime;             // Unix timestamp for auction opening
        uint64 endTime;               // Unix timestamp for auction expiration
        uint64 cliffDuration;         // Post-sale vesting cliff in seconds
        uint64 vestingDuration;       // Post-sale linear vesting duration in seconds
        bool isPrivatePlacement;      // If true, strictly enforces 200-investor cap
    }

    struct Bid {
        address bidder;               // Address of the participating investor
        uint256 depositedAmount;      // Total payment tokens deposited
        uint256 maxPrice;             // Maximum price bidder is willing to accept
        uint256 allocatedTokens;      // Tokens allocated after clearing
        uint256 refundAmount;         // Excess payment tokens eligible for refund
        bool claimed;                 // True if refund and vesting allocation processed
    }

    struct ClearingResult {
        uint256 clearingPrice;        // Uniform price paid by all successful bidders
        uint256 totalTokensSold;      // Quantity of tokens cleared
        uint256 grossProceeds;        // Total capital raised in payment tokens
        uint256 platformFee;          // 0.00% fee (No fee at all) deducted from gross proceeds
        uint64 clearedTimestamp;      // Block timestamp when auction cleared
    }

    struct FeeDistribution {
        uint256 totalFee;
        uint256 treasuryShare;        // Governed by FeeController (0.00% at launch)
        uint256 coreSgfShare;         // Governed by FeeController (0.00% at launch)
        uint256 ipfShare;             // Governed by FeeController (0.00% at launch)
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event AuctionCreated(
        uint256 indexed auctionId,
        address indexed issuer,
        address indexed tokenToSell,
        uint256 totalTokens,
        uint256 startPrice,
        uint256 reservePrice,
        uint64 startTime,
        uint64 endTime
    );

    event BidPlaced(
        uint256 indexed auctionId,
        address indexed bidder,
        uint256 depositedAmount,
        uint256 maxPrice,
        uint256 totalInvestorCount
    );

    event AuctionCleared(
        uint256 indexed auctionId,
        uint256 clearingPrice,
        uint256 totalTokensSold,
        uint256 grossProceeds
    );

    event AuctionCancelled(uint256 indexed auctionId, string reason);

    event RefundAndVestingClaimed(
        uint256 indexed auctionId,
        address indexed bidder,
        uint256 refundAmount,
        uint256 allocatedTokens,
        bytes32 vestingScheduleId
    );

    event PlatformFeeDistributed(
        uint256 indexed auctionId,
        uint256 totalFee,
        uint256 treasuryAmount,
        uint256 coreSgfAmount,
        uint256 ipfAmount
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error AuctionNotPending();
    error AuctionNotActive();
    error AuctionAlreadyCleared();
    error AuctionNotCleared();
    error InvalidAuctionParameters();
    error InvalidPriceCurve();
    error InvalidTimeWindow();
    error BidAmountZero();
    error BidPriceBelowCurrentPrice(uint256 bidPrice, uint256 currentPrice);
    error InvestorCapExceeded(uint256 currentCount, uint256 maxCap);
    error UnauthorizedCaller();
    error TransferFailed();
    error AlreadyClaimed();
    error NoBidFound();
    error ComplianceCheckFailed(address account);

    // -------------------------------------------------------------------------
    // External Functions
    // -------------------------------------------------------------------------

    function createAuction(AuctionConfig calldata config) external returns (uint256 auctionId);
    function placeBid(uint256 auctionId, uint256 paymentAmount, uint256 maxPrice) external;
    function finalizeAuction(uint256 auctionId) external;
    function cancelAuction(uint256 auctionId, string calldata reason) external;
    function claimRefundAndVesting(uint256 auctionId, address bidder) external returns (uint256 refund, uint256 tokens);
    function getCurrentPrice(uint256 auctionId) external view returns (uint256 price);
    function getAuctionConfig(uint256 auctionId) external view returns (AuctionConfig memory);
    function getClearingResult(uint256 auctionId) external view returns (ClearingResult memory);
    function getBid(uint256 auctionId, address bidder) external view returns (Bid memory);
    function getInvestorCount(uint256 auctionId) external view returns (uint256 count);
}
```

### 2. Linear Vesting Vault Interface (`contracts/interfaces/launchpad/ILinearVestingVault.sol`)

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ILinearVestingVault
 * @notice Interface for the On-Chain Second-by-Second Streaming Linear Vesting Vault on Hyperledger Besu.
 * @dev Manages continuous token streaming unlocks with cliff periods, revocability controls, and gasless claim integration.
 */
interface ILinearVestingVault {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum VestingState {
        ACTIVE,       // Schedule is streaming tokens normally
        REVOKED,      // Schedule was revoked by authorized administrator
        COMPLETED     // All tokens have been fully claimed
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct VestingSchedule {
        bytes32 scheduleId;           // Unique schedule identifier
        address beneficiary;          // Recipient of vested tokens
        address token;                // ERC-20 / ERC-3643 asset token address
        uint256 totalAmount;          // Total tokens subject to vesting schedule
        uint256 claimedAmount;        // Total tokens claimed to date
        uint64 startTime;             // Unix timestamp when vesting calculation begins
        uint64 cliffDuration;         // Duration in seconds before first tokens unlock
        uint64 vestingDuration;       // Total duration in seconds from startTime to 100% unlock
        bool revocable;               // Whether schedule can be revoked by governance
        VestingState state;           // Current lifecycle state of the schedule
    }

    struct VestingParameters {
        address beneficiary;
        address token;
        uint256 totalAmount;
        uint64 startTime;
        uint64 cliffDuration;
        uint64 vestingDuration;
        bool revocable;
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event VestingScheduleCreated(
        bytes32 indexed scheduleId,
        address indexed beneficiary,
        address indexed token,
        uint256 totalAmount,
        uint64 startTime,
        uint64 cliffDuration,
        uint64 vestingDuration,
        bool revocable
    );

    event TokensClaimed(
        bytes32 indexed scheduleId,
        address indexed beneficiary,
        address indexed token,
        uint256 amountClaimed,
        uint256 remainingVestingBalance
    );

    event VestingScheduleRevoked(
        bytes32 indexed scheduleId,
        address indexed beneficiary,
        uint256 unvestedAmountRefunded,
        uint256 vestedAmountPaidOut
    );

    event VestingScheduleSynchronized(bytes32 indexed scheduleId, uint256 verifiedClaimable);

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error ScheduleNotFound();
    error ScheduleNotActive();
    error CliffNotReached(uint64 currentTime, uint64 cliffTime);
    error NoTokensDue();
    error ScheduleNotRevocable();
    error InvalidVestingParameters();
    error ZeroAddress();
    error ZeroAmount();
    error TransferFailed();
    error ComplianceCheckFailed(address account);

    // -------------------------------------------------------------------------
    // External Functions
    // -------------------------------------------------------------------------

    function createVestingSchedule(VestingParameters calldata params) external returns (bytes32 scheduleId);
    function claimVestedTokens(bytes32 scheduleId) external returns (uint256 claimed);
    function claimAllVested(address token) external returns (uint256 totalClaimed);
    function revokeVestingSchedule(bytes32 scheduleId) external;
    function syncVestingSchedule(bytes32 scheduleId) external returns (uint256 claimable);
    function computeReleasableAmount(bytes32 scheduleId) external view returns (uint256 releasable);
    function getVestingSchedule(bytes32 scheduleId) external view returns (VestingSchedule memory);
    function getBeneficiarySchedules(address beneficiary) external view returns (bytes32[] memory);
}
```

### 3. Implementation Contract Overview (`contracts/src/launchpad/DutchAuction.sol` & `LinearVestingVault.sol`)

The implementation architecture delivers:
- **Inherited OpenZeppelin Modules:**
  - `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `ERC2771ContextUpgradeable`.
- **DutchAuction Storage Elements:**
  - `mapping(uint256 => AuctionConfig) public auctions;`
  - `mapping(uint256 => mapping(address => Bid)) public bids;`
  - `mapping(uint256 => address[]) internal _auctionBidders;`
  - `mapping(uint256 => ClearingResult) public clearingResults;`
  - `uint256 public auctionCounter;`
  - `uint256 public constant MAX_PRIVATE_INVESTORS = 200;` (Section 42 Companies Act 2013).
  - `address public treasuryVault;`
  - `address public coreSgfVault;`
  - `address public ipfVault;`
- **LinearVestingVault Storage Elements:**
  - `mapping(bytes32 => VestingSchedule) private _schedules;`
  - `mapping(address => bytes32[]) private _beneficiaryScheduleIds;`
  - `mapping(address => mapping(address => uint256)) private _totalTokensVested;`
  - `address public trustedForwarder;`

## Security & Compliance Notes

### 1. Linear Price Decay & High-Precision Arithmetic
- **Arithmetic Safety:** The descending price formula incorporates integer division over time. To avoid precision truncation and intermediate multiplication overflow, price calculations utilize OpenZeppelin `Math.mulDiv`:
  $$\Delta P = \text{Math.mulDiv}(P_{start} - P_{reserve}, \text{block.timestamp} - t_{start}, t_{end} - t_{start})$$
- **Monotonicity Invariant:** For any two timestamps $t_1 < t_2 \in [t_{start}, t_{end}]$, the smart contract guarantees $P(t_1) \ge P(t_2) \ge P_{reserve}$. The price curve never increases and never decays below $P_{reserve}$.

### 2. Uniform Clearing Price & MEV / Front-Running Mitigation
- **Fair Valuation Incentive:** In a uniform price Dutch auction, all allocated bidders pay $P_{clear}$. Bidders have a dominant strategy to bid their personal reserve valuation early, knowing that if the clearing price ends lower, they will receive a pro-rata refund for the surplus capital.
- **Bot Front-Running Neutralization:** Because the price decays predictably and all winning participants receive the identical clearing price, priority gas auctions (PGA) and sandwich MEV bots gain zero economic advantage over genuine retail or institutional participants.

### 3. Indian Companies Act 2013 Section 42 Private Placement Cap (200 Investors)
- Under Section 42 of the Companies Act, 2013 (read with Rule 14 of the Companies (Prospectus and Allotment of Securities) Rules, 2014), any private placement offering of securities to 200 or more persons in a financial year (excluding Qualified Institutional Buyers and employees under ESOP) is statutorily deemed to be a public offer.
- `DutchAuction.sol` enforces this requirement on-chain:
  ```solidity
  if (config.isPrivatePlacement && bids[auctionId][bidder].depositedAmount == 0) {
      if (_auctionBidders[auctionId].length >= MAX_PRIVATE_INVESTORS) {
          revert InvestorCapExceeded(_auctionBidders[auctionId].length, MAX_PRIVATE_INVESTORS);
      }
      _auctionBidders[auctionId].push(bidder);
  }
  ```
  If a 201st distinct investor attempts to deposit funds, the transaction reverts immediately, protecting the issuer from statutory non-compliance.

### 4. Continuous Second-by-Second Streaming Vesting vs Lump-Sum Unlocks
- Traditional milestone vesting unlocks substantial token batches at 30-day or 90-day intervals, triggering predictable market sell-offs and localized liquidity crunches.
- `LinearVestingVault.sol` computes token release continuously:
  $$\text{Vested}(t) = \text{TotalTokens} \times \frac{t - t_{start}}{t_{end} - t_{start}}$$
  Tokens unlock continuously per block. This eliminates discrete unlock volatility cliffs and supports smooth, continuous institutional treasury streaming.

### 5. ERC-3643 Permissioned Token Compliance & Sanctions Screening
- Both `DutchAuction.sol` and `LinearVestingVault.sol` interact directly with ERC-3643 compliant asset tokens.
- When `placeBid` is invoked, or when `claimVestedTokens` transfers tokens to a beneficiary, the contract verifies with the asset's `ICompliance` and `IIdentityRegistry` contracts that the investor's ONCHAINID identity claims are valid, non-expired, and non-sanctioned. Transfers to non-compliant or blacklisted addresses revert automatically.

### 6. RUNBOOK-32 Mitigation: Paymaster Gas Stalls & Administrative Sync
- In high-load conditions, EIP-2771 / ERC-4337 Paymaster relayers may experience nonce desynchronization or gas pool exhaustion, preventing users from submitting gasless claims.
- In compliance with `RUNBOOK-32` (`RBK-OPS-032`), `LinearVestingVault.sol` provides an administrative fallback `syncVestingSchedule(bytes32 scheduleId)`. If relayer infrastructure fails, the operator or beneficiary can initiate direct fallback synchronization or manual transaction submission without corrupting stream state.

### 7. Reentrancy Protection and Checks-Effects-Interactions (CEI)
- All fund-handling functions (`placeBid`, `claimRefundAndVesting`, `claimVestedTokens`, `revokeVestingSchedule`) strictly adhere to the Checks-Effects-Interactions (CEI) design pattern.
- State variables (such as `claimedAmount`, `bids[].claimed`, and auction states) are updated in storage prior to invoking external token transfers.
- Native transient storage reentrancy guards (`ReentrancyGuardUpgradeable`) prevent recursive reentrancy attacks across arbitrary ERC-20 token hooks.

### 8. Growww Universal Platform Fee Integrity
- The contract automatically levies Growww's canonical 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on gross cleared auction capital.
- Split distribution is immutably enforced on-chain:
  - Treasury Vault (`treasuryVault`) per FeeController governance
  - 25% to Core Settlement Guarantee Fund (`coreSgfVault`)
  - 15% to Investor Protection Fund (`ipfVault`)
- Unsuccessful bids refunded in full incur exactly 0% platform fee.

## Acceptance Criteria
- [ ] `IDutchAuction.sol`, `ILinearVestingVault.sol`, `DutchAuction.sol`, and `LinearVestingVault.sol` compile cleanly under Solidity ^0.8.24 with zero compiler warnings.
- [ ] Linear price decay function computes exact values across all block timestamps and never drops below `reservePrice`.
- [ ] High-precision arithmetic utilizes `Math.mulDiv` and prevents integer overflow and rounding errors.
- [ ] Section 42 Companies Act 2013 private placement cap is strictly enforced: exactly 200 distinct bidders can participate, and the 201st distinct bidder reverts with `InvestorCapExceeded`.
- [ ] Uniform clearing price calculates accurately when demand equals supply, and all winning bidders are charged the identical clearing price $P_{clear}$.
- [ ] Over-subscribed auctions correctly compute pro-rata token allocations and release excess deposited capital as refunds.
- [ ] Unsuccessful bidders whose bid price is below $P_{clear}$ receive a 100% refund of deposited payment tokens with zero fee deduction.
- [ ] Platform fee of exactly 0.00% (Zero Fee) is deducted from gross cleared proceeds and routed per FeeController governance (0.00% at launch) without rounding leaks.
- [ ] Allocated tokens are deposited directly into `LinearVestingVault.sol` upon auction finalization.
- [ ] Linear vesting vault enforces cliff period: `claimVestedTokens` reverts with `CliffNotReached` prior to cliff expiry.
- [ ] Linear vesting vault streams tokens second-by-second after cliff expiration, reaching exactly 100% claimable at `startTime + vestingDuration`.
- [ ] Revocable vesting schedules calculate accrued vested balance correctly and return unvested tokens to the authorized authority upon revocation.
- [ ] Gasless EIP-2771 forwarder support correctly resolves `_msgSender()` across both contracts.
- [ ] Administrative synchronization entrypoint `syncVestingSchedule` functions in accordance with RUNBOOK-32.
- [ ] ERC-3643 compliance hooks successfully reject bids or vesting claims from non-whitelisted addresses.
- [ ] Foundry test suite achieves >95% statement and branch coverage across happy path, edge cases, and adversary fuzzing.
- [ ] Static security analysis using Slither and Mythril confirms zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across the entire specification document.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301`: Permissioned Blockchain Evaluation & Selection (Hyperledger Besu).
  - Prompt `303`: Permissioned Asset Token Issuance Smart Contract (ERC-3643).
  - Prompt `305`: Transfer Compliance Hooks Smart Contract & Whitelist Registry.
  - Prompt `307`: Multi-Party Authorization & Multisig Governance Smart Contract.
- **Parallel Tasks:**
  - Prompt `259`: Account Abstraction Bundler & Paymaster Service (EIP-2771 / ERC-4337 gas sponsorship).
  - Prompt `266`: Primary Market Launchpad Engine (off-chain auction orchestration).
  - Prompt `213`: Custodian Depository Integration Service (NSDL/CDSL & custody verification).
  - Prompt `210`: Fee & Realized PnL Engine.
  - Prompt `230`: Settlement Guarantee Fund (SGF) Service.
- **Subsequent Prompts Enabled:**
  - Prompt `309`: Blockchain Event Indexing & Query Service (indexing launchpad events).
  - Prompt `315`: On-Chain Settlement Guarantee Fund Contract (serving as default liquidity buffer under zero-fee model).
  - Prompt `336`: Automated Tax Withholding and eTDS Ledger Smart Contract.
  - Prompt `530`: Flutter Primary Market Launchpad & Dutch Auction Bidding Flow.
