# 337 - ERC-4337 Modular Smart Account & Ephemeral Session Keys Contract (Solidity)

## Purpose
In modern institutional and high-frequency retail capital markets, forcing users to manually sign raw private-key transactions for every discrete market action (order submission, amendment, cancellation, settlement escrow lock, and dividend claim) introduces severe friction, unacceptable latency, and catastrophic private key leakage vectors. Furthermore, conventional Externally Owned Accounts (EOAs) cannot enforce granular on-chain access controls, daily spending velocity limits, or court-mandated guardian recovery workflows required by financial regulators such as the Securities and Exchange Board of India (SEBI) and the International Financial Services Centres Authority (IFSCA).

This prompt specifies the architecture, smart contract interfaces, cryptographic verification mechanics, and deployment pipeline for the **ERC-4337 & ERC-7579 Modular Smart Account and Ephemeral Session Keys Suite** (`GrowwwSmartAccount.sol`, `SessionKeyModule.sol`, `GuardianRecoveryModule.sol`, `GrowwwAccountFactory.sol`). Deployed on the permissioned Hyperledger Besu consortium ledger with QBFT consensus, this suite provides institutional and retail investors with high-performance account abstraction. 

Investors benefit from seamless biometric passkey onboarding (secp256r1 / WebAuthn P-256 and secp256k1 ECDSA), gasless transactions via sponsored Paymasters, multi-tiered daily spending velocity caps, scoped ephemeral session keys for 1-click ultra-low latency trading, and multi-signature guardian social recovery with mandatory regulatory timelocks-all while strictly enforcing Growww's zero-PII invariant and Universal Zero-Fee Model (0.00% fee - No fee at all) accounting.

## What You Are Building
A production-grade, modular, gas-optimized Solidity smart contract suite under `contracts/src/account/` and `contracts/src/account/modules/`, conforming strictly to ERC-4337 (EntryPoint v0.7) and ERC-7579 (Minimal Modular Smart Accounts). Concrete deliverables include:
- `GrowwwSmartAccount.sol`: Core modular smart contract wallet implementing `IAccount` (ERC-4337 v0.7) and `IERC7579Account`. Handles execution routing (`execute`, `executeBatch`, `executeFromExecutor`), modular plugin validation hooks, native/token asset management, and direct compatibility with EntryPoint v0.7.
- `SessionKeyModule.sol`: ERC-7579 compliant validator module (`IValidator`) enabling time-bounded, permission-scoped ephemeral session keys. Enforces strict whitelisting of destination contract targets, allowed 4-byte function selectors, cumulative daily spending velocity caps (in sub-paise / INR terms), and single-transaction value thresholds.
- `GuardianRecoveryModule.sol`: ERC-7579 compliant execution/recovery module (`IExecutor`) implementing an $M$-of-$N$ threshold recovery mechanism for institutional custodians and retail guardians. Integrates configurable timelock challenge periods (e.g., 48 hours) to prevent unauthorized takeovers while emitting immutable audit events for SEBI/IFSCA statutory scrutiny.
- `GrowwwAccountFactory.sol`: Deterministic factory contract utilizing `CREATE2` and ERC-1967 minimal proxy clones to deploy user smart accounts at predictable addresses calculated prior to contract deployment (counterfactual deployment during initial user operation).
- Comprehensive Solidity interfaces: `IGrowwwSmartAccount.sol`, `ISessionKeyModule.sol`, `IGuardianRecoveryModule.sol`, `IGrowwwAccountFactory.sol`.
- Comprehensive Foundry test harness (`test/account/`): Unit, fuzz, invariant, and integration test suites validating EntryPoint v0.7 execution, P-256 WebAuthn signature verifications, session key boundary enforcements, velocity cap overflow protections, and guardian recovery state transitions.

## Scope Boundaries
- **In Scope:**
  - ERC-4337 v0.7 `PackedUserOperation` validation, signature decoding, and gas accounting.
  - Dual cryptographic signature schemes: secp256k1 (standard EVM ECDSA) and secp256r1 / P-256 (RIP-7212 precompile or gas-optimized modular verifier for biometric WebAuthn/FIDO2 hardware keys).
  - ERC-7579 modular account architecture: dynamic installation, uninstallation, and execution of validation modules, executor modules, hooks, and fallback handlers.
  - Ephemeral session key lifecycle: key registration, scoping parameters (expiry timestamp, target contracts, method selectors, transaction limit, daily velocity limit), usage tracking, and emergency revocation.
  - Institutional & retail guardian recovery: $M$-of-$N$ threshold multi-sig guardian approval, timelock delay, cancel-by-owner challenge window, and state recovery execution.
  - Reentrancy protection and storage layout isolation across all proxy clones and modular plugins.
  - Zero on-chain Personally Identifiable Information (PII): Account addresses map only to pseudonymous cryptographic hashes verified by `IdentityRegistry.sol` (Prompt 305).
