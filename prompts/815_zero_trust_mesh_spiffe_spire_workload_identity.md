# 815 - Zero-Trust SPIFFE/SPIRE Microservice Workload Identity Architecture

## Purpose
In a regulated, high-frequency financial and blockchain platform, perimeter-based security and static credentials introduce severe systemic risks. Static database passwords, long-lived API tokens, and persistent Kubernetes service account secrets can be leaked, exfiltrated, or inadvertently logged, creating attack vectors that violate modern zero-trust mandates.

This specification establishes the zero-trust workload identity architecture for Growww using SPIFFE (Secure Production Identity Framework for Everyone) and SPIRE (the SPIFFE Runtime Environment). By issuing cryptographically verifiable, ephemeral X.509 SVIDs (SPIFFE Verifiable Identity Documents) and JWT-SVIDs with short lifetimes (1-hour maximum, rotated every 30 minutes), the platform completely eliminates static database passwords, hardcoded credentials, and long-lived tokens across all microservices, backend workers, and consortium blockchain nodes.

Every microservice must continuously attest its identity to a local SPIRE Agent backed by hardware TPM 2.0 and Kubernetes kernel-level primitives before receiving cryptographic credentials. This architecture directly satisfies the SEBI Cybersecurity and Cyber Resilience Framework (CSCRF), RBI Master Directions on IT Governance, and IFSCA IT security regulations for cryptographic non-repudiation, automated credential rotation, mutual TLS (mTLS) wire encryption, and least-privilege identity federation.

## What You Are Building
A production-grade enterprise SPIRE deployment architecture rooted under `infra/spire/`:
- `infra/spire/server/`: High-availability SPIRE Server cluster manifests (StatefulSet backed by AWS RDS PostgreSQL datastore), integrated with HashiCorp Vault PKI as upstream intermediate CA (Prompt 109) and AWS KMS for KeyManager encryption.
- `infra/spire/agent/`: SPIRE Agent DaemonSet configurations running on every Kubernetes worker node, executing dual hardware (TPM 2.0) and kernel-level node attestation.
- `infra/spire/csi-driver/`: SPIFFE CSI Driver for Kubernetes (`spiffe-csi-driver`) enabling unprivileged pods to mount the SPIRE Workload API Unix Domain Socket directly into container filesystems without requiring insecure `hostPath` volumes.
- `infra/spire/k8s-workload-registrar/`: Controller reconciling Kubernetes pod lifecycles, CRDs, and annotations into deterministic SPIRE Server registration entries.
- `infra/spire/envoy/`: Envoy proxy sidecar templates leveraging SPIFFE Secret Discovery Service (SDS) to enforce dynamic mTLS encryption, certificate reloading, and L7 authorization filtering between microservices.
- `infra/spire/db-auth/`: Passwordless database authentication architecture configuring PostgreSQL and PgBouncer connection pools to authenticate incoming connections via X.509 client certificate Subject Alternative Name (SAN) mapping directly to database roles.
- `infra/spire/federation/`: SPIFFE OIDC Discovery Provider (`/.well-known/openid-configuration`) and trust bundle federation configs for external consortium banking and depository partners.

## Scope Boundaries
- **In Scope:**
  - SPIRE Server HA deployment architecture, storage backend, and upstream Vault PKI CA integration.
  - SPIRE Agent DaemonSet architecture with TPM 2.0 hardware and Kubernetes node attestation.
  - SPIFFE Workload API injection via Kubernetes SPIFFE CSI Driver.
  - Envoy proxy sidecar integration via dynamic SPIFFE Secret Discovery Service (SDS) for mTLS.
  - Passwordless PostgreSQL and PgBouncer client certificate authentication (`cert` method with SPIFFE SAN parsing).
  - SPIFFE JWT-SVID authentication to HashiCorp Vault eliminating static AppRole SecretIDs.
  - Workload attestation for Hyperledger Besu validator daemons and settlement transaction relayers.
  - Automated SVID issuance, key rotation (1-hour TTL, 30-minute renewal window), and emergency revocation.
