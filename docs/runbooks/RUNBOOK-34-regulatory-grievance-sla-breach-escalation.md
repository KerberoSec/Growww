# Runbook 34: Regulatory Grievance SLA Breach Escalation & SCORES ATR Submission

**Runbook ID:** RBK-OPS-034  
**Severity Tier:** P1 (Critical)  
**Authority:** Principal Compliance Officer / Legal Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- An investor grievance ticket in Level 2 or Level 3 approaches statutory SLA limits ($> 15\text{ days}$ for SEBI SCORES 2.0 or $> 20\text{ days}$ for RBI Ombudsman).
- SEBI SCORES 2.0 API issues an automated Action Taken Report (ATR) reminder.

---

## 2. Remediation Workflow

```
[1. Lock & Prioritize Grievance Dossier]
  - Tag docket as SEVERITY_STATUTORY_URGENT in compliance management portal
  - Assign dedicated senior compliance counsel to dossier
                  |
                  v
[2. Aggregate Cryptographic Audit Trail]
  - Extract all relevant records:
      - Double-entry ledger journal entries with immutable block hashes
      - Matching engine execution reports and cancellation receipts
      - Bank UTR payment logs and KYC verification dossiers
                  |
                  v
[3. Compile Action Taken Report (ATR)]
  - Generate digitally signed PDF ATR with compliance officer digital signature (DSC)
  - Submit ATR directly via SEBI SCORES 2.0 API:
      POST /api/v1/compliance/scores/submit-atr
      {"complaint_id": "<SCORES_ID>", "atr_payload": "<SIGNED_ATR_PDF>"}
                  |
                  v
[4. Direct Investor Communication & Closure]
  - Dispatch formal resolution letter to complainant via registered email and DLT SMS
  - Archive full case audit log for 8-year statutory regulatory retention requirement
```
