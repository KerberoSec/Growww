# 331 - Primary Market Order Routing & Clearing Bridge Smart Contracts (PrimaryMarketBridge.sol, MarketSessionQueue.sol)

## Purpose
Traditional equity exchanges in India (National Stock Exchange [NSE] and Bombay Stock Exchange [BSE]) operate under strict, time-delimited trading windows (09:15 to 15:30 IST on non-holiday business weekdays). In contrast, the Growww National Blockchain Stock Exchange (NBSE) architecture facilitates 24/7 continuous secondary market trading and instant atomic Delivery-versus-Payment (DvP) settlement for tokenized Indian equities on a permissioned Hyperledger Besu ledger.

This fundamental temporal mismatch creates a liquidity and settlement barrier: off-market retail and institutional orders placed outside 09:15-15:30 IST cannot interact directly with primary exchange order books or depository settlement systems (NSDL/CDSL). Conversely, 24/7 on-chain participants require continuous liquidity, off-market order staging, automated internalized crossing, and deterministic morning session primary market batch routing.

This prompt specifies the **Primary Market Order Routing & Clearing Bridge Smart Contract Suite (`PrimaryMarketBridge.sol`, `MarketSessionQueue.sol`)**. The contracts manage the hybrid lifecycle between traditional exchange trading sessions and 24/7 on-chain secondary markets. The suite enforces session-aware order queuing, off-market escrow lockup, off-chain crossing and liquidity pooling, consolidated morning batch execution against primary exchange FIX gateways, on-chain custody rebalancing (minting/burning ERC-3643 tokens upon physical demat delivery), and Growww's canonical Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with a 0.00% fee at launch (governed by FeeController.sol) revenue split (with FIFO capital gains computed strictly for user tax compliance under Section 111A/112A).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/bridge/` comprising:
- `MarketSessionQueue.sol`: High-throughput, session-aware order staging and escrow vault. It tracks Indian Standard Time (IST = UTC+5:30) trading sessions (PRE_OPEN, REGULAR_OPEN, POST_CLOSE, CLOSED, WEEKEND_HOLIDAY), cryptographically queues off-market limit and market orders with locked collateral (eINR / CBDC or ERC-3643 tokens), and supports off-market peer-to-peer internalization crossing.
- `PrimaryMarketBridge.sol`: Central clearing bridge connecting on-chain queued orders with domestic exchange routing gateways. It aggregates queued orders into deterministic batch manifests for morning primary exchange routing (09:15 IST), ingests cryptographically signed FIX trade execution reports from authorized broker relayers, validates DvP settlement receipts, triggers atomic custody mint/burn adjustments, and routes Universal Zero-Fee Model (0.00% fee - No fee at all)s to the domestic Treasury (60%), Core SGF (25%), and IPF (15%) reserve vaults.
- `IMarketSessionQueue.sol` and `IPrimaryMarketBridge.sol`: Comprehensive Solidity interface specifications with detailed data structures, enums, events, custom error definitions, and view/mutator function signatures.
- Comprehensive Foundry test harness (`test/bridge/`): Unit, fuzz, invariant, and session transition test suites validating 100% escrow backing, zero front-running, batch reconciliation accuracy, and strict compliance with the fixed 0.00% transaction fee (No fee at all) invariant.

