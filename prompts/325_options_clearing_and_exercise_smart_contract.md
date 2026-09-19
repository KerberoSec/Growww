# 325 - Options Clearing, Collateral Lockup & Automated In-The-Money Exercise Smart Contracts (OptionsTokenFactory.sol, OptionsClearingHouse.sol, PhysicalAndCashSettler.sol)

## Purpose
In modern electronic financial markets, equity derivatives (Call and Put options) provide essential price discovery, risk hedging, and capital allocation mechanisms for retail and institutional participants. In traditional clearinghouses, options clearing and exercise pipelines depend on batch-processed end-of-day cycles, manual physical delivery notices, discretionary broker margin calls, and opaque settlement fees. These centralized processes introduce counterparty credit risk, settlement delay, and operational friction.

The **Options Clearing, Collateral Lockup & Automated In-The-Money Exercise Smart Contract Suite (`OptionsTokenFactory.sol`, `OptionsClearingHouse.sol`, `PhysicalAndCashSettler.sol`)** establishes an institutional-grade, programmable derivatives infrastructure on Hyperledger Besu (QBFT consensus). It tokenizes European-style and American-style Call and Put option contracts as compliant ERC-1155 multi-tokens, locks 100% collateral in segregated on-chain vaults, ingests cryptographically signed settlement price oracles at expiry, executes automated batch in-the-money (ITM) exercises without requiring manual holder transaction submission, unlocks out-of-the-money (OTM) collateral back to writers, enforces Growww's canonical Fixed 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on trade notional turnover with a 0.00% fee at launch (governed by FeeController.sol) revenue split, and records cryptographic attestations for FIFO capital gains computed strictly for user tax compliance (Section 111A/112A).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/derivatives/` comprising:
- `OptionsTokenFactory.sol`: Core ERC-1155 multi-token factory that standardizes, deploys, mints, and burns option series contracts. It computes deterministic Token IDs encoding underlying ISIN, option type (CALL/PUT), strike price, expiration timestamp, and settlement type (Physical vs Cash).
- `OptionsClearingHouse.sol`: Central on-chain clearinghouse and collateral management vault. It enforces 100% full collateralization for option writers (underlying ERC-3643 tokens for covered Calls; tokenized eINR / CBDC cash for cash-secured Puts), tracks position balances, ingests verified expiration benchmark prices via EIP-712 oracle signatures, and coordinates automated batch expiration settlement.
- `PhysicalAndCashSettler.sol`: Modular settlement engine executing atomic physical delivery (swapping underlying ERC-3643 equity tokens against strike cash) or cash settlement (calculating intrinsic payoff value and transferring net cash from writer collateral to holders). It assesses the canonical 0.00% platform fee (No fee at all) on trade notional turnover, splits fee proceeds 0.00% fee at launch (governed by FeeController.sol), records FIFO capital gains data strictly for user tax compliance (Section 111A/112A), and refunds unencumbered collateral to option writers.
- `IOptionsTokenFactory.sol`, `IOptionsClearingHouse.sol`, `IPhysicalAndCashSettler.sol`: Complete Solidity interfaces specifying all structs, enums, custom errors, events, view functions, and state-modifying function signatures.
- Comprehensive Foundry test harness (`test/derivatives/`): Unit, fuzz, invariant, and integration test suites validating 100% collateral coverage, zero-reentrancy resistance, automated ITM/OTM state transitions, and mathematical accuracy of the Universal Zero-Fee Model (0.00% fee - No fee at all) deduction and 0.00% fee at launch (future fee parameters governed by FeeController.sol).

