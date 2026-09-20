# Growww Platform Secrets Management, Encryption Keys & HSM Architecture

## 1. Zero-Plaintext Security Mandate
In strict compliance with SEBI Cyber Security and Cyber Resilience Framework (CSCRF) and RBI digital payment standards, the Growww platform enforces a zero-plaintext secrets mandate:
- No hardcoded API keys, database credentials, or private keys in Git repositories.
- All secrets are managed centrally by HashiCorp Vault Enterprise with automatic key rotation and lease revocation.
- Blockchain consensus keys and settlement signing keys reside exclusively in FIPS 140-2 Level 3 Hardware Security Modules (AWS CloudHSM).

---

## 2. HashiCorp Vault Architecture
- **Topology:** 3-node High Availability cluster with Integrated Raft Storage.
- **Auto-Unseal:** Hardware-backed unseal via AWS KMS / CloudHSM.
- **Authentication:**
  - Kubernetes Auth (`auth/kubernetes`): Microservice pods authenticate using short-lived pod service account tokens.
  - JWT/OIDC Auth (`auth/jwt`): CI/CD runners authenticate via ephemeral OpenID Connect claims.
  - AppRole (`auth/approle`): Machine-to-machine background workers.

---

## 3. Dynamic Database Credentials
Microservices do not hold static database passwords. The Vault Database Secrets Engine generates ephemeral credentials with a 1-hour TTL:
```sql
CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO "{{name}}";
```
When a pod terminates or its lease expires without renewal, Vault automatically revokes the PostgreSQL user role.

---

## 4. FIPS 140-2 Level 3 HSM Custody & PKCS#11
- **Besu Validator Keys:** Generated inside AWS CloudHSM partitions; keys never leave hardware boundary (`CKA_EXTRACTABLE = FALSE`).
- **Settlement Relayer:** Signs DvP execution transactions via PKCS#11 session pool or Vault Transit Engine.
- **Anti-Equivocation Guard:** Prevents signing two different block headers at the same block height to protect against validator slashing.