## Scope Boundaries
- **In Scope:**
  - On-chain IST trading session state machine tracking and time-window validation.
  - Off-market order queuing with 100% escrowed collateral (eINR for buys, ERC-3643 digital security tokens for sells).
  - Priority queue management (price-time priority and FIFO bucket ordering for off-market orders).
  - Internalized crossing settlement for matching off-market buy/sell orders during closed sessions.
  - Morning session batch manifest creation and cryptographic dispatch for primary exchange execution (NSE/BSE).
  - Primary execution report ingestion via EIP-712 multi-signature attestations from licensed broker relayers.
  - Partial fill handling, remaining order rollover, and user cancellation workflows with instant escrow refund.
  - 1:1 Physical-to-digital custody synchronizer hooks (minting ERC-3643 on primary share purchase, burning on primary sale).
  - Growww fixed 0.00% transaction fee (No fee at all) calculation on turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), with FIFO capital gains for tax compliance (Section 111A/112A).
  - Emergency circuit breakers, pause mechanisms, and timelocked multi-signature administrative controls.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain FIX 5.0 SP2 / ITCH / OUCH trading gateway implementation (handled in Prompt 225).
  - Real-time pre-trade risk and margin check engine (handled in Prompt 206).
  - High-frequency matching engine for 24/7 on-chain continuous trading (handled in Prompt 205).
  - Depository participant API adapter for physical NSDL/CDSL demat execution (handled in Prompt 213).
  - Real-time market surveillance and dynamic volatility circuit breakers (handled in Prompt 228).
  - Fiat banking rail integrations and UPI payment hold reservation (handled in Prompt 203 and Prompt 212).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD`, and custom errors).
- **Token Standards:** **ERC-3643** (Permissioned Digital Securities) for tokenized Indian equities; **ERC-20** for tokenized fiat cash settlement (eINR / CBDC).
- **Security & Modularity:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Signature & Cryptography:** OpenZeppelin `ECDSA` and `EIP712Upgradeable` for relayer trade execution attestations and session oracle price feeds.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation and invariant fuzzing, `cast` for Besu RPC interactions).
- **Static Analysis & Auditing:** Slither, Mythril, and Solhint automated CI pipelines.

## Backend / Infra Touchpoints
- **FIX Protocol Gateway (Prompt 225):** Bridges primary market batch manifests into standard FIX `NewOrderSingle` and `NewOrderList` messages routed to NSE NEAT / BSE BOLT systems.
- **Trade Settlement Service (Prompt 208):** Coordinates execution reports between domestic exchange clearing corporations (NSCCL/ICCL) and `PrimaryMarketBridge.sol`.
- **Custodian & Depository Integration Service (Prompt 213):** Synchronizes physical NSDL/CDSL demat account balances with on-chain ERC-3643 token mint/burn events.
- **Market Data & Session Oracle Service (Prompt 207):** Publishes cryptographically attested exchange market session state transitions and official closing/opening tick prices.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes fixed 0.00% transaction fee (No fee at all) deductions (0.00% fee at launch (governed by FeeController.sol) split) and FIFO capital gains tax compliance records.
- **Blockchain Event Indexer (Prompt 309):** Ingests events (`OrderQueued`, `OrderInternalized`, `BatchManifestDispatched`, `PrimaryTradeCleared`, `EscrowRefunded`) for real-time mobile/web status streaming.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Executes on Hyperledger Besu private consortium network with QBFT consensus, 2-second deterministic block times, and immediate finality without chain reorganizations.
- **100% Escrow Collateral Invariant:** No buy order can be queued without full eINR cash locked in `MarketSessionQueue.sol`; no sell order can be queued without full ERC-3643 equity tokens transferred into escrow. Naked routing or unbacked orders are rejected at the contract level.
- **1:1 Custody Parity:** Token creation and destruction are strictly coupled to primary market trade clearing. Buying shares on NSE/BSE results in an ERC-3643 minting transaction only when depository receipt is attested; selling on NSE/BSE results in an ERC-3643 burn before physical demat delivery.
- **Zero On-Chain PII:** Smart contract state stores only `bytes32 orderId`, `bytes32 batchId`, `address investorLedgerAddress`, `bytes32 isinHash`, `uint256 quantity`, `uint256 limitPrice`, and cryptographic execution hashes. No names, PANs, demat numbers, or banking credentials exist on-chain.
- **Multi-Party Governance:** Administrative modifications (whitelisting bridge relayers, setting session parameters, updating fee treasury vaults) require 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:** Initialize `contracts/bridge/PrimaryMarketBridge.sol`, `contracts/bridge/MarketSessionQueue.sol`, interfaces under `contracts/interfaces/bridge/`, and test suites under `test/bridge/`.
2. **Define Data Models and Interfaces:** Write `IMarketSessionQueue.sol` and `IPrimaryMarketBridge.sol` declaring all enums (`SessionState`, `OrderSide`, `OrderType`, `OrderStatus`, `ClearingStatus`), structs, custom errors, and events.
3. **Implement Session State Machine in Queue:**
   - Define session timetable mapping IST hours/minutes into `SessionState` (`CLOSED`, `PRE_OPEN`, `REGULAR_OPEN`, `POST_CLOSE`, `HOLIDAY`).
   - Implement oracle-driven and timestamp-driven session transition logic with authorized session updater roles.
4. **Implement Escrow Lockup and Order Enqueueing:**
   - In `MarketSessionQueue.sol`, implement `queueOrder` accepting buy orders (locking `IERC20` eINR via `SafeERC20`) and sell orders (locking `IERC3643` tokens).
   - Validate investor KYC status via `IIdentityRegistry` compliance hooks.
   - Assign deterministic `orderId = keccak256(abi.encode(msg.sender, isin, side, price, quantity, nonce, block.timestamp))`.
   - Maintain doubly-linked queue index per ISIN to enforce price-time priority.
5. **Implement Off-Market Order Cancellation and Escrow Refund:**
   - Implement `cancelQueuedOrder` permitting the original investor to cancel unexecuted orders during `CLOSED` or `PRE_OPEN` states.
   - Atomically refund locked eINR or ERC-3643 tokens using Checks-Effects-Interactions (CEI).
6. **Implement Off-Market Internalized Crossing:**
   - Implement `internalizeCrossing` in `MarketSessionQueue.sol` allowing authorized settlement operators to match complementary buy/sell queued orders off-market.
   - Execute atomic DvP swap between matched buyer and seller escrows.
   - Assess 0.00% fee (No fee at all) on trade notional turnover with 0.00% fee at launch (future fee parameters governed by FeeController.sol) and record tax compliance metadata.
7. **Implement Batch Manifest Construction:**
   - In `PrimaryMarketBridge.sol`, implement `generateMorningBatchManifest` to aggregate uncrossed queued orders per ISIN when market transitions to `PRE_OPEN` / `REGULAR_OPEN`.
   - Pack orders into optimized arrays (`bytes32[] orderIds`, `uint256 netQuantity`, `uint256 limitPriceCap`) and emit `BatchManifestDispatched`.
8. **Implement EIP-712 Execution Report Ingestion:**
   - Define EIP-712 domain separator: `EIP712("GrowwwPrimaryMarketBridge", "1")`.
   - Implement `clearPrimaryExecutionReport` accepting signed broker execution receipts (`manifestId`, `executedQuantity`, `executionPricePaise`, `brokerExecutionId`, `depositoryReceiptHash`).
   - Verify relayer ECDSA signature against whitelisted broker relayer addresses.
9. **Implement Custody Mint / Burn Synchronizer Hooks:**
   - For executed buy orders: call `IDigitalSecurityToken.mint(buyer, executedUnits, depositoryReceiptHash)` upon primary market demat credit confirmation.
   - For executed sell orders: call `IDigitalSecurityToken.burn(escrowVault, executedUnits)` to extinguish on-chain tokens corresponding to demat debit.
   - Refund unexecuted escrow balances for partial fills.
10. **Implement Fixed Transaction Fee & Multi-Vault Allocation Engine:**
    - Compute trade turnover: `grossTurnover = executionPrice * executedUnits`.
    - Calculate fixed platform fee: `fee = (grossTurnover * feeBps) / 10000 (where feeBps == 0 at launch)` (exact 0.00% (Zero Fee) / 0 bps (0.00% fee at launch)).
    - Distribute fee: routed dynamically per FeeController governance (0.00% at launch).
    - Record tax compliance leaf hash for off-chain FIFO capital gains reporting (Section 111A/112A).
11. **Implement Circuit Breakers and Governance Controls:**
    - Inherit OpenZeppelin `PausableUpgradeable` and `AccessControlUpgradeable`.
    - Restrict emergency pause, token unfreezing, and parameter adjustment to `DEFAULT_ADMIN_ROLE` (`MultiSigGovernance.sol`).
12. **Author Comprehensive Foundry Unit and Invariant Tests:**
    - Test order queuing during closed market sessions and verify 100% escrow balance locking.
    - Test session transition triggers (e.g. 09:00 PRE_OPEN, 09:15 REGULAR_OPEN, 15:30 CLOSED).
    - Invariant test: escrowed token balance in queue must exactly equal sum of all active queued order obligations.
    - Fuzz test partial fill distributions and residual refund calculations across 10,000 randomized trade batches.
    - Validate 0.00% (Zero Fee) fixed transaction fee arithmetic and 0.00% fee launch policy revenue split.
13. **Perform Slither Static Analysis and Gas Profiling:**
    - Run `slither contracts/bridge/` and verify zero high/medium security vulnerabilities.
    - Profile gas consumption on Hyperledger Besu for batch clearing of 100 orders per transaction.
14. **Deploy and Validate on Hyperledger Besu Devnet:**
    - Deploy UUPS proxy implementations using Foundry script `script/DeployBridge.s.sol`.
    - Connect with mock FIX gateway (Prompt 225) and verify end-to-end off-market queuing, batch routing, and DvP settlement.

## Interfaces / Contracts

### 1. Market Session Queue Interface (`IMarketSessionQueue.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IMarketSessionQueue
 * @notice Interface for the session-aware order queuing and escrow contract.
 * @dev Manages off-market order staging, IST session transitions, and internalized crossing.
 */
