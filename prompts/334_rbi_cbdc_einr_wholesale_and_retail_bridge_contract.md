# 334 - RBI CBDC eINR Wholesale and Retail Programmable Bridge Smart Contracts (RBIeINRBridge.sol, WrappedeINR.sol)

## Purpose
The Reserve Bank of India (RBI) Central Bank Digital Currency (CBDC) ecosystem operates across two distinct sovereign digital currency tiers: Wholesale Digital Rupee (eINR-W) for interbank clearing and secondary market settlement of government securities, and Retail Digital Rupee (eINR-R) for direct tokenized legal tender distribution to individuals and merchants. Integrating sovereign CBDC liquidity into the Growww National Blockchain Stock Exchange (NBSE) architecture on Hyperledger Besu requires an institutional, programmable, and regulatory-grade bridge.

Traditional fiat gateways introduce counterparty settlement latency (T+1 or T+0 batch settlement), depository reconciliation friction, and fragmented cash-leg operations. Conversely, pure crypto stablecoins introduce private issuer balance-sheet risks and regulatory non-compliance under Indian banking statutes (RBI Payment and Settlement Systems Act, 2007).

This prompt specifies the **RBI CBDC eINR Wholesale and Retail Programmable Bridge Smart Contract Suite (`RBIeINRBridge.sol`, `WrappedeINR.sol`)**. The contracts govern the bi-directional 1:1 conversion between sovereign off-chain RBI CBDC ledgers and on-chain Wrapped eINR (`weINR`) digital legal tender on Hyperledger Besu. The suite enforces cryptographic ISO 20022 message validation (`pacs.008` for retail credit transfers, `pacs.009` for wholesale financial institution transfers), Delivery-versus-Payment (DvP) Type 1 instant atomic settlement against ERC-3643 permissioned security tokens, multi-party threshold verification across RBI nodal accounts and scheduled commercial custodian banks, and Growww's canonical Universal Zero-Fee Model (0.00% fee - No fee at all) with a 0.00% fee at launch (governed by FeeController.sol) revenue split (with FIFO capital gains computed strictly for user tax compliance under Section 111A/112A).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/cbdc/` comprising:
- `WrappedeINR.sol`: Sovereign-backed, compliant fiat token contract representing wrapped digital rupee (`weINR`) with 4 decimal places ($10^{-4}$ precision, where 1 unit = 0.01 paise = 0.0001 INR; 10,000 units = 1 INR). The contract implements ERC-20 and ERC-3643 compliance hooks, restricted mint/burn roles bound strictly to verified bridge deposits, and account freezing / seize primitives for statutory court or RBI regulatory orders.
- `RBIeINRBridge.sol`: Central programmable bridge contract connecting RBI CBDC Wholesale/Retail gateways with the Besu consortium ledger. It verifies multi-party threshold signatures from authorized RBI verification nodes and commercial bank custodians, validates canonical ISO 20022 message payload hashes (`pacs.008` and `pacs.009`), processes 1:1 wrapped minting upon CBDC escrow locking, executes DvP Type 1 atomic security token settlement, coordinates burning for off-chain CBDC redemption payouts, and routes the Universal Zero-Fee Model (0.00% fee - No fee at all) to Treasury (60%), Core SGF (25%), and IPF (15%) reserve vaults.
- `IRBIeINRBridge.sol` and `IWrappedeINR.sol`: Comprehensive Solidity interface specifications with detailed data structures, enums, events, custom error definitions, and view/mutator function signatures.
- Comprehensive Foundry test harness (`test/cbdc/`): Unit, fuzz, invariant, and threshold verification test suites validating 100% sovereign reserve parity, ISO 20022 hash validation, replay attack prevention, sub-paise precision arithmetic, and strict compliance with the fixed 0.00% transaction fee (No fee at all) invariant.

