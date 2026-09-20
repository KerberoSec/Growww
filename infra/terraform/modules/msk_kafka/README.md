# Terraform AWS MSK Kafka Cluster Module

Institutional-grade, high-availability Apache Kafka cluster module provisioned on AWS Managed Streaming for Apache Kafka (MSK).

## Architecture Highlights
- **Multi-AZ Resiliency:** 3 broker nodes distributed across 3 independent availability zones in `ap-south-1` (Mumbai) or `ap-south-2` (Hyderabad).
- **Authentication & Security:** SASL/SCRAM and AWS IAM authentication with enforced TLS in-transit encryption and customer-managed KMS key at-rest encryption.
- **Financial Market Broker Optimizations:**
  - `auto.create.topics.enable=false`: Enforces strict topic schema governance.
  - `min.insync.replicas=2` with `default.replication.factor=3`: Zero data-loss guarantees for settlement and order match feeds.
  - `unclean.leader.election.enable=false`: Prevents out-of-sync broker failover.
  - `compression.type=lz4`: Ultra-low CPU overhead with high compression throughput.
- **Observability:** Open Monitoring with Prometheus JMX and Node exporters for microsecond lag tracking.
- **Compliance:** 7-year WORM CloudWatch log retention adhering to SEBI and RBI audit guidelines.

## Usage
```hcl
module "msk_cluster" {
  source = "./modules/msk_kafka"

  cluster_name         = "growww-nbse-prod-msk"
  environment          = "production"
  kafka_version        = "3.6.0"
  broker_instance_type = "kafka.m7g.xlarge"
  vpc_id               = var.vpc_id
  subnet_ids           = var.private_subnet_ids
  kms_key_arn          = var.kms_key_arn
  ebs_volume_size_gb   = 1000
}
```