interface IMarketSessionQueue {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum SessionState {
        CLOSED,           // Market closed (15:30 - 09:00 IST, overnight)
        PRE_OPEN,         // Pre-open session (09:00 - 09:15 IST)
        REGULAR_OPEN,     // Normal primary market trading (09:15 - 15:30 IST)
        POST_CLOSE,       // Post-closing session (15:40 - 16:00 IST)
        HOLIDAY           // Weekend or exchange statutory trading holiday
    }

    enum OrderSide {
        BUY,
        SELL
    }

    enum OrderType {
        LIMIT,
        MARKET
    }

    enum OrderStatus {
        QUEUED,           // Actively waiting in queue
        PARTIALLY_FILLED, // Partially matched or routed
        FILLED,           // Completely executed
        CANCELLED,        // Cancelled by user with escrow refunded
        ROUTED_TO_BRIDGE  // Transferred to PrimaryMarketBridge for exchange routing
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct QueuedOrder {
        bytes32 orderId;             // Unique cryptographic order identifier
        address investor;            // On-chain wallet address of investor
        address tokenAddress;        // Address of ERC-3643 DigitalSecurityToken
        bytes32 isin;                // Standardized 12-byte equity ISIN hash
        OrderSide side;              // BUY or SELL
        OrderType orderType;         // LIMIT or MARKET
        OrderStatus status;          // Current lifecycle status
        uint256 totalQuantity;       // Total shares requested (18 decimals)
        uint256 filledQuantity;      // Shares already filled (18 decimals)
        uint256 limitPricePaise;     // Limit price in paise (INR * 100, 0 for MARKET)
        uint256 escrowedAmount;      // Locked eINR paise (for BUY) or token units (for SELL)
        uint256 costBasisPaise;      // Cost basis per share for tax reporting
        uint64 queuedTimestamp;      // Unix timestamp when order was enqueued
        uint64 expiryTimestamp;      // Expiry timestamp after which order auto-cancels
        uint256 nonce;               // Replay protection nonce
    }

