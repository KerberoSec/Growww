# Runbook 29: Options IV Surface Discontinuity & Portfolio Margin Triage

**Runbook ID:** RBK-OPS-029  
**Severity Tier:** P1 (Critical)  
**Authority:** Derivatives Risk Lead / Chief Risk Officer  

---

## 1. Description & Trigger Conditions
Triggered when:
- SABR volatility calibration fails or produces negative local variances.
- An anomalous implied volatility spike causes false margin calls across options traders.

---

## 2. Remediation Workflow

```
[1. Isolate Affected Expiry Cycle & Strikes]
  - Query metrics: options_iv_surface_fit_errors, options_portfolio_margin_spike_total
  - Identify affected underlying (e.g. BTC-OPTIONS) and expiry date
                  |
                  v
[2. Switch to Fallback Volatility Surface]
  - Disable active SABR model; fall back to 30-minute historical IV surface:
      POST /api/v1/derivatives/admin/set-iv-fallback
      {"underlying": "BTC", "model": "HISTORICAL_30M"}
                  |
                  v
[3. Re-Evaluate Portfolio Margin Requirements]
  - Trigger batch portfolio re-evaluation for all open options positions
  - Suppress false margin liquidation triggers while fallback surface is active
                  |
                  v
[4. Re-Calibrate SABR Parameters & Restore]
  - Fit SABR parameters with cleaned strike data; restore automated real-time calibration
  - Log root-cause report in risk operations console
```
