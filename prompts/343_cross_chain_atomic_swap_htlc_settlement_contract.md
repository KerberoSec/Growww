# 343 - Cross-Chain Hashed Time-Locked Contract (HTLC) Atomic Settlement (Solidity)

## Purpose
Cross-border institutional clearing and settlement between foreign digital asset rails (Bitcoin on-chain, Lightning Network payment channels, Ethereum, and foreign fiat-backed stablecoins) and the Growww National Blockchain Stock Exchange (NBSE) permissioned Hyperledger Besu ledger traditionally depend on centralized multi-signature federations, trusted escrow custodians, or centralized third-party crypto-fiat bridges. These legacy mechanisms introduce single points of failure, counterparty credit risk, settlement latency, capital inefficiencies, and severe regulatory exposure under cross-border exchange frameworks.

In securities settlement, Delivery-versus-Payment (DvP) requires that asset transfer occurs if and only if the corresponding payment leg is simultaneously and irreversibly consummated. Cross-chain atomic settlement achieves trustless DvP across disjoint blockchains by binding both settlement legs to an identical cryptographic secret through Hashed Time-Locked Contracts (HTLCs).

This prompt specifies the enterprise-grade implementation of the **Cross-Chain Hashed Time-Locked Contract (HTLC) Atomic Settlement Smart Contract (`CrossChainHTLC.sol`, `ICrossChainHTLC.sol`)** on Hyperledger Besu. The contract coordinates trustless, non-custodial DvP settlement between foreign ledgers (principally Bitcoin/Lightning via Prompt 234 and cross-chain routing via Prompt 238) and the permissioned Besu ledger. By executing SHA-256 pre-image verification against EVM precompiles, enforcing mathematical timelock safety delta thresholds ($\Delta t \ge 24\text{ hours}$) to eliminate timelock griefing and race conditions, providing multi-stage timeout refund escalations, and requiring counterparty cryptographic signature authorizations, the contract guarantees atomic clearing with zero principal risk while complying with IFSCA foreign funding regulations and Growww's canonical Universal Zero-Fee Model (0.00% fee - No fee at all) Model (0.00% fee at launch; future fee parameters governed by FeeController.sol).

## What You Are Building
A production-grade, highly secure, upgradeable Solidity smart contract system located under `contracts/src/settlement/` and `contracts/interfaces/settlement/` comprising:
- `CrossChainHTLC.sol`: Enterprise-grade HTLC settlement engine deployed on Hyperledger Besu. It handles the complete lifecycle of cross-chain bilateral atomic swaps (lock creation, claim with SHA-256 pre-image verification, initiator timeout refunds, and multi-stage escalation arbitration). It supports ERC-20 tokenized equities, ERC-3643 permissioned real-world assets, and tokenized sovereign debt / digital rupee (`eINR`).
- `ICrossChainHTLC.sol`: Complete Solidity interface specifying swap records, multi-stage timelock configurations, fee allocations, EIP-712 authorization structs, custom errors, events, and all state-modifying and view function signatures.
- Cryptographic Precompile Integration: Native EVM SHA-256 precompile utilization (`0x02`) ensuring bit-for-bit hashlock compatibility with Bitcoin Native SegWit (BIP 141), Taproot Tapscript (BIP 341/342), and Lightning Network (BOLT 2/3/11/12) payment hashes without Keccak translation mismatches.
- Multi-Stage Timeout Refund Escalation Subsystem: A three-tier timeout framework (Claim Window $T_1$, Counterparty Grace Window $T_2$, and Administrative/Arbitration Escalation Window $T_3$) preventing unilateral fund freezing and mitigating network congestion or relayer failure risks.
- Counterparty Signature Authorization (EIP-712): Optional signed claim authorization from the designated receiver or clearing relayer to prevent public mempool front-running of pre-images and unauthorized fee-siphoning attacks.
- Fixed Transaction Fee & Multi-Vault Revenue Splitter: Automated assessment of Growww's fixed 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on gross swap turnover, routing per FeeController governance (0.00% at launch), while writing on-chain tax compliance hashes for Section 111A/112A capital gains reporting.
- Comprehensive Foundry Test Harness (`test/settlement/CrossChainHTLC.t.sol`): Complete suite of unit, fuzz, invariant, and cryptographic boundary tests covering happy-path pre-image reveals, timelock expiry refunds, griefing attack delta boundaries, reentrancy attacks, and ERC-3643 compliance transfer restrictions.

