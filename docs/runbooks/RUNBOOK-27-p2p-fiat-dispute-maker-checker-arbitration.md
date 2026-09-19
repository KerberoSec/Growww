# Runbook 27: P2P Fiat Dispute Maker-Checker Arbitration

**Runbook ID:** RBK-OPS-027  
**Severity Tier:** P2 (High)  
**Authority:** P2P Dispute Operations Lead / MLRO  

---

## 1. Description & Trigger Conditions
Triggered when:
- Either buyer or seller opens a dispute in a P2P fiat-crypto trade due to claimed non-payment, incorrect amount, or payment account name mismatch.

---

## 2. Remediation Workflow

```
[1. Lock Escrow & Review Evidence]
  - Escrow status locks in STATUS_IN_DISPUTE (Account 2800)
  - Ingest Buyer proof: Bank statement PDF, UTR reference, payment screenshot
  - Ingest Seller proof: Bank statement showing uncredited status
                  |
                  v
[2. Automated Bank UTR Verification]
  - Query banking API / Open Banking aggregator for UTR clearance
  - Verify payer name matches buyer KYC record
                  |
                  v
[3. Maker-Checker Arbitration Decision]
  - Dispute Officer 1 (Maker): Evaluates evidence and recommends "RELEASE_TO_BUYER" or "REFUND_TO_SELLER"
  - Dispute Supervisor 2 (Checker): Reviews evidence independently and approves release
                  |
                  v
[4. Execute Escrow Settlement & Close Ticket]
  - If approved release to Buyer: Crypto transferred to Buyer; Seller penalized
  - If approved refund to Seller: Crypto returned to Seller; Buyer account suspended
  - Record audit trail with signed officer IDs and evidence checksums
```
