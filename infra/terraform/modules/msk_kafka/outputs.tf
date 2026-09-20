output "cluster_arn" {
  description = "Amazon Resource Name (ARN) of the MSK Kafka cluster"
  value       = aws_msk_cluster.msk_cluster.arn
}

output "cluster_name" {
  description = "Name of the MSK Kafka cluster"
  value       = aws_msk_cluster.msk_cluster.cluster_name
}

output "bootstrap_brokers_tls" {
  description = "TLS encrypted connection host:port pairs for Kafka brokers"
  value       = aws_msk_cluster.msk_cluster.bootstrap_brokers_tls
}

output "bootstrap_brokers_sasl_scram" {
  description = "SASL/SCRAM encrypted connection host:port pairs for Kafka brokers"
  value       = aws_msk_cluster.msk_cluster.bootstrap_brokers_sasl_scram
}

output "bootstrap_brokers_sasl_iam" {
  description = "SASL/IAM connection host:port pairs for Kafka brokers"
  value       = aws_msk_cluster.msk_cluster.bootstrap_brokers_sasl_iam
}

output "zookeeper_connect_string" {
  description = "Apache ZooKeeper connection host:port string"
  value       = aws_msk_cluster.msk_cluster.zookeeper_connect_string
}

output "zookeeper_connect_string_tls" {
  description = "Apache ZooKeeper TLS connection host:port string"
  value       = aws_msk_cluster.msk_cluster.zookeeper_connect_string_tls
}

output "security_group_id" {
  description = "Security group ID attached to MSK brokers"
  value       = aws_security_group.msk_sg.id
}

output "configuration_arn" {
  description = "ARN of the custom MSK configuration"
  value       = aws_msk_configuration.msk_config.arn
}

output "configuration_latest_revision" {
  description = "Latest revision number of the custom MSK configuration"
  value       = aws_msk_configuration.msk_config.latest_revision
}

output "cloudwatch_log_group_name" {
  description = "CloudWatch log group name for MSK broker logs"
  value       = var.enable_cloudwatch_logging ? aws_cloudwatch_log_group.msk_logs[0].name : null
}
