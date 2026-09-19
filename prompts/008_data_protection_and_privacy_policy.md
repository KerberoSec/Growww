# 008 - Data Protection, Privacy & Zero-PII Ledger Policy (DPDP & GDPR)

## Purpose
Operating a regulated fintech and permissioned blockchain infrastructure handling sensitive financial, identity, and banking data requires ironclad data protection governance. The platform must comply with India's Digital Personal Data Protection (DPDP) Act 2023, the European Union General Data Protection Regulation (GDPR) for international investors, and the strict non-negotiable architectural invariant: **ZERO Personally Identifiable Information (PII) on the blockchain ledger**.

This document defines the data classification framework, cryptographic tokenization and pseudonymization standards, consent management lifecycle, off-chain PII vault architecture, and data subject rights handling (including Right to Access, Correction, and Cryptographic Erasure / Right to be Forgotten) across all Growww systems.

## What You Are Building
A comprehensive data privacy architecture and engineering policy document (`docs/security/data_protection_policy.md`) covering:
- The 4-Tier Data Classification Matrix (Tier 1: Highly Sensitive PII; Tier 2: Sensitive Financial & KYC Data; Tier 3: Pseudonymous System Identifiers; Tier 4: Public & Ledger Data).
- The Zero-PII Blockchain Policy: strict technical mechanisms prohibiting PII from ever reaching transaction calldata, events, logs, or smart contract storage on Hyperledger Besu.
- The Cryptographic Tokenization & Pseudonymization Architecture (HashiCorp Vault Transit Engine + HMAC-SHA256).
- The DPDP Act 2023 Consent Manager Specification: tracking granular purpose-specific consent artifacts with cryptographic timestamps.
- Data Subject Rights Protocols: automated handling of Right to Access (data export packages) and Right to Erasure (cryptographic key destruction for off-chain records while preserving on-chain ledger integrity).
- Cross-Border Data Transfer Policies: strict data localization for Indian domestic records in Mumbai and IFSCA records in GIFT City.

## Scope Boundaries
- **In Scope:**
 - Data classification standards, encryption-at-rest and in-transit requirements.
 - Architectural blueprint for the PII Vault and Tokenization Service.
 - Consent management state machine and data erasure algorithms.
 - CI/CD automated secret/PII scanning rules blocking accidental PII leakage.
- **Out of Scope / Handled Elsewhere:**
 - HashiCorp Vault production deployment manifests (covered in Prompt 109).
 - Cloud KMS and HSM infrastructure provisioning (covered in Prompt 805).
 - User authentication and OAuth2 token handling (covered in Prompt 105).

## Technology to Use
- Tokenization & Cryptography: HashiCorp Vault Transit Secret Engine (Format-Preserving Encryption [FPE] and AES-256-GCM envelope encryption).
- Hashing & Pseudonymization: HMAC-SHA256 with HSM-managed rotation keys for investor on-chain commitments.
- Data Validation & Linters: Automated pre-commit hooks and CI static analysis using Trivy, Trufflehog, and custom Semgrep AST rules detecting PII in blockchain event emissions.
- Justification: Vault Transit Engine allows tokenizing sensitive PII into irreversible surrogate tokens that can safely traverse backend microservices while raw PII remains encrypted in an isolated, auditable enclave.

## Backend / Infra Touchpoints
- HashiCorp Vault Cluster (Dedicated Transit Engine).
- Isolated PII Vault Database (PostgreSQL with field-level encryption).
- User Service (Prompt 201).
- Immutable Audit Log Service (Prompt 218).
- Data Privacy Consent Management DB.

