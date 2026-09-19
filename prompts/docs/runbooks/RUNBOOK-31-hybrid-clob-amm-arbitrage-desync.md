# Runbook 31: Hybrid CLOB-AMM Arbitrage Desynchronization Remediation

**Runbook ID:** RBK-OPS-031  
**Severity Tier:** P2 (High)  
**Authority:** Market Structure Lead / Lead Quantitative Engineer  

---

## 1. Description & Trigger Conditions
Triggered when:
- Price divergence between the Central Limit Order Book (CLOB) and the Concentrated Liquidity AMM pool exceeds 0.25% for more than 30 seconds.
- Internal JIT arbitrage relayer fails to execute rebalancing swaps.

---

## 2. Remediation Workflow

```
[1. Identify Divergent Token Pair]
  - Query metrics: hybrid_routing_clob_amm_price_spread, jit_arbitrage_execution_errors
  - Identify asset ISIN and affected AMM pool address
                  |
                  v
[2. Evaluate AMM Pool Virtual Reserves & Gas Relayer]
  - Inspect AMM tick liquidity: GET /api/v1/amm/pools/{id}/ticks
  - Check JIT relayer funded gas balance on Hyperledger Besu
  - If relayer gas exhausted, trigger automated treasury top-up
                  |
                  v
[3. Trigger Manual Equilibrium Rebalance]
  - Trigger administrative rebalance sweep:
      POST /api/v1/amm/admin/trigger-rebalance
      {"pool_id": "<POOL_ID>", "max_slippage": "0.001"}
  - Swap excess inventory between CLOB and AMM pool; capture spread profit to SGF
                  |
                  v
[4. Verify Smart Order Router Normalization]
  - Confirm SOR resumes optimal multi-venue split execution
  - Log arbitrage capture metrics in SGF treasury ledger
```
