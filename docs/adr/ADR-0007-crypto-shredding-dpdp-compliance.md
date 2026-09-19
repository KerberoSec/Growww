# ADR-0007: Crypto-Shredding for DPDP Act 2023 and GDPR Compliance

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Data Protection Officer, Chief Information Security Officer  

---

## 1. Context
The Digital Personal Data Protection Act (DPDP Act 2023) and GDPR grant data principals the right to erasure of personal data. However, immutable append-only ledgers and distributed event logs cannot delete historical records without breaking cryptographic hash chains.

---

## 2. Decision
Adopt **Crypto-Shredding** combined with zero on-chain Personally Identifiable Information (PII).
- No low-entropy PII (Aadhaar, PAN, phone number, physical address) is ever written to the blockchain or Kafka topics, even in hashed or salted form.
- Off-chain PII records are stored encrypted in dedicated PostgreSQL tables using unique per-user Data Encryption Keys (DEK) managed in an HSM/KMS.
- Fulfilling a statutory right-to-erasure request permanently destroys the subject's specific DEK in the HSM, rendering the off-chain ciphertext permanently unrecoverable while leaving immutable distributed logs cryptographically intact.

---

## 3. Alternatives Considered
- **Direct PII Hashing on Ledger:** Rejected because low-entropy identifiers (10-digit PAN, 12-digit Aadhaar) are easily brute-forced via GPU hash tables, violating DPDP anonymization requirements.
- **Mutable Distributed Ledgers:** Violates blockchain immutability and regulatory audit integrity.

---

## 4. Consequences
- **Positive:** Reconciles the statutory right-to-erasure with permanent immutable audit logs; guarantees privacy against rainbow table attacks.
- **Trade-offs:** Requires robust Key Management Service (KMS) governance and key lifecycle monitoring.