- **Out of Scope / Handled Elsewhere:**
  - Root Certificate Authority generation inside FIPS 140-2 Level 3 HSM (handled in Prompt 109).
  - End-user identity, OAuth2/OIDC, Biometric WebAuthn, and customer JWT tokens (handled in Prompt 105).
  - Calico / Cilium L3/L4 network packet filtering and egress gateways (handled in Prompt 802).
  - Terraform provisioning of underlying cloud compute, VPCs, and RDS storage (handled in Prompt 805).
  - Edge ingress API gateway SSL termination from public internet clients (handled in Prompt 219).

## Technology to Use
- **SPIFFE / SPIRE 1.9+**: The Cloud Native Computing Foundation (CNCF) graduated standard for cryptographically attesting workload identities across heterogeneous infrastructure without cloud-vendor lock-in.
- **Envoy Proxy 1.30+**: High-performance L7 proxy acting as sidecar service mesh. Justification: Native integration with SPIFFE SDS allows seamless zero-downtime certificate rotation without terminating in-flight financial transactions or dropping TCP connections.
- **SPIFFE CSI Driver v0.5+**: Kubernetes Container Storage Interface driver. Justification: Securely mounts the Workload API Unix Domain Socket into application containers, strictly enforcing Kubernetes Pod Security Standards (`restricted`) and eliminating risky hostPath volume bindings.
- **TPM 2.0 (Trusted Platform Module)**: Cryptographic hardware module present in AWS Nitro EC2 instances and physical bare-metal servers. Justification: Enables hardware-rooted node attestation by validating Platform Configuration Registers (PCRs), guaranteeing that only untampered nodes can join the SPIRE trust domain.
- **HashiCorp Vault 1.16+**: Enterprise secret and PKI manager. Justification: Serves as the upstream intermediate CA for SPIRE Server via the `upstream_authority "vault"` plugin, anchoring workload certificates back to corporate HSM-backed roots.
- **PostgreSQL 16+ & PgBouncer 1.22+**: Core relational database tier configured with TLS client certificate authentication (`ssl_cert_auth`) mapped to SPIFFE identity SANs.

## Backend / Infra Touchpoints
- **All 30+ Core Microservices**: Including `user-service`, `order-matching-engine`, `trade-settlement-service`, `kyc-aml-service`, `risk-margin-service`, `wallet-account-service`, `market-data-service`, and `fee-pnl-engine`. Microservices consume SVIDs from the local Workload API socket or delegate mTLS handshakes directly to Envoy sidecars.
- **HashiCorp Vault (Prompt 109)**: Supplies intermediate CA certificates to SPIRE Server; microservices authenticate to Vault by exchanging SPIFFE JWT-SVIDs against Vault's JWT/OIDC authentication engine, eradicating all static bootstrap tokens.
- **PostgreSQL Connection Pools (PgBouncer & RDS PostgreSQL)**: Configured with `clientcert=verify-full` and `pg_ident.conf` map directives, allowing microservices to connect without passwords by validating X.509 client certificate SANs.
- **Apache Kafka / Redpanda Cluster**: Kafka brokers configure mTLS listeners with `ssl.client.auth=required` and principal builder plugins parsing SPIFFE IDs for fine-grained topic ACL enforcement.
- **Kubernetes API Server (EKS / GKE)**: Leveraged by SPIRE node and workload attestors to verify pod namespace, service account, container image, and label selectors.

