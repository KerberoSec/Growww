# 805 - Cloud & Blockchain Infrastructure as Code (Terraform / OpenTofu)

## Purpose
In a regulated financial and digital asset infrastructure, manual cloud resource provisioning is unacceptable. All cloud resources, network perimeters, data storage layers, Hardware Security Modules (CloudHSM / KMS), managed Kubernetes clusters, and dedicated Hyperledger Besu consortium validator nodes must be completely codified, version-controlled, auditable, and reproducibly provisioned via declarative Infrastructure as Code (IaC).

This prompt establishes the master Terraform / OpenTofu module architecture for Growww. It enforces a strict multi-region, multi-entity isolation strategy separating the domestic SEBI-regulated entity (AWS Mumbai `ap-south-1` / GCP `asia-south1`) from the international GIFT City IFSCA gateway entity, incorporating automated drift detection, security policy scanning (Checkov / tfsec), and state locking.

## What You Are Building
A modular, enterprise-grade Terraform / OpenTofu codebase and automation pipeline:
- `infra/terraform/modules/vpc/`: Multi-tier VPC architecture (Public, Private App, Private DB, Private HSM/Blockchain, Transit Gateway).
- `infra/terraform/modules/eks/`: Production AWS EKS / GCP GKE cluster provisioning with managed node groups, custom launch templates, and IRSA (IAM Roles for Service Accounts).
- `infra/terraform/modules/rds/`: Multi-AZ AWS Aurora PostgreSQL cluster with automated KMS storage encryption, performance insights, and read replicas.
- `infra/terraform/modules/elasticache/`: Redis cluster with in-transit and at-rest encryption, auto-failover, and multi-AZ replication.
- `infra/terraform/modules/msk/`: Apache Kafka (AWS MSK / Strimzi on K8s) cluster with TLS authentication and Schema Registry integration.
- `infra/terraform/modules/cloudhsm/`: AWS CloudHSM / HashiCorp Vault cluster provisioning for validator signing keys and transaction relayer HSMs.
- `infra/terraform/modules/besu-nodes/`: Dedicated EC2/Compute instances and NVMe EBS storage specifically tuned for Hyperledger Besu QBFT validator nodes.
- `infra/terraform/environments/`: Environment root configurations for `dev`, `staging`, `prod-mumbai`, and `prod-gift-city`.
- `.github/workflows/ci-terraform.yml`: Automated Terraform PR validation workflow running `tflint`, `tfsec`, `checkov`, and generating automated speculative execution plans via Atlantis / Terraform Cloud.

## Scope Boundaries
- **In Scope:**
 - Complete cloud infrastructure provisioning across AWS/GCP for domestic and GIFT City entities.
 - Dedicated VPCs, subnet topologies, security groups, NAT gateways, and Route 53 private hosted zones.
 - Managed database and messaging infrastructure (Aurora Postgres, ElastiCache Redis, MSK Kafka).
 - Dedicated blockchain compute infrastructure (Besu validator instances with high-IOPS NVMe storage and HSM key integration).
 - IAM least-privilege policies, OIDC trust providers, and IRSA role mappings.
 - S3 backend state management with DynamoDB distributed locking and server-side KMS encryption.
- **Out of Scope / Handled Elsewhere:**
 - In-cluster Kubernetes workload manifests (Prompt 802).
 - Application GitOps synchronization via ArgoCD (Prompt 804).
 - Centralized application logging ingestion (Prompt 807).

