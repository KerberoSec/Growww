# 350 - Institutional MPC-TSS Hot/Cold Vault Multi-Sig Coordinator Contract (Solidity)

## Purpose
In institutional digital asset clearing, tokenized equity settlement, and central counterparty operations, the security of cryptographic reserves represents the paramount pillar of financial integrity. Under the dual-entity regulatory structure governing the domestic Growww national market and the offshore International Financial Services Centre (IFSC) GIFT City clearinghouse, assets under custody span tokenized Indian equities, sovereign debt securities (G-Secs), fiat-pegged settlement assets (wholesale/retail eINR), and digital liquidity reserves. 

Concentrating withdrawal authority in standard single-key hot wallets creates catastrophic single-point-of-failure vulnerabilities vulnerable to infrastructure compromise, rogue insider actions, or relayer key compromise. Conversely, purely off-chain Multi-Party Computation Threshold Signature Schemes (MPC-TSS), while eliminating single private keys in host memory, require a deterministic, verifiable, on-chain governance anchor to enforce programmatic policies that no off-chain actor can bypass.

The **Institutional MPC-TSS Hot/Cold Vault Multi-Sig Coordinator Contract (`MpcVaultCoordinator.sol`)** establishes an immutable, stateful, on-chain governance and enforcement layer on Hyperledger Besu. It bridges off-chain MPC-TSS key shares with on-chain settlement operations by enforcing:
1. Cryptographic M-of-N threshold authorization across heterogeneous institutional key shards distributed across physically isolated Hardware Security Modules (HSMs).
2. Mandatory time-delayed execution (timelocks) proportional to transfer value and vault tier, guaranteeing an immutable cooling-off window.
3. Statutory compliance veto powers enabling exchange compliance officers to unilaterally cancel suspicious or anomalous pending sweeps prior to execution.
4. Autonomous 24-hour sliding-window daily velocity limits across cold, warm, and hot vault tiers.
5. On-chain Proof of Reserve (PoR) balance checkpoint validation ensuring that cold-to-hot sweeps never breach statutory collateral solvency margins.

## What You Are Building
A production-grade, modular, upgradeable Solidity smart contract suite under `contracts/src/custody/` conforming to ERC-7201 namespaced storage and OpenZeppelin v5.0 upgradeable standards, comprising:
- `MpcVaultCoordinator.sol`: The master custody coordinator contract managing multi-tier vault registries, M-of-N threshold signature verification over EIP-712 structured payloads, timelocked proposal queues, sliding-window daily velocity counters, reserve checkpoints, and emergency freeze controls.
- `IMpcVaultCoordinator.sol`: The comprehensive Solidity interface declaring all operational structs, lifecycle enums, state transition events, descriptive custom errors, and view/execution function signatures.
- **Three-Tier Vault Hierarchy Management:**
  - **Cold Vault:** Deep offline institutional reserve custody. Subject to strict M-of-N threshold quorums (e.g., 4-of-7 or 3-of-5 institutional shards), mandatory extended timelock delays (e.g., 24 to 48 hours), and Proof of Reserve solvency checks before any balance sweep can execute.
  - **Warm Vault:** Operational clearinghouse settlement buffer. Governed by intermediate timelocks (e.g., 2 to 4 hours) and moderate velocity limits, supplying rebalancing liquidity to hot operational pools.
  - **Hot Vault:** High-throughput settlement relayer pool executing real-time Delivery-versus-Payment (DvP) trades. Governed by immediate or zero-delay execution subject to strict 24-hour aggregate velocity caps and threshold co-signing.
- **EIP-712 Threshold Signature Verifier:** Native Secp256k1 ECDSA verification over domain-separated structured data hashes. Enforces ascending order of signer shard addresses to prevent duplicate signatures and guarantee strict M-of-N consensus across designated institutional custodians.
- **Sliding-Window Daily Velocity Engine:** Stateful accounting tracking cumulative asset outflows within dynamic 24-hour rolling epochs, reverting transactions that breach tier-specific volume thresholds.
- **Compliance Officer Timelock Cancellation Veto:** Dedicated administrative hook permitting compliance officers, emergency guardians, or surveillance sentinels to cancel queued sweep proposals during the timelock latency window.
- **Comprehensive Foundry Test Harness (`test/custody/MpcVaultCoordinator.t.sol`):** Exhaustive unit, property-based fuzzing, and state invariant test suites verifying threshold math, timelock progression, replay resistance, and role-based permissions under adversarial scenarios.

