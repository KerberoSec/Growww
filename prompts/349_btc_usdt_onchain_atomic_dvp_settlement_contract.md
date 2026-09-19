# 349 - Real-Money BTC/USDT Atomic Delivery-versus-Payment (DvP) Settlement Contract (Solidity)

## Purpose
In institutional digital asset spot markets and tokenized financial exchanges, principal risk (often termed Herstatt risk or settlement failure risk) is the primary structural vulnerability in bilateral trade execution. If one counterparty transfers their asset (such as Bitcoin) prior to receiving the agreed payment leg (such as USDT), any insolvency, system failure, or network partitioning on either side causes irrevocable capital loss. Under CPMI-IOSCO Principle 8 (Settlement Finality) and international securities clearing standards, financial exchanges must enforce Delivery-versus-Payment (DvP) Model 1: the gross delivery of the base asset from seller to buyer occurs if and only if the simultaneous, final, and unconditional gross payment from buyer to seller is executed in the same atomic transaction.

On the Growww platform, deployed on a high-throughput permissioned Hyperledger Besu consortium ledger operating under QBFT consensus, spot trading between tokenized wrapped Bitcoin (`wBTC`) and tokenized institutional Tether (`weUSDT`) requires an immutable, on-chain atomic settlement gateway. Every matched trade executed off-chain by the institutional matching engine must be finalized on-chain with deterministic finality, cryptographic trade authentication, compliance verification, and automated revenue extraction.

This prompt specifies the **Real-Money BTC/USDT Atomic Delivery-versus-Payment (DvP) Settlement Contract (`BtcUsdtDvPSettlement.sol`, `IBtcUsdtSettlement.sol`)**. Deployed under `contracts/src/settlement/`, this smart contract executes bilateral single-trade and high-density batched trade settlements between verified buyers and sellers. It verifies off-chain matching engine execution attestations signed via EIP-712 typed data hashing, executes simultaneous transfers of `wBTC` (8 decimals) and `weUSDT` (6 decimals) via `SafeERC20`, integrates ERC-3643 permissioning hooks to guarantee investor KYC/AML eligibility, extracts the canonical 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on gross USDT trade notional turnover, and routes fees across Growww's institutional revenue waterfall: 60% to the Platform Operating Treasury, 25% to the Core Settlement Guarantee Fund (SGF), and 15% to the Investor Protection Fund (IPF).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/src/settlement/` comprising:
- `BtcUsdtDvPSettlement.sol`: Core upgradeable contract (UUPS proxy pattern) orchestrating atomic bilateral and batch settlements for BTC/USDT pairs, signature authentication, multi-vault fee splits, and emergency pause controls.
- `IBtcUsdtSettlement.sol`: Master interface defining all trade structs, batch payloads, custom error codes, compliance events, and external mutator/view function signatures.
- EIP-712 Cryptographic Signature Verification Engine: On-chain cryptographic verifier validating matching engine trade execution receipts and counterparty order authorizations, preventing off-chain order tampering and unauthorized settlement injection.
- Atomic Batch Settlement Processor: Calldata-optimized batch execution pipeline settling up to 100 bilateral BTC/USDT trades in a single Besu transaction, maximizing transaction throughput while preventing block gas limit exhaustion.
- Asymmetric Decimal Arithmetic & Rounding Engine: High-precision fixed-point math module handling asymmetric decimals between `wBTC` (8 decimals, 1 Satoshi = $10^{-8}$ BTC) and `weUSDT` (6 decimals, 1 Micro-USDT = $10^{-6}$ USDT) using OpenZeppelin `Math.mulDiv` to guarantee zero truncation error and zero fund leakage.
- Institutional Fee Split Waterfall: Automated on-chain deduction of the Universal Zero-Fee Model (0.00% fee - No fee at all) calculated on the USDT trade turnover, programmatically split into Treasury reserve Vault, Core SGF Vault, and Investor Protection Fund Vault per FeeController governance with exact zero-sum conservation.
- Idempotency, Replay Prevention, and Nonce Registry: Dual-layer replay protection utilizing a global unique trade hash registry (`isTradeSettled[tradeId]`) and per-account sequential nonces, ensuring trades cannot be double-settled, replayed across forks, or executed post-expiry.
- ERC-3643 Permission & Compliance Enforcement: Automated invocation of ERC-3643 identity registry hooks to assert that buyer and seller hold active, KYC-verified identity claims prior to moving token balances.
- Comprehensive Foundry Test Suite (`test/settlement/BtcUsdtDvPSettlement.t.sol`): Exhaustive unit, fuzz, invariant, and differential test suites verifying DvP atomicity, precision math, reentrancy immunity, fee split conservation, and gas efficiency.
- Hardened Deployment Script (`script/DeployBtcUsdtDvPSettlement.s.sol`): Foundry deployment and initialization pipeline configuring roles, vaults, and proxy verification.

