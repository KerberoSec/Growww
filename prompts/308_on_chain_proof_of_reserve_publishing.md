# 308 - On-Chain Proof-of-Reserve Attestation & Merkle Registry Smart Contract

## Purpose
To guarantee investor trust, regulatory compliance, and eliminate fractional reserve risks in asset-backed security tokenization, Growww must provide continuous, cryptographically verifiable evidence that 100% of outstanding digital equity tokens are matched 1:1 by real Indian shares held in physical custody at SEBI-regulated depositories (NSDL/CDSL). 

This prompt specifies the design, implementation, and deployment of the **On-Chain Proof-of-Reserve (PoR) Attestation & Merkle Registry** (`ProofOfReserveRegistry.sol`). At scheduled intraday intervals (and at daily market close), the platform ingests signed custodian holding reports, aggregates individual investor fractional equity balances into a cryptographic Merkle Sum Tree, and publishes the Merkle Root, aggregate token liability, and custodian depository reserve attestations onto the immutable Hyperledger Besu blockchain.

## What You Are Building
A complete cryptographic proof-of-reserve publishing and verification system comprising:
- `ProofOfReserveRegistry.sol`: On-chain smart contract storing historical Merkle root commitments, custodian digital signatures, and asset-by-asset solvency ratios.
- Merkle Sum Tree Engine (Go / Rust service): Off-chain engine that generates cryptographically balanced Merkle Sum Trees mapping all user balances to the published root without leaking individual account balances or PII.
- On-Chain Invariant Verification Library: Computes whether Total Depository Reserves $\ge$ Total Token Supply for each equity ISIN.
- Comprehensive Foundry unit and fuzz test suite (`test/por/ProofOfReserveRegistry.t.sol`) validating cryptographic signature checks, Merkle leaf inclusion verification, and historical attestation queries.

## Scope Boundaries
- **In Scope:**
 - On-chain storage of periodic reserve attestations (Merkle root, timestamp, ISIN-level reserve counts, custodian signature).
 - Smart contract verification of custodian public key ECDSA / RSA signatures.
 - Public on-chain inclusion proof verification (`verifyUserInclusionProof`).
 - Automated solvency assertion: reverts/alerts if Depository Reserves < Minted Supply.
- **Out of Scope / Handled Elsewhere:**
 - Off-chain Custodian Demat statement ingestion adapter (handled in Prompt 213).
 - Triple-ledger reconciliation service (handled in Prompt 215).
 - Public Proof-of-Reserve UI / Web verification tool (handled in Prompt 606 / 007).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Provides efficient hashing via `keccak256` and optimized memory operations for Merkle proof validation.
- **Off-Chain Engine Language:** **Go (v1.22+) or Rust (1.78+)**.
  *Justification:* High-performance multi-threaded Merkle tree generation capable of hashing 1,000,000+ investor accounts in <500ms.
- **Libraries:** OpenZeppelin `MerkleProof.sol`, `ECDSA.sol`, `EIP712Upgradeable`.
- **Tooling:** Foundry (`forge`, `cast`), Slither static analyzer.

## Backend / Infra Touchpoints
- **Reconciliation Service:** Backend microservice (Prompt 215) that triggers PoR publishing jobs upon completing daily settlement cycles.
- **Custodian Depository Service:** Microservice (Prompt 213) providing signed NSDL/CDSL holding statements.
- **Event Indexer & Public Portal:** Indexing pipeline (Prompt 309) making Merkle branches downloadable for self-sovereign user verification.

## Blockchain Interaction
- **Attestation Submission:** Invoked by the automated Proof-of-Reserve Relayer submitting `publishReserveAttestation(...)`.
- **Contract State:** Stores an append-only historical array of `AttestationRecord` structs indexed by `attestationId` and `timestamp`.
- **Zero PII & Privacy-Preserving Trees:** The Merkle Sum Tree leaves contain `hash(account_uuid + salt)` paired with fractional balance amounts. No PAN, investor names, or wallet addresses are publicly linkable from tree leaves.

## Step-by-Step Build Instructions
1. Scaffold repository directory structure under `contracts/por/` and `services/por-engine/`.
2. Define `IProofOfReserveRegistry.sol` interface.
3. Define data structures:
 - `struct AssetReserve { string isin; uint256 depositoryCustodyUnits; uint256 onChainMintedUnits; uint256 blockHeight; }`
 - `struct AttestationRecord { bytes32 attestationId; bytes32 merkleRoot; uint64 timestamp; address custodianSigner; AssetReserve[] reserves; bytes custodianSignature; }`
4. Implement `ProofOfReserveRegistry.sol` inheriting `Initializable`, `AccessControlUpgradeable`, and `PausableUpgradeable`.
5. Implement `initialize(address adminMultisig, address initialCustodianSigner, address initialRelayer)`.
6. Implement `publishReserveAttestation(bytes32 attestationId, bytes32 merkleRoot, AssetReserve[] calldata reserves, bytes calldata custodianSignature)`:
 - Verify caller has `POR_RELAYER_ROLE`.
 - Validate `custodianSignature` recovers to authorized `custodianSigner` using EIP-712 structured data hash.
 - Iterate over `reserves`: verify that for every ISIN, `depositoryCustodyUnits >= onChainMintedUnits`.
 - Verify `DigitalSecurityToken(tokenByISIN[isin]).totalSupply() == onChainMintedUnits`.
 - Store attestation record in historical storage array.
 - Emit `ReserveAttestationPublished(attestationId, merkleRoot, block.timestamp, reserves.length)`.