## Scope Boundaries
- **In Scope:**
  - ERC-1155 multi-token issuance, minting, batch transfer, and burning for Call and Put options.
  - Deterministic series ID generation and registry indexing across underlying asset ISINs.
  - 100% collateral lockup validation prior to option minting (zero uncovered writing).
  - Integration with `IdentityRegistry.sol` (Prompt 303 / Prompt 305) for investor KYC compliance on secondary transfers.
  - Multi-party EIP-712 signed oracle settlement price ingestion at expiration cutoff.
  - Automated batch In-The-Money (ITM) exercise without requiring active holder transaction submission.
  - Automated Out-of-The-Money (OTM) expiration and collateral unlocking for option writers.
  - Physical Delivery Settlement: Atomic transfer of ERC-3643 shares against strike cash.
  - Cash Settlement: Transfer of intrinsic payoff value ($\Delta = |P_{\text{settle}} - K| \times \text{contracts}$) in tokenized eINR / CBDC.
  - Fixed 0.00% (Zero Fee) (0 bps (0.00% fee at launch) / 0 bps at launch) platform transaction fee assessment on turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), and FIFO capital gains computation for tax compliance (Section 111A/112A).
  - Circuit breakers, emergency pause controls, and timelocked multi-signature governance.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain high-frequency order matching and order book management (handled in Prompt 205).
  - Real-time pre-trade Black-Scholes Greeks, implied volatility, and portfolio VaR margin checks (handled in Prompt 206 and Prompt 229).
  - Off-chain settlement price feed aggregation, medianization, and signing relayer (handled in Prompt 207).
  - SGF default waterfall and clearing corporation mutualized risk resolution (handled in Prompt 315).
  - Domestic fiat banking payment rail execution and UPI hold reservations (handled in Prompt 203 and Prompt 212).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Cancun/Shanghai with native checked arithmetic and custom errors).
- **Token Standards:** **ERC-1155** (Multi-Token Standard) for gas-efficient multi-strike option series management; **ERC-3643 / ERC-20** for underlying digital securities and tokenized eINR cash.
- **Security & Standards:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Signature Verification:** OpenZeppelin `ECDSA` and `EIP712Upgradeable` for tamper-proof oracle price attestation.
- **Development & Testing Framework:** **Foundry** (`forge`, `cast`) for deterministic compilation, unit testing, property fuzzing, and stateful invariant verification.
- **Static Analysis & Security:** Slither, Mythril, and Solhint automated CI analysis pipelines.

