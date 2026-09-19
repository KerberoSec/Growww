# 330 - MCX Commodity Token and WDRA Physical Vault Registry Smart Contracts

## Purpose
In the Indian financial regulatory ecosystem governed by SEBI and WDRA (Warehousing Development and Regulatory Authority), commodity markets rely on accredited vaulting and warehousing infrastructure (such as CCRL and NERL electronic Negotiable Warehouse Receipts - e-NWRs) to secure underlying physical assets. Expanding the Growww institutional 24/7 Web3 exchange into tokenized physical commodities requires an immutable, verifiable, and legally compliant on-chain settlement layer on Hyperledger Besu.

This prompt specifies the design, architecture, interface definitions, testing strategy, and deployment lifecycle of the **MCX Commodity Security Token** (`CommoditySecurityToken.sol`), the **Physical Vault Registry** (`PhysicalVaultRegistry.sol`), and the **Commodity Delivery Burner** (`CommodityDeliveryBurner.sol`). These contracts implement the ERC-3643 (T-REX) standard for tokenized commodities (such as `gGOLD` representing 1 gram of 999 fineness LBMA/BIS-hallmarked gold, and `gSILVER` representing 10 grams of 999 purity silver). Every token is backed 1:1 by physical inventory in WDRA-accredited vaults, cryptographically bound to NABL-accredited refinery assay certificates and e-NWR identifiers, supports continuous Proof-of-Reserve Sparse Merkle verification, and executes strict on-chain burn mechanics upon physical delivery redemption.

## What You Are Building
A production-grade Solidity smart contract suite and verification architecture under `contracts/commodities/` containing:
- `ICommoditySecurityToken.sol`: ERC-3643 compliant security token interface for physical commodities with identity registry compliance checks, partition tracking (allocated vs in-transit delivery), and fractional unit precision (18 decimals).
- `IPhysicalVaultRegistry.sol`: Canonical on-chain registry interface tracking WDRA-accredited vault facilities, vault custodian operators, bar/ingot serial numbers, NABL assay certificate hashes, e-NWR receipt numbers, weight, purity, and custodial audit timestamps.
- `ICommodityDeliveryBurner.sol`: Physical redemption coordinator interface orchestrating delivery requests, token escrow locks, vault dispatch confirmations, physical receipt validations, and final supply burning.
- `ICommodityPoRVerifier.sol`: Merkle tree verification interface validating physical vault inventory leaf commitments against total on-chain circulating token supplies.
- Comprehensive Foundry Test Suite (`test/commodities/CommoditySecurityToken.t.sol`, `test/commodities/PhysicalVaultRegistry.t.sol`): Unit tests, fuzz tests, and formal invariant suites validating 1:1 reserve invariants, assay hashing, delivery burn state transitions, and unauthorized mint/burn prevention.

## Scope Boundaries
- **In Scope:**
  - ERC-3643 security token interface for tokenized commodities (`gGOLD`, `gSILVER`, base metals, agricultural commodities).
  - WDRA-accredited physical vault registry tracking warehouse receipts (e-NWR), vault operators, assay certificate IPFS/Arweave digests, bar serial numbers, purity ratings, and allocated bar weights.
  - Multi-signature / HSM-authorized physical deposit ingestion linking verified warehouse credits to token minting.
  - End-to-end physical delivery redemption lifecycle: delivery request initiation, token lockup escrow, vault dispatch verification, proof of physical handoff, and on-chain token burn.
  - Proof of Reserve Merkle verification matching total active vaulted weight against on-chain token `totalSupply()`.
  - Transfer compliance enforcement via `IIdentityRegistry` (ERC-3643 KYC/AML/Sanctions verification on all holders and redemption claimants).
  - Emergency lot quarantine, vault facility suspension, and pausable circuit breakers.
  - Deterministic upgradeability via ERC-1967 Transparent / UUPS proxy pattern.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain WDRA, CCRL, and NERL e-NWR depository API adapters (handled in Prompt 213 and Prompt 242).
  - Off-chain logistics and armored transport (Brink's, Sequel Logistics) tracking services.
  - MCX real-time spot and futures price oracle feeds (handled in Prompt 328 Oracle Aggregator).
  - Physical delivery GST, custom duty, and stamp duty taxation calculation engine (handled in Prompt 223).
  - Web and mobile user interface for physical vault redemption booking (handled in Category 6 and 7).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Provides native checked arithmetic preventing overflow/underflow, user-defined value types for asset quantities, optimized Yul memory operations, and support for transient storage.
- **Contract Standards & Security Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `ECDSA`, `EIP712Upgradeable`, `ERC1967Utils`) and ERC-3643 (T-REX) Compliance interfaces.
- **Development & Testing Framework:** Foundry (`forge` for fast compilation and fuzzing, `cast` for RPC calls, `anvil` for local node simulation).
- **Static Analysis & Formal Verification:** Slither (Trail of Bits), Solhint, and Halmos / Certora for mathematical invariant proofs.
- **Consortium Blockchain Target:** Hyperledger Besu permissioned enterprise ledger running QBFT consensus.

