# 302 - Permissioned Network Topology & QBFT Validator Infrastructure Setup

## Purpose
In compliance with SEBI and RBI regulatory requirements for institutional market infrastructure, the Growww settlement ledger must operate across an independent, multi-organization consortium network rather than a single centralized node cluster. The consortium architecture ensures that no single entity - including Growww's core operating entity - can unilaterally rewrite transaction history, censor legitimate asset transfers, or falsify proof-of-reserve balances.

This prompt specifies the end-to-end design, configuration, provisioning, and automated deployment of the **Hyperledger Besu QBFT Validator Network**. The initial deployment distributes consensus authority across 4 distinct institutional entities: (1) Domestic Regulated Entity (Growww India), (2) SEBI-registered Custodian / Depository Partner (NSDL/CDSL participant), (3) Clearing / Settlement Partner, and (4) GIFT City International Gateway Entity, with dedicated read-only Observer Nodes provisioned for regulatory authorities (SEBI/IFSCA).

## What You Are Building
A production-grade, highly available Infrastructure-as-Code (IaC) and configuration repository (`infra/blockchain/besu-network/`) containing:
- Consortium genesis configuration (`genesis.json`) with tuned QBFT consensus parameters (2-second block period, validator address list, zero base fee, London/Shanghai EVM fork flags).
- On-chain permissioning smart contracts (`AccountRules.sol` and `NodeRules.sol` / Besu Permissioning Management Contracts) to enforce node allowlists and relayer whitelists at the protocol level.
- Kubernetes Helm charts and Terraform modules for deploying geographically distributed validator nodes on AWS/GCP with dedicated secure enclaves.
- Automated bootnode discovery configuration, static peer definitions, and secure RPC gateway ingress with mTLS.

## Scope Boundaries
- **In Scope:**
 - `genesis.json` creation and cryptographic validator address allocation.
 - Multi-org node topology design (4 initial validators + 2 RPC relayers + 2 observer nodes).
 - Besu on-chain node permissioning and account permissioning contract deployment.
 - Network encryption (TLS P2P communication, mTLS for JSON-RPC management).
 - Containerization, Kubernetes manifests, StatefulSets, persistent volume claims (NVMe SSD storage for RocksDB), and health probes.
- **Out of Scope / Handled Elsewhere:**
 - HSM validator key generation and Web3Signer integration (handled in Prompt 311).
 - Continuous node telemetry and Prometheus metrics scraping (handled in Prompt 310).
 - Smart contract business logic (handled in Prompts 303-307).

## Technology to Use
- **Blockchain Node Engine:** **Hyperledger Besu (v24.x+ Enterprise Ethereum Client)**.
  *Justification:* Besu is written in Java, engineered for enterprise reliability, and natively supports QBFT (Quorum Byzantine Fault Tolerance). QBFT provides deterministic finality in 1 block, dynamic validator additions/removals via on-chain voting or contract calls, zero gas economics, and native support for external transaction signing via Web3Signer.
- **Consensus Algorithm:** QBFT ($N=4$ validators minimum, $F=1$ fault tolerance, $2F+1=3$ quorum required for commit).
- **Storage Engine:** RocksDB with persistent NVMe SSD storage.
- **Infrastructure Orchestration:** Kubernetes (EKS/GKE), Helm v3, Terraform, Ansible.
- **Security & Network:** Calico / Cilium CNI with strict Kubernetes NetworkPolicies, mTLS for administrative JSON-RPC.

## Backend / Infra Touchpoints
- **Cloud Infrastructure:** Multi-region AWS (ap-south-1 Mumbai, ap-south-2 Hyderabad) and GCP (asia-south1) cross-cloud infrastructure to avoid single-cloud dependency.
- **Storage Layer:** High-IOPS AWS `gp3` / `io2` persistent EBS volumes attached to Besu StatefulSets.
- **Key Storage:** HashiCorp Vault / CloudHSM storing validator private keys, accessed strictly via Web3Signer sidecars.
- **Ingress / API Gateways:** Envoy proxy handling JSON-RPC rate limiting and mTLS authentication for backend transaction relayers.

## Blockchain Interaction
- **Network ID & Chain ID:** Private Consortium Chain ID (e.g. `ChainID: 13370` / `NetworkID: 13370`).
- **Consensus & Block Times:** QBFT consensus producing blocks exactly every 2,000 milliseconds (2.0s) under active transactions; empty block idling configured to avoid unnecessary ledger bloat.
- **On-Chain Node Permissioning:** Besu executes `NodeRules.sol` on every incoming P2P connection handshake; any node whose ENODE URI is not in the whitelist smart contract is immediately dropped at the transport layer.
- **On-Chain Account Permissioning:** `AccountRules.sol` contract restricts transaction submission rights strictly to authorized institutional relayer addresses and multi-sig controllers.
- **Zero PII & 1:1 Custody Backing:** The ledger state records cryptographic security token balances and atomic DvP settlements representing fractional equity units backed 1:1 by depository demat balances.

## Step-by-Step Build Instructions
1. Scaffold repository directory structure under `infra/blockchain/besu-network/` with subdirectories: `genesis/`, `contracts/permissioning/`, `helm/besu-node/`, `terraform/`, and `scripts/`.
2. Generate initial validator keypairs using `besu operator generate-blockchain-config` within secure offline environment.
3. Configure `genesis.json` defining:
 - `chainId`: 13370
 - `consensus`: `qbft` with `blockperiodseconds`: 2, `epochlength`: 30000, `requesttimeoutseconds`: 4, and `validators`: list of 4 initial validator addresses.
 - `config`: enable all EVM features through Shanghai/Cancun without public EIP-1559 burn mechanics (base fee fixed to 0).
