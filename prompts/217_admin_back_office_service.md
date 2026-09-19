# 217 - Admin & Back-Office Service (Exception Handling & Dual-Control Operations)

## Purpose
The Admin & Back-Office Service provides the operational control plane for Growww's platform operators, compliance officers, risk managers, and exchange administrators. In a regulated financial institution, no single individual can possess unilateral authority to perform sensitive financial or administrative actions (such as freezing user assets, overriding KYC status, manual deposit crediting, or executing corporate action adjustments).

This service enforces strict Maker-Checker (dual-control / four-eyes principle) workflows, granular Role-Based Access Control (RBAC), multi-party on-chain governance triggers, and comprehensive audit trail instrumentation for all operational interventions and exception management cases.

## What You Are Building
A secure, high-assurance Go microservice (`services/admin-service`) providing:
- **Dual-Control (Maker-Checker) Workflow Engine:** Configurable multi-signature approval pipelines for administrative state changes.
- **Regulatory Account Freeze & Lien Manager:** Enforces court orders, SEBI mandates, and AML risk-based account restrictions.
- **Exception Resolution Desk:** Tools for investigating and resolving mismatched deposits, failed settlements, and disputed transactions.
- **On-Chain Governance Proposer:** Interfaces with Hyperledger Besu multisig contracts (`ComplianceRegistry.sol`, `MultiSigGovernance.sol`) to propose and execute administrative ledger operations.
- **Artifacts Delivered:**
 - `services/admin-service/cmd/server/main.go` - Go microservice entry point.
 - `services/admin-service/internal/workflow/maker_checker.go` - Dual-control approval state machine.
 - `services/admin-service/internal/rbac/casbin_enforcer.go` - RBAC / ABAC security enforcer.
 - `services/admin-service/internal/chain/multisig_proposer.go` - On-chain multisig proposal client.
 - `proto/growww/admin/v1/admin.proto` - Internal gRPC service definitions.

## Scope Boundaries
- **In Scope:**
 - Maker-Checker approval state machine across all platform operations.
 - Granular RBAC / Attribute-Based Access Control (ABAC) enforcement with WebAuthn/FIDO2 MFA.
 - Account freezing, unfreezing, and lien placement across off-chain and on-chain registries.
 - Admin incident resolution logging and audit trail publishing.
- **Out of Scope / Handled Elsewhere:**
 - Admin Web UI interface implementation (handled in Category 6, Prompt 603/604).
 - Raw cryptographic key generation and validator management (handled by Prompt 311).
 - Continuous background reconciliation (handled by Prompt 215).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `casbin/v2` for authorization, `pgx/v5` for database persistence, and `go-webauthn/webauthn` for hardware token verification.
- **Justification:** Go delivers high security, type safety, minimal attack surface, fast execution, and straightforward concurrency management necessary for building mission-critical back-office services where unauthorized actions could lead to catastrophic financial or legal liability.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for storing admin user identities, roles, workflow tickets, and resolution history.
 - Redis 7.2+ for ephemeral approval session management and distributed locks.
 - `github.com/casbin/casbin/v2` for flexible RBAC/ABAC policy evaluation.
 - Apache Kafka 3.7+ for emitting administrative event notices.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Admin user roles, approval tickets, workflow logs, and exception cases.
- **Redis 7.2:** Session storage with mandatory hardware MFA binding and workflow lock states.
- **Kafka Topics:**
 - Publishes: `admin.action.proposed`, `admin.action.approved`, `admin.account.frozen`, `admin.exception.resolved`.
 - Subscribes: `kyc.manual_review.required`, `reconciliation.discrepancy.flagged`.
- **Hyperledger Besu (QBFT):** Submits multi-sig proposals to `ComplianceRegistry.sol` and `MultiSigGovernance.sol`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium ledger running QBFT consensus.
- **Multisig Proposal Execution:** When an administrator freezes an investor account or pauses token transfers for an ISIN:
  1. The Maker proposes the action via the Admin Service.
  2. The Checker approves the action using their independent cryptographic credential.
  3. Once dual approval is confirmed off-chain, the service instructs the HSM relayer to call `proposeTransaction()` on `MultiSigGovernance.sol`.
  4. Second authorized officer signs on-chain via `confirmTransaction()`, executing `freezeWallet()` on `ComplianceRegistry.sol`.
- **Zero PII on Ledger:** Only public wallet addresses, proposal hashes, and administrator public key identifiers are stored on-chain.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/admin-service` adhering to standard directory layout (Prompt 106).
2. Define Protobuf definitions in `proto/growww/admin/v1/admin.proto` and generate Go gRPC stubs.
3. Configure PostgreSQL schema migrations for `admin_users`, `admin_roles`, `workflow_proposals`, and `exception_tickets`.
4. Implement Casbin RBAC/ABAC authorization model defining roles: `SUPPORT_REP`, `COMPLIANCE_MAKER`, `COMPLIANCE_CHECKER`, `SUPER_ADMIN`.
5. Implement WebAuthn / FIDO2 hardware token authentication interceptor for all administrative endpoints.
6. Build the Maker-Checker workflow engine requiring distinct identities for creation (`Maker`) and authorization (`Checker`).
7. Implement the Account Freeze & Lien Manager orchestrating lockouts across Auth, Wallet, and Blockchain registries.
8. Implement the Exception Resolution desk for manual payment reconciliations with mandatory document attachment proofs.
9. Implement on-chain multisig proposal submission client interacting with Hyperledger Besu smart contracts via HSM.
10. Integrate Kafka event publishers producing `admin.action.*` events consumed by Audit Log Service (Prompt 218).
11. Add Prometheus metrics (`admin_proposals_pending`, `admin_action_latency_seconds`) and health checks.
12. Write comprehensive unit and integration tests enforcing maker-checker separation, privilege escalation defense, and error handling.

## Interfaces / Contracts

### Protobuf Definition (`admin.proto`)
```protobuf
syntax = "proto3";

