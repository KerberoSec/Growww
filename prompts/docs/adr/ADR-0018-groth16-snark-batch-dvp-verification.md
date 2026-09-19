# ADR-0018: Groth16 Snark Batch Verification on Native alt_bn128 Precompiles

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Cryptographer, Smart Contracts Architect  

---

## 1. Context
Computing Poseidon hashes in Solidity bytecode costs $> 65,000\text{ gas}$ per hash, causing multi-trade batch DvP settlements to exceed block gas limits ($>3.2\text{M gas}$).

---

## 2. Decision
Verify batch identity and balance proofs off-chain, generating a single **Groth16 Snark Proof** verified on-chain via the native `alt_bn128` pairing precompile at address `0x08`.
- Verifies 50 to 100 trade identity commitments in constant $O(1)$ time.
- Gas consumption is capped at a flat $\approx 180,000\text{ gas}$ per batch.

---

## 3. Consequences
- **Positive:** Lowers verification gas overhead by $94\%$; scales batch settlement throughput.
- **Trade-offs:** Requires off-chain Snark proof generation infrastructure.
