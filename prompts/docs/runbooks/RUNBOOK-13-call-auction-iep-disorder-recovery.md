# Runbook 13: Call Auction IEP Equilibrium Discovery & Anomaly Recovery

**Runbook ID:** RBK-OPS-013  
**Severity Tier:** P1 (Critical)  
**Authority:** Market Operations Officer / Matching Engine SRE  

---

## 1. Description & Trigger Conditions
Triggered when the automated Indicative Equilibrium Price (IEP) discovery algorithm in a circuit-halted call auction fails to converge or exhibits anomalous price volatility exceeding $\pm 10\%$ of previous close.

---

## 2. Recovery Workflow

```
[1. Evaluate Fenwick Tree Demand Curves]
  - Inspect cumulative Bid and Ask arrays via Matching Engine Debug RPC
  - Verify tradable volume intersection point P_IEP
                  |
                  v
[2. If Zero Tradable Intersection Exists]
  - Transition instrument state to CALL_AUCTION_EXTENDED (+5 minutes)
  - Broadcast order imbalance metrics to market data feed
                  |
                  v
[3. If Severe Anomaly Detected (Fat-Finger Quote Injection)]
  - Execute four-eyes administrative quote purge on erroneous orders
  - Recalculate IEP using remaining valid orders
                  |
                  v
[4. Finalize Matching & Transition to Continuous]
  - Execute all matched orders at single uniform clearing price P_IEP
  - Transition session state to CONTINUOUS_TRADING
```
