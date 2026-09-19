# Runbook 11: WebSocket Connection Storm & Market Data Conflation

**Runbook ID:** RBK-OPS-011  
**Severity Tier:** P2 (High)  
**Authority:** Frontend Platform Engineer / Edge Gateway SRE  

---

## 1. Description & Alert Triggers
- PromQL Alert: `WebSocketConnectionRate > 50000/sec` during market open window (09:14:00 to 09:16:00 IST).
- Edge gateway memory utilization $> 85\%$.

---

## 2. Remediation Workflow
1. **Activate Ticker Conflation Window:**
   Dynamically increase Level 2 broadcast conflation from 20ms to 100ms via Envoy runtime config:
   ```bash
   curl -X POST http://127.0.0.1:8001/runtime_modify?market_data.conflation_ms=100
   ```
2. **Apply Ingress Connection Rate Limiting:**
   Throttle incoming new TCP connections at the edge load balancer to 20,000 handshakes/second with client backoff headers.
3. **Scale Edge Gateway Replicas:**
   Trigger HPA scaling on Envoy ingress gateway pods from 20 to 60 replicas.