- **Out of Scope / Handled Elsewhere:**
  - ERC-4337 Bundler node and JSON-RPC infrastructure daemon (`eth_sendUserOperation`, handled in Prompt 259).
  - Paymaster sponsorship logic and gas policy credit verification (handled in Prompt 259).
  - Client-side WebAuthn challenge generation, Secure Enclave signature generation, and biometrics UI (handled in Flutter Client, Prompt 505).
  - Off-chain Aadhaar/PAN KYC document verification and verification status updates (handled in Prompts 201 & 202).
  - Secondary order matching engine and order lifecycle routing (handled in Prompts 204 & 205).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD` for reentrancy locks, and custom error specifications).
- **Account Abstraction Standard:** **ERC-4337 v0.7** (`PackedUserOperation`, EntryPoint v0.7 reference implementations).
- **Modular Smart Account Standard:** **ERC-7579** (Minimal Modular Smart Accounts) with standardized module types: `MODULE_TYPE_VALIDATOR = 1`, `MODULE_TYPE_EXECUTOR = 2`, `MODULE_TYPE_FALLBACK = 3`, `MODULE_TYPE_HOOK = 4`.
- **Base Libraries:** OpenZeppelin Contracts v5.0 (`ECDSA`, `SignatureChecker`, `Create2`, `Address`, `Initializable`, `ReentrancyGuardTransient`).
- **P-256 / WebAuthn Verification:** RIP-7212 standard precompile (`0x0100`) with fallback to optimized Daimo/FreshCryptoLib secp256r1 verification contracts for Besu compatibility.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, gas tracking, and invariant fuzz testing; `cast` for RPC calls against Besu devnet).
- **Static Analysis & Formal Verification:** Slither, Mythril, and Solhint.

## Backend / Infra Touchpoints
- **Bundler & Paymaster Service (Prompt 259):** Ingests `PackedUserOperation` payloads from client apps, bundles them into Besu transactions calling `IEntryPoint(entryPoint).handleOps()`, and validates gas sponsorship policies against the off-chain user ledger.
- **Flutter Mobile & Web Client (Prompt 505):** Generates ephemeral session keypairs in hardware Secure Storage/KeyStore, initiates WebAuthn/Passkey registration, signs session scoped operations, and requests counterfactual smart account addresses via `GrowwwAccountFactory.sol`.
- **Identity & KYC Service (Prompt 201 & Prompt 202):** Provides verified `bytes32 identityId` claims to `IdentityRegistry.sol` to allow counterfactually derived smart accounts to hold regulatory asset tokens.
- **DvP Trade Settlement Service (Prompt 208):** Interacts with `GrowwwSmartAccount.sol` as the institutional settlement counterparty for atomic DvP execution.
- **Audit Log Service (Prompt 218) & Blockchain Event Indexer (Prompt 309):** Listens for `SessionKeyRegistered`, `SessionKeyRevoked`, `RecoveryInitiated`, and `RecoveryExecuted` events for real-time compliance reporting and forensic audit trails.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Operates on the private consortium Hyperledger Besu network under QBFT consensus with 2-second deterministic block times and zero chain reorganizations.
- **EntryPoint v0.7 Integration:** All execution flows pass through canonical `EntryPoint` (v0.7). The account verifies that `msg.sender == address(entryPoint)` during `validateUserOp`. Self-execution or module-authorized execution is supported through explicit entrypoints.
- **Nonce Handling:** Utilizes ERC-4337 2-dimensional nonces (`uint192 key`, `uint64 seq`) allowing distinct transaction lanes: Key `0` for sequential financial transfers, Key `100` for ephemeral session key trading operations, and Key `200` for administrative guardian operations, preventing nonce head-of-line blocking.
- **Signature Schemes & Precompiles:** 
  - Mode 0: secp256k1 ECDSA signatures verified via native `ecrecover`.
  - Mode 1: secp256r1 / P-256 signatures verified via RIP-7212 precompile (`0x0000000000000000000000000000000000000100`) or audited polynomial curve verifier fallback.
- **Zero On-Chain PII Invariant:** The contract stores zero investor names, phone numbers, email addresses, or tax IDs. Accounts are identified strictly by their counterfactually generated EVM address and cryptographic public keys.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure & Tooling:** Initialize Foundry repository layout under `contracts/src/account/`, `contracts/src/account/modules/`, `contracts/src/account/interfaces/`, and `test/account/`. Configure `foundry.toml` with `solc_version = "0.8.24"`, EVM version `cancun`, and optimizer runs set to `200`.
2. **Define Data Models & Core Interfaces:** Author `IGrowwwSmartAccount.sol`, `ISessionKeyModule.sol`, `IGuardianRecoveryModule.sol`, and `IGrowwwAccountFactory.sol` incorporating ERC-4337 v0.7 `PackedUserOperation` structs, ERC-7579 module types, error codes, and audit events.
3. **Implement Base Storage & Initialization:** In `GrowwwSmartAccount.sol`, define ERC-7201 namespaced storage slots (`growww.storage.smartaccount`) to prevent storage collisions during upgradeability or modular delegatecalls. Implement `initialize(address initialOwner, bytes calldata initData)` with strict initialization guards.
4. **Implement EntryPoint v0.7 `validateUserOp`:** In `GrowwwSmartAccount.sol`, implement:
   - Verification that `msg.sender == entryPoint()`.
   - Parsing validation mode from signature header or routing to installed `IValidator` modules based on ERC-7579 module flags.
   - Calling `_validateOwnerSignature` or `ISessionKeyModule.validateSessionKeyUserOp`.
   - Pre-funding the EntryPoint for missing gas requirements (`missingAccountFunds`).
   - Returning validation data encoded with `validAfter`, `validUntil`, and validation failure bit flags.
5. **Implement ERC-7579 Execution Engine:** Implement `execute(bytes32 execMode, bytes calldata executionCalldata)` and `executeBatch(...)` allowing single or atomic multicall interactions. Protect direct calls with `onlyEntryPointOrSelf` or `onlyExecutorModule`.
6. **Implement Dynamic Module Management in `GrowwwSmartAccount.sol`:** Implement `installModule(uint256 moduleType, address module, bytes calldata initData)` and `uninstallModule(uint256 moduleType, address module, bytes calldata deInitData)` enforcing access control and preventing duplicate module bindings.
7. **Construct `SessionKeyModule.sol` Validation Engine:** 
   - Maintain mapping `mapping(address => mapping(address => SessionKeyConfig)) public sessionKeys`.
   - Implement `registerSessionKey(address sessionKey, SessionKeyConfig calldata config)` callable only by the smart account owner.
   - Enforce session boundary parameters: `validAfter`, `validUntil`, `dailyVelocityLimitSubPaise`, `singleTxLimitSubPaise`, and allowed contract target and selector bitmaps.
8. **Implement Velocity Limiter & Daily Cap Tracking:**
   - In `SessionKeyModule.sol`, track rolling calendar-day spend: `mapping(address => mapping(address => DailySpendTracker)) public dailySpends`.
   - Verify transaction value and calldata financial value against configured velocity caps.
   - Revert with `VelocityLimitExceeded(uint256 currentSpend, uint256 limit)` if cumulative daily spend exceeds statutory or user thresholds.
9. **Build Emergency Session Revocation:** Implement `revokeSessionKey(address sessionKey)` callable directly by the smart account owner or designated security guardians, instantly invalidating the session key on-chain.
10. **Construct `GuardianRecoveryModule.sol`:**
    - Maintain guardian configuration: `mapping(address => GuardianConfig) public accountGuardians`.
    - Implement `configureGuardians(address[] calldata guardians, uint256 threshold, uint256 timelockSeconds)` callable only by the smart account.
    - Implement `initiateRecovery(address targetAccount, address proposedOwner, bytes[] calldata guardianSignatures)` verifying that at least $M$-of-$N$ valid guardian signatures approve the key rotation.
11. **Implement Recovery Timelock & Challenge Window:**
    - In `GuardianRecoveryModule.sol`, record `RecoveryRequest` with timestamp `executeAfter = block.timestamp + timelockSeconds`.
    - Implement `cancelRecovery(address targetAccount)` enabling the legitimate account owner to cancel fraudulent recovery attempts during the timelock window.
    - Implement `executeRecovery(address targetAccount)` post-timelock to invoke the account's internal `setOwner` function via ERC-7579 execution interface.
12. **Build `GrowwwAccountFactory.sol`:** Implement deterministic `createAccount(address owner, bytes32 salt, bytes calldata initData)` returning the counterfactual smart contract address calculated via `LibClone` / `Create2`. Deploy accounts lazily during the first `UserOperation` if `initCode` is provided.
13. **Integrate Universal Zero-Fee Model (0.00% fee - No fee at all) & Treasury Routing:** Ensure that transaction executions routed through the account for secondary trades or settlement contracts properly preserve and emit accounting references for Growww's canonical 0.00% fee (no fee at all) (0.00% fee at launch; future fee parameters governed by FeeController.sol).
14. **Author Comprehensive Foundry Unit & Fuzz Tests:**
    - Test `validateUserOp` with valid and invalid secp256k1 signatures.
    - Test P-256 WebAuthn signature verification with mock and real WebAuthn authenticator assertion vectors.
    - Fuzz session key daily spend limits across varying transaction amounts, verifying that velocity caps never underflow or breach limits.
    - Test guardian recovery lifecycle: initialization, threshold validation, owner cancellation, timelock expiry, and final owner replacement.
15. **Static Analysis & Formal Verification:** Execute Slither static analysis, ensure zero critical/high warnings, verify transient storage opcodes adhere to Cancun specifications, and check gas benchmarks for Besu QBFT deployment.

## Interfaces / Contracts

### 1. Smart Account Interface (`IGrowwwSmartAccount.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import { PackedUserOperation } from "./PackedUserOperation.sol";

