# 342 - Fractional Share Rights & On-Chain Corporate Action Splitter Contract (Solidity)

## Purpose
In modern electronic financial markets, corporate actions such as stock splits (sub-divisions), reverse stock splits (consolidations), bonus issues (stock dividends), and rights offerings are fundamental corporate finance operations governed by statutory company law and securities regulations. In traditional Indian capital markets regulated by the Securities and Exchange Board of India (SEBI) and administered via central depositories (NSDL and CDSL), corporate actions are executed strictly against whole, integer demat share balances held as of the official Record Date.

On the Growww National Blockchain Stock Exchange (NBSE) deployed on a permissioned Hyperledger Besu consortium ledger, retail investors are empowered to hold and trade arbitrary fractional units of Indian blue-chip equities down to 18 decimal places (e.g., 0.354129 shares of a tokenized equity). When an issuer executes a corporate action (such as a 2:1 forward split, a 1:5 consolidation, or a 1:1 bonus issue), applying split ratios across arbitrary fractional balances introduces micro-token residual fractions. In addition, when corporate actions dictate that fractional entitlements cannot be credited as equity (such as in consolidations where non-integer entitlements must be liquidated as cash-in-lieu), the system must aggregate non-divisible fractional dust with mathematical rigor and route it to an escrow liquidation pool without diluting the physical 1:1 demat custody backing.

This prompt specifies the enterprise implementation of the **Fractional Share Rights & On-Chain Corporate Action Splitter Contract (`CorporateActionSplitter.sol`)**. The contract delivers an automated, verifiable, on-chain execution engine that applies mathematically exact token splits, stock dividends, and fractional cash-in-lieu (CIL) dust allocations directly onto ERC-3643 compliant security tokens (`TokenizedEquity.sol`). It enforces strict ex-date trading freezes in coordination with the on-chain circuit breaker, supports both atomic supply rebasing and batched balance adjustments, and guarantees a zero-dilution invariant verified against official NSDL/CDSL corporate action master records.

## What You Are Building
A production-grade, upgradeable smart contract suite under `contracts/src/tokens/` comprising:
- `CorporateActionSplitter.sol`: Core corporate action execution smart contract managing corporate event state transitions (Scheduled, Trading Halted, Balance Snapshot, Batch Split / Bonus Execution, Dust Allocation, and Finalization).
- `ICorporateActionSplitter.sol`: Complete Solidity interface specifying corporate action data structures, supported action types, execution entrypoints, custom error definitions, and regulatory audit event schemas.
- High-Precision Fixed-Point Ratio Engine: Mathematical calculation module supporting arbitrary corporate action ratios (e.g., forward splits 2:1, 5:1, 10:1; reverse splits 1:2, 1:5; bonus issues 1:1, 1:2, 3:1) utilizing fixed-point math with 10^6 precision scaling factor (`RATIO_PRECISION = 1,000,000`).
- Dual Execution Modalities:
  - **Atomic Global Supply Rebase:** High-efficiency global scale-factor multiplier adjustment on `TokenizedEquity.sol` for uniform forward stock splits (e.g., 2:1 or 10:1), instantly updating all effective investor balances in O(1) complexity.
  - **Batched Fractional Adjustment & Dust Extraction:** Iterative multi-transaction batch processor for non-uniform actions, bonus distributions, and reverse splits, calculating individual entitlements, minting bonus tokens or burning consolidated units, and collecting micro-token residuals.
- Micro-Token Dust & Cash-in-Lieu (CIL) Escrow: Dedicated dust-sweeping logic that aggregates fractional remnants below minimum unit precision or resulting from fractional reverse-split truncation into a designated `CUSTODY_DUST_ESCROW` address for custodial liquidation and fiat INR wallet payout.
- Circuit Breaker Synchronization: Integration with `CircuitBreakerHalt.sol` (Prompt 341) to automatically invoke `haltISIN` during the corporate action freeze window, preventing balance divergence and in-flight transfer race conditions.
- Comprehensive Foundry Test Suite (`test/tokens/CorporateActionSplitter.t.sol`): Invariant fuzzing, boundary division testing, dust conservation proofs, and role-based authorization verification.

