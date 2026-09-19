# Besu Enterprise Consortium Architecture & Consensus Specification

## 1. Network Topology & QBFT Consensus
- **Consensus**: Quorum Byzantine Fault Tolerance (QBFT) with a fixed 2.0-second block time and deterministic 1-block finality.
- **Validator Set**: 4 validator nodes distributed across separate Availability Zones in AWS Mumbai and Hyderabad.
  - Fault tolerance: $F = (4 - 1) / 3 = 1$ validator node failure tolerated without chain halt.
- **RPC & Read Nodes**: 8 load-balanced JSON-RPC nodes behind Envoy gateways with mTLS and rate limiting.

## 2. Storage Architecture & Pruning
- **Storage Engine**: Bonsai Trie storage layout for high-throughput state reads and minimal disk amplification.
- **State Pruning**: Continuous background trie pruning keeping disk usage under 500 GB per node.
- **Archive Cluster**: 2 dedicated full archive nodes with unpruned historical state for regulatory audits and compliance queries.

## 3. Gas Pricing & Relayer Policies
- Permissioned network running with zero base gas fee (`min-gas-price = 0`) or fixed nominal gas fee to prevent spam while guaranteeing predictable settlement costs.
- Dedicated transaction pools reserved for authorized exchange relayer addresses (`0xRelayer00` through `0xRelayer31`).
