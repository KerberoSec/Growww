# 907 - Regulatory Sandbox Pilot Launch Plan & Phased Rollout

## Purpose
Defines the operational execution plan and risk-mitigated rollout strategy for launching Growww into live operations under the SEBI / IFSCA Regulatory Sandbox framework. The sandbox pilot allows a controlled group of real retail investors to trade fractional, asset-backed digital security tokens using real INR while strictly enforcing regulatory risk caps, phased cohort expansions, real-time supervisory telemetry, and multi-layered emergency kill switches.

## What You Are Building
A complete Sandbox Pilot Launch & Operations Framework (`ops/sandbox/` and `services/pilot-manager/`):
- Phased Pilot Rollout Architecture with feature-flagged user cohort gating (Cohort 1: 50 internal staff, Cohort 2: 150 invited beta users, Cohort 3: 500 public retail investors).
- Hard Regulatory Enforcers: Maximum portfolio limit per user (e.g., INR 50,000), total sandbox aggregate volume caps, and permitted stock whitelist (top 10 Nifty 50 liquid equities).
- Emergency Kill Switch Architecture capable of instantly freezing new order intake, pausing on-chain settlement contracts, or initiating graceful fund repatriation.
- Automated Regulatory Telemetry Exporter transmitting real-time transaction health and risk metrics directly to SEBI / IFSCA supervisory endpoints.

## Scope Boundaries
- **In Scope:**
 - Defining cohort entry criteria, user whitelisting, and mandatory risk consent flows.
 - Hard limit enforcement in the pre-trade risk engine and smart contracts.
 - Phased launch milestone gates (Phase 0: Dry Run, Phase 1: Friends & Family, Phase 2: Closed Beta, Phase 3: Public Sandbox).
 - Emergency freeze and off-boarding procedures.
 - Weekly and monthly regulatory reporting automation for SEBI sandbox oversight committee.
- **Out of Scope / Handled Elsewhere:**
 - Pre-launch sandbox UAT testing (Prompt 906).
 - Full-scale production launch and unconstrained marketing rollout (Prompt 908).
 - Customer grievance resolution procedures (Prompt 910).

## Technology to Use
- **Feature Flagging & Cohort Gating:** Unleash (Self-hosted open-source enterprise feature management) integrated with user authentication tokens.
- **Pre-Trade Risk Enforcers:** Rust/Go Risk Service (`services/risk-engine`) evaluating in-memory real-time user balances and sandbox volume thresholds.
- **Regulatory Telemetry:** Prometheus + Grafana + custom Python exporter pushing structured JSON telemetry to SEBI SFTP/API endpoints over mTLS.
- **Emergency Operations:** HashiCorp Consul / Redis distributed flags + MultiSig governance on Hyperledger Besu.

*Justification:* Unleash combined with deterministic risk-engine checks guarantees that non-whitelisted users cannot access sandbox features and ensures limits are enforced at both application and smart contract layers.

## Backend / Infra Touchpoints
- Staging and Production-Isolated Sandbox Kubernetes Cluster.
- Unleash Feature Flag Server with mTLS.
- User Service & KYC Service (with Sandbox Cohort Tagging).
- Risk & Margin Pre-Trade Checks Service (Prompt 206).
- Hyperledger Besu Consortium Sandbox Network (4 Validators + 1 SEBI Observer Node).

## Blockchain Interaction
- Pilot user addresses are explicitly registered in `ComplianceRegistry.sol` with the `SANDBOX_PILOT_ROLE` attribute.
- `DigitalSecurityToken.sol` enforces hard smart contract-level minting caps ensuring the total tokenized market value cannot exceed the SEBI sandbox ceiling (e.g., INR 5 Crores aggregate).
- `SettlementDvP.sol` includes an emergency multi-sig pause function (`pauseSettlement()`) accessible only by a 2-of-3 threshold multisig held by the Chief Compliance Officer, Head of Engineering, and Custodian Trustee.
- Daily cryptographic proof-of-reserve roots are published to `ProofOfReserveRegistry.sol` and matched against physical custodian demat statements.