## Backend / Infra Touchpoints
- **Custodian & Depository Integration Service (Prompt 213):** Ingests WDRA e-NWR warehouse credits and refinery assay reports to construct signed multi-sig mint proposals.
- **Reconciliation Engine (Prompt 215):** Validates three-way consistency between on-chain token supplies, vault registry records, and physical depository holding logs.
- **Proof-of-Reserve Publishing Service (Prompt 308 & Prompt 327):** Submits periodic Sparse Merkle Tree roots of allocated physical inventory to the on-chain registry.
- **Blockchain Event Indexer (Prompt 309):** Ingests `VaultDepositRegistered`, `DeliveryRequested`, `DeliveryDispatched`, and `TokensBurnedForDelivery` events for downstream accounting and mobile push updates.
- **Double-Entry Wallet Account Ledger Service (Prompt 203):** Mirrors on-chain token balance changes to internal off-chain fiat/commodity ledgers.

## Blockchain Interaction
- **Deployment Model:** Deployed as ERC-1967 upgradeable proxies on the Hyperledger Besu enterprise consortium network.
- **1:1 Custodial Invariant:** Tokens can only be minted when a physical commodity bar/lot is registered in `PhysicalVaultRegistry.sol` with status `VAULTED` and signed by authorized WDRA vault custodians. `totalSupply()` must never exceed total verified vaulted weight.
- **Assay Certificate & Bar Serial Hashing:** Every vaulted bar record contains a deterministic cryptographic hash (`keccak256(abi.encodePacked(vaultId, barSerialNumber, refineryCode, purity, grossWeightGrams, assayCertDigest))`) stored immutably on-chain.
- **Physical Delivery Burn Mechanics:** When an investor requests physical delivery of vaulted metal, the corresponding token amount is moved to an escrow partition in `CommodityDeliveryBurner.sol`. Upon confirmed physical handoff and vault gate pass generation, the escrowed tokens are burned permanently (`_burn`), decrementing `totalSupply()` and updating the vault lot status to `DELIVERED`.
- **Zero PII Ledger Invariant:** Delivery requests reference off-chain shipping/vault collection orders via blinded cryptographic hashes (`deliveryCommitmentHash`). No investor legal names, PANs, delivery addresses, or phone numbers exist on-chain.

## Step-by-Step Build Instructions
1. Scaffold Foundry smart contract directory structure under `contracts/commodities/`:
   - `src/commodities/interfaces/ICommoditySecurityToken.sol`
   - `src/commodities/interfaces/IPhysicalVaultRegistry.sol`
   - `src/commodities/interfaces/ICommodityDeliveryBurner.sol`
   - `src/commodities/interfaces/ICommodityPoRVerifier.sol`
   - `src/commodities/CommoditySecurityToken.sol`
   - `src/commodities/PhysicalVaultRegistry.sol`
   - `src/commodities/CommodityDeliveryBurner.sol`
   - `test/commodities/CommoditySecurityToken.t.sol`
   - `test/commodities/PhysicalVaultRegistry.t.sol`
   - `script/DeployCommodityContracts.s.sol`