## Scope Boundaries
- **In Scope:**
  - Solidity ^0.8.24 implementation of `CrossChainHTLC.sol` and `ICrossChainHTLC.sol` on Hyperledger Besu.
  - Native EVM SHA-256 precompile verification matching foreign blockchain hashlocks (`hashlock = sha256(preimage)`).
  - Multi-token support: Standard ERC-20 tokens, ERC-3643 compliant permissioned security tokens, and native/wrapped `eINR`.
  - On-chain pre-image publication via indexed events (`HTLCClaimed`) enabling automated cross-chain relayers to claim foreign lockups.
  - Strict timelock safety delta validation ($\Delta t = T_{\text{refund}} - T_{\text{now}} \ge \text{MIN\_SAFETY\_DELTA} = 24\text{ hours}$).
  - Multi-stage timeout refund escalations with grace periods and optional mediator/arbitrator dispute intervention.
  - Counterparty authorization via EIP-712 typed structured signatures.
  - Growww canonical 0.00% (Zero Fee) platform fee calculation and 0.00% fee at launch (governed by FeeController.sol) distribution.
  - Tax compliance leaf hash logging for statutory Section 111A/112A capital gains compliance.
  - OpenZeppelin v5.0 UUPS upgradeability, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, and role-based access control.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain Bitcoin Core Taproot/SegWit script generation, PSBT signing, and Lightning invoice creation (handled in Prompt 234).
  - Cross-chain relayer watcher bots, liquidity routing, and synthetic FX conversion (handled in Prompt 238).
  - Centralized matching engine order matching and execution (handled in Prompt 205).
  - Domestic fiat INR banking rails (UPI, IMPS, RTGS) and CBDC retail interfaces (handled in Prompt 212 and Prompt 334).
  - Identity registry KYC attestation and claim issuer signature generation (handled in Prompt 305 and Prompt 703).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD` for gas-efficient reentrancy locks, and custom errors for minimal bytecode footprint).
- **Cryptographic Primitives:**
  - EVM SHA-256 precompile (`address(0x02)`) called directly or via Solidity built-in `sha256(bytes)` to guarantee bitwise compatibility with Bitcoin/Lightning hashlocks.
  - EVM `ecrecover` and OpenZeppelin `ECDSA` / `MessageHashUtils` for EIP-712 domain-separated signature validation.
- **Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`).
  - OpenZeppelin `SafeERC20` for non-standard ERC-20 return values.
  - ERC-3643 Token and Compliance Interfaces (`IERC3643`, `ICompliance`).
- **Development & Testing Toolchain:**
  - **Foundry** (`forge` for building and executing property-based fuzz tests, `cast` for RPC testing against Besu nodes).
  - **Static Security Analysis:** Slither, Mythril, and Solhint linters integrated into continuous delivery pipelines.

## Backend / Infra Touchpoints
- **Bitcoin Lightning & Taproot Ingress Service (Prompt 234):**
  - Consumes `HTLCLocked` events on Besu to generate matching Bitcoin Tapscript HTLC UTXOs or Lightning Network hold invoices locked with the identical SHA-256 `hashlock`.
  - Ingests `HTLCClaimed` events from Besu, extracts the revealed 32-byte `preimage`, and settles the Bitcoin on-chain HTLC or settles the Lightning payment channel invoice off-chain.
- **Cross-Chain Collateral & Synthetic FX Router Service (Prompt 238):**
  - Evaluates cross-chain exchange rates, enforces margin haircuts (BTC 20%, ETH 25%, SOL 30%), calculates swap notional in `eINR`, and coordinates bilateral swap creation across ledgers.
  - Monitors expiry countdowns and triggers automated refunds or escalations if counterparty fails to claim.
- **Trade Settlement Service (Prompt 208):**
  - Orchestrates Delivery-versus-Payment (DvP) trade packaging, generates matched swap parameters, and signs EIP-712 authorizations as the NBSE settlement coordinator.
- **Blockchain Event Indexer (Prompt 309):**
  - Ingests and indexes `HTLCLocked`, `HTLCClaimed`, `HTLCRefunded`, and `HTLCEscalated` events in real-time, updating investor order statuses in sub-second latency.
- **Settlement Guarantee Fund & Treasury Vaults (Prompt 230 / Prompt 315):**
  - Receives allocations per FeeController governance from the 0.00% (Zero Fee) platform fee on each settled swap.
- **Automated Tax Withholding & eTDS Ledger (Prompt 336):**
  - Captures the emitted `taxProofHash` to record cost basis, sale proceeds, and FIFO capital gains classification under Section 111A/112A.

