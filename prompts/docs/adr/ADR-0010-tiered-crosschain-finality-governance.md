# ADR-0010: Tiered Cross-Chain Finality Governance & Reorg Buffer

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Lead Gateway Architect  

---

## 1. Context
External blockchains have heterogeneous finality characteristics (Bitcoin 6-block PoW, Ethereum 2-epoch Casper, Solana 32-slot confirmation). Crediting unfinalized deposits onto the instant Besu ledger risks platform insolvency during deep external chain reorganizations.

---

## 2. Decision
Enforce a **Strict Tiered Finality Confirmation Matrix** at the GIFT City funding gateway:
- **Bitcoin (BTC):** 6 on-chain confirmations (~60 minutes) before purchasing power credit.
- **Ethereum (ETH / USDC):** 64 blocks (2 finalized epochs, ~12.8 minutes).
- **Solana (SOL / USDC):** 64 finalized commitment slots (~25 seconds).
- **Instant Credit Buffer:** Institutional participants can obtain instant execution credit only by staking pre-funded margin in the SGF Escrow Pool.

---

## 3. Consequences
- **Positive:** Mathematically prevents unbacked token mints caused by external blockchain reorganizations.
- **Trade-offs:** Introduces predictable deposit confirmation latency for non-staked external deposits.