interface IGrowwwSmartAccount {
    // --- Data Types ---
    enum SignatureType {
        ECDSA_SECP256K1,
        PASSKEY_SECP256R1,
        MODULE_VALIDATOR
    }

    struct Execution {
        address target;
        uint256 value;
        bytes callData;
    }

    // --- Events ---
    event SmartAccountInitialized(address indexed entryPoint, address indexed owner);
    event OwnerUpdated(address indexed previousOwner, address indexed newOwner);
    event ModuleInstalled(uint256 indexed moduleType, address indexed module);
    event ModuleUninstalled(uint256 indexed moduleType, address indexed module);
    event Executed(address indexed target, uint256 value, bytes callData);
    event BatchExecuted(uint256 count);

    // --- Custom Errors ---
    error OnlyEntryPoint();
    error OnlyOwner();
    error OnlyEntryPointOrSelf();
    error OnlyAuthorizedModule();
    error InvalidSignature(address recoveredSigner);
    error InvalidSignatureLength(uint256 length);
    error ModuleAlreadyInstalled(uint256 moduleType, address module);
    error ModuleNotInstalled(uint256 moduleType, address module);
    error UnsupportedModuleType(uint256 moduleType);
    error ExecutionFailed(bytes returnData);
    error AccountAlreadyInitialized();

