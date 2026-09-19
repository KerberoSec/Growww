# Runbook 30: Copy Trading Replication Lag & Follower Circuit Breaker

**Runbook ID:** RBK-OPS-030  
**Severity Tier:** P2 (High)  
**Authority:** Copy Trading Operations Lead / Lead Matching Engine Engineer  

---

## 1. Description & Trigger Conditions
Triggered when:
- Copy execution dispatcher experiences replication fanout latency $> 50\text{ms}$.
- Follower child order rejection rate breaches threshold ($> 5\%$ rejected due to slippage).

---

## 2. Remediation Workflow

```
[1. Identify Congested Master Trader Queue]
  - Query metrics: copytrading_replication_latency_ms, copytrading_slippage_rejections_total
  - Identify master trader ID and follower queue depth
                  |
                  v
[2. Scale Copy Execution Workers]
  - Trigger horizontal pod autoscaling (HPA) for copy-trading worker pods
  - Repartition Kafka topic growww.copytrading.master.v1 to 16 partitions
                  |
                  v
[3. Trigger Follower Max-Drawdown Detachments]
  - If master incurs sharp intraday loss, verify automated stop-loss detaches execute cleanly
  - Cancel any orphaned follower child orders resting in matching engine book
                  |
                  v
[4. Verify Normalization & Notify Users]
  - Verify fanout latency returns to $< 15\text{ms}$
  - Dispatch status updates to master traders and active followers
```
