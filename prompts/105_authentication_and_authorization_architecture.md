# 105 - Authentication & Authorization Architecture (OAuth2/OIDC, RBAC/ABAC & mTLS Service Mesh)

## Purpose
Establishes the enterprise-grade zero-trust identity, authentication, and fine-grained authorization architecture for the Growww investment platform. Growww operates a high-security financial infrastructure where identity must be securely asserted across 5 client platforms (Flutter native mobile/desktop and web), internal polyglot microservices, administrative back-office portals, and permissioned blockchain transactions.

This prompt provides developers and security engineers with the unified architecture for OpenID Connect (OIDC) / OAuth 2.1 authentication, short-lived cryptographic JWT tokens, biometric MPIN / WebAuthn device authentication, Open Policy Agent (OPA) policy engines for fine-grained Attribute-Based Access Control (ABAC), and Istio mTLS SPIFFE/SPIRE identities for service-to-service communication.

## What You Are Building
A comprehensive IAM architecture specification (`docs/architecture/auth_and_iam.md`), token schemas, Open Policy Agent (OPA) Rego policy bundles, and Envoy Gateway JWT filter configurations:
- External User Authentication: OAuth 2.1 / OIDC architecture with Proof Key for Code Exchange (PKCE), biometric login (FIDO2 / WebAuthn), and short-lived Ed25519 JWT access tokens (15-minute TTL) with secure refresh token rotation.
- Internal Service Authentication: Zero-trust service mesh with Istio mTLS and SPIFFE/SPIRE cryptographic identities (`spiffe://growww.internal/ns/{namespace}/sa/{service-account}`).
- Fine-Grained Authorization: Role-Based Access Control (RBAC) and Attribute-Based Access Control (ABAC) using Open Policy Agent (OPA) with policies written in Rego.
- Administrative Governance: Multi-Factor Authentication (MFA), Hardware FIDO2 security keys, and dual-control Maker-Checker authorization workflows for sensitive operations.

## Scope Boundaries
- **In Scope:**
 - Complete OAuth 2.1 / OIDC flow design for Flutter, Web, and Admin applications.
 - JWT token structure, signing key management, and claims taxonomy.
 - Service-to-service mTLS authentication and SPIFFE ID schemas.
 - OPA Rego policy bundles for RBAC/ABAC across all API routes.
 - Token revocation and blacklisting via Redis.
- **Out of Scope / Handled Elsewhere:**
 - Concrete User Service implementation (handled in Prompt 201).
 - Admin Console UI development (handled in Prompt 604 & 605).
 - Secrets and KMS management (handled in Prompt 109).
 - Internal staff directory & corporate IAM (handled in Prompt 702).

## Technology to Use
- **Identity & Protocol Standards:** OAuth 2.1, OpenID Connect Core 1.0, RFC 7519 (JWT), RFC 7636 (PKCE), FIDO2 / WebAuthn.
- **Cryptographic Algorithms:** Ed25519 (Asymmetric JWT signing and verification) and AES-256-GCM (Session / Token encryption).
- **Policy Engine:** Open Policy Agent (OPA) v0.60+ embedded as Envoy sidecar filter and Go library.
- **Service Mesh & Identity:** Istio 1.21+ with SPIFFE/SPIRE mTLS certificate issuance and automatic rotation.
- **Token State & Blacklisting:** Redis Enterprise Cluster for distributed, sub-millisecond token blacklisting and session revocation.

## Backend / Infra Touchpoints
- **Envoy API Gateway:** Validates incoming JWT tokens at the perimeter, extracts claims into internal gRPC metadata, and invokes OPA authorization filters.
- **User & Auth Service (FastAPI):** Issues JWT tokens, manages refresh tokens in PostgreSQL, and interacts with Redis.
- **HashiCorp Vault / CloudHSM:** Securely generates and rotates Ed25519 private keys used for signing JWT tokens.
- **Istio Citadel / SPIRE Agent:** Automatically injects and rotates X.509 client certificates on every Kubernetes pod.

## Blockchain Interaction
Binds off-chain investor identity to on-chain permissioned smart contract interactions:
- **On-Chain Identity Claims:** In line with ERC-3643 (T-REX standard), the authentication service validates that the authenticated `user_id` possesses an active, compliant identity claim on the `ComplianceRegistry.sol` smart contract.
- **EIP-712 Signature Delegation:** For on-chain actions (DvP approvals, order authorizations), client devices sign EIP-712 structured data hashes using keys stored in the device Secure Enclave / KeyStore (Prompt 521), forwarded by the backend transaction relayer.
- **Zero PII on Ledger:** Only cryptographic hashes of KYC approval tokens (`claim_hash`) and pseudo-anonymous Ethereum addresses are linked to on-chain whitelist contracts; user identities and credentials remain strictly off-chain.

## Step-by-Step Build Instructions
1. Initialize documentation file `docs/architecture/auth_and_iam.md` and policy folder `policies/opa/`.
2. Define the OAuth 2.1 authentication flow with PKCE for Flutter mobile/desktop clients and Next.js web applications.
3. Design the JWT token claim schema:
 - Header: `alg: "EdDSA"`, `typ: "JWT"`, `kid: "<key-fingerprint>"`.
 - Standard Claims: `iss`, `sub` (User UUID), `aud`, `exp` (15 minutes), `nbf`, `iat`, `jti`.
 - Growww Custom Claims: `entity` (`DOMESTIC` / `GIFT_CITY`), `kyc_tier` (`TIER_1` / `TIER_2`), `roles` (`INVESTOR`, `TRADER`, `COMPLIANCE_OFFICER`, `ADMIN`), `device_id`, `mfa_verified`.
