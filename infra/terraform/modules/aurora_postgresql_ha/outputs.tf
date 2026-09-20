output "cluster_id" {
  description = "The ID of the Aurora PostgreSQL Cluster"
  value       = aws_rds_cluster.aurora_cluster.id
}

output "cluster_arn" {
  description = "Amazon Resource Name (ARN) of the Aurora cluster"
  value       = aws_rds_cluster.aurora_cluster.arn
}

output "endpoint" {
  description = "Primary writer endpoint for the Aurora cluster"
  value       = aws_rds_cluster.aurora_cluster.endpoint
}

output "reader_endpoint" {
  description = "Read-only load-balanced endpoint for read replicas"
  value       = aws_rds_cluster.aurora_cluster.reader_endpoint
}

output "port" {
  description = "Port the PostgreSQL database listens on (5432)"
  value       = aws_rds_cluster.aurora_cluster.port
}

output "database_name" {
  description = "Default database name"
  value       = aws_rds_cluster.aurora_cluster.database_name
}

output "master_username" {
  description = "Master username for database admin access"
  value       = aws_rds_cluster.aurora_cluster.master_username
}

output "master_user_secret_arn" {
  description = "AWS Secrets Manager Secret ARN containing the automatically managed master password"
  value       = aws_rds_cluster.aurora_cluster.master_user_secret[0].secret_arn
}

output "security_group_id" {
  description = "Security Group ID protecting the Aurora cluster"
  value       = aws_security_group.aurora_sg.id
}

output "instance_endpoints" {
  description = "Individual instance connection endpoints"
  value       = aws_rds_cluster_instance.aurora_instances[*].endpoint
}
