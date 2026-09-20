/**
 * Growww / NBSE Institutional Exchange Infrastructure
 * Module: Disaster Recovery Warm Standby Hyderabad (ap-south-2)
 * Compliance: SEBI Business Continuity Plan (BCP) & 15-Minute RTO / Zero RPO Requirement
 */

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# Route 53 Application Recovery Controller (ARC) Cluster for Regional Routing
resource "aws_route53recoverycontrolconfig_cluster" "dr_arc_cluster" {
  name = "${var.project_prefix}-dr-arc-cluster"

  tags = merge(
    var.tags,
    {
      Name        = "${var.project_prefix}-dr-arc-cluster"
      Environment = var.environment
      Role        = "disaster-recovery-arc"
    }
  )
}

resource "aws_route53recoverycontrolconfig_control_panel" "dr_control_panel" {
  name        = "${var.project_prefix}-dr-control-panel"
  cluster_arn = aws_route53recoverycontrolconfig_cluster.dr_arc_cluster.arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "mumbai_primary_routing_control" {
  name              = "${var.project_prefix}-mumbai-primary-active"
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.dr_arc_cluster.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.dr_control_panel.arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "hyderabad_standby_routing_control" {
  name              = "${var.project_prefix}-hyderabad-standby-active"
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.dr_arc_cluster.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.dr_control_panel.arn
}

# Safety Rule: Ensure at least one region is always active and prevent double-writer split-brain
resource "aws_route53recoverycontrolconfig_safety_rule" "single_primary_assertion" {
  name                   = "${var.project_prefix}-prevent-split-brain-assertion"
  control_panel_arn      = aws_route53recoverycontrolconfig_control_panel.dr_control_panel.arn
  wait_period_ms         = 5000
  inverted               = false

  rule_config {
    type      = "ASSERTION"
    threshold = 1
    inverted  = false
  }

  asserted_controls = [
    aws_route53recoverycontrolconfig_routing_control.mumbai_primary_routing_control.arn,
    aws_route53recoverycontrolconfig_routing_control.hyderabad_standby_routing_control.arn
  ]
}

# Health Checks for Primary Region (Mumbai)
resource "aws_route53_health_check" "mumbai_primary_health" {
  fqdn              = var.mumbai_api_domain
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health/ready"
  failure_threshold = 3
  request_interval  = 10
  enable_sni        = true
  measure_latency   = true

  tags = merge(
    var.tags,
    {
      Name   = "${var.project_prefix}-mumbai-primary-health"
      Region = "ap-south-1"
    }
  )
}

# S3 Cross-Region Replication IAM Role
resource "aws_iam_role" "s3_crr_role" {
  name = "${var.project_prefix}-s3-crr-replication-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_policy" "s3_crr_policy" {
  name = "${var.project_prefix}-s3-crr-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetReplicationConfiguration",
          "s3:ListBucket"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:s3:::${var.mumbai_source_bucket_name}"
      },
      {
        Action = [
          "s3:GetObjectVersionForReplication",
          "s3:GetObjectVersionAcl",
          "s3:GetObjectVersionTagging"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:s3:::${var.mumbai_source_bucket_name}/*"
      },
      {
        Action = [
          "s3:ReplicateObject",
          "s3:ReplicateDelete",
          "s3:ReplicateTags"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:s3:::${var.hyderabad_destination_bucket_name}/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "s3_crr_attachment" {
  role       = aws_iam_role.s3_crr_role.name
  policy_arn = aws_iam_policy.s3_crr_policy.arn
}

# Standby Aurora PostgreSQL Secondary Cluster in Hyderabad (ap-south-2)
resource "aws_rds_cluster" "aurora_hyderabad_standby" {
  count                               = var.enable_aurora_standby ? 1 : 0
  cluster_identifier                  = "${var.project_prefix}-aurora-hyderabad-standby"
  engine                              = "aurora-postgresql"
  engine_version                      = var.engine_version
  kms_key_id                          = var.hyderabad_kms_key_arn
  storage_encrypted                   = true
  storage_type                        = "aurora-iopt1"
  global_cluster_identifier           = var.aurora_global_cluster_identifier
  db_subnet_group_name                = var.hyderabad_db_subnet_group_name
  vpc_security_group_ids              = [var.hyderabad_aurora_security_group_id]
  db_cluster_parameter_group_name     = var.hyderabad_cluster_parameter_group_name
  enabled_cloudwatch_logs_exports     = ["postgresql", "upgrade"]
  skip_final_snapshot                 = true
  copy_tags_to_snapshot               = true

  tags = merge(
    var.tags,
    {
      Name        = "${var.project_prefix}-aurora-hyderabad-standby"
      Environment = var.environment
      Region      = "ap-south-2"
      Role        = "warm-standby-aurora"
    }
  )
}

resource "aws_rds_cluster_instance" "aurora_hyderabad_instances" {
  count                               = var.enable_aurora_standby ? var.hyderabad_instance_count : 0
  identifier                          = "${var.project_prefix}-aurora-hyd-node-${count.index + 1}"
  cluster_identifier                  = aws_rds_cluster.aurora_hyderabad_standby[0].id
  instance_class                      = var.hyderabad_instance_class
  engine                              = aws_rds_cluster.aurora_hyderabad_standby[0].engine
  engine_version                      = aws_rds_cluster.aurora_hyderabad_standby[0].engine_version
  db_subnet_group_name                = var.hyderabad_db_subnet_group_name
  publicly_accessible                = false
  auto_minor_version_upgrade          = false
  performance_insights_enabled        = true
  performance_insights_retention_period = 731
  performance_insights_kms_key_id     = var.hyderabad_kms_key_arn

  tags = merge(
    var.tags,
    {
      Name   = "${var.project_prefix}-aurora-hyd-node-${count.index + 1}"
      Region = "ap-south-2"
      Role   = "standby-read-replica"
    }
  )
}

# Cross-Region Replication Lag Alarm
resource "aws_cloudwatch_metric_alarm" "cross_region_replication_lag" {
  count               = var.enable_aurora_standby ? 1 : 0
  alarm_name          = "${var.project_prefix}-cross-region-aurora-lag-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "AuroraGlobalDBReplicationLag"
  namespace           = "AWS/RDS"
  period              = 60
  statistic           = "Maximum"
  threshold           = 1000 # 1 second in milliseconds
  alarm_description   = "Aurora Cross-Region Replication lag between Mumbai and Hyderabad exceeded 1s (violating RPO SLA)"
  alarm_actions       = var.alarm_action_arns

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.aurora_hyderabad_standby[0].id
  }
}