## Blockchain Interaction (Permissioned Hyperledger Besu Ledger with 1:1 Custody Backing, Zero PII, QBFT)
- **Consensus & Finality:** Deployed on the Growww NBSE Hyperledger Besu consortium network operating QBFT consensus with 2-second deterministic block times and immediate finality, eliminating chain reorganizations and orphan block risks on the domestic ledger.
- **Trustless Atomic DvP Settlement:** Neither party can seize or steal assets during the swap. If the pre-image is revealed on Hyperledger Besu, the secret becomes public on the blockchain, allowing the initiating party to claim the corresponding foreign leg (Bitcoin UTXO or Lightning channel) before the foreign timelock expires. If the pre-image is never revealed, both parties receive a full refund after their respective timelocks elapse.
- **Cross-Chain Pre-Image Publication:** When `claim(bytes32 swapId, bytes calldata preimage)` is executed, the Besu contract verifies `sha256(preimage) == swap.hashlock`. Upon verification, the full 32-byte pre-image is emitted in the `HTLCClaimed` event log and stored in contract state, publishing it irreversibly to all cross-chain indexers and relayer daemons.
- **Zero On-Chain PII Invariant:** The contract stores only abstract identifiers (`bytes32 swapId`, `bytes32 hashlock`, `bytes32 taxProofHash`, `address initiator`, `address recipient`, `uint256 amount`, `uint64 timelock`). No personal identifiable information (PAN, Aadhaar, name, email, IP address) is ever transmitted or written to storage.
- **ERC-3643 Permissioned Transfer Compliance:** When locking or claiming regulated securities tokens, the contract interacts with the token's ERC-3643 compliance hooks to verify that both initiator and recipient maintain valid ONCHAINID identity claims and meet regulatory eligibility criteria.
- **Multi-Signature Governance:** Contract upgrades, safety delta threshold adjustments, fee collector vault addresses, and emergency pauses require 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM hardware keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:**
   - Create directories `contracts/src/settlement/` and `contracts/interfaces/settlement/`.
   - Scaffold `contracts/src/settlement/CrossChainHTLC.sol` and `contracts/interfaces/settlement/ICrossChainHTLC.sol`.
   - Set up test harness in `test/settlement/CrossChainHTLC.t.sol` using Foundry.
2. **Define Data Structures and Interface (`ICrossChainHTLC.sol`):**
   - Define enums: `SwapStatus` (`UNINITIALIZED`, `LOCKED`, `CLAIMED`, `REFUNDED`, `ESCALATED`).
   - Define structs: `HTLCSwap`, `MultiStageTimelock`, `FeeBreakdown`, `ClaimAuthorization`.
   - Declare events: `HTLCLocked`, `HTLCClaimed`, `HTLCRefunded`, `HTLCEscalated`, `PlatformFeeDistributed`, `SafetyDeltaUpdated`.
   - Declare custom errors with rich contextual arguments to avoid string bloat.
