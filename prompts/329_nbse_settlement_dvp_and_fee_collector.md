# 329 - NBSE Delivery-versus-Payment (DvP) Settlement & Automated Fee Collector Smart Contracts (NBSESettlementDvP.sol, NBSEFeeCollector.sol)

## Purpose
In institutional capital markets and modern securities clearinghouses, Delivery-versus-Payment (DvP) Model 1 represents the gold standard of settlement integrity. Under Model 1 DvP, securities and sovereign fiat cash funds are settled simultaneously on a gross, trade-by-trade atomic basis: the final legal transfer of fractional digital equity from seller to buyer occurs if and only if the corresponding transfer of cash liquidity from buyer to seller is definitively and unconditionally finalized in the exact same state transition. Without atomic DvP, market participants remain exposed to catastrophic counterparty replacement cost and principal default risk during multi-day settlement cycles.

This prompt specifies the design, architecture, interface definitions, gas-optimized data layouts, and security invariants for the **NBSE Delivery-versus-Payment Settlement & Automated Fee Collector Smart Contract Suite** (`NBSESettlementDvP.sol` and `NBSEFeeCollector.sol`) deployed on the permissioned Hyperledger Besu consortium ledger under QBFT consensus. 

The suite enforces three critical, non-negotiable financial mechanisms:
1. **Atomic Simultaneous DvP Settlement:** Gross bilateral equity token delivery against tokenized sovereign fiat (eINR / CBDC) in a single atomic Besu transaction.
2. **Strict Fixed 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) Transaction Fee:** Programmatic fee deduction on every trade execution strictly calculated as `feeAmount = (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)`, guaranteeing zero fee leakage and transparent post-trade accounting.
3. **Automated Fee Distribution & Non-Bypassable SGF Default Waterfall:** Real-time routing of collected transaction fees to designated statutory treasuries (Exchange Treasury, Clearing Corporation Reserve, Investor Protection Fund, Infrastructure Maintenance Pool) and immediate programmatic escalation to the Settlement Guarantee Fund (`SettlementGuaranteeFund.sol`, Prompt 315) if counterparty liquidity or margin obligations fail.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/settlement/` comprising:
- `NBSESettlementDvP.sol`: Primary settlement execution contract executing atomic single and batched bilateral DvP trades, validating EIP-712 cryptographic signatures from authorized settlement relayers, enforcing ERC-3643 compliance hooks, deducting the 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) transaction fee via `NBSEFeeCollector.sol`, and invoking the non-bypassable SGF Default Waterfall in default scenarios.
- `NBSEFeeCollector.sol`: Dedicated statutory fee accounting and automated distribution vault. It ingests 0 bps (0.00% fee at launch) trade fees directly from settlement transactions, maintains per-currency fee accounting ledgers, and splits collected proceeds across statutory stakeholder pools according to governance-approved basis point allocations.
- `INBSESettlementDvP.sol`: Complete interface defining all DvP data structures, enums, custom errors, events, view methods, single-trade settlement, batch settlement, and SGF default escalation routines.
- `INBSEFeeCollector.sol`: Complete interface defining fee collection hooks, treasury distribution channels, basis point allocation structs, events, custom errors, and accounting views.
- Comprehensive Foundry unit, fuzz, and invariant test specifications (`test/settlement/NBSESettlementDvP.t.sol` and `test/settlement/NBSEFeeCollector.t.sol`) validating DvP atomicity, exact 0.00% fee (no fee at all) calculation, multi-currency compatibility, fee distribution splits, and default waterfall invocation.

## Scope Boundaries
- **In Scope:**
  - Gross atomic Delivery-versus-Payment (DvP Model 1) execution for tokenized equities (ERC-3643 / ERC-20) and settlement currency (eINR / CBDC ERC-20).
  - Multi-trade batching (processing up to 100 bilateral trades atomically per Besu transaction block).
  - Strict mathematical calculation and enforcement of the 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) fixed transaction fee (`feeAmount = (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)`).
  - Automated fee distribution across 4 statutory channels: NBSE Exchange Treasury, Clearing Corporation Operating Reserve, Investor Protection Fund (IPF), and Infrastructure Maintenance Pool.
  - Replay protection, trade idempotency, and nonce tracking per market participant.
  - Non-bypassable escalation and liquidity drawdown interface with the on-chain Settlement Guarantee Fund (`ISettlementGuaranteeFund.sol`, Prompt 315) upon counterparty payment failure.
  - Integration with OpenZeppelin ERC-1967 UUPS upgradeability and role-based access control (`AccessControlUpgradeable`).
- **Out of Scope / Handled Elsewhere:**
  - Off-chain limit order book matching and trade execution algorithms (handled in Prompt 205).
  - Off-chain pre-trade margin and VaR calculations (handled in Prompt 206 and Prompt 229).
  - Off-chain Trade Settlement & DvP Orchestration Service (handled in Prompt 208).
  - Core SGF multi-tranche slashing mechanics and pro-rata mutualized assessments (handled in Prompt 315 and Prompt 230).
  - Commercial bank RTGS / UPI fiat hold reservations (handled in Prompt 212 and Prompt 203).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, custom errors, and transient storage `TSTORE`/`TLOAD` support).
  *Justification:* Guarantees mathematical safety against integer overflow/underflow, minimizes gas costs during large batch iterations via custom errors, and provides deterministic execution under Besu QBFT.
- **Contract Libraries & Standards:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`, `ECDSA`, `EIP712Upgradeable`).
- **Token Standards:** ERC-3643 (Permissioned Security Tokens) for underlying equity shares and ERC-20 (`SafeERC20`) for eINR / CBDC settlement currency.
- **Development & Testing Framework:** **Foundry** (`forge` for high-throughput compilation and fuzzing, `cast` for RPC calls, `anvil` for local EVM testing).
- **Static Analysis & Formal Verification:** Slither (Trail of Bits), Mythril, and Solhint automated security linting pipelines.

