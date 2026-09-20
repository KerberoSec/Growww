/**
 * Growww / NBSE Institutional Exchange Infrastructure
 * Module: Terraform AWS Aurora PostgreSQL High-Availability (HA) Multi-AZ Cluster
 * Compliance: SEBI, RBI, IFSCA Sovereign Ledger & ACID Resiliency Standard
 */

resource "aws_db_subnet_group" "aurora_subnet_group" {
  name        = "${var.cluster_identifier}-subnet-group"
  description = "Subnet group for Aurora PostgreSQL Multi-AZ cluster ${var.cluster_identifier}"
  subnet_ids  = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name        = "${var.cluster_identifier}-subnet-group"
      Environment = var.environment
    }
  )
}

resource "aws_security_group" "aurora_sg" {
  name        = "${var.cluster_identifier}-sg"
  description = "Security group for Growww Aurora PostgreSQL Cluster"
  vpc_id      = var.vpc_id

  ingress {
    description     = "PostgreSQL Client Ingress (5432)"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    cidr_blocks     = var.allowed_cidr_blocks
    security_groups = var.allowed_security_group_ids
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.cluster_identifier}-sg"
      Environment = var.environment
    }
  )
}

resource "aws_rds_cluster_parameter_group" "aurora_cluster_pg" {
  name        = "${var.cluster_identifier}-cluster-pg"
  family      = "aurora-postgresql16"
  description = "Institutional trading and settlement parameter group for ${var.cluster_identifier}"

  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements,auto_explain,pgaudit"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "synchronous_commit"
    value = "on"
  }

  parameter {
    name  = "wal_compression"
    value = "lz4"
  }

  parameter {
    name  = "random_page_cost"
    value = "1.1"
  }

  parameter {
    name  = "statement_timeout"
    value = "30000"
  }

  parameter {
    name  = "idle_in_transaction_session_timeout"
    value = "10000"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "250"
  }

  parameter {
    name  = "rds.force_ssl"
    value = "1"
  }

  parameter {
    name  = "pgaudit.log"
    value = "ddl,role,write"
  }

  parameter {
    name  = "auto_explain.log_min_duration"
    value = "1000"
  }

  parameter {
    name  = "pg_stat_statements.track"
    value = "all"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_parameter_group" "aurora_db_pg" {
  name        = "${var.cluster_identifier}-db-pg"
  family      = "aurora-postgresql16"
  description = "Instance parameter group for Aurora PostgreSQL instances"

  parameter {
    name  = "work_mem"
    value = "65536" # 64MB
  }

  parameter {
    name  = "maintenance_work_mem"
    value = "2097152" # 2GB
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_iam_role" "rds_enhanced_monitoring" {
  name = "${var.cluster_identifier}-rds-enhanced-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "rds_enhanced_monitoring_attach" {
  role       = aws_iam_role.rds_enhanced_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

resource "aws_rds_cluster" "aurora_cluster" {
  cluster_identifier                  = var.cluster_identifier
  engine                              = "aurora-postgresql"
  engine_version                      = var.engine_version
  database_name                       = var.database_name
  master_username                     = var.master_username
  manage_master_user_password         = true
  master_user_secret_kms_key_id       = var.kms_key_arn
  kms_key_id                          = var.kms_key_arn
  storage_encrypted                   = true
  storage_type                        = "aurora-iopt1" # I/O-Optimized tier
  backup_retention_period             = var.backup_retention_period
  preferred_backup_window             = "02:00-03:00"
  preferred_maintenance_window        = "sun:03:30-sun:04:30"
  db_subnet_group_name                = aws_db_subnet_group.aurora_subnet_group.name
  vpc_security_group_ids              = [aws_security_group.aurora_sg.id]
  db_cluster_parameter_group_name     = aws_rds_cluster_parameter_group.aurora_cluster_pg.name
  enabled_cloudwatch_logs_exports     = ["postgresql", "upgrade"]
  deletion_protection                 = var.enable_deletion_protection
  copy_tags_to_snapshot               = true
  skip_final_snapshot                 = var.skip_final_snapshot
  final_snapshot_identifier           = "${var.cluster_identifier}-final-snapshot"
  backtrack_window                    = 0
  allow_major_version_upgrade         = false

  dynamic "serverlessv2_scaling_configuration" {
    for_each = var.enable_serverless_v2 ? [1] : []
    content {
      min_capacity = var.serverless_min_acu
      max_capacity = var.serverless_max_acu
    }
  }

  tags = merge(
    var.tags,
    {
      Name        = var.cluster_identifier
      Environment = var.environment
      Role        = "aurora-postgresql-ha"
      Compliance  = "SEBI-RBI-IFSCA"
    }
  )

  lifecycle {
    ignore_changes = [
      availability_zones
    ]
  }
}

resource "aws_rds_cluster_instance" "aurora_instances" {
  count                               = var.instance_count
  identifier                          = "${var.cluster_identifier}-node-${count.index + 1}"
  cluster_identifier                  = aws_rds_cluster.aurora_cluster.id
  instance_class                      = var.enable_serverless_v2 ? "db.serverless" : var.instance_class
  engine                              = aws_rds_cluster.aurora_cluster.engine
  engine_version                      = aws_rds_cluster.aurora_cluster.engine_version
  db_subnet_group_name                = aws_db_subnet_group.aurora_subnet_group.name
  db_parameter_group_name             = aws_db_parameter_group.aurora_db_pg.name
  publicly_accessible                = false
  auto_minor_version_upgrade          = false
  promotion_tier                      = count.index
  performance_insights_enabled        = true
  performance_insights_retention_period = 731
  performance_insights_kms_key_id     = var.kms_key_arn
  monitoring_interval                 = 1 # 1-second enhanced monitoring
  monitoring_role_arn                 = aws_iam_role.rds_enhanced_monitoring.arn

  tags = merge(
    var.tags,
    {
      Name        = "${var.cluster_identifier}-node-${count.index + 1}"
      Environment = var.environment
      Role        = count.index == 0 ? "primary-writer" : "read-replica"
    }
  )
}

# CloudWatch Alarms for Aurora Health, CPU & Replica Lag
resource "aws_cloudwatch_metric_alarm" "aurora_replica_lag" {
  alarm_name          = "${var.cluster_identifier}-replica-lag-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "AuroraReplicaLag"
  namespace           = "AWS/RDS"
  period              = 60
  statistic           = "Maximum"
  threshold           = 50 # Milliseconds
  alarm_description   = "Aurora PostgreSQL replica lag exceeded 50ms - potential replication divergence"
  alarm_actions       = var.alarm_action_arns
  ok_actions          = var.alarm_action_arns

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.aurora_cluster.id
  }
}

resource "aws_cloudwatch_metric_alarm" "aurora_cpu_high" {
  alarm_name          = "${var.cluster_identifier}-cpu-high"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Aurora PostgreSQL cluster CPU exceeded 80% for 3 minutes"
  alarm_actions       = var.alarm_action_arns

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.aurora_cluster.id
  }
}

resource "aws_cloudwatch_metric_alarm" "aurora_freeable_memory_low" {
  alarm_name          = "${var.cluster_identifier}-freeable-memory-low"
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "FreeableMemory"
  namespace           = "AWS/RDS"
  period              = 60
  statistic           = "Average"
  threshold           = 1073741824 # 1GB in bytes
  alarm_description   = "Aurora PostgreSQL cluster freeable memory low (<1GB)"
  alarm_actions       = var.alarm_action_arns

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.aurora_cluster.id
  }
}