3. **Initialize Base Contract & Access Control:**
   - Inherit `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, and `PausableUpgradeable`.
   - Define roles: `DEFAULT_ADMIN_ROLE`, `OPERATOR_ROLE`, `ARBITRATOR_ROLE`, `EMERGENCY_ROLE`.
   - Initialize EIP-712 domain separator: `EIP712("GrowwwCrossChainHTLC", "1")`.
4. **Implement Storage Layout and Parameter Constants:**
   - Define storage mapping: `mapping(bytes32 => HTLCSwap) private _swaps`.
   - Define pre-image registry: `mapping(bytes32 => bytes) private _revealedPreimages`.
   - Establish minimum safety delta: `uint64 public minSafetyDelta` (initialized to `86400` seconds / 24 hours).
   - Establish maximum timelock window: `uint64 public maxTimelockWindow` (e.g., `2592000` seconds / 30 days).
   - Store vault addresses for Treasury (60%), Core SGF (25%), and IPF (15%).
5. **Implement Safe Lock Creation (`createLock`):**
   - Accept parameters: `recipient`, `token`, `amount`, `hashlock`, `timelocks` ($T_1, T_2, T_3$), `notionalPaise`, `taxProofHash`.
   - Validate that `swapId = keccak256(abi.encode(initiator, recipient, token, amount, hashlock, timelocks.claimExpiry))` does not already exist.
   - Enforce minimum safety delta check: require $T_1 - \text{block.timestamp} \ge \text{minSafetyDelta}$.
   - Enforce multi-stage ordering: require $\text{block.timestamp} < T_1 < T_2 < T_3$.
   - Pull tokens into contract escrow using `SafeERC20.safeTransferFrom(msg.sender, address(this), amount)`.
   - Record `HTLCSwap` with status `LOCKED` and emit `HTLCLocked`.
6. **Implement SHA-256 Pre-Image Verification & Claim (`claim`):**
   - Accept `swapId` and `bytes calldata preimage`.
   - Verify swap status is `LOCKED`.
   - Verify that $\text{block.timestamp} \le T_1$ (claim window active).
   - Verify cryptographic pre-image against stored hashlock: `sha256(preimage) == swap.hashlock`. Revert with `InvalidPreimage` on mismatch.
   - Restrict caller: must be designated `recipient` or authorized settlement relayer.
   - Update swap status to `CLAIMED` and store `_revealedPreimages[swapId] = preimage`.
7. **Implement Counterparty Signature Authorization (`claimWithAuthorization`):**
   - Implement EIP-712 structured claim authorization allowing a third-party relayer or the exchange settlement router to submit the pre-image on behalf of the recipient.
   - Verify EIP-712 typed signature from `recipient` authorizing the transaction to prevent front-running and gas siphoning.
8. **Implement Growww 0.00% fee (No fee at all) Deduction & Revenue Split:**
   - Upon successful claim, calculate platform fee on `notionalPaise`: $\text{feePaise} = (\text{notionalPaise} \times 1) / 10000$ (0 bps (0.00% fee at launch) / 0.00% (Zero Fee)).
   - Calculate token-denominated fee equivalent based on swap amount proportion: $\text{feeTokens} = (\text{swap.amount} \times 1) / 10000$.
   - Calculate split: Treasury reserve, Core SGF, Investor Protection Fund per FeeController governance.
   - Transfer net tokens ($\text{amount} - \text{feeTokens}$) to recipient.
   - Transfer fee tokens to the respective vaults and emit `PlatformFeeDistributed`.
9. **Implement Primary Expiry Refund (`refund`):**
   - Accept `swapId`.
   - Verify swap status is `LOCKED`.
   - Enforce timelock condition: $\text{block.timestamp} > T_1$ (primary claim window expired).
   - If a multi-stage dispute was not opened, verify $\text{block.timestamp} > T_2$ (counterparty grace window expired) or caller is `initiator`.
   - Update swap status to `REFUNDED`.
   - Transfer 100% of escrowed tokens back to `initiator` without fee deduction (zero fee on aborted trades).
   - Emit `HTLCRefunded`.
10. **Implement Multi-Stage Escalation and Dispute Resolution:**
    - Implement `escalateSwap(bytes32 swapId, bytes calldata reason)`: Allows either party or the cross-chain router to escalate during window $[T_1, T_2]$ if foreign leg confirmation was delayed or disputed.
    - Set status to `ESCALATED`.
    - Implement `resolveEscalation(bytes32 swapId, address beneficiary, bytes calldata resolutionProof)`: Callable only by `ARBITRATOR_ROLE` or multi-sig governance prior to $T_3$, disbursing funds to the legitimate claimant based on verified foreign chain proofs.
    - If $T_3$ expires without arbitration resolution, automatically allow `initiator` to trigger final default refund.
11. **Implement View Methods & Pre-Image Inspection:**
    - Implement `getSwap(bytes32 swapId)` returning full `HTLCSwap` struct.
    - Implement `getRevealedPreimage(bytes32 swapId)` returning stored pre-image bytes.
    - Implement `isClaimable(bytes32 swapId)` returning status and time remaining.
    - Implement `computeFee(uint256 notionalPaise)` returning fee breakdown.
12. **Implement Administrative & Safety Delta Controls:**
    - Implement `setMinSafetyDelta(uint64 newDelta)` restricted to `DEFAULT_ADMIN_ROLE` (enforcing hard floor of at least 12 hours).
    - Implement `setRevenueVaults(address treasury, address coreSgf, address ipf)` with non-zero address validation.
    - Implement `pause()` and `unpause()` circuit breakers for emergency stop.
13. **Construct Comprehensive Foundry Unit & Fuzz Tests:**
    - Test standard SHA-256 pre-image claim lifecycle with exact event verification.
    - Test initiator refund after $T_1$ expiration.
    - Fuzz test varying pre-image lengths (standard 32 bytes and dynamic payloads).
    - Verify that submission with Keccak-256 hash or invalid SHA-256 pre-image immediately reverts.
    - Verify that setting timelock below `minSafetyDelta` ($< 24\text{ hours}$) reverts with `SafetyDeltaTooLow`.
14. **Construct Griefing & Reentrancy Invariant Tests:**
    - Construct adversarial test simulating counterparty delay and mempool front-running.
    - Construct malicious ERC-20 token attempting reentrancy during `claim` and `refund`. Verify `ReentrancyGuard` reverts call.
    - Invariant test: Contract balance must always equal sum of all active `LOCKED` and `ESCALATED` swap amounts.
    - Invariant test: Total fee deducted on claimed swaps must strictly equal 0 bps (0.00% fee at launch) with exact 0.00% fee launch policy distribution.
15. **Execute Static Analysis & Gas Profiling:**
    - Run Slither static analyzer and verify zero high, medium, or reentrancy warnings.
    - Profile gas usage for `claim` and `createLock` ensuring execution gas well below Besu block gas limits.

## Interfaces / Contracts

### 1. Cross-Chain HTLC Interface (`contracts/interfaces/settlement/ICrossChainHTLC.sol`)

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ICrossChainHTLC
 * @notice Interface for the Cross-Chain Hashed Time-Locked Contract (HTLC) Atomic Settlement Engine on Hyperledger Besu.
 * @dev Coordinates trustless, atomic DvP settlement between Bitcoin/Lightning or foreign chains and the permissioned Besu ledger.
 */
interface ICrossChainHTLC {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum SwapStatus {
        UNINITIALIZED,
        LOCKED,
        CLAIMED,
        REFUNDED,
        ESCALATED
    }

    enum ChainIdentifier {
        BITCOIN_MAINNET,
        BITCOIN_LIGHTNING,
        ETHEREUM_MAINNET,
        SOLANA_MAINNET,
        BESU_CONSORTIUM
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct MultiStageTimelock {
        uint64 claimExpiry;       // T1: Primary receiver claim deadline (minimum safety delta enforced)
        uint64 graceExpiry;       // T2: Counterparty grace / escalation initiation deadline
        uint64 arbitrationExpiry; // T3: Final arbitration cut-off after which default refund is permitted
    }

    struct HTLCSwap {
        bytes32 swapId;                // Unique swap identifier (keccak256 hash of core parameters)
        bytes32 hashlock;              // Cryptographic SHA-256 lock (sha256(preimage))
        address initiator;             // Party escrowing the assets on Besu
        address recipient;             // Authorized counterparty eligible to claim on Besu
        address token;                 // Escrowed ERC-20 or ERC-3643 token address
        uint256 amount;                // Quantity of tokens escrowed
        uint256 notionalPaise;         // Trade valuation in paise for fee assessment
        MultiStageTimelock timelocks;  // Multi-stage timeout milestones
        SwapStatus status;             // Current lifecycle state of the swap
        ChainIdentifier foreignChain;  // Foreign chain containing the counterparty leg
        bytes32 foreignTxOrChannelId;  // Foreign transaction hash or Lightning channel ID
        bytes32 taxProofHash;          // Attestation hash for Section 111A/112A capital gains compliance
    }

    struct LockParameters {
        address recipient;
        address token;
        uint256 amount;
        bytes32 hashlock;
        MultiStageTimelock timelocks;
        uint256 notionalPaise;
        ChainIdentifier foreignChain;
        bytes32 foreignTxOrChannelId;
        bytes32 taxProofHash;
    }

    struct FeeBreakdown {
        uint256 totalFeeTokens;
        uint256 treasuryTokens;
        uint256 coreSgfTokens;
        uint256 ipfTokens;
    }

    struct ClaimAuthorization {
        bytes32 swapId;
        bytes32 hashlock;
        address authorizedRelayer;
        uint64 validUntil;
        bytes signature;
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event HTLCLocked(
        bytes32 indexed swapId,
        bytes32 indexed hashlock,
        address indexed initiator,
        address recipient,
        address token,
        uint256 amount,
        uint64 claimExpiry,
        ChainIdentifier foreignChain,
        bytes32 foreignTxOrChannelId
    );

    event HTLCClaimed(
        bytes32 indexed swapId,
        bytes32 indexed hashlock,
        address indexed recipient,
        address claimCaller,
        bytes preimage,
        uint256 netAmountReceived,
        uint256 platformFeeTokens,
        bytes32 taxProofHash
    );

    event HTLCRefunded(
        bytes32 indexed swapId,
        bytes32 indexed hashlock,
        address indexed initiator,
        uint256 amountRefunded,
        string reason
    );

    event HTLCEscalated(
        bytes32 indexed swapId,
        address indexed reporter,
        bytes reasonPayload
    );

    event HTLCEscalationResolved(
        bytes32 indexed swapId,
        address indexed beneficiary,
        uint256 amountDisbursed,
        address indexed arbitrator
    );

    event PlatformFeeDistributed(
        bytes32 indexed swapId,
        uint256 totalFeeTokens,
        uint256 treasuryShareTokens,
        uint256 coreSgfShareTokens,
        uint256 ipfShareTokens
    );

    event SafetyDeltaUpdated(
        uint64 oldDeltaSeconds,
        uint64 newDeltaSeconds
    );

    event RevenueVaultsUpdated(
        address treasuryVault,
        address coreSgfVault,
        address ipfVault
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error SwapAlreadyExists(bytes32 swapId);
    error SwapNotFound(bytes32 swapId);
    error SwapNotInLockedState(bytes32 swapId, SwapStatus currentStatus);
    error SwapNotInEscalatedState(bytes32 swapId, SwapStatus currentStatus);
    error InvalidPreimage(bytes32 expectedHashlock, bytes32 computedHash);
    error PreimageLengthInvalid(uint256 length);
    error TimelockNotExpired(uint64 expiryTimestamp, uint64 currentTimestamp);
    error TimelockExpired(uint64 expiryTimestamp, uint64 currentTimestamp);
    error SafetyDeltaTooLow(uint64 providedDelta, uint64 requiredMinDelta);
    error InvalidTimelockSequence(uint64 t1, uint64 t2, uint64 t3);
    error UnauthorizedClaimant(address caller, address expectedRecipient);
    error InvalidAuthorizationSignature(bytes32 swapId, address signer);
    error AuthorizationExpired(uint64 validUntil, uint64 currentTimestamp);
    error ZeroAmount();
    error ZeroAddress();
    error TokenTransferFailed(address token, address from, address to, uint256 amount);
    error VaultAddressInvalid(address vault);
    error DisputeWindowClosed(uint64 graceExpiry, uint64 currentTimestamp);
    error ArbitrationWindowExpired(uint64 arbitrationExpiry, uint64 currentTimestamp);

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    /**
     * @notice Creates and escrows assets in a new cross-chain Hashed Time-Locked Contract.
     * @param params Complete struct of lock parameters.
     * @return swapId The unique bytes32 identifier of the escrowed swap.
     */
    function createLock(LockParameters calldata params) external returns (bytes32 swapId);

    /**
     * @notice Claims escrowed tokens by revealing the cryptographic SHA-256 pre-image.
     * @param swapId The unique identifier of the swap.
     * @param preimage The secret pre-image revealing the hashlock (sha256(preimage) == hashlock).
     * @return netAmountReceived The token amount transferred to recipient after 0.00% fee (No fee at all) deduction.
     */
    function claim(bytes32 swapId, bytes calldata preimage) external returns (uint256 netAmountReceived);

    /**
     * @notice Claims escrowed tokens via an authorized third-party relayer using recipient EIP-712 authorization.
     * @param swapId The unique identifier of the swap.
     * @param preimage The secret pre-image matching the hashlock.
     * @param auth EIP-712 authorization signed by the recipient.
     * @return netAmountReceived The token amount transferred to recipient after 0.00% fee (No fee at all) deduction.
     */
    function claimWithAuthorization(
        bytes32 swapId,
        bytes calldata preimage,
        ClaimAuthorization calldata auth
    ) external returns (uint256 netAmountReceived);

    /**
     * @notice Refunds escrowed tokens back to initiator upon expiration of the primary claim timelock.
     * @param swapId The unique identifier of the swap.
     * @return amountRefunded The total token amount returned to initiator without fee deduction.
     */
    function refund(bytes32 swapId) external returns (uint256 amountRefunded);

    /**
     * @notice Escalates a locked swap into a disputed status during the grace window (T1 < block.timestamp <= T2).
     * @param swapId The unique identifier of the swap.
     * @param reasonPayload Arbitrary calldata detailing the cross-chain timeout or delivery dispute.
     */
    function escalateSwap(bytes32 swapId, bytes calldata reasonPayload) external;

    /**
     * @notice Resolves an escalated swap by disbursing funds to the verified beneficiary.
     * @dev Restricted to ARBITRATOR_ROLE or MultiSig Governance before T3 expires.
     * @param swapId The unique identifier of the swap.
     * @param beneficiary Address receiving the escrowed funds.
     * @param resolutionProof Cryptographic or multi-sig proof attesting to foreign settlement outcome.
     */
    function resolveEscalation(
        bytes32 swapId,
        address beneficiary,
        bytes calldata resolutionProof
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getSwap(bytes32 swapId) external view returns (HTLCSwap memory);
    function getRevealedPreimage(bytes32 swapId) external view returns (bytes memory);
    function isClaimable(bytes32 swapId) external view returns (bool claimable, uint64 secondsRemaining);
    function minSafetyDelta() external view returns (uint64);
    function maxTimelockWindow() external view returns (uint64);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function computeFeeBreakdown(uint256 tokenAmount) external pure returns (FeeBreakdown memory fees);
}
```