## Scope Boundaries
- **In Scope:**
  - 1:1 Sovereign-backed wrapped digital rupee token implementation (`weINR`) with 4 decimal places ($10^{-4}$ sub-paise precision).
  - Bi-directional bridge lifecycle: minting `weINR` against verified off-chain eINR-W / eINR-R deposits and burning `weINR` for off-chain CBDC redemption payouts.
  - Verification of cryptographically signed ISO 20022 banking messages (`pacs.008.001.10` for retail and `pacs.009.001.10` for wholesale interbank transfers).
  - Multi-party threshold signature verification (M-of-N) across authorized RBI nodal gateways and scheduled commercial bank verification nodes.
  - Delivery-versus-Payment (DvP) Type 1 instant atomic settlement engine pairing `weINR` cash legs with ERC-3643 permissioned security token asset legs.
  - Account-level compliance checks interfacing with `IIdentityRegistry` (ERC-3643) for KYC/AML verification on all `weINR` holders and bridge participants.
  - Growww fixed 0.00% transaction fee (No fee at all) calculation on gross turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), with FIFO capital gains for tax compliance (Section 111A/112A).
  - Emergency circuit breakers, pause mechanisms, rate-limiting deposit/withdrawal windows, and timelocked multi-signature administrative controls.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain RBI CBDC distributed ledger core network and CBDC retail mobile wallet app (handled by RBI / NPCI infrastructure).
  - ISO 20022 XML parsing and HSM signing middleware (handled in Prompt 213 and Prompt 242).
  - Centralized bank account UPI / IMPS fiat on-ramps (handled in Prompt 203 and Prompt 212).
  - Matching engine for continuous limit order books (handled in Prompt 205).
  - Commodity and equity depository settlement adapters (handled in Prompt 213 and Prompt 331).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD`, and custom errors).
- **Token Standards:** Custom **ERC-20** with 4 decimal places ($10^{-4}$ sub-paise precision) and **ERC-3643** (T-REX) compliance hooks for identity-restricted digital legal tender.
- **Security & Modularity:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Signature & Cryptography:** OpenZeppelin `ECDSA`, `EIP712Upgradeable`, and `MessageHashUtils` for M-of-N threshold EIP-712 attestation verification.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation and invariant fuzzing, `cast` for Besu RPC interactions).
- **Static Analysis & Auditing:** Slither, Mythril, and Solhint automated CI pipelines.

## Backend / Infra Touchpoints
- **RBI CBDC Relayer Gateway (Prompt 213 & Prompt 242):** Listens to off-chain RBI Wholesale/Retail CBDC settlement events, constructs ISO 20022 `pacs.008` / `pacs.009` structured payloads, obtains threshold signatures from custodian bank nodes, and submits bridge mint transactions.
- **Trade Settlement Service (Prompt 208):** Orchestrates DvP Type 1 atomic execution across `RBIeINRBridge.sol` and `SettlementDvP.sol` (Prompt 306).
- **Double-Entry Wallet Account Ledger (Prompt 203):** Mirrors on-chain `weINR` mint, burn, and transfer events into off-chain double-entry ledger accounts with sub-paise precision.
- **Fee & Realized PnL Engine (Prompt 210):** Reconciles Universal Zero-Fee Model (0.00% fee - No fee at all) collections (0.00% fee at launch (governed by FeeController.sol) split) and computes Section 111A/112A capital gains tax compliance records.
- **Blockchain Event Indexer (Prompt 309):** Ingests events (`CBDCDepositMinted`, `CBDCRedemptionBurned`, `DvPTradeSettledAtomic`, `PlatformFeeDistributed`) for real-time portfolio balance streaming and compliance audit trails.
- **Proof-of-Reserve Registry (Prompt 308):** Regularly verifies that total circulating `weINR.totalSupply()` exactly matches verified sovereign CBDC deposits held in RBI nodal escrow accounts.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Executes on Hyperledger Besu private consortium network with QBFT consensus, 2-second deterministic block times, and immediate finality without chain reorganizations.
- **1:1 Sovereign Backing Invariant:** `weINR` tokens can only be minted when an equivalent amount of sovereign eINR-W or eINR-R is verified as locked in the RBI designated nodal escrow account. Total circulating supply `weINR.totalSupply()` must strictly equal or remain less than total verified sovereign CBDC reserves.
- **Sub-Paise Token Precision ($10^{-4}$ INR):** `weINR` utilizes 4 decimals ($10^{-4}$ INR / 0.01 paise per unit) to eliminate rounding loss in micro-fee splits, high-volume fractional security trading, and statutory tax deductions.
- **Zero On-Chain PII:** Smart contract state stores only `bytes32 messageId`, `bytes32 utrHash`, `bytes32 isinHash`, `address investorWallet`, `uint256 amountSubUnits`, and cryptographic public key signatures. No legal names, PAN numbers, bank account numbers, or mobile numbers are stored on-chain.
- **Multi-Party Sovereign Governance:** Administrative modifications (updating RBI verification node sets, adjusting bridge rate limits, whitelisting bank custodians, updating fee treasury vaults) require 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:** Initialize `contracts/cbdc/RBIeINRBridge.sol`, `contracts/cbdc/WrappedeINR.sol`, interfaces under `contracts/interfaces/cbdc/`, and test suites under `test/cbdc/`.
2. **Define Data Models and Interfaces:** Write `IWrappedeINR.sol` and `IRBIeINRBridge.sol` declaring all enums (`CBDCTier`, `DepositStatus`, `RedemptionStatus`, `DvPSettlementType`), structs, custom errors, and events.
3. **Implement Wrapped eINR Token Contract (`WrappedeINR.sol`):**
   - Implement ERC-20 standard with `decimals() = 4` ($10^{-4}$ precision).
   - Implement restricted `mint(address to, uint256 amount)` and `burn(address from, uint256 amount)` functions callable solely by the authorized `RBIeINRBridge` contract.
   - Implement `IIdentityRegistry` transfer hooks to enforce investor KYC validation on all peer-to-peer transfers.
   - Implement statutory regulatory functions (`freezeAccount`, `unfreezeAccount`, `seizeFrozenFunds`) restricted to `DEFAULT_ADMIN_ROLE` backed by judicial/RBI court order hashes.
4. **Implement ISO 20022 Message Hash Verification in Bridge:**
   - Define message verification structs for `pacs.008.001.10` (Retail Customer Credit Transfer) and `pacs.009.001.10` (Wholesale Financial Institution Credit Transfer).
   - Implement deterministic cryptographic message hashing: `keccak256(abi.encode(messageId, endToEndId, uetr, tier, debtorAgent, creditorAgent, amountSubUnits, timestamp, salt))`.
   - Maintain mapping `mapping(bytes32 => bool) isMessageProcessed` to guarantee strict replay protection.
5. **Implement Multi-Party Threshold Signature Verification:**
   - Define EIP-712 domain separator: `EIP712("GrowwwRBIeINRBridge", "1")`.
   - Implement ECDSA signature recovery validating that at least `requiredThreshold` valid, unique signatures are provided from whitelisted RBI and custodian verification nodes.
6. **Implement Bi-Directional Minting Lifecycle:**
   - In `RBIeINRBridge.sol`, implement `mintWrappedeINRPacs008` (for retail deposits) and `mintWrappedeINRPacs009` (for wholesale deposits).
   - Validate recipient investor KYC compliance via `IIdentityRegistry`.
   - Call `WrappedeINR.mint(investor, amountSubUnits)` and emit `CBDCDepositMinted`.
7. **Implement Bi-Directional Redemption & Burn Lifecycle:**
   - Implement `requestCBDCRedemption` allowing investors to lock and burn `weINR` for fiat CBDC payout to their designated RBI nodal wallet.
   - Transfer `weINR` to bridge escrow and execute `WrappedeINR.burn(address(this), amountSubUnits)`.
   - Generate off-chain redemption ticket and emit `CBDCRedemptionBurned` containing `redemptionId`, `beneficiaryHash`, and `amountSubUnits`.
   - Implement `confirmCBDCOffChainDisbursement` callable by custodian relayers to mark the redemption lifecycle finalized.
8. **Implement DvP Type 1 Instant Atomic Settlement Engine:**
   - Implement `settleDvPType1Atomic` accepting paired trade parameters: buyer address, seller address, security token address (ERC-3643), token quantity, `weINR` consideration amount, and trade execution nonce.
   - Atomically execute simultaneous transfer: `weINR` from buyer to seller (net of 0.00% fee (No fee at all)) and security tokens from seller to buyer in a single transaction.
   - Enforce all-or-nothing atomicity: if either cash or asset leg transfer fails, revert entire transaction state.
9. **Implement Fixed Transaction Fee & Multi-Vault Allocation Engine:**
   - Compute gross trade consideration: `turnoverSubUnits = considerationAmount`.
   - Calculate fixed platform fee: `feeSubUnits = (turnoverSubUnits * feeBps) / 10000 (where feeBps == 0 at launch)` (exact 0.00% (Zero Fee) / 0 bps (0.00% fee at launch)).
   - Distribute fee: routed dynamically per FeeController governance (0.00% at launch).
   - Record tax compliance leaf hash for off-chain FIFO capital gains reporting (Section 111A/112A).
10. **Implement Rate Limiting and Dynamic Velocity Controls:**
    - Implement per-block and per-epoch (24-hour) mint/burn volume caps.
    - Implement per-tier velocity limits (`maxRetailSingleDepositLimit = 5,00,000 INR`, `maxWholesaleSingleDepositLimit = 500,00,00,000 INR`).
11. **Implement Circuit Breakers and Governance Controls:**
    - Inherit OpenZeppelin `PausableUpgradeable` and `AccessControlUpgradeable`.
    - Restrict emergency pause, unpause, node management, and velocity limit adjustments to `DEFAULT_ADMIN_ROLE` (`MultiSigGovernance.sol`).
12. **Author Comprehensive Foundry Unit and Invariant Tests:**
    - Test `pacs.008` and `pacs.009` message hash verification and duplicate replay rejection.
    - Test threshold signature verification across varying M-of-N custodian quorum sizes (e.g. 3-of-5, 4-of-7).
    - Test DvP Type 1 atomic settlement ensuring simultaneous asset and cash leg transfer with exact 0.00% fee (No fee at all) deduction.
    - Invariant test: `weINR.totalSupply()` must equal cumulative verified minted deposits minus cumulative burned redemptions.
    - Invariant test: fee distribution must mathematically equal 0.00% fee at launch (governed by FeeController.sol) without unit truncation loss.
13. **Perform Slither Static Analysis and Formal Verification:**
    - Run `slither contracts/cbdc/` and verify zero high/medium security vulnerabilities.
    - Verify absence of reentrancy vectors in minting, burning, and DvP settlement routines.
14. **Deploy and Validate on Hyperledger Besu Devnet:**
    - Deploy UUPS proxy implementations using Foundry script `script/DeployCBDCBridge.s.sol`.
    - Connect with mock RBI CBDC relayer nodes and execute end-to-end wholesale minting, retail DvP settlement, and redemption burning.

## Interfaces / Contracts

### 1. Wrapped eINR Interface (`IWrappedeINR.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title IWrappedeINR
 * @notice Interface for the Wrapped Central Bank Digital Currency (weINR) token on Hyperledger Besu.
 * @dev Implements 4 decimal places (10^-4 INR / sub-paise precision) and compliance hooks.
 */
