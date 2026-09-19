# 007 - Proof-of-Reserve & Custody Verification Architecture

## Purpose
The foundational promise of Growww is absolute solvency and transparency: every single fractional equity unit issued on the permissioned ledger is backed 1:1 by real, physical shares held in depository custody (NSDL/CDSL) by licensed SEBI-registered custodians. 

To eliminate counterparty risk and prevent fractional reserve practices, Growww implements a public, cryptographically verifiable Proof-of-Reserve (PoR) architecture. This system enables any investor, auditor, or regulatory body to mathematically verify that total on-chain token supply matches physical custody balances in real time, and allows individual users to verify their balance inclusion using Merkle proofs without exposing any personal data or balances of other investors.

## What You Are Building
A comprehensive Proof-of-Reserve architecture specification (`docs/architecture/proof_of_reserve_spec.md`) that details:
- Daily automated custody holding ingestion and reconciliation protocol with SEBI-registered custodians.
- The Cryptographic Merkle Tree construction algorithm: generating Sparse Merkle Trees (SMT) of all user token balances and depository holdings.
- The on-chain attestation publishing mechanism on `ProofOfReserveRegistry.sol` on Hyperledger Besu.
- The Self-Service Investor Verification Protocol (enabling users to download Merkle inclusion paths and verify their balances client-side).
- The Public Transparency Dashboard data feeds and open-source CLI verification tool (`growww-por-verifier`).
- Multi-party attestation signing: dual cryptographic signatures from independent auditor HSMs and custodian APIs.

## Scope Boundaries
- **In Scope:**
 - Merkle tree structure, hashing algorithms, and zero-knowledge / salted leaf schemas.
 - Custodian reconciliation data pipeline and reserve discrepancy alert protocols.
 - Smart contract attestation registry interfaces and public verification APIs.
 - Standalone verification CLI specification.
- **Out of Scope / Handled Elsewhere:**
 - Custodian API adapter microservice implementation (covered in Prompt 213).
 - Internal reconciliation engine implementation (covered in Prompt 215).
 - Flutter UI screen for Proof-of-Reserve (covered in Prompt 514).
 - Next.js Public Transparency Portal (covered in Prompt 605).

## Technology to Use
- Cryptographic Structures: Sparse Merkle Trees (SMT) with SHA-256 / Keccak-256; salted leaf hashes for zero-PII privacy.
- Verifier Tooling: Rust / Go standalone CLI (`growww-por-verifier`) compiled for Linux, macOS, and Windows.
- On-Chain Registry: Solidity smart contract (`ProofOfReserveRegistry.sol`) deployed on Hyperledger Besu.
- Public Storage: Encrypted Amazon S3 / Cloudflare R2 bucket distributing daily Merkle root artifacts and proof trees.
- Justification: SMTs with salted leaves allow O(log N) cryptographic proof generation while ensuring individual user balances and identity hashes cannot be reverse-engineered or enumerated by third parties.

## Backend / Infra Touchpoints
- Custodian Depository Integration Service (Prompt 213).
- Daily Reconciliation Engine (Prompt 215).
- Hyperledger Besu Consortium Node RPC.
- Public Content Delivery Network (CDN) hosting daily attestation files.

## Blockchain Interaction
Establishes the on-chain attestation and verification pipeline on Hyperledger Besu:
- **`ProofOfReserveRegistry.sol` Contract:** Stores immutable daily attestation records containing:
 - `isin`: Security identifier.
 - `depositoryShareCount`: Total physical shares verified by custodian API.
 - `onChainTokenSupply`: Total digital tokens minted and circulating.
 - `merkleRoot`: 32-byte root hash of the user balance tree.
 - `custodianSignature`: Cryptographic signature from custodian's HSM.
 - `auditorSignature`: Cryptographic signature from independent auditor's HSM.