### 2. Implementation Specification Contract Overview (`contracts/src/settlement/CrossChainHTLC.sol`)

The `CrossChainHTLC.sol` contract implements `ICrossChainHTLC` along with:
- Inherited OpenZeppelin modules: `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `EIP712Upgradeable`.
- Storage variables:
  - `mapping(bytes32 => HTLCSwap) private _swaps;`
  - `mapping(bytes32 => bytes) private _revealedPreimages;`
  - `uint64 public minSafetyDelta;` (initialized to 86,400 seconds / 24 hours).
  - `uint64 public maxTimelockWindow;` (initialized to 2,592,000 seconds / 30 days).
  - `address public treasuryVault;`
  - `address public coreSgfVault;`
  - `address public ipfVault;`
- Roles:
  - `bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");`
  - `bytes32 public constant ARBITRATOR_ROLE = keccak256("ARBITRATOR_ROLE");`
  - `bytes32 public constant EMERGENCY_ROLE = keccak256("EMERGENCY_ROLE");`

## Security & Compliance Notes

### 1. Timelock Griefing Attack Mitigation & Safety Delta ($\Delta t \ge 24\text{ hours}$)
- **The Cross-Chain Griefing Threat:** In an asynchronous cross-chain atomic swap between Bitcoin/Lightning and Hyperledger Besu, the initiating counterparty on Ledger A faces timing vulnerability if the timelocks on both chains expire too close to one another ($T_A \approx T_B$). If the receiver claims assets on Ledger B by revealing the pre-image close to $T_B$, the initiator on Ledger A may experience network congestion, mempool eviction, or relayer failure, causing $T_A$ to expire before they can confirm their claim transaction on Ledger A. This would permit the counterparty to both keep the asset on Ledger B and refund the asset on Ledger A, resulting in double-spend theft.
- **Enforced Safety Delta Invariant:** `CrossChainHTLC.sol` strictly enforces:
  $$\Delta t = T_1 - \text{block.timestamp} \ge \text{minSafetyDelta} \ge 86400\text{ seconds (24 hours)}$$
  When coordinating with Bitcoin/Lightning, the Besu lock must maintain at least a 24-hour delta margin relative to the Bitcoin UTXO lock time (or Lightning CLTV expiry delta). If an initiator attempts to escrow funds with a timelock less than $\text{block.timestamp} + \text{minSafetyDelta}$, the transaction immediately reverts with `SafetyDeltaTooLow`.