## Scope Boundaries
- **In Scope:**
  - On-chain atomic DvP settlement swapping `wBTC` from seller to buyer and `weUSDT` from buyer to seller.
  - EIP-712 structured data hashing and ECDSA signature verification for matching engine execution attestations.
  - Single-trade settlement (`settleTrade`) and atomic multi-trade batch settlement (`settleBatch`).
  - 0.00% platform fee (No fee at all) calculation on gross USDT turnover, routed per FeeController governance (0.00% at launch).
  - Reentrancy protection via OpenZeppelin `ReentrancyGuardUpgradeable` utilizing EVM transient storage (`TSTORE`/`TLOAD`).
  - Enforcing strict batch size limits (`MAX_BATCH_SIZE = 100`) to guarantee deterministic block gas consumption.
  - Trade execution idempotency via `isTradeSettled[tradeId]` mapping and trade expiry timestamps.
  - ERC-3643 identity verification ensuring both counterparty wallets are verified and unflagged.
  - Role-Based Access Control via `AccessControlUpgradeable` (`DEFAULT_ADMIN_ROLE`, `SETTLEMENT_OPERATOR_ROLE`, `EMERGENCY_ADMIN_ROLE`).
  - Emergency pausing and settlement suspension restricted to multi-sig governance.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain Order Matching Engine and Limit Order Book matching (handled in Prompt 205).
  - Trade packaging, batch construction, and submission orchestration (handled in Trade Settlement Service, Prompt 208).
  - Real-time off-chain fee calculation and realized PnL accounting (handled in Fee Engine, Prompt 210).
  - Multi-partition relayer nonce management and 1.25x gas price escalation (handled in Prompt 245).
  - Bitcoin SPV proof verification and custodial bridge reserve pegging (handled in Prompt 322).
  - External cross-chain stablecoin liquidity bridging (handled in Prompt 323).
  - Statutory Section 194S e-TDS tax deduction and FIU-IND regulatory reporting (handled in Prompt 336 and Prompt 216).