## Backend / Infra Touchpoints
- **Market Data & Oracle Relayer (Prompt 207):** Pushes EIP-712 cryptographically signed official underlying closing settlement prices at expiration cutoff (15:30 IST / designated expiry timestamp).
- **Trade Settlement Service (Prompt 208):** Orchestrates batch calls to `OptionsClearingHouse.sol` and `PhysicalAndCashSettler.sol` for automated expiration processing.
- **Wallet & Account Service (Prompt 203):** Reconciles on-chain eINR / CBDC collateral movements and option premium transfers with domestic fiat ledgers.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes on-chain fixed 0.00% transaction fee (No fee at all) deductions (with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol)) and FIFO capital gains tax records (Prompt 223) with platform accounting and treasury ledgers.
- **Blockchain Event Indexer (Prompt 309):** Ingests events (`OptionSeriesCreated`, `OptionMinted`, `CollateralLocked`, `OptionExercised`, `OptionExpiredOTM`, `PlatformFeeDeducted`) to update off-chain PostgreSQL databases.
- **Custodian Depository Integration (Prompt 213):** Verifies that underlying digital equity tokens held as Call collateral map 1:1 to physical demat accounts at NSDL/CDSL.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Operates on Hyperledger Besu private permissioned network utilizing QBFT consensus with 2-second deterministic block finality and zero transaction reordering.
- **Direct Asset Movement:** Locks and transfers underlying equity tokens (`IERC3643`) and settlement currency (`IERC20` eINR) via `SafeERC20` with strict checks-effects-interactions ordering.
- **Zero On-Chain PII:** The smart contracts record only `bytes32 seriesId`, `uint256 tokenId`, `address writer`, `address holder`, `uint256 strikePrice`, `uint256 amount`, `uint256 costBasis`, and cryptographic signature hashes.
- **100% Solvency Invariant:** At all times, the total locked collateral held in `OptionsClearingHouse.sol` must equal or exceed the maximum potential payoff obligation of all unexercised active options.
- **Multi-Sig Governance:** High-privilege actions (contract upgrades, parameter changes, oracle whitelists, emergency pauses) require 3-of-5 threshold authorization from `MultiSigGovernance.sol` (Prompt 307) backed by HSM validator keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:** Initialize contracts under `contracts/derivatives/`, interfaces under `contracts/interfaces/derivatives/`, test suites under `test/derivatives/`, and deployment scripts under `script/derivatives/`.
2. **Define Data Structures & Enums:** Define enums (`OptionType`, `SettlementType`, `SeriesState`, `ExerciseType`) and structs (`OptionSeriesParams`, `CollateralPosition`, `SettlementPriceData`, `ExerciseReceipt`) in `IOptionsTokenFactory.sol`, `IOptionsClearingHouse.sol`, and `IPhysicalAndCashSettler.sol`.
3. **Implement ERC-1967 / UUPS Upgradeability:** Configure `OptionsTokenFactory.sol`, `OptionsClearingHouse.sol`, and `PhysicalAndCashSettler.sol` with OpenZeppelin `UUPSUpgradeable`, `AccessControlUpgradeable`, and `ReentrancyGuardUpgradeable` storage gap reservations.
4. **Implement Deterministic Series Token ID Hashing:** In `OptionsTokenFactory.sol`, implement deterministic ERC-1155 `tokenId` calculation:
   $$\text{tokenId} = \text{uint256}(\text{keccak256}(\text{abi.encode}(\text{underlyingToken}, \text{optionType}, \text{strikePrice}, \text{expiryTimestamp}, \text{settlementType})))$$
5. **Implement Series Creation & Parameter Whitelisting:** Build `createOptionSeries(...)` in `OptionsTokenFactory.sol`, validating that `underlyingToken` is approved, `expiryTimestamp > block.timestamp`, `strikePrice > 0`, and registering the series in an indexed registry.
6. **Implement Full Collateral Locking Module:** In `OptionsClearingHouse.sol`, build `writeAndMintOption(...)`:
   - For `OptionType.CALL`: Transfer `amount` units of `underlyingToken` from writer to the clearinghouse vault.
   - For `OptionType.PUT`: Transfer $(\text{strikePrice} \times \text{amount}) / 10^{18}$ units of `settlementToken` (eINR) from writer to the clearinghouse vault.
   - Record `CollateralPosition` and invoke `OptionsTokenFactory.mintOption(writer, tokenId, amount)`.
7. **Implement Secondary Transfer Compliance Hooks:** In `OptionsTokenFactory.sol`, override `_update(...)` (or `safeTransferFrom` / `safeBatchTransferFrom`) to call `IIdentityRegistry(identityRegistry).isVerified(to)` and `ICompliance(compliance).canValidateTokenAction(...)`, ensuring non-KYC accounts cannot receive option tokens.
8. **Implement EIP-712 Expiration Oracle Ingestion:** In `OptionsClearingHouse.sol`, build `submitSettlementPrice(bytes32 seriesId, SettlementPriceData calldata data, bytes calldata signature)`:
   - Validate `block.timestamp >= series.expiryTimestamp`.
   - Verify EIP-712 structured data signature against whitelisted oracle relayer public keys.
   - Store settlement price and transition series state from `ACTIVE` to `EXPIRED_PRICE_LOCKED`.
9. **Implement Automated In-The-Money (ITM) Evaluation Logic:** In `OptionsClearingHouse.sol`, implement ITM classification:
   - For Calls: ITM if $P_{\text{settle}} > K$; intrinsic payoff per unit is $P_{\text{settle}} - K$.
   - For Puts: ITM if $P_{\text{settle}} < K$; intrinsic payoff per unit is $K - P_{\text{settle}}$.
   - If intrinsic payoff is zero or negative, mark series as `EXPIRED_WORTHLESS_OTM`.
