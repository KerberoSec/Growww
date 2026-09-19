# ADR-0036: Multi-Asset Cross-Margin & Tiered Leverage Risk Brackets

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Head of Derivatives Trading  

---

## 1. Context
Traders holding diverse portfolios (Tokenized G-Secs, Equities, Crypto) require capital efficiency to use multi-asset balances as shared collateral, while the exchange requires tiered leverage limits to avoid catastrophic liquidation market impact on large block positions.

---

## 2. Decision
1. **Multi-Asset Cross-Margin Collateral Pool:** Enables cross-margining across G-Secs (5% haircut), Blue-chip Equities (20% haircut), and BTC/ETH/USDT (20% haircut).
2. **Tiered Position Leverage Limits:** Implements 4 risk brackets reducing maximum leverage dynamically from 20x for positions $< \text{₹50 Lakhs}$ down to 2x for large positions $> \text{₹10 Crores}$.

---

## 3. Consequences
- **Positive:** Maximum capital efficiency for diversified traders; reduces liquidation cascades on large positions.
- **Trade-offs:** Requires continuous real-time Mark-to-Market collateral revaluation.
