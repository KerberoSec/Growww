# 311 - Validator & Relayer Key Management via Hardware Security Modules (HSM/KMS)

## Purpose
Under SEBI, RBI, and global institutional cybersecurity standards, cryptographic private keys controlling blockchain validator consensus, transaction submission relayers, and multi-signature governance wallets represent critical infrastructure assets. If a validator or relayer private key is stored as a raw plaintext file on disk or in container memory, a single server compromise could allow an attacker to sign fraudulent blocks, censor transactions, or forge settlement instructions.

This prompt specifies the architecture, configuration, and deployment of **Enterprise Key Management & Signing Proxy Infrastructure** utilizing **ConsenSys Web3Signer** backed by **FIPS 140-3 Level 3 Hardware Security Modules (HSM)** and **HashiCorp Vault Transit Engine**. All cryptographic signing operations (secp256k1 ECDSA) occur strictly inside the secure cryptographic boundaries of the HSM, ensuring private keys are non-exportable and never touch application memory in plaintext.

## What You Are Building
A secure key management and signing subsystem (`infra/blockchain/key-management/`) containing:
- Web3Signer Proxy Cluster: Highly available signing proxy interfacing between Hyperledger Besu nodes and the underlying HSM.
- HashiCorp Vault Transit Engine / AWS CloudHSM Configuration: Provisions secp256k1 key rings for (1) QBFT Validator Consensus Signers, (2) DvP Settlement Relayers, (3) Custodian KYC Whitelist Agents, and (4) Proof-of-Reserve Attestation Signers.
- mTLS & IAM Access Policies: Granular role-based access ensuring only authorized Kubernetes service accounts can request transaction signatures.
- Local Mock HSM / SoftHSM2 Suite (`test/hsm/`): For automated CI/CD pipeline integration and local developer environments.

## Scope Boundaries
- **In Scope:**
 - Web3Signer deployment and configuration for Hyperledger Besu validator nodes.
 - AWS CloudHSM / HashiCorp Vault Transit Engine secp256k1 key provisioning.
 - Relayer signing proxy daemon for backend microservices (DvP Settlement, KYC, PoR).
 - Fine-grained signing rules (transaction destination contract allowlisting, gas/value caps).
 - Key rotation procedures and disaster recovery key backup protocols.
- **Out of Scope / Handled Elsewhere:**
 - Microservice OAuth2/OIDC user token authentication (handled in Prompt 105).
 - General application database credentials management (handled in Prompt 109).
 - Multisig smart contract governance execution (handled in Prompt 307).

## Technology to Use
- **Signing Proxy:** **ConsenSys Web3Signer (v24.x+)**.
  *Justification:* Web3Signer is an open-source, enterprise signing service designed natively for Hyperledger Besu and Ethereum. It supports external cryptographic key providers (Vault, AWS KMS/CloudHSM, Azure Key Vault) via REST API and PKCS#11, decodes incoming payloads, and verifies transaction semantics before signing.
- **Key Store / HSM:** **HashiCorp Vault Transit Secrets Engine** (primary software HSM) and **AWS CloudHSM / Azure Dedicated HSM (FIPS 140-3 Level 3)** for production validators.
- **Local Dev / CI:** SoftHSM2 and Vault in dev mode.
- **Protocols:** mTLS (TLS 1.3 with client certificates), REST over HTTPS, Unix domain sockets.

## Backend / Infra Touchpoints
- **Hyperledger Besu Validator Nodes:** Configured with `--key-source=web3signer` pointing to the Web3Signer instance over private mTLS.
- **Settlement Service (Prompt 208):** Submits raw unsigned DvP batches to the signing relayer.
- **KYC Service (Prompt 202):** Submits identity registration transactions to the whitelist signer.
- **Vault Cluster:** Enterprise Vault cluster storing audit logs and access token policies.

## Blockchain Interaction
- **Validator Consensus Signing:** Besu delegates QBFT block proposal and commit signatures to Web3Signer; Web3Signer sends the payload hash to the HSM, receives the `(r, s, v)` signature, and returns it to Besu.
- **Relayer Transaction Signing:** Microservices request transaction signatures without holding keys; the signing proxy validates that the target `to` address matches a known contract (`SettlementDvP.sol`, `DigitalSecurityToken.sol`) and signs the `EIP-155` transaction.
- **Zero Raw Key Storage:** Private keys cannot be extracted from the HSM even by infrastructure administrators.

## Step-by-Step Build Instructions
1. Scaffold repository directory under `infra/blockchain/key-management/` with subdirectories: `web3signer/`, `vault/policies/`, `helm/`, `scripts/`.
2. Configure HashiCorp Vault Transit secrets engine: enable `transit/` mount path and provision secp256k1 keyrings:
 - `growww-qbft-validator-1`
 - `growww-relayer-settlement`
 - `growww-relayer-kyc`
 - `growww-relayer-por`