10. **Implement Cash Settlement Engine:** In `PhysicalAndCashSettler.sol`, build `settleCash(...)`:
    - Calculate aggregate intrinsic payoff: $\text{Total Payoff} = \text{intrinsicPayoffPerUnit} \times \text{contractAmount}$.
    - Calculate fixed platform fee: $\text{Fee} = (\text{Total Payoff} \times 1) / 10000$ (0.00% (Zero Fee) of notional settlement turnover).
    - Route fee according to 0.00% fee at launch (governed by FeeController.sol) allocation.
    - Transfer $(\text{Total Payoff} - \text{Fee})$ to option holder wallet; transfer $\text{Fee}$ to designated multi-vault reserves.
    - Record tax compliance leaf hash for Section 111A/112A capital gains reporting.
    - Release remaining unused collateral back to writer wallet.
11. **Implement Physical Delivery Settlement Engine:** In `PhysicalAndCashSettler.sol`, build `settlePhysical(...)`:
    - For Calls: Pull strike cash $(\text{strikePrice} \times \text{amount}) / 10^{18}$ from holder, transfer underlying equity tokens from clearinghouse vault to holder, transfer strike cash to writer, and deduct 0.00% (Zero Fee) platform fee on trade turnover.
    - For Puts: Pull underlying equity tokens from holder, transfer strike cash from clearinghouse vault to holder, transfer underlying equity tokens to writer, and deduct 0.00% (Zero Fee) platform fee on trade turnover.
12. **Implement Automated Batch Expiration Execution:** In `OptionsClearingHouse.sol`, build `batchAutoExerciseAndSettle(bytes32 seriesId, address[] calldata holders, uint256[] calldata amounts, uint256[] calldata costBases)` enabling clearing relayers to process up to 100 ITM holder positions in a single atomic Besu transaction.
13. **Implement OTM Collateral Reclamation:** In `OptionsClearingHouse.sol`, build `reclaimOTMCollateral(bytes32 seriesId, uint256 positionId)`:
    - Verify series is marked `EXPIRED_WORTHLESS_OTM`.
    - Burn expired ERC-1155 tokens and release 100% of locked collateral back to writer.
14. **Implement Circuit Breakers & Multi-Sig Governance:** Add emergency pause/unpause functions guarded by `EMERGENCY_GUARDIAN_ROLE`, and bind parameter updates to 48-hour timelocked multi-sig controls.
15. **Build Extensive Foundry Test Suite & Static Scans:** Implement unit, fuzz, and invariant test suites (`test/derivatives/OptionsClearing.t.sol`), verify zero reentrancy with Slither, and confirm 100% test branch coverage.

## Interfaces / Contracts