- **Atomic Invariant Check:** If `onChainTokenSupply > depositoryShareCount`, the smart contract automatically locks new minting and triggers high-severity emergency alerts.
- **Zero-PII Invariant:** Leaves in the Merkle tree are computed as `leaf = keccak256(investor_commitment + salt + balance_string)`. No names or raw identifiers exist in the tree.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review the 1:1 custody backing invariant established in Prompt 000.
2. Design the Custody Ingestion Pipeline: at market close (15:30 IST), automatically query NSDL/CDSL holding statements via Custodian API.
3. Design the On-Chain Supply Aggregator: query `totalSupply()` for all `DigitalSecurityToken` contracts on Hyperledger Besu.
4. Define the Balance Reconciliation Invariant: assert $\text{Depository Holding} \ge \text{Total On-Chain Supply}$.
5. Formulate the Merkle Tree Construction Algorithm: for each investor holding fractional units of ISIN $k$, generate a leaf:
   $$\text{Leaf}_i = \text{SHA256}(\text{Investor Commitment}_i \parallel \text{User Secret Salt}_i \parallel \text{Balance}_i)$$
6. Construct the Merkle Tree, calculate the intermediate nodes, and determine the canonical 32-byte `MerkleRoot`.
7. Prepare the Daily Attestation Payload containing depository certificate data, token supply, block height, and Merkle root.
8. Implement the Multi-Signature Signing Flow: require M-of-N threshold signatures from Custodian HSM, Auditor HSM, and Platform HSM.
9. Submit the signed attestation to `ProofOfReserveRegistry.sol` on Hyperledger Besu via relayer.
10. Generate and store individual Merkle inclusion paths in S3/R2 for all active investors, accessible via authenticated API.
11. Specify the open-source CLI verification tool (`growww-por-verifier`) enabling anyone to verify the Merkle root against the on-chain registry.
12. Establish the Breach Protocol: any reserve deficit immediately freezes new token minting and broadcasts an automated notification to SEBI/IFSCA.
13. Document schemas and review with Independent Custodian Auditors, CISO, and Lead Blockchain Architect.

## Interfaces / Contracts
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProofOfReserveRegistry {
    struct Attestation {
        string isin;
        uint256 depositoryShareCount;   // 18-decimal precision
        uint256 onChainTokenSupply;     // 18-decimal precision
        bytes32 merkleRoot;
        uint256 blockNumber;
        uint256 timestamp;
        bytes custodianSignature;
        bytes auditorSignature;
    }

    event ReserveAttested(
        string indexed isin,
        uint256 depositoryShareCount,
        uint256 onChainTokenSupply,
        bytes32 merkleRoot,
        uint256 timestamp
    );

    event ReserveDiscrepancyDetected(
        string indexed isin,
        uint256 depositoryShareCount,
        uint256 onChainTokenSupply,
        uint256 timestamp
    );

    function submitDailyAttestation(Attestation calldata attestation) external;
    function getLatestAttestation(string calldata isin) external view returns (Attestation memory);
    function verifyInclusionProof(
        string calldata isin,
        bytes32 leaf,
        bytes32[] calldata merkleProof
    ) external view returns (bool);
}
```

## Security & Compliance Notes
- Individual user salts ensure that bad actors cannot rainbow-table user balances or estimate other investors' portfolio sizes from public Merkle dumps.
- Depository holding reports must originate directly from SEBI-registered custodians via signed cryptographic channels (mTLS + PKCS#7 signatures).
- Public disclosure complies with SEBI sandbox transparency mandates and IFSCA client asset segregation rules.

## Acceptance Criteria
- [ ] `docs/architecture/proof_of_reserve_spec.md` is complete with Merkle tree construction algorithms, reconciliation invariants, and breach workflows.
- [ ] Solidity interface `IProofOfReserveRegistry` is fully specified with multi-signature verification logic.
- [ ] Merkle leaf generation algorithm with privacy-preserving salts is mathematically validated against deanonymization attacks.
- [ ] Standalone verifier CLI architecture is detailed and tested with sample proof paths.
- [ ] Formal sign-off obtained from Independent Custodian Auditor and Chief Technology Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary), 002 (Two-Entity Structure).
- Parallel Tasks: 213 (Custodian Integration), 215 (Reconciliation), 306 (Registry Smart Contract).
- Downstream Blockers: Blocks Prompt 215, Prompt 306, Prompt 514, and Prompt 605.