    // --- ERC-4337 EntryPoint Handlers ---
    function validateUserOp(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash,
        uint256 missingAccountFunds
    ) external returns (uint256 validationData);

    // --- Execution Handlers (ERC-7579 Standard) ---
    function execute(bytes32 execMode, bytes calldata executionCalldata) external payable;
    function executeFromExecutor(bytes32 execMode, bytes calldata executionCalldata) external payable returns (bytes[] memory returnData);

    // --- Module Management ---
    function installModule(uint256 moduleType, address module, bytes calldata initData) external;
    function uninstallModule(uint256 moduleType, address module, bytes calldata deInitData) external;
    function isModuleInstalled(uint256 moduleType, address module) external view returns (bool);

    // --- Account State & Getters ---
    function owner() external view returns (address);
    function setOwner(address newOwner) external;
    function entryPoint() external view returns (address);
    function getNonce(uint192 key) external view returns (uint256);
    function isValidSignature(bytes32 hash, bytes calldata signature) external view returns (bytes4);
}
```

### 2. Ephemeral Session Key Module Interface (`ISessionKeyModule.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import { PackedUserOperation } from "./PackedUserOperation.sol";

interface ISessionKeyModule {
    // --- Data Types ---
    struct SessionScope {
        uint48 validAfter;
        uint48 validUntil;
        uint256 singleTxLimitSubPaise;      // Max value per transaction (in sub-paise: 10^-4 INR)
        uint256 dailyVelocityLimitSubPaise;  // Max cumulative value per 24 hours
        address allowedTarget;              // Contract allowed (address(0) for any whitelisted)
        bytes4 allowedSelector;             // Function selector allowed (bytes4(0) for any)
    }

    struct SessionState {
        uint256 currentDayBucket;           // Unix day index (block.timestamp / 86400)
        uint256 dailySpendSubPaise;         // Cumulative spend in current day bucket
        bool isRevoked;
    }

