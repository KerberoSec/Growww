# 004 - Domestic KYC/AML & Prevention of Money Laundering Policy

## Purpose
Operating as a regulated fintech intermediary in India requires rigorous, automated, and legally compliant customer due diligence (CDD). Under the Prevention of Money Laundering Act (PMLA) 2002 and SEBI Master Circular on KYC norms, Growww must ensure that only thoroughly verified domestic Indian citizens and entities can deposit INR, acquire fractional equity tokens, and trade on the platform.

This document establishes the domestic identity verification policy, tiered risk-scoring framework, politically exposed persons (PEP) screening, automated suspicious transaction reporting (STR) rules, and bank account validation protocols for domestic users. It translates statutory mandates into concrete algorithmic rules and schema definitions for the engineering organization.

## What You Are Building
A comprehensive domestic KYC/AML policy and technical requirement specification (`docs/compliance/kyc_aml_domestic_policy.md`) covering:
- Aadhaar verification workflows: Paperless Offline eKYC (Aadhaar XML / QR) and UIDAI OTP-based eKYC with mandatory 8-digit masking.
- Income Tax Department PAN validation and active status verification via NSDL/UTIITSL databases.
- Central KYC Registry (CKYCR / CERSAI) integration protocol for automated 14-digit CKYC retrieval and update.
- Bank Account Verification via UPI Penny Drop / IMPS Reverse Penny Drop verifying name-match score >= 85%.
- Anti-Money Laundering (AML) risk categorization (Low, Medium, High Risk tiers) and enhanced due diligence (EDD) triggers.
- Politically Exposed Persons (PEP) and domestic sanctions (MHA / UAPA) screening rules.
- Automated generation of Suspicious Transaction Reports (STR) and Cash Transaction Reports (CTR) for FIU-IND FinGate.

## Scope Boundaries
- **In Scope:**
 - Policy definitions, rule matrices, validation thresholds, and risk scoring algorithms for Indian domestic investors.
 - PII masking rules and Aadhaar data vault architecture requirements.
 - KYC verification state machine transitions and re-KYC periodicity rules.
 - Hashed identity commitment generation for permissioned ledger whitelisting.
- **Out of Scope / Handled Elsewhere:**
 - Foreign investor onboarding and FATF/FATCA compliance (covered in Prompt 005).
 - Microservice implementation for KYC API integrations (covered in Prompt 202).
 - Client-side KYC UI screens in Flutter (covered in Prompt 503).

## Technology to Use
- Primary Format: Markdown policy specification (`docs/compliance/kyc_aml_domestic_policy.md`) and JSON Schema definitions for domestic KYC validation states.
- Hashing & Cryptography: SHA-256 with secret system salt for PII hashing; AES-256-GCM for Aadhaar Data Vault encrypted storage.
- Name Matching Algorithm: Jaro-Winkler distance and Double Metaphone phonetic matching algorithm (minimum threshold 0.85).
- Justification: Standardized phonetic and distance-based matching eliminates false rejections caused by minor name formatting variations between PAN, Aadhaar, and bank records while maintaining zero-fraud tolerance.

## Backend / Infra Touchpoints
- UIDAI eKYC / Offline XML Gateway.
- NSDL / Income Tax Department PAN Verification API.
- CKYCR (Central KYC Records Registry, CERSAI).
- NPCI / Bank IMPS/UPI Penny Drop verification APIs.
- FIU-IND FinGate portal for automated STR/CTR XML upload.
- Isolated Domestic PostgreSQL database (Aadhaar Data Vault / PII Vault).