interface IWrappedeINR is IERC20 {
    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event MintedByBridge(
        address indexed bridge,
        address indexed recipient,
        uint256 amountSubUnits,
        bytes32 indexed messageHash
    );

    event BurnedByBridge(
        address indexed bridge,
        address indexed account,
        uint256 amountSubUnits,
        bytes32 indexed redemptionId
    );

    event AccountFrozen(address indexed account, bytes32 indexed legalOrderHash);
    event AccountUnfrozen(address indexed account, bytes32 indexed legalOrderHash);
    event FundsSeized(address indexed account, address indexed recipient, uint256 amountSubUnits, bytes32 legalOrderHash);

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error UnauthorizedCaller(address caller);
    error AccountIsFrozen(address account);
    error InvalidPrecisionAmount(uint256 amount);
    error IdentityComplianceCheckFailed(address account);
    error InsufficientFreeBalance(address account, uint256 available, uint256 requested);
    error ZeroAddressNotAllowed();

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function mint(
        address recipient,
        uint256 amountSubUnits,
        bytes32 messageHash
    ) external;

    function burn(
        address account,
        uint256 amountSubUnits,
        bytes32 redemptionId
    ) external;

    function freezeAccount(address account, bytes32 legalOrderHash) external;

    function unfreezeAccount(address account, bytes32 legalOrderHash) external;