### 1. Options Token Factory Interface (`IOptionsTokenFactory.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IOptionsTokenFactory {
    enum OptionType {
        CALL,
        PUT
    }

    enum SettlementType {
        PHYSICAL_DELIVERY,
        CASH_SETTLED
    }

    enum SeriesState {
        UNINITIALIZED,
        ACTIVE,
        EXPIRED_PRICE_LOCKED,
        EXPIRED_WORTHLESS_OTM,
        FULLY_SETTLED,
        CANCELLED
    }

    struct OptionSeriesParams {
        bytes32 seriesId;
        address underlyingToken;
        address settlementToken;
        OptionType optionType;
        SettlementType settlementType;
        uint256 strikePrice;
        uint256 expiryTimestamp;
        bool isAmerican;
        SeriesState state;
    }

    // Events
    event OptionSeriesCreated(
        bytes32 indexed seriesId,
        uint256 indexed tokenId,
        address indexed underlyingToken,
        address settlementToken,
        OptionType optionType,
        SettlementType settlementType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        bool isAmerican
    );
    event OptionsMinted(bytes32 indexed seriesId, uint256 indexed tokenId, address indexed recipient, uint256 amount);
    event OptionsBurned(bytes32 indexed seriesId, uint256 indexed tokenId, address indexed holder, uint256 amount);
    event SeriesStateUpdated(bytes32 indexed seriesId, SeriesState indexed oldState, SeriesState indexed newState);

    // Custom Errors
    error SeriesAlreadyExists(bytes32 seriesId);
    error SeriesDoesNotExist(bytes32 seriesId);
    error InvalidUnderlyingToken(address token);
    error InvalidSettlementToken(address token);
    error InvalidStrikePrice();
    error InvalidExpiryTimestamp(uint256 provided, uint256 current);
    error UnauthorizedCaller(address caller);
    error NonKYCRecipient(address recipient);
    error SeriesNotActive(bytes32 seriesId, SeriesState currentState);

    // Factory Functions
    function createOptionSeries(
        address underlyingToken,
        address settlementToken,
        OptionType optionType,
        SettlementType settlementType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        bool isAmerican
    ) external returns (bytes32 seriesId, uint256 tokenId);

    function mintOption(address recipient, uint256 tokenId, uint256 amount) external;
    function burnOption(address holder, uint256 tokenId, uint256 amount) external;
    function updateSeriesState(bytes32 seriesId, SeriesState newState) external;

    // View Functions
    function getSeriesParams(bytes32 seriesId) external view returns (OptionSeriesParams memory);
    function getSeriesParamsByTokenId(uint256 tokenId) external view returns (OptionSeriesParams memory);
    function computeTokenId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) external pure returns (uint256);
    function computeSeriesId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) external pure returns (bytes32);
    function isSeriesActive(bytes32 seriesId) external view returns (bool);
}
```

### 2. Options Clearinghouse Interface (`IOptionsClearingHouse.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "./IOptionsTokenFactory.sol";

interface IOptionsClearingHouse {
    struct CollateralPosition {
        uint256 positionId;
        bytes32 seriesId;
        address writer;
        address collateralToken;
        uint256 lockedAmount;
        uint256 mintedContracts;
        uint256 unexercisedContracts;
        bool isReleased;
    }

    struct SettlementPriceData {
        bytes32 seriesId;
        uint256 settlementPrice;
        uint256 priceTimestamp;
        uint256 sequenceNumber;
    }

    struct BatchExerciseParams {
        bytes32 seriesId;
        address[] holders;
        uint256[] contractAmounts;
        uint256[] costBases;
    }

    // Events
    event CollateralDepositedAndOptionMinted(
        uint256 indexed positionId,
        bytes32 indexed seriesId,
        address indexed writer,
        address collateralToken,
        uint256 collateralAmount,
        uint256 contractsMinted
    );
    event SettlementPriceSubmitted(
        bytes32 indexed seriesId,
        uint256 settlementPrice,
        uint256 priceTimestamp,
        address indexed oracleRelayer
    );
    event AutomatedExerciseExecuted(
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 contractAmount,
        uint256 netPayout,
        uint256 platformFeeAmount
    );
    event CollateralUnlocked(uint256 indexed positionId, bytes32 indexed seriesId, address indexed writer, uint256 amountReleased);
    event OracleRelayerUpdated(address indexed oracleRelayer, bool isWhitelisted);

    // Custom Errors
    error InsufficientCollateralProvided(uint256 required, uint256 provided);
    error ExpiryNotReached(uint256 currentTimestamp, uint256 expiryTimestamp);
    error ExpiryAlreadyProcessed(bytes32 seriesId);
    error StaleOraclePrice(uint256 priceTimestamp, uint256 currentTimestamp);
    error InvalidOracleSignature();
    error OracleNotWhitelisted(address relayer);
    error PositionAlreadyReleased(uint256 positionId);
    error InactiveOptionSeries(bytes32 seriesId);
    error ArrayLengthMismatch();
    error OptionNotITM(bytes32 seriesId, uint256 settlementPrice, uint256 strikePrice);

