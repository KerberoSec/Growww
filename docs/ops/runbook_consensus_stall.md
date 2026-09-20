# Runbook: Hyperledger Besu QBFT Consensus Stall Incident Response & Disaster Recovery

**Runbook ID:** RBK-OPS-310  
**Severity Tier:** P1 (Critical) - SEBI Material Market Infrastructure Incident  
**Authority:** Blockchain Infrastructure Lead / SRE Incident Commander / Clearing & Settlement Operations Lead  
**SLA:** Block production recovery within 5 minutes; SEBI regulatory preliminary notification within 2 hours.  

---

## 1. Executive Summary & Architectural Context

The Growww / NBSE financial exchange relies on a private Hyperledger Besu permissioned blockchain network for delivery-versus-payment (DvP) trade settlement, clearing waterfalls, and real-time proof-of-reserve attestation.

### Consortium Topology & Quorum Constraints
- **Consensus Engine:** Istanbul/QBFT (Quorum Byzantine Fault Tolerance).
- **Target Block Time:** 2.0 seconds with deterministic 1-block finality.
- **Validator Set ($N=4$):**
  1. `val-growww-core` (Growww Exchange Node, AWS Mumbai `ap-south-1a`)
  2. `val-custodian-nsdl` (NSDL Depository Node, AWS Mumbai `ap-south-1b`)
  3. `val-clearing-corp` (Clearing Corporation Node, AWS Mumbai `ap-south-1c`)
  4. `val-gift-city` (GIFT City Regulatory/Fintech Node, AWS Hyderabad `ap-south-2a`)
- **Byzantine Fault Tolerance ($F$):**
  $$F = \left\lfloor \frac{N - 1}{3} \right\rfloor = \left\lfloor \frac{4 - 1}{3} \right\rfloor = 1$$
- **Consensus Quorum ($Q$):**
  $$Q = 2F + 1 = 3 \text{ validators}$$
  Consensus requires commit signatures from at least **3 out of 4 validators** for every block. If **2 or more validators** fail or become partitioned, block production immediately halts.

---

## 2. Alert Triggers & PagerDuty Escalation

This runbook is triggered automatically when any of the following Prometheus / Alertmanager alerts fire:

| Alert Name | PromQL Expression | Threshold | Severity | Immediate Action |
| :--- | :--- | :--- | :--- | :--- |
| **`BesuConsensusStalled`** | `increase(besu_blockchain_chain_head_block_number{job="besu-validators"}[30s]) == 0` | 10s | **P1 (Critical)** | Page Blockchain Primary On-Call |
| **`BesuMultipleValidatorsDown`** | `count(up{job="besu-validators"} == 0) >= 2` | 10s | **P1 (Critical)** | Declare Major Severity Incident |
| **`BesuQBFTHighRoundNumber`** | `besu_qbft_round_number{job="besu-validators"} > 2` | 10s | **P1 (Critical)** | Investigate proposer deadlock |
| **`BesuPeerCountCriticallyLow`** | `besu_peers_current{job="besu-validators"} < 3` | 20s | **P1 (Critical)** | Inspect P2P mesh & security groups |
| **`Web3SignerDown`** | `up{job="web3signer-sidecars"} == 0` | 15s | **P1 (Critical)** | Verify HSM connectivity & tokens |

---

## 3. Immediate Diagnostic Checklist (Triage in < 60 Seconds)

Execute the diagnostic script or run JSON-RPC health queries across all 4 validator endpoints:

### Step 3.1: Query Block Height & Synchronization Status
Run from an internal operations bastion or management pod:

```bash
for NODE in val-growww-core val-custodian-nsdl val-clearing-corp val-gift-city; do
  echo "=== Checking Node: $NODE ==="
  curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    -H "Content-Type: application/json" http://$NODE:8545 | jq -r '.result' | xargs -I {} printf "Hex: {} | Decimal: %d\n" {}
  curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":2}' \
    -H "Content-Type: application/json" http://$NODE:8545 | jq -c '.result'
done
```

### Step 3.2: Inspect P2P Peer Count & Network Connectivity
```bash
for NODE in val-growww-core val-custodian-nsdl val-clearing-corp val-gift-city; do
  PEERS=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":3}' \
    -H "Content-Type: application/json" http://$NODE:8545 | jq -r '.result')
  printf "%-25s Connected Peers: %d\n" "$NODE" "$PEERS"
done
```
*Expected: Each validator must report at least 3 connected peers.*

