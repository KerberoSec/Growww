# 702 - Identity & Access Management (IAM), RBAC & Privileged Access Management

## Purpose
Architects enterprise-grade Identity and Access Management (IAM), fine-grained Role-Based Access Control (RBAC), Attribute-Based Access Control (ABAC), and Privileged Access Management (PAM) across all internal Growww operations. Ensures strict enforcement of the principle of least privilege, zero persistent administrative credentials in production, mandatory multi-party approvals (dual-control / "four-eyes" principle) for high-impact operations (e.g., account freezes, manual ledger adjustments, token pausing), and complete segregation of duties (SoD) between developers, trade operations, compliance officers, and custody auditors in strict compliance with SEBI and RBI mandates.

## What You Are Building
- **IAM & Access Policy Blueprint:** `docs/security/iam_and_rbac_policy.md` detailing the master role matrix, permission sets, session lifetime rules, and separation of duties matrix.
- **Keycloak / IdP Realm Configuration:** Declarative realm definitions, client scopes, user federation (SAML 2.0 / OIDC), and WebAuthn / FIDO2 hardware MFA enforcement configurations.
- **Open Policy Agent (OPA) Rego Policies:** Centralized Rego authorization policies (`policies/rbac/*.rego`) evaluating fine-grained REST/gRPC API request permissions and multi-party quorum states.
- **Privileged Access Management (PAM) Workflow:** Teleport / HashiCorp Boundary configuration enabling ephemeral, audited Just-In-Time (JIT) access to production clusters and databases with auto-expiring certificates (<= 1 hour).
- **Dual-Control Quorum Interceptor:** Middleware / gRPC interceptor verifying multi-signature authorization tokens for critical back-office administrative actions before downstream execution.

## Scope Boundaries
- **In Scope:**
 - Internal staff IAM, role matrices, permission hierarchies, and group inheritance.
 - Mandatory hardware-backed MFA (FIDO2/WebAuthn YubiKeys) for all administrative and operational access.
 - Decoupled authorization engine using Open Policy Agent (OPA) with API Gateway sidecar integration.
 - Ephemeral Just-In-Time (JIT) access automation for break-glass operational troubleshooting.
 - Multi-party approval ("four-eyes") authorization flows for privileged operations.
- **Out of Scope / Handled Elsewhere:**
 - Retail and institutional investor authentication/registration (handled in Prompts 201 and 505).
 - Admin Console UI web interfaces (handled in Prompts 604 and 605).
 - Low-level secrets storage and KMS/Vault infrastructure provisioning (handled in Prompts 109 and 805).

## Technology to Use
- **Identity Provider (IdP):** Keycloak 24+ / Okta enterprise identity federation supporting OIDC, SAML 2.0, SCIM provisioning, and WebAuthn/FIDO2 hardware security keys.
- **Authorization Engine:** Open Policy Agent (OPA) / Rego evaluating sub-millisecond decision policies at the API Gateway and service boundaries.
- **Privileged Access Management (PAM):** Teleport 15+ / HashiCorp Boundary for certificate-based, identity-aware SSH/Kubernetes/Database access with full session recording.
- **Justification:** OPA decouples business logic from security authorization rules, allowing compliance policies to be updated, unit-tested, and audited independently without redeploying application microservices. Teleport eliminates static SSH keys and long-lived database credentials across production infrastructure.

## Backend / Infra Touchpoints
- **API Gateway / Admin BFF:** Kong / Envoy Gateway (219) integrated with OPA sidecar.
- **Admin & Back-Office Service:** Admin Service (217), User Service (201), KYC Service (202).
- **Identity & Vault Infrastructure:** Keycloak IdP Cluster, HashiCorp Vault (109), AWS IAM / CloudTrail.
- **Audit Logging Mesh:** Centralized Audit Service (218) consuming `audit.iam.events` Kafka topic.

## Blockchain Interaction
Role mapping to on-chain multisig signers and administrative gates:
- **On-Chain Role Segregation:** Administrative actions requiring on-chain execution - such as blacklisting/freezing an investor address in `ComplianceRegistry.sol`, emergency pausing `DigitalSecurityToken.sol`, or registering a new token series - are strictly gated behind `MultiSigGovernance.sol`.
- **Hardware Signer Quorums:** The IAM system maps the "Compliance Officer" and "Chief Security Officer" roles to individual FIPS 140-2 Level 3 HSM-backed Ethereum signing keys. An on-chain administrative action requires an $M$-of-$N$ threshold (e.g., 3-of-5 signers) where no single internal operator possesses unilateral execution authority.
- **Audit Linkage:** Every internal approval workflow generates a cryptographic approval bundle recorded in the off-chain audit ledger (Prompt 218) with a SHA-256 hash anchored to the on-chain multisig transaction payload.