    // --- Events ---
    event SessionKeyRegistered(
        address indexed account,
        address indexed sessionKey,
        uint48 validAfter,
        uint48 validUntil,
        uint256 dailyVelocityLimitSubPaise
    );
    event SessionKeyRevoked(address indexed account, address indexed sessionKey);
    event SessionKeyUsed(
        address indexed account,
        address indexed sessionKey,
        address indexed target,
        bytes4 selector,
        uint256 amountSubPaise
    );

    // --- Errors ---
    error SessionKeyExpired(uint48 validUntil, uint256 currentTimestamp);
    error SessionKeyNotYetValid(uint48 validAfter, uint256 currentTimestamp);
    error SessionKeyRevokedError();
    error SessionKeyNotFound();
    error UnauthorizedTarget(address target);
    error UnauthorizedSelector(bytes4 selector);
    error SingleTxLimitExceeded(uint256 amount, uint256 limit);
    error DailyVelocityLimitExceeded(uint256 currentSpend, uint256 attemptedAmount, uint256 limit);
    error InvalidSessionParameters();

    // --- External Functions ---
    function registerSessionKey(
        address sessionKey,
        SessionScope calldata scope
    ) external;

    function revokeSessionKey(address sessionKey) external;

    function validateSessionUserOp(
        address account,
        PackedUserOperation calldata userOp,
        bytes32 userOpHash
    ) external returns (uint256 validationData);