### Step 3.3: Verify Active QBFT Validator Set
```bash
curl -s -X POST --data '{"jsonrpc":"2.0","method":"qbft_getValidatorsByBlockNumber","params":["latest"],"id":4}' \
  -H "Content-Type: application/json" http://val-growww-core:8545 | jq '.result'
```
*Confirm all 4 institutional addresses are listed in the validator set.*

### Step 3.4: Inspect Besu Node Logs for Consensus Round Timeouts
Search the live log streams for QBFT phase failures:
```bash
kubectl -n blockchain-system logs -l app.kubernetes.io/name=besu-validator --tail=100 | \
  grep -E "(RoundChange|Prepare|Commit|timeout|Unable to reach consensus|New round)"
```

---

## 4. Remediation Workflows by Failure Scenario

```
                    [BesuConsensusStalled Alert Fires]
                                   |
                     [Step 3.1-3.4 Initial Triage]
                                   |
         -------------------------------------------------------
        |                       |               |               |
        v                       v               v               v
  [Scenario A]            [Scenario B]    [Scenario C]    [Scenario D]
  P2P Partition           Node Crash      Web3Signer/HSM  Round Deadlock
  Peers < 3               Pod OOM/Term    Signing Errors  Round > 2
        |                       |               |               |
  Re-add static           Safe restart    Refresh token   Discard stuck
  enode peers             with lock-check restart signer  block proposal
        |                       |               |               |
         -------------------------------------------------------
                                   |
                         [Consensus Resumed?]
                           /              \
                        YES                NO
                         |                  |
               [Post-Check Step 5]    [Scenario E: Quorum Loss]
               Verify block cadence   Consortium Coordinated Reboot
```

---

### Scenario A: Network Partition / P2P Disconnect (Peers < 3)

**Root Cause:** Security group modification, cross-AZ transit gateway flap, or AWS Direct Connect degradation between Mumbai and Hyderabad.

1. **Verify P2P Listening Port (30303 TCP/UDP):**
   ```bash
   nc -zv $TARGET_NODE_IP 30303
   ```
2. **Extract Enode URI of Healthy Validator:**
   ```bash
   ENODE=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' \
     -H "Content-Type: application/json" http://val-growww-core:8545 | jq -r '.result.enode')
   echo "Target Enode: $ENODE"
   ```
3. **Manually Force P2P Peering on Partitioned Node:**
   ```bash
   curl -X POST --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE\"],\"id\":2}" \
     -H "Content-Type: application/json" http://$PARTITIONED_NODE:8545
   ```
4. Confirm peer count increments back to $\ge 3$ within 10 seconds.

---

### Scenario B: Validator Node Crash / Host Failure

**Root Cause:** Out-of-memory (OOM) killer invoked on JVM, NVMe disk full, or underlying hypervisor hardware retirement.

> [!CAUTION]
> **ANTI-SLASHING WARNING:** Never start two instances of the same validator using the same private key at the same time. Doing so will produce conflicting QBFT proposals/commits and trigger protocol slashing or invalid state fork.

1. **Check Pod Exit Code & OOM Status:**
   ```bash
   kubectl -n blockchain-system describe pod val-growww-core-0
   ```
2. **Verify Lock File Removal:**
   If Besu crashed abruptly, the RocksDB lock file may prevent JVM restart:
   ```bash
   # Check if process is truly dead before removing lock
   kubectl -n blockchain-system exec -it val-growww-core-0 -- rm -f /var/lib/besu/data/DATABASE_METADATA.lock
   ```
3. **Restart Single Validator Pod:**
   ```bash
   kubectl -n blockchain-system rollout restart statefulset/val-growww-core
   ```
4. **Monitor Syncing to Tip:**
   ```bash
   watch -n 1 "curl -s -X POST --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}' -H 'Content-Type: application/json' http://val-growww-core:8545 | jq -r '.result'"
   ```

---

### Scenario C: Web3Signer HSM Cryptographic Failure

**Root Cause:** IAM role expiration, AWS CloudHSM / KMS endpoint throttling, or invalid Web3Signer signing key configuration.

1. **Check Web3Signer Upstream Health:**
   ```bash
   curl -i http://signer-val-growww-core:9001/healthcheck
   ```
2. **Inspect Signing Error Logs:**
   ```bash
   kubectl -n blockchain-system logs -l app.kubernetes.io/name=web3signer --tail=100 | grep -i "error"
   ```
3. **Verify HSM Key Availability:**
   ```bash
   curl -s http://signer-val-growww-core:9001/api/v1/eth2/publicKeys
   ```
4. **Restart Web3Signer Sidecar:**
   ```bash
   kubectl -n blockchain-system rollout restart deployment/signer-val-growww-core
   ```
   Confirm Besu can request consensus signatures without `SignerException`.

