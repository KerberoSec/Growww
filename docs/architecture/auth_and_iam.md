# Growww Platform Authentication & Authorization (IAM) Architecture

## 1. Zero-Trust Identity & Authentication Core
The Growww / NBSE trading platform implements a defense-in-depth zero-trust security architecture across client perimeters, internal microservices, and permissioned Hyperledger Besu smart contracts.

---

## 2. External Authentication: OAuth 2.1 & OIDC
- **Clients:** Flutter iOS, Android, macOS, Linux, Windows desktop, and Next.js web portal.
- **Grant Types:** Authorization Code Flow with PKCE (RFC 7636) strictly enforced. Client secrets are never embedded in native applications.
- **Biometric & Hardware Keys:** FIDO2 / WebAuthn passwordless authentication with hardware TPM / Secure Enclave binding.
- **Access Tokens:** Ed25519 (EdDSA) signed JWTs with short 15-minute TTL.
- **Refresh Token Rotation (RTR):** Single-use cryptographic tokens with Argon2id hashing in PostgreSQL. Reuse detection immediately invalidates all active sessions for the compromised user.

### 2.1 Canonical JWT Claims Schema
```json
{
  "iss": "https://auth.growww.trade",
  "sub": "usr_018d9f45-728b-7000-848f-39589dfd0001",
  "aud": "https://api.growww.trade",
  "exp": 1774088900,
  "nbf": 1774088000,
  "iat": 1774088000,
  "jti": "jti_884920184",
  "entity": "DOMESTIC",
  "kyc_tier": "TIER_2",
  "roles": ["INVESTOR", "PRO_TRADER"],
  "device_id": "dev_mac_arm64_001",
  "mfa_verified": true
}
```

---

## 3. Internal Service-to-Service: Istio mTLS & SPIFFE
- All Kubernetes namespaces enforce `PeerAuthentication` mode `STRICT`.
- Service accounts receive automatic X.509 certificates via SPIFFE/SPIRE with 12-hour rotation:
  - `spiffe://growww.internal/ns/growww-trading/sa/matching-engine`
  - `spiffe://growww.internal/ns/growww-settlement/sa/besu-relayer`

---

## 4. Fine-Grained Authorization: Open Policy Agent (OPA)
- Envoy Gateway delegates incoming HTTP/gRPC requests to OPA sidecar filter evaluating Rego policies:
  - Role-Based Access Control (RBAC): `INVESTOR`, `PRO_TRADER`, `MARKET_MAKER`, `COMPLIANCE_OFFICER`, `SUPER_ADMIN`.
  - Attribute-Based Access Control (ABAC): Dynamic checks including entity jurisdiction matching, KYC tier minimums, and daily loss circuit limits.
