# Runbook 03: DvP Settlement Backlog Remediation

**Runbook ID:** RBK-OPS-003  
**Severity Tier:** P2 (High)  
**Authority:** Settlement Operations Lead / On-Duty SRE  

---

## 1. Symptoms & Alert Triggers
- PromQL Alert: `SettlementPendingTransactions > 500` for > 5 minutes.
- Kafka topic `growww.settlement.instruction.v1` consumer lag increasing steadily.
- CloudHSM Relayer account gas balances dropping below low-water mark (0.5 ETH equivalent).

---

## 2. Diagnostic Steps
1. **Check Relayer Account Nonces & Balances:**
   ```bash
   curl -s http://127.0.0.1:8008/api/v1/relayers/health
   ```
2. **Inspect Besu Node Transaction Pool:**
   ```bash
   curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_besuStatistics","params":[],"id":1}' \
     -H "Content-Type: application/json" http://127.0.0.1:8545
   ```
3. **Verify Depository Gateway Connectivity:**
   Check connectivity status to NSDL/CDSL APIs.

---

## 3. Remediation Workflow
1. If a specific relayer nonce is stuck, trigger the gas escalator watchdog to submit a 1.25x gas replacement transaction.
2. If the bottleneck is downstream depository latency, scale out `settlement-service` consumer workers up to 16 replicas.
3. If the underlying depository connection is offline, pause new settlement dispatches while keeping trades safely queued in the outbox.
