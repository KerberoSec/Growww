# 005 - International Investor KYC/AML Policy (GIFT City IFSCA)

## Purpose
Enabling international retail and institutional investors to invest in Indian equities through the GIFT City IFSC gateway requires strict adherence to international regulatory standards, including the Financial Action Task Force (FATF) Recommendations, IFSCA (Anti Money Laundering, Counter-Terrorist Financing and Know Your Customer) Guidelines 2022, and global sanctions frameworks (UN, US OFAC, EU, UK HM Treasury).

This document establishes the comprehensive Customer Due Diligence (CDD), Enhanced Due Diligence (EDD), international tax classification (FATCA/CRS), biometric liveness verification, and global sanctions screening policy for foreign investors entering via the GIFT City IFSC Entity.

## What You Are Building
A complete international compliance specification (`docs/compliance/kyc_aml_foreign_policy.md`) detailing:
- Standardized Foreign Investor Onboarding Pipeline: Passport OCR, Machine Readable Zone (MRZ) validation, and ICAO 9303 compliance verification.
- 3D Biometric AI Liveness & Facial Comparison against government-issued identity documents.
- Proof of Address (PoA) verification (utility bills, foreign bank statements < 90 days old) with automated geo-fencing.
- Global Sanctions & Watchlist Screening against OFAC (SDN list), UN Security Council, EU Consolidated List, and UK HMT.
- Politically Exposed Persons (PEP) global database screening and adverse media continuous monitoring.
- Foreign Account Tax Compliance Act (FATCA) and Common Reporting Standard (CRS) self-certification schema.
- Non-Permitted Jurisdictions policy (FATF Blacklist / High-Risk Jurisdictions subject to a Call for Action).

## Scope Boundaries
- **In Scope:**
 - Onboarding policies, risk matrices, sanctions rules, and tax residency schemas for international retail/institutional investors.
 - Foreign investor KYC state lifecycle, re-verification periodicity, and country risk scoring.
 - Whitelisting of international investor identity commitments on the permissioned ledger.
- **Out of Scope / Handled Elsewhere:**
 - Foreign currency payment gateway and FX conversion services (covered in Prompt 214).
 - Domestic Indian investor KYC verification (covered in Prompt 004).
 - Microservice implementation for external international KYC vendors (covered in Prompt 202).

## Technology to Use
- Primary Format: Markdown specification (`docs/compliance/kyc_aml_foreign_policy.md`) with JSON Schema validation schemas.
- External Integration Adapters: SumSub / Onfido / LexisNexis World-Check / Dow Jones Risk & Compliance REST APIs.
- Tax Residency & Country Standards: ISO 3166-1 alpha-3 country codes and OECD Common Reporting Standard XML schema v2.0.
- Justification: Utilizing ISO country codes and OECD CRS standards guarantees seamless tax exchange reporting between IFSCA and global tax authorities while preventing human error in jurisdictional classification.

## Backend / Infra Touchpoints
- Third-Party Identity Verification Services (Passport OCR, Biometric Liveness).
- Global Sanctions Screening Engine (LexisNexis World-Check / ComplyAdvantage).
- Isolated GIFT City IFSC PostgreSQL Database (Foreign Investor KYC Data Store).
- GIFT City Vault KMS cluster (envelope encryption for international documents).

## Blockchain Interaction
Establishes international investor authorization on the Hyperledger Besu consortium ledger:
- **Zero-PII On-Chain Invariant:** No foreign passport numbers, home addresses, or tax IDs are ever stored on-chain.
- **Cross-Border Identity Commitment:** Upon successful verification in GIFT City, the IFSC compliance engine creates an anonymous 32-byte cryptographic identifier: `foreign_investor_commitment = keccak256(abi.encodePacked(passport_hash, ifsc_investor_id, gift_city_salt))`.
- **Global Compliance Registry:** The GIFT City gateway node registers this commitment on `ComplianceRegistry.sol` on the permissioned Hyperledger Besu network, assigning jurisdictional flag `JURISDICTION_IFSCA`.
- **Trade Execution Gating:** `SettlementDvP.sol` verifies that foreign participants possess active `JURISDICTION_IFSCA` whitelisting before executing cross-border settlement against the domestic depository pool.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review FATF 40 Recommendations and IFSCA AML/CFT/KYC Guidelines 2022.
2. Compile the FATF High-Risk & Monitored Jurisdictions list (Blacklist / Grey List) and establish automated IP and nationality blocking.
3. Define the 5-stage International Onboarding Pipeline: (1) Country Eligibility Check, (2) Passport MRZ Extraction & NFC Verification, (3) 3D Biometric Liveness, (4) Proof of Address Verification, (5) FATCA/CRS Tax Declaration.
4. Establish the passport verification criteria: require ICAO 9303 compliant passports, validate checksums in MRZ line 1 and line 2, and verify expiry date > 6 months.
5. Define the 3D Biometric Liveness standard: require ISO/IEC 30107-3 Level 2 PAD (Presentation Attack Detection) compliance to prevent deepfakes and spoofing.
6. Detail the Global Sanctions Screening protocol: real-time screening against OFAC, EU, UK, and UN lists prior to account activation, with automated daily delta re-screening.
7. Design the PEP risk matrix: classify foreign PEPs, domestic PEPs, and close associates; mandate Senior Management sign-off for any PEP onboarding.
8. Specify the FATCA / CRS questionnaire schema: capture Country of Tax Residence, Tax Identification Number (TIN), US Person status (Form W-9/W-8BEN equivalent), and source of funds declaration.
9. Establish the ongoing monitoring policy: high-risk accounts reviewed annually; automated screening for material adverse media mentions.
10. Design the cross-border Identity Commitment algorithm for `ComplianceRegistry.sol` on Hyperledger Besu.
11. Build the JSON schema for International Investor Compliance Profiles.
12. Review and finalize with GIFT City Legal Counsel, Head of International Compliance, and Chief Security Architect.

