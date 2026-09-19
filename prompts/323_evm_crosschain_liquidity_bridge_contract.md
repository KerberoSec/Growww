# 323 - EVM Cross-Chain Liquidity Bridge Contract (Chainlink CCIP / LayerZero v2 / Multi-Token Collateral)

## Purpose
The Growww institutional trading ecosystem requires high-throughput, low-latency, and provably secure multi-token collateral ingestion and liquidity bridging across EVM-compatible networks (Ethereum Mainnet, Arbitrum One, Base, Polygon, Optimism, and the Growww Permissioned Appchain). Institutional investors, market makers, and liquidity providers deposit high-grade collateral (such as USDC, USDT, WETH, WBTC, tokenized sovereign bonds, and regulated equity securities) on external public or permissioned chains to fund 24/7 continuous trading accounts and DvP settlement accounts on Growww.

However, cross-chain bridging represents one of the highest risk vectors in decentralized and institutional finance, vulnerable to bridge validator compromise, smart contract reentrancy, replay attacks, oracle manipulation, and sudden liquidity drains. To establish an institutional-grade, fault-tolerant infrastructure, this prompt defines the architecture, specifications, state models, and Solidity interfaces for the **EVM Cross-Chain Liquidity Bridge Contract Suite** (`CrossChainLiquidityBridge.sol`, `CCIPReceiver.sol`, `LayerZeroReceiverAdapter.sol`, and `BridgeRateLimiter.sol`).

The smart contract suite implements dual-transport cross-chain messaging (supporting both Chainlink CCIP v1.5+ and LayerZero v2), multi-token collateral lock-and-mint and burn-and-mint architectures, granular per-token sliding-window rate limiters, multi-tier emergency pause circuit breakers, and deterministic execution retry mechanisms.

## What You Are Building
A modular, production-ready EVM smart contract suite and testing infrastructure under `contracts/bridge/crosschain/` comprising:
- `CrossChainLiquidityBridge.sol`: The central bridge controller contract managing multi-token collateral deposits, lock-and-mint, burn-and-mint, release-and-unlock operations, cross-chain fee handling, and destination payload dispatch.
- `CCIPReceiver.sol`: Chainlink CCIP receiver adapter inheriting from Chainlink CCIP standards (`Client.Any2EVMMessage`), validating source chain selectors, verifying authenticated remote sender contracts, handling gas limits, and routing decoded payloads to the bridge core.
- `LayerZeroReceiverAdapter.sol`: LayerZero v2 receiver adapter implementing `ILayerZeroReceiver` and OFT/OApp messaging patterns, validating endpoint IDs, verifying remote peers, and executing inbound message payloads.
- `BridgeRateLimiter.sol`: Adaptive sliding-window rate limiter contract enforcing configurable hourly capacity, daily volume limits, single-transaction maximums, and chain-specific throughput quotas.
- `EmergencyPauseCircuit.sol`: Multi-tier emergency circuit breaker enabling instant single-token, single-chain, or global operational halts triggered by automated risk monitoring daemons or multi-sig guardians.
- Comprehensive Foundry unit, integration, invariant, and differential fuzz testing suites (`test/bridge/CrossChainLiquidityBridge.t.sol`) simulating multi-chain message delivery, reentrancy attacks, rate limit overflows, and recovery flows.

## Scope Boundaries
- **In Scope:**
  - Solidity smart contract interfaces, data structures, custom errors, events, and state specifications for `CrossChainLiquidityBridge.sol`, `CCIPReceiver.sol`, `LayerZeroReceiverAdapter.sol`, and `BridgeRateLimiter.sol`.
  - Multi-protocol transport routing for Chainlink CCIP and LayerZero v2.
  - Multi-token collateral ingestion architecture supporting standard ERC-20 tokens, rebasing/fee-on-transfer protection, and ERC-3643 compliant permissioned security tokens.
  - Both lock-and-mint (vault escrow on origin chain -> mint synthetic representation on destination chain) and burn-and-mint (burn synthetic representation on origin chain -> mint or unlock collateral on destination chain) bridging modes.
  - Dynamic sliding-window rate limiting and volume threshold throttles.
  - Multi-tier emergency pause circuits (global pause, token-specific freeze, destination-chain circuit breaker) with guardian role controls.
  - Inbound message retry queue for handling execution failures without fund loss.
  - Comprehensive Foundry invariant and fuzz test specifications.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain relayer daemon implementation and chain monitoring microservices (specified in Prompt 309, Prompt 310, and Prompt 319).
  - Inter-entity Domestic vs GIFT City legal and tax ledger bridge (specified in Prompt 313).
  - Appchain rollup sequencer and state root settlement contracts (specified in Prompt 316).
  - Off-chain fiat FX conversion and banking on-ramps (specified in Prompt 214).

