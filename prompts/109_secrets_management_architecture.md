# 109 - Secrets Management, Encryption Keys & HSM Architecture

## Purpose
Establishes the enterprise secrets lifecycle, Hardware Security Module (HSM) key custody, automated secret rotation, and dynamic credential architecture for the Growww investment platform. Growww operates in a high-security financial domain governed by SEBI and RBI cybersecurity mandates, requiring zero plaintext secrets in source code, configuration files, container images, or memory dumps.

This prompt provides security architects, DevOps engineers, and backend developers with the end-to-end architecture for HashiCorp Vault integration, Kubernetes External Secrets Operator (ESO), dynamic short-lived database credentials, and FIPS 140-2 Level 3 Hardware Security Modules (AWS CloudHSM / Vault Transit Engine) for validator and transaction relayer key custody.

## What You Are Building
A comprehensive secrets and key management specification (`docs/architecture/secrets_management_and_hsm.md`), Vault HCL policy templates, Kubernetes External Secrets manifests, and automated key rotation runbooks:
- Centralized Secrets Management: HashiCorp Vault deployment architecture with auto-unseal via AWS KMS / GCP KMS and multi-region replication.
- Dynamic Database Credentials: Vault database secrets engine generating short-lived (1-hour TTL), dynamically provisioned PostgreSQL credentials for each microservice.
- Kubernetes External Secrets Operator (ESO): Automated synchronization of Vault secrets into native Kubernetes `Secret` objects with automated pod rolling on rotation.
- FIPS 140-2 Level 3 HSM Key Custody: Dedicated CloudHSM / Vault Transit PKI for signing blockchain transactions, validator consensus blocks, and JWT authorization tokens.
- Secret Rotation & Revocation Runbooks: Zero-downtime automated rotation for API keys, database passwords, TLS certificates, and cryptographic signing keys.

