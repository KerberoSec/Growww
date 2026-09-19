# ADR-0001: Hyperledger Besu with QBFT Consensus for Settlement Ledger

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Blockchain Architect, VP Engineering, Chief Compliance Officer  

---

## 1. Context
The Growww institutional capital platform requires a permissioned distributed ledger to orchestrate atomic Delivery-versus-Payment (DvP) settlement for tokenized real-world assets. The ledger must provide immediate deterministic finality (zero reorganizations), enterprise privacy, EVM compatibility, and high transaction throughput under regulatory scrutiny (SEBI and IFSCA).

---

## 2. Decision
Adopt **Hyperledger Besu** configured with **Quorum Byzantine Fault Tolerance (QBFT)** consensus as the primary settlement blockchain.

- **Block Period:** 2.0 seconds
- **Round Change Timeout:** 8.0 seconds
- **EVM Target:** London hardfork (`londonBlock: 0`, zero base fee)
- **Validator Topology:** Multi-region consortium nodes across Mumbai, GIFT City, and regulatory trust nodes.

---

## 3. Alternatives Considered

| Option | Pros | Cons | Reason for Rejection |
|---|---|---|---|
| **Public Ethereum / L2s** | Large ecosystem | Gas fee volatility, public PII leakage risks, non-deterministic finality | Non-compliant with Indian regulatory data residency and privacy mandates |
| **R3 Corda** | Strong enterprise banking adoption | Non-EVM, proprietary development tooling, complex smart contract model | Lacks compatibility with standardized ERC-3643 security token ecosystems |
| **Hyperledger Fabric** | Granular channel privacy | Endorsement latency, non-standard smart contract lifecycle | Inflexible DvP composability across multiple institutional counterparties |

---

## 4. Consequences
- **Positive:** Deterministic single-block finality with zero chain reorganizations; native compatibility with standard Solidity tooling (Foundry, Hardhat, OpenZeppelin).
- **Trade-offs:** Requires running dedicated validator infrastructure; Poseidon hashing requires careful gas budget management.

---

## 5. Revisit Triggers
- Sustained peak DvP settlement load exceeds 2,500 TPS requiring L2 rollup scaling.
- SEBI or IFSCA issues statutory mandates requiring a specific sovereign ledger standard.
