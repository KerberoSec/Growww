output "vpc_id" {
  description = "ID of the Besu VPC"
  value       = aws_vpc.besu_vpc.id
}

output "subnet_ids" {
  description = "IDs of the Besu subnets"
  value       = aws_subnet.besu_subnets[*].id
}

output "validator_security_group_id" {
  description = "ID of the validator security group"
  value       = aws_security_group.besu_validator_sg.id
}

output "ebs_volume_ids" {
  description = "IDs of the persistent EBS volumes for validators"
  value       = aws_ebs_volume.validator_storage[*].id
}