    // Clearing Operations
    function writeAndMintOption(
        bytes32 seriesId,
        uint256 contractAmount
    ) external returns (uint256 positionId);

    function submitSettlementPrice(
        bytes32 seriesId,
        SettlementPriceData calldata data,
        bytes calldata signature
    ) external;

    function batchAutoExerciseAndSettle(
        BatchExerciseParams calldata params
    ) external;

    function reclaimOTMCollateral(
        uint256 positionId
    ) external;

    function reclaimResidualCollateral(
        uint256 positionId
    ) external;

    // View Functions
    function getCollateralPosition(uint256 positionId) external view returns (CollateralPosition memory);
    function getSettlementPrice(bytes32 seriesId) external view returns (uint256 price, uint256 timestamp, bool isLocked);
    function isOptionITM(bytes32 seriesId) external view returns (bool isITM, uint256 intrinsicPayoffPerUnit);
    function totalLockedCollateral(address token) external view returns (uint256);
}
```

### 3. Physical & Cash Settler Interface (`IPhysicalAndCashSettler.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "./IOptionsTokenFactory.sol";

interface IPhysicalAndCashSettler {
    struct SettlementExecutionReceipt {
        bytes32 executionId;
        bytes32 seriesId;
        address holder;
        address writer;
        IOptionsTokenFactory.SettlementType settlementType;
        uint256 contractAmount;
        uint256 grossPayoutOrAssetAmount;
        uint256 feeDeducted;
        uint256 netDisbursed;
        bytes32 taxProofHash;
        uint256 timestamp;
    }

    // Events
    event CashSettlementExecuted(
        bytes32 indexed executionId,
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 contractAmount,
        uint256 grossPayoff,
        uint256 feeDeducted,
        uint256 netPayout,
        bytes32 taxProofHash
    );
    event PhysicalSettlementExecuted(
        bytes32 indexed executionId,
        bytes32 indexed seriesId,
        address indexed holder,
        address indexed writer,
        uint256 contractAmount,
        uint256 underlyingDelivered,
        uint256 cashExchanged,
        uint256 feeDeducted,
        bytes32 taxProofHash
    );
    event PlatformFeeCollected(
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 turnoverAmount,
        uint256 feeAmount,
        address treasuryVault,
        address coreSgfVault,
        address ipfVault
    );
    event FeeVaultsUpdated(address indexed treasury, address indexed coreSgf, address indexed ipf);

    // Custom Errors
    error UnauthorizedClearingHouse(address caller);
    error InvalidFeeVaultAddress();
    error TransferFailed(address token, address from, address to, uint256 amount);
    error MathOverflowOrUnderflow();
    error ZeroContractsSettled();

    // Settlement Operations
    function executeCashSettlement(
        bytes32 seriesId,
        address holder,
        uint256 contractAmount,
        uint256 intrinsicPayoffPerUnit,
        uint256 costBasisTotal,
        address settlementToken
    ) external returns (uint256 netPayout, uint256 feeDeducted);

    function executePhysicalDelivery(
        bytes32 seriesId,
        address holder,
        address writer,
        uint256 contractAmount,
        uint256 strikePrice,
        IOptionsTokenFactory.OptionType optionType,
        uint256 costBasisTotal,
        address underlyingToken,
        address settlementToken
    ) external returns (SettlementExecutionReceipt memory receipt);

