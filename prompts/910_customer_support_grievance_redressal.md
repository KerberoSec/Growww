# 910 - Customer Support, Dispute Resolution & Grievance Redressal (SEBI SCORES / ODR Integration)

## Purpose
Establishes the customer support, dispute resolution, and regulatory grievance redressal platform for Growww, complying with SEBI and RBI investor protection frameworks. Because investors hold fractional, asset-backed securities settled via on-chain and off-chain rails, customer disputes (such as deposit delays, failed order executions, fractional dividend claims, and unauthorized account access claims) must be resolved with rapid Turnaround Times (TAT) through an auditable, multi-tier escalation system integrated with SEBI SCORES 2.0 and the SMART ODR (Online Dispute Resolution) portal.

## What You Are Building
An integrated Customer Support & Grievance Redressal Subsystem (`services/grievance-service/` and `apps/admin-support/`):
- Multi-tier support escalation engine (Tier 1: In-App Chat / Helpdesk, Tier 2: Grievance Redressal Officer, Tier 3: Principal Compliance Officer & SEBI SCORES).
- Automated Dispute Evidence Package Generator compiling unified timelines with bank gateway reference IDs, order execution logs, matching engine tickets, and Hyperledger Besu on-chain settlement receipts.
- Direct bidirectional API adapter integrating with SEBI SCORES 2.0 and SMART ODR platform for seamless regulatory ticket synchronization.
- In-app Flutter Support UI (`lib/features/support/`) allowing users to file disputes, track ticket progress, submit additional documents, and escalate unresolved issues.

## Scope Boundaries
- **In Scope:**
 - Automated ticketing workflow with strict SLA tracking (48h internal first response, 7-day resolution target, 21-day SEBI statutory maximum).
 - PII-masked customer support console for Tier-1/Tier-2 support agents.
 - Cryptographic dispute verification linking off-chain trade records to on-chain transaction hashes.
 - Integration with SEBI SCORES 2.0 webhooks and SMART ODR dispute resolution APIs.
 - Monthly regulatory grievance reporting (SEBI Investor Grievance Report Form 1A).
- **Out of Scope / Handled Elsewhere:**
 - Investor Web and Admin console foundational UI scaffolding (Prompt 601, 604).
 - Production monitoring and SRE alerting (Prompt 909).
 - Security incident response for platform-wide breaches (Prompt 706).

## Technology to Use
- **Workflow & SLA Orchestration:** Temporal.io (or Python Celery/FastAPI) - manages long-running regulatory dispute workflows with deterministic timers, SLA escalations, and automated state persistence.
- **Backend Service:** Python/FastAPI (`services/grievance-service`) with PostgreSQL 16+ storing encrypted ticket data, message threads, and immutable audit logs.
- **Support UI:** Next.js React Admin Console for agents + Flutter Support Feature module for mobile/desktop investors.
- **Regulatory Gateways:** REST/JSON client with mTLS for SEBI SCORES 2.0 API and SMART ODR dispute integration.

*Justification:* Temporal guarantees durable execution of multi-week regulatory dispute timers without risking dropped escalation events during server restarts.

## Backend / Infra Touchpoints
- Staging and Production PostgreSQL (`grievance_db`).
- Temporal Server Cluster for workflow and timer orchestration.
- Integration with Core Services: User Service (Prompt 201), Wallet Service (Prompt 203), Order Service (Prompt 204), Settlement Service (Prompt 208), Audit Log Service (Prompt 218).
- External Regulatory Gateways: SEBI SCORES 2.0 API endpoint and SMART ODR Portal.

## Blockchain Interaction
- Support and Grievance officers utilize read-only smart contract verification tools to validate disputed transactions:
 - Queries `SettlementDvP.sol` with the trade UUID to inspect on-chain execution status, gas usage, and block timestamp.
 - Queries `ComplianceRegistry.sol` to verify if an account was frozen or blacklisted at the exact timestamp an order was rejected.
 - Queries `ProofOfReserveRegistry.sol` to prove that fractional asset balances were fully backed in physical custody at the trade settlement date.
 - Exports verifiable cryptographic receipts (tx hash, block header, validator signatures) attached to the regulatory dispute evidence dossier.