## Blockchain Interaction
The permissioned Hyperledger Besu consortium network requires stringent workload attestation and cryptographic non-repudiation:
- **Besu Validator Node Attestation**: Besu validator daemons running in the `growww-blockchain` namespace are attested by SPIRE Agents using TPM 2.0 and Linux workload attestation (verifying cgroup, process binary SHA-256 hash, and execution user ID). SPIRE issues X.509 SVIDs with SPIFFE ID `spiffe://growww.internal/ns/growww-blockchain/sa/besu-validator`.
- **Besu JSON-RPC & WebSocket Security**: The JSON-RPC (`8545`) and WebSocket (`8546`) interfaces of Besu nodes are fronted by Envoy proxies configured with SPIFFE SDS. Ingress is restricted exclusively to authorized transaction relayers presenting valid X.509 SVIDs issued to `spiffe://growww.internal/ns/growww-settlement/sa/trade-settlement-service`.
- **Relayer Vault HSM Signing Delegation**: When DvP settlement relayers execute on-chain smart contract transactions (`DvPSettlement.sol`), the relayer presents its SPIFFE JWT-SVID to HashiCorp Vault. Vault validates the JWT-SVID against the configured role and permits the relayer to invoke `/v1/transit/sign/growww-relayer` without exposing private keys.
- **Consortium Trust Federation**: SPIRE OIDC Discovery Provider publishes trust bundles allowing external validator nodes hosted by regulatory or partner institutions (e.g. NSDL/CDSL mock consortium nodes) to federate trust without sharing common private keys.

## Step-by-Step Build Instructions
1. Scaffold the SPIRE infrastructure directory tree at `infra/spire/{server,agent,csi,envoy,policies,db-auth,federation}`.
2. Formulate the canonical SPIFFE ID naming schema and trust domain definition (`spiffe://growww.internal/...`).
3. Configure the HashiCorp Vault PKI secrets engine to issue short-lived intermediate CA certificates for the SPIRE Server cluster.
4. Deploy the SPIRE Server HA cluster as a Kubernetes `StatefulSet` backed by AWS RDS PostgreSQL using the `sql` datastore plugin.
5. Configure SPIRE Server key management plugins (`aws_kms` or `vault`) and node attestors (`k8s_psat` and `tpm`).
6. Deploy the SPIFFE CSI Driver DaemonSet across all cluster worker nodes to provide secure socket mounting at `/spiffe-workload-api/spiffe-workload-api.sock`.
7. Deploy the SPIRE Agent DaemonSet on all Kubernetes nodes, configured with the `tpm` node attestor and `k8s` workload attestor.
8. Deploy the `k8s-workload-registrar` in CRD or reconciliation mode to automatically register pods into SPIRE based on namespace and ServiceAccount names.
9. Construct the Envoy sidecar container specification with SPIFFE SDS configuration for automatic X.509 SVID certificate and trust bundle ingestion.
10. Integrate Envoy mTLS filters across core inter-service communication paths (`order-service` to `order-matching-engine`, `order-matching-engine` to `trade-settlement-service`).
11. Configure RDS PostgreSQL and PgBouncer instances for TLS certificate verification (`clientcert=verify-full`) and populate `pg_ident.conf` mappings for SPIFFE identities.
12. Configure HashiCorp Vault with the JWT/OIDC authentication engine configured to validate SPIFFE JWT-SVIDs from the SPIRE OIDC Discovery Provider.
13. Implement workload attestation profiles for Hyperledger Besu validator nodes using binary hash and Linux cgroup selectors.
14. Configure SPIRE Server Prometheus metric endpoints and OpenTelemetry collector integration for monitoring SVID issuance rates, renewal latencies, and attestation rejections.
15. Perform chaos and failure validation: verify zero-downtime mTLS rotation, agent restart survival, and immediate connection rejection for unauthorized or tampered pods.

## Interfaces / Contracts

### SPIFFE ID Naming Taxonomy
All workload identities issued within the Growww infrastructure follow a deterministic URI hierarchy:
- Trust Domain: `spiffe://growww.internal`
- Core Kubernetes Microservices: `spiffe://growww.internal/ns/{namespace}/sa/{service-account}`
- Blockchain Validator Nodes: `spiffe://growww.internal/ns/growww-blockchain/sa/besu-validator`
- Transaction Relayers: `spiffe://growww.internal/ns/growww-settlement/sa/settlement-relayer`
- Infrastructure Daemons: `spiffe://growww.internal/infra/{component}/{cluster-id}`

