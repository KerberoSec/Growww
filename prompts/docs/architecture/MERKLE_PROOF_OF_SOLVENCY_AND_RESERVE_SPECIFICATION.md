# Cryptographic Merkle-Tree Proof of Solvency & Reserve Verification Specification

This document defines the cryptographic protocols, zk-SNARK proof systems, blinded Merkle sum trees, and self-verification workflows for exchange-wide Proof of Solvency (Proof of Reserves + Proof of Liabilities) on the Growww RWA Exchange.

---

## 1. Mathematical Solvency Guarantee

An exchange is cryptographically solvent if and only if its total verified on-chain assets meet or exceed its total client liabilities:

$$\sum_{i=1}^N \text{AssetReserves}_i \ge \sum_{u=1}^M \text{UserLiabilities}_u, \quad \forall \text{Asset } i$$

---

## 2. zk-SNARK Blinded Merkle Sum Tree (Proof of Liabilities)

### 2.1 Merkle Sum Tree Architecture
Traditional Merkle trees prove inclusion but leak user account balances. Growww implements a **Blinded Merkle Sum Tree** with Poseidon hashing and zero-knowledge validity proofs:
1. **Leaf Node Construction ($L_u$):**
   $$L_u = H_{Poseidon}\left(\text{UserUUID}_u \parallel \text{UserBlindingNonce}_u \parallel \text{Balances}_u\right)$$
   $$\text{SumValue}(L_u) = \text{Balances}_u$$
2. **Intermediate Parent Node ($N_{parent}$):**
   $$N_{parent} = H_{Poseidon}\left(N_{left} \parallel \text{SumValue}(N_{left}) \parallel N_{right} \parallel \text{SumValue}(N_{right})\right)$$
   $$\text{SumValue}(N_{parent}) = \text{SumValue}(N_{left}) + \text{SumValue}(N_{right})$$
3. **Root Node ($R_{solvency}$):**
   Contains the cryptographic commitment of all user balances and the total aggregated exchange liabilities $\Sigma \text{Liabilities}$.

### 2.2 zk-SNARK Constraint Invariants
A Groth16 zk-SNARK circuit proves:
- Every leaf balance is non-negative: $\text{Balance}_u \ge 0$ (prevents fake negative balances from masking deficits).
- The summation of child nodes equals parent nodes at every level: $\text{SumValue}(N_{parent}) \equiv \text{SumValue}(N_{left}) + \text{SumValue}(N_{right})$.
- The total liabilities match the published Merkle Root $\Sigma \text{Liabilities}$.

---

## 3. Proof of Assets & On-Chain Signature Verification

### 3.1 Cold Vault Address Ownership Verification
1. **Address Declaration:** The exchange publishes all Cold Storage Vault addresses across Bitcoin, Ethereum, Solana, and Hyperledger Besu.
2. **Cryptographic Proof of Control:**
   - For EVM addresses: EIP-1271 / EIP-712 digital signatures over a challenge string containing the current block hash and timestamp:
     $$\text{Sign}\left(\text{"Growww-Solvency-Proof:"} \parallel \text{BlockNumber} \parallel \text{Timestamp} \parallel R_{solvency}\right)$$
   - For Bitcoin addresses: BIP-322 signature proofs.
   - For Solana: Ed25519 off-chain signature challenge.

---

## 4. User Self-Verification Protocol

Every registered user can cryptographically verify their inclusion in the solvency root directly from the web/mobile app:
1. **Fetch Proof:** User queries `GET /api/v1/audit/proof-of-solvency/my-proof`.
2. **Client-Side Verification:** The client app downloads the Merkle sibling path $(\pi_1, \pi_2, \dots, \pi_k)$ and verifies:
   $$\text{RootComputed} = \text{VerifyPath}(L_u, \pi, \text{PathIndices}) == R_{solvency}$$
3. **Tamper Detection:** If $\text{RootComputed} \ne R_{solvency}$, the client flags an immediate cryptographic fraud break (RUNBOOK-26).
