# Runbook 24: Perpetual Funding Rate Anomaly & Emergency Clamping

**Runbook ID:** RBK-OPS-024  
**Severity Tier:** P1 (Critical)  
**Authority:** Risk Operations Lead / Derivatives Systems Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- An external index price feed provider reports stale data (> 30s) or erratic pricing ($|P_i - \text{Median}| > 3.0\%$).
- Calculated 8-hour funding rate breaches the hard safety clamp ($\pm 0.75\%$).

---

## 2. Remediation Workflow

```
[1. Identify Affected Contract & Stalled Index Feed]
  - Query metrics: derivatives_index_price_feed_staleness, derivatives_funding_rate_clamp_events
  - Identify outlier exchange feed (e.g., Feed 3 reporting +5% spread)
                  |
                  v
[2. Isolate Outlier Feed & Re-Calculate Index]
  - Disable corrupted feed provider in index-calculator engine:
      POST /api/v1/derivatives/admin/disable-feed-source
      {"source": "EXCHANGE_3", "contract": "BTC-PERP"}
  - Index engine falls back to median of remaining healthy feeds
                  |
                  v
[3. Enforce Clamped Funding Rate Calculation]
  - Re-run funding rate evaluation with clamped bounds: F = clamp(F_calc, -0.75%, +0.75%)
  - Verify funding payment settlement ledger batch before 00:00/08:00/16:00 UTC execution
                  |
                  v
[4. Audit Logging & Resolution]
  - Log automated failover event in compliance audit trail
  - Restore feed provider once external API confirms data normalization
```
