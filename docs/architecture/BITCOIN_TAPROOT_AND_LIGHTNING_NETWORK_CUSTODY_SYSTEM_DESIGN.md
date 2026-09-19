# Bitcoin Taproot & Lightning Network Custody Architecture Specification

**Specification ID:** SPEC-ARCH-003-BTC  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Core Bitcoin Infrastructure & Layer-2 Lightning Custody  
**Owner:** Bitcoin Core Engineering & Lightning Infrastructure Group  

---

## 1. Executive Summary & Primitives
The Bitcoin custody subsystem manages institutional native BTC custody, Layer-2 Lightning Network channels, and 1:1 wrapped token minting on Hyperledger Besu.
- **Taproot (BIP 341/342)**: Uses Schnorr signatures and Merkelized Alternative Script Trees (MAST) to aggregate complex institutional multi-sig policies into indistinguishable single-key Taproot outputs.
- **MuSig2 (BIP 327)**: Two-round Schnorr threshold key aggregation across cold HSMs, reducing on-chain footprint and transaction fees by 65%.
- **Lightning Network Channels**: High-capacity zero-confirmation channel factories providing instant, fee-free Bitcoin deposits and withdrawals for retail traders.

---

## 2. Lightning Network Channel Management & Anti-Griefing

```
+----------------------------------------------------------------------------------------------------+
| LIGHTNING NETWORK INGRESS & HTLC ANTI-GRIEFING PIPELINE                                            |
|                                                                                                    |
|  [ Trader Lightning Deposit ] ---> [ Dedicated LND / Core Lightning Node Clustered Hub ]           |
|                                                 |                                                  |
|                                                 v                                                  |
|                                    [ Settled HTLC Invoice? ]                                       |
|                                                 |                                                  |
|                                   YES           v                                                  |
|                   [ Instant TigerBeetle Credit (Ledger ID 1) ]                                     |
|                   (Zero Confirmation Delay, Sub-Second Trading)                                    |
|                                                 |                                                  |
|                                                 v                                                  |
|                        [ Channel Liquidity Rebalancing Daemon (Submarine Swaps) ]                  |
|                        (Loop In / Loop Out via On-Chain Taproot Vault)                             |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 HTLC Timeout & Griefing Defense:
- **Strict CLTV Delta**: Minimum 80-block delta for all accepted incoming HTLCs to ensure safe on-chain claim resolution during mempool congestion.
- **Peer Reputation Scoring**: Disconnects and re-routes around channels exhibiting high rates of unresolved or stalled HTLCs.
- **Submarine Swaps**: Automated rebalancing using Boltz / Loop protocols to convert on-chain UTXO reserves into Lightning inbound capacity dynamically.
