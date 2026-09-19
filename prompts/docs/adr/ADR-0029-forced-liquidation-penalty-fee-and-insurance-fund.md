# ADR-0029: Forced Liquidation 1.0% Penalty Fee & Insurance Fund Capitalization

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Head of Trading, Principal Financial Architect  

---

## 1. Context
Standard trading carries a single 0.00% (No fee at all) all-inclusive platform fee with zero brokerage. However, when an under-collateralized position breaches its Maintenance Margin Requirement (MMR) and requires automated forced liquidation by the exchange matching engine and liquidation engine, the platform absorbs execution market impact, latency slippage, and deficit bankruptcy risk. Without a liquidation penalty fee, traders are disincentivized from closing distressed positions voluntarily before insolvency.

---

## 2. Decision
1. **Flat 1.0% Forced Liquidation Fee:** 
   - A flat **1.0% (100 bps / 0.01)** fee is levied strictly on the liquidated notional value when a position is forcefully closed or taken over by the Exchange Liquidation Engine.
   - Normal voluntary trades placed by users continue to pay strictly the standard flat **0.00% (Zero Fee)** platform fee.
2. **Mathematical Formulation:**
   $$\text{Fee}_{liq} = \text{round}\left(\text{Liquidated Notional} \times 0.0100\right)$$
3. **Liquidation Fee Capital Allocation:**
   - **80% to Insurance Fund / SGF (Settlement Guarantee Fund):** Directly capitalizes the exchange backstop reserve to absorb negative balance deficit bankruptcies.
   - **20% to Treasury Liquidation Relayer Pool:** Covers gas sponsorship and compute overhead for the liquidation execution bots.
4. **Bankruptcy Protocol:** If equity after liquidation is negative ($\text{Account Equity} < 0$), the remaining deficit is absorbed 100% by the Insurance Fund / SGF without clawback from non-defaulting traders.

---

## 3. Consequences
- **Positive:** Protects the exchange solvency; incentivizes timely user risk management; mirrors institutional derivatives and crypto exchange standards (e.g., Binance Insurance Fund model).
- **Trade-offs:** Distressed accounts lose an additional 1.0% penalty buffer upon breach of MMR.