## Scope Boundaries
- **In Scope:**
  - On-chain lifecycle state machine for corporate actions: `SCHEDULED`, `FROZEN`, `EXECUTING`, `COMPLETED`, and `CANCELLED`.
  - Registration of corporate actions with official parameters: corporate action ID, ISIN, action type, ratio numerator and denominator, ex-date timestamp, record-date cutoff, and depository authorization hash.
  - Mathematical ratio calculation engine with 10^6 fixed-point scaling factor and explicit rounding control (floor vs ceil).
  - Direct interaction with `TokenizedEquity.sol` / `DigitalSecurityToken.sol` (ERC-3643) for atomic rebase, batch bonus minting, and batch consolidation burning.
  - Fractional dust isolation and aggregation into custodial escrow for Cash-in-Lieu (CIL) processing.
  - Role-based access control protecting corporate action declaration, snapshot triggers, execution, and rollback.
  - Zero share dilution invariant enforcement ensuring post-split token supply plus aggregated dust matches demat custody reserves.
  - Rich audit logging emitting events for indexers and back-office reconciliation.
- **Out of Scope / Handled Elsewhere:**
  - Automated ingestion and parsing of corporate action announcements from BSE, NSE, and NSDL/CDSL SFTP/XML feeds (handled in Prompt 222).
  - Physical share custody transfers and demat account settlement with NSDL/CDSL (handled in Prompt 213).
  - Off-chain matching engine order cancellation and resting quote purge on ex-date (handled in Prompt 205 and Prompt 713).
  - Bank account fiat payouts for cash dividends and cash-in-lieu proceeds via payment gateways (handled in Prompt 212).
  - Section 194 TDS calculation and annual tax reporting (handled in Prompt 222 and Prompt 223).
  - Secondary market trade settlement and delivery-versus-payment execution (handled in Prompt 306 and Prompt 329).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD`, custom user-defined value types, and native custom errors for gas minimization).
- **Fixed-Point Mathematics:** High-precision integer arithmetic using a standard scaling factor of 10^6 (`RATIO_PRECISION = 1_000_000`), operating seamlessly on 18-decimal token balances (`10^18`). OpenZeppelin `Math.sol` full 512-bit multiplication and division (`mulDiv`) with explicit `Rounding.Floor` and `Rounding.Ceil` modes to eliminate precision loss and prevent integer overflow.
- **Base Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `UUPSUpgradeable`).
  - OpenZeppelin Contracts v5.0 `Math.sol` for safe 512-bit arithmetic.
  - ERC-3643 (T-REX) Permissioned Security Token standard interfaces (`IERC3643`, `IToken`, `IIdentityRegistry`).
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, invariant property fuzzing, and gas profiling; `cast` for Besu JSON-RPC diagnostic interactions).
- **Static Analysis & Formal Verification:** Slither, Mythril, and Solhint automated CI checks verifying zero privilege escalations, correct storage layout preservation, and absence of rounding vulnerabilities.

## Backend / Infra Touchpoints
- **Corporate Actions Service (Prompt 222):** Core microservice that detects corporate events from exchange and depository notifications, snapshots off-chain portfolio balances, stages action parameters, and coordinates on-chain proposal submissions via multi-sig relayer.
- **Custodian Depository Integration Service (Prompt 213):** Validates physical share receipt or sub-division at NSDL/CDSL; produces cryptographic signature proofs linking the on-chain action to the official Depository Event Reference ID.
- **Fee Engine (Prompt 210):** Evaluates corporate action handling fees, depository pass-through tariffs, or rights processing charges, ensuring all operational fees are accounted for prior to balance finalization.
- **Circuit Breaker & Emergency Halt Contract (Prompt 341):** Triggered by `CorporateActionSplitter` to enforce `haltISIN` on ex-date, guaranteeing no secondary transfers occur while balances are snapshotted and adjusted.
- **Blockchain Event Indexer (Prompt 309):** Ingests `CorporateActionDeclared`, `BatchSplitExecuted`, `DustAllocated`, and `CorporateActionFinalized` events, updating PostgreSQL holdings tables and notifying frontend portfolio views within 50ms.
- **Daily Reconciliation Service (Prompt 215):** Compares post-split on-chain `totalSupply()` plus aggregated dust reserves against NSDL/CDSL end-of-day holding statements to guarantee continuous 1:1 asset backing.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Deployed on Hyperledger Besu permissioned consortium network running QBFT consensus with 2-second deterministic block times and instant single-block transaction finality (zero reorganizations).
- **Direct Interaction with TokenizedEquity.sol:**
  - For pure forward splits (e.g., 2:1 or 10:1), calls `TokenizedEquity.rebaseMultiplier(uint256 newMultiplier, bytes32 corporateActionId)` which adjusts the internal balance scalar atomically across all token holders.
  - For bonus issues (e.g., 1:1 bonus), invokes `TokenizedEquity.mint(address recipient, uint256 amount)` across pre-verified investor addresses, gated by ERC-3643 identity verification.
  - For reverse splits / consolidations (e.g., 1:5 consolidation), executes `TokenizedEquity.burn(address holder, uint256 burnAmount)` while routing fractional residuals to `CUSTODY_DUST_ESCROW`.
- **Zero Share Dilution Invariant:** At all times, the contract guarantees that:
  $$\text{TotalSupply}_{post} + \text{Dust}_{escrow} = \text{TotalSupply}_{pre} \times \frac{\text{RatioNumerator}}{\text{RatioDenominator}}$$
  Any operation that violates this invariant reverts the transaction atomically.
- **Zero On-Chain PII:** The contract stores zero investor names, tax identifiers (PAN), or personal demat details. All state variables and events reference strictly `bytes32 indexed isin`, `bytes32 indexed corporateActionId`, `address indexed investorAddress`, and `uint256` token amounts.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & Directory Structure:** Create Foundry project layout under `contracts/`:
   - `contracts/src/tokens/CorporateActionSplitter.sol`
   - `contracts/src/interfaces/ICorporateActionSplitter.sol`
   - `test/tokens/CorporateActionSplitter.t.sol`
   - `test/mocks/MockTokenizedEquity.sol`
   - `script/DeployCorporateActionSplitter.s.sol`
2. **Define Data Structures, Enums, and Custom Errors:** Implement `ICorporateActionSplitter.sol` specifying `ActionType` (`STOCK_SPLIT`, `REVERSE_SPLIT`, `BONUS_ISSUE`, `RIGHTS_ISSUE`, `CASH_IN_LIEU`), `ActionStatus` (`SCHEDULED`, `FROZEN`, `EXECUTING`, `COMPLETED`, `CANCELLED`), `CorporateAction` struct, and comprehensive custom errors.
3. **Configure Role-Based Access Control:** Implement OpenZeppelin `AccessControlUpgradeable` with four distinct enterprise roles: `DEFAULT_ADMIN_ROLE`, `CORPORATE_ACTION_OPERATOR_ROLE`, `DEPOSITORY_RELAYER_ROLE`, and `EMERGENCY_GUARDIAN_ROLE`.
4. **Implement Fixed-Point Math Engine:** Build mathematical functions using OpenZeppelin `Math.mulDiv` configured with `RATIO_PRECISION = 1_000_000`:
   - `calculateEntitlement(uint256 balance, uint256 numerator, uint256 denominator)` applying `Rounding.Floor`.
   - `calculateDustResidual(uint256 balance, uint256 numerator, uint256 denominator)` capturing fractional rounding remnants.
5. **Implement Corporate Action Declaration (`declareCorporateAction`):** Build entrypoint enabling authorized operators to register an action. Store ISIN, action type, ratio numerator, ratio denominator, ex-date timestamp, record-date cutoff, and depository reference hash. Validate numerator > 0 and denominator > 0.
6. **Implement Circuit Breaker Integration Hooks:** Connect to `ICircuitBreakerHalt` (Prompt 341). On transition to `FROZEN`, verify or trigger `haltISIN` on the target equity token to prevent concurrent trading or settlement during balance recalculation.
7. **Implement Atomic Global Rebase Execution (`executeRebaseSupply`):** For forward stock splits with uniform integer multipliers, execute an atomic supply rebase by invoking `rebaseMultiplier` on the target `TokenizedEquity.sol` contract, transitioning action status to `COMPLETED`.
8. **Implement Batched Bonus Share Distribution (`executeBatchBonus`):** Build batch processing function accepting arrays of investor addresses and their snapshotted balances. Calculate bonus units, verify ERC-3643 compliance via `IIdentityRegistry`, invoke `TokenizedEquity.mint`, and emit `BatchBonusExecuted`.
9. **Implement Batched Reverse Split Consolidation (`executeBatchReverseSplit`):** Build batch processing function for consolidations. Calculate new consolidated balances, burn the difference from investor holdings, aggregate fractional remainder units into the corporate action dust accumulator, and emit `BatchReverseSplitExecuted`.
10. **Implement Micro-Token Dust Sweep & CIL Finalization (`finalizeDustAllocation`):** Sweep accumulated micro-token dust to `CUSTODY_DUST_ESCROW` address. Record total dust units, emit `DustAllocated` event with IPFS reference hash of off-chain Cash-in-Lieu payout ledger, and transition action to `COMPLETED`.
11. **Implement Depository Dual-Signer Verification Gate:** Require that before execution begins, `DEPOSITORY_RELAYER_ROLE` must submit `confirmDepositoryRecord(bytes32 actionId, bytes32 nsdlCdslMasterHash, bytes calldata signature)` proving physical demat parity.
12. **Implement Emergency Cancellation & Rollback:** Allow `EMERGENCY_GUARDIAN_ROLE` or governance multi-sig to cancel actions in `SCHEDULED` or `FROZEN` state prior to execution, automatically lifting ISIN trading halts.
13. **Build Comprehensive Foundry Mock Suite:** Create `MockTokenizedEquity.sol` implementing ERC-3643 mint/burn/rebase hooks and `MockCircuitBreaker.sol` tracking halt calls.
14. **Implement Property-Based Invariant Fuzz Tests:** Write Foundry invariant tests asserting:
    - Dilution invariant holds across all random holder balances and split ratios: post-supply + dust strictly equals expected scaled supply.
    - Dust residual per holder is strictly bounded: $0 \le \text{dust} < 10^{-6}$ fractional share units.
    - No unauthorized address can declare, execute, or cancel actions.
    - Zero execution possible after `COMPLETED` or `CANCELLED` status.
15. **Conduct Slither Static Analysis & Gas Profiling:** Run Slither analysis, ensuring zero high or medium severity findings. Profile gas usage to guarantee batch functions process up to 200 accounts per Besu transaction block within gas limits.

## Interfaces / Contracts

### Corporate Action Splitter Interface (`ICorporateActionSplitter.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface ICorporateActionSplitter {
    /// @notice Type of corporate action executed on-chain
    enum ActionType {
        STOCK_SPLIT,       // Forward sub-division (e.g., 2:1, 10:1)
        REVERSE_SPLIT,     // Consolidation (e.g., 1:5, 1:10)
        BONUS_ISSUE,       // Stock dividend (e.g., 1:1 bonus)
        RIGHTS_ISSUE,      // Pro-rata subscription entitlement
        CASH_IN_LIEU       // Fractional cash liquidation distribution
    }

    /// @notice Lifecycle state of a corporate action event
    enum ActionStatus {
        SCHEDULED,         // Declared, awaiting ex-date
        FROZEN,            // Trading halted on target ISIN, snapshot active
        EXECUTING,         // Batched balance adjustments in progress
        COMPLETED,         // All balances adjusted, dust swept, trading resumed
        CANCELLED          // Aborted by governance or depository desync
    }

    /// @notice Core corporate action configuration and state container
    struct CorporateAction {
        bytes32 corporateActionId;    // Unique hash: keccak256(isin, actionType, recordDate)
        bytes32 isin;                 // Security identifier (e.g., INE467B01029)
        address tokenAddress;         // Target ERC-3643 TokenizedEquity contract
        ActionType actionType;        // Type of corporate action
        ActionStatus status;          // Current lifecycle status
        uint32 ratioNumerator;        // Split ratio numerator (e.g., 2 in 2:1 split)
        uint32 ratioDenominator;      // Split ratio denominator (e.g., 1 in 2:1 split)
        uint64 declaredAt;            // Timestamp of declaration
        uint64 exDate;                // Official Ex-Date timestamp (IST 00:00:00 equivalent)
        uint64 recordDate;            // Official Record Date cutoff timestamp
        bytes32 depositoryEventRef;   // NSDL/CDSL official corporate action notice hash
        uint256 totalPreSupply;       // Total supply recorded at snapshot
        uint256 totalPostSupply;      // Total supply calculated post-adjustment
        uint256 totalDustAccumulated; // Accumulated fractional micro-token dust
        uint32 processedAccounts;     // Count of accounts adjusted in batches
    }

    /// @notice Batch execution progress tracking container
    struct BatchProgress {
        bytes32 corporateActionId;
        uint32 batchIndex;
        uint32 accountsInBatch;
        uint256 tokensMintedOrBurned;
        uint256 dustGenerated;
        bytes32 batchRootHash;        // Merkle root of snapshotted balances in batch
    }

    // --- Events ---
    event CorporateActionDeclared(
        bytes32 indexed corporateActionId,
        bytes32 indexed isin,
        address indexed tokenAddress,
        ActionType actionType,
        uint32 ratioNumerator,
        uint32 ratioDenominator,
        uint64 exDate,
        uint64 recordDate,
        bytes32 depositoryEventRef
    );

    event CorporateActionStatusUpdated(
        bytes32 indexed corporateActionId,
        ActionStatus previousStatus,
        ActionStatus newStatus,
        address updatedBy
    );

    event DepositoryParityConfirmed(
        bytes32 indexed corporateActionId,
        bytes32 indexed nsdlCdslMasterHash,
        address confirmedBy
    );

    event RebaseSupplyExecuted(
        bytes32 indexed corporateActionId,
        address indexed tokenAddress,
        uint256 oldSupply,
        uint256 newSupply,
        uint256 multiplier
    );

    event BatchBonusExecuted(
        bytes32 indexed corporateActionId,
        uint32 indexed batchIndex,
        uint256 totalMinted,
        uint256 accountsProcessed
    );

    event BatchReverseSplitExecuted(
        bytes32 indexed corporateActionId,
        uint32 indexed batchIndex,
        uint256 totalBurned,
        uint256 dustCollected,
        uint256 accountsProcessed
    );

    event DustAllocated(
        bytes32 indexed corporateActionId,
        address indexed dustEscrowAddress,
        uint256 totalDustTokens,
        bytes32 cilLedgerHash
    );

    event CorporateActionFinalized(
        bytes32 indexed corporateActionId,
        uint256 finalTotalSupply,
        uint256 totalDustAccumulated,
        uint32 totalAccountsAdjusted
    );

    event CorporateActionCancelled(
        bytes32 indexed corporateActionId,
        bytes32 reasonHash,
        address cancelledBy
    );

    // --- Custom Errors ---
    error InvalidActionState(bytes32 actionId, ActionStatus currentStatus, ActionStatus requiredStatus);
    error InvalidRatio(uint32 numerator, uint32 denominator);
    error ExDateNotReached(uint64 exDate, uint64 currentTimestamp);
    error RecordDateCutoffPassed(uint64 recordDate, uint64 currentTimestamp);
    error DepositoryProofMismatch(bytes32 expectedRef, bytes32 providedRef);
    error ArrayLengthMismatch();
    error DilutionInvariantBreached(uint256 expectedSupply, uint256 actualSupply);
    error UnauthorizedOperator(address caller, bytes32 requiredRole);
    error ZeroAddressDetected();
    error TradingNotHalted(bytes32 isin);
    error ZeroTokenBalance(address investor);
    error DustEscrowFailed();

    // --- Core View Methods ---
    function getCorporateAction(bytes32 corporateActionId) external view returns (CorporateAction memory);
    function calculateEntitlement(
        uint256 balance,
        uint32 numerator,
        uint32 denominator
    ) external pure returns (uint256 newBalance, uint256 dustResidual);
    function getAccumulatedDust(bytes32 corporateActionId) external view returns (uint256);
    function isActionFinalized(bytes32 corporateActionId) external view returns (bool);

    // --- Lifecycle State Execution Methods ---
    function declareCorporateAction(
        bytes32 isin,
        address tokenAddress,
        ActionType actionType,
        uint32 ratioNumerator,
        uint32 ratioDenominator,
        uint64 exDate,
        uint64 recordDate,
        bytes32 depositoryEventRef
    ) external returns (bytes32 corporateActionId);

    function confirmDepositoryRecord(
        bytes32 corporateActionId,
        bytes32 nsdlCdslMasterHash,
        bytes calldata custodianSignature
    ) external;

    function freezeTradingAndSnapshot(bytes32 corporateActionId) external;

    function executeRebaseSupply(bytes32 corporateActionId) external;

    function executeBatchBonus(
        bytes32 corporateActionId,
        uint32 batchIndex,
        address[] calldata recipients,
        uint256[] calldata snapshottedBalances
    ) external;

    function executeBatchReverseSplit(
        bytes32 corporateActionId,
        uint32 batchIndex,
        address[] calldata holders,
        uint256[] calldata snapshottedBalances
    ) external;

    function finalizeDustAllocation(
        bytes32 corporateActionId,
        address dustEscrowAddress,
        bytes32 cilLedgerHash
    ) external;

    function finalizeCorporateAction(bytes32 corporateActionId) external;

    function cancelCorporateAction(bytes32 corporateActionId, bytes32 reasonHash) external;
}
```

### TokenizedEquity ERC-3643 Corporate Action Hook Interface Snippet
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface ITokenizedEquityCorporateActionHook {
    /// @notice Rebase global supply multiplier for forward stock splits
    function rebaseMultiplier(
        uint256 newMultiplierPaise,
        bytes32 corporateActionId
    ) external;

    /// @notice Execute privileged corporate action bonus minting
    function corporateActionMint(
        address to,
        uint256 amount,
        bytes32 corporateActionId
    ) external;

    /// @notice Execute privileged corporate action consolidation burn
    function corporateActionBurn(
        address from,
        uint256 amount,
        bytes32 corporateActionId
    ) external;
}
```

