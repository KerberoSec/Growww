# 327 - Multi-Chain Proof-of-Reserve Registry Smart Contract

## Purpose
In a regulated multi-asset and cross-chain financial investment ecosystem, proving real-time solvency and 1:1 reserve backing is essential to eliminate fractional reserve risk and maintain regulatory compliance. As Growww expands its institutional tokenization infrastructure across multiple asset classes and settlement layers, reserves are distributed across diverse custody environments: Bitcoin institutional multi-sig cold wallets, Ethereum and Solana lockbox escrow contracts, SEBI-regulated depository participant accounts (NSDL/CDSL) for Indian equities, and scheduled commercial bank escrow accounts for fiat cash reserves.

This prompt specifies the design, architecture, interface specification, testing strategy, and deployment lifecycle of the **Multi-Chain Proof-of-Reserve Registry Smart Contract** (`MultiChainProofOfReserve.sol`). The contract serves as the on-chain canonical source of truth on the permissioned Hyperledger Besu settlement ledger. It ingests, cryptographically validates, and records Sparse Merkle Sum Tree (SMST) root commitments, asset-by-asset custodial reserve attestations, and multi-custodian threshold digital signatures. The contract enforces on-chain mathematical solvency checks (Reserves >= Liabilities), provides public cryptographic inclusion verification for self-sovereign user audits, and triggers emergency circuit breakers if reserve deficits or stale attestation heartbeats are detected.

## What You Are Building
A production-grade Solidity smart contract suite and verification architecture under `contracts/por/` containing:
- `MultiChainProofOfReserve.sol`: Core registry smart contract implementing Sparse Merkle Sum Tree root management, multi-chain reserve tracking, multi-custodian signature verification, and automated solvency assertion.
- `IMultiChainProofOfReserve.sol`: Comprehensive Solidity interface defining data structures, enumerations, events, custom errors, and external/public function signatures.
- `SparseMerkleSumTreeVerifier.sol`: Optimized cryptographic verification library for validating user leaf inclusion proofs and branch sum additions on-chain.
- Multi-Asset Custodial Registry: Data schemas and invariant engines tracking Bitcoin cold wallet UTXO balances, EVM lockbox balances, Solana SPL lockbox tokens, NSDL/CDSL demat share quantities, and bank fiat cash reserves.
- Emergency Solvency Circuit Breaker: Automated pause and quarantine hooks invoked if any individual asset or aggregate reserve ratio falls below the mandatory 100% threshold.
- Comprehensive Foundry Test Suite (`test/por/MultiChainProofOfReserve.t.sol`): Unit tests, fuzz tests, and formal invariant suites verifying signature recovery, Merkle sum validation, multi-asset aggregation, and edge-case reverts.

## Scope Boundaries
- **In Scope:**
  - On-chain storage of periodic multi-chain proof-of-reserve attestations across five asset categories: Bitcoin cold wallets, EVM lockboxes, Solana lockboxes, NSDL/CDSL demat equities, and bank cash reserves.
  - Verification of multi-party EIP-712 structured signatures from authorized custodians, independent auditors, and automated reserve oracles.
  - On-chain Sparse Merkle Sum Tree (SMST) root commitment recording and historical epoch querying.
  - Public on-chain inclusion proof verification (`verifyUserInclusionProof`) for individual account balance validation.
  - Solvency invariant evaluation: ensuring Total Custodial Reserves >= Total On-Chain Token Liabilities across each registered asset identifier.
  - Stale attestation heartbeat detection and automated contract pausing via OpenZeppelin `PausableUpgradeable`.
  - Upgradeability pattern via ERC-1967 Transparent/UUPS proxy standard.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain Bitcoin RPC node indexing and UTXO balance calculation (handled in Prompt 213 and Prompt 319).
  - Off-chain Ethereum and Solana lockbox event indexing (handled in Prompt 309 and Prompt 313).
  - Off-chain Sparse Merkle Sum Tree construction engine (handled in Prompt 308 and Prompt 317).
  - Web-based public Proof-of-Reserve verification user interface (handled in Prompt 606 and Prompt 007).
  - Daily three-way accounting ledger reconciliation engine (handled in Prompt 215).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Provides native checked arithmetic preventing integer overflow/underflow, optimized Yul memory operations, and support for transient storage opcodes.