    function seizeFrozenFunds(
        address account,
        address recipient,
        uint256 amountSubUnits,
        bytes32 legalOrderHash
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function decimals() external view returns (uint8);
    function isFrozen(address account) external view returns (bool);
    function bridgeContract() external view returns (address);
    function identityRegistry() external view returns (address);
}
```

### 2. RBI CBDC eINR Bridge Interface (`IRBIeINRBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IRBIeINRBridge
 * @notice Interface for the RBI CBDC Wholesale & Retail Programmable Bridge.
 * @dev Manages ISO 20022 message validation, M-of-N threshold verifications, DvP Type 1 settlement, and 0.00% fee (No fee at all) split.
 */
interface IRBIeINRBridge {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum CBDCTier {
        RETAIL,     // eINR-R: Retail digital rupee (pacs.008 message format)
        WHOLESALE   // eINR-W: Wholesale digital rupee (pacs.009 message format)
    }

    enum DepositStatus {
        NONE,
        PENDING_CONFIRMATION,
        MINTED,
        REJECTED
    }

    enum RedemptionStatus {
        NONE,
        REQUESTED_LOCKED,
        BURNED_PENDING_DISBURSEMENT,
        DISBURSED_CONFIRMED,
        CANCELLED_REFUNDED
    }

    enum DvPSettlementType {
        TYPE_1_ATOMIC_GROSS, // Instant gross simultaneous cash and asset leg settlement
        TYPE_2_NET_CASH      // Gross security leg with batched net cash leg
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct ISO20022Pacs008Payload {
        bytes32 messageId;            // Unique ISO 20022 Message Identifier
        bytes32 endToEndId;           // End-to-End identification reference
        bytes32 uetr;                 // Unique End-to-end Transaction Reference (UETR)
        address recipientWallet;      // On-chain recipient ledger address
        uint256 amountSubUnits;       // Amount in sub-paise (10^-4 INR precision)
        bytes32 debtorAgentBic;       // Originating bank BIC / IFSC hash
        bytes32 creditorAgentBic;     // Destination settlement bank BIC hash
        uint64 settlementTimestamp;   // ISO 20022 settlement timestamp
        uint256 nonce;                // Replay prevention nonce
    }