## Scope Boundaries
- **In Scope:**
  - On-chain smart contract implementation of `MpcVaultCoordinator.sol` and interface `IMpcVaultCoordinator.sol`.
  - M-of-N threshold signature verification using EIP-712 typed structured data hashing with ascending shard deduplication.
  - Multi-tier vault registration, tier assignment (Cold, Warm, Hot), and parameter configuration.
  - Stateful timelock lifecycle: Propose -> Queue -> Timelock Delay -> Execute (or Cancel).
  - Unilateral compliance officer cancellation and proposal invalidation during the active timelock window.
  - 24-hour sliding-window cumulative velocity limit tracking and enforcement per vault and asset.
  - Proof of Reserve balance checkpoint verification preventing sweeps that compromise solvency boundaries.
  - Granular single-vault freeze and consortium-wide global emergency freeze controls.
  - Comprehensive event logging for all lifecycle transitions to drive real-time indexing.
  - Complete Foundry unit, invariant fuzzing, and static analysis audit suites.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain MPC-TSS distributed key generation (DKG) protocols, round communication, and CGGMP21/FROST co-signing daemons (handled in Prompt 237).
  - CloudHSM PKCS#11 hardware key share storage, envelope encryption, and local signing engines (handled in Prompt 717).
  - Off-chain physical custody reconciliation with NSDL/CDSL depositories and Merkle tree generation (handled in Prompt 213 and Prompt 215).
  - On-chain Proof of Reserve root publishing registry (handled in Prompt 308 and Prompt 327).
  - Off-chain matching engine partition freezing and order cancellation (handled in Prompt 205 and Prompt 226).
  - Front-end administrative dashboard and compliance portal UI (handled in Prompt 217).
  - Core DvP atomic trade settlement logic (handled in Prompt 306 and Prompt 329).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Cancun with transient storage opcodes `TSTORE`/`TLOAD`, custom user-defined value types, and native custom errors for optimal gas efficiency).
