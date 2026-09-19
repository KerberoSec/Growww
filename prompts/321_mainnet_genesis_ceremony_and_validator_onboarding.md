# 321 - Mainnet Genesis Ceremony, FIPS 140-2 Level 3 HSM Key Generation & Multi-Institutional Validator Onboarding

## Purpose
Transitioning a regulated financial settlement ledger to production requires absolute cryptographic integrity, institutional non-repudiation, and the verifiable elimination of centralized administrative backdoors. In accordance with SEBI market infrastructure institution guidelines, RBI settlement oversight principles, and global cybersecurity mandates, no single entity (including Growww's parent organization) may possess unilateral control over block production, validator consensus, contract upgrades, or the genesis block allocation.

This prompt specifies the protocol, automated runbooks, cryptographic toolchains, and multi-party signing procedures for the **Mainnet Genesis Ceremony, FIPS 140-2 Level 3 HSM Key Generation, and Multi-Institutional Validator Onboarding** (`ops/genesis-ceremony/` and `infra/hsm/ceremony/`). It establishes an air-gapped ceremony protocol for generating validator keypairs inside hardware security modules, computes a deterministic and tamper-evident `genesis.json` configuration, orchestrates the onboarding of 4 founding institutional validators, and executes a verifiable, atomic on-chain governance handover to a 3-of-5 institutional MultiSig timelock while irrevocably burning deployment keys.

## What You Are Building
An enterprise-grade production initialization framework and ceremony toolkit comprising:
- **Air-Gapped HSM Key Generation Runbook & Tooling (`infra/hsm/ceremony/`):** Hardened scripts and procedural checklists for generating secp256k1 validator keys inside FIPS 140-2 Level 3 compliant Hardware Security Modules (AWS CloudHSM / Thales Luna PCIe HSM) with non-exportable private key attributes and Web3Signer integration.
- **Multi-Party Shamir's Secret Sharing (SSS) Backup Protocol:** Cryptographically splits emergency disaster recovery seed material into a 3-of-5 threshold scheme distributed across institutional legal trustees.
- **Deterministic Genesis Configuration Generator (`ops/genesis-ceremony/genesis_builder.go`):** Produces the immutable Besu QBFT `genesis.json` (Chain ID `13370`), calculates the genesis block header hash, embeds pre-compiled permissioning contracts, and binds the initial 4 institutional validator public addresses.
- **Multi-Institutional Validator Onboarding & Peering Protocol:** Automated coordination workflows for connecting the 4 founding institutional nodes: (1) Growww India Operating Entity, (2) SEBI-Registered Depository/Custodian (NSDL/CDSL Participant), (3) Clearing & Settlement Corporation, and (4) GIFT City International Gateway Entity, alongside read-only regulatory Observer Nodes (SEBI/IFSCA).
- **MultiSig Governance Handover & Deployer Key Revocation Engine:** Executes atomic on-chain transactions transferring contract ownership, proxy admin rights, and permissioning controls to a 3-of-5 Gnosis Safe MultiSig governed by a 48-hour TimelockController, followed by the cryptographic proof of deployer key burn.
- **Ceremony Audit Trail & Cryptographic Attestation Suite:** Generates an open-source, reproducible witness transcript signed with PGP/GPG keys by all institutional participants.

## Scope Boundaries
- **In Scope:**
  - Standard Operating Procedures (SOP) and runbooks for physical and cryptographic air-gapped key generation.
  - FIPS 140-2 Level 3 HSM partition initialization and Web3Signer secure RPC endpoint binding.
  - Deterministic `genesis.json` compilation, validation, and zero-state allocation verification.
  - Pre-deployment and initialization of core system contracts (`NodeRules.sol`, `AccountRules.sol`, `AdminRegistry.sol`, `SafeProxyFactory`).
  - Step-by-step onboarding protocol for 4 external institutional validator partners.
  - On-chain atomic ownership transfer to the 3-of-5 Institutional Governance MultiSig and 48-hour TimelockController.
  - Cryptographic witness attestation manifest (`ceremony-attestation.json`) generation and verification.
- **Out of Scope / Handled Elsewhere:**
  - Testnet cluster and faucet deployment (handled in Prompt 320).
  - Ongoing node monitoring and Prometheus alerting (handled in Prompt 310).
  - Ongoing smart contract logic and token mechanics (handled in Prompts 303-307).
  - Disaster recovery snapshot restoration procedures (handled in Prompt 314).

## Technology to Use
- **Blockchain Node Engine:** **Hyperledger Besu v24.x+** Enterprise Ethereum Client.
- **Consensus Algorithm:** QBFT (Quorum Byzantine Fault Tolerance) with $N=4$ validators, $F=1$ fault tolerance, 2-second block intervals.
- **Key Generation & Storage:** **FIPS 140-2 Level 3 HSM** (AWS CloudHSM cluster or Thales Luna Network HSM) interfacing with **Web3Signer (v24.x+)** via PKCS#11 / REST.
- **Threshold Cryptography:** **Shamir's Secret Sharing (SSS / SLIP-0039)** 3-of-5 threshold configuration for emergency recovery shards.
- **MultiSig & Governance:** **Safe (v1.4.1+)** multi-signature smart contracts and **OpenZeppelin TimelockController** (48-hour minimum execution delay).
- **Ceremony Tooling:** **Go (v1.22+)** for deterministic genesis hashing and witness verification; **OpenSSL / GPG (v2.4+)** for participant cryptographic signatures.
- **Operating Environment:** Hardened, air-gapped Ubuntu 24.04 LTS live boot environments with disabled network interfaces for offline key generation.

## Backend / Infra Touchpoints
- **Validator Key Management HSM (Prompt 311):** Integrates directly with the FIPS 140-2 Level 3 HSM partitions provisioned during the ceremony.
- **Network Topology & Validator Setup (Prompt 302):** Deploys the physical node StatefulSets, Kubernetes NetworkPolicies, and peering configurations using the addresses generated in this ceremony.
- **MultiSig Governance Smart Contract (Prompt 307):** Receives full administrative ownership, upgradeability control, and emergency pause permissions at the conclusion of the handover.
- **Secrets Management Architecture (Prompt 109):** Stores encrypted configuration manifests, mTLS certificates, and non-sensitive metadata in HashiCorp Vault.

## Blockchain Interaction
- **Chain ID & Network Parameters:** Chain ID `13370`, Network ID `13370`.
- **Genesis Block Allocation (`alloc`):** Zero pre-mine for private individuals or unbacked addresses. Genesis balance is strictly allocated to the Settlement Guarantee Fund contract (Prompt 315) and pre-deployed system permissioning contracts.
- **On-Chain Validator Set:** Initial 4 validator addresses encoded in the `extraData` header field of Block 0. Future validator additions or removals require on-chain QBFT voting or MultiSig governance proposals.
- **Zero PII & 1:1 Depository Invariance:** The genesis state instantiates clean ledger registries with zero initial token supply. All subsequent security tokens must originate through verified 1:1 demat share lock transactions via the Depository Gateway (Prompt 213).

## Step-by-Step Build Instructions
1. Scaffold repository under `ops/genesis-ceremony/` and `infra/hsm/ceremony/`:
   - `ops/genesis-ceremony/scripts/` (Genesis generation and verification scripts).
   - `ops/genesis-ceremony/docs/` (Ceremony runbooks, witness protocols, legal sign-off templates).
   - `infra/hsm/ceremony/pkcs11/` (HSM initialization and PKCS#11 key generation templates).
   - `infra/hsm/ceremony/web3signer/` (Web3Signer configuration templates).
2. Prepare the Air-Gapped Key Generation Environment:
   - Provision 4 identical, factory-sealed offline workstations running air-gapped Ubuntu 24.04 Live USBs inside physically secure rooms.
   - Verify cryptographic hashes (SHA-256) of all ceremony binaries (`besu`, `web3signer`, `genesis_builder`).
3. Execute FIPS 140-2 Level 3 HSM Validator Key Generation:
   - For each of the 4 institutional participants, initialize dedicated HSM cryptographic partitions (Crypto Officer and Crypto User roles).
   - Generate secp256k1 keypairs directly inside the HSM partition with attributes `CKA_EXTRACTABLE=FALSE`, `CKA_PRIVATE=TRUE`, `CKA_SIGN=TRUE`.
   - Export strictly the uncompressed public key (`0x04...`) and derive the 20-byte Ethereum validator address (`0x...`).
   - Store public validator addresses in the ceremony registry: `val-growww-india`, `val-custodian-nsdl`, `val-clearing-corp`, `val-gift-city`.
4. Generate Emergency Disaster Recovery Shards using Shamir's Secret Sharing (SSS):
   - In an offline enclave, generate an emergency master recovery seed.
   - Split the seed into 5 SSS shares with a 3-of-5 recovery threshold.
   - Encrypt each share with the respective institutional trustee's public PPG key, print onto tamper-evident physical steel plates, and deposit into institutional bank vaults.
5. Compile Pre-Deployed Governance and Permissioning Contracts:
   - Compile `AdminRegistry.sol`, `NodeRules.sol`, `AccountRules.sol`, `TimelockController.sol`, and `Safe.sol` (3-of-5 threshold) using Solidity 0.8.24 with deterministic bytecode settings.
   - Calculate deterministic contract deployment addresses via `CREATE2`.
6. Run Deterministic Genesis Generator (`genesis_builder.go`):
   - Input: Chain ID `13370`, QBFT block time `2s`, 4 validator public addresses, pre-deployed contract bytecodes.
   - Output: Canonical `genesis.json`.
   - Calculate and publish SHA-256 hash of `genesis.json`: `GENESIS_HASH`.
7. Cryptographic Witness Sign-Off Protocol:
   - Authorized technical officers from each of the 4 institutions inspect the `genesis.json` and compute the SHA-256 hash independently.
   - Each witness signs the attestation document containing `GENESIS_HASH`, public addresses, and timestamp with their institutional GPG/PGP hardware key.
   - Commit signed witness statements to `ops/genesis-ceremony/attestations/`.
8. Configure and Launch Institutional Validator Nodes:
   - Configure Web3Signer on each validator node to connect to its respective institutional HSM partition via mTLS.
   - Initialize node storage using the canonical `genesis.json`: `besu --data-path=/var/lib/besu/data template-genesis /etc/genesis.json`.
   - Establish secure P2P peering between the 4 validator nodes and 2 bootnodes using static ENODE URIs.
9. Verify Mainnet Block 0 (Genesis) and QBFT Consensus Ignition:
   - Start validator nodes simultaneously at the designated ceremony timestamp ($T_{\text{launch}}$).
   - Observe QBFT round 0 proposal, commit phase, and the successful production of Block 1 within 2.0 seconds.
   - Verify that all 4 validator nodes produce valid cryptographic signatures in the QBFT block headers.
10. Deploy and Initialize Core Application Smart Contracts:
    - Deploy `TokenIssuance.sol`, `TokenRedemption.sol`, `TransferCompliance.sol`, and `SettlementDvP.sol` using an ephemeral deployment address.
11. Execute MultiSig Governance Handover:
    - Transfer ownership of all core contracts from the ephemeral deployment address to the 48-hour `TimelockController`.
    - Set the admin of `TimelockController` to the 3-of-5 Institutional `Safe` MultiSig.
    - Transfer proxy administration rights for all upgradeable ERC-1967 contracts to the MultiSig.
12. Irrevocable Deployer Key Revocation / Burn:
    - Renounce all administrative roles previously held by the deployer key.
    - Cryptographically prove that the deployer private key has been zeroized from memory and the temporary signing key discarded.
13. Deploy Regulatory Observer Nodes:
    - Provision read-only observer nodes for SEBI and IFSCA regulatory sandboxes.
    - Verify observer nodes synchronize mainnet blocks via P2P without participating in QBFT consensus voting.
14. Finalize Production Readiness Sign-Off:
    - Run automated ledger sanity test asserting zero unauthorized token balances.
    - Publish public ceremony transcript and validator addresses on the public Growww transparency portal.

## Interfaces / Contracts

### Production Genesis Block Configuration (`genesis.json`)
```json
{
  "config": {
    "chainId": 13370,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "shanghaiTime": 0,
    "cancunTime": 0,
    "zeroBaseFee": true,
    "fixedBaseFee": 0,
    "qbft": {
      "blockperiodseconds": 2,
      "epochlength": 30000,
      "requesttimeoutseconds": 4,
      "validatorcontractaddress": "0x0000000000000000000000000000000000007777",
      "validators": [
        "0x10A8C28bF3cD216892556515b13C8e59275F9001",
        "0x20B9D39cE4dE327903667626c24D9f60386E0002",
        "0x30CaE40dF5eF438014778737d35Ea071497F1003",
        "0x40DbF51eA6fA549125889848e46Fb18250802004"
      ]
    },
    "permissions": {
      "node": "0x0000000000000000000000000000000000008888",
      "account": "0x0000000000000000000000000000000000009999"
    }
  },
  "nonce": "0x0",
  "timestamp": "0x69600000",
  "extraData": "0xf87aa000000000000000000000000000000000000000000000000000000000000000009410a8c28bf3cd216892556515b13c8e59275f90019420b9d39ce4de327903667626c24d9f60386e00029430cae40df5ef438014778737d35ea071497f10039440dbf51ea6fa549125889848e46fb18250802004c880c0",
  "gasLimit": "0x1C9C380",
  "difficulty": "0x1",
  "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
  "alloc": {
    "0x0000000000000000000000000000000000008888": {
      "comment": "Pre-compiled NodeRules Permissioning Contract",
      "balance": "0x0"
    },
    "0x0000000000000000000000000000000000009999": {
      "comment": "Pre-compiled AccountRules Permissioning Contract",
      "balance": "0x0"
    },
    "0x5555555555555555555555555555555555555555": {
      "comment": "Settlement Guarantee Fund Contract Reserve",
      "balance": "0x0"
    }
  }
}
```

### Ceremony Witness Attestation Schema (`ceremony-attestation.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "MainnetGenesisWitnessAttestation",
  "type": "object",
  "required": [
    "ceremonyId",
    "chainId",
    "genesisSha256",
    "blockZeroTimestamp",
    "institutionalSigners",
    "governanceMultiSigAddress",
    "timelockDelaySeconds"
  ],
  "properties": {
    "ceremonyId": {
      "type": "string",
      "example": "GROWWW-MAINNET-CEREMONY-2026-V1"
    },
    "chainId": {
      "type": "integer",
      "example": 13370
    },
    "genesisSha256": {
      "type": "string",
      "pattern": "^[a-fA-F0-9]{64}$",
      "example": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    },
    "blockZeroTimestamp": {
      "type": "integer",
      "example": 1767859200
    },
    "institutionalSigners": {
      "type": "array",
      "minItems": 4,
      "items": {
        "type": "object",
        "required": [
          "organizationName",
          "role",
          "validatorAddress",
          "hsmModelFipsLevel",
          "officerPgpKeyFingerprint",
          "pgpSignature"
        ],
        "properties": {
          "organizationName": {
            "type": "string"
          },
          "role": {
            "type": "string",
            "enum": ["OPERATOR", "DEPOSITORY_CUSTODIAN", "CLEARING_CORP", "GIFT_CITY_GATEWAY", "REGULATORY_OBSERVER"]
          },
          "validatorAddress": {
            "type": "string",
            "pattern": "^0x[a-fA-F0-9]{40}$"
          },
          "hsmModelFipsLevel": {
            "type": "string",
            "example": "AWS CloudHSM FIPS 140-2 Level 3"
          },
          "officerPgpKeyFingerprint": {
            "type": "string"
          },
          "pgpSignature": {
            "type": "string"
          }
        }
      }
    },
    "governanceMultiSigAddress": {
      "type": "string",
      "pattern": "^0x[a-fA-F0-9]{40}$"
    },
    "timelockDelaySeconds": {
      "type": "integer",
      "minimum": 172800,
      "example": 172800
    }
  }
}
```

### Web3Signer PKCS#11 HSM Configuration Template (`hsm-signer-config.yaml`)
```yaml
type: "file-keystore"
# Production Web3Signer CloudHSM / Hardware PKCS#11 Key Provider
keyType: "SECP256K1"
hardwareModule:
  pkcs11LibraryPath: "/opt/cloudhsm/lib/libcloudhsm_pkcs11.so"
  slotIndex: 0
  keyLabel: "GROWWW-PROD-VAL-01"
  pinSecretEnvVar: "HSM_CRYPTO_USER_PIN"
signingPolicy:
  allowedChainIds: [13370]
  enforceBlockHeaderValidation: true
  maxConsecutiveSigningFailures: 3
```

## Security & Compliance Notes
- **FIPS 140-2 Level 3 Hardware Enforcement:** All validator private signing keys are generated inside dedicated hardware partitions with strict non-extractability flags (`CKA_EXTRACTABLE=FALSE`). Private key bytes never exist in plaintext memory on host compute instances.
- **Multi-Institutional Quorum (No Dictator Key):** Consensus requires at least 3-of-4 validator confirmations to produce or finalize any block. Governance upgrades and parameter changes strictly require a 3-of-5 threshold of independent institutional legal entities.
- **Air-Gapped Genesis Verification:** The genesis configuration hash (`GENESIS_HASH`) must be computed on disconnected machines and independently cross-signed by cryptographic officers before any node connects to the live consortium network.
- **Atomic Deployer Key Burning:** After contract deployment and verification, the temporary deployer address executes `renounceOwnership()`, transferring 100% control to the `TimelockController` and rendering the deployer key completely powerless on-chain.
- **Strict Zero-PII Policy:** Block headers, genesis transactions, and ceremony attestations contain only cryptographic public keys, hexadecimal hashes, and institutional entity names. No personal identifiable information is ever recorded to the distributed ledger.

## Acceptance Criteria
- [ ] Air-gapped key generation protocol successfully executed with 4 independent institutional HSM partitions (FIPS 140-2 Level 3).
- [ ] Canonical `genesis.json` generated with deterministic SHA-256 hash verified by all 4 institutional signers.
- [ ] 3-of-5 Shamir's Secret Sharing (SSS) key shards encrypted and securely deposited with designated legal trustees.
- [ ] QBFT consensus ignited at $T_{\text{launch}}$ with 4 validators achieving steady 2.0s block production and deterministic finality.
- [ ] MultiSig governance handover completed: contract ownership transferred to `TimelockController` with 48-hour delay under 3-of-5 Safe MultiSig control.
- [ ] Deployer private key revoked and on-chain ownership renunciation cryptographically verified.
- [ ] Complete ceremony transcript (`ceremony-attestation.json`) signed with GPG keys and committed to the audit archive.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `302` (Network Topology & Validator Setup), Prompt `307` (MultiSig Governance Smart Contract), Prompt `311` (Validator Key Management HSM), Prompt `320` (Permissioned Testnet Cluster & Faucet).
- **Parallel Tasks:** Prompt `314` (Chain Disaster Recovery & Backup Plan), Prompt `315` (Settlement Guarantee Fund Contract).
- **Subsequent Prompts Enabled:** Prompt `208` (Trade Settlement Service Production Cutover), Prompt `216` (Regulatory Reporting Service Integration).
