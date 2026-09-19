# ADR-0040: Crypto Earn Yield Vaults & PoS Validator Slashing Insurance

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Treasury, Lead Staking Architect, Chief Risk Officer  

---

## 1. Context
Users holding idle crypto and fiat balances seek yield generation, while Proof-of-Stake validator delegation carries technical risks (slashing penalties, unbonding lockups).

---

## 2. Decision
1. **Flexible Savings Vaults:** Deploy idle USDT/INR in sovereign repo and institutional lending, paying daily auto-compounding interest with instant zero-penalty redemption.
2. **Validator Slashing Insurance:** Liquid staking for ETH and SOL delegates to institutional-grade validators backed by an internal Slashing Insurance Fund guaranteeing 100% principal protection.
3. **Instant Exit Liquidity Pools:** Provide instant unbonding exit routes for staked assets at a flat 0.05% swap fee.

---

## 3. Consequences
- **Positive:** Attracts long-term capital deposits; eliminates staking lockup friction for retail users.
- **Trade-offs:** Requires liquidity buffering in instant exit pools.