| Out-of-Scope Component | Responsible System | Relevant Prompt |
| :--- | :--- | :--- |
| Matching Engine & Order Book | Order Matching Engine | Prompt 205 |
| Settlement Batch Construction & Orchestration | Trade Settlement Service | Prompt 208 |
| Gross Fee Calculation & PnL Ledger | Fee and Realized PnL Engine | Prompt 210 |
| Relayer Nonce Partitioning & Gas Escalation | Settlement Relayer Service | Prompt 245 |
| Custodial Bitcoin SPV & Bridge Peg | Bitcoin SPV & DLC Bridge Contract | Prompt 322 |
| Cross-Chain EVM Liquidity Bridge | EVM Cross-Chain Liquidity Bridge | Prompt 323 |
| Tax Withholding & e-TDS Compliance | Automated Tax Withholding Contract | Prompt 336 |
| Multi-Sig Governance Root | Multi-Party Multisig Contract | Prompt 307 |

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with native support for transient storage opcodes `TSTORE`/`TLOAD`, custom errors, and checked arithmetic).
- **Base Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`).
  - OpenZeppelin Contracts v5.0 `SafeERC20` for secure transfer of `wBTC` and `weUSDT`.
  - OpenZeppelin Contracts v5.0 `Math.sol` for full 512-bit safe fixed-point multiplication and division (`mulDiv`).
  - OpenZeppelin Contracts v5.0 `ECDSA` and `EIP712Upgradeable` for cryptographic signature validation.
- **Token Standards & Interfaces:**
  - ERC-20 (`IERC20`, `SafeERC20`) for wrapped Bitcoin (`wBTC`, 8 decimals) and wrapped electronic USDT (`weUSDT`, 6 decimals).
  - ERC-3643 (`IERC3643`, `IIdentityRegistry`) for real-time KYC/AML compliance checking.
- **Mathematical Standards:**
  - Basis-point scaling: `BPS_PRECISION = 10_000` (where 100 bps = 1.00%, 0 bps (0.00% fee at launch) = 0.00% (Zero Fee)).
  - Fee rate: `PLATFORM_FEE_BPS = 0` (0.00% Zero Fee at launch; governed by FeeController.sol with 50 bps cap).
  - Fee split allocation: governed dynamically by FeeController.sol.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, gas profiling, and property fuzzing; `cast` for JSON-RPC Besu interactions).
- **Static Analysis & Auditing:** Slither (Trail of Bits), Mythril, and Solhint.

## Backend / Infra Touchpoints
- **Trade Settlement Service (Prompt 208):** Ingests matched BTC/USDT trades from the Order Matching Engine (Prompt 205), generates EIP-712 execution attestations, bundles trades into optimal settlement batches, and dispatches them to the relayer queue.
- **Fee Engine (Prompt 210):** Validates the 0.00% (Zero Fee) platform fee calculation against off-chain trade notional records, tracks gross platform revenue, and reconciles fee receipts across Treasury, SGF, and IPF vaults.
- **Settlement Relayer Nonce Partitioning & Gas Escalator (Prompt 245):** Dedicated Go service utilizing 32 partitioned relayer accounts to sign and broadcast `settleBatch` transactions to Hyperledger Besu with automatic 1.25x gas price escalation on pending transactions.
- **Order Matching Engine (Prompt 205):** Ultra-low-latency matching engine generating matched trade execution logs, trade identifiers (`tradeId`), execution timestamps, and price/quantity commitments.
- **Settlement Guarantee Fund Contract (Prompt 315):** On-chain SGF repository receiving the 25% fee allocation to maintain mutualized capital reserves against systemic default.
- **Blockchain Event Indexer (Prompt 309):** Real-time Go/Rust indexer consuming `BtcUsdtTradeSettled` and `BatchSettlementExecuted` events to update portfolio holdings and order statuses in PostgreSQL within 50 milliseconds.
- **Real-Time Market Surveillance Engine (Prompt 228):** Ingests on-chain settlement receipts to detect wash trading, spoofing, or abnormal BTC/USDT price deviations, issuing automated emergency halt signals when necessary.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Block Finality:** Deployed on Hyperledger Besu private permissioned network governed by QBFT consensus with 2-second block times and deterministic single-block finality (zero chain reorganizations).
- **Dual-Leg Atomic Asset Movement:**
  - Base leg: `wBTC` is transferred from seller to buyer via `SafeERC20.safeTransferFrom(trade.seller, trade.buyer, trade.btcAmount)`.
  - Payment leg: `weUSDT` net amount (`grossUsdtAmount - totalFee`) is transferred from buyer to seller via `SafeERC20.safeTransferFrom(trade.buyer, trade.seller, netUsdtAmount)`.
  - Fee leg: `weUSDT` fee amount is transferred from buyer to the settlement contract, which routes:
    - 60% to `treasuryVault`
    - 25% to `sgfVault`
    - 15% to `ipfVault`
- **All-or-Nothing Settlement Guarantee:** If any transfer fails, if an authorization signature is invalid, if counterparty compliance checks fail, or if an account has insufficient allowance/balance, the entire transaction reverts atomically, ensuring no partial settlement or unhedged exposure.
- **Zero On-Chain PII Guarantee:** In compliance with India's Digital Personal Data Protection Act (DPDPA 2023) and global privacy mandates, no raw investor names, tax IDs, or off-chain credentials exist in contract calldata or state. Trades are identified exclusively by cryptographic hashes (`tradeId = keccak256(...)`), EVM addresses, and EIP-712 signature digests.
- **ERC-3643 Investor Verification:** Settlement checks `IIdentityRegistry.isVerified(trade.buyer)` and `IIdentityRegistry.isVerified(trade.seller)` before executing token transfers, ensuring compliance with institutional regulatory standards.
- **Hardware Security Module (HSM) Signing:** Settlement relayers and matching engine signers utilize secp256k1 ECDSA private keys hosted in FIPS 140-2 Level 3 HSMs managed under Prompt 311.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & File Layout:** Configure Foundry project structure under `contracts/`:
   - `contracts/src/settlement/BtcUsdtDvPSettlement.sol`
   - `contracts/src/interfaces/IBtcUsdtSettlement.sol`
   - `test/settlement/BtcUsdtDvPSettlement.t.sol`
   - `test/mocks/MockERC20.sol`
   - `test/mocks/MockIdentityRegistry.sol`
   - `script/DeployBtcUsdtDvPSettlement.s.sol`
2. **Define Master Interface & Structs:** Implement `IBtcUsdtSettlement.sol` specifying `BtcUsdtTrade`, `SettlementBatch`, `FeeSplit`, custom errors (`InvalidBatchSize`, `TradeAlreadySettled`, `TradeExpired`, `InvalidSignature`, `InvalidFeeAmount`, `UnauthorizedCaller`, `ComplianceFailed`), and regulatory events.
3. **Configure Storage Layout & ERC-7201 Namespaces:** Establish an upgrade-safe storage layout using the ERC-7201 standard:
   - Namespace slot: `keccak256(abi.encode(uint256(keccak256("growww.storage.BtcUsdtDvPSettlement")) - 1)) & ~bytes32(uint256(0xff))`
   - Maintain token bindings (`wBTC`, `weUSDT`), vault addresses (`treasuryVault`, `sgfVault`, `ipfVault`), identity registry, and settlement state.
4. **Implement Role-Based Access Control:** Integrate OpenZeppelin `AccessControlUpgradeable` configuring three enterprise roles:
   - `DEFAULT_ADMIN_ROLE`: Multi-sig governance address for proxy upgrades, vault address updates, and contract configuration.
   - `SETTLEMENT_OPERATOR_ROLE`: Authorized relayer accounts (Prompt 245) permitted to invoke `settleTrade` and `settleBatch`.
   - `EMERGENCY_ADMIN_ROLE`: High-speed automated surveillance role capable of triggering `pause()` and `unpause()`.
5. **Configure EIP-712 Domain Separator:** Implement EIP-712 initialization with domain name `"GrowwwBtcUsdtDvP"` and version `"1"`. Define the canonical typehash:
   - `TRADE_TYPEHASH = keccak256("BtcUsdtTrade(bytes32 tradeId,address buyer,address seller,uint256 btcAmount,uint256 usdtGrossAmount,uint256 feeAmount,uint64 matchedTimestamp,uint64 expiryTimestamp,uint256 nonce)")`
6. **Implement Asymmetric Decimal Fee Calculation Engine:** Build internal routine `_calculateAndVerifyFee`:
   - Enforce exact fee calculation: `expectedFee = Math.mulDiv(usdtGrossAmount, PLATFORM_FEE_BPS, BPS_PRECISION, Math.Rounding.Floor)`.
   - Assert `trade.feeAmount == expectedFee` to eliminate fee manipulation.
   - Compute fee split: `treasuryAmount = (expectedFee * 6000) / 10000`, `sgfAmount = (expectedFee * 2500) / 10000`, and `ipfAmount = expectedFee - treasuryAmount - sgfAmount` (allocating remainder dust to IPF, ensuring zero-loss conservation).
7. **Implement Cryptographic Signature Verification:** Implement internal helper `_verifyTradeSignature`:
   - Reconstruct EIP-712 typed data digest for `BtcUsdtTrade`.
   - Recover signer via OpenZeppelin `ECDSA.recover`.
   - Assert recovered address equals the authorized `matchingEngineSigner`.
8. **Implement ERC-3643 Investor Compliance Checks:** Implement internal helper `_verifyCompliance`:
   - Call `IIdentityRegistry(identityRegistry).isVerified(buyer)` and `IIdentityRegistry(identityRegistry).isVerified(seller)`.
   - Revert with `ComplianceCheckFailed` if either participant is unverified or suspended.
9. **Implement Single-Trade Settlement Logic (`settleTrade`):**
   - Verify caller has `SETTLEMENT_OPERATOR_ROLE`.
   - Ensure contract is not paused (`whenNotPaused`) and guard against reentrancy (`nonReentrant`).
   - Validate timestamps: `trade.expiryTimestamp >= block.timestamp`.
   - Check idempotency: require `!isTradeSettled[trade.tradeId]`, then set `isTradeSettled[trade.tradeId] = true`.
   - Execute signature verification and compliance validation.
   - Transfer `wBTC` from seller to buyer via `SafeERC20.safeTransferFrom`.
   - Transfer `weUSDT` net amount (`usdtGrossAmount - feeAmount`) from buyer to seller.
   - Transfer and split `weUSDT` fee amount to Treasury, SGF, and IPF vaults.
   - Emit `BtcUsdtTradeSettled` event.
10. **Implement Batched Settlement Engine (`settleBatch`):**
    - Enforce batch size boundaries: require `batch.trades.length > 0` and `batch.trades.length <= MAX_BATCH_SIZE` (100).
    - Cache storage pointers and vault addresses in memory prior to the execution loop to minimize gas.
    - Iterate over `batch.trades`, executing atomic dual-leg transfers and accumulating aggregate volume and fee metrics.
    - Emit `BatchSettlementExecuted(batch.batchId, batch.trades.length, totalBtcVolume, totalUsdtVolume, totalFees)`.
11. **Implement Trade Cancellation & Invalidation:** Implement `cancelTrade(bytes32 tradeId, string calldata reason)` restricted to `SETTLEMENT_OPERATOR_ROLE` to mark expired or canceled off-chain orders as permanently settled/invalidated, preventing stale submission.
12. **Implement Multi-Sig Governance & Parameter Setters:** Implement administrative functions protected by `onlyRole(DEFAULT_ADMIN_ROLE)`:
    - `setVaultAddresses(address treasury, address sgf, address ipf)`
    - `setMatchingEngineSigner(address newSigner)`
    - `setBatchSizeLimit(uint256 newLimit)`
    - `setIdentityRegistry(address newRegistry)`
13. **Build Comprehensive Foundry Mock & Test Harness:** Create `MockERC20` supporting arbitrary decimals (8 for wBTC, 6 for weUSDT) and `MockIdentityRegistry`. Construct unit test suite verifying happy-path settlement, signature verification, fee distribution, and custom error reverts.
14. **Implement Property-Based Fuzz & Invariant Tests:** Write invariant tests asserting:
    - **Solvency Invariant:** Contract token balances of `wBTC` and `weUSDT` remain exactly zero post-settlement (zero custody retention).
    - **Fee Conservation Invariant:** Sum of fee payments strictly equals `treasuryAmount + sgfAmount + ipfAmount == trade.feeAmount`.
    - **Replay Invariant:** Calling `settleTrade` with an identical `tradeId` reverts unconditionally on subsequent attempts.
15. **Execute Gas Profiling & Slither Static Analysis:** Run Slither to verify absence of reentrancy vulnerabilities, unchecked returns, or storage layout shadowing. Benchmark gas consumption to guarantee single trade settlement $< 85,000$ gas and batched settlement $< 50,000$ gas per trade.

## Interfaces / Contracts

### 1. BTC/USDT Settlement Interface (`IBtcUsdtSettlement.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IBtcUsdtSettlement
 * @author Growww National Blockchain Stock Exchange (NBSE)
 * @notice Master interface for on-chain atomic Delivery-versus-Payment (DvP) settlement
 *         of executed BTC/USDT trades with simultaneous 0.00% fee (No fee at all) extraction and 0.00% fee launch policy revenue splitting.
 */
