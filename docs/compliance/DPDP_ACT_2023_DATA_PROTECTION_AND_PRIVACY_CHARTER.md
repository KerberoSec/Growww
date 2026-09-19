# DPDP Act 2023 Data Protection, Privacy & Zero-PII Ledger Charter

## 1. Statutory Scope & Legal Classification

- **Governing Law:** Digital Personal Data Protection Act 2023 (DPDP Act 2023), Gazette of India.
- **Classification:** Significant Data Fiduciary (SDF) under Section 10 of the DPDP Act 2023.
- **Regulatory Oversight:** Data Protection Board of India (DPBI).
- **Data Protection Officer (DPO):** Resident Indian officer appointed to oversee compliance, conduct periodic Data Protection Impact Assessments (DPIAs), and manage investor privacy grievances.

---

## 2. The Zero-PII Blockchain Ledger Invariant

A fundamental tenet of the Growww / NBSE enterprise blockchain architecture is the absolute prohibition of Personally Identifiable Information (PII) on the Hyperledger Besu distributed ledger:

```
+-----------------------------------------------------------------------------------+
|                            Zero-PII Invariant Rule                                |
+-----------------------------------------------------------------------------------+
| 1. Prohibited On-Chain: Names, PAN, Aadhaar, phone numbers, emails, physical      |
|    addresses, IP addresses, and plaintext bank account numbers.                   |
| 2. Permitted On-Chain: Cryptographic addresses (0x...), token balances, batch DvP  |
|    trade hashes, and Poseidon zero-knowledge identity commitments.                |
+-----------------------------------------------------------------------------------+
```

### 2.1 Identity Commitment Architecture
User identities are validated off-chain by an authorized Claim Issuer. An un-linkable, blinded commitment is written to the on-chain Identity Registry contract:

$$\text{IdentityCommitment} = H_{\text{Poseidon}}\left(\text{UserSecret} \parallel \text{JurisdictionCode} \parallel \text{KYCEpoch}\right)$$

This allows smart contracts to verify investor eligibility and FPI sectoral caps without revealing the user's real-world identity to blockchain node operators or auditors.

---

## 3. Data Subject Rights & Crypto-Shredding Erasure

Under Section 12 of the DPDP Act 2023, data principals possess the statutory right to correction, completion, updating, and erasure of their personal data.

### 3.1 The Blockchain Immutability vs Right-to-Erasure Dilemma
Traditional distributed ledgers cannot delete historical transaction blocks without breaking cryptographic hash chains. The platform resolves this conflict through **Envelope Encryption with HSM-Governed Crypto-Shredding (ADR-0007)**:
1. All off-chain relational data containing user PII is encrypted with a unique, user-specific Data Encryption Key (DEK).
2. The user's DEK is generated and protected inside a FIPS 140-2 Level 3 Hardware Security Module (HSM).
3. Upon receiving a valid erasure request:
   - The user's primary DEK is permanently destroyed within the HSM.
   - All historical off-chain ciphertexts stored in immutable append-only logs, TimescaleDB, and backups become mathematically impossible to decrypt.
   - The user's on-chain identity commitment is revoked, rendering historical on-chain addresses permanently anonymous.

---

## 4. Cross-Border Data Transfer & Localization

1. **Domestic Data Localization:** All primary databases, transaction logs, and key management systems storing Indian citizen data are located exclusively within the territory of India (AWS Mumbai `ap-south-1` and GIFT City).
2. **Whitelisted Transfers:** Cross-border transfer of encrypted telemetry to international entities (e.g. GIFT City IFSC Gateway) is conducted strictly in compliance with Central Government notifications under Section 16 of the DPDP Act 2023.