    struct ISO20022Pacs009Payload {
        bytes32 messageId;            // Unique ISO 20022 Message Identifier
        bytes32 transactionId;        // Interbank Transaction Identification
        bytes32 uetr;                 // Unique End-to-end Transaction Reference
        address institutionalWallet;  // On-chain institutional participant address
        uint256 amountSubUnits;       // Amount in sub-paise (10^-4 INR precision)
        bytes32 debtorInstBic;        // Originating Financial Institution BIC
        bytes32 creditorInstBic;      // Creditor Financial Institution BIC
        uint64 settlementTimestamp;   // Interbank RTGS/CBDC settlement timestamp
        uint256 nonce;                // Replay prevention nonce
    }

    struct RedemptionRequest {
        bytes32 redemptionId;         // Unique redemption identifier
        address investorWallet;       // On-chain investor address requesting payout
        uint256 amountSubUnits;       // Amount of weINR burned
        CBDCTier tier;                // RETAIL (eINR-R) or WHOLESALE (eINR-W)
        bytes32 rbiNodalWalletHash;   // Hash of destination RBI CBDC wallet identifier
        RedemptionStatus status;      // Current status of redemption lifecycle
        uint64 requestedTimestamp;    // Timestamp when burn was executed
        uint64 finalizedTimestamp;    // Timestamp when bank disbursement was confirmed
    }