## Scope Boundaries
- **In Scope:**
 - HashiCorp Vault architecture, authentication methods, and HCL access policies.
 - HSM hardware integration (PKCS#11 standard) for blockchain signing keys.
 - Dynamic credential generation for PostgreSQL and Redis.
 - Kubernetes External Secrets Operator (ESO) configuration and SecretStore CRDs.
 - Automated key rotation protocols and emergency secret revocation procedures.
- **Out of Scope / Handled Elsewhere:**
 - User session tokens and JWT payload specs (handled in Prompt 105).
 - General data encryption at rest and in transit policy (handled in Prompt 707).
 - Infrastructure-as-Code Terraform provisioning of HSM hardware (handled in Prompt 805).

## Technology to Use
- **Secrets & Key Storage:**
 - *HashiCorp Vault Enterprise (v1.16+):* Secrets engine (KV-v2, Database, Transit, PKI).
 - *AWS CloudHSM / Vault Transit Engine:* FIPS 140-2 Level 3 validated cryptographic hardware.
 - *External Secrets Operator (ESO v0.9+):* Kubernetes operator synchronizing secrets from Vault.
- **Cryptographic Protocols & APIs:**
 - *PKCS#11 / JCE:* Standardized cryptographic token interface for interacting with HSM hardware.
 - *Ed25519 & ECDSA (secp256k1):* Asymmetric signing algorithms for JWTs and Ethereum transactions.
 - *AES-256-GCM:* Envelope encryption for database field-level encryption.

## Backend / Infra Touchpoints
- **Kubernetes Cluster (EKS/GKE):** Injects secrets into microservice pods via Kubernetes native secrets managed by ESO.
- **PostgreSQL Databases:** Configured with Vault database secrets engine for dynamic user creation/revocation.
- **Hyperledger Besu Nodes:** Integrated with HashiCorp Vault Transit / CloudHSM via Web3j / ethsigner for transaction signing.
- **CI/CD Pipelines:** Ephemeral OIDC authentication to Vault for short-lived build and deployment tokens.

## Blockchain Interaction
Defines cryptographic key custody and transaction signing for the permissioned ledger:
- **Validator Consensus Keys:** Besu validator QBFT private keys are generated inside and never exported from FIPS 140-2 Level 3 HSM hardware modules.
- **Transaction Relayer Signing:** Backend transaction relayer services (e.g., DvP settlement executor) sign Ethereum transactions using Vault Transit Engine (`/v1/transit/sign/growww-relayer`) or AWS CloudHSM over PKCS#11; no raw private keys reside on the application server or filesystem.
- **Multisig Governance Key Custody:** Administrative smart contract operations (`MultiSigGovernance.sol`) require M-of-N threshold signatures from designated officers holding independent hardware FIDO2 / HSM tokens.

## Step-by-Step Build Instructions
1. Author `docs/architecture/secrets_management_and_hsm.md` detailing secrets taxonomy, threat model, and HSM custody architecture.
2. Design the Vault cluster topology: 3-node HA deployment with Raft storage, auto-unseal via cloud KMS, and disaster recovery replication.
3. Configure Vault Authentication Engines:
 - `auth/kubernetes`: Service-account-based authentication for microservice pods.
 - `auth/jwt`: OIDC-based authentication for CI/CD runners (GitHub Actions).
 - `auth/approle`: Machine-to-machine authentication for standalone daemon workers.
4. Define least-privilege Vault HCL policies for each microservice bounded context (e.g., `policies/order-service-policy.hcl`).
5. Configure Vault Database Secrets Engine for PostgreSQL:
 - Configure connection to `growww_platform` database.
 - Define dynamic role `order-service-role` with SQL template generating short-lived users (`ttl=1h`, `max_ttl=24h`).
6. Configure Vault Transit Secrets Engine for cryptographic signing and field-level envelope encryption:
 - Create key `growww-jwt-signer` (Ed25519) for auth service.
 - Create key `growww-besu-relayer` (ECDSA secp256k1) for on-chain transaction signing.
7. Configure Kubernetes External Secrets Operator (ESO):
 - Deploy ESO via Helm chart.
 - Create `ClusterSecretStore` manifest referencing Vault Kubernetes auth.
 - Define `ExternalSecret` custom resources for microservice deployments.
8. Implement PKCS#11 / Vault Transit signing client in Go and Rust for the blockchain relayer service.
9. Establish automated key rotation pipelines:
 - Database root credentials rotated every 90 days.
 - JWT signing keys rotated every 90 days with a 7-day dual-verification overlap.
 - API partner secrets rotated every 180 days.
10. Define Emergency Secret Revocation Runbook: Automated script revoking compromised Vault tokens, cycling database users, and blacklisting JWT signing keys.
11. Integrate pre-commit secret scanning (TruffleHog / GitGuardian) to ensure zero credentials are committed to version control.
12. Review compliance against SEBI Cybersecurity and Cyber Resilience Framework guidelines.

## Interfaces / Contracts

### Vault HCL Policy Template (`policies/order-service-policy.hcl`)
```hcl
# Read dynamic database credentials
path "database/creds/order-service-role" {
  capabilities = ["read"]
}

# Read static application secrets
path "secret/data/growww/prod/order-service/*" {
  capabilities = ["read"]
}

# Sign transactions via Transit Engine (zero private key exposure)
path "transit/sign/growww-besu-relayer" {
  capabilities = ["update"]
}

# Encrypt sensitive database fields
path "transit/encrypt/order-field-encryption" {
  capabilities = ["update"]
}

path "transit/decrypt/order-field-encryption" {
  capabilities = ["update"]
}
```

### Kubernetes ExternalSecret CRD Manifest
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: order-service-secrets
  namespace: growww-core
spec:
  refreshInterval: "1h"
  secretStoreRef:
    name: vault-cluster-store
    kind: ClusterSecretStore
  target:
    name: order-service-k8s-secret
    creationPolicy: Owner
  data:
 - secretKey: DATABASE_PASSWORD
      remoteRef:
        key: database/creds/order-service-role
        property: password
 - secretKey: DATABASE_USER
      remoteRef:
        key: database/creds/order-service-role
        property: username
 - secretKey: REDIS_AUTH_TOKEN
      remoteRef:
        key: secret/data/growww/prod/order-service/redis
        property: auth_token
```

## Security & Compliance Notes
- **FIPS 140-2 Level 3 Compliance:** All master cryptographic keys securing financial assets, token minting, and validator consensus must reside exclusively inside certified HSM hardware.
- **Dynamic Credential Lifecycle:** Static database passwords are prohibited in production. Every microservice pod receives a unique, short-lived database username and password that automatically expires and rotates every hour.
- **Immutable Secret Audit Trail:** Vault audit logging is enabled and streamed to a tamper-proof syslog cluster; every secret read, lease renewal, or cryptographic signing operation is immutably logged for forensic auditability.

## Acceptance Criteria
- [ ] Complete Secrets Management and HSM architecture document (`docs/architecture/secrets_management_and_hsm.md`) published.
- [ ] Vault HCL policies authored with least-privilege permissions for all 10 bounded contexts.
- [ ] Dynamic PostgreSQL credential generation configured and verified with automated 1-hour lease expiration.
- [ ] Kubernetes External Secrets Operator (ESO) integration tested with automated secret synchronization.
- [ ] FIPS 140-2 Level 3 HSM / Vault Transit signing interface specified and validated for blockchain relayers.
- [ ] Emergency Secret Revocation and Key Rotation runbooks finalized.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 105 (Auth Architecture), Prompt 108 (Environment Strategy).
- **Parallel Work:** Prompt 110 (Inter-Entity Communication), Prompt 707 (Data Encryption).
- **Blocks:** Prompt 208 (Settlement Service), Prompt 311 (Validator Key Management), Prompt 805 (Terraform Infra).