4. Compile and configure Besu Enterprise Permissioning Smart Contracts (`NodeRules.sol`, `AccountRules.sol`, `AdminRegistry.sol`).
5. Insert pre-compiled / pre-deployed permissioning contracts into the `genesis.json` `alloc` section at deterministic addresses (e.g. `0x0000000000000000000000000000000000008888` and `0x0000000000000000000000000000000000009999`).
6. Build Docker images for Hyperledger Besu incorporating corporate security baselines and non-root runtime users.
7. Create Helm charts (`helm/besu-node/`) supporting three node profiles: `validator`, `rpc-relayer`, and `observer-node`.
8. Configure Kubernetes StatefulSets with persistent volume claim templates requesting 500GB NVMe storage backed by `gp3` storage class with 3000 IOPS / 125 MB/s throughput.
9. Configure P2P networking: set up static bootnodes in separate availability zones with static external IP addresses and Kubernetes Service NodePorts/LoadBalancers.
10. Configure Besu CLI flags in entrypoint: `--network=dev`, `--genesis-file=/etc/genesis.json`, `--data-path=/var/lib/besu/data`, `--p2p-enabled=true`, `--rpc-http-enabled=true`, `--permissions-nodes-contract-enabled=true`, `--metrics-enabled=true`.
11. Deploy 2 dedicated Bootnodes across primary cloud regions and verify P2P discovery between nodes.
12. Deploy the 4 initial Validator nodes across isolated VPCs/Kubernetes clusters (Growww Core, Custodian, Clearing, GIFT City Gateway).
13. Verify QBFT consensus initiation: inspect Besu logs for `QBFT: Block proposed`, `QBFT: Commit phase complete`, and block height increments every 2 seconds.
14. Deploy 2 high-throughput RPC Relayer nodes with WebSocket/HTTP JSON-RPC enabled behind an Envoy load balancer for internal microservices.
15. Deploy 1 read-only Observer Node for regulatory sandbox simulation and verify it synchronizes blocks without participating in consensus rounds.

## Interfaces / Contracts

### Besu Genesis Configuration (`genesis.json` schema)
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
    "zeroBaseFee": true,
    "qbft": {
      "blockperiodseconds": 2,
      "epochlength": 30000,
      "requesttimeoutseconds": 4,
      "validatorcontractaddress": "0x0000000000000000000000000000000000007777",
      "validators": [
        "0x9e8b7c6d5e4f3a2b1c0d9e8b7c6d5e4f3a2b1c01",
        "0x8d7c6b5a4f3e2d1c0b9a8d7c6b5a4f3e2d1c0b02",
        "0x7c6b5a4d3e2f1c0a9b8c7c6b5a4d3e2f1c0a9b03",
        "0x6b5a4d3c2e1f0a9b8c7d6b5a4d3c2e1f0a9b8c04"
      ]
    },
    "permissions": {
      "node": "0x0000000000000000000000000000000000008888",
      "account": "0x0000000000000000000000000000000000009999"
    }
  },
  "nonce": "0x0",
  "timestamp": "0x66000000",
  "extraData": "0xf83a...",
  "gasLimit": "0x1fffffffffffff",
  "difficulty": "0x1",
  "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
  "alloc": {}
}
```

### On-Chain Node Permissioning Interface (`NodeRules.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface INodeRules {
    event NodeAdded(bool indexed active, bytes32 indexed enodeHigh, bytes32 indexed enodeLow, bytes16 ip, uint16 port);
    event NodeRemoved(bool indexed active, bytes32 indexed enodeHigh, bytes32 indexed enodeLow, bytes16 ip, uint16 port);

    function enodeAllowed(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external view returns (bool isAllowed);

    function addEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external;

    function removeEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external;
}
```

## Security & Compliance Notes
- **Consensus Resilience:** 4-validator QBFT setup tolerates 1 failed/offline validator ($N=4, F=1, Quorum=3$). If 2 validators go offline simultaneously, the chain safely halts block creation to preserve state integrity rather than forking.
- **Node Isolation & Network Firewalls:** Validator nodes must NOT expose JSON-RPC HTTP/WS ports to external networks; only the P2P port (`30303`) is exposed via mutual TLS / restricted IP security groups between consortium participants.
- **Zero Raw Private Key Exposure:** Validator consensus keys are held in HSMs and accessed exclusively via Web3Signer sidecars over Unix domain sockets or mTLS endpoints (Prompt 311).
- **Audit & Regulatory Compliance:** Read-only Observer Nodes allow SEBI / RBI compliance officers to independently verify blocks and smart contract state in real time without having custody of assets or voting power.

## Acceptance Criteria
- [ ] Multi-node Besu QBFT network successfully bootstraps and produces blocks at deterministic 2.0-second intervals.
- [ ] Dynamic node permissioning verified: non-whitelisted enodes are rejected by `NodeRules.sol` during P2P handshake.
- [ ] Fault tolerance tested: gracefully taking 1 validator offline allows consensus to continue producing blocks; taking a second node offline halts block production safely without corrupting RocksDB state.
- [ ] Terraform and Helm charts deploy identical, repeatable clusters in staging and production environments.
- [ ] Observer node synchronizes 100% of historical blocks and logs with zero consensus permissions.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Platform Selection), Prompt `101` (System Architecture Overview), Prompt `802` (Kubernetes Cluster Architecture).
- **Parallel Tasks:** Prompt `311` (Validator Key Management HSM), Prompt `805` (Infrastructure as Code).
- **Subsequent Prompts Enabled:** Prompt `303` (Token Issuance Contract), Prompt `306` (Settlement DvP Contract), Prompt `310` (Chain Node Monitoring).
