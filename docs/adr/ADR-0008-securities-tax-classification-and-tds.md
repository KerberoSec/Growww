# ADR-0008: Zero-Fee Starting Policy, Dynamic Fee Controller & Decoupled Tax Architecture

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Technology Officer, Head of Financial Architecture, Lead Blockchain Architect  

---

## 1. Context
Launching a high-throughput Web3 spot crypto and RWA exchange requires frictionless user onboarding and deep liquidity. Starting with non-zero fees or mandatory on-chain tax withholding creates execution drag, high gas costs, and barrier to adoption. However, long-term sustainability may require the platform to introduce or adjust fees in the future.

---

## 2. Decision
1. **0.00% Starting Fee Policy ("0 Means 0 in All")**:
   - Initial trading fees set to strictly 0.00% for all orders (both Maker and Taker).
   - Zero gas fees for users via ERC-4337 Paymaster sponsorship.
   - Zero on-chain TDS (0% tax withholding).
2. **Dynamic Future Fee Controller**:
   - Implement `FeeController.sol` on Hyperledger Besu.
   - Fee adjustments require a mandatory **48-Hour Timelock** and 3-of-5 multi-sig authorization.
   - Enforce an immutable safety cap: `MAX_FEE_CEILING = 50 bps` (0.50%).
   - Dynamic parameters propagate to off-chain matching and settlement engines via Kafka configuration streams without service restarts.
3. **Decoupled Tax Architecture**:
   - Clean on-chain DvP settlement with optional off-chain voluntary reporting statements (CSV/JSON/PDF).

---

## 3. Consequences
- **Positive:** Maximum initial adoption and zero trading friction; full future flexibility to adjust fees under strict governance; 35% lower smart contract gas consumption.
