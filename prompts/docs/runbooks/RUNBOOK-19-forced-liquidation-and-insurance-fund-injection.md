# Runbook 19: Forced Liquidation & Insurance Fund Deficit Injection

**Runbook ID:** RBK-OPS-019  
**Severity Tier:** P2 (High)  
**Authority:** Risk Operations Lead / Clearing Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- An under-margined account drops below the Maintenance Margin Requirement (MMR) and enters the automated forced liquidation cycle.
- A liquidated position incurs a bankruptcy deficit ($\text{Account Equity} < 0$), requiring Insurance Fund / SGF capital injection.

---

## 2. Remediation Workflow

```
[1. Detect MMR Breach & Order Cancellation]
  - Risk Engine flags account when Margin Ratio (MR) < Maintenance Margin Requirement (MMR)
  - Cancel all open unexecuted limit orders to release collateral reservations
                  |
                  v
[2. Execute Automated Forced Liquidation]
  - Liquidation Engine takes over position and places aggressive IOC/FOK market orders
  - Levy flat 1.0% Forced Liquidation Fee on liquidated notional:
      Liquidation_Fee = Liquidated_Notional * 0.0100
  - Route 80% of fee to SGF / Insurance Fund, 20% to Treasury Relayer Pool
                  |
                  v
[3. Check Account Equity & Deficit Resolution]
  - If Remaining Equity >= 0:
      - Credit remaining balance to user fiat/token wallet; close liquidation cycle
  - If Remaining Equity < 0 (Bankruptcy Deficit):
      - Transfer exact deficit amount from SGF / Insurance Fund to clear negative balance
      - Post atomic double-entry journal entry:
          Debit: Account 3200 (SGF Insurance Fund)
          Credit: Account 1100 (Client Trading Balance)
                  |
                  v
[4. Telemetry & Regulatory Notification]
  - Publish LIQUIDATION_COMPLETED event on Kafka topic: growww.risk.liquidation.v1
  - Log audit entry on Hyperledger Besu with transaction hash and deficit metrics
```