    struct InternalizedMatch {
        bytes32 buyOrderId;          // Queued buy order identifier
        bytes32 sellOrderId;         // Queued sell order identifier
        uint256 matchQuantity;       // Quantity to cross atomically (18 decimals)
        uint256 matchPricePaise;     // Execution price in paise
        uint256 sellerCostBasis;     // Seller cost basis in paise for tax compliance
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event SessionStateUpdated(
        SessionState indexed previousState,
        SessionState indexed newState,
        uint64 timestamp,
        address indexed updater
    );

    event OrderEnqueued(
        bytes32 indexed orderId,
        address indexed investor,
        bytes32 indexed isin,
        OrderSide side,
        OrderType orderType,
        uint256 quantity,
        uint256 limitPricePaise,
        uint256 escrowedAmount
    );

    event OrderCancelled(
        bytes32 indexed orderId,
        address indexed investor,
        uint256 refundedAmount,
        string reason
    );

    event OrdersInternalizedCross(
        bytes32 indexed buyOrderId,
        bytes32 indexed sellOrderId,
        bytes32 indexed isin,
        uint256 matchQuantity,
        uint256 matchPricePaise,
        uint256 grossVolumePaise,
        uint256 platformFeePaise
    );

    event OrdersLockedForRouting(
        bytes32 indexed batchManifestId,
        bytes32[] orderIds,
        uint256 totalQuantity
    );