## Backend / Infra Touchpoints
- **Trade Settlement Service (Prompt 208):** Off-chain Go microservice ingesting executed matches from the matching engine (Prompt 205), packaging trades into atomic batches, signing EIP-712 settlement payloads using CloudHSM, and submitting transactions to `NBSESettlementDvP.sol`.
- **Wallet & Account Service (Prompt 203):** Manages double-entry fiat ledgers and coordinates eINR / CBDC token balances for buyer and seller wallets.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes on-chain 0.00% fee (No fee at all) deductions with off-chain accounting records and tax reporting systems.
- **Settlement Guarantee Fund Service (Prompt 230 / Prompt 315):** Ingests default escalation events from `NBSESettlementDvP.sol` to trigger multi-tranche collateral drawdowns.
- **Blockchain Event Indexer (Prompt 309):** Ingests `DvPTradeSettled`, `DvPBatchSettled`, `FeeCollected`, and `SettlementDefaultTriggered` events to update relational read replicas.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Block Finality:** Hyperledger Besu private permissioned network running QBFT consensus with 2-second block intervals and deterministic immediate finality (zero probabilistic forks).
- **Atomic Two-Legged Asset Swapping:** `NBSESettlementDvP.sol` pulls settlement currency from the buyer and equity tokens from the seller via `SafeERC20.safeTransferFrom`, executing both legs simultaneously.
- **Strict Fee Interception:** During DvP execution, the contract deducts `feeAmount = (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)` from the gross cash leg and routes it to `NBSEFeeCollector.sol`. The net cash (`tradeVolumeINR - feeAmount`) is transferred to the seller.
- **Zero On-Chain PII:** The contracts store only cryptographic identifiers (`bytes32 tradeId`, `bytes32 batchId`), pseudonymous wallet addresses (`address buyer`, `address seller`), token addresses, raw token quantities, and rupee amounts in paise/wei precision. Zero investor names, PANs, Aadhaar numbers, or banking credentials exist on-chain.
- **Multi-Sig Governance:** Upgradeability, fee split parameters, SGF contract bindings, and emergency pause controls require 3-of-5 multisig authorization (`MultiSigGovernance.sol`, Prompt 307) backed by CloudHSM validator keys.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:** Initialize smart contract directory under `contracts/settlement/` with `src/NBSESettlementDvP.sol`, `src/NBSEFeeCollector.sol`, `src/interfaces/INBSESettlementDvP.sol`, `src/interfaces/INBSEFeeCollector.sol`, and test directories under `test/settlement/`.
2. **Define Interface Specifications:** Create `INBSESettlementDvP.sol` and `INBSEFeeCollector.sol` defining all structs (`TradeOrder`, `DvPSettlementBatch`, `FeeDistributionSplit`, `FeeAccountingRecord`), enums, custom errors, events, and method signatures.
3. **Configure EIP-712 Domain Separator:** Establish typed data hashing for `Growww-NBSE-DvP` with domain version `1`, chain ID, and verifying contract address to prevent cross-network replay attacks.
4. **Implement Fee Calculation Logic:** Define the standard 0 bps (0.00% fee at launch) mathematical formula `feeAmount = (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)` in both contracts with strict equality verification against submitted trade orders.
5. **Implement NBSEFeeCollector Vault:**
   - Inherit from `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, and `PausableUpgradeable`.
   - Implement `collectTradeFee` callable only by authorized `NBSESettlementDvP` contracts.
   - Maintain per-token accounting mappings tracking `totalFeesCollected`, `totalFeesDistributed`, and `totalTradesProcessed`.
   - Implement `distributeAccumulatedFees` to compute and disburse statutory splits (`exchangeBps`, `clearingCorpBps`, `investorProtectionBps`, `infrastructureBps`) verifying that total BPS equals 10,000 (100.00%).
6. **Implement NBSESettlementDvP Core Architecture:**
   - Inherit from OpenZeppelin upgradeable contracts, `ReentrancyGuardUpgradeable`, and `EIP712Upgradeable`.
   - Store immutable references to `NBSEFeeCollector` and `ISettlementGuaranteeFund`.
   - Establish role mappings: `DEFAULT_ADMIN_ROLE`, `SETTLEMENT_OPERATOR_ROLE`, `EMERGENCY_GUARDIAN_ROLE`, and `RISK_COMMITTEE_ROLE`.
7. **Implement Single Trade DvP Execution (`settleSingleDvP`):**
   - Verify caller authorization or validate EIP-712 relayer signature.
   - Verify trade expiry (`trade.expiryTimestamp >= block.timestamp`) and idempotency (`!isTradeSettled[trade.tradeId]`).
   - Validate strict 0.00% fee (no fee at all): require `trade.expectedFeeAmount == 0` (0 bps at launch; FeeController governed).
   - Mark `isTradeSettled[trade.tradeId] = true` prior to token transfers (Checks-Effects-Interactions pattern).
   - Transfer gross settlement currency from buyer: route `feeAmount` to `NBSEFeeCollector` and net cash (`tradeVolumeINR - feeAmount`) to seller.
   - Transfer fractional equity units (`trade.tokenAmount`) from seller to buyer via `SafeERC20.safeTransferFrom`.
   - Emit `DvPTradeSettled`.
8. **Implement Batch DvP Settlement (`settleBatchDvP`):**
   - Accept `DvPSettlementBatch` containing up to 100 trades.
   - Enforce atomic all-or-nothing execution: if any single trade in the batch fails compliance, balance checks, or signature validation, revert the entire transaction cleanly.
   - Aggregate batch volume and total fees, emitting `DvPBatchSettled`.
9. **Implement SGF Default Waterfall Trigger (`executeDefaultWaterfallSettlement`):**
   - If seller fails to deliver securities or buyer cash transfer fails, permit authorized `RISK_COMMITTEE_ROLE` to trigger default resolution.
   - Invoke `ISettlementGuaranteeFund.declareMemberDefault` and execute sequential tranche drawdowns to cover liquidity deficits.
   - Emit `SettlementDefaultTriggered`.
10. **Implement Governance & Parameter Management:**
    - Add `updateFeeSplitConfig` and `updateTreasuryAddress` in `NBSEFeeCollector.sol` guarded by multi-sig timelock.
    - Add `setFeeCollector` and `setSGFContract` in `NBSESettlementDvP.sol` restricted to `DEFAULT_ADMIN_ROLE`.
11. **Implement Emergency Controls & Circuit Breakers:**
    - Implement `pauseSettlement` and `unpauseSettlement` in `NBSESettlementDvP.sol` callable by `EMERGENCY_GUARDIAN_ROLE`.
    - Implement `cancelTrade(bytes32 tradeId, string calldata reason)` restricted to authorized clearing operators.
12. **Write Foundry Unit Test Suite:**
    - Test valid single DvP trade settlement with exact 0.00% fee (no fee at all) deduction.
    - Test atomic batch settlement with 10, 50, and 100 trades.
    - Test rejection on fee mismatch (e.g. 0 bps, 2 bps, or incorrect rounding).
    - Test rejection on expired timestamps, duplicate `tradeId` replay, and invalid EIP-712 signatures.
13. **Write Invariant & Fuzz Tests:**
    - Invariant: `totalTokensDeliveredToBuyers == totalTokensDeductedFromSellers`.
    - Invariant: `totalCashPaidByBuyers == totalCashReceivedBySellers + totalFeesCollected`.
    - Invariant: `feeAmount == (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)` across all fuzzed volume amounts.
    - Invariant: sum of fee distribution splits across all channels strictly equals 10,000 bps (100%).
14. **Run Static Analysis & Formal Verification:**
    - Execute Slither and Mythril to verify zero reentrancy, unhandled return values, or access control flaws.
15. **Prepare Deployment Scripts & Besu Migration:**
    - Write `script/DeployNBSESettlement.s.sol` for deterministic proxy deployment on Besu QBFT testnet and mainnet.

## Interfaces / Contracts

### NBSE Settlement DvP Interface (`INBSESettlementDvP.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface INBSESettlementDvP {
    enum SettlementStatus {
        PENDING,
        SETTLED,
        FAILED_INSUFFICIENT_BALANCE,
        DEFAULT_WATERFALL_TRIGGERED,
        CANCELLED
    }

    struct TradeOrder {
        bytes32 tradeId;
        bytes32 orderIdBuy;
        bytes32 orderIdSell;
        address buyer;
        address seller;
        address tokenAddress;
        address currencyToken;
        uint256 tokenAmount;
        uint256 tradeVolumeINR;
        uint256 expectedFeeAmount;
        uint64 matchedTimestamp;
        uint64 expiryTimestamp;
        uint256 nonce;
    }

    struct DvPSettlementBatch {
        bytes32 batchId;
        TradeOrder[] trades;
        bytes[] signatures;
    }

    struct DefaultInterventionRecord {
        bytes32 tradeId;
        bytes32 defaultWaterfallId;
        address defaultingParty;
        uint256 deficitAmount;
        uint64 timestamp;
        bool isResolved;
    }

    // Events
    event DvPTradeSettled(
        bytes32 indexed tradeId,
        bytes32 indexed batchId,
        address indexed tokenAddress,
        address buyer,
        address seller,
        address currencyToken,
        uint256 tokenAmount,
        uint256 tradeVolumeINR,
        uint256 feeAmount
    );

    event DvPBatchSettled(
        bytes32 indexed batchId,
        uint256 tradeCount,
        uint256 totalVolumeINR,
        uint256 totalFeesCollected,
        uint64 timestamp
    );

    event SettlementDefaultTriggered(
        bytes32 indexed tradeId,
        address indexed defaultingParty,
        uint256 deficitAmount,
        bytes32 defaultWaterfallId,
        uint64 timestamp
    );

    event FeeCollectorUpdated(address indexed oldCollector, address indexed newCollector);
    event SGFContractUpdated(address indexed oldSGF, address indexed newSGF);
    event TradeCancelled(bytes32 indexed tradeId, string reason, uint64 timestamp);
    event MaxBatchSizeUpdated(uint256 oldMaxSize, uint256 newMaxSize);

    // Errors
    error TradeAlreadySettled(bytes32 tradeId);
    error TradeExpired(bytes32 tradeId, uint64 expiryTimestamp, uint64 currentTimestamp);
    error InvalidTradeFee(bytes32 tradeId, uint256 providedFee, uint256 requiredFee);
    error ZeroAddressDetected();
    error ZeroAmountProvided();
    error UnauthorizedSettlementRelayer(address caller);
    error InvalidSignature(bytes32 tradeId);
    error BatchSizeExceeded(uint256 size, uint256 maxSize);
    error BatchArrayLengthMismatch(uint256 tradesCount, uint256 signaturesCount);
    error TokenTransferFailed(address token, address from, address to, uint256 amount);
    error CashTransferFailed(address currency, address from, address to, uint256 amount);
    error SGFWaterfallExecutionFailed(bytes32 tradeId, bytes32 defaultWaterfallId);
    error ComplianceCheckFailed(address participant, address token);
    error InvalidNonce(address account, uint256 providedNonce, uint256 expectedNonce);

    // Read Functions
    function isTradeSettled(bytes32 tradeId) external view returns (bool);
    function getTradeStatus(bytes32 tradeId) external view returns (SettlementStatus);
    function getDefaultRecord(bytes32 tradeId) external view returns (DefaultInterventionRecord memory);
    function calculateTransactionFee(uint256 tradeVolumeINR) external pure returns (uint256 feeAmount);
    function feeCollector() external view returns (address);
    function sgfContract() external view returns (address);
    function maxBatchSize() external view returns (uint256);
    function userNonces(address account) external view returns (uint256);

    // Mutating Functions
    function settleSingleDvP(
        TradeOrder calldata trade,
        bytes calldata signature
    ) external returns (bool success);

    function settleBatchDvP(
        DvPSettlementBatch calldata batch
    ) external returns (uint256 settledCount, uint256 totalVolumeINR);

    function executeDefaultWaterfallSettlement(
        bytes32 tradeId,
        TradeOrder calldata trade,
        bytes32 defaultWaterfallId
    ) external returns (bool success);

    function cancelTrade(bytes32 tradeId, string calldata reason) external;
    function setFeeCollector(address newFeeCollector) external;
    function setSGFContract(address newSGFContract) external;
    function setMaxBatchSize(uint256 newMaxSize) external;
    function pauseSettlement() external;
    function unpauseSettlement() external;
}
```

### NBSE Fee Collector Interface (`INBSEFeeCollector.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface INBSEFeeCollector {
    enum FeeDistributionChannel {
        EXCHANGE_TREASURY,
        CLEARING_CORP_RESERVE,
        INVESTOR_PROTECTION_FUND,
        INFRASTRUCTURE_MAINTENANCE_POOL
    }

    struct FeeDistributionSplit {
        address exchangeTreasury;
        address clearingCorpReserve;
        address investorProtectionFund;
        address infrastructurePool;
        uint16 exchangeBps;
        uint16 clearingCorpBps;
        uint16 investorProtectionBps;
        uint16 infrastructureBps;
    }

    struct FeeAccountingRecord {
        uint256 totalFeesCollected;
        uint256 totalFeesDistributed;
        uint256 totalTradesProcessed;
        uint256 undistributedBalance;
        uint64 lastDistributionTimestamp;
    }

    // Events
    event FeeCollected(
        bytes32 indexed tradeId,
        address indexed currencyToken,
        address indexed payer,
        uint256 feeAmount,
        uint256 volumeINR,
        uint64 timestamp
    );

    event FeesDistributed(
        address indexed currencyToken,
        uint256 totalDistributed,
        uint256 exchangeAmount,
        uint256 clearingCorpAmount,
        uint256 ipfAmount,
        uint256 infraAmount,
        uint64 timestamp
    );

    event FeeSplitConfigUpdated(
        address indexed updatedBy,
        uint16 exchangeBps,
        uint16 clearingCorpBps,
        uint16 investorProtectionBps,
        uint16 infrastructureBps
    );

    event TreasuryAddressUpdated(
        FeeDistributionChannel indexed channel,
        address indexed oldAddress,
        address indexed newAddress
    );

    event SettlementContractWhitelisted(address indexed settlementContract, bool isWhitelisted);

    // Errors
    error InvalidFeeSplitSum(uint16 totalBps);
    error UnauthorizedCaller(address caller);
    error ZeroAddressForbidden();
    error InsufficientCollectedFees(address currencyToken, uint256 availableBalance);
    error DistributionTransferFailed(address recipient, uint256 amount);
    error InvalidTransactionFeeRate(uint256 expectedFee, uint256 actualFee);
    error SettlementContractNotWhitelisted(address caller);

    // Read Functions
    function calculateExpectedFee(uint256 tradeVolumeINR) external pure returns (uint256);
    function getFeeSplitConfig() external view returns (FeeDistributionSplit memory);
    function getFeeAccounting(address currencyToken) external view returns (FeeAccountingRecord memory);
    function isWhitelistedSettlementContract(address settlementContract) external view returns (bool);

    // Mutating Functions
    function collectTradeFee(
        bytes32 tradeId,
        address currencyToken,
        address payer,
        uint256 tradeVolumeINR,
        uint256 feeAmount
    ) external returns (bool success);

    function distributeAccumulatedFees(address currencyToken) external returns (uint256 totalDistributed);
    function updateFeeSplitConfig(FeeDistributionSplit calldata newSplit) external;
    function updateTreasuryAddress(FeeDistributionChannel channel, address newAddress) external;
    function setSettlementContractWhitelist(address settlementContract, bool isWhitelisted) external;
    function pauseFeeCollection() external;
    function unpauseFeeCollection() external;
}
```

## Security & Compliance Notes
- **Strict 0.00% fee (no fee at all) Invariant:** The contract enforces `feeAmount == (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)` with strict mathematical equality. Any trade payload containing an inconsistent or manipulated fee amount will revert immediately with `InvalidTradeFee`, preventing malicious under-billing or fee evasion.
- **Model 1 DvP Atomicity:** Both the asset delivery leg (`tokenAmount` transfer from seller to buyer) and payment leg (`tradeVolumeINR` transfer from buyer to seller and fee collector) execute within the same EVM transaction context. Failure in any single leg reverts the entire state transition, eliminating principal counterparty risk.
- **Replay & Idempotency Protection:** Every trade utilizes a unique cryptographic `tradeId = keccak256(orderIdBuy, orderIdSell, matchedTimestamp, nonce)`. Once settled, `isTradeSettled[tradeId]` is set to `true`, preventing double-settlement or replay attacks across blocks.
- **EIP-712 Domain Separation:** Settlement authorization signatures are validated against a structured EIP-712 typed data domain containing the Besu chain ID and verifying contract address, preventing signature replays across testnets or external networks.
- **Non-Bypassable SGF Default Waterfall:** If a trade cannot settle due to counterparty insolvency or missing collateral, the contract provides direct hooks into `ISettlementGuaranteeFund.declareMemberDefault`. The statutory waterfall executes sequentially through defaulter margins, member SGF contributions, Clearing Corporation dedicated capital, and mutualized pool funds without centralized bypass capability.
- **Multi-Signature Access Control:** Parameter modifications (fee split distributions, treasury destination addresses, SGF contract bindings, contract upgrades) are strictly restricted to 3-of-5 institutional multisig governance (`MultiSigGovernance.sol`, Prompt 307) backed by CloudHSM validator keys.
- **Zero On-Chain PII Compliance:** All data stored on-chain consists of pseudonymous Ethereum addresses and cryptographic hashes, ensuring full compliance with India's Digital Personal Data Protection (DPDP) Act 2023, SEBI cybersecurity guidelines, and international privacy standards.

## Acceptance Criteria
- [ ] `INBSESettlementDvP.sol` and `INBSEFeeCollector.sol` interface specifications compile cleanly under Solidity 0.8.24 with zero compiler warnings.
- [ ] Zero application implementation code included in interface specifications (only function signatures, structs, enums, events, and custom errors).
- [ ] Fixed 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) transaction fee calculation strictly adheres to `feeAmount = (tradeVolumeINR * feeBps) / 10000 (where feeBps == 0 at launch)` with zero precision loss.
- [ ] Model 1 DvP settlement guarantees atomic simultaneous transfer of equity tokens to buyer and net cash funds to seller in the same transaction.
- [ ] Batch settlement engine successfully executes up to 100 trades per Besu transaction block, reverting atomically if any trade fails.
- [ ] `NBSEFeeCollector.sol` accurately records and distributes accumulated fees across Exchange Treasury, Clearing Corp Reserve, IPF, and Infrastructure Pool in strict accordance with the 10,000 BPS allocation sum.
- [ ] SGF Default Waterfall interface executes deterministic escalation to `ISettlementGuaranteeFund.sol` upon counterparty default.
- [ ] 100% test coverage achieved across Foundry unit, fuzz, and invariant test suites.
- [ ] Slither and Mythril static security analyzers pass with zero critical, high, or medium severity findings.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Platform Evaluation & Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `303` (Token Issuance Smart Contract), Prompt `307` (Multi-Party Authorization & Multisig Governance), Prompt `315` (Settlement Guarantee Fund & Default Waterfall Smart Contract).
- **Parallel Tasks:** Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `230` (Settlement Guarantee Fund & Default Waterfall Service).
- **Subsequent Prompts Enabled:** Prompt `309` (Blockchain Event Indexer & State Sync), Prompt `215` (Reconciliation Engine), Prompt `216` (Regulatory Reporting Service).
