variable "cluster_identifier" {
  type        = string
  description = "Identifier for the Aurora PostgreSQL cluster"
  default     = "growww-nbse-aurora-primary"
}

variable "environment" {
  type        = string
  description = "Deployment environment (production, staging, testnet)"
  default     = "production"
}

variable "engine_version" {
  type        = string
  description = "Aurora PostgreSQL database engine version"
  default     = "16.1"
}

variable "database_name" {
  type        = string
  description = "Default database name created upon initialization"
  default     = "growww_exchange_ledger"
}

variable "master_username" {
  type        = string
  description = "Master username for database admin access"
  default     = "growww_admin"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where the Aurora cluster is provisioned"
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of isolated database subnet IDs spanning at least 3 AZs"
}

variable "allowed_cidr_blocks" {
  type        = list(string)
  description = "List of CIDR blocks permitted to connect to PostgreSQL port 5432"
  default     = ["10.0.0.0/16"]
}

variable "allowed_security_group_ids" {
  type        = list(string)
  description = "List of Security Group IDs (e.g. EKS services) permitted to connect"
  default     = []
}

variable "instance_count" {
  type        = number
  description = "Total number of Aurora instances (1 primary writer + N read replicas)"
  default     = 3
}

variable "instance_class" {
  type        = string
  description = "Instance class for provisioned nodes (e.g. db.r7g.2xlarge)"
  default     = "db.r7g.2xlarge"
}

variable "enable_serverless_v2" {
  type        = bool
  description = "Whether to use Aurora Serverless v2 scaling"
  default     = false
}

variable "serverless_min_acu" {
  type        = number
  description = "Minimum ACU capacity for Serverless v2"
  default     = 2.0
}

variable "serverless_max_acu" {
  type        = number
  description = "Maximum ACU capacity for Serverless v2"
  default     = 64.0
}

variable "kms_key_arn" {
  type        = string
  description = "ARN of KMS customer-managed key for storage encryption and Secrets Manager"
}

variable "backup_retention_period" {
  type        = number
  description = "Number of days to retain automated backups (1-35 days)"
  default     = 35
}

variable "enable_deletion_protection" {
  type        = bool
  description = "Prevent accidental deletion of the database cluster"
  default     = true
}

variable "skip_final_snapshot" {
  type        = bool
  description = "Skip final snapshot upon cluster destroy"
  default     = false
}

variable "alarm_action_arns" {
  type        = list(string)
  description = "List of SNS topic ARNs for CloudWatch alarms"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to all resources"
  default = {
    Project    = "Growww-NBSE"
    ManagedBy  = "Terraform"
    Compliance = "SEBI-RBI-IFSCA"
  }
}
