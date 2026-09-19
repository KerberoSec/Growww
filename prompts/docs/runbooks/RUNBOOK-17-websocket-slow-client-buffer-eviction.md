# Runbook 17: WebSocket Slow Client Buffer Eviction & Conflation Degradation

**Runbook ID:** RBK-OPS-017  
**Severity Tier:** P2 (High)  
**Authority:** API Gateway / Network Operations Lead  

---

## 1. Description & Trigger Conditions
Triggered when API Gateway WebSocket egress buffers experience systemic slow consumer saturation:
- Percentage of connections in Conflation Mode ($> 80\%$ buffer occupancy) exceeds 15% across a cluster node.
- Involuntary disconnects due to ring buffer overflow exceed threshold (> 50 disconnects/min).

---

## 2. Remediation Workflow

```
[1. Isolate Congested Gateway Pods]
  - Query metrics: websocket_conflation_mode_connections, websocket_buffer_overflow_disconnects_total
  - Identify affected gateway nodes and target consumer IP subnets or user IDs
                  |
                  v
[2. Evaluate Upstream Ingestion & Network Saturation]
  - Check gateway CPU, network egress bandwidth, and memory pressure
  - If network bandwidth saturated, trigger horizontal autoscaling (HPA) of API Gateway pods
                  |
                  v
[3. Transition High-Lag Clients to Conflation Mode]
  - Ensure 200ms conflated snapshot aggregation is active for lagging subscriptions
  - Force disconnect clients lagging beyond 2000ms SLA with code 1008 (ERR_STREAM_DESYNC)
                  |
                  v
[4. Verify Normalization]
  - Confirm WebSocket egress queue depth returns below 20% capacity
  - Verify client reconnections bootstrap cleanly via full Level-2 REST snapshot
```