## Step-by-Step Build Instructions
1. Scaffold the pilot operations repository structure (`ops/sandbox/`) with `cohorts/`, `policies/`, `telemetry/`, and `runbooks/`.
2. Configure Unleash feature flags (`flags/sandbox_pilot_enabled.json`, `flags/user_tier_limits.json`) with strict user ID targeting.
3. Configure the Pre-Trade Risk Engine (`services/risk-engine`) with sandbox hard limits: INR 50,000 max portfolio per investor, INR 10,000 max single order, 5 orders/day limit.
4. Restrict trading universe to the SEBI-approved sandbox stock whitelist (e.g., `RELIANCE`, `TCS`, `HDFCBANK`, `INFY`, `ICICIBANK`).
5. Configure the on-chain smart contracts on Besu with sandbox supply ceilings and multi-sig pause keys.
6. Implement the mandatory in-app Sandbox Risk Disclosure and Consent Modal in Flutter (`lib/features/onboarding/sandbox_consent_dialog.dart`) requiring explicit digital signing before trading.
7. Implement the Emergency Kill Switch API and Admin Console dashboard button with two-person rule authorization.
8. Build the Automated Regulatory Telemetry Exporter (`services/regulatory-exporter/`) generating daily transaction digests, slippage metrics, and grievance counts.
9. Execute Phase 0 (Internal Dry Run, 5 days): 10 internal engineers executing end-to-end cycles with real bank accounts.
10. Execute Phase 1 (Alpha Cohort, 14 days): 50 selected users trading up to INR 10,000 each; verify daily reconciliation and latency.
11. Review Phase 1 results with SEBI Sandbox Review Committee and obtain written approval for Phase 2 expansion.
12. Execute Phase 2 (Beta Cohort, 30 days): 200 users trading up to INR 25,000 each; evaluate customer support ticket resolution times.
13. Execute Phase 3 (Public Sandbox, 60 days): 500 users trading up to INR 50,000 each; validate continuous proof-of-reserve disclosure.
14. Compile the Final Sandbox Performance & Exit Report for SEBI sandbox graduation.

## Interfaces / Contracts
```json
// Sandbox Cohort Policy & Limits Specification
// File: ops/sandbox/policies/cohort_limits.json
{
  "sandbox_program": "SEBI-REG-SANDBOX-COHORT-4",
  "effective_date": "2026-10-01",
  "jurisdiction": "SEBI / IFSCA",
  "limits": {
    "max_total_participants": 500,
    "max_portfolio_per_investor_inr": 50000.00,
    "max_single_trade_inr": 10000.00,
    "max_daily_trades_per_user": 5,
    "aggregate_sandbox_exposure_inr": 25000000.00
  },
  "approved_securities_whitelist": [
    { "ticker": "RELIANCE", "isin": "INE002A01018", "lot_precision": 4 },
    { "ticker": "TCS",      "isin": "INE467B01029", "lot_precision": 4 },
    { "ticker": "INFY",     "isin": "INE009A01021", "lot_precision": 4 },
    { "ticker": "HDFCBANK", "isin": "INE040A01034", "lot_precision": 4 }
  ],
  "kill_switch": {
    "enabled": true,
    "requires_multisig": true,
    "multisig_threshold": "2-of-3",
    "authorized_roles": ["CHIEF_COMPLIANCE_OFFICER", "HEAD_OF_OPERATIONS", "LEAD_ARCHITECT"]
  }
}
```

```yaml
# Regulatory Daily Telemetry Schema (ops/sandbox/telemetry/daily_digest_schema.yaml)
type: object
properties:
  report_date: { type: string, format: date }
  active_investors_count: { type: integer, maximum: 500 }
  total_trades_executed: { type: integer }
  total_turnover_inr: { type: number }
  avg_matching_latency_ms: { type: number }
  settlement_success_rate_percent: { type: number, minimum: 99.0 }
  proof_of_reserve_discrepancy_count: { type: integer, const: 0 }
  customer_grievances_opened: { type: integer }
  customer_grievances_resolved: { type: integer }
required:
 - report_date
 - active_investors_count
 - total_trades_executed
 - proof_of_reserve_discrepancy_count
```

## Security & Compliance Notes
- In strict adherence to SEBI Sandbox Regulations, all investor funds in the pilot are ring-fenced in a designated Escrow Account with an RBI-regulated Scheduled Commercial Bank.
- An Investor Protection Guarantee Reserve is funded prior to Phase 1 launch to cover any potential technical shortfall.
- The Emergency Kill Switch triggers an immediate order book cancel-all, suspends new deposits, and locks token minting within < 500ms of activation.

## Acceptance Criteria
- [ ] Cohort gating system strictly blocks non-whitelisted users from participating in the sandbox pilot.
- [ ] Risk engine and smart contracts enforce INR 50,000 investor limits with zero exceptions.
- [ ] Emergency Kill Switch passes multi-sig activation tests with instant order cancellation and settlement pausing.
- [ ] Automated regulatory telemetry generates and transmits daily compliance digests with 100% timeliness.
- [ ] Successful completion of all Sandbox rollout phases with zero unresolved custody discrepancies.

## Suggested Order / Dependencies
- **Prerequisites:** 003 (Regulatory Pathway), 206 (Risk Engine), 307 (MultiSig Governance), 906 (UAT Plan).
- **Parallel Tasks:** 908 (Production Launch Runbook), 910 (Customer Grievance Redressal).
