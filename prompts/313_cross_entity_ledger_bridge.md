# 313 - Inter-Entity Ledger Bridge (Domestic Depository <-> GIFT City Gateway)

## Purpose
The Growww architecture operates across two distinct legal and regulatory jurisdictions: (1) The **Domestic Regulated Entity** (operating under SEBI / RBI domestic jurisdiction for Indian investors and holding custody of underlying Indian equities), and (2) The **GIFT City International Gateway Entity** (operating under IFSCA jurisdiction for foreign/NRI investors). To enable foreign investors in GIFT City to trade fractionalized Indian equities without violating cross-border data privacy (DPDP Act / GDPR) or foreign exchange regulations (FEMA / LRS), the two ledgers must synchronize settlement state, custodial allocations, and proof-of-reserve records through a cryptographically secure, privacy-preserving bridge.

This prompt specifies the architecture, implementation, and deployment of the **Inter-Entity Cross-Ledger Bridge & Mirror Registry** (`services/cross-entity-bridge/` and `contracts/bridge/`). The bridge guarantees that every fractional equity token mirrored on the GIFT City ledger corresponds to an immutable, locked custodial allocation on the Domestic ledger, verifying state across jurisdictions via Merkle Mountain Range (MMR) proofs and zero-knowledge / hash-commitment relays without ever leaking domestic or foreign investor PII.

## What You Are Building
A high-assurance cross-entity synchronization system comprising:
- `DomesticLedgerBridge.sol`: Smart contract on the Domestic Besu ledger managing cross-border custody allocations, partition locks, and cryptographic state commitments.
- `GiftCityLedgerBridge.sol`: Smart contract on the GIFT City Besu ledger managing mirrored international equity tokens, foreign settlement proofs, and redemption triggers.
- Cross-Entity Bridge Relayer Service (Go / Rust): Dual-headed relayer daemon operating over an audited, encrypted mTLS tunnel connecting domestic and GIFT City enclaves.
- Cryptographic State Verifier: Validates block header commitments and transaction receipt Merkle proofs across both ledgers.
- Real-Time Cross-Entity Auditor Daemon: Continuously asserts that $\sum \text{GIFT City Mirrored Tokens} \le \text{Domestic Custody Allocation Locked}$ for every equity ISIN.

## Scope Boundaries
- **In Scope:**
 - Cross-ledger cryptographic state relay and Merkle receipt proof verification.
 - Locking and unlocking custodial allocations on Domestic Ledger.
 - Minting and burning mirrored international security tokens on GIFT City Ledger.
 - Audited mTLS communication between Domestic and GIFT City cloud VPCs.
 - Zero PII transfer across the international boundary.
- **Out of Scope / Handled Elsewhere:**
 - Foreign investor KYC/AML onboarding in GIFT City (handled in Prompt 005 & 202).
 - FX conversion and international USD/INR banking flows (handled in Prompt 214).
 - Inter-entity secure network tunnel provisioning (handled in Prompt 110).

## Technology to Use
- **Relayer Language:** **Go (v1.22+) or Rust (v1.78+)**.
  *Justification:* Provides native concurrency, robust mTLS gRPC support, and high-performance cryptographic hashing (`keccak256`, SHA-256) for multi-chain event listening and verification.
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
- **Cryptographic Primitives:** Merkle Mountain Range (MMR) inclusion proofs, EIP-712 structured dual-attestations.
- **Transport Security:** gRPC over mTLS with hardware-backed client certificates.

## Backend / Infra Touchpoints
- **Domestic Besu RPC:** Connects to Domestic Consortium Ledger (`ChainID: 13370`).
- **GIFT City Besu RPC:** Connects to GIFT City International Consortium Ledger (`ChainID: 13371`).
- **Foreign Investor Funding & FX Service (Prompt 214):** Informs bridge of incoming international capital allocations.
- **Inter-Entity Tunnel (Prompt 110):** Encrypted IPsec / WireGuard VPN tunnel connecting Mumbai data center with GIFT City IFSC data center.