## Security & Compliance Notes
- **Ex-Date Reconciliation with NSDL/CDSL Corporate Action Master Files:**
  - Before any corporate action can transition to `EXECUTING` status, `DEPOSITORY_RELAYER_ROLE` must submit the cryptographic hash of the official NSDL/CDSL corporate action master file (CAD / CA Master Feed).
  - The contract verifies that the on-chain parameters (ISIN, ratio numerator, ratio denominator, and record date) match the depository announcement hash signed by the institutional custodian's Hardware Security Module (HSM).
- **Zero Share Dilution Invariant:**
  - To prevent synthetic inflation or unauthorized share creation, every batch adjustment strictly enforces the conservation law:
    $$\text{ActualSupply}_{post} + \text{AccumulatedDust} = \text{InitialSupply}_{pre} \times \frac{\text{RatioNumerator}}{\text{RatioDenominator}}$$
  - If any rounding discrepancy occurs, the transaction reverts with `DilutionInvariantBreached`, ensuring that the on-chain capitalization remains mathematically identical to the physical shares in custodial demat accounts.
- **Fractional Cash-in-Lieu (CIL) Statutory Alignment:**
  - Under SEBI Listing Obligations and Disclosure Requirements (LODR) Regulations and the Companies Act 2013, non-integer fractional entitlements resulting from corporate actions (such as 1:5 reverse splits or odd bonus distributions) cannot be distributed as unbacked fractions.
  - The contract isolates all residual micro-tokens down to the 18th decimal place, transfers them to the `CUSTODY_DUST_ESCROW` pool, and records an IPFS hash (`cilLedgerHash`) referencing the off-chain fiat disbursement ledger executed via Prompt 210 and Prompt 222.