## Technology to Use
- **Smart Contract Language:** Solidity 0.8.24 (Target EVM: Cancun / Prague).
  *Justification:* Native support for transient storage (`TSTORE`/`TLOAD`), custom errors for gas optimization, checked arithmetic, and robust user-defined value types.
- **Base Framework:** OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `SafeERC20`, `ERC1967Utils`, `UUPSUpgradeable`).
- **Cross-Chain Interoperability Protocols:**
  - **Chainlink CCIP (v1.5+):** Utilizing `IRouterClient`, `CCIPReceiver`, `Client.EVM2AnyMessage`, `Client.Any2EVMMessage`, and CCIP Risk Management Network integration.
  - **LayerZero v2:** Utilizing `ILayerZeroEndpointV2`, `ILayerZeroReceiver`, `Origin`, and LayerZero OApp messaging standards.
- **Development & Testing Toolchain:** Foundry (`forge` for compilation, unit testing, fuzzing, and invariant checks; `cast` for RPC calls and contract interaction scripts).
- **Static Analysis & Formal Verification:** Slither (Trail of Bits), Aderyn, Certora / Halmos for symbolic execution of invariant properties.

## Backend / Infra Touchpoints
- **Blockchain Event Indexer (Prompt 309):** Ingests `TokensBridged`, `TokensReceived`, `BridgeDepositLocked`, and `EmergencyHaltTriggered` events to maintain off-chain collateral balances in PostgreSQL and Kafka topics.
- **Chain Node Monitoring & Alerting (Prompt 310):** Monitors RPC connectivity, CCIP router status, LayerZero endpoint liveliness, and contract gas balances across all connected networks.
- **Institutional Custody Bridge (Prompt 319):** Interacts with institutional custody accounts and multi-sig timelocks for transfers exceeding large institutional thresholds.
- **Wallet & Account Ledger Service (Prompt 203):** Updates off-chain user ledger balances upon successful inbound collateral minting or unlock events.
- **Pre-Trade Risk & Margin Engine (Prompt 206):** Verifies deposited collateral availability and liquidity weighting factors for trading margin calculation.
- **Vault KMS / HSM (Prompt 311):** Secures guardian signing keys and multi-sig operational keys in FIPS 140-2 Level 3 Hardware Security Modules.

## Blockchain Interaction
- **Outbound Multi-Token Collateral Bridge Flow:**
  1. Investor or institutional participant calls `bridgeTokens(...)` on `CrossChainLiquidityBridge.sol`, supplying token address, amount, destination chain selector/endpoint ID, recipient address, chosen protocol (CCIP or LayerZero), and protocol-specific execution gas parameters.
  2. The bridge contract checks global, token, and chain pause states. If any circuit breaker is active, execution reverts with custom error `BridgeCircuitBreakerActive()`.
  3. The contract validates caller permissions (if token is ERC-3643 compliant, KYC verification hook is queried).
  4. The contract calls `BridgeRateLimiter.sol` to verify that the transfer does not exceed the single-transfer maximum, hourly capacity, or daily sliding-window limit. If limits are exceeded, execution reverts with `RateLimitExceeded()`.
  5. The bridge handles token custody based on collateral mode:
     - For Lock-and-Mint: Transfers tokens from caller to the bridge contract vault using OpenZeppelin `SafeERC20.safeTransferFrom`, verifying pre- and post-balance deltas to prevent fee-on-transfer discrepancies.
     - For Burn-and-Mint: Calls `IBurnableMintableToken(token).burnFrom(msg.sender, amount)` to destroy synthetic units.
  6. The contract computes protocol messaging fees (in native gas asset or LINK) and dispatches the payload via the designated adapter (`CCIPRouter` or `LayerZeroEndpointV2`).
  7. Emits `TokensBridged(transferId, sender, recipient, token, amount, destinationChain, protocol, nonce)`.

