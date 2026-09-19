# 906 - User Acceptance Testing (UAT) Plan for Real-World Regulatory Sandbox Scenarios

## Purpose
Defines the comprehensive User Acceptance Testing (UAT) plan and regulatory simulation framework designed to validate Growww within the SEBI and IFSCA Regulatory Sandbox environments. Before onboarding pilot retail investors, the system must undergo end-to-end scenario testing with business stakeholders, compliance officers, and regulatory observers - verifying complex real-world market scenarios such as KYC rejections, circuit breakers, DvP bank reversals, corporate action splits, and daily custody reconciliation.

## What You Are Building
A structured UAT testing suite, scenario catalog, and regulatory observer verification harness (`tests/uat/` and `docs/uat/`):
- Comprehensive UAT Test Scenario Matrix (`docs/uat/SEBI_Sandbox_UAT_Matrix.xlsx` / markdown catalog) covering 40+ complex edge cases.
- Regulatory Sandbox Observer Verification setup providing SEBI/IFSCA auditors with read-only dashboard access and dedicated Hyperledger Besu observer nodes.
- Depository & Banking Exception Simulation Harness (`tests/uat/simulators/`) injecting realistic failure modes (e.g., bank timeout on UPI collect, NSDL settlement rejection, corporate action stock split).
- End-to-end UAT execution workflows and sign-off templates for internal compliance and external regulatory stakeholders.

## Scope Boundaries
- **In Scope:**
 - Validating complete business journeys with cross-functional testers (Compliance, Legal, Product, Operations).
 - Simulating SEBI sandbox operational boundaries (maximum 500 whitelisted sandbox participants, aggregate INR transaction limits).
 - Regulatory observer verification: Real-time ledger queries, proof-of-reserve validation, and automated compliance reporting.
 - Edge case simulations: Circuit breaker halts, partial order fills, deposit refunds, KYC re-verification requests, fractional corporate actions (dividends/splits).
- **Out of Scope / Handled Elsewhere:**
 - Automated developer unit/integration tests (Prompt 901).
 - High-throughput load testing (Prompt 903).
 - Formal sandbox launch and cohort management (Prompt 907).

## Technology to Use
- **Test Management & Execution Tracking:** Xray / TestRail integrated with Jira for auditable scenario execution, evidence attachment (screenshots, tx hashes), and compliance sign-offs.
- **Regulatory Observer Interface:** Grafana Compliance Portal + Hyperledger Besu Explorer (Blockscout / customized read-only portal) connected to a zero-gas observer node.
- **Scenario Automation & Mock Feeds:** Python-based UAT scenario runner (`pytest-bdd`) with configurable state injectors for mock banking, KYC, and depository systems.

*Justification:* Combining structured test management with dedicated regulatory observer portals provides transparent, verifiable evidence to SEBI/IFSCA that all regulatory controls are functioning correctly.

## Backend / Infra Touchpoints
- Staging UAT Environment mirroring production configurations.
- Admin & Back-Office Console (Prompt 604, 605, 606).
- User & KYC Services with manual compliance review overrides.
- Settlement Service connected to mock NSDL/CDSL batch files.
- Dedicated Regulatory Sandbox Hyperledger Besu Observer Node.

## Blockchain Interaction
- SEBI and IFSCA regulatory observers connect to dedicated, read-only Hyperledger Besu observer nodes.
- UAT testers execute trades and verify that every settlement generates an on-chain event on `SettlementDvP.sol` with immutable transaction receipts.
- UAT verifies that non-whitelisted or frozen accounts (`ComplianceRegistry.sol`) are strictly prevented from transferring or receiving digital security tokens.
- Testers verify that daily end-of-day proof-of-reserve Merkle roots (`ProofOfReserveRegistry.sol`) match physical share quantities reported by the mock NSDL/CDSL depository ledger.