    event EscrowClaimedByBridge(
        bytes32 indexed orderId,
        address indexed bridgeAddress,
        uint256 claimedAmount
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error InvalidSessionTransition(SessionState current, SessionState requested);
    error TradingSessionRestricted(SessionState current);
    error InvalidOrderParameters(string reason);
    error InsufficientEscrowAllowance(uint256 required, uint256 provided);
    error OrderNotFound(bytes32 orderId);
    error OrderNotActive(bytes32 orderId, OrderStatus status);
    error UnauthorizedOrderCancellation(address caller, address owner);
    error OrderExpired(bytes32 orderId, uint64 expiry);
    error NonceAlreadyUsed(address investor, uint256 nonce);
    error UnauthorizedBridge(address caller);
    error UnauthorizedSessionOracle(address caller);
    error KYCVerificationFailed(address investor);

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function queueOrder(
        bytes32 isin,
        address token,
        OrderSide side,
        OrderType orderType,
        uint256 quantity,
        uint256 limitPricePaise,
        uint256 costBasisPaise,
        uint64 expiryTimestamp,
        uint256 nonce
    ) external returns (bytes32 orderId);

    function cancelQueuedOrder(bytes32 orderId, string calldata reason) external;
    function internalizeCrossing(InternalizedMatch calldata matchDetails) external;
    function lockOrdersForPrimaryRouting(
        bytes32[] calldata orderIds,
        bytes32 manifestId
    ) external;
    function releaseRefundEscrow(bytes32 orderId, uint256 refundAmount) external;
    function setSessionState(SessionState newState) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function currentSession() external view returns (SessionState);
    function getOrder(bytes32 orderId) external view returns (QueuedOrder memory);
    function isOrderActive(bytes32 orderId) external view returns (bool);
    function getQueuedOrdersByISIN(bytes32 isin) external view returns (bytes32[] memory);
    function getInvestorActiveOrders(address investor) external view returns (bytes32[] memory);
    function primaryBridge() external view returns (address);
    function paymentToken() external view returns (address);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
}
```

### 2. Primary Market Bridge Interface (`IPrimaryMarketBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import { IMarketSessionQueue } from "./IMarketSessionQueue.sol";

/**
 * @title IPrimaryMarketBridge
 * @notice Interface for the primary exchange routing and batch clearing bridge.
 * @dev Coordinates morning batch dispatch, FIX execution ingestion, and custody DvP sync.
 */
interface IPrimaryMarketBridge {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum ManifestStatus {
        STAGED,          // Manifest constructed and orders locked
        DISPATCHED,      // Sent to primary exchange FIX gateway
        PARTIALLY_EXECUTED, // Part of the manifest executed on primary market
        FULLY_EXECUTED,  // Manifest completely executed and cleared
        RECONCILED,      // Demat custody and on-chain ledger completely balanced
        CANCELLED        // Manifest cancelled and escrows refunded
    }

    enum ClearingStatus {
        PENDING_EXECUTION,
        EXECUTED_ON_EXCHANGE,
        DEMAT_CONFIRMED,
        SETTLED_DVP,
        REJECTED
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct BatchManifest {
        bytes32 manifestId;          // Unique routing batch identifier
        bytes32 isin;                // Security ISIN identifier hash
        address tokenAddress;        // ERC-3643 token address
        IMarketSessionQueue.OrderSide side; // BUY or SELL
        uint256 aggregateQuantity;   // Total shares across all included orders (18 decimals)
        uint256 executedQuantity;    // Cumulative shares executed on primary exchange
        uint256 weightedAveragePricePaise; // Volume-weighted average execution price in paise
        uint256 totalGrossPaise;     // Total fiat gross amount in paise
        uint256 totalPlatformFeePaise; // Total 0.00% (Zero Fee) platform fees collected
        ManifestStatus status;       // Current batch lifecycle status
        uint64 createdAt;            // Unix creation timestamp
        uint64 dispatchedAt;         // Unix dispatch timestamp
        uint64 settledAt;            // Unix final DvP settlement timestamp
        bytes32[] orderIds;          // Constituent queued order IDs
    }