## Blockchain Interaction
Enforces the supreme Zero-PII architectural mandate across the Hyperledger Besu ledger:
- **Absolute PII Ban:** No names, email addresses, phone numbers, Aadhaar numbers, PANs, passport numbers, bank account numbers, or IP addresses may ever be transmitted in RPC payloads, transaction inputs, event logs, or contract state.
- **Pseudonymous Account Hashing:** On-chain accounts, whitelist registrations in `ComplianceRegistry.sol`, and order settlements in `SettlementDvP.sol` strictly use opaque 32-byte cryptographic hashes (`bytes32 investorCommitment = keccak256(abi.encodePacked(uuid, salt))`).
- **Right to Erasure Compatibility:** When a user invokes their right to be forgotten, the off-chain PII encryption key in HashiCorp Vault is cryptographically destroyed (crypto-shredded). The on-chain ledger remains mathematically valid and immutable, but the on-chain commitment hash can never again be linked to any physical human identity.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Analyze statutory compliance requirements under India DPDP Act 2023, EU GDPR, and SEBI/RBI cybersecurity circulars.
2. Establish the 4-Tier Data Classification Matrix defining handling, encryption, and retention rules for every data attribute across the system.
3. Define the Zero-PII Blockchain Policy, specifying prohibited fields, linting rules, and relayer rejection filters.
4. Architect the PII Vault: design the isolated, restricted-access database schema where raw identities are encrypted using AES-256-GCM.
5. Define the Tokenization Protocol: specify how microservices exchange surrogate UUIDs and Vault tokens without ever accessing raw PII.
6. Design the Consent Management Architecture: implement the Consent Artifact schema capturing Purpose, Notice Version, Timestamp, and Expiry.
7. Formulate the Cryptographic Shredding / Right to be Forgotten Protocol: detail the exact sequence for deleting off-chain database records and destroying Vault encryption keys upon verified user request.
8. Establish Data Localization and Cross-Border Transfer Rules: domestic Indian investor PII strictly stored in AWS `ap-south-1`; international investor data stored in GIFT City IFSC.
9. Implement pre-commit and CI/CD automated linting rules (Semgrep / Trufflehog) scanning for PII keywords in protobuf contracts and Solidity events.
10. Specify third-party vendor data sharing controls (minimal payload sharing with KYC/sanctions providers over encrypted mTLS).
11. Define Data Breach Incident Response procedures: mandatory notification to the Data Protection Board of India (DPBI) and CERT-In within 6 hours of confirmed breach.
12. Review specification with Data Protection Officer (DPO), Legal Counsel, and CISO.

## Interfaces / Contracts
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ConsentArtifact",
  "type": "object",
  "properties": {
    "consent_id": { "type": "string", "format": "uuid" },
    "investor_uuid": { "type": "string", "format": "uuid" },
    "purposes": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "purpose_code": { 
            "type": "string", 
            "enum": ["KYC_VERIFICATION", "TRADE_EXECUTION", "TAX_REPORTING", "MARKETING"] 
          },
          "is_mandatory": { "type": "boolean" },
          "status": { "type": "string", "enum": ["GRANTED", "REVOKED", "EXPIRED"] },
          "granted_at": { "type": "string", "format": "date-time" },
          "revoked_at": { "type": ["string", "null"], "format": "date-time" }
        },
        "required": ["purpose_code", "is_mandatory", "status", "granted_at"]
      }
    },
    "notice_version": { "type": "string" },
    "ip_address_hash": { "type": "string", "pattern": "^[a-f0-9]{64}$" },
    "user_agent": { "type": "string" },
    "digital_signature": { "type": "string" }
  },
  "required": ["consent_id", "investor_uuid", "purposes", "notice_version", "ip_address_hash", "digital_signature"]
}
```

## Security & Compliance Notes
- Under the DPDP Act 2023, penalties for failure to prevent personal data breaches reach up to INR 250 Crores per incident; strict architectural containment is a critical operational safeguard.
- Zero PII on-chain ensures that blockchain immutability does not conflict with statutory Right to be Forgotten mandates under GDPR and DPDP.

## Acceptance Criteria
- [ ] `docs/security/data_protection_policy.md` is complete with 4-Tier Data Classification Matrix and Zero-PII Ledger Policy.
- [ ] Automated CI/CD Semgrep rules detecting PII in protobufs and Solidity contracts are configured and passing.
- [ ] Cryptographic shredding and Right to be Forgotten workflow is documented with mathematical privacy guarantees.
- [ ] Consent Artifact JSON schema is defined and validated.
- [ ] Formal sign-off obtained from Data Protection Officer (DPO) and Chief Information Security Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary), 002 (Two-Entity Structure).
- Parallel Tasks: 105 (Auth Architecture), 109 (Secrets Management), 201 (User Service).
- Downstream Blockers: Blocks Prompt 109, Prompt 201, Prompt 202, and all smart contract designs in Category 3.
