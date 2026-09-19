# Growww / NBSE Data Protection, Privacy & Zero-PII Ledger Policy (DPDP & GDPR)

## 1. Statutory Governance & Objectives
This policy establishes mandatory data protection standards for Growww / NBSE under:
- **India DPDP Act 2023:** Digital Personal Data Protection Act compliance governing Indian domestic users.
- **EU GDPR:** General Data Protection Regulation compliance for international GIFT City investors.
- **SEBI & RBI Regulatory Norms:** Data localization, financial record retention, and cyber resilience frameworks.
- **Supreme Architectural Invariant:** **ZERO Personally Identifiable Information (PII) on the blockchain ledger**.

---

## 2. 4-Tier Data Classification Matrix

| Classification Tier | Data Types Included | Storage Location | Encryption Standard | Access Controls |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 1: Highly Sensitive PII** | Aadhaar numbers, biometric vectors, passport NFC data, PAN | Isolated PII Vault (PostgreSQL VPC) | AES-256-GCM + FIPS 140-2 Level 3 HSM | Zero direct query; microservice surrogate tokens only |
| **Tier 2: Sensitive Financial & KYC** | Bank account numbers, IFSC, trade history, net worth | Encrypted Core RDS / PostgreSQL | AES-256-XTS at rest, TLS 1.3 in transit | RBAC with 2FA, automated WORM audit trails |
| **Tier 3: Pseudonymous Identifiers** | Internal `user_id`, salted wallet addresses, session tokens | Redis Cluster, Kafka Topics | TLS 1.3, ephemeral session encryption | Service mesh mTLS token verification |
| **Tier 4: Public & Ledger Data** | Orderbook depth, aggregated ticks, smart contract bytecode | Public nodes, Hyperledger Besu | Opaque cryptographic hashes (`bytes32`) | Public read access, cryptographic QBFT consensus |

---

## 3. The Zero-PII Blockchain Policy

### 3.1 Strict Prohibition
Under no circumstances shall any plaintext or reversibly encrypted PII appear in:
1. Transaction calldata or method arguments
2. Solidity smart contract storage or state variables
3. Event logs emitted by contracts
4. RPC trace payloads or node logs

### 3.2 Cryptographic Pseudonymization Commitment
All identity assertions on Hyperledger Besu are verified through opaque 32-byte commitments:
$$\text{InvestorCommitment} = \text{keccak256}(\text{abi.encodePacked}(\text{investor\_uuid}, \text{salt}_{\text{vault}}))$$
Smart contracts (`ComplianceRegistry.sol`, `SettlementDvP.sol`) validate permissions strictly via `InvestorCommitment`, ensuring mathematical auditability with zero human identity leakage.

---

## 4. Cryptographic Erasure & "Right to be Forgotten"

### 4.1 Harmonization with Statutory Record Retention
- Financial regulations (PMLA 2002) require a 5-year post-closure retention of financial transaction audit trails.
- DPDP Act 2023 & GDPR Article 17 grant users the right to erasure of personal data once processing purposes are fulfilled.

### 4.2 Crypto-Shredding Mechanism
1. Each data principal's PII is encrypted with a unique Data Encryption Key (DEK) wrapped by an HSM Master Key.
2. Upon verified erasure request:
   - The user's individual DEK in HashiCorp Vault is permanently deleted.
   - Any historical encrypted database ciphertext becomes mathematically impossible to decrypt ($2^{256}$ brute-force complexity).
   - The on-chain `InvestorCommitment` remains permanently valid for cryptographic proof of historical DvP settlement balance without ever being re-linkable to physical identity.

---

## 5. DPDP Act 2023 Consent Manager Lifecycle
- Granular consent items collected for specific processing purposes (`KYC_VERIFICATION`, `TRADE_EXECUTION`, `REGULATORY_FILING`).
- Digitally signed consent artifacts with ISO 8601 timestamps and withdrawal mechanisms accessible via the user privacy dashboard.
- Automated quarterly consent renewal notifications and compliance audit logging.