interface IBtcUsdtSettlement {
    // --- Data Structures ---

    /**
     * @notice Struct representing a single bilateral BTC/USDT trade execution.
     * @param tradeId Unique 32-byte identifier generated by the matching engine.
     * @param buyer Address of the tokenized Bitcoin buyer (USDT payer).
     * @param seller Address of the tokenized Bitcoin seller (BTC deliverer).
     * @param btcAmount Gross Bitcoin amount transferred in Satoshis (8 decimals).
     * @param usdtGrossAmount Gross USDT consideration transferred in Micro-USDT (6 decimals).
     * @param feeAmount Total 0.00% (Zero Fee) platform fee extracted in Micro-USDT (6 decimals).
     * @param matchedTimestamp Timestamp when matching engine executed the order (Unix seconds).
     * @param expiryTimestamp Timestamp after which the trade execution is void (Unix seconds).
     * @param nonce Sequential nonce of the trade order for replay mitigation.
     */
    struct BtcUsdtTrade {
        bytes32 tradeId;
        address buyer;
        address seller;
        uint256 btcAmount;
        uint256 usdtGrossAmount;
        uint256 feeAmount;
        uint64 matchedTimestamp;
        uint64 expiryTimestamp;
        uint256 nonce;
    }

    /**
     * @notice Struct representing a batch of trades submitted for atomic settlement.
     * @param batchId Unique 32-byte identifier for the settlement batch.
     * @param trades Array of individual bilateral trades to settle atomically.
     * @param signatures Array of EIP-712 matching engine execution signatures corresponding to each trade.
     */
    struct SettlementBatch {
        bytes32 batchId;
        BtcUsdtTrade[] trades;
        bytes[] signatures;
    }