3. Write Vault ACL policies restricting signing capabilities per Kubernetes service account (e.g. `policy-settlement-signer.hcl` allows only `update` on `transit/sign/growww-relayer-settlement`).
4. Configure Web3Signer configuration files (`web3signer-config.yaml` and key definition `.yaml` files):
 - Define `key-type: SECP256K1`
 - Configure Vault endpoint, token authentication, and path `transit/sign/key-name`.
5. Implement signing validation rules in Web3Signer:
 - Restrict chain ID strictly to `13370` (prevents cross-chain replay).
 - Enforce destination contract whitelist (only approved Growww ecosystem contracts).
 - Reject any transaction with non-zero ETH `value` (consortium chain has zero native token transfers).
6. Configure Besu validator nodes to use Web3Signer: add flags `--key-source=web3signer`, `--web3signer-url=https://web3signer.blockchain.internal:9000`, `--web3signer-tls-enabled=true`.
7. Implement a lightweight transaction signing client library in Go/Rust (`pkg/signer/hsm_signer.go`):
 - Constructs standard EIP-155 transaction payloads.
 - Fetches nonce from Besu node.
 - Calls Web3Signer / Vault Transit API to compute ECDSA signature.
 - Broadcasts signed raw transaction via `eth_sendRawTransaction`.
8. Create SoftHSM2 mock container setup for local developer environments and GitHub Actions CI pipelines.
9. Write automated integration tests: simulate 1,000 concurrent signing requests and verify latency is $<15\text{ ms}$ per signature.
10. Implement Prometheus metrics exporter on Web3Signer: track `signing_requests_total`, `signing_latency_ms`, and `hsm_error_rate`.
11. Document key generation and M-of-N key recovery ceremony procedures (`docs/security/key_ceremony_runbook.md`).
12. Build hardened Web3Signer Docker image with read-only root filesystem and minimal Alpine base.
13. Deploy Web3Signer StatefulSet in Kubernetes with dedicated sidecar containers communicating via localhost mTLS.

## Interfaces / Contracts

### Web3Signer Key Definition Configuration (`validator-key-1.yaml`)
```yaml
type: "hashicorp"
keyType: "SECP256K1"
serverHost: "vault.internal.growww.in"
serverPort: 8200
timeout: 3000
tlsEnabled: true
tlsTrustStorePath: "/etc/web3signer/certs/vault-ca.p12"
tlsTrustStorePassword: "vault-keystore-password"
keyPath: "transit/sign/growww-qbft-validator-1"
token: "${VAULT_SIGNING_TOKEN}"
```

### Relayer Signing Request Interface (REST API)
```json
{
  "request_id": "REQ-SIGN-20260918-001",
  "key_alias": "growww-relayer-settlement",
  "chain_id": 13370,
  "to": "0x3333333333333333333333333333333333333333",
  "nonce": 4210,
  "gas_limit": "0x100000",
  "gas_price": "0x0",
  "data": "0x1a2b3c4d000000000000000000000000...",
  "value": "0x0"
}
```

### Vault Transit Policy Definition (`policy-settlement-signer.hcl`)
```hcl
path "transit/sign/growww-relayer-settlement" {
  capabilities = ["update"]
}

path "transit/verify/growww-relayer-settlement" {
  capabilities = ["read", "update"]
}

path "transit/keys/growww-relayer-settlement" {
  capabilities = ["read"]
}
```

## Security & Compliance Notes
- **FIPS 140-3 Level 3 Compliance:** Production consensus and relayer keys reside within dedicated HSM modules meeting FIPS 140-3 Level 3 tamper-resistance standards.
- **Zero Key Exportability:** Private keys are generated directly inside the HSM with the `exportable=false` attribute permanently set; extraction of the raw private scalar is cryptographically impossible.
- **Audit Logging of Every Signature:** Vault logs every signing request with client IP, Kubernetes service account name, key alias, and timestamp to an immutable append-only audit sink (Prompt 218).

## Acceptance Criteria
- [ ] Besu validator nodes successfully propose and commit QBFT blocks using keys managed in Web3Signer/Vault.
- [ ] Relayer microservices sign transactions via signing API without holding raw private keys.
- [ ] Web3Signer strictly rejects signing requests for unapproved destination contract addresses.
- [ ] Signing latency p95 is $<20\text{ ms}$ under sustained load of 500 signatures/second.
- [ ] SoftHSM2 CI pipeline passes 100% of integration test suites without external cloud dependencies.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `302` (Network Topology & Validator Setup), Prompt `109` (Secrets Management Architecture).
- **Parallel Tasks:** Prompt `306` (Settlement DvP Contract), Prompt `310` (Chain Node Monitoring).
- **Subsequent Prompts Enabled:** Prompt `208` (Trade Settlement Service), Prompt `202` (KYC Backend Service).