## Blockchain Interaction
Establishes how verified domestic identity commitments interact with the permissioned ledger:
- **Zero-PII On-Chain Invariant:** No Aadhaar numbers, PANs, phone numbers, or full names are ever recorded on the Hyperledger Besu blockchain.
- **Cryptographic Whitelist Attestation:** Upon successful domestic KYC verification, the KYC Service generates an identity commitment hash: `commitment = keccak256(abi.encodePacked(pan_hash, investor_uuid, salt))`.
- **Compliance Registry Smart Contract:** The backend submits this 32-byte commitment hash and a KYC validity timestamp to `ComplianceRegistry.sol` on Hyperledger Besu.
- **Transfer Gating:** `DigitalSecurityToken.sol` and `SettlementDvP.sol` query `ComplianceRegistry.sol` during every trade to verify that both buyer and seller possess active, unexpired, and non-frozen KYC commitments before allowing execution.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review the statutory mandates of PMLA 2002, PMLA (Maintenance of Records) Rules 2005, and SEBI KYC Master Circulars.
2. Define the mandatory 4-step domestic onboarding pipeline: (a) PAN Verification, (b) Aadhaar Offline XML / CKYC Fetch, (c) Bank Account Penny Drop, (d) AI Face Liveness / Geolocation check.
3. Establish the PAN Verification algorithm: validate 10-character PAN format, query NSDL database, verify status is `OPERATIVE`, and ensure PAN-Aadhaar linking is confirmed.
4. Establish the Aadhaar masking policy: enforce cryptographic redaction of the first 8 digits in all images, logs, and database records, storing only the last 4 digits in plain text alongside the encrypted vault token.
5. Define the Bank Account Penny Drop matching rule: execute Re 1 penny drop, extract beneficiary name from IMPS response, and calculate Jaro-Winkler match against PAN/Aadhaar name.
6. Design the AML Risk Scoring Matrix classifying users into Low, Medium, or High risk based on annual income, occupation, geographical location, and transaction patterns.
7. Detail Enhanced Due Diligence (EDD) triggers: sudden deposits exceeding 5x declared monthly income, rapid turnover, or PEP associations trigger manual review by the Principal Officer.
8. Define the FIU-IND reporting triggers: cash deposits > INR 10 Lakhs (CTR) or anomalous round-tripping transactions (STR) automatically queue XML reports for FIU submission.
9. Specify the Re-KYC lifecycle: Low risk (every 10 years), Medium risk (every 8 years), High risk (every 2 years) with automated in-app alerts.
10. Design the cryptographic Identity Commitment generation algorithm for `ComplianceRegistry.sol` on Hyperledger Besu.
11. Construct the JSON schema representing the complete domestic KYC verification state.
12. Review the specification with the Principal Officer, Legal Compliance, and CISO.

## Interfaces / Contracts
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "DomesticKYCVerificationState",
  "type": "object",
  "properties": {
    "investor_uuid": { "type": "string", "format": "uuid" },
    "pan_verification": {
      "type": "object",
      "properties": {
        "pan_sha256": { "type": "string", "pattern": "^[a-f0-9]{64}$" },
        "pan_status": { "type": "string", "enum": ["OPERATIVE", "INOPERATIVE", "DELETION"] },
        "aadhaar_linked": { "type": "boolean" },
        "nsdl_verified_at": { "type": "string", "format": "date-time" }
      },
      "required": ["pan_sha256", "pan_status", "aadhaar_linked", "nsdl_verified_at"]
    },
    "bank_verification": {
      "type": "object",
      "properties": {
        "account_number_vault_token": { "type": "string" },
        "ifsc": { "type": "string", "pattern": "^[A-Z]{4}0[A-Z0-9]{6}$" },
        "penny_drop_status": { "type": "string", "enum": ["SUCCESS", "FAILED"] },
        "name_match_score": { "type": "number", "minimum": 0.0, "maximum": 1.0 }
      },
      "required": ["account_number_vault_token", "ifsc", "penny_drop_status", "name_match_score"]
    },
    "aml_risk_profile": {
      "type": "object",
      "properties": {
        "risk_tier": { "type": "string", "enum": ["LOW", "MEDIUM", "HIGH"] },
        "pep_status": { "type": "boolean" },
        "sanctions_clear": { "type": "boolean" },
        "fiu_str_flagged": { "type": "boolean" }
      },
      "required": ["risk_tier", "pep_status", "sanctions_clear", "fiu_str_flagged"]
    },
    "on_chain_commitment": {
      "type": "object",
      "properties": {
        "identity_commitment_hash": { "type": "string", "pattern": "^0x[a-fA-F0-9]{64}$" },
        "whitelist_expiry": { "type": "integer" }
      },
      "required": ["identity_commitment_hash", "whitelist_expiry"]
    }
  },
  "required": ["investor_uuid", "pan_verification", "bank_verification", "aml_risk_profile", "on_chain_commitment"]
}
```

## Security & Compliance Notes
- Strict compliance with UIDAI regulations: storing raw Aadhaar numbers or unredacted physical copies is punishable under the Aadhaar Act 2016. All Aadhaar numbers must be stored in a dedicated, isolated Aadhaar Data Vault with HSM-managed keys.
- PMLA records, verification logs, and audit trails must be retained in immutable storage for a minimum statutory period of 5 years following account closure.

## Acceptance Criteria
- [ ] `docs/compliance/kyc_aml_domestic_policy.md` is complete with full verification pipelines, name-matching algorithms, and AML risk tiers.
- [ ] Aadhaar Data Vault and 8-digit masking rules are formally specified with zero-plaintext exceptions.
- [ ] On-chain Identity Commitment generation algorithm for `ComplianceRegistry.sol` is defined with zero PII leakage.
- [ ] JSON Schema for domestic KYC verification state passes all validation tests.
- [ ] Formal sign-off obtained from Principal Officer (PMLA) and Chief Information Security Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary).
- Parallel Tasks: 005 (Foreign KYC/AML), 202 (KYC Service), 701 (Identity Architecture).
- Downstream Blockers: Blocks Prompt 202, Prompt 503, and Prompt 701.