    /**
     * @notice Fee split breakdown for statutory revenue routing.
     * @param treasuryAmount Portion allocated to Platform Operating Treasury (60%).
     * @param sgfAmount Portion allocated to Core Settlement Guarantee Fund (25%).
     * @param ipfAmount Portion allocated to Investor Protection Fund (15%).
     */
    struct FeeSplit {
        uint256 treasuryAmount;
        uint256 sgfAmount;
        uint256 ipfAmount;
    }

    // --- Custom Errors ---

    error InvalidBatchSize(uint256 size, uint256 maxLimit);
    error ArrayLengthMismatch(uint256 tradesLength, uint256 signaturesLength);
    error TradeAlreadySettled(bytes32 tradeId);
    error TradeExpired(bytes32 tradeId, uint64 expiryTimestamp, uint256 currentTimestamp);
    error InvalidTradeSignature(bytes32 tradeId);
    error InvalidFeeAmount(bytes32 tradeId, uint256 providedFee, uint256 expectedFee);
    error ComplianceCheckFailed(address participant);
    error ZeroAddressNotAllowed();
    error InvalidTokenConfiguration();
    error SettlementPaused();
    error UnauthorizedCaller(address caller);

    // --- Events ---

    /**
     * @notice Emitted when an individual bilateral BTC/USDT trade settles successfully.
     */
    event BtcUsdtTradeSettled(
        bytes32 indexed tradeId,
        bytes32 indexed batchId,
        address indexed buyer,
        address seller,
        uint256 btcAmount,
        uint256 usdtNetAmount,
        uint256 feeAmount,
        uint256 nonce
    );

    /**
     * @notice Emitted when a batch of trades is settled atomically.
     */
    event BatchSettlementExecuted(
        bytes32 indexed batchId,
        uint256 tradeCount,
        uint256 totalBtcVolume,
        uint256 totalUsdtVolume,
        uint256 totalFeeUsdt
    );

    /**
     * @notice Emitted when transaction fees are split across platform vaults.
     */
    event FeeSplitDistributed(
        bytes32 indexed tradeId,
        uint256 treasuryAmount,
        uint256 sgfAmount,
        uint256 ipfAmount
    );

    /**
     * @notice Emitted when settlement vault addresses are updated by governance.
     */
    event VaultAddressesUpdated(
        address indexed treasuryVault,
        address indexed sgfVault,
        address indexed ipfVault
    );

    /**
     * @notice Emitted when the matching engine signer address is updated.
     */
    event MatchingEngineSignerUpdated(address indexed previousSigner, address indexed newSigner);

    /**
     * @notice Emitted when an unexecuted or stale trade is invalidated.
     */
    event TradeCancelled(bytes32 indexed tradeId, string reason);

    // --- External Functions ---

    function settleTrade(BtcUsdtTrade calldata trade, bytes calldata signature) external;

    function settleBatch(SettlementBatch calldata batch) external;

    function cancelTrade(bytes32 tradeId, string calldata reason) external;

    function setVaultAddresses(address treasury, address sgf, address ipf) external;

    function setMatchingEngineSigner(address newSigner) external;

    function setBatchSizeLimit(uint256 newLimit) external;

    function setIdentityRegistry(address newRegistry) external;

    function getTradeTypehash() external pure returns (bytes32);

    function isTradeSettled(bytes32 tradeId) external view returns (bool);

