# 009 - Risk Disclosure & Investor Protection Framework

## Purpose
Democratizing fractional access to Indian equities via permissioned distributed ledger technology introduces distinct operational, liquidity, and technological nuances that retail investors must fully comprehend. Under SEBI Investor Protection guidelines, SEBI (Research Analysts/Investment Advisers) Regulations, and IFSCA Conduct of Business rules, financial intermediaries must enforce transparent risk disclosures, pre-trade suitability evaluations, robust grievance redressal, and circuit breaker protections.

This document establishes the comprehensive investor protection architecture, mandatory statutory risk warnings, suitability assessment scoring algorithms, client grievance resolution workflows (including automated SEBI SCORES and IFSCA SPARK integration), and platform-wide circuit breaker safety mechanisms for Growww.

## What You Are Building
A comprehensive investor protection framework and risk disclosure specification (`docs/compliance/investor_protection_framework.md`) covering:
- Standardized Risk Disclosure Taxonomy (Market Risk, Fractional Liquidity Risk, Custody Operational Risk, Blockchain Technology Risk, Tax Implications).
- Investor Suitability Assessment Engine: algorithmic scoring of investor risk tolerance, financial literacy, and investment objectives prior to enabling trading.
- Mandatory Pre-Trade Risk Acknowledgement & Disclosure Prompts (modal disclosures in Flutter clients and Web console).
- Grievance Redressal Mechanism & Tiered Escalation Workflow: internal resolution SLA (48 hours), escalation to Principal Compliance Officer (7 days), and automated bidirectional integration with SEBI SCORES (SEBI Complaints Redress System) and IFSCA SPARK portals.
- Platform Circuit Breaker and Trading Halt Architecture: automatic trading suspension matching exchange halts (NSE/BSE circuit breakers) and ledger anomaly triggers.
- Investor Protection & Education Fund (IPEF) contribution tracking and educational onboarding modules.

## Scope Boundaries
- **In Scope:**
 - Risk disclosure text, versioning lifecycle, and mandatory client acceptance tracking.
 - Mathematical suitability scoring algorithms and risk threshold matrices.
 - Grievance state machine and regulatory ticketing protocol specifications.
 - Market halt triggers, emergency circuit breaker definitions, and safety limits.
- **Out of Scope / Handled Elsewhere:**
 - Flutter UI implementation of risk disclaimer popups and questionnaires (covered in Prompts 503 and 513).
 - Web console implementation for support and complaints (covered in Prompt 604).
 - Real-time market data feed ingestion for circuit breakers (covered in Prompt 207).

## Technology to Use
- Primary Format: Markdown specification (`docs/compliance/investor_protection_framework.md`) accompanied by YAML suitability matrices (`docs/compliance/suitability_matrix.yaml`).
- Regulatory Integration Protocols: REST / SOAP APIs for SEBI SCORES 2.0 and IFSCA SPARK grievance portal synchronization.
- Audit Storage: Cryptographically signed consent and acknowledgement logs stored in WORM (Write Once, Read Many) compliant storage.
- Justification: WORM storage guarantees non-repudiation of investor risk acknowledgements, providing irrefutable legal evidence during regulatory audits and ombudsman dispute proceedings.

## Backend / Infra Touchpoints
- User Service (Prompt 201).
- Risk & Margin Checks Service (Prompt 206).
- Market Data Service (Prompt 207).
- Support & Admin Back Office Service (Prompt 217).
- SEBI SCORES 2.0 API / IFSCA SPARK Gateway.