    function getSessionScope(
        address account,
        address sessionKey
    ) external view returns (SessionScope memory scope, SessionState memory state);
}
```

### 3. Guardian Social Recovery Module Interface (`IGuardianRecoveryModule.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface IGuardianRecoveryModule {
    // --- Data Types ---
    struct RecoveryConfig {
        address[] guardians;
        uint256 threshold;                 // Required guardian signatures (M of N)
        uint48 timelockDurationSeconds;    // Mandatory cooling/challenge window
    }

    struct ActiveRecovery {
        address proposedNewOwner;
        uint48 executionTime;              // block.timestamp + timelockDurationSeconds
        uint256 approvalCount;
        bool isExecuted;
        bool isCancelled;
    }

    // --- Events ---
    event GuardiansConfigured(address indexed account, address[] guardians, uint256 threshold, uint48 timelockSeconds);
    event RecoveryInitiated(
        address indexed account,
        address indexed proposedNewOwner,
        uint48 executionTime,
        uint256 approvals
    );
    event RecoveryApproved(address indexed account, address indexed guardian, address proposedNewOwner);
    event RecoveryCancelled(address indexed account, address indexed cancelledBy);
    event RecoveryExecuted(address indexed account, address indexed previousOwner, address indexed newOwner);

    // --- Errors ---
    error InvalidThreshold(uint256 threshold, uint256 guardianCount);
    error TimelockTooShort(uint48 provided, uint48 minimumRequired);
    error NotAGuardian(address caller);
    error RecoveryAlreadyActive();
    error NoActiveRecovery();
    error TimelockNotExpired(uint48 executionTime, uint256 currentTimestamp);
    error RecoveryCancelledOrExecuted();
    error ThresholdNotMet(uint256 approvals, uint256 threshold);
    error DuplicateGuardianSignature(address guardian);

    // --- External Functions ---
    function configureGuardians(
        address[] calldata guardians,
        uint256 threshold,
        uint48 timelockDurationSeconds
    ) external;

    function initiateRecovery(
        address targetAccount,
        address proposedNewOwner,
        bytes[] calldata guardianSignatures
    ) external;

    function approveRecovery(address targetAccount) external;

    function cancelRecovery(address targetAccount) external;

    function executeRecovery(address targetAccount) external;

    function getRecoveryConfig(address account) external view returns (RecoveryConfig memory);
    function getActiveRecovery(address account) external view returns (ActiveRecovery memory);
}
```

### 4. Deterministic Factory Interface (`IGrowwwAccountFactory.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface IGrowwwAccountFactory {
    event AccountCreated(address indexed account, address indexed owner, bytes32 salt);

    function createAccount(
        address owner,
        bytes32 salt,
        bytes calldata initData
    ) external returns (address account);

    function getAddress(
        address owner,
        bytes32 salt,
        bytes calldata initData
    ) external view returns (address);

    function accountImplementation() external view returns (address);
}
```

## Security & Compliance Notes
- **Reentrancy Protection:** All state-modifying functions that route execution externally (`execute`, `executeBatch`, `executeFromExecutor`) must employ transient-storage reentrancy locks (`TSTORE` / `TLOAD` via OpenZeppelin `ReentrancyGuardTransient`) to prevent cross-function and cross-module reentrancy at zero persistent storage write cost.
- **Strict EntryPoint Verification:** Functions performing user operation validation (`validateUserOp`) must enforce `require(msg.sender == address(entryPoint), OnlyEntryPoint())`. Any deviation opens an immediate unauthorized impersonation vulnerability.
- **Session Expiry & Velocity Bounds:** Ephemeral session keys must enforce absolute timestamp bounds (`validAfter` and `validUntil`). Daily velocity limits must reset based on deterministic 24-hour day buckets (`block.timestamp / 86400`) and sub-paise arithmetic ($10^{-4}$ INR precision) to eliminate integer truncation exploits.
- **SEBI & Regulatory Compliance Audit Trail:** In compliance with SEBI operational guidelines and IFSCA sandbox mandates, all administrative module actions, key rotations, session registrations, and guardian recoveries emit structured events indexed by `account` and `timestamp`. No plaintext PII (names, Aadhaar, PAN) is stored on-chain; only cryptographic identities validated via `IdentityRegistry.sol` (Prompt 305) are referenced.
- **Storage Namespace Safety (ERC-7201):** Modular smart account state and module storage variables must reside in standardized ERC-7201 namespaced storage slots. This prevents storage slot collisions across upgradeable proxies or when executing dynamic delegatecalls.
- **Guardian Recovery Challenge Window:** The guardian recovery module must enforce a minimum immutable timelock period (e.g., 48 hours / 172,800 seconds). The account owner retains absolute authority to invoke `cancelRecovery` at any point during this challenge window, completely neutralizing malicious or compromised guardian collusions.

## Acceptance Criteria
- [ ] Foundry test suite executes cleanly and passes all unit, fuzz, and invariant tests with $\ge 95\%$ code coverage.
- [ ] `GrowwwSmartAccount.sol` strictly conforms to ERC-4337 v0.7 specifications, validating `PackedUserOperation` payloads and returning correct packed `validationData` bitmasks (`authorizer || validUntil || validAfter`).
- [ ] `GrowwwSmartAccount.sol` supports dual-mode signature validation: secp256k1 ECDSA via `ecrecover` and secp256r1 P-256 WebAuthn passkey signatures via RIP-7212 precompile with fallback.
- [ ] `SessionKeyModule.sol` successfully validates scoped user operations, reverting immediately if destination target, function selector, or valid time range is violated.
- [ ] `SessionKeyModule.sol` enforces daily spending velocity caps in sub-paise precision, preventing cumulative daily spends above the configured limit.
- [ ] Owner can instantly revoke an active session key, causing subsequent `validateUserOp` attempts with that session key to revert with `SessionKeyRevokedError`.
- [ ] `GuardianRecoveryModule.sol` successfully initiates $M$-of-$N$ guardian recovery, enforces the full duration of the timelock window, and correctly prevents premature execution.
- [ ] Smart account owner can cancel an in-flight guardian recovery request during the timelock window.
- [ ] Counterfactual deployment via `GrowwwAccountFactory.sol` correctly predicts the exact smart account address and deploys the contract via `CREATE2` during the first `UserOperation`.
- [ ] Slither static analysis runs cleanly with zero high or critical security findings.
- [ ] Gas benchmarks show `validateUserOp` consumes $< 65,000$ gas for secp256k1 validation and $< 110,000$ gas for RIP-7212 P-256 validation.

## Suggested Order / Dependencies
- **Prerequisites:**
  - `301_permissioned_blockchain_evaluation_selection.md` (Hyperledger Besu QBFT network foundation).
  - `305_transfer_compliance_hooks_smart_contract.md` (Identity Registry & KYC/AML compliance checks).
  - `105_authentication_and_authorization_architecture.md` (WebAuthn / Passkey cryptographic architecture).
- **Parallel Tasks:**
  - `259_bundler_and_paymaster_service.md` (ERC-4337 Bundler node and gas sponsorship paymaster daemon).
  - `306_settlement_dvp_smart_contract.md` (Atomic DvP settlement contract interacting with smart accounts).
- **Downstream Blockers:**
  - `505_flutter_authentication_ui.md` (Passkey registration & ephemeral session key generation on client devices).
  - `601_nextjs_investor_web_app_scaffolding.md` (Investor web portal smart account connection).
