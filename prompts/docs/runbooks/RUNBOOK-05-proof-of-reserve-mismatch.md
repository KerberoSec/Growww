# Runbook 05: Proof of Reserve Mismatch Remediation

**Runbook ID:** RBK-OPS-005  
**Severity Tier:** P1 (Critical)  
**Authority:** Custody Operations Officer / Chief Risk Officer  

---

## 1. Alert Description
Triggered when the automated daily cryptographic Proof of Reserve (PoR) Merkle root verification fails:
$$\text{TotalOnChainSupply}(a) > \text{AttestedDepositoryBalance}(a) + \varepsilon_{\text{dust}}$$

---

## 2. Remediation Protocol

```
[PoR Mismatch Alert]
       |
       v
[Automate Immediate Asset Trading Halt]
       |
       v
[Block Minting & Burning for Asset]
       |
       v
[Fetch Fresh Direct Custodian API Statement]
       |
       +--(If Custodian Balance Validates Reserve)--> [Update Oracle Attestation] -> [Resume Trading via Call Auction]
       |
       +--(If True Shortfall Confirmed)
       |
       v
[Notify Statutory Custody Board & SEBI Surveillance]
       |
       v
[Drawdown Settlement Guarantee Fund / Insurance to Settle Shortfall]
```

All actions and forensic logs are sealed into an immutable audit package and submitted to regulatory trustees.