4. Establish the Refresh Token Rotation (RTR) lifecycle: Refresh tokens are single-use, stored as argon2id hashes in PostgreSQL with 30-day absolute expiration; reuse detection triggers immediate revocation of all active sessions for that user.
5. Configure Redis token blacklisting: When a user logs out or is suspended, the `jti` is stored in Redis with TTL equal to the remaining access token lifetime.
6. Design the Istio mTLS zero-trust service mesh configuration: Enforce `PeerAuthentication` mode `STRICT` across all Kubernetes namespaces (`growww-core`, `growww-trading`, `growww-settlement`).
7. Formulate SPIFFE ID conventions: `spiffe://growww.internal/ns/{namespace}/sa/{service-account-name}`.
8. Define the RBAC / ABAC permission model matrix mapping roles to actions and resource scopes.
9. Author standard OPA Rego policy bundles (`policies/opa/authz.rego`) for:
 - Validating JWT integrity and claim constraints.
 - Restricting trading actions to users with verified KYC status (`kyc_tier >= TIER_1`).
 - Restricting entity boundaries (Domestic investors cannot access GIFT City endpoints directly and vice versa).
 - Enforcing Maker-Checker authorization on administrative back-office endpoints.
10. Define Envoy Gateway filter configuration integrating JWT authentication and external OPA authorization checks.
11. Design biometric authentication integration (WebAuthn / FIDO2 / Passkeys) allowing passwordless biometric authentication backed by hardware keystores.
12. Establish automated JWT signing key rotation runbook (quarterly key rotation via Vault with dual-key verification grace period).
13. Perform security threat modeling against token theft, session hijacking, replay attacks, and privilege escalation.
14. Publish `docs/architecture/auth_and_iam.md` to repository.

## Interfaces / Contracts

### JWT Payload Specification
```json
{
  "iss": "https://auth.growww.in",
  "sub": "usr_01HZX89AB72K9M12P5QRSTUVWX",
  "aud": ["https://api.growww.in"],
  "exp": 1718812800,
  "nbf": 1718811900,
  "iat": 1718811900,
  "jti": "tok_01HZX89AB98CDEF1234567890A",
  "entity": "DOMESTIC",
  "kyc_status": "APPROVED",
  "kyc_tier": "TIER_2",
  "roles": ["INVESTOR"],
  "permissions": ["trading:orders:create", "trading:orders:cancel", "wallet:deposit", "wallet:withdraw"],
  "device_id": "dev_01HZX77BC9900AABBCCDDEEFF",
  "mfa_authenticated": true,
  "wallet_address": "0x71C...3a9"
}
```

### OPA Authorization Policy (`policies/opa/authz.rego`)
```rego
package growww.authz

import future.keywords.in

default allow = false

# Allow public health check and metrics endpoints
allow {
    input.path = ["healthz"]
}

# General API Authorization Logic
allow {
    # 1. Verify token is valid and not expired
    input.jwt.valid == true
    
    # 2. Match entity boundary
    input.jwt.payload.entity == input.headers["x-growww-entity"]
    
    # 3. Check role and permission
    required_permission := get_required_permission(input.method, input.path)
    required_permission in input.jwt.payload.permissions
    
    # 4. Enforce KYC verification for trading operations
    is_trading_action(input.path)
    input.jwt.payload.kyc_status == "APPROVED"
}

# Helper to identify trading paths
is_trading_action(path) {
    path[0] == "api"
    path[1] == "v1"
    path[2] == "orders"
}

# Helper mapping method & path to permission
get_required_permission("POST", ["api", "v1", "orders"]) = "trading:orders:create"
get_required_permission("DELETE", ["api", "v1", "orders", _]) = "trading:orders:cancel"
get_required_permission("GET", ["api", "v1", "portfolio"]) = "portfolio:read"
```

### Istio Strict mTLS PeerAuthentication Manifest
```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: growww-core
spec:
  mtls:
    mode: STRICT
---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: order-service-rbac
  namespace: growww-core
spec:
  selector:
    matchLabels:
      app: order-service
  action: ALLOW
  rules:
 - from:
 - source:
        principals: ["cluster.local/ns/growww-edge/sa/envoy-gateway-sa"]
```

## Security & Compliance Notes
- **Zero-Trust Network Principle:** No implicit trust is granted based on internal network IP location. Every single RPC between microservices requires cryptographically validated mTLS and SPIFFE ID authorization.
- **Short-Lived Access Tokens:** Access tokens expire after 15 minutes to minimize exposure in the event of client-side token extraction.
- **Device-Binding & Session Pinning:** Access tokens are bound to cryptographic device fingerprints; tokens presented from unmatched client fingerprints trigger immediate session invalidation and fraud alerts.
- **SEBI Audit Requirements:** All authentication events (logins, failed attempts, password resets, MFA challenges, permission denials) must be immutably recorded in the security audit log.

## Acceptance Criteria
- [ ] Complete Authentication & Authorization architecture document (`docs/architecture/auth_and_iam.md`) committed.
- [ ] OAuth 2.1 + PKCE flow fully documented with sequence diagrams for Flutter and Web clients.
- [ ] Ed25519 JWT claim schema and Refresh Token Rotation (RTR) specification finalized.
- [ ] OPA Rego policy suite (`policies/opa/authz.rego`) authored with unit tests verifying permit/deny rules for investors and admins.
- [ ] Istio STRICT mTLS and SPIFFE ID naming conventions specified across all Kubernetes namespaces.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 103 (API Standards).
- **Parallel Work:** Prompt 109 (Secrets Management), Prompt 110 (Inter-Entity Communication).
- **Blocks:** Prompt 201 (User Service), Prompt 505 (Auth UI), Prompt 604 (Admin Console IAM), Prompt 702 (Internal IAM).