## Blockchain Interaction
Establishes on-chain investor protection and emergency circuit breaker controls on Hyperledger Besu:
- **Emergency Circuit Breaker Hooks:** `DigitalSecurityToken.sol` and `SettlementDvP.sol` implement Pausable governance hooks. When NSE/BSE triggers a market-wide circuit breaker (e.g., 10%, 15%, 20% index movement), the automated market monitor invokes `emergencyPause()` on-chain via multi-sig relayer.
- **Non-Transferability of Frozen Assets:** In the event of an individual regulatory freeze or court order, `ComplianceRegistry.sol` marks the investor's commitment hash as `FROZEN`, preventing unauthorized on-chain liquidation while protecting beneficial ownership.
- **Zero-PII Grievance Logging:** Grievance ticket identifiers and settlement resolution receipts recorded on-chain reference strictly pseudonymous transaction hashes and anonymous dispute IDs.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review SEBI Master Circulars on Investor Grievance Redressal and IFSCA Conduct of Business Regulations 2022.
2. Compile the 5 Core Statutory Risk Disclosures: Market Volatility, Fractional Liquidity Mechanics, Physical Depository Custody Model, Blockchain Ledger Finality, and Tax Liability.
3. Design the Investor Suitability Assessment Algorithm: formulate an 8-question questionnaire evaluating age, annual income, investment horizon, financial experience, and risk appetite, scoring users from 0 to 100.
4. Establish Suitability Tiers: Conservative (0-40: limited to large-cap blue-chip equities), Moderate (41-75: broad market equities), Aggressive (76-100: full equity catalog).
5. Define the Mandatory Pre-Trade Consent Protocol: enforce explicit, unskippable risk disclosure acceptance before an investor's first trade and upon any material update to regulatory terms.
6. Design the Grievance Redressal State Machine: `TICKET_OPEN` -> `UNDER_INVESTIGATION` (max 48h) -> `PROPOSED_RESOLUTION` -> `INVESTOR_ACCEPTED` or `ESCALATED_TO_COMPLIANCE` (max 7 days) -> `EXTERNAL_OMBUDSMAN_SCORES`.
7. Detail the bidirectional API sync with SEBI SCORES 2.0: automated polling for pending complaints, action-taken report (ATR) uploads, and resolution status updates.
8. Establish the Platform Circuit Breaker Policy: map real-time market data triggers from NSE/BSE to automatic platform-wide and ISIN-specific trading halts.
9. Define the Negative Balance Protection rule: ensure retail accounts can never incur debt or negative cash balances under any execution condition.
10. Specify the WORM audit logging format capturing investor IP address, timestamp, device fingerprint, and accepted disclosure version hash.
11. Construct the YAML schema for suitability questions and scoring weights.
12. Review specification with Chief Compliance Officer, Investor Relations Lead, and Legal Counsel.

## Interfaces / Contracts
```yaml
# Suitability Assessment & Risk Matrix Schema (docs/compliance/suitability_matrix.yaml)
suitability_framework:
  version: "1.0.0"
  max_score: 100
  tiers:
 - tier: "CONSERVATIVE"
      min_score: 0
      max_score: 40
      allowed_asset_classes: ["NIFTY_50_INDEX_CONSTITUENTS"]
      max_single_order_inr: 25000.00
 - tier: "MODERATE"
      min_score: 41
      max_score: 75
      allowed_asset_classes: ["NIFTY_100_CONSTITUENTS", "LARGE_CAP_EQUITIES"]
      max_single_order_inr: 100000.00
 - tier: "AGGRESSIVE"
      min_score: 76
      max_score: 100
      allowed_asset_classes: ["ALL_LISTED_EQUITIES"]
      max_single_order_inr: 500000.00

grievance_sla:
  acknowledgement_hours: 2
  first_response_hours: 24
  internal_resolution_hours: 48
  statutory_escalation_days: 7
  regulatory_bodies: ["SEBI_SCORES", "IFSCA_SPARK"]
```

## Security & Compliance Notes
- Strict compliance with SEBI circular SEBI/HO/OIAE/OIAE_IAD-1/P/CIR/2023/131 on the Online Dispute Resolution (ODR) portal.
- All risk acknowledgement logs must be cryptographically signed using the client's session key and stored in immutable WORM audit logs with a 5-year retention schedule.

## Acceptance Criteria
- [ ] `docs/compliance/investor_protection_framework.md` is complete with statutory risk disclosures, suitability engine algorithms, and grievance escalation workflows.
- [ ] Automated SEBI SCORES 2.0 and IFSCA SPARK integration specifications are documented.
- [ ] Suitability scoring matrix (`suitability_matrix.yaml`) is fully codified with concrete scoring formulas and asset tier gates.
- [ ] Platform circuit breaker policy and on-chain emergency pause mechanisms are specified.
- [ ] Sign-off obtained from Principal Compliance Officer and Chief Risk Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary), 003 (Regulatory Pathway).
- Parallel Tasks: 206 (Risk & Margin Checks), 217 (Admin Back Office), 513 (Risk UI Screens).
- Downstream Blockers: Blocks Prompt 206, Prompt 217, and Prompt 513.
