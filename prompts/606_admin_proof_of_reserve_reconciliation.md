# 606 - Admin & Public Proof-of-Reserve Verification Dashboard

## Purpose
Implements both an internal administrative reconciliation monitor (`apps/growww_admin/app/(dashboard)/reconciliation/`) and a transparent public proof-of-reserve (PoR) verification portal (`apps/growww_web/app/(marketing)/reserves/`). Demonstrates real-time 1:1 mathematical backing between physical equity shares held in SEBI-registered depositories (NSDL/CDSL) and digital fractional security tokens minted on the permissioned Hyperledger Besu blockchain. Empowers any retail investor, institutional partner, auditor, or regulator to independently verify Merkle inclusion proofs for their individual account holdings without compromising investor financial privacy or exposing PII.

## What You Are Building
- Public Proof-of-Reserve Transparency Portal (`apps/growww_web/app/(marketing)/reserves/`).
- Interactive Client-Side Merkle Tree Verifier tool: allows an investor to paste their anonymous balance leaf hash and verify Merkle branch inclusion against the on-chain root published in `ProofOfReserveRegistry.sol`.
- Live 1:1 Reserve Ratio Health Gauge (Total Physical Custody Shares vs Total Minted Digital Tokens).
- ISIN-level asset breakdown table displaying listed equities, underlying depository custody certificates, and daily attestation timestamps.
- Depository Attestation & Auditor Certificate Viewer with downloadable digitally signed PDF/CSV reports.
- Internal Admin Reconciliation Workstation (`apps/growww_admin/app/(dashboard)/reconciliation/`) with automated discrepancy alerts, break investigation workflows, and settlement timing adjustment tools.

## Scope Boundaries
- **In Scope:**
 - Public transparency portal UI, live reserve ratio gauge, and asset breakdown table.
 - Client-side cryptographic Merkle leaf hash computation and branch verification algorithms.
 - Depository statement preview and signed audit certificate download manager.
 - Admin reconciliation workstation, exception queue, and break resolution forms.
 - Real-time on-chain Merkle root polling and contract event listeners.
- **Out of Scope / Handled Elsewhere:**
 - Backend reconciliation calculation engine (Prompt 215).
 - `ProofOfReserveRegistry.sol` smart contract implementation (Prompt 308).
 - Custodian/Depository API adapter service (Prompt 213).
 - Public proof-of-reserve policy definition (Prompt 007).

## Technology to Use
- **Next.js 14 App Router, React 18/19, TypeScript 5.4+:** Provides server-side rendered public transparency pages for instant load times and SEO indexing, combined with reactive client components for cryptographic verification.
- **Viem 2.x & `keccak256`:** Executes client-side cryptographic hashing and on-chain contract state queries directly against Hyperledger Besu RPC. Justification: Native Viem hashing functions guarantee cryptographic consistency with EVM smart contract logic.
- **`merkletreejs`:** Fast, lightweight TypeScript library for reconstructing Merkle proof branches and validating inclusion proofs in the browser.
- **Recharts / Tremor & Tailwind CSS:** Clean, financial-grade data visualization charts (reserve ratio over time, asset allocation distribution).
- **Framer Motion:** Interactive animated visual step-through of the Merkle verification tree path from leaf to root.

## Backend / Infra Touchpoints
- **Public Proof-of-Reserve API (`/api/v1/public/reserves`):** Fetches current aggregate reserve metrics, ISIN breakdowns, and published Merkle root metadata from the Reconciliation Service (Prompt 215).
- **Custodian Integration Service (Prompt 213):** Fetches verified depository holding snapshots from NSDL/CDSL.
- **Hyperledger Besu JSON-RPC Node (Prompt 302):** Directly queries `ProofOfReserveRegistry.sol` to fetch latest attested on-chain roots and event logs.
- **IPFS / Decentralized Object Storage (Prompt 401):** Resolves immutable audit report URIs.

## Blockchain Interaction
- **On-Chain Root Verification:** The verification widget reads the active Merkle root directly from `ProofOfReserveRegistry.sol` using Viem: `getLatestAttestation(string isin) returns (bytes32 merkleRoot, uint256 totalShares, uint256 timestamp, string ipfsUri)`.
- **Event Monitoring:** Subscribes to `ReserveAttestationPublished` events on Hyperledger Besu to refresh the public health gauge automatically whenever a new custody snapshot is attested.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `ProofOfReserveRegistry.sol` (Maintains cryptographic Merkle roots, total custody counts, and auditor attestations per ISIN).
 - `DigitalSecurityToken.sol` (Queries `totalSupply()` to compare against physical custody holdings).
- **Privacy-Preserving Architecture:** User account balances are obfuscated in the Merkle tree via salt/nonce hashing: $\text{Leaf} = \text{keccak256}(\text{accountId} \parallel \text{balance} \parallel \text{salt})$. No public user names or raw financial figures are exposed on-chain.

