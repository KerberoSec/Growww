terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Dedicated VPC for Besu Blockchain Network
resource "aws_vpc" "besu_vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "besu-${var.environment}-vpc"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Subnets across multi-AZ
resource "aws_subnet" "besu_subnets" {
  count                   = 3
  vpc_id                  = aws_vpc.besu_vpc.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = false

  tags = {
    Name        = "besu-${var.environment}-subnet-${count.index + 1}"
    Environment = var.environment
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# Security Group for Validators: strictly P2P port 30303 and Prometheus scraping
resource "aws_security_group" "besu_validator_sg" {
  name        = "besu-${var.environment}-validator-sg"
  description = "Security group for consortium Besu validator nodes"
  vpc_id      = aws_vpc.besu_vpc.id

  ingress {
    description = "P2P Consortium Communication"
    from_port   = 30303
    to_port     = 30303
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  ingress {
    description = "Prometheus Metrics Exporter"
    from_port   = 9545
    to_port     = 9545
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "besu-${var.environment}-validator-sg"
    Environment = var.environment
  }
}

# Storage for RocksDB data
resource "aws_ebs_volume" "validator_storage" {
  count             = var.validator_count
  availability_zone = data.aws_availability_zones.available.names[count.index % 3]
  size              = var.ebs_volume_size
  type              = "gp3"
  iops              = var.ebs_iops
  throughput        = 125
  encrypted         = true

  tags = {
    Name        = "besu-validator-${count.index + 1}-data"
    Environment = var.environment
  }
}