- **Inbound Collateral Ingestion & Receipt Flow:**
  1. The external messaging protocol (CCIP OffRamp or LayerZero Endpoint) calls the registered receiver contract (`CCIPReceiver.sol` or `LayerZeroReceiverAdapter.sol`).
  2. The receiver contract verifies that the source chain and source sender contract match the authorized peer registry. If unauthorized, execution reverts with `UnauthorizedSourceSender()`.
  3. The receiver decodes the payload: `(bytes32 transferId, address originalSender, address recipient, address originToken, address destinationToken, uint256 amount, uint8 collateralMode)`.
  4. The receiver contract forwards the execution call to `CrossChainLiquidityBridge.sol` via internal authorized call.
  5. The bridge contract checks destination rate limits, replay protection status (asserting `!processedTransfers[transferId]`), and pauses.
  6. Inbound delivery execution:
     - For Lock-and-Mint: Calls `IBurnableMintableToken(destinationToken).mint(recipient, amount)`.
     - For Burn-and-Mint: Transfers collateral from bridge vault escrow to `recipient` using `SafeERC20.safeTransfer`.
  7. Marks `processedTransfers[transferId] = true` and emits `TokensReceived(transferId, recipient, destinationToken, amount, sourceChain, originSender)`.
  8. If an inbound execution fails due to recipient contract revert or temporary condition, the message is stored in the `failedMessages` queue with its full payload hash, enabling authorized retry or admin recovery without loss of funds.

## Step-by-Step Build Instructions
1. Initialize Foundry project workspace under `contracts/bridge/crosschain/` with directories: `src/`, `src/interfaces/`, `src/adapters/`, `src/security/`, `test/`, and `script/`.
2. Install OpenZeppelin Contracts Upgradeable v5.0, Chainlink CCIP contracts, and LayerZero v2 contracts as dependencies.
3. Define core interface definitions:
   - `ICrossChainLiquidityBridge.sol`: Primary bridge interface specifying collateral deposit, burn, unlock, mint, rate limiting, and pause controls.
   - `ICCIPReceiverAdapter.sol`: Interface for Chainlink CCIP receiver and dispatch routing.
   - `ILayerZeroEndpointV2Adapter.sol`: Interface for LayerZero v2 endpoint integration and payload delivery.
   - `IBurnableMintableToken.sol`: Standardized interface for mintable and burnable synthetic token representations.
   - `IBridgeRateLimiter.sol`: Interface for sliding-window rate limiting logic.
   - `IEmergencyPauseCircuit.sol`: Interface for multi-tier pause circuits.
4. Implement `BridgeRateLimiter.sol` as an upgradeable contract:
   - Define data structure `RateLimitState` storing `hourlyCapacity`, `dailyCapacity`, `maxSingleTransfer`, `currentHourlyUsage`, `currentDailyUsage`, `lastHourlyWindowStart`, and `lastDailyWindowStart`.
   - Implement sliding-window calculation logic to decay and reset usage windows deterministically based on `block.timestamp`.
   - Implement `checkAndConsumeRateLimit(address token, uint256 amount, uint64 destinationChain)` returning boolean status and updating consumed volume.
   - Implement management functions restricted to `RATE_LIMIT_ADMIN_ROLE` for setting per-token and per-chain parameters.
5. Implement `EmergencyPauseCircuit.sol`:
   - Support three granular pause levels: `GlobalPause` (halts all bridge operations), `TokenPause` (halts specific token collateral), and `ChainPause` (halts specific remote chain).
   - Implement multi-role permissions: `EMERGENCY_GUARDIAN_ROLE` (can instantly trigger pause) and `TIMELOCK_ADMIN_ROLE` (required to unpause after safety review).
   - Emit detailed audit events on every state transition.