## Blockchain Interaction
- **Domestic Leg:** When foreign capital enters GIFT City to buy Indian shares, the domestic broker purchases physical shares at NSDL/CDSL; `DomesticLedgerBridge.sol` locks the underlying `DigitalSecurityToken` units in an `InternationalEscrowPartition` and emits `CustodyAllocationCommitted(batchId, isin, amount, rootHash)`.
- **GIFT City Leg:** The bridge relayer presents the signed cryptographic receipt to `GiftCityLedgerBridge.sol`; upon verification, mirrored tokens (`GROWWW-GIFT-RELIANCE`) are minted to the foreign investor's verified GIFT City wallet.
- **Zero PII Guarantee:** Only anonymous batch IDs, ISIN codes, unit amounts, and cryptographic state roots are transmitted across the bridge. No investor personal details, PAN, or passport numbers ever cross the boundary.

## Step-by-Step Build Instructions
1. Scaffold project directory layout: `contracts/bridge/domestic/`, `contracts/bridge/gift-city/`, `services/cross-entity-bridge/`, `proto/bridge/`.
2. Define Protobuf definitions for inter-entity state relay (`proto/bridge/cross_ledger.proto`): message types `StateCommitmentBatch`, `ReceiptProof`, and `BridgeSyncResponse`.
3. Implement `DomesticLedgerBridge.sol`:
 - Store authorized GIFT City bridge address and relayer public keys.
 - Implement `lockCustodialAllocation(string calldata isin, uint256 amount, bytes32 giftCityRecipientHash)`:
 - Transfer `DigitalSecurityToken` from domestic inventory to bridge escrow partition.
 - Generate incremental batch ID and update state root.
 - Emit `CustodialAllocationLocked(batchId, isin, amount, giftCityRecipientHash, block.number)`.
4. Implement `releaseCustodialAllocation(bytes32 batchId, string calldata isin, uint256 amount, bytes calldata giftCityBurnProof)`:
 - Verify cryptographic proof of token burn on GIFT City ledger.
 - Release tokens from escrow partition back to domestic inventory.
 - Emit `CustodialAllocationReleased(batchId, isin, amount)`.
5. Implement `GiftCityLedgerBridge.sol`:
 - Maintain mapping of authorized domestic state roots.
 - Implement `mintMirroredTokenWithProof(bytes32 batchId, string calldata isin, address foreignInvestor, uint256 amount, bytes32[] calldata merkleProof)`:
 - Verify proof against latest domestic state root.
 - Verify `foreignInvestor` is verified in GIFT City Identity Registry.
 - Mint mirrored security token.
 - Emit `MirroredTokenMinted(batchId, isin, foreignInvestor, amount)`.
6. Implement `burnMirroredToken(string calldata isin, uint256 amount, bytes32 domesticDestinationHash)`:
 - Burn mirrored tokens from foreign investor wallet.
 - Emit `MirroredTokenBurned(batchId, isin, amount, domesticDestinationHash)`.
7. Implement the Bridge Relayer daemon in Go/Rust (`services/cross-entity-bridge/`):
 - Subscribe to WebSocket event streams on Domestic Ledger (`ChainID: 13370`) and GIFT City Ledger (`ChainID: 13371`).
 - Listen for `CustodialAllocationLocked` and `MirroredTokenBurned` events.
 - Generate Merkle receipt inclusion proofs.
 - Submit cross-ledger verification transactions via Web3Signer HSM.
8. Implement the Real-Time Cross-Entity Auditor Daemon:
 - Polls `totalLocked(isin)` on Domestic Bridge and `totalSupply(isin)` on GIFT City Bridge every 10 seconds.
 - Asserts invariant: $\text{TotalLocked}_{\text{Domestic}} \ge \text{TotalSupply}_{\text{GIFT}}$.
 - Alerts via PagerDuty and halts bridge if any discrepancy $> 0$ occurs.
