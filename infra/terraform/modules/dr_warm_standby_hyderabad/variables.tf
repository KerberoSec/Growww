variable "project_prefix" {
  type        = string
  description = "Project name prefix for resource naming"
  default     = "growww-nbse"
}

variable "environment" {
  type        = string
  description = "Deployment environment (production, staging)"
  default     = "production"
}

variable "mumbai_api_domain" {
  type        = string
  description = "FQDN of the primary Mumbai API endpoint for health probes"
  default     = "api.mumbai.growww-nbse.exchange"
}

variable "mumbai_source_bucket_name" {
  type        = string
  description = "Source S3 bucket in Mumbai (ap-south-1) for replication"
  default     = "growww-nbse-audit-ledger-mumbai"
}

variable "hyderabad_destination_bucket_name" {
  type        = string
  description = "Destination S3 bucket in Hyderabad (ap-south-2) for replication"
  default     = "growww-nbse-audit-ledger-hyderabad"
}

variable "enable_aurora_standby" {
  type        = bool
  description = "Deploy secondary warm standby Aurora PostgreSQL cluster in ap-south-2"
  default     = true
}

variable "aurora_global_cluster_identifier" {
  type        = string
  description = "Identifier of the Aurora Global Database spanning ap-south-1 and ap-south-2"
  default     = "growww-nbse-global-ledger"
}

variable "engine_version" {
  type        = string
  description = "Aurora PostgreSQL engine version"
  default     = "16.1"
}

variable "hyderabad_kms_key_arn" {
  type        = string
  description = "KMS Key ARN in ap-south-2 for encryption at rest"
  default     = ""
}

variable "hyderabad_db_subnet_group_name" {
  type        = string
  description = "DB subnet group name in Hyderabad (ap-south-2)"
  default     = "growww-hyd-db-subnet-group"
}

variable "hyderabad_aurora_security_group_id" {
  type        = string
  description = "Security Group ID for Aurora instances in Hyderabad"
  default     = ""
}

variable "hyderabad_cluster_parameter_group_name" {
  type        = string
  description = "Cluster parameter group name in ap-south-2"
  default     = "default.aurora-postgresql16"
}

variable "hyderabad_instance_count" {
  type        = number
  description = "Number of standby read replica instances in Hyderabad"
  default     = 2
}

variable "hyderabad_instance_class" {
  type        = string
  description = "EC2 instance class for standby Aurora nodes"
  default     = "db.r7g.2xlarge"
}

variable "alarm_action_arns" {
  type        = list(string)
  description = "SNS topic ARNs for DR replication alerts"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Resource tags"
  default = {
    Project    = "Growww-NBSE"
    Compliance = "SEBI-BCP-DR"
  }
}