    function getVaultAddresses() external view returns (address treasury, address sgf, address ipf);
}
```

### 2. Core DvP Settlement Contract Skeleton (`BtcUsdtDvPSettlement.sol` Reference Layout)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {EIP712Upgradeable} from "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {IBtcUsdtSettlement} from "../interfaces/IBtcUsdtSettlement.sol";

interface IIdentityRegistry {
    function isVerified(address user) external view returns (bool);
}

/**
 * @title BtcUsdtDvPSettlement
 * @notice Production smart contract executing atomic BTC/USDT Delivery-versus-Payment settlements on Hyperledger Besu.
 */
contract BtcUsdtDvPSettlement is
    Initializable,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable,
    EIP712Upgradeable,
    IBtcUsdtSettlement
{
    using SafeERC20 for IERC20;

    // --- Roles ---
    bytes32 public constant SETTLEMENT_OPERATOR_ROLE = keccak256("SETTLEMENT_OPERATOR_ROLE");
    bytes32 public constant EMERGENCY_ADMIN_ROLE = keccak256("EMERGENCY_ADMIN_ROLE");

    // --- Constants ---
    uint256 public constant BPS_PRECISION = 10_000;
    uint256 public constant PLATFORM_FEE_BPS = 0; // 0.00% Zero Fee at launch (governed by FeeController)
    uint256 public constant TREASURY_SHARE_BPS = 0; // Governed dynamically by FeeController
    uint256 public constant SGF_SHARE_BPS = 0; // Governed dynamically by FeeController
    uint256 public constant IPF_SHARE_BPS = 0; // Governed dynamically by FeeController

    bytes32 public constant TRADE_TYPEHASH = keccak256(
        "BtcUsdtTrade(bytes32 tradeId,address buyer,address seller,uint256 btcAmount,uint256 usdtGrossAmount,uint256 feeAmount,uint64 matchedTimestamp,uint64 expiryTimestamp,uint256 nonce)"
    );

    // --- ERC-7201 Storage Layout ---
    struct SettlementStorage {
        IERC20 wBtc;
        IERC20 weUsdt;
        IIdentityRegistry identityRegistry;
        address matchingEngineSigner;
        address treasuryVault;
        address sgfVault;
        address ipfVault;
        uint256 maxBatchSize;
        mapping(bytes32 => bool) isTradeSettled;
    }

    // keccak256(abi.encode(uint256(keccak256("growww.storage.BtcUsdtDvPSettlement")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant SETTLEMENT_STORAGE_LOCATION =
        0xb7e56a73c52a0e44f83b63297127e539659b87fcf999cebb5f15d18e87498c00;

    function _getSettlementStorage() private pure returns (SettlementStorage storage $) {
        assembly {
            $.slot := SETTLEMENT_STORAGE_LOCATION
        }
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address settlementRelayer,
        address matchingEngineSigner,
        address wBtcAddress,
        address weUsdtAddress,
        address identityRegistryAddress,
        address treasuryVault,
        address sgfVault,
        address ipfVault
    ) external initializer {
        if (
            admin == address(0) ||
            settlementRelayer == address(0) ||
            matchingEngineSigner == address(0) ||
            wBtcAddress == address(0) ||
            weUsdtAddress == address(0) ||
            identityRegistryAddress == address(0) ||
            treasuryVault == address(0) ||
            sgfVault == address(0) ||
            ipfVault == address(0)
        ) revert ZeroAddressNotAllowed();

        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();
        __ReentrancyGuard_init();
        __EIP712_init("GrowwwBtcUsdtDvP", "1");

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(SETTLEMENT_OPERATOR_ROLE, settlementRelayer);
        _grantRole(EMERGENCY_ADMIN_ROLE, admin);

        SettlementStorage storage $ = _getSettlementStorage();
        $.wBtc = IERC20(wBtcAddress);
        $.weUsdt = IERC20(weUsdtAddress);
        $.identityRegistry = IIdentityRegistry(identityRegistryAddress);
        $.matchingEngineSigner = matchingEngineSigner;
        $.treasuryVault = treasuryVault;
        $.sgfVault = sgfVault;
        $.ipfVault = ipfVault;
        $.maxBatchSize = 100;
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(DEFAULT_ADMIN_ROLE) {}

    // --- Settlement Functions ---

    function settleTrade(
        BtcUsdtTrade calldata trade,
        bytes calldata signature
    ) external override onlyRole(SETTLEMENT_OPERATOR_ROLE) whenNotPaused nonReentrant {
        SettlementStorage storage $ = _getSettlementStorage();
        _executeSettlement($, bytes32(0), trade, signature);
    }

    function settleBatch(
        SettlementBatch calldata batch
    ) external override onlyRole(SETTLEMENT_OPERATOR_ROLE) whenNotPaused nonReentrant {
        SettlementStorage storage $ = _getSettlementStorage();
        uint256 tradeCount = batch.trades.length;
        if (tradeCount == 0 || tradeCount > $.maxBatchSize) revert InvalidBatchSize(tradeCount, $.maxBatchSize);
        if (tradeCount != batch.signatures.length) revert ArrayLengthMismatch(tradeCount, batch.signatures.length);

        uint256 totalBtc;
        uint256 totalUsdt;
        uint256 totalFee;

        for (uint256 i = 0; i < tradeCount; ) {
            _executeSettlement($, batch.batchId, batch.trades[i], batch.signatures[i]);
            totalBtc += batch.trades[i].btcAmount;
            totalUsdt += batch.trades[i].usdtGrossAmount;
            totalFee += batch.trades[i].feeAmount;
            unchecked { ++i; }
        }

        emit BatchSettlementExecuted(batch.batchId, tradeCount, totalBtc, totalUsdt, totalFee);
    }

    function _executeSettlement(
        SettlementStorage storage $,
        bytes32 batchId,
        BtcUsdtTrade calldata trade,
        bytes calldata signature
    ) internal {
        if ($.isTradeSettled[trade.tradeId]) revert TradeAlreadySettled(trade.tradeId);
        if (block.timestamp > trade.expiryTimestamp) revert TradeExpired(trade.tradeId, trade.expiryTimestamp, block.timestamp);

        // Verify EIP-712 signature
        bytes32 structHash = keccak256(
            abi.encode(
                TRADE_TYPEHASH,
                trade.tradeId,
                trade.buyer,
                trade.seller,
                trade.btcAmount,
                trade.usdtGrossAmount,
                trade.feeAmount,
                trade.matchedTimestamp,
                trade.expiryTimestamp,
                trade.nonce
            )
        );
        bytes32 digest = _hashTypedDataV4(structHash);
        address recoveredSigner = ECDSA.recover(digest, signature);
        if (recoveredSigner != $.matchingEngineSigner) revert InvalidTradeSignature(trade.tradeId);

        // Verify ERC-3643 KYC/AML compliance
        if (!$.identityRegistry.isVerified(trade.buyer)) revert ComplianceCheckFailed(trade.buyer);
        if (!$.identityRegistry.isVerified(trade.seller)) revert ComplianceCheckFailed(trade.seller);

        // Verify 0.00% fee (No fee at all) calculation
        uint256 feeBps = feeController.takerFeeBps(); // 0 at launch
            uint256 expectedFee = Math.mulDiv(trade.usdtGrossAmount, feeBps, BPS_PRECISION, Math.Rounding.Floor);
        if (trade.feeAmount != expectedFee) revert InvalidFeeAmount(trade.tradeId, trade.feeAmount, expectedFee);

        // Mark settled
        $.isTradeSettled[trade.tradeId] = true;

        // Split fee: Treasury reserve, Core SGF reserve, Investor Protection Fund
        uint256 treasuryAmount = expectedFee > 0 ? Math.mulDiv(expectedFee, 6_000, BPS_PRECISION, Math.Rounding.Floor) : 0;
        uint256 sgfAmount = Math.mulDiv(expectedFee, SGF_SHARE_BPS, BPS_PRECISION, Math.Rounding.Floor);
        uint256 ipfAmount = expectedFee - treasuryAmount - sgfAmount;

        uint256 usdtNetAmount = trade.usdtGrossAmount - expectedFee;

        // Dual-leg atomic transfer
        // Leg 1: Transfer wBTC (8 decimals) from seller to buyer
        $.wBtc.safeTransferFrom(trade.seller, trade.buyer, trade.btcAmount);

        // Leg 2: Transfer weUSDT net amount (6 decimals) from buyer to seller
        $.weUsdt.safeTransferFrom(trade.buyer, trade.seller, usdtNetAmount);

        // Leg 3: Transfer and allocate weUSDT fee to vaults
        if (expectedFee > 0) {
            $.weUsdt.safeTransferFrom(trade.buyer, $.treasuryVault, treasuryAmount);
            $.weUsdt.safeTransferFrom(trade.buyer, $.sgfVault, sgfAmount);
            $.weUsdt.safeTransferFrom(trade.buyer, $.ipfVault, ipfAmount);

            emit FeeSplitDistributed(trade.tradeId, treasuryAmount, sgfAmount, ipfAmount);
        }

        emit BtcUsdtTradeSettled(
            trade.tradeId,
            batchId,
            trade.buyer,
            trade.seller,
            trade.btcAmount,
            usdtNetAmount,
            expectedFee,
            trade.nonce
        );
    }

    // --- Admin / View Functions ---

    function cancelTrade(bytes32 tradeId, string calldata reason) external override onlyRole(SETTLEMENT_OPERATOR_ROLE) {
        SettlementStorage storage $ = _getSettlementStorage();
        if ($.isTradeSettled[tradeId]) revert TradeAlreadySettled(tradeId);
        $.isTradeSettled[tradeId] = true;
        emit TradeCancelled(tradeId, reason);
    }

    function setVaultAddresses(address treasury, address sgf, address ipf) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (treasury == address(0) || sgf == address(0) || ipf == address(0)) revert ZeroAddressNotAllowed();
        SettlementStorage storage $ = _getSettlementStorage();
        $.treasuryVault = treasury;
        $.sgfVault = sgf;
        $.ipfVault = ipf;
        emit VaultAddressesUpdated(treasury, sgf, ipf);
    }

    function setMatchingEngineSigner(address newSigner) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newSigner == address(0)) revert ZeroAddressNotAllowed();
        SettlementStorage storage $ = _getSettlementStorage();
        address oldSigner = $.matchingEngineSigner;
        $.matchingEngineSigner = newSigner;
        emit MatchingEngineSignerUpdated(oldSigner, newSigner);
    }

    function setBatchSizeLimit(uint256 newLimit) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newLimit == 0 || newLimit > 500) revert InvalidBatchSize(newLimit, 500);
        _getSettlementStorage().maxBatchSize = newLimit;
    }

    function setIdentityRegistry(address newRegistry) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newRegistry == address(0)) revert ZeroAddressNotAllowed();
        _getSettlementStorage().identityRegistry = IIdentityRegistry(newRegistry);
    }

    function getTradeTypehash() external pure override returns (bytes32) {
        return TRADE_TYPEHASH;
    }

    function isTradeSettled(bytes32 tradeId) external view override returns (bool) {
        return _getSettlementStorage().isTradeSettled[tradeId];
    }

    function getVaultAddresses() external view override returns (address treasury, address sgf, address ipf) {
        SettlementStorage storage $ = _getSettlementStorage();
        return ($.treasuryVault, $.sgfVault, $.ipfVault);
    }
}
```