## Step-by-Step Build Instructions
1. Establish the `tests/uat/` directory structure with `scenarios/`, `simulators/`, `checklists/`, and `evidence/`.
2. Author the Master UAT Scenario Catalog (`docs/uat/SEBI_Sandbox_UAT_Matrix.md`) detailing test cases, preconditions, steps, expected results, and regulatory mapping.
3. Deploy a dedicated, read-only Hyperledger Besu observer node for regulatory auditor access with an isolated RPC interface.
4. Configure the Admin Console UAT workspace allowing compliance testers to review KYC applications, initiate manual freezes, and approve corporate actions.
5. Implement the Depository Exception Simulator (`tests/uat/simulators/depository_mock.py`) capable of simulating settlement delays, partial deliveries, and ISIN suspensions.
6. Implement the Banking Rail Exception Simulator (`tests/uat/simulators/bank_mock.py`) simulating UPI mandate expirations, partial debit recoveries, and bank reversals.
7. Execute Scenario Group 1: Onboarding & KYC compliance (valid Aadhaar, rejected PAN, foreign investor passport via GIFT City pathway).
8. Execute Scenario Group 2: Fractional order execution and matching (limit buy, limit sell, partial fills across multiple buyers, cancel-replace).
9. Execute Scenario Group 3: Atomic DvP settlement and banking failure handling (simulating bank API failure after off-chain match, verifying automatic order reversal).
10. Execute Scenario Group 4: Market halt & Circuit Breakers (simulating 10% stock price move, asserting instant order book halt and notification dispatch).
11. Execute Scenario Group 5: Corporate actions (simulating 2:1 stock split on TCS, verifying fractional token balance doubling and proof-of-reserve update).
12. Execute Scenario Group 6: End-of-day reconciliation and regulatory report generation (verifying automated generation of SEBI Form X reports and ledger hash parity).
13. Compile test execution evidence, audit logs, and on-chain transaction hashes into the final UAT Sign-Off Package.

## Interfaces / Contracts
```markdown
# SEBI Regulatory Sandbox UAT Scenario Definition (Excerpt)
# File: docs/uat/scenarios/UAT-SCEN-014-dvp-settlement-failure.md

### Scenario ID: UAT-SCEN-014
**Title:** Atomic DvP Settlement Rollback on Upstream Bank Failure
**Regulatory Ref:** SEBI Interoperability & Settlement Finality Guidelines
**Category:** Settlement & Funds Flow

#### Preconditions:
1. Investor A has INR 10,000 wallet balance.
2. Investor B holds 1.0 unit of RELIANCE digital token.
3. Bank Mock is set to return `PAYMENT_TIMEOUT` on settlement transfer.

#### Execution Steps:
1. Investor A submits Limit Buy for 1.0 RELIANCE at INR 2,500.
2. Matching Engine matches order against Investor B's sell order.
3. Settlement Service initiates DvP transaction and calls Bank Mock for fiat transfer.
4. Bank Mock returns HTTP 504 (Timeout / Settlement Failure).
5. Settlement Service invokes `cancelDvP(tradeId)` on `SettlementDvP.sol`.

#### Expected Results:
- [ ] On-chain token remains locked in escrow and is safely refunded to Investor B.
- [ ] Investor A's reserved INR 2,500 balance is released back to available balance.
- [ ] Kafka event `settlement.dvp.failed` is emitted with audit log entry.
- [ ] Both investors receive in-app push notification: "Order settlement unsuccessful. Funds restored."
- [ ] Zero state desynchronization between PostgreSQL balances and Besu token balances.
```

## Security & Compliance Notes
- UAT test accounts must strictly operate in sandboxed environments with mock INR currencies and test securities.
- All regulatory observer access must be authenticated via multi-factor authentication (MFA) and logged in immutable audit trails.
- Any discrepancy found during UAT reconciliation (even by 0.0001 fractional unit or 1 paisa) is classified as a Blocker and halts release.

## Acceptance Criteria
- [ ] 100% of defined UAT regulatory scenarios (40+ test cases) executed and passed with documented evidence.
- [ ] SEBI/IFSCA Regulatory Observer node successfully syncs blocks and queries contract state with zero errors.
- [ ] Corporate actions (splits/dividends) accurately adjust fractional token balances and proof-of-reserve roots.
- [ ] Bank and depository failure simulations demonstrate 100% clean rollbacks with zero orphaned funds or locked tokens.
- [ ] Formal UAT Sign-Off received from Compliance Officer, Product Head, and Lead System Architect.

## Suggested Order / Dependencies
- **Prerequisites:** 215 (Reconciliation), 216 (Reporting), 604-607 (Admin Consoles), 901, 902.
- **Parallel Tasks:** 907 (Sandbox Pilot Launch Plan), 908 (Production Launch Runbook).
