/**
 * Growww / NBSE Institutional Exchange Infrastructure
 * Module: Terraform AWS MSK (Managed Streaming for Apache Kafka) Cluster
 * Compliance: SEBI, RBI, IFSCA WORM Audit & High-Availability Resiliency Standard
 */

resource "aws_msk_configuration" "msk_config" {
  name           = "${var.cluster_name}-config"
  kafka_versions = [var.kafka_version]
  description    = "Ultra-low latency institutional exchange broker configuration for ${var.cluster_name}"

  server_properties = join("\n", [
    "auto.create.topics.enable=false",
    "default.replication.factor=3",
    "min.insync.replicas=2",
    "unclean.leader.election.enable=false",
    "compression.type=lz4",
    "log.retention.hours=168",
    "log.roll.hours=24",
    "log.segment.bytes=1073741824",
    "log.cleaner.enable=true",
    "log.cleanup.policy=delete",
    "message.max.bytes=10485760",
    "replica.fetch.max.bytes=10485760",
    "replica.fetch.response.max.bytes=52428800",
    "num.partitions=12",
    "num.replica.fetchers=4",
    "num.network.threads=8",
    "num.io.threads=16",
    "socket.request.max.bytes=104857600",
    "socket.send.buffer.bytes=1048576",
    "socket.receive.buffer.bytes=1048576",
    "zookeeper.connection.timeout.ms=18000",
    "connections.max.idle.ms=30000",
    "log.flush.interval.messages=10000",
    "log.flush.interval.ms=1000"
  ])

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group" "msk_sg" {
  name        = "${var.cluster_name}-sg"
  description = "Security group for Growww NBSE MSK Kafka Cluster"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Kafka TLS Client Ingress (9094)"
    from_port       = 9094
    to_port         = 9094
    protocol        = "tcp"
    cidr_blocks     = var.allowed_cidr_blocks
    security_groups = var.allowed_security_group_ids
  }

  ingress {
    description     = "Kafka SASL/SCRAM Client Ingress (9096)"
    from_port       = 9096
    to_port         = 9096
    protocol        = "tcp"
    cidr_blocks     = var.allowed_cidr_blocks
    security_groups = var.allowed_security_group_ids
  }

  ingress {
    description     = "Prometheus JMX Exporter (11001)"
    from_port       = 11001
    to_port         = 11001
    protocol        = "tcp"
    cidr_blocks     = var.monitoring_cidr_blocks
    security_groups = var.allowed_security_group_ids
  }

  ingress {
    description     = "Prometheus Node Exporter (11002)"
    from_port       = 11002
    to_port         = 11002
    protocol        = "tcp"
    cidr_blocks     = var.monitoring_cidr_blocks
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
      Name        = "${var.cluster_name}-sg"
      Environment = var.environment
      Role        = "msk-security-group"
    }
  )
}

resource "aws_cloudwatch_log_group" "msk_logs" {
  count             = var.enable_cloudwatch_logging ? 1 : 0
  name              = "/aws/msk/${var.cluster_name}"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_arn

  tags = merge(
    var.tags,
    {
      Name        = "${var.cluster_name}-logs"
      Environment = var.environment
    }
  )
}

resource "aws_msk_cluster" "msk_cluster" {
  cluster_name           = var.cluster_name
  kafka_version          = var.kafka_version
  number_of_broker_nodes = var.number_of_broker_nodes

  broker_node_group_info {
    instance_type   = var.broker_instance_type
    client_subnets  = var.subnet_ids
    security_groups = [aws_security_group.msk_sg.id]

    storage_info {
      ebs_storage_info {
        volume_size = var.ebs_volume_size_gb

        provisioned_throughput {
          enabled           = var.ebs_provisioned_throughput_enabled
          volume_throughput = var.ebs_provisioned_throughput_mb_per_sec
        }
      }
    }

    connectivity_info {
      public_access {
        type = "DISABLED"
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.msk_config.arn
    revision = aws_msk_configuration.msk_config.latest_revision
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = var.kms_key_arn

    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  client_authentication {
    sasl {
      scram = true
      iam   = var.enable_sasl_iam
    }
    unauthenticated = false
  }

  open_monitoring {
    prometheus {
      jmx_exporter {
        enabled_in_broker = true
      }
      node_exporter {
        enabled_in_broker = true
      }
    }
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = var.enable_cloudwatch_logging
        log_group = var.enable_cloudwatch_logging ? aws_cloudwatch_log_group.msk_logs[0].name : null
      }
      s3 {
        enabled = var.enable_s3_logging
        bucket  = var.s3_logging_bucket_name
        prefix  = var.s3_logging_prefix
      }
    }
  }

  tags = merge(
    var.tags,
    {
      Name        = var.cluster_name
      Environment = var.environment
      ManagedBy   = "Terraform"
      Compliance  = "SEBI-RBI-IFSCA"
    }
  )

  depends_on = [
    aws_msk_configuration.msk_config,
    aws_cloudwatch_log_group.msk_logs
  ]
}

resource "aws_msk_scram_secret_association" "scram_association" {
  count           = length(var.scram_secret_association_arns) > 0 ? 1 : 0
  cluster_arn     = aws_msk_cluster.msk_cluster.arn
  secret_arn_list = var.scram_secret_association_arns
}

# CloudWatch Alarms for Broker Health & Partition Safety
resource "aws_cloudwatch_metric_alarm" "under_replicated_partitions" {
  alarm_name          = "${var.cluster_name}-under-replicated-partitions"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "UnderReplicatedPartitions"
  namespace           = "AWS/Kafka"
  period              = 60
  statistic           = "Maximum"
  threshold           = 0
  alarm_description   = "Kafka under-replicated partitions detected in cluster ${var.cluster_name} - potential data loss or broker failure"
  alarm_actions       = var.alarm_action_arns
  ok_actions          = var.alarm_action_arns

  dimensions = {
    "Cluster Name" = var.cluster_name
  }
}

resource "aws_cloudwatch_metric_alarm" "offline_partitions" {
  alarm_name          = "${var.cluster_name}-offline-partitions"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "OfflinePartitionsCount"
  namespace           = "AWS/Kafka"
  period              = 60
  statistic           = "Maximum"
  threshold           = 0
  alarm_description   = "Kafka offline partitions count > 0 in cluster ${var.cluster_name} - immediate trading partition halt"
  alarm_actions       = var.alarm_action_arns
  ok_actions          = var.alarm_action_arns

  dimensions = {
    "Cluster Name" = var.cluster_name
  }
}

resource "aws_cloudwatch_metric_alarm" "broker_cpu_high" {
  alarm_name          = "${var.cluster_name}-broker-cpu-high"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 3
  metric_name         = "CpuUser"
  namespace           = "AWS/Kafka"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "MSK Broker CPU User utilization exceeded 80% for 3 minutes"
  alarm_actions       = var.alarm_action_arns

  dimensions = {
    "Cluster Name" = var.cluster_name
  }
}

resource "aws_cloudwatch_metric_alarm" "broker_disk_free_low" {
  alarm_name          = "${var.cluster_name}-broker-disk-free-low"
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "KafkaDataLogsDiskUsed"
  namespace           = "AWS/Kafka"
  period              = 300
  statistic           = "Maximum"
  threshold           = 85
  alarm_description   = "MSK Broker Disk utilization exceeded 85% - auto-expansion required"
  alarm_actions       = var.alarm_action_arns

  dimensions = {
    "Cluster Name" = var.cluster_name
  }
}