## Security & Compliance Notes
- **Reentrancy Protection & Transient Storage Guard:**
  - Settlement operations mutate external token contract states across multiple calls (`safeTransferFrom`).
  - OpenZeppelin's `ReentrancyGuardUpgradeable` is strictly enforced across `settleTrade` and `settleBatch`.
  - On Shanghai/Cancun EVM deployments on Hyperledger Besu, transient storage (`TSTORE`/`TLOAD`) is utilized, keeping reentrancy lock overhead under 100 gas while preventing cross-function or cross-token reentrancy.
- **Strict Batch Size Bounding (`MAX_BATCH_SIZE = 100`):**
  - To prevent unbounded loop vulnerabilities and Besu block gas exhaustion (Besu default block gas limit: 30,000,000 gas), batches are hard-capped at 100 bilateral trades per transaction.
  - Submitting an empty batch (`length == 0`) or exceeding `maxBatchSize` reverts immediately with `InvalidBatchSize`.
- **Cryptographic Trade Authentication & Nonces:**
  - Every settlement requires an EIP-712 structured cryptographic signature generated by the authorized matching engine key (`matchingEngineSigner`).
  - The EIP-712 domain separator includes `verifyingContract` and `chainId`, preventing cross-contract or cross-chain replay attacks.
  - Nonces and execution timestamps (`expiryTimestamp`) guarantee that stale or expired orders cannot be mined by malicious or delayed relayers.
- **Idempotency & Double-Settlement Immunity:**
  - The contract maintains a global mapping `mapping(bytes32 => bool) isTradeSettled`.
  - Before executing any asset transfer, the contract checks `!isTradeSettled[tradeId]`.
  - The flag is set to `true` prior to external calls, strictly adhering to the Checks-Effects-Interactions (CEI) design pattern.
- **Zero-Custody Solvency Invariant:**
  - The settlement contract never holds custody of client assets between transactions.
  - All token transfers move directly from seller to buyer (`wBTC`) and from buyer to seller/vaults (`weUSDT`).
  - At the end of every transaction, the contract's own balance of `wBTC` and `weUSDT` must remain exactly 0. Any residual balance indicates an exploitable state defect.
