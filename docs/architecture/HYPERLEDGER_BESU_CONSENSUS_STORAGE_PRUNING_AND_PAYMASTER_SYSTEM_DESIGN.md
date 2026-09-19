# Hyperledger Besu Consensus, Bonsai Trie Pruning & Paymaster System Design

**Specification ID:** SPEC-ARCH-026-BESU  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Distributed Ledger Technology & Consensus Infrastructure  
**Owner:** Blockchain Core Infrastructure & EVM Architecture Group  

---

## 1. Executive Summary & Consensus Parameters
The platform operates on a sovereign Hyperledger Besu enterprise blockchain consortium:
- **Consensus Protocol**: Quorum Byzantine Fault Tolerance (QBFT) with a fixed 2.0-second block time and deterministic 1-block finality (zero probabilistic reorgs).
- **Cluster Topology**: 4 Validator nodes distributed across isolated Availability Zones in AWS Mumbai and Hyderabad. Survives failure of 1 node ($F = (4-1)/3 = 1$).
- **Bonsai Trie Storage & Background Pruning**: Keeps disk storage under 500GB per validator through continuous trie log pruning and flat key-value reads.
- **Zero-Gas User Experience**: 100% sponsored transaction execution via ERC-4337 Account Abstraction and `PaymasterRelayer.sol`.

---

## 2. Block Production & Relayer Priority Topology

```
+----------------------------------------------------------------------------------------------------+
| BESU CONSENSUS & ACCOUNT ABSTRACTION PAYMASTER PIPELINE                                            |
|                                                                                                    |
|  [ User Signs Trade via EIP-712 ] ---> [ Bundler RPC Gateway ]                                     |
|                                                    |                                               |
|                                                    v                                               |
|                                    [ PaymasterRelayer.sol ]                                        |
|                                    - Validates User Signature                                      |
|                                    - Sponsoring Node Subsidizes Gas (User Pays $0.00 Gas)          |
|                                                    |                                               |
|                                                    v                                               |
|                                    [ 32-Relayer Nonce Sharding Pool ]                              |
|                                    - Monotonic In-Memory Sequence Controllers                      |
|                                                    |                                               |
|                                                    v                                               |
|                        [ Besu QBFT Validator Set (4 Active Nodes, 2.0s Blocks) ]                   |
|                        - Consensus Round Propose -> Prepare -> Commit                             |
|                        - 1-Block Immediate Finality                                                |
|                                                    |                                               |
|                                                    v                                               |
|                        [ Bonsai Trie Storage Engine (RocksDB Flat State) ]                         |
+----------------------------------------------------------------------------------------------------+
```
