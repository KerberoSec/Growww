# Terraform AWS Aurora PostgreSQL HA Module

Institutional-grade, multi-AZ Aurora PostgreSQL 16 high-availability cluster module designed for core clearing, double-entry ledger bookkeeping, and high-frequency trade persistence.

## Architecture Highlights
- **Multi-AZ Replication:** Primary writer with automated failover to multi-AZ read replicas with sub-minute failover SLA.
- **I/O-Optimized Storage Tier (`aurora-iopt1`):** Predictable high-throughput pricing with ultra-low write latency for ledger state persistence.
- **Compliance & Auditing:**
  - `pgaudit` extension enabled for full DDL/Role/Write audit logging.
  - `pg_stat_statements` enabled for query profiling and latency hotspot detection.
  - Enforced SSL connections (`rds.force_ssl = 1`).
  - Automatic AWS Secrets Manager master credential management and rotation.
  - 35-day automated point-in-time recovery backup retention.
- **Observability:** 1-second enhanced monitoring and 2-year Performance Insights retention.

## Usage
```hcl
module "aurora_postgres" {
  source = "./modules/aurora_postgresql_ha"

  cluster_identifier = "growww-nbse-prod-aurora"
  environment        = "production"
  engine_version     = "16.1"
  database_name      = "growww_exchange_ledger"
  master_username    = "growww_admin"
  instance_count     = 3
  instance_class     = "db.r7g.2xlarge"
  vpc_id             = var.vpc_id
  subnet_ids         = var.db_subnet_ids
  kms_key_arn        = var.kms_key_arn
}
```
