variable "aws_region" {
  description = "Primary AWS region for Growww Besu cluster"
  type        = string
  default     = "ap-south-1"
}

variable "secondary_aws_region" {
  description = "Secondary AWS region for DR standby"
  type        = string
  default     = "ap-south-2"
}

variable "environment" {
  description = "Deployment environment name"
  type        = string
  default     = "production"
}

variable "vpc_cidr" {
  description = "VPC CIDR block for Besu validator network"
  type        = string
  default     = "10.100.0.0/16"
}

variable "validator_count" {
  description = "Number of initial consortium validators"
  type        = number
  default     = 4
}

variable "ebs_volume_size" {
  description = "Size in GB of EBS volume per node"
  type        = number
  default     = 500
}

variable "ebs_iops" {
  description = "Provisioned IOPS for gp3 storage"
  type        = number
  default     = 3000
}