7. Implement `verifyUserInclusionProof(bytes32 attestationId, bytes32 leafHash, bytes32[] calldata merkleProof) external view returns (bool)` using OpenZeppelin `MerkleProof.verify`.
8. Implement getter functions: `getLatestAttestation()`, `getAttestationById(bytes32 attestationId)`, `getAssetReserve(bytes32 attestationId, string calldata isin)`.
9. Implement the off-chain Go/Rust Merkle Sum Tree generator under `services/por-engine/`:
 - Load all user fractional balances from PostgreSQL.
 - Salt and hash investor identifiers (`keccak256(userId + salt)`).
 - Construct binary Merkle Sum Tree where parent node = `hash(left.hash + left.sum + right.hash + right.sum)` and parent sum = `left.sum + right.sum`.
 - Output root hash and individual cryptographic inclusion proofs (leaf + sibling branch).
10. Write Foundry unit tests in `test/por/ProofOfReserveRegistry.t.sol` testing valid attestation publication, invalid custodian signature rejection, and reserve deficiency reversion.
11. Write fuzz tests verifying Merkle leaf verification for 1,000 randomized tree depths (depth 1 to 24).
12. Run Slither analysis and verify zero storage collision or reentrancy issues.
13. Deploy `ProofOfReserveRegistry.sol` on Hyperledger Besu devnet.
14. Perform end-to-end integration test: generate Merkle root from 10,000 synthetic investor accounts, publish to Besu node, and verify single user inclusion proof via JSON-RPC.
15. Export ABI and Go/Rust bindings using `abigen` / Foundry artifacts.

## Interfaces / Contracts

### Proof of Reserve Registry Interface (`IProofOfReserveRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IProofOfReserveRegistry {
    struct AssetReserve {
        string isin;
        uint256 depositoryCustodyUnits; // 18 decimals, confirmed by NSDL/CDSL
        uint256 onChainMintedUnits;      // 18 decimals, on-chain total supply
        uint256 verifiedBlockHeight;
    }

    struct AttestationRecord {
        bytes32 attestationId;
        bytes32 merkleRoot;
        uint64 timestamp;
        address custodianSigner;
        bytes custodianSignature;
    }

    event ReserveAttestationPublished(
        bytes32 indexed attestationId,
        bytes32 indexed merkleRoot,
        uint64 timestamp,
        uint256 assetsCount
    );

    event SolvencyBreachAlert(
        bytes32 indexed attestationId,
        string isin,
        uint256 custodyUnits,
        uint256 tokenUnits
    );

    event CustodianSignerUpdated(address indexed oldSigner, address indexed newSigner);

    // Errors
    error AttestationAlreadyExists(bytes32 attestationId);
    error AttestationNotFound(bytes32 attestationId);
    error InvalidCustodianSignature();
    error ReserveDeficiency(string isin, uint256 custodyUnits, uint256 tokenUnits);
    error UnauthorizedRelayer(address caller);
    error EmptyReservesList();

    // Publication
    function publishReserveAttestation(
        bytes32 attestationId,
        bytes32 merkleRoot,
        AssetReserve[] calldata reserves,
        bytes calldata custodianSignature
    ) external;

    // Proof Verification
    function verifyUserInclusionProof(
        bytes32 attestationId,
        bytes32 leafHash,
        bytes32[] calldata merkleProof
    ) external view returns (bool isValid);

    // Query Views
    function getLatestAttestationId() external view returns (bytes32);
    function getAttestation(bytes32 attestationId) external view returns (AttestationRecord memory, AssetReserve[] memory);
    function getCustodianSigner() external view returns (address);
}
```

## Security & Compliance Notes
- **Independent Custodian Cryptographic Signature:** The custodian bank/depository signs the exact balance payload using their HSM key before submission. The smart contract independently validates this signature, eliminating reliance on Growww's internal relayers.
- **Solvency Invariant:** `depositoryCustodyUnits >= onChainMintedUnits` is enforced on-chain. If any asset fails this check, the contract emits a high-severity `SolvencyBreachAlert` and automatically halts secondary trading via circuit-breaker hooks.
- **Privacy-Preserving Auditability:** By combining salted hashes with Merkle Sum Trees, individual investors can verify their balance inclusion without revealing their identity or balance to any other investor or external observer.

## Acceptance Criteria
- [ ] `ProofOfReserveRegistry.sol` deployed and verified on Hyperledger Besu.
- [ ] Automated Go/Rust Merkle tree generator builds balanced trees for 100,000 accounts in under 1 second.
- [ ] Valid custodian signatures pass verification; tampered signatures revert with `InvalidCustodianSignature`.
- [ ] On-chain proof verification correctly confirms valid user inclusion branches and rejects forged branches.
- [ ] 100% test coverage in Foundry with Slither analysis passing cleanly.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance), Prompt `007` (Proof of Reserve Public Disclosure), Prompt `215` (Reconciliation Service).
- **Parallel Tasks:** Prompt `213` (Custodian Integration Service), Prompt `606` (Admin PoR Dashboard).
- **Subsequent Prompts Enabled:** Prompt `510` (Flutter Portfolio PoR View), Prompt `608` (Public Landing Disclosure).