- **Core Frameworks & Standards:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlUpgradeable`, `UUPSUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`).
  - OpenZeppelin Cryptography: `ECDSA`, `MessageHashUtils`, `EIP712Upgradeable`.
  - ERC-7201: Namespaced Storage Layout for deterministic upgrade safety.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, property-based testing, and invariant fuzzing; `cast` for Besu JSON-RPC interaction and diagnostic scripting).
- **Static Analysis & Formal Verification:** Slither, Mythril, and Solhint automated CI checks ensuring zero unhandled external calls, storage collisions, or privilege escalations.
- **Target Network:** Permissioned Hyperledger Besu enterprise ledger operating QBFT consensus with 2-second block intervals and zero chain reorganizations.

## Backend / Infra Touchpoints
- **Institutional Multi-Chain MPC-TSS Vault & Custody Service (Prompt 237):** Distributed microservice running across institutional nodes that coordinates off-chain threshold signing rounds, constructs EIP-712 sweep payloads, aggregates M-of-N shard co-signatures, and broadcasts execution transactions to Besu.
- **CloudHSM Key Lifecycle & Signing Daemon (Prompt 717):** Dedicated hardware daemon interfacing via PKCS#11 with FIPS 140-2/3 Level 3 HSM partitions located in distinct geographic regions (Mumbai, Singapore, Frankfurt) to produce ECDSA Secp256k1 shard signatures.
- **Reconciliation Service & Proof of Reserve (Prompt 215 / Prompt 308):** Ingests physical custodial demat records and bank balances, computes aggregate backing figures, and calls `recordReserveCheckpoint()` to establish on-chain balance watermarks before cold vault sweeps are authorized.
- **Admin Back-Office Service (Prompt 217):** Enterprise web interface used by authorized clearinghouse directors, compliance officers, and risk managers to monitor queued timelock proposals, inspect sliding-window velocity utilization, or execute emergency cancellations.
- **Circuit Breaker & Emergency Halt Contract (Prompt 341):** Cross-contract safety hook enabling `CircuitBreakerHalt.sol` to trigger instant vault freezes across all tiers when market-wide circuit limits or critical depository desync events are detected.
- **Blockchain Event Indexer (Prompt 309):** Real-time Kafka consumer ingesting `SweepProposed`, `SweepExecuted`, `SweepCancelled`, `VaultFrozen`, and `DailyLimitExceeded` events within < 50ms to maintain synchronized state across operational caches.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Settlement Assurances:** Deployed on a private, permissioned Hyperledger Besu ledger utilizing QBFT consensus. Transactions achieve absolute finality upon block inclusion (2-second block time), preventing chain re-organizations from reversing timelocked proposals or sweep executions.
- **EIP-712 Domain Separated Authorization:** All sweep intents are signed as structured data adhering to EIP-712, binding the authorization cryptographically to the specific Besu network `chainId`, the `MpcVaultCoordinator` proxy address, a monotonically increasing per-vault nonce, and an explicit expiration timestamp.
- **Stateful Timelock Lifecycle & Velocity Controls:**
  - Sweeps below the low-value operational threshold execute immediately upon threshold signature verification.
  - Sweeps exceeding the low-value threshold enter a mandatory timelock queue: `eta = block.timestamp + minTimelockSeconds`.
  - The sweep cannot execute until `block.timestamp >= eta`.
  - The sweep must execute before the grace period expires: `block.timestamp <= eta + GRACE_PERIOD`.
  - During the pending window, designated compliance officers holding `COMPLIANCE_OFFICER_ROLE` possess unilateral veto power to call `cancelSweep()`.
- **Reserve Solvency Guard Rails:** Cold vault sweeps require active reserve checkpoints verified within the preceding 24 hours. If a sweep would reduce the total cold reserve below the statutory minimum solvency ratio (e.g., 80% of total collateral backing), the transaction reverts automatically with `InsufficientReserveBalance`.
- **Zero On-Chain PII:** The coordinator records exclusively cryptographic hashes, asset contract addresses, vault identifier hashes (`bytes32`), basis points, and numeric token amounts. Zero customer identifiers, corporate names, or off-chain investor details are ever stored on-chain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project Scaffolding & Dependencies:** Set up Foundry directory structure: `contracts/src/custody/MpcVaultCoordinator.sol`, `contracts/src/interfaces/IMpcVaultCoordinator.sol`, and `test/custody/MpcVaultCoordinator.t.sol`. Configure `foundry.toml` with EVM version `cancun`, optimizer runs `200`, and strict via-IR pipeline.
2. **Author the Formal Interface (`IMpcVaultCoordinator.sol`):** Define all public enums (`VaultTier`, `ProposalStatus`), core data structures (`VaultConfig`, `SweepProposal`, `ReserveCheckpoint`, `KeyShardSignature`), custom errors, and event definitions with indexed query fields.
3. **Establish ERC-7201 Namespaced Storage Layout:** Implement a structured storage struct `MpcVaultCoordinatorStorage` located at a deterministic storage slot derived via `keccak256(abi.encode(uint256(keccak256("growww.storage.MpcVaultCoordinator")) - 1)) & ~bytes32(uint256(0xff))` to eliminate storage collision risks across UUPS upgrade cycles.
4. **Implement Role-Based Access Control (RBAC):** Inherit OpenZeppelin `AccessControlUpgradeable`. Define distinct role identifiers: `DEFAULT_ADMIN_ROLE`, `GOVERNANCE_ROLE`, `COMPLIANCE_OFFICER_ROLE`, `EMERGENCY_GUARDIAN_ROLE`, `RESERVE_ORACLE_ROLE`, and `RELAYER_ROLE`.
5. **Configure EIP-712 Domain Separator & Typehashes:** Inherit `EIP712Upgradeable`. Declare immutable constant typehashes for `SweepIntent`, `EmergencyFreezeIntent`, and `ReserveCheckpointAttestation`.
6. **Implement M-of-N Cryptographic Signature Verification:** Construct an internal verification library function `_verifyThresholdSignatures(bytes32 structHash, KeyShardSignature[] calldata signatures, uint8 requiredThreshold)` that:
   - Verifies each shard address is registered and active in `isKeyShardActive`.
   - Strictly enforces ascending address ordering (`signatures[i].signerAddress < signatures[i+1].signerAddress`) to prevent duplicate shard submissions.
   - Recovers signer addresses via `ECDSA.recover` over the formatted EIP-712 digest.
   - Asserts that valid recovered signatures equal or exceed `requiredThreshold`.
7. **Build Vault Registration & Configuration Routines:** Implement `registerVault(bytes32 vaultId, VaultTier tier, address vaultAddress, uint256 dailyLimit, uint32 minTimelockSeconds)` restricted to `GOVERNANCE_ROLE`. Ensure cold vaults cannot be assigned zero timelock delays.
8. **Build Sliding-Window Daily Velocity Engine:** Implement internal logic `_updateDailySpend(bytes32 vaultId, uint256 amount)`:
   - Evaluate whether `block.timestamp >= lastWindowReset + 1 days`.
   - If true, reset `dailySpent = amount` and update `lastWindowReset = block.timestamp`.
   - If false, assert `dailySpent + amount <= dailyLimit`, reverting with `DailyLimitExceeded(vaultId, attempted, limit)`.
9. **Implement Timelocked Sweep Proposal & Queueing:** Build `proposeSweep(bytes32 vaultId, address tokenAddress, address destination, uint256 amount, bytes32 rationaleHash, KeyShardSignature[] calldata signatures)`:
   - Validate vault is not frozen and destination is whitelisted.
   - Verify threshold signatures against the EIP-712 `SweepIntent` digest.
   - If amount exceeds `instantExecutionThreshold`, transition status to `QUEUED`, calculate `eta = block.timestamp + config.minTimelockSeconds`, record proposal, and emit `SweepQueued`.
   - If amount is within instant limits and within daily velocity, execute immediately, update spend, and emit `SweepExecuted`.
10. **Implement Timelocked Sweep Execution:** Build `executeSweep(uint256 proposalId)`:
    - Assert proposal status is `QUEUED`.
    - Enforce timelock delay: revert with `TimelockNotElapsed` if `block.timestamp < proposal.eta`.
    - Enforce grace period: revert with `GracePeriodElapsed` if `block.timestamp > proposal.eta + GRACE_PERIOD`.
    - Verify vault is not currently frozen.
    - Check reserve balance constraints if source is a Cold Vault.
    - Update daily velocity counters.
    - Execute asset transfer via ERC-20 `safeTransfer` or native Besu transfer.
    - Update status to `EXECUTED` and emit `SweepExecuted`.
11. **Implement Compliance Officer Veto & Proposal Cancellation:** Build `cancelSweep(uint256 proposalId, bytes32 cancellationReasonHash)` restricted to `COMPLIANCE_OFFICER_ROLE` or `EMERGENCY_GUARDIAN_ROLE`:
    - Invalidate queued proposal by setting status to `CANCELLED`.
    - Free any reserved daily velocity quotas.
    - Emit `SweepCancelled(proposalId, msg.sender, cancellationReasonHash)`.
12. **Implement Proof-of-Reserve Balance Checkpoints:** Build `recordReserveCheckpoint(ReserveCheckpoint calldata checkpoint, KeyShardSignature[] calldata oracleSignatures)` restricted to `RESERVE_ORACLE_ROLE` or validated via multi-oracle threshold signatures. Store the latest verified reserve watermark for cold sweep solvency verification.
13. **Implement Multi-Tier Emergency Freeze State Machine:**
    - Single-vault freeze: `freezeVault(bytes32 vaultId, bytes32 reasonHash)` callable by `EMERGENCY_GUARDIAN_ROLE`.
    - Global emergency freeze: `freezeAllVaults(bytes32 reasonHash)` callable by `EMERGENCY_GUARDIAN_ROLE` or directly from `CircuitBreakerHalt.sol`.
    - Unfreeze routines requiring multi-party authorization from `GOVERNANCE_ROLE`.
14. **Develop Comprehensive Foundry Test Suite:** Author `test/custody/MpcVaultCoordinator.t.sol` covering:
    - Correct threshold signature validation and invalid signature rejection.
    - Prevention of duplicate shard signatures via strict ascending sort assertion.
    - Sliding-window daily limit rollover and threshold breach reverts.
    - Timelock delay enforcement and grace period expiration.
    - Compliance officer unilateral cancellation during active timelock.
    - Invariant testing asserting that frozen vaults can never execute sweeps under any condition.
15. **Audit Static Analysis & Formulate Deployment Script:** Run Slither and Solhint with zero warnings. Construct Foundry deployment script `script/DeployMpcVaultCoordinator.s.sol` establishing UUPS proxy, initializing key shard registries, configuring role grants, and wiring initial vault parameters.

## Interfaces / Contracts

### Custody Coordinator Interface (`IMpcVaultCoordinator.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface IMpcVaultCoordinator {
    /// @notice Operational tier of custody vault
    enum VaultTier {
        COLD,   // Deep offline custody, maximum timelock, high quorum
        WARM,   // Intermediate clearing buffer, moderate timelock
        HOT     // Real-time settlement pool, sliding-window velocity caps
    }

    /// @notice Lifecycle status of a sweep proposal
    enum ProposalStatus {
        NONE,
        QUEUED,
        EXECUTED,
        CANCELLED,
        EXPIRED
    }

    /// @notice Vault operational configuration and velocity state
    struct VaultConfig {
        VaultTier tier;
        bool isFrozen;
        address vaultAddress;
        uint256 dailyLimit;
        uint256 dailySpent;
        uint256 lastWindowReset;
        uint32 minTimelockSeconds;
        uint8 requiredThreshold; // M required signers
        uint8 totalShards;       // N total registered shards
    }

    /// @notice Timelocked sweep proposal record
    struct SweepProposal {
        uint256 proposalId;
        bytes32 vaultId;
        address tokenAddress;    // address(0) denotes native Besu token
        address destination;
        uint256 amount;
        uint64 proposedAt;
        uint64 eta;
        uint64 expiresAt;
        ProposalStatus status;
        bytes32 rationaleHash;   // Off-chain audit dossier hash
        address proposer;
    }

    /// @notice Cryptographic attestation from an institutional key shard
    struct KeyShardSignature {
        uint8 shardId;
        address signerAddress;
        bytes signature;         // 65-byte Secp256k1 signature (r, s, v)
    }

    /// @notice Proof-of-Reserve solvency checkpoint
    struct ReserveCheckpoint {
        address asset;
        uint256 totalPhysicalReserve; // Demat/bank balance from custody audit
        uint256 totalOnChainSupply;   // Minted digital token balance
        uint64 checkpointTimestamp;
        bytes32 proofHash;            // Merkle root or IPFS audit report hash
    }

    // --- Events ---
    event VaultRegistered(
        bytes32 indexed vaultId,
        VaultTier indexed tier,
        address indexed vaultAddress,
        uint256 dailyLimit,
        uint32 minTimelockSeconds,
        uint8 requiredThreshold
    );

    event VaultUpdated(
        bytes32 indexed vaultId,
        uint256 newDailyLimit,
        uint32 newMinTimelockSeconds,
        uint8 newThreshold
    );

    event KeyShardRegistered(bytes32 indexed vaultId, uint8 indexed shardId, address indexed shardAddress);
    event KeyShardRevoked(bytes32 indexed vaultId, uint8 indexed shardId, address indexed shardAddress);

    event SweepProposed(
        uint256 indexed proposalId,
        bytes32 indexed vaultId,
        address indexed tokenAddress,
        address destination,
        uint256 amount,
        bytes32 rationaleHash
    );

    event SweepQueued(
        uint256 indexed proposalId,
        bytes32 indexed vaultId,
        uint64 eta,
        uint64 expiresAt
    );

    event SweepExecuted(
        uint256 indexed proposalId,
        bytes32 indexed vaultId,
        address indexed destination,
        uint256 amount,
        address executor
    );

    event SweepCancelled(
        uint256 indexed proposalId,
        bytes32 indexed vaultId,
        address indexed cancelledBy,
        bytes32 reasonHash
    );

    event VaultFrozen(bytes32 indexed vaultId, address indexed initiatedBy, bytes32 reasonHash);
    event VaultUnfrozen(bytes32 indexed vaultId, address indexed initiatedBy);
    event GlobalFreezeTriggered(address indexed initiatedBy, bytes32 reasonHash);
    event GlobalFreezeLifted(address indexed initiatedBy);

    event DailyLimitExceeded(
        bytes32 indexed vaultId,
        uint256 attemptedAmount,
        uint256 dailySpent,
        uint256 dailyLimit
    );

    event ReserveCheckpointRecorded(
        address indexed asset,
        uint256 totalPhysicalReserve,
        uint256 totalOnChainSupply,
        uint64 indexed timestamp,
        bytes32 proofHash
    );

    // --- Custom Errors ---
    error VaultAlreadyExists(bytes32 vaultId);
    error VaultNotFound(bytes32 vaultId);
    error VaultIsFrozen(bytes32 vaultId);
    error GlobalEmergencyFreezeActive();
    error InvalidThresholdConfiguration(uint8 threshold, uint8 totalShards);
    error ShardAlreadyRegistered(bytes32 vaultId, uint8 shardId);
    error ShardNotActive(bytes32 vaultId, address signer);
    error InvalidSignerOrder(address current, address previous);
    error InsufficientValidSignatures(uint8 valid, uint8 required);
    error DailyLimitExceededError(bytes32 vaultId, uint256 requested, uint256 remaining);
    error TimelockNotElapsed(uint256 proposalId, uint64 currentTimestamp, uint64 eta);
    error GracePeriodElapsed(uint256 proposalId, uint64 currentTimestamp, uint64 expiresAt);
    error ProposalNotQueued(uint256 proposalId, ProposalStatus currentStatus);
    error UnauthorizedCaller(address caller, bytes32 requiredRole);
    error InvalidDestinationAddress(address destination);
    error InsufficientReserveBalance(address asset, uint256 required, uint256 available);
    error ProposalExpired(uint256 proposalId);
    error DestinationNotWhitelisted(address destination);
    error ArrayLengthMismatch();

    // --- Core External Functions ---
    function proposeSweep(
        bytes32 vaultId,
        address tokenAddress,
        address destination,
        uint256 amount,
        bytes32 rationaleHash,
        KeyShardSignature[] calldata signatures
    ) external returns (uint256 proposalId);

    function executeSweep(uint256 proposalId) external;

    function cancelSweep(uint256 proposalId, bytes32 cancellationReasonHash) external;

    function recordReserveCheckpoint(
        ReserveCheckpoint calldata checkpoint,
        KeyShardSignature[] calldata oracleSignatures
    ) external;

    function freezeVault(bytes32 vaultId, bytes32 reasonHash) external;
    function unfreezeVault(bytes32 vaultId) external;
    function freezeAllVaults(bytes32 reasonHash) external;
    function liftGlobalFreeze() external;

    function registerVault(
        bytes32 vaultId,
        VaultTier tier,
        address vaultAddress,
        uint256 dailyLimit,
        uint32 minTimelockSeconds,
        uint8 requiredThreshold
    ) external;

    function configureKeyShard(
        bytes32 vaultId,
        uint8 shardId,
        address shardAddress,
        bool active
    ) external;

    // --- View Inspection Methods ---
    function getVaultConfig(bytes32 vaultId) external view returns (VaultConfig memory);
    function getProposal(uint256 proposalId) external view returns (SweepProposal memory);
    function getRemainingDailyLimit(bytes32 vaultId) external view returns (uint256);
    function isVaultFrozen(bytes32 vaultId) external view returns (bool);
    function isGlobalFrozen() external view returns (bool);
    function getLatestReserveCheckpoint(address asset) external view returns (ReserveCheckpoint memory);
}
```

### Integration Pattern (`MpcVaultCoordinator` Execution Flow Snippet)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { IMpcVaultCoordinator } from "../interfaces/IMpcVaultCoordinator.sol";

abstract contract MpcVaultCoordinatorBase is IMpcVaultCoordinator {
    using SafeERC20 for IERC20;

    bytes32 public constant SWEEP_INTENT_TYPEHASH = keccak256(
        "SweepIntent(bytes32 vaultId,address tokenAddress,address destination,uint256 amount,uint256 nonce,bytes32 rationaleHash,uint256 deadline)"
    );

    uint64 public constant GRACE_PERIOD = 3 days;

    function _verifySignatures(
        bytes32 digest,
        bytes32 vaultId,
        KeyShardSignature[] calldata signatures,
        uint8 requiredThreshold
    ) internal view {
        if (signatures.length < requiredThreshold) {
            revert InsufficientValidSignatures(uint8(signatures.length), requiredThreshold);
        }

        address lastSigner = address(0);
        uint8 validSignatures = 0;

        for (uint256 i = 0; i < signatures.length; i = _uncheckedInc(i)) {
            address currentSigner = signatures[i].signerAddress;

            // Enforce strictly ascending signer addresses to guarantee uniqueness
            if (currentSigner <= lastSigner) {
                revert InvalidSignerOrder(currentSigner, lastSigner);
            }

            // Verify shard is registered and active for the specified vault
            if (!_isShardActive(vaultId, signatures[i].shardId, currentSigner)) {
                revert ShardNotActive(vaultId, currentSigner);
            }

            // Recover and validate signer
            address recovered = ECDSA.recover(digest, signatures[i].signature);
            if (recovered != currentSigner) {
                revert ShardNotActive(vaultId, recovered);
            }

            lastSigner = currentSigner;
            validSignatures++;
        }

        if (validSignatures < requiredThreshold) {
            revert InsufficientValidSignatures(validSignatures, requiredThreshold);
        }
    }

    function _uncheckedInc(uint256 i) internal pure returns (uint256) {
        unchecked { return i + 1; }
    }

    function _isShardActive(bytes32 vaultId, uint8 shardId, address signer) internal view virtual returns (bool);
}
```