9. Write Foundry unit tests in `test/bridge/` verifying cross-chain proof validation and replay prevention.
10. Write integration tests executing a full cycle: Domestic share lock $\rightarrow$ Bridge relay $\rightarrow$ GIFT City mirror mint $\rightarrow$ GIFT City trade $\rightarrow$ GIFT City mirror burn $\rightarrow$ Domestic unlock.
11. Perform Slither security audit on bridge contracts to ensure reentrancy resistance and proof verification correctness.
12. Configure automated Docker container builds and Kubernetes deployments in domestic and GIFT City namespaces.

## Interfaces / Contracts

### Domestic Ledger Bridge Interface (`IDomesticLedgerBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IDomesticLedgerBridge {
    event CustodialAllocationLocked(
        bytes32 indexed batchId,
        string indexed isin,
        uint256 amount,
        bytes32 indexed giftCityRecipientHash,
        uint256 blockNumber
    );

    event CustodialAllocationReleased(
        bytes32 indexed batchId,
        string indexed isin,
        uint256 amount,
        address recipient
    );

    function lockCustodialAllocation(
        string calldata isin,
        uint256 amount,
        bytes32 giftCityRecipientHash
    ) external returns (bytes32 batchId);

    function releaseCustodialAllocation(
        bytes32 batchId,
        string calldata isin,
        uint256 amount,
        address domesticRecipient,
        bytes calldata giftCityBurnProof
    ) external;

    function getLockedAllocation(string calldata isin) external view returns (uint256);
}
```

### GIFT City Ledger Bridge Interface (`IGiftCityLedgerBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IGiftCityLedgerBridge {
    event MirroredTokenMinted(
        bytes32 indexed batchId,
        string indexed isin,
        address indexed foreignInvestor,
        uint256 amount
    );

    event MirroredTokenBurned(
        bytes32 indexed batchId,
        string indexed isin,
        address indexed foreignInvestor,
        uint256 amount,
        bytes32 domesticDestinationHash
    );

    function mintMirroredTokenWithProof(
        bytes32 batchId,
        string calldata isin,
        address foreignInvestor,
        uint256 amount,
        bytes calldata domesticLockProof
    ) external;

    function burnMirroredToken(
        string calldata isin,
        uint256 amount,
        bytes32 domesticDestinationHash
    ) external returns (bytes32 batchId);
}
```

## Security & Compliance Notes
- **Strict Custodial Isolation:** Mirrored tokens in GIFT City cannot exist without an identical quantity of real shares locked in custody on the domestic ledger. Synthetic or unbacked issuance is strictly impossible.
- **Cross-Border Privacy Protection (Zero PII):** No investor names, PAN numbers, Aadhaar details, or foreign passport details cross the bridge network. Transactions exchange only cryptographic hashes and numerical amounts.
- **Replay & Fork Resistance:** Batch IDs include chain IDs, domain separators, and incremental nonces. Replaying a lock proof from testnet or a different chain will revert during cryptographic verification.
- **Emergency Partition Freeze:** If an audit discrepancy or network breach is detected, either entity's compliance officer can trigger an immediate freeze of the bridge contracts via multi-sig.

## Acceptance Criteria
- [ ] `DomesticLedgerBridge.sol` and `GiftCityLedgerBridge.sol` deployed on respective Besu ledgers.
- [ ] Dual-headed Go/Rust bridge relayer successfully synchronizes lock/mint and burn/unlock events across ledgers.
- [ ] Invariant holds under all test workloads: $\text{Domestic Locked} \ge \text{GIFT Mirrored Supply}$.
- [ ] Cross-chain proof verification fails and reverts if proof payload is tampered with or replayed.
- [ ] Slither / Mythril security analysis produces zero high/medium severity findings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `002` (Two-Entity Legal/Technical Structure), Prompt `110` (Inter-Entity Secure Communication), Prompt `303` (Token Issuance), Prompt `304` (Token Redemption).
- **Parallel Tasks:** Prompt `214` (Foreign Investor Funding & FX Service), Prompt `005` (Foreign KYC Policy).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (Proof of Reserve).