    struct DvPTradeOrder {
        bytes32 tradeId;              // Unique trade identifier
        address buyer;                // Buyer address (providing weINR cash leg)
        address seller;               // Seller address (providing security tokens)
        address securityToken;        // Address of ERC-3643 security token
        uint256 tokenUnits;           // Number of security tokens to transfer (18 decimals)
        uint256 considerationSubUnits;// Gross weINR consideration (4 decimals)
        uint256 sellerCostBasis;      // Seller cost basis in sub-units for tax compliance
        uint64 tradeTimestamp;        // Off-chain matching timestamp
        uint256 nonce;                // Replay prevention nonce
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event CBDCDepositMinted(
        bytes32 indexed messageHash,
        CBDCTier indexed tier,
        address indexed recipient,
        uint256 amountSubUnits,
        bytes32 uetr
    );

    event CBDCRedemptionBurned(
        bytes32 indexed redemptionId,
        CBDCTier indexed tier,
        address indexed investor,
        uint256 amountSubUnits,
        bytes32 rbiNodalWalletHash
    );

    event CBDCRedemptionDisbursed(
        bytes32 indexed redemptionId,
        bytes32 indexed bankDisbursementRefHash,
        uint64 finalizedTimestamp
    );

    event DvPTradeSettledAtomic(
        bytes32 indexed tradeId,
        address indexed buyer,
        address indexed seller,
        address securityToken,
        uint256 tokenUnits,
        uint256 grossConsiderationSubUnits,
        uint256 platformFeeSubUnits,
        bytes32 taxProofHash
    );

    event PlatformFeeDistributed(
        bytes32 indexed tradeId,
        uint256 totalFeeSubUnits,
        uint256 treasuryShareSubUnits,
        uint256 coreSgfShareSubUnits,
        uint256 ipfShareSubUnits
    );

    event VerificationNodeAdded(address indexed node, string institutionName);
    event VerificationNodeRemoved(address indexed node);
    event VerificationThresholdUpdated(uint256 previousThreshold, uint256 newThreshold);
    event VelocityLimitsUpdated(CBDCTier tier, uint256 singleLimit, uint256 dailyLimit);

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error MessageAlreadyProcessed(bytes32 messageHash);
    error InvalidThresholdSignatures(uint256 provided, uint256 required);
    error DuplicateSignerDetected(address signer);
    error SignerNotAuthorizedNode(address signer);
    error InvalidMessagePayload(string reason);
    error StaleMessageTimestamp(uint64 timestamp, uint64 thresholdWindow);
    error VelocityLimitExceeded(CBDCTier tier, uint256 requested, uint256 allowed);
    error RedemptionNotFound(bytes32 redemptionId);
    error InvalidRedemptionState(bytes32 redemptionId, RedemptionStatus current);
    error DvPAssetTransferFailed(address token, address from, address to, uint256 amount);
    error DvPCashTransferFailed(address from, address to, uint256 amount);
    error FeeCalculationMismatch(uint256 expectedFee, uint256 actualFee);
    error NonceAlreadyUsed(bytes32 identifier, uint256 nonce);

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function mintWrappedeINRPacs008(
        ISO20022Pacs008Payload calldata payload,
        bytes[] calldata thresholdSignatures
    ) external returns (bytes32 messageHash);

    function mintWrappedeINRPacs009(
        ISO20022Pacs009Payload calldata payload,
        bytes[] calldata thresholdSignatures
    ) external returns (bytes32 messageHash);

    function requestCBDCRedemption(
        uint256 amountSubUnits,
        CBDCTier tier,
        bytes32 rbiNodalWalletHash
    ) external returns (bytes32 redemptionId);

    function confirmCBDCOffChainDisbursement(
        bytes32 redemptionId,
        bytes32 bankDisbursementRefHash,
        bytes[] calldata thresholdSignatures
    ) external;