2. Define complete enumerations, data structures, custom errors, and events in `IPhysicalVaultRegistry.sol` covering vault operators, commodity asset codes, bar/lot metadata, assay certificates, and e-NWR IDs.
3. Define complete interfaces for `ICommoditySecurityToken.sol` combining ERC-3643 compliance hooks, partition tracking, minting with vault lot binding, and physical delivery burning.
4. Define complete interfaces for `ICommodityDeliveryBurner.sol` covering delivery request creation, vault operator dispatch authorization, delivery confirmation, cancellation timeouts, and escrow handling.
5. Define Merkle proof verification methods in `ICommodityPoRVerifier.sol` for verifying inclusion of individual vault bar lots in the published Proof-of-Reserve root.
6. Implement `PhysicalVaultRegistry.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, and `ReentrancyGuardUpgradeable`:
   - Implement vault facility onboarding (`registerVaultFacility`) restricted to `VAULT_ADMIN_ROLE`.
   - Implement commodity lot registration (`depositCommodityLot`) requiring dual-signature attestation from accredited vault custodian and NABL refinery assayer.
   - Implement lot status transitions (`lockLotForDelivery`, `markLotDelivered`, `quarantineLot`).
   - Implement view functions for querying lot details, total active vaulted weight, and vault audit history.
7. Implement `CommoditySecurityToken.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, and `ICommoditySecurityToken`:
   - Enforce ERC-3643 transfer restrictions by querying `IIdentityRegistry` before every mint, transfer, and burn.
   - Implement `mintWithVaultLot(address to, uint256 amount, bytes32 lotHash)` restricted to `MINTER_ROLE` (granted exclusively to the governance multi-sig after vault lot confirmation).
   - Implement `burnForDelivery(address from, uint256 amount, bytes32 deliveryId)` restricted to `DELIVERY_BURNER_ROLE`.
   - Implement address-level freezing (`freezeAddress`, `unfreezeAddress`) and recovery features mandated by regulatory authorities.