- **Contract Standard & Security Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `ECDSA`, `EIP712Upgradeable`, `ERC1967Utils`).
- **Development and Testing Framework:** Foundry (`forge` for high-throughput compilation and fuzzing, `cast` for RPC interactions, `anvil` for local testnet simulation).
- **Static Analysis & Formal Verification:** Slither (Trail of Bits), Solhint, and Halmos / Certora for invariant property verification.
- **Consortium Blockchain Target:** Hyperledger Besu permissioned ledger running QBFT consensus.

## Backend / Infra Touchpoints
- **Custodian Depository Integration Service (Prompt 213):** Supplies cryptographically signed demat equity holding statements from NSDL and CDSL.
- **Institutional Custody & Cross-Chain Bridge Service (Prompt 319):** Supplies signed Bitcoin cold wallet UTXO balance proofs and Ethereum/Solana lockbox states.
- **Banking & Payment Gateway Service (Prompt 212):** Transmits signed ISO 20022 cash balance statements from partner scheduled commercial banks.
- **Reconciliation Engine (Prompt 215):** Computes system-wide user liabilities, initiates attestation epochs, and coordinates multi-party signing.
- **Blockchain Event Indexer (Prompt 309):** Ingests `AttestationPublished` and `SolvencyThresholdBreached` events to maintain low-latency query APIs.

## Blockchain Interaction
- **Deployment Model:** Deployed as an ERC-1967 upgradeable proxy on the Hyperledger Besu consortium ledger.
- **Attestation Submission:** At regular scheduled intervals (e.g. hourly for digital assets, market-close for demat equities), authorized relayer accounts submit epoch attestations via `publishMultiChainAttestation(...)`.
- **Cryptographic Signature Verification:** Contract recovers signer addresses from EIP-712 typed data payloads and verifies that the required threshold of institutional custodian and auditor keys have signed the exact epoch root.
- **Solvency Check Execution:** During attestation ingestion, the contract iterates across all asset entries, comparing reported reserves against on-chain token supplies (`DigitalSecurityToken.totalSupply()`). If any asset has Reserves < Liabilities, the transaction reverts or trips the emergency solvency breaker.
- **Zero PII Storage:** All investor leaves in the Sparse Merkle Sum Tree contain blinded identity hashes (`keccak256(userId, userSalt, assetId)`) paired with balance amounts, ensuring complete privacy and compliance with India's Digital Personal Data Protection (DPDP) Act 2023.

## Step-by-Step Build Instructions
1. Scaffold Foundry smart contract directory structure under `contracts/por/`:
   - `src/por/MultiChainProofOfReserve.sol`: Core contract implementation.
   - `src/por/interfaces/IMultiChainProofOfReserve.sol`: Complete interface definition.
   - `src/por/libraries/SparseMerkleSumTreeVerifier.sol`: Cryptographic SMST verification library.
   - `test/por/MultiChainProofOfReserve.t.sol`: Foundry unit, fuzz, and invariant test suite.
   - `script/DeployMultiChainPoR.s.sol`: Deterministic deployment script for Besu QBFT network.
2. Define complete data structures in `IMultiChainProofOfReserve.sol` including `AssetCategory`, `AttestationStatus`, `AssetReserveEntry`, `AttestationEpoch`, and `InclusionProofNode`.
3. Implement `SparseMerkleSumTreeVerifier.sol` library:
   - Implement leaf hash computation: `keccak256(abi.encodePacked(leafId, balance, assetId))`.
   - Implement node hash and sum combination: `parentHash = keccak256(abi.encodePacked(leftHash, leftSum, rightHash, rightSum))` with `parentSum = leftSum + rightSum`.
   - Enforce non-negativity constraint preventing balance underflow attacks.