    struct PrimaryExecutionReport {
        bytes32 manifestId;          // Associated batch manifest ID
        bytes32 brokerExecutionId;   // Unique exchange/broker trade identifier
        uint256 executedUnits;       // Number of shares executed (18 decimals)
        uint256 executionPricePaise; // Executed price per share in paise
        uint256 totalConsiderationPaise; // Gross executed value in paise
        uint64 executionTimestamp;   // Primary exchange matching timestamp
        bytes32 depositoryReceiptHash; // NSDL/CDSL demat settlement commitment hash
        bytes relayerSignature;      // EIP-712 cryptographic attestation by authorized broker
    }

    struct OrderClearingReceipt {
        bytes32 orderId;             // Queued order ID
        uint256 filledUnits;         // Units allocated to this order (18 decimals)
        uint256 grossAmountPaise;    // Gross consideration in paise
        uint256 feeDeductedPaise;    // 0.00% (Zero Fee) platform fee deducted
        uint256 refundAmount;        // Unfilled escrow returned
        bytes32 taxProofHash;        // Section 111A/112A tax compliance attestation
        ClearingStatus status;       // Final clearing status
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event BatchManifestCreated(
        bytes32 indexed manifestId,
        bytes32 indexed isin,
        IMarketSessionQueue.OrderSide indexed side,
        uint256 aggregateQuantity,
        uint256 orderCount
    );

    event BatchManifestDispatched(
        bytes32 indexed manifestId,
        uint64 dispatchedTimestamp,
        address indexed relayer
    );

    event PrimaryExecutionCleared(
        bytes32 indexed manifestId,
        bytes32 indexed brokerExecutionId,
        uint256 executedUnits,
        uint256 executionPricePaise,
        bytes32 depositoryReceiptHash
    );

    event OrderSettledDvP(
        bytes32 indexed manifestId,
        bytes32 indexed orderId,
        address indexed investor,
        uint256 filledUnits,
        uint256 grossPaise,
        uint256 platformFeePaise,
        bytes32 taxProofHash
    );

    event CustodyTokenMinted(
        address indexed tokenAddress,
        address indexed recipient,
        uint256 amount,
        bytes32 depositoryReceiptHash
    );

    event CustodyTokenBurned(
        address indexed tokenAddress,
        address indexed from,
        uint256 amount,
        bytes32 depositoryReceiptHash
    );