### 2. Native EVM SHA-256 Precompile Compatibility
- Ethereum/EVM platforms natively favor Keccak-256 for internal operations. However, Bitcoin script (`OP_SHA256`, `OP_HASH160`) and Lightning Network BOLT specifications mandate SHA-256 for payment pre-images and hashlocks.
- `CrossChainHTLC.sol` computes hashlocks using native SHA-256 via EVM precompile address `0x02` (or built-in `sha256(preimage)`). This ensures exact bit-level parity across Bitcoin Tapscript, Lightning channel invoices, and Hyperledger Besu, preventing mismatched hashes or costly off-chain proof translation layers.

### 3. Multi-Stage Timeout Refund Escalations
- **Stage 1 (Primary Claim Window $[0, T_1]$):** The designated recipient holds exclusive right to claim the escrowed funds by presenting the valid 32-byte pre-image.
- **Stage 2 (Counterparty Grace & Dispute Window $(T_1, T_2]$):** If $T_1$ expires without claim, the swap enters a grace window. Either party or authorized router monitors can trigger `escalateSwap()` if foreign chain confirmation was submitted but delayed. This prevents instant refund execution while cross-chain proof verification is underway.
- **Stage 3 (Arbitration Window $(T_2, T_3]$):** If escalated, designated clearing arbitrators (`ARBITRATOR_ROLE` or multi-sig governance) review verified foreign chain proofs (e.g., via SPV verification or ZK light client verification) to disburse funds to the legitimate owner.
- **Post $T_3$ Default Expiry:** If $T_3$ elapses without resolution, default refund to `initiator` becomes permanently unlocked, ensuring escrowed assets can never be permanently frozen in the contract.