## Technology to Use
- **Terraform 1.8+ / OpenTofu 1.7+**: Declarative IaC tool. Justification: Universal industry standard for multi-cloud infrastructure automation with massive ecosystem support, deterministic state management, and strict module composition.
- **AWS Provider (v5.x) / GCP Provider (v5.x)**: Official cloud infrastructure providers.
- **Checkov & tfsec**: Static analysis and policy-as-code security scanners for IaC.
- **Terragrunt / Terraform Workspaces**: Configuration wrapper to maintain DRY (Don't Repeat Yourself) multi-environment deployments.
- **Atlantis / Terraform Cloud**: Automated pull-request-driven plan and apply execution engine with dual-approval enforcement.

## Backend / Infra Touchpoints
- **AWS VPC / GCP VPC**: Isolated VPCs across `ap-south-1` (Mumbai) with Transit Gateway peering to GIFT City enclave.
- **AWS EKS 1.30+**: Managed control plane across 3 Availability Zones.
- **AWS Aurora PostgreSQL 16**: Multi-AZ transactional engine.
- **AWS ElastiCache Redis 7**: Low-latency caching cluster.
- **AWS CloudHSM (FIPS 140-2 Level 3)**: Secure cryptographic key custody cluster.
- **AWS S3 & DynamoDB**: Encrypted remote backend state storage and distributed lock table.

## Blockchain Interaction
The `besu-nodes` Terraform module provisions the dedicated infrastructure for the permissioned Hyperledger Besu consortium ledger:
- **Dedicated Validator Compute**: Provisions dedicated EC2 compute instances (`c6i.2xlarge` or `c7g.2xlarge`) with dedicated CPU pinning and zero hyperthread sharing to prevent consensus latency jitter.
- **High-Throughput Storage**: Provisions dedicated `gp3` / `io2` EBS volumes configured with 10,000 IOPS and 500 MB/s throughput, or direct NVMe instance store for the Besu chain data directory (`/var/lib/besu/data`).
- **Network Isolation & Elastic IPs**: Assigns static Elastic IPs and configures Security Groups allowing QBFT consensus P2P traffic (TCP/UDP 30303) exclusively between authenticated consortium validator nodes.
- **CloudHSM / KMS Validator Key Integration**: Provisions CloudHSM instances and IAM roles enabling validator nodes and transaction relayer services to sign blocks and settlement transactions using hardware-isolated private keys.

## Step-by-Step Build Instructions
1. Establish S3 remote state bucket with versioning, default AWS KMS CMK encryption, and DynamoDB lock table in `ap-south-1`.
2. Scaffold directory structure `infra/terraform/{modules/{vpc,eks,rds,redis,msk,cloudhsm,besu-nodes},environments/{dev,staging,prod-mumbai,prod-gift-city}}`.
3. Build the `vpc` module defining 4 subnet tiers (Public, App, Database, HSM/Blockchain) across 3 Availability Zones with VPC Flow Logs enabled.
4. Build the `eks` module provisioning EKS cluster, managed node groups with custom user-data launch templates, and enabling AWS OIDC provider for IRSA.
5. Build the `rds` module provisioning Aurora PostgreSQL Multi-AZ cluster with KMS encryption, performance insights, and automated daily snapshots.
6. Build the `redis` module provisioning ElastiCache cluster with encryption-in-transit, auth token, and multi-AZ auto-failover.
7. Build the `msk` module provisioning a 3-broker Apache Kafka cluster with IAM / TLS authentication and CloudWatch log streaming.
8. Build the `cloudhsm` module provisioning a dedicated CloudHSM cluster spanning 2 AZs with automated backup policies.
9. Build the `besu-nodes` module provisioning validator instances, Elastic IPs, NVMe volume attachments, and Besu security groups.
10. Implement environment root configurations (`prod-mumbai/main.tf`, `prod-gift-city/main.tf`) invoking child modules with environment-specific parameters.
11. Write `.github/workflows/ci-terraform.yml` to run `terraform fmt -check`, `tflint`, and `checkov -d .` on all PRs.
12. Configure Atlantis or Terraform Cloud integration to generate automated speculative plans and post plan diffs to PR comments.
13. Execute `terraform plan` and `terraform apply` in `dev` and `staging` environments to verify zero-error provisioning.
14. Validate multi-region connectivity between Mumbai VPC and GIFT City VPC over AWS Transit Gateway.

## Interfaces / Contracts
```hcl
# infra/terraform/modules/besu-nodes/main.tf (Excerpt)
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "subnet_ids" { type = list(string) }
variable "consortium_peer_cidrs" { type = list(string) }
variable "instance_type" { default = "c6i.2xlarge" }

resource "aws_security_group" "besu_validator_sg" {
  name        = "growww-${var.environment}-besu-validator-sg"
  description = "Security group for Hyperledger Besu QBFT Validator Nodes"
  vpc_id      = var.vpc_id

  # QBFT P2P Peering - Restricted to consortium members
  ingress {
    description = "QBFT Consensus P2P TCP"
    from_port   = 30303
    to_port     = 30303
    protocol    = "tcp"
    cidr_blocks = var.consortium_peer_cidrs
  }

  ingress {
    description = "QBFT Consensus P2P UDP Discovery"
    from_port   = 30303
    to_port     = 30303
    protocol    = "udp"
    cidr_blocks = var.consortium_peer_cidrs
  }

  # JSON-RPC - Restricted to Internal Settlement VPC subnets
  ingress {
    description = "Internal Besu JSON-RPC"
    from_port   = 8545
    to_port     = 8545
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"] # App subnet CIDR
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "growww-${var.environment}-besu-sg"
    Environment = var.environment
    Compliance  = "SEBI-RBI-Regulated"
  }
}

resource "aws_instance" "besu_validator" {
  count                  = length(var.subnet_ids)
  ami                    = data.aws_ami.ubuntu_lts.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_ids[count.index]
  vpc_security_group_ids = [aws_security_group.besu_validator_sg.id]
  iam_instance_profile   = aws_iam_instance_profile.besu_node_profile.name

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 50
    encrypted             = true
    kms_key_id            = var.kms_key_arn
    delete_on_termination = false
  }

  ebs_block_device {
    device_name           = "/dev/xvdf"
    volume_type           = "io2"
    volume_size           = 500
    iops                  = 10000
    encrypted             = true
    kms_key_id            = var.kms_key_arn
    delete_on_termination = false
  }

  tags = {
    Name        = "growww-${var.environment}-besu-validator-${count.index + 1}"
    Environment = var.environment
  }
}
```

## Security & Compliance Notes
- RBI Data Localization & Residency: All compute, databases, S3 buckets, and KMS keys for domestic operations are strictly pinned to AWS Mumbai (`ap-south-1`) / GCP India regions.
- Zero Public Databases: RDS, Redis, Kafka, and Besu nodes have zero public IP addresses and reside exclusively in isolated private subnets.
- Customer Managed Encryption (CMEK): All storage volumes (EBS, RDS, S3, ElastiCache) are encrypted using AWS KMS Customer Managed Keys with annual automatic key rotation.
- Static IaC Security Gate: Checkov and tfsec scans must pass with zero failed checks for CIS AWS Foundations Benchmark compliance.

## Acceptance Criteria
- [ ] Terraform modules provision complete VPC, EKS, RDS, Redis, Kafka, CloudHSM, and Besu validator infrastructure cleanly.
- [ ] Multi-region separation between Mumbai domestic entity and GIFT City international entity is fully enforced in module parameters.
- [ ] Besu validator nodes are provisioned with dedicated `io2` / `gp3` high-IOPS storage and secure consortium firewall rules.
- [ ] Automated CI pipeline executes `tflint`, `checkov`, and `tfsec` security scans, passing all compliance checks.
- [ ] State files are securely encrypted in S3 with DynamoDB distributed locking preventing concurrent state modifications.

## Suggested Order / Dependencies
- Prerequisites: Prompt 002 (Two-Entity Structure), Prompt 108 (Environment Strategy), Prompt 109 (Secrets Management), Prompt 302 (Network Topology).
- Parallel Tasks: Prompt 802 (Kubernetes Architecture), Prompt 806 (Observability Stack).