    event PlatformFeeDistributed(
        bytes32 indexed manifestId,
        bytes32 indexed orderId,
        uint256 feeAmountPaise,
        uint256 treasurySharePaise,
        uint256 coreSgfSharePaise,
        uint256 ipfSharePaise
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error ManifestNotFound(bytes32 manifestId);
    error InvalidManifestState(bytes32 manifestId, ManifestStatus status);
    error InvalidBrokerSignature(bytes32 manifestId, address recovered);
    error DuplicateExecutionReport(bytes32 brokerExecutionId);
    error ExecutionQuantityExceedsManifest(uint256 executed, uint256 aggregate);
    error SettlementMismatch(uint256 expected, uint256 actual);
    error UnauthorizedRelayer(address caller);
    error StaleExecutionTimestamp(uint64 timestamp, uint64 cutoff);
    error FeeCalculationViolation(uint256 expectedFee, uint256 providedFee);
    error DepositoryProofInvalid(bytes32 receiptHash);

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function createBatchManifest(
        bytes32 isin,
        IMarketSessionQueue.OrderSide side,
        bytes32[] calldata orderIds
    ) external returns (bytes32 manifestId);

    function markManifestDispatched(bytes32 manifestId) external;

    function processPrimaryExecutionReport(
        PrimaryExecutionReport calldata report
    ) external;

    function finalizeBatchOrderSettlement(
        bytes32 manifestId,
        OrderClearingReceipt[] calldata orderReceipts
    ) external;

    function cancelBatchManifest(bytes32 manifestId, string calldata reason) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getManifest(bytes32 manifestId) external view returns (BatchManifest memory);
    function isExecutionProcessed(bytes32 brokerExecutionId) external view returns (bool);
    function marketSessionQueue() external view returns (address);
    function eip712DomainSeparator() external view returns (bytes32);
    function isWhitelistedRelayer(address relayer) external view returns (bool);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function computeTransactionFee(
        uint256 grossTurnoverPaise
    ) external pure returns (uint256 feePaise, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare);
}
```

## Security & Compliance Notes
- **100% Escrow and Solvency Invariant:** Orders entering `MarketSessionQueue.sol` require upfront transfer and locking of full collateral (eINR for buys, ERC-3643 tokens for sells). The contract enforces strict conservation of value: total locked token/fiat balances must always equal or exceed active queued liabilities.
- **Fixed Platform Fee Integrity:** In strict accordance with the Growww Core Fee Model (Prompt 006), the platform fee is assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover with an automated 0.00% fee at launch (governed by FeeController.sol) revenue split. Capital gains are computed strictly for off-chain user tax compliance (Section 111A/112A). Zero holding or custody fees are charged.
- **Cryptographic Replay and Tamper Resistance:** Primary market execution reports are signed using structured EIP-712 typed signatures (`PrimaryExecutionReport`). Every execution contains a unique `brokerExecutionId` recorded in a permanent mapping (`mapping(bytes32 => bool) isExecutionProcessed`) preventing duplicate clearance or replay attacks.
- **Session Enforcement and Anti-Front-Running:** The smart contract suite validates session timestamps and prevents unauthorized crossing during primary market active trading windows (`REGULAR_OPEN`), ensuring all off-market crossings occur deterministically at verified closing or mid-market oracle benchmark prices.
- **Identity & KYC Transfer Hooks:** Secondary settlements and internalized crossings invoke `IIdentityRegistry.isVerified(investor)` on both buyer and seller accounts via underlying ERC-3643 compliance hooks (Prompt 305), ensuring full SEBI/RBI KYC and sanctions compliance.
- **Multi-Signature Governance:** Privileged functions (setting session oracles, whitelisting broker relayers, updating fee vaults, and UUPS contract upgrades) require 3-of-5 multi-signature authorization from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Acceptance Criteria
- [ ] `IMarketSessionQueue.sol` and `IPrimaryMarketBridge.sol` compiled under Solidity 0.8.24 with zero warnings.
- [ ] Order queuing strictly requires and verifies 100% collateral transfer (eINR for buys, ERC-3643 tokens for sells).
- [ ] Session state machine accurately reflects IST trading sessions (`CLOSED`, `PRE_OPEN`, `REGULAR_OPEN`, `POST_CLOSE`, `HOLIDAY`).
- [ ] User order cancellation during off-market sessions immediately returns 100% escrowed collateral with zero lockup penalty.
- [ ] Off-market internalized crossing executes atomic DvP balance transfer between buyer and seller with accurate 0.00% transaction fee (No fee at all) deduction and 0.00% fee at launch (future fee parameters governed by FeeController.sol).
- [ ] Morning batch manifest aggregation locks queued orders and computes correct volume-weighted parameters for FIX dispatch.
- [ ] EIP-712 broker execution report ingestion validates cryptographic ECDSA signatures, timestamp freshness, and replay nonce uniqueness.
- [ ] 1:1 Custody synchronization hooks successfully mint ERC-3643 tokens on verified demat credit receipts and burn on demat debits.
- [ ] Partial execution handling correctly calculates remaining escrow refunds and updates order statuses.
- [ ] Fixed turnover fee evaluation tests prove mathematically that fee equals 0 bps (0.00% fee at launch) of turnover with 0.00% fee at launch (governed by FeeController.sol) distribution, with capital gains calculated for Section 111A/112A tax compliance.
- [ ] Invariant fuzz tests in Foundry execute 10,000 runs confirming total locked escrow balances equal active unfulfilled order obligations.
- [ ] Slither and Mythril static analysis suites pass with zero high or medium severity warnings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract - ERC-3643), Prompt `305` (Transfer Compliance Hooks), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `205` (Order Matching Engine), Prompt `208` (Trade Settlement Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `213` (Custodian Depository Integration Service), Prompt `225` (FIX Protocol Gateway).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (Proof of Reserve Registry), Prompt `309` (Event Indexing Service), Prompt `509` (Flutter Order Placement Flow).