### SPIRE Server Configuration Manifest
```hcl
# infra/spire/server/server.conf
server {
  bind_address = "0.0.0.0"
  bind_port = "8081"
  trust_domain = "growww.internal"
  data_dir = "/run/spire/data"
  log_level = "INFO"
  ca_ttl = "720h"
  default_x509_svid_ttl = "1h"
  default_jwt_svid_ttl = "1h"

  ca_key_type = "ec-p256"

  jwt_issuer = "https://oidc.spire.growww.internal"
}

plugins {
  DataStore "sql" {
    plugin_data {
      database_type = "postgres"
      connection_string = "host=postgres-spire.growww-secops.svc port=5432 user=spire_admin dbname=spire sslmode=verify-full sslrootcert=/run/spire/certs/rds-ca.pem"
    }
  }

  NodeAttestor "k8s_psat" {
    plugin_data {
      clusters = {
        "growww-prod-mumbai" = {
          service_account_allow_list = ["growww-secops:spire-agent"]
        }
      }
    }
  }

  NodeAttestor "tpm" {
    plugin_data {
      ca_path = "/run/spire/tpm/ek-ca-bundle.pem"
    }
  }

  KeyManager "vault" {
    plugin_data {
      vault_addr = "https://vault.growww-secops.svc:8200"
      transit_mount = "transit"
      key_name = "spire-server-ca-key"
      insecure_skip_verify = false
      ca_cert_path = "/run/spire/certs/vault-ca.pem"
    }
  }

  UpstreamAuthority "vault" {
    plugin_data {
      vault_addr = "https://vault.growww-secops.svc:8200"
      pki_mount_path = "pki_intermediate_spire"
      role_name = "spire-server"
      ca_cert_path = "/run/spire/certs/vault-ca.pem"
    }
  }
}
```

### SPIRE Agent Configuration Manifest
```hcl
# infra/spire/agent/agent.conf
agent {
  data_dir = "/run/spire"
  log_level = "INFO"
  server_address = "spire-server.growww-secops.svc"
  server_port = "8081"
  socket_path = "/run/spiffe-workload-api/spiffe-workload-api.sock"
  trust_bundle_path = "/run/spire/bundle/root-bundle.crt"
  trust_domain = "growww.internal"
}

plugins {
  NodeAttestor "k8s_psat" {
    plugin_data {
      cluster = "growww-prod-mumbai"
    }
  }

  KeyManager "memory" {
    plugin_data {}
  }

  WorkloadAttestor "k8s" {
    plugin_data {
      skip_kubelet_verification = false
      max_retries = 3
      retry_interval = "500ms"
    }
  }

  WorkloadAttestor "unix" {
    plugin_data {}
  }
}
```

### Envoy SDS Workload API Secret Configuration
```yaml
# infra/spire/envoy/sds-cluster-config.yaml
resources:
  - "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret
    name: "spiffe://growww.internal/ns/growww-trading/sa/order-matching-engine"
    tls_certificate:
      certificate_chain:
        filename: "/run/spiffe-workload-api/svid.crt"
      private_key:
        filename: "/run/spiffe-workload-api/svid.key"
  - "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret
    name: "SPIRE_ROOT_BUNDLE"
    validation_context:
      trusted_ca:
        filename: "/run/spiffe-workload-api/bundle.crt"
      match_typed_subject_alt_names:
        - sanitizer:
            name: envoy.tls.cert_validator.spiffe
          matcher:
            exact: "spiffe://growww.internal/ns/growww-trading/sa/order-service"
```

### PostgreSQL Passwordless Identity Mapping
```text
# infra/spire/db-auth/pg_ident.conf
# MAPNAME       SYSTEM-USERNAME (SPIFFE SAN)                                                  PG-USERNAME
spiffe_map      /^spiffe:\/\/growww\.internal\/ns\/growww-core\/sa\/user-service$             user_service_app
spiffe_map      /^spiffe:\/\/growww\.internal\/ns\/growww-trading\/sa\/order-service$          order_service_app
spiffe_map      /^spiffe:\/\/growww\.internal\/ns\/growww-settlement\/sa\/trade-settlement$    trade_settlement_app

# infra/spire/db-auth/pg_hba.conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD  OPTIONS
hostssl all             all             10.100.0.0/16           cert    clientcert=verify-full map=spiffe_map
```