- **Asymmetric Precision & Dust Conservation:**
  - `wBTC` operates at 8 decimals (1 Satoshi = $10^{-8}$ BTC) while `weUSDT` operates at 6 decimals (1 Micro-USDT = $10^{-6}$ USDT).
  - Platform fees are calculated exclusively on the USDT consideration:
    $$\text{Fee} = \left\lfloor \frac{\text{usdtGrossAmount} \times 1}{10000} \right\rfloor$$
  - Fee split calculations guarantee zero fund loss:
    $$\text{Treasury} = \left\lfloor \frac{\text{Fee} \times 6000}{10000} \right\rfloor, \quad \text{SGF} = \left\lfloor \frac{\text{Fee} \times 2500}{10000} \right\rfloor, \quad \text{IPF} = \text{Fee} - (\text{Treasury} + \text{SGF})$$
  - The remainder dust (if any fractional unit exists) is programmatically credited to the Investor Protection Fund (IPF), ensuring that:
    $$\text{Treasury} + \text{SGF} + \text{IPF} \equiv \text{Fee}$$
- **ERC-3643 Investor KYC/AML Compliance:**
  - In accordance with SEBI guidelines, FIU-IND regulations, and institutional consortium bylaws, participants must possess active KYC identity claims on-chain.
  - The settlement contract queries `IIdentityRegistry.isVerified()` for both buyer and seller. If either account has been frozen, sanctioned, or de-listed by compliance authorities, the trade settlement immediately reverts.
- **Zero On-Chain PII Guarantee:**
  - No personal identification data (such as names, addresses, PAN, Aadhaar, passport numbers, or tax identifiers) is stored or emitted.
  - Identities are represented strictly by cryptographic address commitments and EIP-712 signature hashes in compliance with India's DPDPA 2023.

## Acceptance Criteria
- [ ] Production Solidity contracts `BtcUsdtDvPSettlement.sol` and `IBtcUsdtSettlement.sol` compiled under Solidity ^0.8.24 with zero warnings and strict ERC-7201 upgradeable storage layout.
- [ ] Slither static analysis runs with zero high or medium severity warnings, verifying absence of unchecked external calls, reentrancy vulnerabilities, or storage collisions.
- [ ] Atomic DvP execution verified: `wBTC` delivers to buyer and `weUSDT` delivers to seller if and only if both transfers succeed simultaneously.
- [ ] EIP-712 cryptographic signature verification verified: Orders signed by unauthorized keys revert with `InvalidTradeSignature`.
- [ ] Fee waterfall verified: Canonical 0.00% (Zero Fee) platform fee is deducted from gross USDT amount and routed per FeeController governance (0.00% at launch) with zero-dust leakage.
- [ ] Idempotency verified: Re-submitting a previously settled `tradeId` reverts immediately with `TradeAlreadySettled`.
- [ ] Expiry protection verified: Submitting a trade where `block.timestamp > trade.expiryTimestamp` reverts with `TradeExpired`.
- [ ] Batch size bounds verified: `settleBatch` reverts with `InvalidBatchSize` when submitted with 0 trades or more than 100 trades.
- [ ] Compliance hooks verified: Trades involving an unverified buyer or seller address revert with `ComplianceCheckFailed`.
- [ ] Invariant fuzz testing: Foundry invariant test suite runs $\ge 10,000$ iterations verifying:
  - Zero-Custody Invariant: Contract balances of `wBTC` and `weUSDT` are 0 at transaction termination.
  - Conservation of Value: `usdtNetAmount + treasuryFee + sgfFee + ipfFee == usdtGrossAmount`.
  - Replay Impossibility: No single `tradeId` can settle more than once under any execution sequence.
- [ ] Gas efficiency benchmark: Single bilateral trade settlement consumes $< 85,000$ gas; batched settlements consume $< 50,000$ gas per trade on Hyperledger Besu.
- [ ] Foundry test suite achieves $\ge 95\%$ line and branch coverage across all execution paths, access control guards, and revert conditions.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301` (Permissioned Blockchain Platform Evaluation & Selection - Hyperledger Besu)
  - Prompt `302` (Network Topology & QBFT Validator Infrastructure Setup)
  - Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract - SettlementDvP.sol)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
  - Prompt `006` (Fee Model Specification: Fixed Fee Structure & Treasury Allocation)
- **Parallel Tasks:**
  - Prompt `208` (Trade Settlement Service - Batch Packaging & Settlement Ingestion)
  - Prompt `210` (Transaction Fee & Realized PnL Calculation Engine)
  - Prompt `245` (Settlement Relayer Nonce Partitioning & Gas Escalator)
  - Prompt `322` (Bitcoin SPV & DLC Bridge Contract - wBTC Custody Peg)
  - Prompt `315` (On-Chain Settlement Guarantee Fund & Default Waterfall Smart Contract)
  - Prompt `309` (Blockchain Event Indexing & Query Service)
- **Subsequent Prompts Enabled:**
  - Prompt `326` (Perpetual Futures Clearing Smart Contract)
  - Prompt `336` (Automated Tax Withholding and e-TDS Compliance Ledger Smart Contract)
  - Prompt `612` (Web Institutional Trading Desk & Volume Rebate Portal)
  - Prompt `708` (Market Manipulation & Anti-Wash Trading Detection Pipeline)
  - Prompt `811` (Testnet vs Mainnet Dual Environment CI/CD Automation)