8. Implement `CommodityDeliveryBurner.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, and `EIP712Upgradeable`:
   - Implement `requestDelivery(bytes32 commodityId, uint256 lotCount, bytes32 deliveryCommitmentHash)` transferring tokens from user to contract escrow.
   - Implement `authorizeDispatch(bytes32 deliveryId, bytes32[] calldata lotHashes)` restricted to authorized vault custodian operators.
   - Implement `confirmDeliveryAndBurn(bytes32 deliveryId, bytes calldata receiptProof)` executing `burnForDelivery` on the token contract and updating lot statuses to `DELIVERED`.
   - Implement `cancelDelivery(bytes32 deliveryId)` with expiration cooldown and penalty deduction if the claimant fails to collect within the designated pickup window.
9. Implement Proof-of-Reserve validation logic connecting `PhysicalVaultRegistry.sol` with `ICommodityPoRVerifier.sol` to ensure total vaulted commodity grams >= on-chain token supply.
10. Write comprehensive unit tests in `test/commodities/CommoditySecurityToken.t.sol`:
    - Test token initialization, symbol/name validation (`gGOLD`, `gSILVER`), and 18-decimal precision.
    - Test compliant vs non-compliant recipient minting and transfers.
    - Test unauthorized mint and burn attempts reverting with custom errors.
11. Write comprehensive unit tests in `test/commodities/PhysicalVaultRegistry.t.sol`:
    - Test vault facility registration and status updates.
    - Test lot deposit with valid vs invalid assay certificate hashes and duplicate bar serial numbers.
    - Test end-to-end delivery lifecycle: request -> lock -> dispatch -> confirm -> burn.
    - Test lot quarantine and emergency facility pausing.
12. Write Foundry property-based fuzz tests:
    - Fuzz test lot weight aggregations ensuring no arithmetic overflow across millions of fractional lots.
    - Fuzz test Proof-of-Reserve Merkle inclusion proofs across random tree depths (depth 1 to 32).
    - Invariant test: `token.totalSupply() == vaultRegistry.getTotalActiveVaultedWeight(commodityId)`.
13. Run Slither static analyzer (`slither src/commodities/`) and ensure zero high, medium, or reentrancy vulnerabilities.
14. Configure Foundry deployment script (`DeployCommodityContracts.s.sol`) with deterministic salt and initialize proxies on local Besu testnet.
15. Document contract ABIs, EIP-712 domain separators, method signatures, events, and custom errors for downstream integration.

## Interfaces / Contracts

### Commodity Security Token Interface (`ICommoditySecurityToken.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICommoditySecurityToken {
    enum CommodityType {
        GOLD_999,
        SILVER_999,
        PLATINUM_995,
        CRUDE_OIL,
        AGRICULTURAL
    }

    struct CommodityMetadata {
        bytes32 commodityId;
        string standardSymbol;
        CommodityType commodityType;
        uint8 unitPurity;
        uint256 gramsPerTokenUnit;
        address vaultRegistry;
        address identityRegistry;
        address complianceRegistry;
    }

    event CommodityTokenMinted(
        address indexed to,
        uint256 amount,
        bytes32 indexed lotHash,
        bytes32 indexed eNwrId
    );

    event CommodityTokenBurnedForDelivery(
        address indexed from,
        uint256 amount,
        bytes32 indexed deliveryId,
        bytes32 indexed lotHash
    );

    event AddressFrozen(
        address indexed target,
        bool isFrozen,
        address indexed operator
    );

    event IdentityRegistryUpdated(
        address indexed oldRegistry,
        address indexed newRegistry
    );

    event ComplianceRegistryUpdated(
        address indexed oldCompliance,
        address indexed newCompliance
    );

    event VaultRegistryUpdated(
        address indexed oldRegistry,
        address indexed newRegistry
    );

    error CallerNotMinter(address caller);
    error CallerNotBurner(address caller);
    error RecipientNotVerified(address recipient);
    error SenderNotVerified(address sender);
    error TransferRestrictedByCompliance(uint256 reasonCode);
    error AddressIsFrozen(address account);
    error InvalidZeroAddress();
    error InvalidZeroAmount();
    error VaultLotAlreadyMinted(bytes32 lotHash);
    error VaultLotNotActive(bytes32 lotHash);
    error SupplyExceedsVaultedReserves(uint256 requestedSupply, uint256 availableReserves);

    function mintWithVaultLot(
        address to,
        uint256 amount,
        bytes32 lotHash,
        bytes32 eNwrId
    ) external;

    function batchMintWithVaultLots(
        address[] calldata recipients,
        uint256[] calldata amounts,
        bytes32[] calldata lotHashes,
        bytes32[] calldata eNwrIds
    ) external;

    function burnForDelivery(
        address from,
        uint256 amount,
        bytes32 deliveryId,
        bytes32 lotHash
    ) external;

    function freezeAddress(address target) external;

    function unfreezeAddress(address target) external;

    function isAddressFrozen(address target) external view returns (bool);

    function getCommodityMetadata() external view returns (CommodityMetadata memory);

    function getVaultRegistry() external view returns (address);

    function getIdentityRegistry() external view returns (address);

    function isCompliant(address account) external view returns (bool);
}
```

### Physical Vault Registry Interface (`IPhysicalVaultRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IPhysicalVaultRegistry {
    enum VaultStatus {
        ACTIVE,
        SUSPENDED,
        DECOMMISSIONED
    }

    enum LotStatus {
        VAULTED,
        LOCKED_FOR_DELIVERY,
        DELIVERED,
        QUARANTINED,
        AUDIT_DISPUTED
    }

    struct VaultFacility {
        bytes32 vaultId;
        string wdraAccreditationNumber;
        string operatorName;
        string locationCity;
        string locationState;
        VaultStatus status;
        uint256 maxStorageCapacityGrams;
        uint256 currentStoredGrams;
        uint256 lastAuditTimestamp;
        address primarySigner;
    }

    struct CommodityLot {
        bytes32 lotHash;
        bytes32 commodityId;
        bytes32 vaultId;
        string barSerialNumber;
        string refineryCode;
        uint16 purityPermille;
        uint256 grossWeightGrams;
        uint256 netPureWeightGrams;
        bytes32 assayCertificateDigest;
        bytes32 eNwrReceiptId;
        LotStatus status;
        uint256 depositTimestamp;
        uint256 lastVerifiedTimestamp;
    }

    struct CustodianAttestation {
        bytes32 lotHash;
        bytes32 vaultId;
        bytes32 assayDigest;
        uint256 timestamp;
        bytes custodianSignature;
        bytes assayerSignature;
    }

    event VaultFacilityRegistered(
        bytes32 indexed vaultId,
        string wdraAccreditationNumber,
        string operatorName,
        address indexed primarySigner
    );

    event VaultFacilityStatusUpdated(
        bytes32 indexed vaultId,
        VaultStatus oldStatus,
        VaultStatus newStatus
    );

    event CommodityLotDeposited(
        bytes32 indexed lotHash,
        bytes32 indexed commodityId,
        bytes32 indexed vaultId,
        string barSerialNumber,
        uint256 netPureWeightGrams,
        bytes32 assayDigest,
        bytes32 eNwrId
    );

    event CommodityLotStatusUpdated(
        bytes32 indexed lotHash,
        LotStatus oldStatus,
        LotStatus newStatus,
        bytes32 indexed deliveryId
    );

    event VaultAudited(
        bytes32 indexed vaultId,
        uint256 auditTimestamp,
        bytes32 auditorReportDigest,
        bool auditPassed
    );

    event ProofOfReserveRootUpdated(
        bytes32 indexed commodityId,
        bytes32 indexed smtRoot,
        uint256 totalVaultedGrams,
        uint256 epochId
    );

    error UnauthorizedCaller(address caller, bytes32 requiredRole);
    error VaultAlreadyExists(bytes32 vaultId);
    error VaultNotFound(bytes32 vaultId);
    error VaultNotActive(bytes32 vaultId);
    error DuplicateBarSerialNumber(string barSerialNumber, string refineryCode);
    error LotAlreadyExists(bytes32 lotHash);
    error LotNotFound(bytes32 lotHash);
    error InvalidLotStatus(bytes32 lotHash, LotStatus currentStatus, LotStatus expectedStatus);
    error InvalidAssayCertificate(bytes32 assayDigest);
    error InvalidCustodianSignature(address recoveredSigner);
    error InvalidAssayerSignature(address recoveredSigner);
    error WeightCapacityExceeded(bytes32 vaultId, uint256 requestedWeight, uint256 remainingCapacity);
    error StaleAuditTimestamp(bytes32 vaultId, uint256 lastAuditTimestamp);

    function registerVaultFacility(
        bytes32 vaultId,
        string calldata wdraAccreditationNumber,
        string calldata operatorName,
        string calldata locationCity,
        string calldata locationState,
        uint256 maxStorageCapacityGrams,
        address primarySigner
    ) external;

    function updateVaultStatus(bytes32 vaultId, VaultStatus newStatus) external;

    function depositCommodityLot(
        CommodityLot calldata lot,
        CustodianAttestation calldata attestation
    ) external;

    function batchDepositCommodityLots(
        CommodityLot[] calldata lots,
        CustodianAttestation[] calldata attestations
    ) external;

    function lockLotForDelivery(bytes32 lotHash, bytes32 deliveryId) external;

    function markLotDelivered(bytes32 lotHash, bytes32 deliveryId) external;

    function quarantineLot(bytes32 lotHash, string calldata reason) external;

    function releaseQuarantine(bytes32 lotHash) external;

    function recordVaultAudit(
        bytes32 vaultId,
        bytes32 auditorReportDigest,
        bool auditPassed
    ) external;

    function updateProofOfReserveRoot(
        bytes32 commodityId,
        bytes32 smtRoot,
        uint256 totalVaultedGrams,
        uint256 epochId
    ) external;

    function getCommodityLot(bytes32 lotHash) external view returns (CommodityLot memory);

    function getVaultFacility(bytes32 vaultId) external view returns (VaultFacility memory);

    function getTotalActiveVaultedWeight(bytes32 commodityId) external view returns (uint256);

    function getProofOfReserveRoot(bytes32 commodityId) external view returns (bytes32 root, uint256 totalWeight, uint256 epochId);

    function isLotActive(bytes32 lotHash) external view returns (bool);
}
```

### Commodity Delivery Burner Interface (`ICommodityDeliveryBurner.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICommodityDeliveryBurner {
    enum DeliveryStatus {
        REQUESTED,
        APPROVED,
        DISPATCHED,
        COMPLETED,
        CANCELLED,
        EXPIRED
    }

    struct DeliveryRequest {
        bytes32 deliveryId;
        address claimant;
        bytes32 commodityId;
        uint256 tokenAmount;
        uint256 lotCount;
        bytes32 deliveryCommitmentHash;
        bytes32[] assignedLotHashes;
        DeliveryStatus status;
        uint256 requestedTimestamp;
        uint256 expiryTimestamp;
        uint256 completedTimestamp;
    }

    event DeliveryInitiated(
        bytes32 indexed deliveryId,
        address indexed claimant,
        bytes32 indexed commodityId,
        uint256 tokenAmount,
        uint256 lotCount,
        bytes32 deliveryCommitmentHash,
        uint256 expiryTimestamp
    );

    event DeliveryLotsAssigned(
        bytes32 indexed deliveryId,
        bytes32[] lotHashes,
        bytes32 indexed vaultId
    );

    event DeliveryDispatched(
        bytes32 indexed deliveryId,
        bytes32 gatePassDigest,
        bytes32 logisticsAirwayBillHash
    );

    event DeliveryCompleted(
        bytes32 indexed deliveryId,
        address indexed claimant,
        uint256 burnedTokenAmount,
        bytes32 receiptProofDigest
    );

    event DeliveryCancelled(
        bytes32 indexed deliveryId,
        address indexed claimant,
        uint256 refundedAmount,
        uint256 penaltyFeeAmount
    );

    error DeliveryAlreadyExists(bytes32 deliveryId);
    error DeliveryNotFound(bytes32 deliveryId);
    error InvalidDeliveryStatus(bytes32 deliveryId, DeliveryStatus current, DeliveryStatus expected);
    error DeliveryWindowExpired(bytes32 deliveryId, uint256 currentTimestamp, uint256 expiryTimestamp);
    error DeliveryWindowNotExpired(bytes32 deliveryId, uint256 currentTimestamp, uint256 expiryTimestamp);
    error InsufficientTokenEscrow(uint256 provided, uint256 required);
    error LotWeightMismatch(uint256 totalLotWeightGrams, uint256 requiredWeightGrams);
    error UnauthorizedVaultOperator(address caller, bytes32 vaultId);
    error ClaimantIdentityNotVerified(address claimant);

    function requestDelivery(
        bytes32 commodityId,
        uint256 tokenAmount,
        uint256 lotCount,
        bytes32 deliveryCommitmentHash,
        uint256 validUntil
    ) external returns (bytes32 deliveryId);

    function assignLotsAndAuthorizeDispatch(
        bytes32 deliveryId,
        bytes32[] calldata lotHashes,
        bytes32 gatePassDigest,
        bytes32 logisticsAirwayBillHash
    ) external;

    function confirmDeliveryAndBurn(
        bytes32 deliveryId,
        bytes calldata receiptSignature,
        bytes32 receiptProofDigest
    ) external;

    function cancelDelivery(bytes32 deliveryId) external;

    function getDeliveryRequest(bytes32 deliveryId) external view returns (DeliveryRequest memory);

    function calculateDeliveryFee(bytes32 commodityId, uint256 tokenAmount) external view returns (uint256 feeInTokens);
}
```

### Commodity Proof of Reserve Verifier Interface (`ICommodityPoRVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICommodityPoRVerifier {
    struct InclusionProofNode {
        bytes32 siblingHash;
        uint256 siblingWeight;
        bool isRightSibling;
    }

    function computeLotLeafHash(
        bytes32 lotHash,
        bytes32 commodityId,
        bytes32 vaultId,
        uint256 netPureWeightGrams,
        bytes32 assayDigest
    ) external pure returns (bytes32);

    function verifyLotInclusionProof(
        bytes32 expectedRoot,
        bytes32 lotHash,
        bytes32 commodityId,
        bytes32 vaultId,
        uint256 netPureWeightGrams,
        bytes32 assayDigest,
        InclusionProofNode[] calldata proofNodes
    ) external pure returns (bool isValid);
}
```

## Security & Compliance Notes
- **WDRA & SEBI Regulatory Alignment:** All physical vaults registered in `PhysicalVaultRegistry.sol` must hold valid accreditation from the Warehousing Development and Regulatory Authority (WDRA) and comply with SEBI Electronic Gold Receipt (EGR) and commodity trading guidelines.
- **1:1 Strict Physical Reserve Invariant:** The smart contracts programmatically ensure that `CommoditySecurityToken.totalSupply()` never exceeds the sum of `netPureWeightGrams` of active (`VAULTED`) lots in `PhysicalVaultRegistry.sol`. Any state transition attempting to violate this invariant immediately reverts.
- **Assay Certificate Cryptographic Anchoring:** Every lot entry mandates an immutable SHA-256 / Keccak-256 digest of the physical NABL/BIS assay certificate. Tampering with purity or gross weight yields an invalid leaf hash that fails Proof-of-Reserve verification.
- **Dual-Control Minting and Redemption:** Minting requires two distinct authorized cryptographic attestations: one from the WDRA vault custodian confirming physical bar receipt, and one from an accredited refinery assayer confirming purity.
- **Physical Delivery Escrow Lock:** When physical redemption is requested, tokens are escrowed into `CommodityDeliveryBurner.sol` and segregated from circulating supply. Tokens cannot be transferred, sold, or double-redeemed while delivery is pending.
- **Zero PII Ledger Protection:** All personal investor information relating to delivery logistics, residential addresses, and tax identity is kept strictly off-chain in encrypted storage. Delivery requests use salted zero-knowledge / cryptographic commitment hashes (`deliveryCommitmentHash`), ensuring full compliance with India's DPDP Act 2023.

## Acceptance Criteria
- [ ] `ICommoditySecurityToken.sol`, `IPhysicalVaultRegistry.sol`, `ICommodityDeliveryBurner.sol`, and `ICommodityPoRVerifier.sol` compile cleanly with Solidity 0.8.24 with zero warnings.
- [ ] 100% test coverage across Foundry unit, fuzz, and invariant test suites under `test/commodities/`.
- [ ] `mintWithVaultLot` strictly enforces ERC-3643 KYC compliance via `IIdentityRegistry` and validates that the referenced lot is registered and active in `PhysicalVaultRegistry.sol`.
- [ ] `depositCommodityLot` validates dual-party signatures (vault custodian and refinery assayer) and prevents duplicate bar serial numbers.
- [ ] End-to-end physical delivery flow (Request -> Assign Lots -> Dispatch -> Confirm Handover -> Permanent Burn) operates deterministically, decrementing total supply and updating lot status to `DELIVERED`.
- [ ] Fuzz tests verify that Proof-of-Reserve Merkle validation correctly confirms valid inclusion proofs and rejects manipulated lot weights or forged sibling nodes.
- [ ] Slither static analysis returns zero high, medium, or reentrancy issues across the entire commodity contract suite.
- [ ] Deterministic deployment scripts deploy ERC-1967 upgradeable proxies to local Besu QBFT network with verified initialization parameters.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract), Prompt `304` (Token Redemption Smart Contract), Prompt `307` (MultiSig Governance Smart Contract), Prompt `327` (Multi-Chain Proof-of-Reserve Registry), Prompt `007` (Proof of Reserve Public Disclosure).
- **Parallel Tasks:** Prompt `213` (Custodian Depository Integration Service), Prompt `228` (Real-Time Market Surveillance Engine), Prompt `328` (Decentralized Oracle Aggregation Smart Contract).
- **Subsequent Prompts Enabled:** Prompt `242` (Commodity Vault Logistics & Physical Settlement Service), Prompt `608` (Commodity Vault & Physical Delivery Web Portal).