## Security & Compliance Notes
- **SEBI CSCRF Section 3 Compliance**: Mandates strict identity verification, micro-segmentation, non-repudiation, and ephemeral credential lifecycles across regulated trading and demat custody systems.
- **Short-Lived Ephemeral SVIDs**: Workload X.509 SVIDs have a maximum time-to-live (TTL) of 1 hour (3600 seconds) and must be rotated automatically by the SPIRE Agent at 50% lifespan (30 minutes). If a workload process is compromised, the exfiltrated credential window is strictly bounded.
- **Zero Hardcoded Secrets in Git / Configs**: All microservice communication and database queries operate via mutual TLS using ephemeral client certificates. No passwords or API tokens exist in Kubernetes Secrets, ConfigMaps, or source code.
- **Hardware-Enforced Node Integrity**: The TPM 2.0 node attestor validates hardware endorsements and PCR values before granting node credentials. Worker nodes with modified kernels, unauthorized boot components, or rootkits are automatically rejected from the SPIRE trust domain.
- **Micro-Segmentation and Least Privilege**: Access between microservices is enforced via Envoy L7 RBAC filters evaluating incoming SPIFFE IDs. The matching engine accepts requests exclusively from verified `order-service` SVIDs, rejecting all other cluster traffic even if network-level firewalls are breached.
- **Audit Logging and Cryptographic Non-Repudiation**: Every SVID issuance, attestation failure, and renewal event is forwarded to the immutable audit logging pipeline (Prompt 218 and Prompt 807) with pod UID, node identity, and timestamp.

## Acceptance Criteria
- [ ] SPIRE Server cluster deployed in high-availability mode backed by RDS PostgreSQL and integrated with HashiCorp Vault PKI upstream authority.
- [ ] SPIRE Agent DaemonSet deployed on all cluster nodes with verified TPM 2.0 and Kubernetes PSAT node attestation.
- [ ] SPIFFE CSI Driver successfully mounts the Workload API Unix Domain Socket into microservice pods under Pod Security Standard `restricted` without hostPath volume mounts.
- [ ] Automatic registration controller assigns deterministic SPIFFE IDs (`spiffe://growww.internal/ns/{namespace}/sa/{service}`) to microservice pods upon scheduling.
- [ ] Inter-service communication between `order-service` and `order-matching-engine` executes over Envoy mTLS with dynamic certificate reloading via SPIFFE SDS.
- [ ] PostgreSQL database connections authenticate successfully using X.509 client certificates and `pg_ident.conf` SPIFFE SAN mapping with zero database passwords.
- [ ] Microservices successfully authenticate to HashiCorp Vault using SPIFFE JWT-SVIDs via Vault's JWT/OIDC auth engine.
- [ ] Hyperledger Besu validator nodes and settlement RPC relayers receive attested SPIFFE SVIDs, and unauthorized callers to Besu RPC endpoints are rejected.
- [ ] Workload SVIDs automatically rotate every 30 minutes with zero in-flight connection drops or 5xx error spikes.
- [ ] Pods with spoofed or unauthorized labels are denied SVID issuance by the SPIRE Agent workload attestor.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101: System Architecture Overview
  - Prompt 102: Service Boundary Map and Bounded Contexts
  - Prompt 109: Secrets Management, Encryption Keys & HSM Architecture (Vault PKI upstream CA)
  - Prompt 802: Multi-Tenant Kubernetes Cluster Architecture & Namespace Segmentation
  - Prompt 805: Infrastructure-as-Code Terraform (VPC, EKS, RDS provisioning)
- **Downstream Dependents:**
  - Prompt 803: CI Pipeline Design
  - Prompt 804: CD Progressive Delivery Pipeline (Envoy mTLS verification gates)
  - Prompt 806: Observability Stack Telemetry (SPIRE metrics and mTLS traces)
  - Prompt 811: Dual-Environment GitOps CI/CD Pipeline (Testnet vs Mainnet identity federation)
