# Runbook 01: Trading Halt & Emergency Kill Switch Execution

**Runbook ID:** RBK-OPS-001  
**Severity Tier:** P1 (Critical)  
**Authority:** On-Duty Market Operations Officer / Risk Officer (Single Authority to Halt; Four-Eyes to Resume)  
**Target Execution Latency:** < 1.0 Second  

---

## 1. Trigger Conditions
- Matching engine logic error generating invalid trade fills or prices.
- Circuit breaker band breach exceeding statutory limits without automated halt.
- Statutory regulatory order from SEBI, RBI, or IFSCA.
- Critical proof-of-reserve mismatch or depository connectivity disruption.
- Security incident or suspected cryptographic key compromise.

---

## 2. Emergency Halt Procedure (Single-Authority)

Execute the emergency market halt command via the Admin CLI or API Gateway endpoint:

```bash
# Venue-wide immediate halt
curl -X POST https://127.0.0.1:8019/api/v1/admin/circuit/halt \
  -H "Authorization: Bearer ${OPERATOR_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "scope": "GLOBAL_VENUE",
    "reason": "Suspected market pricing anomaly - emergency halt",
    "retain_resting_orders": true
  }'
```

### Verification Steps
1. Verify `trading-calendar-service` transitions all instrument states to `HALTED_CIRCUIT` or `HALTED_REGULATORY`.
2. Inspect API Gateway logs confirming all new order submissions receive `HTTP 423 Locked` (Market Halted).
3. Confirm Kafka topic `growww.instrument.state.v1` broadcasts the halt event to all consumers.

---

## 3. Resume Procedure (Mandatory Four-Eyes Control)

1. Root cause must be diagnosed, resolved, and documented in an incident ticket.
2. Initiator submits a market resume request specifying a **Pre-Open Call Auction Window (15 minutes)**.
3. Second authorized officer verifies system integrity and cryptographically approves the resume request in `growww_admin`.
4. Continuous matching resumes strictly following completion of the call auction equilibrium price discovery.
