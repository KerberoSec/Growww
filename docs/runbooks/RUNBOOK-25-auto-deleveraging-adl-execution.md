# Runbook 25: Auto-Deleveraging (ADL) Protocol Execution & Triage

**Runbook ID:** RBK-OPS-025  
**Severity Tier:** P1 (Critical)  
**Authority:** Chief Risk Officer / Derivatives Clearing Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- Extreme market gap or illiquid book causes a liquidated account's bankruptcy deficit to exceed the total available Insurance Fund / SGF capital balance.
- Automated ADL engine initiates emergency counterparty position de-leveraging.

---

## 2. Remediation Workflow

```
[1. Confirm Insurance Fund Depletion & ADL Activation]
  - Risk engine detects: Insurance_Fund_Balance == 0 AND Bankruptcy_Deficit > 0
  - Matching engine pauses new order placements on affected contract for 30 seconds
                  |
                  v
[2. Rank Counterparty Positions by ADL Priority]
  - Calculate ADL Score for all profitable opposing positions:
      ADL_Score = ProfitRankingPercentile * EffectiveLeverage
  - Sort queue in descending order of ADL Score
                  |
                  v
[3. Execute Emergency Deleveraging Matches]
  - Match bankrupt position against highest-ranked ADL counterparties at Bankruptcy Price
  - Emit ADL_POSITION_CLOSED execution reports to affected traders with code ERR_ADL_EXECUTED
                  |
                  v
[4. SGF Capital Top-Up & Session Resumption]
  - CRO authorizes emergency capital injection into Insurance Fund from Corporate Treasury
  - Resume continuous trading on contract; notify regulatory authorities
```