## Security & Compliance Notes
- **Timelock Cancellation Window by Compliance Officers (Statutory Veto):**
  - All sweep proposals exceeding operational thresholds enter a mandatory timelock delay (minimum 48 hours for Cold Vaults, 2 hours for Warm Vaults).
  - During this latency window, exchange compliance officers holding `COMPLIANCE_OFFICER_ROLE` or risk trustees holding `EMERGENCY_GUARDIAN_ROLE` possess unambiguous unilateral power to cancel the proposal on-chain via `cancelSweep()`.
  - Invalidation is instantaneous and irreversible, preventing unauthorized transfers even if an attacker compromises the requisite M-of-N private key shares off-chain.
- **Strict M-of-N Cryptographic Threshold Enforcement & Shard Deduplication:**
  - Every proposal payload requires explicit ECDSA signatures over an EIP-712 structured digest binding network chain ID, contract address, monotonically increasing nonce, and transfer parameters.
  - Signatures must be submitted in strictly ascending order of signer address (`signatures[i].signerAddress < signatures[i+1].signerAddress`). This programmatic check guarantees that duplicate signatures from the same key shard cannot be re-used to satisfy threshold quorums.
- **Geographically Distributed Key Share Defense in Depth:**
  - Key shards corresponding to registered signers reside inside FIPS 140-2/3 Level 3 Hardware Security Modules across independent physical cloud and bare-metal regions (e.g., Shard 1: AWS CloudHSM Mumbai, Shard 2: Azure Dedicated HSM Central India, Shard 3: AWS CloudHSM Singapore, Shard 4: On-Premises HSM GIFT City, Shard 5: Offline Disaster Recovery Vault).
  - No single region, cloud provider, or network boundary failure can produce a valid M-of-N threshold signature on-chain.