## Step-by-Step Build Instructions
1. Author the master IAM specification under `docs/security/iam_and_rbac_policy.md`, defining the complete Role Matrix: Super Admin, Compliance Officer, AML Analyst, Trade Operations Lead, Custody Auditor, Support L1/L2, and Developer/SRE.
2. Establish the Separation of Duties (SoD) conflict table (e.g., an AML Analyst cannot execute trade adjustments; Developers cannot hold production data write roles).
3. Configure Keycloak realms, defining custom user attributes, role mappings, composite roles, and mandatory WebAuthn FIDO2 MFA execution.
4. Implement the Open Policy Agent (OPA) Rego policy bundle (`policies/rbac/authz.rego`) defining role-to-endpoint access rules, HTTP method restrictions, and data-level tenant isolation.
5. Write unit tests for OPA Rego policies using `opa test`, ensuring 100% test coverage over positive and negative authorization cases.
6. Integrate OPA as an Envoy/Kong sidecar filter for the Admin API Gateway (Prompt 219), enforcing policy checks on every incoming request.
7. Implement the Dual-Control / Four-Eyes authorization interceptor in the Admin Service (Prompt 217), requiring two distinct authorized signatures for actions flagged as `HIGH_IMPACT` (e.g., user account unfreezes, balance reconciliations).
8. Configure Teleport PAM for production Kubernetes clusters and PostgreSQL instances, enforcing short-lived X.509 client certificates (validity max 60 minutes) requested via Slack/Jira approval workflows.
9. Configure automated SCIM provisioning and de-provisioning workflows, ensuring immediate account revocation across all services upon employee off-boarding.
10. Implement immutable IAM audit event emission to Kafka (`audit.iam.events`), capturing user ID, source IP, role, requested action, approval chain, and cryptographic signature.
11. Implement an automated quarterly access review and recertification pipeline generating access attestation reports for the CISO.
12. Conduct automated end-to-end authorization tests simulating privilege escalation, expired token usage, and SoD violation attempts.

## Interfaces / Contracts
```rego
# policies/rbac/authz.rego
package growww.authz

default allow = false

# Role Definitions & Hierarchy
user_roles := input.user.roles
is_compliance_officer { user_roles[_] == "ROLE_COMPLIANCE_OFFICER" }
is_super_admin { user_roles[_] == "ROLE_SUPER_ADMIN" }
is_trade_ops { user_roles[_] == "ROLE_TRADE_OPS" }

# Read-only access for compliance
allow {
    input.method == "GET"
    startswith(input.path, "/api/v1/compliance/")
    is_compliance_officer
}

# High-impact operations require verified dual-control multi-party approval
allow {
    input.method == "POST"
    input.path == "/api/v1/compliance/freeze-investor"
    is_compliance_officer
    has_valid_dual_approval(input.action_id, input.user.id)
}

has_valid_dual_approval(action_id, requester_id) {
    approval := data.approvals[action_id]
    approval.status == "APPROVED"
    approval.approver_id != requester_id # Enforce strict 4-eyes separation
    approval.approver_role == "ROLE_COMPLIANCE_LEAD"
    time.now_ns() < approval.expires_at_ns
}
```

```protobuf
// schemas/iam/v1/iam_events.proto
syntax = "proto3";
package growww.iam.v1;

message IamAuditEvent {
  string event_id = 1;
  string timestamp_utc = 2;
  string actor_user_id = 3;
  string actor_email = 4;
  repeated string actor_roles = 5;
  string action_name = 6;
  string resource_uri = 7;
  string client_ip = 8;
  string mfa_type = 9; // e.g. "WEBAUTHN_FIDO2"
  bool authorization_decision = 10;
  repeated string dual_control_approvers = 11;
  string digital_signature_sha256 = 12;
}
```

## Security & Compliance Notes
- **SEBI CSCRF Compliance:** Strictly satisfies Section 3 (Access Control and Identity Management) requiring hardware MFA, segregation of operational environments, and quarterly access reviews.
- **CERT-In Guidelines:** Enforces centralized identity logging and immediate revocation capabilities for suspected compromised identities.
- **Zero Standing Privileges:** Developers and operators have zero persistent write or SSH access to production infrastructure; all access is ephemeral and session-recorded.

## Acceptance Criteria
- [ ] Master IAM specification and Separation of Duties matrix published in `docs/security/iam_and_rbac_policy.md`.
- [ ] Keycloak realm configured with WebAuthn/FIDO2 hardware MFA enforced for 100% of internal staff.
- [ ] OPA Rego policies authored, tested (`opa test` passing with 100% coverage), and integrated into API Gateway sidecars.
- [ ] Dual-control four-eyes approval workflow verified for all high-impact operational and compliance endpoints.
- [ ] Teleport PAM operational with session recording and max 60-minute ephemeral certificate TTLs.
- [ ] Automated SCIM de-provisioning verified: test account access revoked across all microservices in < 5 seconds.
- [ ] All administrative actions emit structured, immutable audit records to Kafka `audit.iam.events`.

## Suggested Order / Dependencies
- **Prerequisites:** 105 (Auth Architecture), 109 (Secrets Management), 217 (Admin Back-Office Service).
- **Parallel Tasks:** 701 (Threat Model), 707 (Data Encryption & Key Custody).
- **Downstream Dependents:** 604 (Admin User Review), 605 (Admin Exception & Multi-Party Approval UI), 802 (Kubernetes RBAC), 807 (Centralized Logging).