4. Implement `MultiChainProofOfReserve.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, and `EIP712Upgradeable`.
5. Define constructor disabling initializers and initialize function accepting admin multisig, initial relayer, and initial custodian signers.
6. Implement asset registration and management functions:
   - `registerAsset(string calldata assetId, AssetCategory category, address tokenAddress, uint256 maxHeartbeatInterval)` restricted to `POR_ADMIN_ROLE`.
   - `updateAssetConfig(string calldata assetId, bool isActive, uint256 maxHeartbeatInterval)`.
7. Implement custodian signer threshold management:
   - `setCustodianSigner(address signer, bool isAuthorized)` restricted to admin multisig.
   - `setRequiredSignerThreshold(AssetCategory category, uint8 threshold)`.
8. Implement core attestation ingestion function `publishMultiChainAttestation(...)`:
   - Enforce caller has `ATTESTATION_RELAYER_ROLE`.
   - Verify epoch timestamp is greater than previous epoch and within valid clock drift bounds.
   - Validate EIP-712 multi-party signatures against authorized custodian signers for each asset category.
   - Validate that for each asset entry, reported custodial reserves equal or exceed total on-chain liabilities.
   - Store historical `AttestationEpoch` record in contract state mapping.
   - Emit `AttestationPublished` and `AssetReserveUpdated` events.
9. Implement user self-sovereign balance inclusion verification function `verifyUserInclusionProof(...)`:
   - Accept `epochId`, `leafId`, `balance`, `assetId`, and array of sibling `InclusionProofNode` structs.
   - Recompute Merkle Sum root using `SparseMerkleSumTreeVerifier`.
   - Verify recomputed root matches the committed SMST root for the specified epoch.
   - Emit `UserInclusionVerified` event.
10. Implement emergency circuit breaker mechanisms:
    - `triggerEmergencySolvencyPause(string calldata assetId, string calldata reason)` callable by `EMERGENCY_GUARDIAN_ROLE` or automatically triggered upon deficit detection.
    - `resumeSolvencyOperations()` restricted to multi-sig governance.
11. Write comprehensive Foundry unit tests in `test/por/MultiChainProofOfReserve.t.sol`:
    - Test successful multi-chain attestation publishing across all five asset categories.
    - Test rejection of invalid EIP-712 signatures, expired timestamps, and unauthorized relayers.
    - Test deterministic revert when reserve deficit occurs (Reserves < Liabilities).
    - Test SMST inclusion verification with valid proofs and rejection of tampered balances or corrupted proof nodes.
12. Write property-based fuzz tests in Foundry:
    - Fuzz test `SparseMerkleSumTreeVerifier` across random tree depths (depth 1 to 32) and random balance distributions.
    - Verify sum invariant conservation: root sum equals sum of all leaf balances.
13. Run Slither static analyzer (`slither src/por/MultiChainProofOfReserve.sol`) and resolve all informational, warning, and high-severity findings.
14. Configure Foundry deployment scripts and simulate deployment to local Hyperledger Besu QBFT node.
15. Document contract ABI, EIP-712 type hashes, error signatures, and event schemas for downstream services.

## Interfaces / Contracts

### Multi-Chain Proof of Reserve Interface (`IMultiChainProofOfReserve.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IMultiChainProofOfReserve {
    enum AssetCategory {
        DEMAT_EQUITY,
        BITCOIN_COLD_WALLET,
        EVM_LOCKBOX,
        SOLANA_LOCKBOX,
        BANK_FIAT_ESCROW
    }

    enum AttestationStatus {
        PENDING,
        VERIFIED,
        DISPUTED,
        EXPIRED
    }

    struct AssetConfig {
        string assetId;
        AssetCategory category;
        address tokenAddress;
        uint256 maxHeartbeatInterval;
        bool isActive;
    }

    struct AssetReserveEntry {
        string assetId;
        AssetCategory category;
        uint256 custodialReserveUnits;
        uint256 onChainLiabilityUnits;
        bytes32 custodyStatementHash;
        uint256 lastVerifiedTimestamp;
    }

    struct AttestationEpoch {
        uint256 epochId;
        bytes32 smstMerkleRoot;
        uint256 totalLiabilitySum;
        uint256 timestamp;
        uint256 blockNumber;
        AttestationStatus status;
        bytes32 attestationDigest;
    }

    struct InclusionProofNode {
        bytes32 siblingHash;
        uint256 siblingSum;
        bool isRightSibling;
    }

    struct CustodianSignature {
        address signer;
        bytes signature;
    }

    event AssetRegistered(
        string indexed assetId,
        AssetCategory indexed category,
        address indexed tokenAddress,
        uint256 maxHeartbeatInterval
    );

    event AssetConfigUpdated(
        string indexed assetId,
        bool isActive,
        uint256 maxHeartbeatInterval
    );

    event CustodianSignerStatusUpdated(
        address indexed signer,
        AssetCategory indexed category,
        bool isAuthorized
    );

    event SignerThresholdUpdated(
        AssetCategory indexed category,
        uint8 newThreshold
    );

    event AttestationPublished(
        uint256 indexed epochId,
        bytes32 indexed smstMerkleRoot,
        uint256 totalLiabilitySum,
        uint256 assetCount,
        uint256 timestamp
    );

    event AssetReserveUpdated(
        uint256 indexed epochId,
        string indexed assetId,
        AssetCategory indexed category,
        uint256 custodialReserveUnits,
        uint256 onChainLiabilityUnits
    );

    event UserInclusionVerified(
        uint256 indexed epochId,
        bytes32 indexed leafId,
        string indexed assetId,
        uint256 balance,
        bool isValid
    );

    event SolvencyThresholdBreached(
        string indexed assetId,
        AssetCategory indexed category,
        uint256 custodialReserves,
        uint256 onChainLiabilities,
        uint256 timestamp
    );

    event EmergencySolvencyPaused(
        string indexed assetId,
        string reason,
        address indexed triggeredBy
    );

    event EmergencySolvencyResumed(address indexed triggeredBy);

    error UnauthorizedCaller(address caller, bytes32 requiredRole);
    error InvalidZeroAddress();
    error AssetAlreadyRegistered(string assetId);
    error AssetNotFound(string assetId);
    error AssetNotActive(string assetId);
    error InvalidHeartbeatInterval();
    error InvalidEpochSequence(uint256 providedEpoch, uint256 expectedEpoch);
    error StaleAttestationTimestamp(uint256 timestamp, uint256 currentBlockTimestamp);
    error ClockDriftExceeded(uint256 timestamp, uint256 currentBlockTimestamp);
    error SignerThresholdNotMet(AssetCategory category, uint8 providedSigners, uint8 requiredThreshold);
    error InvalidCustodianSignature(address recoveredSigner, AssetCategory category);
    error DuplicateSignerDetected(address signer);
    error ReserveDeficitDetected(string assetId, uint256 reserves, uint256 liabilities);
    error InvalidSMSTProof(bytes32 calculatedRoot, bytes32 expectedRoot);
    error NegativeBalanceDisallowed(uint256 balance);
    error ArrayLengthMismatch();
    error ContractIsPaused();

    function registerAsset(
        string calldata assetId,
        AssetCategory category,
        address tokenAddress,
        uint256 maxHeartbeatInterval
    ) external;

    function updateAssetConfig(
        string calldata assetId,
        bool isActive,
        uint256 maxHeartbeatInterval
    ) external;

    function setCustodianSigner(
        address signer,
        AssetCategory category,
        bool isAuthorized
    ) external;

    function setSignerThreshold(
        AssetCategory category,
        uint8 threshold
    ) external;

    function publishMultiChainAttestation(
        uint256 epochId,
        bytes32 smstMerkleRoot,
        uint256 totalLiabilitySum,
        uint256 timestamp,
        AssetReserveEntry[] calldata reserveEntries,
        CustodianSignature[] calldata custodianSignatures
    ) external;

    function verifyUserInclusionProof(
        uint256 epochId,
        bytes32 leafId,
        uint256 balance,
        string calldata assetId,
        InclusionProofNode[] calldata proofNodes
    ) external view returns (bool isValid);

    function getLatestEpoch() external view returns (AttestationEpoch memory);

    function getEpochById(uint256 epochId) external view returns (AttestationEpoch memory);

    function getAssetReserveByEpoch(
        uint256 epochId,
        string calldata assetId
    ) external view returns (AssetReserveEntry memory);

    function getAssetConfig(string calldata assetId) external view returns (AssetConfig memory);

    function isSolvent(string calldata assetId) external view returns (bool solvent, uint256 surplusOrDeficit);

    function triggerEmergencySolvencyPause(string calldata assetId, string calldata reason) external;

    function resumeSolvencyOperations() external;
}
```

### Sparse Merkle Sum Tree Verifier Library Interface (`ISparseMerkleSumTreeVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ISparseMerkleSumTreeVerifier {
    struct InclusionProofNode {
        bytes32 siblingHash;
        uint256 siblingSum;
        bool isRightSibling;
    }

    function computeLeafHash(
        bytes32 leafId,
        uint256 balance,
        string calldata assetId
    ) external pure returns (bytes32);

    function verifyProof(
        bytes32 expectedRoot,
        uint256 expectedTotalSum,
        bytes32 leafId,
        uint256 balance,
        string calldata assetId,
        InclusionProofNode[] calldata proof
    ) external pure returns (bool isValid);
}
```

## Security & Compliance Notes
- **Multi-Party Custodian Quorum:** To prevent malicious or compromised single-key attestation publication, attestation ingestion strictly requires a k-of-n threshold signature per asset category. Custodian public keys are managed by multi-sig institutional governance.
- **Mathematical Invariant Enforcement:** The contract programmatically evaluates `custodialReserveUnits >= onChainLiabilityUnits` for every registered asset during attestation execution. If any asset is undercollateralized by even 1 wei, the transaction reverts or halts contract operations.
- **Sparse Merkle Sum Tree Non-Negativity:** Tree summation prevents malicious masking of liabilities by enforcing unsigned 256-bit integer addition without negative node values, ensuring that liability totals cannot be spoofed.
- **EIP-712 Domain Separation:** Attestation digests utilize EIP-712 structured data hashing including `chainId`, contract verifying address, and epoch nonce, preventing cross-chain signature replay or reentrancy attacks.
- **Zero PII & Data Privacy Compliance:** Investor leaf identifiers are one-way cryptographic hashes (`keccak256(userId, salt, assetId)`) containing no personal identifiable information, adhering to India's DPDP Act 2023, SEBI cybersecurity standards, and EU GDPR.
- **Emergency Circuit Breaker:** If an oracle feed or custodian attestation fails to arrive within `maxHeartbeatInterval` or reports a reserve discrepancy, the `EMERGENCY_GUARDIAN_ROLE` can immediately pause settlement operations to protect market integrity.

## Acceptance Criteria
- [ ] `IMultiChainProofOfReserve.sol` and `MultiChainProofOfReserve.sol` compile cleanly with Solidity 0.8.24 and zero warnings.
- [ ] 100% test coverage across Foundry unit, fuzz, and invariant test suites.
- [ ] EIP-712 signature verification strictly validates multi-custodian threshold quorums and rejects invalid or unauthorized signers.
- [ ] `publishMultiChainAttestation` atomically reverts whenever any asset reserve is less than its on-chain liability (`Reserves < Liabilities`).
- [ ] `verifyUserInclusionProof` correctly validates authentic SMST leaf inclusion paths and rejects modified balances, altered sibling nodes, or mismatched asset identifiers.
- [ ] Slither static analysis returns zero high, medium, or reentrancy vulnerabilities.
- [ ] Contract proxy deploys deterministically on Hyperledger Besu testnet and executes gas-optimized attestation publishing.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract), Prompt `307` (MultiSig Governance Smart Contract), Prompt `308` (On-Chain Proof of Reserve Publishing), Prompt `007` (Proof of Reserve Public Disclosure), Prompt `215` (Reconciliation Service).
- **Parallel Tasks:** Prompt `317` (ZK Proof of Solvency Verifier), Prompt `319` (Institutional Custody Bridge), Prompt `606` (Admin Proof of Reserve Reconciliation UI).
- **Subsequent Prompts Enabled:** Prompt `318` (Shareholder Voting Governance Contract), Prompt `233` (Continuous 24x7 Regulatory Reporting).