    // Mathematical Calculation Views
    function computeTransactionFee(
        uint256 turnoverAmount
    ) external pure returns (uint256 feeAmount, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare);

    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function clearingHouse() external view returns (address);
}
```

## Security & Compliance Notes
- **100% Full Collateralization Invariant:** The clearinghouse strictly prohibits naked or uncovered option writing. Every Call contract is backed by 1:1 locked underlying equity tokens (`IERC3643`) or full notional cash; every Put contract is backed by $(K \times \text{contracts})$ in locked `eINR` settlement tokens before the factory can mint tokens.
- **Fixed Platform Fee Integrity:** In strict accordance with the Growww Fee Model (Prompt 006), platform fees are assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover with a 0.00% fee at launch (governed by FeeController.sol) revenue split. FIFO capital gains calculations are conducted strictly for user tax compliance (Section 111A/112A) and anchored via cryptographic leaf hashes. Zero holding or custody charges are levied.
- **Oracle Manipulation & Front-Running Protections:** Settlement prices are ingested only after `block.timestamp >= series.expiryTimestamp` using EIP-712 structured multi-attestation signatures from whitelisted Oracle Relayers (Prompt 207). Stale prices (> 15 minutes old at submission) or unauthorized signers revert immediately.
- **Reentrancy Protection & CEI Pattern:** All collateral locking, settlement disbursement, and token burn operations strictly follow the Checks-Effects-Interactions pattern and are protected by OpenZeppelin `ReentrancyGuardUpgradeable` modifiers.
- **KYC & Regulatory Compliance Hooks:** Secondary transfers of ERC-1155 option tokens automatically query the `IdentityRegistry.sol` (Prompt 305) to guarantee that only KYC-verified domestic or GIFT City accounts can hold or trade options. Court-ordered freezes or SEBI directives can freeze specific participant holdings via multi-sig `freezeAddress`.
- **Multi-Party Governance & Timelock:** Upgrades to UUPS contract implementations or changes to core settlement parameters (fee vaults, whitelisted oracles) require a 3-of-5 threshold signature from `MultiSigGovernance.sol` (Prompt 307) and a mandatory 48-hour timelock delay.

## Acceptance Criteria
- [ ] `IOptionsTokenFactory.sol`, `IOptionsClearingHouse.sol`, and `IPhysicalAndCashSettler.sol` contracts created and compiled under Solidity 0.8.24 with zero warnings.
- [ ] Deterministic ERC-1155 `tokenId` and `seriesId` generation verified across Call/Put and Physical/Cash series permutations.
- [ ] Covered Call minting strictly verifies and locks 100% underlying `IERC3643` tokens in `OptionsClearingHouse.sol`.
- [ ] Cash-Secured Put minting strictly verifies and locks 100% strike cash (`IERC20` eINR) in `OptionsClearingHouse.sol`.
- [ ] EIP-712 oracle settlement price submission verified with cryptographic ECDSA signature checks and timestamp validation.
- [ ] Automated batch ITM exercise executes atomic settlements for up to 100 positions per block without manual holder intervention.
- [ ] OTM option expiration automatically unlocks 100% unencumbered collateral back to writers.
- [ ] Physical delivery swaps underlying equity tokens against strike cash atomically.
- [ ] Cash settlement transfers exact intrinsic value minus 0.00% (Zero Fee) platform fee to holders with 0.00% fee launch policy revenue split.
- [ ] Mathematical tests confirm 0.00% fee (No fee at all) is assessed accurately on turnover and capital gains are computed for tax compliance (Section 111A/112A).
- [ ] Invariant fuzz tests in Foundry pass 10,000 runs confirming total contract token balance >= aggregate claimable collateral at all times.
- [ ] Slither and Mythril static analysis tools execute with zero high or medium severity vulnerabilities.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Selection), Prompt `303` (Token Issuance Smart Contract - ERC-3643), Prompt `305` (Transfer Compliance Hooks), Prompt `306` (Settlement DvP Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `206` (Risk & Margin Checks Service), Prompt `208` (Trade Settlement Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `229` (Real-Time VaR Margin Engine).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (On-Chain Proof of Reserve), Prompt `309` (Event Indexing Service), Prompt `509` (Flutter Order Placement Flow).