---

### Scenario D: QBFT Round Change Deadlock / Stuck Proposal

**Root Cause:** A proposed block contains an invalid transaction or gas discrepancy that peers reject, but the current round leader repeatedly retries or fails to advance round-robin proposer index.

1. **Check Current Round Number:**
   ```bash
   curl -s http://val-growww-core:9545/metrics | grep "besu_qbft_round_number"
   ```
2. **Discard Poisonous Proposal on Current Proposer:**
   ```bash
   curl -X POST --data '{"jsonrpc":"2.0","method":"qbft_discardProposal","params":[],"id":1}' \
     -H "Content-Type: application/json" http://$PROPOSER_NODE:8545
   ```
3. **Trigger Immediate Round Advance:**
   If the round does not increment automatically, restart the currently scheduled round leader to allow round change timer to advance to the next honest validator.

---

### Scenario E: Total Quorum Loss ($F \ge 2$ Offline) - Disaster Recovery

**Root Cause:** Catastrophic multi-AZ cloud outage or corrupted shared block state affecting 2 or more validators simultaneously. Quorum ($3/4$) is broken.

1. **Declare Level-1 Disaster Recovery Incident:**
   Notify Chief Technology Officer, Head of Custody, and SEBI Compliance Officer.
2. **Freeze Trading Matching Engines:**
   Engage trading halt kill switch ([`RUNBOOK-01-trading-halt-kill-switch.md`](file:///home/kali/Growww/docs/runbooks/RUNBOOK-01-trading-halt-kill-switch.md)) to prevent unmatched settlement liabilities while the chain is halted.
3. **Determine Canonical Chain Head:**
   Inspect all 4 nodes and locate the node with the highest valid block number and valid extraData QBFT seal:
   ```bash
   for IP in 10.0.1.10 10.0.2.10 10.0.3.10 10.0.4.10; do
     echo "Checking node $IP:"
     curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' \
       -H "Content-Type: application/json" http://$IP:8545 | jq '{number: .result.number, hash: .result.hash}'
   done
   ```
4. **Re-synchronize Partitioned Nodes:**
   Export canonical RocksDB backup snapshot from highest valid node and restore to lagging/corrupted nodes.
5. **Simultaneous Coordinated Node Resumption:**
   Start all 4 validators within 10 seconds of each other. Once 3 nodes handshake over P2P port 30303, QBFT will achieve commit quorum and produce the next block.

---

## 5. Post-Recovery Verification & Health Validation

Before resuming normal exchange trading and DvP settlement:

- [ ] **Continuous Block Cadence:** Observe block height incrementing every 2.0s ($\pm 200\text{ms}$) for at least 30 consecutive blocks.
  ```bash
  for i in {1..30}; do
    curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      -H "Content-Type: application/json" http://val-growww-core:8545 | jq -r '.result' | xargs printf "%d\n"
    sleep 2
  done
  ```
- [ ] **Consensus Seal Verification:** Inspect latest block `extraData` to confirm it contains $\ge 3$ valid ECDSA commit seals:
  ```bash
  curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' \
    -H "Content-Type: application/json" http://val-growww-core:8545 | jq -r '.result.extraData'
  ```
- [ ] **Mempool Drain:** Confirm `besu_transaction_pool_transactions_count` decreases to baseline (< 50).
- [ ] **Zero Divergence:** Confirm `max(besu_blockchain_chain_head_block_number) - min(besu_blockchain_chain_head_block_number) == 0`.
- [ ] **Unfreeze Trading Ingress:** Re-enable trading matching engines and settlement relayer gateways.

---

## 6. Regulatory Compliance & SEBI Reporting Requirements

Under SEBI Master Circular for Clearing Corporations and Institutional Trading Platforms:

1. **Preliminary Incident Notification (Within 2 Hours):**
   - Submit Form Annexure-A incident report to SEBI Technology Advisory Committee (`tech-incident@sebi.gov.in`).
   - Report must state: exact outage timestamp (UTC / IST), duration of block stall, total number of uncommitted DvP settlement batches, and immediate remediation applied.
2. **Forensic Evidence Preservation:**
   - Archive Besu log files: `/var/log/besu/*.log` across all 4 validators.
   - Take Prometheus metric snapshot from 30 minutes prior to outage to 30 minutes post-recovery.
   - Preserve Web3Signer audit trails.
3. **Comprehensive Root-Cause Analysis (RCA) Submission (Within 14 Days):**
   - Must include timeline, hardware/network root cause, preventive architecture upgrades, and auditor sign-off.
