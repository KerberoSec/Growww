variable "cluster_name" {
  type        = string
  description = "Name of the MSK Kafka Cluster"
  default     = "growww-nbse-msk-primary"
}

variable "environment" {
  type        = string
  description = "Deployment environment (production, staging, testnet)"
  default     = "production"
}

variable "kafka_version" {
  type        = string
  description = "Apache Kafka engine version"
  default     = "3.6.0"
}

variable "number_of_broker_nodes" {
  type        = number
  description = "Total number of broker nodes across availability zones (must be multiple of number of subnets, min 3)"
  default     = 3
}

variable "broker_instance_type" {
  type        = string
  description = "EC2 instance type for Kafka brokers"
  default     = "kafka.m7g.xlarge"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where MSK cluster is deployed"
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of private subnet IDs across 3 AZs for broker deployment"
}

variable "allowed_cidr_blocks" {
  type        = list(string)
  description = "List of CIDR blocks permitted to communicate with MSK brokers"
  default     = ["10.0.0.0/16"]
}

variable "monitoring_cidr_blocks" {
  type        = list(string)
  description = "List of CIDR blocks permitted to scrape Prometheus metrics (JMX/Node)"
  default     = ["10.0.0.0/16"]
}

variable "allowed_security_group_ids" {
  type        = list(string)
  description = "List of security group IDs permitted to communicate with MSK brokers (e.g. EKS worker nodes)"
  default     = []
}

variable "ebs_volume_size_gb" {
  type        = number
  description = "Size in GiB of the EBS volume attached to each broker node"
  default     = 1000
}

variable "ebs_provisioned_throughput_enabled" {
  type        = bool
  description = "Enable provisioned throughput for EBS storage volumes"
  default     = true
}

variable "ebs_provisioned_throughput_mb_per_sec" {
  type        = number
  description = "Throughput in MiB/s for provisioned storage"
  default     = 250
}

variable "kms_key_arn" {
  type        = string
  description = "ARN of customer-managed KMS key for encryption at rest and CloudWatch logs"
}

variable "enable_sasl_iam" {
  type        = bool
  description = "Enable IAM client authentication alongside SASL/SCRAM"
  default     = true
}

variable "scram_secret_association_arns" {
  type        = list(string)
  description = "List of AWS Secrets Manager secret ARNs containing SASL/SCRAM credentials"
  default     = []
}

variable "enable_cloudwatch_logging" {
  type        = bool
  description = "Enable broker log delivery to CloudWatch Log Groups"
  default     = true
}

variable "log_retention_days" {
  type        = number
  description = "Retention period in days for CloudWatch broker logs (7 years = 2557 days for WORM compliance)"
  default     = 2557
}

variable "enable_s3_logging" {
  type        = bool
  description = "Enable long-term immutable S3 broker log archiving"
  default     = false
}

variable "s3_logging_bucket_name" {
  type        = string
  description = "S3 bucket name for archival broker logs"
  default     = null
}

variable "s3_logging_prefix" {
  type        = string
  description = "Prefix within S3 bucket for broker logs"
  default     = "msk-broker-logs/"
}

variable "alarm_action_arns" {
  type        = list(string)
  description = "List of SNS topic ARNs for CloudWatch alarms notifications (PagerDuty / OpsGenie)"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Tags to attach to all provisioned resources"
  default = {
    Project     = "Growww-NBSE"
    ManagedBy   = "Terraform"
    Compliance  = "SEBI-RBI-IFSCA"
  }
}