- **Sliding-Window Daily Velocity Limits (Blast Radius Mitigation):**
  - Every vault is configured with a strict 24-hour rolling velocity limit (`dailyLimit`).
  - Attempted transfers that exceed the unspent balance within the active 24-hour window revert immediately with `DailyLimitExceededError`.
  - This architecture limits the maximum aggregate exposure of the clearinghouse to pre-funded, bounded capital allocations even in the catastrophic event of a compromised threshold quorum.
- **Proof-of-Reserve Solvency Guard Rails:**
  - The contract maintains verified physical custody balance checkpoints ingested from the Reconciliation Service (Prompt 215) and Proof of Reserve publisher (Prompt 308).
  - Cold vault sweeps evaluate the post-transfer solvency ratio. If a sweep would cause digital token issuance to exceed physical custody reserves, the transaction reverts deterministically with `InsufficientReserveBalance`.
- **Zero On-Chain Personally Identifiable Information (PII):**
  - All sweep requests record strictly operational financial data: `bytes32 vaultId`, token contract addresses, `uint256 amount`, and a `bytes32 rationaleHash` referencing off-chain encrypted audit packages. No investor identity data is ever committed to the ledger.

## Acceptance Criteria
- [ ] `MpcVaultCoordinator.sol` and `IMpcVaultCoordinator.sol` compile successfully with Solidity ^0.8.24 using Foundry with zero warnings or deprecations.
- [ ] Pass all Slither static analysis and Solhint linting suites with zero high, medium, or informational findings regarding reentrancy, access control, or storage alignment.
- [ ] M-of-N threshold signature verification rigorously enforced: valid signatures from M distinct active shards succeed; M-1 signatures or signatures with duplicate signers revert with `InsufficientValidSignatures` or `InvalidSignerOrder`.
- [ ] Cold vault sweeps exceeding operational threshold enter the `QUEUED` state with `eta = block.timestamp + minTimelockSeconds` and cannot be executed prior to `eta`.
- [ ] Calling `executeSweep()` before `eta` reverts with `TimelockNotElapsed`. Calling after `eta + GRACE_PERIOD` reverts with `GracePeriodElapsed`.
- [ ] Authorized compliance officers can unilaterally cancel pending proposals during the timelock window via `cancelSweep()`, transitioning status to `CANCELLED` and preventing execution.
- [ ] 24-hour sliding-window daily limit properly tracks cumulative outflows and resets after 24 hours. Attempting to sweep beyond the daily limit reverts with `DailyLimitExceededError`.
- [ ] Single-vault freeze (`freezeVault`) and global freeze (`freezeAllVaults`) immediately prevent all sweep proposals and executions across affected scopes.
- [ ] Cold vault sweeps revert with `InsufficientReserveBalance` if the transfer would breach statutory reserve solvency margins based on recorded checkpoints.
- [ ] Foundry test suite achieves $\ge 95\%$ line and branch test coverage, including invariant fuzz tests asserting that frozen vaults never execute transfers under randomized state fuzzing.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301` (Permissioned Blockchain Evaluation & Selection - Besu Architecture)
  - Prompt `302` (Network Topology and Validator Setup)
  - Prompt `303` (Token Issuance Smart Contract)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
  - Prompt `341` (On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract)
- **Parallel Tasks:**
  - Prompt `237` (Institutional Multi-Chain MPC-TSS Vault & Custody Service)
  - Prompt `717` (HSM / KMS CloudHSM Key Lifecycle & Signing Daemon)
  - Prompt `215` (Reconciliation Service & Automated Depository Audit)
  - Prompt `308` (On-Chain Proof of Reserve Publishing)
  - Prompt `217` (Admin Back-Office Service - Custody & Compliance Portal)
- **Subsequent Prompts Enabled:**
  - Prompt `315` (Settlement Guarantee Fund Contract)
  - Prompt `319` (Institutional Custody Bridge)
  - Prompt `327` (Multichain Proof of Reserve Registry Contract)
  - Prompt `329` (NBSE Delivery-versus-Payment DvP Settlement & Automated Fee Collector)
