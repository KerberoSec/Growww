# Runbook 22: Travel Rule Quarantine & AML Suspicious Deposit Resolution

**Runbook ID:** RBK-OPS-022  
**Severity Tier:** P2 (High)  
**Authority:** MLRO (Money Laundering Reporting Officer) / Compliance Operations Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- An incoming external crypto deposit is flagged by Chainalysis / TRM Labs with high risk score ($> 7.5/10$), darknet connection, mixer interaction (e.g. Tornado Cash), or sanctioned address link (OFAC / UN / FIU-IND).
- An incoming transfer $> \$1,000\text{ USD}$ is missing required FATF Travel Rule originator information from the sending VASP.

---

## 2. Remediation Workflow

```
[1. Automated Ledger Quarantine & Ingestion Isolation]
  - Tag deposit transaction as STATUS_QUARANTINED
  - Credit virtual Quarantine Ledger (Account 2900 - Quarantined Crypto Escrow)
  - Prevent funds from entering user trading or withdrawal balance
                  |
                  v
[2. Evaluate Alert Type & Compliance Triage]
  - Case A (Sanctions / Darknet / Mixer):
      - Lock associated user account profile immediately
      - Generate Suspicious Transaction Report (STR) / Suspicious Activity Report (SAR)
      - File filing with FIU-IND / relevant national financial intelligence unit
      - Maintain permanent asset freeze under FreezeManager protocol
  - Case B (Missing Travel Rule Data):
      - Send automated Travel Rule Request message to Originating VASP via TRISA/Notabene
      - Prompt user in mobile/web UI to declare originator details or provide unhosted proof
                  |
                  v
[3. Resolution & Decision Execution]
  - Option 1 (Travel Rule Verified):
      - Clear quarantine hold; release funds to user active trading balance
  - Option 2 (Returned to Originator):
      - If rejected by compliance, execute automated Return-to-Sender transfer
      - Subtract network miner gas fee; refund net amount to verified origin address
                  |
                  v
[4. Audit Trail & Case Closure]
  - Record compliance officer sign-off and rationale in immutable audit database
  - Archive case record in compliance management portal
```
