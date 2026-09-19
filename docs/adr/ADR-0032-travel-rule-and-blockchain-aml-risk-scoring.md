# ADR-0032: FATF Travel Rule Protocol & Automated Blockchain AML Risk Scoring

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Compliance Officer, Money Laundering Reporting Officer (MLRO), Lead Security Engineer  

---

## 1. Context
To operate in full compliance with Financial Action Task Force (FATF) Recommendation 16, FIU-IND directives, and global Virtual Asset Service Provider (VASP) mandates, external deposits and withdrawals must be screened for financial crime risk, sanctioned entities, mixers, and originator/beneficiary identification.

---

## 2. Decision
1. **Pre-Deposit & Pre-Withdrawal AML Risk Scoring:**
   - Real-time querying against blockchain analytics engines (Chainalysis KYT / TRM Labs / Elliptic).
   - Direct deposits from sanctioned entities (OFAC, UN, FIU-IND), mixers (e.g., Tornado Cash), ransomware addresses, or high-risk darknet markets are immediately quarantined into a restricted holding ledger and flagged for Suspicious Transaction Reporting (STR).
2. **FATF Travel Rule Integration:**
   - For all incoming and outgoing crypto transfers exceeding statutory thresholds ($> \$1,000\text{ USD}$ or equivalent INR ₹83,000), the platform exchanges cryptographic Travel Rule data packets via the TRISA / Notabene / Sygna interoperability networks.
   - PII payloads (Originator Name, Physical Address, Account ID, Beneficiary Name) are encrypted point-to-point using recipient VASP public keys.
3. **Unhosted (Self-Custody) Wallet Verification:**
   - For transfers to/from non-VASP unhosted wallets, proof of ownership is validated via Satoshi micro-transfers or EIP-712 / BIP-137 cryptographic signature challenges.

---

## 3. Consequences
- **Positive:** Protects platform against regulatory sanction exposure; prevents illicit asset contamination.
- **Trade-offs:** Outgoing withdrawals to other exchanges require automated VASP discovery and IVMS-101 payload negotiation.
