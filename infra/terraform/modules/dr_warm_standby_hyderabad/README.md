# Terraform AWS Disaster Recovery Warm Standby Hyderabad (ap-south-2) Module

Institutional Disaster Recovery (DR) warm-standby infrastructure connecting primary region `ap-south-1` (Mumbai) with standby region `ap-south-2` (Hyderabad).

## Architectural Compliance
- **Regulatory SLA:** Meets SEBI Business Continuity Plan (BCP) requirements for clearing corporations and stock exchanges:
  - **RTO (Recovery Time Objective):** < 15 minutes.
  - **RPO (Recovery Point Objective):** < 1 second.
- **Route 53 Application Recovery Controller (ARC):** Multi-region routing controls with automated assertion safety rules preventing double-writer split-brain anomalies.
- **Aurora Global Database:** Sub-second physical storage replication to standby read replicas in Hyderabad.
- **S3 Cross-Region Replication (CRR):** Immutable WORM audit trail log replication with customer KMS key encryption.