### 4. Counterparty Signature Authorization (EIP-712)
- To prevent front-running attacks in public or consortium transaction pools where malicious actors could observe a broadcast pre-image and front-run the `claim()` invocation to redirect fees, `claimWithAuthorization` verifies an EIP-712 typed signature signed by `recipient`. The signature binds the `swapId`, `hashlock`, and designated `authorizedRelayer`.

### 5. Reentrancy Protection and Checks-Effects-Interactions
- All mutator functions (`createLock`, `claim`, `refund`, `resolveEscalation`) strictly adhere to the Checks-Effects-Interactions (CEI) design pattern.
- State transitions (updating status to `CLAIMED` or `REFUNDED`) occur before external token transfers are executed.
- Uses OpenZeppelin `ReentrancyGuardUpgradeable` to block reentrant re-entrancy vectors across non-standard or malicious ERC-20 tokens.

### 6. IFSCA Foreign Funding Rules & GIFT City Regulatory Compliance
- Under International Financial Services Centres Authority (IFSCA) Capital Market Intermediaries Regulations, cross-border inward remittances and crypto-fiat DvP settlements must strictly enforce:
  - **KYC/AML Whitelisting:** Escrowed tokens governed by ERC-3643 (Prompt 305) invoke identity registry compliance checks on both `initiator` and `recipient`. Unregistered or sanctioned addresses cannot lock or receive tokens.
  - **Zero On-Chain PII:** Swap identifiers and tax compliance proofs (`taxProofHash`) use cryptographic commitments without storing investor names, permanent account numbers (PAN), or banking credentials on the Besu ledger.
  - **Statutory Tax Leaf Hashes:** Emitted events capture `taxProofHash` for off-chain capital gains calculation under Section 111A/112A and statutory eTDS reporting under Section 194S.

