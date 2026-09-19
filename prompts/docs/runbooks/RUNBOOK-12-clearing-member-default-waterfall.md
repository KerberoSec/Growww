# Runbook 12: Clearing Member Default & SGF Waterfall Execution

**Runbook ID:** RBK-OPS-012  
**Severity Tier:** P1 (Critical)  
**Authority:** Chief Risk Officer / Clearing Corporation Managing Director  

---

## 1. Trigger Conditions
A clearing member fails to meet daily cash margin settlement obligations by 16:30 IST and deficit exceeds available intraday credit.

---

## 2. Waterfall Execution Sequence

```
[1. Declare Member Default]
  - Issue formal default declaration under Clearing Corporation Bye-Laws
  - Suspend member trading terminals across all segments in < 1 second
                  |
                  v
[2. Liquidate Defaulter Collateral]
  - Seize and auction member's pledged securities and bank fixed deposits
  - Apply proceeds to settle outstanding clearing obligations
                  |
                  v
[3. Drawdown Defaulter Core SGF Contribution]
  - Deduct member's deposited corpus from the Settlement Guarantee Fund
                  |
                  v
[4. Drawdown Exchange First-Loss Capital]
  - Transfer up to 25% of retained platform earnings to cover remaining deficit
                  |
                  v
[5. Drawdown Core SGF Corpus & Pro-Rata Assessment]
  - Execute multi-sig transaction on Hyperledger Besu to release SGF reserves
  - If required, assess non-defaulting clearing members pro-rata (capped at 100% margin)
```

File formal regulatory default disclosure with SEBI within 2 hours.