    function settleDvPType1Atomic(
        DvPTradeOrder calldata tradeOrder,
        bytes calldata buyerSignature,
        bytes calldata sellerSignature
    ) external returns (uint256 platformFeePaid);

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function wrappedeINR() external view returns (address);
    function getThresholdParameters() external view returns (uint256 threshold, uint256 totalNodes);
    function isVerificationNode(address node) external view returns (bool);
    function isMessageProcessed(bytes32 messageHash) external view returns (bool);
    function getRedemption(bytes32 redemptionId) external view returns (RedemptionRequest memory);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function computeTransactionFee(
        uint256 grossTurnoverSubUnits
    ) external pure returns (uint256 totalFee, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare);
}
```

## Security & Compliance Notes
- **100% Reserve Parity Invariant:** `weINR` tokens can only be minted against cryptographically verified ISO 20022 `pacs.008` (retail) or `pacs.009` (wholesale) settlement receipts signed by authorized RBI nodal verification nodes. Total on-chain supply must strictly remain less than or equal to off-chain sovereign CBDC reserves held in designated escrow accounts.
- **Fixed Platform Fee Integrity:** In strict accordance with the Growww Core Fee Model (Prompt 006), the platform fee is assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover with an automated 0.00% fee at launch (governed by FeeController.sol) revenue split. Capital gains are computed strictly for off-chain user tax compliance (Section 111A/112A). Zero holding or custody fees are charged.
- **M-of-N Cryptographic Quorum:** All bridge minting and redemption disbursement operations require a minimum threshold quorum of valid ECDSA signatures (e.g. 3-of-5) from distinct, whitelisted institutional verification nodes running FIPS 140-2 Level 3 HSMs.
- **Replay Protection & Idempotency:** Every ISO 20022 message payload is hashed deterministically (`messageHash`) and permanently committed in storage. Duplicate message submissions with matching hashes or reused nonces are rejected immediately.
- **Atomic DvP Settlement Safety:** DvP Type 1 settlement uses Checks-Effects-Interactions (CEI) patterns and non-reentrant execution hooks to ensure simultaneous transfer of both cash and security legs. Any settlement failure on either leg reverts the entire transaction state atomically.
- **Zero On-Chain PII:** The smart contract stores only cryptographic hashes (`messageId`, `uetr`, `rbiNodalWalletHash`, `taxProofHash`) and public Ethereum addresses. All personally identifiable information (PII) including Aadhaar, PAN, bank account numbers, and retail CBDC wallet IDs remains off-chain in encrypted compliance vaults.
- **Statutory Court & RBI Freezing Hooks:** `WrappedeINR.sol` contains specialized judicial freezing functions (`freezeAccount`, `seizeFrozenFunds`) restricted strictly to multi-sig administrative governance requiring verified Indian judicial / RBI regulatory attachment order hashes.

## Acceptance Criteria
- [ ] `IWrappedeINR.sol` and `IRBIeINRBridge.sol` compiled under Solidity 0.8.24 with zero warnings.
- [ ] `WrappedeINR.sol` implements ERC-20 standard with 4 decimal places ($10^{-4}$ precision) and strictly restricts `mint()` and `burn()` to the authorized bridge contract.
- [ ] `RBIeINRBridge.sol` validates structured ISO 20022 `pacs.008` (retail) and `pacs.009` (wholesale) message payloads.
- [ ] M-of-N threshold signature verification validates distinct authorized signers, rejecting duplicate signatures or unapproved nodes.
- [ ] Replay protection mapping prevents reprocessing of previously minted ISO 20022 message hashes.
- [ ] Off-chain redemption burn flow locks and destroys `weINR`, emitting structured events for off-chain bank relayer disbursement.
- [ ] DvP Type 1 settlement executes simultaneous atomic swap of `weINR` and ERC-3643 security tokens with complete reversion on partial leg failure.
- [ ] Platform fee computation mathematically enforces exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) deduction on trade turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol).
- [ ] Invariant tests in Foundry prove across 10,000 runs that `weINR.totalSupply()` never exceeds verified deposited sovereign reserves.
- [ ] Slither and Mythril static analysis suites pass with zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across entire specification documentation.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract - ERC-3643), Prompt `305` (Transfer Compliance Hooks), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `203` (Double-Entry Ledger Service), Prompt `208` (Trade Settlement Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `213` (Custodian Depository Integration Service), Prompt `242` (ISO 20022 Banking Gateway).
- **Subsequent Prompts Enabled:** Prompt `308` (Proof of Reserve Registry), Prompt `309` (Event Indexing Service), Prompt `331` (Primary Market Order Routing & Clearing Bridge), Prompt `509` (Flutter Order Placement Flow).