### 7. Growww Core Fee Model Integrity
- Platform fee is fixed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover.
- Fee allocation is automated on-chain per FeeController.sol governance (0.00% at launch).
- Zero custody or holding fees are assessed. Aborted swaps that are refunded after timeout incur exactly 0% platform fee.

## Acceptance Criteria
- [ ] `ICrossChainHTLC.sol` and `CrossChainHTLC.sol` compile cleanly under Solidity ^0.8.24 with zero warnings.
- [ ] SHA-256 pre-image verification utilizes EVM precompile (`0x02`) and matches native Bitcoin/Lightning SHA-256 outputs bit-for-bit.
- [ ] Swaps cannot be created with a claim timelock delta less than `minSafetyDelta` ($T_1 - \text{block.timestamp} \ge 24\text{ hours}$), reverting with `SafetyDeltaTooLow`.
- [ ] Multi-stage timelock ordering ($T_1 < T_2 < T_3$) is strictly enforced during lock creation.
- [ ] Valid pre-image submission within $T_1$ releases net escrowed tokens to recipient, emits `HTLCClaimed` with full pre-image in calldata, and deducts exactly 0.00% fee (No fee at all).
- [ ] Platform fee is split strictly per FeeController governance (0.00% at launch) without rounding leaks.
- [ ] Pre-image mismatches revert immediately with custom error `InvalidPreimage`.
- [ ] Timelocked refunds cannot be executed prior to $T_1$ expiration, reverting with `TimelockNotExpired`.
- [ ] Aborted swaps refunded after expiry return 100% of escrowed principal to initiator with zero fee deduction.
- [ ] Multi-stage escalation allows dispute submission during $(T_1, T_2]$ and prevents unauthorized refunds during pending arbitration.
- [ ] EIP-712 structured claim authorization verifies recipient signature and rejects unauthorized relayers or expired authorizations.
- [ ] Reentrancy attacks against `claim` and `refund` fail and revert cleanly.
- [ ] ERC-3643 compliance hooks successfully block transfers to non-whitelisted investor addresses.
- [ ] Contract is fully upgradeable via UUPS pattern and guarded by multi-signature role-based access control.
- [ ] Foundry test suite achieves >95% statement and branch coverage across happy path, timelock boundaries, and failure cases.
- [ ] Slither and Mythril static analysis tools report zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across the entire specification document.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301`: Permissioned Blockchain Evaluation & Selection (Hyperledger Besu).
  - Prompt `303`: Token Issuance Smart Contract (ERC-3643 & ERC-20).
  - Prompt `305`: Transfer Compliance Hooks Smart Contract.
  - Prompt `306`: Atomic DvP Settlement Smart Contract (`SettlementDvP.sol`).
  - Prompt `307`: MultiSig Governance Smart Contract.
- **Parallel Tasks:**
  - Prompt `234`: Bitcoin (BTC), Lightning Network & Taproot Collateral Ingress Service (Go microservice handling foreign HTLC legs).
  - Prompt `238`: Cross-Chain Collateral & Synthetic FX Router Service (Go/Python routing engine coordinating bilateral swaps).
  - Prompt `208`: Trade Settlement Service.
  - Prompt `210`: Fee & Realized PnL Engine.
  - Prompt `230`: Settlement Guarantee Fund (SGF) Service.
- **Subsequent Prompts Enabled:**
  - Prompt `309`: Event Indexing Service (indexing `HTLCLocked` and `HTLCClaimed` events).
  - Prompt `315`: Settlement Guarantee Fund Smart Contract (serving as default liquidity buffer under zero-fee model).
  - Prompt `336`: Automated Tax Withholding & eTDS Ledger Contract (consuming `taxProofHash`).
  - Prompt `509`: Flutter Cross-Chain Collateral & Order Placement Flow.