- **Anti-Front-Running Trading Halt Enforcement:**
  - Execution of corporate actions requires synchronous integration with `CircuitBreakerHalt.sol` (Prompt 341).
  - Calling `freezeTradingAndSnapshot` asserts that `circuitBreaker.isHalted(isin)` returns `true`. If the market has not been halted on the target ISIN, execution reverts with `TradingNotHalted(isin)`, eliminating front-running or balance mutation during batch calculation.
- **Reentrancy and Storage Safety:**
  - All state-changing routines employ OpenZeppelin `ReentrancyGuardUpgradeable` to block cross-function reentrancy.
  - Storage variables are structured according to ERC-7201 Namespaced Storage Layout standards to guarantee seamless upgradeability without storage collision risks.

## Acceptance Criteria
- [ ] `CorporateActionSplitter.sol` and `ICorporateActionSplitter.sol` implemented, compiled with Solidity ^0.8.24, and passing Slither static analysis with zero high or medium severity findings.
- [ ] Complete corporate action lifecycle state machine (`SCHEDULED`, `FROZEN`, `EXECUTING`, `COMPLETED`, `CANCELLED`) enforced with deterministic state transition guards.
- [ ] Fixed-point mathematical calculation module verified for 10^6 precision (`RATIO_PRECISION = 1_000_000`) across forward splits (2:1, 10:1), reverse splits (1:5, 1:10), and bonus distributions (1:1, 1:2).
- [ ] Atomic global rebase routine (`executeRebaseSupply`) successfully updates `TokenizedEquity.sol` multiplier in O(1) complexity for pure forward splits.
- [ ] Batched bonus minting (`executeBatchBonus`) and reverse split consolidation (`executeBatchReverseSplit`) correctly process arrays of up to 200 investor accounts per block.
- [ ] Fractional micro-token dust correctly aggregated and transferred to `CUSTODY_DUST_ESCROW` with emitted `DustAllocated` event containing the CIL payout ledger hash.
- [ ] Dual-custody authorization gate strictly verifies NSDL/CDSL corporate action master hash signed by `DEPOSITORY_RELAYER_ROLE` before execution commences.
- [ ] Zero Share Dilution Invariant verified via Foundry invariant fuzz test running $\ge 10,000$ iterations with zero balance variance.
- [ ] Integration with `CircuitBreakerHalt.sol` verified: corporate action execution reverts with `TradingNotHalted` if target ISIN is not frozen.
- [ ] Foundry test suite achieves $\ge 95\%$ line and branch test coverage across all core execution paths, custom error reverts, and boundary rounding conditions.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301` (Permissioned Blockchain Evaluation & Selection)
  - Prompt `302` (Network Topology and Validator Setup)
  - Prompt `303` (Permissioned Asset Token Issuance Smart Contract - ERC-3643)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
  - Prompt `341` (On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract)
- **Parallel Tasks:**
  - Prompt `222` (Corporate Actions Service - Dividends, Splits, Bonuses & Token Adjustments)
  - Prompt `213` (Custodian Depository Integration Service - NSDL & CDSL Integration)
  - Prompt `210` (Transaction Fee & Regulatory Tariff Calculation Engine)
  - Prompt `215` (Reconciliation & Anomaly Detection Service)
  - Prompt `309` (Blockchain Event Indexer & Audit Trail)
- **Subsequent Prompts Enabled:**
  - Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract)
  - Prompt `329` (NBSE Delivery-versus-Payment DvP Settlement & Automated Fee Collector)
  - Prompt `331` (Primary Market Order Routing & Clearing Bridge)
  - Prompt `811` (Testnet vs Mainnet Dual Environment CI/CD)
