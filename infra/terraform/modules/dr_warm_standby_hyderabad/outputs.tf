output "arc_cluster_arn" {
  description = "ARN of the Route 53 Application Recovery Controller (ARC) Cluster"
  value       = aws_route53recoverycontrolconfig_cluster.dr_arc_cluster.arn
}

output "mumbai_routing_control_arn" {
  description = "Routing control ARN for Mumbai (ap-south-1) primary region"
  value       = aws_route53recoverycontrolconfig_routing_control.mumbai_primary_routing_control.arn
}

output "hyderabad_routing_control_arn" {
  description = "Routing control ARN for Hyderabad (ap-south-2) standby region"
  value       = aws_route53recoverycontrolconfig_routing_control.hyderabad_standby_routing_control.arn
}

output "s3_crr_role_arn" {
  description = "IAM Role ARN used for cross-region S3 audit log replication"
  value       = aws_iam_role.s3_crr_role.arn
}

output "hyderabad_aurora_cluster_id" {
  description = "Cluster identifier for secondary Aurora cluster in Hyderabad"
  value       = var.enable_aurora_standby ? aws_rds_cluster.aurora_hyderabad_standby[0].id : null
}

output "hyderabad_aurora_reader_endpoint" {
  description = "Reader endpoint for standby Aurora cluster in Hyderabad"
  value       = var.enable_aurora_standby ? aws_rds_cluster.aurora_hyderabad_standby[0].reader_endpoint : null
}