package growww.admin.v1;

option go_package = "github.com/growww/services/admin-service/gen/v1;adminv1";

service AdminBackOfficeService {
  rpc CreateWorkflowProposal (CreateProposalRequest) returns (CreateProposalResponse);
  rpc ApproveWorkflowProposal (ApproveProposalRequest) returns (ApproveProposalResponse);
  rpc RejectWorkflowProposal (RejectProposalRequest) returns (RejectProposalResponse);
  rpc FreezeUserAccount (FreezeAccountRequest) returns (FreezeAccountResponse);
  rpc ResolveException (ResolveExceptionRequest) returns (ResolveExceptionResponse);
}

message CreateProposalRequest {
  string action_type = 1; // FREEZE_ACCOUNT / UNFREEZE / MANUAL_LEDGER_ADJUST / ASSET_PAUSE
  string target_resource_id = 2;
  string justification = 3;
  string supporting_document_hash = 4;
  string maker_admin_id = 5;
}

message CreateProposalResponse {
  string proposal_id = 1;
  string status = 2; // PENDING_APPROVAL
  int64 created_at = 3;
}

message ApproveProposalRequest {
  string proposal_id = 1;
  string checker_admin_id = 2;
  string fido2_auth_token = 3;
  string comments = 4;
}

message ApproveProposalResponse {
  string proposal_id = 1;
  string status = 2; // APPROVED / EXECUTED / ON_CHAIN_PENDING
  string on_chain_tx_hash = 3;
  int64 executed_at = 4;
}

message RejectProposalRequest {
  string proposal_id = 1;
  string checker_admin_id = 2;
  string rejection_reason = 3;
}

message RejectProposalResponse {
  string proposal_id = 1;
  string status = 2; // REJECTED
  int64 rejected_at = 3;
}

message FreezeAccountRequest {
  string user_id = 1;
  string freeze_scope = 2; // TOTAL / TRADING_ONLY / WITHDRAWAL_ONLY
  string regulatory_notice_ref = 3;
  string maker_admin_id = 4;
}

message FreezeAccountResponse {
  string proposal_id = 1;
  string status = 2;
}

message ResolveExceptionRequest {
  string exception_id = 1;
  string resolution_action = 2;
  string adjustment_amount_inr = 3;
  string maker_admin_id = 4;
}

message ResolveExceptionResponse {
  string proposal_id = 1;
  string status = 2;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE admin_users (
    admin_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(128) UNIQUE NOT NULL,
    role VARCHAR(32) NOT NULL CHECK (role IN ('SUPPORT_REP', 'COMPLIANCE_MAKER', 'COMPLIANCE_CHECKER', 'SUPER_ADMIN')),
    fido2_credential_id BYTEA NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow_proposals (
    proposal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_type VARCHAR(64) NOT NULL,
    target_resource_id VARCHAR(128) NOT NULL,
    justification TEXT NOT NULL,
    maker_id UUID NOT NULL REFERENCES admin_users(admin_id),
    checker_id UUID REFERENCES admin_users(admin_id),
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_APPROVAL',
    on_chain_proposal_id NUMERIC(78, 0),
    on_chain_tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_maker_checker_distinct CHECK (maker_id != checker_id)
);

CREATE TABLE exception_tickets (
    ticket_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(64) NOT NULL,
    source_service VARCHAR(64) NOT NULL,
    payload_json JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    assigned_to UUID REFERENCES admin_users(admin_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **Strict Dual-Control Enforcement:** The database and application layers enforce `maker_id != checker_id` at all times, making self-approval impossible.
- **Hardware FIDO2 MFA:** All Checker approvals strictly require WebAuthn / FIDO2 hardware cryptographic confirmation.
- **Privilege Separation:** Support representatives cannot access KYC document unmasked details; only Compliance Officers possess unmasking privileges with mandatory justification logging.
- **Immutable Action Auditing:** Every administrative interaction is published to the Audit Log Service (Prompt 218) and retained under regulatory compliance guidelines.

## Acceptance Criteria
- [ ] Go admin service compiles cleanly with strict static analysis and zero linter warnings.
- [ ] Maker-Checker workflow completely blocks attempts by a Maker to approve their own proposal.
- [ ] FIDO2 authentication middleware correctly validates WebAuthn hardware credentials.
- [ ] Account freeze triggers immediate restriction propagation to User Service, Wallet Service, and Blockchain registry.
- [ ] On-chain multisig proposals are generated and confirmed via test Besu node contracts.
- [ ] Comprehensive unit and integration test suite achieves >=85% test coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 105 (Auth Architecture), Prompt 201 (User Service), Prompt 307 (Multi-Party Governance).
- **Subsequent / Parallel Tasks:** Prompt 218 (Audit Log Service), Prompt 603 (Admin Portal Web).
