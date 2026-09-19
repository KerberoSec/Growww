# ADR-0037: zk-SNARK Blinded Merkle Sum Tree Proof of Solvency

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Technology Officer, Lead Cryptographer, Head of Compliance  

---

## 1. Context
Standard proof of reserve audits either leak individual client balances or fail to prove that total exchange liabilities do not exceed verified on-chain assets.

---

## 2. Decision
Implement a **zk-SNARK Blinded Merkle Sum Tree** using Poseidon hashes and Groth16 zero-knowledge proofs:
1. Every user balance is cryptographically blinded with a private nonce in the leaf nodes.
2. zk-SNARK circuit proves zero negative balances and exact mathematical summation of all account liabilities.
3. Cold storage vault address ownership is proven on-chain via cryptographic digital signatures.
4. Users can self-verify inclusion in the published solvency root in $O(\log N)$ time.

---

## 3. Consequences
- **Positive:** 100% mathematical solvency verification without revealing confidential user account balances.
- **Trade-offs:** Requires generating SNARK proofs during daily solvency snapshots.
