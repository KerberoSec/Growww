# Runbook 33: Multi-Region Disaster Recovery (DR) Failover Execution

**Runbook ID:** RBK-OPS-033  
**Severity Tier:** P0 (Critical Outage)  
**Authority:** Chief Technology Officer / Head of Site Reliability Engineering (SRE)  

---

## 1. Description & Trigger Conditions
Triggered when:
- Complete catastrophic outage of AWS Primary Region (Mumbai - `ap-south-1`).
- Loss of quorum across primary datacenter nodes requiring emergency failover to Secondary Region (Frankfurt - `eu-central-1`).

---

## 2. Remediation Workflow

```
[1. Declare Severity P0 Disaster State]
  - CTO and SRE Lead declare formal disaster recovery activation
  - Notify executive leadership, compliance, and regulatory bodies
                  |
                  v
[2. Promote Secondary Aurora Global Database]
  - Promote Frankfurt Aurora read replica to standalone primary writer:
      aws rds failover-global-cluster --global-cluster-identifier growww-global-db --target-db-cluster-identifier growww-db-frankfurt
  - Verify Aurora promotion completes in < 30 seconds
                  |
                  v
[3. Activate Secondary Kubernetes & Matching Cluster]
  - Scale up matching engine, API gateway, and core microservices in Frankfurt EKS cluster
  - Replay matching engine state from last uncommitted Raft WAL snapshot
  - Start Hyperledger Besu validator nodes in secondary region
                  |
                  v
[4. Shift Global DNS Traffic via Route 53]
  - Update Route 53 health-check weighted DNS records to 100% Frankfurt ingress
  - Invalidate CloudFront CDN cache
                  |
                  v
[5. Post-Failover Verification]
  - Execute automated smoke test suite: make smoke
  - Verify sub-10μs matching, WebSocket streaming, and ledger journal balance
  - Total Target RTO: < 60 seconds; Target RPO: 0 seconds for engine Raft state
```
