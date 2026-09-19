# Hyperledger Besu QBFT Consortium Network

This directory contains the production Infrastructure-as-Code (IaC), genesis specifications, on-chain permissioning contracts, Helm charts, and Terraform definitions for the Growww permissioned securities settlement ledger.

## Consortium Topology

Consensus authority is distributed across four distinct institutional entities:
1. **Growww Core Entity:** Primary settlement relayer and validator node (AWS ap-south-1 Mumbai).
2. **Custodian Partner:** Regulated depository custodian node (GIFT City Equinix Data Center).
3. **Clearing Partner:** Clearing and settlement corporation node (Mumbai Primary DC).
4. **GIFT City Gateway Entity:** International gateway validator node (GIFT City).

In addition, dedicated read-only Observer Nodes are provisioned for regulatory authorities (SEBI, IFSCA, RBI).

## Directory Structure

- `genesis/`: Consortium genesis configuration (`genesis.json`) with QBFT 2-second block intervals, zero gas base fee, and pre-allocated permissioning contract addresses.
- `contracts/permissioning/`: On-chain permissioning smart contracts (`NodeRules.sol`, `AccountRules.sol`, `AdminRegistry.sol`).
- `helm/besu-node/`: Kubernetes Helm chart supporting `validator`, `rpc-relayer`, and `observer-node` profiles.
- `terraform/`: AWS Terraform modules for multi-AZ VPC, security groups, and NVMe EBS storage provisioning.
- `scripts/`: Network bootstrap and health verification automation.

## Quick Start

### 1. Review Genesis Configuration
```bash
cat genesis/genesis.json
```

### 2. Deploy via Helm
```bash
helm upgrade --install besu-validator-1 ./helm/besu-node \
  --set nodeProfile=validator \
  --set replicaCount=1
```

### 3. Verify Health
```bash
./scripts/verify_network.sh http://localhost:8545
```