## Interfaces / Contracts
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "InternationalInvestorComplianceProfile",
  "type": "object",
  "properties": {
    "foreign_investor_uuid": { "type": "string", "format": "uuid" },
    "jurisdiction_code": { "type": "string", "pattern": "^[A-Z]{3}$" },
    "passport_verification": {
      "type": "object",
      "properties": {
        "passport_hash_sha256": { "type": "string", "pattern": "^[a-f0-9]{64}$" },
        "mrz_valid": { "type": "boolean" },
        "issuing_country_iso3": { "type": "string", "pattern": "^[A-Z]{3}$" },
        "expiry_date": { "type": "string", "format": "date" }
      },
      "required": ["passport_hash_sha256", "mrz_valid", "issuing_country_iso3", "expiry_date"]
    },
    "liveness_verification": {
      "type": "object",
      "properties": {
        "liveness_score": { "type": "number", "minimum": 0.0, "maximum": 1.0 },
        "face_match_score": { "type": "number", "minimum": 0.0, "maximum": 1.0 },
        "anti_spoof_passed": { "type": "boolean" }
      },
      "required": ["liveness_score", "face_match_score", "anti_spoof_passed"]
    },
    "tax_declaration": {
      "type": "object",
      "properties": {
        "tax_residency_country_iso3": { "type": "string", "pattern": "^[A-Z]{3}$" },
        "tin_vault_token": { "type": "string" },
        "is_us_person": { "type": "boolean" },
        "fatca_status": { "type": "string", "enum": ["COMPLIANT_W8BEN", "COMPLIANT_W9", "EXEMPT"] }
      },
      "required": ["tax_residency_country_iso3", "tin_vault_token", "is_us_person", "fatca_status"]
    },
    "sanctions_screening": {
      "type": "object",
      "properties": {
        "ofac_cleared": { "type": "boolean" },
        "un_sanctions_cleared": { "type": "boolean" },
        "pep_detected": { "type": "boolean" },
        "screened_at": { "type": "string", "format": "date-time" }
      },
      "required": ["ofac_cleared", "un_sanctions_cleared", "pep_detected", "screened_at"]
    },
    "on_chain_whitelist": {
      "type": "object",
      "properties": {
        "foreign_commitment_hash": { "type": "string", "pattern": "^0x[a-fA-F0-9]{64}$" },
        "jurisdiction_flag": { "type": "string", "enum": ["JURISDICTION_IFSCA"] }
      },
      "required": ["foreign_commitment_hash", "jurisdiction_flag"]
    }
  },
  "required": [
    "foreign_investor_uuid",
    "jurisdiction_code",
    "passport_verification",
    "liveness_verification",
    "tax_declaration",
    "sanctions_screening",
    "on_chain_whitelist"
  ]
}
```

## Security & Compliance Notes
- Strict adherence to FATF recommendations: accounts originating from FATF blacklist jurisdictions (e.g., DPRK, Iran) are blocked immediately at network and registration levels.
- All international identity documents, biometric data, and tax forms must be stored strictly within the GIFT City IFSC jurisdictional cloud boundary, encrypted at rest using AES-256-GCM.
- In accordance with GDPR and international data protection standards, international investors retain the right to rectify information and request cryptographic erasure of off-chain records upon account termination.

## Acceptance Criteria
- [ ] `docs/compliance/kyc_aml_foreign_policy.md` is complete, detailing passport verification, 3D biometric liveness, and sanctions screening.
- [ ] FATCA/CRS self-certification schema is documented and compliant with OECD standards.
- [ ] Sanctions screening rules for OFAC, UN, EU, and UK lists are formally codified with real-time and daily delta checks.
- [ ] On-chain international identity commitment structure for `ComplianceRegistry.sol` is specified with zero PII.
- [ ] Sign-off obtained from GIFT City Compliance Officer and International Legal Counsel.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 002 (Two-Entity Structure), 004 (Domestic KYC/AML).
- Parallel Tasks: 110 (Inter-Entity Communication), 214 (Foreign Investor Funding).
- Downstream Blockers: Blocks Prompt 214 and international onboarding flows in Prompt 503.