6. Implement `CCIPReceiver.sol`:
   - Inherit from Chainlink CCIP `CCIPReceiver` base contract.
   - Maintain mapping `allowedSourceChains` (uint64 chainSelector => bool) and `trustedSourceSenders` (uint64 chainSelector => bytes32 senderAddress).
   - Implement `_ccipReceive(Client.Any2EVMMessage memory message)` internal override:
     - Validate source chain and remote sender.
     - Decode message payload containing transfer details.
     - Invoke `CrossChainLiquidityBridge.deliverInboundTokens(...)`.
     - Catch execution reverts and store message in `failedMessageQueue` with `MessageFailed` event for manual resolution.
7. Implement `LayerZeroReceiverAdapter.sol`:
   - Implement `ILayerZeroReceiver` handling `lzReceive(Origin calldata _origin, bytes32 _guid, bytes calldata _message, address _executor, bytes calldata _extraData)`.
   - Verify `_origin.srcEid` against authorized endpoint registry and `_origin.sender` against configured peer addresses.
   - Forward decoded transfer payload to `CrossChainLiquidityBridge.deliverInboundTokens(...)`.
   - Enforce replay protection by tracking processed LayerZero GUIDs.
8. Implement `CrossChainLiquidityBridge.sol`:
   - Inherit from `Initializable`, `AccessControlUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, and `UUPSUpgradeable`.
   - Configure storage layout with storage gap for safe future upgradeability.
   - Implement `initialize(address defaultAdmin, address emergencyGuardian, address rateLimiter)` initializer with reinitialization protection.
   - Implement `bridgeTokens(...)` supporting both native and LINK fee payments for CCIP, as well as native fee quotes for LayerZero v2.
   - Implement `deliverInboundTokens(...)` callable only by authorized adapter contracts (`CCIPReceiver` and `LayerZeroReceiverAdapter`).
   - Implement `retryFailedMessage(bytes32 messageId)` allowing recovery of queued failed messages once prerequisite conditions (e.g. liquidity or recipient readiness) are satisfied.
   - Implement emergency withdrawal function restricted to multi-sig timelock governance for rescue of non-collateral tokens or stuck assets.
9. Write unit tests in Foundry (`test/bridge/CrossChainLiquidityBridge.t.sol`):
   - Test full deposit-lock-mint flow across simulated local and remote mocks.
   - Test burn-and-mint redemption flow.
   - Test fee-on-transfer token rejection and exact balance delta accounting.
   - Test unauthorized caller reverts on privileged functions.
10. Write fuzz and invariant test suite (`test/bridge/BridgeInvariants.t.sol`):
    - Invariant 1: Vault balance on origin chain must always be greater than or equal to the sum of locked collateral across all active users (`vaultBalance >= totalLockedCollateral`).
    - Invariant 2: Total synthetic tokens minted across destination chains must never exceed total locked collateral in the origin bridge vault (`sum(syntheticSupply) <= vaultLockedCollateral`).
    - Invariant 3: Rate limiter usage must never exceed configured `hourlyCapacity` or `dailyCapacity` in any given sliding window.
    - Invariant 4: No transfer can be executed twice (strict single-execution replay invariant).
11. Execute static analysis with Slither (`slither contracts/bridge/crosschain/src/`) and ensure 0 high, medium, or informational findings.
12. Create deployment scripts in Foundry (`script/DeployCrossChainBridge.s.sol`) configured for Ethereum Mainnet, Arbitrum, Base, Polygon, and Hyperledger Besu Appchain with deterministic CREATE2 proxy deployment.

## Interfaces / Contracts

### Primary Bridge Contract Interface (`ICrossChainLiquidityBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICrossChainLiquidityBridge {
    // Enums
    enum BridgeProtocol {
        CHAINLINK_CCIP,
        LAYERZERO_V2
    }

    enum CollateralMode {
        LOCK_AND_MINT,
        BURN_AND_MINT
    }

    enum TransferStatus {
        NON_EXISTENT,
        PENDING,
        COMPLETED,
        FAILED,
        REFUNDED
    }

    // Structs
    struct BridgeRequest {
        address token;
        uint256 amount;
        uint64 destinationChainSelector;
        uint32 destinationEid;
        bytes32 recipient;
        BridgeProtocol protocol;
        CollateralMode collateralMode;
        bytes extraArgs;
    }

    struct InboundTransferPayload {
        bytes32 transferId;
        bytes32 originSender;
        address recipient;
        address originToken;
        address destinationToken;
        uint256 amount;
        CollateralMode collateralMode;
        uint64 originChainSelector;
        uint32 originEid;
        uint256 nonce;
    }

    struct FailedMessageRecord {
        bytes32 messageId;
        BridgeProtocol protocol;
        uint64 sourceChainSelector;
        uint32 sourceEid;
        bytes32 sourceSender;
        bytes payload;
        bytes32 reasonHash;
        uint256 timestamp;
        bool resolved;
    }

    // Events
    event TokensBridged(
        bytes32 indexed transferId,
        address indexed sender,
        bytes32 indexed recipient,
        address token,
        uint256 amount,
        uint64 destinationChainSelector,
        uint32 destinationEid,
        BridgeProtocol protocol,
        CollateralMode collateralMode,
        uint256 feePaid
    );

    event TokensReceived(
        bytes32 indexed transferId,
        address indexed recipient,
        address indexed destinationToken,
        uint256 amount,
        uint64 originChainSelector,
        uint32 originEid,
        bytes32 originSender,
        CollateralMode collateralMode
    );

    event InboundMessageFailed(
        bytes32 indexed messageId,
        BridgeProtocol indexed protocol,
        uint64 sourceChainSelector,
        bytes32 sourceSender,
        bytes errorData
    );

    event FailedMessageRetried(
        bytes32 indexed messageId,
        bool success
    );

    event AdapterRegistered(
        BridgeProtocol indexed protocol,
        address indexed adapterAddress
    );

    event TokenMappingUpdated(
        address indexed localToken,
        uint64 indexed remoteChainSelector,
        address indexed remoteToken,
        CollateralMode collateralMode,
        bool isSupported
    );

    event EmergencyVaultWithdrawal(
        address indexed token,
        address indexed recipient,
        uint256 amount,
        address indexed adminCaller
    );

    // Custom Errors
    error BridgeCircuitBreakerActive();
    error TokenNotSupported(address token);
    error DestinationChainNotSupported(uint64 chainSelector, uint32 eid);
    error InvalidAmount();
    error InvalidRecipient();
    error InsufficientFee(uint256 provided, uint256 required);
    error RateLimitExceeded(address token, uint256 requested, uint256 available);
    error TransferAlreadyProcessed(bytes32 transferId);
    error UnauthorizedAdapter(address caller);
    error InboundDeliveryFailed(bytes32 transferId);
    error MessageNotFailed(bytes32 messageId);
    error MessageAlreadyResolved(bytes32 messageId);
    error CollateralInvariantViolation(uint256 vaultBalance, uint256 requiredCollateral);
    error ZeroAddressDetected();
    error ArrayLengthMismatch();

    // Outbound Functions
    function bridgeTokens(
        BridgeRequest calldata request
    ) external payable returns (bytes32 transferId);

    function getBridgeFee(
        BridgeRequest calldata request
    ) external view returns (uint256 fee);

    // Inbound Functions (Privileged to Registered Adapters)
    function deliverInboundTokens(
        InboundTransferPayload calldata payload
    ) external returns (bool success);

    function recordFailedInboundMessage(
        bytes32 messageId,
        BridgeProtocol protocol,
        uint64 sourceChainSelector,
        uint32 sourceEid,
        bytes32 sourceSender,
        bytes calldata payload,
        bytes calldata errorData
    ) external;

    function retryFailedMessage(
        bytes32 messageId
    ) external;

    // View / Inspection Functions
    function isTransferProcessed(bytes32 transferId) external view returns (bool);
    function getFailedMessage(bytes32 messageId) external view returns (FailedMessageRecord memory);
    function getLocalTokenForRemote(uint64 remoteChainSelector, address remoteToken) external view returns (address localToken, CollateralMode collateralMode, bool isSupported);
    function getVaultCollateralBalance(address token) external view returns (uint256);

    // Governance & Configuration
    function setAdapter(BridgeProtocol protocol, address adapter) external;
    function setTokenSupport(
        address localToken,
        uint64 remoteChainSelector,
        uint32 remoteEid,
        address remoteToken,
        CollateralMode collateralMode,
        bool isSupported
    ) external;
}
```

### Chainlink CCIP Receiver Adapter Interface (`ICCIPReceiverAdapter.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICCIPReceiverAdapter {
    // Events
    event CCIPMessageDispatched(
        bytes32 indexed messageId,
        uint64 indexed destinationChainSelector,
        bytes32 receiver,
        address feeToken,
        uint256 fees
    );

    event CCIPMessageReceived(
        bytes32 indexed messageId,
        uint64 indexed sourceChainSelector,
        bytes32 sender
    );

    event TrustedSenderUpdated(
        uint64 indexed sourceChainSelector,
        bytes32 indexed trustedSender,
        bool isTrusted
    );

    // Custom Errors
    error SourceChainNotAllowed(uint64 sourceChainSelector);
    error SourceSenderNotTrusted(uint64 sourceChainSelector, bytes32 sender);
    error InvalidRouter(address router);
    error CCIPPayloadExecutionReverted(bytes32 messageId, bytes reason);

    // External Functions
    function sendCCIPMessage(
        uint64 destinationChainSelector,
        bytes32 receiver,
        bytes calldata payload,
        uint256 gasLimit,
        address feeToken
    ) external payable returns (bytes32 messageId);

    function getCCIPFee(
        uint64 destinationChainSelector,
        bytes32 receiver,
        bytes calldata payload,
        uint256 gasLimit,
        address feeToken
    ) external view returns (uint256 fee);

    function setTrustedSender(
        uint64 sourceChainSelector,
        bytes32 sender,
        bool isTrusted
    ) external;

    function getRouter() external view returns (address);
    function isTrustedSender(uint64 sourceChainSelector, bytes32 sender) external view returns (bool);
}
```

### LayerZero v2 Endpoint Adapter Interface (`ILayerZeroEndpointV2Adapter.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ILayerZeroEndpointV2Adapter {
    // Structs
    struct Origin {
        uint32 srcEid;
        bytes32 sender;
        uint64 nonce;
    }

    // Events
    event LZMessageDispatched(
        bytes32 indexed guid,
        uint32 indexed dstEid,
        bytes32 receiver,
        uint256 nativeFee,
        uint256 lzTokenFee
    );

    event LZMessageReceived(
        bytes32 indexed guid,
        uint32 indexed srcEid,
        bytes32 sender
    );

    event PeerConfigured(
        uint32 indexed eid,
        bytes32 peerAddress
    );

    // Custom Errors
    error EndpointIdNotSupported(uint32 eid);
    error UnauthorizedPeer(uint32 srcEid, bytes32 sender);
    error InvalidEndpointV2(address endpoint);
    error LZPayloadExecutionReverted(bytes32 guid, bytes reason);

    // External Functions
    function sendLZMessage(
        uint32 dstEid,
        bytes32 receiver,
        bytes calldata payload,
        bytes calldata options
    ) external payable returns (bytes32 guid);

    function getLZFee(
        uint32 dstEid,
        bytes32 receiver,
        bytes calldata payload,
        bytes calldata options
    ) external view returns (uint256 nativeFee, uint256 lzTokenFee);

    function setPeer(uint32 eid, bytes32 peer) external;
    function getPeer(uint32 eid) external view returns (bytes32);
    function getEndpoint() external view returns (address);
}
```

### Bridge Rate Limiter Interface (`IBridgeRateLimiter.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IBridgeRateLimiter {
    // Structs
    struct TokenRateLimitConfig {
        uint256 hourlyCapacity;
        uint256 dailyCapacity;
        uint256 maxSingleTransfer;
        bool isEnabled;
    }

    struct SlidingWindowState {
        uint256 currentHourlyUsage;
        uint256 currentDailyUsage;
        uint256 windowStartTimestamp;
        uint256 lastUpdateTimestamp;
    }

    // Events
    event RateLimitConfigSet(
        address indexed token,
        uint64 indexed destinationChainSelector,
        uint256 hourlyCapacity,
        uint256 dailyCapacity,
        uint256 maxSingleTransfer,
        bool isEnabled
    );

    event RateLimitConsumed(
        address indexed token,
        uint64 indexed destinationChainSelector,
        uint256 amount,
        uint256 currentHourlyUsage,
        uint256 currentDailyUsage
    );

    // Custom Errors
    error MaxSingleTransferExceeded(address token, uint256 requested, uint256 maxAllowed);
    error HourlyRateLimitExceeded(address token, uint256 requested, uint256 availableCapacity);
    error DailyRateLimitExceeded(address token, uint256 requested, uint256 availableCapacity);
    error RateLimitingDisabledForToken(address token);

    // Operational Functions
    function checkAndConsumeRateLimit(
        address token,
        uint256 amount,
        uint64 destinationChainSelector
    ) external returns (bool allowed);

    function getAvailableHourlyCapacity(
        address token,
        uint64 destinationChainSelector
    ) external view returns (uint256 available);

    function getAvailableDailyCapacity(
        address token,
        uint64 destinationChainSelector
    ) external view returns (uint256 available);

    function getRateLimitConfig(
        address token,
        uint64 destinationChainSelector
    ) external view returns (TokenRateLimitConfig memory);

    function setRateLimitConfig(
        address token,
        uint64 destinationChainSelector,
        TokenRateLimitConfig calldata config
    ) external;
}
```

### Emergency Pause Circuit Interface (`IEmergencyPauseCircuit.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IEmergencyPauseCircuit {
    // Enums
    enum CircuitStatus {
        OPERATIONAL,
        TOKEN_PAUSED,
        CHAIN_PAUSED,
        GLOBAL_EMERGENCY_HALT
    }

    // Events
    event GlobalPauseTriggered(address indexed guardian, string reason);
    event GlobalUnpaused(address indexed admin);
    event TokenCircuitBreakerTriggered(address indexed token, address indexed guardian, string reason);
    event TokenCircuitBreakerReset(address indexed token, address indexed admin);
    event ChainCircuitBreakerTriggered(uint64 indexed chainSelector, address indexed guardian, string reason);
    event ChainCircuitBreakerReset(uint64 indexed chainSelector, address indexed admin);

    // Custom Errors
    error GlobalEmergencyHaltActive();
    error TokenCircuitBreakerActive(address token);
    error ChainCircuitBreakerActive(uint64 chainSelector);
    error CallerNotGuardian(address caller);
    error CallerNotTimelockAdmin(address caller);

    // View Functions
    function isOperationAllowed(address token, uint64 chainSelector) external view returns (bool);
    function isGlobalHalted() external view returns (bool);
    function isTokenHalted(address token) external view returns (bool);
    function isChainHalted(uint64 chainSelector) external view returns (bool);

    // State Mutation Functions
    function triggerGlobalEmergencyHalt(string calldata reason) external;
    function unpauseGlobal() external;
    function triggerTokenCircuitBreaker(address token, string calldata reason) external;
    function resetTokenCircuitBreaker(address token) external;
    function triggerChainCircuitBreaker(uint64 chainSelector, string calldata reason) external;
    function resetChainCircuitBreaker(uint64 chainSelector) external;
}
```

### Synthetic Burnable & Mintable Token Interface (`IBurnableMintableToken.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IBurnableMintableToken is IERC20 {
    // Events
    event SyntheticTokensMinted(address indexed to, uint256 amount, bytes32 indexed crossChainTransferId);
    event SyntheticTokensBurned(address indexed from, uint256 amount, bytes32 indexed crossChainTransferId);

    // Custom Errors
    error CallerNotAuthorizedBridge(address caller);
    error ZeroMintAmount();
    error ZeroBurnAmount();

    // Privileged Functions
    function mint(address to, uint256 amount) external;
    function burn(uint256 amount) external;
    function burnFrom(address account, uint256 amount) external;
    function bridgeContract() external view returns (address);
}
```

## Security & Compliance Notes
- **Dual-Transport Protocol Redundancy:** Relying on a single cross-chain messaging protocol creates a catastrophic single point of failure. The Growww bridge allows dynamic routing across Chainlink CCIP and LayerZero v2, preventing complete network lockout if one protocol suffers downtime or reorganization.
- **Strict Invariant Assertion (Solvency Guarantee):** For lock-and-mint tokens, the contract suite guarantees that on-chain vault reserves never fall below total synthetic supply issued across connected networks ($\text{VaultBalance}(\text{token}) \ge \sum \text{RemoteSyntheticSupply}(\text{token})$). Off-chain reconciliation daemons (Prompt 215) continually assert this mathematical invariant.
- **Sliding-Window Rate Limiting Defense:** Bridge rate limiting limits maximum extractable value during any exploit scenario. Even if a private key or adapter is compromised, cumulative outflow is strictly bounded by deterministic hourly and daily caps.
- **Fail-Safe Inbound Execution Queue:** Inbound messages that encounter execution reverts (e.g. transient receiver contract failure, gas estimation variance, temporary KYC locks) are never silently dropped or burned. They are preserved in the `failedMessages` storage mapping, ensuring complete auditability and manual or automated retry once prerequisite conditions are restored.
- **Transient Storage Reentrancy Locks:** Implements reentrancy protection utilizing EVM transient storage (`TSTORE`/`TLOAD`) in Solidity 0.8.24, minimizing gas consumption while eliminating cross-function and cross-contract reentrancy attack vectors.
- **Dual-Role Pause Controls (Guardian vs Admin):** The `EMERGENCY_GUARDIAN_ROLE` (assigned to automated risk daemons and security operations centers) can immediately halt operations upon anomaly detection with zero latency. Resuming operations requires `TIMELOCK_ADMIN_ROLE` multi-sig threshold approval after a mandatory security review.
- **Zero PII on Chain:** All cross-chain messages contain purely technical parameters (token contract addresses, recipient cryptographic identifiers, amounts, nonces, and chain IDs). No customer names, tax IDs, or personal records are ever transmitted over cross-chain payload data.

## Acceptance Criteria
- [ ] `CrossChainLiquidityBridge.sol`, `CCIPReceiver.sol`, `LayerZeroReceiverAdapter.sol`, `BridgeRateLimiter.sol`, and `EmergencyPauseCircuit.sol` are completely implemented adhering to Solidity 0.8.24 standards.
- [ ] Full support for both Lock-and-Mint and Burn-and-Mint collateral bridging architectures with multi-token support.
- [ ] CCIP adapter correctly constructs `Client.EVM2AnyMessage`, calculates fee overheads in native gas asset or LINK, and validates `Client.Any2EVMMessage` inbound deliveries.
- [ ] LayerZero v2 adapter correctly interfaces with `ILayerZeroEndpointV2` and validates peer authentication.
- [ ] Sliding-window rate limiter accurately enforces `hourlyCapacity`, `dailyCapacity`, and `maxSingleTransfer` caps across independent test simulations.
- [ ] Emergency pause circuits allow instant global, token-specific, and chain-specific halting, and reject bridging attempts with `BridgeCircuitBreakerActive()`.
- [ ] Replay attack protection rejects duplicate `transferId` and GUID submissions across all test scenarios.
- [ ] Inbound message execution failures cleanly capture and store payloads in `failedMessages` without losing investor funds.
- [ ] 100% test pass rate in Foundry covering unit tests, fuzz tests, and formal invariant checks (`vaultBalance >= totalLockedCollateral`).
- [ ] Slither and Aderyn static analyzers pass with zero high or medium security warnings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Platform Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `303` (Token Issuance Smart Contract), Prompt `305` (Transfer Compliance Hooks).
- **Parallel Tasks:** Prompt `307` (Multisig Governance Smart Contract), Prompt `311` (Validator Key Management HSM), Prompt `319` (Institutional Custody Bridge).
- **Subsequent Prompts Enabled:** Prompt `203` (Wallet & Account Ledger Service), Prompt `206` (Pre-Trade Risk & Margin Engine), Prompt `309` (Event Indexing Service), Prompt `310` (Chain Node Monitoring & Alerting).