## Step-by-Step Build Instructions
1. Scaffold public `/reserves` route under `apps/growww_web/app/(marketing)/` and internal `/reconciliation` route under `apps/growww_admin/app/(dashboard)/`.
2. Implement Top Hero Section on the public portal featuring a prominent "100.00% Fully Backed" reserve status gauge with live pulsing indicator.
3. Build Asset Overview Card displaying Total Physical Shares held in NSDL/CDSL Custody vs Total Minted Digital Tokens across all listed equities.
4. Build ISIN Reserve Breakdown Table with columns: Stock Name, Ticker, ISIN, Depository (NSDL/CDSL), Physical Custody Count, Digital Token Supply, Reserve Ratio %, and Last Attestation Timestamp.
5. Implement Interactive Merkle Leaf Verifier Component with user inputs: "Anonymous Account ID", "Fractional Token Balance", "Account Salt/Nonce", and "Merkle Proof Siblings Array" (or single JSON proof paste).
6. Implement client-side `keccak256` leaf generator and Merkle path traversal verifying calculated root against `ProofOfReserveRegistry.sol` on-chain root.
7. Build animated visual Merkle Tree Path visualizer (Leaf $\to$ Intermediate Nodes $\to$ Merkle Root) with instant green "Cryptographically Verified" or red "Proof Mismatch" validation badge.
8. Implement Depository Attestation & Auditor Certificate Viewer featuring signed PDF preview modals and direct download links.
9. In Admin Console: Build Real-Time Reconciliation Dashboard displaying ISIN balance comparisons, settlement lag indicators (T+1 timing adjustments), and break thresholds.
10. Build Admin Discrepancy Alert Queue flagging any variance $>0.000001$ units between custody holdings and on-chain token supply with automated escalation triggers.
11. Implement Break Investigation Drawer allowing operations officers to review pending DvP settlements, custodian sync logs, and attach resolution commentary.
12. Build Historical Reserve Chart displaying daily 1:1 backing ratios over 30d, 90d, and 1-year historical ranges.
13. Write comprehensive unit tests verifying client-side Merkle proof validation with synthetic tree datasets and Playwright E2E tests.

## Interfaces / Contracts
```typescript
import { type Address, type Hex } from 'viem';

export interface IsinReserveSummary {
  isin: string;
  ticker: string;
  companyName: string;
  depository: 'NSDL' | 'CDSL';
  physicalSharesInCustody: string; // e.g. "1500000.000000"
  digitalTokensMinted: string; // e.g. "1500000.000000"
  reserveRatioPercent: number; // e.g. 100.00
  status: 'FULLY_BACKED' | 'SURPLUS' | 'DEFICIT';
  lastAttestationTimestamp: string;
  onChainMerkleRoot: Hex;
  ipfsReportUri: string;
  contractAddress: Address;
}

export interface MerkleVerificationInput {
  accountId: string;
  tokenBalance: string;
  salt: Hex;
  isin: string;
  proof: Hex[];
}

export interface MerkleVerificationResult {
  isValid: boolean;
  computedLeaf: Hex;
  computedRoot: Hex;
  onChainRoot: Hex;
  matchedBlockNumber: number;
  attestationTimestamp: string;
  auditorSignatureVerified: boolean;
}

export interface AdminReconciliationBreak {
  breakId: string;
  isin: string;
  varianceAmount: string;
  reason: 'SETTLEMENT_TIMING_LAG' | 'CUSTODIAN_SYNC_DELAY' | 'UNRECONCILED_MINT';
  detectedAt: string;
  status: 'OPEN' | 'INVESTIGATING' | 'RESOLVED';
  assignedOfficer?: string;
}
```

## Security & Compliance Notes
- **Zero-Knowledge Privacy Guarantee:** Retail investor identity remains completely protected; verification uses salted cryptographic hashes so third parties cannot deduce individual balances.
- **Decentralized RPC Fallbacks:** Verification tool connects directly to multiple Hyperledger Besu consortium nodes to prevent centralized man-in-the-middle tampering.
- **SEBI Proof-of-Reserve Mandate:** Daily custodian attestations and cryptographic roots are preserved immutably for regulatory audits.
- **Anti-Scraping Protection:** Public proof verification endpoints implement rate-limiting to prevent automated brute-force leaf enumeration.

## Acceptance Criteria
- [ ] Public proof-of-reserve portal renders live aggregate reserve gauge showing 1:1 custody backing.
- [ ] ISIN breakdown table lists all active securities with custody shares, token supply, and depository identifiers.
- [ ] Client-side Merkle proof verifier accurately calculates leaf hash from inputs and validates proof against on-chain root in `ProofOfReserveRegistry.sol`.
- [ ] Merkle path animation visually highlights the exact proof branch from leaf to root.
- [ ] Auditor attestation documents and custodian certificates download properly with valid digital signatures.
- [ ] Admin reconciliation workstation highlights breaks and triggers operational investigation workflows.
- [ ] Unit tests for Merkle proof verification algorithm achieve 100% test coverage with standard and edge test vectors.
- [ ] Lighthouse performance score >95 for public transparency portal.

## Suggested Order / Dependencies
- **Prerequisites:** 007 (PoR Public Disclosure), 213 (Custodian Adapter), 215 (Reconciliation Service), 308 (PoR Smart Contract), 601 (Web Scaffolding).
- **Direct Successors / Parallel:** 605 (Multi-Party Approval UI), 607 (Regulatory Reporting), 608 (Marketing Site).