## Step-by-Step Build Instructions
1. Scaffold the `services/grievance-service/` repository with standard FastAPI directory structure and database migrations.
2. Define the PostgreSQL schema for tickets, dispute categories, attachments, SLA timers, and escalation levels.
3. Configure Temporal workflows (`workflows/grievance_sla_workflow.py`) implementing automated escalation timers: Day 2 (Warning to Lead), Day 5 (Escalate to Grievance Officer), Day 14 (Escalate to Compliance Officer), Day 21 (SEBI Hard Breach).
4. Implement the Dispute Evidence Builder (`services/dispute_evidence.py`) that aggregates user profile, bank payment logs, order matching logs, and Besu on-chain settlement receipts into a tamper-evident PDF/JSON dossier.
5. Build the SEBI SCORES 2.0 API connector (`integrations/sebi_scores_client.py`) supporting bi-directional ticket creation, status synchronization, and resolution submission.
6. Build the SMART ODR API connector (`integrations/smart_odr_client.py`) for automated dispute referral when mutual resolution fails.
7. Implement the Flutter in-app support module (`lib/features/support/`) with categorized dispute intake forms (e.g., "Funds Deposited but Not Credited", "Order Executed with Unexpected Price", "Dividend Not Received").
8. Implement the Next.js Support Agent Console (`apps/admin-support/`) with role-based PII masking (Tier-1 sees masked PAN `XXXXX1234F` and masked bank account).
9. Integrate with Notification Service (Prompt 211) to send automated SMS, email, and push notifications to investors upon every ticket status change.
10. Implement the Monthly Regulatory Grievance Report generator exporting SEBI Form 1A compliance tables (opened, resolved, pending > 7 days, pending > 15 days).
11. Build an automated audit logger recording every agent view, comment, and resolution action in an immutable append-only audit store.
12. Conduct end-to-end integration tests simulating a disputed trade, automated evidence package generation, SCORES ticket sync, and compliant resolution within SLA.

## Interfaces / Contracts
```python
# Grievance & Dispute Service Contract (services/grievance-service/models.py)
from pydantic import BaseModel, Field
from datetime import datetime
from enum import Enum

class DisputeCategory(str, Enum):
    FUNDS_DEPOSIT_DELAY = "FUNDS_DEPOSIT_DELAY"
    FUNDS_WITHDRAWAL_DELAY = "FUNDS_WITHDRAWAL_DELAY"
    ORDER_EXECUTION_DISPUTE = "ORDER_EXECUTION_DISPUTE"
    FRACTIONAL_SETTLEMENT_ISSUE = "FRACTIONAL_SETTLEMENT_ISSUE"
    CORPORATE_ACTION_DIVIDEND = "CORPORATE_ACTION_DIVIDEND"
    KYC_REJECTION_APPEAL = "KYC_REJECTION_APPEAL"
    UNAUTHORIZED_ACCESS_CLAIM = "UNAUTHORIZED_ACCESS_CLAIM"

class GrievanceEscalationTier(str, Enum):
    TIER_1_CUSTOMER_CARE = "TIER_1_CUSTOMER_CARE"
    TIER_2_GRIEVANCE_OFFICER = "TIER_2_GRIEVANCE_OFFICER"
    TIER_3_PRINCIPAL_COMPLIANCE = "TIER_3_PRINCIPAL_COMPLIANCE"
    TIER_4_SEBI_SCORES_ODR = "TIER_4_SEBI_SCORES_ODR"

class DisputeEvidencePackage(BaseModel):
    ticket_id: str
    investor_id: str
    dispute_category: DisputeCategory
    order_id: str | None
    trade_id: str | None
    besu_tx_hash: str | None
    block_number: int | None
    bank_reference_number: str | None
    nsdl_custody_cert_id: str | None
    timeline_events: list[dict]
    generated_at: datetime
```

```yaml
# SEBI SCORES 2.0 Webhook Payload Schema
type: object
properties:
  scores_complaint_id: { type: string }
  growww_ticket_id: { type: string }
  investor_pan_masked: { type: string }
  complainant_name: { type: string }
  complaint_description: { type: string }
  sebi_deadline_date: { type: string, format: date }
  status: { type: string, enum: ["PENDING_INTERMEDIARY", "CLARIFICATION_REQUESTED", "DISPOSED", "ESCALATED_ODR"] }
required:
 - scores_complaint_id
 - sebi_deadline_date
 - status
```

## Security & Compliance Notes
- Strictly complies with the SEBI Master Circular on Investor Grievance Redressal Mechanism and SMART ODR circulars.
- Tier-1 support staff are restricted from viewing unmasked PANs, Aadhaar numbers, or full bank account details.
- Every modification to ticket status, internal note, or external communication is cryptographically timestamped and stored in the immutable audit log.

## Acceptance Criteria
- [ ] In-app Flutter grievance filing flow operates seamlessly across mobile and desktop clients with document upload capability.
- [ ] Temporal workflow enforces SLA timers and escalates overdue tickets to the Grievance Redressal Officer automatically.
- [ ] Dispute Evidence Package Generator automatically links bank gateway logs, matching engine trades, and Besu on-chain transaction hashes.
- [ ] SEBI SCORES 2.0 and SMART ODR API adapters successfully sync complaints and submit resolutions.
- [ ] Monthly SEBI Form 1A investor grievance compliance reports generate automatically with zero manual spreadsheet reconciliation.

## Suggested Order / Dependencies
- **Prerequisites:** 201 (User Service), 208 (Settlement Service), 218 (Audit Log Service), 604 (Admin Console), 908 (Production Launch).
- **Parallel Tasks:** 909 (Post-Launch Monitoring & SLOs).
